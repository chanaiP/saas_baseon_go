<script setup lang="ts">
defineOptions({ name: 'TenantView' })
import { ElMessage, ElMessageBox } from 'element-plus'
import { computed, onMounted, ref } from 'vue'

import {
  fetchFeatures,
  fetchPlanFeatures,
  fetchPlanQuotas,
  fetchPlans,
  fetchTenantQuotaOverrides,
  fetchTenantSubscription,
  saveTenantQuotaOverrides,
} from '@/api/plan'
import type { Feature, Plan, PlanQuotaValue, Quota } from '@/api/plan'
import {
  createTenantWithPackage,
  deleteTenant,
  fetchTenant,
  fetchTenantQuotaRecords,
  fetchTenantPrimaryAdmin,
  fetchTenants,
  patchTenantStatus,
  resetTenantPrimaryAdminPassword,
  updateTenant,
} from '@/api/tenant'
import { archiveSuccessMessage, confirmArchiveAction } from '@/composables/useArchiveConfirm'
import type { TenantPrimaryAdminPasswordResetResult } from '@/api/tenant'
import type { Tenant, TenantBusinessUnitQuotaRecord, TenantCreatePayload, TenantOrgQuotaRecord } from '@/api/tenant'
import { usePermissionStore } from '@/stores/permission'
import { useTenantBrandingStore } from '@/stores/tenantBranding'
import { isValidOptionalPhone, normalizePhoneInput, sanitizePhoneInput } from '@/utils/phone'
import type { TableColumn } from '@/views/components/NeuroAgentListPage.vue'
import NeuroAgentDialog from '@/views/components/NeuroAgentDialog.vue'
import NeuroAgentListPage from '@/views/components/NeuroAgentListPage.vue'
import NeuroAgentPageShell from '@/views/components/NeuroAgentPageShell.vue'
import CardListView, { type CardListItem, type FilterOption } from '@/views/components/CardListView.vue'

function usageText(t: Tenant | undefined): string {
  if (!t?.created_at) return ''
  const diffMs = Date.now() - new Date(t.created_at).getTime()
  /** 已满 24 小时的整天数；创建后 24 小时内为 0 */
  const days = Math.max(0, Math.floor(diffMs / 86400000))
  const years = Math.floor(days / 365)
  const months = Math.floor((days % 365) / 30)
  const remainDays = days % 30

  const parts: string[] = []
  if (years > 0) parts.push(`${years}年`)
  if (months > 0) parts.push(`${months}月`)
  if (remainDays > 0) parts.push(`${remainDays}天`)
  if (parts.length === 0) return '已使用 0天'
  return '已使用 ' + parts.join(' ')
}

function formatDate(dateStr: string | Date): string {
  const d = new Date(dateStr)
  return `${d.getFullYear()}/${String(d.getMonth() + 1).padStart(2, '0')}/${String(d.getDate()).padStart(2, '0')}`
}

const perm = usePermissionStore()
const tenantBrand = useTenantBrandingStore()

const loading = ref(false)
const items = ref<Tenant[]>([])
const total = ref(0)
const page = ref(1)
const limit = ref(10)
const sizeDropdownOpen = ref(false)
/** 点击卡片后展示主体详情（含下属公司等） */
const detailDrawerVisible = ref(false)

const searchKeyword = ref('')
const statusFilter = ref('')
const shouldShowPager = computed(() => total.value > limit.value)

const tenantFilterOptions: FilterOption[] = [
  { label: '全部', value: '' },
  { label: '启用', value: 'active' },
  { label: '停用', value: 'inactive' },
]

/** 按套餐编码映射胶囊配色；未知编码用稳定哈希落入调色盘。 */
function planPillToneClass(planCode?: string | null, planName?: string | null): string {
  const raw = (planCode && String(planCode).trim()) || (planName && String(planName).trim()) || ''
  const c = raw.toUpperCase()
  const byCode: Record<string, string> = {
    TRIAL: 'tc-plan-pill--trial',
    BASIC: 'tc-plan-pill--basic',
    PRO: 'tc-plan-pill--pro',
    ENTERPRISE: 'tc-plan-pill--enterprise',
  }
  if (byCode[c]) return byCode[c]
  let h = 0
  for (let i = 0; i < c.length; i++) h = (h * 31 + c.charCodeAt(i)) >>> 0
  return `tc-plan-pill--tone-${h % 8}`
}

function filteredTenants(): Tenant[] {
  let result = [...items.value]
  if (searchKeyword.value.trim()) {
    const q = searchKeyword.value.trim().toLowerCase()
    result = result.filter((t) =>
      t.name.toLowerCase().includes(q) || t.code.toLowerCase().includes(q)
    )
  }
  if (statusFilter.value) {
    if (statusFilter.value === 'active') result = result.filter((t) => t.status === 1)
    if (statusFilter.value === 'inactive') result = result.filter((t) => t.status === 0)
  }
  return result
}

function quotaProgressPct(limit: number | null | undefined, used: number | null | undefined): number {
  const cap = Number(limit ?? 0)
  const usedValue = Math.max(0, Number(used ?? 0))
  if (cap <= 0 || cap === -1) return 0
  return Math.min(100, Math.round((usedValue / cap) * 100))
}

function quotaUsageText(limit: number | null | undefined, used: number | null | undefined): string {
  const usedValue = Math.max(0, Number(used ?? 0))
  const cap = Number(limit ?? 0)
  if (cap === -1) return `${usedValue}/∞（不限）`
  return `${usedValue}/${cap > 0 ? cap : 0}`
}

function tenantToCard(t: Tenant): CardListItem {
  const companyPct = quotaProgressPct(t.max_companies, t.used_companies)
  const storePct = quotaProgressPct(t.max_stores, t.used_stores)
  const businessUnitPct = quotaProgressPct(t.max_business_units, t.used_business_units)
  const userPct = quotaProgressPct(t.max_users, t.used_users)
  return {
    id: t.id,
    title: t.name,
    subtitle: t.code,
    planName: (t.plan_name && String(t.plan_name).trim()) || '',
    planCode: (t.plan_code && String(t.plan_code).trim()) || '',
    icon: t.status === 1 ? '🏢' : '📦',
    status: t.status === 1 ? 'active' : 'inactive',
    createdAt: t.created_at ?? undefined,
    companyPct,
    storePct,
    businessUnitPct,
    userPct,
    companies: quotaUsageText(t.max_companies, t.used_companies),
    stores: quotaUsageText(t.max_stores, t.used_stores),
    businessUnits: quotaUsageText(t.max_business_units, t.used_business_units),
    users: quotaUsageText(t.max_users, t.used_users),
    contactName: t.contact_name ?? '',
    contactPhone: t.contact_phone ?? '',
  }
}

const cardData = computed(() => filteredTenants().map(tenantToCard))

const showTenantCardMoreMenu = computed(
  () =>
    perm.can('tenant:quota_config') || perm.can('tenant:status') || perm.can('tenant:reset_primary_password') || perm.can('tenant:delete'),
)

const isTrulyEmpty = computed(() => items.value.length === 0)

const selected = ref<Tenant | null>(null)
/** 详情抽屉内配额条展示（与卡片一致） */
const selectedCardMetrics = computed(() => (selected.value ? tenantToCard(selected.value) : null))
const detailCompanies = ref<TenantOrgQuotaRecord[]>([])
const detailStores = ref<TenantOrgQuotaRecord[]>([])
const detailBusinessUnits = ref<TenantBusinessUnitQuotaRecord[]>([])
const detailLoading = ref(false)

const dlg = ref(false)
const dlgEdit = ref(false)
const editLoading = ref(false)
const today = new Date().toISOString().slice(0, 10)
/** 新增主体套餐订阅：默认「结束日期 − 开始日期」为 7 天（结束日 = 开始日 + 7 个自然日） */
const TENANT_SUBSCRIPTION_DEFAULT_SPAN_DAYS = 7

