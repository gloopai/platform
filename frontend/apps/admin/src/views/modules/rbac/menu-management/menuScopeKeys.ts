import type { RbacAdminMenu } from './types'

/** 当前菜单项关联的 menu_key 集合：页面为自身；分组为自身 + 所有子孙页面。 */
export function collectScopeMenuKeys(m: RbacAdminMenu, all: RbacAdminMenu[]): string[] {
  const keys = new Set<string>()
  const mk = (m.menu_key || '').trim()
  if (mk) keys.add(mk)

  function walkLeaves(parentId: number) {
    for (const c of all.filter((x) => x.parent_id === parentId)) {
      if (c.kind === 1) {
        const k = (c.menu_key || '').trim()
        if (k) keys.add(k)
      } else if (c.kind === 2) {
        walkLeaves(c.id)
      }
    }
  }
  if (m.kind === 2) walkLeaves(m.id)
  return [...keys].filter(Boolean)
}
