import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

import {
  fetchMenuBundles,
  fetchTenantMenuOverrides,
  saveTenantMenuOverrides,
} from '@/api/permission'
import type { MenuBundle, TenantMenuOverride, TenantMenuOverrideValue } from '@/api/permission'
import type { MenuNode } from '@/types/menu'

const STORAGE_KEY = 'sidebar_menu_config_v2'

function defaultTree(): MenuNode[] {
  return [
    {
      id: 'home',
      type: 'menu',
      title: '首页',
      path: '/home',
      icon: 'House',
      children: [],
    },
    {
      id: 'sys',
      type: 'directory',
      title: '系统管理',
      icon: 'Setting',
      children: [
        {
          id: 'tenant',
          type: 'menu',
          title: '主体管理',
          path: '/tenants',
          icon: 'OfficeBuilding',
          children: [
            { id: 'tenant-c', type: 'button', title: '新增', permissionCode: 'tenant:create', enabled: true },
            { id: 'tenant-e', type: 'button', title: '编辑', permissionCode: 'tenant:edit', enabled: true },
            { id: 'tenant-s', type: 'button', title: '启停', permissionCode: 'tenant:status', enabled: true },
            { id: 'tenant-rp', type: 'button', title: '重置主管理员密码', permissionCode: 'tenant:reset_primary_password', enabled: true },
            { id: 'tenant-d', type: 'button', title: '删除', permissionCode: 'tenant:delete', enabled: true },
          ],
        },
        {
          id: 'plan',
          type: 'menu',
          title: '套餐中心',
          path: '/plans',
          icon: 'Tickets',
          children: [
            { id: 'plan-c', type: 'button', title: '新增', permissionCode: 'plan:create', enabled: true },
            { id: 'plan-e', type: 'button', title: '编辑', permissionCode: 'plan:edit', enabled: true },
            { id: 'plan-conf', type: 'button', title: '配置', permissionCode: 'plan:config', enabled: true },
          ],
        },
        {
          id: 'org',
          type: 'menu',
          title: '组织架构',
          path: '/organization',
          icon: 'Share',
          children: [
            { id: 'org-c', type: 'button', title: '新增', permissionCode: 'org:create', enabled: true },
            { id: 'org-e', type: 'button', title: '编辑', permissionCode: 'org:edit', enabled: true },
            { id: 'org-d', type: 'button', title: '删除', permissionCode: 'org:delete', enabled: true },
          ],
        },
        {
          id: 'pos',
          type: 'menu',
          title: '岗位管理',
          path: '/positions',
          icon: 'Postcard',
          children: [
            { id: 'pos-c', type: 'button', title: '新增', permissionCode: 'pos:create', enabled: true },
            { id: 'pos-e', type: 'button', title: '编辑', permissionCode: 'pos:edit', enabled: true },
            { id: 'pos-d', type: 'button', title: '删除', permissionCode: 'pos:delete', enabled: true },
          ],
        },
        {
          id: 'business-unit',
          type: 'menu',
          title: '业务单元',
          path: '/business-units',
          icon: 'DataAnalysis',
          children: [
            { id: 'business-unit-c', type: 'button', title: '新增', permissionCode: 'business_unit:create', enabled: true },
            { id: 'business-unit-e', type: 'button', title: '编辑', permissionCode: 'business_unit:edit', enabled: true },
            { id: 'business-unit-d', type: 'button', title: '删除', permissionCode: 'business_unit:delete', enabled: true },
          ],
        },
        {
          id: 'user',
          type: 'menu',
          title: '用户管理',
          path: '/users',
          icon: 'User',
          children: [
            { id: 'user-c', type: 'button', title: '新增', permissionCode: 'user:create', enabled: true },
            { id: 'user-e', type: 'button', title: '编辑', permissionCode: 'user:edit', enabled: true },
            { id: 'user-rp', type: 'button', title: '重置密码', permissionCode: 'user:reset_password', enabled: true },
            { id: 'user-d', type: 'button', title: '删除', permissionCode: 'user:delete', enabled: true },
          ],
        },
        {
          id: 'role',
          type: 'menu',
          title: '角色权限',
          path: '/roles',
          icon: 'UserFilled',
          children: [
            { id: 'role-c', type: 'button', title: '新增', permissionCode: 'role:create', enabled: true },
            { id: 'role-e', type: 'button', title: '编辑', permissionCode: 'role:edit', enabled: true },
            { id: 'role-d', type: 'button', title: '删除', permissionCode: 'role:delete', enabled: true },
            { id: 'role-p', type: 'button', title: '权限设置', permissionCode: 'role:permission', enabled: true },
          ],
        },
        {
          id: 'menu',
          type: 'menu',
          title: '菜单管理',
          path: '/menus',
          icon: 'Menu',
          children: [
            { id: 'menu-c', type: 'button', title: '新增', permissionCode: 'menu:create', enabled: true },
            { id: 'menu-e', type: 'button', title: '编辑', permissionCode: 'menu:edit', enabled: true },
            { id: 'menu-d', type: 'button', title: '删除', permissionCode: 'menu:delete', enabled: true },
            { id: 'menu-pkg', type: 'button', title: '套餐中心收录', permissionCode: 'menu:package_feature', enabled: true },
          ],
        },
        {
          id: 'dict',
          type: 'menu',
          title: '数据字典',
          path: '/dict',
          icon: 'Collection',
          children: [
            { id: 'dict-type-c', type: 'button', title: '字典类型-新增', permissionCode: 'dict_type:create', enabled: true },
            { id: 'dict-type-e', type: 'button', title: '字典类型-编辑', permissionCode: 'dict_type:edit', enabled: true },
            { id: 'dict-type-d', type: 'button', title: '字典类型-删除', permissionCode: 'dict_type:delete', enabled: true },
            { id: 'dict-item-c', type: 'button', title: '字典项-新增', permissionCode: 'dict_item:create', enabled: true },
            { id: 'dict-item-e', type: 'button', title: '字典项-编辑', permissionCode: 'dict_item:edit', enabled: true },
            { id: 'dict-item-d', type: 'button', title: '字典项-删除', permissionCode: 'dict_item:delete', enabled: true },
          ],
        },
        {
          id: 'param',
          type: 'menu',
          title: '参数管理',
          path: '/params',
          icon: 'Tools',
          children: [
            { id: 'param-c', type: 'button', title: '新增', permissionCode: 'param:create', enabled: true },
            { id: 'param-e', type: 'button', title: '编辑', permissionCode: 'param:edit', enabled: true },
            { id: 'param-d', type: 'button', title: '删除', permissionCode: 'param:delete', enabled: true },
          ],
        },
        {
          id: 'audit-log',
          type: 'menu',
          title: '操作日志',
          path: '/audit-logs',
          icon: 'Document',
          children: [],
        },
        {
          id: 'login-log',
          type: 'menu',
          title: '登录日志',
          path: '/login-logs',
          icon: 'Key',
          children: [],
        },
      ],
    },
    {
      id: 'monitor',
      type: 'directory',
      title: '系统监控',
      icon: 'Monitor',
      children: [
        {
          id: 'mon-health',
          type: 'menu',
          title: '健康检查',
          path: '/monitor/health',
          icon: 'CircleCheck',
          children: [],
        },
        {
          id: 'mon-srv',
          type: 'menu',
          title: '服务器信息',
          path: '/monitor/server',
          icon: 'Cpu',
          children: [],
        },
        {
          id: 'mon-jobs',
          type: 'menu',
          title: '定时任务',
          path: '/monitor/jobs',
          icon: 'Timer',
          children: [],
        },
        {
          id: 'mon-svc',
          type: 'menu',
          title: '服务监控',
          path: '/monitor/services',
          icon: 'Connection',
          children: [],
        },
        {
          id: 'mon-cache',
          type: 'menu',
          title: '缓存监控',
          path: '/monitor/cache',
          icon: 'Histogram',
          children: [],
        },
        {
          id: 'mon-cache-keys',
          type: 'menu',
          title: '缓存列表',
          path: '/monitor/cache-keys',
          icon: 'List',
          children: [],
        },
      ],
    },
  ]
}

