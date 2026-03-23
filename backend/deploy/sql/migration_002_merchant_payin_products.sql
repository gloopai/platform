-- 商户可用支付产品白名单；未配置任何行时商户在收银台无可用支付方式（需总后台分配）。
CREATE TABLE IF NOT EXISTS merchant_payin_products (
  merchant_id VARCHAR(64) NOT NULL,
  payin_product_id BIGINT UNSIGNED NOT NULL,
  enabled TINYINT NOT NULL DEFAULT 1,
  sort_order INT NOT NULL DEFAULT 0,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (merchant_id, payin_product_id),
  KEY idx_merchant (merchant_id, enabled),
  CONSTRAINT fk_mpp_product FOREIGN KEY (payin_product_id) REFERENCES payin_products (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

ALTER TABLE orders
  ADD COLUMN channel_locked TINYINT NOT NULL DEFAULT 0 COMMENT '1=商户下单已指定通道，收银台不可改支付方式' AFTER payin_product_code;