function formatYmd(d: Date): string {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

function addDaysToYmd(startYmd: string, days: number): string {
  const parts = startYmd.split('-').map(Number)
  const y = parts[0]
  const m = parts[1]
  const d = parts[2]
  if (!y || !m || !d) return startYmd
  const dt = new Date(y, m - 1, d)
  dt.setDate(dt.getDate() + days)
  return formatYmd(dt)
}

function emptyCreateForm(): TenantCreatePayload {
  return {
    code: '', name: '', status: 1,
    admin_name: '超级管理员', admin_employee_no: '', admin_phone: '', admin_password: '',
  }
}
const form = ref<TenantCreatePayload>(emptyCreateForm())
const editForm = ref({
  code: '', name: '', status: 1,
  contact_name: '', contact_phone: '',
  brand_display_name: '',
})
const createStep = ref<1 | 2>(1)
const plans = ref<Plan[]>([])
const features = ref<Feature[]>([])
const selectedPlanId = ref<number | null>(null)
const planQuotas = ref<PlanQuotaValue[]>([])
const quotaValues = ref<Record<number, number>>({})
const packageLoading = ref(false)
const packageLoadError = ref('')
const packageForm = ref({
  subscription_status: 'ACTIVE',
  start_time: today,
  end_time: addDaysToYmd(today, TENANT_SUBSCRIPTION_DEFAULT_SPAN_DAYS),
  trial_end_time: '',
  auto_renew: false,
  frozen_reason: '',
})
const planDetailDlg = ref(false)
const planDetailFeatures = ref<Feature[]>([])
const planDetailQuotas = ref<PlanQuotaValue[]>([])
const editRow = ref<Tenant | null>(null)
const quotaOverrideDlg = ref(false)
const quotaOverrideSaving = ref(false)
const quotaOverrideTenant = ref<Tenant | null>(null)
const quotaOverrideRows = ref<Quota[]>([])
const quotaOverrideEnabled = ref<Record<number, boolean>>({})
const quotaOverrideValues = ref<Record<number, number>>({})
const quotaOverrideReasons = ref<Record<number, string>>({})

function dateTimeStart(date: string) {
  return date ? `${date}T00:00:00` : ''
}

function dateTimeEnd(date: string) {
  return date ? `${date}T23:59:59` : ''
}

function quotaText(value: number) {
  if (value === -1) return '无限制'
  if (value === 0) return '不可用'
  return String(value)
}

async function loadPackageOptions() {
  packageLoading.value = true
  packageLoadError.value = ''
  try {
    const [planRes, featureRes] = await Promise.all([
      fetchPlans(0, 100),
      fetchFeatures(0, 300),
    ])
    plans.value = planRes.items.filter((item) => item.status === 1)
    features.value = featureRes.items
    if (!selectedPlanId.value) {
      selectedPlanId.value = plans.value.find((item) => item.is_default)?.id ?? plans.value[0]?.id ?? null
    }
    if (selectedPlanId.value) await loadSelectedPlanQuotas()
  } catch (e) {
    packageLoadError.value = e instanceof Error ? e.message : '套餐数据加载失败'
    plans.value = []
    features.value = []
    selectedPlanId.value = null
    planQuotas.value = []
    quotaValues.value = {}
  } finally {
    packageLoading.value = false
  }
}

async function loadSelectedPlanQuotas() {
  if (!selectedPlanId.value) {
    planQuotas.value = []
    quotaValues.value = {}
    return
  }
  const data = await fetchPlanQuotas(selectedPlanId.value)
  planQuotas.value = data.quotas
  quotaValues.value = Object.fromEntries(data.quotas.map((item) => [item.quota_id, item.quota_value]))
}

async function onCreatePlanChange() {
  await loadSelectedPlanQuotas()
}

function onPackageStartDateChange() {
  const s = packageForm.value.start_time
  if (!s) return
  packageForm.value.end_time = addDaysToYmd(s, TENANT_SUBSCRIPTION_DEFAULT_SPAN_DAYS)
}

async function openPlanDetail() {
  if (!selectedPlanId.value) return
  const [featureData, quotaData] = await Promise.all([
    fetchPlanFeatures(selectedPlanId.value),
    fetchPlanQuotas(selectedPlanId.value),
  ])
  const featureIds = new Set(featureData.feature_ids)
  planDetailFeatures.value = features.value.filter((item) => featureIds.has(item.id))
  planDetailQuotas.value = quotaData.quotas
  planDetailDlg.value = true
}

async function load() {
  loading.value = true
  try {
    const skip = (page.value - 1) * limit.value
    const res = await fetchTenants(skip, limit.value)
    items.value = res.items
    total.value = res.total
  } finally {
    loading.value = false
  }
}

async function loadDetailForTenant(row: Tenant) {
  detailLoading.value = true
  try {
    const [fresh, recordData] = await Promise.all([
      fetchTenant(row.id),
      fetchTenantQuotaRecords(row.id),
    ])
    selected.value = fresh
    detailCompanies.value = recordData.companies
    detailStores.value = recordData.stores
    detailBusinessUnits.value = recordData.business_units
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '加载主体详情失败')
    throw e
  } finally {
    detailLoading.value = false
  }
}

async function openTenantDetail(row: Tenant) {
  selected.value = row
  detailDrawerVisible.value = true
  try {
    await loadDetailForTenant(row)
  } catch {
    detailDrawerVisible.value = false
    selected.value = null
    detailCompanies.value = []
    detailStores.value = []
    detailBusinessUnits.value = []
  }
}

function onDetailDrawerClosed() {
  selected.value = null
  detailCompanies.value = []
  detailStores.value = []
  detailBusinessUnits.value = []
}

function openCreate() {
  form.value = emptyCreateForm()
  createStep.value = 1
  selectedPlanId.value = null
  planQuotas.value = []
  quotaValues.value = {}
  packageLoadError.value = ''
  packageForm.value = {
    subscription_status: 'ACTIVE',
    start_time: today,
    end_time: addDaysToYmd(today, TENANT_SUBSCRIPTION_DEFAULT_SPAN_DAYS),
    trial_end_time: '',
    auto_renew: false,
    frozen_reason: '',
  }
  void loadPackageOptions()
  dlg.value = true
}

const createSaving = ref(false)
/** 第一步仅校验并进入套餐页，不落库；第二步与主体信息一并提交。 */
function saveCreate() {
  if (!form.value.code || !form.value.name) return ElMessage.warning('请填写主体编码和名称')
  if (!form.value.admin_name || !form.value.admin_employee_no || !form.value.admin_password)
    return ElMessage.warning('请填写管理员姓名、工号和初始密码')
  if (form.value.admin_password.length < 6) return ElMessage.warning('初始密码至少 6 位')
  if (!isValidOptionalPhone(form.value.admin_phone || '')) return ElMessage.warning('手机号需为 10-15 位数字')
  form.value.admin_phone = form.value.admin_phone ? normalizePhoneInput(form.value.admin_phone) : ''
  createStep.value = 2
}

async function savePackageConfig() {
  if (!selectedPlanId.value) return ElMessage.warning('请选择套餐')
  if (!packageForm.value.start_time || !packageForm.value.end_time) return ElMessage.warning('请填写订阅起止日期')
  if (packageForm.value.end_time <= packageForm.value.start_time) return ElMessage.warning('订阅结束日期须晚于开始日期')
  createSaving.value = true
  try {
    await createTenantWithPackage({
      tenant: { ...form.value },
      package: {
        plan_id: selectedPlanId.value,
        subscription_status: packageForm.value.subscription_status,
        start_time: dateTimeStart(packageForm.value.start_time),
        end_time: dateTimeEnd(packageForm.value.end_time),
        trial_end_time: packageForm.value.trial_end_time ? dateTimeEnd(packageForm.value.trial_end_time) : null,
        auto_renew: packageForm.value.auto_renew,
        frozen_reason: packageForm.value.frozen_reason || null,
        quotas: planQuotas.value.map((quota) => ({
          quota_id: quota.quota_id,
          quota_value: quotaValues.value[quota.quota_id] ?? quota.quota_value,
        })),
      },
    })
    dlg.value = false
    ElMessage.success('主体与套餐已创建，管理员可登录使用')
    await load()
  } finally {
    createSaving.value = false
  }
}

async function toggleStatus(row: Tenant) {
  const action = row.status === 1 ? '停用' : '启用'
  try {
    await ElMessageBox.confirm(`确定${action}「${row.name}」？`, '操作确认', { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' })
  } catch { return }
  await patchTenantStatus(row.id, row.status === 1 ? 0 : 1)
  await load()
  if (detailDrawerVisible.value && selected.value?.id === row.id) {
    const next = items.value.find((x) => x.id === row.id)
    if (next) await loadDetailForTenant(next)
  }
}

async function openQuotaOverrideDialog(row: Tenant) {
  quotaOverrideTenant.value = row
  quotaOverrideRows.value = []
  quotaOverrideEnabled.value = {}
  quotaOverrideValues.value = {}
  quotaOverrideReasons.value = {}
  try {
    const subscription = await fetchTenantSubscription(row.id)
    const [planQuotasData, overrideData] = await Promise.all([
      fetchPlanQuotas(subscription.plan_id),
      fetchTenantQuotaOverrides(row.id),
    ])
    const overrideByQuota = new Map(overrideData.overrides.map((item) => [item.quota_id, item]))
    quotaOverrideRows.value = planQuotasData.quotas.map((q) => ({
      id: q.quota_id,
      quota_code: q.quota_code,
      quota_name: q.quota_name,
      quota_type: 'STATIC',
      period_type: q.period_type ?? 'NONE',
      unit: q.unit ?? 'COUNT',
      status: 1,
      description: null,
    }))
    for (const quota of quotaOverrideRows.value) {
      const override = overrideByQuota.get(quota.id)
      const planVal = planQuotasData.quotas.find((q) => q.quota_id === quota.id)?.quota_value ?? 0
      quotaOverrideEnabled.value[quota.id] = !!override
      quotaOverrideValues.value[quota.id] = override?.quota_value ?? planVal
      quotaOverrideReasons.value[quota.id] = override?.reason || ''
    }
    quotaOverrideDlg.value = true
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '加载配额覆盖失败')
  }
}

async function saveQuotaOverridesForTenant() {
  if (!quotaOverrideTenant.value) return
  quotaOverrideSaving.value = true
  try {
    const overrides = quotaOverrideRows.value
      .filter((quota) => quotaOverrideEnabled.value[quota.id])
      .map((quota) => ({
        quota_id: quota.id,
        quota_value: quotaOverrideValues.value[quota.id] ?? 0,
        reason: quotaOverrideReasons.value[quota.id] || null,
      }))
    await saveTenantQuotaOverrides(quotaOverrideTenant.value.id, overrides)
    quotaOverrideDlg.value = false
    ElMessage.success('配额覆盖已保存')
    await load()
    if (selected.value?.id === quotaOverrideTenant.value.id) {
      const fresh = items.value.find((x) => x.id === quotaOverrideTenant.value?.id)
      if (fresh) await loadDetailForTenant(fresh)
    }
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '保存配额覆盖失败')
  } finally {
    quotaOverrideSaving.value = false
  }
}

