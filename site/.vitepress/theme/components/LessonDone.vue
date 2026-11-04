<script setup lang="ts">
import { computed } from 'vue'
import { withBase } from 'vitepress'
import { allLessons } from '../../course'
import { toggleLesson } from '../progress'
import { update } from '../store'
import { useCourse } from '../useCourse'

const { current, isDone, ready, lessonPath } = useCourse()

const next = computed(() => {
  const c = current.value
  if (!c) return null
  const i = allLessons.findIndex(({ l }) => l.id === c.l.id)
  return allLessons.slice(i + 1).find(({ l }) => ready.value.has(l.id)) ?? null
})

const toggle = () => {
  const id = current.value!.l.id
  update((p) => toggleLesson(p, id, new Date().toISOString()))
}
</script>

<template>
  <section v-if="current" class="done-bar" :style="{ '--line': current.m.color }">
    <button type="button" class="done-btn" :aria-pressed="isDone(current.l.id)" @click="toggle">
      <span class="ring" :class="{ on: isDone(current.l.id) }" aria-hidden="true" />
      {{ isDone(current.l.id) ? 'Урок пройден' : 'Отметить урок пройденным' }}
    </button>
    <a v-if="next" class="next" :href="withBase(lessonPath(next.m, next.l))">
      Следующая станция: {{ next.l.title }} →
    </a>
  </section>
</template>

<style scoped>
.done-bar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-top: 56px;
  padding-top: 24px;
  border-top: 4px solid var(--line);
}
.done-btn {
  display: inline-flex;
  align-items: center;
  gap: 12px;
  min-height: 44px;
  padding: 0 18px 0 12px;
  border-radius: 4px;
  border: 1px solid var(--vp-c-divider);
  font-family: var(--gs-font-display);
  font-weight: 700;
  font-size: 15px;
  color: var(--vp-c-text-1);
  background: var(--vp-c-bg);
}
.done-btn:hover { border-color: var(--line); }
.done-btn:focus-visible { outline: 2px solid var(--line); outline-offset: 2px; }
.ring {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  border: 4px solid var(--line);
  transition: background 0.35s cubic-bezier(0.2, 0.8, 0.2, 1), transform 0.35s cubic-bezier(0.2, 0.8, 0.2, 1);
}
.ring.on { background: var(--line); transform: scale(1.15); }
.next { font-weight: 600; color: var(--vp-c-text-1); }
.next:hover { text-decoration: underline; }
@media (prefers-reduced-motion: reduce) { .ring { transition: none; } }
</style>
