import { onMounted, ref } from 'vue'
import { fetchMerchantStatsByProduct, fetchMerchantSummary } from '@/api/console'
import { fetchMerchantOrders } from '@/api/orders'
import type { MerchantOrderItem, MerchantProductStatsResp, MerchantSummary } from '@/types/merchant.api'

const DEFAULT_ERR = '数据加载失败：请确认已登录且网关服务正常运行。'

/**
 * 控制台首页汇总数据；调试时可在 Vue DevTools 查看 summary / error。
 */
export function useMerchantSummary() {
  const summary = ref<MerchantSummary | null>(null)
  const byProduct = ref<MerchantProductStatsResp | null>(null)
  const payoutOverview = ref({
    count: 0,
    successCount: 0,
    failedCount: 0,
    amount: 0,
    successRate: 0,
  })
  const error = ref('')
  const loading = ref(false)

  async function load() {
    loading.value = true
    error.value = ''
    try {
      const [sum, productStats, payoutOrders] = await Promise.all([
        fetchMerchantSummary(),
        fetchMerchantStatsByProduct(),
        fetchMerchantOrders({ limit: 200 }, 'payout'),
      ])
      summary.value = sum
      byProduct.value = productStats
      payoutOverview.value = calcPayoutOverview(payoutOrders.orders || [])
    } catch {
      error.value = DEFAULT_ERR
    } finally {
      loading.value = false
    }
  }

  onMounted(() => {
    void load()
  })

  return { summary, byProduct, payoutOverview, error, loading, load }
}

function calcPayoutOverview(orders: MerchantOrderItem[]) {
  let count = 0
  let successCount = 0
  let failedCount = 0
  let amount = 0
  for (const o of orders) {
    count += 1
    amount += Number.isFinite(o.amount) ? o.amount : 0
    if (o.status === 1) successCount += 1
    if (o.status === 2) failedCount += 1
  }
  return {
    count,
    successCount,
    failedCount,
    amount,
    successRate: count > 0 ? successCount / count : 0,
  }
}
