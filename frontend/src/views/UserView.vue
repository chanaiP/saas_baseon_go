<script setup lang="ts">
defineOptions({ name: 'UserView' })
import { CaretBottom, CaretTop } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { computed, markRaw, onMounted, ref, shallowRef } from 'vue'

import { fetchDictItemsByCode } from '@/api/dict'
import type { DictItemRow } from '@/api/dict'
import { fetchOrgTree } from '@/api/organization'
import type { OrgNode } from '@/api/organization'
import { fetchSysParamBatch } from '@/api/param'
import { fetchPositions } from '@/api/position'
import type { PositionRow } from '@/api/position'
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

const scopeLabel = ref('未限定组织（全部用户）')

const dlg = ref(false)
const edit = ref<UserRow | null>(null)

const resetResultDlg = ref(false)
const resetResult = ref<{ employee_no: string; phone: string; password: string } | null>(null)
const form = ref({
  employee_no: '',
  password: '',
  name: '',
  phone: '',
  role_ids: [] as number[],
  department_ids: [] as number[],
  position_ids: [] as number[],
})

function flattenDepartmentOptions(nodes: OrgNode[], companyName = ''): { value: number; label: string }[] {
  const out: { value: number; label: string }[] = []
  for (const n of nodes) {
    if (n.node_type === 'company') {
      if (n.children?.length) out.push(...flattenDepartmentOptions(n.children, n.name))
    } else if (n.node_type === 'department') {
      out.push({ value: n.id, label: companyName ? `${companyName} · ${n.name}` : n.name })
      if (n.children?.length) out.push(...flattenDepartmentOptions(n.children, companyName))
    } else if (n.children?.length) {
      out.push(...flattenDepartmentOptions(n.children, companyName))
    }
  }
  return out
}

const departmentOptions = computed(() => flattenDepartmentOptions(org.value))
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

const roleNameById = computed(() => Object.fromEntries(roles.value.map((r) => [r.id, r.name])))

function firstDept(nodes: OrgNode[], companyId: number | null): { companyId?: number; departmentId: number } | null {
  for (const n of nodes) {
    const nextC = n.node_type === 'company' ? n.id : companyId
    if (n.node_type === 'department') return { companyId: nextC ?? undefined, departmentId: n.id }
    if (n.children?.length) {
      const r = firstDept(n.children, nextC)
      if (r) return r
    }
  }
  return null
}

