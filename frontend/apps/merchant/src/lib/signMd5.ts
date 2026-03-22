import md5 from 'blueimp-md5'

/** 商户开放平台签名：参数名小写排序，拼接 key=secret 后 MD5 */
export function md5Sign(params: Record<string, string>, secret: string): string {
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
