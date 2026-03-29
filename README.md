# pay-platform（`scaffold/platform-admin`）

本分支为 **管理端平台脚手架**：

- **进程**：**`gateway`（HTTP）** + **`service-hub`（gRPC，直连 MySQL）**；已移除 **core / trade / notice-consumer** 等支付域服务。
- **网关**：仅 **Admin** 路由（登录、RBAC、后台用户、展示配置）；`ServiceContext` 只连接 **service-hub** 与 **Redis（登录限流）**，不再依赖 NSQ、代收付 gRPC。
- **前端**：仅管理端（权限与安全、系统与运维）。

## 本地运行

1. **MySQL**：创建库（默认名 `pay`），执行 `bash backend/deploy/init_demo.sh`（按脚本内变量配置账号；默认演示账号 **`admin` / `admin123`**）。**`backend/services/service-hub/etc/service-hub.yaml`** 中的 `Mysql.DataSource` 需与本机一致。
2. **依赖**：**Consul**、**Redis**（与 `gateway-api.yaml` / 限流配置一致）。**无需 NSQ**（`service-hub` 中 `Nsq.NsqdTCPAddr` 可为空）。
3. **启动**：`bash dev-up.sh`（仅拉起 **service-hub**、**gateway**、可选 **fe-admin**）。
4. **前端**：`cd frontend && npm install && npm run dev`。

- 网关 Admin：`http://127.0.0.1:8080`
- 管理端：`http://127.0.0.1:5176`
- service-hub gRPC：`127.0.0.1:8094`（Consul：`payment.rpc.service-hub`）

说明见 [`docs/项目脚手架.md`](docs/项目脚手架.md)、[`docs/管理端前端开发规范.md`](docs/管理端前端开发规范.md)。

完整聚合支付能力与多前端请使用 **`0.1` / `main`** 等业务分支。
