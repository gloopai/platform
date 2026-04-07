<template>
  <div class="min-h-full bg-slate-100">
    <div class="mx-auto flex max-w-lg flex-col gap-4 px-4 py-12">
      <div class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
        <div class="text-xl font-semibold text-slate-900">绑定谷歌验证器</div>
        <p class="mt-2 text-sm text-slate-600">
          管理后台要求使用谷歌验证器（Google Authenticator 等 TOTP 应用）保护账号。请完成绑定后再使用其他功能。
        </p>

        <div v-if="loading" class="mt-8 text-center text-sm text-slate-500">加载中...</div>

        <div v-else class="mt-6 space-y-6">
          <!-- 有二维码数据时始终展示（与 mfa_pending 无关；此前误用 !mfaPending 导致首次生成后二维码被隐藏） -->
          <div v-if="qrDataUrl" class="space-y-3">
            <div class="text-sm font-medium text-slate-800">1. 使用谷歌验证器扫描下方二维码</div>
            <div class="flex justify-center">
              <img
                :src="qrDataUrl"
                alt="谷歌验证器绑定二维码"
                class="h-52 w-52 rounded-xl border border-slate-200 bg-white p-2 shadow-sm"
              />
            </div>
            <div class="rounded-lg bg-slate-50 p-3 text-xs text-slate-600">
              无法扫描？在应用中选择「输入设置密钥」，并填写：
              <span class="mt-1 block break-all font-mono text-sm text-slate-900">{{ secret }}</span>
            </div>
          </div>

          <div v-if="mfaPending && !qrDataUrl" class="rounded-lg border border-amber-200 bg-amber-50 p-3 text-sm text-amber-900">
            系统里已有未完成的密钥。若当前手机已添加该账号，直接输入 6 位动态码；若需重新扫码，请点击「重新生成密钥」以显示二维码。
          </div>

          <div class="space-y-2">
            <div class="text-sm font-medium text-slate-800">
              {{ qrDataUrl ? '2. 输入应用中的 6 位验证码' : '输入应用中的 6 位验证码' }}
            </div>
            <input
              v-model.trim="code"
              maxlength="8"
              autocomplete="one-time-code"
              inputmode="numeric"
              placeholder="6 位数字"
              class="w-full rounded-md border border-slate-200 px-3 py-2 font-mono text-sm tracking-widest"
              @keyup.enter="confirm"
            />
          </div>

          <div class="flex flex-wrap gap-2">
            <button
              type="button"
              class="rounded-lg bg-slate-900 px-4 py-2 text-xs font-semibold text-white disabled:opacity-40"
              :disabled="saving || !code"
              @click="confirm"
            >
              {{ saving ? '提交中...' : '完成绑定' }}
            </button>
            <button
              v-if="mfaPending || qrDataUrl"
              type="button"
              class="rounded-lg border border-slate-300 bg-white px-4 py-2 text-xs font-semibold text-slate-700 disabled:opacity-40"
              :disabled="regenerating"
              @click="regenerate"
            >
              {{ regenerating ? '生成中...' : '重新生成密钥' }}
            </button>
            <button
              type="button"
              class="rounded-lg px-4 py-2 text-xs font-medium text-slate-500 hover:text-slate-800"
              @click="logout"
            >
              退出登录
            </button>
          </div>

          <div v-if="error" class="rounded-xl border border-rose-200 bg-rose-50 p-3 text-sm text-rose-800">
            {{ error }}
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { adminGet, adminPost, clearAdminSession, setAdminMfaGate } from '../lib/adminApi'

type Me = {
  mfa_enabled: number
  mfa_pending: number
}

const router = useRouter()
const loading = ref(true)
const saving = ref(false)
const regenerating = ref(false)
const error = ref('')
const mfaPending = ref(false)
const qrDataUrl = ref('')
const secret = ref('')
const code = ref('')

async function loadMe() {
  const me = await adminGet<Me>('/v1/admin/me')
  if (me.mfa_enabled === 1) {
    setAdminMfaGate(true)
    await router.replace('/home')
    return false
  }
  mfaPending.value = me.mfa_pending === 1
  return true
}

async function runSetup() {
  error.value = ''
  const r = await adminPost<{ secret: string; qr_data_url: string }>('/v1/admin/mfa/setup', {})
  secret.value = r.secret
  qrDataUrl.value = r.qr_data_url
}

onMounted(async () => {
  loading.value = true
  error.value = ''
  try {
    const cont = await loadMe()
    if (!cont) return
    if (!mfaPending.value) {
      await runSetup()
    }
  } catch (e) {
    error.value = e instanceof Error ? e.message : '加载失败'
  } finally {
    loading.value = false
  }
})

async function confirm() {
  const c = code.value.trim()
  if (!c) return
  saving.value = true
  error.value = ''
  try {
    await adminPost('/v1/admin/mfa/confirm', { code: c })
    setAdminMfaGate(true)
    await router.replace('/home')
  } catch (e) {
    error.value = e instanceof Error ? e.message : '绑定失败'
  } finally {
    saving.value = false
  }
}

async function regenerate() {
  regenerating.value = true
  error.value = ''
  code.value = ''
  try {
    await runSetup()
  } catch (e) {
    error.value = e instanceof Error ? e.message : '生成失败'
  } finally {
    regenerating.value = false
  }
}

async function logout() {
  try {
    await adminPost('/v1/admin/logout', {})
  } catch {
  }
  clearAdminSession()
  localStorage.removeItem('admin_allowed_paths')
  await router.replace('/login')
}
</script>
