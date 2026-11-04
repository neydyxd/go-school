<script setup lang="ts">
import { ref } from 'vue'
import { exportJson, importJson } from '../progress'
import { progress, update } from '../store'

const msg = ref<{ ok: boolean; text: string } | null>(null)

function download() {
  const blob = new Blob([exportJson(progress.value)], { type: 'application/json' })
  const a = document.createElement('a')
  a.href = URL.createObjectURL(blob)
  a.download = `go-school-progress-${new Date().toISOString().slice(0, 10)}.json`
  a.click()
  URL.revokeObjectURL(a.href)
  msg.value = { ok: true, text: 'Файл с прогрессом сохранён.' }
}

async function upload(e: Event) {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file) return
  if (file.size > 1_000_000) {
    msg.value = { ok: false, text: 'Файл слишком большой для прогресса курса.' }
    return
  }
  const r = importJson(await file.text())
  if (!r.ok) {
    msg.value = { ok: false, text: `Не загрузил: ${r.error}. Текущий прогресс не тронут.` }
    return
  }
  update(() => r.p)
  msg.value = { ok: true, text: 'Прогресс восстановлен из файла.' }
}
</script>

<template>
  <div class="io">
    <p class="io-title">Резервная копия прогресса</p>
    <p class="io-hint">Прогресс хранится в этом браузере. Раз в пару недель сохраняй файл.</p>
    <div class="io-actions">
      <button type="button" @click="download">Сохранить в файл</button>
      <label class="as-btn">
        Загрузить из файла
        <input type="file" accept="application/json,.json" class="sr-only" @change="upload" />
      </label>
    </div>
    <p v-if="msg" class="io-msg" :class="{ bad: !msg.ok }" role="status">{{ msg.text }}</p>
  </div>
</template>

<style scoped>
.io-title { margin: 0; font-family: var(--gs-font-display); font-weight: 700; }
.io-hint { margin: 4px 0 12px; font-size: 14px; color: var(--vp-c-text-2); }
.io-actions { display: flex; flex-wrap: wrap; gap: 8px; }
button, .as-btn {
  display: inline-flex;
  align-items: center;
  min-height: 40px;
  padding: 0 14px;
  border-radius: 4px;
  border: 1px solid var(--vp-c-divider);
  background: var(--vp-c-bg);
  color: var(--vp-c-text-1);
  font-size: 14px;
  cursor: pointer;
}
button:hover, .as-btn:hover { border-color: var(--vp-c-text-2); }
button:focus-visible, .as-btn:focus-within { outline: 2px solid var(--vp-c-brand-1); outline-offset: 2px; }
.sr-only { position: absolute; width: 1px; height: 1px; overflow: hidden; clip: rect(0 0 0 0); }
.io-msg { margin: 10px 0 0; font-size: 14px; color: var(--gs-ok); }
.io-msg.bad { color: var(--gs-bad); }
</style>