const pwdResultDlg = ref(false)
const pwdResult = ref<TenantPrimaryAdminPasswordResetResult | null>(null)
const pwdResultTenantName = ref('')

function loginAccountLines(r: TenantPrimaryAdminPasswordResetResult) {
  const parts = [`工号：${r.employee_no}`]
  if (r.phone) parts.push(`手机号：${r.phone}`)
  parts.push('（登录时可使用工号或手机号作为账号）')
  return parts.join('\n')
}

function fullHandoffText(r: TenantPrimaryAdminPasswordResetResult, tenantName: string) {
  return [
    `主体：${tenantName}`,
    `姓名：${r.name}`,
    loginAccountLines(r),
    `新密码：${r.new_password}`,
  ].join('\n')
}

async function copyToClipboard(text: string) {
  try {
    await navigator.clipboard.writeText(text)
    ElMessage.success('已复制到剪贴板')
  } catch {
    ElMessage.warning('复制失败，请手动选择文本复制')
  }
}

function copyPwdResultLoginHint() {
  const r = pwdResult.value
  if (!r) return
  void copyToClipboard(loginAccountLines(r))
}

function copyPwdResultFull() {
  const r = pwdResult.value
  if (!r) return
  void copyToClipboard(fullHandoffText(r, pwdResultTenantName.value))
}

function copyPwdResultPassword() {
  const r = pwdResult.value
  if (!r) return
  void copyToClipboard(r.new_password)
}

async function resetPrimaryAdminPassword(row: Tenant) {
  let brief
  try {
    brief = await fetchTenantPrimaryAdmin(row.id)
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '获取主管理员信息失败')
    return
  }
  const loginHint = brief.phone
    ? `工号 ${brief.employee_no}、手机号 ${brief.phone}`
    : `工号 ${brief.employee_no}`
  try {
    await ElMessageBox.confirm(
      `主体「${row.name}」\n主管理员：${brief.name}\n登录账号：${loginHint}\n\n确定重置密码？系统将生成随机密码（字母与数字），请在弹窗中复制发给对方。`,
      '重置主管理员密码',
      { type: 'warning', confirmButtonText: '确定重置', cancelButtonText: '取消' },
    )
  } catch {
    return
  }
  try {
    const data = await resetTenantPrimaryAdminPassword(row.id)
    pwdResult.value = data
    pwdResultTenantName.value = row.name
    pwdResultDlg.value = true
    ElMessage.success('主管理员密码已重置')
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '重置失败')
  }
}

async function remove(row: Tenant) {
  try {
    await confirmArchiveAction({
      name: row.name,
      title: '归档主体',
      detail: '主体归档后将停用登录与业务入口，主体下公司、部门、用户、角色、权限、字典等历史数据仍会保留以供审计追溯。',
    })
  } catch {
    return
  }
  try {
    await deleteTenant(row.id)
    ElMessage.success(archiveSuccessMessage(row.name))
    if (selected.value?.id === row.id) {
      detailDrawerVisible.value = false
      onDetailDrawerClosed()
    }
    await load()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '删除失败')
  }
}

async function openEdit(row: Tenant) {
  editRow.value = row
  dlgEdit.value = true
  editLoading.value = true
  try {
    const d = await fetchTenant(row.id)
    editForm.value = {
      code: d.code,
      name: d.name,
      status: d.status,
      contact_name: d.contact_name ?? '',
      contact_phone: d.contact_phone ?? '',
      brand_display_name: d.brand_display_name ?? '',
    }
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '加载主体详情失败')
    editForm.value = {
      code: row.code,
      name: row.name,
      status: row.status,
      contact_name: row.contact_name ?? '',
      contact_phone: row.contact_phone ?? '',
      brand_display_name: row.brand_display_name ?? '',
    }
  } finally {
    editLoading.value = false
  }
}

async function saveEdit() {
  if (!editRow.value) return
  if (!isValidOptionalPhone(editForm.value.contact_phone)) return ElMessage.warning('手机号需为 10-15 位数字')
  const tid = editRow.value.id
  try {
    await updateTenant(tid, {
      name: editForm.value.name,
      status: editForm.value.status,
      contact_name: editForm.value.contact_name.trim(),
      contact_phone: editForm.value.contact_phone ? normalizePhoneInput(editForm.value.contact_phone) : '',
      brand_display_name: editForm.value.brand_display_name || null,
    })
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '保存失败')
    return
  }
  dlgEdit.value = false
  if (tid === perm.profile?.tenant_id) {
    await tenantBrand.load()
  }
  await load()
  if (detailDrawerVisible.value && selected.value?.id === tid) {
    const next = items.value.find((x) => x.id === tid)
    if (next) await loadDetailForTenant(next)
  }
}

function openSelectedTenantEdit() {
  const s = selected.value
  if (s) void openEdit(s)
}

function onDetailDrawerDropdownCommand(cmd: string) {
  const s = selected.value
  if (!s) return
  onDropdownCommand(cmd, s)
}

function onDropdownCommand(cmd: string, row: Tenant) {
  if (cmd === 'toggleStatus') void toggleStatus(row)
  else if (cmd === 'delete') void remove(row)
  else if (cmd === 'resetPwd') void resetPrimaryAdminPassword(row)
  else if (cmd === 'quotaConfig') void openQuotaOverrideDialog(row)
}

function expiryStatus(row: Tenant): { label: string; type: 'success' | 'warning' | 'danger' | 'info' } {
  if (!row.expire_date) return { label: '永久', type: 'info' }
  const now = new Date().toISOString().slice(0, 10)
  if (row.expire_date < now) return { label: '已到期', type: 'danger' }
  const daysLeft = Math.ceil((new Date(row.expire_date).getTime() - Date.now()) / 86400000)
  if (daysLeft <= 60) return { label: `还剩 ${daysLeft} 天到期`, type: 'warning' }
  if (row.start_date) {
    const used = Math.floor((Date.now() - new Date(row.start_date).getTime()) / 86400000)
    return { label: `已使用 ${Math.max(0, used)} 天`, type: 'success' }
  }
  return { label: row.expire_date, type: 'success' }
}

function handleTenantSearch(value: string) {
  searchKeyword.value = value
}

function handleTenantFilter(value: string) {
  statusFilter.value = value
}

function handleCardClick(card: CardListItem) {
  const tenant = items.value.find((t) => t.id === card.id)
  if (tenant) void openTenantDetail(tenant)
}

const coColumns: TableColumn[] = [
  { key: 'name', title: '名称', minWidth: 200 },
  { key: 'code', title: '编码', width: 120 },
  { key: 'company_name', title: '所属公司', minWidth: 140 },
  { key: 'type', title: '类型', width: 120 },
  { key: 'status', title: '状态', width: 88, align: 'center' },
]

const visibleTenantCount = computed(() => filteredTenants().length)
const activeTenantCount = computed(() => filteredTenants().filter((t) => t.status === 1).length)
const inactiveTenantCount = computed(() => filteredTenants().filter((t) => t.status === 0).length)

onMounted(async () => {
  await load()
})
</script>

