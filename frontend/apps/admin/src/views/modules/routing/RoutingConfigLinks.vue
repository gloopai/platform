<template>
  <div class="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
    <RouterLink
      to="/merchant-products"
      class="group rounded-xl border border-slate-200 bg-white p-4 shadow-sm transition hover:border-indigo-300 hover:shadow-md"
    >
      <div class="text-sm font-semibold text-slate-900">代收产品与通道</div>
      <p class="mt-2 text-sm text-slate-600">
        为每个对外<strong>代收</strong>产品绑定上游通道（须开启代收能力）及权重；费率在通道与商户侧配置。
      </p>
      <div class="mt-3 text-sm font-medium text-indigo-600 group-hover:underline">去配置 →</div>
    </RouterLink>

    <RouterLink
      to="/merchant-payout-products"
      class="group rounded-xl border border-slate-200 bg-white p-4 shadow-sm transition hover:border-indigo-300 hover:shadow-md"
    >
      <div class="text-sm font-semibold text-slate-900">代付产品与通道</div>
      <p class="mt-2 text-sm text-slate-600">
        <strong>代付</strong>对外产品与通道独立配置；通道需开启代付能力。与代收路由互不影响。
      </p>
      <div class="mt-3 text-sm font-medium text-indigo-600 group-hover:underline">去配置 →</div>
    </RouterLink>

    <RouterLink
      to="/merchants"
      class="group rounded-xl border border-slate-200 bg-white p-4 shadow-sm transition hover:border-indigo-300 hover:shadow-md"
    >
      <div class="text-sm font-semibold text-slate-900">商户管理</div>
      <p class="mt-2 text-sm text-slate-600">
        为商户分别勾选<strong>代收 / 代付</strong>可用产品并设置默认费率；白名单与开放模式规则见文档。
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
