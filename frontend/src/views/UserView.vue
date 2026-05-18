<script setup lang="ts">
defineOptions({ name: 'UserView' })
import { CaretBottom, CaretTop } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { computed, markRaw, onMounted, ref, shallowRef } from 'vue'

import { fetchDictItemsByCode } from '@/api/dict'
import type { DictItemRow } from '@/api/dict'
import { fetchOrgTree } from '@/api/organization'
import type { OrgNode } from '@/api/organization'
import { fetchSysParamBatch } from '@/api/param'
import { fetchPositions, fetchPositionTypes } from '@/api/position'
import type { PositionRow, PositionTypeRow } from '@/api/position'
import { createUser, deleteUser, fetchAssignableRoles, fetchUsers, resetUserPassword, updateUser } from '@/api/user'
import type { UserRow } from '@/api/user'
import { archiveSuccessMessage, confirmArchiveAction } from '@/composables/useArchiveConfirm'
import type { FilterField, TableColumn } from '@/views/components/NeuroAgentListPage.vue'
import NeuroAgentDialog from '@/views/components/NeuroAgentDialog.vue'
import NeuroAgentListPage from '@/views/components/NeuroAgentListPage.vue'
import { isValidOptionalPhone, normalizePhoneInput, sanitizePhoneInput } from '@/utils/phone'

const loading = ref(false)
const org = shallowRef<OrgNode[]>([])
/** 左侧组织树：与 el-tree default-expand-all 联动（:key 重挂） */
const orgTreeExpanded = ref(true)
const companyId = ref<number | undefined>(undefined)
const departmentId = ref<number | undefined>(undefined)
/** 与列表「查询」同步；对应接口 keyword / status */
const appliedKeyword = ref('')
const appliedStatus = ref<number | undefined>(undefined)
const statusOptions = ref<DictItemRow[]>([])
const orgNodeTypeOptions = ref<DictItemRow[]>([])
const items = ref<UserRow[]>([])
const total = ref(0)
const page = ref(1)
const limit = ref(10)

const roles = ref<{ id: number; name: string }[]>([])
const positions = ref<PositionRow[]>([])
const positionTypes = ref<PositionTypeRow[]>([])
const positionKeyword = ref('')

const scopeLabel = ref('未限定组织（全部用户）')

const dlg = ref(false)
const edit = ref<UserRow | null>(null)

const resetResultDlg = ref(false)
const resetResult = ref<{ employee_no: string; phone: string; password: string } | null>(null)
const credentialDialogTitle = ref('密码已重置')
const credentialPasswordLabel = ref('重置后的密码')
const resetConfirmDlg = ref(false)
const resetConfirmRow = ref<UserRow | null>(null)
const resetConfirmLoading = ref(false)
const statusConfirmDlg = ref(false)
const statusConfirmRow = ref<UserRow | null>(null)
const statusConfirmLoading = ref(false)
const form = ref({
  employee_no: '',
  name: '',
  phone: '',
  role_ids: [] as number[],
  department_ids: [] as number[],
  position_ids: [] as number[],
})

function flattenDepartmentOptions(nodes: OrgNode[], path: string[] = []): { value: number; label: string }[] {
  const out: { value: number; label: string }[] = []
  for (const n of nodes) {
    const nextPath = [...path, n.name]
    const typeLabel = orgNodeTypeLabel(n.node_type)
    out.push({ value: n.id, label: `${nextPath.join(' · ')}（${typeLabel}）` })
    if (n.children?.length) out.push(...flattenDepartmentOptions(n.children, nextPath))
  }
  return out
}

type DepartmentTreeOption = {
  value: number
  label: string
  searchText: string
  disabled?: boolean
  children?: DepartmentTreeOption[]
}

type PositionCascaderOption = {
  value: number | string
  label: string
  searchText: string
  children?: PositionCascaderOption[]
}

function buildDepartmentTreeOptions(nodes: OrgNode[], path: string[] = []): DepartmentTreeOption[] {
  return nodes.map((n) => {
    const typeLabel = orgNodeTypeLabel(n.node_type)
    const nextPath = [...path, n.name]
    const item: DepartmentTreeOption = {
      value: n.id,
      label: `${n.name}（${typeLabel}）`,
      searchText: `${nextPath.join(' ')} ${n.name} ${n.code ?? ''} ${typeLabel} ${n.node_type ?? ''}`.toLowerCase(),
    }
    if (n.children?.length) item.children = buildDepartmentTreeOptions(n.children, nextPath)
    return item
  })
}

