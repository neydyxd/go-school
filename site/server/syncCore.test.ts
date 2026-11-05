import { describe, expect, it } from 'vitest'
import { syncState, type Db, type Row } from './syncCore'
import type { SyncDoc } from '../.vitepress/theme/sync'

// Память вместо базы: проверяем логику доступа и слияния, а не SQL.
function memDb(): Db & { row: Row | null; failUpdates: number } {
  const db = {
    row: null as Row | null,
    failUpdates: 0,
    async claim(keyHash: string, doc: SyncDoc) {
      if (!db.row) db.row = { keyHash, doc, version: 1 }
    },
    async read() {
      return db.row
    },
    async update(doc: SyncDoc, version: number) {
      if (!db.row || db.row.version !== version) return false
      if (db.failUpdates > 0) {
        db.failUpdates--
        db.row = { ...db.row, version: db.row.version + 1 } // кто-то успел записать раньше
        return false
      }
      db.row = { ...db.row, doc, version: version + 1 }
      return true
    },
  }
  return db
}

const KEY_A = 'a'.repeat(43)
const KEY_B = 'b'.repeat(43)
const T1 = '2026-11-05T10:00:00.000Z'
const T2 = '2026-11-05T11:00:00.000Z'
const doc = (lessons: SyncDoc['progress']['lessons']): SyncDoc => ({
  progress: { v: 2, lessons, hw: {} },
  cards: { v: 1, cards: {} },
})

describe('syncState', () => {
  it('первое устройство занимает хранилище своим ключом', async () => {
    const db = memDb()
    const r = await syncState(db, KEY_A, doc({ x: { done: true, at: T1 } }))
    expect(r).toEqual({ status: 200, doc: doc({ x: { done: true, at: T1 } }) })
    expect(db.row?.keyHash).not.toBe(KEY_A) // ключ хранится только хешем
  })
  it('чужой ключ после занятия получает 403 и ничего не меняет', async () => {
    const db = memDb()
    await syncState(db, KEY_A, doc({ x: { done: true, at: T1 } }))
    const r = await syncState(db, KEY_B, doc({ y: { done: true, at: T1 } }))
    expect(r.status).toBe(403)
    expect(db.row?.doc).toEqual(doc({ x: { done: true, at: T1 } }))
  })
  it('второе устройство с тем же ключом получает слитый прогресс', async () => {
    const db = memDb()
    await syncState(db, KEY_A, doc({ x: { done: true, at: T1 } }))
    const r = await syncState(db, KEY_A, doc({ y: { done: true, at: T2 } }))
    expect(r).toEqual({ status: 200, doc: doc({ x: { done: true, at: T1 }, y: { done: true, at: T2 } }) })
  })
  it('при гонке записи перечитывает и повторяет', async () => {
    const db = memDb()
    await syncState(db, KEY_A, doc({ x: { done: true, at: T1 } }))
    db.failUpdates = 2
    const r = await syncState(db, KEY_A, doc({ y: { done: true, at: T2 } }))
    expect(r.status).toBe(200)
  })
  it('если гонка не кончается, отвечает 409, а не крутится вечно', async () => {
    const db = memDb()
    await syncState(db, KEY_A, doc({}))
    db.failUpdates = 100
    expect((await syncState(db, KEY_A, doc({}))).status).toBe(409)
  })
  it.each([['пустой', ''], ['короткий', 'abc'], ['с недопустимыми символами', '!'.repeat(43)]])(
    'ключ %s отвергается с 401 до обращения к базе',
    async (_, key) => {
      const db = memDb()
      expect((await syncState(db, key, doc({}))).status).toBe(401)
      expect(db.row).toBeNull()
    },
  )
})
