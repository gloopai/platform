-- 系统配置：新建商户自动数字 merchant_id 的起始值（含），与 Core 取号 floor 一致
INSERT INTO global_settings (setting_key, setting_value) VALUES ('merchant_numeric_id_start', '5000000000')
ON DUPLICATE KEY UPDATE setting_value = VALUES(setting_value);
