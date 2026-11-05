// Синхронизация прогресса и карточек между устройствами через /api/sync.
// Источник правды на устройстве остаётся localStorage: сеть только сливает копии.
// Ключ синхронизации живёт в localStorage и передаётся между устройствами
// ссылкой вида /#sync=<ключ> (фрагмент после # не уходит на сервер в логи).

import { reactive } from 'vue'
import { load as loadProgress, save as saveProgress } from './progress'
import { load as loadCards, save as saveCards } from './cards'
import { mergeDoc, parseDoc, type SyncDoc } from './sync'

export const SYNC_KEY = 'go-school:sync-key'
export const CHANGED = 'go-school:changed' // локальное изменение, пора отправить
export const SYNCED = 'go-school:synced' // localStorage обновлён с сервера, перечитать

const KEY_RE = /^[A-Za-z0-9_-]{43}$/
const DEBOUNCE_MS = 1500

export const syncStatus = reactive({ enabled: false, busy: false, last: '', error: '' })

const storage = (): Storage | null => {
  try {
    return typeof window === 'undefined' ? null : window.localStorage
  } catch {
    return null
  }
}

function getKey(): string | null {
  try {
    const k = storage()?.getItem(SYNC_KEY) ?? null
    return k && KEY_RE.test(k) ? k : null
  } catch {
    return null
  }
}

function setKey(k: string | null) {
  try {
    if (k) storage()?.setItem(SYNC_KEY, k)
    else storage()?.removeItem(SYNC_KEY)
  } catch {
    // ponytail: без localStorage синхронизация невозможна, статус покажет ошибку
  }
  syncStatus.enabled = !!k
}

function newKey(): string {
  const b = crypto.getRandomValues(new Uint8Array(32))
  return btoa(String.fromCharCode(...b)).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '')
}

const localDoc = (): SyncDoc => ({ progress: loadProgress(storage()), cards: loadCards(storage()) })

let inflight: Promise<number> | null = null
let again = false

// Возвращает HTTP-статус (0: нет сети или ключа).
export function syncNow(): Promise<number> {
  if (inflight) {
    again = true
    return inflight
  }
  inflight = run().finally(() => {
    inflight = null
    if (again) {
      again = false
      void syncNow()
    }
  })
  return inflight
}

async function run(): Promise<number> {
  const key = getKey()
  if (!key) return 0
  syncStatus.busy = true
  try {
    const res = await fetch('/api/sync', {
      method: 'POST',
      headers: { authorization: `Bearer ${key}`, 'content-type': 'application/json' },
      body: JSON.stringify(localDoc()),
    })
    if (res.status !== 200) {
      syncStatus.error =
        res.status === 403 ? 'Ключ не подходит: подключи устройство заново по ссылке с основного.'
        : res.status === 503 ? 'Сервер синхронизации ещё не настроен.'
        : `Сервер ответил ${res.status}, попробую позже.`
      return res.status
    }
    const remote = parseDoc(await res.json())
    if (!remote) {
      syncStatus.error = 'Сервер прислал непонятный ответ.'
      return 500
    }
    // Сливаем со свежей локальной копией: пока шёл запрос, могли отметить что-то ещё.
    const merged = mergeDoc(localDoc(), remote)
    saveProgress(storage(), merged.progress)
    saveCards(storage(), merged.cards)
    window.dispatchEvent(new Event(SYNCED))
    syncStatus.last = new Date().toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })
    syncStatus.error = ''
    return 200
  } catch {
    syncStatus.error = 'Нет связи с сервером, синхронизирую позже.'
    return 0
  } finally {
    syncStatus.busy = false
  }
}

let timer: ReturnType<typeof setTimeout> | undefined
function schedule() {
  if (!getKey()) return
  clearTimeout(timer)
  timer = setTimeout(() => void syncNow(), DEBOUNCE_MS)
}

export async function enable(): Promise<void> {
  setKey(newKey())
  if ((await syncNow()) === 403) {
    setKey(null)
    syncStatus.error = 'Синхронизация уже включена на другом устройстве. Открой там «Ссылку для другого устройства» и перейди по ней здесь.'
  }
}

// Принимает ссылку /#sync=<ключ> целиком или сам ключ.
export async function connect(input: string): Promise<boolean> {
  const m = input.trim().match(/(?:#sync=)?([A-Za-z0-9_-]{43})$/)
  if (!m) {
    syncStatus.error = 'Это не похоже на ссылку синхронизации.'
    return false
  }
  setKey(m[1])
  const status = await syncNow()
  if (status === 403) setKey(null)
  return status === 200
}

export function disconnect() {
  setKey(null)
  syncStatus.last = ''
  syncStatus.error = ''
}

export const connectLink = (): string => {
  const k = getKey()
  return k ? `${location.origin}/#sync=${k}` : ''
}

let started = false
export function initSync() {
  if (started || typeof window === 'undefined') return
  started = true
  syncStatus.enabled = !!getKey()
  const hash = location.hash.match(/^#sync=([A-Za-z0-9_-]{43})$/)
  if (hash) {
    history.replaceState(null, '', location.pathname + location.search)
    void connect(hash[1])
  } else {
    void syncNow()
  }
  window.addEventListener(CHANGED, schedule)
  window.addEventListener('online', () => void syncNow())
  document.addEventListener('visibilitychange', () => {
    if (document.visibilityState === 'visible') void syncNow()
  })
}
