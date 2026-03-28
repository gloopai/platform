# 印度上游 PSP 对接说明

本目录描述**一家**印度四方/聚合支付上游的 HTTP API（商户侧调上游）。平台内部实现上游通道时，应以本文为契约；**不要求**与 pay-platform 对商户开放的 OpenAPI 路径或字段名一致。

---

## 1. 通用约定

- **Base path**：`/exposed/v1`（具体域名由上游提供）。
- **请求头**：`Content-Type: application/json`。
- **公共参数**（除回调外多数接口需要）：

| 参数 | 必填 | 说明 |
|------|------|------|
| `appId` | 是 | 应用 ID，开户后提供 |
| `timestamp` | 是 | Unix 毫秒时间戳（13 位），与上游服务器时间误差不超过 5 分钟 |
| `sign` | 是 | 请求签名，规则见上游《签名规则》文档（原文未附于本仓库） |

- **响应信封**：

| 字段 | 类型 | 说明 |
|------|------|------|
| `code` | int | `1` 表示成功，其他为失败 |
| `msg` | string | `OK` 表示成功，其他为错误说明 |
| `data` | object | 业务数据 |

---

## 2. 代收（Payin）

### 2.1 创建代收订单

- **路径**：`POST /exposed/v1/order/payment`

| 参数 | 必填 | 类型 | 说明 |
|------|------|------|------|
| `orderNo` | 是 | string | 商户订单号，商户内唯一 |
| `amount` | 是 | string | 金额，单位分，无小数 |
| `name` | 是 | string | 付款人姓名，建议真实 |
| `phone` | 是 | string | 付款人手机 |
| `email` | 是 | string | 付款人邮箱 |
| `userIP` | 否 | string | 用户 IP |
| `notifyUrl` | 否 | string | 订单状态变更时的异步通知地址 |

**成功时 `data`：**

| 字段 | 说明 |
|------|------|
| `sysOrderNo` | 平台订单号 |
| `payUrl` | 收银台地址 |

### 2.2 查询代收订单

- **路径**：`POST /exposed/v1/query/payment`

| 参数 | 必填 | 说明 |
|------|------|------|
| `orderNo` | 是 | 商户订单号 |

**成功时 `data` 含**：`appId`、`orderNo`、`sysOrderNo`、`amount`（分，string）、`status`、`referenceNo`、`failReason` 等。

**`status`**：`1` 处理中，`2` 成功，`3` 失败。

### 2.3 代收补单

- **路径**：`POST /exposed/v1/makeup`

| 参数 | 必填 | 说明 |
|------|------|------|
| `orderNo` | 是 | 商户订单号 |
| `referenceNo` | 是 | UTR |

若调用失败，可将补单截图与订单号发到 TG 商户群，由机器人协助补单追踪。

### 2.4 代收异步回调（上游 → 商户）

- **URL**：创建代收订单时传入的 `notifyUrl`。
- **方法**：`POST`，`Content-Type: application/json`。

| 参数 | 必填 | 说明 |
|------|------|------|
| `timestamp` | 是 | Unix 毫秒时间戳 |
| `sign` | 是 | 签名 |
| `orderNo` | 是 | 商户订单号 |
| `sysOrderNo` | 是 | 平台订单号 |
| `status` | 是 | `1` 处理中，`2` 成功，`3` 失败 |
| `amount` | 是 | 实收金额，单位分（string） |

**响应**：HTTP 200，body 为纯文本，**仅** `SUCCESS` 或 `FAIL`（忽略大小写）。非 200 或非上述字符串会触发重试（共 6 次，间隔递增）。可能重复回调，须幂等；重复处理成功时仍应返回 `SUCCESS`。

---

## 3. 代付（Payout）

### 3.1 创建代付订单

- **路径**：`POST /exposed/v1/order/payout`

**资损防范**：代付下单若失败，须先调**查询代付订单**确认实际状态；若查询不到（须处理超时、反序列化失败）或状态为失败，再换通道重试。

