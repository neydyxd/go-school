// Прогресс курса в localStorage. Чистые функции: вход не мутируется,
// любые сбои хранилища (нет storage, приватный режим, квота) глотаются.

export const KEY = 'go-school:progress'

// Отметка со временем. Снятая отметка хранится как done:false, а не удаляется:
// иначе при синхронизации другое устройство «воскресит» её.
export type Mark = { done: boolean; at: string }

export type Progress = {
  v: 2
  lessons: Record<string, Mark>
  hw: Record<string, Mark>
}

export const empty = (): Progress => ({ v: 2, lessons: {}, hw: {} })

// v1 хранил домашку как true без времени: при переходе на v2 ставим самое раннее время,
// чтобы любая новая отметка с другого устройства была «позже».
const EPOCH = '1970-01-01T00:00:00.000Z'

const isObj = (x: unknown): x is Record<string, unknown> =>
  typeof x === 'object' && x !== null && !Array.isArray(x)

const isMark = (x: unknown): x is Mark => isObj(x) && typeof x.done === 'boolean' && typeof x.at === 'string'

// Разбор прогресса v1 или v2 в v2. Строка вместо результата: текст ошибки для пользователя.
export function parseProgress(x: unknown): Progress | string {
  if (!isObj(x)) return 'Файл не похож на прогресс курса: ожидался JSON-объект'
  if (x.v !== 1 && x.v !== 2) return `Неизвестная версия формата: ${String(x.v)}`
  if (!isObj(x.lessons) || !isObj(x.hw)) return 'Поля lessons и hw должны быть объектами'
  const lessons: Record<string, Mark> = {}
  for (const [id, l] of Object.entries(x.lessons)) {
    if (!isMark(l)) return `Урок ${id}: ожидались поля done (boolean) и at (строка)`
    lessons[id] = { done: l.done, at: l.at }
  }
  const hw: Record<string, Mark> = {}
  for (const [id, v] of Object.entries(x.hw)) {
    if (x.v === 1) {
      if (typeof v !== 'boolean') return `Домашка ${id}: значение должно быть true/false`
      if (v) hw[id] = { done: true, at: EPOCH }
    } else {
      if (!isMark(v)) return `Домашка ${id}: ожидались поля done (boolean) и at (строка)`
      hw[id] = { done: v.done, at: v.at }
    }
  }
  return { v: 2, lessons, hw }
}

export function load(storage: Storage | null): Progress {
  try {
    const raw = storage?.getItem(KEY)
    if (!raw) return empty()
    const p = parseProgress(JSON.parse(raw))
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
  return { ...p, lessons: { ...p.lessons, [id]: { done: !p.lessons[id]?.done, at: now } } }
}

export function toggleHw(p: Progress, id: string, now: string): Progress {
  return { ...p, hw: { ...p.hw, [id]: { done: !p.hw[id]?.done, at: now } } }
}

export const exportJson = (p: Progress): string => JSON.stringify(p, null, 2)

export function importJson(text: string): { ok: true; p: Progress } | { ok: false; error: string } {
  let data: unknown
  try {
    data = JSON.parse(text)
  } catch {
    return { ok: false, error: 'Это не JSON: файл повреждён или выбран не тот файл' }
  }
  const p = parseProgress(data)
  return typeof p === 'string' ? { ok: false, error: p } : { ok: true, p }
}