function filterDepartmentNode(keyword: string, data: DepartmentTreeOption) {
  const q = keyword.trim().toLowerCase()
  if (!q) return true
  return data.searchText.includes(q)
}

const departmentOptions = computed(() => flattenDepartmentOptions(org.value))
const departmentTreeOptions = computed(() => buildDepartmentTreeOptions(org.value))
const positionCascaderProps = {
  multiple: true,
  emitPath: false,
  value: 'value',
  label: 'label',
  children: 'children',
} as const
const positionCascaderOptions = computed<PositionCascaderOption[]>(() => {
  const typeById = new Map(positionTypes.value.map((t) => [t.id, t]))
  const positionsByType = new Map<number, PositionRow[]>()
  const unknownPositions: PositionRow[] = []

  for (const position of positions.value) {
    if (!typeById.has(position.position_type_id)) {
      unknownPositions.push(position)
      continue
    }
    const list = positionsByType.get(position.position_type_id) ?? []
    list.push(position)
    positionsByType.set(position.position_type_id, list)
  }

  const options = positionTypes.value
    .filter((type) => positionsByType.has(type.id))
    .map((type) => ({
      value: `type-${type.id}`,
      label: type.name,
      searchText: `${type.name} ${type.code}`.toLowerCase(),
      children: (positionsByType.get(type.id) ?? []).map((position) => ({
        value: position.id,
        label: position.name,
        searchText: `${type.name} ${type.code} ${position.name} ${position.code}`.toLowerCase(),
      })),
    }))

  if (unknownPositions.length) {
    options.push({
      value: 'type-unknown',
      label: '未分类',
      searchText: '未分类',
      children: unknownPositions.map((position) => ({
        value: position.id,
        label: position.name,
        searchText: `未分类 ${position.name} ${position.code}`.toLowerCase(),
      })),
    })
  }

  const q = positionKeyword.value.trim().toLowerCase()
  if (!q) return options
  const filtered: PositionCascaderOption[] = []
  for (const type of options) {
    const typeMatched = type.searchText.includes(q)
    const children = (type.children ?? []).filter((position) => typeMatched || position.searchText.includes(q))
    if (children.length) filtered.push({ ...type, children })
  }
  return filtered
})

function onPositionCascaderVisibleChange(visible: boolean) {
  if (!visible) positionKeyword.value = ''
}

const nodeTypeLabelByValue = computed(() =>
  Object.fromEntries(orgNodeTypeOptions.value.map((x) => [String(x.value).trim(), x.label])),
)

function orgNodeTypeLabel(nodeType: string | null | undefined): string {
  const t = String(nodeType ?? '').trim()
  if (!t) return '组织'
  return nodeTypeLabelByValue.value[t] ?? (
    t === 'company' ? '公司' : t === 'department' ? '部门' : t === 'store' ? '门店' : t
  )
}

const positionNameById = computed(() => Object.fromEntries(positions.value.map((p) => [p.id, p.name])))
const validPositionIdSet = computed(() => new Set(positions.value.map((p) => p.id)))

const roleNameById = computed(() => Object.fromEntries(roles.value.map((r) => [r.id, r.name])))

function normalizedSelectedPositionIDs(ids: Array<number | string>) {
  return ids
    .map((id) => Number(id))
    .filter((id) => Number.isFinite(id) && validPositionIdSet.value.has(id))
}

async function loadOrg() {
  try {
    org.value = markRaw(await fetchOrgTree())
  } catch {
    /** 无组织架构菜单权限、套餐未含 org_manage、或接口异常时不应阻断用户列表加载 */
    org.value = []
  }
}

async function loadRoles() {
  try {
    const r = await fetchAssignableRoles({ skip: 0, limit: 500 })
    roles.value = r.items.map((x) => ({ id: x.id, name: x.name }))
  } catch {
    roles.value = []
  }
}

async function loadPositions() {
  try {
    const r = await fetchPositions({ skip: 0, limit: 500 })
    positions.value = r.items
  } catch {
    positions.value = []
  }
}

async function loadPositionTypes() {
  try {
    const r = await fetchPositionTypes(0, 500)
    positionTypes.value = r.items
  } catch {
    positionTypes.value = []
  }
}

