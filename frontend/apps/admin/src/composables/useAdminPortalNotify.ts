import { onMounted, onUnmounted, ref, watch, type Ref } from 'vue'
import { useRouter } from 'vue-router'
import { adminSseUrl, loadAdminToken } from '../lib/adminApi'
import type { PortalNotifyEnvelope, PortalNotifyListItem } from '../lib/portalNotifyTypes'
import { useUiToast } from './useUiToast'

function sleep(ms: number) {
  return new Promise((r) => setTimeout(r, ms))
}

function parseLinkQuery(s: string | undefined): Record<string, string> {
  if (!s || s === '{}') return {}
  try {
    const o = JSON.parse(s) as Record<string, unknown>
    const q: Record<string, string> = {}
    for (const [k, v] of Object.entries(o)) {
      q[k] = v == null ? '' : String(v)
    }
    return q
  } catch {
    return {}
  }
}

function parseSSEBlocks(buffer: string): { frames: Array<{ event?: string; data?: string }>; rest: string } {
  const frames: Array<{ event?: string; data?: string }> = []
  let rest = buffer
  const sep = '\n\n'
  while (true) {
    const idx = rest.indexOf(sep)
    if (idx === -1) break
    const block = rest.slice(0, idx)
    rest = rest.slice(idx + sep.length)
    if (!block.trim()) continue
    let event: string | undefined
    let data: string | undefined
    for (const line of block.split('\n')) {
      if (line.startsWith(':')) continue
      if (line.startsWith('event:')) event = line.slice(6).trim()
      if (line.startsWith('data:')) data = line.slice(5).trim()
    }
    frames.push({ event, data })
  }
  return { frames, rest }
}

export function useAdminPortalNotify(adminToken: Ref<string>) {
  const router = useRouter()
  const toast = useUiToast()

  const recent = ref<PortalNotifyListItem[]>([])
  const pendingCount = ref(0)
  const connected = ref(false)

  let abort: AbortController | null = null
  let stopped = false
  let loopPromise: Promise<void> | null = null

  function clearPending() {
    pendingCount.value = 0
  }

  function pushRecent(env: PortalNotifyEnvelope) {
    const item: PortalNotifyListItem = {
      id: env.id,
      title: env.title || '通知',
      body: env.body || '',
      severity: env.severity || 'info',
      link_path: env.link_path || '',
      link_query_json: env.link_query_json,
      at: Date.now(),
    }
    recent.value = [item, ...recent.value.filter((x) => x.id !== env.id)].slice(0, 50)
    pendingCount.value++

    const sev = (env.severity || 'info').toLowerCase()
    const msg = env.title || '新通知'
    if (sev === 'error') toast.error(msg, 5200)
    else if (sev === 'warning') toast.warning(msg, 4200)
    else toast.info(msg, 3600)
  }

  function handleFrame(event: string | undefined, data: string | undefined) {
    if (event === 'connected') {
      connected.value = true
      return
    }
    if (event !== 'notification' || !data) return
    try {
      const env = JSON.parse(data) as PortalNotifyEnvelope
      if (env && env.id) pushRecent(env)
    } catch {
      // ignore
    }
  }

  function navigateTo(item: PortalNotifyListItem) {
    const raw = (item.link_path || '').trim()
    if (!raw) return
    const path = raw.startsWith('/') ? raw : `/${raw}`
    const query = parseLinkQuery(item.link_query_json)
    void router.push({ path, query })
  }

  async function runStream() {
    const url = adminSseUrl('/v1/admin/notifications/stream')
    const tok = loadAdminToken()
    if (!tok) {
      connected.value = false
      return
    }

    abort = new AbortController()
    const ac = abort
    let resp: Response
    try {
      resp = await fetch(url, {
        method: 'GET',
        headers: {
          Accept: 'text/event-stream',
          'X-Admin-Token': tok,
        },
        signal: ac.signal,
      })
    } catch {
      connected.value = false
      return
    }
    if (!resp.ok || !resp.body) {
      connected.value = false
      return
    }

    connected.value = true
    const reader = resp.body.getReader()
    const decoder = new TextDecoder()
    let buffer = ''

    try {
      while (!stopped) {
        const { done, value } = await reader.read()
        if (done) break
        buffer += decoder.decode(value, { stream: true })
        const { frames, rest } = parseSSEBlocks(buffer)
        buffer = rest
        for (const f of frames) {
          handleFrame(f.event, f.data)
        }
      }
    } catch {
      connected.value = false
    }
  }

  async function loop() {
    while (!stopped) {
      if (!loadAdminToken()) {
        connected.value = false
        await sleep(1500)
        continue
      }
      await runStream()
      if (stopped) break
      connected.value = false
      await sleep(3000)
    }
  }

  function start() {
    if (loopPromise) return
    stopped = false
    loopPromise = loop().finally(() => {
      loopPromise = null
    })
  }

  function stop() {
    stopped = true
    abort?.abort()
    abort = null
    connected.value = false
  }

  watch(adminToken, () => {
    abort?.abort()
    if (!adminToken.value) {
      recent.value = []
      pendingCount.value = 0
      connected.value = false
    }
  })

  onMounted(() => {
    start()
  })

  onUnmounted(() => {
    stop()
  })

  return {
    recent,
    pendingCount,
    connected,
    clearPending,
    navigateTo,
    stop,
  }
}
