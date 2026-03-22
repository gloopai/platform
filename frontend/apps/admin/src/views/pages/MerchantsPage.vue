<template>
  <div class="grid gap-4">
    <div class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
      <div class="flex flex-wrap items-start justify-between gap-3">
        <div>
          <div class="text-sm font-semibold text-slate-900">商户管理</div>
          <div class="mt-1 text-sm text-slate-600">进件、费率、密钥与状态；在此为商户勾选收银台可见的支付产品（与「支付产品」菜单中的定义配合使用）。</div>
        </div>
        <div class="flex items-center gap-2">
          <button class="rounded-md bg-slate-900 px-3 py-2 text-sm font-semibold text-white" @click="openCreate">
            新建商户
          </button>
          <button class="rounded-md border border-slate-200 bg-white px-3 py-2 text-sm font-semibold text-slate-700" @click="reload">
            刷新
          </button>
        </div>
      </div>
    </div>

    <div class="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm">
      <table class="w-full text-left text-sm">
        <thead class="bg-slate-50 text-xs font-semibold text-slate-600">
          <tr>
            <th class="px-4 py-3">商户 ID</th>
            <th class="px-4 py-3">状态</th>
            <th class="px-4 py-3">费率(bps)</th>
            <th class="px-4 py-3">余额</th>
            <th class="px-4 py-3">支付产品</th>
            <th class="px-4 py-3">Notify URL</th>
            <th class="px-4 py-3">操作</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-slate-200">
          <tr v-if="loading">
            <td class="px-4 py-3 text-slate-600" colspan="7">加载中...</td>
          </tr>
          <tr v-else-if="merchants.length === 0">
            <td class="px-4 py-3 text-slate-600" colspan="7">暂无数据</td>
          </tr>
          <tr v-for="m in merchants" :key="m.merchant_id">
            <td class="px-4 py-3">
              <div class="font-medium text-slate-900">{{ m.merchant_id }}</div>
              <div class="mt-1 text-xs text-slate-500">secret: {{ maskSecret(m.api_secret) }}</div>
            </td>
            <td class="px-4 py-3">
              <span class="rounded-full px-2 py-0.5 text-xs font-semibold" :class="m.status === 1 ? 'bg-emerald-100 text-emerald-700' : 'bg-rose-100 text-rose-700'">
                {{ m.status === 1 ? '启用' : '锁定' }}
              </span>
            </td>
            <td class="px-4 py-3 text-slate-700">{{ m.rate_bps }}</td>
            <td class="px-4 py-3 text-slate-700">{{ formatMoney(m.balance) }}</td>
            <td class="px-4 py-3 text-xs text-slate-600">
              <span v-if="!m.pay_product_ids?.length" class="text-slate-400">未分配</span>
              <span v-else class="line-clamp-2">{{ payProductLabels(m.pay_product_ids) }}</span>
            </td>
            <td class="px-4 py-3 text-slate-700">
              <div class="max-w-xs break-all">{{ m.notify_url || '-' }}</div>
            </td>
            <td class="px-4 py-3">
              <div class="flex flex-wrap gap-2">
                <button class="rounded-md border border-slate-200 bg-white px-2 py-1 text-xs font-semibold text-slate-700 hover:bg-slate-50" @click="openEdit(m)">
                  编辑
                </button>
                <button
                  class="rounded-md border border-slate-200 bg-white px-2 py-1 text-xs font-semibold text-slate-700 hover:bg-slate-50"
                  @click="toggleLock(m)"
                >
                  {{ m.status === 1 ? '锁定' : '解锁' }}
                </button>
                <button
                  class="rounded-md border border-slate-200 bg-white px-2 py-1 text-xs font-semibold text-slate-700 hover:bg-slate-50"
                  @click="resetSecret(m)"
                >
                  重置密钥
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="error" class="rounded-xl border border-rose-200 bg-rose-50 p-4 text-sm text-rose-800">{{ error }}</div>

    <div v-if="modalOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
      <div class="w-full max-w-xl rounded-2xl bg-white shadow-xl">
        <div class="flex items-center justify-between border-b border-slate-200 px-5 py-4">
          <div class="text-sm font-semibold text-slate-900">{{ isEdit ? '编辑商户' : '新建商户' }}</div>
          <button class="text-sm font-semibold text-slate-600 hover:text-slate-900" @click="modalOpen = false">关闭</button>
        </div>
        <div class="grid gap-4 px-5 py-4">
          <label class="grid gap-1">
            <span class="text-xs font-medium text-slate-600">merchant_id</span>
            <input v-model.trim="form.merchant_id" class="rounded-md border border-slate-200 px-3 py-2 text-sm" :disabled="isEdit" />
          </label>
          <label v-if="!isEdit" class="grid gap-1">
            <span class="text-xs font-medium text-slate-600">api_secret（留空自动生成）</span>
            <input v-model.trim="form.api_secret" class="rounded-md border border-slate-200 px-3 py-2 text-sm" />
          </label>
          <div class="grid grid-cols-12 gap-4">
            <label class="col-span-12 grid gap-1 md:col-span-6">
              <span class="text-xs font-medium text-slate-600">rate_bps</span>
              <input v-model.number="form.rate_bps" type="number" min="0" class="rounded-md border border-slate-200 px-3 py-2 text-sm" />
            </label>
            <label class="col-span-12 grid gap-1 md:col-span-6">
              <span class="text-xs font-medium text-slate-600">状态</span>
              <select v-model.number="form.status" class="rounded-md border border-slate-200 px-3 py-2 text-sm">
                <option :value="1">启用</option>
                <option :value="0">锁定</option>
              </select>
            </label>
          </div>
          <label class="grid gap-1">
            <span class="text-xs font-medium text-slate-600">Notify URL</span>
            <input v-model.trim="form.notify_url" class="rounded-md border border-slate-200 px-3 py-2 text-sm" />
          </label>
          <label class="grid gap-1">
            <span class="text-xs font-medium text-slate-600">Return URL</span>
            <input v-model.trim="form.return_url" class="rounded-md border border-slate-200 px-3 py-2 text-sm" />
          </label>
          <label class="grid gap-1">
            <span class="text-xs font-medium text-slate-600">IP 白名单</span>
            <textarea v-model="form.ip_whitelist" rows="4" class="rounded-md border border-slate-200 px-3 py-2 font-mono text-xs" />
          </label>

          <div class="grid gap-2">
            <span class="text-xs font-medium text-slate-600">收银台可用支付产品</span>
            <p class="text-[11px] text-slate-500">仅勾选的产品会出现在该商户收银台；未分配则收银台无可用方式。</p>
            <div v-if="loadingProducts" class="text-xs text-slate-500">加载产品中…</div>
            <div v-else class="max-h-40 space-y-2 overflow-y-auto rounded-md border border-slate-200 p-3">
              <label v-for="p in payProducts" :key="p.id" class="flex cursor-pointer items-center gap-2 text-sm">
                <input v-model="form.pay_product_ids" type="checkbox" :value="p.id" class="rounded border-slate-300" />
                <span class="font-mono text-xs text-slate-800">{{ p.code }}</span>
                <span class="text-slate-600">{{ p.name }}</span>
              </label>
              <div v-if="!payProducts.length" class="text-xs text-slate-500">暂无支付产品，请先在「支付产品」中配置。</div>
            </div>
          </div>

          <div v-if="modalError" class="rounded-xl border border-rose-200 bg-rose-50 p-4 text-sm text-rose-800">{{ modalError }}</div>

          <div class="flex flex-wrap items-center gap-3">
            <button
              class="rounded-lg bg-slate-900 px-4 py-2 text-sm font-semibold text-white disabled:opacity-40"
              :disabled="saving || !form.merchant_id"
              @click="save"
            >
              {{ saving ? '保存中...' : '保存' }}
            </button>
            <button class="rounded-lg border border-slate-200 bg-white px-4 py-2 text-sm font-semibold text-slate-700" @click="modalOpen = false">
              取消
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { adminGet, adminPost, adminPut } from '../../lib/adminApi'

