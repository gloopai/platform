-- 清理历史演示商户，统一只保留 m_demo 测试数据
DELETE FROM merchant_payout_products WHERE merchant_id IN ('m_rate_mix', 'm_zero_fee');
DELETE FROM merchant_payin_products WHERE merchant_id IN ('m_rate_mix', 'm_zero_fee');
DELETE FROM merchants WHERE merchant_id IN ('m_rate_mix', 'm_zero_fee');

INSERT INTO merchants (merchant_id, api_secret, status, default_payin_rate_bps, default_payout_rate_bps, ip_whitelist, payin_balance, available_balance, notify_url)
VALUES
  ('m_demo', 'demo_secret', 1, 60, 80, '127.0.0.1', 100000, 100000, '')
ON DUPLICATE KEY UPDATE
  api_secret = VALUES(api_secret),
  status = VALUES(status),
  default_payin_rate_bps = VALUES(default_payin_rate_bps),
  default_payout_rate_bps = VALUES(default_payout_rate_bps),
  payin_balance = VALUES(payin_balance),
  available_balance = VALUES(available_balance),
  ip_whitelist = VALUES(ip_whitelist);

INSERT INTO channels (
  name, payin_type, gateway_url, upstream_merchant_no, rsa_private_key, sign_secret, weight, min_amount, max_amount,
  supports_payin, supports_payout, upstream_payin_rate_bps, upstream_payout_rate_bps, upstream_payout_fee_mode, upstream_payout_fixed_fee, enabled, fuse_enabled
)
VALUES
  ('mock-channel', 'mock', '', '', '', 'channel_secret', 100, 0, 0, 1, 1, 50, 70, 1, 0, 1, 0),
  ('mock-channel-b', 'mock', '', '', '', 'channel_secret_b', 90, 0, 0, 1, 1, 45, 0, 2, 180, 1, 0),
  ('wechat-channel-rate', 'wechat', '', '', '', 'channel_secret_wechat', 100, 0, 0, 1, 1, 35, 65, 1, 0, 1, 0),
  ('alipay-channel-mix', 'alipay', '', '', '', 'channel_secret_alipay', 100, 0, 0, 1, 1, 40, 50, 3, 120, 1, 0)
ON DUPLICATE KEY UPDATE
  sign_secret = VALUES(sign_secret),
  supports_payin = VALUES(supports_payin),
  supports_payout = VALUES(supports_payout),
  upstream_payin_rate_bps = VALUES(upstream_payin_rate_bps),
  upstream_payout_rate_bps = VALUES(upstream_payout_rate_bps),
  upstream_payout_fee_mode = VALUES(upstream_payout_fee_mode),
  upstream_payout_fixed_fee = VALUES(upstream_payout_fixed_fee),
  enabled = VALUES(enabled),
  fuse_enabled = VALUES(fuse_enabled),
  weight = VALUES(weight),
  payin_type = VALUES(payin_type);

INSERT INTO payin_products (code, name, sort_order, enabled) VALUES
  ('mock', 'Mock支付', 10, 1),
  ('wechat', '微信支付', 20, 1),
  ('alipay', '支付宝', 30, 1)
ON DUPLICATE KEY UPDATE name = VALUES(name), sort_order = VALUES(sort_order), enabled = VALUES(enabled);

INSERT INTO payout_products (code, name, sort_order, enabled) VALUES
  ('bank_card', '银行卡代付', 10, 1),
  ('wallet', '钱包代付', 20, 1)
ON DUPLICATE KEY UPDATE name = VALUES(name), sort_order = VALUES(sort_order), enabled = VALUES(enabled);

INSERT INTO payin_product_channels (payin_product_id, channel_id, weight, enabled)
SELECT pp.id, c.id, w.w, 1
FROM payin_products pp
JOIN (
  SELECT 'mock' AS code, 'mock-channel' AS ch, 60 AS w
  UNION ALL SELECT 'mock', 'mock-channel-b', 40
  UNION ALL SELECT 'wechat', 'wechat-channel-rate', 100
  UNION ALL SELECT 'alipay', 'alipay-channel-mix', 100
) w ON pp.code = w.code
JOIN channels c ON c.name = w.ch
ON DUPLICATE KEY UPDATE weight = VALUES(weight), enabled = VALUES(enabled);

