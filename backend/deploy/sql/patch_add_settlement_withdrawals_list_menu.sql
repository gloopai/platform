-- 侧栏增加「提现申请列表」；已拥有「提现申请」菜单的角色同步获得本项
SET @gid_merchant := (SELECT id FROM admin_menus WHERE menu_key = 'group.merchant' LIMIT 1);

INSERT INTO admin_menus (parent_id, menu_key, label, icon, kind, path, sort_order)
SELECT @gid_merchant, 'menu.settlement_withdrawals_list', '提现申请列表', '', 1, '/settlement/withdrawals/list', 35
FROM DUAL
WHERE @gid_merchant IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM admin_menus WHERE menu_key = 'menu.settlement_withdrawals_list');

-- 超级管理员
INSERT INTO admin_role_menus (role_id, menu_id)
SELECT ar.id, am.id
FROM admin_roles ar
JOIN admin_menus am ON am.menu_key = 'menu.settlement_withdrawals_list'
WHERE ar.code = 'super_admin'
ON DUPLICATE KEY UPDATE menu_id = VALUES(menu_id);

-- 已勾选「提现申请」的角色一并勾选「提现申请列表」
INSERT INTO admin_role_menus (role_id, menu_id)
SELECT rm.role_id, am.id
FROM admin_role_menus rm
JOIN admin_menus src ON src.id = rm.menu_id AND src.menu_key = 'menu.settlement_withdrawals'
JOIN admin_menus am ON am.menu_key = 'menu.settlement_withdrawals_list'
ON DUPLICATE KEY UPDATE menu_id = VALUES(menu_id);