<template>
  <div class="tenant-page">
    <NeuroAgentPageShell>
      <template #title>主体管理</template>
      <template #subtitle>平铺浏览全部主体；点击卡片打开详情，下属公司为参考信息。</template>
      <template #meta>
        <span>总共 {{ visibleTenantCount }} 个主体</span>
        <span class="stat-dot">·</span>
        <span class="stat stat-active">{{ activeTenantCount }} 个启用</span>
        <span class="stat-dot">·</span>
        <span class="stat stat-inactive">{{ inactiveTenantCount }} 个停用</span>
      </template>
      <template #actions>
        <el-button v-permission="'tenant:create'" class="btn-gradient" @click="openCreate">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" width="16" height="16">
            <path d="M12 5v14M5 12h14" />
          </svg>
          新增主体
        </el-button>
      </template>

      <div class="tenant-panel-stack">
        <div class="tenant-grid-scroll nav-body">
          <CardListView
            :loading="loading"
            :cards="cardData"
            :filter-options="tenantFilterOptions"
            :col-count="3"
            default-layout="grid"
            search-placeholder="搜索主体名称、编码..."
            :empty-title="isTrulyEmpty ? '暂无主体' : '无数据'"
            :empty-description="isTrulyEmpty ? '点击「新增主体」创建第一个租户' : '当前筛选条件下暂无结果'"
            :show-create="false"
            @search="handleTenantSearch"
            @filter="handleTenantFilter"
            @card-click="handleCardClick"
          >
            <template #default="{ card }">
              <div
                class="tc-card-wrapper"
                :class="{ active: detailDrawerVisible && selected?.id === card.id }"
                @click="handleCardClick(card)"
              >
                  <div class="tc-accent"></div>

                  <!-- Header: icon + name/subtitle + badge -->
                  <div class="tc-header">
                    <div class="tc-header-left">
                      <span class="tc-icon">{{ card.icon }}</span>
                      <div class="tc-title-group">
                        <div class="tc-title-row">
                          <span class="tc-title">{{ card.title }}</span>
                          <span
                            v-if="card.planName"
                            class="tc-plan-pill"
                            :class="planPillToneClass(card.planCode, card.planName)"
                            :title="String(card.planName)"
                          >{{ card.planName }}</span>
                        </div>
                        <span class="tc-subtitle">{{ card.subtitle }}</span>
                      </div>
                    </div>
                    <span
                      class="tc-badge"
                      :class="
                        items.find(x => x.id === card.id)?.status === 1 ? 'tc-badge--enabled' : 'nm-pill--inactive'
                      "
                    >
                      {{ items.find(x => x.id === card.id)?.status === 1 ? '启用' : '停用' }}
                    </span>
                  </div>

                  <!-- Contact info -->
                  <div class="tc-contact">
                    <span class="tc-contact-text">
                      {{ [card.contactName, card.contactPhone].filter(Boolean).join(' · ') || '暂无联系人' }}
                    </span>
                  </div>

                  <!-- Usage metrics (cards only; drawer keeps progress bars) -->
                  <div class="tc-usage-grid">
                    <div class="tc-usage-item">
                      <span class="tc-usage-item__label">公司</span>
                      <strong class="tc-usage-item__value">{{ card.companies }}</strong>
                    </div>
                    <div class="tc-usage-item">
                      <span class="tc-usage-item__label">门店</span>
                      <strong class="tc-usage-item__value">{{ card.stores }}</strong>
                    </div>
                    <div class="tc-usage-item">
                      <span class="tc-usage-item__label">业务单元</span>
                      <strong class="tc-usage-item__value">{{ card.businessUnits }}</strong>
                    </div>
                    <div class="tc-usage-item">
                      <span class="tc-usage-item__label">用户</span>
                      <strong class="tc-usage-item__value">{{ card.users }}</strong>
                    </div>
                  </div>

                  <!-- Footer: timestamps + actions -->
                  <div class="tc-footer">
                    <span class="tc-footer-item">
                      <span class="tc-footer-icon">🕒</span>
                      <span>{{ card.createdAt ? formatDate(card.createdAt) : '' }}</span>
                    </span>
                    <span class="tc-footer-item">
                      <span class="tc-footer-icon">🔄</span>
                      <span>{{ usageText(items.find(x => x.id === card.id)) }}</span>
                    </span>
                    <div class="tc-footer-actions">
                      <el-button v-permission="'tenant:edit'" size="default" @click.stop="openEdit(items.find(x => x.id === card.id)!)" class="tc-action-btn">
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
                        编辑
                      </el-button>
                      <el-dropdown
                        v-if="showTenantCardMoreMenu"
                        @command="(cmd: string) => onDropdownCommand(cmd, items.find(x => x.id === card.id)!)"
                        trigger="click"
                      >
                        <button class="tc-action-btn tc-action-btn--more" @click.stop>
                          <span>⋯</span>
                        </button>
                        <template #dropdown>
                          <el-dropdown-menu>
                            <el-dropdown-item v-if="perm.can('tenant:quota_config')" command="quotaConfig">
                              调整配额
                            </el-dropdown-item>
                            <el-dropdown-item v-if="perm.can('tenant:reset_primary_password')" command="resetPwd">
                              重置密码
                            </el-dropdown-item>
                            <el-dropdown-item v-if="perm.can('tenant:status')" command="toggleStatus">
                              {{ items.find((x) => x.id === card.id)?.status === 1 ? '停用' : '启用' }}
                            </el-dropdown-item>
                            <el-dropdown-item v-if="perm.can('tenant:delete')" command="delete" divided>
                              删除
                            </el-dropdown-item>
                          </el-dropdown-menu>
                        </template>
                      </el-dropdown>
                    </div>
                  </div>
                </div>
              </template>
          </CardListView>
        </div>
        <div v-if="shouldShowPager" class="nav-footer">
          <div class="pager">
            <div class="pagination-size" @click="sizeDropdownOpen = !sizeDropdownOpen" tabindex="0" @blur="sizeDropdownOpen = false">
              <span class="size-text">每页 {{ limit }} 条</span>
              <span class="size-arrow" :class="{ 'arrow-up': sizeDropdownOpen }">▾</span>
              <div class="size-dropdown" v-if="sizeDropdownOpen">
                <div
                  v-for="size in [10, 20, 50]"
                  :key="size"
                  class="size-option"
                  :class="{ 'option-active': size === limit }"
                  @click="limit = size; page = 1; sizeDropdownOpen = false; load()"
                >{{ size }} 条</div>
              </div>
            </div>
            <button
              class="pagination-btn pagination-prev"
              :disabled="page === 1"
              @click="page = page - 1; load()"
            >←</button>
            <div class="pagination-pages">
              <button
                v-for="p in Array.from({ length: Math.min(5, Math.ceil(total / limit)) }, (_, i) => i + 1)"
                :key="p"
                class="pagination-page"
                :class="{ 'page-active': p === page }"
                @click="page = p; load()"
              >{{ p }}</button>
            </div>
            <button
              class="pagination-btn pagination-next"
              :disabled="page >= Math.ceil(total / limit)"
              @click="page = page + 1; load()"
            >→</button>
            <span class="info-total">共 {{ total }} 条</span>
          </div>
        </div>
      </div>
    </NeuroAgentPageShell>

    <el-drawer
      v-model="detailDrawerVisible"
      class="tenant-detail-drawer"
      :title="selected?.name ?? '主体详情'"
      direction="rtl"
      size="min(720px, 94vw)"
      append-to-body
      @closed="onDetailDrawerClosed"
    >
      <div v-if="selected" v-loading="detailLoading" class="tenant-detail-drawer-body">
        <div class="tenant-detail-actions">
          <el-button v-permission="'tenant:edit'" type="primary" plain @click="openSelectedTenantEdit">
            编辑主体
          </el-button>
          <el-dropdown
            v-if="showTenantCardMoreMenu"
            @command="onDetailDrawerDropdownCommand"
            trigger="click"
          >
            <el-button>
              更多操作
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item v-if="perm.can('tenant:quota_config')" command="quotaConfig">
                  调整配额
                </el-dropdown-item>
                <el-dropdown-item v-if="perm.can('tenant:reset_primary_password')" command="resetPwd">
                  重置主管理员密码
                </el-dropdown-item>
                <el-dropdown-item v-if="perm.can('tenant:status')" command="toggleStatus">
                  {{ selected.status === 1 ? '停用主体' : '启用主体' }}
                </el-dropdown-item>
                <el-dropdown-item v-if="perm.can('tenant:delete')" command="delete" divided>
                  归档主体
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>

        <section class="tenant-detail-section tenant-detail-section--overview">
          <div class="detail-header tenant-detail-header-block">
            <div class="detail-title">
              <span class="detail-name">{{ selected.name }}</span>
              <span
                v-if="selected.plan_name"
                class="tc-plan-pill tc-plan-pill--inline"
                :class="planPillToneClass(selected.plan_code, selected.plan_name)"
                :title="selected.plan_name"
              >{{ selected.plan_name }}</span>
              <el-tag
                size="small"
                round
                effect="plain"
                :type="selected.status === 1 ? 'success' : undefined"
                :class="{ 'nm-status-pill--inactive': selected.status !== 1 }"
              >
                {{ selected.status === 1 ? '启用' : '停用' }}
              </el-tag>
              <el-tag size="small" :type="expiryStatus(selected).type" effect="plain">
                {{ expiryStatus(selected).label }}
              </el-tag>
            </div>
            <div class="detail-meta">
              <span class="detail-item">
                <span class="detail-label">编码</span>
                <span class="detail-value">{{ selected.code }}</span>
              </span>
              <span class="detail-divider"></span>
              <span class="detail-item">
                <span class="detail-label">创建时间</span>
                <span class="detail-value">{{ selected.created_at ? formatDate(selected.created_at) : '—' }}</span>
              </span>
              <span class="detail-divider"></span>
              <span class="detail-item">
                <span class="detail-label">联系人</span>
                <span class="detail-value">{{ [selected.contact_name, selected.contact_phone].filter(Boolean).join(' · ') || '—' }}</span>
              </span>
            </div>
          </div>
        </section>

        <template v-if="selectedCardMetrics">
          <section class="tenant-detail-section tenant-detail-section--quota">
            <div class="tenant-detail-section-label tenant-detail-section-label--usage">配额使用</div>
            <div class="tc-progress-group tenant-detail-quota">
              <div class="tc-progress-row">
                <span class="tc-progress-label">公司</span>
                <span class="tc-progress-value">{{ selectedCardMetrics.companies }}</span>
              </div>
              <div class="tc-progress-bar">
                <div class="tc-progress-fill" :style="{ width: selectedCardMetrics.companyPct + '%' }"></div>
              </div>
              <div class="tc-progress-row tc-mt-sm">
                <span class="tc-progress-label">门店</span>
                <span class="tc-progress-value">{{ selectedCardMetrics.stores }}</span>
              </div>
              <div class="tc-progress-bar">
                <div class="tc-progress-fill" :style="{ width: selectedCardMetrics.storePct + '%' }"></div>
              </div>
              <div class="tc-progress-row tc-mt-sm">
                <span class="tc-progress-label">业务单元</span>
                <span class="tc-progress-value">{{ selectedCardMetrics.businessUnits }}</span>
              </div>
              <div class="tc-progress-bar">
                <div class="tc-progress-fill" :style="{ width: selectedCardMetrics.businessUnitPct + '%' }"></div>
              </div>
              <div class="tc-progress-row tc-mt-sm">
                <span class="tc-progress-label">用户</span>
                <span class="tc-progress-value">{{ selectedCardMetrics.users }}</span>
              </div>
              <div class="tc-progress-bar">
                <div class="tc-progress-fill" :style="{ width: selectedCardMetrics.userPct + '%' }"></div>
              </div>
            </div>
          </section>
        </template>

        <section class="tenant-detail-section tenant-detail-section--records">
          <div class="tenant-detail-section-label tenant-detail-section-label--records">配额明细记录</div>
          <p class="tenant-detail-section-hint">
            公司 {{ detailCompanies.length }} 条 · 门店 {{ detailStores.length }} 条 · 业务单元 {{ detailBusinessUnits.length }} 条
          </p>
          <div class="tenant-detail-companies">
            <div v-if="detailCompanies.length" class="tenant-detail-record-group">
              <div class="tenant-detail-record-title">公司记录</div>
              <NeuroAgentListPage
                mode="el-table"
                title=""
                :columns="coColumns"
                :data="detailCompanies.map((x) => ({ ...x, type: '公司', company_name: x.company_name || '—' }))"
                :loading="detailLoading"
                :show-create="false"
                :show-selection="false"
                :show-pagination="false"
              />
            </div>
            <div v-if="detailStores.length" class="tenant-detail-record-group">
              <div class="tenant-detail-record-title">门店记录</div>
              <NeuroAgentListPage
                mode="el-table"
                title=""
                :columns="coColumns"
                :data="detailStores.map((x) => ({ ...x, type: '门店', company_name: x.company_name || '—' }))"
                :loading="detailLoading"
                :show-create="false"
                :show-selection="false"
                :show-pagination="false"
              />
            </div>
            <div v-if="detailBusinessUnits.length" class="tenant-detail-record-group">
              <div class="tenant-detail-record-title">业务单元记录</div>
              <NeuroAgentListPage
                mode="el-table"
                title=""
                :columns="coColumns"
                :data="detailBusinessUnits.map((x) => ({ ...x, type: x.bu_type || '业务单元', company_name: '—' }))"
                :loading="detailLoading"
                :show-create="false"
                :show-selection="false"
                :show-pagination="false"
              />
            </div>
          </div>
        </section>
      </div>
    </el-drawer>

    <NeuroAgentDialog
      v-model="dlg"
      title="新增主体"
      icon="🏢"
      size="large"
      width="min(980px, 94vw)"
      height="min(760px, 92vh)"
      :close-on-overlay-click="false"
    >
      <div class="nm-form">
        <div class="nm-step-strip">
          <span :class="{ active: createStep === 1, done: createStep > 1 }">1 创建主体</span>
          <span :class="{ active: createStep === 2 }">2 配置套餐与配额</span>
        </div>
        <div v-if="createStep === 1" class="nm-form-section">
          <div class="nm-form-section-title">主体信息</div>
          <div class="nm-form-item">
            <label class="nm-form-label">编码 <span class="nm-form-required">*</span></label>
            <el-input v-model="form.code" placeholder="唯一标识，如 acme" />
          </div>
          <div class="nm-form-item">
            <label class="nm-form-label">名称 <span class="nm-form-required">*</span></label>
            <el-input v-model="form.name" placeholder="主体名称，同时作为默认公司名" />
          </div>
        </div>
        <div v-if="createStep === 1" class="nm-form-section">
          <div class="nm-form-section-title">超级管理员（联系人）</div>
          <div class="nm-form-row">
            <div class="nm-form-item">
              <label class="nm-form-label">姓名 <span class="nm-form-required">*</span></label>
              <el-input v-model="form.admin_name" placeholder="默认超级管理员，可修改" />
            </div>
            <div class="nm-form-item">
              <label class="nm-form-label">工号 <span class="nm-form-required">*</span></label>
              <el-input v-model="form.admin_employee_no" placeholder="登录账号" />
            </div>
          </div>
          <div class="nm-form-row">
            <div class="nm-form-item">
              <label class="nm-form-label">手机号</label>
              <el-input
                v-model="form.admin_phone"
                inputmode="tel"
                placeholder="选填"
                @input="form.admin_phone = sanitizePhoneInput(String($event))"
              />
            </div>
            <div class="nm-form-item">
              <label class="nm-form-label">初始密码 <span class="nm-form-required">*</span></label>
              <el-input v-model="form.admin_password" type="password" show-password placeholder="至少 6 位" />
            </div>
          </div>
        </div>
        <div v-if="createStep === 2" class="nm-form-section">
          <div class="nm-form-section-title">套餐订阅</div>
          <el-alert
            v-if="packageLoadError"
            :title="packageLoadError"
            type="error"
            show-icon
            :closable="false"
          />
          <div v-if="packageLoading" class="package-state">
            <span class="package-state__spinner"></span>
            <span>正在加载套餐与默认配额...</span>
          </div>
          <el-empty v-else-if="!packageLoadError && plans.length === 0" description="暂无可用套餐，请先在套餐中心启用套餐" />
          <div class="nm-form-row">
            <div class="nm-form-item">
              <label class="nm-form-label">套餐 <span class="nm-form-required">*</span></label>
              <el-select
                v-model="selectedPlanId"
                style="width: 100%"
                :teleported="false"
                @change="onCreatePlanChange"
              >
                <el-option v-for="plan in plans" :key="plan.id" :label="`${plan.plan_name} / ${plan.plan_code}`" :value="plan.id" />
              </el-select>
            </div>
            <div class="nm-form-item nm-form-item--inline-end">
              <el-button :disabled="!selectedPlanId" @click="openPlanDetail">套餐详情</el-button>
            </div>
          </div>
          <div class="nm-form-row">
            <div class="nm-form-item">
              <label class="nm-form-label">开始日期 <span class="nm-form-required">*</span></label>
              <el-date-picker
                v-model="packageForm.start_time"
                type="date"
                value-format="YYYY-MM-DD"
                style="width: 100%"
                :teleported="false"
                @change="onPackageStartDateChange"
              />
            </div>
            <div class="nm-form-item">
              <label class="nm-form-label">结束日期 <span class="nm-form-required">*</span></label>
              <el-date-picker
                v-model="packageForm.end_time"
                type="date"
                value-format="YYYY-MM-DD"
                style="width: 100%"
                :teleported="false"
              />
            </div>
          </div>
        </div>
        <div v-if="createStep === 2" class="nm-form-section">
          <div class="nm-form-section-title">配额覆盖</div>
          <p class="quota-editor-tip">填写说明：<b>-1</b> 表示不限，<b>0</b> 表示不可用。</p>
          <el-empty
            v-if="!packageLoading && selectedPlanId && planQuotas.length === 0"
            description="当前套餐未开放与功能矩阵关联的配额"
          />
          <div v-else class="quota-editor">
            <div v-for="quota in planQuotas" :key="quota.quota_id" class="quota-editor-row">
              <div>
                <strong>{{ quota.quota_name }}</strong>
                <small>{{ quota.quota_code }} · {{ quota.unit || '-' }} · 默认 {{ quotaText(quota.quota_value) }}</small>
              </div>
              <el-input-number v-model="quotaValues[quota.quota_id]" :min="-1" controls-position="right" />
            </div>
          </div>
        </div>
      </div>
      <template #footer-right>
        <button v-if="createStep === 1" class="nm-btn nm-btn--primary" @click="saveCreate">下一步</button>
        <button
          v-else
          class="nm-btn nm-btn--primary"
          :disabled="createSaving || packageLoading || !!packageLoadError || !selectedPlanId"
          @click="savePackageConfig"
        >
          <span v-if="createSaving" class="nm-btn__spinner"></span>
          保存套餐配置
        </button>
      </template>
    </NeuroAgentDialog>

    <NeuroAgentDialog v-model="dlgEdit" title="编辑主体" icon="🏢" size="large" :loading="editLoading">
      <div class="nm-form">
        <div class="nm-form-section">
          <div class="nm-form-item">
            <label class="nm-form-label">编码</label>
            <el-input v-model="editForm.code" disabled />
          </div>
          <div class="nm-form-item">
            <label class="nm-form-label">名称</label>
            <el-input v-model="editForm.name" />
          </div>
          <div class="nm-form-item">
            <label class="nm-form-label">状态</label>
            <el-radio-group v-model="editForm.status">
              <el-radio :value="1">启用</el-radio>
              <el-radio :value="0">停用</el-radio>
            </el-radio-group>
          </div>
        </div>
        <div class="nm-form-section">
          <div class="nm-form-section-title">联系人</div>
          <div class="nm-form-row">
            <div class="nm-form-item">
              <label class="nm-form-label">姓名</label>
              <el-input v-model="editForm.contact_name" placeholder="联系人姓名" />
            </div>
            <div class="nm-form-item">
              <label class="nm-form-label">手机号</label>
              <el-input
                v-model="editForm.contact_phone"
                inputmode="tel"
                placeholder="联系人手机号"
                @input="editForm.contact_phone = sanitizePhoneInput(String($event))"
              />
            </div>
          </div>
        </div>
        <div class="nm-form-section">
          <div class="nm-form-section-title">侧栏品牌（可选）</div>
          <div class="nm-form-item">
            <label class="nm-form-label">展示名称</label>
            <el-input v-model="editForm.brand_display_name" maxlength="128" show-word-limit placeholder="留空则使用主体名称" />
          </div>
        </div>
      </div>
      <template #footer-right>
        <button class="nm-btn nm-btn--primary" :disabled="editLoading" @click="saveEdit">保存</button>
      </template>
    </NeuroAgentDialog>

    <NeuroAgentDialog v-model="planDetailDlg" title="套餐内容" icon="📦" size="large">
      <div class="package-detail">
        <div class="package-detail-section">
          <h3>功能点</h3>
          <div class="package-chip-list">
            <span v-for="feature in planDetailFeatures" :key="feature.id" class="package-chip">
              {{ feature.feature_name }}
              <small>{{ feature.feature_code }}</small>
            </span>
          </div>
        </div>
        <div class="package-detail-section">
          <h3>默认配额</h3>
          <div class="quota-editor">
            <div v-for="quota in planDetailQuotas" :key="quota.quota_id" class="quota-editor-row readonly">
              <div>
                <strong>{{ quota.quota_name }}</strong>
                <small>{{ quota.quota_code }} · {{ quota.unit || '-' }}</small>
              </div>
              <b>{{ quotaText(quota.quota_value) }}</b>
            </div>
          </div>
        </div>
      </div>
      <template #footer-right>
        <button class="nm-btn nm-btn--primary" @click="planDetailDlg = false">关闭</button>
      </template>
    </NeuroAgentDialog>

    <NeuroAgentDialog
      v-model="quotaOverrideDlg"
      title="调整配额"
      icon="⚖️"
      size="large"
      width="min(1080px, 96vw)"
      height="min(780px, 92vh)"
      :loading="quotaOverrideSaving"
    >
      <div class="tenant-quota-override-surface">
        <div class="quota-override-head">
          <strong>{{ quotaOverrideTenant?.name || '' }}</strong>
          <small>仅调整当前主体，不影响套餐默认值。-1 表示不限，0 表示不可用。</small>
        </div>
        <el-table :data="quotaOverrideRows" row-key="id" max-height="560" class="tenant-quota-override-table">
          <el-table-column label="覆盖" width="88">
            <template #default="{ row }">
              <el-switch v-model="quotaOverrideEnabled[row.id]" />
            </template>
          </el-table-column>
          <el-table-column prop="quota_name" label="配额" min-width="250">
            <template #default="{ row }">
              <div class="quota-name">
                <strong>{{ row.quota_name }}</strong>
                <small>{{ row.quota_code }}</small>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="period_type" label="周期" width="90" />
          <el-table-column prop="unit" label="单位" width="90" />
          <el-table-column label="覆盖值" width="190">
            <template #default="{ row }">
              <el-input-number
                v-model="quotaOverrideValues[row.id]"
                :disabled="!quotaOverrideEnabled[row.id]"
                :min="-1"
              />
            </template>
          </el-table-column>
          <el-table-column label="原因" min-width="220">
            <template #default="{ row }">
              <el-input
                v-model="quotaOverrideReasons[row.id]"
                class="quota-reason-input"
                :disabled="!quotaOverrideEnabled[row.id]"
              />
            </template>
          </el-table-column>
        </el-table>
      </div>
      <template #footer-right>
        <button class="nm-btn nm-btn--ghost" :disabled="quotaOverrideSaving" @click="quotaOverrideDlg = false">取消</button>
        <button class="nm-btn nm-btn--primary" :disabled="quotaOverrideSaving" @click="saveQuotaOverridesForTenant">
          <span v-if="quotaOverrideSaving" class="nm-btn__spinner"></span>
          保存
        </button>
      </template>
    </NeuroAgentDialog>

    <NeuroAgentDialog v-model="pwdResultDlg" title="密码已重置" icon="🔑" size="medium" @close="pwdResult = null">
      <template v-if="pwdResult">
        <p class="nm-dialog-text">请将以下信息复制发给对方管理员，并提醒其妥善保管密码。</p>
        <div class="nm-info-grid">
          <div class="nm-info-row">
            <span class="nm-info-label">主体</span>
            <span class="nm-info-value">{{ pwdResultTenantName }}</span>
          </div>
          <div class="nm-info-row">
            <span class="nm-info-label">姓名</span>
            <span class="nm-info-value">{{ pwdResult.name }}</span>
          </div>
          <div class="nm-info-row">
            <span class="nm-info-label">工号</span>
            <span class="nm-info-value">{{ pwdResult.employee_no }}</span>
          </div>
          <div v-if="pwdResult.phone" class="nm-info-row">
            <span class="nm-info-label">手机号</span>
            <span class="nm-info-value">{{ pwdResult.phone }}</span>
          </div>
          <div class="nm-info-row">
            <span class="nm-info-label">新密码</span>
            <span class="nm-info-value nm-info-value--mono">{{ pwdResult.new_password }}</span>
          </div>
        </div>
        <p class="nm-dialog-text nm-dialog-text--muted">登录时可使用工号或手机号作为账号。</p>
      </template>
      <template #footer-left>
        <button class="nm-btn nm-btn--ghost" @click="copyPwdResultLoginHint">复制账号说明</button>
        <button class="nm-btn nm-btn--ghost" @click="copyPwdResultPassword">复制密码</button>
      </template>
      <template #footer-right>
        <button class="nm-btn nm-btn--primary" @click="copyPwdResultFull">一键复制全部</button>
      </template>
    </NeuroAgentDialog>
  </div>
