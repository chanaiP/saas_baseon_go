import type { MenuNode } from '@/types/menu'

export type MenuCrumb = { title: string; path?: string }

function pathMatches(routePath: string, menuPath: string) {
  return routePath === menuPath || routePath.startsWith(`${menuPath}/`)
}

/** 根据侧栏菜单树生成面包屑（目录 + 当前菜单） */
export function menuBreadcrumbs(tree: MenuNode[], routePath: string): MenuCrumb[] {
  const out: MenuCrumb[] = []

  function walk(nodes: MenuNode[], ancestors: MenuCrumb[]): boolean {
    for (const n of nodes) {
      if (n.type === 'directory' && n.children?.length) {
        const nextA = [...ancestors, { title: n.title }]
        if (walk(n.children, nextA)) return true
      } else if (n.type === 'menu' && n.path) {
        if (pathMatches(routePath, n.path)) {
          out.push(...ancestors, { title: n.title, path: n.path })
          return true
        }
      }
    }
    return false
  }

  walk(tree, [])
  return out
}
