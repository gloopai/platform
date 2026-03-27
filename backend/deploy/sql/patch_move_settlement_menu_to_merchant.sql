-- 将「资金流水 / 提现申请」入口移动到「商户与接入」分组下
SET @gid_merchant := (SELECT id FROM admin_menus WHERE menu_key = 'group.merchant' LIMIT 1);
SET @rid_super_admin := (SELECT id FROM admin_roles WHERE code = 'super_admin' LIMIT 1);

DELETE FROM admin_menus WHERE menu_key = 'menu.settlement';

INSERT INTO admin_menus (parent_id, menu_key, label, icon, kind, path, sort_order) VALUES
  (@gid_merchant, 'menu.settlement_logs', '资金流水', '', 1, '/settlement/logs', 20),
  (@gid_merchant, 'menu.settlement_withdrawals', '提现申请', '', 1, '/settlement/withdrawals', 30)
ON DUPLICATE KEY UPDATE
  parent_id = VALUES(parent_id),
  label = VALUES(label),
  path = VALUES(path),
  sort_order = VALUES(sort_order);

INSERT INTO admin_role_menus (role_id, menu_id)
SELECT @rid_super_admin, am.id
FROM admin_menus am
WHERE am.menu_key IN ('menu.settlement_logs', 'menu.settlement_withdrawals')
ON DUPLICATE KEY UPDATE menu_id = VALUES(menu_id);
