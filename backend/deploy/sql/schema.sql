CREATE TABLE IF NOT EXISTS merchants (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  merchant_id VARCHAR(64) NOT NULL,
  api_secret VARCHAR(128) NOT NULL,
  status TINYINT NOT NULL DEFAULT 1,
  default_payin_rate_bps INT NOT NULL DEFAULT 0 COMMENT '代收：未单独配置产品费率时使用',
  default_payout_rate_bps INT NOT NULL DEFAULT 0 COMMENT '代付：未单独配置产品费率时使用',
  ip_whitelist TEXT NULL,
  payin_balance BIGINT NOT NULL DEFAULT 0,
  available_balance BIGINT NOT NULL DEFAULT 0,
  frozen_balance BIGINT NOT NULL DEFAULT 0,
  withdrawn_amount BIGINT NOT NULL DEFAULT 0,
  notify_url VARCHAR(512) NULL,
  return_url VARCHAR(512) NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_merchant_id (merchant_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 上游通道：可单独开通代收/代付能力；平台相对上游的费率在通道级配置
CREATE TABLE IF NOT EXISTS channels (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  name VARCHAR(64) NOT NULL,
  payin_type VARCHAR(32) NULL,
  gateway_url VARCHAR(512) NULL,
  upstream_merchant_no VARCHAR(128) NULL,
  rsa_private_key TEXT NULL,
  sign_secret VARCHAR(128) NULL,
  weight INT NOT NULL DEFAULT 100,
  min_amount BIGINT NOT NULL DEFAULT 0,
  max_amount BIGINT NOT NULL DEFAULT 0,
  supports_payin TINYINT NOT NULL DEFAULT 1,
  supports_payout TINYINT NOT NULL DEFAULT 0,
  upstream_payin_rate_bps INT NOT NULL DEFAULT 0 COMMENT '代收：平台相对上游费率（万分比）',
  upstream_payout_rate_bps INT NOT NULL DEFAULT 0 COMMENT '代付：平台相对上游费率（万分比）',
  upstream_payout_fee_mode TINYINT NOT NULL DEFAULT 1 COMMENT '代付上游费率模式：1=比例 2=固定 3=固定+比例',
  upstream_payout_fixed_fee BIGINT NOT NULL DEFAULT 0 COMMENT '代付上游固定手续费（分）',
  enabled TINYINT NOT NULL DEFAULT 1,
  fuse_enabled TINYINT NOT NULL DEFAULT 0,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_enabled_payintype (enabled, payin_type),
  KEY idx_enabled_fuse (enabled, fuse_enabled)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 对外代收产品（微信、支付宝等）
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

-- 对外代付产品（代付到卡/钱包等）
CREATE TABLE IF NOT EXISTS payout_products (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  code VARCHAR(32) NOT NULL,
  name VARCHAR(64) NOT NULL,
  sort_order INT NOT NULL DEFAULT 0,
  enabled TINYINT NOT NULL DEFAULT 1,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_payout_product_code (code),
  KEY idx_enabled_sort (enabled, sort_order)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 代收产品 ↔ 通道（仅权重与启用；费率见 channels / 商户授权）
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

-- 代付产品 ↔ 通道（仅权重与启用；费率见 channels / 商户授权）
CREATE TABLE IF NOT EXISTS payout_product_channels (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  payout_product_id BIGINT UNSIGNED NOT NULL,
  channel_id BIGINT UNSIGNED NOT NULL,
  weight INT NOT NULL DEFAULT 100,
  enabled TINYINT NOT NULL DEFAULT 1,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_payout_product_channel (payout_product_id, channel_id),
  KEY idx_channel (channel_id),
  CONSTRAINT fk_ppoc_product FOREIGN KEY (payout_product_id) REFERENCES payout_products (id) ON DELETE CASCADE,
  CONSTRAINT fk_ppoc_channel FOREIGN KEY (channel_id) REFERENCES channels (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 商户代收产品白名单；merchant_rate_bps NULL 表示使用 merchants.default_payin_rate_bps
CREATE TABLE IF NOT EXISTS merchant_payin_products (
  merchant_id VARCHAR(64) NOT NULL,
  payin_product_id BIGINT UNSIGNED NOT NULL,
  enabled TINYINT NOT NULL DEFAULT 1,
  sort_order INT NOT NULL DEFAULT 0,
  merchant_rate_bps INT NULL COMMENT '对该商户该代收产品的费率，NULL=用商户默认代收费率',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (merchant_id, payin_product_id),
  KEY idx_merchant (merchant_id, enabled),
  CONSTRAINT fk_mpp_product FOREIGN KEY (payin_product_id) REFERENCES payin_products (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 商户代付产品白名单
CREATE TABLE IF NOT EXISTS merchant_payout_products (
  merchant_id VARCHAR(64) NOT NULL,
  payout_product_id BIGINT UNSIGNED NOT NULL,
  enabled TINYINT NOT NULL DEFAULT 1,
  sort_order INT NOT NULL DEFAULT 0,
  fee_mode TINYINT NOT NULL DEFAULT 1 COMMENT '1=比例 2=固定 3=固定+比例',
  merchant_rate_bps INT NULL COMMENT '对该商户该代付产品的费率，NULL=用商户默认代付费率',
  fee_fixed_amount BIGINT NOT NULL DEFAULT 0 COMMENT '固定手续费（分）',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (merchant_id, payout_product_id),
  KEY idx_merchant (merchant_id, enabled),
  CONSTRAINT fk_mppo_product FOREIGN KEY (payout_product_id) REFERENCES payout_products (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS payin_orders (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  order_no VARCHAR(64) NOT NULL,
  merchant_id VARCHAR(64) NOT NULL,
  merchant_order_no VARCHAR(64) NOT NULL,
  amount BIGINT NOT NULL,
  currency VARCHAR(8) NOT NULL DEFAULT 'CNY',
  status TINYINT NOT NULL DEFAULT 0,
  channel_id BIGINT NOT NULL DEFAULT 0,
  payin_product_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  payin_product_code VARCHAR(32) NULL,
  channel_locked TINYINT NOT NULL DEFAULT 0 COMMENT '1=商户指定通道，收银台不可切换',
  paid_amount BIGINT NOT NULL DEFAULT 0,
  fee_mode TINYINT NOT NULL DEFAULT 1 COMMENT '1=比例 2=固定 3=固定+比例',
  fee_rate_bps INT NOT NULL DEFAULT 0,
  fee_fixed_amount BIGINT NOT NULL DEFAULT 0,
  fee_amount BIGINT NOT NULL DEFAULT 0,
  net_amount BIGINT NOT NULL DEFAULT 0,
  return_url VARCHAR(512) NULL,
  notify_url VARCHAR(512) NULL,
  upstream_trade_no VARCHAR(128) NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_pay_order_no (order_no),
  UNIQUE KEY uk_pay_merchant_order (merchant_id, merchant_order_no),
  KEY idx_merchant_created (merchant_id, created_at),
  KEY idx_status_updated (status, updated_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

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
  fee_mode TINYINT NOT NULL DEFAULT 1 COMMENT '1=比例 2=固定 3=固定+比例',
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

-- 提现申请单（phase2）
CREATE TABLE IF NOT EXISTS merchant_withdrawals (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  withdraw_no VARCHAR(64) NOT NULL COMMENT '平台提现单号',
  merchant_id VARCHAR(64) NOT NULL,
  apply_amount BIGINT NOT NULL COMMENT '申请金额（分）',
  fee_amount BIGINT NOT NULL DEFAULT 0 COMMENT '手续费（分）',
  net_amount BIGINT NOT NULL COMMENT '实付金额（分）',
  fiat_debit_amount BIGINT NOT NULL DEFAULT 0 COMMENT '审核通过时扣减的法币余额（分）',
  status TINYINT NOT NULL DEFAULT 0 COMMENT '0=待审核 1=已驳回 2=待打款 3=打款中 4=成功 5=失败',
  receive_account VARCHAR(128) NOT NULL DEFAULT '' COMMENT '收款账号/卡号',
  receive_name VARCHAR(64) NOT NULL DEFAULT '' COMMENT '收款人',
  bank_name VARCHAR(128) NOT NULL DEFAULT '' COMMENT '银行/机构名称',
  apply_note VARCHAR(255) NOT NULL DEFAULT '',
  review_note VARCHAR(255) NOT NULL DEFAULT '',
  payout_note VARCHAR(255) NOT NULL DEFAULT '',
  reviewed_by VARCHAR(64) NOT NULL DEFAULT '',
  reviewed_at TIMESTAMP NULL DEFAULT NULL,
  payouted_by VARCHAR(64) NOT NULL DEFAULT '',
  payouted_at TIMESTAMP NULL DEFAULT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_withdraw_no (withdraw_no),
  KEY idx_merchant_created (merchant_id, created_at),
  KEY idx_status_created (status, created_at)
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

CREATE TABLE IF NOT EXISTS global_settings (
  setting_key VARCHAR(64) NOT NULL,
  setting_value VARCHAR(255) NOT NULL,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (setting_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 管理台 RBAC：角色/菜单（当前只做到菜单粒度；操作权限预留）
CREATE TABLE IF NOT EXISTS admin_roles (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  code VARCHAR(64) NOT NULL,
  name VARCHAR(64) NOT NULL,
  status TINYINT NOT NULL DEFAULT 1,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_admin_role_code (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS admin_menus (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  parent_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  menu_key VARCHAR(64) NOT NULL,
  label VARCHAR(64) NOT NULL,
  icon VARCHAR(32) NOT NULL DEFAULT '',
  kind TINYINT NOT NULL DEFAULT 1 COMMENT '1=leaf 2=group',
  path VARCHAR(128) NULL COMMENT 'leaf 路由路径，形如 /stats',
  sort_order INT NOT NULL DEFAULT 0,
  placement VARCHAR(16) NOT NULL DEFAULT 'left' COMMENT 'left=左侧导航 avatar=头像下拉',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_admin_menu_key (menu_key),
  KEY idx_parent_sort (parent_id, sort_order, id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS admin_user_roles (
  admin_user_id BIGINT UNSIGNED NOT NULL,
  role_id BIGINT UNSIGNED NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (admin_user_id, role_id),
  KEY idx_role (role_id),
  CONSTRAINT fk_admin_user_roles_user FOREIGN KEY (admin_user_id) REFERENCES admin_users (id) ON DELETE CASCADE,
  CONSTRAINT fk_admin_user_roles_role FOREIGN KEY (role_id) REFERENCES admin_roles (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS admin_role_menus (
  role_id BIGINT UNSIGNED NOT NULL,
  menu_id BIGINT UNSIGNED NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (role_id, menu_id),
  KEY idx_menu (menu_id),
  CONSTRAINT fk_admin_role_menus_role FOREIGN KEY (role_id) REFERENCES admin_roles (id) ON DELETE CASCADE,
  CONSTRAINT fk_admin_role_menus_menu FOREIGN KEY (menu_id) REFERENCES admin_menus (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 管理台 RBAC：操作权限点（接口/按钮级）
CREATE TABLE IF NOT EXISTS admin_permissions (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  perm_key VARCHAR(128) NOT NULL,
  label VARCHAR(128) NOT NULL,
  category VARCHAR(64) NOT NULL DEFAULT '',
  menu_key VARCHAR(64) NOT NULL DEFAULT '' COMMENT '对应 admin_menus.menu_key',
  status TINYINT NOT NULL DEFAULT 1,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_admin_perm_key (perm_key),
  KEY idx_category (category, status, id),
  KEY idx_perm_menu_key (menu_key, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS admin_role_permissions (
  role_id BIGINT UNSIGNED NOT NULL,
  perm_id BIGINT UNSIGNED NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (role_id, perm_id),
  KEY idx_perm (perm_id),
  CONSTRAINT fk_admin_role_perms_role FOREIGN KEY (role_id) REFERENCES admin_roles (id) ON DELETE CASCADE,
  CONSTRAINT fk_admin_role_perms_perm FOREIGN KEY (perm_id) REFERENCES admin_permissions (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 管理台 RBAC：接口规则（方法 + path pattern -> perm_key），用于免改代码配置权限映射
CREATE TABLE IF NOT EXISTS admin_api_rules (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  method VARCHAR(16) NOT NULL,
  path_pattern VARCHAR(255) NOT NULL COMMENT '支持 :param 段，如 /v1/admin/merchants/:merchant_id',
  perm_key VARCHAR(128) NOT NULL,
  status TINYINT NOT NULL DEFAULT 1,
  remark VARCHAR(255) NOT NULL DEFAULT '',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_method_path (method, path_pattern),
  KEY idx_perm (perm_key, status, id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
