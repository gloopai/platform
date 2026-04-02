# gloopai/platform

**管理端平台脚手架**（自 `pay-platform` 的 `scaffold/platform-admin` 拆出，独立仓库）。

- **本地路径**：建议放在 **`gloopai/platform-admin`**（与 `gloopai/pay/pay-platform` 同级；从 `pay-platform` 根目录看为 **`../../platform-admin`**）。
- **进程**：`gateway`（Admin HTTP）+ `service-hub`（gRPC）+ `job-worker`（定时任务）。
- **网关**：登录、RBAC、后台用户、展示配置、运维、定时任务与操作日志等；`ServiceContext` 主要连接 **service-hub** 与 **Redis（登录限流）**。
- **前端**：管理端（`frontend/`），npm 包 **`@platform/admin`**。
- **Go 模块路径**：**`github.com/gloopai/platform/...`**（`common`、`gateway`、`service-hub`、`job-worker`）。

## 本地运行

1. **MySQL**：默认库名 **`platform`**（见 `backend/deploy/init_demo.sh`）。执行 `bash backend/deploy/init_demo.sh`；演示账号 **`admin` / `admin123`**。各服务 YAML 中 `Mysql.DataSource` 需与本机一致。
2. **依赖**：**Consul**、**Redis**（与 gateway 限流配置一致）。`service-hub` 中 `Nsq.NsqdTCPAddr` 可为空。
3. **启动**：`bash dev-up.sh`（拉起 **service-hub**、**job-worker**、**gateway**、可选前端）。
4. **前端**：`cd frontend && npm install && npm run dev`。

- 网关：`http://127.0.0.1:8080`
- 管理端：`http://127.0.0.1:5176`
- service-hub gRPC：`127.0.0.1:8094`（Consul：`platform.rpc.service-hub`）

## 与业务仓库同步

本仓库的 Go 模块为 **`github.com/gloopai/platform/...`**，前端为 **`@platform/*`**。合并进 **pay** 或 **ec** 后，必须把前缀改回各产品自己的（下面有脚本，避免手改遗漏）。

### 通用步骤（两个产品都一样）

1. 在 **pay-platform** 或 **ec-platform** 里加一次 upstream（若已加可跳过）：

   ```bash
   git remote add platform https://github.com/gloopai/platform.git
   ```

2. 取回并合并（在你当前开发分支上）：

   ```bash
   git fetch platform
   git merge platform/main
   ```

   解决冲突、提交。

3. 按目标产品执行 **改写 import / 包名 / 部分 YAML**（见下节），再在各 `backend/**/` 里 **`go mod tidy`**，前端目录 **`npm install`**，最后 **`go build`** / 前端构建自测。

### pay-platform

合并完成后在 **pay-platform 根目录**执行：

```bash
./scripts/platform-admin-repo.sh rewrite-imports-pay
```

会把 `gloopai/platform` → `gloopai/pay`、`@platform/` → `@pay/`、`platform.notify.portal` → `pay.notify.portal`，并把 Consul 里的 **`platform.rpc.*`** 改成与现有支付栈一致的 **`payment.rpc.*`**。

说明与首次推送用法见 **`scripts/platform-admin-repo.sh`** 文件头注释。

### ec-platform

合并完成后在 **ec-platform 根目录**执行：

```bash
bash scripts/sync-from-platform.sh
```

会把 `gloopai/platform` → `gloopai/ec`、`@platform/` → `@ec/`、**`platform.notify.portal`** → **`ec.notify.portal`**。Consul 服务名在本仓库与脚手架中均为 **`platform.rpc.*`**，一般无需再改。

（首次使用可先 `chmod +x scripts/sync-from-platform.sh`。）
