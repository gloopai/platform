-- 脚手架演示数据：管理员、RBAC（仅权限与安全 + 系统与运维）、global_settings
-- 不含商户 / 通道 / 订单等业务演示数据。
--
-- MySQL 8.0.20+：ON DUPLICATE KEY UPDATE 中勿用 VALUES(col)，改用 INSERT ... AS new + new.col。
-- 重置菜单/RBAC：TRUNCATE 父表时 MySQL 会因外键拒绝，故先 SET FOREIGN_KEY_CHECKS=0（仅本会话）。

INSERT INTO admin_users (username, password_hash, status)
VALUES ('admin', '$2a$10$KT9JCR/85vRqDuRyUGR28O.69/Y5VjbtqmkyX7epzLsKAfcny/rpK', 1) AS new
ON DUPLICATE KEY UPDATE password_hash = new.password_hash, status = new.status;

INSERT INTO admin_roles (code, name, status)
VALUES ('super_admin', '超级管理员', 1) AS new
ON DUPLICATE KEY UPDATE name = new.name, status = new.status;

SET @OLD_FK_CHECKS = @@SESSION.FOREIGN_KEY_CHECKS;
SET SESSION FOREIGN_KEY_CHECKS = 0;
TRUNCATE TABLE admin_role_menus;
TRUNCATE TABLE admin_role_permissions;
TRUNCATE TABLE admin_api_rules;
TRUNCATE TABLE admin_permissions;
TRUNCATE TABLE admin_menus;
SET SESSION FOREIGN_KEY_CHECKS = @OLD_FK_CHECKS;

-- 根级：工作台（登录后默认页），再为分组（sort_order 决定侧栏顺序）
INSERT INTO admin_menus (parent_id, menu_key, label, icon, kind, path, sort_order) VALUES
  (0, 'menu.home', '工作台', 'chart', 1, '/home', 5),
  (0, 'group.rbac', '权限与安全', 'shield', 2, NULL, 10),
  (0, 'group.system', '系统与运维', 'cog', 2, NULL, 20);

SET @gid_rbac := (SELECT id FROM admin_menus WHERE menu_key = 'group.rbac' LIMIT 1);
SET @gid_system := (SELECT id FROM admin_menus WHERE menu_key = 'group.system' LIMIT 1);

INSERT INTO admin_menus (parent_id, menu_key, label, icon, kind, path, sort_order) VALUES
  (@gid_rbac, 'menu.rbac_overview', '配置总览', '', 1, '/rbac/overview', 10),
  (@gid_rbac, 'menu.rbac_menus', '菜单管理', '', 1, '/rbac/menus', 15),
  (@gid_rbac, 'menu.rbac_features', '功能点', '', 1, '/rbac/features', 20),
  (@gid_rbac, 'menu.rbac_api_rules', '接口规则', '', 1, '/rbac/api-rules', 25),
  (@gid_rbac, 'menu.rbac_roles', '角色与授权', '', 1, '/rbac/roles', 30),
  (@gid_rbac, 'menu.rbac_admin_users', '后台用户', '', 1, '/rbac/admin-users', 35),
  (@gid_system, 'menu.system', '系统管理', '', 1, '/system', 10),
  (@gid_system, 'menu.ops', '运维监控', '', 1, '/ops', 20);

INSERT INTO admin_permissions (perm_key, label, category, menu_key, status) VALUES
  ('admin.auth.logout', '退出登录', 'auth', 'menu.system', 1),
  ('admin.ops.read', '运维监控-读取', 'ops', 'menu.ops', 1),
  ('admin.system.read_settings', '系统管理-展示配置读取', 'system', 'menu.system', 1),
  ('admin.system.write_settings', '系统管理-展示配置写入', 'system', 'menu.system', 1),
  ('admin.admin_users.manage', '后台用户管理（列表、增删改、角色、密码与 MFA）', 'admin_users', 'menu.rbac_admin_users', 1),
  ('admin.rbac.my_menu', 'RBAC-读取我的菜单', 'rbac', 'menu.rbac_overview', 1),
  ('admin.rbac.manage', 'RBAC-配置管理（菜单/角色/权限/接口）', 'rbac', 'menu.rbac_roles', 1);

