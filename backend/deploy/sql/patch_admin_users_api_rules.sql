-- 补齐后台用户管理相关接口的 RBAC 映射（中间件对未注册接口 fail-closed 会返回 403）。
-- 已与 seed_demo 对齐；在已有库上可单独执行本文件（幂等）。

INSERT INTO admin_api_rules (method, path_pattern, perm_key, status, remark) VALUES
  ('POST', '/v1/admin/admin_users', 'admin.admin_users.manage', 1, ''),
  ('PUT', '/v1/admin/admin_users/:id', 'admin.admin_users.manage', 1, ''),
  ('DELETE', '/v1/admin/admin_users/:id', 'admin.admin_users.manage', 1, ''),
  ('POST', '/v1/admin/admin_users/:id/reset_password', 'admin.admin_users.manage', 1, ''),
  ('POST', '/v1/admin/admin_users/:id/mfa/setup', 'admin.admin_users.manage', 1, ''),
  ('POST', '/v1/admin/admin_users/:id/mfa/confirm', 'admin.admin_users.manage', 1, ''),
  ('POST', '/v1/admin/admin_users/:id/mfa/disable', 'admin.admin_users.manage', 1, '')
ON DUPLICATE KEY UPDATE
  perm_key = VALUES(perm_key),
  status = VALUES(status),
  remark = VALUES(remark);