export function buildDefaultMenuTreeSnapshot(): MenuNode[] {
  return defaultTree()
}

/** 仅系统管理员可见的侧栏顶级菜单 id（与 defaultTree 中节点 id 一致） */
export const PLATFORM_ONLY_MENU_ID = 'tenant'
export const PLATFORM_ONLY_ROOT_MENU_IDS = new Set(['tenant', 'plan', 'monitor'])

/** 非平台管理员侧栏/菜单配置展示用：移除平台级节点 */
export function filterPlatformOnlyMenus(nodes: MenuNode[]): MenuNode[] {
  return nodes
    .filter((n) => !PLATFORM_ONLY_ROOT_MENU_IDS.has(n.id))
    .map((n) => {
      if (!n.children?.length) return { ...n }
      const children = filterPlatformOnlyMenus(n.children)
      return children.length ? { ...n, children } : { ...n, children: undefined }
    })
}

function cloneTenantBranchFrom(nodes: MenuNode[]): MenuNode | null {
  for (const n of nodes) {
    if (n.id === PLATFORM_ONLY_MENU_ID) return JSON.parse(JSON.stringify(n)) as MenuNode
    if (n.children?.length) {
      const t = cloneTenantBranchFrom(n.children)
      if (t) return t
    }
  }
  return null
}

