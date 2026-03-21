<template>
  <div class="min-h-full bg-slate-100">
    <div class="mx-auto flex max-w-md flex-col gap-4 px-4 py-16">
      <div class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
        <div class="text-xl font-semibold text-slate-900">商户平台登录</div>
        <div class="mt-2 text-sm text-slate-600">使用 merchant_id 与 api_secret 登录。</div>

        <div class="mt-6 grid gap-3">
          <label class="grid gap-1">
            <span class="text-xs font-medium text-slate-600">merchant_id</span>
            <input v-model.trim="merchantId" class="rounded-md border border-slate-200 px-3 py-2 text-sm" />
          </label>
          <label class="grid gap-1">
            <span class="text-xs font-medium text-slate-600">api_secret</span>
            <input v-model.trim="apiSecret" class="rounded-md border border-slate-200 px-3 py-2 text-sm" type="password" />
          </label>
          <button
            class="mt-2 rounded-lg bg-slate-900 px-4 py-2 text-sm font-semibold text-white disabled:opacity-40"
            :disabled="loading || !merchantId || !apiSecret"
            @click="login"
          >
            {{ loading ? '登录中...' : '登录' }}
          </button>

          <div v-if="error" class="mt-2 rounded-xl border border-rose-200 bg-rose-50 p-4 text-sm text-rose-800">
            {{ error }}
          </div>
        </div>

        <div class="mt-6 rounded-xl bg-slate-50 p-4 text-xs text-slate-600">
          Demo 账号：m_demo / demo_secret
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { saveMerchantAuth, saveMerchantSession } from '../lib/merchantApi'

type MerchantLoginResp = {
  token: string
  expires_at: number
  merchant_id: string
}

const router = useRouter()
const merchantId = ref(localStorage.getItem('merchant_id') || 'm_demo')
const apiSecret = ref(localStorage.getItem('merchant_secret') || 'demo_secret')
const loading = ref(false)
const error = ref('')

async function login() {
  loading.value = true
  error.value = ''
  try {
    const resp = await fetch('/v1/merchant/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ merchant_id: merchantId.value, api_secret: apiSecret.value }),
    })
    if (!resp.ok) {
      error.value = `登录失败(${resp.status})`
      return
    }
    const data = (await resp.json()) as MerchantLoginResp
    saveMerchantAuth({ merchantId: merchantId.value, apiSecret: apiSecret.value })
    saveMerchantSession({ token: data.token, expiresAt: data.expires_at, merchantId: data.merchant_id })
    await router.replace('/console')
  } catch {
    error.value = '网络错误'
  } finally {
    loading.value = false
  }
}
</script>
