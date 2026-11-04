<script setup lang="ts">
import { computed } from 'vue'
import { withBase } from 'vitepress'
import { useCourse } from '../useCourse'

// Табло вагона: линия модуля, станции-уроки, текущая подсвечена.
const { current, ready, isDone, lessonPath } = useCourse()

const pos = computed(() => {
  const c = current.value
  if (!c) return null
  const i = c.m.lessons.findIndex((l) => l.id === c.l.id)
  return { i, n: c.m.lessons.length, next: c.m.lessons[i + 1] ?? null }
})
</script>

<template>
  <nav v-if="current && pos" class="board" :style="{ '--line': current.m.color }" aria-label="Положение на линии модуля">
    <div class="board-head">
      <span class="board-line">Линия «{{ current.m.title }}»</span>
      <span class="board-pos">станция {{ pos.i + 1 }} из {{ pos.n }}</span>
    </div>
    <ol class="track">
      <li
        v-for="(l, i) in current.m.lessons"
        :key="l.id"
        class="stop"
        :class="{ done: isDone(l.id), here: i === pos.i, soon: !ready.has(l.id) }"
      >
        <a v-if="ready.has(l.id) && i !== pos.i" :href="withBase(lessonPath(current.m, l))" :title="l.title">
          <span class="dot" /><span class="name">{{ l.title }}</span>
        </a>
        <span v-else :title="l.title" :aria-current="i === pos.i ? 'step' : undefined">
          <span class="dot" /><span class="name">{{ l.title }}</span>
        </span>
      </li>
    </ol>
    <p v-if="pos.next" class="board-next">Следующая станция: <b>{{ pos.next.title }}</b></p>
    <p v-else class="board-next">Конечная линии. Дальше пересадка на следующий модуль.</p>
  </nav>
</template>

<style scoped>
.board {
  margin: 0 0 40px;
  padding: 16px 18px 14px;
  border-radius: 4px;
  background: var(--gs-board-bg);
  color: var(--gs-board-ink);
}
.board-head {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  font-family: var(--gs-font-display);
  font-size: 13px;
  letter-spacing: 0.04em;
  text-transform: uppercase;
}
.board-line { font-weight: 700; }
.board-pos { opacity: 0.7; font-variant-numeric: tabular-nums; white-space: nowrap; }
.track {
  --gap: 1fr;
  display: grid;
  grid-auto-flow: column;
  grid-auto-columns: minmax(0, 1fr);
  list-style: none;
  margin: 18px 0 10px;
  padding: 0;
  position: relative;
}
.track::before {
  content: '';
  position: absolute;
  left: 8px;
  right: 8px;
  top: 7px;
  height: 4px;
  border-radius: 2px;
  background: var(--line);
}
.stop { position: relative; margin: 0; min-width: 0; }
.stop a, .stop > span { display: flex; flex-direction: column; gap: 8px; color: inherit; text-decoration: none; }
.dot {
  width: 18px;
  height: 18px;
  border-radius: 50%;
  background: var(--gs-board-bg);
  border: 4px solid var(--line);
  position: relative;
  z-index: 1;
  transition: background 0.3s;
}
.done .dot { background: var(--line); }
.here .dot { width: 26px; height: 26px; margin-top: -4px; margin-left: -4px; border-width: 5px; box-shadow: 0 0 0 4px var(--gs-board-bg); }
.soon .dot { border-style: dashed; opacity: 0.6; }
.name {
  font-size: 12px;
  line-height: 1.25;
  opacity: 0.75;
  overflow: hidden;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  padding-right: 6px;
}
.here .name { opacity: 1; font-weight: 600; }
.stop a:hover .name { opacity: 1; text-decoration: underline; }
.stop a:focus-visible { outline: 2px solid var(--line); outline-offset: 4px; border-radius: 4px; }
.board-next { margin: 4px 0 0; font-size: 14px; }
@media (max-width: 640px) {
  .name { display: none; }
  .here .name { display: block; position: absolute; top: 28px; left: 0; width: max-content; max-width: 60vw; }
  .track { margin-bottom: 40px; }
}
</style>
