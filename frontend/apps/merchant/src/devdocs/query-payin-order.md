# 查询代收订单状态

接口：`GET /v1/payin/query`

## 请求参数（Query）

- `merchant_id`：商户号（必填）
- `order_no`：平台订单号（与 `merchant_order_no` 二选一）
- `merchant_order_no`：商户订单号（与 `order_no` 二选一）
- `sign`：签名（必填）

## 成功返回

```json
{
  "order": {
    "order_no": "PO202603231234560001",
    "merchant_id": "m_demo",
    "merchant_order_no": "MO-20260323-0001",
    "amount": 1000,
    "currency": "CNY",
    "status": 1,
    "channel_id": 1,
    "notify_url": "https://merchant.example.com/notify",
    "upstream_trade_no": "UP-20260323-0001"
  }
}
```

## 状态值

- `0`：待处理
- `1`：成功
- `2`：失败
- `3`：关闭
