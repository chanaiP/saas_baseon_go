<script setup lang="ts">
defineOptions({ name: 'MenuView' })
import { ArrowDown, ArrowDownBold, ArrowUp, ArrowUpBold } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { computed, nextTick, onMounted, ref, watch } from 'vue'

import type { MenuBundle } from '@/api/permission'
import { fetchPermissionMenuBundles, updatePermission } from '@/api/permission'
import { confirmArchiveAction } from '@/composables/useArchiveConfirm'
import { usePermissionStore } from '@/stores/permission'
import { filterPlatformOnlyMenus, useSidebarMenuStore } from '@/stores/sidebarMenu'
import type { MenuNode, MenuNodeType } from '@/types/menu'
import type { FilterField, TableColumn } from '@/views/components/NeuroAgentListPage.vue'
import NeuroAgentDialog from '@/views/components/NeuroAgentDialog.vue'
import MenuIconSelect from '@/views/components/MenuIconSelect.vue'
import NeuroAgentListPage from '@/views/components/NeuroAgentListPage.vue'
import { menuIconComponent } from '@/utils/menuIconPicker'

/** 工具栏「上级节点」中表示顶级挂载的哨兵值（与首页、系统管理等同级） */
const ROOT_MENU_PARENT_ID = '__menu_tree_root__'

/** 根级目录（与侧栏 el-sub-menu 同级），同一时间只展开其中一个 */
function rootDirectoryNodes(nodes: MenuNode[]): MenuNode[] {
  return nodes.filter((n) => n.type === 'directory' && n.children?.length)
}

function collectDescendantIds(node: MenuNode): string[] {
  const ids: string[] = []
  for (const c of node.children || []) {
    ids.push(c.id, ...collectDescendantIds(c))
  }
  return ids
}

/** 所有带子节点的行 id（用于一键展开树表） */
function collectIdsWithChildren(nodes: MenuNode[]): string[] {
  const ids: string[] = []
  function walk(list: MenuNode[]) {
    for (const n of list) {
      if (n.children?.length) {
        ids.push(n.id)
        walk(n.children)
      }
    }
  }
  walk(nodes)
  return ids
}

/** 默认只展开第一个根目录，与侧栏 unique-opened 维护习惯一致 */
function defaultExpandedKeys(nodes: MenuNode[]): string[] {
  const dirs = rootDirectoryNodes(nodes)
  if (dirs.length) return [dirs[0].id]
  const any = nodes.find((n) => n.children?.length)
  return any ? [any.id] : []
}

const perm = usePermissionStore()
const store = useSidebarMenuStore()

/** 平台管理员菜单归属展示（来自后端菜单包元数据 is_platform_only 等） */
const menuBundlesForScope = ref<MenuBundle[]>([])

const bundleByMenuPath = computed(() => {
  const m = new Map<string, MenuBundle>()
  for (const b of menuBundlesForScope.value) {
    m.set(b.path, b)
  }
  return m
})

async function loadMenuBundlesForScope() {
  if (!perm.profile?.is_platform_admin) {
    menuBundlesForScope.value = []
    return
  }
  try {
    menuBundlesForScope.value = await fetchPermissionMenuBundles()
  } catch {
    menuBundlesForScope.value = []
  }
}

/** 目录无归属；菜单按路由 path；按钮按 permissionCode 在 bundles.operations 中匹配 */
function menuScopeTag(row: MenuNode): { text: string; type: 'warning' | 'success' | 'info' } | null {
  if (row.type === 'directory') return null
  if (row.type === 'menu' && row.path) {
    const b = bundleByMenuPath.value.get(row.path)
    if (!b) return { text: '自定义', type: 'info' }
    if (b.is_platform_only) return { text: '仅平台', type: 'warning' }
    return { text: '租户菜单', type: 'success' }
  }
  if (row.type === 'button' && row.permissionCode) {
    for (const b of menuBundlesForScope.value) {
      const op = b.operations.find((o) => o.path === row.permissionCode)
      if (op) {
        if (op.is_platform_only) return { text: '仅平台', type: 'warning' }
        return { text: '租户菜单', type: 'success' }
      }
    }
    return { text: '自定义', type: 'info' }
  }
  return null
}

const isPlatformAdmin = computed(
  () => !!(perm.profile?.is_platform_admin || perm.profile?.tenant_is_platform),
)
const canCreateMenu = computed(() => perm.canUseAction('menu:create'))
const canEditMenu = computed(() => perm.canUseAction('menu:edit'))
const canDeleteMenu = computed(() => perm.canUseAction('menu:delete'))

const selectedMenuNode = ref<MenuNode | null>(null)
/** 从表格行「增加子项」打开时为该行；工具栏打开时为 null，须通过 globalAddParentId 指定上级 */
const addContextRow = ref<MenuNode | null>(null)
/** 工具栏「新增子项」时必选：父级节点 id（仅目录或菜单） */
const globalAddParentId = ref<string | null>(null)
const addDlg = ref(false)
const addForm = ref<{
  nodeType: MenuNodeType
  title: string
  path: string
  permissionCode: string
  icon: string
  dataPermMode: 'NONE' | 'ORG' | 'BU' | 'ORG_BU'
}>({ nodeType: 'menu', title: '新菜单', path: '', permissionCode: '', icon: 'Document', dataPermMode: 'ORG' })

