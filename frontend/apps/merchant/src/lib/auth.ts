export type MerchantAuth = {
  appId: string
  apiSecret: string
}

export type MerchantSession = {
  token: string
  expiresAt: number
  merchantId: string
}

const MERCHANT_DISPLAY_NAME_KEY = 'merchant_display_name'

export function loadMerchantAuth(): MerchantAuth {
  return {
    appId: localStorage.getItem('merchant_app_id') || 'app_demo',
    apiSecret: localStorage.getItem('merchant_secret') || 'demo_secret',
  }
}

export function saveMerchantAuth(auth: MerchantAuth) {
  localStorage.setItem('merchant_app_id', auth.appId)
  localStorage.setItem('merchant_secret', auth.apiSecret)
}

export function loadMerchantToken(): string {
  return localStorage.getItem('merchant_token') || ''
}

export function saveMerchantSession(sess: MerchantSession) {
  localStorage.setItem('merchant_token', sess.token)
  localStorage.setItem('merchant_token_expires_at', String(sess.expiresAt))
  localStorage.setItem('merchant_id', sess.merchantId)
}

export function resolveMerchantDisplayName(merchantId: string): string {
  const custom = localStorage.getItem(MERCHANT_DISPLAY_NAME_KEY)
  if (custom?.trim()) return custom.trim()
  const id = merchantId.trim()
  const known: Record<string, string> = {
    m_demo: '演示商户',
  }
  if (id && known[id]) return known[id]
  return id ? `商户 · ${id}` : '商户'
}

export function merchantMonogram(displayName: string, merchantId: string): string {
  const s = (displayName.trim() || merchantId || '?').trim()
  if (!s) return '?'
  const first = s[0]!
  if (/[\u4e00-\u9fff]/.test(first)) return first
  return first.toUpperCase()
}

export function clearMerchantSession() {
  localStorage.removeItem('merchant_token')
  localStorage.removeItem('merchant_token_expires_at')
  localStorage.removeItem(MERCHANT_DISPLAY_NAME_KEY)
}
