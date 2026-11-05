import { existsSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import container from 'markdown-it-container'
import { defineConfig, type DefaultTheme } from 'vitepress'
import { lessonPath, modules } from './course'

const contentDir = fileURLToPath(new URL('../content', import.meta.url))
const isReady = (m: string, slug: string) => existsSync(`${contentDir}/${m}/${slug}.md`)

// Готовые уроки: по ним строятся ссылки; остальные показываются как «скоро».
const ready = modules.flatMap((m) => m.lessons.filter((l) => isReady(m.id, l.slug)).map((l) => l.id))

const sidebar: DefaultTheme.SidebarItem[] = modules.map((m) => ({
  text: m.title,
  collapsed: true,
  items: m.lessons.map((l) =>
    ready.includes(l.id) ? { text: l.title, link: lessonPath(m, l) } : { text: `${l.title} · скоро` },
  ),
}))

// Блоки урока: ::: hood / ::: trap / ::: interview [заголовок]
const blocks: Record<string, string> = {
  hood: 'Под капотом',
  trap: 'Грабли',
  interview: 'На собесе спросят',
}

export default defineConfig<DefaultTheme.Config & { ready: string[] }>({
  lang: 'ru-RU',
  title: 'Go School',
  description: 'Личная школа Go: от нуля до джуниор-бэкендера',
  srcDir: 'content',
  cleanUrls: true,
  lastUpdated: false,
  head: [
    ['link', { rel: 'preconnect', href: 'https://fonts.googleapis.com' }],
    ['link', { rel: 'preconnect', href: 'https://fonts.gstatic.com', crossorigin: '' }],
    ['link', {
      rel: 'stylesheet',
      href: 'https://fonts.googleapis.com/css2?family=Science+Gothic:wdth,wght@50..200,100..900&family=Commissioner:wght@400..700&family=Ubuntu+Sans+Mono:wght@400..600&display=swap',
    }],
  ],
  markdown: {
    theme: { light: 'github-light', dark: 'github-dark' },
    config(md) {
      for (const [name, title] of Object.entries(blocks)) {
        md.use(container, name, {
          render(tokens: { nesting: number; info: string }[], idx: number) {
            const t = tokens[idx]
            if (t.nesting !== 1) return '</div>\n'
            const custom = t.info.trim().slice(name.length).trim()
            return `<div class="gs-block gs-${name}"><p class="gs-block-title">${md.utils.escapeHtml(custom || title)}</p>\n`
          },
        })
      }
    },
  },
  themeConfig: {
    ready, // id готовых уроков, читают компоненты темы
    nav: [{ text: 'Маршрут', link: '/' }, { text: 'Диплом', link: '/diploma' }],
    sidebar,
    outline: { level: [2, 3], label: 'На этой странице' },
    docFooter: { prev: 'Предыдущая станция', next: 'Следующая станция' },
    darkModeSwitchLabel: 'Тема',
    lightModeSwitchTitle: 'Светлая тема',
    darkModeSwitchTitle: 'Тёмная тема',
    sidebarMenuLabel: 'Уроки',
    returnToTopLabel: 'Наверх',
    notFound: { title: 'Такой станции нет', quote: 'Похоже, этот урок ещё не построен.', linkText: 'К схеме маршрута' },
    search: {
      provider: 'local',
      options: {
        translations: {
          button: { buttonText: 'Поиск', buttonAriaLabel: 'Поиск' },
          modal: {
            noResultsText: 'Ничего не нашлось',
            resetButtonTitle: 'Сбросить',
            footer: { selectText: 'открыть', navigateText: 'выбор', closeText: 'закрыть' },
          },
        },
      },
    },
  },
})
