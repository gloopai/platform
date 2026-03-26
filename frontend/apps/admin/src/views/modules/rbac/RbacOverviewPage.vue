<template>
  <div class="grid gap-4 lg:grid-cols-2">
    <div class="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
      <div class="text-sm font-semibold text-slate-900">配置顺序（建议）</div>
      <ol class="mt-3 list-decimal space-y-2 pl-5 text-sm text-slate-700">
        <li><strong>权限点</strong>：定义业务动作（如 <span class="font-mono text-xs">admin.merchants.read</span>）。</li>
        <li><strong>接口规则</strong>：HTTP 方法 + 路径模式 → 绑定到某个权限点。</li>
        <li><strong>角色与授权</strong>：为角色勾选菜单 + 权限点。</li>
        <li>用户登录后仅能看到被授权的菜单；接口由网关按规则校验。</li>
      </ol>
    </div>
    <div class="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
      <div class="text-sm font-semibold text-slate-900">快速入口</div>
      <div class="mt-3 flex flex-col gap-2">
        <RouterLink
          to="/rbac/roles"
          class="rounded-lg border border-slate-200 bg-slate-50 px-4 py-3 text-sm font-medium text-slate-800 transition hover:border-indigo-300 hover:bg-indigo-50"
        >
          角色与授权 → 菜单 + 权限点
        </RouterLink>
        <RouterLink
          to="/rbac/permissions"
          class="rounded-lg border border-slate-200 bg-slate-50 px-4 py-3 text-sm font-medium text-slate-800 transition hover:border-indigo-300 hover:bg-indigo-50"
        >
          权限点目录 → 新建 / 查看
        </RouterLink>
        <RouterLink
          to="/rbac/api-rules"
          class="rounded-lg border border-slate-200 bg-slate-50 px-4 py-3 text-sm font-medium text-slate-800 transition hover:border-indigo-300 hover:bg-indigo-50"
        >
          接口规则 → 路径映射
        </RouterLink>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { RouterLink } from 'vue-router'
</script>
