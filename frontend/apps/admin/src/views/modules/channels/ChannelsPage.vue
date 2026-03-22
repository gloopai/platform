<template>
  <div class="grid gap-4">
    <ChannelsHeader @new-channel="newChannel" @refresh="reload" />

    <div class="grid grid-cols-12 gap-4">
      <ChannelList
        :channels="channels"
        :loading="loading"
        :selected-id="selectedId"
        @select="select"
      />

      <ChannelFormCard
        v-model="form"
        :saving="saving"
        :saved="saved"
        :error="error"
        :can-save="!!adminTokenValue && !!form.name"
        @save="save"
        @reset="resetForm"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, inject, onMounted, onUnmounted, ref } from 'vue'

import { adminGet, adminPost, adminPut } from '../../../lib/adminApi'

import ChannelFormCard from './ChannelFormCard.vue'
import ChannelList from './ChannelList.vue'
import ChannelsHeader from './ChannelsHeader.vue'
import type { AdminChannel } from './types'
import { emptyChannelForm } from './types'

const adminToken = inject('adminToken') as { value: string } | undefined
const registerRefresh = inject('registerRefresh') as ((fn: () => void) => () => void) | undefined
const adminTokenValue = computed(() => adminToken?.value || '')

const loading = ref(false)
const saving = ref(false)
const error = ref('')
const saved = ref(false)

const channels = ref<AdminChannel[]>([])
const selectedId = ref<number | null>(null)

const form = ref<AdminChannel>(emptyChannelForm())
const selected = computed(() => channels.value.find((c) => c.id === selectedId.value) || null)

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

function select(id: number) {
  selectedId.value = id
  applySelected()
  saved.value = false
  error.value = ''
}

function newChannel() {
  selectedId.value = null
  form.value = emptyChannelForm()
  saved.value = false
  error.value = ''
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
