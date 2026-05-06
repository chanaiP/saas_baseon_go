<script setup lang="ts">
defineOptions({ name: 'BusinessUnitView' })

import type { ElTree } from 'element-plus'
import { ElMessage } from 'element-plus'
import { computed, nextTick, onMounted, ref, watch } from 'vue'

import {
  createBusinessUnit,
  deleteBusinessUnit,
  fetchBusinessUnitOrgMappings,
  fetchBusinessUnitPage,
  updateBusinessUnit,
} from '@/api/businessUnit'
import { archiveSuccessMessage, confirmArchiveAction } from '@/composables/useArchiveConfirm'
import { usePermissionStore } from '@/stores/permission'
import type { BusinessUnitRow } from '@/api/businessUnit'
import { fetchDictItemsByCode } from '@/api/dict'
import type { DictItemRow } from '@/api/dict'
import { fetchOrgTree } from '@/api/organization'
import type { OrgNode } from '@/api/organization'
import type { FilterField, TableColumn } from '@/views/components/NeuroAgentListPage.vue'
import NeuroAgentDialog from '@/views/components/NeuroAgentDialog.vue'
import NeuroAgentListPage from '@/views/components/NeuroAgentListPage.vue'

const permissionStore = usePermissionStore()

const loading = ref(false)
const rows = ref<BusinessUnitRow[]>([])
const total = ref(0)
const page = ref(1)
const limit = ref(20)
const keyword = ref('')
const filterStatus = ref<number | undefined>(undefined)

const orgTree = ref<OrgNode[]>([])
const orgTreeRef = ref<InstanceType<typeof ElTree>>()
const orgFilterText = ref('')

const buTypeItems = ref<DictItemRow[]>([])

/** 字典「自定义」项的 value（一般为 CUSTOM；也兼容仅中文标签的项） */
const buDictCustomValue = computed(() => {
  const exact = buTypeItems.value.find((i) => String(i.value).toUpperCase() === 'CUSTOM')
  if (exact) return String(exact.value)
  const byLabel = buTypeItems.value.find((i) => /自定义/.test(i.label))
  return byLabel ? String(byLabel.value) : 'CUSTOM'
})

/** 下拉当前选中的预设类型 value；选「自定义」时为字典中自定义项的 value */
const buTypePreset = ref('')
/** 选自定义时的自由文案，将写入后端的 bu_type */
const buCustomTypeText = ref('')

const isBuTypeCustom = computed(() => buTypePreset.value === buDictCustomValue.value)

const statusOptions = [
  { label: '启用', value: 1 },
  { label: '停用', value: 0 },
]

const statusFilterOptions = [{ label: '全部', value: '' as const }, ...statusOptions]

/** 单一弹窗：create | edit */
const dialogMode = ref<'create' | 'edit' | null>(null)
const dialogVisible = computed({
  get: () => dialogMode.value != null,
  set: (v: boolean) => {
    if (!v) dialogMode.value = null
  },
})

const currentEditId = ref<number | null>(null)
const form = ref({
  name: '',
  code: '',
  status: 1,
  remark: '' as string,
})

function flattenOrg(nodes: OrgNode[]): OrgNode[] {
  const out: OrgNode[] = []
  function walk(list: OrgNode[]) {
    for (const n of list) {
      out.push(n)
      if (n.children?.length) walk(n.children)
    }
  }
  walk(nodes)
  return out
}

const orgFlat = computed(() => flattenOrg(orgTree.value))

const buTypeLabelByValue = computed(() =>
  Object.fromEntries(buTypeItems.value.map((i) => [i.value, i.label] as const)),
)

function buTypeLabel(raw: string | null | undefined): string {
  if (!raw) return '—'
  const preset = buTypeLabelByValue.value[raw]
  if (preset) return preset
  return `自定义（${raw}）`
}

/** 从已保存的 bu_type 反推出下拉与自定义框 */
function hydrateBuTypeFields(stored: string) {
  const vals = new Set(buTypeItems.value.map((i) => i.value))
  const customV = buDictCustomValue.value
  const s = (stored || '').trim()
  if (s && vals.has(s) && s !== customV) {
    buTypePreset.value = s
    buCustomTypeText.value = ''
  } else if (s) {
    buTypePreset.value = customV
    buCustomTypeText.value = s
  } else {
    buTypePreset.value = ''
    buCustomTypeText.value = ''
  }
}

