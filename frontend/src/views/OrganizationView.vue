<script setup lang="ts">
defineOptions({ name: 'OrganizationView' })
import { Box, CaretBottom, CaretTop, FolderOpened, House, OfficeBuilding, Shop } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { computed, markRaw, onMounted, ref, shallowRef } from 'vue'

import { fetchDictItemsByCode } from '@/api/dict'
import type { DictItemRow } from '@/api/dict'
import {
  createOrgNode,
  deleteOrgNode,
  fetchOrgTree,
  updateOrgNode,
} from '@/api/organization'
import { archiveSuccessMessage, confirmArchiveAction } from '@/composables/useArchiveConfirm'
import type { OrgNode } from '@/api/organization'
import { usePermissionStore } from '@/stores/permission'
import type { FilterField, TableColumn } from '@/views/components/NeuroAgentListPage.vue'
import NeuroAgentDialog from '@/views/components/NeuroAgentDialog.vue'
import NeuroAgentListPage from '@/views/components/NeuroAgentListPage.vue'

const permStore = usePermissionStore()
const canCreateOrg = computed(() => permStore.canUseAction('org:create'))

const loading = ref(false)
const tree = shallowRef<OrgNode[]>([])
const tenantId = ref(1)

const appliedNameHint = ref('')
const appliedCompanyType = ref('')
const appliedNodeType = ref('')
const orgExpandMode = ref(1)
const manualExpandedRowKeys = ref<string[]>([])

const companyTypeOptions = ref<DictItemRow[]>([])
const statusDictItems = ref<DictItemRow[]>([])
const orgNodeTypeItems = ref<DictItemRow[]>([])

const nodeTypeLabelByValue = computed(() =>
  Object.fromEntries(orgNodeTypeItems.value.map((x) => [x.value, x.label])),
)
const companyTypeLabelByValue = computed(() =>
  Object.fromEntries(companyTypeOptions.value.flatMap((x) => [[x.value, x.label], [x.value.toLowerCase(), x.label]])),
)

function nodeTypeLabel(nt: string): string {
  return nodeTypeLabelByValue.value[nt] ?? nt
}

function companyTypeLabel(code: string | null | undefined): string {
  const k = String(code || '').trim()
  if (!k) return '—'
  return companyTypeLabelByValue.value[k] ?? companyTypeLabelByValue.value[k.toLowerCase()] ?? k
}

const orgFilterFields = computed<FilterField[]>(() => [
  { key: 'nameHint', label: '名称 / 编码', type: 'text', placeholder: '模糊匹配' },
  {
    key: 'nodeType',
    label: '节点类型',
    type: 'select',
    placeholder: '全部',
    options: [
      { label: '全部', value: '' },
      ...orgNodeTypeItems.value.map((x) => ({ label: x.label, value: x.value })),
    ],
  },
  {
    key: 'companyType',
    label: '公司类型',
    type: 'select',
    placeholder: '全部',
    options: [
      { label: '全部', value: '' },
      ...companyTypeOptions.value.map((x) => ({ label: x.label, value: x.value })),
    ],
  },
])

function filterOrgSubtree(nodes: OrgNode[], nameQ: string, typeQ: string, nodeTypeQ: string): OrgNode[] {
  const nq = nameQ.trim().toLowerCase()
  const tq = typeQ.trim()
  const ntq = nodeTypeQ.trim()
  const out: OrgNode[] = []
  for (const node of nodes) {
    const raw = node.children ?? []
    const children = raw.length ? filterOrgSubtree(raw, nameQ, typeQ, nodeTypeQ) : []
    const selfNodeTypeHit = !ntq || node.node_type === ntq
    const selfCompanyTypeHit = !tq || (requiresCompanyType(node.node_type) && String(node.company_type ?? '').trim() === tq)
    const nameHit =
      !nq ||
      node.name.toLowerCase().includes(nq) ||
      String(node.code ?? '')
        .toLowerCase()
        .includes(nq)
    const selfHit = selfNodeTypeHit && selfCompanyTypeHit && nameHit
    if (selfHit || children.length > 0) {
      const next: OrgNode = { ...node, children: children.length > 0 ? children : [] }
      out.push(next)
    }
  }
  return out
}

