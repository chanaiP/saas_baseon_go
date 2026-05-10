import type {
  CapabilityCellState,
  CapabilityQuotaValue,
  Plan,
  PlanCapabilityMatrixData,
  PlanCapabilityNode,
} from '@/api/plan'

export interface CapabilityFlatRow {
  node: PlanCapabilityNode
  depth: number
  path: string[]
  featureIds: number[]
  /** 父节点 id，根级为 null（用于折叠时判断祖先是否展开） */
  parentId: string | null
  /** 是否有子节点（用于显示折叠控件） */
  hasChildren: boolean
}

export type PlanFeatureSelection = Record<number, number[]>
export type PlanQuotaValues = Record<number, Record<number, number>>

/** 收集树上有子节点的节点 id（全部展开用） */
export function collectExpandableNodeIds(nodes: PlanCapabilityNode[]): string[] {
  const out: string[] = []
  for (const node of nodes) {
    const children = node.children || []
    if (children.length) {
      out.push(node.id, ...collectExpandableNodeIds(children))
    }
  }
  return out
}

/** 业务域 / 分组：展开后可见其下一层（通常为菜单行） */
export function collectStructuralDomainGroupIds(nodes: PlanCapabilityNode[]): string[] {
  const out: string[] = []
  function walk(list: PlanCapabilityNode[]) {
    for (const n of list) {
      const children = n.children || []
      if ((n.node_type === 'domain' || n.node_type === 'group') && children.length) {
        out.push(n.id)
        walk(children)
      } else if (children.length) {
        walk(children)
      }
    }
  }
  walk(nodes)
  return out
}

/** 菜单型节点：再展开可见其下操作等子节点 */
export function collectMenuFeatureIdsWithChildren(nodes: PlanCapabilityNode[]): string[] {
  const out: string[] = []
  function walk(list: PlanCapabilityNode[]) {
    for (const n of list) {
      const children = n.children || []
      if (n.feature_type === 'MENU' && children.length) {
        out.push(n.id)
      }
      if (children.length) walk(children)
    }
  }
  walk(nodes)
  return out
}

export type MatrixFoldLevel = 'domain' | 'menu' | 'operation' | 'full'

/** 能力矩阵折叠层级对应的 expandedIds（与 PlanCapabilityMatrix 行可见逻辑一致） */
export function expandedIdsForMatrixFoldLevel(nodes: PlanCapabilityNode[], level: MatrixFoldLevel): string[] {
  if (level === 'domain') return []
  const structural = collectStructuralDomainGroupIds(nodes)
  if (level === 'menu') return structural
  const menus = collectMenuFeatureIdsWithChildren(nodes)
  if (level === 'operation') return [...structural, ...menus]
  return collectExpandableNodeIds(nodes)
}

export function foldPresetExpandedIds(nodes: PlanCapabilityNode[], level: MatrixFoldLevel): Set<string> {
  return new Set(expandedIdsForMatrixFoldLevel(nodes, level))
}

export function sameExpandedIdSet(a: Set<string>, b: Set<string>): boolean {
  if (a.size !== b.size) return false
  for (const x of a) {
    if (!b.has(x)) return false
  }
  return true
}

/** 将当前展开集合与各预设比对，供能力矩阵复合按钮使用；非预设（手动点行折叠）时回退为 menu */
export function inferMatrixFoldLevel(nodes: PlanCapabilityNode[], expanded: Set<string>): MatrixFoldLevel {
  const order: MatrixFoldLevel[] = ['full', 'operation', 'menu', 'domain']
  for (const level of order) {
    if (sameExpandedIdSet(expanded, foldPresetExpandedIds(nodes, level))) return level
  }
  return 'menu'
}

export function flattenCapabilityNodes(
  nodes: PlanCapabilityNode[],
  depth = 0,
  path: string[] = [],
  parentId: string | null = null,
): CapabilityFlatRow[] {
  const rows: CapabilityFlatRow[] = []
  for (const node of nodes) {
    const nextPath = [...path, node.label]
    const children = node.children || []
    const hasChildren = children.length > 0
    rows.push({
      node,
      depth,
      path: nextPath,
      featureIds: collectFeatureIds(node),
      parentId,
      hasChildren,
    })
    rows.push(...flattenCapabilityNodes(children, depth + 1, nextPath, node.id))
  }
  return rows
}

