-- 已有库增量：商户/产品自由 JSON 与 Consul 双写（与 schema.sql 对齐）。
-- 若列已存在会报错，可逐条执行并忽略 Duplicate column。
ALTER TABLE merchants ADD COLUMN merchant_config TEXT NULL COMMENT '商户自由 JSON（与 Consul 双写）' AFTER return_url;
ALTER TABLE payin_products ADD COLUMN product_config TEXT NULL COMMENT '产品自由 JSON（与 Consul 双写）' AFTER enabled;
ALTER TABLE payout_products ADD COLUMN product_config TEXT NULL COMMENT '产品自由 JSON（与 Consul 双写）' AFTER enabled;
