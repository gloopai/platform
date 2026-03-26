-- 权限点绑定侧栏菜单 menu_key，便于按「页面」做配置总览

ALTER TABLE admin_permissions
  ADD COLUMN menu_key VARCHAR(64) NOT NULL DEFAULT '' COMMENT '对应 admin_menus.menu_key，空表示未绑定菜单' AFTER category,
  KEY idx_perm_menu_key (menu_key, status);
