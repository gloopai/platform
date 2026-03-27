-- 移除商户表默认费率字段（费率仅通过商户-产品授权行配置）
ALTER TABLE merchants
  DROP COLUMN default_payin_rate_bps,
  DROP COLUMN default_payout_rate_bps;
