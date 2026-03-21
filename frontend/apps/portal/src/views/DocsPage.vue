<template>
  <div class="mx-auto max-w-6xl px-4 py-10">
    <h1 class="text-2xl font-semibold text-slate-900">开发者中心</h1>
    <p class="mt-2 text-sm text-slate-600">快速集成、SDK 下载与在线文档入口。</p>

    <div class="mt-8 rounded-2xl border border-slate-200 bg-white p-6">
      <div class="text-sm font-semibold text-slate-900">快速集成</div>
      <div class="mt-4 grid gap-3 md:grid-cols-3">
        <div class="rounded-xl bg-slate-50 p-4">
          <div class="text-xs font-semibold text-slate-700">1. 创建订单</div>
          <div class="mt-1 text-sm text-slate-600">商户调用 CreateOrder，获取 order_no。</div>
        </div>
        <div class="rounded-xl bg-slate-50 p-4">
          <div class="text-xs font-semibold text-slate-700">2. 跳转收银台</div>
          <div class="mt-1 text-sm text-slate-600">用户完成支付，平台异步处理回调。</div>
        </div>
        <div class="rounded-xl bg-slate-50 p-4">
          <div class="text-xs font-semibold text-slate-700">3. Webhook 通知</div>
          <div class="mt-1 text-sm text-slate-600">NSQ 消费者阶梯重试推送通知商户。</div>
        </div>
      </div>
    </div>

    <div class="mt-6 grid gap-4 md:grid-cols-2">
      <div class="rounded-2xl border border-slate-200 bg-white p-6">
        <div class="text-sm font-semibold text-slate-900">SDK 下载</div>
        <div class="mt-3 grid gap-2 text-sm">
          <a class="font-medium text-slate-700 hover:text-slate-900" href="#" @click.prevent>Go SDK（示例）</a>
          <a class="font-medium text-slate-700 hover:text-slate-900" href="#" @click.prevent>PHP SDK（示例）</a>
          <a class="font-medium text-slate-700 hover:text-slate-900" href="#" @click.prevent>Java SDK（示例）</a>
          <a class="font-medium text-slate-700 hover:text-slate-900" href="#" @click.prevent>Python SDK（示例）</a>
        </div>
      </div>
      <div class="rounded-2xl border border-slate-200 bg-white p-6">
        <div class="text-sm font-semibold text-slate-900">在线文档</div>
        <div class="mt-3 text-sm text-slate-600">可接入 Swagger / Redoc 或自建 GitBook 样式文档站。</div>
        <div class="mt-4 flex flex-wrap gap-2">
          <a class="rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm font-semibold text-slate-700 hover:bg-slate-50" href="#" @click.prevent>
            Swagger
          </a>
          <a class="rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm font-semibold text-slate-700 hover:bg-slate-50" href="#" @click.prevent>
            GitBook
          </a>
        </div>
      </div>
    </div>
  </div>
</template>
