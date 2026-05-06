import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import type { RouteLocationNormalizedLoaded } from 'vue-router'

import router from '@/router'

export interface TabItem {
  path: string
  fullPath: string
  title: string
  /** 与路由 name、页面 defineOptions.name 一致，供 keep-alive include */
  name: string
}

function tabFromRoute(route: RouteLocationNormalizedLoaded): TabItem | null {
  if (route.meta?.hideInTabs) return null
  const name = route.name
  if (name == null || typeof name !== 'string') return null
  const title = (route.meta?.title as string) || name
  return {
    path: route.path,
    fullPath: route.fullPath,
    title,
    name,
  }
}

function pruneRefreshSeqForVisited(
  seq: Record<string, number>,
  visitedPaths: Set<string>,
): Record<string, number> {
  const next: Record<string, number> = {}
  for (const [k, v] of Object.entries(seq)) {
    if (visitedPaths.has(k)) next[k] = v
  }
  return next
}

export const useTabsStore = defineStore('tabs', () => {
  const visited = ref<TabItem[]>([])
  /** 按 path 累加，配合 AdminLayout 中 router-view 的 key 强制刷新 keep-alive 页 */
  const pathRefreshSeq = ref<Record<string, number>>({})

  const cachedComponentNames = computed(() => [...new Set(visited.value.map((t) => t.name))])

  function syncRefreshSeq() {
    const paths = new Set(visited.value.map((t) => t.path))
    pathRefreshSeq.value = pruneRefreshSeqForVisited(pathRefreshSeq.value, paths)
  }

  /** 供 keep-alive 子组件 key，刷新后同 path 会变化以触发重挂载 */
  function viewKey(fullPath: string, path: string): string {
    const n = pathRefreshSeq.value[path] ?? 0
    return n ? `${fullPath}__r${n}` : fullPath
  }

  function refreshCurrentTab() {
    const p = router.currentRoute.value.path
    pathRefreshSeq.value = { ...pathRefreshSeq.value, [p]: (pathRefreshSeq.value[p] || 0) + 1 }
  }

  /** 仅保留首页一条页签并跳转（与「关闭全部」一致） */
  function goHomeAsSingleTab() {
    const r = router.resolve({ path: '/home' })
    const leaf = r.matched[r.matched.length - 1]
    const name = leaf?.name
    if (typeof name !== 'string') {
      visited.value = []
      pathRefreshSeq.value = {}
      void router.push('/home')
      return
    }
    const title = (leaf.meta?.title as string) || name
    visited.value = [
      {
        path: r.path,
        fullPath: r.fullPath,
        title,
        name,
      },
    ]
    pathRefreshSeq.value = pruneRefreshSeqForVisited(pathRefreshSeq.value, new Set([r.path]))
    void router.push(r.fullPath)
  }

  function addFromRoute(route: RouteLocationNormalizedLoaded) {
    const tab = tabFromRoute(route)
    if (!tab) return
    const i = visited.value.findIndex((t) => t.path === tab.path)
    if (i >= 0) {
      visited.value[i] = { ...visited.value[i], fullPath: tab.fullPath, title: tab.title }
      return
    }
    visited.value.push(tab)
  }

  function closeTab(path: string) {
    const i = visited.value.findIndex((t) => t.path === path)
    if (i < 0) return
    const currentPath = router.currentRoute.value.path
    visited.value.splice(i, 1)
    syncRefreshSeq()
    if (visited.value.length === 0) {
      goHomeAsSingleTab()
      return
    }
    if (currentPath === path) {
      const next = visited.value[i] ?? visited.value[i - 1]
      if (next) void router.push(next.fullPath)
    }
  }

  function closeOthers(path: string) {
    const keep = visited.value.find((t) => t.path === path)
    if (!keep) return
    visited.value = [keep]
    syncRefreshSeq()
    if (router.currentRoute.value.path !== path) void router.push(keep.fullPath)
  }

  /** 关闭指定页签左侧所有页签（不含自身） */
  function closeLeft(path: string) {
    const i = visited.value.findIndex((t) => t.path === path)
    if (i <= 0) return
    const currentPath = router.currentRoute.value.path
    visited.value.splice(0, i)
    syncRefreshSeq()
    if (!visited.value.some((t) => t.path === currentPath)) {
      const next = visited.value[0]
      if (next) void router.push(next.fullPath)
    }
  }

  /** 关闭指定页签右侧所有页签（不含自身） */
  function closeRight(path: string) {
    const i = visited.value.findIndex((t) => t.path === path)
    if (i < 0 || i >= visited.value.length - 1) return
    const currentPath = router.currentRoute.value.path
    visited.value.splice(i + 1)
    syncRefreshSeq()
    if (!visited.value.some((t) => t.path === currentPath)) {
      const next = visited.value[visited.value.length - 1]
      if (next) void router.push(next.fullPath)
    }
  }

  function closeAll() {
    goHomeAsSingleTab()
  }

  function clear() {
    visited.value = []
    pathRefreshSeq.value = {}
  }

  return {
    visited,
    cachedComponentNames,
    viewKey,
    refreshCurrentTab,
    addFromRoute,
    closeTab,
    closeOthers,
    closeLeft,
    closeRight,
    closeAll,
    clear,
  }
})
