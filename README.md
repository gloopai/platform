# pay-platform

四方聚合支付相关单体/微服务与前端（管理台、商户台、收银台等）。

> **分支 `scaffold/platform-admin`**：管理端脚手架线。网关进程 **仅监听 Admin HTTP**（`:8080`）；`dev-up.sh` 只拉起 **管理端前端**；`frontend/package.json` 的 workspaces 仅 **`apps/admin`**。说明见 [`docs/项目脚手架.md`](docs/项目脚手架.md)。合并回主线前请评估是否恢复多路 HTTP 与多前端。

## 本地把主路径跑通

1. 初始化数据库：`bash backend/deploy/init_demo.sh`（按脚本内说明配置 MySQL 账号）  
2. 启动：`bash dev-up.sh`（网关 Admin `http://127.0.0.1:8080`，管理端 `http://127.0.0.1:5176`）  
3. **脚手架分支**：登录管理台做配置 / RBAC 等即可；完整支付链路见 [`docs/端到端联调一遍.md`](docs/端到端联调一遍.md)（需使用含商户 / OpenAPI / 收银台网关端口的分支）。

更多：[`docs/开发计划.md`](docs/开发计划.md)、[`docs/MVP与后续迭代.md`](docs/MVP与后续迭代.md)、开放接口错误格式 [`docs/开放API错误码.md`](docs/开放API错误码.md)。