async function loadUsers() {
  loading.value = true
  try {
    const skip = (page.value - 1) * limit.value
    const res = await fetchUsers({
      skip,
      limit: limit.value,
      keyword: appliedKeyword.value || undefined,
      status: appliedStatus.value,
      company_id: companyId.value,
      department_id: departmentId.value,
    })
    items.value = res.items
    total.value = res.total
  } finally {
    loading.value = false
  }
}

function onOrgClick(data: OrgNode) {
  if (data.node_type === 'company') {
    companyId.value = data.id
    departmentId.value = undefined
    scopeLabel.value = `${orgNodeTypeLabel(data.node_type)}：${data.name}`
  } else {
    departmentId.value = data.id
    companyId.value = undefined
    scopeLabel.value = `${orgNodeTypeLabel(data.node_type)}：${data.name}`
  }
  page.value = 1
  loadUsers()
}

function clearOrg() {
  companyId.value = undefined
  departmentId.value = undefined
  scopeLabel.value = '未限定组织（全部用户）'
  page.value = 1
  loadUsers()
}

const userListTitle = computed(() => `${scopeLabel.value} · 用户列表`)

const userFilterFields = computed<FilterField[]>(() => {
  const statusOpts: { label: string; value: number | string }[] = [{ label: '全部', value: '' }]
  if (statusOptions.value.length) {
    for (const it of statusOptions.value) {
      statusOpts.push({ label: it.label, value: Number(it.value) })
    }
  } else {
    statusOpts.push({ label: '启用', value: 1 }, { label: '停用', value: 0 })
  }
  return [
    {
      key: 'keyword',
      label: '工号 / 姓名 / 手机',
      type: 'text',
      placeholder: '模糊匹配',
    },
    {
      key: 'status',
      label: '状态',
      type: 'select',
      placeholder: '全部',
      options: statusOpts,
    },
  ]
})

function openCreate() {
  if (!positionTypes.value.length) void loadPositionTypes()
  if (!positions.value.length) void loadPositions()
  if (!roles.value.length) void loadRoles()
  edit.value = null
  form.value = {
    employee_no: '',
    name: '',
    phone: '',
    role_ids: [],
    department_ids: [],
    position_ids: [],
  }
  dlg.value = true
}

function openEdit(row: UserRow) {
  if (row.can_edit === false) {
    ElMessage.warning('初始超级管理员不能编辑')
    return
  }
  edit.value = row
  const dIds =
    row.department_ids?.length ? [...row.department_ids] : row.department_id != null ? [row.department_id] : []
  form.value = {
    employee_no: row.employee_no,
    name: row.name,
    phone: row.phone || '',
    role_ids: [...row.role_ids],
    department_ids: dIds,
    position_ids: [...(row.position_ids ?? [])],
  }
  dlg.value = true
}

async function save() {
  if (!isValidOptionalPhone(form.value.phone)) {
    return ElMessage.warning('手机号需为 10-15 位数字')
  }
  const positionIds = normalizedSelectedPositionIDs(form.value.position_ids)
  try {
    if (edit.value) {
      const body: Parameters<typeof updateUser>[1] = {
        name: form.value.name,
        phone: form.value.phone ? normalizePhoneInput(form.value.phone) : null,
        role_ids: form.value.role_ids,
        company_id: edit.value.company_id,
        department_ids: form.value.department_ids,
        position_ids: positionIds,
        status: edit.value.status,
      }
      await updateUser(edit.value.id, body)
    } else {
      const primaryDepartmentId = form.value.department_ids[0]
      const created = await createUser({
        employee_no: form.value.employee_no,
        password: '',
        name: form.value.name,
        phone: form.value.phone ? normalizePhoneInput(form.value.phone) : undefined,
        company_id: companyId.value,
        department_id: primaryDepartmentId ?? null,
        department_ids: form.value.department_ids.length ? form.value.department_ids : undefined,
        position_ids: positionIds.length ? positionIds : undefined,
        role_ids: form.value.role_ids,
      })
      credentialDialogTitle.value = '用户已创建'
      credentialPasswordLabel.value = '初始密码'
      resetResult.value = {
        employee_no: created.employee_no,
        phone: (created.phone ?? '').trim(),
        password: created.initial_password || '',
      }
      resetResultDlg.value = true
    }
    dlg.value = false
    ElMessage.success('已保存')
    await loadUsers()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '保存用户失败')
  }
}