/**
 * 租户管理员编辑菜单 JSON 时不含主体管理节点，保存前从当前完整树补回，避免误删平台菜单配置。
 */
export function mergeTenantMenuBranchForSave(edited: MenuNode[], baseline: MenuNode[]): MenuNode[] {
  const tenantBranch = cloneTenantBranchFrom(baseline)
  if (!tenantBranch) return edited
  const tenantMenu: MenuNode = tenantBranch
  const copy = JSON.parse(JSON.stringify(edited)) as MenuNode[]
  function insertUnderSys(nodes: MenuNode[]): boolean {
    for (let i = 0; i < nodes.length; i++) {
      if (nodes[i].id === 'sys') {
        const ch = [...(nodes[i].children || [])]
        const idx = ch.findIndex((c) => c.id === PLATFORM_ONLY_MENU_ID)
        if (idx >= 0) ch[idx] = tenantMenu
        else ch.unshift(tenantMenu)
        nodes[i] = { ...nodes[i], children: ch }
        return true
      }
      if (nodes[i].children?.length && insertUnderSys(nodes[i].children!)) return true
    }
    return false
  }
  insertUnderSys(copy)
  return copy
}

function loadFromStorage(): MenuNode[] {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return defaultTree()
    const parsed = JSON.parse(raw) as MenuNode[]
    if (!(Array.isArray(parsed) && parsed.length)) return defaultTree()
    const migrated = migrateDefaultNodes(parsed)
    const missingMenus = mergeMissingSysMenusFromDefaults(migrated)
    const { next, changed } = mergeBuiltinMenuButtonsWithDefaults(missingMenus.next)
    if (missingMenus.changed || changed) {
      try {
        localStorage.setItem(STORAGE_KEY, JSON.stringify(next))
      } catch {
        /* 存储满或隐私模式等：内存树已对齐，不写盘亦可 */
      }
    }
    return next
  } catch {
    return defaultTree()
  }
}

/**
 * 升级后 defaultTree 在「系统管理」下新增菜单项时，将缺失项按内置顺序插入并保留用户额外节点。
 */
function mergeMissingSysMenusFromDefaults(nodes: MenuNode[]): { next: MenuNode[]; changed: boolean } {
  const defaults = defaultTree()
  const sysDef = defaults.find((n) => n.id === 'sys')
  if (!sysDef?.children?.length) return { next: nodes, changed: false }
  const next = JSON.parse(JSON.stringify(nodes)) as MenuNode[]
  const sysIdx = next.findIndex((n) => n.id === 'sys')
  if (sysIdx < 0) return { next: nodes, changed: false }
  const sys = next[sysIdx]
  const children = [...(sys.children || [])]
  const before = JSON.stringify(children)
  const ordered: MenuNode[] = []
  const used = new Set<string>()
  for (const defChild of sysDef.children) {
    const existing = children.find((c) => c.id === defChild.id)
    if (existing) {
      ordered.push(existing)
      used.add(defChild.id)
    } else if (defChild.type === 'menu' && defChild.path) {
      ordered.push(JSON.parse(JSON.stringify(defChild)) as MenuNode)
      used.add(defChild.id)
    }
  }
  for (const c of children) {
    if (!used.has(c.id)) ordered.push(c)
  }
  if (JSON.stringify(ordered) === before) return { next: nodes, changed: false }
  next[sysIdx] = { ...sys, children: ordered }
  return { next, changed: true }
}

