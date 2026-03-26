<template>
  <div class="grid gap-4">
    <ChannelsHeader @new-channel="openNew" @refresh="reload" />

    <div class="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm">
      <div class="flex flex-col gap-3 border-b border-slate-200 p-4 sm:flex-row sm:items-center sm:justify-between">
        <input
          v-model.trim="searchQuery"
          type="search"
          autocomplete="off"
          placeholder="搜索名称、ID、支付类型…"
          class="w-full max-w-md rounded-lg border border-slate-200 px-3 py-2 text-sm placeholder:text-slate-400"
        />
        <label class="flex items-center gap-2 text-sm text-slate-600">
          <span class="text-slate-500">匹配</span>
          <span class="font-mono text-slate-900">{{ filteredChannels.length }}</span>
          <span class="text-slate-500">条</span>
        </label>
      </div>

      <div class="overflow-x-auto">
        <table class="min-w-full text-left text-sm">
          <thead class="border-b border-slate-200 bg-slate-50 text-xs font-semibold uppercase tracking-wide text-slate-500">
            <tr>
              <th class="whitespace-nowrap px-4 py-3">ID</th>
              <th class="whitespace-nowrap px-4 py-3">名称</th>
              <th class="whitespace-nowrap px-4 py-3">支付类型</th>
              <th class="whitespace-nowrap px-4 py-3">代收</th>
              <th class="whitespace-nowrap px-4 py-3">代付</th>
              <th class="whitespace-nowrap px-4 py-3">状态</th>
              <th class="sticky right-0 z-20 whitespace-nowrap bg-slate-50 px-4 py-3 text-right">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="loading">
              <td colspan="7" class="px-4 py-8 text-center text-slate-500">加载中...</td>
            </tr>
            <tr v-else-if="!filteredChannels.length">
              <td colspan="7" class="px-4 py-8 text-center text-slate-500">暂无数据</td>
            </tr>
            <tr
              v-for="c in pagedChannels"
              v-else
              :key="c.id"
              class="group border-b border-slate-100 transition hover:bg-slate-50/80"
            >
              <td class="px-4 py-3 font-mono text-slate-800">#{{ c.id }}</td>
              <td class="px-4 py-3 font-medium text-slate-900">{{ c.name }}</td>
              <td class="px-4 py-3 font-mono text-xs text-slate-600">{{ c.payin_type || '—' }}</td>
              <td class="px-4 py-3">
                <span
                  class="rounded-full px-2 py-0.5 text-xs font-semibold"
                  :class="c.supports_payin ? 'bg-emerald-100 text-emerald-800' : 'bg-slate-100 text-slate-600'"
                >
                  {{ c.supports_payin ? '是' : '否' }}
                </span>
              </td>
              <td class="px-4 py-3">
                <span
                  class="rounded-full px-2 py-0.5 text-xs font-semibold"
                  :class="c.supports_payout ? 'bg-emerald-100 text-emerald-800' : 'bg-slate-100 text-slate-600'"
                >
                  {{ c.supports_payout ? '是' : '否' }}
                </span>
              </td>
              <td class="px-4 py-3">
                <span
                  v-if="c.fuse_enabled"
                  class="rounded-full bg-rose-100 px-2 py-0.5 text-xs font-semibold text-rose-700"
                >
                  熔断
                </span>
                <span
                  v-else-if="c.enabled"
                  class="rounded-full bg-emerald-100 px-2 py-0.5 text-xs font-semibold text-emerald-700"
                >
                  运行
                </span>
                <span v-else class="rounded-full bg-slate-100 px-2 py-0.5 text-xs font-semibold text-slate-600">停用</span>
              </td>
              <td class="sticky right-0 z-10 bg-white px-4 py-3 text-right group-hover:bg-slate-50/80">
                <button
                  type="button"
                  class="rounded-lg border border-slate-200 bg-white px-3 py-1.5 text-xs font-semibold text-slate-800 hover:border-slate-300"
                  @click="openEdit(c.id)"
                >
                  编辑
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <AdminPaginationBar
        v-if="!loading && filteredChannels.length"
        :total="total"
        :page="page"
        :page-size="pageSize"
        :page-count="pageCount"
        @update:page="page = $event"
        @update:page-size="pageSize = $event"
      />
    </div>

    <UiDrawer
      v-model="drawerOpen"
      :title="drawerTitle"
      subtitle="保存后路由与绑定校验将按通道能力生效。"
      max-width-class="max-w-2xl"
    >
      <ChannelFormCard
        v-if="drawerOpen"
        v-model="form"
        embedded
        hide-footer-actions
        :saving="saving"
        :saved="saved"
        :error="error"
        :can-save="!!adminTokenValue && !!form.name"
        @save="save"
        @reset="resetForm"
      />

      <template #footer>
        <div class="flex flex-wrap items-center justify-start gap-3">
          <button
            type="button"
            class="rounded-lg border border-slate-200 bg-white px-4 py-2 text-xs font-semibold text-slate-700"
            @click="resetForm"
          >
            重置
          </button>
          <button
            type="button"
            class="rounded-lg bg-slate-900 px-4 py-2 text-xs font-semibold text-white disabled:opacity-40"
            :disabled="saving || !adminTokenValue || !form.name"
            @click="save"
          >
            {{ saving ? '保存中...' : '保存配置' }}
          </button>
          <button
            type="button"
            class="rounded-lg border border-slate-200 bg-white px-4 py-2 text-xs font-semibold text-slate-700"
            @click="closeDrawer"
          >
            关闭
          </button>
        </div>
      </template>
    </UiDrawer>
  </div>
