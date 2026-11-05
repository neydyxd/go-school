<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { withBase } from 'vitepress'
import { modules } from '../../course'
import { data as bank } from '../cards.data'
import { counts, empty, isWeak, KEY, load, mark, pickDeck, save, shuffle, view, type Card } from '../cards'

// Карточки собеса: вопросы из всех блоков «На собесе спросят». Состояние только на клиенте.
const storage = (): Storage | null => {
  try {
    return window.localStorage
  } catch {
    return null
  }
}

const state = ref(empty())
const mounted = ref(false)
const selected = ref<string[]>([]) // пусто = все модули
const weakOnly = ref(false)
const deck = ref<Card[]>([])
const idx = ref(0)
const revealed = ref(false)
const cardEl = ref<HTMLElement | null>(null)
const answerEl = ref<HTMLElement | null>(null)
const panelEl = ref<HTMLElement | null>(null) // итог круга или пустое состояние

const byModule = modules
  .map((m) => ({ m, n: bank.filter((c) => c.m === m.id).length }))
  .filter((x) => x.n > 0)
const lessonName = new Map(modules.flatMap((m) => m.lessons.map((l) => [l.id, `${m.short} · ${l.title}`])))
const colorOf = new Map(modules.map((m) => [m.id, m.color]))

const scope = computed(() => pickDeck(bank, state.value, { modules: selected.value, weakOnly: false }))
const stats = computed(() => counts(scope.value, state.value))
const current = computed(() => deck.value[idx.value] ?? null)
const round = computed(() => counts(deck.value, state.value))
const v = computed(() => view({ deckLen: deck.value.length, idx: idx.value, weakOnly: weakOnly.value, weakLeft: stats.value.weak }))

function rebuild() {
  deck.value = shuffle(pickDeck(bank, state.value, { modules: selected.value, weakOnly: weakOnly.value }))
  idx.value = 0
  revealed.value = false
}

const onStorage = (e: StorageEvent) => {
  if (e.key === KEY) state.value = load(storage())
}

onMounted(() => {
  state.value = load(storage())
  mounted.value = true
  rebuild()
  window.addEventListener('storage', onStorage)
})
onUnmounted(() => window.removeEventListener('storage', onStorage))
watch([selected, weakOnly], rebuild)

// Фокус следует за содержимым: карточка, иначе итог круга или пустое состояние,
// чтобы клавиатура и скринридер не теряли место.
const focusCard = () => nextTick(() => (cardEl.value ?? panelEl.value)?.focus({ preventScroll: true }))

function showAll() {
  weakOnly.value = false // watch пересоберёт колоду
  focusCard()
}

function toggleModule(id: string) {
  const s = selected.value
  selected.value = s.includes(id) ? s.filter((x) => x !== id) : [...s, id]
}

function reveal() {
  if (!current.value || revealed.value) return
  revealed.value = true
  // фокус на ответе: скринридер прочитает его, стрелки продолжают работать (обработчик на секции)
  nextTick(() => answerEl.value?.focus({ preventScroll: true }))
}

function answer(known: boolean) {
  const c = current.value
  if (!c || !revealed.value) return
  state.value = mark(state.value, c.id, known, Date.now())
  save(storage(), state.value)
  idx.value++
  revealed.value = false
  focusCard()
}

// Пробел: показать ответ; стрелка вправо: знаю; влево: повторить.
// На кнопках, ссылках и чекбоксах пробел и Enter оставляем браузеру.
function onKey(e: KeyboardEvent) {
  if (e.altKey || e.ctrlKey || e.metaKey) return
  const tag = (e.target as HTMLElement).tagName
  if (tag === 'INPUT') return
  if (e.key === ' ' && tag !== 'BUTTON' && tag !== 'A') {
    if (!current.value) return
    e.preventDefault()
    reveal()
  } else if (e.key === 'ArrowRight' && revealed.value) {
    e.preventDefault()
    answer(true)
  } else if (e.key === 'ArrowLeft' && revealed.value) {
    e.preventDefault()
    answer(false)
  }
}

const markOf = (c: Card) => state.value.cards[c.id]
</script>

