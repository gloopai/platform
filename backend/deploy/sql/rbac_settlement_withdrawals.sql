-- Settlement/Withdraw phase2 RBAC SQL (idempotent)
-- 执行顺序：权限点 -> API规则 -> 角色授权

-- 1) 权限点
INSERT INTO admin_permissions (perm_key, label, category, menu_key, status)
SELECT 'admin.settlement.withdraw.read', '查看提现申请', 'settlement', 'settlement', 1
WHERE NOT EXISTS (
  SELECT 1 FROM admin_permissions WHERE perm_key = 'admin.settlement.withdraw.read'
);

INSERT INTO admin_permissions (perm_key, label, category, menu_key, status)
SELECT 'admin.settlement.withdraw.apply', '创建提现申请', 'settlement', 'settlement', 1
WHERE NOT EXISTS (
  SELECT 1 FROM admin_permissions WHERE perm_key = 'admin.settlement.withdraw.apply'
);

INSERT INTO admin_permissions (perm_key, label, category, menu_key, status)
SELECT 'admin.settlement.withdraw.review', '审核提现申请', 'settlement', 'settlement', 1
WHERE NOT EXISTS (
  SELECT 1 FROM admin_permissions WHERE perm_key = 'admin.settlement.withdraw.review'
);

INSERT INTO admin_permissions (perm_key, label, category, menu_key, status)
SELECT 'admin.settlement.withdraw.payout', '执行提现打款', 'settlement', 'settlement', 1
WHERE NOT EXISTS (
  SELECT 1 FROM admin_permissions WHERE perm_key = 'admin.settlement.withdraw.payout'
);

-- 2) API 规则（预留路径，后端接入时保持一致）
INSERT INTO admin_api_rules (method, path_pattern, perm_key, status, remark)
SELECT 'GET', '/v1/admin/settlement/withdrawals', 'admin.settlement.withdraw.read', 1, '结算提现：查询申请单'
WHERE NOT EXISTS (
  SELECT 1 FROM admin_api_rules WHERE method = 'GET' AND path_pattern = '/v1/admin/settlement/withdrawals'
);

INSERT INTO admin_api_rules (method, path_pattern, perm_key, status, remark)
SELECT 'POST', '/v1/admin/settlement/withdrawals', 'admin.settlement.withdraw.apply', 1, '结算提现：创建申请'
WHERE NOT EXISTS (
  SELECT 1 FROM admin_api_rules WHERE method = 'POST' AND path_pattern = '/v1/admin/settlement/withdrawals'
);

INSERT INTO admin_api_rules (method, path_pattern, perm_key, status, remark)
SELECT 'PUT', '/v1/admin/settlement/withdrawals/:id/review', 'admin.settlement.withdraw.review', 1, '结算提现：审核'
WHERE NOT EXISTS (
  SELECT 1 FROM admin_api_rules WHERE method = 'PUT' AND path_pattern = '/v1/admin/settlement/withdrawals/:id/review'
);

INSERT INTO admin_api_rules (method, path_pattern, perm_key, status, remark)
SELECT 'PUT', '/v1/admin/settlement/withdrawals/:id/payout', 'admin.settlement.withdraw.payout', 1, '结算提现：打款'
WHERE NOT EXISTS (
  SELECT 1 FROM admin_api_rules WHERE method = 'PUT' AND path_pattern = '/v1/admin/settlement/withdrawals/:id/payout'
);

-- 3) 给 super_admin 默认授权（若存在）
INSERT INTO admin_role_permissions (role_id, perm_id)
SELECT r.id, p.id
FROM admin_roles r
JOIN admin_permissions p ON p.perm_key IN (
  'admin.settlement.withdraw.read',
  'admin.settlement.withdraw.apply',
  'admin.settlement.withdraw.review',
  'admin.settlement.withdraw.payout'
)
WHERE r.code = 'super_admin'
AND NOT EXISTS (
  SELECT 1 FROM admin_role_permissions rp
  WHERE rp.role_id = r.id AND rp.perm_id = p.id
);