function migrateDefaultNodes(nodes: MenuNode[]): MenuNode[] {
  const hasPlans = !!findMenuByPath(nodes, '/plans')
  if (hasPlans) return nodes
  const next = JSON.parse(JSON.stringify(nodes)) as MenuNode[]
  const defaults = defaultTree()
  const planNode = findMenuByPath(defaults, '/plans')
  if (!planNode) return next
  const sys = next.find((node) => node.id === 'sys')
  if (sys) {
    const children = [...(sys.children || [])]
    const tenantIndex = children.findIndex((node) => node.id === 'tenant')
    children.splice(tenantIndex >= 0 ? tenantIndex + 1 : 0, 0, planNode)
    sys.children = children
  } else {
    next.push({
      id: 'sys',
      type: 'directory',
      title: '系统管理',
      icon: 'Setting',
      children: [planNode],
    })
  }
  localStorage.setItem(STORAGE_KEY, JSON.stringify(next))
  return next
}

/** 建立 id → 菜单节点（仅 type===menu），用于与 defaultTree 对齐内置子按钮 */
function indexMenuNodesById(nodes: MenuNode[], out: Map<string, MenuNode>): void {
  for (const n of nodes) {
    if (n.type === 'menu' && n.id) out.set(n.id, n)
    if (n.children?.length) indexMenuNodesById(n.children, out)
  }
}

const deprecatedBuiltinPermissionCodes = new Set([
  'dict:create',
  'dict:edit',
  'dict:delete',
  'home:view',
  'audit:view',
  'login:view',
  'brand:edit',
  'monhealth:view',
  'monserver:view',
  'monjobs:view',
  'monservices:view',
  'moncache:view',
  'moncachekeys:view',
])

function isDeprecatedBuiltinButton(node: MenuNode): boolean {
  return node.type === 'button' && !!node.permissionCode && deprecatedBuiltinPermissionCodes.has(node.permissionCode)
}

/**
 * 将 localStorage 中的旧侧栏树与当前 defaultTree 的内置「按钮」子项对齐。
 * 解决升级后仍只显示「维护」或菜单仅「编辑」：旧 JSON 不会自动出现新 permissionCode。
 */
function mergeBuiltinMenuButtonsWithDefaults(nodes: MenuNode[]): { next: MenuNode[]; changed: boolean } {
  const defaults = defaultTree()
  const defByMenuId = new Map<string, MenuNode>()
  indexMenuNodesById(defaults, defByMenuId)
  const before = JSON.stringify(nodes)

  function walk(list: MenuNode[]): MenuNode[] {
    return list.filter((node) => !isDeprecatedBuiltinButton(node)).map((node) => {
      const walkedChildren = node.children?.length ? walk(node.children) : undefined
      let nextNode: MenuNode = { ...node, children: walkedChildren }
      if (nextNode.type !== 'menu' || !nextNode.id) return nextNode

      const currentChildren = nextNode.children || []
      const defMenu = defByMenuId.get(nextNode.id)
      if (!defMenu?.children?.length) {
        nextNode.children = currentChildren
        return nextNode
      }

      const defButtons = defMenu.children.filter(
        (c) => c.type === 'button' && c.permissionCode && !deprecatedBuiltinPermissionCodes.has(c.permissionCode),
      )
      if (!defButtons.length) {
        nextNode.children = currentChildren
        return nextNode
      }

      const nonButtons = currentChildren.filter((c) => c.type !== 'button')
      const storedButtons = currentChildren.filter(
        (c) =>
          c.type === 'button' &&
          c.permissionCode &&
          !deprecatedBuiltinPermissionCodes.has(c.permissionCode),
      )
      const byCode = new Map(storedButtons.map((b) => [b.permissionCode!, b]))
      const mergedButtons: MenuNode[] = []
      for (const defBtn of defButtons) {
        const code = defBtn.permissionCode!
        const existing = byCode.get(code)
        if (existing) {
          const title =
            existing.title === '维护' ? defBtn.title : existing.title || defBtn.title
          mergedButtons.push({ ...defBtn, ...existing, title })
        } else {
          mergedButtons.push({ ...defBtn })
        }
      }
      for (const sb of storedButtons) {
        if (deprecatedBuiltinPermissionCodes.has(sb.permissionCode!)) continue
        if (!defButtons.some((d) => d.permissionCode === sb.permissionCode)) {
          mergedButtons.push(sb)
        }
      }
      nextNode.children = [...nonButtons, ...mergedButtons]
      return nextNode
    })
  }

  const next = walk(nodes)
  return { next, changed: before !== JSON.stringify(next) }
}

