-- 10 位数字 merchant_id 自增序列（与 schema 一致；存量库补表并尽量与已有纯数字 merchant_id 对齐）
CREATE TABLE IF NOT EXISTS merchant_numeric_seq (
  slot TINYINT UNSIGNED NOT NULL,
  next_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  PRIMARY KEY (slot)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO merchant_numeric_seq (slot, next_id) VALUES (1, 0)
ON DUPLICATE KEY UPDATE next_id = merchant_numeric_seq.next_id;

SET @max_numeric_mid := (
  SELECT COALESCE(MAX(CAST(merchant_id AS UNSIGNED)), 0)
  FROM merchants
  WHERE merchant_id REGEXP '^[0-9]{1,10}$'
);

UPDATE merchant_numeric_seq
SET next_id = GREATEST(next_id, @max_numeric_mid)
WHERE slot = 1;
