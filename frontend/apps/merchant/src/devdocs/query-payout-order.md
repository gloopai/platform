# 查询代付订单状态

接口：`GET /v1/payout/query`

## 请求参数（Query）

- `merchant_id`：商户号（必填）
- `order_no`：平台代付订单号（与 `merchant_order_no` 二选一）
- `merchant_order_no`：商户代付订单号（与 `order_no` 二选一）
- `sign`：签名（必填）

## 成功返回

返回结构与代收查单一致，核心看 `order.status`。

## 状态值

- `0`：待处理
- `1`：成功
- `2`：失败
- `3`：关闭

