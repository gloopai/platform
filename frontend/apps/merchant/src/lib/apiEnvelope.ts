/** 与网关约定一致：{ code, message, data }；成功 code=2000 */
export const API_CODE_SUCCESS = 2000

export type ApiEnvelope<T = unknown> = {
  code: number
  message?: string
  data?: T
}

export function parseApiEnvelope<T>(bodyText: string): ApiEnvelope<T> {
  let j: unknown
  try {
    j = bodyText.trim() ? JSON.parse(bodyText) : {}
  } catch {
    throw new Error('响应不是合法 JSON')
  }
  const o = j as ApiEnvelope<T>
  if (!o || typeof o.code !== 'number') {
    throw new Error('响应格式无效')
  }
  return o
}

export function unwrapApiData<T>(env: ApiEnvelope<T>): T {
  if (env.code !== API_CODE_SUCCESS) {
    throw new Error(env.message?.trim() || `错误码 ${env.code}`)
  }
  return (env.data ?? {}) as T
}

export function apiEnvelopeErrorMessage(env: ApiEnvelope): string {
  return env.message?.trim() || `错误码 ${env.code}`
}