INSERT INTO payout_product_channels (payout_product_id, channel_id, weight, enabled)
SELECT pp.id, c.id, 100, 1
FROM payout_products pp
CROSS JOIN channels c
WHERE pp.code = 'bank_card' AND c.name IN ('mock-channel', 'mock-channel-b') AND c.supports_payout = 1
ON DUPLICATE KEY UPDATE weight = VALUES(weight), enabled = VALUES(enabled);

INSERT INTO merchant_payin_products (merchant_id, payin_product_id, enabled, sort_order, merchant_rate_bps)
SELECT m.merchant_id, pp.id, 1, pp.sort_order,
  CASE
    WHEN m.merchant_id = 'm_demo' AND pp.code = 'mock' THEN NULL
    WHEN m.merchant_id = 'm_demo' AND pp.code = 'wechat' THEN 120
    WHEN m.merchant_id = 'm_demo' AND pp.code = 'alipay' THEN 0
    ELSE NULL
  END
FROM payin_products pp
JOIN merchants m ON m.merchant_id IN ('m_demo')
WHERE pp.code IN ('mock', 'wechat', 'alipay')
ON DUPLICATE KEY UPDATE
  enabled = VALUES(enabled),
  sort_order = VALUES(sort_order),
  merchant_rate_bps = VALUES(merchant_rate_bps);

INSERT INTO merchant_payout_products (merchant_id, payout_product_id, enabled, sort_order, fee_mode, merchant_rate_bps, fee_fixed_amount)
SELECT m.merchant_id, pp.id, 1, pp.sort_order,
  CASE
    WHEN m.merchant_id = 'm_demo' THEN 3
    ELSE 1
  END AS fee_mode,
  CASE
    WHEN m.merchant_id = 'm_demo' THEN 40
    ELSE 0
  END AS merchant_rate_bps,
  CASE
    WHEN m.merchant_id = 'm_demo' THEN 60
    ELSE 0
  END AS fee_fixed_amount
FROM payout_products pp
JOIN merchants m ON m.merchant_id IN ('m_demo')
WHERE pp.code = 'bank_card'
ON DUPLICATE KEY UPDATE
  enabled = VALUES(enabled),
  sort_order = VALUES(sort_order),
  fee_mode = VALUES(fee_mode),
  merchant_rate_bps = VALUES(merchant_rate_bps),
  fee_fixed_amount = VALUES(fee_fixed_amount);

INSERT INTO admin_users (username, password_hash, status)
VALUES ('admin', '$2a$10$KT9JCR/85vRqDuRyUGR28O.69/Y5VjbtqmkyX7epzLsKAfcny/rpK', 1)
ON DUPLICATE KEY UPDATE password_hash = VALUES(password_hash), status = VALUES(status);

-- ---- 管理台 RBAC 初始化（菜单级）----
INSERT INTO admin_roles (code, name, status)
VALUES ('super_admin', '超级管理员', 1)
ON DUPLICATE KEY UPDATE name = VALUES(name), status = VALUES(status);

-- 顶层菜单：单页 + 分组
INSERT INTO admin_menus (parent_id, menu_key, label, icon, kind, path, sort_order) VALUES
  (0, 'menu.stats', '系统概览', 'chart', 1, '/stats', 10),
  (0, 'group.merchant', '商户与接入', 'briefcase', 2, NULL, 20),
  (0, 'group.channel', '通道与路由', 'layers', 2, NULL, 30),
  (0, 'group.trade', '交易与资金', 'credit', 2, NULL, 40),
  (0, 'group.rbac', '权限与安全', 'shield', 2, NULL, 45),
  (0, 'group.system', '系统与运维', 'cog', 2, NULL, 50)
ON DUPLICATE KEY UPDATE
  label = VALUES(label),
  icon = VALUES(icon),
  kind = VALUES(kind),
  path = VALUES(path),
  sort_order = VALUES(sort_order);

-- 分组子菜单
DELETE FROM admin_menus WHERE menu_key = 'menu.rbac';
DELETE FROM admin_menus WHERE menu_key = 'menu.rbac_permissions';
DELETE FROM admin_menus WHERE menu_key = 'menu.rbac_api_rules';
DELETE FROM admin_menus WHERE menu_key = 'menu.rbac_features';

SET @gid_merchant := (SELECT id FROM admin_menus WHERE menu_key = 'group.merchant' LIMIT 1);
SET @gid_channel := (SELECT id FROM admin_menus WHERE menu_key = 'group.channel' LIMIT 1);
SET @gid_trade := (SELECT id FROM admin_menus WHERE menu_key = 'group.trade' LIMIT 1);
SET @gid_rbac := (SELECT id FROM admin_menus WHERE menu_key = 'group.rbac' LIMIT 1);
SET @gid_system := (SELECT id FROM admin_menus WHERE menu_key = 'group.system' LIMIT 1);

