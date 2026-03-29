-- scaffold/platform-admin：仅平台管理端 + service-hub 所需表（无商户/通道/订单等业务表）

CREATE TABLE IF NOT EXISTS admin_users (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  username VARCHAR(64) NOT NULL,
  password_hash VARCHAR(128) NOT NULL,
  status TINYINT NOT NULL DEFAULT 1,
  mfa_secret VARCHAR(128) NOT NULL DEFAULT '',
  mfa_enabled TINYINT NOT NULL DEFAULT 0,
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
  path VARCHAR(128) NULL,
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

CREATE TABLE IF NOT EXISTS admin_api_rules (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  method VARCHAR(16) NOT NULL,
  path_pattern VARCHAR(255) NOT NULL COMMENT '支持 :param 段',
  perm_key VARCHAR(128) NOT NULL,
  status TINYINT NOT NULL DEFAULT 1,
  remark VARCHAR(255) NOT NULL DEFAULT '',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_method_path (method, path_pattern),
  KEY idx_perm (perm_key, status, id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 门户通知落库（可选 NSQ 推送）；service-hub PublishPortalNotification 使用
CREATE TABLE IF NOT EXISTS portal_notifications (
  id CHAR(36) NOT NULL,
  portal VARCHAR(16) NOT NULL COMMENT 'admin | merchant',
  broadcast TINYINT NOT NULL DEFAULT 0,
  title VARCHAR(512) NOT NULL DEFAULT '',
  body TEXT,
  severity VARCHAR(16) NOT NULL DEFAULT 'info',
  link_path VARCHAR(512) NOT NULL DEFAULT '',
  link_query_json TEXT,
  meta_json TEXT,
  target_admin_ids TEXT COMMENT 'JSON array of int64',
  target_merchant_ids TEXT COMMENT 'JSON array of merchant id strings',
  created_at DATETIME(3) NOT NULL,
  PRIMARY KEY (id),
  KEY idx_portal_created (portal, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
