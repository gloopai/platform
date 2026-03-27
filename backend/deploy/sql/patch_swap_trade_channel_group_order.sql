-- 侧栏顶层分组：「交易与资金」排在「产品与通道」之前（与 seed_demo 一致）
UPDATE admin_menus SET sort_order = 30 WHERE menu_key = 'group.trade';
UPDATE admin_menus SET sort_order = 40 WHERE menu_key = 'group.channel';
