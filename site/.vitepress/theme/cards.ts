// Карточки собеса: разбор блоков ::: interview и состояние «знаю / повторить» в localStorage.
// Чистые функции: вход не мутируется, сбои хранилища глотаются (как в progress.ts).

export const KEY = 'go-school:cards'

export type CardMark = { known: boolean; at: number; seen: number }
export type CardsState = { v: 1; cards: Record<string, CardMark> }

// Карточка банка: q и a уже отрендерены в HTML на этапе сборки (cards.data.ts).
export type Card = { id: string; m: string; lesson: string; title: string; url: string; q: string; a: string }

export type QA = { q: string; a: string }

export const empty = (): CardsState => ({ v: 1, cards: {} })

const isObj = (x: unknown): x is Record<string, unknown> =>
  typeof x === 'object' && x !== null && !Array.isArray(x)

const isMark = (x: unknown): x is CardMark =>
  isObj(x) && typeof x.known === 'boolean' && Number.isFinite(x.at) && Number.isInteger(x.seen) && (x.seen as number) >= 0

// Чужая схема целиком даёт пустое состояние; отдельные битые записи просто отбрасываются.
export function parseCards(x: unknown): CardsState | null {
  if (!isObj(x) || x.v !== 1 || !isObj(x.cards)) return null
  const cards: Record<string, CardMark> = {}
  for (const [id, m] of Object.entries(x.cards)) if (isMark(m)) cards[id] = { known: m.known, at: m.at, seen: m.seen }
  return { v: 1, cards }
}

export function load(storage: Storage | null): CardsState {
  try {
    const raw = storage?.getItem(KEY)
    return (raw && parseCards(JSON.parse(raw))) || empty()
  } catch {
    return empty()
  }
}

export function save(storage: Storage | null, s: CardsState): void {
  try {
    storage?.setItem(KEY, JSON.stringify(s))
  } catch {
    // ponytail: квота или приватный режим, отметки живут только до перезагрузки
  }
}

export function mark(s: CardsState, id: string, known: boolean, now: number): CardsState {
  const seen = (s.cards[id]?.seen ?? 0) + 1
  return { ...s, cards: { ...s.cards, [id]: { known, at: now, seen } } }
}

export const isWeak = (s: CardsState, id: string) => s.cards[id]?.known === false

export function counts(cards: { id: string }[], s: CardsState) {
  let known = 0
  let weak = 0
  for (const c of cards) {
    const m = s.cards[c.id]
    if (m?.known) known++
    else if (m) weak++
  }
  return { total: cards.length, known, weak, fresh: cards.length - known - weak }
}

// Колода: карточки выбранных модулей (пустой список = все), в режиме weakOnly только «повторить».
export function pickDeck<T extends { id: string; m: string }>(
  cards: T[],
  s: CardsState,
  opts: { modules: string[]; weakOnly: boolean },
): T[] {
  return cards.filter(
    (c) => (opts.modules.length === 0 || opts.modules.includes(c.m)) && (!opts.weakOnly || isWeak(s, c.id)),
  )
}

// Что показать вместо карточки и какие кнопки дать. weakLeft: сколько карточек
// «повторить» осталось в выбранных модулях (а не в пройденном круге).
export type View =
  | { kind: 'card' }
  | { kind: 'no-cards' }
  | { kind: 'no-weak'; showAll: true }
  | { kind: 'done'; again: boolean; onlyWeak: boolean; showAll: boolean }

export function view(o: { deckLen: number; idx: number; weakOnly: boolean; weakLeft: number }): View {
  if (o.deckLen === 0) return o.weakOnly ? { kind: 'no-weak', showAll: true } : { kind: 'no-cards' }
  if (o.idx < o.deckLen) return { kind: 'card' }
  // в режиме «только повторить» новый круг возможен, только пока слабые остались
  const again = !o.weakOnly || o.weakLeft > 0
  return { kind: 'done', again, onlyWeak: !o.weakOnly && o.weakLeft > 0, showAll: o.weakOnly }
}

// Фишер-Йетс на копии; rand подменяется в тестах.
export function shuffle<T>(xs: readonly T[], rand: () => number = Math.random): T[] {
  const out = [...xs]
  for (let i = out.length - 1; i > 0; i--) {
    const j = Math.floor(rand() * (i + 1))
    ;[out[i], out[j]] = [out[j], out[i]]
  }
  return out
}

// FNV-1a 32 бита: короткий стабильный хеш текста вопроса для id карточки.
export function hash(text: string): string {
  let h = 0x811c9dc5
  for (let i = 0; i < text.length; i++) {
    h ^= text.charCodeAt(i)
    h = Math.imul(h, 0x01000193)
  }
  return (h >>> 0).toString(36).padStart(7, '0')
}

// id = модуль-урок-хеш вопроса: m01-l05-0abc123. Пробелы в вопросе нормализуются.
export const cardId = (m: string, slug: string, q: string) =>
  `${m}-${slug.slice(0, 3)}-${hash(q.trim().replace(/\s+/g, ' '))}`

// Ответ карточки показывается на странице l01, а не в своём уроке, поэтому
// относительные ссылки (./l03, ../m07/l05) переписываем в абсолютные от урока-источника.
export function absLinks(html: string, pageUrl: string, base = '/'): string {
  return html.replace(/href="(\.{1,2}\/[^"]*)"/g, (_, href: string) => {
    const u = new URL(href, `http://x${pageUrl}`)
    return `href="${(base + u.pathname).replace(/\/+/g, '/')}${u.search}${u.hash}"`
  })
}

const FENCE = /^\s*(`{3,}|~{3,})/
const OPEN = /^:::+\s*\S/
const CLOSE = /^:::+\s*$/
const Q_LINE = /^<details><summary>(.*)<\/summary>\s*$/
const A_END = /^<\/details>\s*$/

// Достаёт пары «вопрос / ответ» (сырой markdown) из всех блоков ::: interview урока.
// Внутри ограждённого кода ::: и </details> не считаются разметкой.
export function parseInterview(src: string): QA[] {
  const out: QA[] = []
  let fence = '' // открытый маркер ``` или ~~~, пусто вне кода
  let depth = 0 // вложенность контейнеров внутри interview
  let inBlock = false
  let q: string | null = null
  let a: string[] = []
  for (const line of src.split(/\r?\n/)) {
    const f = line.match(FENCE)
    if (fence) {
      if (f && f[1][0] === fence[0] && f[1].length >= fence.length && line.trim() === f[1]) fence = ''
      if (q !== null) a.push(line)
      continue
    }
    if (f) {
      fence = f[1]
      if (q !== null) a.push(line)
      continue
    }
    if (!inBlock) {
      if (/^:::+\s*interview\b/.test(line)) inBlock = true
      continue
    }
    if (q === null) {
      const m = line.match(Q_LINE)
      if (m) q = m[1].trim()
      else if (OPEN.test(line)) depth++
      else if (CLOSE.test(line) && depth-- === 0) {
        inBlock = false
        depth = 0
      }
      continue
    }
    if (A_END.test(line)) {
      out.push({ q, a: a.join('\n').trim() })
      q = null
      a = []
    } else a.push(line)
  }
  return out
}