function findMenuByPath(nodes: MenuNode[], path: string): MenuNode | null {
  for (const n of nodes) {
    if (n.type === 'menu' && n.path && (path === n.path || path.startsWith(n.path + '/'))) {
      return n
    }
    if (n.children?.length) {
      const hit = findMenuByPath(n.children, path)
      if (hit) return hit
    }
  }
  return null
}

function applyTenantOverrides(
  nodes: MenuNode[],
  bundles: MenuBundle[],
  overrides: TenantMenuOverride[],
): MenuNode[] {
  if (!bundles.length || !overrides.length) return nodes
  const bundleByPath = new Map(bundles.map((item) => [item.path, item]))
  const overrideByPermissionId = new Map(overrides.map((item) => [item.permission_id, item]))

  function walk(items: MenuNode[]): MenuNode[] {
    const next: MenuNode[] = []
    for (const node of items) {
      const cloned: MenuNode = { ...node }
      if (cloned.children?.length) cloned.children = walk(cloned.children)
      if (cloned.type === 'menu' && cloned.path) {
        const bundle = bundleByPath.get(cloned.path)
        const override = bundle ? overrideByPermissionId.get(bundle.menu_permission_id) : undefined
        if (override) {
          if (override.custom_name) cloned.title = override.custom_name
          if (override.custom_icon) cloned.icon = override.custom_icon
          if (override.enabled === false || override.visible === false) cloned.enabled = false
          if (override.enabled === true && override.visible !== false) delete cloned.enabled
        }
      }
      next.push(cloned)
    }
    return next
  }

  return walk(nodes)
}

/** 定位节点所在同级列表与下标（用于排序）；根级时 siblings 即为整棵树 */
function findSiblingContext(
  nodes: MenuNode[],
  id: string,
): { siblings: MenuNode[]; index: number } | null {
  for (let i = 0; i < nodes.length; i++) {
    if (nodes[i].id === id) return { siblings: nodes, index: i }
    const ch = nodes[i].children
    if (ch?.length) {
      const hit = findSiblingContext(ch, id)
      if (hit) return hit
    }
  }
  return null
}

