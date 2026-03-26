export type RbacAdminMenu = {
  id: number
  parent_id: number
  menu_key: string
  label: string
  icon: string
  kind: number
  path: string
  sort_order: number
  placement?: string
}

export type AdminPermission = {
  id: number
  perm_key: string
  label: string
  category: string
  menu_key: string
  status: number
}

export type ApiRule = {
  id: number
  method: string
  path_pattern: string
  perm_key: string
  status: number
  remark: string
}

export type MenuMgmtTab = 'left' | 'avatar' | 'other'
