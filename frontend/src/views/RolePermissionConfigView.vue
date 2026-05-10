<script setup lang="ts">
defineOptions({ name: 'RolePermissionConfigView' })
import { ArrowLeft } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { computed, markRaw, reactive, ref, shallowRef, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { fetchOrgTree } from '@/api/organization'
import type { OrgNode } from '@/api/organization'
import { fetchBusinessUnitPage } from '@/api/businessUnit'
import type { BusinessUnitRow } from '@/api/businessUnit'
import { fetchMenuBundles } from '@/api/permission'
import type { MenuBundle } from '@/api/permission'
import { fetchRole, updateRolePermissions } from '@/api/role'
import type { RoleDataOverride, RoleRow } from '@/api/role'
import { fetchUsers } from '@/api/user'
import type { UserRow } from '@/api/user'
import OrgUserPicker from '@/components/OrgUserPicker.vue'
import NeuroAgentDialog from '@/views/components/NeuroAgentDialog.vue'

const DATA_SCOPE_OPTIONS = [
  { value: 'ALL', label: '全部' },
  { value: 'ORG', label: '本公司/部门' },
  { value: 'ORG_SUB', label: '本公司/部门及子部门' },
  { value: 'SELF', label: '仅看自己' },
  { value: 'CUSTOM', label: '自定义' },
] as const
const BU_ACCESS_MODE_OPTIONS = [
  { value: 'CURRENT_ORG_BU', label: '当前组织关联业务单元' },
  { value: 'SPECIFIED_BU', label: '指定业务单元' },
] as const
const DATA_PERM_MODE_ORG = new Set(['ORG', 'ORG_BU'])
const DATA_PERM_MODE_BU = new Set(['BU', 'ORG_BU'])

const ALLOWED_SCOPE = new Set<string>(DATA_SCOPE_OPTIONS.map((o) => o.value))

const route = useRoute()
const router = useRouter()
const roleId = computed(() => Number(route.params.roleId))

const loading = ref(true)
const saving = ref(false)
const role = ref<RoleRow | null>(null)
const bundles = ref<MenuBundle[]>([])
/** 「应用到所有菜单」时的模板行：随任意下拉变更更新 */
const scopeTemplatePath = ref<string | null>(null)
const orgTree = shallowRef<OrgNode[]>([])
const userOptions = ref<UserRow[]>([])
const businessUnitOptions = ref<BusinessUnitRow[]>([])
const confirmDlg = ref(false)
const confirmTitle = ref('操作确认')
const confirmMessage = ref('')
let confirmResolver: ((ok: boolean) => void) | null = null

function openConfirmDialog(title: string, message: string): Promise<boolean> {
  confirmTitle.value = title
  confirmMessage.value = message
  confirmDlg.value = true
  return new Promise<boolean>((resolve) => {
    confirmResolver = resolve
  })
}

function resolveConfirmDialog(ok: boolean) {
  confirmDlg.value = false
  const resolver = confirmResolver
  confirmResolver = null
  if (resolver) resolver(ok)
}

interface MenuRowState {
  menuOn: boolean
  opIds: number[]
  data_scope: string
  customCompanyIds: number[]
  customDeptIds: number[]
  customUserIds: number[]
  customBusinessUnitIds: number[]
  buDataAccessMode: 'CURRENT_ORG_BU' | 'SPECIFIED_BU' | ''
}

const menuState = reactive<Record<string, MenuRowState>>({})

/** 不在当前 menu bundles 中的权限 id（套餐裁剪或自定义权限），保存时必须合并进 payload，否则会误删库里的绑定 */
const preservedPermissionIds = ref<number[]>([])

function idsCoveredByBundles(bundleList: MenuBundle[]): Set<number> {
  const s = new Set<number>()
  for (const b of bundleList) {
    if (b.menu_permission_id > 0) s.add(b.menu_permission_id)
    if (b.data_permission_id > 0) s.add(b.data_permission_id)
    for (const o of b.operations) s.add(o.id)
  }
  return s
}

function syncPreservedPermissionIds(r: RoleRow) {
  const covered = idsCoveredByBundles(bundles.value)
  preservedPermissionIds.value = r.permission_ids.filter((id) => !covered.has(id))
}

function normalizeDataScope(s: string | undefined) {
  if (s && ALLOWED_SCOPE.has(s)) return s
  return ''
}

function ensureStatePath(path: string) {
  if (!menuState[path]) {
    menuState[path] = {
      menuOn: false,
      opIds: [],
      data_scope: '',
      customCompanyIds: [],
      customDeptIds: [],
      customUserIds: [],
      customBusinessUnitIds: [],
      buDataAccessMode: '',
    }
  }
}

function resetMenuState() {
  for (const k of Object.keys(menuState)) delete menuState[k]
}

function initMenuStateFromRole(r: RoleRow) {
  resetMenuState()
  for (const b of bundles.value) {
    const menuOn = r.permission_ids.includes(b.menu_permission_id)
    const opIds = b.operations.filter((o) => r.permission_ids.includes(o.id)).map((o) => o.id)
    const ov = r.data_overrides?.find((d) => d.permission_id === b.data_permission_id)
    menuState[b.path] = {
      menuOn,
      opIds: [...opIds],
      data_scope: normalizeDataScope(ov?.data_scope),
      customCompanyIds: [...(ov?.custom_company_ids ?? [])],
      customDeptIds: [...(ov?.custom_department_ids ?? [])],
      customUserIds: [...(ov?.custom_user_ids ?? [])],
      customBusinessUnitIds: [...(ov?.custom_business_unit_ids ?? [])],
      buDataAccessMode: (ov?.bu_data_access_mode as 'CURRENT_ORG_BU' | 'SPECIFIED_BU' | null) || '',
    }
  }
  syncPreservedPermissionIds(r)
}

function onMenuCheck(b: MenuBundle, checked: boolean) {
  ensureStatePath(b.path)
  const s = menuState[b.path]
  s.menuOn = checked
  if (checked) {
    s.opIds = b.operations.map((o) => o.id)
    if (!supportsOrgScope(b.path) || !ALLOWED_SCOPE.has(s.data_scope)) s.data_scope = ''
  } else {
    s.opIds = []
  }
}

const allMenuChecked = computed(
  () => bundles.value.length > 0 && bundles.value.every((b) => menuState[b.path]?.menuOn),
)
const someMenuChecked = computed(
  () => !allMenuChecked.value && bundles.value.some((b) => menuState[b.path]?.menuOn),
)

function resolveApplySourcePath(): string | null {
  if (scopeTemplatePath.value && menuState[scopeTemplatePath.value]) return scopeTemplatePath.value
  return bundles.value.find((b) => menuState[b.path]?.menuOn)?.path || null
}

const applySourceMenuLabel = computed(() => {
  const path = resolveApplySourcePath()
  if (!path) return '当前'
  return SCOPE_MENU_LABEL(path)
})

async function toggleAllMenus(checked: boolean) {
  const msg = checked ? '确定全选所有菜单权限？' : '确定取消全选？所有菜单权限将被关闭。'
  if (!(await openConfirmDialog('操作确认', msg))) return
  for (const b of bundles.value) onMenuCheck(b, checked)
}

const SCOPE_LABEL_MAP = Object.fromEntries(DATA_SCOPE_OPTIONS.map((o) => [o.value, o.label]))

async function applyToAllMenus() {
  const src = resolveApplySourcePath()
  if (!src || !menuState[src]) {
    ElMessage.warning('请先开启至少一个菜单')
    return
  }
  const cur = menuState[src]
  const srcSupportsOrg = supportsOrgScope(src)
  const srcSupportsBu = supportsBuScope(src)
  const shouldApplyOrg = srcSupportsOrg && !!cur.data_scope
  const shouldApplyBu = srcSupportsBu && !!cur.buDataAccessMode
  if (!shouldApplyOrg && !shouldApplyBu) {
    ElMessage.warning('当前菜单未选择组织架构权限或业务单元权限，无法批量应用')
    return
  }
  const scopeLabel = SCOPE_LABEL_MAP[cur.data_scope] ?? cur.data_scope
  const applyParts: string[] = []
  if (shouldApplyOrg) applyParts.push(`组织架构权限「${scopeLabel}」${cur.data_scope === 'CUSTOM' ? '（含自定义选择）' : ''}`)
  if (shouldApplyBu) {
    const buLabel = BU_ACCESS_MODE_OPTIONS.find((o) => o.value === cur.buDataAccessMode)?.label || cur.buDataAccessMode
    applyParts.push(`业务单元权限「${buLabel}」`)
  }
  if (
    !(await openConfirmDialog(
      '批量应用确认',
      `将「${SCOPE_MENU_LABEL(src)}」的${applyParts.join('，')}应用到所有已开启的菜单，确定？`,
    ))
  ) return
  for (const b of bundles.value) {
    const s = menuState[b.path]
    if (!s?.menuOn || b.path === src) continue
    if (shouldApplyOrg && supportsOrgScope(b.path)) {
      s.data_scope = cur.data_scope
      s.customCompanyIds = [...cur.customCompanyIds]
      s.customDeptIds = [...cur.customDeptIds]
      s.customUserIds = [...cur.customUserIds]
    }
    if (shouldApplyBu && supportsBuScope(b.path)) {
      s.buDataAccessMode = cur.buDataAccessMode
      s.customBusinessUnitIds =
        cur.buDataAccessMode === 'SPECIFIED_BU'
          ? [...cur.customBusinessUnitIds]
          : []
    }
  }
  ElMessage.success('已应用到所有已开启菜单')
}

function SCOPE_MENU_LABEL(path: string): string {
  return bundles.value.find((b) => b.path === path)?.title ?? path
}

function menuDataPermMode(path: string): 'NONE' | 'ORG' | 'BU' | 'ORG_BU' {
  const rawMode = bundles.value.find((b) => b.path === path)?.data_perm_mode
  const mode = String(rawMode || '').trim().toUpperCase().replace('-', '_')
  if (mode === 'NONE' || mode === 'ORG' || mode === 'BU' || mode === 'ORG_BU') return mode
  return 'ORG'
}

function supportsOrgScope(path: string): boolean {
  return DATA_PERM_MODE_ORG.has(menuDataPermMode(path))
}

function supportsBuScope(path: string): boolean {
  return DATA_PERM_MODE_BU.has(menuDataPermMode(path))
}

function buOptionLabel(item: BusinessUnitRow): string {
  return item.status === 1 ? item.name : `${item.name}（已停用）`
}

const PERMISSION_PREFIX_LABELS: Record<string, string> = {
  tenant: '主体',
  plan: '套餐',
  user: '用户',
  org: '组织',
  pos: '岗位',
  role: '角色',
  perm: '权限',
  menu: '菜单',
  dict: '字典',
  dict_type: '字典类型',
  dict_item: '字典项',
  param: '参数',
  business_unit: '业务单元',
  file: '文件',
  brand: '品牌',
}

const PERMISSION_ACTION_LABELS: Record<string, string> = {
  create: '新增',
  edit: '编辑',
  delete: '删除',
  status: '启停',
  reset_password: '重置密码',
  reset_primary_password: '重置主管理员密码',
  quota_config: '调整配额',
  permission: '权限设置',
  package_feature: '套餐中心收录',
  import: '导入',
  export: '导出',
  upload: '上传',
  download: '下载',
}

function splitPermissionCode(code: string): { prefix: string; action: string } | null {
  const raw = code.trim()
  if (!raw) return null
  const colon = raw.indexOf(':')
  if (colon > 0) {
    return { prefix: raw.slice(0, colon), action: raw.slice(colon + 1) }
  }
  for (const prefix of ['business_unit', 'dict_type', 'dict_item']) {
    if (raw.startsWith(`${prefix}_`)) {
      return { prefix, action: raw.slice(prefix.length + 1) }
    }
  }
  const underscore = raw.indexOf('_')
  if (underscore > 0) {
    return { prefix: raw.slice(0, underscore), action: raw.slice(underscore + 1) }
  }
  return null
}

function operationDisplayName(op: { name?: string; path?: string }): string {
  const name = String(op.name || '').trim()
  const path = String(op.path || '').trim()
  const shouldTranslate = !name || name === path || /^[a-z][a-z0-9_]*[:_][a-z0-9_]+$/i.test(name)
  if (!shouldTranslate) return name
  const parsed = splitPermissionCode(path || name)
  if (!parsed) return name || path
  const action = PERMISSION_ACTION_LABELS[parsed.action]
  if (!action) return name || path
  const prefix = PERMISSION_PREFIX_LABELS[parsed.prefix]
  return prefix ? `${prefix}-${action}` : action
}

function onDataScopeChange(path: string, newScope: string) {
  ensureStatePath(path)
  const s = menuState[path]
  s.data_scope = newScope
  scopeTemplatePath.value = path
  if (newScope !== 'CUSTOM') {
    s.customCompanyIds = []
    s.customDeptIds = []
    s.customUserIds = []
  }
}

function onBuAccessModeChange(path: string, value: 'CURRENT_ORG_BU' | 'SPECIFIED_BU' | '') {
  ensureStatePath(path)
  const s = menuState[path]
  s.buDataAccessMode = value
  if (value !== 'SPECIFIED_BU') s.customBusinessUnitIds = []
}

function readCustomKeys(path: string): string[] {
  ensureStatePath(path)
  const s = menuState[path]
  const keys: string[] = []
  for (const id of s.customCompanyIds) {
    keys.push(`c_${id}`)
    for (const u of userOptions.value) if (u.company_id === id && u.department_id == null) keys.push(`u_${u.id}`)
  }
  for (const id of s.customDeptIds) {
    keys.push(`d_${id}`)
    for (const u of userOptions.value) if (u.department_id === id) keys.push(`u_${u.id}`)
  }
  for (const id of s.customUserIds) keys.push(`u_${id}`)
  return [...new Set(keys)]
}

function writeCustomKeys(path: string, keys: string[]) {
  ensureStatePath(path)
  const s = menuState[path]
  s.customCompanyIds = []
  s.customDeptIds = []
  s.customUserIds = []
  for (const k of keys) {
    const id = Number(k.slice(2))
    if (k.startsWith('c_')) s.customCompanyIds.push(id)
    else if (k.startsWith('d_')) s.customDeptIds.push(id)
    else if (k.startsWith('u_')) s.customUserIds.push(id)
  }
}

function buildPayload(): { permission_ids: number[]; data_overrides: RoleDataOverride[] } {
  const ids = new Set<number>()
  const overrides: RoleDataOverride[] = []
  for (const b of bundles.value) {
    const s = menuState[b.path]
    if (!s?.menuOn) continue
    if (b.menu_permission_id > 0) ids.add(b.menu_permission_id)
    if (b.data_permission_id > 0) ids.add(b.data_permission_id)
    for (const oid of s.opIds) ids.add(oid)
    const orgEnabled = supportsOrgScope(b.path)
    const buEnabled = supportsBuScope(b.path)
    const hasOrgConfigured = orgEnabled && !!s.data_scope
    const hasBuConfigured = buEnabled && !!s.buDataAccessMode
    if (!hasOrgConfigured && !hasBuConfigured) continue
    const dataScope = orgEnabled && s.data_scope ? s.data_scope : 'ALL'
    const isCustom = orgEnabled && dataScope === 'CUSTOM'
    overrides.push({
      permission_id: b.data_permission_id,
      data_scope: dataScope,
      custom_company_ids: isCustom ? [...s.customCompanyIds] : [],
      custom_department_ids: isCustom ? [...s.customDeptIds] : [],
      custom_user_ids: isCustom ? [...s.customUserIds] : [],
      custom_business_unit_ids: buEnabled && s.buDataAccessMode === 'SPECIFIED_BU' ? [...s.customBusinessUnitIds] : [],
      bu_data_access_mode: buEnabled && s.buDataAccessMode ? s.buDataAccessMode : null,
    })
  }
  for (const pid of preservedPermissionIds.value) ids.add(pid)
  return { permission_ids: [...ids], data_overrides: overrides }
}

async function load() {
  loading.value = true
  try {
    const id = roleId.value
    if (Number.isNaN(id) || id < 1) {
      router.replace('/roles')
      return
    }
    try {
      const [r, b] = await Promise.all([fetchRole(id), fetchMenuBundles()])
      role.value = r
      bundles.value = b
      initMenuStateFromRole(r)
      scopeTemplatePath.value = b[0]?.path ?? null
    } catch (e) {
      ElMessage.error(e instanceof Error ? e.message : '加载失败')
      role.value = null
      bundles.value = []
      resetMenuState()
      scopeTemplatePath.value = null
      preservedPermissionIds.value = []
      orgTree.value = []
      userOptions.value = []
      businessUnitOptions.value = []
      return
    }
    try {
      const [org, usersPage, buPage] = await Promise.all([
        fetchOrgTree(),
        fetchUsers({ skip: 0, limit: 500, status: 1 }),
        fetchBusinessUnitPage({ skip: 0, limit: 500 }),
      ])
      orgTree.value = markRaw(org)
      userOptions.value = usersPage.items
      businessUnitOptions.value = buPage.items
    } catch {
      orgTree.value = []
      userOptions.value = []
      businessUnitOptions.value = []
      ElMessage.warning('组织架构、用户或业务单元加载失败，部分权限选项可能不可用')
    }
  } finally {
    loading.value = false
  }
}

async function save() {
  if (!role.value) return
  for (const b of bundles.value) {
    const s = menuState[b.path]
    if (!s?.menuOn) continue
    const orgEnabled = supportsOrgScope(b.path)
    const buEnabled = supportsBuScope(b.path)
    const hasOrgConfigured = orgEnabled && !!s.data_scope
    const hasBuConfigured = buEnabled && !!s.buDataAccessMode
    if ((orgEnabled || buEnabled) && !hasOrgConfigured && !hasBuConfigured) {
      ElMessage.warning(`「${b.title}」请至少配置组织架构权限或业务单元权限`)
      return
    }
    if (supportsOrgScope(b.path) && s.data_scope === 'CUSTOM' && !s.customCompanyIds.length && !s.customDeptIds.length && !s.customUserIds.length) {
      ElMessage.warning(`「${b.title}」为自定义范围时，请至少选择组织架构或用户`)
      return
    }
    if (supportsBuScope(b.path) && s.buDataAccessMode === 'SPECIFIED_BU' && !s.customBusinessUnitIds.length) {
      ElMessage.warning(`「${b.title}」选择指定业务单元时，请至少选择一个业务单元`)
      return
    }
  }
  saving.value = true
  try {
    const { permission_ids, data_overrides } = buildPayload()
    await updateRolePermissions(role.value.id, {
      permission_ids,
      data_overrides,
    })
    ElMessage.success('已保存')
    await load()
  } finally {
    saving.value = false
  }
}

function goBack() {
  void router.push('/roles')
}

watch(roleId, (rid) => {
  if (Number.isNaN(rid) || rid < 1) {
    void router.replace('/roles')
    return
  }
  void load()
}, { immediate: true })
</script>

<template>
  <div class="page" v-loading="loading">
    <template v-if="role">
      <div class="toolbar">
        <div class="toolbar-mid">
          <h1 class="title">{{ role.name }}（{{ role.code }}）<span class="title-tag">权限配置</span></h1>
        </div>
        <el-button text type="primary" class="back" @click="goBack">
          <el-icon><ArrowLeft /></el-icon>
          返回角色权限
        </el-button>
        <el-button type="primary" :loading="saving" @click="save">保存</el-button>
      </div>

      <el-card shadow="never" class="card">
        <template #header>
          <span>菜单 / 操作 / 权限范围</span>
          <span class="card-hint">按菜单配置显示组织架构权限、业务单元权限；支持单选或多选业务单元</span>
          <div class="card-header-actions">
            <el-button size="small" type="primary" link @click="applyToAllMenus">
              把 <span class="apply-menu-name">{{ applySourceMenuLabel }}</span> 菜单数据权限 应用到所有已开启菜单
            </el-button>
            <el-checkbox
              class="check-all"
              :model-value="allMenuChecked"
              :indeterminate="someMenuChecked"
              @change="(v: boolean | string | number) => toggleAllMenus(!!v)"
            >全选菜单</el-checkbox>
          </div>
        </template>
        <div class="bundle-list">
          <div
            v-for="b in bundles"
            :key="b.path"
            class="bundle-card"
          >
            <div class="bundle-card-hd">
              <el-checkbox
                :model-value="menuState[b.path]?.menuOn ?? false"
                @update:model-value="(v: boolean | string | number) => onMenuCheck(b, !!v)"
              >
                <span class="menu-label">{{ b.title }}</span>
              </el-checkbox>
              <div v-if="supportsOrgScope(b.path) || supportsBuScope(b.path)" class="scope-select-cluster">
                <div
                  v-if="supportsOrgScope(b.path)"
                  class="scope-select-wrap"
                  :class="{ 'scope-select-wrap--off': !menuState[b.path]?.menuOn }"
                >
                  <span class="scope-select-prefix">组织架构权限：</span>
                  <el-select
                    class="scope-select"
                  :model-value="menuState[b.path]?.data_scope ?? ''"
                    :disabled="!menuState[b.path]?.menuOn"
                  placeholder="请选择"
                    filterable
                  clearable
                    aria-label="组织架构权限"
                    @update:model-value="(v: string) => onDataScopeChange(b.path, v)"
                    @click.stop
                  >
                  <el-option label="请选择" value="" />
                    <el-option
                      v-for="opt in DATA_SCOPE_OPTIONS"
                      :key="opt.value"
                      :label="opt.label"
                      :value="opt.value"
                    />
                  </el-select>
                </div>
                <div
                  v-if="supportsBuScope(b.path)"
                  class="scope-select-wrap"
                  :class="{ 'scope-select-wrap--off': !menuState[b.path]?.menuOn }"
                >
                  <span class="scope-select-prefix">业务单元权限：</span>
                  <el-select
                    class="scope-select"
                    :model-value="menuState[b.path]?.buDataAccessMode || ''"
                    :disabled="!menuState[b.path]?.menuOn"
                  placeholder="请选择"
                  clearable
                    aria-label="业务单元权限"
                    @update:model-value="(v: 'CURRENT_ORG_BU' | 'SPECIFIED_BU' | '') => onBuAccessModeChange(b.path, v)"
                    @click.stop
                  >
                    <el-option label="不配置" value="" />
                    <el-option v-for="opt in BU_ACCESS_MODE_OPTIONS" :key="opt.value" :label="opt.label" :value="opt.value" />
                  </el-select>
                </div>
              </div>
            </div>
            <!-- 操作权限 -->
            <div v-show="menuState[b.path]?.menuOn && b.operations.length" class="bundle-ops">
              <el-checkbox-group v-model="menuState[b.path]!.opIds" class="op-group" @click.stop>
                <el-checkbox v-for="op in b.operations" :key="op.id" :label="op.id">
                  <span class="op-name" :title="op.path">{{ operationDisplayName(op) }}</span>
                </el-checkbox>
              </el-checkbox-group>
            </div>
            <!-- 组织架构自定义权限 -->
            <div
              v-if="menuState[b.path]?.menuOn && supportsOrgScope(b.path) && menuState[b.path]?.data_scope === 'CUSTOM'"
              class="bundle-custom"
              @click.stop
            >
              <div class="custom-label">组织架构权限（自定义）</div>
              <OrgUserPicker
                :bundle-key="`${b.path}-custom`"
                :model-value="readCustomKeys(b.path)"
                :org-tree="orgTree"
                :users="userOptions"
                @update:model-value="(keys: string[]) => writeCustomKeys(b.path, keys)"
              />
              <div v-if="!orgTree.length" class="empty-hint">
                暂无组织架构数据，请先在「组织架构」页面添加公司与部门。
              </div>
            </div>
            <div
              v-if="menuState[b.path]?.menuOn && supportsBuScope(b.path) && menuState[b.path]?.buDataAccessMode === 'SPECIFIED_BU'"
              class="bundle-custom"
              @click.stop
            >
              <div class="custom-label">业务单元权限（指定）</div>
              <el-select
                v-model="menuState[b.path]!.customBusinessUnitIds"
                multiple
                collapse-tags
                collapse-tags-tooltip
                filterable
                style="width: 100%"
                placeholder="选择业务单元（支持多选）"
              >
                <el-option
                  v-for="bu in businessUnitOptions"
                  :key="bu.id"
                  :label="buOptionLabel(bu)"
                  :value="bu.id"
                />
              </el-select>
              <div v-if="!businessUnitOptions.length" class="empty-hint">暂无业务单元数据</div>
            </div>
          </div>
        </div>
      </el-card>
    </template>
    <NeuroAgentDialog
      v-model="confirmDlg"
      :title="confirmTitle"
      icon="⚠️"
      size="small"
      confirm-text="确定"
      cancel-text="取消"
      @confirm="resolveConfirmDialog(true)"
      @cancel="resolveConfirmDialog(false)"
      @close="resolveConfirmDialog(false)"
    >
      <p class="confirm-msg">{{ confirmMessage }}</p>
    </NeuroAgentDialog>
  </div>
</template>

<style scoped>
.page {
  padding: 16px;
  min-height: 360px;
}
.toolbar {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}
.back {
  flex-shrink: 0;
}
.toolbar-mid {
  flex: 1;
  min-width: 0;
}
.title {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  display: inline-flex;
  align-items: baseline;
  gap: 10px;
}
.title-tag {
  font-size: 13px;
  font-weight: 500;
  color: var(--el-text-color-secondary);
}
:deep(.el-card__header) {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
}
.card-hint {
  margin-left: 8px;
  font-size: 12px;
  font-weight: normal;
  color: var(--el-text-color-secondary);
  flex: 1;
  min-width: 200px;
}
.card-header-actions {
  margin-left: auto;
  display: flex;
  align-items: center;
  gap: 12px;
  flex-shrink: 0;
}
.check-all {
  flex-shrink: 0;
}

.bundle-card {
  margin-bottom: 12px;
  padding: 12px 14px;
  border-radius: 8px;
  border: 1px solid var(--el-border-color-lighter);
  transition: border-color 0.15s, background 0.15s;
}
.bundle-card:hover {
  border-color: var(--el-color-primary-light-5);
}
.bundle-card-hd {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}
.menu-label {
  font-weight: 600;
  font-size: 14px;
}
/* 与 el-select 一体的单框：前缀在框内左侧，右侧为无边框的下拉 */
.scope-select-wrap {
  display: inline-flex;
  align-items: stretch;
  flex-shrink: 0;
  max-width: 100%;
  width: 260px;
  min-width: 260px;
  border-radius: var(--el-border-radius-base);
  background-color: var(--el-fill-color-blank);
  box-shadow: 0 0 0 1px var(--el-border-color) inset;
  transition: box-shadow 0.2s cubic-bezier(0.645, 0.045, 0.355, 1);
  overflow: hidden;
}
.scope-select-cluster {
  margin-left: auto;
  display: inline-flex;
  align-items: center;
  gap: 10px;
  flex-wrap: nowrap;
  justify-content: flex-end;
  min-width: 0;
}
.scope-select-wrap:hover:not(.scope-select-wrap--off):not(:focus-within) {
  box-shadow: 0 0 0 1px var(--el-border-color-hover) inset;
}
.scope-select-wrap:focus-within:not(.scope-select-wrap--off) {
  box-shadow: 0 0 0 1px var(--el-color-primary) inset;
}
.scope-select-wrap--off {
  cursor: not-allowed;
  background-color: var(--el-fill-color-light);
  box-shadow: 0 0 0 1px var(--el-border-color-lighter) inset;
}
.scope-select-prefix {
  display: inline-flex;
  align-items: center;
  padding: 0 10px 0 11px;
  font-size: var(--el-font-size-base);
  color: var(--el-text-color-regular);
  white-space: nowrap;
  flex-shrink: 0;
  border-right: 1px solid var(--el-border-color-lighter);
  background-color: transparent;
  user-select: none;
}
.scope-select-wrap--off .scope-select-prefix {
  color: var(--el-text-color-placeholder);
}
.scope-select {
  flex: 1;
  min-width: 0;
  width: auto !important;
}
.scope-select-wrap :deep(.scope-select.el-select) {
  width: 100% !important;
}
.scope-select-wrap :deep(.scope-select .el-select__wrapper),
.scope-select-wrap :deep(.scope-select .el-select__wrapper:hover),
.scope-select-wrap :deep(.scope-select .el-select__wrapper.is-hovering),
.scope-select-wrap :deep(.scope-select .el-select__wrapper.is-focused) {
  box-shadow: none !important;
  border-radius: 0 !important;
  background-color: transparent !important;
  min-height: var(--el-component-size);
}
.scope-select-wrap--off :deep(.scope-select .el-select__wrapper) {
  background-color: transparent !important;
}

.bundle-ops {
  margin: 8px 0 0 24px;
  padding-top: 6px;
  border-top: 1px dashed var(--el-border-color-lighter);
}
.op-group {
  display: flex;
  flex-wrap: wrap;
  gap: 4px 14px;
}
.op-name {
  font-size: 13px;
}

.bundle-custom {
  margin: 12px 0 0 24px;
  padding: 12px;
  border-radius: 8px;
  background: var(--el-fill-color-light);
  border: 1px solid var(--el-border-color-lighter);
}
.custom-label {
  font-size: 13px;
  font-weight: 600;
  margin-bottom: 8px;
  color: var(--el-text-color-regular);
}
.empty-hint {
  margin-top: 8px;
  font-size: 12px;
  color: var(--el-color-warning);
}
.apply-menu-name {
  color: var(--el-color-primary);
  font-weight: 700;
  text-shadow: 0 0 10px color-mix(in srgb, var(--el-color-primary) 35%, transparent);
}
.confirm-msg {
  margin: 0;
  font-size: 15px;
  line-height: 1.7;
  color: var(--el-text-color-primary);
}
</style>