type AdminMerchantInfo = {
  merchant_id: string
  api_secret: string
  status: number
  rate_bps: number
  notify_url: string
  return_url: string
  ip_whitelist: string
  balance: number
  pay_product_ids?: number[]
}

type PayProductRow = { id: number; code: string; name: string }

const merchants = ref<AdminMerchantInfo[]>([])
const payProducts = ref<PayProductRow[]>([])
const loadingProducts = ref(false)
const loading = ref(false)
const saving = ref(false)
const error = ref('')

const modalOpen = ref(false)
const modalError = ref('')
const isEdit = ref(false)
const form = ref<AdminMerchantInfo & { pay_product_ids: number[] }>({
  merchant_id: '',
  api_secret: '',
  status: 1,
  rate_bps: 0,
  notify_url: '',
  return_url: '',
  ip_whitelist: '',
  balance: 0,
  pay_product_ids: [],
})

function maskSecret(s: string) {
  if (!s) return '-'
  if (s.length <= 10) return s
  return `${s.slice(0, 6)}...${s.slice(-4)}`
}

function formatMoney(v: number) {
  return `¥ ${(v / 100).toFixed(2)}`
}

function payProductLabels(ids: number[]) {
  const map = new Map(payProducts.value.map((p) => [p.id, p.code]))
  return (ids || []).map((id) => map.get(id) || String(id)).join('、')
}