<template>
  <section class="fc" aria-label="Карточки собеса" @keydown="onKey">
    <div class="fc-filters" role="group" aria-label="Модули">
      <button type="button" class="chip" :aria-pressed="selected.length === 0" @click="selected = []">
        Все <span class="n">{{ bank.length }}</span>
      </button>
      <button
        v-for="{ m, n } in byModule"
        :key="m.id"
        type="button"
        class="chip"
        :style="{ '--c': m.color }"
        :aria-pressed="selected.includes(m.id)"
        :title="m.title"
        @click="toggleModule(m.id)"
      >
        <span class="dot" aria-hidden="true" />{{ m.short }} <span class="n">{{ n }}</span>
      </button>
    </div>

    <div v-if="!mounted" class="fc-panel fc-note">В банке {{ bank.length }} вопросов. Загружаю карточки…</div>
    <template v-else>
      <div class="fc-meta">
        <p class="fc-stats">
          <span><b>{{ stats.total }}</b> в выборке</span>
          <span class="ok"><b>{{ stats.known }}</b> знаю</span>
          <span class="bad"><b>{{ stats.weak }}</b> повторить</span>
          <span><b>{{ stats.fresh }}</b> новых</span>
        </p>
        <div class="fc-modes">
          <label class="weak">
            <input v-model="weakOnly" type="checkbox" />
            Только «повторить»
          </label>
          <button type="button" class="ghost" @click="rebuild(); focusCard()">Перемешать</button>
        </div>
      </div>

      <div v-if="v.kind === 'no-weak'" ref="panelEl" class="fc-panel fc-note fc-done" tabindex="-1">
        <p class="done-title">Слабых карточек не осталось</p>
        <p>В выбранных модулях нет карточек с отметкой «повторить». Можно пройти все или выбрать другие модули.</p>
        <div class="fc-actions">
          <button type="button" class="act" @click="showAll">Показать все карточки</button>
        </div>
      </div>

      <div v-else-if="v.kind === 'no-cards'" ref="panelEl" class="fc-panel fc-note fc-done" tabindex="-1">
        В выбранных модулях пока нет вопросов.
      </div>

      <div v-else-if="v.kind === 'done'" ref="panelEl" class="fc-panel fc-note fc-done" tabindex="-1">
        <p class="done-title">{{ v.again ? 'Круг пройден' : 'Все слабые карточки выучены' }}</p>
        <p>За этот круг: знаю {{ round.known }}, повторить {{ round.weak }}.</p>
        <div class="fc-actions">
          <button v-if="v.again" type="button" class="act" @click="rebuild(); focusCard()">Ещё круг</button>
          <button v-if="v.onlyWeak" type="button" class="act" @click="weakOnly = true; focusCard()">Только «повторить»</button>
          <button v-if="v.showAll" type="button" class="act" @click="showAll">Показать все карточки</button>
        </div>
      </div>

      <article
        v-else-if="current"
        ref="cardEl"
        class="fc-card"
        :style="{ '--c': colorOf.get(current.m) }"
        tabindex="0"
        :aria-label="`Карточка ${idx + 1} из ${deck.length}`"
      >
        <header class="fc-head">
          <span class="pos">{{ idx + 1 }} / {{ deck.length }}</span>
          <a class="to-lesson" :href="withBase(current.url)">{{ lessonName.get(current.lesson) }} · к уроку</a>
        </header>
        <div class="fc-q" v-html="current.q" />
        <p v-if="markOf(current)" class="fc-was">
          {{ isWeak(state, current.id) ? 'В прошлый раз: повторить' : 'В прошлый раз: знаю' }} · показов {{ markOf(current)!.seen }}
        </p>

        <button v-if="!revealed" type="button" class="act reveal" @click="reveal">
          Показать ответ <kbd>Пробел</kbd>
        </button>
        <template v-else>
          <div ref="answerEl" class="fc-a" tabindex="-1" aria-label="Ответ" v-html="current.a" />
          <div class="fc-actions">
            <button type="button" class="act again" @click="answer(false)"><kbd>←</kbd> Повторить</button>
            <button type="button" class="act know" @click="answer(true)">Знаю <kbd>→</kbd></button>
          </div>
        </template>
      </article>
      <p class="fc-help">Клавиши, когда фокус на карточке: пробел открывает ответ, стрелка вправо «знаю», влево «повторить».</p>
    </template>
  </section>
</template>

<style scoped>
.fc { margin: 24px 0 32px; }
.fc-filters { display: flex; flex-wrap: wrap; gap: 6px; }
.chip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  min-height: 36px;
  padding: 0 10px;
  border-radius: 999px;
  border: 1px solid var(--vp-c-divider);
  background: var(--vp-c-bg);
  color: var(--vp-c-text-2);
  font-size: 14px;
  font-weight: 600;
}
.chip:hover { border-color: var(--vp-c-text-2); }
.chip[aria-pressed='true'] { color: var(--vp-c-text-1); border-color: var(--vp-c-text-1); background: var(--vp-c-bg-soft); }
.chip .dot { width: 12px; height: 12px; border-radius: 50%; border: 3px solid var(--c); background: var(--vp-c-bg); }
.chip[aria-pressed='true'] .dot { background: var(--c); }
.chip .n { font-weight: 400; font-size: 12px; color: var(--vp-c-text-2); font-variant-numeric: tabular-nums; }

