<template>
  <div class="min-h-full bg-slate-100">
    <div class="mx-auto flex max-w-md flex-col gap-4 px-4 py-16">
      <div class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
        <div class="text-xl font-semibold text-slate-900">总管理台登录</div>
        <div class="mt-2 text-sm text-slate-600">登录后才能访问管理功能。</div>

        <div class="mt-6 grid gap-3">
          <label class="grid gap-1">
            <span class="text-xs font-medium text-slate-600">用户名</span>
            <input v-model.trim="username" class="rounded-md border border-slate-200 px-3 py-2 text-sm" />
          </label>
          <label class="grid gap-1">
            <span class="text-xs font-medium text-slate-600">密码</span>
            <input v-model.trim="password" class="rounded-md border border-slate-200 px-3 py-2 text-sm" type="password" />
          </label>
          <button
            class="mt-2 rounded-lg bg-slate-900 px-4 py-2 text-sm font-semibold text-white disabled:opacity-40"
            :disabled="loading || !username || !password"
            @click="login"
          >
            {{ loading ? '登录中...' : '登录' }}
          </button>

          <div v-if="error" class="mt-2 rounded-xl border border-rose-200 bg-rose-50 p-4 text-sm text-rose-800">
            {{ error }}
          </div>
        </div>

        <div class="mt-6 rounded-xl bg-slate-50 p-4 text-xs text-slate-600">
          Demo 账号：admin / admin123
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { saveAdminSession } from '../lib/adminApi'

type AdminLoginResp = {
  token: string
  expires_at: number
}

const router = useRouter()
const username = ref('admin')
const password = ref('admin123')
const loading = ref(false)
const error = ref('')

async function login() {
  loading.value = true
  error.value = ''
  try {
    const resp = await fetch('/v1/admin/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username: username.value, password: password.value }),
    })
    if (!resp.ok) {
      error.value = `登录失败(${resp.status})`
      return
    }
    const data = (await resp.json()) as AdminLoginResp
    saveAdminSession({ token: data.token, expiresAt: data.expires_at })
    await router.replace('/stats')
  } catch {
    error.value = '网络错误'
  } finally {
    loading.value = false
  }
}
</script>