export const useSidebarMenuStore = defineStore('sidebarMenu', () => {
  const tree = ref<MenuNode[]>(loadFromStorage())
  const menuBundles = ref<MenuBundle[]>([])
  const tenantOverrides = ref<TenantMenuOverride[]>([])
  const overridesLoaded = ref(false)
  const overridesLoadedAt = ref(0)
  const tenantTree = computed(() => applyTenantOverrides(tree.value, menuBundles.value, tenantOverrides.value))

  function persist() {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(tree.value))
  }

  function resetDefault() {
    tree.value = defaultTree()
    persist()
  }

  function replace(next: MenuNode[]) {
    tree.value = next
    persist()
  }

  function normalizeBuiltinTree(): boolean {
    const { next, changed } = mergeBuiltinMenuButtonsWithDefaults(tree.value)
    if (!changed) return false
    tree.value = next
    persist()
    return true
  }

  async function loadTenantMenuRuntime(options: { force?: boolean; isPlatformAdmin?: boolean } = {}) {
    if (options.isPlatformAdmin) {
      overridesLoaded.value = true
      overridesLoadedAt.value = Date.now()
      tenantOverrides.value = []
      return
    }
    const fresh = overridesLoaded.value && Date.now() - overridesLoadedAt.value < 5 * 60_000
    if (fresh && !options.force) return
    const [bundles, data] = await Promise.all([
      fetchMenuBundles(),
      fetchTenantMenuOverrides(),
    ])
    menuBundles.value = bundles
    tenantOverrides.value = data.overrides || []
    overridesLoaded.value = true
    overridesLoadedAt.value = Date.now()
  }

  function clearTenantMenuRuntime() {
    menuBundles.value = []
    tenantOverrides.value = []
    overridesLoaded.value = false
    overridesLoadedAt.value = 0
  }

  function menuPermissionIdForPath(path: string): number | null {
    const bundle = menuBundles.value.find((item) => item.path === path)
    return bundle?.menu_permission_id ?? null
  }

  async function saveTenantMenuOverrideForPath(path: string, patch: Omit<TenantMenuOverrideValue, 'permission_id'>) {
    if (!menuBundles.value.length) {
      await loadTenantMenuRuntime({ force: true })
    }
    const permissionId = menuPermissionIdForPath(path)
    if (!permissionId) throw new Error('未找到菜单权限映射，无法保存租户菜单覆盖')
    const current = tenantOverrides.value.find((item) => item.permission_id === permissionId)
    const merged: TenantMenuOverrideValue = {
      permission_id: permissionId,
      custom_name: current?.custom_name ?? null,
      custom_icon: current?.custom_icon ?? null,
      enabled: current?.enabled ?? null,
      visible: current?.visible ?? null,
      sort_order: current?.sort_order ?? null,
      ...patch,
    }
    const rest = tenantOverrides.value
      .filter((item) => item.permission_id !== permissionId)
      .map((item) => ({
        permission_id: item.permission_id,
        custom_name: item.custom_name ?? null,
        custom_icon: item.custom_icon ?? null,
        enabled: item.enabled ?? null,
        visible: item.visible ?? null,
        sort_order: item.sort_order ?? null,
      }))
    const data = await saveTenantMenuOverrides([...rest, merged])
    tenantOverrides.value = data.overrides || []
    overridesLoaded.value = true
    overridesLoadedAt.value = Date.now()
  }

  /**
   * 与同父节点下的相邻项交换顺序（侧栏、菜单管理树表共用同一份 tree，持久化后即时生效）。
   * 使用深拷贝替换根树，避免仅改嵌套数组时部分计算属性不刷新。
   */
  function moveMenuNode(id: string, dir: 'up' | 'down'): boolean {
    const next = JSON.parse(JSON.stringify(tree.value)) as MenuNode[]
    const ctx = findSiblingContext(next, id)
    if (!ctx) return false
    const { siblings, index } = ctx
    const j = dir === 'up' ? index - 1 : index + 1
    if (j < 0 || j >= siblings.length) return false
    const tmp = siblings[index]
    siblings[index] = siblings[j]
    siblings[j] = tmp
    tree.value = next
    persist()
    return true
  }

  /** `enabled === false` 时目录/菜单/按钮均不进入侧栏（平台与主体租户一致）。 */
  function setNodeEnabled(id: string, enabled: boolean) {
    function walk(nodes: MenuNode[]): boolean {
      for (const n of nodes) {
        if (n.id === id) {
          if (enabled) delete n.enabled
          else n.enabled = false
          return true
        }
        if (n.children?.length && walk(n.children)) return true
      }
      return false
    }
    const next = JSON.parse(JSON.stringify(tree.value)) as MenuNode[]
    if (walk(next)) {
      tree.value = next
      persist()
    }
  }

  function setNodeTitle(id: string, title: string) {
    const normalized = title.trim()
    if (!normalized) return
    function walk(nodes: MenuNode[]): boolean {
      for (const n of nodes) {
        if (n.id === id) {
          n.title = normalized
          return true
        }
        if (n.children?.length && walk(n.children)) return true
      }
      return false
    }
    const next = JSON.parse(JSON.stringify(tree.value)) as MenuNode[]
    if (walk(next)) {
      tree.value = next
      persist()
    }
  }

  function menuMoveState(id: string): { canUp: boolean; canDown: boolean } {
    const ctx = findSiblingContext(tree.value, id)
    if (!ctx) return { canUp: false, canDown: false }
    const { siblings, index } = ctx
    return {
      canUp: index > 0,
      canDown: index < siblings.length - 1,
    }
  }

  /** 在侧栏树根级追加节点（与首页、系统管理等顶级项同级，持久化到 localStorage） */
  function appendRootMenuNode(node: MenuNode): boolean {
    const next = JSON.parse(JSON.stringify(tree.value)) as MenuNode[]
    next.push({ ...node, children: node.children ?? [] })
    tree.value = next
    persist()
    return true
  }

  /** 在指定父节点下追加子节点（平台菜单结构，持久化到 localStorage） */
  function addMenuChild(parentId: string, node: MenuNode): boolean {
    const next = JSON.parse(JSON.stringify(tree.value)) as MenuNode[]
    function walk(nodes: MenuNode[]): boolean {
      for (const n of nodes) {
        if (n.id === parentId) {
          if (!n.children) n.children = []
          n.children.push({ ...node, children: node.children ?? [] })
          return true
        }
        if (n.children?.length && walk(n.children)) return true
      }
      return false
    }
    if (!walk(next)) return false
    tree.value = next
    persist()
    return true
  }

  /** 更新节点展示字段（平台菜单树，持久化到 localStorage） */
  function updateMenuNode(
    nodeId: string,
    patch: Partial<Pick<MenuNode, 'title' | 'path' | 'icon' | 'permissionCode' | 'isPlatformOnly' | 'enabled' | 'dataPermMode'>>,
  ): boolean {
    const next = JSON.parse(JSON.stringify(tree.value)) as MenuNode[]
    function walk(nodes: MenuNode[]): boolean {
      for (const n of nodes) {
        if (n.id === nodeId) {
          if (patch.title !== undefined) n.title = patch.title
          if (patch.path !== undefined) n.path = patch.path
          if (patch.icon !== undefined) n.icon = patch.icon
          if (patch.permissionCode !== undefined) n.permissionCode = patch.permissionCode
          if (patch.isPlatformOnly !== undefined) n.isPlatformOnly = patch.isPlatformOnly
          if (patch.enabled !== undefined) n.enabled = patch.enabled
          if (patch.dataPermMode !== undefined) n.dataPermMode = patch.dataPermMode
          return true
        }
        if (n.children?.length && walk(n.children)) return true
      }
      return false
    }
    if (!walk(next)) return false
    tree.value = next
    persist()
    return true
  }

  /** 从树中移除节点（仅应由调用方限制为自定义节点 id） */
  function removeMenuNode(nodeId: string): boolean {
    const next = JSON.parse(JSON.stringify(tree.value)) as MenuNode[]
    function removeFrom(nodes: MenuNode[]): boolean {
      const i = nodes.findIndex((n) => n.id === nodeId)
      if (i >= 0) {
        nodes.splice(i, 1)
        return true
      }
      for (const n of nodes) {
        if (n.children?.length && removeFrom(n.children)) return true
      }
      return false
    }
    if (!removeFrom(next)) return false
    tree.value = next
    persist()
    return true
  }

  function menuForRoute(routePath: string) {
    return findMenuByPath(tenantTree.value, routePath)
  }

  return {
    tree,
    tenantTree,
    menuBundles,
    tenantOverrides,
    overridesLoaded,
    persist,
    resetDefault,
    replace,
    normalizeBuiltinTree,
    loadTenantMenuRuntime,
    clearTenantMenuRuntime,
    saveTenantMenuOverrideForPath,
    moveMenuNode,
    setNodeEnabled,
    setNodeTitle,
    menuMoveState,
    addMenuChild,
    appendRootMenuNode,
    updateMenuNode,
    removeMenuNode,
    menuForRoute,
  }
})