</template>

<script setup lang="ts">
import { computed, inject, onMounted, onUnmounted, ref, watch } from 'vue'

import AdminPaginationBar from '../../../components/AdminPaginationBar.vue'
import { UiDrawer } from '../../../../../../shared/ui'
import { useUiToast } from '../../../composables/useUiToast'
import { useClientPagination } from '../../../composables/useClientPagination'
import { adminGet, adminPost, adminPut } from '../../../lib/adminApi'

import ChannelFormCard from './ChannelFormCard.vue'
import ChannelsHeader from './ChannelsHeader.vue'
import type { AdminChannel } from './types'
import { emptyChannelForm } from './types'

const adminToken = inject('adminToken') as { value: string } | undefined
const registerRefresh = inject('registerRefresh') as ((fn: () => void) => () => void) | undefined
const adminTokenValue = computed(() => adminToken?.value || '')

const toast = useUiToast()
const loading = ref(false)
const saving = ref(false)
const error = ref('')
const saved = ref(false)
const drawerOpen = ref(false)
const searchQuery = ref('')

const channels = ref<AdminChannel[]>([])
const selectedId = ref<number | null>(null)

const form = ref<AdminChannel>(emptyChannelForm())

const isNew = computed(() => selectedId.value === null)

const drawerTitle = computed(() => (isNew.value ? '新建通道' : `编辑通道 · #${form.value.id}`))

const selected = computed(() => channels.value.find((c) => c.id === selectedId.value) || null)

const filteredChannels = computed(() => {
  const list = channels.value
  const s = searchQuery.value.trim().toLowerCase()
  if (!s) return list
  return list.filter((c) => {
    const idStr = String(c.id)
    const name = (c.name || '').toLowerCase()
    const pt = (c.payin_type || '').toLowerCase()
    return idStr.includes(s) || name.includes(s) || pt.includes(s)
  })
})

const { page, pageSize, total, pageCount, slice: pagedChannels } = useClientPagination(filteredChannels, 10)

watch(searchQuery, () => {
  page.value = 1
})

function applySelected() {
  if (!selected.value) return
  form.value = { ...selected.value }
}

function resetForm() {
  if (selected.value) applySelected()
  else form.value = emptyChannelForm()
  saved.value = false
  error.value = ''
}

function openEdit(id: number) {
  selectedId.value = id
  applySelected()
  saved.value = false
  error.value = ''
  drawerOpen.value = true
}

function openNew() {
  selectedId.value = null
  form.value = emptyChannelForm()
  saved.value = false
  error.value = ''
  drawerOpen.value = true
}

async function reload() {
  error.value = ''
  saved.value = false
  loading.value = true
  try {
    const data = await adminGet<{ channels: AdminChannel[] }>('/v1/admin/channels')
    channels.value = data.channels || []
    if (selectedId.value && channels.value.some((c) => c.id === selectedId.value)) {
      applySelected()
    }
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    error.value = msg
    toast.error(`加载通道列表失败：${msg}`)
  } finally {
    loading.value = false
  }
}

function closeDrawer() {
  drawerOpen.value = false
}

watch(drawerOpen, (open, wasOpen) => {
  if (wasOpen === true && open === false) void reload()
})

async function save() {
  saving.value = true
  error.value = ''
  saved.value = false
  const creating = form.value.id <= 0
  try {
    const isUpdate = form.value.id > 0
    const url = isUpdate ? `/v1/admin/channels/${form.value.id}` : '/v1/admin/channels'
    const body = { ...form.value } as Record<string, unknown>
    const data = isUpdate
      ? await adminPut<{ channel: AdminChannel }>(url, body)
      : await adminPost<{ channel: AdminChannel }>(url, body)
    const ch = data.channel
    const idx = channels.value.findIndex((c) => c.id === ch.id)
    if (idx >= 0) channels.value[idx] = ch
    else channels.value.unshift(ch)
    selectedId.value = ch.id
    form.value = { ...ch }
    saved.value = true
    closeDrawer()
    toast.success(creating ? '通道已创建' : '编辑已保存')
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    error.value = msg
    toast.error(`保存通道失败：${msg}`)
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
