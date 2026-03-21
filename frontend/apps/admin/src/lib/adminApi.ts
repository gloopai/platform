export type AdminSession = {
  token: string
  expiresAt: number
}

export function loadAdminToken(): string {
  return localStorage.getItem('admin_token') || ''
}

export function saveAdminSession(sess: AdminSession) {
  localStorage.setItem('admin_token', sess.token)
  localStorage.setItem('admin_token_expires_at', String(sess.expiresAt))
}

export function clearAdminSession() {
  localStorage.removeItem('admin_token')
  localStorage.removeItem('admin_token_expires_at')
}

export async function adminGet<T>(path: string): Promise<T> {
  const tok = loadAdminToken()
  const resp = await fetch(path, { headers: { 'X-Admin-Token': tok } })
  if (!resp.ok) throw new Error(String(resp.status))
  return (await resp.json()) as T
}

export async function adminPost<T>(path: string, body?: Record<string, unknown>): Promise<T> {
  const tok = loadAdminToken()
  const resp = await fetch(path, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', 'X-Admin-Token': tok },
    body: JSON.stringify(body || {}),
  })
  if (!resp.ok) throw new Error(String(resp.status))
  return (await resp.json()) as T
}

export async function adminPut<T>(path: string, body?: Record<string, unknown>): Promise<T> {
  const tok = loadAdminToken()
  const resp = await fetch(path, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json', 'X-Admin-Token': tok },
    body: JSON.stringify(body || {}),
  })
  if (!resp.ok) throw new Error(String(resp.status))
  return (await resp.json()) as T
}