const displayTree = computed(() => {
  const nq = appliedNameHint.value.trim()
  const tq = appliedCompanyType.value.trim()
  const nt = appliedNodeType.value.trim()
  if (!nq && !tq && !nt) return tree.value
  return filterOrgSubtree(tree.value, nq, tq, nt)
})

function orgRowKeyStr(row: OrgNode) {
  return `${row.node_type}-${row.id}`
}

function collectExpandKeysDepth1(nodes: OrgNode[]): string[] {
  const keys: string[] = []
  for (const n of nodes) {
    if (n.children?.length) keys.push(orgRowKeyStr(n))
  }
  return keys
}

function collectExpandKeysAll(nodes: OrgNode[]): string[] {
  const keys: string[] = []
  const walk = (arr: OrgNode[]) => {
    for (const n of arr) {
      if (n.children?.length) {
        keys.push(orgRowKeyStr(n))
        walk(n.children)
      }
    }
  }
  walk(nodes)
  return keys
}

const orgExpandedRowKeys = computed(() => {
  const roots = displayTree.value
  if (orgExpandMode.value === -1) return manualExpandedRowKeys.value
  if (orgExpandMode.value === 0) return []
  if (orgExpandMode.value === 1) return collectExpandKeysDepth1(roots)
  return collectExpandKeysAll(roots)
})

function cycleOrgExpandMode() {
  if (orgExpandMode.value === -1) {
    orgExpandMode.value = 2
    manualExpandedRowKeys.value = []
    return
  }
  orgExpandMode.value = (orgExpandMode.value + 1) % 3
}

const orgExpandButtonLabel = computed(() => {
  if (orgExpandMode.value === 0) return '展开一层'
  if (orgExpandMode.value === 1) return '全部展开'
  return '全部收起'
})

const orgExpandButtonTitle = computed(() => {
  if (orgExpandMode.value === 0) return '仅显示根节点下的第一层子节点（推荐，加载更快）'
  if (orgExpandMode.value === 1) return '展开所有层级（节点很多时可能较慢）'
  return '收起为仅显示根节点'
})

function onOrgSearch(payload: { keyword: string; filters: Record<string, unknown> }) {
  appliedNameHint.value = String(payload.filters?.nameHint ?? '').trim()
  appliedCompanyType.value = String(payload.filters?.companyType ?? '').trim()
  appliedNodeType.value = String(payload.filters?.nodeType ?? '').trim()
  orgExpandMode.value = 2
  manualExpandedRowKeys.value = []
}

function onOrgExpandChange(row: OrgNode, expanded: OrgNode[] | boolean) {
  if (Array.isArray(expanded)) {
    orgExpandMode.value = -1
    manualExpandedRowKeys.value = expanded.map((r) => orgRowKeyStr(r))
    return
  }
  const nextKeys = new Set(orgExpandedRowKeys.value)
  orgExpandMode.value = -1
  const rowKey = orgRowKeyStr(row)
  if (expanded) {
    nextKeys.add(rowKey)
  } else {
    nextKeys.delete(rowKey)
  }
  manualExpandedRowKeys.value = Array.from(nextKeys)
}

function requiresCompanyType(nodeType: string): boolean {
  return ['group', 'company', 'store'].includes(String(nodeType || '').trim())
}

function orgStatusLabel(status: number) {
  const hit = statusDictItems.value.find((x) => x.value === String(status))
  return hit?.label ?? (status === 1 ? '启用' : '停用')
}

type NodeFormState = {
  /** 仅从行「添加子级」打开时为 true，锁定父级不可改 */
  lock_parent: boolean
  parent_id: number | null
  parent_hint: string
  node_type: string
  name: string
  code: string
  company_type: string
  status: number
}

const dlgNode = ref(false)
const nodeEditId = ref<number | null>(null)
const nodeSaving = ref(false)
const nodeForm = ref<NodeFormState>({
  lock_parent: false,
  parent_id: null,
  parent_hint: '',
  node_type: '',
  name: '',
  code: '',
  company_type: '',
  status: 1,
})

