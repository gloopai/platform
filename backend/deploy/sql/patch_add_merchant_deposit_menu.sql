-- 侧栏「商户管理」下增加「资金存入」；超级管理员与已拥有「商户列表」菜单的角色同步授权
SET @gid_merchant := (SELECT id FROM admin_menus WHERE menu_key = 'group.merchant' LIMIT 1);

INSERT INTO admin_menus (parent_id, menu_key, label, icon, kind, path, sort_order)
SELECT @gid_merchant, 'menu.merchant_deposit', '资金存入', '', 1, '/merchants/deposit', 15
FROM DUAL
WHERE @gid_merchant IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM admin_menus WHERE menu_key = 'menu.merchant_deposit');

INSERT INTO admin_permissions (perm_key, label, category, menu_key, status)
SELECT 'admin.merchants.deposit', '商户-资金存入', 'merchants', 'menu.merchant_deposit', 1
FROM DUAL
WHERE NOT EXISTS (SELECT 1 FROM admin_permissions WHERE perm_key = 'admin.merchants.deposit');

INSERT INTO admin_role_permissions (role_id, perm_id)
SELECT ar.id, ap.id
FROM admin_roles ar
JOIN admin_permissions ap ON ap.perm_key = 'admin.merchants.deposit'
WHERE ar.code = 'super_admin'
ON DUPLICATE KEY UPDATE perm_id = VALUES(perm_id);

INSERT INTO admin_role_menus (role_id, menu_id)
SELECT ar.id, am.id
FROM admin_roles ar
JOIN admin_menus am ON am.menu_key = 'menu.merchant_deposit'
WHERE ar.code = 'super_admin'
ON DUPLICATE KEY UPDATE menu_id = VALUES(menu_id);

INSERT INTO admin_role_menus (role_id, menu_id)
SELECT rm.role_id, am.id
FROM admin_role_menus rm
JOIN admin_menus src ON src.id = rm.menu_id AND src.menu_key = 'menu.merchants'
JOIN admin_menus am ON am.menu_key = 'menu.merchant_deposit'
ON DUPLICATE KEY UPDATE menu_id = VALUES(menu_id);

INSERT INTO admin_api_rules (method, path_pattern, perm_key, status, remark)
SELECT 'POST', '/v1/admin/settlement/deposit', 'admin.merchants.deposit', 1, ''
FROM DUAL
WHERE NOT EXISTS (
  SELECT 1 FROM admin_api_rules WHERE method = 'POST' AND path_pattern = '/v1/admin/settlement/deposit'
);
