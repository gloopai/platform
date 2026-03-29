# pay-platform（`scaffold/platform-admin`）

本分支为 **管理端平台脚手架**：仅保留 **Admin 网关路由**（登录、RBAC、后台用户、展示配置）、**管理端前端**（权限与安全、系统与运维），已删除商户台/收银台/开放 API 等前端应用、支付业务网关 handler 与演示数据。

## 本地运行

1. **MySQL**：创建库（默认名 `pay`），执行 `bash backend/deploy/init_demo.sh`（按脚本内变量配置账号；默认演示账号 **`admin` / `admin123`**）。
2. **依赖**：Consul、Redis、NSQ — 与 `dev-up.sh` 一致（网关 `Ready` 与 `svc` 依赖下游 gRPC / Redis）。
3. **启动**：`bash dev-up.sh`（网关 Admin `http://127.0.0.1:8080`，管理端 `http://127.0.0.1:5176`）。
4. **前端**：`cd frontend && npm install && npm run dev`。

说明与裁剪清单见 [`docs/项目脚手架.md`](docs/项目脚手架.md)；管理端 UI 约定见 [`docs/管理端前端开发规范.md`](docs/管理端前端开发规范.md)。

若需完整聚合支付能力与多前端，请使用 **`0.1` / `main`** 等业务分支。
