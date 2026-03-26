<template>
  <div class="space-y-6">
    <PageHeader title="账户与安全" description="账户信息、接入参数与密码管理" />

    <div class="overflow-hidden rounded-2xl border border-slate-200/90 bg-white shadow-sm">
      <div class="border-b border-slate-100 bg-slate-50/80 px-4 py-3 text-sm font-semibold text-slate-900">账户信息</div>
      <div class="grid gap-4 px-4 py-4 sm:grid-cols-2">
        <InfoField label="商户号" :value="summary?.merchant_id || '-'" copyable />
        <InfoField label="登录邮箱" :value="summary?.email || '-'" />
        <InfoField label="账户状态" :value="statusText" />
      </div>
    </div>

    <div class="overflow-hidden rounded-2xl border border-slate-200/90 bg-white shadow-sm">
      <div class="border-b border-slate-100 bg-slate-50/80 px-4 py-3 text-sm font-semibold text-slate-900">API 接入信息</div>
      <div class="grid gap-4 px-4 py-4 sm:grid-cols-2">
        <InfoField label="AppID" :value="summary?.app_id || '-'" copyable />
        <InfoField
          label="AppSecret"
          :value="showSecret ? summary?.app_secret || '-' : maskSecret(summary?.app_secret)"
          copyable
        >
          <template #extra>
            <button
              type="button"
              class="rounded-lg border border-slate-200 bg-white px-2 py-1 text-xs font-medium text-slate-700 hover:border-slate-300"
              @click="showSecret = !showSecret"
            >
              {{ showSecret ? '隐藏' : '显示' }}
            </button>
          </template>
        </InfoField>
        <InfoField label="通知地址" :value="summary?.notify_url || '-'" copyable />
        <InfoField label="返回地址" :value="summary?.return_url || '-'" copyable />
        <InfoField label="IP 白名单" :value="summary?.ip_whitelist || '-'" />
      </div>
    </div>

    <div class="overflow-hidden rounded-2xl border border-slate-200/90 bg-white shadow-sm">
      <div class="border-b border-slate-100 bg-slate-50/80 px-4 py-3 text-sm font-semibold text-slate-900">修改密码</div>
      <form class="grid gap-4 px-4 py-4 sm:max-w-lg" @submit.prevent="submitChangePassword">
        <label class="grid gap-1.5">
          <span class="text-xs font-medium text-slate-600">当前密码</span>
          <input
            v-model.trim="passwordForm.old_password"
            type="password"
            autocomplete="current-password"
            class="rounded-lg border border-slate-200 px-3 py-2 text-sm text-slate-900 outline-none ring-slate-400/30 focus:ring-2"
          />
        </label>
        <label class="grid gap-1.5">
          <span class="text-xs font-medium text-slate-600">新密码</span>
          <input
            v-model.trim="passwordForm.new_password"
            type="password"
            autocomplete="new-password"
            class="rounded-lg border border-slate-200 px-3 py-2 text-sm text-slate-900 outline-none ring-slate-400/30 focus:ring-2"
          />
        </label>
        <label class="grid gap-1.5">
          <span class="text-xs font-medium text-slate-600">确认新密码</span>
          <input
            v-model.trim="passwordForm.confirm_password"
            type="password"
            autocomplete="new-password"
            class="rounded-lg border border-slate-200 px-3 py-2 text-sm text-slate-900 outline-none ring-slate-400/30 focus:ring-2"
          />
        </label>
        <p class="text-xs text-slate-500">密码需 8-64 位，并同时包含字母和数字。</p>
        <div class="flex items-center gap-2">
          <button
            type="submit"
            class="rounded-lg bg-slate-900 px-3 py-2 text-xs font-semibold text-white disabled:opacity-40"
            :disabled="changeLoading"
          >
            {{ changeLoading ? '更新中…' : '更新密码' }}
          </button>
          <span v-if="changeSuccess" class="text-xs font-medium text-emerald-700">{{ changeSuccess }}</span>
          <span v-else-if="changeError" class="text-xs font-medium text-rose-700">{{ changeError }}</span>
        </div>
      </form>
    </div>

    <ErrorCallout v-if="loadError" :message="loadError" />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { fetchMerchantSummary } from '@/api/console'
import { postMerchantChangePassword, postMerchantLogout } from '@/api/session'
import PageHeader from '@/components/layout/PageHeader.vue'
import ErrorCallout from '@/components/ui/ErrorCallout.vue'
import { clearMerchantSession } from '@/lib/merchantApi'
import type { MerchantSummary } from '@/types/merchant.api'
import InfoField from '@/views/pages/account/InfoField.vue'

const router = useRouter()
const summary = ref<MerchantSummary | null>(null)
const showSecret = ref(false)
const loadError = ref('')
const changeLoading = ref(false)
const changeError = ref('')
const changeSuccess = ref('')

const passwordForm = reactive({
  old_password: '',
  new_password: '',
  confirm_password: '',
})

const statusText = computed(() => {
  const s = summary.value?.status
  if (s === 1) return '正常'
  if (s === 0) return '停用'
  return '-'
})

onMounted(async () => {
  try {
    summary.value = await fetchMerchantSummary()
  } catch {
    loadError.value = '账户信息加载失败，请稍后重试。'
  }
})

function maskSecret(secret?: string): string {
  if (!secret) return '-'
  if (secret.length <= 8) return '********'
  return `${secret.slice(0, 4)}********${secret.slice(-4)}`
}

function validatePassword() {
  if (!passwordForm.old_password || !passwordForm.new_password || !passwordForm.confirm_password) return '请完整填写密码信息'
  if (passwordForm.new_password !== passwordForm.confirm_password) return '两次输入的新密码不一致'
  if (passwordForm.new_password === passwordForm.old_password) return '新密码不能与当前密码相同'
  if (passwordForm.new_password.length < 8 || passwordForm.new_password.length > 64) return '新密码长度需为 8-64 位'
  const hasLetter = /[A-Za-z]/.test(passwordForm.new_password)
  const hasDigit = /\d/.test(passwordForm.new_password)
  if (!hasLetter || !hasDigit) return '新密码需包含字母和数字'
  return ''
}

async function submitChangePassword() {
  changeError.value = ''
  changeSuccess.value = ''
  const err = validatePassword()
  if (err) {
    changeError.value = err
    return
  }
  changeLoading.value = true
  try {
    await postMerchantChangePassword({
      old_password: passwordForm.old_password,
      new_password: passwordForm.new_password,
    })
    changeSuccess.value = '密码修改成功，请重新登录。'
    await forceLogout()
  } catch (e) {
    changeError.value = e instanceof Error ? `修改失败：${e.message}` : '修改失败，请稍后重试。'
  } finally {
    changeLoading.value = false
  }
}

async function forceLogout() {
  try {
    await postMerchantLogout()
  } catch {
  } finally {
    clearMerchantSession()
    await router.replace('/login')
  }
}
</script>
