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

export function clearMerchantSession() {
  localStorage.removeItem('merchant_token')
  localStorage.removeItem('merchant_token_expires_at')
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
