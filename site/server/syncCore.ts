// Логика синхронизации без привязки к конкретной базе: api/sync.ts даёт ей Neon,
// тесты дают память. Хранилище рассчитано на одного владельца: первый ключ
// занимает его, остальные получают 403. Сам ключ не хранится, только SHA-256.

import { createHash, timingSafeEqual } from 'node:crypto'
import { mergeDoc, type SyncDoc } from '../.vitepress/theme/sync.js'

export type Row = { keyHash: string; doc: SyncDoc; version: number }

export type Db = {
  claim(keyHash: string, doc: SyncDoc): Promise<void> // вставить, если хранилище пустое
  read(): Promise<Row | null>
  update(doc: SyncDoc, version: number): Promise<boolean> // false: версия уже сменилась
}

export type Result = { status: 200; doc: SyncDoc } | { status: 401 | 403 | 409; error: string }

const KEY_RE = /^[A-Za-z0-9_-]{43}$/ // 32 случайных байта в base64url
const RETRIES = 5

export const hashKey = (key: string) => createHash('sha256').update(key).digest('hex')

const sameHash = (a: string, b: string) =>
  a.length === b.length && timingSafeEqual(Buffer.from(a), Buffer.from(b))

export async function syncState(db: Db, key: string, incoming: SyncDoc): Promise<Result> {
  if (!KEY_RE.test(key)) return { status: 401, error: 'нет ключа синхронизации' }
  const hash = hashKey(key)
  await db.claim(hash, incoming)
  for (let i = 0; i < RETRIES; i++) {
    const row = await db.read()
    if (!row || !sameHash(row.keyHash, hash)) return { status: 403, error: 'ключ не подходит' }
    const merged = mergeDoc(row.doc, incoming)
    if (await db.update(merged, row.version)) return { status: 200, doc: merged }
  }
  return { status: 409, error: 'слишком много одновременных записей, повтори позже' }
}
