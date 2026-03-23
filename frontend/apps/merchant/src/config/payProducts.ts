/**
 * 开放下单参数 `payin_type`（即支付产品编码），与库表 `payin_products.code`、网关文档一致。
 * 演示环境 seed 见 `backend/deploy/sql/seed_demo.sql`；联调页下拉默认值应与此保持同步。
 */
export const DEMO_PAY_PRODUCT_OPTIONS = [
  { code: 'mock', label: 'Mock支付' },
  { code: 'wechat', label: '微信支付' },
  { code: 'alipay', label: '支付宝' },
] as const
