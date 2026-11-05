<script setup lang="ts">
import { computed } from 'vue'
import { withBase } from 'vitepress'
import { inkOn, type Module } from '../../course'
import { useCourse } from '../useCourse'
import ProgressIO from './ProgressIO.vue'
import SyncPanel from './SyncPanel.vue'

// Схема маршрута: основная линия модулей + параллельная линия алгоритмов.
const { modules, ready, isDone, nextUp, total, doneCount, lessonPath } = useCourse()

const main = modules.filter((m) => m.id !== 'm11')
const algo = modules.find((m) => m.id === 'm11')!

const doneIn = (m: Module) => m.lessons.filter((l) => isDone(l.id)).length
const readyIn = (m: Module) => m.lessons.filter((l) => ready.value.has(l.id)).length
const pct = computed(() => Math.round((doneCount.value / total) * 100))
const num = (m: Module) => String(Number(m.id.slice(1)))

// Конечная станция основной линии: дипломный проект (site/content/diploma.md).
const diploma = [
  { id: 'booking', title: 'Бронирование переговорок' },
  { id: 'flags', title: 'Сервис фича-флагов' },
  { id: 'orders', title: 'Заказы с outbox' },
]
</script>

<template>
  <div class="map">
    <header class="intro">
      <h1 class="title">Go School</h1>
      <p class="lead">
        Маршрут от первой строчки на Go до джуниор-бэкендера: язык и его внутренности,
        конкурентность, PostgreSQL, Gin, фоновые задачи на River и алгоритмы для собеседований.
      </p>
      <div class="status">
        <div class="meter" role="img" :aria-label="`Пройдено ${doneCount} из ${total} уроков`">
          <span class="meter-big">{{ doneCount }}</span>
          <span class="meter-of">из {{ total }} станций · {{ pct }}%</span>
        </div>
        <a v-if="nextUp" class="go" :href="withBase(lessonPath(nextUp.m, nextUp.l))" :style="{ '--line': nextUp.m.color }">
          <span class="go-dot" aria-hidden="true" />
          <span>
            <span class="go-label">{{ doneCount ? 'Продолжить' : 'Начать' }}</span>
            <span class="go-name">{{ nextUp.l.title }}</span>
          </span>
        </a>
      </div>
    </header>

    <div class="scheme">
      <ol class="line" aria-label="Основная линия">
        <li v-for="m in main" :key="m.id" class="station" :class="{ complete: doneIn(m) === m.lessons.length }" :style="{ '--line': m.color, '--pin-ink': inkOn(m.color) }">
          <div class="pin" aria-hidden="true"><span>{{ num(m) }}</span></div>
          <div class="body">
            <h2 class="st-name">{{ m.title }}</h2>
            <p class="st-meta">
              {{ doneIn(m) }}/{{ m.lessons.length }} пройдено
              <template v-if="readyIn(m) < m.lessons.length"> · {{ readyIn(m) ? `готово ${readyIn(m)}` : 'строится' }}</template>
            </p>
            <ul class="stops">
              <li v-for="l in m.lessons" :key="l.id" :class="{ done: isDone(l.id), soon: !ready.has(l.id) }">
                <a v-if="ready.has(l.id)" :href="withBase(lessonPath(m, l))">{{ l.title }}</a>
                <span v-else>{{ l.title }}</span>
              </li>
            </ul>
            <p v-if="m.id === 'm02'" class="transfer" :style="{ '--line': algo.color }">
              Пересадка на линию «{{ algo.short }}»: решай 2-3 задачи в неделю параллельно с основной линией.
            </p>
          </div>
        </li>
        <li class="station terminus">
          <div class="pin" aria-hidden="true" />
          <div class="body">
            <h2 class="st-name"><a :href="withBase('/diploma')">Диплом</a></h2>
            <p class="st-meta">Конечная · свой сервис со всем стеком курса, тема на выбор</p>
            <ul class="stops">
              <li v-for="t in diploma" :key="t.id">
                <a :href="withBase(`/diploma#${t.id}`)">{{ t.title }}</a>
              </li>
            </ul>
          </div>
        </li>
      </ol>

      <aside class="side">
        <section class="branch station" :style="{ '--line': algo.color, '--pin-ink': inkOn(algo.color) }" aria-label="Параллельная линия: алгоритмы">
          <div class="pin" aria-hidden="true"><span>{{ num(algo) }}</span></div>
          <div class="body">
            <h2 class="st-name">{{ algo.title }}</h2>
            <p class="st-meta">Параллельная линия · {{ doneIn(algo) }}/{{ algo.lessons.length }}</p>
            <ul class="stops">
              <li v-for="l in algo.lessons" :key="l.id" :class="{ done: isDone(l.id), soon: !ready.has(l.id) }">
                <a v-if="ready.has(l.id)" :href="withBase(lessonPath(algo, l))">{{ l.title }}</a>
                <span v-else>{{ l.title }}</span>
              </li>
            </ul>
          </div>
        </section>
        <ProgressIO class="io-box" />
        <SyncPanel class="io-box" />
      </aside>
    </div>
  </div>
</template>

