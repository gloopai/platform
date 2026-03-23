INSERT INTO merchants (merchant_id, api_secret, status, default_collect_rate_bps, default_payout_rate_bps, ip_whitelist, balance, notify_url)
VALUES
  ('m_demo', 'demo_secret', 1, 60, 80, '127.0.0.1', 0, ''),
  ('m_rate_mix', 'demo_secret_mix', 1, 90, 120, '127.0.0.1', 0, ''),
  ('m_zero_fee', 'demo_secret_zero', 1, 0, 0, '127.0.0.1', 0, '')
ON DUPLICATE KEY UPDATE
  api_secret = VALUES(api_secret),
  status = VALUES(status),
  default_collect_rate_bps = VALUES(default_collect_rate_bps),
  default_payout_rate_bps = VALUES(default_payout_rate_bps),
  ip_whitelist = VALUES(ip_whitelist);

INSERT INTO channels (
  name, pay_type, gateway_url, upstream_merchant_no, rsa_private_key, sign_secret, weight, min_amount, max_amount,
  supports_collect, supports_payout, upstream_collect_rate_bps, upstream_payout_rate_bps, upstream_payout_fee_mode, upstream_payout_fixed_fee, enabled, fuse_enabled
)
VALUES
  ('mock-channel', 'mock', '', '', '', 'channel_secret', 100, 0, 0, 1, 1, 50, 70, 1, 0, 1, 0),
  ('mock-channel-b', 'mock', '', '', '', 'channel_secret_b', 90, 0, 0, 1, 1, 45, 0, 2, 180, 1, 0),
  ('wechat-channel-rate', 'wechat', '', '', '', 'channel_secret_wechat', 100, 0, 0, 1, 1, 35, 65, 1, 0, 1, 0),
  ('alipay-channel-mix', 'alipay', '', '', '', 'channel_secret_alipay', 100, 0, 0, 1, 1, 40, 50, 3, 120, 1, 0)
ON DUPLICATE KEY UPDATE
  sign_secret = VALUES(sign_secret),
  supports_collect = VALUES(supports_collect),
  supports_payout = VALUES(supports_payout),
  upstream_collect_rate_bps = VALUES(upstream_collect_rate_bps),
  upstream_payout_rate_bps = VALUES(upstream_payout_rate_bps),
  upstream_payout_fee_mode = VALUES(upstream_payout_fee_mode),
  upstream_payout_fixed_fee = VALUES(upstream_payout_fixed_fee),
  enabled = VALUES(enabled),
  fuse_enabled = VALUES(fuse_enabled),
  weight = VALUES(weight),
  pay_type = VALUES(pay_type);

INSERT INTO pay_products (code, name, sort_order, enabled) VALUES
  ('mock', 'Mock支付', 10, 1),
  ('wechat', '微信支付', 20, 1),
  ('alipay', '支付宝', 30, 1)
ON DUPLICATE KEY UPDATE name = VALUES(name), sort_order = VALUES(sort_order), enabled = VALUES(enabled);

INSERT INTO payout_products (code, name, sort_order, enabled) VALUES
  ('bank_card', '银行卡代付', 10, 1),
  ('wallet', '钱包代付', 20, 1)
ON DUPLICATE KEY UPDATE name = VALUES(name), sort_order = VALUES(sort_order), enabled = VALUES(enabled);

INSERT INTO pay_product_channels (pay_product_id, channel_id, weight, enabled)
SELECT pp.id, c.id, w.w, 1
FROM pay_products pp
JOIN (
  SELECT 'mock' AS code, 'mock-channel' AS ch, 60 AS w
  UNION ALL SELECT 'mock', 'mock-channel-b', 40
  UNION ALL SELECT 'wechat', 'wechat-channel-rate', 100
  UNION ALL SELECT 'alipay', 'alipay-channel-mix', 100
) w ON pp.code = w.code
JOIN channels c ON c.name = w.ch
ON DUPLICATE KEY UPDATE weight = VALUES(weight), enabled = VALUES(enabled);

INSERT INTO payout_product_channels (payout_product_id, channel_id, weight, enabled)
SELECT pp.id, c.id, 100, 1
FROM payout_products pp
CROSS JOIN channels c
WHERE pp.code = 'bank_card' AND c.name IN ('mock-channel', 'mock-channel-b') AND c.supports_payout = 1
ON DUPLICATE KEY UPDATE weight = VALUES(weight), enabled = VALUES(enabled);

