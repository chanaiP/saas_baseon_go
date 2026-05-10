<script setup lang="ts">
import { ArrowDown, ArrowRight } from '@element-plus/icons-vue'
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'

import type { Plan, PlanCapabilityMatrixData, PlanCapabilityNode } from '@/api/plan'
import {
  expandedIdsForMatrixFoldLevel,
  flattenCapabilityNodes,
  inferMatrixFoldLevel,
  nodeStateForPlan,
  type MatrixFoldLevel,
  type PlanFeatureSelection,
  type PlanQuotaValues,
} from '@/composables/usePlanCapabilityMatrix'

const props = defineProps<{
  matrix: PlanCapabilityMatrixData
  featureSelection: PlanFeatureSelection
  quotaValues: PlanQuotaValues
  savingPlanId?: number | null
}>()

const emit = defineEmits<{
  toggle: [payload: { plan: Plan; node: PlanCapabilityNode; enabled: boolean; cascade?: boolean }]
  savePlan: [plan: Plan]
  planAction: [payload: { plan: Plan; action: 'edit' | 'copy' | 'toggle' | 'delete' }]
  editFeature: [featureId: number]
}>()

const rows = computed(() => flattenCapabilityNodes(props.matrix.nodes))
const matrixShellRef = ref<HTMLElement | null>(null)
const matrixShellWidth = ref(0)

const CAPABILITY_COLUMN_MIN = 300
const CAPABILITY_COLUMN_MAX = 380
const PLAN_COLUMN_MIN = 190
const PLAN_COLUMN_MAX = 360

let matrixResizeObserver: ResizeObserver | null = null

function clampColumnWidth(value: number, min: number, max: number) {
  return Math.min(max, Math.max(min, value))
}

function updateMatrixShellWidth() {
  matrixShellWidth.value = matrixShellRef.value?.clientWidth ?? 0
}

onMounted(() => {
  updateMatrixShellWidth()
  if (!matrixShellRef.value) return
  matrixResizeObserver = new ResizeObserver(updateMatrixShellWidth)
  matrixResizeObserver.observe(matrixShellRef.value)
})

onBeforeUnmount(() => {
  matrixResizeObserver?.disconnect()
  matrixResizeObserver = null
})

const matrixGridColumns = computed(() => {
  const planCount = props.matrix.plans.length
  const shellWidth = matrixShellWidth.value
  if (!planCount || shellWidth <= 0) {
    return `minmax(${CAPABILITY_COLUMN_MIN}px, ${CAPABILITY_COLUMN_MAX}px)`
  }
  const capabilityWidth = clampColumnWidth(Math.round(shellWidth * 0.2), CAPABILITY_COLUMN_MIN, CAPABILITY_COLUMN_MAX)
  const availableForPlans = Math.max(0, shellWidth - capabilityWidth)
  const planWidth = clampColumnWidth(Math.floor(availableForPlans / planCount), PLAN_COLUMN_MIN, PLAN_COLUMN_MAX)
  return `${capabilityWidth}px repeat(${planCount}, ${planWidth}px)`
})

/** 展开的节点 id（有子节点才可折叠；默认全部展开） */
const expandedIds = ref<Set<string>>(new Set())

watch(
  () => props.matrix.nodes,
  (nodes) => {
    // 默认展开到「菜单」行可见，不自动展开菜单下的操作/按钮
    expandedIds.value = new Set(expandedIdsForMatrixFoldLevel(nodes, 'menu'))
  },
  { deep: true, immediate: true },
)

function isRowVisible(row: { parentId: string | null }) {
  let pid: string | null = row.parentId
  const all = rows.value
  while (pid) {
    if (!expandedIds.value.has(pid)) return false
    const parent = all.find((r) => r.node.id === pid)
    pid = parent?.parentId ?? null
  }
  return true
}

const visibleRows = computed(() => rows.value.filter((row) => isRowVisible(row)))

function isExpanded(nodeId: string) {
  return expandedIds.value.has(nodeId)
}

