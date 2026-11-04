// Проверяет все внешние ссылки в уроках. Exit 1, если есть битые.
// Запуск: node scripts/check-links.mjs [папка=content]
import { readdirSync, readFileSync, statSync } from 'node:fs'
import { join } from 'node:path'

const root = process.argv[2] ?? new URL('../content', import.meta.url).pathname
// Сайты, которые режут ботов (403/429), но живые в браузере: только предупреждение.
const botWalls = ['leetcode.com', 'habr.com', 'stepik.org', 'youtube.com', 'medium.com']

const files = (dir) =>
  readdirSync(dir).flatMap((f) => {
    const p = join(dir, f)
    return statSync(p).isDirectory() ? files(p) : p.endsWith('.md') ? [p] : []
  })

const links = new Map() // url -> [файлы]
for (const f of files(root)) {
  for (const [, url] of readFileSync(f, 'utf8').matchAll(/\]\((https?:\/\/[^)\s]+)\)/g)) {
    links.set(url, [...(links.get(url) ?? []), f.slice(root.length + 1)])
  }
}

async function check(url) {
  for (const method of ['HEAD', 'GET']) {
    try {
      const r = await fetch(url, {
        method,
        redirect: 'follow',
        signal: AbortSignal.timeout(15000),
        headers: { 'user-agent': 'Mozilla/5.0 (go-school link check)' },
      })
      if (r.ok) return r.status
      if (method === 'GET') return r.status
    } catch (e) {
      if (method === 'GET') return e.name === 'TimeoutError' ? 'timeout' : 'error'
    }
  }
}

const queue = [...links.keys()]
const bad = []
const warn = []
await Promise.all(
  Array.from({ length: 8 }, async () => {
    while (queue.length) {
      const url = queue.shift()
      const status = await check(url)
      if (typeof status === 'number' && status < 400) continue
      const wall = botWalls.some((h) => new URL(url).hostname.endsWith(h)) && (status === 403 || status === 429)
      ;(wall ? warn : bad).push(`${status}  ${url}  (${links.get(url).join(', ')})`)
    }
  }),
)

console.log(`Проверено ссылок: ${links.size}`)
if (warn.length) console.log(`\nПредупреждения (сайт режет ботов, проверь руками):\n${warn.join('\n')}`)
if (bad.length) {
  console.log(`\nБитые:\n${bad.join('\n')}`)
  process.exit(1)
}
console.log('Битых нет')