function flattenOrgOptions(nodes: OrgNode[], depth = 0): { id: number; label: string }[] {
  const pad = '　'.repeat(depth)
  const out: { id: number; label: string }[] = []
  for (const n of nodes) {
    out.push({
      id: n.id,
      label: `${pad}${nodeTypeLabel(n.node_type)} · ${n.name}`,
    })
    if (n.children?.length) out.push(...flattenOrgOptions(n.children, depth + 1))
  }
  return out
}

function findOrgNodeById(nodes: OrgNode[], targetId: number): OrgNode | null {
  for (const n of nodes) {
    if (n.id === targetId) return n
    const childHit = n.children?.length ? findOrgNodeById(n.children, targetId) : null
    if (childHit) return childHit
  }
  return null
}

function collectDescendantIds(node: OrgNode, out: Set<number>) {
  for (const ch of node.children || []) {
    out.add(ch.id)
    collectDescendantIds(ch, out)
  }
}

const flatParentOptions = computed(() => {
  const all = flattenOrgOptions(tree.value)
  if (nodeEditId.value == null) return all
  const current = findOrgNodeById(tree.value, nodeEditId.value)
  if (!current) return all
  const blocked = new Set<number>([current.id])
  collectDescendantIds(current, blocked)
  return all.filter((opt) => !blocked.has(opt.id))
})

async function load() {
  loading.value = true
  try {
    if (permStore.profile?.tenant_id != null) {
      tenantId.value = permStore.profile.tenant_id
    }
    const orgData = await fetchOrgTree()
    tree.value = markRaw(orgData)
  } finally {
    loading.value = false
  }
  Promise.all([
    fetchDictItemsByCode('company_type'),
    fetchDictItemsByCode('common_status'),
    fetchDictItemsByCode('org_node_type'),
  ])
    .then(([coDict, stDict, ntDict]) => {
      companyTypeOptions.value = coDict.items
      statusDictItems.value = stDict.items
      orgNodeTypeItems.value = ntDict.items
    })
    .catch(() => {
      /* 字典失败不影响树展示 */
    })
}

function openCreate(parent?: OrgNode | null) {
  nodeEditId.value = null
  if (parent) {
    nodeForm.value = {
      lock_parent: true,
      parent_id: parent.id,
      parent_hint: `${nodeTypeLabel(parent.node_type)} · ${parent.name}`,
      node_type: '',
      name: '',
      code: '',
      company_type: '',
      status: 1,
    }
  } else {
    nodeForm.value = {
      lock_parent: false,
      parent_id: null,
      parent_hint: '',
      node_type: '',
      name: '',
      code: '',
      company_type: '',
      status: 1,
    }
  }
  dlgNode.value = true
}

function openEdit(row: OrgNode) {
  nodeEditId.value = row.id
  nodeForm.value = {
    lock_parent: false,
    parent_id: row.parent_id ?? null,
    parent_hint: '',
    node_type: row.node_type,
    name: row.name,
    code: row.code || '',
    company_type: row.company_type || '',
    status: row.status,
  }
  dlgNode.value = true
}

async function saveNode() {
  if (nodeSaving.value) return
  const name = nodeForm.value.name.trim()
  if (!name) return ElMessage.warning('请填写名称')
  if (!nodeForm.value.node_type?.trim()) {
    return ElMessage.warning('请选择节点类型')
  }
  if (requiresCompanyType(nodeForm.value.node_type) && !nodeForm.value.company_type?.trim()) {
    return ElMessage.warning('该节点类型必须选择公司类型')
  }
  nodeSaving.value = true
  try {
    if (nodeEditId.value != null) {
      const body: Record<string, unknown> = {
        name,
        code: nodeForm.value.code.trim() || undefined,
        status: nodeForm.value.status,
        node_type: nodeForm.value.node_type.trim(),
        parent_id: nodeForm.value.parent_id ?? null,
      }
      if (requiresCompanyType(nodeForm.value.node_type)) {
        body.company_type = nodeForm.value.company_type || undefined
      }
      await updateOrgNode(tenantId.value, nodeEditId.value, body)
    } else {
      const body: Record<string, unknown> = {
        node_type: nodeForm.value.node_type,
        name,
        code: nodeForm.value.code.trim() || undefined,
        status: nodeForm.value.status,
      }
      if (nodeForm.value.parent_id != null) {
        body.parent_id = nodeForm.value.parent_id
      }
      if (requiresCompanyType(nodeForm.value.node_type)) {
        body.company_type = nodeForm.value.company_type || undefined
      }
      await createOrgNode(tenantId.value, body)
    }
    dlgNode.value = false
    ElMessage.success('已保存')
    await load()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '保存失败')
  } finally {
    nodeSaving.value = false
  }
}

