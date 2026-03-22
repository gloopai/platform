INSERT INTO merchants (merchant_id, api_secret, status, default_collect_rate_bps, default_payout_rate_bps, ip_whitelist, balance, notify_url)
VALUES ('m_demo', 'demo_secret', 1, 60, 80, '127.0.0.1', 0, '')
ON DUPLICATE KEY UPDATE api_secret = VALUES(api_secret), status = VALUES(status);

INSERT INTO channels (name, pay_type, gateway_url, upstream_merchant_no, rsa_private_key, sign_secret, weight, min_amount, max_amount,
  supports_collect, supports_payout, upstream_collect_cost_bps, upstream_payout_cost_bps, enabled, fuse_enabled)
VALUES ('mock-channel', 'mock', '', '', '', 'channel_secret', 100, 0, 0, 1, 1, 50, 70, 1, 0)
ON DUPLICATE KEY UPDATE sign_secret = VALUES(sign_secret), enabled = VALUES(enabled), fuse_enabled = VALUES(fuse_enabled), weight = VALUES(weight);

INSERT INTO channels (name, pay_type, gateway_url, upstream_merchant_no, rsa_private_key, sign_secret, weight, min_amount, max_amount,
  supports_collect, supports_payout, upstream_collect_cost_bps, upstream_payout_cost_bps, enabled, fuse_enabled)
SELECT 'mock-channel-b', 'mock', '', '', '', 'channel_secret_b', 100, 0, 0, 1, 0, 45, 0, 1, 0
FROM DUAL
WHERE NOT EXISTS (SELECT 1 FROM channels WHERE name = 'mock-channel-b' LIMIT 1);

INSERT INTO pay_products (code, name, sort_order, enabled) VALUES
  ('mock', 'Mock支付', 10, 1),
  ('wechat', '微信支付', 20, 1),
  ('alipay', '支付宝', 30, 1)
ON DUPLICATE KEY UPDATE name = VALUES(name), sort_order = VALUES(sort_order), enabled = VALUES(enabled);

INSERT INTO payout_products (code, name, sort_order, enabled) VALUES
  ('bank_card', '银行卡代付', 10, 1),
  ('wallet', '钱包代付', 20, 1)
ON DUPLICATE KEY UPDATE name = VALUES(name), sort_order = VALUES(sort_order), enabled = VALUES(enabled);

INSERT INTO pay_product_channels (pay_product_id, channel_id, weight, cost_rate_bps, enabled)
SELECT pp.id, c.id, w.w, NULL, 1
FROM pay_products pp
JOIN (
  SELECT 'mock' AS code, 'mock-channel' AS ch, 60 AS w
  UNION ALL SELECT 'mock', 'mock-channel-b', 40
) w ON pp.code = w.code
JOIN channels c ON c.name = w.ch
ON DUPLICATE KEY UPDATE weight = VALUES(weight), enabled = VALUES(enabled);

INSERT INTO payout_product_channels (payout_product_id, channel_id, weight, cost_rate_bps, enabled)
SELECT pp.id, c.id, 100, NULL, 1
FROM payout_products pp
CROSS JOIN channels c
WHERE pp.code = 'bank_card' AND c.name = 'mock-channel' AND c.supports_payout = 1
ON DUPLICATE KEY UPDATE weight = VALUES(weight), enabled = VALUES(enabled);

INSERT INTO merchant_pay_products (merchant_id, pay_product_id, enabled, sort_order, merchant_rate_bps)
SELECT 'm_demo', pp.id, 1, pp.sort_order, NULL
FROM pay_products pp
WHERE pp.code IN ('mock', 'wechat', 'alipay')
ON DUPLICATE KEY UPDATE enabled = VALUES(enabled), sort_order = VALUES(sort_order);

INSERT INTO merchant_payout_products (merchant_id, payout_product_id, enabled, sort_order, merchant_rate_bps)
SELECT 'm_demo', pp.id, 1, pp.sort_order, NULL
FROM payout_products pp
WHERE pp.code = 'bank_card'
ON DUPLICATE KEY UPDATE enabled = VALUES(enabled), sort_order = VALUES(sort_order);

INSERT INTO admin_users (username, password_hash, status)
VALUES ('admin', '$2a$10$KT9JCR/85vRqDuRyUGR28O.69/Y5VjbtqmkyX7epzLsKAfcny/rpK', 1)
ON DUPLICATE KEY UPDATE password_hash = VALUES(password_hash), status = VALUES(status);