const editDlg = ref(false)
const editTargetId = ref<string | null>(null)
const editForm = ref({
  title: '',
  path: '',
  permissionCode: '',
  icon: 'Document',
  dataPermMode: 'ORG' as 'NONE' | 'ORG' | 'BU' | 'ORG_BU',
})

function filterTenantManageableTree(nodes: MenuNode[]): MenuNode[] {
  const out: MenuNode[] = []
  for (const node of nodes) {
    if (node.type === 'directory') {
      const children = filterTenantManageableTree(node.children || [])
      if (children.length) out.push({ ...node, children })
      continue
    }
    if (node.type === 'menu' && node.path) {
      const filteredChildren = filterTenantManageableTree(node.children || [])
      /** 必须递归过滤子按钮：否则仅菜单或未开通删除时仍会带出整棵默认子节点（看起来「套餐关了删除却还在」） */
      if (perm.canUseMenuPath(node.path) || filteredChildren.length > 0) {
        out.push({ ...node, children: filteredChildren })
      }
      continue
    }
    if (node.type === 'button' && node.permissionCode && perm.canUseAction(node.permissionCode)) {
      out.push({ ...node })
    }
  }
  return out
}

const displayTree = computed(() => {
  if (isPlatformAdmin.value) return store.tree
  return filterTenantManageableTree(filterPlatformOnlyMenus(store.tenantTree))
})

/** 点击「查询」后生效的名称筛选（与 NeuroAgentListPage 筛选行一致） */
const menuAppliedTitleFilter = ref('')

const menuFilterFields: FilterField[] = [
  { key: 'menuTitle', label: '名称', type: 'text', placeholder: '模糊匹配' },
]

/** 按显示名称筛选树行；父节点命中时保留整棵子树 */
function filterMenuTreeByTitle(nodes: MenuNode[], query: string): MenuNode[] {
  const q = query.trim().toLowerCase()
  if (!q) return nodes
  const out: MenuNode[] = []
  for (const n of nodes) {
    const filteredChildren =
      n.children && n.children.length ? filterMenuTreeByTitle(n.children, query) : []
    const selfMatch = n.title.toLowerCase().includes(q)
    if (selfMatch) {
      out.push({ ...n, children: n.children })
    } else if (filteredChildren.length) {
      out.push({ ...n, children: filteredChildren })
    }
  }
  return out
}

const menuTableTree = computed(() => {
  const q = menuAppliedTitleFilter.value.trim()
  if (!q) return displayTree.value
  return filterMenuTreeByTitle(displayTree.value, menuAppliedTitleFilter.value)
})

const expandedRowKeys = ref<string[]>([])
const editingTitleId = ref<string | null>(null)
const editingTitleValue = ref('')

function typeZh(t: string) {
  const m: Record<string, string> = { directory: '目录', menu: '菜单', button: '按钮' }
  return m[t] || t
}

function moveState(id: string) {
  return store.menuMoveState(id)
}

function moveUp(id: string) {
  if (!isPlatformAdmin.value || !canEditMenu.value) return
  store.moveMenuNode(id, 'up')
}

function moveDown(id: string) {
  if (!isPlatformAdmin.value || !canEditMenu.value) return
  store.moveMenuNode(id, 'down')
}

function reset() {
  if (!isPlatformAdmin.value || !canDeleteMenu.value) return
  void ElMessageBox.confirm('将侧栏菜单恢复为系统默认结构，确定继续？', '恢复默认', { type: 'warning' })
    .then(() => {
      store.resetDefault()
      selectedMenuNode.value = null
      ElMessage.success('已恢复默认')
    })
    .catch(() => {})
}

function canToggle(row: MenuNode) {
  return isPlatformAdmin.value || (canEditMenu.value && row.type === 'menu' && !!row.path)
}

async function onRowEnabled(row: MenuNode, enabled: boolean) {
  if (!canToggle(row)) return
  if (isPlatformAdmin.value) {
    if (!canEditMenu.value) return
    store.setNodeEnabled(row.id, enabled)
    return
  }
  try {
    await store.saveTenantMenuOverrideForPath(row.path!, { enabled, visible: enabled })
    ElMessage.success('菜单覆盖已保存')
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '保存菜单覆盖失败')
  }
}

function canRename(row: MenuNode) {
  return !isPlatformAdmin.value && canEditMenu.value && row.type === 'menu'
}

function startTitleEdit(row: MenuNode) {
  if (!canRename(row)) return
  editingTitleId.value = row.id
  editingTitleValue.value = row.title
}

function findNodeById(nodes: MenuNode[], id: string): MenuNode | null {
  for (const n of nodes) {
    if (n.id === id) return n
    if (n.children?.length) {
      const hit = findNodeById(n.children, id)
      if (hit) return hit
    }
  }
  return null
}

function findParentNode(nodes: MenuNode[], childId: string): MenuNode | null {
  for (const n of nodes) {
    if (n.children?.some((c) => c.id === childId)) return n
    if (n.children?.length) {
      const hit = findParentNode(n.children, childId)
      if (hit) return hit
    }
  }
  return null
}

type ResolveParentResult = { ok: true; parent: MenuNode } | { ok: false; message: string }

