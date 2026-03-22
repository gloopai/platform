/**
 * 占位页文案：路径 -> 展示内容（与 router 中占位路由一致）。
 */
export type MerchantPlaceholderBlock = {
  summary: string
  bullets: string[]
  apiNote?: string
}

export const merchantPlaceholderByPath: Record<string, MerchantPlaceholderBlock> = {
  '/products': {
    summary: '查看平台为您开通的支付方式、费率档位与结算周期；支持下载对账说明与费率确认函（能力随平台开通）。',
    bullets: [
      '支付方式：微信、支付宝、银联等开关与限额说明',
      '费率：按 MDR / 笔数展示，与签约合同一致',
      '结算：T+N、D+1 等周期与提现规则说明',
    ],
    apiNote: '待接入：GET /v1/merchant/products 或商户签约查询接口',
  },
  '/account': {
    summary: '管理登录安全、联系方式与基础资料；敏感操作建议开启二次验证（若平台提供）。',
    bullets: [
      '登录与密钥：与「开发配置」中 api_secret 说明联动',
      'IP 白名单、回调域名：以服务端配置为准',
      '子账号与角色（若平台提供商户多用户）',
    ],
    apiNote: '待接入：GET/PUT /v1/merchant/profile、安全设置类接口',
  },
}
