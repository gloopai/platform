# 开放 API 错误响应（网关）

适用于 **商户签名** 的 `POST /v1/pay/order`、`GET /v1/pay/query`，以及 **无签名** 的收银台 `GET /v1/terminal/order`、`POST /v1/terminal/pay` 等由网关 `openapi` 包统一写出 JSON 错误的接口。

> `POST /v1/callback/notify` 为兼容上游习惯，成功与失败都返回 `{"ok": true|false}`；不走本文的 `code` 错误体。

## 回调接口返回（`/v1/callback/notify`）

该接口固定返回 HTTP 200，响应体：

```json
{
  "ok": false,
  "reason_code": "INVALID_SIGN",
  "reason": "invalid sign"
}
```

- `ok=true`：回调被接受（含幂等重放接受场景）。
- `ok=false`：回调未被接受。
- `reason_code`：机器可读，建议商户优先按此分支处理。
- `reason`：可读文本，便于排障，不保证长期稳定。

### `reason_code` 对照（MVP）

| `reason_code` | 含义 | 建议处理 |
|---------------|------|----------|
| `INVALID_NOTIFY_PARAMS` | 回调参数缺失/非法 | 检查 `order_no`/`paid_amount`/`channel_id`/`upstream_trade_no` |
| `CHANNEL_NOT_FOUND` | 通道不存在或签名密钥不可查 | 校验 `channel_id` 与通道配置 |
| `INVALID_SIGN` | 回调签名错误 | 校验签名算法与 `channel_sign_secret` |
| `ORDER_NOT_FOUND` | 平台订单不存在 | 校验 `order_no` 是否平台单号 |
| `ORDER_NOT_PENDING` | 订单不在待支付状态 | 仅待支付订单可置成功 |
| `REPLAY_PAYLOAD_MISMATCH` | 已支付但重放快照不一致 | 保证重复通知参数与首笔一致 |
| `MARK_PAID_FAILED` | 落支付状态失败 | 查看网关/trade 日志 |
| `MARK_PAID_RACE` | 并发竞争，读取最终态失败 | 短暂重试后查询订单最终态 |
| `MARK_PAID_RACE_MISMATCH` | 并发竞争且最终快照不一致 | 以平台最终订单快照为准排查上游重复回调 |
| `IDEMPOTENT_REPLAY_ACCEPTED` | 已支付且回放快照一致（成功）；会再次尝试入账（幂等）并入队商户通知 | 可视为成功；通知队列应按 `order_no` 幂等处理 |
| `IDEMPOTENT_RACE_ACCEPTED` | 并发竞争后确认同快照（成功） | 可视为成功，无需补偿 |
| `CREDIT_FAILED` | 订单已置支付成功，但入款（充值到商户 payin 余额）失败 | 根据 `reason` 重试回调；平台侧需查 settle/对账 |
| `NOTIFY_MARSHAL_FAILED` | 组装通知消息失败 | 重试回调 |
| `NOTIFY_PUBLISH_FAILED` | 消息队列发布失败 | 重试回调 |

**成功**：仍为各接口原有 JSON 结构。

**失败**：HTTP 状态码 + `Content-Type: application/json`，正文为：

```json
{
  "code": "INVALID_ARGUMENT",
  "message": "human readable message"
}
```

商户集成应优先按 **`code`** 分支；`message` 供排障与展示，**不保证**长期稳定英文/中文混排。

---

## 业务码（`code`）与典型 HTTP 状态