async function loadPayProducts() {
  loadingProducts.value = true
  try {
    const res = await adminGet<{ products: PayProductRow[] }>('/v1/admin/pay_products')
    payProducts.value = res.products || []
  } catch {
    payProducts.value = []
  } finally {
    loadingProducts.value = false
  }
}

async function reload() {
  loading.value = true
  error.value = ''
  try {
    const res = await adminGet<{ merchants: AdminMerchantInfo[] }>('/v1/admin/merchants')
    merchants.value = res.merchants || []
  } catch {
    error.value = '加载失败：请确认已登录且网关已启动。'
  } finally {
    loading.value = false
  }
}

function openCreate() {
  isEdit.value = false
  modalError.value = ''
  form.value = {
    merchant_id: '',
    api_secret: '',
    status: 1,
    rate_bps: 0,
    notify_url: '',
    return_url: '',
    ip_whitelist: '',
    balance: 0,
    pay_product_ids: [],
  }
  modalOpen.value = true
}

function openEdit(m: AdminMerchantInfo) {
  isEdit.value = true
  modalError.value = ''
  form.value = {
    ...m,
    pay_product_ids: m.pay_product_ids ? [...m.pay_product_ids] : [],
  }
  modalOpen.value = true
}

function normalizePayProductIds(ids: unknown): number[] {
  const arr = Array.isArray(ids) ? ids : []
  const out: number[] = []
  for (const x of arr) {
    const n = typeof x === 'number' ? x : parseInt(String(x), 10)
    if (!Number.isFinite(n) || n <= 0) continue
    if (!out.includes(n)) out.push(n)
  }
  return out
}

async function save() {
  saving.value = true
  modalError.value = ''
  const payIds = normalizePayProductIds(form.value.pay_product_ids)
  try {
    if (isEdit.value) {
      const resp = await adminPut<{ merchant: AdminMerchantInfo }>(`/v1/admin/merchants/${encodeURIComponent(form.value.merchant_id)}`, {
        status: form.value.status,
        rate_bps: form.value.rate_bps,
        notify_url: form.value.notify_url,
        return_url: form.value.return_url,
        ip_whitelist: form.value.ip_whitelist,
        pay_product_ids: payIds,
      })
      const idx = merchants.value.findIndex((x) => x.merchant_id === resp.merchant.merchant_id)
      if (idx >= 0) merchants.value[idx] = resp.merchant
    } else {
      const resp = await adminPost<{ merchant: AdminMerchantInfo }>('/v1/admin/merchants', {
        merchant_id: form.value.merchant_id,
        api_secret: form.value.api_secret,
        rate_bps: form.value.rate_bps,
        notify_url: form.value.notify_url,
        return_url: form.value.return_url,
        ip_whitelist: form.value.ip_whitelist,
        pay_product_ids: payIds,
      })
      merchants.value.push(resp.merchant)
      merchants.value.sort((a, b) => a.merchant_id.localeCompare(b.merchant_id))
    }
    modalOpen.value = false
  } catch {
    modalError.value = '保存失败'
  } finally {
    saving.value = false
  }
}

async function toggleLock(m: AdminMerchantInfo) {
  error.value = ''
  try {
    const target = m.status === 1 ? 0 : 1
    const resp = await adminPut<{ merchant: AdminMerchantInfo }>(`/v1/admin/merchants/${encodeURIComponent(m.merchant_id)}`, { status: target })
    const idx = merchants.value.findIndex((x) => x.merchant_id === resp.merchant.merchant_id)
    if (idx >= 0) merchants.value[idx] = resp.merchant
  } catch {
    error.value = '更新状态失败'
  }
}

async function resetSecret(m: AdminMerchantInfo) {
  error.value = ''
  try {
    const resp = await adminPut<{ merchant: AdminMerchantInfo }>(`/v1/admin/merchants/${encodeURIComponent(m.merchant_id)}`, { reset_secret: true })
    const idx = merchants.value.findIndex((x) => x.merchant_id === resp.merchant.merchant_id)
    if (idx >= 0) merchants.value[idx] = resp.merchant
  } catch {
    error.value = '重置密钥失败'
  }
}

onMounted(() => {
  void loadPayProducts()
  void reload()
})
</script>