async function removeNode(row: OrgNode) {
  await confirmArchiveAction({ name: row.name, title: '归档组织节点', detail: '仅无子节点、无用户关联的组织节点可归档；历史关系仍会保留。' })
  await deleteOrgNode(tenantId.value, row.id)
  ElMessage.success(archiveSuccessMessage(row.name))
  await load()
}

function rowKey(row: OrgNode) {
  return orgRowKeyStr(row)
}

function nodeTypeTag(t: string) {
  if (t === 'company') return 'primary'
  if (t === 'store') return 'warning'
  if (t === 'group') return 'danger'
  if (t === 'warehouse') return 'info'
  return 'success'
}

function orgNameIcon(t: string) {
  if (t === 'company') return OfficeBuilding
  if (t === 'store') return Shop
  if (t === 'group') return House
  if (t === 'warehouse') return Box
  return FolderOpened
}

function orgRowClassName({ row }: { row: OrgNode }) {
  return row.node_type === 'store' ? 'org-row--store' : ''
}

const dlgTitle = computed(() => {
  if (nodeEditId.value != null) return '编辑组织节点'
  return nodeForm.value.lock_parent ? '新建子组织节点' : '新建组织节点'
})

const columns: TableColumn[] = [
  // 自适应布局：仅用 minWidth 保底，剩余宽度由表格自动分配；操作列保留更宽空间。
  { key: 'name', title: '名称', minWidth: 320 },
  { key: 'code', title: '编码', minWidth: 140, align: 'left' },
  { key: 'node_type', title: '类型', minWidth: 120, align: 'center' },
  { key: 'company_type', title: '公司类型', minWidth: 140, align: 'center' },
  { key: 'status', title: '状态', minWidth: 120, align: 'center' },
  { key: 'actions', title: '操作', minWidth: 210, align: 'left' as const },
]

onMounted(load)
</script>

