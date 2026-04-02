# gloopai/platform

**管理端平台脚手架**（自 `pay-platform` 的 `scaffold/platform-admin` 拆出，独立仓库）。

- **进程**：`gateway`（Admin HTTP）+ `service-hub`（gRPC）+ `job-worker`（定时任务）。
- **网关**：登录、RBAC、后台用户、展示配置、运维、定时任务与操作日志等；`ServiceContext` 主要连接 **service-hub** 与 **Redis（登录限流）**。
- **前端**：管理端（`frontend/`）。
- **Go 模块路径**仍为 `github.com/gloopai/pay/...`，便于合并进 **pay-platform** / **ec-platform** 时无需改 import。

## 本地运行

1. **MySQL**：默认库名 **`platform`**（见 `backend/deploy/init_demo.sh`）。执行 `bash backend/deploy/init_demo.sh`；演示账号 **`admin` / `admin123`**。各服务 YAML 中 `Mysql.DataSource` 需与本机一致。
2. **依赖**：**Consul**、**Redis**（与 gateway 限流配置一致）。`service-hub` 中 `Nsq.NsqdTCPAddr` 可为空。
3. **启动**：`bash dev-up.sh`（拉起 **service-hub**、**job-worker**、**gateway**、可选前端）。
4. **前端**：`cd frontend && npm install && npm run dev`。

- 网关：`http://127.0.0.1:8080`
- 管理端：`http://127.0.0.1:5176`
- service-hub gRPC：`127.0.0.1:8094`（Consul：`platform.rpc.service-hub`）

## 与业务仓库同步

在 **pay-platform** / **ec-platform** 中添加 remote 后合并：

```bash
git remote add platform https://github.com/gloopai/platform.git
git fetch platform
git merge platform/main
```

详见业务仓库内 `scripts/platform-admin-repo.sh` 说明。
