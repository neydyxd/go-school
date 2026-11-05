<script setup lang="ts">
import { ref } from 'vue'
import { connect, connectLink, disconnect, enable, syncNow, syncStatus } from '../syncClient'

const link = ref('')
const input = ref('')
const copied = ref(false)

async function copyLink() {
  link.value = connectLink()
  try {
    await navigator.clipboard.writeText(link.value)
    copied.value = true
    setTimeout(() => (copied.value = false), 2500)
  } catch {
    // ponytail: без доступа к буферу ссылка просто видна в поле ниже
  }
}

async function onConnect() {
  if (await connect(input.value)) input.value = ''
}
</script>

<template>
  <div class="io">
    <p class="io-title">Синхронизация между устройствами</p>

    <template v-if="!syncStatus.enabled">
      <p class="io-hint">Включи на основном устройстве, затем открой выданную ссылку на телефоне. Аккаунт не нужен.</p>
      <div class="io-actions">
        <button type="button" :disabled="syncStatus.busy" @click="enable">Включить на этом устройстве</button>
      </div>
      <form class="connect" @submit.prevent="onConnect">
        <label for="gs-sync-link" class="io-hint">Уже включено на другом устройстве? Вставь ссылку оттуда:</label>
        <div class="io-actions">
          <input id="gs-sync-link" v-model="input" type="text" autocomplete="off" spellcheck="false" placeholder="https://…/#sync=…" />
          <button type="submit" :disabled="syncStatus.busy || !input.trim()">Подключить</button>
        </div>
      </form>
    </template>

    <template v-else>
      <p class="io-hint" role="status">
        <template v-if="syncStatus.busy">Синхронизирую…</template>
        <template v-else-if="syncStatus.last && !syncStatus.error">Синхронизировано в {{ syncStatus.last }}</template>
        <template v-else-if="!syncStatus.error">Включена</template>
      </p>
      <div class="io-actions">
        <button type="button" :disabled="syncStatus.busy" @click="syncNow">Синхронизировать сейчас</button>
        <button type="button" @click="copyLink">{{ copied ? 'Ссылка скопирована' : 'Ссылка для другого устройства' }}</button>
        <button type="button" @click="disconnect">Отключить здесь</button>
      </div>
      <template v-if="link">
        <label for="gs-sync-out" class="io-hint">Открой эту ссылку на телефоне (например, отправь себе в «Избранное»). Это ключ к твоему прогрессу, никому не пересылай.</label>
        <input id="gs-sync-out" :value="link" type="text" readonly @focus="($event.target as HTMLInputElement).select()" />
      </template>
    </template>

    <p v-if="syncStatus.error" class="io-msg bad" role="alert">{{ syncStatus.error }}</p>
  </div>
</template>

<style scoped>
.io { margin-top: 20px; }
.io-title { margin: 0; font-family: var(--gs-font-display); font-weight: 700; }
.io-hint { display: block; margin: 4px 0 12px; font-size: 14px; color: var(--vp-c-text-2); }
.io-actions { display: flex; flex-wrap: wrap; gap: 8px; }
.connect { margin-top: 14px; }
button, input {
  min-height: 40px;
  padding: 0 14px;
  border-radius: 4px;
  border: 1px solid var(--vp-c-divider);
  background: var(--vp-c-bg);
  color: var(--vp-c-text-1);
  font-size: 14px;
}
button { cursor: pointer; }
button:disabled { opacity: 0.6; cursor: default; }
input { flex: 1 1 200px; min-width: 0; width: 100%; font-family: var(--vp-font-family-mono); font-size: 13px; }
button:hover:not(:disabled) { border-color: var(--vp-c-text-2); }
button:focus-visible, input:focus-visible { outline: 2px solid var(--vp-c-brand-1); outline-offset: 2px; }
.io-msg { margin: 10px 0 0; font-size: 14px; }
.io-msg.bad { color: var(--gs-bad); }
</style>
