import md5 from 'blueimp-md5'

export type MerchantAuth = {
  merchantId: string
  apiSecret: string
}

export type MerchantSession = {
  token: string
  expiresAt: number
  merchantId: string
}

export function loadMerchantAuth(): MerchantAuth {
  return {
    merchantId: localStorage.getItem('merchant_id') || 'm_demo',
    apiSecret: localStorage.getItem('merchant_secret') || 'demo_secret',
  }
}

export function saveMerchantAuth(auth: MerchantAuth) {
  localStorage.setItem('merchant_id', auth.merchantId)
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

/** 可选：后台将来下发或本地覆盖的展示名（与 merchant_id 对应会话） */
const MERCHANT_DISPLAY_NAME_KEY = 'merchant_display_name'

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

/** 头像字母：中文取一字，英文取首字母 */
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

function md5Sign(params: Record<string, string>, secret: string): string {
  const keys = Object.keys(params)
    .map((k) => k.toLowerCase())
    .filter((k) => k !== 'sign')
    .sort()
  const parts: string[] = []
  for (const k of keys) {
    const v = params[k]
    if (!v) continue
    parts.push(`${k}=${v}`)
  }
  parts.push(`key=${secret}`)
  return md5(parts.join('&'))
}

export async function signedGet<T>(path: string, params: Record<string, string | number | boolean | undefined>): Promise<T> {
  const auth = loadMerchantAuth()
  const p: Record<string, string> = { merchant_id: auth.merchantId }
  for (const [k, v] of Object.entries(params)) {
    if (v === undefined || v === null) continue
    p[k] = String(v)
  }
  const sign = md5Sign(p, auth.apiSecret)
  const qs = new URLSearchParams({ ...p, sign })
  const resp = await fetch(`${path}?${qs.toString()}`)
  if (!resp.ok) {
    throw new Error(String(resp.status))
  }
  return (await resp.json()) as T
}

export async function signedPost<T>(
  path: string,
  body: Record<string, string | number | boolean | undefined>,
): Promise<T> {
  const auth = loadMerchantAuth()
  const p: Record<string, string> = { merchant_id: auth.merchantId }
  for (const [k, v] of Object.entries(body)) {
    if (v === undefined || v === null) continue
    p[k] = String(v)
  }
  const sign = md5Sign(p, auth.apiSecret)
  const resp = await fetch(path, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ ...p, sign }),
  })
  if (!resp.ok) {
    throw new Error(String(resp.status))
  }
  return (await resp.json()) as T
}

export async function merchantConsoleGet<T>(path: string, params?: Record<string, string | number | boolean | undefined>): Promise<T> {
  const tok = loadMerchantToken()
  const qs = new URLSearchParams()
  if (params) {
    for (const [k, v] of Object.entries(params)) {
      if (v === undefined || v === null) continue
      qs.set(k, String(v))
    }
  }
  const url = qs.toString() ? `${path}?${qs.toString()}` : path
  const resp = await fetch(url, { headers: { 'X-Merchant-Token': tok } })
  if (!resp.ok) throw new Error(String(resp.status))
  return (await resp.json()) as T
}

export async function merchantConsolePost<T>(path: string, body?: Record<string, unknown>): Promise<T> {
  const tok = loadMerchantToken()
  const resp = await fetch(path, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', 'X-Merchant-Token': tok },
    body: JSON.stringify(body || {}),
  })
  if (!resp.ok) throw new Error(String(resp.status))
  return (await resp.json()) as T
}