| 参数 | 必填 | 说明 |
|------|------|------|
| `orderNo` | 是 | 商户订单号，商户内唯一 |
| `wayCode` | 是 | `1` 银行卡，`2` UPI；文档写明当前仅支持 `1` |
| `amount` | 是 | 金额，单位分（string） |
| `bankName` | 是 | 开户银行名称，无则固定 `IndiaBank` |
| `bankCode` | 是 | IFSC（四位大写 + `0` + 六位数字） |
| `accountNo` | 是 | `wayCode=1` 为银行卡号，`wayCode=2` 为 UPI 账号 |
| `name` / `phone` / `email` | 是 | 收款人信息，建议真实 |
| `notifyUrl` | 否 | 状态变更异步通知 |

**成功时 `data.sysOrderNo`**：平台订单号。

### 3.2 查询代付订单

原文档未提供，**待上游补充**（路径、参数、状态枚举应对齐代收查询风格）。

### 3.3 代付异步回调（上游 → 商户）

- **URL**：创建代付订单时的 `notifyUrl`。
- **方法**：`POST`，`Content-Type: application/json`。

| 参数 | 必填 | 说明 |
|------|------|------|
| `orderNo` | 是 | 商户订单号 |
| `sysOrderNo` | 是 | 平台订单号 |
| `status` | 是 | `1` 处理中，`2` 成功，`3` 失败 |
| `amount` | 是 | 交易金额，单位分（string） |
| `referenceNo` | 是 | UTR |
| `timestamp` | 是 | Unix 毫秒时间戳 |
| `sign` | 是 | 签名 |

**响应与重试、幂等**：同「代收异步回调」（返回 `SUCCESS` / `FAIL` 纯文本，6 次重试等）。

---

## 4. 查询商户余额

- **路径**：`POST /exposed/v1/query/balance`

仅需公共参数 `appId`、`timestamp`、`sign`。

**成功时 `data`：**

| 字段 | 说明 |
|------|------|
| `availableBalance` | 可用于代付的金额（分，string） |
| `unsettledAmount` | 待结算（代收成功但尚不可代付）（分） |
| `frozenAmount` | 冻结（代付处理中）（分） |

---

## 5. 上游通道对接模块（实现指引）

**代码位置（已实现类型与接口骨架）**：[`backend/channeldriver`](../../backend/channeldriver)（独立模块 `github.com/gloopai/pay/channeldriver`）。

以下与 **本文档 API** 对齐；由具体通道实现，从 `channels` 表映射 `appId`、密钥、网关 Base URL 等到 [`ChannelConfig`](../../backend/channeldriver/base/config.go)。

**代收侧 [`PayinChannel`](../../backend/channeldriver/base/payin.go)**

- `CreatePayment` → 对应 `POST .../order/payment`
- `QueryPayment` → `POST .../query/payment`
- `Makeup` → `POST .../makeup`（可 `ErrUnsupported`）
- `VerifyPayinNotify` → 验签并解析回调
- `PayinNotifyResponse` → `SUCCESS` / `FAIL` 等响应体

**代付侧 [`PayoutChannel`](../../backend/channeldriver/base/payout.go)**

- `CreatePayout` → `POST .../order/payout`
- `QueryPayout` → **待上游文档补齐后实现**
- `VerifyPayoutNotify` / `PayoutNotifyResponse`

**余额 [`BalanceChannel`](../../backend/channeldriver/base/balance.go)**

- `QueryBalance` → `POST .../query/balance`

**多实现注册 [`Registry`](../../backend/channeldriver/base/registry.go)**：按 `driver_key` 注册；**通知分发**见 [`base/dispatch.go`](../../backend/channeldriver/base/dispatch.go) 的 `HandlePayinNotify` / `HandlePayoutNotify`（需传入 `PayinNotifyRoute` / `PayoutNotifyRoute` 从路径或查库解析 `ChannelConfig`）。根包 [`channeldriver`](../../backend/channeldriver) 重新导出 `base` 中的符号，便于单路径 import。

**说明**：平台内部订单号、路由、清结算仍走现有 trade/core/gateway；本包只负责 **与上游的 HTTP、签名、字段映射**，不复制商户 OpenAPI 形态。