/** 提交用的 bu_type 字符串 */
function resolveBuTypeForApi(): string | null {
  if (!buTypePreset.value) return null
  if (buTypePreset.value === buDictCustomValue.value) {
    const t = buCustomTypeText.value.trim()
    return t.length ? t : null
  }
  return buTypePreset.value
}

function orgNodePath(nodeId: number): string {
  const parts: string[] = []
  let cur: OrgNode | undefined = orgFlat.value.find((n) => n.id === nodeId)
  const guard = new Set<number>()
  while (cur && !guard.has(cur.id)) {
    guard.add(cur.id)
    parts.unshift(cur.name)
    const pid = cur.parent_id
    cur = pid != null ? orgFlat.value.find((n) => n.id === pid) : undefined
  }
  return parts.join(' / ') || `#${nodeId}`
}

const checkedOrgSummary = computed(() => {
  const keys = orgTreeRef.value?.getCheckedKeys(false) as TreeKey[]
  const ids = (keys ?? []).map((k) => Number(k)).filter((n) => Number.isFinite(n))
  return ids.map((id) => ({ id, path: orgNodePath(id) }))
})

type TreeKey = string | number

function filterOrgNode(value: string, data: OrgNode) {
  if (!value) return true
  const q = value.toLowerCase()
  return (
    data.name.toLowerCase().includes(q) || String(data.code ?? '')
      .toLowerCase()
      .includes(q)
  )
}

watch(orgFilterText, (v) => {
  orgTreeRef.value?.filter(v)
})

const filterFields: FilterField[] = [
  { key: 'keyword', label: '名称/编码', type: 'text', placeholder: '模糊查询' },
  { key: 'status', label: '状态', type: 'select', options: statusFilterOptions },
]

const columns: TableColumn[] = [
  { key: 'name', title: '名称', minWidth: 160 },
  { key: 'code', title: '编码', minWidth: 120 },
  { key: 'bu_type', title: '类型', minWidth: 100 },
  { key: 'remark', title: '备注', minWidth: 140 },
  { key: 'status', title: '状态', minWidth: 88 },
  { key: 'actions', title: '操作', width: 200, fixed: 'right', tooltip: false },
]

const showCreateBu = computed(() => permissionStore.canUseAction('business_unit:create'))

const dialogTitle = computed(() => (dialogMode.value === 'edit' ? '编辑业务单元' : '新增业务单元'))

async function loadBuTypes() {
  try {
    const r = await fetchDictItemsByCode('business_unit_type')
    buTypeItems.value = (r.items ?? []).filter((i) => i.enabled !== false)
    if (!buTypeItems.value.length) {
      ElMessage.warning('未找到字典「业务单元类型」，请在数据字典中维护 business_unit_type')
    }
  } catch {
    buTypeItems.value = []
    ElMessage.warning('加载业务单元类型字典失败')
  }
}

async function loadOrg() {
  orgTree.value = await fetchOrgTree()
}

async function load() {
  loading.value = true
  try {
    const skip = (page.value - 1) * limit.value
    const data = await fetchBusinessUnitPage({
      skip,
      limit: limit.value,
      keyword: keyword.value || undefined,
      status: filterStatus.value,
    })
    rows.value = data.items
    total.value = data.total
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '加载业务单元列表失败')
  } finally {
    loading.value = false
  }
}

function generateCode() {
  form.value.code = `BU_${Date.now().toString(36).toUpperCase()}`
}

function collectCheckedOrgIds(): number[] {
  const raw = orgTreeRef.value?.getCheckedKeys(false) as TreeKey[]
  const ids = (raw ?? []).map((k) => Number(k)).filter((n) => Number.isFinite(n) && n > 0)
  return [...new Set(ids)]
}