INSERT INTO admin_menus (parent_id, menu_key, label, icon, kind, path, sort_order) VALUES
  (@gid_merchant, 'menu.merchants', '商户管理', '', 1, '/merchants', 10),
  (@gid_merchant, 'menu.merchant_payin_products', '代收产品', '', 1, '/merchant-payin-products', 20),
  (@gid_merchant, 'menu.merchant_payout_products', '代付产品', '', 1, '/merchant-payout-products', 30),

  (@gid_channel, 'menu.channels', '通道管理', '', 1, '/channels', 10),
  (@gid_channel, 'menu.routing', '路由策略', '', 1, '/routing', 20),
  (@gid_channel, 'menu.channel_health', '通道监控', '', 1, '/channel-health', 30),

  (@gid_trade, 'menu.payin_orders', '代收订单', '', 1, '/payin-orders', 10),
  (@gid_trade, 'menu.payout_orders', '代付订单', '', 1, '/payout-orders', 20),
  (@gid_trade, 'menu.refunds', '退款与差错', '', 1, '/refunds', 30),
  (@gid_trade, 'menu.reconcile', '对账中心', '', 1, '/reconcile', 40),
  (@gid_trade, 'menu.settlement', '结算与提现', '', 1, '/settlement', 50),

  (@gid_rbac, 'menu.rbac_overview', '配置总览', '', 1, '/rbac/overview', 10),
  (@gid_rbac, 'menu.rbac_menus', '菜单管理', '', 1, '/rbac/menus', 15),
  (@gid_rbac, 'menu.rbac_roles', '角色与授权', '', 1, '/rbac/roles', 20),
  (@gid_rbac, 'menu.rbac_admin_users', '后台用户', '', 1, '/rbac/admin-users', 22),

  (@gid_system, 'menu.system', '系统管理', '', 1, '/system', 10),
  (@gid_system, 'menu.ops', '运维监控', '', 1, '/ops', 20)
ON DUPLICATE KEY UPDATE
  parent_id = VALUES(parent_id),
  label = VALUES(label),
  icon = VALUES(icon),
  kind = VALUES(kind),
  path = VALUES(path),
  sort_order = VALUES(sort_order);

-- 默认把 demo admin 绑定为超级管理员，并授予全部菜单
INSERT INTO admin_user_roles (admin_user_id, role_id)
SELECT au.id, ar.id
FROM admin_users au
JOIN admin_roles ar ON ar.code = 'super_admin'
WHERE au.username = 'admin'
ON DUPLICATE KEY UPDATE role_id = VALUES(role_id);

INSERT INTO admin_role_menus (role_id, menu_id)
SELECT ar.id, am.id
FROM admin_roles ar
JOIN admin_menus am
WHERE ar.code = 'super_admin'
ON DUPLICATE KEY UPDATE menu_id = VALUES(menu_id);

-- ---- 管理台操作权限点（接口/按钮级）；menu_key 对应侧栏项，便于按页面做配置总览 ----
INSERT INTO admin_permissions (perm_key, label, category, menu_key, status) VALUES
  ('admin.auth.logout', '退出登录', 'auth', 'menu.system', 1),
  ('admin.ops.read', '运维监控-读取', 'ops', 'menu.ops', 1),

  ('admin.channels.read', '通道管理-读取', 'channels', 'menu.channels', 1),
  ('admin.channels.write', '通道管理-写入', 'channels', 'menu.channels', 1),

  ('admin.merchants.read', '商户管理-读取', 'merchants', 'menu.merchants', 1),
  ('admin.merchants.write', '商户管理-写入', 'merchants', 'menu.merchants', 1),
  ('admin.merchants.transfer', '商户划转', 'merchants', 'menu.merchants', 1),

  ('admin.payin_products.read', '代收产品-读取', 'products', 'menu.merchant_payin_products', 1),
  ('admin.payin_products.write', '代收产品-写入', 'products', 'menu.merchant_payin_products', 1),
  ('admin.payout_products.read', '代付产品-读取', 'products', 'menu.merchant_payout_products', 1),
  ('admin.payout_products.write', '代付产品-写入', 'products', 'menu.merchant_payout_products', 1),

  ('admin.routing.read', '路由策略-读取', 'routing', 'menu.routing', 1),
  ('admin.stats.read', '系统概览-读取', 'stats', 'menu.stats', 1),

  ('admin.orders.read', '订单-读取（代收/代付订单列表）', 'orders', '', 1),
  ('admin.orders.mock', '订单-模拟打款成功', 'orders', 'menu.payout_orders', 1),
  ('admin.refunds.read', '退款与差错-读取', 'refunds', 'menu.refunds', 1),
  ('admin.reconcile.read', '对账-读取', 'reconcile', 'menu.reconcile', 1),
  ('admin.settlement.read', '结算-读取', 'settlement', 'menu.settlement', 1),

  ('admin.system.read_settings', '系统管理-展示配置读取', 'system', 'menu.system', 1),
  ('admin.system.write_settings', '系统管理-展示配置写入', 'system', 'menu.system', 1),

  ('admin.admin_users.manage', '后台用户-查看列表与分配角色', 'admin_users', 'menu.rbac_admin_users', 1),

  ('admin.rbac.my_menu', 'RBAC-读取我的菜单', 'rbac', 'menu.rbac_overview', 1),
  ('admin.rbac.manage', 'RBAC-配置管理（菜单/角色/权限/接口）', 'rbac', 'menu.rbac_roles', 1)
