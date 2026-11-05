import { describe, expect, it } from 'vitest'
import { mergeCards, mergeDoc, mergeProgress, parseDoc } from './sync'
import { empty as emptyProgress, type Progress } from './progress'
import { empty as emptyCards } from './cards'

const T1 = '2026-11-05T10:00:00.000Z'
const T2 = '2026-11-05T11:00:00.000Z'

const prog = (lessons: Progress['lessons'] = {}, hw: Progress['hw'] = {}): Progress => ({ v: 2, lessons, hw })

describe('mergeProgress', () => {
  it('объединяет отметки с разных устройств', () => {
    const a = prog({ x: { done: true, at: T1 } })
    const b = prog({ y: { done: true, at: T2 } }, { 'y/ex1': { done: true, at: T2 } })
    expect(mergeProgress(a, b)).toEqual(prog(
      { x: { done: true, at: T1 }, y: { done: true, at: T2 } },
      { 'y/ex1': { done: true, at: T2 } },
    ))
  })
  it('более поздняя отметка побеждает, в том числе снятие', () => {
    const a = prog({ x: { done: true, at: T1 } }, { h: { done: true, at: T1 } })
    const b = prog({ x: { done: false, at: T2 } }, { h: { done: false, at: T2 } })
    expect(mergeProgress(a, b)).toEqual(b)
    expect(mergeProgress(b, a)).toEqual(b)
  })
  it('при равном времени побеждает done:true, результат не зависит от порядка', () => {
    const a = prog({ x: { done: false, at: T1 } })
    const b = prog({ x: { done: true, at: T1 } })
    expect(mergeProgress(a, b)).toEqual(b)
    expect(mergeProgress(b, a)).toEqual(b)
  })
  it('не мутирует входы', () => {
    const a = prog({ x: { done: true, at: T1 } })
    const b = prog({ x: { done: false, at: T2 } })
    mergeProgress(a, b)
    expect(a.lessons.x).toEqual({ done: true, at: T1 })
  })
  it('с пустым даёт исходное', () => {
    const a = prog({ x: { done: true, at: T1 } })
    expect(mergeProgress(a, emptyProgress())).toEqual(a)
  })
})

describe('mergeCards', () => {
  it('позднее время побеждает, при равном больше показов', () => {
    const a = { v: 1 as const, cards: { c1: { known: true, at: 1, seen: 3 }, c2: { known: true, at: 5, seen: 1 } } }
    const b = { v: 1 as const, cards: { c1: { known: false, at: 2, seen: 1 }, c2: { known: false, at: 5, seen: 4 }, c3: { known: true, at: 1, seen: 1 } } }
    const m = mergeCards(a, b)
    expect(m.cards.c1).toEqual({ known: false, at: 2, seen: 1 })
    expect(m.cards.c2).toEqual({ known: false, at: 5, seen: 4 })
    expect(m.cards.c3).toEqual({ known: true, at: 1, seen: 1 })
    expect(mergeCards(b, a)).toEqual(m)
  })
})

describe('mergeDoc и parseDoc', () => {
  it('сливает обе части документа', () => {
    const a = { progress: prog({ x: { done: true, at: T1 } }), cards: emptyCards() }
    const b = { progress: emptyProgress(), cards: { v: 1 as const, cards: { c: { known: true, at: 1, seen: 1 } } } }
    expect(mergeDoc(a, b)).toEqual({ progress: a.progress, cards: b.cards })
  })
  it('parseDoc принимает валидный документ, в том числе прогресс v1', () => {
    const d = parseDoc({ progress: { v: 1, lessons: {}, hw: { h: true } }, cards: { v: 1, cards: {} } })
    expect(d && d.progress.v).toBe(2)
  })
  it.each([
    ['не объект', 'x'],
    ['null', null],
    ['без progress', { cards: { v: 1, cards: {} } }],
    ['битый progress', { progress: { v: 2, lessons: [], hw: {} }, cards: { v: 1, cards: {} } }],
    ['битые cards', { progress: { v: 2, lessons: {}, hw: {} }, cards: { v: 7, cards: {} } }],
  ])('parseDoc отвергает: %s', (_, x) => {
    expect(parseDoc(x)).toBeNull()
  })
  it('parseDoc отбрасывает битые отдельные карточки, а не весь документ', () => {
    const d = parseDoc({ progress: { v: 2, lessons: {}, hw: {} }, cards: { v: 1, cards: { ok: { known: true, at: 1, seen: 1 }, bad: { known: 'yes' } } } })
    expect(d && Object.keys(d.cards.cards)).toEqual(['ok'])
  })
})