</template>

<style scoped>
@import '@/styles/theme/neuro-theme.css';

.nm-step-strip {
  display: flex;
  gap: 8px;
  margin-bottom: 18px;
}

.nm-step-strip span {
  flex: 1;
  padding: 10px 12px;
  border: 1px solid var(--neuro-border);
  border-radius: var(--neuro-radius-md);
  color: var(--neuro-text-secondary);
  font-weight: 700;
  text-align: center;
}

.nm-step-strip span.active,
.nm-step-strip span.done {
  border-color: var(--neuro-primary);
  color: var(--neuro-primary);
  background: var(--neuro-primary-10);
}

.nm-form-item--inline-end {
  align-self: end;
}

.quota-editor {
  display: grid;
  gap: 10px;
}

.quota-editor-tip {
  margin: 0 0 8px;
  color: var(--neuro-text-secondary);
  font-size: 12px;
}

.package-state {
  display: flex;
  align-items: center;
  gap: 10px;
  min-height: 56px;
  padding: 12px 14px;
  border: 1px solid var(--neuro-border);
  border-radius: var(--neuro-radius-md);
  background: var(--neuro-surface-soft);
  color: var(--neuro-text-secondary);
  font-size: 13px;
}

.package-state__spinner {
  width: 16px;
  height: 16px;
  border: 2px solid var(--neuro-primary-10);
  border-top-color: var(--neuro-primary);
  border-radius: 50%;
  animation: nm-spin 0.8s linear infinite;
}

