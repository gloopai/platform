-- Upgrades pre-rename databases: channels.payin_type -> driver_key (PSP driver identifier).
-- Skip if your schema already defines driver_key.

ALTER TABLE channels
  CHANGE COLUMN payin_type driver_key VARCHAR(32) NULL COMMENT 'PSP 驱动标识，与 Registry 注册键一致';

-- Optional: align index name with new column (MySQL 8+)
ALTER TABLE channels RENAME INDEX idx_enabled_payintype TO idx_enabled_driverkey;