type AddParentScenario = 'none' | 'root' | 'row' | 'global-node'

const addParentScenario = computed((): AddParentScenario => {
  if (addContextRow.value) return 'row'
  if (globalAddParentId.value === ROOT_MENU_PARENT_ID) return 'root'
  if (globalAddParentId.value) return 'global-node'
  return 'none'
})

/** 当前要挂载到的「直接父节点」：行操作为该行（按钮行则取其所属菜单）；工具栏为下拉所选（不含根哨兵）。 */
const effectiveDirectAddParent = computed((): MenuNode | null => {
  const ctx = addContextRow.value
  const tree = displayTree.value
  if (ctx?.type === 'directory' || ctx?.type === 'menu') return ctx
  if (ctx?.type === 'button') {
    const pm = findParentNode(tree, ctx.id)
    return pm?.type === 'menu' ? pm : null
  }
  if (!ctx && globalAddParentId.value && globalAddParentId.value !== ROOT_MENU_PARENT_ID) {
    const n = findNodeById(tree, globalAddParentId.value)
    if (n?.type === 'directory' || n?.type === 'menu') return n
    return null
  }
  return null
})

/** 目录下仅允许子目录、菜单；菜单下仅允许按钮；顶级根下仅允许目录或菜单（与现有首页等同级） */
const allowedAddNodeTypes = computed((): MenuNodeType[] => {
  if (addParentScenario.value === 'root') return ['directory', 'menu']
  if (addParentScenario.value === 'none') return []
  const p = effectiveDirectAddParent.value
  if (!p) return []
  if (p.type === 'directory') return ['directory', 'menu']
  if (p.type === 'menu') return ['button']
  return []
})

/** 工具栏打开且无行上下文时，须先选上级 */
const addDialogNeedsGlobalParentPicker = computed(() => !addContextRow.value)

function collectDirMenuPickerOptions(nodes: MenuNode[], depth: number, out: { value: string; label: string }[]) {
  for (const n of nodes) {
    if (n.type === 'directory' || n.type === 'menu') {
      const indent = `${'\u3000'.repeat(depth)}${depth ? '├ ' : ''}`
      const typ = n.type === 'directory' ? '目录' : '菜单'
      const pathHint = n.path ? ` ${n.path}` : ''
      out.push({ value: n.id, label: `${indent}${n.title}（${typ}${pathHint}）` })
      if (n.children?.length) collectDirMenuPickerOptions(n.children, depth + 1, out)
    }
  }
}

const parentPickerOptions = computed(() => {
  const out: { value: string; label: string }[] = [
    {
      value: ROOT_MENU_PARENT_ID,
      label: '根节点（顶级，与首页、系统管理等侧栏顶级项同级）',
    },
  ]
  collectDirMenuPickerOptions(displayTree.value, 0, out)
  return out
})

/** 目录下仅子目录+菜单；菜单下仅按钮；父节点非法则失败 */
function resolveDirectParentForNewNode(nodeType: MenuNodeType, parentNode: MenuNode | null): ResolveParentResult {
  if (!parentNode) {
    return {
      ok: false,
      message:
        '请先指定上级：在下方选择「根节点」或某一目录/菜单，或在表格中点击「增加子项」从当前行挂载。',
    }
  }
  if (nodeType === 'button') {
    if (parentNode.type !== 'menu') {
      return { ok: false, message: '按钮（操作）只能添加在「菜单」下。' }
    }
    return { ok: true, parent: parentNode }
  }
  if (nodeType === 'menu') {
    if (parentNode.type !== 'directory') {
      return { ok: false, message: '菜单只能添加在「目录」下；菜单下只能添加按钮，不能添加子菜单。' }
    }
    return { ok: true, parent: parentNode }
  }
  if (nodeType === 'directory') {
    if (parentNode.type !== 'directory') {
      return { ok: false, message: '子目录只能添加在「目录」下。' }
    }
    return { ok: true, parent: parentNode }
  }
  return { ok: false, message: '类型无效' }
}

function openAddDlg() {
  addContextRow.value = null
  globalAddParentId.value = null
  addForm.value = { nodeType: 'menu', title: '新菜单', path: '', permissionCode: '', icon: 'Document', dataPermMode: 'ORG' }
  addDlg.value = true
}

function openAddDlgUnderRow(row: MenuNode) {
  if (!isPlatformAdmin.value || !canCreateMenu.value) return
  globalAddParentId.value = null
  addContextRow.value = row
  if (row.type === 'menu') {
    addForm.value = { nodeType: 'button', title: '新按钮', path: '', permissionCode: '', icon: 'Document', dataPermMode: 'ORG' }
  } else if (row.type === 'directory') {
    addForm.value = { nodeType: 'menu', title: '新菜单', path: '', permissionCode: '', icon: 'Document', dataPermMode: 'ORG' }
  } else {
    addForm.value = { nodeType: 'button', title: '新按钮', path: '', permissionCode: '', icon: 'Document', dataPermMode: 'ORG' }
  }
  addDlg.value = true
}

function onAddDlgClosed() {
  addContextRow.value = null
  globalAddParentId.value = null
}