ON DUPLICATE KEY UPDATE
  label = VALUES(label),
  category = VALUES(category),
  menu_key = VALUES(menu_key),
  status = VALUES(status);

-- 默认超级管理员拥有全部权限点
INSERT INTO admin_role_permissions (role_id, perm_id)
SELECT ar.id, ap.id
FROM admin_roles ar
JOIN admin_permissions ap
WHERE ar.code = 'super_admin'
ON DUPLICATE KEY UPDATE perm_id = VALUES(perm_id);

-- ---- 接口规则：把 admin 接口映射到权限点（可在后台界面维护）----
INSERT INTO admin_api_rules (method, path_pattern, perm_key, status, remark) VALUES
  ('POST', '/v1/admin/logout', 'admin.auth.logout', 1, ''),

  ('GET', '/v1/admin/ops/services', 'admin.ops.read', 1, ''),

  ('GET', '/v1/admin/channels', 'admin.channels.read', 1, ''),
  ('POST', '/v1/admin/channels', 'admin.channels.write', 1, ''),
  ('PUT', '/v1/admin/channels/:id', 'admin.channels.write', 1, ''),

  ('GET', '/v1/admin/merchants', 'admin.merchants.read', 1, ''),
  ('POST', '/v1/admin/merchants', 'admin.merchants.write', 1, ''),
  ('PUT', '/v1/admin/merchants/:merchant_id', 'admin.merchants.write', 1, ''),
  ('POST', '/v1/admin/merchants/:merchant_id/transfer_payin_to_payout', 'admin.merchants.transfer', 1, ''),

  ('GET', '/v1/admin/payin_products', 'admin.payin_products.read', 1, ''),
  ('POST', '/v1/admin/payin_products', 'admin.payin_products.write', 1, ''),
  ('PUT', '/v1/admin/payin_products/:id', 'admin.payin_products.write', 1, ''),
  ('GET', '/v1/admin/payin_products/:id/bindings', 'admin.payin_products.read', 1, ''),
  ('POST', '/v1/admin/payin_products/:id/bindings', 'admin.payin_products.write', 1, ''),
  ('PUT', '/v1/admin/payin_product_bindings/:id', 'admin.payin_products.write', 1, ''),
  ('DELETE', '/v1/admin/payin_product_bindings/:id', 'admin.payin_products.write', 1, ''),

  ('GET', '/v1/admin/payout_products', 'admin.payout_products.read', 1, ''),
  ('POST', '/v1/admin/payout_products', 'admin.payout_products.write', 1, ''),
  ('PUT', '/v1/admin/payout_products/:id', 'admin.payout_products.write', 1, ''),
  ('GET', '/v1/admin/payout_products/:id/bindings', 'admin.payout_products.read', 1, ''),
  ('POST', '/v1/admin/payout_products/:id/bindings', 'admin.payout_products.write', 1, ''),
  ('PUT', '/v1/admin/payout_product_bindings/:id', 'admin.payout_products.write', 1, ''),
  ('DELETE', '/v1/admin/payout_product_bindings/:id', 'admin.payout_products.write', 1, ''),

  ('GET', '/v1/admin/routing/summary', 'admin.routing.read', 1, ''),
  ('GET', '/v1/admin/stats/overview', 'admin.stats.read', 1, ''),
  ('GET', '/v1/admin/payin_orders', 'admin.orders.read', 1, ''),
  ('GET', '/v1/admin/payout_orders', 'admin.orders.read', 1, ''),
  ('POST', '/v1/admin/payout_orders/:order_no/mock_success', 'admin.orders.mock', 1, ''),
  ('GET', '/v1/admin/refunds', 'admin.refunds.read', 1, ''),
  ('GET', '/v1/admin/reconcile/day', 'admin.reconcile.read', 1, ''),
  ('GET', '/v1/admin/settlement/logs', 'admin.settlement.read', 1, ''),

  ('GET', '/v1/admin/admin_users', 'admin.admin_users.manage', 1, ''),
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
  ('DELETE', '/v1/admin/rbac/api_rules/:id', 'admin.rbac.manage', 1, '')
