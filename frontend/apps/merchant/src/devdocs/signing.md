# 接口签名说明

签名算法：`MD5`

## 规则

1. 取参与签名的参数（不含 `sign`）。
2. 参数名转小写后按字典序排序。
3. 拼接为 `k1=v1&k2=v2...`（空值可忽略）。
4. 末尾追加 `&key=app_secret`。
5. 对最终字符串做 MD5，得到 `sign`。

## 防重放必填参数

所有需要商户签名的开放接口都必须携带以下字段并参与签名：

- `timestamp`：秒级 Unix 时间戳
- `nonce`：随机串（建议 16~32 位，每次请求唯一）

服务端会执行：

- 时间窗校验（默认 `timestamp` 在当前时间 `±300s` 内）；
- 重放校验（相同 `merchant_id + nonce + timestamp` 仅允许一次）。

命中重放时返回：

- HTTP `409`
- `code = REPLAY_REQUEST`

## 示例

参数：

- `merchant_id=m_demo`
- `merchant_order_no=MO-20260323-0001`
- `amount=1000`
- `currency=CNY`
- `pay_type=mock`
- `timestamp=1774368000`
- `nonce=a1b2c3d4e5f6a7b8`

secret：

- `demo_secret`

待签名串：

```text
amount=1000&currency=CNY&merchant_id=m_demo&merchant_order_no=MO-20260323-0001&nonce=a1b2c3d4e5f6a7b8&pay_type=mock&timestamp=1774368000&key=demo_secret
```

## 注意事项

- 参数名大小写不一致会导致验签失败，建议统一小写。
- 不要把 `sign` 自己也拼进去。
- `timestamp`、`nonce` 必须与请求体/Query 完全一致并参与签名。
- 同一签名请求体不要重复发送，否则会被判定为重放。

