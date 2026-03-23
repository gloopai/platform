# 接口签名说明

签名算法：`MD5`

## 规则

1. 取参与签名的参数（不含 `sign`）。
2. 参数名转小写后按字典序排序。
3. 拼接为 `k1=v1&k2=v2...`（空值可忽略）。
4. 末尾追加 `&key=app_secret`。
5. 对最终字符串做 MD5，得到 `sign`。

## 示例

参数：

- `merchant_id=m_demo`
- `merchant_order_no=MO-20260323-0001`
- `amount=1000`
- `currency=CNY`
- `pay_type=mock`

secret：

- `demo_secret`

待签名串：

```text
amount=1000&currency=CNY&merchant_id=m_demo&merchant_order_no=MO-20260323-0001&pay_type=mock&key=demo_secret
```

## 注意事项

- 参数名大小写不一致会导致验签失败，建议统一小写。
- 不要把 `sign` 自己也拼进去。
- 时间戳/随机串如有参与签名，请保证请求和签名内容一致。