async function remove(row: UserRow) {
  if (row.can_delete === false) {
    ElMessage.warning('初始超级管理员不能删除')
    return
  }
  await confirmArchiveAction({ name: row.name || row.employee_no, title: '归档用户' })
  await deleteUser(row.id)
  ElMessage.success(archiveSuccessMessage(row.name || row.employee_no))
  await loadUsers()
}

function resetResultPlainText() {
  const r = resetResult.value
  if (!r) return ''
  const phoneLine = r.phone ? r.phone : '（未绑定）'
  return `工号：${r.employee_no}\n手机号：${phoneLine}\n密码：${r.password}`
}

function displayResetPhone() {
  const r = resetResult.value
  if (!r?.phone) return '—'
  return r.phone
}

async function copyText(text: string, okMsg: string) {
  try {
    await navigator.clipboard.writeText(text)
    ElMessage.success(okMsg)
  } catch {
    try {
      const ta = document.createElement('textarea')
      ta.value = text
      ta.style.position = 'fixed'
      ta.style.left = '-9999px'
      document.body.appendChild(ta)
      ta.select()
      document.execCommand('copy')
      document.body.removeChild(ta)
      ElMessage.success(okMsg)
    } catch {
      ElMessage.error('复制失败，请手动选中内容复制')
    }
  }
}

function copyResetPasswordOnly() {
  const r = resetResult.value
  if (!r) return
  void copyText(r.password, '已复制密码')
}

function copyResetAllInfo() {
  void copyText(resetResultPlainText(), '已复制全部信息')
}

function closeResetResultDlg() {
  resetResultDlg.value = false
  resetResult.value = null
}

function onResetResultDlgClose() {
  resetResult.value = null
}

function resetPwd(row: UserRow) {
  resetConfirmRow.value = row
  resetConfirmDlg.value = true
}

function closeResetConfirmDlg() {
  if (resetConfirmLoading.value) return
  resetConfirmDlg.value = false
  resetConfirmRow.value = null
}

async function confirmResetPassword() {
  const row = resetConfirmRow.value
  if (!row) return
  resetConfirmLoading.value = true
  try {
    const res = await resetUserPassword(row.id)
    credentialDialogTitle.value = '密码已重置'
    credentialPasswordLabel.value = '重置后的密码'
    resetResult.value = {
      employee_no: row.employee_no,
      phone: (row.phone ?? '').trim(),
      password: res.new_password,
    }
    resetConfirmDlg.value = false
    resetConfirmRow.value = null
    resetResultDlg.value = true
    ElMessage.success('密码已重置')
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '重置密码失败')
  } finally {
    resetConfirmLoading.value = false
  }
}

function openStatusConfirm(row: UserRow) {
  if (row.can_edit === false) {
    ElMessage.warning('初始超级管理员不能停用')
    return
  }
  statusConfirmRow.value = row
  statusConfirmDlg.value = true
}

function closeStatusConfirmDlg() {
  if (statusConfirmLoading.value) return
  statusConfirmDlg.value = false
  statusConfirmRow.value = null
}

async function confirmToggleStatus() {
  const row = statusConfirmRow.value
  if (!row) return
  const nextStatus = row.status === 1 ? 0 : 1
  statusConfirmLoading.value = true
  try {
    await updateUser(row.id, { status: nextStatus })
    statusConfirmDlg.value = false
    statusConfirmRow.value = null
    ElMessage.success(nextStatus === 1 ? '用户已启用' : '用户已停用')
    await loadUsers()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '更新用户状态失败')
  } finally {
    statusConfirmLoading.value = false
  }
}

function userStatusLabel(s: number) {
  const hit = statusOptions.value.find((x) => x.value === String(s))
  return hit?.label ?? (s === 1 ? '启用' : '停用')
}

function onPageChange(p: number) {
  page.value = p
  void loadUsers()
}

function onPageSizeChange(s: number) {
  limit.value = s
  page.value = 1
  void loadUsers()
}

function onUserSearch(payload: { keyword: string; filters: Record<string, unknown> }) {
  appliedKeyword.value = String(payload.keyword || payload.filters?.keyword || payload.filters?.userKeyword || '').trim()
  const raw = payload.filters?.status
  if (raw === '' || raw === null || raw === undefined) {
    appliedStatus.value = undefined
  } else {
    appliedStatus.value = Number(raw)
  }
  page.value = 1
  void loadUsers()
}

function formatDeptCell(row: UserRow) {
  const ids = row.department_ids?.length ? row.department_ids : row.department_id != null ? [row.department_id] : []
  if (!ids.length) return '—'
  return ids
    .map((id, index) => row.department_names?.[index] || departmentOptions.value.find((o) => o.value === id)?.label || `#${id}`)
    .join('；')
}

