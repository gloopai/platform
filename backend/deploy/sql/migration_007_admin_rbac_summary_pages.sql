-- 恢复 RBAC 汇总入口菜单：功能点 / 接口规则

SET @gid_rbac := (SELECT id FROM admin_menus WHERE menu_key = 'group.rbac' LIMIT 1);

INSERT INTO admin_menus (parent_id, menu_key, label, icon, kind, path, sort_order, placement)
SELECT @gid_rbac, 'menu.rbac_features', '功能点', '', 1, '/rbac/features', 20, 'left'
WHERE @gid_rbac IS NOT NULL
ON DUPLICATE KEY UPDATE
  parent_id = VALUES(parent_id),
  label = VALUES(label),
  icon = VALUES(icon),
  kind = VALUES(kind),
  path = VALUES(path),
  sort_order = VALUES(sort_order),
  placement = VALUES(placement);

INSERT INTO admin_menus (parent_id, menu_key, label, icon, kind, path, sort_order, placement)
SELECT @gid_rbac, 'menu.rbac_api_rules', '接口规则', '', 1, '/rbac/api-rules', 25, 'left'
WHERE @gid_rbac IS NOT NULL
ON DUPLICATE KEY UPDATE
  parent_id = VALUES(parent_id),
  label = VALUES(label),
  icon = VALUES(icon),
  kind = VALUES(kind),
  path = VALUES(path),
  sort_order = VALUES(sort_order),
  placement = VALUES(placement);

-- 兜底：把新增菜单授予 super_admin
INSERT INTO admin_role_menus (role_id, menu_id)
SELECT r.id, m.id
FROM admin_roles r
JOIN admin_menus m ON m.menu_key IN ('menu.rbac_features', 'menu.rbac_api_rules')
WHERE r.code = 'super_admin'
ON DUPLICATE KEY UPDATE menu_id = VALUES(menu_id);
