import { onMounted, ref } from 'vue'
import { fetchMerchantStatsByProduct, fetchMerchantSummary } from '@/api/console'
import type { MerchantProductStatsResp, MerchantSummary } from '@/types/merchant.api'

const DEFAULT_ERR = '数据加载失败：请确认已登录且网关服务正常运行。'

/**
 * 控制台首页汇总数据；调试时可在 Vue DevTools 查看 summary / error。
 */
export function useMerchantSummary() {
  const summary = ref<MerchantSummary | null>(null)
  const byProduct = ref<MerchantProductStatsResp | null>(null)
  const error = ref('')
  const loading = ref(false)

  async function load() {
    loading.value = true
    error.value = ''
    try {
      const [sum, productStats] = await Promise.all([
        fetchMerchantSummary(),
        fetchMerchantStatsByProduct(),
      ])
      summary.value = sum
      byProduct.value = productStats
    } catch {
      error.value = DEFAULT_ERR
    } finally {
      loading.value = false
    }
  }

  onMounted(() => {
    void load()
  })

  return { summary, byProduct, error, loading, load }
}
