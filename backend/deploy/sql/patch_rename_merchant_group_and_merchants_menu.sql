-- 侧栏分组「商户与接入」→「商户管理」；子菜单「商户管理」→「商户列表」
UPDATE admin_menus SET label = '商户管理' WHERE menu_key = 'group.merchant';
UPDATE admin_menus SET label = '商户列表' WHERE menu_key = 'menu.merchants';
