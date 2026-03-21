<template>
  <div class="grid gap-4">
    <div class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
      <div class="flex items-start justify-between gap-3">
        <div>
          <div class="text-sm font-semibold text-slate-900">通道路由配置</div>
          <div class="mt-1 text-sm text-slate-600">左侧通道列表，右侧配置详情（权重、限额、熔断）。</div>
        </div>
        <div class="flex items-center gap-2">
          <button class="rounded-md bg-slate-900 px-3 py-2 text-sm font-semibold text-white" @click="newChannel">
            新建
          </button>
          <button class="rounded-md border border-slate-200 bg-white px-3 py-2 text-sm font-semibold text-slate-700" @click="reload">
            刷新
          </button>
        </div>
      </div>
    </div>

    <div class="grid grid-cols-12 gap-4">
      <div class="col-span-12 rounded-2xl border border-slate-200 bg-white p-4 shadow-sm md:col-span-4">
        <div class="text-xs font-semibold text-slate-500">通道列表</div>

        <div v-if="loading" class="mt-3 text-sm text-slate-500">加载中...</div>
        <div v-else class="mt-3 space-y-2">
          <button
            v-for="c in channels"
            :key="c.id"
            class="w-full rounded-xl border px-3 py-3 text-left hover:bg-slate-50"
            :class="selectedId === c.id ? 'border-slate-900' : 'border-slate-200'"
            @click="select(c.id)"
          >
            <div class="flex items-start justify-between gap-3">
              <div>
                <div class="text-sm font-semibold text-slate-900">{{ c.name }}</div>
                <div class="mt-1 text-xs text-slate-500">#{{ c.id }} · {{ c.pay_type || '-' }}</div>
              </div>
              <div class="flex flex-col items-end gap-1">
                <span
                  v-if="c.fuse_enabled"
                  class="rounded-full bg-rose-100 px-2 py-0.5 text-xs font-semibold text-rose-700"
                >
                  熔断中
                </span>
                <span
                  v-else-if="c.enabled"
                  class="rounded-full bg-emerald-100 px-2 py-0.5 text-xs font-semibold text-emerald-700"
                >
                  运行中
                </span>
                <span v-else class="rounded-full bg-slate-100 px-2 py-0.5 text-xs font-semibold text-slate-600">
                  已停用
                </span>
              </div>
            </div>
          </button>
        </div>
      </div>

      <div class="col-span-12 rounded-2xl border border-slate-200 bg-white p-6 shadow-sm md:col-span-8">
        <div class="flex items-start justify-between gap-3">
          <div class="text-xs text-slate-500">当前：{{ form.id ? `#${form.id}` : '新建' }}</div>
          <div v-if="saved" class="text-xs font-semibold text-emerald-700">已保存</div>
        </div>

        <div class="mt-4 grid grid-cols-12 gap-4">
          <label class="col-span-12 grid gap-1 md:col-span-6">
            <span class="text-xs font-medium text-slate-600">通道名称</span>
            <input v-model.trim="form.name" class="rounded-md border border-slate-200 px-3 py-2 text-sm" />
          </label>
          <label class="col-span-12 grid gap-1 md:col-span-6">
            <span class="text-xs font-medium text-slate-600">支付类型</span>
            <input v-model.trim="form.pay_type" class="rounded-md border border-slate-200 px-3 py-2 text-sm" />
          </label>

          <label class="col-span-12 grid gap-1">
            <span class="text-xs font-medium text-slate-600">上游 API 地址</span>
            <input v-model.trim="form.gateway_url" class="rounded-md border border-slate-200 px-3 py-2 text-sm" />
          </label>

          <label class="col-span-12 grid gap-1 md:col-span-6">
            <span class="text-xs font-medium text-slate-600">上游商户号</span>
            <input v-model.trim="form.upstream_merchant_no" class="rounded-md border border-slate-200 px-3 py-2 text-sm" />
          </label>
          <label class="col-span-12 grid gap-1 md:col-span-6">
            <span class="text-xs font-medium text-slate-600">签名密钥（Sign Secret）</span>
            <input v-model.trim="form.sign_secret" class="rounded-md border border-slate-200 px-3 py-2 text-sm" />
          </label>

          <label class="col-span-12 grid gap-1">
            <span class="text-xs font-medium text-slate-600">RSA 私钥</span>
            <textarea v-model="form.rsa_private_key" rows="7" class="rounded-md border border-slate-200 px-3 py-2 font-mono text-xs" />
          </label>

          <label class="col-span-12 grid gap-1 md:col-span-4">
            <span class="text-xs font-medium text-slate-600">权重（0-100）</span>
            <input v-model.number="form.weight" type="number" min="0" max="100" class="rounded-md border border-slate-200 px-3 py-2 text-sm" />
          </label>
          <label class="col-span-12 grid gap-1 md:col-span-4">
            <span class="text-xs font-medium text-slate-600">单笔最小金额（分）</span>
            <input v-model.number="form.min_amount" type="number" min="0" class="rounded-md border border-slate-200 px-3 py-2 text-sm" />
          </label>
          <label class="col-span-12 grid gap-1 md:col-span-4">
            <span class="text-xs font-medium text-slate-600">单笔最大金额（分）</span>
            <input v-model.number="form.max_amount" type="number" min="0" class="rounded-md border border-slate-200 px-3 py-2 text-sm" />
          </label>

          <div class="col-span-12 grid grid-cols-12 gap-4">
            <label class="col-span-12 flex items-center justify-between rounded-lg border border-slate-200 px-3 py-2 md:col-span-6">
              <div class="text-sm text-slate-700">启用通道</div>
              <input v-model="form.enabled" type="checkbox" class="h-4 w-4" />
            </label>
            <label class="col-span-12 flex items-center justify-between rounded-lg border border-slate-200 px-3 py-2 md:col-span-6">
              <div class="text-sm text-slate-700">熔断开关</div>
              <input v-model="form.fuse_enabled" type="checkbox" class="h-4 w-4" />
            </label>
          </div>
        </div>

        <div v-if="error" class="mt-4 rounded-lg border border-rose-200 bg-rose-50 p-3 text-sm text-rose-800">
          {{ error }}
        </div>

        <div class="mt-6 flex flex-wrap items-center gap-3">
          <button
            class="rounded-lg bg-slate-900 px-4 py-2 text-sm font-semibold text-white disabled:opacity-40"
            :disabled="saving || !adminTokenValue || !form.name"
            @click="save"
          >
            {{ saving ? '保存中...' : '保存配置' }}
          </button>
          <button class="rounded-lg border border-slate-200 bg-white px-4 py-2 text-sm font-semibold text-slate-700" @click="resetForm">
            重置
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, inject, onMounted, onUnmounted, ref } from 'vue'

