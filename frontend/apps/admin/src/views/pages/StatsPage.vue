<template>
  <div class="grid gap-4">
    <div class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
      <div class="text-sm font-semibold text-slate-900">实时大屏</div>
      <div class="mt-2 text-sm text-slate-600">TPS、成功率、地域分布（占位，待接入数据源）。</div>
      <div class="mt-4 grid gap-3 md:grid-cols-3">
        <div class="rounded-xl bg-slate-50 p-4">
          <div class="text-xs text-slate-500">TPS</div>
          <div class="mt-2 text-2xl font-semibold text-slate-900">-</div>
        </div>
        <div class="rounded-xl bg-slate-50 p-4">
          <div class="text-xs text-slate-500">成功率</div>
          <div class="mt-2 text-2xl font-semibold text-slate-900">-</div>
        </div>
        <div class="rounded-xl bg-slate-50 p-4">
          <div class="text-xs text-slate-500">通道延迟</div>
          <div class="mt-2 text-2xl font-semibold text-slate-900">-</div>
        </div>
      </div>
    </div>
  </div>
</template>
