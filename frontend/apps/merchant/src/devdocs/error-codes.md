# 开放接口错误码（商户侧）

统一错误体：

```json
{
  "code": "SOME_CODE",
  "message": "human readable message"
}
```

## 常用错误码

- `INVALID_PARAMS`：请求体/参数解析失败
- `PAYLOAD_TOO_LARGE`：JSON 请求体超过服务端限制（默认约 256KB）
- `TIMESTAMP_REQUIRED`：缺少 `timestamp`
- `NONCE_REQUIRED`：缺少 `nonce`
- `INVALID_TIMESTAMP`：`timestamp` 非法或超出时间窗
- `INVALID_ARGUMENT`：参数不合法
- `INVALID_SIGN`：签名错误
- `REPLAY_REQUEST`：请求被判定为重放（同 `merchant_id + nonce + timestamp`）
- `TOO_MANY_REQUESTS`：触发接口限流
- `UNAUTHENTICATED`：认证失败
- `PAY_PRODUCT_NOT_ENABLED`：产品未开通
- `NO_AVAILABLE_CHANNEL`：无可用通道
- `INSUFFICIENT_AVAILABLE_BALANCE`：可用余额不足
- `PAYOUT_ORDER_ALREADY_EXISTS_PENDING`：代付单号已存在且待处理
- `ORDER_NOT_FOUND`：订单不存在
- `INTERNAL_ERROR`：服务内部错误
- `UNAVAILABLE`：依赖服务不可用

## 代付相关建议

- 收到 `INSUFFICIENT_AVAILABLE_BALANCE`：先充值/划转可用余额，再重试下单。
- 收到 `PAYOUT_ORDER_ALREADY_EXISTS_PENDING`：不要重复提交同单号，换 `merchant_order_no`。
- 收到 `REPLAY_REQUEST`：更换 `nonce`（必要时刷新 `timestamp`）后重试。
- 收到 `TOO_MANY_REQUESTS`：降低请求频率，等待窗口后重试。