<style scoped>
.map { max-width: 1180px; margin: 0 auto; padding: 48px 24px 96px; }
.intro { display: grid; grid-template-columns: minmax(0, 1.4fr) minmax(0, 1fr); gap: 24px 56px; align-items: end; margin-bottom: 64px; }
.title {
  grid-column: 1 / -1;
  margin: 0;
  font-family: var(--gs-font-display);
  font-variation-settings: 'wdth' 62;
  font-weight: 800;
  font-size: clamp(64px, 14vw, 190px);
  line-height: 0.82;
  letter-spacing: -0.02em;
  color: var(--vp-c-text-1);
}
.lead { margin: 0; font-size: clamp(18px, 1.6vw, 21px); line-height: 1.5; max-width: 52ch; color: var(--vp-c-text-1); }
.status { display: flex; flex-direction: column; gap: 16px; }
.meter { display: flex; align-items: baseline; gap: 10px; }
.meter-big { font-family: var(--gs-font-text); font-size: 56px; font-weight: 700; line-height: 1; font-variant-numeric: tabular-nums; }
.meter-of { color: var(--vp-c-text-2); font-size: 15px; }
.go {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 14px 18px;
  border-radius: 4px;
  background: var(--gs-board-bg);
  color: var(--gs-board-ink);
  text-decoration: none;
  border-left: 8px solid var(--line);
}
.go:hover { filter: brightness(1.08); }
.go:focus-visible { outline: 3px solid var(--line); outline-offset: 3px; }
.go-dot { width: 22px; height: 22px; border-radius: 50%; border: 5px solid var(--line); flex: none; }
.go-label { display: block; font-family: var(--gs-font-display); font-size: 12px; text-transform: uppercase; letter-spacing: 0.06em; opacity: 0.75; }
.go-name { display: block; font-weight: 700; font-size: 18px; }

.scheme { display: grid; grid-template-columns: minmax(0, 1fr) 360px; gap: 56px; align-items: start; }
.line { list-style: none; margin: 0; padding: 0; }
.station { position: relative; display: grid; grid-template-columns: 56px minmax(0, 1fr); gap: 20px; padding-bottom: 36px; }
.line .station::before {
  content: '';
  position: absolute;
  left: 23px;
  top: 28px;
  bottom: -28px;
  width: 10px;
  background: var(--line);
}
.line .station:last-child::before { display: none; }
.pin {
  position: relative;
  z-index: 1;
  width: 56px;
  height: 56px;
  border-radius: 50%;
  background: var(--line);
  display: grid;
  place-items: center;
  box-shadow: 0 0 0 6px var(--vp-c-bg);
  font-family: var(--gs-font-text);
  font-weight: 700;
  font-size: 22px;
  color: var(--pin-ink);
}
.st-name {
  margin: 6px 0 2px;
  padding: 0;
  border: 0;
  font-family: var(--gs-font-display);
  font-variation-settings: 'wdth' 75;
  font-weight: 800;
  font-size: clamp(26px, 3vw, 36px);
  line-height: 1.05;
  letter-spacing: -0.01em;
}
.st-meta { margin: 0 0 12px; font-size: 14px; color: var(--vp-c-text-2); font-variant-numeric: tabular-nums; }
.stops { list-style: none; margin: 0; padding: 0; display: flex; flex-wrap: wrap; gap: 6px 18px; }
.stops li { position: relative; padding-left: 16px; font-size: 15px; line-height: 1.4; }
.stops li::before {
  content: '';
  position: absolute;
  left: 0;
  top: 0.38em;
  width: 9px;
  height: 9px;
  border-radius: 50%;
  border: 2px solid var(--line);
}
.stops li.done::before { background: var(--line); }
.stops li.soon { color: var(--vp-c-text-3); }
.stops li.soon::before { border-style: dashed; }
.stops a { color: var(--vp-c-text-1); text-decoration: none; border-bottom: 1px solid var(--vp-c-divider); }
.stops a:hover { border-bottom-color: var(--line); }
.terminus { --line: var(--vp-c-text-1); padding-bottom: 0; }
.terminus .pin { background: var(--vp-c-bg); border: 10px solid var(--vp-c-text-1); }
.terminus .pin::after { content: ''; width: 14px; height: 14px; border-radius: 50%; background: var(--vp-c-text-1); }
.terminus .st-name a { color: inherit; text-decoration: none; border-bottom: 3px solid var(--vp-c-divider); }
.terminus .st-name a:hover { border-bottom-color: var(--vp-c-text-1); }
.terminus .st-name a:focus-visible { outline: 3px solid var(--vp-c-text-1); outline-offset: 3px; }
.transfer {
  margin: 16px 0 0;
  padding: 8px 12px;
  font-size: 14px;
  border-left: 6px solid var(--line);
  background: var(--vp-c-bg-soft);
  border-radius: 0 4px 4px 0;
}

.side { position: sticky; top: calc(var(--vp-nav-height) + 24px); display: flex; flex-direction: column; gap: 40px; }
.branch { padding: 20px; border-radius: 4px; background: var(--vp-c-bg-soft); border-top: 10px solid var(--line); }
.branch .stops { flex-direction: column; }
.io-box { padding: 0 4px; }

@media (max-width: 960px) {
  .intro { grid-template-columns: 1fr; }
  .scheme { grid-template-columns: 1fr; gap: 24px; }
  .side { position: static; }
}
@media (max-width: 520px) {
  .terminus .pin { border-width: 7px; }
  .terminus .pin::after { width: 10px; height: 10px; }
  .map { padding: 28px 16px 64px; }
  .station { grid-template-columns: 40px minmax(0, 1fr); gap: 14px; }
  .pin { width: 40px; height: 40px; font-size: 17px; }
  .line .station::before { left: 15px; top: 20px; width: 10px; }
  .meter-big { font-size: 44px; }
}
</style>