ON DUPLICATE KEY UPDATE
  perm_key = VALUES(perm_key),
  status = VALUES(status),
  remark = VALUES(remark);

-- 幂等兜底：确保 demo 管理员绑定 super_admin 并拥有全部权限点（避免部分语句未执行或历史数据导致无角色/无权限）
INSERT INTO admin_user_roles (admin_user_id, role_id)
SELECT u.id, r.id
FROM admin_users u
CROSS JOIN admin_roles r
WHERE u.username = 'admin' AND r.code = 'super_admin' AND r.status = 1
ON DUPLICATE KEY UPDATE role_id = VALUES(role_id);

INSERT INTO admin_role_permissions (role_id, perm_id)
SELECT ar.id, ap.id
FROM admin_roles ar
JOIN admin_permissions ap
WHERE ar.code = 'super_admin'
ON DUPLICATE KEY UPDATE perm_id = VALUES(perm_id);

INSERT INTO global_settings (setting_key, setting_value) VALUES
  ('country_code', 'CN'),
  ('currency_code', 'CNY'),
  ('currency_symbol', '¥')
ON DUPLICATE KEY UPDATE setting_value = VALUES(setting_value);

INSERT INTO payin_orders (
  order_no, merchant_id, merchant_order_no, amount, currency, status, channel_id,
  payin_product_id, payin_product_code, channel_locked, paid_amount, fee_mode, fee_rate_bps, fee_fixed_amount, fee_amount, net_amount,
  return_url, notify_url, upstream_trade_no
)
SELECT
  'C-DEMO-001', 'm_demo', 'MO-C-DEMO-001', 1000, 'CNY', 1, c.id, pp.id, pp.code, 0, 1000, 1, 60, 0, 6, 994,
  '', '', 'UP-C-DEMO-001'
FROM channels c
JOIN payin_products pp ON pp.code = 'mock'
WHERE c.name = 'mock-channel'
ON DUPLICATE KEY UPDATE status = VALUES(status), paid_amount = VALUES(paid_amount), fee_amount = VALUES(fee_amount), net_amount = VALUES(net_amount);

-- 与 C-DEMO-001 一致：已支付代收应对应一笔 ORDER_PAID（入账金额 = net_amount），否则不变量检查会报错
INSERT INTO fund_logs (merchant_id, order_no, change_type, amount, balance_before, balance_after, reason, created_at)
VALUES ('m_demo', 'C-DEMO-001', 'ORDER_PAID', 994, 999006, 100000, 'ORDER_PAID', NOW())
ON DUPLICATE KEY UPDATE
  amount = VALUES(amount),
  balance_before = VALUES(balance_before),
  balance_after = VALUES(balance_after);

INSERT INTO payout_orders (
  order_no, merchant_id, merchant_order_no, amount, currency, status, channel_id,
  payout_product_id, payout_product_code, paid_amount, fee_mode, fee_rate_bps, fee_fixed_amount, fee_amount, net_amount,
  notify_url, upstream_trade_no
)
SELECT
  'P-DEMO-001', 'm_demo', 'MO-P-DEMO-001', 2000, 'CNY', 1, c.id, pp.id, pp.code, 2000, 3, 40, 60, 68, 1932,
  '', 'UP-P-DEMO-001'
FROM channels c
JOIN payout_products pp ON pp.code = 'bank_card'
WHERE c.name = 'mock-channel-b'
ON DUPLICATE KEY UPDATE status = VALUES(status), paid_amount = VALUES(paid_amount), fee_amount = VALUES(fee_amount), net_amount = VALUES(net_amount);
