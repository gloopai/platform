import { parseApiEnvelope, unwrapApiData } from './apiEnvelope'

export type AdminSession = {
  token: string
  expiresAt: number
}

const ADMIN_IDENTITY_KEY = 'admin_identity'

function trimTrailingSlash(s: string): string {
  return s.endsWith('/') ? s.slice(0, -1) : s
}

const ADMIN_API_BASE = trimTrailingSlash(String(import.meta.env.VITE_ADMIN_API_BASE || '').trim())
const ADMIN_SSE_BASE = trimTrailingSlash(String(import.meta.env.VITE_ADMIN_SSE_BASE || '').trim())

export function adminApiUrl(path: string): string {
  if (!path.startsWith('/')) return path
  if (!ADMIN_API_BASE) return path
  return `${ADMIN_API_BASE}${path}`
}

/** 仅用于管理台通知 SSE：可直连 gateway，避免 Vite 代理缓冲流式响应；登录与其它 API 仍用 adminApiUrl（同域代理）。 */
export function adminSseUrl(path: string): string {
  if (!path.startsWith('/')) return path
  if (ADMIN_SSE_BASE) return `${ADMIN_SSE_BASE}${path}`
  return adminApiUrl(path)
}

export function loadAdminToken(): string {
  return localStorage.getItem('admin_token') || ''
}

export function saveAdminSession(sess: AdminSession) {
  localStorage.setItem('admin_token', sess.token)
  localStorage.setItem('admin_token_expires_at', String(sess.expiresAt))
}

export function loadAdminIdentity(): string {
  return (localStorage.getItem(ADMIN_IDENTITY_KEY) || '').trim()
}

export function saveAdminIdentity(identity: string) {
  const v = identity.trim()
  if (!v) {
    localStorage.removeItem(ADMIN_IDENTITY_KEY)
    return
  }
  localStorage.setItem(ADMIN_IDENTITY_KEY, v)
}

export function clearAdminSession() {
  localStorage.removeItem('admin_token')
  localStorage.removeItem('admin_token_expires_at')
  localStorage.removeItem(ADMIN_IDENTITY_KEY)
}

/** 与需登录后台接口统一的请求选项（自动带 X-Admin-Token；JSON body 自动序列化） */
export type AdminRequestOptions = {
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH'
  /** 对象会按 JSON 发送并设置 Content-Type；无需传则不发 body */
  body?: Record<string, unknown> | null
  /** 额外头（会与默认头合并；勿覆盖鉴权键名除非你有意为之） */
  headers?: Record<string, string>
}

/**
 * 所有需携带管理台 Token 的接口统一走此方法，避免各处重复拼 headers。
 * 未登录场景（如 POST /v1/admin/login）请单独 fetch，勿用本方法。
 */
export async function adminRequest<T>(path: string, options: AdminRequestOptions = {}): Promise<T> {
  const method = options.method ?? 'GET'
  const tok = loadAdminToken()
  const headers: Record<string, string> = {
    'X-Admin-Token': tok,
    ...options.headers,
  }

  let body: string | undefined
  if (options.body != null && method !== 'GET') {
    headers['Content-Type'] = 'application/json'
    body = JSON.stringify(options.body)
  }

  const resp = await fetch(adminApiUrl(path), { method, headers, body })
  const text = await resp.text()
  const env = parseApiEnvelope<T>(text)
  try {
    return unwrapApiData(env)
  } catch (e) {
    if (e instanceof Error) throw e
    throw new Error(String(e))
  }
}

export async function adminGet<T>(path: string): Promise<T> {
  return adminRequest<T>(path, { method: 'GET' })
}

export async function adminPost<T>(path: string, body?: Record<string, unknown>): Promise<T> {
  return adminRequest<T>(path, { method: 'POST', body: body ?? {} })
}

export async function adminPut<T>(path: string, body?: Record<string, unknown>): Promise<T> {
  return adminRequest<T>(path, { method: 'PUT', body: body ?? {} })
}

export async function adminDelete<T>(path: string): Promise<T> {
  return adminRequest<T>(path, { method: 'DELETE' })
}