function formatPositionCell(row: UserRow) {
  const ids = row.position_ids ?? []
  if (!ids.length) return '—'
  return ids.map((id) => positionNameById.value[id] ?? `#${id}`).join('；')
}

function formatRolesTitle(row: UserRow) {
  if (row.is_platform_admin) return '平台超级管理员'
  if (row.is_tenant_admin) return '租户超级管理员'
  const ids = row.role_ids ?? []
  if (!ids.length) return ''
  return ids.map((id) => roleNameById.value[id] ?? `#${id}`).join('；')
}

const columns: TableColumn[] = [
  { key: 'employee_no', title: '工号', minWidth: 136 },
  { key: 'name', title: '姓名', minWidth: 112 },
  { key: 'departments', title: '部门', minWidth: 132 },
  { key: 'positions', title: '岗位', minWidth: 132 },
  { key: 'roles', title: '角色', minWidth: 178 },
  { key: 'phone', title: '手机', minWidth: 136 },
  { key: 'status', title: '状态', minWidth: 92 },
  { key: 'actions', title: '操作', minWidth: 364, tooltip: false },
]

onMounted(async () => {
  await loadOrg()
  await Promise.all([loadRoles(), loadPositionTypes(), loadPositions()])
  const [st, orgTypeDict, params] = await Promise.all([
    fetchDictItemsByCode('common_status'),
    fetchDictItemsByCode('org_node_type'),
    fetchSysParamBatch(['user.list_default_page_size']),
  ])
  statusOptions.value = st.items
  orgNodeTypeOptions.value = orgTypeDict.items
  const ps = params.values['user.list_default_page_size']
  if (ps) {
    const n = parseInt(ps, 10)
    if (!Number.isNaN(n) && n > 0) limit.value = n
  }
  await loadUsers()
})
</script>

