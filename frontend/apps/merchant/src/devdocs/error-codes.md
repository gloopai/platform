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
- `INVALID_ARGUMENT`：参数不合法
- `INVALID_SIGN`：签名错误
- `UNAUTHENTICATED`：认证失败
- `PAY_PRODUCT_NOT_ENABLED`：产品未开通
- `NO_AVAILABLE_CHANNEL`：无可用通道
- `INSUFFICIENT_PAYOUT_BALANCE`：代付余额不足
- `PAYOUT_ORDER_ALREADY_EXISTS_PENDING`：代付单号已存在且待处理
- `ORDER_NOT_FOUND`：订单不存在
- `INTERNAL_ERROR`：服务内部错误
- `UNAVAILABLE`：依赖服务不可用

## 代付相关建议

- 收到 `INSUFFICIENT_PAYOUT_BALANCE`：先充值/划转代付余额，再重试下单。
- 收到 `PAYOUT_ORDER_ALREADY_EXISTS_PENDING`：不要重复提交同单号，换 `merchant_order_no`。