function confirmAdd() {
  if (!allowedAddNodeTypes.value.includes(addForm.value.nodeType)) {
    ElMessage.warning('当前上级下不允许该类型，请检查目录/菜单/按钮层级规则。')
    return
  }
  const t = addForm.value.title.trim()
  if (!t) {
    ElMessage.warning('请输入名称')
    return
  }
  const nt = addForm.value.nodeType
  if (nt === 'menu' && !addForm.value.path.trim().startsWith('/')) {
    ElMessage.warning('菜单路由需以 / 开头')
    return
  }
  if (nt === 'button' && !addForm.value.permissionCode.trim()) {
    ElMessage.warning('按钮须填写权限码')
    return
  }
  const id = `custom_menu_${Date.now()}`
  const node: MenuNode = {
    id,
    type: nt,
    title: t,
    icon: addForm.value.icon.trim() || 'Document',
    enabled: true,
  }
  if (nt === 'menu') node.path = addForm.value.path.trim()
  if (nt === 'button') node.permissionCode = addForm.value.permissionCode.trim()
  if (nt === 'menu') node.dataPermMode = addForm.value.dataPermMode

  if (addParentScenario.value === 'root') {
    if (nt === 'button') {
      ElMessage.warning('根节点下不能添加按钮，请选择某一菜单作为上级。')
      return
    }
    store.appendRootMenuNode(node)
    addDlg.value = false
    ElMessage.success('已添加到侧栏顶级')
    return
  }

  const resolved = resolveDirectParentForNewNode(nt, effectiveDirectAddParent.value)
  if (!resolved.ok) {
    ElMessage.warning(resolved.message)
    return
  }
  const parent = resolved.parent
  if (store.addMenuChild(parent.id, node)) {
    addDlg.value = false
    ElMessage.success('已添加')
  } else {
    ElMessage.error('添加失败，请确认父节点仍存在')
  }
}

const editTargetRow = computed(() => {
  const id = editTargetId.value
  if (!id) return null
  return findNodeById(displayTree.value, id)
})

function openEditRow(row: MenuNode) {
  if (!isPlatformAdmin.value || !canEditMenu.value) return
  editTargetId.value = row.id
  editForm.value = {
    title: row.title,
    path: row.path || '',
    permissionCode: row.permissionCode || '',
    icon: row.icon || 'Document',
    dataPermMode: row.dataPermMode || 'ORG',
  }
  editDlg.value = true
}

function onEditDlgClosed() {
  editTargetId.value = null
}

async function confirmEdit() {
  const id = editTargetId.value
  if (!id) return
  const row = findNodeById(displayTree.value, id)
  if (!row) return
  const t = editForm.value.title.trim()
  if (!t) {
    ElMessage.warning('请输入名称')
    return
  }
  if (row.type === 'menu' && !editForm.value.path.trim().startsWith('/')) {
    ElMessage.warning('菜单路由须以 / 开头')
    return
  }
  if (row.type === 'button' && !editForm.value.permissionCode.trim()) {
    ElMessage.warning('按钮须填写权限码')
    return
  }
  const patch: Partial<MenuNode> = { title: t, icon: editForm.value.icon.trim() || 'Document' }
  if (row.type === 'menu') patch.path = editForm.value.path.trim()
  if (row.type === 'button') patch.permissionCode = editForm.value.permissionCode.trim()
  if (row.type === 'menu') patch.dataPermMode = editForm.value.dataPermMode
  if (!store.updateMenuNode(id, patch)) return
  if (row.type === 'menu' && row.path) {
    const bundle = bundleByMenuPath.value.get(row.path)
    if (bundle?.menu_permission_id) {
      try {
        await updatePermission(bundle.menu_permission_id, {
          data_perm_mode: editForm.value.dataPermMode,
        })
        await loadMenuBundlesForScope()
      } catch (e) {
        ElMessage.error(e instanceof Error ? e.message : '权限类型保存失败')
        return
      }
    } else {
      ElMessage.error('未找到该菜单的后端权限映射，已阻止保存。请刷新页面后重试。')
      return
    }
  }
  editDlg.value = false
  ElMessage.success('已保存')
}

async function removeCustomRow(row: MenuNode) {
  if (!isPlatformAdmin.value || !canDeleteMenu.value) return
  if (!row.id.startsWith('custom_menu_')) {
    ElMessage.warning('仅可删除以「新增子项」创建的自定义节点')
    return
  }
  await confirmArchiveAction({ name: row.title, title: '移除自定义菜单', detail: '仅移除当前自定义菜单配置；平台菜单源定义和历史审计不会被删除。', confirmText: '移除' })
  if (store.removeMenuNode(row.id)) {
    if (selectedMenuNode.value?.id === row.id) selectedMenuNode.value = null
    ElMessage.success('已移除')
  }
}

async function submitTitleEdit(row: MenuNode) {
  if (editingTitleId.value !== row.id) return
  const nextTitle = editingTitleValue.value.trim()
  if (nextTitle && nextTitle !== row.title) {
    if (isPlatformAdmin.value) {
      if (!canEditMenu.value) return
      store.setNodeTitle(row.id, nextTitle)
    } else {
      try {
        await store.saveTenantMenuOverrideForPath(row.path!, { custom_name: nextTitle })
        ElMessage.success('菜单名称已保存')
      } catch (e) {
        ElMessage.error(e instanceof Error ? e.message : '保存菜单名称失败')
      }
    }
  }
  editingTitleId.value = null
  editingTitleValue.value = ''
}