INSERT INTO admin_api_rules (method, path_pattern, perm_key, status, remark) VALUES
  ('POST', '/v1/admin/logout', 'admin.auth.logout', 1, ''),
  ('GET', '/v1/admin/ops/services', 'admin.ops.read', 1, ''),
  ('GET', '/v1/admin/admin_users', 'admin.admin_users.manage', 1, ''),
  ('POST', '/v1/admin/admin_users', 'admin.admin_users.manage', 1, ''),
  ('PUT', '/v1/admin/admin_users/:id', 'admin.admin_users.manage', 1, ''),
  ('DELETE', '/v1/admin/admin_users/:id', 'admin.admin_users.manage', 1, ''),
  ('POST', '/v1/admin/admin_users/:id/reset_password', 'admin.admin_users.manage', 1, ''),
  ('POST', '/v1/admin/admin_users/:id/mfa/setup', 'admin.admin_users.manage', 1, ''),
  ('POST', '/v1/admin/admin_users/:id/mfa/confirm', 'admin.admin_users.manage', 1, ''),
  ('POST', '/v1/admin/admin_users/:id/mfa/disable', 'admin.admin_users.manage', 1, ''),
  ('GET', '/v1/admin/display_settings', 'admin.system.read_settings', 1, ''),
  ('PUT', '/v1/admin/display_settings', 'admin.system.write_settings', 1, ''),
  ('GET', '/v1/admin/rbac/my_menu', 'admin.rbac.my_menu', 1, ''),
  ('GET', '/v1/admin/rbac/roles', 'admin.rbac.manage', 1, ''),
  ('POST', '/v1/admin/rbac/roles', 'admin.rbac.manage', 1, ''),
  ('PUT', '/v1/admin/rbac/roles/:id', 'admin.rbac.manage', 1, ''),
  ('DELETE', '/v1/admin/rbac/roles/:id', 'admin.rbac.manage', 1, ''),
  ('GET', '/v1/admin/rbac/menus', 'admin.rbac.manage', 1, ''),
  ('POST', '/v1/admin/rbac/menus', 'admin.rbac.manage', 1, ''),
  ('PUT', '/v1/admin/rbac/menus/:id', 'admin.rbac.manage', 1, ''),
  ('DELETE', '/v1/admin/rbac/menus/:id', 'admin.rbac.manage', 1, ''),
  ('GET', '/v1/admin/rbac/roles/:id/menus', 'admin.rbac.manage', 1, ''),
  ('PUT', '/v1/admin/rbac/roles/:id/menus', 'admin.rbac.manage', 1, ''),
  ('GET', '/v1/admin/rbac/admin_users/:id/roles', 'admin.admin_users.manage', 1, ''),
  ('PUT', '/v1/admin/rbac/admin_users/:id/roles', 'admin.admin_users.manage', 1, ''),
  ('GET', '/v1/admin/rbac/permissions', 'admin.rbac.manage', 1, ''),
  ('POST', '/v1/admin/rbac/permissions', 'admin.rbac.manage', 1, ''),
  ('PUT', '/v1/admin/rbac/permissions/:id', 'admin.rbac.manage', 1, ''),
  ('DELETE', '/v1/admin/rbac/permissions/:id', 'admin.rbac.manage', 1, ''),
  ('GET', '/v1/admin/rbac/roles/:id/perm_keys', 'admin.rbac.manage', 1, ''),
  ('PUT', '/v1/admin/rbac/roles/:id/perm_keys', 'admin.rbac.manage', 1, ''),
  ('GET', '/v1/admin/rbac/api_rules', 'admin.rbac.manage', 1, ''),
  ('POST', '/v1/admin/rbac/api_rules', 'admin.rbac.manage', 1, ''),
  ('PUT', '/v1/admin/rbac/api_rules/:id', 'admin.rbac.manage', 1, ''),
  ('DELETE', '/v1/admin/rbac/api_rules/:id', 'admin.rbac.manage', 1, '')
AS new
ON DUPLICATE KEY UPDATE
  perm_key = new.perm_key,
  status = new.status,
  remark = new.remark;

INSERT INTO admin_user_roles (admin_user_id, role_id)
SELECT au.id, ar.id
FROM admin_users au
JOIN admin_roles ar ON ar.code = 'super_admin'
WHERE au.username = 'admin'
ON DUPLICATE KEY UPDATE role_id = ar.id;

INSERT INTO admin_role_menus (role_id, menu_id)
SELECT ar.id, am.id
FROM admin_roles ar
JOIN admin_menus am
WHERE ar.code = 'super_admin'
ON DUPLICATE KEY UPDATE menu_id = am.id;

INSERT INTO admin_role_permissions (role_id, perm_id)
SELECT ar.id, ap.id
FROM admin_roles ar
JOIN admin_permissions ap
WHERE ar.code = 'super_admin'
ON DUPLICATE KEY UPDATE perm_id = ap.id;

INSERT INTO global_settings (setting_key, setting_value) VALUES
  ('country_code', 'CN'),
  ('currency_code', 'CNY'),
  ('currency_symbol', '¥'),
  ('merchant_numeric_id_start', '5000000000')
AS new
ON DUPLICATE KEY UPDATE setting_value = new.setting_value;
