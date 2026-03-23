-- 停机窗口执行：订单主表拆分为代收/代付
START TRANSACTION;

CREATE TABLE IF NOT EXISTS payin_orders LIKE orders;
CREATE TABLE IF NOT EXISTS payout_orders (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  order_no VARCHAR(64) NOT NULL,
  merchant_id VARCHAR(64) NOT NULL,
  merchant_order_no VARCHAR(64) NOT NULL,
  amount BIGINT NOT NULL,
  currency VARCHAR(8) NOT NULL DEFAULT 'CNY',
  status TINYINT NOT NULL DEFAULT 0,
  channel_id BIGINT NOT NULL DEFAULT 0,
  payout_product_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  payout_product_code VARCHAR(32) NULL,
  paid_amount BIGINT NOT NULL DEFAULT 0,
  fee_mode TINYINT NOT NULL DEFAULT 1,
  fee_rate_bps INT NOT NULL DEFAULT 0,
  fee_fixed_amount BIGINT NOT NULL DEFAULT 0,
  fee_amount BIGINT NOT NULL DEFAULT 0,
  net_amount BIGINT NOT NULL DEFAULT 0,
  notify_url VARCHAR(512) NULL,
  upstream_trade_no VARCHAR(128) NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_payout_order_no (order_no),
  UNIQUE KEY uk_payout_merchant_order (merchant_id, merchant_order_no),
  KEY idx_merchant_created (merchant_id, created_at),
  KEY idx_status_updated (status, updated_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO payin_orders (
  order_no, merchant_id, merchant_order_no, amount, currency, status, channel_id,
  payin_product_id, payin_product_code, channel_locked, paid_amount, fee_mode, fee_rate_bps,
  fee_fixed_amount, fee_amount, net_amount, return_url, notify_url, upstream_trade_no, created_at, updated_at
)
SELECT
  order_no, merchant_id, merchant_order_no, amount, currency, status, channel_id,
  payin_product_id, payin_product_code, channel_locked, paid_amount, fee_mode, fee_rate_bps,
  fee_fixed_amount, fee_amount, net_amount, return_url, notify_url, upstream_trade_no, created_at, updated_at
FROM orders
ON DUPLICATE KEY UPDATE
  status = VALUES(status),
  paid_amount = VALUES(paid_amount),
  upstream_trade_no = VALUES(upstream_trade_no),
  updated_at = VALUES(updated_at);

COMMIT;
