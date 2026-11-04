import DefaultTheme from 'vitepress/theme'
import type { Theme } from 'vitepress'
import { defineComponent, h, onMounted } from 'vue'
import { useCourse } from './useCourse'
import { initProgress } from './store'
import LineBoard from './components/LineBoard.vue'
import LessonDone from './components/LessonDone.vue'
import Predict from './components/Predict.vue'
import Hw from './components/Hw.vue'
import Roadmap from './components/Roadmap.vue'
import './style.css'

export default {
  extends: DefaultTheme,
  Layout: defineComponent({
    name: 'GsLayout',
    setup() {
      // Цвет линии текущего модуля прокидываем в CSS всей страницы урока.
      const { current } = useCourse()
      onMounted(initProgress)
      return () =>
        h(
          'div',
          { class: 'gs-root', style: current.value ? { '--gs-line': current.value.m.color } : undefined },
          h(DefaultTheme.Layout, null, {
            'doc-before': () => h(LineBoard),
            'doc-after': () => h(LessonDone),
          }),
        )
    },
  }),
  enhanceApp({ app }) {
    app.component('Predict', Predict)
    app.component('Hw', Hw)
    app.component('Roadmap', Roadmap)
  },
} satisfies Theme
