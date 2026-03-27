<template>
  <div class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
    <div class="flex items-start justify-between gap-3">
      <div>
        <div class="text-sm font-semibold text-slate-900">商户列表</div>
        <div class="mt-1 text-sm text-slate-600">
          「新建」仅录入邮箱并生成凭据；列表中点击「编辑」在抽屉中维护资料与代收/代付授权。
        </div>
      </div>
      <div class="flex items-center gap-2">
        <button
          type="button"
          class="rounded-md bg-slate-900 px-3 py-2 text-sm font-semibold text-white"
          @click="$emit('new-merchant')"
        >
          新建
        </button>
        <button
          type="button"
          class="rounded-md border border-slate-200 bg-white px-3 py-2 text-sm font-semibold text-slate-700"
          @click="$emit('refresh')"
        >
          刷新
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
defineEmits<{
  'new-merchant': []
  refresh: []
}>()
</script>
