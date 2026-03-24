# 创建代付订单

接口：`POST /v1/payout/order`

## 请求参数（JSON）

- `merchant_id`：商户号（必填）
- `merchant_order_no`：商户订单号（必填，建议唯一）
- `amount`：代付金额，单位分（必填，>0）
- `currency`：币种（可选，默认 `CNY`）
- `payout_product_code`：代付产品编码（推荐）
- `pay_type`：兼容字段，未传 `payout_product_code` 时可用
- `notify_url`：异步通知地址（可选）
- `timestamp`：Unix 时间戳（秒，必填，参与签名）
- `nonce`：随机串（必填，参与签名，建议每次请求唯一）
- `sign`：签名（必填）

## 关键校验

- 创建代付订单时会进行可用余额校验。
- 扣减金额 = `amount + fee_amount`（手续费快照来自商户配置）。
- 防重放校验：
  - `timestamp` 需在服务端允许时间窗内（默认 ±300 秒）；
  - 相同 `merchant_id + nonce + timestamp` 仅允许一次通过签名校验。
- 若余额不足，返回：
  - HTTP `422`
  - `code = INSUFFICIENT_AVAILABLE_BALANCE`
- 首次请求若在扣款阶段失败（如余额不足、资金服务异常）会创建代付单并置为失败态（`status=2`），避免单号长期停留 `pending`。

## 幂等说明（轻量）

- 同一 `merchant_order_no` 重试不会重复扣款。
- 若该单号对应订单仍待处理，返回：
  - HTTP `422`
  - `code = PAYOUT_ORDER_ALREADY_EXISTS_PENDING`
- 若该单号对应订单已失败/已成功，重试返回已存在订单（HTTP `200`），请按返回 `status` 判断后续流程。