@keyframes nm-spin {
  to { transform: rotate(360deg); }
}

.quota-editor-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 180px;
  gap: 16px;
  align-items: center;
  padding: 12px;
  border: 1px solid var(--neuro-border);
  border-radius: var(--neuro-radius-md);
  background: var(--neuro-surface-soft);
}

.quota-editor-row strong {
  display: block;
  color: var(--neuro-text);
}

.quota-editor-row small {
  display: block;
  margin-top: 4px;
  color: var(--neuro-text-secondary);
}

.quota-editor-row.readonly b {
  justify-self: end;
  color: var(--neuro-primary);
}

.package-detail {
  display: grid;
  gap: 18px;
}

.package-detail-section h3 {
  margin: 0 0 10px;
  color: var(--neuro-text);
}

.package-chip-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.package-chip {
  display: inline-flex;
  flex-direction: column;
  gap: 2px;
  min-width: 140px;
  padding: 10px 12px;
  border: 1px solid var(--neuro-border);
  border-radius: var(--neuro-radius-md);
  background: var(--neuro-surface-soft);
  color: var(--neuro-text);
}

.package-chip small {
  color: var(--neuro-text-secondary);
}

/* === Page ===
 * 勿再用 width/min-height calc(100% + 48px) + 负 margin：会多出一段可滚空白，且打断 flex 链导致主体内容被裁切。 */
.tenant-page {
  width: 100%;
  max-width: 100%;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  flex: 1 1 auto;
  min-height: 0;
  background: transparent;
  box-sizing: border-box;
}

.tenant-page :deep(.neuro-page-shell) {
  flex: 1 1 auto;
  min-height: 0;
}

.tenant-panel-stack {
  flex: 1 1 auto;
  min-height: 0;
  width: 100%;
  display: flex;
  flex-direction: column;
}

.tenant-grid-scroll {
  flex: 0 1 auto;
  min-height: 0;
}

:deep(.tenant-detail-drawer .el-drawer__body) {
  padding: 12px 16px 20px;
  overflow-y: auto;
  box-sizing: border-box;
}

.tenant-detail-drawer-body {
  min-height: 120px;
}

.tenant-detail-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-bottom: 16px;
}

.tenant-detail-header-block {
  border-radius: var(--neuro-radius-lg);
  margin-bottom: 0;
}