function toggleFold(nodeId: string) {
  const next = new Set(expandedIds.value)
  if (next.has(nodeId)) next.delete(nodeId)
  else next.add(nodeId)
  expandedIds.value = next
}

function setMatrixFoldLevel(level: MatrixFoldLevel) {
  expandedIds.value = new Set(expandedIdsForMatrixFoldLevel(props.matrix.nodes, level))
}

const inferredFoldLevel = computed(() => inferMatrixFoldLevel(props.matrix.nodes, expandedIds.value))

/** 菜单层：仅业务域 ↔ 展示菜单行；若已展开操作/全部则先收到菜单层 */
function toggleMenuLayer() {
  const lv = inferredFoldLevel.value
  if (lv === 'domain') setMatrixFoldLevel('menu')
  else if (lv === 'menu') setMatrixFoldLevel('domain')
  else setMatrixFoldLevel('menu')
}

/** 操作层：在菜单层基础上展开/收起菜单下操作等子节点 */
function toggleOperationLayer() {
  const lv = inferredFoldLevel.value
  if (lv === 'domain' || lv === 'menu') setMatrixFoldLevel('operation')
  else setMatrixFoldLevel('menu')
}

/** 全部展开 ↔ 收起全部（回到仅顶层业务域） */
function toggleFullMatrix() {
  if (inferredFoldLevel.value === 'full') setMatrixFoldLevel('domain')
  else setMatrixFoldLevel('full')
}

const menuLayerToggleLabel = computed(() => (inferredFoldLevel.value === 'domain' ? '展开菜单' : '收起菜单'))
const menuLayerToggleTitle = computed(() =>
  inferredFoldLevel.value === 'domain'
    ? '展开业务域，展示其下菜单行'
    : inferredFoldLevel.value === 'menu'
      ? '仅展示业务域行（隐藏其下菜单）'
      : '收起至菜单层（隐藏菜单下的操作等子节点）',
)

const operationLayerToggleLabel = computed(() =>
  inferredFoldLevel.value === 'domain' || inferredFoldLevel.value === 'menu' ? '展开操作' : '收起操作',
)
const operationLayerToggleTitle = computed(() =>
  inferredFoldLevel.value === 'domain' || inferredFoldLevel.value === 'menu'
    ? '在菜单层基础上，展开各菜单下的操作等子节点'
    : '收起至菜单层（不展开操作子节点）',
)

const fullMatrixToggleLabel = computed(() => (inferredFoldLevel.value === 'full' ? '收起全部' : '展开全部'))
const fullMatrixToggleTitle = computed(() =>
  inferredFoldLevel.value === 'full'
    ? '收起全部展开，仅保留顶层业务域行'
    : '展开业务域、菜单及全部子层级',
)

function stateLabel(state: string) {
  if (state === 'enabled') return '已开'
  if (state === 'partial') return '部分'
  return '未开'
}

/** 类型胶囊文案 */
function typePillLabel(node: PlanCapabilityNode) {
  if (node.node_type === 'domain') return '业务域'
  if (node.node_type === 'group') return '目录'
  if (node.feature_type === 'BUTTON' && !String(node.feature_code || '').startsWith('button_')) {
    return '能力'
  }
  const map: Record<string, string> = {
    MENU: '菜单',
    BUTTON: '操作',
    API: '接入',
    SERVICE: '服务',
    CONFIG: '配置',
  }
  return map[node.feature_type || ''] || '能力'
}

/** 类型胶囊样式（与业务域 / 菜单 / 操作等区分） */
function typePillClass(node: PlanCapabilityNode) {
  if (node.node_type === 'domain' || node.node_type === 'group') return 'type-pill--domain'
  if (node.feature_type === 'MENU') return 'type-pill--menu'
  if (node.feature_type === 'BUTTON') return 'type-pill--op'
  if (node.feature_type === 'API') return 'type-pill--api'
  if (node.feature_type === 'SERVICE') return 'type-pill--service'
  if (node.feature_type === 'CONFIG') return 'type-pill--config'
  return 'type-pill--misc'
}

