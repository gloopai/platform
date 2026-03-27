<template>
  <div class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
    <div class="flex flex-wrap items-start justify-between gap-3">
      <div class="min-w-0 max-w-3xl">
        <div class="text-sm font-semibold text-slate-900">通道管理</div>
        <p class="mt-1 text-sm leading-relaxed text-slate-600">
          列表中点击「编辑」在抽屉内分段维护；保存后路由与产品绑定按通道能力生效。路由策略见「路由策略」菜单。
        </p>
        <p class="mt-2 text-[11px] leading-snug text-slate-500">新建或编辑均可分「基本设置 / 上游对接 / 费率与能力」填写，底部一次保存全部字段。</p>
      </div>
      <div class="flex shrink-0 items-center gap-2">
        <button
          type="button"
          class="rounded-md bg-slate-900 px-3 py-2 text-sm font-semibold text-white"
          @click="$emit('new-channel')"
        >
          新建
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
defineEmits<{
  'new-channel': []
  refresh: []
}>()
</script>
