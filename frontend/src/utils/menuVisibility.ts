import type { MenuNode } from '@/types/menu'
import { menuBreadcrumbs } from '@/utils/menuBreadcrumb'

/** 与侧栏一致：有菜单路由权限，或有任一页面内操作权限，则可见 */
export function isMenuVisibleForSidebar(
  n: MenuNode,
  can: (code: string) => boolean,
  canPath: (path: string) => boolean = can,
): boolean {
  if (n.enabled === false) return false
  if (n.type !== 'menu' || !n.path) return false
  if (canPath(n.path)) return true
  const buttons = (n.children || []).filter((c) => c.type === 'button' && c.enabled !== false)
  return buttons.some((b) => b.permissionCode && can(b.permissionCode))
}

/** 过滤目录与菜单，仅保留当前用户可见的侧栏节点 */
export function filterVisibleMenuTree(
  nodes: MenuNode[],
  can: (code: string) => boolean,
  canPath: (path: string) => boolean = can,
): MenuNode[] {
  const out: MenuNode[] = []
  for (const n of nodes) {
    if (n.enabled === false) continue
    if (n.type === 'directory') {
      const ch = filterVisibleMenuTree(n.children || [], can, canPath)
      if (ch.length) out.push({ ...n, children: ch })
    } else if (n.type === 'menu') {
      if (isMenuVisibleForSidebar(n, can, canPath)) out.push(n)
    }
  }
  return out
}

export type FlatMenuSearchItem = {
  path: string
  title: string
  /** 父级目录路径文案，用于副标题 */
  group: string
}

/**
 * 扁平化可访问菜单项；`treeForCrumb` 用完整平台菜单树生成面包屑（与顶栏面包屑一致）。
 */
export function flattenAccessibleMenuItems(
  visibleTree: MenuNode[],
  treeForCrumb: MenuNode[],
  can: (code: string) => boolean,
  canPath: (path: string) => boolean = can,
): FlatMenuSearchItem[] {
  const filtered = filterVisibleMenuTree(visibleTree, can, canPath)
  const items: FlatMenuSearchItem[] = []

  function walk(nodes: MenuNode[]) {
    for (const n of nodes) {
      if (n.type === 'menu' && n.path) {
        const bc = menuBreadcrumbs(treeForCrumb, n.path)
        const group =
          bc.length > 1 ? bc.slice(0, -1).map((c) => c.title).join(' / ') : ''
        items.push({ path: n.path, title: n.title, group })
      }
      if (n.children?.length) walk(n.children)
    }
  }

  walk(filtered)
  return items
}