/** 名称下方的辅助说明（不再重复类型前缀） */
function nodeSubtitle(row: { node: PlanCapabilityNode; path: string[] }) {
  if (row.node.node_type === 'domain') {
    // 与侧栏菜单树对齐的业务域用「一级目录」；「其他能力」为套餐内非 MENU 树能力聚合，单独说明
    if (row.node.id === 'domain-other') return '特殊业务域分组'
    return '一级目录'
  }
  if (row.node.node_type === 'group') return row.path.join(' / ')
  return row.node.feature_code || row.path.slice(0, -1).join(' / ') || ''
}

function emitPlanAction(plan: Plan, action: string | number | object) {
  if (typeof action !== 'string') return
  emit('planAction', { plan, action: action as 'edit' | 'copy' | 'toggle' | 'delete' })
}

function canEditFeatureNode(node: PlanCapabilityNode) {
  if (node.node_type !== 'feature' || !node.feature_id) return false
  if (node.feature_type === 'MENU') return false
  if (node.feature_type === 'BUTTON') {
    return !String(node.feature_code || '').startsWith('button_')
  }
  return ['API', 'SERVICE', 'CONFIG'].includes(node.feature_type || '')
}

function editableFeatureTitle(node: PlanCapabilityNode) {
  return canEditFeatureNode(node)
    ? '编辑功能点'
    : '目录、菜单和操作由菜单管理维护；套餐列可控制开关'
}
</script>

