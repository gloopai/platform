-- 将演示环境双 mock 通道合并为单一 mock_psp（与 channeldriver/mockpsp 对齐）
-- 已存在库可手工执行；新建库请直接用 seed_demo.sql

UPDATE channels
SET name = 'mock-psp',
    payin_type = 'mock_psp',
    upstream_merchant_no = 'mock_app_id'
WHERE name IN ('mock-channel', 'mock-channel-b');
