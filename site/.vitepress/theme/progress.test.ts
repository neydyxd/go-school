import { describe, expect, it } from 'vitest'
import { empty, exportJson, importJson, KEY, load, save, toggleHw, toggleLesson } from './progress'

function memStorage(init: Record<string, string> = {}): Storage {
  const m = new Map(Object.entries(init))
  return {
    get length() { return m.size },
    clear: () => m.clear(),
    getItem: (k) => m.get(k) ?? null,
    key: (i) => [...m.keys()][i] ?? null,
    removeItem: (k) => { m.delete(k) },
    setItem: (k, v) => { m.set(k, String(v)) },
  }
}

const NOW = '2026-11-04T10:00:00.000Z'

describe('load', () => {
  it('без storage возвращает пустой прогресс', () => {
    expect(load(null)).toEqual(empty())
  })
  it('битый JSON даёт пустой прогресс', () => {
    expect(load(memStorage({ [KEY]: '{oops' }))).toEqual(empty())
  })
  it('чужая версия схемы даёт пустой прогресс', () => {
    expect(load(memStorage({ [KEY]: JSON.stringify({ v: 99, lessons: {}, hw: {} }) }))).toEqual(empty())
  })
  it('getItem бросает (приватный режим) и это не ломает загрузку', () => {
    const s = memStorage()
    s.getItem = () => { throw new Error('denied') }
    expect(load(s)).toEqual(empty())
  })
  it('сохранённое читается обратно', () => {
    const s = memStorage()
    const p = toggleLesson(empty(), 'm01-l05', NOW)
    save(s, p)
    expect(load(s)).toEqual(p)
  })
})

describe('save', () => {
  it('без storage и при исключении молча ничего не делает', () => {
    expect(() => save(null, empty())).not.toThrow()
    const s = memStorage()
    s.setItem = () => { throw new Error('quota') }
    expect(() => save(s, empty())).not.toThrow()
  })
})

describe('toggle', () => {
  it('toggleLesson отмечает и снимает отметку, не мутируя вход', () => {
    const a = empty()
    const b = toggleLesson(a, 'm01-l05', NOW)
    expect(a.lessons).toEqual({})
    expect(b.lessons['m01-l05']).toEqual({ done: true, at: NOW })
    const c = toggleLesson(b, 'm01-l05', NOW)
    expect(c.lessons['m01-l05']).toBeUndefined()
    expect(b.lessons['m01-l05']?.done).toBe(true)
  })
  it('toggleHw переключает и не мутирует вход', () => {
    const a = empty()
    const b = toggleHw(a, 'm01-l05/ex1')
    expect(a.hw).toEqual({})
    expect(b.hw['m01-l05/ex1']).toBe(true)
    expect(toggleHw(b, 'm01-l05/ex1').hw['m01-l05/ex1']).toBeUndefined()
  })
})

describe('import/export', () => {
  it('export -> import даёт тот же прогресс', () => {
    const p = toggleHw(toggleLesson(empty(), 'm02-l01', NOW), 'm02-l01/ex2')
    const r = importJson(exportJson(p))
    expect(r).toEqual({ ok: true, p })
  })
  it.each([
    ['не JSON', 'hello'],
    ['не объект', '[1,2]'],
    ['чужая версия', JSON.stringify({ v: 2, lessons: {}, hw: {} })],
    ['lessons не объект', JSON.stringify({ v: 1, lessons: [], hw: {} })],
    ['урок без done', JSON.stringify({ v: 1, lessons: { a: { at: NOW } }, hw: {} })],
    ['hw не boolean', JSON.stringify({ v: 1, lessons: {}, hw: { a: 'yes' } })],
  ])('отклоняет: %s', (_, text) => {
    const r = importJson(text)
    expect(r.ok).toBe(false)
    if (!r.ok) expect(r.error).toMatch(/[а-я]/i)
  })
})
