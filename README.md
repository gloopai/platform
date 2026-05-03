# gloopai/platform

管理端平台脚手架：从 `pay-platform` 的 `scaffold/platform-admin` 拆出，独立维护。

## 概览

| 项 | 说明 |
| --- | --- |
| **进程** | `gateway`（Admin HTTP）、`service-hub`（gRPC）、`job-worker`（定时任务；`dev-up.sh` 会起两个 worker 实例模拟多节点） |
| **网关职责** | 登录、RBAC、后台用户、展示配置、运维、定时任务与操作日志等；`ServiceContext` 主要连 **service-hub** 与 **Redis**（登录限流） |
| **前端** | `frontend/`，npm 包 **`@platform/admin`** |
| **Go 模块** | **`github.com/gloopai/platform/...`**（`common`、`gateway`、`service-hub`、`job-worker`） |

**本地目录建议**：与 `gloopai/pay/pay-platform` 同级时，可命名为 **`gloopai/platform-admin`**（从 `pay-platform` 根目录看为 **`../../platform-admin`**）。

## 本地运行

### 1. 数据库

- 默认库名 **`platform`**（见 `backend/deploy/init_demo.sh`）。
- 执行：`bash backend/deploy/init_demo.sh`
- 演示账号：**`admin` / `admin123`**
- 各服务 YAML 里的 `Mysql.DataSource` 需与本机 MySQL 一致。

### 2. 依赖服务

- **Consul**、**Redis**（与 gateway 限流配置一致）。
- **`dev-up.sh`**：若本机已有 `consul` / `redis-server` 且对应端口未占用，会尝试在后台拉起开发实例；否则请自行保证 **8500**（Consul）、**6379**（Redis）可用。
- **`service-hub`**：`Nsq.NsqdTCPAddr` 可为空。

### 3. 启动后端与前端

```bash
bash dev-up.sh
```

日志目录：**`.dev-logs/`**（各进程 stdout/stderr）。

前端也可单独安装与开发：

```bash
cd frontend && npm install && npm run dev
```

### 4. 常用地址

| 服务 | 地址 |
| --- | --- |
| 网关（Admin） | `http://127.0.0.1:8080`（Admin API：`/v1/admin/*`） |
| 管理端前端 | `http://127.0.0.1:5176` |
| service-hub gRPC | `127.0.0.1:8094`（Consul：`platform.rpc.service-hub`） |

若 MySQL 未在 **3306** 监听，`dev-up.sh` 会提示，**service-hub** 可能启动失败。

---

## 与业务仓库同步

本仓库 Go 模块为 **`github.com/gloopai/platform/...`**，前端为 **`@platform/*`**。合并进 **pay** 或 **ec** 后，必须把前缀改回各产品自己的命名（见下方脚本，避免手改遗漏）。

### 通用步骤

1. 在 **pay-platform** 或 **ec-platform** 中添加 upstream（若已存在可跳过）：

   ```bash
   git remote add platform https://github.com/gloopai/platform.git
   ```

2. 在当前开发分支上取回并合并：

   ```bash
   git fetch platform
   git merge platform/main
   ```

   解决冲突后提交。

3. 按目标产品执行 **改写 import / 包名 / 部分 YAML**（见下），再在 `backend/**/` 下执行 **`go mod tidy`**，前端目录 **`npm install`**，最后 **`go build`** 与前端构建自测。

### pay-platform

在 **pay-platform 根目录**执行：

```bash
./scripts/platform-admin-repo.sh rewrite-imports-pay
```

效果概要：`gloopai/platform` → `gloopai/pay`，`@platform/` → `@pay/`，`platform.notify.portal` → `pay.notify.portal`，Consul 中 **`platform.rpc.*`** 改为与支付栈一致的 **`payment.rpc.*`**。

说明与首次推送用法见 **`scripts/platform-admin-repo.sh`** 文件头注释。

### ec-platform

在 **ec-platform 根目录**执行：

```bash
bash scripts/sync-from-platform.sh
```

效果概要：`gloopai/platform` → `gloopai/ec`，`@platform/` → `@ec/`，**`platform.notify.portal`** → **`ec.notify.portal`**。Consul 服务名在本仓库与脚手架中为 **`platform.rpc.*`**，一般无需再改。

首次使用可先：`chmod +x scripts/sync-from-platform.sh`。
