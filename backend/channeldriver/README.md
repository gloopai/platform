# channeldriver

PSP / 支付通道对接模块：`driver_key` 对应一套协议实现，与 `channels` 表中的配置行组合使用。业务订单、清结算仍在 trade/core/gateway；本模块只负责 **与通道(PSP)的 HTTP、签名、字段映射**。

## 目录结构

| 路径 | 说明 |
|------|------|
| [`base/`](base/) | **契约与基础设施**：`ChannelConfig`、错误码、`PayinChannel` / `PayoutChannel` / `BalanceChannel`、`Registry`、回调辅助 `HandlePayinNotify` / `HandlePayoutNotify`、HTTP 写回 `WriteChannelNotify` |
| 根包 `channeldriver` | 对 `base` 的 **重新导出**（`reexport.go`），服务可继续只 `import "github.com/gloopai/pay/channeldriver"` |
| [`mockpsp/`](mockpsp/) | 内存模拟 PSP，用于联调与单测 |

## 如何对接新通道

### 1. 约定 `driver_key`

- 与库表 **`channels.payin_type`**（及路由选用的通道行）一致，例如 `mock_psp`、`psp_india_a`。
- 同一 `driver_key` 可对应多行 `channels`（不同 `app_id` / 密钥），运行时由 **`ChannelConfig`** 区分。

### 2. 实现接口（建议新建子包）

在 `backend/channeldriver/<yourpsp>/` 下实现：

- **`PayinChannel`**（[`base/payin.go`](base/payin.go)）：`CreatePayment`、`QueryPayment`、`Makeup`（可选 `ErrUnsupported`）、`VerifyPayinNotify`、`PayinNotifyResponse`。
- **`PayoutChannel`**（[`base/payout.go`](base/payout.go)）：代付创建/查询/回调验签（若该通道暂不接代付，可先不注册）。
- **`BalanceChannel`**（[`base/balance.go`](base/balance.go)）：可选。

实现须 **并发安全**；每次调用带上该通道行的 **`ChannelConfig`**（网关从 DB 组装，见下）。

参考 **`mockpsp`**：[`mockpsp/driver.go`](mockpsp/driver.go)、[`mockpsp/register.go`](mockpsp/register.go)。

### 3. 注册到 `Registry`

进程启动时（与 **trade**、**gateway** 各自一致）：

```go
reg := channeldriver.NewRegistry()
_ = yourpsp.RegisterAll(reg, yourpsp.New(yourpsp.DefaultDriverKey))
```

或分别 `RegisterPayin` / `RegisterPayout` / `RegisterBalance`。见 [`base/registry.go`](base/registry.go)。

### 4. 平台侧接线（本仓库现状）

| 环节 | 说明 |
|------|------|
| **trade** | `PrepareTerminalPay`：若 `payin_type` 命中已注册驱动且配置了 `Upstream.CheckoutNotifyBaseURL`，则调 `CreatePayment` 并填通道异步 `notifyUrl`。 |
| **gateway OpenAPIServer**（`/v1/callback/*` 与开放接口同端口 `:8090`） | 通道异步回调入口（如 `POST /v1/callback/upstream/payin`）内：`ConfigFromDriverKey` + `VerifyPayinNotify`，成功后走平台入账逻辑；响应体通常用 `WriteChannelNotify`。 |
| **数据库** | 插入/迁移 `channels` 行：`payin_type` = `driver_key`，`channel_merchant_no`、`sign_secret`、`gateway_url`（若需）等与 PSP 文档一致。 |

更细的 HTTP 字段约定见 [`docs/in/README.md`](../../docs/in/README.md)（印度示例 PSP）。

### 5. 依赖与替换路径

- `go.mod` 中：

  ```text
  require github.com/gloopai/pay/channeldriver v0.0.0-00010101000000-000000000000
  replace github.com/gloopai/pay/channeldriver => ../channeldriver
  ```

- 新实现可 **只依赖** `github.com/gloopai/pay/channeldriver/base`，避免与根包循环引用；平台代码仍多用根包 re-export。

## 与 `base` 的关系

- **新增通道代码**：优先只 import **`channeldriver/base`** 定义结构体与接口实现。
- **服务装配**：继续 **`channeldriver.NewRegistry()`** 即可，类型与 `base` 中一致（别名 + 包装函数）。
