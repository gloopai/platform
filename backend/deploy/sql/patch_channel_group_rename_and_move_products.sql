-- 分组「通道与路由」更名为「产品与通道」，并将「代收产品 / 代付产品」从「商户与接入」移入该分组
SET @gid_channel := (SELECT id FROM admin_menus WHERE menu_key = 'group.channel' LIMIT 1);

UPDATE admin_menus SET label = '产品与通道' WHERE menu_key = 'group.channel';

UPDATE admin_menus
SET parent_id = @gid_channel, sort_order = 10
WHERE menu_key = 'menu.merchant_payin_products' AND @gid_channel IS NOT NULL;

UPDATE admin_menus
SET parent_id = @gid_channel, sort_order = 20
WHERE menu_key = 'menu.merchant_payout_products' AND @gid_channel IS NOT NULL;

UPDATE admin_menus SET sort_order = 30 WHERE menu_key = 'menu.channels';
UPDATE admin_menus SET sort_order = 40 WHERE menu_key = 'menu.routing';
UPDATE admin_menus SET sort_order = 50 WHERE menu_key = 'menu.channel_health';
