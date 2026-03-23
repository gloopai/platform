<template>
  <div class="space-y-6">
    <header>
      <h1 class="text-xl font-semibold tracking-tight text-slate-900 sm:text-2xl">开发者中心</h1>
      <p class="mt-1 text-sm text-slate-600">包含开发配置、接口文档和联调工具，便于商户快速接入。</p>
    </header>

    <div class="rounded-2xl border border-slate-200/90 bg-white p-3 shadow-sm">
      <div class="flex flex-wrap gap-2">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          type="button"
          class="rounded-lg border px-3 py-1.5 text-xs font-semibold transition"
          :class="activeTab === tab.key ? 'border-slate-900 bg-slate-900 text-white' : 'border-slate-200 bg-white text-slate-700 hover:border-slate-300'"
          @click="activeTab = tab.key"
        >
          {{ tab.label }}
        </button>
      </div>
    </div>

    <DeveloperConfigTab
      v-show="activeTab === 'config'"
      v-model:merchant-id="merchantId"
      v-model:api-secret="apiSecret"
      v-model:ip-whitelist="ipWhitelist"
      v-model:notify-url="notifyUrl"
      :saving="configSaving"
      :save-error="configSaveError"
      :save-success="configSaveSuccess"
      @save="saveConfig"
    />
    <DeveloperDocsTab v-show="activeTab === 'docs'" />
    <DeveloperToolsTab v-show="activeTab === 'tools'" :merchant-id="merchantId" :api-secret="apiSecret" :notify-url="notifyUrl" />
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import DeveloperConfigTab from './developers/DeveloperConfigTab.vue'
import DeveloperDocsTab from './developers/DeveloperDocsTab.vue'
import DeveloperToolsTab from './developers/DeveloperToolsTab.vue'
import { loadMerchantAuth, saveMerchantAuth } from '@/lib/merchantApi'
import { fetchMerchantSummary, updateMerchantConfig } from '@/api/console'

const auth = loadMerchantAuth()
const merchantId = ref(auth.merchantId)
const apiSecret = ref(auth.apiSecret)
const ipWhitelist = ref('')
const notifyUrl = ref('')
const configSaving = ref(false)
const configSaveError = ref('')
const configSaveSuccess = ref('')

const tabs = [
  { key: 'config', label: '开发配置' },
  { key: 'docs', label: '开发文档' },
  { key: 'tools', label: '联调工具' },
] as const
const activeTab = ref<(typeof tabs)[number]['key']>('config')

watch([merchantId, apiSecret], () => {
  saveMerchantAuth({ merchantId: merchantId.value, apiSecret: apiSecret.value })
})

onMounted(async () => {
  try {
    const s = await fetchMerchantSummary()
    merchantId.value = s.merchant_id || merchantId.value
    apiSecret.value = s.api_secret || apiSecret.value
    notifyUrl.value = s.notify_url || ''
    ipWhitelist.value = s.ip_whitelist || ''
  } catch {
    // keep defaults for first load or unauthenticated state
  }
})

async function saveConfig() {
  configSaving.value = true
  configSaveError.value = ''
  configSaveSuccess.value = ''
  try {
    const resp = await updateMerchantConfig({
      notify_url: notifyUrl.value.trim(),
      ip_whitelist: ipWhitelist.value.trim(),
    })
    notifyUrl.value = resp.notify_url || ''
    ipWhitelist.value = resp.ip_whitelist || ''
    merchantId.value = resp.merchant_id || merchantId.value
    apiSecret.value = resp.api_secret || apiSecret.value
    configSaveSuccess.value = '保存成功'
    setTimeout(() => {
      configSaveSuccess.value = ''
    }, 1500)
  } catch (err) {
    configSaveError.value = err instanceof Error ? `保存失败：${err.message}` : '保存失败'
  } finally {
    configSaving.value = false
  }
}
</script>
