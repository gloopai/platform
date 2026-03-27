<template>
  <div class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
    <div class="flex flex-wrap items-start justify-between gap-3">
      <div class="min-w-0 max-w-3xl">
        <div class="text-sm font-semibold text-slate-900">{{ title }}</div>
        <p class="mt-1 text-sm leading-relaxed text-slate-600">{{ subtitle }}</p>
        <p class="mt-2 text-[11px] leading-snug text-slate-500">
          列表中点击「编辑」在抽屉内分段维护；产品与通道绑定分别保存。
        </p>
      </div>
      <div class="flex shrink-0 items-center gap-2">
        <button
          type="button"
          class="rounded-md bg-slate-900 px-3 py-2 text-sm font-semibold text-white"
          @click="$emit('new-product')"
        >
          新建产品
        </button>
        <button
          type="button"
          class="rounded-md border border-slate-200 bg-white px-3 py-2 text-sm font-semibold text-slate-700 hover:bg-slate-50"
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
      '维护对外展示编码（code）、排序与启用；为每个产品绑定上游通道及权重。手续费规则在「通道管理」与「商户授权」中配置。',
  },
)

defineEmits<{
  'new-product': []
  refresh: []
}>()
</script>