<template>
  <div class="pro-page page page--list user-mgmt-root">
    <div class="user-mgmt-layout">
      <aside class="user-mgmt-aside">
        <el-card class="user-mgmt-aside-card" shadow="never">
          <template #header>
            <div class="user-mgmt-aside-header">
              <span class="user-mgmt-aside-title">组织架构</span>
              <el-button
                type="default"
                size="small"
                class="user-mgmt-aside-expand-btn"
                :disabled="!org.length"
                :title="orgTreeExpanded ? '全部收起' : '全部展开'"
                :aria-label="orgTreeExpanded ? '全部收起' : '全部展开'"
                @click="orgTreeExpanded = !orgTreeExpanded"
              >
                <span>{{ orgTreeExpanded ? '全部收起' : '全部展开' }}</span>
                <el-icon class="user-mgmt-aside-expand-btn__icon">
                  <CaretTop v-if="orgTreeExpanded" />
                  <CaretBottom v-else />
                </el-icon>
              </el-button>
            </div>
          </template>
          <div class="aside-toolbar">
            <el-button link type="primary" @click="clearOrg">查看全部</el-button>
          </div>
          <div class="aside-tree-wrap">
            <el-tree
              :key="'user-org-tree-' + String(orgTreeExpanded)"
              :data="org"
              node-key="id"
              :default-expand-all="orgTreeExpanded"
              :props="{ label: 'name', children: 'children' }"
              @node-click="onOrgClick"
            />
          </div>
        </el-card>
      </aside>
      <main class="user-mgmt-main">
        <NeuroAgentListPage
          class="user-mgmt-list"
          mode="el-table"
          :title="userListTitle"
          :columns="columns"
          :data="items"
          :loading="loading"
          :total="total"
          :page="page"
          :page-size="limit"
          :page-sizes="[10, 20, 50]"
          :show-create="true"
          :show-selection="false"
          :has-filters="false"
          :filter-fields="userFilterFields"
          @create="openCreate"
          @search="onUserSearch"
          @page-change="onPageChange"
          @page-size-change="onPageSizeChange"
        >
          <template #actions>
            <el-button v-permission="'user:create'" class="btn-gradient" @click="openCreate">新增用户</el-button>
          </template>
          <template #col-departments="{ row }">
            <span class="cell-multi-line" :title="formatDeptCell(row)">{{ formatDeptCell(row) }}</span>
          </template>
          <template #col-positions="{ row }">
            <span class="cell-multi-line" :title="formatPositionCell(row)">{{ formatPositionCell(row) }}</span>
          </template>
          <template #col-roles="{ row }">
            <div class="role-tags-cell" :title="formatRolesTitle(row)">
              <el-tag
                v-if="row.is_platform_admin || row.is_tenant_admin"
                size="small"
                effect="plain"
                class="role-tag-chip role-tag-chip--admin"
              >
                {{ row.is_platform_admin ? '平台超级管理员' : '租户超级管理员' }}
              </el-tag>
              <template v-if="(row.role_ids ?? []).length">
                <el-tag
                  v-for="rid in row.role_ids"
                  :key="`${row.id}-${rid}`"
                  size="small"
                  effect="plain"
                  class="role-tag-chip"
                >
                  {{ roleNameById[rid] ?? `#${rid}` }}
                </el-tag>
              </template>
              <span v-else-if="!row.is_platform_admin && !row.is_tenant_admin" class="muted">—</span>
            </div>
          </template>
          <template #col-status="{ row }">
            <el-tag
              effect="plain"
              size="small"
              round
              :type="row.status === 1 ? 'success' : undefined"
              :class="{ 'nm-status-pill--inactive': row.status !== 1 }"
            >{{ userStatusLabel(row.status) }}</el-tag>
          </template>
          <template #col-actions="{ row }">
            <span class="op-btns user-mgmt-op-btns">
              <el-button v-permission="'user:edit'" size="small" :disabled="row.can_edit === false" @click="openEdit(row)">编辑</el-button>
              <el-button v-permission="'user:reset_password'" size="small" @click="resetPwd(row)">重置密码</el-button>
              <el-button
                v-permission="'user:edit'"
                size="small"
                :disabled="row.can_edit === false"
                @click="openStatusConfirm(row)"
              >
                {{ row.status === 1 ? '停用' : '启用' }}
              </el-button>
              <el-button v-permission="'user:delete'" size="small" :disabled="row.can_delete === false" type="danger" @click="remove(row)">删除</el-button>
            </span>
          </template>
        </NeuroAgentListPage>
      </main>
    </div>

    <NeuroAgentDialog v-model="dlg" :title="edit ? '编辑用户' : '新增用户'" icon="👤" size="large">
      <div class="nm-form">
        <div class="nm-form-item" v-if="!edit">
          <label class="nm-form-label">工号</label>
          <el-input v-model="form.employee_no" />
        </div>
        <div class="nm-form-row">
          <div class="nm-form-item">
            <label class="nm-form-label">姓名</label>
            <el-input v-model="form.name" />
          </div>
          <div class="nm-form-item">
            <label class="nm-form-label">手机</label>
            <el-input
              v-model="form.phone"
              inputmode="tel"
              placeholder="请输入手机号"
              @input="form.phone = sanitizePhoneInput(String($event))"
            />
          </div>
        </div>
        <div class="nm-form-item">
          <label class="nm-form-label">任职组织</label>
          <el-tree-select
            v-model="form.department_ids"
            :data="departmentTreeOptions"
            multiple
            filterable
            node-key="value"
            :props="{ label: 'label', children: 'children', disabled: 'disabled' }"
            check-strictly
            :filter-node-method="filterDepartmentNode"
            collapse-tags
            collapse-tags-tooltip
            render-after-expand
            default-expand-all
            placeholder="搜索并选择任职组织；多项时第一项为主组织"
            style="width: 100%"
          />
        </div>
        <div class="nm-form-item">
          <label class="nm-form-label">岗位</label>
          <el-cascader
            v-model="form.position_ids"
            :options="positionCascaderOptions"
            :props="positionCascaderProps"
            clearable
            collapse-tags
            collapse-tags-tooltip
            :show-all-levels="false"
            placeholder="按岗位类型选择岗位"
            style="width: 100%"
            @visible-change="onPositionCascaderVisibleChange"
          >
            <template #header>
              <div class="position-cascader-search">
                <el-input
                  v-model="positionKeyword"
                  clearable
                  placeholder="搜索岗位类型、岗位名称或编码"
                  @click.stop
                  @keydown.stop
                />
              </div>
            </template>
            <template #empty>
              <div class="position-cascader-empty">未找到匹配岗位</div>
            </template>
          </el-cascader>
        </div>
        <div class="nm-form-item">
          <label class="nm-form-label">角色</label>
          <el-select
            v-model="form.role_ids"
            multiple
            filterable
            collapse-tags
            collapse-tags-tooltip
            placeholder="可多选，多个角色权限叠加"
            style="width: 100%"
          >
            <el-option v-for="r in roles" :key="r.id" :label="r.name" :value="r.id" />
          </el-select>
        </div>
      </div>
      <template #footer-right>
        <button class="nm-btn nm-btn--primary" @click="save">保存</button>
      </template>
    </NeuroAgentDialog>

    <NeuroAgentDialog
      v-model="statusConfirmDlg"
      :title="statusConfirmRow?.status === 1 ? '停用用户' : '启用用户'"
      :icon="statusConfirmRow?.status === 1 ? '⏸️' : '▶️'"
      size="small"
      :confirm-text="statusConfirmRow?.status === 1 ? '确认停用' : '确认启用'"
      cancel-text="取消"
      :loading="statusConfirmLoading"
      @confirm="confirmToggleStatus"
      @cancel="closeStatusConfirmDlg"
      @close="closeStatusConfirmDlg"
    >
      <div class="user-status-confirm-panel">
        <p v-if="statusConfirmRow?.status === 1">停用后该用户将无法登录，当前登录会话也会失效。</p>
        <p v-else>启用后该用户可以重新登录系统。</p>
        <div v-if="statusConfirmRow" class="user-status-confirm-target">
          <span>{{ statusConfirmRow.name || statusConfirmRow.employee_no }}</span>
          <small>{{ statusConfirmRow.employee_no }}</small>
        </div>
      </div>
    </NeuroAgentDialog>

    <NeuroAgentDialog
      v-model="resetConfirmDlg"
      title="重置密码"
      icon="⚠️"
      size="small"
      confirm-text="确认重置"
      cancel-text="取消"
      :loading="resetConfirmLoading"
      @confirm="confirmResetPassword"
      @cancel="closeResetConfirmDlg"
      @close="closeResetConfirmDlg"
    >
      <div class="reset-confirm-panel">
        <p>确定重置该用户登录密码？系统将生成随机密码（字母与数字）。</p>
        <div v-if="resetConfirmRow" class="reset-confirm-target">
          <span>{{ resetConfirmRow.name || resetConfirmRow.employee_no }}</span>
          <small>{{ resetConfirmRow.employee_no }}</small>
        </div>
        <p class="reset-confirm-note">请妥善保管后续弹窗中的密码内容。</p>
      </div>
    </NeuroAgentDialog>

    <NeuroAgentDialog
      v-model="resetResultDlg"
      :title="credentialDialogTitle"
      icon="🔑"
      size="medium"
      :show-cancel="false"
      :show-confirm="false"
      @close="onResetResultDlgClose"
    >
      <div v-if="resetResult" class="reset-result-panel nm-form">
        <p class="reset-result-hint">请复制并安全交付给用户；关闭窗口后如需再次查看请重新重置。</p>
        <div class="nm-form-item">
          <label class="nm-form-label">工号</label>
          <div class="reset-result-value">{{ resetResult.employee_no }}</div>
        </div>
        <div class="nm-form-item">
          <label class="nm-form-label">手机号</label>
          <div class="reset-result-value">{{ displayResetPhone() }}</div>
        </div>
        <div class="nm-form-item">
          <label class="nm-form-label">{{ credentialPasswordLabel }}</label>
          <div class="reset-result-value reset-result-value--mono">{{ resetResult.password }}</div>
        </div>
      </div>
      <template #footer-right>
        <button type="button" class="nm-btn nm-btn--ghost" @click="closeResetResultDlg">关闭</button>
        <button type="button" class="nm-btn nm-btn--ghost" @click="copyResetPasswordOnly">复制密码</button>
        <button type="button" class="nm-btn nm-btn--primary" @click="copyResetAllInfo">复制全部</button>
      </template>
    </NeuroAgentDialog>
  </div>
