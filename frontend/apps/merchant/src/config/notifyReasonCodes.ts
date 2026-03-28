export type NotifyReasonRow = {
  code: string
  meaning: string
  action: string
}

/**
 * 回调 reason_code 对照（MVP）。
 * 与 docs/开放API错误码.md、docs/端到端联调一遍.md 保持一致。
 */
export const NOTIFY_REASON_ROWS: NotifyReasonRow[] = [
  { code: 'INVALID_NOTIFY_PARAMS', meaning: '回调参数缺失/非法', action: '检查 order_no/paid_amount/upstream_trade_no 等' },
  { code: 'CHANNEL_NOT_FOUND', meaning: '（平台侧）上游验签配置异常', action: '属平台与 PSP 联调问题，非商户集成参数' },
  { code: 'INVALID_SIGN', meaning: '回调签名错误', action: '校验签名算法与 channel_sign_secret' },
  { code: 'ORDER_NOT_FOUND', meaning: '平台订单不存在', action: '校验 order_no 是否平台单号' },
  { code: 'ORDER_NOT_PENDING', meaning: '订单非待支付状态', action: '仅待支付订单可置成功' },
  { code: 'REPLAY_PAYLOAD_MISMATCH', meaning: '已支付但重放快照不一致', action: '确保重复通知参数与首笔一致' },
  { code: 'MARK_PAID_FAILED', meaning: '落支付状态失败', action: '检查网关/trade 日志' },
  { code: 'MARK_PAID_RACE', meaning: '并发竞争，读取最终态失败', action: '短暂重试并查询订单最终状态' },
  { code: 'MARK_PAID_RACE_MISMATCH', meaning: '并发竞争且快照不一致', action: '以平台最终订单快照为准排查' },
  { code: 'IDEMPOTENT_REPLAY_ACCEPTED', meaning: '已支付同快照重放被接受', action: '可视为成功，无需补偿' },
  { code: 'IDEMPOTENT_RACE_ACCEPTED', meaning: '并发竞争后同快照被接受', action: '可视为成功，无需补偿' },
]