function openCreate() {
  dialogMode.value = 'create'
  currentEditId.value = null
  form.value = { name: '', code: '', status: 1, remark: '' }
  buTypePreset.value = ''
  buCustomTypeText.value = ''
  orgFilterText.value = ''
  void nextTick(() => {
    orgTreeRef.value?.setCheckedKeys([])
    orgTreeRef.value?.filter('')
  })
}

async function openEdit(row: BusinessUnitRow) {
  dialogMode.value = 'edit'
  currentEditId.value = row.id
  form.value = {
    name: row.name,
    code: row.code,
    status: row.status,
    remark: row.remark ?? '',
  }
  hydrateBuTypeFields(row.bu_type ?? '')
  orgFilterText.value = ''
  let mapIds: number[] = []
  try {
    const maps = await fetchBusinessUnitOrgMappings(row.id)
    mapIds = maps.map((m) => m.org_id)
  } catch {
    mapIds = []
  }
  await nextTick()
  orgTreeRef.value?.setCheckedKeys(mapIds)
  orgTreeRef.value?.filter('')
}

async function saveDialog() {
  if (!form.value.name.trim() || !form.value.code.trim()) {
    ElMessage.warning('请填写名称与编码')
    return
  }
  const buTypeFinal = resolveBuTypeForApi()
  if (!buTypeFinal) {
    ElMessage.warning(isBuTypeCustom.value ? '请填写自定义类型' : '请选择业务单元类型')
    return
  }
  const orgIds = collectCheckedOrgIds()
  if (orgIds.length < 1) {
    ElMessage.warning('请至少选择一个关联组织节点')
    return
  }
  try {
    if (dialogMode.value === 'create') {
      await createBusinessUnit({
        name: form.value.name.trim(),
        code: form.value.code.trim(),
        bu_type: buTypeFinal,
        org_node_ids: orgIds,
        status: form.value.status,
        remark: form.value.remark.trim() || null,
      })
    } else if (currentEditId.value) {
      await updateBusinessUnit(currentEditId.value, {
        name: form.value.name.trim(),
        code: form.value.code.trim(),
        bu_type: buTypeFinal,
        status: form.value.status,
        remark: form.value.remark.trim() || null,
        org_node_ids: orgIds,
      })
    }
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '保存失败')
    return
  }
  dialogMode.value = null
  await load()
}

async function removeRow(row: BusinessUnitRow) {
  try {
    await confirmArchiveAction({
      name: row.name,
      title: '归档业务单元',
      detail: '归档后组织映射会停用，业务历史、审计记录和编码追溯仍会保留。',
    })
  } catch {
    return
  }
  try {
    await deleteBusinessUnit(row.id)
    ElMessage.success(archiveSuccessMessage(row.name))
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '归档失败')
    return
  }
  if (currentEditId.value === row.id) {
    dialogMode.value = null
  }
  await load()
}

function onSearch(payload: { keyword?: string; filters: Record<string, unknown> }) {
  const topKw = String(payload.keyword ?? '').trim()
  const filterKw = String(payload.filters?.keyword ?? '').trim()
  keyword.value = topKw || filterKw
  const st = payload.filters?.status
  filterStatus.value = st === '' || st === undefined || st === null ? undefined : Number(st)
  page.value = 1
  void load()
}

onMounted(async () => {
  await loadBuTypes()
  await loadOrg()
  await load()
})
</script>