</template>

<style scoped>
.page--list {
  padding: 0;
  max-width: none;
  width: 100%;
}

.user-mgmt-root {
  display: flex;
  flex-direction: column;
  gap: 16px;
  width: 100%;
  min-height: calc(100dvh - 160px);
  min-height: calc(100vh - 160px);
}

.user-mgmt-layout {
  flex: 1;
  display: grid;
  grid-template-columns: minmax(240px, 300px) minmax(0, 1fr);
  gap: 16px;
  min-height: 0;
  width: 100%;
  align-items: stretch;
}

@media (max-width: 991px) {
  .user-mgmt-layout {
    grid-template-columns: 1fr;
  }
}

.user-mgmt-aside {
  min-width: 0;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.user-mgmt-aside-card {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.user-mgmt-aside-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  width: 100%;
}

.user-mgmt-aside-title {
  font-weight: 600;
  font-size: 14px;
  min-width: 0;
}

.user-mgmt-aside-expand-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  flex-shrink: 0;
}

.user-mgmt-aside-expand-btn__icon {
  font-size: 13px;
}

.user-mgmt-aside-card :deep(.el-card__body) {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
  overflow: hidden;
}

.aside-toolbar {
  flex-shrink: 0;
  margin-bottom: 8px;
}

.aside-tree-wrap {
  flex: 1;
  min-height: 0;
  overflow: auto;
}

