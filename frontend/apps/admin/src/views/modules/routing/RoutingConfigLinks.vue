<template>
  <div class="grid gap-4 md:grid-cols-3">
    <RouterLink
      to="/merchant-products"
      class="group rounded-xl border border-slate-200 bg-white p-4 shadow-sm transition hover:border-indigo-300 hover:shadow-md"
    >
      <div class="text-sm font-semibold text-slate-900">支付产品与通道</div>
      <p class="mt-2 text-sm text-slate-600">
        为每个对外支付产品绑定多条上游通道，设置<strong>权重</strong>与启用状态；交易服务在同一产品内按权重做加权随机选路。
      </p>
      <div class="mt-3 text-sm font-medium text-indigo-600 group-hover:underline">去配置 →</div>
    </RouterLink>

    <RouterLink
      to="/merchants"
      class="group rounded-xl border border-slate-200 bg-white p-4 shadow-sm transition hover:border-indigo-300 hover:shadow-md"
    >
      <div class="text-sm font-semibold text-slate-900">商户管理</div>
      <p class="mt-2 text-sm text-slate-600">
        为商户勾选<strong>可用支付产品</strong>。配置至少一条后进入严格白名单；未配置时平台对该商户为开放模式（见文档）。
      </p>
      <div class="mt-3 text-sm font-medium text-indigo-600 group-hover:underline">去配置 →</div>
    </RouterLink>

    <RouterLink
      to="/channels"
      class="group rounded-xl border border-slate-200 bg-white p-4 shadow-sm transition hover:border-indigo-300 hover:shadow-md"
    >
      <div class="text-sm font-semibold text-slate-900">通道管理</div>
      <p class="mt-2 text-sm text-slate-600">
        设置通道<strong>启用 / 熔断</strong>、单笔限额等。熔断通道在路由时会被跳过。
      </p>
      <div class="mt-3 text-sm font-medium text-indigo-600 group-hover:underline">去配置 →</div>
    </RouterLink>
  </div>
</template>