export function collectFeatureIds(node: PlanCapabilityNode): number[] {
  const ids = new Set<number>()
  if (node.feature_id != null) ids.add(node.feature_id)
  for (const child of node.children || []) {
    for (const id of collectFeatureIds(child)) ids.add(id)
  }
  return [...ids].sort((a, b) => a - b)
}

export function collectChildFeatureIds(node: PlanCapabilityNode): number[] {
  const ids = new Set<number>()
  for (const child of node.children || []) {
    for (const id of collectFeatureIds(child)) ids.add(id)
  }
  return [...ids].sort((a, b) => a - b)
}

export function collectQuotaValues(node: PlanCapabilityNode, planId: number): CapabilityQuotaValue[] {
  const byId = new Map<number, CapabilityQuotaValue>()
  for (const cell of node.cells || []) {
    if (cell.plan_id !== planId) continue
    for (const quota of cell.quota_values || []) byId.set(quota.quota_id, quota)
  }
  for (const child of node.children || []) {
    for (const quota of collectQuotaValues(child, planId)) byId.set(quota.quota_id, quota)
  }
  return [...byId.values()].sort((a, b) => a.quota_id - b.quota_id)
}

export function selectionFromMatrix(matrix: PlanCapabilityMatrixData | null): PlanFeatureSelection {
  if (!matrix) return {}
  const rows = flattenCapabilityNodes(matrix.nodes)
  const out: PlanFeatureSelection = {}
  for (const plan of matrix.plans) {
    const ids = new Set<number>()
    for (const row of rows) {
      const cell = row.node.cells.find((item) => item.plan_id === plan.id)
      if (!cell?.enabled && cell?.state !== 'enabled') continue
      for (const id of cell?.feature_ids || []) ids.add(id)
    }
    out[plan.id] = [...ids].sort((a, b) => a - b)
  }
  return out
}

export function quotaValuesFromMatrix(matrix: PlanCapabilityMatrixData | null): PlanQuotaValues {
  if (!matrix) return {}
  const out: PlanQuotaValues = {}
  const rows = flattenCapabilityNodes(matrix.nodes)
  for (const plan of matrix.plans) {
    out[plan.id] = {}
    for (const row of rows) {
      for (const quota of collectQuotaValues(row.node, plan.id)) {
        out[plan.id][quota.quota_id] = quota.quota_value
      }
    }
  }
  return out
}

export function nodeStateForPlan(node: PlanCapabilityNode, planId: number, selection: PlanFeatureSelection): CapabilityCellState {
  const selected = new Set(selection[planId] || [])
  const childIds = collectChildFeatureIds(node)
  if (childIds.length) {
    const childEnabledCount = childIds.filter((id) => selected.has(id)).length
    const selfEnabled = node.feature_id != null && selected.has(node.feature_id)
    if (childEnabledCount === 0 && !selfEnabled) return 'disabled'
    if (childEnabledCount === childIds.length) return 'enabled'
    return 'partial'
  }

  const ids = node.feature_id != null ? [node.feature_id] : []
  if (!ids.length) return 'disabled'
  const enabledCount = ids.filter((id) => selected.has(id)).length
  if (enabledCount === 0) return 'disabled'
  if (enabledCount === ids.length) return 'enabled'
  return 'partial'
}

export function setNodeEnabled(
  selection: PlanFeatureSelection,
  node: PlanCapabilityNode,
  plan: Plan,
  enabled: boolean,
  options: { cascade?: boolean } = {},
): PlanFeatureSelection {
  const next = { ...selection }
  const ids = options.cascade
    ? collectFeatureIds(node)
    : node.feature_id != null
      ? [node.feature_id]
      : []
  const selected = new Set(next[plan.id] || [])
  for (const id of ids) {
    if (enabled) selected.add(id)
    else selected.delete(id)
  }
  next[plan.id] = [...selected].sort((a, b) => a - b)
  return next
}

export function patchPlanQuotaValues(values: PlanQuotaValues, planId: number, updates: Record<number, number>): PlanQuotaValues {
  return {
    ...values,
    [planId]: {
      ...(values[planId] || {}),
      ...updates,
    },
  }
}