.fc-meta { display: flex; flex-wrap: wrap; align-items: center; justify-content: space-between; gap: 8px 16px; margin: 16px 0 10px; }
.fc .fc-stats { display: flex; flex-wrap: wrap; gap: 4px 14px; margin: 0; font-size: 14px; color: var(--vp-c-text-2); }
.fc-stats b { color: var(--vp-c-text-1); font-variant-numeric: tabular-nums; }
.fc-stats .ok b { color: var(--gs-ok); }
.fc-stats .bad b { color: var(--gs-bad); }
.fc-modes { display: flex; align-items: center; gap: 14px; font-size: 14px; }
.weak { display: inline-flex; align-items: center; gap: 8px; cursor: pointer; }
.weak input { width: 18px; height: 18px; accent-color: var(--vp-c-text-1); }
.ghost { color: var(--vp-c-text-2); text-decoration: underline; text-underline-offset: 3px; }
.ghost:hover { color: var(--vp-c-text-1); }

.fc-panel { padding: 18px 20px; border-radius: 4px; background: var(--vp-c-bg-soft); }
.fc .fc-note p { margin: 4px 0; }
.done-title { font-family: var(--gs-font-display); font-weight: 800; font-size: 20px; }

/* Карточка как табло: тёмная плашка вопроса и полоса линии модуля слева */
.fc-card {
  border-radius: 4px;
  border: 1px solid var(--vp-c-divider);
  border-left: 6px solid var(--c);
  background: var(--vp-c-bg-elv);
  padding: 0 18px 16px;
  outline: none;
}
.fc-card:focus-visible, .fc-card:has(.fc-a:focus-visible) { box-shadow: 0 0 0 2px var(--vp-c-bg), 0 0 0 4px var(--vp-c-text-1); }
.fc-a:focus, .fc-done:focus { outline: none; }
.fc-done:focus-visible { box-shadow: 0 0 0 2px var(--vp-c-text-1); }
.fc-head {
  display: flex;
  flex-wrap: wrap;
  justify-content: space-between;
  gap: 4px 12px;
  margin: 0 -18px;
  padding: 8px 18px;
  background: var(--gs-board-bg);
  color: var(--gs-board-ink);
  font-size: 13px;
}
.pos { font-family: var(--gs-font-display); font-weight: 800; letter-spacing: 0.04em; font-variant-numeric: tabular-nums; }
.fc .to-lesson { color: inherit; opacity: 0.85; text-decoration-color: var(--c); }
.fc .to-lesson:hover { opacity: 1; }
.fc .to-lesson:focus-visible { outline: 2px solid var(--gs-board-ink); outline-offset: 2px; }
.fc-q {
  margin: 16px 0 6px;
  font-family: var(--gs-font-display);
  font-variation-settings: 'wdth' 85;
  font-size: 22px;
  font-weight: 700;
  line-height: 1.25;
  overflow-wrap: anywhere;
}
.fc .fc-was { margin: 0 0 8px; font-size: 13px; color: var(--vp-c-text-2); }
.fc-a { margin-top: 10px; padding-top: 4px; border-top: 1px dashed var(--vp-c-divider); overflow-wrap: anywhere; }
.fc-a :deep(div[class*='language-']) { margin: 12px 0 !important; border-radius: 4px !important; }
.fc-a :deep(pre) { overflow-x: auto; }
.fc-a :deep(table) { display: block; overflow-x: auto; }

.fc-actions { display: flex; flex-wrap: wrap; gap: 10px; margin-top: 14px; }
.act {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  min-height: 44px;
  padding: 0 16px;
  border-radius: 4px;
  border: 1px solid var(--vp-c-divider);
  background: var(--vp-c-bg);
  color: var(--vp-c-text-1);
  font-family: var(--gs-font-display);
  font-weight: 700;
  font-size: 15px;
}
.act:hover { border-color: var(--vp-c-text-1); }
.reveal { margin-top: 10px; }
.again { border-color: var(--gs-bad); box-shadow: inset 4px 0 0 var(--gs-bad); }
.know { border-color: var(--gs-ok); box-shadow: inset -4px 0 0 var(--gs-ok); }
.fc-actions .act { flex: 1 1 140px; }
.chip:focus-visible, .act:focus-visible, .ghost:focus-visible, .weak input:focus-visible {
  outline: 2px solid var(--vp-c-text-1);
  outline-offset: 2px;
}
kbd {
  padding: 1px 6px;
  border-radius: 4px;
  border: 1px solid var(--vp-c-divider);
  font-family: var(--gs-font-mono);
  font-size: 12px;
  font-weight: 500;
  color: var(--vp-c-text-2);
}
.fc .fc-help { margin: 8px 0 0; font-size: 13px; color: var(--vp-c-text-2); }
</style>
