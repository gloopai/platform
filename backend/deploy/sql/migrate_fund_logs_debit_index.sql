-- 已有库增量：加速 DebitPayout 幂等查询（merchant_id + order_no + change_type）。
-- 若索引已存在会报错，可忽略 Duplicate key name。
ALTER TABLE fund_logs ADD INDEX idx_fund_logs_merchant_order_type (merchant_id, order_no, change_type);