function cancelTitleEdit() {
  editingTitleId.value = null
  editingTitleValue.value = ''
}

function isRootNode(row: MenuNode): boolean {
  return menuTableTree.value.some((n) => n.id === row.id)
}

/** 树表行在整棵树中的深度：0=根（与侧栏一级一致），1=二级，2+=按钮等 */
function depthOfRow(row: MenuNode, nodes: MenuNode[] = menuTableTree.value, depth = 0): number | null {
  for (const n of nodes) {
    if (n.id === row.id) return depth
    if (n.children?.length) {
      const d = depthOfRow(row, n.children, depth + 1)
      if (d !== null) return d
    }
  }
  return null
}

function menuTableRowClassName(data: { row: MenuNode }) {
  const d = depthOfRow(data.row) ?? 0
  if (d <= 0) return 'menu-tbl-row--l0'
  if (d === 1) return 'menu-tbl-row--l1'
  return 'menu-tbl-row--l2'
}

function titleCellClass(row: MenuNode) {
  const d = depthOfRow(row) ?? 0
  const cls = ['menu-tbl-title-cell']
  if (d === 0) cls.push('menu-tbl-title--root')
  return cls.join(' ')
}

function sortUseBoldArrows(row: MenuNode) {
  return (depthOfRow(row) ?? 0) === 0
}

/**
 * 与侧栏 `unique-opened` 对齐：仅根级 **目录** 互斥展开；根级单页菜单（如首页）可与目录并存；目录下多行菜单可同时展开以查看按钮。
 */
function onTreeExpandChange(row: MenuNode, expandedRows: any[]) {
  const expanded = expandedRows.some((r: any) => r.id === row.id)
  const roots = menuTableTree.value
  const dirRoots = rootDirectoryNodes(roots)
  const dirRootIds = new Set(dirRoots.map((n) => n.id))

  if (!expanded) {
    const drop = new Set([row.id, ...collectDescendantIds(row)])
    expandedRowKeys.value = expandedRowKeys.value.filter((id) => !drop.has(id))
    return
  }

  if (dirRootIds.has(row.id) && isRootNode(row)) {
    const toRemove = new Set<string>()
    for (const o of dirRoots) {
      if (o.id === row.id) continue
      toRemove.add(o.id)
      for (const id of collectDescendantIds(o)) toRemove.add(id)
    }
    expandedRowKeys.value = expandedRowKeys.value.filter((id) => !toRemove.has(id))
    if (!expandedRowKeys.value.includes(row.id)) {
      expandedRowKeys.value = [...expandedRowKeys.value, row.id]
    }
    return
  }

  if (!expandedRowKeys.value.includes(row.id)) {
    expandedRowKeys.value = [...expandedRowKeys.value, row.id]
  }
}

function onTitleToggle(row: MenuNode) {
  const isExpanded = expandedRowKeys.value.includes(row.id)
  if (isExpanded) {
    const drop = new Set([row.id, ...collectDescendantIds(row)])
    expandedRowKeys.value = expandedRowKeys.value.filter((id) => !drop.has(id))
  } else {
    if (!expandedRowKeys.value.includes(row.id)) {
      expandedRowKeys.value = [...expandedRowKeys.value, row.id]
    }
  }
}

async function expandAllTree() {
  expandedRowKeys.value = [...collectIdsWithChildren(menuTableTree.value)]
  await nextTick()
}

function collapseAllTree() {
  expandedRowKeys.value = []
}

function onMenuSearch(payload: { keyword: string; filters: Record<string, unknown> }) {
  menuAppliedTitleFilter.value = String(payload.filters?.menuTitle ?? '').trim()
}

watch(
  () =>
    [
      displayTree.value.map((n) => n.id).join(','),
      menuAppliedTitleFilter.value,
      menuTableTree.value.map((n) => n.id).join(','),
    ] as const,
  () => {
    if (menuAppliedTitleFilter.value.trim()) {
      expandedRowKeys.value = [...collectIdsWithChildren(menuTableTree.value)]
    } else {
      expandedRowKeys.value = defaultExpandedKeys(displayTree.value)
    }
  },
  { immediate: true },
)

watch(
  () => [allowedAddNodeTypes.value.join(','), addForm.value.nodeType] as const,
  () => {
    const allowed = allowedAddNodeTypes.value
    if (!allowed.length) return
    if (!allowed.includes(addForm.value.nodeType)) {
      addForm.value.nodeType = allowed[0]!
    }
  },
)

const columns = computed<TableColumn[]>(() => [
  { key: 'sort', title: '排序', width: 104, align: 'center', fixed: 'left', hidden: !isPlatformAdmin.value || !canEditMenu.value },
  { key: 'title', title: '名称', minWidth: 240 },
  { key: 'sidebar', title: '侧栏', width: 120, minWidth: 120, align: 'center', hidden: !canEditMenu.value },
  { key: 'type', title: '类型', width: 104, minWidth: 96, align: 'center' },
  {
    key: 'scope',
    title: '归属',
    width: 108,
    minWidth: 96,
    align: 'center',
    hidden: !isPlatformAdmin.value,
  },
  { key: 'path', title: '路由', minWidth: 200, width: 220, hidden: !isPlatformAdmin.value },
  { key: 'permissionCode', title: '权限码', minWidth: 160, width: 180, hidden: !isPlatformAdmin.value },
  {
    key: 'ops',
    title: '操作',
    minWidth: 268,
    width: 280,
    align: 'center',
    fixed: 'right',
    tooltip: false,
    hidden: !isPlatformAdmin.value || (!canCreateMenu.value && !canEditMenu.value && !canDeleteMenu.value),
  },
])

