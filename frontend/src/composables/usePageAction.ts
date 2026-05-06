import { useRoute } from 'vue-router'

import { usePermissionStore } from '@/stores/permission'
import { useSidebarMenuStore } from '@/stores/sidebarMenu'

export function usePageAction() {
  const route = useRoute()
  const perm = usePermissionStore()
  const sidebar = useSidebarMenuStore()

  function pageAction(code: string): boolean {
    if (!perm.canUseAction(code)) return false
    const menu = sidebar.menuForRoute(route.path)
    if (menu?.enabled === false) return false
    if (!menu?.children?.length) return true
    const buttons = menu.children.filter((c) => c.type === 'button' && c.enabled !== false)
    if (!buttons.length) return true
    return buttons.some((b) => b.permissionCode === code)
  }

  return { pageAction }
}
