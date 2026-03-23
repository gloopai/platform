<template>
  <div class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
    <div class="flex items-start justify-between gap-3">
      <div>
        <div class="text-sm font-semibold text-slate-900">{{ title }}</div>
        <div class="mt-1 text-sm text-slate-600">{{ subtitle }}</div>
      </div>
      <div class="flex items-center gap-2">
        <button
          type="button"
          class="rounded-md bg-slate-900 px-3 py-2 text-sm font-semibold text-white"
          @click="$emit('new-product')"
        >
          新建产品
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
withDefaults(
  defineProps<{
    title?: string
    subtitle?: string
  }>(),
  {
    title: '支付产品与上游通道',
    subtitle:
      '维护对外展示编码（code）、排序与启用；为每个产品绑定上游通道及权重。费率在「通道管理」与「商户管理」中配置。',
  },
)

defineEmits<{
  'new-product': []
  refresh: []
}>()
</script>
