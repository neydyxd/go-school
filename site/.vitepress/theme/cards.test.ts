import { describe, expect, it } from 'vitest'
import { absLinks, cardId, counts, view, empty, hash, KEY, load, mark, parseInterview, pickDeck, save, shuffle } from './cards'

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

const NOW = 1_762_300_000_000

describe('parseInterview', () => {
  it('достаёт вопросы и ответы только из блока interview', () => {
    const src = [
      '# Урок',
      '<details><summary>Не из блока</summary>',
      '',
      'мимо',
      '',
      '</details>',
      '::: interview',
      '<details><summary>Чем `a` отличается от `b`?</summary>',
      '',
      'Первый ответ.',
      '',
      '</details>',
      '<details><summary>Второй?</summary>',
      '',
      '- пункт',
      '',
      '</details>',
      ':::',
      '<details><summary>После блока</summary>',
      '',
      'мимо',
      '',
      '</details>',
    ].join('\n')
    expect(parseInterview(src)).toEqual([
      { q: 'Чем `a` отличается от `b`?', a: 'Первый ответ.' },
      { q: 'Второй?', a: '- пункт' },
    ])
  })

  it('::: и </details> внутри блока кода не считаются разметкой', () => {
    const src = [
      '::: interview',
      '<details><summary>Код?</summary>',
      '',
      '```md',
      ':::',
      '</details>',
      '```',
      '',
      '</details>',
      ':::',
    ].join('\n')
    expect(parseInterview(src)).toEqual([{ q: 'Код?', a: '```md\n:::\n</details>\n```' }])
  })

  it('вложенный контейнер между вопросами не закрывает блок', () => {
    const src = ['::: interview', '::: details x', 'y', ':::', '<details><summary>Q</summary>', '', 'A', '', '</details>', ':::'].join('\n')
    expect(parseInterview(src)).toEqual([{ q: 'Q', a: 'A' }])
  })

  it('урок без блока даёт пустой список', () => {
    expect(parseInterview('# Просто текст\n\n::: hood\nx\n:::')).toEqual([])
  })

  it('CRLF переводы строк понимаются', () => {
    expect(parseInterview('::: interview\r\n<details><summary>Q</summary>\r\n\r\nA\r\n\r\n</details>\r\n:::')).toEqual([{ q: 'Q', a: 'A' }])
  })
})

describe('cardId', () => {
  it('стабилен, содержит модуль и урок, не зависит от лишних пробелов', () => {
    const id = cardId('m01', 'l05-slices', 'Что такое  слайс?')
    expect(id).toMatch(/^m01-l05-[0-9a-z]{7}$/)
    expect(cardId('m01', 'l05-slices', ' Что такое слайс? ')).toBe(id)
  })
  it('разные вопросы дают разные хеши', () => {
    expect(hash('Что такое слайс?')).not.toBe(hash('Что такое мапа?'))
    expect(hash('ab')).not.toBe(hash('ba'))
  })
})

describe('load/save', () => {
  it('без storage и без ключа возвращает пустое состояние', () => {
    expect(load(null)).toEqual(empty())
    expect(load(memStorage())).toEqual(empty())
  })
  it.each([
    ['битый JSON', '{oops'],
    ['не объект', '[1]'],
    ['чужая версия', JSON.stringify({ v: 2, cards: {} })],
    ['cards не объект', JSON.stringify({ v: 1, cards: [] })],
    ['схема прогресса уроков', JSON.stringify({ v: 1, lessons: {}, hw: {} })],
  ])('%s даёт пустое состояние', (_, raw) => {
    expect(load(memStorage({ [KEY]: raw }))).toEqual(empty())
  })
  it('битые записи отбрасываются, целые остаются', () => {
    const raw = JSON.stringify({
      v: 1,
      cards: {
        ok: { known: true, at: NOW, seen: 2 },
        noKnown: { at: NOW, seen: 1 },
        badSeen: { known: false, at: NOW, seen: -1 },
        str: 'yes',
      },
    })
    expect(load(memStorage({ [KEY]: raw }))).toEqual({ v: 1, cards: { ok: { known: true, at: NOW, seen: 2 } } })
  })
  it('getItem и setItem бросают, это не ломает страницу', () => {
    const s = memStorage()
    s.getItem = () => { throw new Error('denied') }
    s.setItem = () => { throw new Error('quota') }
    expect(load(s)).toEqual(empty())
    expect(() => save(s, empty())).not.toThrow()
    expect(() => save(null, empty())).not.toThrow()
  })
  it('сохранённое читается обратно', () => {
    const s = memStorage()
    const st = mark(mark(empty(), 'a', true, NOW), 'b', false, NOW)
    save(s, st)
    expect(load(s)).toEqual(st)
  })
})

