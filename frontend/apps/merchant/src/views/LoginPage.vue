<template>
  <div class="relative flex min-h-full flex-col overflow-hidden bg-gradient-to-br from-slate-100/90 via-white to-slate-50">
    <div class="pointer-events-none absolute -left-24 top-20 h-72 w-72 rounded-full bg-slate-300/25 blur-3xl" />
    <div class="pointer-events-none absolute -right-16 bottom-10 h-64 w-64 rounded-full bg-slate-400/15 blur-3xl" />

    <div class="relative z-10 flex flex-1 flex-col items-center justify-center px-4 py-12 sm:py-16">
      <div class="mb-8 flex flex-col items-center text-center">
        <div
          class="flex h-14 w-14 items-center justify-center rounded-2xl bg-gradient-to-br from-slate-600 to-slate-800 text-lg font-bold text-white shadow-xl shadow-slate-900/20"
        >
          P
        </div>
        <h1 class="mt-4 text-2xl font-semibold tracking-tight text-slate-900 sm:text-3xl">欢迎回来</h1>
        <p class="mt-2 max-w-sm text-sm text-slate-600">登录商户中心，管理收款、对账与接入配置</p>
      </div>

      <div class="w-full max-w-md rounded-3xl border border-white/80 bg-white/90 p-8 shadow-2xl shadow-slate-200/60 backdrop-blur-sm">
        <div class="grid gap-4">
          <label class="grid gap-1.5">
            <span class="text-xs font-medium text-slate-600">商户号 merchant_id</span>
            <input
              v-model.trim="merchantId"
              class="rounded-xl border border-slate-200 bg-white px-4 py-3 text-sm text-slate-900 shadow-inner transition placeholder:text-slate-400 focus:border-slate-400 focus:outline-none focus:ring-2 focus:ring-slate-400/25"
              autocomplete="username"
            />
          </label>
          <label class="grid gap-1.5">
            <span class="text-xs font-medium text-slate-600">API 密钥 api_secret</span>
            <input
              v-model.trim="apiSecret"
              class="rounded-xl border border-slate-200 bg-white px-4 py-3 text-sm text-slate-900 shadow-inner transition focus:border-slate-400 focus:outline-none focus:ring-2 focus:ring-slate-400/25"
              type="password"
              autocomplete="current-password"
            />
          </label>
          <button
            type="button"
            class="mt-2 flex w-full items-center justify-center rounded-xl bg-slate-800 px-4 py-3 text-sm font-semibold text-white shadow-lg shadow-slate-900/15 transition hover:bg-slate-700 focus:outline-none focus:ring-2 focus:ring-slate-500/40 disabled:cursor-not-allowed disabled:opacity-40"
            :disabled="loading || !merchantId || !apiSecret"
            @click="login"
          >
            {{ loading ? '登录中…' : '进入商户中心' }}
          </button>

          <div
            v-if="error"
            class="rounded-2xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-800"
          >
            {{ error }}
          </div>
        </div>

        <div class="mt-6 rounded-2xl border border-slate-100 bg-slate-50/80 px-4 py-3 text-center text-xs text-slate-600">
          体验账号：<span class="font-mono font-medium text-slate-800">m_demo</span> /
          <span class="font-mono font-medium text-slate-800">demo_secret</span>
        </div>
      </div>

      <p class="mt-8 text-center text-xs text-slate-500">安全提示：请勿在公共设备保存密钥，定期轮换 api_secret。</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { MERCHANT_API } from '@/api/endpoints'
import { saveMerchantAuth, saveMerchantSession } from '@/lib/merchantApi'
import type { MerchantLoginResponse } from '@/types/merchant.api'

const router = useRouter()
const merchantId = ref(localStorage.getItem('merchant_id') || 'm_demo')
const apiSecret = ref(localStorage.getItem('merchant_secret') || 'demo_secret')
const loading = ref(false)
const error = ref('')

async function login() {
  loading.value = true
  error.value = ''
  try {
    const resp = await fetch(MERCHANT_API.login, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ merchant_id: merchantId.value, api_secret: apiSecret.value }),
    })
    if (!resp.ok) {
      error.value = `登录失败（${resp.status}）`
      return
    }
    const data = (await resp.json()) as MerchantLoginResponse
    saveMerchantAuth({ merchantId: merchantId.value, apiSecret: apiSecret.value })
    saveMerchantSession({ token: data.token, expiresAt: data.expires_at, merchantId: data.merchant_id })
    await router.replace('/console')
  } catch {
    error.value = '网络错误，请稍后重试'
  } finally {
    loading.value = false
  }
}
</script>
