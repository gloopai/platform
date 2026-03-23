-- 已有库升级：支付产品与订单字段（执行前请备份）
-- MySQL 8+ / InnoDB

CREATE TABLE IF NOT EXISTS payin_products (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  code VARCHAR(32) NOT NULL,
  name VARCHAR(64) NOT NULL,
  sort_order INT NOT NULL DEFAULT 0,
  enabled TINYINT NOT NULL DEFAULT 1,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_payin_product_code (code),
  KEY idx_enabled_sort (enabled, sort_order)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS payin_product_channels (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  payin_product_id BIGINT UNSIGNED NOT NULL,
  channel_id BIGINT UNSIGNED NOT NULL,
  weight INT NOT NULL DEFAULT 100,
  enabled TINYINT NOT NULL DEFAULT 1,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_product_channel (payin_product_id, channel_id),
  KEY idx_channel (channel_id),
  CONSTRAINT fk_ppc_product FOREIGN KEY (payin_product_id) REFERENCES payin_products (id) ON DELETE CASCADE,
  CONSTRAINT fk_ppc_channel FOREIGN KEY (channel_id) REFERENCES channels (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

ALTER TABLE orders
  ADD COLUMN payin_product_id BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER channel_id,
  ADD COLUMN payin_product_code VARCHAR(32) NULL AFTER payin_product_id;
