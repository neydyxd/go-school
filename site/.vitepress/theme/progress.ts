// Прогресс курса в localStorage. Чистые функции: вход не мутируется,
// любые сбои хранилища (нет storage, приватный режим, квота) глотаются.

export const KEY = 'go-school:progress'

export type Progress = {
  v: 1
  lessons: Record<string, { done: boolean; at: string }>
  hw: Record<string, boolean>
}

export const empty = (): Progress => ({ v: 1, lessons: {}, hw: {} })

const isObj = (x: unknown): x is Record<string, unknown> =>
  typeof x === 'object' && x !== null && !Array.isArray(x)

function validate(x: unknown): Progress | string {
  if (!isObj(x)) return 'Файл не похож на прогресс курса: ожидался JSON-объект'
  if (x.v !== 1) return `Неизвестная версия формата: ${String(x.v)}`
  if (!isObj(x.lessons) || !isObj(x.hw)) return 'Поля lessons и hw должны быть объектами'
  for (const [id, l] of Object.entries(x.lessons)) {
    if (!isObj(l) || typeof l.done !== 'boolean' || typeof l.at !== 'string')
      return `Урок ${id}: ожидались поля done (boolean) и at (строка)`
  }
  for (const [id, v] of Object.entries(x.hw)) {
    if (typeof v !== 'boolean') return `Домашка ${id}: значение должно быть true/false`
  }
  return x as Progress
}

export function load(storage: Storage | null): Progress {
  try {
    const raw = storage?.getItem(KEY)
    if (!raw) return empty()
    const p = validate(JSON.parse(raw))
    return typeof p === 'string' ? empty() : p
  } catch {
    return empty()
  }
}

export function save(storage: Storage | null, p: Progress): void {
  try {
    storage?.setItem(KEY, JSON.stringify(p))
  } catch {
    // ponytail: квота или приватный режим, прогресс живёт только в памяти вкладки
  }
}

export function toggleLesson(p: Progress, id: string, now: string): Progress {
  const lessons = { ...p.lessons }
  if (lessons[id]?.done) delete lessons[id]
  else lessons[id] = { done: true, at: now }
  return { ...p, lessons }
}

export function toggleHw(p: Progress, id: string): Progress {
  const hw = { ...p.hw }
  if (hw[id]) delete hw[id]
  else hw[id] = true
  return { ...p, hw }
}

export const exportJson = (p: Progress): string => JSON.stringify(p, null, 2)

export function importJson(text: string): { ok: true; p: Progress } | { ok: false; error: string } {
  let data: unknown
  try {
    data = JSON.parse(text)
  } catch {
    return { ok: false, error: 'Это не JSON: файл повреждён или выбран не тот файл' }
  }
  const p = validate(data)
  return typeof p === 'string' ? { ok: false, error: p } : { ok: true, p }
}
