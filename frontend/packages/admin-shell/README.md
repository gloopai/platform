# `@platform/admin-shell`

本包位于 **platform** 仓库，供 **pay-platform**、**ec-platform** 等通过 **本地路径** 引用，无需发布到 npm registry。

## 在 pay 中引用

`pay-platform/frontend/apps/admin/package.json`：

```json
"@platform/admin-shell": "file:../../../../../platform-admin/frontend/packages/admin-shell"
```

（路径以 `.../pay/pay-platform/frontend/apps/admin` 与 `.../gloopai/platform-admin` 同级为前提。）

然后在业务里：

```ts
import { PLATFORM_ADMIN_SHELL_VERSION } from '@platform/admin-shell'
```

## 后续

将登录页、AdminLayout、路由守卫、通用 API 封装等逐步迁入本包；业务仓库只保留领域页面与菜单配置。
