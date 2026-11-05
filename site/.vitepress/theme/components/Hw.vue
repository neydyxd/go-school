<script setup lang="ts">
import { computed } from 'vue'
import { toggleHw } from '../progress'
import { progress, update } from '../store'
import { useCourse } from '../useCourse'

// Задание домашки с отметкой «сделано». <Hw id="ex1" level="база" title="...">описание</Hw>
const props = defineProps<{ id: string; title: string; level?: 'база' | 'посложнее' | 'со звёздочкой' }>()
const { current } = useCourse()
const key = computed(() => `${current.value?.l.id ?? 'x'}/${props.id}`)
const done = computed(() => !!progress.value.hw[key.value])
</script>

<template>
  <div class="hw" :class="{ done }">
    <label class="hw-head">
      <input type="checkbox" :checked="done" @change="update((p) => toggleHw(p, key))" />
      <span class="hw-title">{{ title }}</span>
      <span v-if="level" class="hw-level">{{ level }}</span>
    </label>
    <div class="hw-body"><slot /></div>
  </div>
</template>

<style scoped>
.hw { margin: 14px 0; padding: 12px 16px 4px; border-radius: 4px; background: var(--vp-c-bg-soft); }
.hw-head { display: flex; align-items: center; gap: 10px; cursor: pointer; }
/* на узком экране метка уровня уходит на вторую строку, десктоп как был */
@media (max-width: 640px) { .hw-head { flex-wrap: wrap; row-gap: 4px; } }
.hw-head input { width: 18px; height: 18px; accent-color: var(--gs-line, var(--vp-c-brand-1)); flex: none; }
.hw-title { font-family: var(--gs-font-display); font-weight: 700; font-size: 16px; }
.hw-level { margin-left: auto; font-size: 12px; color: var(--vp-c-text-2); white-space: nowrap; }
.done .hw-title { text-decoration: line-through; text-decoration-thickness: 2px; opacity: 0.7; }
.hw-body :deep(p) { margin: 8px 0; }
</style>