function onMenuTableRowClick(row: MenuNode) {
  selectedMenuNode.value = row
}

watch(
  isPlatformAdmin,
  () => {
    void loadMenuBundlesForScope()
  },
  { immediate: true },
)

onMounted(() => {
  void store.loadTenantMenuRuntime({
    isPlatformAdmin: isPlatformAdmin.value,
    force: true,
  })
})
</script>

<template>
  <div class="page page-menu-mgmt">
    <el-alert
      v-if="perm.profile && !isPlatformAdmin"
      type="warning"
      show-icon
      :closable="false"
      title="当前主体只显示套餐与角色权限内的菜单；平台级菜单不会生成到租户侧。租户只能关闭/开启或更换显示名称。"
      class="page-alert"
    />
    <el-alert
      v-if="isPlatformAdmin"
      type="info"
      show-icon
      :closable="false"
      title="「归属」列便于辨认菜单与操作在后端权限模型中是「仅平台」还是「租户菜单」；实际边界由代码（如 PLATFORM_ONLY_PERMISSION_PATHS）与库表同步写入决定，此页不提供修改归属，避免与租户安全模型冲突。"
      class="page-alert"
    />
    <el-alert
      type="info"
      show-icon
      :closable="false"
      title="层级规则：目录下只能添加子目录或菜单；菜单下只能添加按钮（操作）。工具栏「新增子项」须在弹窗选择上级：可选「根节点」挂到侧栏顶级（与首页等同级），或选某一目录/菜单。从按钮行「增加子项」时，新按钮挂到该按钮所属菜单下。"
      class="page-alert"
    />
    <NeuroAgentListPage
      mode="el-table"
      :title="isPlatformAdmin ? '平台菜单管理（全量菜单结构）' : '租户菜单管理（套餐内菜单覆盖）'"
      :columns="columns"
      :data="menuTableTree"
      :show-create="false"
      :show-selection="false"
      :show-pagination="false"
      :filter-fields="menuFilterFields"
      :tree-props="{ children: 'children' }"
      :row-key="'id'"
      :expanded-row-keys="expandedRowKeys"
      :row-class-name="menuTableRowClassName"
      :default-expand-all="false"
      @expand-change="onTreeExpandChange"
      @row-click="onMenuTableRowClick"
      @search="onMenuSearch"
    >
      <template #search-trailing-actions>
        <el-button size="default" @click="expandAllTree">全部展开</el-button>
        <el-button size="default" @click="collapseAllTree">全部收起</el-button>
      </template>
      <template #actions>
        <el-button
          v-if="isPlatformAdmin"
          v-permission="'menu:create'"
          class="btn-gradient"
          title="将打开弹窗，需先选择上级目录或菜单"
          @click="openAddDlg"
        >
          新增子项
        </el-button>
        <el-button v-if="isPlatformAdmin" v-permission="'menu:delete'" @click="reset">恢复默认</el-button>
      </template>
      <template #col-sort="{ row }">
        <span class="sort-btns" @click.stop>
          <el-button
            link
            type="primary"
            class="sort-arrow-btn"
            :disabled="!moveState(row.id).canUp"
            title="上移"
            @click="moveUp(row.id)"
          >
            <el-icon>
              <ArrowUpBold v-if="sortUseBoldArrows(row)" />
              <ArrowUp v-else />
            </el-icon>
          </el-button>
          <el-button
            link
            type="primary"
            class="sort-arrow-btn"
            :disabled="!moveState(row.id).canDown"
            title="下移"
            @click="moveDown(row.id)"
          >
            <el-icon>
              <ArrowDownBold v-if="sortUseBoldArrows(row)" />
              <ArrowDown v-else />
            </el-icon>
          </el-button>
        </span>
      </template>
      <template #col-title="{ row }">
        <span v-if="editingTitleId !== row.id" :class="titleCellClass(row)">
          <el-tooltip v-if="row.icon && menuIconComponent(row.icon)" :content="row.icon" placement="top">
            <span class="menu-tbl-title-ico">
              <el-icon><component :is="menuIconComponent(row.icon)!" /></el-icon>
            </span>
          </el-tooltip>
          <span class="menu-tbl-title-text">{{ row.title }}</span>
          <el-button v-if="canRename(row)" link type="primary" class="rename-btn" @click.stop="startTitleEdit(row)">
            改名
          </el-button>
        </span>
        <el-input
          v-else
          v-model="editingTitleValue"
          size="small"
          class="title-edit-input"
          @keyup.enter="submitTitleEdit(row)"
          @keyup.esc="cancelTitleEdit"
          @blur="submitTitleEdit(row)"
          @click.stop
        />
        <span v-if="row.children && row.children.length" class="title-expand" @click.stop="onTitleToggle(row)">
          <svg v-if="expandedRowKeys.includes(row.id)" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" width="14" height="14"><path d="M6 9l6 6 6-6"/></svg>
          <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" width="14" height="14"><path d="M9 6l6 6-6 6"/></svg>
        </span>
      </template>
      <template #col-sidebar="{ row }">
        <el-switch
          :model-value="row.enabled !== false"
          size="small"
          :disabled="!canToggle(row)"
          @change="(v: boolean) => onRowEnabled(row, v)"
          @click.stop
        />
      </template>
      <template #col-type="{ row }">{{ typeZh(row.type) }}</template>
      <template #col-scope="{ row }">
        <span v-for="t in [menuScopeTag(row)]" :key="`${row.id}-scope`">
          <el-tag v-if="t" size="small" :type="t.type">{{ t.text }}</el-tag>
          <span v-else class="scope-dash">—</span>
        </span>
      </template>
      <template #col-ops="{ row }">
        <span class="op-btns">
          <el-button v-permission="'menu:edit'" size="small" @click.stop="openEditRow(row)">编辑</el-button>
          <el-button v-permission="'menu:create'" size="small" @click.stop="openAddDlgUnderRow(row)">增加子项</el-button>
          <el-button
            v-if="row.id.startsWith('custom_menu_')"
            v-permission="'menu:delete'"
            type="danger"
            size="small"
            @click.stop="removeCustomRow(row)"
          >
            删除
          </el-button>
        </span>
      </template>
    </NeuroAgentListPage>

    <NeuroAgentDialog
      v-model="addDlg"
      title="新增菜单子项"
      icon="🧭"
      size="medium"
      confirm-text="确定"
      @confirm="confirmAdd"
      @close="onAddDlgClosed"
    >
      <div class="nm-form">
        <template v-if="addDialogNeedsGlobalParentPicker">
          <div class="nm-form-item">
            <label class="nm-form-label">上级节点</label>
            <el-select
              v-model="globalAddParentId"
              class="nm-form-control"
              filterable
              clearable
              placeholder="必选：根节点（顶级）或某一目录/菜单"
            >
              <el-option v-for="opt in parentPickerOptions" :key="opt.value" :label="opt.label" :value="opt.value" />
            </el-select>
          </div>
        </template>
        <div v-else class="nm-form-item nm-form-item--muted">
          <label class="nm-form-label">上级</label>
          <span v-if="effectiveDirectAddParent">
            {{ effectiveDirectAddParent.title }}（{{ typeZh(effectiveDirectAddParent.type) }}）
          </span>
          <span v-else>未解析到上级节点</span>
        </div>
        <div class="nm-form-item">
          <label class="nm-form-label">类型</label>
          <el-select v-model="addForm.nodeType" class="nm-form-control" :disabled="!allowedAddNodeTypes.length">
            <el-option v-if="allowedAddNodeTypes.includes('directory')" label="目录" value="directory" />
            <el-option v-if="allowedAddNodeTypes.includes('menu')" label="菜单" value="menu" />
            <el-option v-if="allowedAddNodeTypes.includes('button')" label="按钮（操作）" value="button" />
          </el-select>
        </div>
        <div class="nm-form-item">
          <label class="nm-form-label">名称</label>
          <el-input v-model="addForm.title" placeholder="显示名称" />
        </div>
        <div v-if="addForm.nodeType === 'menu'" class="nm-form-item">
          <label class="nm-form-label">路由</label>
          <el-input v-model="addForm.path" placeholder="须以 / 开头，如 /reports" />
        </div>
        <div v-if="addForm.nodeType === 'menu'" class="nm-form-item">
          <label class="nm-form-label">数据权限需求类型</label>
          <el-select v-model="addForm.dataPermMode" class="nm-form-control">
            <el-option label="不需要（NONE）" value="NONE" />
            <el-option label="组织架构权限（ORG）" value="ORG" />
            <el-option label="业务单元权限（BU）" value="BU" />
            <el-option label="组织 + 业务单元（ORG_BU）" value="ORG_BU" />
          </el-select>
        </div>
        <div v-if="addForm.nodeType === 'button'" class="nm-form-item">
          <label class="nm-form-label">权限码</label>
          <el-input v-model="addForm.permissionCode" placeholder="如 report:view" />
        </div>
        <div class="nm-form-item">
          <label class="nm-form-label">图标</label>
          <MenuIconSelect v-model="addForm.icon" placeholder="选择 Element Plus 图标" />
        </div>
        <p class="add-hint">
          目录下仅允许子目录或菜单；菜单下仅允许按钮（操作）。上级可选「根节点」以新增顶级目录或菜单（不可在根下直接加按钮）。从工具栏打开须先选上级。自定义节点可在表格「操作」列删除。
        </p>
      </div>
    </NeuroAgentDialog>

    <NeuroAgentDialog
      v-model="editDlg"
      title="编辑节点"
      icon="✏️"
      size="medium"
      confirm-text="保存"
      @confirm="confirmEdit"
      @close="onEditDlgClosed"
    >
      <div v-if="editTargetRow" class="nm-form">
        <div class="nm-form-item">
          <label class="nm-form-label">类型</label>
          <el-input :model-value="typeZh(editTargetRow.type)" disabled />
        </div>
        <div class="nm-form-item">
          <label class="nm-form-label">名称</label>
          <el-input v-model="editForm.title" placeholder="显示名称" />
        </div>
        <div v-if="editTargetRow.type === 'menu'" class="nm-form-item">
          <label class="nm-form-label">路由</label>
          <el-input v-model="editForm.path" placeholder="须以 / 开头" />
        </div>
        <div v-if="editTargetRow.type === 'menu'" class="nm-form-item">
          <label class="nm-form-label">数据权限需求类型</label>
          <el-select v-model="editForm.dataPermMode" class="nm-form-control">
            <el-option label="不需要（NONE）" value="NONE" />
            <el-option label="组织架构权限（ORG）" value="ORG" />
            <el-option label="业务单元权限（BU）" value="BU" />
            <el-option label="组织 + 业务单元（ORG_BU）" value="ORG_BU" />
          </el-select>
        </div>
        <div v-if="editTargetRow.type === 'button'" class="nm-form-item">
          <label class="nm-form-label">权限码</label>
          <el-input v-model="editForm.permissionCode" placeholder="如 report:view" />
        </div>
        <div class="nm-form-item">
          <label class="nm-form-label">图标</label>
          <MenuIconSelect v-model="editForm.icon" placeholder="选择图标" />
        </div>
      </div>
    </NeuroAgentDialog>
  </div>
