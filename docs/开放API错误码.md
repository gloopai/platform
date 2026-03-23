# 开放 API 错误响应（网关）

适用于 **商户签名** 的 `POST /v1/pay/order`、`GET /v1/pay/query`，以及 **无签名** 的收银台 `GET /v1/terminal/order`、`POST /v1/terminal/pay`、上游回调 `POST /v1/callback/notify` 等由网关 `openapi` 包统一写出 JSON 错误的接口。

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
| `MERCHANT_ID_REQUIRED` | 400 | 缺少 `merchant_id` |
| `SIGN_REQUIRED` | 400 | 缺少 `sign` |
| `INVALID_ARGUMENT` | 400 | 参数不合法（如缺必填字段） |
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
| `FAILED_PRECONDITION` | 422 | 其它前置条件不满足（如锁定通道校验失败） |
| `INTERNAL_ERROR` | 500 | 内部错误或未单独映射的 gRPC 码 |
| `UNAVAILABLE` | 503 | 依赖服务不可用 |

---

## 与 gRPC 的映射

网关内部错误多为 **gRPC Status**（trade 等），映射规则见 `backend/services/gateway/internal/openapi/errresp.go` 中 `mapGRPC`。`NOT_FOUND` 且消息含 `order` 时，业务码为 **`ORDER_NOT_FOUND`**。

---

## 修订记录

| 日期 | 说明 |
|------|------|
| 2026-03-23 | 首版：统一 JSON 错误体、`code` 表；商户开发页与收银台展示错误时优先解析 `code` + `message`。 |
