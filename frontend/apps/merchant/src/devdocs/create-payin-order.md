# 创建代收订单

接口：`POST /v1/payin/order`

## 请求参数（JSON）

- `merchant_id`：商户号（必填）
- `merchant_order_no`：商户订单号（必填，建议唯一）
- `amount`：订单金额，单位分（必填，>0）
- `currency`：币种（可选，默认 `CNY`）
- `payin_type`：支付产品编码（必填，例如 `mock`、`wechat`）
- `notify_url`：异步通知地址（可选）
- `timestamp`：Unix 时间戳（秒，必填，参与签名）
- `nonce`：随机串（必填，参与签名，建议每次请求唯一）
- `sign`：签名（必填）

## 成功返回

```json
{
  "order_no": "PO202603231234560001",
  "status": 0,
  "channel_id": 1,
  "checkout_url": "http://127.0.0.1:5174/?order_no=PO202603231234560001"
}
```

## 失败返回

统一错误体：

```json
{
  "code": "INVALID_ARGUMENT",
  "message": "human readable message"
}
```

常见错误补充：

- `REPLAY_REQUEST`：相同 `merchant_id + nonce + timestamp` 重复请求
- `TOO_MANY_REQUESTS`：触发限流