<template>
  <section class="capability-matrix">
    <!-- 单层 grid + 每行 subgrid：列数变化时仍与表头共用轨道，避免多列挤压换行后左右纵不对齐 -->
    <div ref="matrixShellRef" class="matrix-shell">
      <div
        class="matrix-grid"
        :style="{
          gridTemplateColumns: matrixGridColumns,
        }"
      >
        <div class="matrix-row matrix-row--head">
          <div class="capability-head sticky-capability">
            <strong>业务能力</strong>
            <div class="fold-toolbar" aria-label="矩阵展开层级">
              <button
                type="button"
                class="fold-toolbar-btn"
                :title="menuLayerToggleTitle"
                @click="toggleMenuLayer"
              >
                {{ menuLayerToggleLabel }}
              </button>
              <button
                type="button"
                class="fold-toolbar-btn"
                :title="operationLayerToggleTitle"
                @click="toggleOperationLayer"
              >
                {{ operationLayerToggleLabel }}
              </button>
              <button
                type="button"
                class="fold-toolbar-btn"
                :title="fullMatrixToggleTitle"
                @click="toggleFullMatrix"
              >
                {{ fullMatrixToggleLabel }}
              </button>
            </div>
          </div>
          <div v-for="plan in matrix.plans" :key="plan.id" class="plan-head" :class="{ 'is-disabled-plan': plan.status !== 1 }">
            <div class="plan-title-line">
              <div>
                <div class="plan-name-line">
                  <strong>{{ plan.plan_name }}</strong>
                  <span v-if="plan.status !== 1" class="plan-status-pill">已停用</span>
                </div>
                <small>{{ plan.plan_code }}</small>
              </div>
              <el-dropdown trigger="click" @command="(action: string | number | object) => emitPlanAction(plan, action)">
                <button class="plan-menu-btn">操作</button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="edit">编辑套餐</el-dropdown-item>
                    <el-dropdown-item command="copy">复制套餐</el-dropdown-item>
                    <el-dropdown-item command="toggle">{{ plan.status === 1 ? '停用套餐' : '启用套餐' }}</el-dropdown-item>
                    <el-dropdown-item command="delete" divided>删除套餐</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
            <button
              type="button"
              class="save-link"
              :disabled="savingPlanId === plan.id"
              @click.stop="emit('savePlan', plan)"
            >
              {{ savingPlanId === plan.id ? '保存中' : '保存本列' }}
            </button>
          </div>
        </div>

        <div
          v-for="row in visibleRows"
          :key="row.node.id"
          class="matrix-row"
          :class="[`matrix-row--${row.node.node_type}`]"
        >
          <div class="capability-cell sticky-capability" :style="{ paddingLeft: `${12 + row.depth * 22}px` }">
            <div class="capability-main">
              <strong class="node-title">{{ row.node.label }}</strong>
              <span class="type-pill" :class="typePillClass(row.node)">{{ typePillLabel(row.node) }}</span>
              <button
                v-if="canEditFeatureNode(row.node)"
                type="button"
                class="feature-edit-btn"
                :title="editableFeatureTitle(row.node)"
                @click.stop="emit('editFeature', row.node.feature_id!)"
              >
                编辑功能点
              </button>
              <button
                v-if="row.hasChildren"
                type="button"
                class="fold-btn"
                :aria-expanded="isExpanded(row.node.id)"
                :title="isExpanded(row.node.id) ? '收起' : '展开'"
                @click.stop="toggleFold(row.node.id)"
              >
                <el-icon class="fold-icon">
                  <ArrowDown v-if="isExpanded(row.node.id)" />
                  <ArrowRight v-else />
                </el-icon>
              </button>
              <span v-else class="fold-spacer" aria-hidden="true" />
            </div>
            <small v-if="nodeSubtitle(row)" class="node-sub">{{ nodeSubtitle(row) }}</small>
          </div>

          <div v-for="plan in matrix.plans" :key="plan.id" class="plan-cell" :class="{ 'is-disabled-plan': plan.status !== 1 }">
            <div v-if="row.node.node_type === 'domain'" class="domain-action-row">
              <button
                type="button"
                class="domain-action-btn domain-action-btn--enable"
                @click="emit('toggle', { plan, node: row.node, enabled: true, cascade: true })"
              >
                全部开启
              </button>
              <button
                type="button"
                class="domain-action-btn domain-action-btn--disable"
                @click="emit('toggle', { plan, node: row.node, enabled: false, cascade: true })"
              >
                全部关闭
              </button>
            </div>
            <div v-else class="plan-cell-pill-row">
              <button
                class="state-pill"
                :class="nodeStateForPlan(row.node, plan.id, featureSelection)"
                @click="emit('toggle', {
                  plan,
                  node: row.node,
                  enabled: nodeStateForPlan(row.node, plan.id, featureSelection) !== 'enabled',
                })"
              >
                {{ stateLabel(nodeStateForPlan(row.node, plan.id, featureSelection)) }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.capability-matrix {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.fold-toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;
}

.fold-toolbar-btn {
  border: 1px solid rgba(120, 140, 190, 0.28);
  border-radius: 7px;
  background: rgba(255, 255, 255, 0.12);
  color: var(--plan-text, #344054);
  cursor: pointer;
  font-size: 11px;
  font-weight: 700;
  line-height: 1.2;
  padding: 5px 8px;
  white-space: nowrap;
}

.fold-toolbar-btn:hover {
  border-color: color-mix(in srgb, var(--plan-accent, #4778ff) 45%, transparent);
  color: var(--plan-accent, #3154c9);
}

.matrix-shell {
  max-width: 100%;
  overflow: auto;
  border: 1px solid rgba(120, 140, 190, 0.18);
  border-radius: var(--neuro-radius-2xl, 24px);
  background: var(--plan-panel-bg, rgba(255, 255, 255, 0.9));
  -ms-overflow-style: none;
  scrollbar-width: none;
}

.matrix-shell::-webkit-scrollbar {
  display: none;
  width: 0;
  height: 0;
}

.matrix-grid {
  display: grid;
  width: max-content;
  min-width: 100%;
  grid-auto-rows: auto;
  align-items: stretch;
}

.matrix-row {
  display: grid;
  grid-column: 1 / -1;
  grid-template-columns: subgrid;
  border-bottom: 1px solid rgba(120, 140, 190, 0.14);
}

.matrix-row:last-child {
  border-bottom: none;
}

.matrix-row--head {
  position: sticky;
  top: 0;
  z-index: 5;
  background: var(--plan-panel-strong, linear-gradient(180deg, #f8fbff, #eef4ff));
}

.sticky-capability {
  position: sticky;
  left: 0;
  z-index: 3;
  background: var(--plan-panel-bg, rgba(255, 255, 255, 0.94));
  border-right: none !important;
}

.matrix-row--head .sticky-capability {
  z-index: 7;
  background: var(--plan-panel-strong, linear-gradient(180deg, #f8fbff, #eef4ff));
}

.capability-head,
.plan-head,
.capability-cell,
.plan-cell {
  position: relative;
  padding: 14px 16px;
  border-right: 1px solid rgba(120, 140, 190, 0.12);
}

.capability-head {
  display: flex;
  min-width: 0;
  flex-direction: column;
  align-items: flex-start;
  justify-content: center;
  gap: 10px;
}

.capability-head > strong {
  color: var(--plan-text, #101828);
  font-size: 16px;
  font-weight: 700;
  line-height: 1.25;
}

.sticky-capability::after {
  position: absolute;
  top: 0;
  right: 0;
  bottom: 0;
  width: 1px;
  background: rgba(120, 140, 190, 0.12);
  content: '';
}

.plan-head {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.plan-head.is-disabled-plan,
.plan-cell.is-disabled-plan {
  filter: grayscale(0.9);
  opacity: 0.48;
}

.plan-head.is-disabled-plan {
  background: color-mix(in srgb, var(--nm-bg-elevated) 78%, var(--nm-bg-deep));
}

.plan-cell.is-disabled-plan {
  background: color-mix(in srgb, var(--nm-bg-elevated) 68%, var(--nm-bg-deep));
}

.plan-title-line {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 10px;
}

.plan-title-line > div {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 2px;
}

.plan-name-line {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 6px;
}

.plan-name-line strong {
  min-width: 0;
}

.plan-status-pill {
  flex: 0 0 auto;
  border: 1px solid rgba(148, 163, 184, 0.32);
  border-radius: var(--neuro-radius-full, 999px);
  background: rgba(148, 163, 184, 0.16);
  color: var(--plan-text-muted, #7a8599);
  font-size: 11px;
  font-weight: 800;
  line-height: 1;
  padding: 4px 7px;
  white-space: nowrap;
}

.plan-head small,
.node-sub {
  color: var(--plan-text-muted, #7a8599);
  font-size: 12px;
}

.plan-menu-btn {
  border: 1px solid rgba(120, 140, 190, 0.22);
  border-radius: var(--neuro-radius-full, 999px);
  background: rgba(255, 255, 255, 0.12);
  color: var(--plan-text, #344054);
  cursor: pointer;
  font-size: 12px;
  font-weight: 800;
  padding: 5px 9px;
  white-space: nowrap;
}

.save-link {
  border: none;
  background: none;
  color: var(--plan-accent, #4778ff);
  cursor: pointer;
  font-weight: 700;
  padding: 0;
  text-align: left;
}

.save-link:disabled {
  color: #94a3b8;
  cursor: default;
}

/* 粘性表头下避免与首行单元格叠层抢点击 */
.plan-head .save-link {
  position: relative;
  z-index: 4;
}

.capability-cell {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  justify-content: flex-start;
  gap: 4px;
  min-width: 0;
}

.capability-main {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
  min-width: 0;
  width: 100%;
}

.node-title {
  font-size: 15px;
  font-weight: 700;
  color: var(--plan-text, #101828);
  line-height: 1.35;
  min-width: 0;
  word-break: break-word;
}

.type-pill {
  flex-shrink: 0;
  padding: 4px 10px;
  border-radius: var(--neuro-radius-full, 999px);
  font-size: 12px;
  font-weight: 800;
  letter-spacing: 0.02em;
  line-height: 1.2;
}

.type-pill--domain {
  background: color-mix(in srgb, var(--plan-warning, #f59e0b) 22%, transparent);
  color: #9a3412;
  border: 1px solid color-mix(in srgb, var(--plan-warning, #f59e0b) 35%, transparent);
}

.type-pill--menu {
  background: color-mix(in srgb, var(--plan-accent, #4778ff) 26%, #ffffff);
  color: #0f172a;
  border: 1px solid color-mix(in srgb, var(--plan-accent, #4778ff) 42%, transparent);
}

.type-pill--op {
  background: rgba(15, 118, 110, 0.12);
  color: #0f766e;
  border: 1px solid rgba(15, 118, 110, 0.22);
}

.type-pill--api {
  background: rgba(99, 102, 241, 0.12);
  color: #4338ca;
  border: 1px solid rgba(99, 102, 241, 0.22);
}

.type-pill--service {
  background: rgba(124, 58, 237, 0.1);
  color: #6b21a8;
  border: 1px solid rgba(124, 58, 237, 0.2);
}

.type-pill--config {
  background: rgba(71, 85, 105, 0.1);
  color: #334155;
  border: 1px solid rgba(71, 85, 105, 0.2);
}

.type-pill--misc {
  background: rgba(100, 116, 139, 0.12);
  color: #475569;
  border: 1px solid rgba(100, 116, 139, 0.2);
}

.feature-edit-btn {
  border: 1px solid color-mix(in srgb, var(--plan-accent, #14dcc8) 45%, transparent);
  border-radius: 999px;
  background: color-mix(in srgb, var(--plan-accent, #14dcc8) 12%, transparent);
  color: var(--plan-accent, #14dcc8);
  cursor: pointer;
  font-size: 11px;
  font-weight: 800;
  line-height: 1;
  padding: 5px 8px;
  white-space: nowrap;
}

.feature-edit-btn:hover {
  background: color-mix(in srgb, var(--plan-accent, #14dcc8) 20%, transparent);
}

.fold-btn {
  display: inline-flex;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  width: 18px;
  height: 18px;
  padding: 0;
  border: 0;
  background: transparent;
  color: #ffffff;
  cursor: pointer;
}

.fold-btn:hover {
  color: var(--plan-accent, #00f5d4);
}

.fold-icon {
  font-size: 16px;
}

.fold-spacer {
  display: inline-block;
  flex-shrink: 0;
  width: 18px;
  height: 18px;
}

.matrix-row--domain .capability-cell {
  background: rgba(71, 120, 255, 0.08);
}

.matrix-row--group .capability-cell {
  background: color-mix(in srgb, var(--plan-accent, #4778ff) 9%, transparent);
}

.plan-cell {
  display: flex;
  min-height: 76px;
  flex-direction: column;
  align-items: flex-start;
  justify-content: center;
  gap: 8px;
}

.plan-cell-pill-row {
  display: flex;
  width: 100%;
  align-items: center;
  gap: 8px;
}

.domain-action-row {
  display: flex;
  align-items: center;
  gap: 6px;
}

.domain-action-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 62px;
  border: 1px solid rgba(120, 140, 190, 0.26);
  border-radius: 7px;
  cursor: pointer;
  font-size: 11px;
  font-weight: 800;
  line-height: 1.2;
  padding: 6px 8px;
  white-space: nowrap;
}

.domain-action-btn--enable {
  background: color-mix(in srgb, var(--plan-accent, #00f5d4) 18%, var(--plan-panel-strong, rgba(20, 25, 38, 0.96)));
  border-color: color-mix(in srgb, var(--plan-accent, #00f5d4) 44%, transparent);
  color: var(--plan-accent, #00f5d4);
}

.domain-action-btn--disable {
  background: color-mix(in srgb, var(--nm-bg-elevated) 86%, var(--nm-bg-deep));
  border-color: color-mix(in srgb, var(--nm-border) 90%, transparent);
  color: var(--nm-text-muted);
}

.state-pill {
  flex: 0 0 auto;
  width: fit-content;
  max-width: 100%;
  border: none;
  border-radius: var(--neuro-radius-full, 999px);
  cursor: pointer;
  font-weight: 800;
  padding: 7px 12px;
}

.state-pill.enabled {
  background: rgba(22, 163, 74, 0.12);
  color: #15803d;
}

.state-pill.partial {
  background: rgba(245, 158, 11, 0.14);
  color: #b45309;
}

.state-pill.disabled {
  background: color-mix(in srgb, var(--nm-bg-elevated) 72%, var(--nm-bg-deep));
  color: var(--nm-text-muted);
  border: 1px solid color-mix(in srgb, var(--nm-border) 90%, transparent);
}

</style>
