<script setup lang="ts">
import { ref } from 'vue'

// «Что выведет программа?»: сначала свой ответ, потом объяснение.
const props = defineProps<{ options: string[]; correct: number }>()
const picked = ref<number | null>(null)
</script>

<template>
  <div class="predict">
    <p class="q">Что выведет программа? Выбери, потом смотри разбор.</p>
    <div class="opts" role="group" aria-label="Варианты ответа">
      <button
        v-for="(o, i) in props.options"
        :key="i"
        type="button"
        class="opt"
        :class="{
          right: picked !== null && i === props.correct,
          wrong: picked === i && i !== props.correct,
        }"
        :disabled="picked !== null"
        @click="picked = i"
      ><code>{{ o }}</code></button>
    </div>
    <div v-if="picked !== null" class="answer" aria-live="polite">
      <p class="verdict">{{ picked === props.correct ? 'Верно.' : 'Мимо. Правильный ответ подсвечен.' }}</p>
      <slot />
      <button type="button" class="again" @click="picked = null">Попробовать ещё раз</button>
    </div>
  </div>
</template>

<style scoped>
.predict { margin: 16px 0 28px; padding: 14px 16px; border-radius: 4px; border: 1px dashed var(--vp-c-divider); }
.q { margin: 0 0 10px; font-weight: 600; }
.opts { display: flex; flex-wrap: wrap; gap: 8px; }
.opt {
  min-height: 40px;
  padding: 6px 12px;
  border-radius: 4px;
  border: 1px solid var(--vp-c-divider);
  background: var(--vp-c-bg-soft);
  color: var(--vp-c-text-1);
  text-align: left;
}
.opt code { background: none; padding: 0; font-size: 14px; white-space: pre-wrap; }
.opt:not(:disabled):hover { border-color: var(--vp-c-text-2); }
.opt:focus-visible { outline: 2px solid var(--vp-c-brand-1); outline-offset: 2px; }
.opt:disabled { cursor: default; }
.opt.right { border-color: var(--gs-ok); box-shadow: inset 4px 0 0 var(--gs-ok); }
.opt.wrong { border-color: var(--gs-bad); box-shadow: inset 4px 0 0 var(--gs-bad); opacity: 0.85; }
.answer { margin-top: 12px; }
.verdict { margin: 0 0 6px; font-weight: 700; }
.again { margin-top: 6px; font-size: 13px; color: var(--vp-c-text-2); text-decoration: underline; background: none; border: 0; padding: 0; }
</style>