describe('mark и counts', () => {
  it('mark считает показы, перезаписывает known и не мутирует вход', () => {
    const a = empty()
    const b = mark(a, 'x', false, NOW)
    const c = mark(b, 'x', true, NOW + 1)
    expect(a.cards).toEqual({})
    expect(b.cards.x).toEqual({ known: false, at: NOW, seen: 1 })
    expect(c.cards.x).toEqual({ known: true, at: NOW + 1, seen: 2 })
  })
  it('counts делит на знаю, повторить и новые', () => {
    const st = mark(mark(empty(), 'a', true, NOW), 'b', false, NOW)
    expect(counts([{ id: 'a' }, { id: 'b' }, { id: 'c' }], st)).toEqual({ total: 3, known: 1, weak: 1, fresh: 1 })
    expect(counts([], st)).toEqual({ total: 0, known: 0, weak: 0, fresh: 0 })
  })
})

describe('pickDeck', () => {
  const cards = [
    { id: 'a', m: 'm01' },
    { id: 'b', m: 'm01' },
    { id: 'c', m: 'm02' },
  ]
  const st = mark(mark(empty(), 'a', true, NOW), 'c', false, NOW)
  it('пустой фильтр модулей значит все', () => {
    expect(pickDeck(cards, st, { modules: [], weakOnly: false }).map((c) => c.id)).toEqual(['a', 'b', 'c'])
  })
  it('фильтр по модулям', () => {
    expect(pickDeck(cards, st, { modules: ['m01'], weakOnly: false }).map((c) => c.id)).toEqual(['a', 'b'])
  })
  it('только слабые: отмеченные «повторить», без новых и знакомых', () => {
    expect(pickDeck(cards, st, { modules: [], weakOnly: true }).map((c) => c.id)).toEqual(['c'])
    expect(pickDeck(cards, st, { modules: ['m01'], weakOnly: true })).toEqual([])
  })
})

describe('shuffle', () => {
  it('возвращает перестановку и не мутирует вход', () => {
    const xs = [1, 2, 3, 4, 5, 6, 7, 8]
    const out = shuffle(xs)
    expect(xs).toEqual([1, 2, 3, 4, 5, 6, 7, 8])
    expect([...out].sort((a, b) => a - b)).toEqual(xs)
  })
  it('детерминирован при заданном rand и реально переставляет', () => {
    const r = () => 0
    expect(shuffle([1, 2, 3, 4], r)).toEqual([2, 3, 4, 1])
    expect(shuffle([], r)).toEqual([])
  })
})

describe('absLinks', () => {
  it('относительные ссылки переписываются от урока-источника, абсолютные и внешние не трогаются', () => {
    const html = '<a href="./l03-db-net-arch#tcp">a</a> <a href="../m07/l05-indexes">b</a> <a href="/m01/l05-slices">c</a> <a href="https://go.dev">d</a>'
    expect(absLinks(html, '/m12/l02-live-coding')).toBe(
      '<a href="/m12/l03-db-net-arch#tcp">a</a> <a href="/m07/l05-indexes">b</a> <a href="/m01/l05-slices">c</a> <a href="https://go.dev">d</a>',
    )
  })
  it('учитывает base сайта', () => {
    expect(absLinks('<a href="./x">x</a>', '/m01/l05-slices', '/school/')).toBe('<a href="/school/m01/x">x</a>')
  })
})

describe('view', () => {
  it('пустая колода: в режиме «повторить» предлагает показать все, иначе просто пусто', () => {
    expect(view({ deckLen: 0, idx: 0, weakOnly: true, weakLeft: 0 })).toEqual({ kind: 'no-weak', showAll: true })
    expect(view({ deckLen: 0, idx: 0, weakOnly: false, weakLeft: 0 })).toEqual({ kind: 'no-cards' })
  })
  it('пока карточки не кончились, показываем карточку', () => {
    expect(view({ deckLen: 3, idx: 2, weakOnly: true, weakLeft: 3 })).toEqual({ kind: 'card' })
  })
  it('конец круга в обычном режиме: ещё круг и переход к слабым, если они есть', () => {
    expect(view({ deckLen: 3, idx: 3, weakOnly: false, weakLeft: 1 })).toEqual({ kind: 'done', again: true, onlyWeak: true, showAll: false })
    expect(view({ deckLen: 3, idx: 3, weakOnly: false, weakLeft: 0 })).toEqual({ kind: 'done', again: true, onlyWeak: false, showAll: false })
  })
  it('конец круга «только повторить»: все слабые выучены, нового круга нет, есть выход', () => {
    expect(view({ deckLen: 2, idx: 2, weakOnly: true, weakLeft: 0 })).toEqual({ kind: 'done', again: false, onlyWeak: false, showAll: true })
    expect(view({ deckLen: 2, idx: 2, weakOnly: true, weakLeft: 1 })).toEqual({ kind: 'done', again: true, onlyWeak: false, showAll: true })
  })
})
