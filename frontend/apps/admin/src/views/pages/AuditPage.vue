<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-lg font-semibold tracking-tight text-slate-900 sm:text-xl">运营与审计</h1>
      <p class="mt-1 text-sm text-slate-600">跨商户运营动作、合规审计与权限管理（能力逐步接入）。</p>
    </div>

    <div class="grid gap-4 md:grid-cols-3">
      <div class="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
        <div class="text-xs font-semibold text-slate-700">全站订单查询</div>
        <p class="mt-2 text-sm text-slate-600">按订单号、商户、状态、时间跨商户检索与导出。</p>
        <p class="mt-3 font-mono text-[10px] text-amber-800">待接入：/v1/admin/orders</p>
      </div>
      <div class="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
        <div class="text-xs font-semibold text-slate-700">人工补单 / 重发通知</div>
        <p class="mt-2 text-sm text-slate-600">异常订单处置、回调补发与备注留痕。</p>
        <p class="mt-3 font-mono text-[10px] text-amber-800">待接入：运营工单 API</p>
      </div>
      <div class="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
        <div class="text-xs font-semibold text-slate-700">管理员与 RBAC</div>
        <p class="mt-2 text-sm text-slate-600">角色：财务 / 运营 / 技术 / 只读；数据按商户隔离。</p>
        <p class="mt-3 font-mono text-[10px] text-amber-800">待接入：/v1/admin/users</p>
      </div>
    </div>

    <div class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
      <div class="text-sm font-semibold text-slate-900">操作审计日志</div>
      <p class="mt-2 text-sm text-slate-600">记录管理员关键操作（配置变更、资金类、权限变更），支持按人与时间检索。</p>
      <div class="mt-4 rounded-xl border border-dashed border-slate-200 bg-slate-50 py-8 text-center text-sm text-slate-500">
        审计流水列表（占位）
      </div>
    </div>
  </div>
</template>
