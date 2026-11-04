import { ref } from 'vue'
import { empty, KEY, load, save, type Progress } from './progress'

// Один реактивный прогресс на всё приложение. На сервере (SSR) пустой,
// на клиенте подгружается в init() и синхронизируется между вкладками.
export const progress = ref<Progress>(empty())

const storage = (): Storage | null => {
  try {
    return typeof window === 'undefined' ? null : window.localStorage
  } catch {
    return null
  }
}

let started = false
export function initProgress() {
  if (started || typeof window === 'undefined') return
  started = true
  progress.value = load(storage())
  window.addEventListener('storage', (e) => {
    if (e.key === KEY) progress.value = load(storage())
  })
}

// Все изменения идут через функцию от предыдущего состояния, а не из замыкания рендера.
export function update(fn: (p: Progress) => Progress) {
  progress.value = fn(progress.value)
  save(storage(), progress.value)
}
