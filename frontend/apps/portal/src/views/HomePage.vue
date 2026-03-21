<template>
  <div class="mx-auto max-w-6xl px-4 py-10">
    <div class="rounded-3xl bg-gradient-to-br from-slate-900 to-slate-700 p-10 text-white">
      <div class="text-sm font-semibold text-slate-200">聚合支付中台</div>
      <h1 class="mt-3 text-3xl font-semibold leading-tight md:text-4xl">稳定 · 高并发 · 多渠道</h1>
      <p class="mt-4 max-w-2xl text-sm text-slate-200 md:text-base">
        统一接入多个支付通道，提供下单、查单、回调、结算与路由调度能力。
      </p>
      <div class="mt-6 flex flex-wrap gap-3">
        <RouterLink class="rounded-lg bg-white px-4 py-2 text-sm font-semibold text-slate-900" to="/docs">
          快速集成
        </RouterLink>
        <RouterLink class="rounded-lg border border-white/30 px-4 py-2 text-sm font-semibold text-white" to="/products">
          查看产品
        </RouterLink>
      </div>
    </div>

    <div class="mt-10 grid gap-4 md:grid-cols-3">
      <div class="rounded-2xl border border-slate-200 bg-white p-6">
        <div class="text-sm font-semibold text-slate-900">快捷收单</div>
        <div class="mt-2 text-sm text-slate-600">扫码支付、H5 支付、App 拉起，统一订单模型。</div>
      </div>
      <div class="rounded-2xl border border-slate-200 bg-white p-6">
        <div class="text-sm font-semibold text-slate-900">代付/结算</div>
        <div class="mt-2 text-sm text-slate-600">支持 D+0 / T+1 结算节奏，流水可审计。</div>
      </div>
      <div class="rounded-2xl border border-slate-200 bg-white p-6">
        <div class="text-sm font-semibold text-slate-900">智能路由</div>
        <div class="mt-2 text-sm text-slate-600">按权重与金额段分流，熔断故障通道。</div>
      </div>
    </div>

    <div class="mt-10 rounded-2xl border border-slate-200 bg-white p-6">
      <div class="text-sm font-semibold text-slate-900">合作渠道</div>
      <div class="mt-4 grid grid-cols-2 gap-3 md:grid-cols-4">
        <div class="rounded-xl bg-slate-50 px-4 py-3 text-center text-sm font-semibold text-slate-700">支付宝</div>
        <div class="rounded-xl bg-slate-50 px-4 py-3 text-center text-sm font-semibold text-slate-700">微信支付</div>
        <div class="rounded-xl bg-slate-50 px-4 py-3 text-center text-sm font-semibold text-slate-700">Stripe</div>
        <div class="rounded-xl bg-slate-50 px-4 py-3 text-center text-sm font-semibold text-slate-700">银联</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { RouterLink } from 'vue-router'
</script>
