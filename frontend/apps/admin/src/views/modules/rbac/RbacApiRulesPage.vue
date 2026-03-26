<template>
  <div class="space-y-6">
    <p v-if="error" class="text-sm text-rose-600">{{ error }}</p>

    <div class="rounded-2xl border border-slate-200/90 bg-white p-5 shadow-sm">
      <div class="text-sm font-semibold text-slate-800">新增 / 更新规则</div>
      <p class="mt-1 text-xs text-slate-500">路径中的动态段写成 <span class="font-mono">:id</span>、<span class="font-mono">:merchant_id</span> 等，需与网关路由一致。</p>
      <div class="mt-4 grid gap-3 lg:grid-cols-12">
        <label class="lg:col-span-2 grid gap-1 text-xs font-medium text-slate-600">
          Method
          <input v-model.trim="newRuleMethod" type="text" class="rounded-lg border border-slate-200 px-3 py-2 font-mono text-sm uppercase" placeholder="GET" />
        </label>
        <label class="lg:col-span-5 grid gap-1 text-xs font-medium text-slate-600">
          Path Pattern
          <input
            v-model.trim="newRulePath"
            type="text"
            class="rounded-lg border border-slate-200 px-3 py-2 font-mono text-sm"
            placeholder="/v1/admin/merchants/:merchant_id"
          />
        </label>
        <label class="lg:col-span-3 grid gap-1 text-xs font-medium text-slate-600">
          perm_key
          <input v-model.trim="newRulePermKey" type="text" class="rounded-lg border border-slate-200 px-3 py-2 font-mono text-sm" />
        </label>
        <div class="lg:col-span-2 flex items-end">
          <button
            type="button"
            class="w-full rounded-lg bg-slate-900 px-4 py-2 text-sm font-semibold text-white disabled:opacity-40"
            :disabled="saving"
            @click="upsertApiRule"
          >
            保存
          </button>
        </div>
      </div>
    </div>

    <div class="overflow-hidden rounded-2xl border border-slate-200/90 bg-white shadow-sm">
      <div class="flex items-center justify-between gap-2 border-b border-slate-100 px-4 py-3">
        <div class="text-sm font-semibold text-slate-800">已配置规则</div>
        <button
          type="button"
          class="rounded-lg border border-slate-200 bg-white px-3 py-1.5 text-xs font-semibold text-slate-700 shadow-sm"
          :disabled="loading"
          @click="loadApiRules"
        >
          刷新
        </button>
      </div>
      <div class="overflow-x-auto">
        <table class="w-full min-w-[860px] text-left text-sm">
          <thead class="border-b border-slate-100 bg-slate-50/90 text-xs font-semibold uppercase tracking-wide text-slate-500">
            <tr>
              <th class="whitespace-nowrap px-4 py-3">Method</th>
              <th class="whitespace-nowrap px-4 py-3">Path Pattern</th>
              <th class="whitespace-nowrap px-4 py-3">perm_key</th>
              <th class="whitespace-nowrap px-4 py-3">状态</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100">
            <tr v-for="r in apiRules" :key="r.id" class="hover:bg-slate-50/70">
              <td class="px-4 py-3 font-mono text-xs text-slate-800">{{ r.method }}</td>
              <td class="px-4 py-3 font-mono text-xs text-slate-700">{{ r.path_pattern }}</td>
              <td class="px-4 py-3 font-mono text-xs text-slate-700">{{ r.perm_key }}</td>
              <td class="px-4 py-3">
                <span
                  class="inline-flex rounded-full px-2 py-0.5 text-[10px] font-semibold"
                  :class="r.status === 1 ? 'bg-emerald-100 text-emerald-800' : 'bg-slate-200 text-slate-700'"
                >
                  {{ r.status === 1 ? '启用' : '停用' }}
                </span>
              </td>
            </tr>
            <tr v-if="!apiRules.length && !loading">
              <td class="px-4 py-12 text-center text-slate-500" colspan="4">暂无规则</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'

import { adminGet, adminPost } from '../../../lib/adminApi'

type AdminApiRule = { id: number; method: string; path_pattern: string; perm_key: string; status: number; remark: string }

const loading = ref(true)
const saving = ref(false)
const error = ref('')
const apiRules = ref<AdminApiRule[]>([])

const newRuleMethod = ref('GET')
const newRulePath = ref('')
const newRulePermKey = ref('')

async function loadApiRules() {
  loading.value = true
  error.value = ''
  try {
    const r = await adminGet<{ rules: AdminApiRule[] }>('/v1/admin/rbac/api_rules')
    apiRules.value = r.rules || []
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
    apiRules.value = []
  } finally {
    loading.value = false
  }
}

async function upsertApiRule() {
  const method = newRuleMethod.value.trim().toUpperCase()
  const path_pattern = newRulePath.value.trim()
  const perm_key = newRulePermKey.value.trim()
  if (!method || !path_pattern || !perm_key) {
    error.value = '请填写 Method / Path Pattern / perm_key'
    return
  }
  saving.value = true
  error.value = ''
  try {
    await adminPost('/v1/admin/rbac/api_rules', { method, path_pattern, perm_key, status: 1, remark: '' })
    newRulePath.value = ''
    newRulePermKey.value = ''
    await loadApiRules()
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    saving.value = false
  }
}

onMounted(() => void loadApiRules())
</script>