<template>
  <div class="page">
    <NeuroAgentListPage
      mode="el-table"
      title="组织架构"
      subtitle="节点类型来自字典「组织节点类型」，新建须在表单中点选类型（必选）；其中集团/公司/门店必须选择公司类型。编辑可调整类型（有子节点时会刷新子树归属公司）。"
      :columns="columns"
      :data="displayTree"
      :loading="loading"
      :show-create="false"
      :show-selection="false"
      :show-pagination="false"
      :tree-props="{ children: 'children' }"
      :default-expand-all="false"
      :expanded-row-keys="orgExpandedRowKeys"
      :row-key="rowKey"
      :row-class-name="orgRowClassName"
      :skip-client-sort="true"
      :filter-fields="orgFilterFields"
      @search="onOrgSearch"
      @expand-change="onOrgExpandChange"
    >
      <template #search-trailing-actions>
        <div class="org-search-trailing">
          <el-button
            type="default"
            class="org-tree-expand-btn"
            :title="orgExpandButtonTitle"
            :aria-label="orgExpandButtonLabel"
            @click="cycleOrgExpandMode"
          >
            <span>{{ orgExpandButtonLabel }}</span>
            <el-icon class="org-tree-expand-btn__icon">
              <CaretTop v-if="orgExpandMode === 2" />
              <CaretBottom v-else />
            </el-icon>
          </el-button>
        </div>
      </template>
      <template #actions>
        <el-button v-permission="'org:create'" class="btn-gradient" @click="openCreate(null)">新建</el-button>
      </template>
      <template #col-name="{ row }">
        <span class="org-name-cell">
          <el-icon class="org-name-cell__icon" :class="'org-name-cell__icon--' + row.node_type">
            <component :is="orgNameIcon(row.node_type)" />
          </el-icon>
          <span class="org-name-cell__text">{{ row.name }}</span>
        </span>
      </template>
      <template #col-node_type="{ row }">
        <el-tag size="small" effect="plain" :type="nodeTypeTag(row.node_type)">
          {{ nodeTypeLabel(row.node_type) }}
        </el-tag>
      </template>
      <template #col-company_type="{ row }">
        <template v-if="requiresCompanyType(row.node_type)">{{ companyTypeLabel(row.company_type) }}</template>
        <span v-else class="org-type-muted">—</span>
      </template>
      <template #col-status="{ row }">
        <el-tag
          size="small"
          effect="plain"
          round
          :type="row.status === 1 ? 'success' : undefined"
          :class="{ 'nm-status-pill--inactive': row.status !== 1 }"
        >
          {{ orgStatusLabel(row.status) }}
        </el-tag>
      </template>
      <template #col-actions="{ row }">
        <span class="op-btns" @click.stop>
          <el-button v-if="canCreateOrg" size="small" type="primary" plain @click.stop="openCreate(row)">
            添加子级
          </el-button>
          <el-button v-permission="'org:edit'" size="small" @click.stop="openEdit(row)">编辑</el-button>
          <el-button
            v-permission="'org:delete'"
            type="danger"
            size="small"
            plain
            @click.stop="removeNode(row)"
          >
            删除
          </el-button>
        </span>
      </template>
    </NeuroAgentListPage>

    <NeuroAgentDialog v-model="dlgNode" :title="dlgTitle" icon="🏢" size="medium">
      <p v-if="nodeEditId == null && nodeForm.lock_parent && nodeForm.parent_hint" class="org-dlg-hint">
        父级：<strong>{{ nodeForm.parent_hint }}</strong>
      </p>
      <p v-if="nodeEditId != null" class="org-dlg-hint org-dlg-hint--muted">
        编辑时可调整节点类型；有子节点时会同步刷新子树的归属公司信息；暂不调整层级。
      </p>
      <div class="nm-form">
        <div class="nm-form-item">
          <label class="nm-form-label">节点类型（必选）</label>
          <p v-if="!orgNodeTypeItems.length" class="org-type-muted org-node-type-empty">
            请先在数据字典中配置「组织节点类型」。
          </p>
          <div v-else class="org-node-type-picker" role="radiogroup" aria-label="节点类型">
            <button
              v-for="it in orgNodeTypeItems"
              :key="it.id"
              type="button"
              role="radio"
              class="org-node-type-chip"
              :class="{ 'org-node-type-chip--selected': nodeForm.node_type === it.value }"
              :aria-checked="nodeForm.node_type === it.value"
              @click="nodeForm.node_type = it.value"
            >
              <el-icon class="org-node-type-chip__icon">
                <component :is="orgNameIcon(it.value)" />
              </el-icon>
              <span class="org-node-type-chip__label">{{ it.label }}</span>
            </button>
          </div>
          <p v-if="orgNodeTypeItems.length && !nodeForm.node_type" class="org-node-type-required-hint">
            请点选一种类型
          </p>
        </div>
        <template v-if="!nodeForm.lock_parent">
          <div class="nm-form-item">
            <label class="nm-form-label">父级</label>
            <el-select
              v-model="nodeForm.parent_id"
              clearable
              filterable
              placeholder="不选表示根节点"
              style="width: 100%"
              :teleported="false"
            >
              <el-option
                v-for="opt in flatParentOptions"
                :key="opt.id"
                :label="opt.label"
                :value="opt.id"
              />
            </el-select>
          </div>
        </template>
        <div class="nm-form-item">
          <label class="nm-form-label">名称</label>
          <el-input v-model="nodeForm.name" />
        </div>
        <div class="nm-form-item">
          <label class="nm-form-label">编码</label>
          <el-input v-model="nodeForm.code" />
        </div>
        <div v-if="requiresCompanyType(nodeForm.node_type)" class="nm-form-item">
          <label class="nm-form-label">公司类型（必选）</label>
          <el-select v-model="nodeForm.company_type" placeholder="请选择" filterable style="width: 100%" :teleported="false">
            <el-option v-for="it in companyTypeOptions" :key="it.id" :label="it.label" :value="it.value" />
          </el-select>
        </div>
        <div class="nm-form-item">
          <label class="nm-form-label">状态</label>
          <el-radio-group v-if="statusDictItems.length" v-model="nodeForm.status">
            <el-radio v-for="it in statusDictItems" :key="it.id" :value="Number(it.value)">{{ it.label }}</el-radio>
          </el-radio-group>
          <el-radio-group v-else v-model="nodeForm.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">停用</el-radio>
          </el-radio-group>
        </div>
      </div>
      <template #footer-right>
        <button class="nm-btn nm-btn--primary" :disabled="nodeSaving" @click="saveNode">
          <span v-if="nodeSaving" class="nm-btn__spinner"></span>
          保存
        </button>
      </template>
    </NeuroAgentDialog>
  </div>