type AdminChannelInfo = {
  id: number
  name: string
  pay_type: string
  gateway_url: string
  upstream_merchant_no: string
  rsa_private_key: string
  sign_secret: string
  weight: number
  min_amount: number
  max_amount: number
  enabled: boolean
  fuse_enabled: boolean
}

const adminToken = inject('adminToken') as { value: string } | undefined
const registerRefresh = inject('registerRefresh') as ((fn: () => void) => () => void) | undefined
const adminTokenValue = computed(() => adminToken?.value || '')

const loading = ref(false)
const saving = ref(false)
const error = ref('')
const saved = ref(false)

const channels = ref<AdminChannelInfo[]>([])
const selectedId = ref<number | null>(null)

const emptyForm: AdminChannelInfo = {
  id: 0,
  name: '',
  pay_type: '',
  gateway_url: '',
  upstream_merchant_no: '',
  rsa_private_key: '',
  sign_secret: '',
  weight: 100,
  min_amount: 0,
  max_amount: 0,
  enabled: true,
  fuse_enabled: false,
}

const form = ref<AdminChannelInfo>({ ...emptyForm })
const selected = computed(() => channels.value.find((c) => c.id === selectedId.value) || null)

function applySelected() {
  if (!selected.value) return
  form.value = { ...selected.value }
}

function resetForm() {
  if (selected.value) applySelected()
  else form.value = { ...emptyForm }
  saved.value = false
  error.value = ''
}

function select(id: number) {
  selectedId.value = id
  applySelected()
  saved.value = false
  error.value = ''
}

function newChannel() {
  selectedId.value = null
  form.value = { ...emptyForm }
  saved.value = false
  error.value = ''
}

async function reload() {
  error.value = ''
  saved.value = false
  loading.value = true
  try {
    const resp = await fetch('/v1/admin/channels', {
      headers: { 'X-Admin-Token': adminTokenValue.value },
    })
    if (!resp.ok) {
      error.value = `加载失败(${resp.status})`
      return
    }
    const data = (await resp.json()) as { channels: AdminChannelInfo[] }
    channels.value = data.channels || []
    if (selectedId.value && channels.value.some((c) => c.id === selectedId.value)) {
      applySelected()
    } else if (channels.value.length > 0) {
      select(channels.value[0].id)
    } else {
      newChannel()
    }
  } catch {
    error.value = '网络错误'
  } finally {
    loading.value = false
  }
}

async function save() {
  saving.value = true
  error.value = ''
  saved.value = false
  try {
    const isUpdate = form.value.id > 0
    const url = isUpdate ? `/v1/admin/channels/${form.value.id}` : '/v1/admin/channels'
    const method = isUpdate ? 'PUT' : 'POST'
    const resp = await fetch(url, {
      method,
      headers: {
        'Content-Type': 'application/json',
        'X-Admin-Token': adminTokenValue.value,
      },
      body: JSON.stringify({ ...form.value }),
    })
    if (!resp.ok) {
      error.value = `保存失败(${resp.status})`
      return
    }
    const data = (await resp.json()) as { channel: AdminChannelInfo }
    const ch = data.channel
    const idx = channels.value.findIndex((c) => c.id === ch.id)
    if (idx >= 0) channels.value[idx] = ch
    else channels.value.unshift(ch)
    selectedId.value = ch.id
    form.value = { ...ch }
    saved.value = true
  } catch {
    error.value = '网络错误'
  } finally {
    saving.value = false
  }
}

let unregister: (() => void) | null = null
onMounted(() => {
  void reload()
  if (registerRefresh) unregister = registerRefresh(() => void reload())
})
onUnmounted(() => {
  if (unregister) unregister()
})
</script>