async function loadOrg() {
  try {
    org.value = markRaw(await fetchOrgTree())
  } catch {
    /** 无组织架构菜单权限、套餐未含 org_manage、或接口异常时不应阻断用户列表加载 */
    org.value = []
  }
  const p = firstDept(org.value, null)
  if (p) {
    departmentId.value = p.departmentId
    companyId.value = undefined
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
      key: 'userKeyword',
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
  edit.value = null
  form.value = {
    employee_no: '',
    password: '112233',
    name: '',
    phone: '',
    role_ids: roles.value[0] ? [roles.value[0].id] : [],
    department_ids: departmentId.value ? [departmentId.value] : [],
    position_ids: [],
  }
  dlg.value = true
}

function openEdit(row: UserRow) {
  edit.value = row
  const dIds =
    row.department_ids?.length ? [...row.department_ids] : row.department_id != null ? [row.department_id] : []
  form.value = {
    employee_no: row.employee_no,
    password: '',
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
  try {
    if (edit.value) {
      const body: Parameters<typeof updateUser>[1] = {
        name: form.value.name,
        phone: form.value.phone ? normalizePhoneInput(form.value.phone) : null,
        role_ids: form.value.role_ids,
        company_id: edit.value.company_id,
        department_ids: form.value.department_ids,
        position_ids: form.value.position_ids,
        status: edit.value.status,
      }
      await updateUser(edit.value.id, body)
    } else {
      await createUser({
        employee_no: form.value.employee_no,
        password: form.value.password,
        name: form.value.name,
        phone: form.value.phone ? normalizePhoneInput(form.value.phone) : undefined,
        company_id: companyId.value,
        department_id: departmentId.value,
        department_ids: form.value.department_ids.length ? form.value.department_ids : undefined,
        position_ids: form.value.position_ids.length ? form.value.position_ids : undefined,
        role_ids: form.value.role_ids,
      })
    }
    dlg.value = false
    ElMessage.success('已保存')
    await loadUsers()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '保存用户失败')
  }
}

async function remove(row: UserRow) {
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

async function resetPwd(row: UserRow) {
  try {
    await ElMessageBox.confirm(
      '确定重置该用户登录密码？系统将生成随机密码（字母与数字），请妥善保管弹窗中的内容。',
      '重置密码',
      { type: 'warning' },
    )
    const res = await resetUserPassword(row.id)
    resetResult.value = {
      employee_no: row.employee_no,
      phone: (row.phone ?? '').trim(),
      password: res.new_password,
    }
    resetResultDlg.value = true
    ElMessage.success('密码已重置')
  } catch (e) {
    if (e === 'cancel') return
    ElMessage.error(e instanceof Error ? e.message : '重置密码失败')
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
  appliedKeyword.value = String(payload.filters?.userKeyword ?? '').trim()
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
    .map((id) => departmentOptions.value.find((o) => o.value === id)?.label ?? `#${id}`)
    .join('；')
}

function formatPositionCell(row: UserRow) {
  const ids = row.position_ids ?? []
  if (!ids.length) return '—'
  return ids.map((id) => positionNameById.value[id] ?? `#${id}`).join('；')
}

function formatRolesTitle(row: UserRow) {
  const ids = row.role_ids ?? []
  if (!ids.length) return ''
  return ids.map((id) => roleNameById.value[id] ?? `#${id}`).join('；')
}

const columns: TableColumn[] = [
  { key: 'employee_no', title: '工号', width: 110 },
  { key: 'name', title: '姓名', width: 100 },
  { key: 'departments', title: '部门', minWidth: 168 },
  { key: 'positions', title: '岗位', minWidth: 140 },
  { key: 'roles', title: '角色', minWidth: 160 },
  { key: 'phone', title: '手机', width: 130 },
  { key: 'status', title: '状态', width: 80 },
  { key: 'actions', title: '操作', minWidth: 300, width: 320, fixed: 'right', tooltip: false },
]

onMounted(async () => {
  await loadOrg()
  await Promise.all([loadRoles(), loadPositions()])
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
              <span v-else class="muted">—</span>
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
              <el-button v-permission="'user:edit'" size="small" @click="openEdit(row)">编辑</el-button>
              <el-button v-permission="'user:reset_password'" size="small" @click="resetPwd(row)">重置密码</el-button>
              <el-button v-permission="'user:delete'" size="small" type="danger" @click="remove(row)">删除</el-button>
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
        <div class="nm-form-item" v-if="!edit">
          <label class="nm-form-label">初始密码</label>
          <el-input v-model="form.password" type="password" />
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
          <label class="nm-form-label">任职部门</label>
          <el-select
            v-model="form.department_ids"
            multiple
            filterable
            collapse-tags
            collapse-tags-tooltip
            placeholder="多项时任职部门列表的第一项为主部门"
            style="width: 100%"
          >
            <el-option v-for="o in departmentOptions" :key="o.value" :label="o.label" :value="o.value" />
          </el-select>
        </div>
        <div class="nm-form-item">
          <label class="nm-form-label">岗位</label>
          <el-select
            v-model="form.position_ids"
            multiple
            filterable
            collapse-tags
            collapse-tags-tooltip
            placeholder="可选择多个岗位"
            style="width: 100%"
          >
            <el-option v-for="p in positions" :key="p.id" :label="p.name" :value="p.id" />
          </el-select>
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
      v-model="resetResultDlg"
      title="密码已重置"
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
          <label class="nm-form-label">重置后的密码</label>
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

/* 命令布局默认禁止末列换行，用户管理操作列按钮多，需放宽以免裁切 */
.user-mgmt-root :deep(.neuro-command-layout .el-table__fixed-right .el-table__cell .cell) {
  flex-wrap: wrap;
  white-space: normal;
}

.user-mgmt-root :deep(.neuro-command-layout .user-mgmt-op-btns.op-btns) {
  flex-wrap: wrap;
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
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  align-items: center;
  max-width: 100%;
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
</style>
