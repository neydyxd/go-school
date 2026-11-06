// POST /api/sync: тело это состояние устройства, ответ это слитое состояние всех устройств.
// Авторизация: Authorization: Bearer <ключ синхронизации>. База: Neon Postgres (DATABASE_URL).

import { neon } from '@neondatabase/serverless'
import { parseDoc, type SyncDoc } from '../.vitepress/theme/sync.js'
import { syncState, type Db } from '../server/syncCore.js'

const MAX_BODY = 1_000_000

const json = (status: number, body: unknown) =>
  new Response(JSON.stringify(body), {
    status,
    headers: { 'content-type': 'application/json; charset=utf-8', 'cache-control': 'no-store' },
  })

let schema: Promise<unknown> | null = null

function neonDb(url: string): Db {
  const sql = neon(url)
  return {
    async claim(keyHash, doc) {
      await sql`INSERT INTO sync_state (slot, key_hash, doc) VALUES (1, ${keyHash}, ${JSON.stringify(doc)}::jsonb)
                ON CONFLICT (slot) DO NOTHING`
    },
    async read() {
      const rows = await sql`SELECT key_hash, doc, version FROM sync_state WHERE slot = 1`
      if (!rows.length) return null
      const r = rows[0] as { key_hash: string; doc: unknown; version: number }
      // ponytail: битый документ в базе считаем пустым, клиентская копия его перезапишет
      const doc: SyncDoc = parseDoc(r.doc) ?? { progress: { v: 2, lessons: {}, hw: {} }, cards: { v: 1, cards: {} } }
      return { keyHash: r.key_hash, doc, version: r.version }
    },
    async update(doc, version) {
      const rows = await sql`UPDATE sync_state SET doc = ${JSON.stringify(doc)}::jsonb, version = version + 1, updated_at = now()
                             WHERE slot = 1 AND version = ${version} RETURNING version`
      return rows.length > 0
    },
  }
}

export async function POST(req: Request): Promise<Response> {
  const url = process.env.DATABASE_URL
  if (!url) return json(503, { error: 'база синхронизации не подключена' })

  const text = await req.text()
  if (text.length > MAX_BODY) return json(413, { error: 'слишком большое тело запроса' })
  let incoming: SyncDoc | null = null
  try {
    incoming = parseDoc(JSON.parse(text))
  } catch {
    incoming = null
  }
  if (!incoming) return json(400, { error: 'тело не похоже на прогресс курса' })

  const sql = neon(url)
  // Одна строка на всё хранилище: slot всегда 1, владелец определяется хешем ключа.
  schema ??= sql`CREATE TABLE IF NOT EXISTS sync_state (
      slot int PRIMARY KEY DEFAULT 1 CHECK (slot = 1),
      key_hash text NOT NULL,
      doc jsonb NOT NULL,
      version int NOT NULL DEFAULT 1,
      updated_at timestamptz NOT NULL DEFAULT now())`.catch((e) => {
    schema = null
    throw e
  })
  await schema

  const key = (req.headers.get('authorization') ?? '').replace(/^Bearer\s+/i, '')
  const r = await syncState(neonDb(url), key, incoming)
  return r.status === 200 ? json(200, r.doc) : json(r.status, { error: r.error })
}