</template>

<style scoped>
.page {
  padding: 16px;
}

.org-search-trailing {
  display: flex;
  width: 100%;
  justify-content: flex-end;
  align-items: center;
}

.org-tree-expand-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.org-tree-expand-btn__icon {
  font-size: 14px;
}

.op-btns {
  display: inline-flex;
  flex-wrap: nowrap;
  justify-content: flex-start;
  align-items: center;
  gap: 8px;
  width: 100%;
}

.org-name-cell {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.org-name-cell__icon {
  flex-shrink: 0;
  font-size: 16px;
  color: var(--el-text-color-secondary);
}

.org-name-cell__icon--company {
  color: var(--el-color-primary);
}

.org-name-cell__icon--department {
  color: var(--el-color-success);
}

.org-name-cell__icon--store {
  color: var(--el-color-warning);
}

.org-name-cell__text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.org-type-muted {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.org-dlg-hint {
  margin: 0 0 12px;
  padding: 10px 12px;
  font-size: 13px;
  line-height: 1.5;
  color: var(--el-text-color-regular);
  background: var(--el-fill-color-light);
  border-radius: 8px;
}

.org-dlg-hint--muted {
  color: var(--el-text-color-secondary);
  background: var(--el-fill-color-blank);
  border: 1px solid var(--el-border-color-lighter);
}

.org-node-type-picker {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.org-node-type-chip {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 10px 14px;
  margin: 0;
  border-radius: 10px;
  border: 1px solid var(--el-border-color);
  background: var(--el-fill-color-blank);
  color: var(--el-text-color-primary);
  font-size: 14px;
  line-height: 1.2;
  cursor: pointer;
  transition:
    border-color 0.15s ease,
    background-color 0.15s ease,
    box-shadow 0.15s ease;
}

.org-node-type-chip:hover {
  border-color: var(--el-color-primary-light-5);
  background: var(--el-fill-color-light);
}

.org-node-type-chip:focus-visible {
  outline: 2px solid var(--el-color-primary);
  outline-offset: 2px;
}

.org-node-type-chip--selected {
  border-color: var(--el-color-primary);
  background: color-mix(in srgb, var(--el-color-primary) 12%, var(--el-fill-color-blank));
  box-shadow: 0 0 0 1px color-mix(in srgb, var(--el-color-primary) 35%, transparent);
}

.org-node-type-chip__icon {
  font-size: 18px;
  color: var(--el-text-color-secondary);
}

.org-node-type-chip--selected .org-node-type-chip__icon {
  color: var(--el-color-primary);
}

.org-node-type-chip__label {
  white-space: nowrap;
}

.org-node-type-required-hint {
  margin: 8px 0 0;
  font-size: 12px;
  color: var(--el-color-warning);
}

.org-node-type-empty {
  margin: 0;
}

:deep(.neuro-el-table tr.org-row--store > td.el-table__cell) {
  background-color: color-mix(in srgb, var(--el-color-warning) 10%, transparent) !important;
}
:deep(.neuro-el-table tr.org-row--store:hover > td.el-table__cell) {
  background-color: color-mix(in srgb, var(--el-color-warning) 16%, transparent) !important;
}
</style>