INSERT INTO merchant_pay_products (merchant_id, pay_product_id, enabled, sort_order, merchant_rate_bps)
SELECT m.merchant_id, pp.id, 1, pp.sort_order,
  CASE
    WHEN m.merchant_id = 'm_demo' THEN NULL
    WHEN m.merchant_id = 'm_rate_mix' AND pp.code = 'mock' THEN 30
    WHEN m.merchant_id = 'm_rate_mix' AND pp.code = 'wechat' THEN 120
    WHEN m.merchant_id = 'm_rate_mix' AND pp.code = 'alipay' THEN NULL
    WHEN m.merchant_id = 'm_zero_fee' THEN 0
    ELSE NULL
  END
FROM pay_products pp
JOIN merchants m ON m.merchant_id IN ('m_demo', 'm_rate_mix', 'm_zero_fee')
WHERE pp.code IN ('mock', 'wechat', 'alipay')
ON DUPLICATE KEY UPDATE
  enabled = VALUES(enabled),
  sort_order = VALUES(sort_order),
  merchant_rate_bps = VALUES(merchant_rate_bps);

INSERT INTO merchant_payout_products (merchant_id, payout_product_id, enabled, sort_order, fee_mode, merchant_rate_bps, fee_fixed_amount)
SELECT m.merchant_id, pp.id, 1, pp.sort_order,
  CASE
    WHEN m.merchant_id = 'm_demo' THEN 1
    WHEN m.merchant_id = 'm_rate_mix' THEN 2
    WHEN m.merchant_id = 'm_zero_fee' THEN 3
    ELSE 1
  END AS fee_mode,
  CASE
    WHEN m.merchant_id = 'm_demo' THEN 80
    WHEN m.merchant_id = 'm_rate_mix' THEN 0
    WHEN m.merchant_id = 'm_zero_fee' THEN 40
    ELSE 0
  END AS merchant_rate_bps,
  CASE
    WHEN m.merchant_id = 'm_demo' THEN 0
    WHEN m.merchant_id = 'm_rate_mix' THEN 150
    WHEN m.merchant_id = 'm_zero_fee' THEN 60
    ELSE 0
  END AS fee_fixed_amount
FROM payout_products pp
JOIN merchants m ON m.merchant_id IN ('m_demo', 'm_rate_mix', 'm_zero_fee')
WHERE pp.code = 'bank_card'
ON DUPLICATE KEY UPDATE
  enabled = VALUES(enabled),
  sort_order = VALUES(sort_order),
  fee_mode = VALUES(fee_mode),
  merchant_rate_bps = VALUES(merchant_rate_bps),
  fee_fixed_amount = VALUES(fee_fixed_amount);

INSERT INTO admin_users (username, password_hash, status)
VALUES ('admin', '$2a$10$KT9JCR/85vRqDuRyUGR28O.69/Y5VjbtqmkyX7epzLsKAfcny/rpK', 1)
ON DUPLICATE KEY UPDATE password_hash = VALUES(password_hash), status = VALUES(status);

INSERT INTO global_settings (setting_key, setting_value) VALUES
  ('country_code', 'CN'),
  ('currency_code', 'CNY'),
  ('currency_symbol', '¥')
ON DUPLICATE KEY UPDATE setting_value = VALUES(setting_value);

INSERT INTO collect_orders (
  order_no, merchant_id, merchant_order_no, amount, currency, status, channel_id,
  pay_product_id, pay_product_code, channel_locked, paid_amount, fee_mode, fee_rate_bps, fee_fixed_amount, fee_amount, net_amount,
  return_url, notify_url, upstream_trade_no
)
SELECT
  'C-DEMO-001', 'm_demo', 'MO-C-DEMO-001', 1000, 'CNY', 1, c.id, pp.id, pp.code, 0, 1000, 1, 60, 0, 6, 994,
  '', '', 'UP-C-DEMO-001'
FROM channels c
JOIN pay_products pp ON pp.code = 'mock'
WHERE c.name = 'mock-channel'
ON DUPLICATE KEY UPDATE status = VALUES(status), paid_amount = VALUES(paid_amount), fee_amount = VALUES(fee_amount), net_amount = VALUES(net_amount);

INSERT INTO payout_orders (
  order_no, merchant_id, merchant_order_no, amount, currency, status, channel_id,
  payout_product_id, payout_product_code, paid_amount, fee_mode, fee_rate_bps, fee_fixed_amount, fee_amount, net_amount,
  notify_url, upstream_trade_no
)
SELECT
  'P-DEMO-001', 'm_rate_mix', 'MO-P-DEMO-001', 2000, 'CNY', 1, c.id, pp.id, pp.code, 2000, 2, 0, 150, 150, 1850,
  '', 'UP-P-DEMO-001'
FROM channels c
JOIN payout_products pp ON pp.code = 'bank_card'
WHERE c.name = 'mock-channel-b'
ON DUPLICATE KEY UPDATE status = VALUES(status), paid_amount = VALUES(paid_amount), fee_amount = VALUES(fee_amount), net_amount = VALUES(net_amount);
