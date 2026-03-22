<template>
  <div class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
    <div class="flex items-start justify-between gap-3">
      <div>
        <div class="text-sm font-semibold text-slate-900">商户管理</div>
        <div class="mt-1 text-sm text-slate-600">左侧选择商户查看与编辑基础信息；支付产品在右侧单独绑定，与「支付产品」菜单中的定义一致。</div>
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
