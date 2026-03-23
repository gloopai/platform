<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-lg font-semibold tracking-tight text-slate-900 sm:text-xl">退款与差错</h1>
      <p class="mt-1 max-w-3xl text-sm text-slate-600">
        <strong>MVP</strong>：当前主路径为代收下单、收银台支付与异步回调；<strong>退款单、差错单与调账流水尚未接入</strong>。需要核对订单状态时请使用
        <router-link to="/orders" class="font-medium text-slate-800 underline decoration-slate-300 underline-offset-2 hover:text-slate-950">
          全站订单
        </router-link>
        。
      </p>
    </div>

    <div class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
      <div class="text-xs font-semibold uppercase tracking-wide text-slate-400">后续规划（未实现）</div>
      <ul class="mt-3 list-inside list-disc space-y-2 text-sm text-slate-700">
        <li>退款申请与审核、原路退回与通道侧状态同步</li>
        <li>差错登记、长款挂账与人工调账</li>
        <li>与对账中心差异类型联动</li>
      </ul>
      <p class="mt-4 rounded-xl border border-amber-200 bg-amber-50 px-4 py-3 font-mono text-xs text-amber-900">
        待接入：/v1/admin/refunds 等；届时本页可替换为列表与详情。
      </p>
    </div>

    <div class="overflow-hidden rounded-2xl border border-slate-200/90 bg-white shadow-sm">
      <div class="border-b border-slate-100 bg-slate-50/90 px-4 py-3 text-sm font-semibold text-slate-800">退款单（占位）</div>
      <div class="px-4 py-12 text-center text-sm text-slate-500">暂无数据</div>
    </div>
  </div>
</template>