</template>

<style scoped>
.page {
  padding: 16px;
}

/* 菜单管理页：隐藏内容区滚动条（保留滚动能力） */
.page-menu-mgmt {
  scrollbar-width: none;
}

.page-menu-mgmt::-webkit-scrollbar {
  width: 0;
  height: 0;
}

.page-menu-mgmt :deep(*) {
  scrollbar-width: none;
}

.page-menu-mgmt :deep(::-webkit-scrollbar) {
  width: 0;
  height: 0;
}

/* 菜单管理树表：列间留白，避免侧栏开关、类型、路由挤在一起 */
.page-menu-mgmt :deep(.neuro-el-table .el-table__cell .cell) {
  padding-left: 14px;
  padding-right: 14px;
}
.page-menu-mgmt :deep(.neuro-el-table .el-table__header .el-table__cell .cell) {
  padding-left: 14px;
  padding-right: 14px;
}

.add-hint {
  margin: 0;
  font-size: 12px;
  color: var(--el-text-color-secondary);
  line-height: 1.5;
}
.nm-form-control {
  width: 100%;
}
.nm-form-item--muted {
  font-size: 13px;
  color: var(--el-text-color-regular);
}
.menu-tbl-title-ico {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  font-size: 17px;
  color: var(--el-text-color-regular);
  cursor: default;
}
.menu-tbl-title-text {
  min-width: 0;
}
.op-btns {
  display: inline-flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: center;
  gap: 6px;
}
.page-alert {
  margin-bottom: 12px;
}
.sort-btns {
  display: inline-flex !important;
  align-items: center !important;
  gap: 0 !important;
  flex-direction: row !important;
}
.sort-arrow-btn {
  padding: 0 !important;
  min-width: auto !important;
  height: auto !important;
  display: inline-flex !important;
  align-items: center !important;
  justify-content: center !important;
}
/* 一级、二级箭头同字号；一级仅用 Bold 图标加粗，三级及以下更小略淡 */
:deep(.menu-tbl-row--l0 .sort-btns .sort-arrow-btn .el-icon),
:deep(.menu-tbl-row--l1 .sort-btns .sort-arrow-btn .el-icon) {
  font-size: 15px;
}
:deep(.menu-tbl-row--l2 .sort-btns .sort-arrow-btn .el-icon) {
  font-size: 12px;
  opacity: 0.82;
}

