import { computed, onMounted, onUnmounted, ref } from 'vue'
import { merchantConsoleUrl } from '@/lib/http'

const offsetMs = ref(0)
const nowMs = ref(Date.now())
let tickTimer: number | null = null
let syncTimer: number | null = null

async function syncServerClock() {
  try {
    const resp = await fetch(merchantConsoleUrl('/health'))
    if (!resp.ok) return
    const j = (await resp.json()) as { timestamp_ms?: number }
    if (typeof j.timestamp_ms === 'number') {
      offsetMs.value = j.timestamp_ms - Date.now()
    }
  } catch {
  }
}

export function useServerClock() {
  const serverTimeText = computed(() => {
    const d = new Date(nowMs.value + offsetMs.value)
    return d.toLocaleString('zh-CN', {
      hour12: false,
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    })
  })

  onMounted(() => {
    void syncServerClock()
    tickTimer = window.setInterval(() => {
      nowMs.value = Date.now()
    }, 1000)
    syncTimer = window.setInterval(() => {
      void syncServerClock()
    }, 15000)
  })
  onUnmounted(() => {
    if (tickTimer != null) window.clearInterval(tickTimer)
    if (syncTimer != null) window.clearInterval(syncTimer)
  })

  return { serverTimeText }
}