.tenant-detail-section {
  border: 1px solid color-mix(in srgb, var(--neuro-border) 80%, transparent);
  border-radius: var(--neuro-radius-lg);
  background: color-mix(in srgb, var(--neuro-surface) 88%, #0a1220 12%);
  box-shadow: 0 6px 16px rgba(6, 14, 26, 0.16);
  padding: 14px 16px;
  margin-bottom: 14px;
}
.tenant-detail-section--overview {
  border: none;
  background: transparent;
  box-shadow: none;
  padding: 0;
}
.tenant-detail-section--quota {
  border-color: color-mix(in srgb, var(--neuro-secondary) 20%, var(--neuro-border));
  background: color-mix(in srgb, var(--neuro-surface) 92%, #0a1220 8%);
  box-shadow: none;
}
.tenant-detail-section--records {
  border-color: color-mix(in srgb, #7f8cff 16%, var(--neuro-border));
  background: color-mix(in srgb, var(--neuro-surface) 93%, #09111f 7%);
  box-shadow: none;
}
.tenant-detail-section-label {
  font-size: 13px;
  font-weight: 600;
  color: var(--neuro-text);
  margin-bottom: 8px;
  letter-spacing: 0.02em;
  display: inline-flex;
  align-items: center;
  gap: 8px;
}
.tenant-detail-section-label::before {
  content: '';
  width: 8px;
  height: 8px;
  border-radius: 999px;
  background: color-mix(in srgb, var(--neuro-primary) 70%, #92fff2 30%);
  box-shadow: 0 0 10px color-mix(in srgb, var(--neuro-primary) 50%, transparent);
}
.tenant-detail-section-label--block {
  margin-bottom: 12px;
}
.tenant-detail-section--overview .tenant-detail-section-label {
  font-size: 14px;
  font-weight: 700;
}
.tenant-detail-section--quota .tenant-detail-section-label,
.tenant-detail-section--records .tenant-detail-section-label {
  font-size: 12px;
  font-weight: 600;
  color: var(--neuro-text-secondary);
}
.tenant-detail-section--quota .tenant-detail-section-label::before,
.tenant-detail-section--records .tenant-detail-section-label::before {
  width: 6px;
  height: 6px;
  opacity: 0.7;
}
.tenant-detail-section-label--usage {
  margin-top: 0;
}
.tenant-detail-section-label--records {
  margin-top: 0;
  margin-bottom: 10px;
}

.tenant-detail-quota {
  margin-bottom: 0;
}

.tenant-detail-section-hint {
  margin: 0 0 12px;
  font-size: 12px;
  color: var(--neuro-text-secondary);
}

.tenant-detail-companies :deep(.neuro-agent-list-page) {
  min-height: auto;
}

.tenant-detail-record-group {
  margin-bottom: 14px;
}

.tenant-detail-record-title {
  margin: 0 0 8px;
  font-size: 13px;
  font-weight: 600;
  color: var(--neuro-text);
}

.tenant-detail-companies :deep(.neuro-agent-header) {
  padding: 8px 0;
}

.tenant-detail-companies :deep(.neuro-table-card) {
  border-radius: var(--neuro-radius-lg);
  overflow: hidden;
}

.stat-active { color: #10b981; font-weight: 600; }
.stat-inactive { color: var(--neuro-text-secondary); }
.stat-dot { opacity: 0.5; }

/* 导航面板主体：仅浏览器主滚动条，不在此层再建纵向滚动 */
.nav-body {
  flex: 0 1 auto;
  min-height: 0;
  overflow-x: hidden;
  overflow-y: visible;
  padding: 8px 12px 12px;
}

/* 导航面板底部 */
.nav-footer {
  padding: 8px 16px;
  border-top: 1px solid var(--neuro-primary-10);
  background: transparent;
  flex-shrink: 0;
}

/* 详情头部（抽屉内复用） */
.detail-header {
  padding: 12px 20px;
  background: linear-gradient(135deg, var(--neuro-primary-10) 0%, transparent 60%);
  border-bottom: 1px solid var(--neuro-primary-10);
  flex-shrink: 0;
}

.detail-title {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 6px;
}

.detail-name {
  font-size: 16px;
  font-weight: 700;
  color: var(--neuro-text);
  letter-spacing: 0.02em;
}

.detail-meta {
  display: flex;
  align-items: center;
  gap: 16px;
  flex-wrap: wrap;
}

.detail-item {
  display: flex;
  align-items: baseline;
  gap: 6px;
  font-size: 13px;
}

.detail-label {
  color: var(--neuro-text-secondary);
  font-weight: 500;
}

.detail-value {
  color: var(--neuro-text);
}

.detail-value--accent {
  color: var(--neuro-primary);
  font-weight: 700;
}

.detail-divider {
  width: 1px;
  height: 14px;
  background: var(--neuro-border);
}

/* === Override CardListView（主体卡片区） === */
.nav-body :deep(.card-list-view) {
  min-height: 0;
}

.nav-body :deep(.card-list-toolbar-wrap) {
  margin-bottom: 10px;
  padding: 0;
  background: transparent !important;
  border: none !important;
  border-radius: 0;
}

.nav-body :deep(.card-list-toolbar--primary),
.nav-body :deep(.card-list-toolbar--filters) {
  padding-left: 0;
  padding-right: 0;
  background: transparent !important;
}

.nav-body :deep(.card-grid) {
  gap: 6px;
}

.nav-body :deep(.card-grid--list) {
  gap: 10px;
}

.nav-body :deep(.card-grid-item) {
  opacity: 1;
  transform: none;
  animation: none;
}

/* === Tenant Card (custom, no NeuroAgentCard) === */
.tc-card-wrapper {
  position: relative;
  background: var(--neuro-surface);
  border: 1px solid var(--neuro-border);
  border-radius: var(--neuro-radius-xl);
  padding: 16px 20px;
  overflow: hidden;
  transition: all var(--neuro-transition-normal);
  cursor: pointer;
}

.tc-card-wrapper:hover {
  border-color: var(--neuro-primary);
  box-shadow: 0 0 24px var(--neuro-primary-20), 0 0 60px var(--neuro-primary-10), 0 6px 20px rgba(0, 0, 0, 0.2);
  transform: translateY(-2px);
}

.tc-card-wrapper.active {
  border-color: var(--neuro-primary-30);
  box-shadow: 0 0 16px var(--neuro-primary-10), 0 4px 12px rgba(0, 0, 0, 0.15);
}

.tc-accent {
  position: absolute;
  left: 0;
  top: 8px;
  bottom: 8px;
  width: 3px;
  border-radius: 2px;
  background: transparent;
  transition: background var(--neuro-transition-fast);
  z-index: 2;
  pointer-events: none;
}

.tc-card-wrapper.active .tc-accent {
  background: var(--neuro-gradient-primary);
}

/* Header */
.tc-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.tc-header-left {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
}

.tc-icon {
  font-size: 24px;
  line-height: 1;
  flex-shrink: 0;
}

.tc-title-group {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.tc-title-row {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
  flex-wrap: wrap;
}

.tc-title {
  font-size: 16px;
  font-weight: 700;
  color: var(--neuro-text);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  min-width: 0;
  flex-shrink: 1;
}

.tc-plan-pill {
  flex-shrink: 0;
  max-width: 140px;
  padding: 4px 12px;
  border-radius: 999px;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.04em;
  line-height: 1.3;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  border: 1px solid transparent;
}

.tc-plan-pill--trial {
  color: #5eead4;
  background: color-mix(in srgb, #5eead4 18%, transparent);
  border-color: color-mix(in srgb, #5eead4 42%, transparent);
}
.tc-plan-pill--basic {
  color: #93c5fd;
  background: color-mix(in srgb, #60a5fa 18%, transparent);
  border-color: color-mix(in srgb, #60a5fa 45%, transparent);
}
.tc-plan-pill--pro {
  color: #c4b5fd;
  background: color-mix(in srgb, #a78bfa 20%, transparent);
  border-color: color-mix(in srgb, #a78bfa 48%, transparent);
}
.tc-plan-pill--enterprise {
  color: #fcd34d;
  background: color-mix(in srgb, #f59e0b 18%, transparent);
  border-color: color-mix(in srgb, #f59e0b 42%, transparent);
}

.tc-plan-pill--tone-0 {
  color: #2dd4bf;
  background: color-mix(in srgb, #2dd4bf 16%, transparent);
  border-color: color-mix(in srgb, #2dd4bf 40%, transparent);
}
.tc-plan-pill--tone-1 {
  color: #a78bfa;
  background: color-mix(in srgb, #a78bfa 18%, transparent);
  border-color: color-mix(in srgb, #a78bfa 42%, transparent);
}
.tc-plan-pill--tone-2 {
  color: #fb7185;
  background: color-mix(in srgb, #fb7185 16%, transparent);
  border-color: color-mix(in srgb, #fb7185 38%, transparent);
}
.tc-plan-pill--tone-3 {
  color: #38bdf8;
  background: color-mix(in srgb, #38bdf8 18%, transparent);
  border-color: color-mix(in srgb, #38bdf8 42%, transparent);
}
.tc-plan-pill--tone-4 {
  color: #a3e635;
  background: color-mix(in srgb, #84cc16 18%, transparent);
  border-color: color-mix(in srgb, #84cc16 40%, transparent);
}
.tc-plan-pill--tone-5 {
  color: #fdba74;
  background: color-mix(in srgb, #fb923c 18%, transparent);
  border-color: color-mix(in srgb, #fb923c 42%, transparent);
}
.tc-plan-pill--tone-6 {
  color: #e879f9;
  background: color-mix(in srgb, #d946ef 16%, transparent);
  border-color: color-mix(in srgb, #d946ef 38%, transparent);
}
.tc-plan-pill--tone-7 {
  color: #67e8f9;
  background: color-mix(in srgb, #22d3ee 18%, transparent);
  border-color: color-mix(in srgb, #22d3ee 42%, transparent);
}

.tc-plan-pill--inline {
  max-width: 200px;
  vertical-align: middle;
}

.tc-subtitle {
  font-size: 13px;
  color: var(--neuro-text-secondary);
  opacity: 0.5;
}

/* Badge */
.tc-badge {
  padding: 6px 16px;
  border-radius: 999px;
  font-size: 13px;
  font-weight: 600;
  letter-spacing: 0.02em;
  border: none;
  flex-shrink: 0;
}
.tc-badge--enabled {
  background: var(--neuro-success-20);
  color: var(--neuro-success);
}

/* Contact */
.tc-contact {
  margin-bottom: 8px;
}
.tc-contact-text {
  font-size: 13px;
  color: var(--neuro-text-secondary);
  opacity: 0.7;
}

/* Card usage metrics */
.tc-usage-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px 14px;
  margin-top: 4px;
}

.tc-usage-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
  padding: 6px 0 2px;
}

.tc-usage-item__label {
  font-size: 12px;
  color: var(--neuro-text-secondary);
  opacity: 0.7;
}

.tc-usage-item__value {
  font-size: 18px;
  line-height: 1.15;
  color: color-mix(in srgb, var(--neuro-primary) 56%, var(--neuro-text-secondary));
  font-weight: 800;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.tc-usage-item::after {
  content: '';
  display: block;
  width: 100%;
  height: 3px;
  border-radius: 999px;
  background:
    linear-gradient(90deg, color-mix(in srgb, var(--neuro-primary) 34%, transparent), transparent 72%);
  opacity: 0.5;
}

/* Progress bars (drawer detail) */
.tc-progress-group {
  display: flex;
  flex-direction: column;
  gap: 14px;
}
.tc-progress-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  min-height: 24px;
}
.tc-mt-sm {
  margin-top: 14px;
}
.tc-progress-label {
  font-size: 14px;
  color: var(--neuro-text-secondary);
  opacity: 0.8;
}
.tc-progress-value {
  font-size: 14px;
  color: var(--neuro-primary);
  font-weight: 700;
}
.tc-progress-bar {
  position: relative;
  height: 8px;
  background: var(--neuro-border);
  border-radius: var(--neuro-radius-sm);
  overflow: hidden;
}
.tc-progress-fill {
  height: 100%;
  background: linear-gradient(90deg, var(--neuro-primary), var(--neuro-secondary));
  border-radius: var(--neuro-radius-sm);
  transition: width 0.5s ease;
}

/* Footer */
.tc-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid var(--neuro-border);
}
.tc-footer-item {
  display: flex;
  align-items: center;
  gap: 6px;
  color: var(--neuro-text-secondary);
  opacity: 0.5;
  font-size: 13px;
}
.tc-footer-icon {
  font-size: 16px;
}
.tc-footer-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-left: auto;
  flex-shrink: 0;
}

.tc-action-btn {
  display: inline-flex !important;
  align-items: center !important;
  gap: 6px;
  padding: 8px 16px !important;
  height: auto !important;
  line-height: 1 !important;
  border: 1px solid var(--nm-border) !important;
  border-radius: 12px !important;
  background: var(--nm-bg-elevated) !important;
  color: var(--nm-text-primary) !important;
  font-size: 13px !important;
  font-weight: 500 !important;
  transition: all 0.25s ease;
  cursor: pointer;
  white-space: nowrap;
}
.tc-action-btn:hover {
  border-color: var(--nm-primary) !important;
  background: rgba(0, 245, 212, 0.06) !important;
  color: var(--nm-primary) !important;
}
.tc-action-btn svg {
  flex-shrink: 0;
}
.tc-action-btn--more {
  display: inline-flex !important;
  align-items: center !important;
  justify-content: center !important;
  gap: 0 !important;
  padding: 8px 14px !important;
  height: auto !important;
  line-height: 1 !important;
  min-height: unset !important;
  border: 1px solid var(--nm-border) !important;
  border-radius: 12px !important;
  background: var(--nm-bg-elevated) !important;
  color: var(--nm-text-secondary) !important;
  font-size: 16px !important;
  font-weight: 400 !important;
  cursor: pointer;
  white-space: nowrap;
}
.tc-action-btn--more:hover {
  background: rgba(0, 245, 212, 0.06) !important;
  border-color: var(--nm-primary) !important;
  color: var(--nm-primary) !important;
}

/* === 分页器 === */
.pager {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
}

.pagination-size {
  position: relative;
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 5px 10px;
  background: var(--neuro-surface);
  border: 1px solid var(--neuro-border);
  border-radius: var(--neuro-radius-md);
  cursor: pointer;
  transition: all var(--neuro-transition-fast);
  user-select: none;
}
.pagination-size:hover {
  border-color: var(--neuro-primary-20);
}
.size-text {
  font-size: 11px;
  color: var(--neuro-text-secondary);
}
.size-arrow {
  font-size: 9px;
  color: var(--neuro-text-secondary);
  transition: transform 0.2s ease;
}
.size-arrow.arrow-up { transform: rotate(180deg); }
.size-dropdown {
  position: absolute;
  bottom: calc(100% + 6px);
  left: 0;
  background: var(--neuro-surface-90);
  border: 1px solid var(--neuro-primary-20);
  border-radius: var(--neuro-radius-md);
  padding: 4px 0;
  min-width: 90px;
  z-index: 100;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.3);
}
.size-option {
  padding: 5px 12px;
  font-size: 11px;
  color: var(--neuro-text-secondary);
  white-space: nowrap;
  transition: all 0.15s ease;
}
.size-option:hover {
  background: var(--neuro-primary-10);
  color: var(--neuro-primary);
}
.size-option.option-active {
  color: var(--neuro-primary);
  font-weight: 600;
}

.pagination-btn {
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--neuro-surface);
  border: 1px solid var(--neuro-border);
  border-radius: var(--neuro-radius-md);
  color: var(--neuro-text-secondary);
  font-size: 11px;
  cursor: pointer;
  transition: all var(--neuro-transition-fast);
}
.pagination-btn:hover:not(:disabled) {
  border-color: var(--neuro-primary-20);
  color: var(--neuro-primary);
}
.pagination-btn:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

.pagination-pages {
  display: flex;
  align-items: center;
  gap: 4px;
}
.pagination-page {
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--neuro-surface);
  border: 1px solid var(--neuro-border);
  border-radius: var(--neuro-radius-md);
  color: var(--neuro-text-secondary);
  font-size: 11px;
  cursor: pointer;
  transition: all var(--neuro-transition-fast);
}
.pagination-page:hover {
  border-color: var(--neuro-primary-20);
  color: var(--neuro-primary);
}
.pagination-page.page-active {
  background: var(--neuro-primary);
  border-color: var(--neuro-primary);
  color: var(--neuro-background);
  font-weight: 700;
}

.info-total {
  font-size: 11px;
  color: var(--neuro-text-secondary);
  margin-left: 4px;
}

.quota-override-head {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin-bottom: 12px;
  padding: 2px 2px 0;
}

.quota-override-head strong {
  color: var(--neuro-text);
}

.quota-override-head small {
  color: var(--neuro-text-secondary);
  font-size: 12px;
}

.tenant-quota-override-surface {
  border: 1px solid color-mix(in srgb, var(--neuro-border-strong) 90%, #ffffff 10%);
  border-radius: 14px;
  padding: 12px;
  background: color-mix(in srgb, var(--neuro-surface) 98%, #0a0f1f 2%);
}

.tenant-quota-override-table {
  --el-table-bg-color: color-mix(in srgb, var(--neuro-surface) 97%, #000 3%);
  --el-table-tr-bg-color: color-mix(in srgb, var(--neuro-surface) 97%, #000 3%);
  --el-table-header-bg-color: color-mix(in srgb, var(--neuro-surface) 90%, #101c2c 10%);
  --el-table-border-color: color-mix(in srgb, var(--neuro-border-strong) 90%, #fff 10%);
  --el-table-text-color: var(--neuro-text);
  --el-table-header-text-color: color-mix(in srgb, var(--neuro-text) 84%, white 16%);
  --el-table-row-hover-bg-color: color-mix(in srgb, var(--neuro-surface) 86%, #16263c 14%);
}

.tenant-quota-override-table :deep(.el-input-number) {
  width: 100%;
}

.tenant-quota-override-table :deep(.el-table__inner-wrapper::before),
.tenant-quota-override-table :deep(.el-table__border-left-patch) {
  background: transparent;
}

.tenant-quota-override-table :deep(.el-table__body tr),
.tenant-quota-override-table :deep(.el-table__body td) {
  background: transparent;
}

.tenant-quota-override-table :deep(.el-input__wrapper),
.tenant-quota-override-table :deep(.el-input-number),
.tenant-quota-override-table :deep(.el-input-number .el-input__wrapper) {
  background: color-mix(in srgb, #0c1424 95%, #1a2740 5%);
  box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--neuro-border-strong) 88%, #fff 12%);
  border-radius: 10px;
}

.tenant-quota-override-table :deep(.el-input__inner) {
  color: var(--neuro-text);
}

.tenant-quota-override-table :deep(.el-input.is-disabled .el-input__wrapper),
.tenant-quota-override-table :deep(.el-input-number.is-disabled .el-input__wrapper) {
  background: color-mix(in srgb, #0a111d 84%, #0a0a0a 16%);
  box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--neuro-border-strong) 82%, #000 18%);
}

.tenant-quota-override-table :deep(.el-input-number__decrease),
.tenant-quota-override-table :deep(.el-input-number__increase) {
  background: color-mix(in srgb, #121c31 94%, #263752 6%);
  color: color-mix(in srgb, var(--neuro-text) 86%, white 14%);
  border-color: color-mix(in srgb, var(--neuro-border-strong) 86%, #fff 14%);
}

.tenant-quota-override-table :deep(.el-input-number__decrease) {
  border-right: 1px solid color-mix(in srgb, var(--neuro-border-strong) 86%, #fff 14%);
}

.tenant-quota-override-table :deep(.el-input-number__increase) {
  border-left: 1px solid color-mix(in srgb, var(--neuro-border-strong) 86%, #fff 14%);
}

.tenant-quota-override-table :deep(.el-input-number .el-input__inner) {
  text-align: center;
}

.tenant-quota-override-table :deep(.el-input),
.tenant-quota-override-table :deep(.quota-reason-input) {
  width: 100%;
}

/* Old dialog styles cleanup */
</style>