| `code` | HTTP | 含义 |
|--------|------|------|
| `INVALID_PARAMS` | 400 | 参数无法解析（如 JSON 非法） |
| `PAYLOAD_TOO_LARGE` | 413 | JSON 请求体超过 `OpenAPI.MaxBodyBytes`（默认 256KiB） |
| `MERCHANT_ID_REQUIRED` | 400 | 缺少 `merchant_id` |
| `SIGN_REQUIRED` | 400 | 缺少 `sign` |
| `TIMESTAMP_REQUIRED` | 400 | 缺少 `timestamp` |
| `NONCE_REQUIRED` | 400 | 缺少 `nonce` |
| `INVALID_TIMESTAMP` | 400 | `timestamp` 非法或超出允许时间窗 |
| `INVALID_ARGUMENT` | 400 | 参数不合法（如缺必填字段） |
| `TOO_MANY_REQUESTS` | 429 | 触发接口限流（登录或开放签名接口） |
| `MERCHANT_NOT_FOUND` | 401 | 商户不存在 |
| `MERCHANT_DISABLED` | 401 | 商户已停用 |
| `INVALID_SIGN` | 401 | 签名错误 |
| `UNAUTHENTICATED` | 401 | 未认证（其它 gRPC 未映射场景） |
| `IP_NOT_ALLOWED` | 403 | 客户端 IP 不在白名单 |
| `PAY_PRODUCT_NOT_ENABLED` | 403 | 支付产品未开通给该商户 |
| `FORBIDDEN` | 403 | 其它权限拒绝 |
| `ORDER_NOT_FOUND` | 404 | 订单不存在（与商户无关或无权） |
| `NOT_FOUND` | 404 | 其它 gRPC NotFound |
| `NO_AVAILABLE_CHANNEL` | 422 | 当前无可用上游通道（路由/熔断/金额区间等） |
| `INSUFFICIENT_AVAILABLE_BALANCE` | 422 | 可用余额不足（创建代付订单时校验失败） |
| `PAYOUT_ORDER_ALREADY_EXISTS_PENDING` | 422 | 同一 `merchant_order_no` 的代付单已存在且仍待处理，需更换单号重试 |
| `REPLAY_REQUEST` | 409 | 签名请求被判定为重放（相同 `merchant_id+nonce+timestamp`） |
| `FAILED_PRECONDITION` | 422 | 其它前置条件不满足（如锁定通道校验失败） |
| `INTERNAL_ERROR` | 500 | 内部错误或未单独映射的 gRPC 码 |
| `UNAVAILABLE` | 503 | 依赖服务不可用 |

---

## 与 gRPC 的映射

网关内部错误多为 **gRPC Status**（trade 等），映射规则见 `backend/services/gateway/internal/openapi/errresp.go` 中 `mapGRPC`。`NOT_FOUND` 且消息含 `order` 时，业务码为 **`ORDER_NOT_FOUND`**。

---

## 代付下单（`POST /v1/payout/order`）补充约定

- 下单会先创建代付订单，再执行可用余额扣减（扣减金额 = `amount + fee_amount`）。
- 开放签名接口统一要求携带 `timestamp`（秒级 Unix 时间戳）与 `nonce`（随机串）参与签名：
  - 时间窗默认 ±300 秒；
  - 相同 `merchant_id + nonce + timestamp` 只允许成功校验一次（重放请求会被拒绝）。
- 当可用余额不足时，返回：
  - HTTP `422`
  - `code = INSUFFICIENT_AVAILABLE_BALANCE`
- 当首次下单在扣款阶段失败（如余额不足或资金服务异常）时，平台会将该代付单标记为失败态（`status=2`），避免该单号长期停留待处理。
- 轻量幂等策略：
  - 同一 `merchant_order_no` 重试不会重复扣款；
  - 若该单号对应订单仍是待处理状态，返回：
    - HTTP `422`
    - `code = PAYOUT_ORDER_ALREADY_EXISTS_PENDING`
  - 若该单号对应订单已失败/已成功，重试返回该已存在订单（HTTP `200`，以返回体 `status` 判定）。

---

## 修订记录

| 日期 | 说明 |
|------|------|
| 2026-03-23 | 首版：统一 JSON 错误体、`code` 表；商户开发页与收银台展示错误时优先解析 `code` + `message`。 |
| 2026-03-23 | 备注：`/v1/callback/notify` 仍返回 `ok` 布尔值，不使用 `code` 错误体。 |
| 2026-03-24 | 代付扣款失败补偿：首次扣款失败后订单置失败态（`status=2`）；同 `merchant_order_no` 重试返回已存在订单，不再卡在 pending。 |
