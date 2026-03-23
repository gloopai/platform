<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-lg font-semibold tracking-tight text-slate-900 sm:text-xl">结算与提现</h1>
      <p class="mt-1 max-w-3xl text-sm text-slate-600">
        <strong>MVP</strong>：主路径先跑通<strong>收单 → 支付成功 → 订单</strong>；结算周期、提现审核、打款与手续费清算后续迭代。当前可先通过
        <router-link to="/orders" class="font-medium text-slate-800 underline decoration-slate-300 underline-offset-2 hover:text-slate-950">
          全站订单
        </router-link>
        与
        <router-link to="/reconcile" class="font-medium text-slate-800 underline decoration-slate-300 underline-offset-2 hover:text-slate-950">
          对账中心
        </router-link>
        核对平台侧数据。
      </p>
    </div>

    <div class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
      <div class="text-xs font-semibold uppercase tracking-wide text-slate-400">后续规划（未实现）</div>
      <ul class="mt-3 list-inside list-disc space-y-2 text-sm text-slate-700">
        <li>结算单生成、商户确认与 T+N / D+1 策略</li>
        <li>提现申请、风控与打款通道</li>
        <li>与对账中心、资金流水闭环</li>
      </ul>
      <p class="mt-4 rounded-xl border border-amber-200 bg-amber-50 px-4 py-3 font-mono text-xs text-amber-900">
        待接入：/v1/admin/settlement 等
      </p>
    </div>
  </div>
</template>
