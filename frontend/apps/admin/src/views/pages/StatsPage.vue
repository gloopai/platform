<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-lg font-semibold tracking-tight text-slate-900 sm:text-xl">系统概览</h1>
      <p class="mt-1 text-sm text-slate-600">聚合支付核心指标与健康度（部分指标待接入数据源）。</p>
    </div>

    <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
      <div class="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
        <div class="text-xs font-medium text-slate-500">今日平台交易额</div>
        <div class="mt-2 text-2xl font-semibold tabular-nums text-slate-900">—</div>
        <div class="mt-1 text-[10px] text-slate-400">待接入：汇总各商户成功订单</div>
      </div>
      <div class="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
        <div class="text-xs font-medium text-slate-500">今日订单笔数</div>
        <div class="mt-2 text-2xl font-semibold tabular-nums text-slate-900">—</div>
        <div class="mt-1 text-[10px] text-slate-400">待接入：全站订单统计</div>
      </div>
      <div class="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
        <div class="text-xs font-medium text-slate-500">支付成功率</div>
        <div class="mt-2 text-2xl font-semibold tabular-nums text-slate-900">—</div>
        <div class="mt-1 text-[10px] text-slate-400">待接入：成功/发起</div>
      </div>
      <div class="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
        <div class="text-xs font-medium text-slate-500">通道可用率</div>
        <div class="mt-2 text-2xl font-semibold tabular-nums text-slate-900">—</div>
        <div class="mt-1 text-[10px] text-slate-400">待接入：健康检查</div>
      </div>
    </div>

    <div class="grid gap-4 lg:grid-cols-2">
      <div class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
        <div class="text-sm font-semibold text-slate-900">实时大屏</div>
        <p class="mt-2 text-sm text-slate-600">TPS、成功率、地域分布、通道维度（占位）。</p>
        <div class="mt-4 grid gap-3 sm:grid-cols-3">
          <div class="rounded-xl bg-slate-50 p-4">
            <div class="text-xs text-slate-500">TPS</div>
            <div class="mt-2 text-xl font-semibold text-slate-900">—</div>
          </div>
          <div class="rounded-xl bg-slate-50 p-4">
            <div class="text-xs text-slate-500">成功率</div>
            <div class="mt-2 text-xl font-semibold text-slate-900">—</div>
          </div>
          <div class="rounded-xl bg-slate-50 p-4">
            <div class="text-xs text-slate-500">通道延迟 P99</div>
            <div class="mt-2 text-xl font-semibold text-slate-900">—</div>
          </div>
        </div>
      </div>
      <div class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
        <div class="text-sm font-semibold text-slate-900">待办与告警</div>
        <ul class="mt-3 space-y-2 text-sm text-slate-600">
          <li class="flex items-center gap-2">
            <span class="h-1.5 w-1.5 rounded-full bg-amber-400" />
            待审核提现 / 结算单（待接入）
          </li>
          <li class="flex items-center gap-2">
            <span class="h-1.5 w-1.5 rounded-full bg-rose-400" />
            通道异常或熔断（待接入）
          </li>
          <li class="flex items-center gap-2">
            <span class="h-1.5 w-1.5 rounded-full bg-slate-300" />
            对账差异笔数（待接入）
          </li>
        </ul>
      </div>
    </div>
  </div>
</template>