<template>
  <NeuroAgentListPage
    mode="el-table"
    title="业务单元"
    subtitle="创建经营归属口径并关联组织节点；配额占用、经营统计与权限可选由系统在启用状态下默认处理，无需手动开关计费/统计。"
    :columns="columns"
    :data="rows"
    :loading="loading"
    :total="total"
    :page="page"
    :page-size="limit"
    :page-sizes="[20, 50, 100]"
    :show-selection="false"
    :show-create="showCreateBu"
    :filter-fields="filterFields"
    @create="openCreate"
    @search="onSearch"
    @page-change="(p:number)=>{page=p;load()}"
    @page-size-change="(s:number)=>{limit=s;page=1;load()}"
  >
    <template #col-bu_type="{ row }">
      <span>{{ buTypeLabel((row as BusinessUnitRow).bu_type) }}</span>
    </template>
    <template #col-remark="{ row }">
      <span class="cell-remark">{{ (row as BusinessUnitRow).remark?.trim() ? (row as BusinessUnitRow).remark : '—' }}</span>
    </template>
    <template #col-status="{ row }">
      <el-tag
        size="small"
        effect="plain"
        round
        :type="Number(row.status) === 1 ? 'success' : undefined"
        :class="{ 'nm-status-pill--inactive': Number(row.status) !== 1 }"
      >
        {{ Number(row.status) === 1 ? '启用' : '停用' }}
      </el-tag>
    </template>
    <template #col-actions="{ row }">
      <span class="op-btns" @click.stop>
        <el-button v-permission="'business_unit:edit'" size="small" @click.stop="openEdit(row as BusinessUnitRow)">
          编辑
        </el-button>
        <el-button
          v-permission="'business_unit:delete'"
          type="danger"
          size="small"
          @click.stop="removeRow(row as BusinessUnitRow)"
        >
          删除
        </el-button>
      </span>
    </template>
  </NeuroAgentListPage>

  <NeuroAgentDialog
    v-model="dialogVisible"
    :title="dialogTitle"
    size="large"
    width="min(920px, 94vw)"
    @confirm="saveDialog"
  >
    <div class="nm-form bu-dialog-form">
      <div class="nm-form-section">
        <div class="nm-form-section-title">基本信息</div>
        <div class="nm-form-row bu-name-code-row">
          <div class="nm-form-item">
            <label class="nm-form-label">名称 <span class="nm-form-required">*</span></label>
            <el-input v-model="form.name" placeholder="如 华东一区" />
          </div>
          <div class="nm-form-item">
            <label class="nm-form-label">编码 <span class="nm-form-required">*</span></label>
            <div class="bu-code-row">
              <el-input v-model="form.code" placeholder="租户内唯一" />
              <button type="button" class="nm-btn nm-btn--ghost bu-code-row__btn" @click="generateCode">自动生成</button>
            </div>
          </div>
        </div>
        <div class="nm-form-row">
          <div class="nm-form-item">
            <label class="nm-form-label">类型 <span class="nm-form-required">*</span></label>
            <el-select
              v-model="buTypePreset"
              placeholder="请选择"
              filterable
              style="width: 100%"
              :teleported="false"
            >
              <el-option v-for="it in buTypeItems" :key="it.id" :label="it.label" :value="it.value" />
            </el-select>
            <div v-if="isBuTypeCustom" class="bu-custom-type">
              <label class="nm-form-label">自定义类型 <span class="nm-form-required">*</span></label>
              <el-input
                v-model="buCustomTypeText"
                maxlength="32"
                show-word-limit
                placeholder="请输入自定义类型标识，如 特渠经营单元"
              />
            </div>
          </div>
          <div class="nm-form-item">
            <label class="nm-form-label">状态 <span class="nm-form-required">*</span></label>
            <el-select v-model="form.status" style="width: 100%" :teleported="false">
              <el-option v-for="item in statusOptions" :key="item.value" :label="item.label" :value="item.value" />
            </el-select>
          </div>
        </div>
      </div>
      <div class="nm-form-section">
        <div class="nm-form-section-title">关联组织</div>
        <p class="nm-dialog-text nm-dialog-text--muted bu-org-hint">
          请选择该业务单元覆盖的组织范围。已被其他启用业务单元关联的节点不可重复选择。
        </p>
        <div class="bu-org-split">
          <div class="nm-form-item bu-org-split__tree">
            <label class="nm-form-label">组织架构 <span class="nm-form-required">*</span></label>
            <div class="org-tree-wrap">
              <el-input
                v-model="orgFilterText"
                class="org-tree-wrap__search"
                placeholder="搜索组织名称或编码"
                clearable
              />
              <div class="org-tree-wrap__scroll">
                <el-tree
                  ref="orgTreeRef"
                  :data="orgTree"
                  show-checkbox
                  node-key="id"
                  default-expand-all
                  :check-strictly="true"
                  :props="{ label: 'name', children: 'children' }"
                  :filter-node-method="filterOrgNode"
                />
              </div>
            </div>
          </div>
          <div class="nm-form-item bu-org-split__picked">
            <label class="nm-form-label">已选节点</label>
            <div class="org-picked-panel">
              <template v-if="checkedOrgSummary.length">
                <div class="checked-title">已选 {{ checkedOrgSummary.length }} 个</div>
                <ul class="checked-list">
                  <li v-for="x in checkedOrgSummary" :key="x.id">{{ x.path }}</li>
                </ul>
              </template>
              <p v-else class="bu-org-split__empty">勾选左侧组织后，在此查看路径</p>
            </div>
          </div>
        </div>
      </div>
      <div class="nm-form-section">
        <div class="nm-form-section-title">其他</div>
        <div class="nm-form-item">
          <label class="nm-form-label">备注</label>
          <el-input v-model="form.remark" type="textarea" :rows="3" placeholder="可选" />
        </div>
      </div>
    </div>
  </NeuroAgentDialog>
