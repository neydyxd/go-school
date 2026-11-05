// Слияние состояния курса между устройствами. Чистые функции, общие для браузера
// и серверной функции api/sync.ts. Правило одно: у каждой отметки есть время,
// более поздняя побеждает; при равном времени результат не зависит от порядка.

import { parseProgress, type Mark, type Progress } from './progress'
import { parseCards, type CardMark, type CardsState } from './cards'

export type SyncDoc = { progress: Progress; cards: CardsState }

function mergeBy<T>(a: Record<string, T>, b: Record<string, T>, pick: (x: T, y: T) => T): Record<string, T> {
  const out: Record<string, T> = { ...a }
  for (const [id, y] of Object.entries(b)) out[id] = id in out ? pick(out[id], y) : y
  return out
}

const laterMark = (x: Mark, y: Mark): Mark =>
  x.at !== y.at ? (x.at > y.at ? x : y) : x.done ? x : y

const laterCard = (x: CardMark, y: CardMark): CardMark =>
  x.at !== y.at ? (x.at > y.at ? x : y) : x.seen !== y.seen ? (x.seen > y.seen ? x : y) : x.known ? x : y

export const mergeProgress = (a: Progress, b: Progress): Progress => ({
  v: 2,
  lessons: mergeBy(a.lessons, b.lessons, laterMark),
  hw: mergeBy(a.hw, b.hw, laterMark),
})

export const mergeCards = (a: CardsState, b: CardsState): CardsState => ({
  v: 1,
  cards: mergeBy(a.cards, b.cards, laterCard),
})

export const mergeDoc = (a: SyncDoc, b: SyncDoc): SyncDoc => ({
  progress: mergeProgress(a.progress, b.progress),
  cards: mergeCards(a.cards, b.cards),
})

// Проверка документа с сети (клиент и сервер не доверяют друг другу на слово).
export function parseDoc(x: unknown): SyncDoc | null {
  if (typeof x !== 'object' || x === null || Array.isArray(x)) return null
  const d = x as Record<string, unknown>
  const progress = parseProgress(d.progress)
  const cards = parseCards(d.cards)
  if (typeof progress === 'string' || !cards) return null
  return { progress, cards }
}
