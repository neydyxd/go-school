// Банк карточек собеса: все блоки ::: interview всех уроков курса.
// Новые уроки подхватываются сами: загрузчик смотрит на content/m*/*.md при сборке и в dev.
import { createContentLoader, createMarkdownRenderer, type SiteConfig } from 'vitepress'
import { modules } from '../course'
import { absLinks, cardId, parseInterview, type Card } from './cards'

declare const data: Card[]
export { data }

export default createContentLoader('m*/*.md', {
  includeSrc: true,
  async transform(pages): Promise<Card[]> {
    // Тот же markdown-it, что и у сайта (Shiki, контейнеры): VitePress держит его одним экземпляром.
    const config = (globalThis as { VITEPRESS_CONFIG?: SiteConfig }).VITEPRESS_CONFIG!
    const md = await createMarkdownRenderer(config.srcDir, config.markdown, config.site.base, config.logger)
    const byUrl = new Map(pages.map((p) => [p.url, p.src ?? '']))
    const cards: Card[] = []
    for (const m of modules) {
      for (const l of m.lessons) {
        const url = `/${m.id}/${l.slug}`
        const src = byUrl.get(url)
        if (src === undefined) continue
        const seen = new Set<string>()
        for (const { q, a } of parseInterview(src)) {
          const id = cardId(m.id, l.slug, q)
          if (seen.has(id)) continue // один и тот же вопрос дважды в уроке
          seen.add(id)
          // env как у страницы урока: cleanUrls влияет на то, как linkPlugin пишет внутренние ссылки
          const env = { cleanUrls: config.cleanUrls, relativePath: `${m.id}/${l.slug}.md` }
          const html = (src: string, inline = false) =>
            absLinks(inline ? md.renderInline(src, env) : md.render(src, env), url, config.site.base)
          cards.push({ id, m: m.id, lesson: l.id, title: l.title, url, q: html(q, true), a: html(a) })
        }
      }
    }
    return cards
  },
})
