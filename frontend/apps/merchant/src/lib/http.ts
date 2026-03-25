/**
 * HTTP 传输层：控制台 Token 请求 + 开放签名的 GET/POST。
 * 业务语义请放在 ../api/*.ts，避免页面直接拼路径。
 */
import { loadMerchantAuth, loadMerchantToken } from './auth'
import { md5Sign } from './signMd5'

function trimTrailingSlash(s: string): string {
  return s.endsWith('/') ? s.slice(0, -1) : s
}

const MERCHANT_API_BASE = trimTrailingSlash(String(import.meta.env.VITE_MERCHANT_API_BASE || '').trim())
const OPENAPI_BASE = trimTrailingSlash(String(import.meta.env.VITE_OPENAPI_BASE || '').trim())

export function merchantConsoleUrl(path: string): string {
  if (!path.startsWith('/')) return path
  if (!MERCHANT_API_BASE) return path
  return `${MERCHANT_API_BASE}${path}`
}

export function openAPIUrl(path: string): string {
  if (!path.startsWith('/')) return path
  if (!OPENAPI_BASE) return path
  return `${OPENAPI_BASE}${path}`
}

export async function merchantConsoleGet<T>(
  path: string,
  params?: Record<string, string | number | boolean | undefined>,
): Promise<T> {
  const tok = loadMerchantToken()
  const qs = new URLSearchParams()
  if (params) {
    for (const [k, v] of Object.entries(params)) {
      if (v === undefined || v === null) continue
      qs.set(k, String(v))
    }
  }
  const url = qs.toString() ? `${path}?${qs.toString()}` : path
  const resp = await fetch(merchantConsoleUrl(url), { headers: { 'X-Merchant-Token': tok } })
  if (!resp.ok) throw new Error(String(resp.status))
  return (await resp.json()) as T
}

export async function merchantConsolePut<T>(path: string, body?: Record<string, unknown>): Promise<T> {
  const tok = loadMerchantToken()
  const resp = await fetch(merchantConsoleUrl(path), {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json', 'X-Merchant-Token': tok },
    body: JSON.stringify(body || {}),
  })
  if (!resp.ok) throw new Error(String(resp.status))
  return (await resp.json()) as T
}

export async function merchantConsolePost<T>(path: string, body?: Record<string, unknown>): Promise<T> {
  const tok = loadMerchantToken()
  const resp = await fetch(merchantConsoleUrl(path), {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', 'X-Merchant-Token': tok },
    body: JSON.stringify(body || {}),
  })
  if (!resp.ok) throw new Error(String(resp.status))
  return (await resp.json()) as T
}

export async function signedGet<T>(
  path: string,
  params: Record<string, string | number | boolean | undefined>,
): Promise<T> {
  const auth = loadMerchantAuth()
  const p: Record<string, string> = { merchant_id: auth.merchantId }
  for (const [k, v] of Object.entries(params)) {
    if (v === undefined || v === null) continue
    p[k] = String(v)
  }
  const sign = md5Sign(p, auth.apiSecret)
  const qs = new URLSearchParams({ ...p, sign })
  const resp = await fetch(openAPIUrl(`${path}?${qs.toString()}`))
  if (!resp.ok) throw new Error(String(resp.status))
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
  const resp = await fetch(openAPIUrl(path), {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ ...p, sign }),
  })
  if (!resp.ok) throw new Error(String(resp.status))
  return (await resp.json()) as T
}