.user-mgmt-main {
  min-width: 0;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.user-mgmt-main :deep(.neuro-agent-list-page) {
  flex: 1;
  min-height: 0 !important;
  display: flex;
  flex-direction: column;
}

.user-mgmt-main :deep(.list-card) {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
  margin-top: 0;
}

.user-mgmt-main :deep(.card-table) {
  flex: 1;
  min-height: 0;
  overflow: auto;
}

/* Chrome 对表格按钮组的最小宽度计算更激进，操作列必须固定单行横排。 */
.user-mgmt-root :deep(.el-table__cell .cell) {
  flex-wrap: nowrap;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.user-mgmt-root :deep(.user-mgmt-op-btns.op-btns) {
  display: inline-flex;
  flex-wrap: nowrap;
  align-items: center;
  gap: 8px;
  width: max-content;
  max-width: none;
  white-space: nowrap;
}

.user-mgmt-root :deep(.user-mgmt-op-btns.op-btns .el-button) {
  flex: 0 0 auto;
  margin-left: 0;
  min-width: 62px;
  padding-left: 12px;
  padding-right: 12px;
}

.cell-multi-line {
  display: inline-block;
  max-width: 100%;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  font-size: 13px;
  vertical-align: middle;
}

.role-tags-cell {
  display: inline-flex;
  flex-wrap: nowrap;
  gap: 4px;
  align-items: center;
  max-width: 100%;
  overflow: hidden;
  white-space: nowrap;
}

.role-tags-cell .role-tag-chip {
  flex: 0 0 auto;
}

.role-tag-chip {
  margin: 0;
}

.muted {
  color: var(--el-text-color-placeholder);
  font-size: 13px;
}

.pager {
  margin-top: 12px;
  display: flex;
  justify-content: flex-end;
}

.user-status-confirm-panel {
  display: grid;
  gap: 12px;
  color: var(--el-text-color-secondary);
  line-height: 1.7;
}

.user-status-confirm-panel p {
  margin: 0;
}

.user-status-confirm-target {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 12px 14px;
  border: 1px solid rgba(0, 245, 212, 0.18);
  border-radius: 10px;
  background: rgba(0, 245, 212, 0.08);
  color: var(--el-text-color-primary);
}

.user-status-confirm-target small {
  color: var(--el-text-color-secondary);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}

.reset-confirm-panel {
  display: grid;
  gap: 12px;
  color: var(--el-text-color-secondary);
  line-height: 1.7;
}

.reset-confirm-panel p {
  margin: 0;
}

.reset-confirm-target {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 12px 14px;
  border: 1px solid rgba(0, 245, 212, 0.18);
  border-radius: 10px;
  background: rgba(0, 245, 212, 0.08);
  color: var(--el-text-color-primary);
}

.reset-confirm-target small {
  color: var(--el-text-color-secondary);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}

.reset-confirm-note {
  color: var(--el-color-warning);
}

.reset-result-panel {
  padding-top: 4px;
}

.reset-result-hint {
  margin: 0 0 14px;
  font-size: 13px;
  color: var(--el-text-color-secondary);
  line-height: 1.5;
}

.reset-result-value {
  font-size: 15px;
  font-weight: 600;
  color: var(--nm-text-primary, var(--el-text-color-primary));
  word-break: break-all;
}

.reset-result-value--mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  letter-spacing: 0.02em;
}

.position-cascader-search {
  padding: 10px 12px 8px;
  min-width: 420px;
  box-sizing: border-box;
}

.position-cascader-empty {
  padding: 16px 20px;
  color: var(--el-text-color-secondary);
  font-size: 13px;
}
</style>