**本地/单测模拟 PSP**：[`backend/channeldriver/mockpsp`](../../backend/channeldriver/mockpsp) 提供内存 `Driver`（`mock_psp`）、`RegisterAll`、`BuildPayinNotifyBody` / `BuildPayoutNotifyBody`、以及可选的 `StartChannelHTTPServer`（`httptest` + `/exposed/v1/...`），便于联调 gateway 与 `channeldriver.HandlePayinNotify`。

各服务在 `go.mod` 中增加：

```text
replace github.com/gloopai/pay/channeldriver => ../channeldriver
require github.com/gloopai/pay/channeldriver v0.0.0-00010101000000-000000000000
```

（路径按服务相对 `backend/channeldriver` 调整；接入后执行 `go mod tidy`。）

### 5.1 多上游、多路通知（平台侧总原则）

本文档只描述 **一家** PSP 的 API。实际会接入 **多家上游**，且每家回调路径、签名、body、响应体（如纯文本 `SUCCESS`）都可能不同。实现上建议：

- **一类协议一个实现**：按 `driver_key` 全局注册（例如 `psp_india_a`），**多行 `channels` 配置**可共用同一实现、不同 `appId`/密钥。
- **入站通知要可路由**：通过「URL 路径带 `driver_key` / `channel_id`」或「先根据回调内容查单得到 `channel_id`」等方式，把请求交给 **对应的实现 + 该行的 ChannelConfig**，再验签、更新订单；避免所有上游共用一个无法扩展的回调 handler。
- **代收通知与代付通知**分开处理（不同方法或不同路径），避免串单。
- **幂等**：上游会重复回调，业务更新须按订单维度幂等；日志建议带 `channel_id`、`driver_key`、请求 ID。

### 5.2 平台侧已接入路径（mock_psp）

- **trade** [`PrepareTerminalPay`](../../backend/services/trade/internal/logic/prepare_terminal_pay.go)：当 `channels.payin_type` 与已注册 `driver_key` 一致（如 `mock_psp`），且配置了 **`Upstream.CheckoutNotifyBaseURL`**（指向 **gateway OpenAPIServer** 基址，与签名开放接口同端口，默认 `:8090`，无尾斜杠）时，会调用 `CreatePayment`，并把上游异步地址设为  
  `{CheckoutNotifyBaseURL}/v1/callback/upstream/payin?channel_id={id}&order_no={平台 order_no}`。  
  本地示例见 [`backend/services/trade/etc/trade.yaml`](../../backend/services/trade/etc/trade.yaml)。
- **gateway OpenAPIServer**（与 `/v1/payin` 等验签接口同进程、同端口）：`POST /v1/callback/upstream/payin` 由 [`UpstreamPayinNotify`](../../backend/services/gateway/internal/logic/checkout/upstream_payin.go) 处理，验签后与 [`upstreamNotifyCore`](../../backend/services/gateway/internal/logic/checkout/upstream_payin.go) 共用入账逻辑（与旧版 MD5 回调入口共用同一套入账）；响应体为纯文本 `SUCCESS` / `FAIL`（与上游文档 §2.4 一致）。回调路由**不**走商户签名中间件。
- 未配置基址或驱动未命中时，仍走 **`gateway_url` / `mock://` 二维码** 等旧行为，便于无上游环境演示。

---

## 6. 修订记录

| 日期 | 说明 |
|------|------|
| 2026-03-28 | 合并原 `docs/in` 下多篇摘录为单页；修正代收回调中 notifyUrl 描述笔误；代收补单标题去重；代付查询标为待补充 |
| 2026-03-28 | §5.1 多上游与多路通知原则 |
| 2026-03-28 | §5 指向 `backend/channeldriver` 已实现骨架 |
| 2026-03-28 | `channeldriver` 独立为 `github.com/gloopai/pay/channeldriver` 模块 |
| 2026-03-28 | §5.2 平台 trade/gateway 与 `mock_psp` 回调路径说明 |
| 2026-03-28 | 契约迁入 `channeldriver/base`，根包 re-export；§5 链接更新 |
