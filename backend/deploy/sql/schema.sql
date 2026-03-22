CREATE TABLE IF NOT EXISTS merchants (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  merchant_id VARCHAR(64) NOT NULL,
  api_secret VARCHAR(128) NOT NULL,
  status TINYINT NOT NULL DEFAULT 1,
  rate_bps INT NOT NULL DEFAULT 0,
  ip_whitelist TEXT NULL,
  balance BIGINT NOT NULL DEFAULT 0,
  frozen_balance BIGINT NOT NULL DEFAULT 0,
  withdrawn_amount BIGINT NOT NULL DEFAULT 0,
  notify_url VARCHAR(512) NULL,
  return_url VARCHAR(512) NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_merchant_id (merchant_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS channels (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  name VARCHAR(64) NOT NULL,
  pay_type VARCHAR(32) NULL,
  gateway_url VARCHAR(512) NULL,
  upstream_merchant_no VARCHAR(128) NULL,
  rsa_private_key TEXT NULL,
  sign_secret VARCHAR(128) NULL,
  weight INT NOT NULL DEFAULT 100,
  min_amount BIGINT NOT NULL DEFAULT 0,
  max_amount BIGINT NOT NULL DEFAULT 0,
  enabled TINYINT NOT NULL DEFAULT 1,
  fuse_enabled TINYINT NOT NULL DEFAULT 0,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_enabled_paytype (enabled, pay_type),
  KEY idx_enabled_fuse (enabled, fuse_enabled)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 对外支付产品（微信、支付宝等）；与 channels 多对多，见 pay_product_channels
CREATE TABLE IF NOT EXISTS pay_products (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  code VARCHAR(32) NOT NULL,
  name VARCHAR(64) NOT NULL,
  sort_order INT NOT NULL DEFAULT 0,
  enabled TINYINT NOT NULL DEFAULT 1,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_pay_product_code (code),
  KEY idx_enabled_sort (enabled, sort_order)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 某支付产品下可用哪些上游通道及路由权重（权重用于同产品内加权随机）
CREATE TABLE IF NOT EXISTS pay_product_channels (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  pay_product_id BIGINT UNSIGNED NOT NULL,
  channel_id BIGINT UNSIGNED NOT NULL,
  weight INT NOT NULL DEFAULT 100,
  enabled TINYINT NOT NULL DEFAULT 1,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_product_channel (pay_product_id, channel_id),
  KEY idx_channel (channel_id),
  CONSTRAINT fk_ppc_product FOREIGN KEY (pay_product_id) REFERENCES pay_products (id) ON DELETE CASCADE,
  CONSTRAINT fk_ppc_channel FOREIGN KEY (channel_id) REFERENCES channels (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 商户可用的支付产品（白名单）；收银台仅展示此处配置且平台侧仍可用的产品
CREATE TABLE IF NOT EXISTS merchant_pay_products (
  merchant_id VARCHAR(64) NOT NULL,
  pay_product_id BIGINT UNSIGNED NOT NULL,
  enabled TINYINT NOT NULL DEFAULT 1,
  sort_order INT NOT NULL DEFAULT 0,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (merchant_id, pay_product_id),
  KEY idx_merchant (merchant_id, enabled),
  CONSTRAINT fk_mpp_product FOREIGN KEY (pay_product_id) REFERENCES pay_products (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS orders (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  order_no VARCHAR(64) NOT NULL,
  merchant_id VARCHAR(64) NOT NULL,
  merchant_order_no VARCHAR(64) NOT NULL,
  amount BIGINT NOT NULL,
  currency VARCHAR(8) NOT NULL DEFAULT 'CNY',
  status TINYINT NOT NULL DEFAULT 0,
  channel_id BIGINT NOT NULL DEFAULT 0,
  pay_product_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  pay_product_code VARCHAR(32) NULL,
  channel_locked TINYINT NOT NULL DEFAULT 0 COMMENT '1=商户指定通道，收银台不可切换',
  paid_amount BIGINT NOT NULL DEFAULT 0,
  return_url VARCHAR(512) NULL,
  notify_url VARCHAR(512) NULL,
  upstream_trade_no VARCHAR(128) NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_order_no (order_no),
  UNIQUE KEY uk_merchant_order (merchant_id, merchant_order_no),
  KEY idx_merchant_created (merchant_id, created_at),
  KEY idx_status_updated (status, updated_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS fund_logs (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  merchant_id VARCHAR(64) NOT NULL,
  order_no VARCHAR(64) NOT NULL,
  change_type VARCHAR(32) NOT NULL,
  amount BIGINT NOT NULL,
  balance_before BIGINT NOT NULL,
  balance_after BIGINT NOT NULL,
  reason VARCHAR(128) NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_order_change (order_no, change_type),
  KEY idx_merchant_created (merchant_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS merchant_notify_logs (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  merchant_id VARCHAR(64) NOT NULL,
  order_no VARCHAR(64) NOT NULL,
  notify_url VARCHAR(512) NOT NULL,
  attempt INT NOT NULL DEFAULT 0,
  http_status INT NOT NULL DEFAULT 0,
  response_body TEXT NULL,
  error_msg VARCHAR(256) NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_order_created (order_no, created_at),
  KEY idx_merchant_created (merchant_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS admin_users (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  username VARCHAR(64) NOT NULL,
  password_hash VARCHAR(128) NOT NULL,
  status TINYINT NOT NULL DEFAULT 1,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_username (username)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS admin_sessions (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  admin_id BIGINT UNSIGNED NOT NULL,
  token_hash CHAR(64) NOT NULL,
  expires_at TIMESTAMP NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_token_hash (token_hash),
  KEY idx_admin_expires (admin_id, expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS merchant_sessions (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  merchant_id VARCHAR(64) NOT NULL,
  token_hash CHAR(64) NOT NULL,
  expires_at TIMESTAMP NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_token_hash (token_hash),
  KEY idx_merchant_expires (merchant_id, expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
