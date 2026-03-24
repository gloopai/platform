# 查询代付订单状态

接口：`GET /v1/payout/query`

## 请求参数（Query）

- `merchant_id`：商户号（必填）
- `order_no`：平台代付订单号（与 `merchant_order_no` 二选一）
- `merchant_order_no`：商户代付订单号（与 `order_no` 二选一）
- `timestamp`：Unix 时间戳（秒，必填，参与签名）
- `nonce`：随机串（必填，参与签名，建议每次请求唯一）
- `sign`：签名（必填）

## 成功返回

返回结构与代收查单一致，核心看 `order.status`。

## 状态值

- `0`：待处理
- `1`：成功
- `2`：失败
- `3`：关闭

## 幂等与补偿语义

- 下单扣款阶段失败后，订单会落为 `status=2`（失败），不会长期停留 `pending`。
- 同一 `merchant_order_no`：
  - 若订单仍 `pending`，重试返回 `422 + PAYOUT_ORDER_ALREADY_EXISTS_PENDING`
  - 若订单已 `failed/success`，重试返回已存在订单（HTTP `200`，按 `status` 分支）

## 错误补充

- `REPLAY_REQUEST`：重复使用相同 `merchant_id + nonce + timestamp`
- `TOO_MANY_REQUESTS`：触发限流

