-- 菜单展示位置：侧栏 / 头像下拉

ALTER TABLE admin_menus
  ADD COLUMN placement VARCHAR(16) NOT NULL DEFAULT 'left' COMMENT 'left=左侧导航 avatar=头像下拉' AFTER sort_order;
