-- 余额拆分迁移：将 merchants.balance 拆为 payin_balance / available_balance
ALTER TABLE merchants
  ADD COLUMN IF NOT EXISTS payin_balance BIGINT NOT NULL DEFAULT 0 AFTER balance,
  ADD COLUMN IF NOT EXISTS available_balance BIGINT NOT NULL DEFAULT 0 AFTER payin_balance;

-- 历史余额默认归到代收余额
UPDATE merchants
SET payin_balance = balance
WHERE payin_balance = 0;
