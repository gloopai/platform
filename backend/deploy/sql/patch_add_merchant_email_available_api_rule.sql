-- 管理台：校验商户邮箱是否可用（新建前检查）
INSERT INTO admin_api_rules (method, path_pattern, perm_key, status, remark) VALUES
  ('GET', '/v1/admin/merchants/email_available', 'admin.merchants.read', 1, '')
ON DUPLICATE KEY UPDATE perm_key = VALUES(perm_key), status = VALUES(status), remark = VALUES(remark);
