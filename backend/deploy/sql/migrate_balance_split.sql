-- 余额拆分迁移：将 merchants.balance 拆为 collect_balance / payout_balance
ALTER TABLE merchants
  ADD COLUMN IF NOT EXISTS collect_balance BIGINT NOT NULL DEFAULT 0 AFTER balance,
  ADD COLUMN IF NOT EXISTS payout_balance BIGINT NOT NULL DEFAULT 0 AFTER collect_balance;

-- 历史余额默认归到代收余额
UPDATE merchants
SET collect_balance = balance
WHERE collect_balance = 0;
