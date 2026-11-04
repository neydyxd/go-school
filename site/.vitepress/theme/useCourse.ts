import { computed } from 'vue'
import { useData, useRoute } from 'vitepress'
import { allLessons, lessonPath, modules } from '../course'
import { progress } from './store'

export function useCourse() {
  const route = useRoute()
  const { theme } = useData()
  const ready = computed(() => new Set<string>(theme.value.ready ?? []))

  // Текущий урок по URL: /m01/l05-slices
  const current = computed(() => {
    const path = route.path.replace(/\.html$/, '').replace(/\/$/, '')
    return allLessons.find(({ m, l }) => lessonPath(m, l) === path) ?? null
  })

  const isDone = (id: string) => !!progress.value.lessons[id]?.done

  // Следующий непройденный готовый урок по порядку курса (для «продолжить»)
  const nextUp = computed(
    () => allLessons.find(({ l }) => ready.value.has(l.id) && !isDone(l.id)) ?? null,
  )

  const total = allLessons.length
  const doneCount = computed(() => allLessons.filter(({ l }) => isDone(l.id)).length)

  return { modules, ready, current, isDone, nextUp, total, doneCount, lessonPath }
}
