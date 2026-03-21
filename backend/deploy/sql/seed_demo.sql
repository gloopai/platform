INSERT INTO merchants (merchant_id, api_secret, status, rate_bps, ip_whitelist, balance, notify_url)
VALUES ('m_demo', 'demo_secret', 1, 0, '127.0.0.1', 0, '')
ON DUPLICATE KEY UPDATE api_secret = VALUES(api_secret), status = VALUES(status);

INSERT INTO channels (name, pay_type, gateway_url, upstream_merchant_no, rsa_private_key, sign_secret, weight, min_amount, max_amount, enabled, fuse_enabled)
VALUES ('mock-channel', 'mock', '', '', '', 'channel_secret', 100, 0, 0, 1, 0)
ON DUPLICATE KEY UPDATE sign_secret = VALUES(sign_secret), enabled = VALUES(enabled), fuse_enabled = VALUES(fuse_enabled), weight = VALUES(weight);

INSERT INTO admin_users (username, password_hash, status)
VALUES ('admin', '$2a$10$KT9JCR/85vRqDuRyUGR28O.69/Y5VjbtqmkyX7epzLsKAfcny/rpK', 1)
ON DUPLICATE KEY UPDATE password_hash = VALUES(password_hash), status = VALUES(status);