</template>

<style scoped>
.bu-name-code-row {
  align-items: flex-start;
}

.bu-custom-type {
  margin-top: 10px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.bu-org-hint {
  margin: -6px 0 0;
}

.bu-org-split {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(200px, 38%);
  gap: 14px;
  align-items: stretch;
}

@media (max-width: 640px) {
  .bu-org-split {
    grid-template-columns: 1fr;
  }
}

.bu-code-row {
  display: flex;
  align-items: stretch;
  gap: 8px;
  width: 100%;
}

.bu-code-row :deep(.el-input) {
  flex: 1;
  min-width: 0;
}

.bu-code-row__btn {
  flex-shrink: 0;
  white-space: nowrap;
  align-self: stretch;
  padding-left: 12px;
  padding-right: 12px;
  font-size: 13px;
}

.org-tree-wrap {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 220px;
  max-height: 320px;
  padding: 12px;
  border-radius: 10px;
  background: rgba(45, 55, 72, 0.2);
  border: 1px solid rgba(45, 55, 72, 0.4);
  min-width: 0;
}

.org-tree-wrap__search {
  flex-shrink: 0;
  margin-bottom: 10px;
}

.org-tree-wrap__scroll {
  flex: 1;
  min-height: 0;
  overflow: auto;
}

.bu-org-split__tree.nm-form-item {
  min-height: 0;
}

.org-picked-panel {
  flex: 1;
  min-height: 220px;
  max-height: 320px;
  overflow: auto;
  padding: 12px;
  border-radius: 10px;
  background: rgba(45,55, 72, 0.2);
  border: 1px solid rgba(45, 55, 72, 0.4);
  font-size: 13px;
  color: var(--el-text-color-regular);
}

html:not(.dark) .org-tree-wrap,
html:not(.dark) .org-picked-panel {
  background: rgba(0, 0, 0, 0.03);
  border-color: rgba(0, 0, 0, 0.1);
}

.org-tree-wrap__scroll :deep(.el-tree) {
  background: transparent;
  --el-tree-node-hover-bg-color: rgba(0, 245, 212, 0.08);
}

.org-tree-wrap__scroll :deep(.el-tree-node__content) {
  border-radius: 6px;
}

.org-tree-wrap__scroll :deep(.el-tree-node__label) {
  font-size: 14px;
  color: var(--el-text-color-primary);
}

.dark .org-tree-wrap__scroll :deep(.el-tree-node__label) {
  color: #e2e8f0;
}

.dark .org-picked-panel {
  color: #cbd5e1;
}

.bu-org-split__empty {
  margin: 0;
  font-size: 13px;
  color: var(--el-text-color-secondary);
  line-height: 1.5;
}

.dark .bu-org-split__empty {
  color: #64748b;
}

.checked-title {
  font-weight: 600;
  margin: 0 0 8px;
  color: var(--el-text-color-primary);
}

.dark .checked-title {
  color: #e2e8f0;
}

.checked-list {
  margin: 0;
  padding-left: 1.2em;
}

.checked-list li {
  margin-bottom: 4px;
  word-break: break-word;
}

.cell-remark {
  display: inline-block;
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
