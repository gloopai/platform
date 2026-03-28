/** 与后端 common/notify Envelope JSON 对齐 */
export type PortalNotifyEnvelope = {
  event: string
  portal: string
  id: string
  title: string
  body: string
  severity: string
  link_path: string
  link_query_json?: string
  meta_json?: string
}

export type PortalNotifyListItem = {
  id: string
  title: string
  body: string
  severity: string
  link_path: string
  link_query_json?: string
  at: number
}
