import { computed, onMounted, onUnmounted, ref } from 'vue'
import { parseApiEnvelope, unwrapApiData } from '@/lib/apiEnvelope'
import { merchantConsoleUrl } from '@/lib/http'

const offsetMs = ref(0)
const nowMs = ref(Date.now())
let tickTimer: number | null = null
let syncTimer: number | null = null

async function syncServerClock() {
  try {
    const resp = await fetch(merchantConsoleUrl('/health'))
    const text = await resp.text()
    try {
      const data = unwrapApiData(parseApiEnvelope<{ timestamp_ms?: number }>(text))
      if (typeof data.timestamp_ms === 'number') {
        offsetMs.value = data.timestamp_ms - Date.now()
      }
    } catch {
      /* ignore */
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
