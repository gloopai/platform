-- 已有库追加：上游对接整段 JSON 文本（与 schema.sql 对齐）
ALTER TABLE channels
  ADD COLUMN upstream_config TEXT NULL COMMENT '上游对接自由 JSON 文本（管理台整段保存）' AFTER sign_secret;