/* 隐藏首列（排序列）的树形展开图标 */
:deep(.el-table__body-wrapper .el-table__row > td:first-child .el-table__expand-icon) {
  display: none !important;
}
:deep(.el-table__body-wrapper .el-table__row > td:first-child .el-table__indent) {
  display: none !important;
}
:deep(.el-table__body-wrapper .el-table__row > td:first-child .el-table__placeholder) {
  display: none !important;
}

/* 名称列折叠箭头 */
.title-expand {
  display: inline-flex;
  align-items: center;
  margin-left: 8px;
  color: var(--el-text-color-secondary);
  cursor: pointer;
  transition: transform 0.2s;
  flex-shrink: 0;
}
.title-expand:hover {
  color: var(--el-primary-color);
}

/* 名称列缩进：二级 24px，三级 48px */
.menu-tbl-title-cell {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}
.rename-btn {
  height: 22px;
  padding: 0 2px;
}
.title-edit-input {
  width: min(220px, 100%);
}
:deep(.menu-tbl-row--l1 .menu-tbl-title-cell) {
  margin-left: 24px;
}
:deep(.menu-tbl-row--l2 .menu-tbl-title-cell) {
  margin-left: 48px;
}

/* 一级行名称：加粗并略大于正文（约 1px），层次清晰又不夸张 */
:deep(.menu-tbl-title--root) {
  font-weight: 600;
  font-size: 15px;
  color: var(--el-text-color-primary);
}
</style>
