<script setup lang="ts">
defineOptions({ name: 'AppCenterListView' })

import { Box, Download, Grid, Monitor, Plus, Refresh, Search, Setting, Upload } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'

import { fetchDictItemsByCode, type DictItemRow } from '@/api/dict'
import { fetchTenants, type Tenant } from '@/api/tenant'
import { fetchUsers, type UserRow } from '@/api/user'
import { formatDateTimeChina } from '@/utils/datetime'
import NeuroAgentDialog from '@/views/components/NeuroAgentDialog.vue'
import NeuroAgentPageShell from '@/views/components/NeuroAgentPageShell.vue'

import { createAppCenterApp, downloadAppManifestTemplate, fetchAppCenterApp, fetchAppCenterApps, fetchAppCenterStats, parseAppManifestFile, updateAppCenterApp, updateAppCenterAppStatus } from '../api'
import type { AppCenterApp, AppCenterClient, AppCenterCreatePayload, AppCenterStats, AppCenterUpdatePayload, AppManifestParseResult } from '../types'

const emptyStats: AppCenterStats = {
  total: 0,
  online: 0,
  beta: 0,
  developing: 0,
  builtin: 0,
  disabled: 0,
  categories: 0,
  client_apps: 0,
  tenant_openings: 0,
  trial_invites: 0,
  manifest_loads: 0,
  audit_logs: 0,
}
const stats = ref<AppCenterStats>({ ...emptyStats })

const dicts = ref<Record<string, DictItemRow[]>>({
  app_type: [],
  app_status: [],
  app_source: [],
  app_charge_mode: [],
  app_visibility_scope: [],
  app_client_type: [],
})

const fallbackLabels: Record<string, Record<string, string>> = {
  app_type: {
    SYSTEM_APP: '系统底座',
    BUSINESS_APP: '业务系统',
    ABILITY_APP: '业务中台',
    API_APP: 'API 应用',
    CONNECTOR_APP: '连接器',
    AI_APP: 'AI / Agent',
    SUITE_APP: '组合套件',
  },
  app_status: {
    INITIATED: '立项',
    DRAFT: '立项',
    PLANNED: '规划中',
    DEVELOPING: '开发中',
    BETA: 'Beta',
    ONLINE: '已上线',
    DISABLED: '已停用',
    ARCHIVED: '已归档',
  },
  app_source: {
    BUILTIN: '系统内置',
    MANUAL: '手工创建',
    MANIFEST: '声明文件装载',
  },
  app_charge_mode: {
    FREE: '免费',
    SUBSCRIPTION: '订阅制',
    BUYOUT: '订阅制',
    USAGE_BASED: '按量收费',
    MIXED: '组合收费',
    NON_SELLABLE: '非售卖',
  },
  app_visibility_scope: {
    PLATFORM_ONLY: '仅平台',
    TENANT: '租户可用',
    GLOBAL: '全局可见',
  },
  app_client_type: {
    PC_WEB: 'PC Web',
    API_ONLY: 'API Only',
    H5: 'H5',
    IOS: 'iOS',
    ANDROID: 'Android',
    HARMONYOS: '鸿蒙',
    WINDOWS: 'Windows',
    MACOS: 'macOS',
    MINIAPP: '小程序',
    WECHAT: '企业微信',
    DINGTALK: '钉钉',
    FEISHU: '飞书',
  },
}

const appTypeOrder = ['SYSTEM_APP', 'BUSINESS_APP', 'ABILITY_APP', 'API_APP', 'CONNECTOR_APP', 'AI_APP', 'SUITE_APP']

const appTypeOptions = computed(() => {
  const order = new Map(appTypeOrder.map((value, index) => [value, index]))
  return dictOptions('app_type')
    .filter((item) => item.value !== 'CLIENT_APP')
    .sort((a, b) => (order.get(a.value) ?? 99) - (order.get(b.value) ?? 99))
})

const deploymentModeOptions = [
  { label: '合并部署', value: 'MERGED' },
  { label: '独立部署', value: 'STANDALONE' },
]

const communicationModeOptions = [
  { key: 'PLATFORM_API', label: '平台 API' },
  { key: 'WEBHOOK', label: 'Webhook' },
  { key: 'DATA_SYNC', label: '数据同步' },
  { key: 'GATEWAY_PROXY', label: '网关代理' },
]

const fallbackClientOptions = [
  { key: 'PC_WEB', label: 'PC Web' },
  { key: 'API_ONLY', label: 'API Only' },
  { key: 'H5', label: 'H5' },
  { key: 'IOS', label: 'iOS' },
  { key: 'ANDROID', label: 'Android' },
  { key: 'HARMONYOS', label: '鸿蒙' },
  { key: 'WINDOWS', label: 'Windows' },
  { key: 'MACOS', label: 'macOS' },
  { key: 'MINIAPP', label: '小程序' },
  { key: 'WECHAT', label: '企业微信' },
  { key: 'DINGTALK', label: '钉钉' },
  { key: 'FEISHU', label: '飞书' },
]

const trialPolicyOptions = [
  { value: '不支持试用', label: '不支持试用', hint: '租户必须通过套餐包含、应用中心开通或邀请码开通获得应用。' },
  { value: '不限期', label: '不限期', hint: '租户获得并开通应用后进入不限期试用，后续仍可通过授权或套餐收口。' },
  { value: '7 天', label: '7 天', hint: '租户获得并开通应用后开始计算 7 天。' },
  { value: '15 天', label: '15 天', hint: '租户获得并开通应用后开始计算 15 天。' },
  { value: '30 天', label: '30 天', hint: '租户获得并开通应用后开始计算 30 天。' },
  { value: 'CUSTOM', label: '自定义', hint: '自定义试用时间，适合灰度、试点或合同约定周期。' },
]

const customTrialPolicyValue = 'CUSTOM'
const trialEligibleChargeModes = new Set(['SUBSCRIPTION', 'USAGE_BASED', 'MIXED'])
const trialTransitionRules = [
  { title: '试用期间付费', content: '立即转入付费状态，并从付费成功时开始计时。' },
  { title: '试用期结束未付费', content: '自动终止试用，不继续保留试用权益。' },
]

const loading = ref(false)
const apps = ref<AppCenterApp[]>([])
const total = ref(0)
const keyword = ref('')
const typeFilter = ref('')
const statusFilter = ref('')
const sourceFilter = ref('')
const activeSection = ref('apps')
const route = useRoute()
const createDialogVisible = ref(false)
const editDialogVisible = ref(false)
const detailDialogVisible = ref(false)
const createMode = ref<'manual' | 'manifest'>('manual')
const manifestFileInputRef = ref<HTMLInputElement | null>(null)
const manifestFileName = ref('')
const manifestImportSummary = ref('')
const saving = ref(false)
const detailLoading = ref(false)
const detailApp = ref<AppCenterApp | null>(null)
const pickerPageSize = 10
const ownerUsers = ref<UserRow[]>([])
const ownerUsersLoading = ref(false)
const ownerUsersTotal = ref(0)
const ownerUsersSkip = ref(0)
const createOwnerIds = ref<string[]>([])
const editOwnerIds = ref<string[]>([])
const ownerPickerVisible = ref(false)
const ownerPickerTarget = ref<'create' | 'edit'>('create')
const ownerPickerKeyword = ref('')
const ownerPickerDraftIds = ref<string[]>([])
const tenantOptions = ref<Tenant[]>([])
const tenantOptionsLoading = ref(false)
const tenantOptionsTotal = ref(0)
const tenantOptionsSkip = ref(0)
const createTenantIds = ref<string[]>([])
const editTenantIds = ref<string[]>([])
const tenantPickerVisible = ref(false)
const tenantPickerTarget = ref<'create' | 'edit'>('create')
const tenantPickerKeyword = ref('')
const tenantPickerDraftIds = ref<string[]>([])
let createBodyScrollTimer: number | undefined
let editBodyScrollTimer: number | undefined
let ownerSearchTimer: number | undefined
let tenantSearchTimer: number | undefined

const statCards = computed(() => [
  { label: '所有应用', value: stats.value.total, status: '', tone: 'default', hint: '全部应用' },
  { label: '已上线', value: stats.value.online, status: 'ONLINE', tone: 'success', hint: '正式可用' },
  { label: 'Beta 版', value: stats.value.beta, status: 'BETA', tone: 'warning', hint: '测试验证中' },
  { label: '规划开发中', value: stats.value.developing, status: 'PLANNED,DEVELOPING', tone: 'info', hint: '建设中' },
  { label: '已停用', value: stats.value.disabled, status: 'DISABLED', tone: 'muted', hint: '不可选用' },
])

const statusFilterOptions = computed(() => [
  ...dictOptions('app_status'),
])

const chargeModeOptions = computed(() => dictOptions('app_charge_mode').filter((item) => item.value !== 'BUYOUT'))

const commercialModeOptions = computed(() =>
  chargeModeOptions.value.map((item) => ({
    ...item,
    description: commercialModeDescription(item.value),
  })),
)

const commercialKindOptions = [
  { value: 'FREE', label: '免费', description: '租户可免费获得，不配置试用期。' },
  { value: 'PAID', label: '收费', description: '需要选择收费模式，可配置是否包含试用。' },
  { value: 'NON_SELLABLE', label: '非售卖', description: '平台内置或治理能力，不作为商品售卖。' },
]

const paidChargeModeOptions = computed(() => commercialModeOptions.value.filter((item) => chargeModeAllowsTrial(item.value)))

const clientTypeOptions = computed(() => {
  const rows = dictOptions('app_client_type')
  if (rows.length === 0) return fallbackClientOptions
  return rows.map((item) => ({ key: item.value, label: item.label }))
})

const createForm = ref<AppCenterCreatePayload>(newCreateForm())
const createStep = ref('basic')
const createScrollRef = ref<HTMLElement | null>(null)
const createBodyScrolling = ref(false)
const createDraft = ref(newCreateDraft())
const editAppId = ref<number | null>(null)
const editAppSnapshot = ref<AppCenterApp | null>(null)
const editForm = ref<AppCenterUpdatePayload>(newEditForm())
const editStep = ref('basic')
const editScrollRef = ref<HTMLElement | null>(null)
const editBodyScrolling = ref(false)
const editDraft = ref(newEditDraft())

const statusSelectValue = computed({
  get: () => (statusFilter.value.includes(',') ? '' : statusFilter.value),
  set: (value: string) => {
    statusFilter.value = value
  },
})

const filteredOwnerUsers = computed(() => {
  const kw = ownerPickerKeyword.value.trim().toLowerCase()
  if (!kw) return ownerUsers.value
  return ownerUsers.value.filter((user) => {
    return [user.name, user.employee_no, user.phone, user.email]
      .filter(Boolean)
      .some((value) => String(value).toLowerCase().includes(kw))
  })
})

const ownerUsersHasMore = computed(() => ownerUsers.value.length < ownerUsersTotal.value)

const filteredTenantOptions = computed(() => {
  const kw = tenantPickerKeyword.value.trim().toLowerCase()
  if (!kw) return tenantOptions.value
  return tenantOptions.value.filter((tenant) => {
    return [tenant.name, tenant.code, tenant.contact_name, tenant.contact_phone]
      .filter(Boolean)
      .some((value) => String(value).toLowerCase().includes(kw))
  })
})

const tenantOptionsHasMore = computed(() => tenantOptions.value.length < tenantOptionsTotal.value)

const createSteps = [
  { key: 'basic', title: '基础信息', hint: 'app_code、类型、部署' },
  { key: 'visibility', title: '可见范围', hint: '全部租户、指定租户' },
  { key: 'commercial', title: '收费策略', hint: '免费、收费、非售卖' },
  { key: 'clients', title: '客户端', hint: 'PC Web、API、移动端' },
]

const selectedCreateClients = computed(() => {
  return createDraft.value.clients.filter((item) => item.checked).map((item) => item.label)
})
function commercialStepCompleted(draft: { commercial_kind?: string; paid_charge_mode?: string; trial_days: string; trial_custom_time?: string }) {
  if (!draft.commercial_kind) return false
  if (draft.commercial_kind !== 'PAID') return true
  return Boolean(draft.paid_charge_mode && resolvedTrialPolicy(draft))
}
const createStepCompletion = computed<Record<string, boolean>>(() => ({
  basic: Boolean(createForm.value.app_code.trim() && createForm.value.app_name.trim()),
  visibility: Boolean(createDraft.value.visibility_mode),
  commercial: commercialStepCompleted(createDraft.value),
  clients: selectedCreateClients.value.length > 0,
}))
const selectedEditClients = computed(() => {
  return editDraft.value.clients.filter((item) => item.checked).map((item) => item.label)
})
const editStepCompletion = computed<Record<string, boolean>>(() => ({
  basic: Boolean(editForm.value.app_name.trim()),
  visibility: Boolean(editDraft.value.visibility_mode),
  commercial: commercialStepCompleted(editDraft.value),
  clients: selectedEditClients.value.length > 0,
}))

const groupedApps = computed(() => {
  const order = new Map(appTypeOptions.value.map((item, index) => [item.value, index]))
  const groups = new Map<string, AppCenterApp[]>()
  apps.value.forEach((app) => {
    const key = app.app_type || 'UNKNOWN'
    const rows = groups.get(key) || []
    rows.push(app)
    groups.set(key, rows)
  })
  return Array.from(groups.entries())
    .sort(([a], [b]) => {
      const aOrder = order.get(a) ?? Number.MAX_SAFE_INTEGER
      const bOrder = order.get(b) ?? Number.MAX_SAFE_INTEGER
      if (aOrder !== bOrder) return aOrder - bOrder
      return labelOf('app_type', a).localeCompare(labelOf('app_type', b), 'zh-Hans-CN')
    })
    .map(([type, items]) => ({
      type,
      label: labelOf('app_type', type),
      items,
    }))
})

const hasActiveFilters = computed(() => {
  return Boolean(keyword.value.trim() || typeFilter.value || statusFilter.value || sourceFilter.value)
})

const directoryItems = computed(() => [
  {
    key: 'apps',
    label: '应用列表',
    count: stats.value.total,
    description: '平台已登记的内置应用、业务应用与装载应用。',
  },
  {
    key: 'clients',
    label: '客户端中心',
    count: stats.value.client_apps,
    description: '已登记的应用客户端形态数量。',
  },
  {
    key: 'openings',
    label: '租户开通总览',
    count: stats.value.tenant_openings,
    description: '租户套餐订阅与开通记录总数。',
  },
  {
    key: 'invites',
    label: '体验邀请总览',
    count: stats.value.trial_invites,
    description: '试用订阅与体验邀请相关记录。',
  },
  {
    key: 'manifest',
    label: 'Manifest 装载记录',
    count: stats.value.manifest_loads,
    description: '通过 Manifest 声明文件装载的应用。',
  },
  {
    key: 'audit',
    label: '应用审计日志',
    count: stats.value.audit_logs,
    description: '应用中心相关审计事件。',
  },
])

const activeDirectoryItem = computed(() => directoryItems.value.find((item) => item.key === activeSection.value) || directoryItems.value[0])

const sectionRouteMap: Record<string, string> = {
  apps: '/apps',
  clients: '/apps/clients',
  openings: '/apps/tenant-openings',
  invites: '/apps/trial-invites',
  manifest: '/apps/manifests',
  audit: '/apps/audit-logs',
}

function sectionFromPath(path: string) {
  const hit = Object.entries(sectionRouteMap).find(([, routePath]) => routePath === path)
  return hit?.[0] || 'apps'
}

function dictOptions(code: keyof typeof dicts.value) {
  const rows = dicts.value[code] || []
  if (rows.length > 0) return rows
  return Object.entries(fallbackLabels[code] || {}).map(([value, label], index) => ({
    id: index,
    dict_type_id: 0,
    label,
    value,
    sort_order: index + 1,
    enabled: true,
  }))
}

function labelOf(code: keyof typeof dicts.value, value: string | null | undefined) {
  if (!value) return '—'
  const row = dictOptions(code).find((item) => item.value === value)
  return row?.label || fallbackLabels[code]?.[value] || value
}

function commercialModeDescription(value: string) {
  switch (value) {
    case 'FREE':
      return '免费开放，不配置试用期。'
    case 'SUBSCRIPTION':
      return '按周期订阅，可配置试用。'
    case 'USAGE_BASED':
      return '按调用量或资源用量计费。'
    case 'MIXED':
      return '订阅制与按量收费的组合。'
    case 'NON_SELLABLE':
      return '平台内置或治理能力，不作为商品售卖。'
    default:
      return '只声明商业模式，价格后续配置。'
  }
}

function chargeModeAllowsTrial(chargeMode: string | null | undefined) {
  return trialEligibleChargeModes.has(String(chargeMode || ''))
}

function trialPolicyLabel(value: string | null | undefined) {
  if (!value) return '未选择'
  const option = trialPolicyOptions.find((item) => item.value === value)
  if (!option) return value
  return option.value === customTrialPolicyValue ? '自定义' : option.label
}

function isPresetTrialPolicy(value: string | null | undefined) {
  return Boolean(value && trialPolicyOptions.some((item) => item.value === value && item.value !== customTrialPolicyValue))
}

function customTrialTimeFrom(value: string | null | undefined) {
  if (!value || isPresetTrialPolicy(value)) return ''
  return value.trim()
}

function resolvedTrialPolicy(draft: { trial_days: string; trial_custom_time?: string }) {
  if (draft.trial_days === customTrialPolicyValue) return (draft.trial_custom_time || '').trim()
  return draft.trial_days
}

function trialPolicyForSubmit(chargeMode: string | null | undefined, draft: { trial_days: string; trial_custom_time?: string }) {
  if (!chargeModeAllowsTrial(chargeMode)) return '不支持试用'
  return resolvedTrialPolicy(draft)
}

function commercialKindFromChargeMode(chargeMode: string | null | undefined) {
  if (chargeMode === 'NON_SELLABLE') return 'NON_SELLABLE'
  if (chargeModeAllowsTrial(chargeMode)) return 'PAID'
  return 'FREE'
}

function chargeModeForSubmit(currentChargeMode: string | null | undefined, draft: { commercial_kind?: string; paid_charge_mode?: string }) {
  if (draft.commercial_kind === 'NON_SELLABLE') return 'NON_SELLABLE'
  if (draft.commercial_kind === 'PAID') return draft.paid_charge_mode || ''
  if (draft.commercial_kind === 'FREE') return 'FREE'
  if (currentChargeMode === 'NON_SELLABLE') return 'NON_SELLABLE'
  return 'FREE'
}

function chooseCreateCommercialKind(kind: string) {
  createDraft.value.commercial_kind = kind
  if (kind === 'PAID' && !createDraft.value.paid_charge_mode) createDraft.value.paid_charge_mode = 'SUBSCRIPTION'
}

function chooseEditCommercialKind(kind: string) {
  editDraft.value.commercial_kind = kind
  if (kind === 'PAID' && !editDraft.value.paid_charge_mode) editDraft.value.paid_charge_mode = 'SUBSCRIPTION'
}

function trialStartRuleFor(policy: string | null | undefined) {
  return policy && policy !== '不支持试用' ? '获得并开通时开始' : null
}

function statusClass(status: string) {
  return {
    'app-status--online': status === 'ONLINE',
    'app-status--developing': status === 'DEVELOPING' || status === 'BETA',
    'app-status--disabled': status === 'DISABLED' || status === 'ARCHIVED',
    'app-status--draft': status === 'INITIATED' || status === 'DRAFT' || status === 'PLANNED',
  }
}

function iconFor(app: AppCenterApp) {
  if (app.app_code === 'system-management') return Setting
  if (app.app_code === 'system-monitor') return Monitor
  if (app.app_code === 'app-center') return Grid
  return Box
}

function appInitial(app: AppCenterApp) {
  return (app.app_name || app.app_code || '应用').trim().slice(0, 1)
}

function newCreateForm(): AppCenterCreatePayload {
  return {
    app_code: '',
    app_name: '',
    icon: '',
    app_type: '',
    deployment_mode: '',
    communication_modes: null,
    status: 'INITIATED',
    charge_mode: '',
    visibility_scope: 'TENANT',
    owner: '',
    owner_user_ids: '',
    version: '',
    description: '',
    detail_description: '',
    sort_order: 0,
  }
}

function newCreateDraft() {
  return {
    visibility_mode: '',
    visible_tenants: '',
    commercial_kind: '',
    paid_charge_mode: '',
    trial_days: '',
    trial_custom_time: '',
    trial_start_rule: '',
    communication_modes: communicationModeOptions.map((item) => ({
      ...item,
      checked: item.key === 'PLATFORM_API',
    })),
    clients: clientTypeOptions.value.map((item) => ({
      ...item,
      checked: false,
    })),
  }
}

function newEditDraft(app?: AppCenterApp | null) {
  const draft = newCreateDraft()
  const clientCodes = new Set((app?.clients || []).filter((item) => item.enabled !== false).map((item) => item.client_code))
  return {
    ...draft,
    communication_modes: draft.communication_modes.map((item) => ({
      ...item,
      checked: app?.communication_modes ? app.communication_modes.split(',').includes(item.key) : item.checked,
    })),
    visibility_mode: app?.visibility_mode || (app?.visibility_scope === 'GLOBAL' ? 'ALL_TENANTS' : 'SPECIFIED_TENANTS'),
    visible_tenants: app?.visible_tenants || (app?.visibility_scope === 'GLOBAL' ? '全部租户' : '按租户可见范围控制'),
    commercial_kind: app?.charge_mode ? commercialKindFromChargeMode(app.charge_mode) : draft.commercial_kind,
    paid_charge_mode: app?.charge_mode && chargeModeAllowsTrial(app.charge_mode) ? app.charge_mode : draft.paid_charge_mode,
    trial_days: app?.trial_policy ? (isPresetTrialPolicy(app.trial_policy) ? app.trial_policy : customTrialPolicyValue) : draft.trial_days,
    trial_custom_time: customTrialTimeFrom(app?.trial_policy),
    trial_start_rule: app?.trial_start_rule || trialStartRuleFor(app?.trial_policy || draft.trial_days) || '',
    clients: draft.clients.map((item) => ({
      ...item,
      checked: clientCodes.size > 0 ? clientCodes.has(item.key) : item.key === 'PC_WEB' || item.key === 'API_ONLY' || (app?.app_type === 'CONNECTOR_APP' && ['WECHAT', 'DINGTALK', 'FEISHU'].includes(item.key)),
    })),
  }
}

function newEditForm(): AppCenterUpdatePayload {
  return {
    app_name: '',
    icon: '',
    app_type: 'BUSINESS_APP',
    charge_mode: '',
    visibility_scope: 'TENANT',
    owner: '',
    owner_user_ids: '',
    version: '',
    description: '',
    detail_description: '',
    sort_order: 0,
  }
}

function appToEditForm(app: AppCenterApp): AppCenterUpdatePayload {
  return {
    app_name: app.app_name,
    icon: app.icon || '',
    app_type: app.app_type,
    deployment_mode: app.deployment_mode || 'MERGED',
    communication_modes: app.communication_modes || null,
    charge_mode: app.charge_mode,
    visibility_scope: app.visibility_scope,
    owner: app.owner || '',
    owner_user_ids: app.owner_user_ids || '',
    version: app.version || '',
    description: app.description || '',
    detail_description: app.detail_description || '',
    visibility_mode: app.visibility_mode || '',
    visible_tenants: app.visible_tenants || '',
    open_method: app.open_method || '',
    trial_policy: app.trial_policy || '',
    trial_start_rule: app.trial_start_rule || '',
    asset_config: app.asset_config || '',
    doc_config: app.doc_config || '',
    release_channel: app.release_channel || '',
    release_note: app.release_note || '',
    sort_order: app.sort_order,
  }
}

function trimNullable(value: string | null | undefined) {
  const trimmed = String(value ?? '').trim()
  return trimmed || null
}

function checkedKeys(items: Array<{ key: string; checked: boolean }>) {
  return items.filter((item) => item.checked).map((item) => item.key).join(',')
}

async function downloadManifestTemplate() {
  try {
    const blob = await downloadAppManifestTemplate()
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = 'app.manifest.yaml'
    link.click()
    URL.revokeObjectURL(url)
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : 'Manifest 模板下载失败')
  }
}

function triggerManifestImport() {
  manifestFileInputRef.value?.click()
}

function manifestValue(source: Record<string, unknown>, paths: string[]) {
  for (const path of paths) {
    const value = path.split('.').reduce<unknown>((current, key) => {
      if (!current || typeof current !== 'object') return undefined
      return (current as Record<string, unknown>)[key]
    }, source)
    if (value !== undefined && value !== null && value !== '') return value
  }
  return undefined
}

function manifestString(source: Record<string, unknown>, paths: string[]) {
  const value = manifestValue(source, paths)
  return value === undefined ? '' : String(value).trim()
}

function manifestArray(source: Record<string, unknown>, paths: string[]) {
  const value = manifestValue(source, paths)
  if (Array.isArray(value)) {
    return value
      .map((item) => {
        if (typeof item === 'object' && item) {
          const row = item as Record<string, unknown>
          return row.client_code || row.code || row.key || row.name
        }
        return item
      })
      .map((item) => String(item || '').trim())
      .filter(Boolean)
  }
  if (typeof value === 'string') {
    return value.split(/[，,\s]+/).map((item) => item.trim()).filter(Boolean)
  }
  return []
}

function allowedValue(value: string, allowed: string[]) {
  const normalized = value.trim().toUpperCase()
  return allowed.includes(normalized) ? normalized : ''
}

function applyManifestToCreateForm(manifest: Record<string, unknown>, fileName: string) {
  const rawChargeMode = manifestString(manifest, ['charge_mode', 'billing_mode', 'app.billing_mode', 'app.charge_policy', 'commercial.charge_mode', 'billing.charge_mode'])
  const chargeMode = allowedValue(rawChargeMode, ['FREE', 'SUBSCRIPTION', 'USAGE_BASED', 'MIXED', 'NON_SELLABLE'])
  const rawVisibilityMode = manifestString(manifest, ['visibility_mode', 'visibility.mode'])
  const rawVisibilityScope = manifestString(manifest, ['visibility_scope', 'app.visibility_scope', 'visibility.scope'])
  const visibilityMode = allowedValue(rawVisibilityMode, ['ALL_TENANTS', 'SPECIFIED_TENANTS'])
    || (allowedValue(rawVisibilityScope, ['GLOBAL', 'TENANT']) === 'GLOBAL' ? 'ALL_TENANTS' : '')
  const rawDeploymentMode = manifestString(manifest, ['deployment_mode', 'app.deployment_mode', 'deployment.mode'])
  const deploymentMode = allowedValue(rawDeploymentMode, ['MERGED', 'STANDALONE'])
  const rawAppType = manifestString(manifest, ['app_type', 'app.app_type', 'type', 'application.type'])
  const appType = allowedValue(rawAppType, appTypeOrder)
  const trialPolicy = manifestString(manifest, ['trial_policy', 'trial.policy', 'commercial.trial_policy', 'billing.trial_policy'])
  const communicationModes = new Set(manifestArray(manifest, ['communication_modes', 'deployment.communication_modes']).map((item) => item.toUpperCase()))
  const clientCodes = new Set(manifestArray(manifest, ['clients', 'client_codes']).map((item) => item.toUpperCase()))

  createForm.value = {
    ...createForm.value,
    app_code: manifestString(manifest, ['app_code', 'app.app_code', 'code', 'application.code']) || createForm.value.app_code,
    app_name: manifestString(manifest, ['app_name', 'app.app_name', 'name', 'application.name']) || createForm.value.app_name,
    icon: manifestString(manifest, ['icon', 'app.icon', 'application.icon']) || createForm.value.icon,
    app_type: appType || createForm.value.app_type,
    deployment_mode: deploymentMode || createForm.value.deployment_mode || 'MERGED',
    charge_mode: chargeMode || createForm.value.charge_mode,
    version: manifestString(manifest, ['version', 'app.version', 'application.version']) || createForm.value.version,
    description: manifestString(manifest, ['description', 'app.description', 'application.description']) || createForm.value.description,
    detail_description: manifestString(manifest, ['detail_description', 'detail', 'application.detail_description']) || createForm.value.detail_description,
    sort_order: Number(manifestString(manifest, ['sort_order', 'application.sort_order']) || createForm.value.sort_order || 0),
  }
  createDraft.value.visibility_mode = visibilityMode || createDraft.value.visibility_mode || 'ALL_TENANTS'
  createDraft.value.commercial_kind = chargeMode ? commercialKindFromChargeMode(chargeMode) : createDraft.value.commercial_kind
  createDraft.value.paid_charge_mode = chargeModeAllowsTrial(chargeMode) ? chargeMode : createDraft.value.paid_charge_mode
  if (trialPolicy) {
    createDraft.value.trial_days = isPresetTrialPolicy(trialPolicy) ? trialPolicy : customTrialPolicyValue
    createDraft.value.trial_custom_time = isPresetTrialPolicy(trialPolicy) ? '' : trialPolicy
  } else if (chargeModeAllowsTrial(chargeMode) && !createDraft.value.trial_days) {
    createDraft.value.trial_days = '不支持试用'
  }
  if (communicationModes.size) {
    createDraft.value.communication_modes = createDraft.value.communication_modes.map((item) => ({
      ...item,
      checked: communicationModes.has(item.key),
    }))
  }
  if (clientCodes.size) {
    createDraft.value.clients = createDraft.value.clients.map((item) => ({
      ...item,
      checked: clientCodes.has(item.key),
    }))
  }
  manifestFileName.value = fileName
  manifestImportSummary.value = 'Manifest 已解析并回填到表单，可继续修改后创建。'
  createMode.value = 'manifest'
  createStep.value = 'basic'
  nextTick(() => createScrollRef.value?.scrollTo({ top: 0 }))
}

function backendManifestToCreateSource(result: AppManifestParseResult) {
  const chargeMode = result.charge_policy === 'FREE' || result.charge_policy === 'NON_SELLABLE'
    ? result.charge_policy
    : result.billing_mode
  return {
    app: {
      app_code: result.app_code,
      app_name: result.app_name,
      app_type: result.app_type,
      source: result.source,
      status: result.status,
      deployment_mode: result.deployment_mode,
      visibility_scope: result.visibility_scope,
      charge_policy: chargeMode,
    },
    clients: result.client_codes,
  }
}

async function handleManifestFileChange(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file) return
  try {
    const result = await parseAppManifestFile(file)
    if (!result.valid || !result.importable) {
      const blocker = result.blockers?.[0] || 'Manifest 当前不能作为新建导入'
      manifestFileName.value = file.name
      manifestImportSummary.value = blocker
      createMode.value = 'manifest'
      ElMessage.error(blocker)
      return
    }
    applyManifestToCreateForm(backendManifestToCreateSource(result), file.name)
    manifestImportSummary.value = `后端已解析：菜单 ${result.counts.menus}、权限 ${result.counts.permissions}、API ${result.counts.apis}、套餐功能 ${result.counts.package_features}、配额 ${result.counts.quotas}。`
    ElMessage.success('Manifest 已由后端解析')
  } catch (error) {
    ElMessage.error(error instanceof Error ? `Manifest 解析失败：${error.message}` : 'Manifest 解析失败')
  }
}

function parseCsv(value: string | null | undefined) {
  return String(value || '')
    .split(',')
    .map((item) => item.trim())
    .filter(Boolean)
}

function parseOwnerIDs(value: string | null | undefined) {
  return parseCsv(value)
}

function ownerNames(ids: string[]) {
  const names = ids
    .map((id) => ownerUsers.value.find((user) => String(user.id) === id)?.name)
    .filter((name): name is string => Boolean(name))
  return names.join(', ')
}

function ownerSelectLabel(ids: string[]) {
  if (ids.length === 0) return ''
  return ownerNames(ids) || `已选择 ${ids.length} 人`
}

async function loadOwnerUsers(reset = false) {
  if (ownerUsersLoading.value) return
  if (!reset && ownerUsers.value.length > 0 && !ownerUsersHasMore.value) return
  if (reset) {
    ownerUsers.value = []
    ownerUsersSkip.value = 0
    ownerUsersTotal.value = 0
  }
  ownerUsersLoading.value = true
  try {
    const result = await fetchUsers({
      skip: ownerUsersSkip.value,
      limit: pickerPageSize,
      status: 1,
      keyword: ownerPickerKeyword.value.trim() || undefined,
    })
    const items = result.items || []
    ownerUsers.value = reset ? items : [...ownerUsers.value, ...items]
    ownerUsersSkip.value += items.length
    ownerUsersTotal.value = result.total ?? ownerUsers.value.length
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '成员列表加载失败')
  } finally {
    ownerUsersLoading.value = false
  }
}

function openOwnerPicker(target: 'create' | 'edit') {
  ownerPickerTarget.value = target
  ownerPickerDraftIds.value = [...(target === 'create' ? createOwnerIds.value : editOwnerIds.value)]
  ownerPickerKeyword.value = ''
  ownerPickerVisible.value = true
  void loadOwnerUsers(true)
}

function closeOwnerPicker() {
  ownerPickerVisible.value = false
}

function isOwnerDraftSelected(id: number | string) {
  return ownerPickerDraftIds.value.includes(String(id))
}

function toggleOwnerDraft(id: number | string) {
  const value = String(id)
  ownerPickerDraftIds.value = isOwnerDraftSelected(value)
    ? ownerPickerDraftIds.value.filter((item) => item !== value)
    : [...ownerPickerDraftIds.value, value]
}

function removeOwner(target: 'create' | 'edit', id: string) {
  const current = target === 'create' ? createOwnerIds : editOwnerIds
  current.value = current.value.filter((item) => item !== id)
}

function confirmOwnerPicker() {
  if (ownerPickerTarget.value === 'create') {
    createOwnerIds.value = [...ownerPickerDraftIds.value]
  } else {
    editOwnerIds.value = [...ownerPickerDraftIds.value]
  }
  ownerPickerVisible.value = false
}

function handleOwnerPickerScroll(event: Event) {
  const target = event.currentTarget as HTMLElement
  if (target.scrollTop + target.clientHeight >= target.scrollHeight - 24) {
    void loadOwnerUsers()
  }
}

function parseTenantNames(value: string | null | undefined) {
  return String(value || '')
    .split(/[，,]/)
    .map((item) => item.trim())
    .filter(Boolean)
}

function tenantNames(ids: string[]) {
  const names = ids
    .map((id) => tenantOptions.value.find((tenant) => String(tenant.id) === id)?.name)
    .filter((name): name is string => Boolean(name))
  return names.join(', ')
}

function tenantSelectLabel(ids: string[]) {
  if (ids.length === 0) return ''
  return tenantNames(ids) || `已选择 ${ids.length} 个租户`
}

async function loadTenantOptions(reset = false) {
  if (tenantOptionsLoading.value) return
  if (!reset && tenantOptions.value.length > 0 && !tenantOptionsHasMore.value) return
  if (reset) {
    tenantOptions.value = []
    tenantOptionsSkip.value = 0
    tenantOptionsTotal.value = 0
  }
  tenantOptionsLoading.value = true
  try {
    const result = await fetchTenants(tenantOptionsSkip.value, pickerPageSize)
    const items = result.items || []
    tenantOptions.value = reset ? items : [...tenantOptions.value, ...items]
    tenantOptionsSkip.value += items.length
    tenantOptionsTotal.value = result.total ?? tenantOptions.value.length
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '租户列表加载失败')
  } finally {
    tenantOptionsLoading.value = false
  }
}

function resolveTenantIdsFromNames(value: string | null | undefined) {
  const tokens = parseTenantNames(value)
  if (tokens.length === 0) return []
  return tenantOptions.value
    .filter((tenant) => tokens.includes(tenant.name) || tokens.includes(tenant.code) || tokens.includes(String(tenant.id)))
    .map((tenant) => String(tenant.id))
}

function openTenantPicker(target: 'create' | 'edit') {
  tenantPickerTarget.value = target
  tenantPickerDraftIds.value = [...(target === 'create' ? createTenantIds.value : editTenantIds.value)]
  tenantPickerKeyword.value = ''
  tenantPickerVisible.value = true
  void loadTenantOptions(true)
}

function closeTenantPicker() {
  tenantPickerVisible.value = false
}

function isTenantDraftSelected(id: number | string) {
  return tenantPickerDraftIds.value.includes(String(id))
}

function toggleTenantDraft(id: number | string) {
  const value = String(id)
  tenantPickerDraftIds.value = isTenantDraftSelected(value)
    ? tenantPickerDraftIds.value.filter((item) => item !== value)
    : [...tenantPickerDraftIds.value, value]
}

function removeTenant(target: 'create' | 'edit', id: string) {
  const current = target === 'create' ? createTenantIds : editTenantIds
  current.value = current.value.filter((item) => item !== id)
  if (target === 'create') {
    createDraft.value.visible_tenants = tenantNames(createTenantIds.value)
  } else {
    editDraft.value.visible_tenants = tenantNames(editTenantIds.value)
  }
}

function confirmTenantPicker() {
  if (tenantPickerTarget.value === 'create') {
    createTenantIds.value = [...tenantPickerDraftIds.value]
    createDraft.value.visible_tenants = tenantNames(createTenantIds.value)
  } else {
    editTenantIds.value = [...tenantPickerDraftIds.value]
    editDraft.value.visible_tenants = tenantNames(editTenantIds.value)
  }
  tenantPickerVisible.value = false
}

function handleTenantPickerScroll(event: Event) {
  const target = event.currentTarget as HTMLElement
  if (target.scrollTop + target.clientHeight >= target.scrollHeight - 24) {
    void loadTenantOptions()
  }
}

function deploymentModeLabel(value: string | null | undefined) {
  return deploymentModeOptions.find((item) => item.value === (value || 'MERGED'))?.label || value || '合并部署'
}

function communicationModesLabel(value: string | null | undefined) {
  const keys = String(value || '')
    .split(',')
    .map((item) => item.trim())
    .filter(Boolean)
  if (keys.length === 0) return '内部调用'
  return keys.map((key) => communicationModeOptions.find((item) => item.key === key)?.label || key).join(' / ')
}

function visibilityScopeFromMode(mode: string | null | undefined) {
  return mode === 'ALL_TENANTS' ? 'GLOBAL' : 'TENANT'
}

function draftClients(items: Array<{ key: string; label: string; checked: boolean }>): AppCenterClient[] {
  return items
    .filter((item) => item.checked)
    .map((item, index) => ({
      client_code: item.key,
      client_name: item.label,
      enabled: true,
      sort_order: index + 1,
    }))
}

function normalizeUpdatePayload(value: AppCenterUpdatePayload): AppCenterUpdatePayload {
  return {
    ...value,
    app_name: value.app_name.trim(),
    icon: trimNullable(value.icon),
    owner: trimNullable(value.owner),
    owner_user_ids: trimNullable(value.owner_user_ids),
    version: trimNullable(value.version),
    description: trimNullable(value.description),
    detail_description: trimNullable(value.detail_description),
    deployment_mode: value.deployment_mode || 'MERGED',
    communication_modes: value.deployment_mode === 'STANDALONE' ? trimNullable(value.communication_modes) : null,
    visibility_mode: trimNullable(value.visibility_mode),
    visible_tenants: trimNullable(value.visible_tenants),
    open_method: trimNullable(value.open_method),
    trial_policy: trimNullable(value.trial_policy),
    trial_start_rule: trimNullable(value.trial_start_rule),
    asset_config: trimNullable(value.asset_config),
    doc_config: trimNullable(value.doc_config),
    release_channel: trimNullable(value.release_channel),
    release_note: trimNullable(value.release_note),
    sort_order: Number(value.sort_order || 0),
  }
}

function openCreateDialog() {
  createForm.value = newCreateForm()
  createOwnerIds.value = []
  createTenantIds.value = []
  createDraft.value = newCreateDraft()
  createMode.value = 'manual'
  manifestFileName.value = ''
  manifestImportSummary.value = ''
  createStep.value = 'basic'
  createDialogVisible.value = true
  void loadOwnerUsers()
  void loadTenantOptions()
  nextTick(() => {
    createScrollRef.value?.scrollTo({ top: 0 })
  })
}

async function setCreateStep(step: string) {
  createStep.value = step
  await nextTick()
  const target = document.getElementById(`create-section-${step}`)
  if (!target || !createScrollRef.value) return
  createScrollRef.value.scrollTo({
    top: target.offsetTop - createScrollRef.value.offsetTop,
    behavior: 'smooth',
  })
}

function handleCreateScroll() {
  createBodyScrolling.value = true
  if (createBodyScrollTimer) window.clearTimeout(createBodyScrollTimer)
  createBodyScrollTimer = window.setTimeout(() => {
    createBodyScrolling.value = false
  }, 700)
  const container = createScrollRef.value
  if (!container) return
  let current = createSteps[0]?.key || 'basic'
  createSteps.forEach((step) => {
    const target = document.getElementById(`create-section-${step.key}`)
    if (target && target.offsetTop - container.offsetTop - container.scrollTop <= 28) {
      current = step.key
    }
  })
  createStep.value = current
}

async function setEditStep(step: string) {
  editStep.value = step
  await nextTick()
  const target = document.getElementById(`edit-section-${step}`)
  if (!target || !editScrollRef.value) return
  editScrollRef.value.scrollTo({
    top: target.offsetTop - editScrollRef.value.offsetTop,
    behavior: 'smooth',
  })
}

function handleEditScroll() {
  editBodyScrolling.value = true
  if (editBodyScrollTimer) window.clearTimeout(editBodyScrollTimer)
  editBodyScrollTimer = window.setTimeout(() => {
    editBodyScrolling.value = false
  }, 700)
  const container = editScrollRef.value
  if (!container) return
  let current = createSteps[0]?.key || 'basic'
  createSteps.forEach((step) => {
    const target = document.getElementById(`edit-section-${step.key}`)
    if (target && target.offsetTop - container.offsetTop - container.scrollTop <= 28) {
      current = step.key
    }
  })
  editStep.value = current
}

async function saveCreateDialog() {
  const chargeMode = chargeModeForSubmit(createForm.value.charge_mode, createDraft.value)
  const trialPolicy = trialPolicyForSubmit(chargeMode, createDraft.value)
  const payload: AppCenterCreatePayload = {
    ...createForm.value,
    app_code: createForm.value.app_code.trim(),
    app_name: createForm.value.app_name.trim(),
    icon: trimNullable(createForm.value.icon),
    owner: trimNullable(ownerNames(createOwnerIds.value)),
    owner_user_ids: trimNullable(createOwnerIds.value.join(',')),
    version: trimNullable(createForm.value.version),
    description: trimNullable(createForm.value.description),
    detail_description: trimNullable(createForm.value.detail_description),
    deployment_mode: createForm.value.deployment_mode,
    charge_mode: chargeMode,
    communication_modes: createForm.value.deployment_mode === 'STANDALONE' ? trimNullable(checkedKeys(createDraft.value.communication_modes)) : null,
    visibility_scope: visibilityScopeFromMode(createDraft.value.visibility_mode),
    visibility_mode: trimNullable(createDraft.value.visibility_mode),
    visible_tenants: createDraft.value.visibility_mode === 'SPECIFIED_TENANTS' ? trimNullable(tenantNames(createTenantIds.value) || createDraft.value.visible_tenants) : null,
    open_method: null,
    trial_policy: trimNullable(trialPolicy),
    trial_start_rule: trialStartRuleFor(trialPolicy),
    asset_config: null,
    doc_config: null,
    release_channel: null,
    release_note: null,
    sort_order: Number(createForm.value.sort_order || 0),
    clients: draftClients(createDraft.value.clients),
  }
  if (!payload.app_code || !payload.app_name) {
    ElMessage.error('请填写应用编码和应用名称')
    return
  }
  if (!payload.app_type || !payload.deployment_mode) {
    ElMessage.error('请选择应用类型和部署方式')
    return
  }
  if (!createDraft.value.visibility_mode) {
    ElMessage.error('请选择可见范围')
    return
  }
  if (!createDraft.value.commercial_kind) {
    ElMessage.error('请选择免费、收费或非售卖')
    return
  }
  if (createDraft.value.commercial_kind === 'PAID' && (!createDraft.value.paid_charge_mode || !resolvedTrialPolicy(createDraft.value))) {
    ElMessage.error('请选择收费模式和试用策略')
    return
  }
  try {
    await ElMessageBox.confirm(
      `确认创建应用主档「${payload.app_name}」？创建后将生成应用身份，并保存当前客户端、可见范围和收费策略。`,
      '创建应用主档',
      {
        confirmButtonText: '确认创建',
        cancelButtonText: '再检查一下',
        type: 'warning',
      },
    )
  } catch {
    return
  }
  saving.value = true
  try {
    const created = await createAppCenterApp(payload)
    ElMessage.success('应用已创建')
    createDialogVisible.value = false
    await loadApps()
    await openDetailDialog(created)
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '创建应用失败')
  } finally {
    saving.value = false
  }
}

function openEditDialog(app: AppCenterApp) {
  editAppId.value = app.id
  editAppSnapshot.value = app
  editForm.value = appToEditForm(app)
  editOwnerIds.value = parseOwnerIDs(app.owner_user_ids)
  editDraft.value = newEditDraft(app)
  editTenantIds.value = resolveTenantIdsFromNames(app.visible_tenants)
  editStep.value = 'basic'
  editDialogVisible.value = true
  void loadOwnerUsers()
  void loadTenantOptions().then(() => {
    editTenantIds.value = resolveTenantIdsFromNames(app.visible_tenants)
  })
  nextTick(() => {
    editScrollRef.value?.scrollTo({ top: 0 })
  })
}

async function saveEditDialog() {
  if (!editAppId.value) return
  const chargeMode = chargeModeForSubmit(editForm.value.charge_mode, editDraft.value)
  const trialPolicy = trialPolicyForSubmit(chargeMode, editDraft.value)
  const payload = normalizeUpdatePayload({
    ...editForm.value,
    charge_mode: chargeMode,
    owner: ownerNames(editOwnerIds.value),
    owner_user_ids: editOwnerIds.value.join(','),
    communication_modes: editForm.value.deployment_mode === 'STANDALONE' ? checkedKeys(editDraft.value.communication_modes) : null,
    visibility_scope: visibilityScopeFromMode(editDraft.value.visibility_mode),
    visibility_mode: editDraft.value.visibility_mode,
    visible_tenants: editDraft.value.visibility_mode === 'SPECIFIED_TENANTS' ? tenantNames(editTenantIds.value) || editDraft.value.visible_tenants : null,
    open_method: null,
    trial_policy: trialPolicy,
    trial_start_rule: trialStartRuleFor(trialPolicy),
    clients: draftClients(editDraft.value.clients),
  })
  if (!payload.app_name) {
    ElMessage.error('请填写应用名称')
    return
  }
  if (!editDraft.value.visibility_mode || !editDraft.value.commercial_kind) {
    ElMessage.error('请选择可见范围和收费策略')
    return
  }
  if (editDraft.value.commercial_kind === 'PAID' && (!editDraft.value.paid_charge_mode || !resolvedTrialPolicy(editDraft.value))) {
    ElMessage.error('请选择收费模式和试用策略')
    return
  }
  saving.value = true
  try {
    const updated = await updateAppCenterApp(editAppId.value, payload)
    ElMessage.success('应用已更新')
    editDialogVisible.value = false
    await loadApps()
    editAppSnapshot.value = updated
    detailApp.value = updated
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '更新应用失败')
  } finally {
    saving.value = false
  }
}

async function toggleAppStatus(app: AppCenterApp) {
  const nextStatus = app.status === 'DISABLED' ? 'ONLINE' : 'DISABLED'
  const actionText = nextStatus === 'DISABLED' ? '停用' : '启用'
  try {
    await ElMessageBox.confirm(`确认${actionText}「${app.app_name}」？`, `${actionText}应用`, {
      confirmButtonText: actionText,
      cancelButtonText: '取消',
      type: nextStatus === 'DISABLED' ? 'warning' : 'info',
    })
    const updated = await updateAppCenterAppStatus(app.id, nextStatus)
    ElMessage.success(`应用已${actionText}`)
    await loadApps()
    if (detailApp.value?.id === updated.id) detailApp.value = updated
  } catch (error) {
    if (error === 'cancel' || error === 'close') return
    ElMessage.error(error instanceof Error ? error.message : `${actionText}应用失败`)
  }
}

async function openDetailDialog(app: AppCenterApp) {
  detailDialogVisible.value = true
  detailLoading.value = true
  detailApp.value = null
  try {
    detailApp.value = await fetchAppCenterApp(app.id)
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '读取应用详情失败')
    detailDialogVisible.value = false
  } finally {
    detailLoading.value = false
  }
}

async function loadDictionaries() {
  const codes = Object.keys(dicts.value) as Array<keyof typeof dicts.value>
  const results = await Promise.allSettled(codes.map((code) => fetchDictItemsByCode(code)))
  results.forEach((result, index) => {
    if (result.status === 'fulfilled') {
      dicts.value[codes[index]] = result.value.items.filter((item) => item.enabled !== false)
    }
  })
}

async function loadApps() {
  loading.value = true
  try {
    const [list, summary] = await Promise.all([
      fetchAppCenterApps({
        skip: 0,
        limit: 100,
        keyword: keyword.value,
        type: typeFilter.value,
        status: statusFilter.value,
        source: sourceFilter.value,
      }),
      fetchAppCenterStats(),
    ])
    apps.value = list.items
    total.value = list.total
    stats.value = summary
  } finally {
    loading.value = false
  }
}

function resetFilters() {
  keyword.value = ''
  typeFilter.value = ''
  statusFilter.value = ''
  sourceFilter.value = ''
  void loadApps()
}

function applyStatusFilter(status: string) {
  statusFilter.value = status
  void loadApps()
}

onMounted(async () => {
  activeSection.value = sectionFromPath(route.path)
  await loadDictionaries()
  await loadApps()
})

onUnmounted(() => {
  if (createBodyScrollTimer) window.clearTimeout(createBodyScrollTimer)
  if (editBodyScrollTimer) window.clearTimeout(editBodyScrollTimer)
  if (ownerSearchTimer) window.clearTimeout(ownerSearchTimer)
  if (tenantSearchTimer) window.clearTimeout(tenantSearchTimer)
})

watch(
  () => route.path,
  (path) => {
    activeSection.value = sectionFromPath(path)
  },
)

watch(ownerPickerKeyword, () => {
  if (!ownerPickerVisible.value) return
  if (ownerSearchTimer) window.clearTimeout(ownerSearchTimer)
  ownerSearchTimer = window.setTimeout(() => {
    void loadOwnerUsers(true)
  }, 220)
})

watch(tenantPickerKeyword, () => {
  if (!tenantPickerVisible.value) return
  if (tenantSearchTimer) window.clearTimeout(tenantSearchTimer)
  tenantSearchTimer = window.setTimeout(() => {
    void loadTenantOptions(true)
  }, 220)
})
</script>

<template>
  <NeuroAgentPageShell class="app-center-page" :show-hero="activeSection === 'apps'">
    <template #title>应用中心</template>
    <template #subtitle>
      统一注册、装载、订阅、安装、授权、计费和运行治理；版本、进度与文档由项目管理应用同步。
    </template>
    <template #actions>
      <button
        v-for="card in statCards"
        :key="card.label"
        class="app-stat"
        :class="[`app-stat--${card.tone}`, { 'app-stat--active': statusFilter === card.status }]"
        type="button"
        @click="applyStatusFilter(card.status)"
      >
        <span>{{ card.label }}</span>
        <strong>{{ card.value }}</strong>
        <em>{{ card.hint }}</em>
      </button>
    </template>

    <div class="app-center-layout">
      <section class="app-center-main">
        <template v-if="activeSection === 'apps'">
          <div class="app-center-toolbar">
            <div class="app-center-toolbar__filters">
              <div class="app-center-search">
                <el-icon :size="16"><Search /></el-icon>
                <input v-model="keyword" type="text" placeholder="搜索应用编码、名称或说明" @keyup.enter="loadApps" />
              </div>
              <select v-model="typeFilter" aria-label="应用类型">
                <option value="">全部类型</option>
                <option v-for="item in appTypeOptions" :key="item.value" :value="item.value">
                  {{ item.label }}
                </option>
              </select>
              <select v-model="statusSelectValue" aria-label="应用状态">
                <option value="">全部状态</option>
                <option v-for="item in statusFilterOptions" :key="item.value" :value="item.value">
                  {{ item.label }}
                </option>
              </select>
              <select v-model="sourceFilter" aria-label="应用来源">
                <option value="">全部来源</option>
                <option v-for="item in dictOptions('app_source')" :key="item.value" :value="item.value">
                  {{ item.label }}
                </option>
              </select>
              <button class="app-btn app-btn--primary" type="button" @click="loadApps">
                查询
              </button>
              <button v-if="hasActiveFilters" class="app-btn" type="button" @click="resetFilters">
                <el-icon :size="15"><Refresh /></el-icon>
                重置
              </button>
            </div>
            <div class="app-center-toolbar__actions">
              <button class="app-btn" type="button" @click="downloadManifestTemplate">
                <el-icon :size="15"><Download /></el-icon>
                Manifest 格式文件
              </button>
              <button class="app-btn app-btn--primary" type="button" @click="openCreateDialog">
                <el-icon :size="15"><Plus /></el-icon>
                新增应用
              </button>
            </div>
          </div>

          <div v-if="loading" class="app-center-empty">加载中...</div>
          <div v-else-if="apps.length === 0" class="app-center-empty">暂无应用</div>
          <div v-else class="app-type-list">
            <section v-for="group in groupedApps" :key="group.type" class="app-type-group">
              <div class="app-type-group__head">
                <div>
                  <strong>{{ group.label }}</strong>
                  <span>{{ group.type }}</span>
                </div>
                <em>{{ group.items.length }}</em>
              </div>
              <div class="app-grid">
                <article
                  v-for="app in group.items"
                  :key="app.app_code"
                  class="app-card"
                  :class="{ 'app-card--disabled': app.status === 'DISABLED' || app.status === 'ARCHIVED' }"
                >
                  <div class="app-card__head">
                    <div class="app-card__icon">
                      <span>{{ appInitial(app) }}</span>
                    </div>
                    <div class="app-card__title">
                      <h3>{{ app.app_name }}</h3>
                      <code>{{ app.app_code }}</code>
                    </div>
                    <span class="app-status" :class="statusClass(app.status)">
                      {{ labelOf('app_status', app.status) }}
                    </span>
                  </div>
                  <div class="app-card__badges">
                    <span>{{ labelOf('app_source', app.source) }}</span>
                    <span>{{ labelOf('app_charge_mode', app.charge_mode) }}</span>
                  </div>
                  <p class="app-card__desc">{{ app.description || '暂无说明' }}</p>
                  <div class="app-card__metrics">
                    <div>
                      <span>负责人</span>
                      <strong>{{ app.owner || '—' }}</strong>
                    </div>
                    <div>
                      <span>当前版本</span>
                      <strong>{{ app.version || '—' }}</strong>
                    </div>
                    <div>
                      <span>来源</span>
                      <strong>{{ labelOf('app_source', app.source) }}</strong>
                    </div>
                    <div>
                      <span>可见范围</span>
                      <strong>{{ labelOf('app_visibility_scope', app.visibility_scope) }}</strong>
                    </div>
                  </div>
                  <div class="app-card__channels">
                    <span>PC Web</span>
                  </div>
                  <div class="app-card__actions">
                    <button class="app-card__detail app-card__detail--primary" type="button" @click="openDetailDialog(app)">
                      进入详情
                    </button>
                    <button class="app-card__detail" type="button" @click="openEditDialog(app)">编辑</button>
                    <button class="app-card__detail" type="button" :disabled="app.is_builtin" @click="toggleAppStatus(app)">
                      {{ app.status === 'DISABLED' ? '启用' : '停用' }}
                    </button>
                    <button class="app-card__detail" type="button" disabled>安装记录</button>
                    <button class="app-card__detail" type="button" disabled>邀请体验</button>
                  </div>
                </article>
              </div>
            </section>
          </div>

        </template>

        <div v-else class="app-center-production-card">
          <strong>{{ activeDirectoryItem.count }}</strong>
          <span>{{ activeDirectoryItem.label }}</span>
          <p>{{ activeDirectoryItem.description }}本页已接入生产统计数据，详细管理视图将在对应模块落地后展开。</p>
        </div>
      </section>
    </div>

    <NeuroAgentDialog
      v-model="createDialogVisible"
      title="新增应用"
      icon="📦"
      size="large"
      width="1040px"
      height="86vh"
      :loading="saving"
      :confirm-disabled="saving"
      confirm-text="创建应用主档"
      @confirm="saveCreateDialog"
    >
      <div class="app-create-wizard">
        <aside class="app-create-steps">
          <button
            v-for="(step, index) in createSteps"
            :key="step.key"
            class="app-create-step"
            :class="{ 'is-active': createStep === step.key, 'is-complete': createStepCompletion[step.key] }"
            type="button"
            @click="setCreateStep(step.key)"
          >
            <i>{{ createStepCompletion[step.key] ? '✓' : index + 1 }}</i>
            <span>
              <strong>{{ step.title }}</strong>
              <em>{{ step.hint }}</em>
            </span>
          </button>
        </aside>

        <section ref="createScrollRef" class="app-create-body" :class="{ 'is-scrolling': createBodyScrolling }" @scroll="handleCreateScroll">
          <div class="app-create-mode-panel">
            <div class="app-create-mode-panel__head">
              <strong>创建方式</strong>
              <span>{{ createMode === 'manifest' ? '已从 Manifest 回填，可继续修改。' : '手工填写应用主档。' }}</span>
            </div>
            <div class="app-create-mode-switch">
              <button type="button" :class="{ 'is-active': createMode === 'manual' }" @click="createMode = 'manual'">
                手工填写
              </button>
              <button type="button" :class="{ 'is-active': createMode === 'manifest' }" @click="triggerManifestImport">
                <el-icon :size="15"><Upload /></el-icon>
                选择 Manifest 文件解析
              </button>
            </div>
            <input ref="manifestFileInputRef" class="app-manifest-input" type="file" accept=".json,.yaml,.yml,application/json,application/x-yaml,text/yaml" @change="handleManifestFileChange" />
            <p v-if="manifestFileName" class="app-create-mode-panel__result">
              <strong>{{ manifestFileName }}</strong>
              <span>{{ manifestImportSummary }}</span>
            </p>
          </div>

          <div id="create-section-basic" class="app-create-panel">
            <div class="app-create-panel__head">
              <div>
                <h3>基础信息</h3>
                <p>创建应用主档，明确应用长期身份。应用编码创建后不建议修改。</p>
              </div>
              <span>必填</span>
            </div>
            <div class="app-form">
              <label>
                <span>应用编码</span>
                <input v-model="createForm.app_code" placeholder="只允许小写字母、数字、中横线" />
              </label>
              <label>
                <span>应用名称</span>
                <input v-model="createForm.app_name" placeholder="请输入应用名称" />
              </label>
              <label class="app-form__full">
                <span>一句话介绍</span>
                <input v-model="createForm.description" maxlength="120" placeholder="用于应用列表卡片，例如：统一管理客户、商机与跟进任务" />
              </label>
              <label>
                <span>应用类型</span>
                <select v-model="createForm.app_type">
                  <option value="">请选择应用类型</option>
                  <option v-for="item in appTypeOptions" :key="item.value" :value="item.value">
                    {{ item.label }}
                  </option>
                </select>
              </label>
              <label>
                <span>部署方式</span>
                <select v-model="createForm.deployment_mode">
                  <option value="">请选择部署方式</option>
                  <option v-for="item in deploymentModeOptions" :key="item.value" :value="item.value">
                    {{ item.label }}
                  </option>
                </select>
              </label>
              <label>
                <span>负责人</span>
                <div class="app-owner-select" role="button" tabindex="0" @click="openOwnerPicker('create')" @keydown.enter.prevent="openOwnerPicker('create')">
                  <span v-if="createOwnerIds.length" class="app-owner-select__chips">
                    <button v-for="id in createOwnerIds" :key="id" type="button" @click.stop="removeOwner('create', id)">
                      {{ ownerSelectLabel([id]) || id }} ×
                    </button>
                  </span>
                  <span v-else class="is-placeholder">选择成员，可多选</span>
                  <em>{{ ownerUsersLoading ? '加载中' : '选择' }}</em>
                </div>
              </label>
              <label class="app-form__full">
                <span>应用详细介绍</span>
                <textarea v-model="createForm.detail_description" class="app-form__rich" placeholder="可填写较完整的应用背景、边界、核心能力和使用说明，支持后续接入富文本内容"></textarea>
              </label>
            </div>
          </div>

          <div id="create-section-visibility" class="app-create-panel">
            <div class="app-create-panel__head">
              <div>
                <h3>可见范围</h3>
                <p>只决定哪些租户能看到并申请开通该应用，不承载收费和试用规则。</p>
              </div>
            </div>
            <div class="app-policy-layout">
              <section class="app-policy-block">
                <header>
                  <strong>可见范围</strong>
                  <span>控制租户是否能看到并申请开通该应用。</span>
                </header>
                <div class="app-choice-grid">
                  <button type="button" :class="{ 'is-active': createDraft.visibility_mode === 'ALL_TENANTS' }" @click="createDraft.visibility_mode = 'ALL_TENANTS'">
                    <strong>全部租户</strong>
                    <span>所有租户都可见，仍需按套餐或授权开通。</span>
                  </button>
                  <button type="button" :class="{ 'is-active': createDraft.visibility_mode === 'SPECIFIED_TENANTS' }" @click="createDraft.visibility_mode = 'SPECIFIED_TENANTS'">
                    <strong>指定租户</strong>
                    <span>只对指定租户可见，适合试点、定制或灰度。</span>
                  </button>
                </div>
                <label v-if="createDraft.visibility_mode === 'SPECIFIED_TENANTS'">
                  <span>指定租户</span>
                  <div class="app-owner-select" role="button" tabindex="0" @click="openTenantPicker('create')" @keydown.enter.prevent="openTenantPicker('create')">
                    <span v-if="createTenantIds.length" class="app-owner-select__chips">
                      <button v-for="id in createTenantIds" :key="id" type="button" @click.stop="removeTenant('create', id)">
                        {{ tenantSelectLabel([id]) || id }} ×
                      </button>
                    </span>
                    <span v-else class="is-placeholder">选择租户，可多选</span>
                    <em>{{ tenantOptionsLoading ? '加载中' : '选择' }}</em>
                  </div>
                </label>
              </section>
            </div>
          </div>

          <div id="create-section-commercial" class="app-create-panel">
            <div class="app-create-panel__head">
              <div>
                <h3>收费策略</h3>
                <p>先区分免费、收费和非售卖；收费应用再声明收费模式和试用规则。</p>
              </div>
            </div>
            <div class="app-policy-layout">

              <section class="app-policy-block">
                <header>
                  <strong>收费类型</strong>
                  <span>先判断应用是否售卖。免费和非售卖不配置试用期。</span>
                </header>
                <div class="app-choice-grid app-choice-grid--three">
                  <button
                    v-for="item in commercialKindOptions"
                    :key="item.value"
                    type="button"
                    :class="{ 'is-active': createDraft.commercial_kind === item.value }"
                    @click="chooseCreateCommercialKind(item.value)"
                  >
                    <strong>{{ item.label }}</strong>
                    <span>{{ item.description }}</span>
                  </button>
                </div>
              </section>

              <section v-if="createDraft.commercial_kind === 'PAID'" class="app-policy-block">
                <header>
                  <strong>收费模式</strong>
                  <span>收费应用继续声明计费方式，价格、周期和计量项后续在套餐或计费配置中维护。</span>
                </header>
                <div class="app-choice-grid app-choice-grid--three">
                  <button
                    v-for="item in paidChargeModeOptions"
                    :key="item.value"
                    type="button"
                    :class="{ 'is-active': createDraft.paid_charge_mode === item.value }"
                    @click="createDraft.paid_charge_mode = item.value"
                  >
                    <strong>{{ item.label }}</strong>
                    <span>{{ item.description }}</span>
                  </button>
                </div>
              </section>

              <section v-if="createDraft.commercial_kind === 'PAID'" class="app-policy-block">
                <span class="app-policy-alert" tabindex="0" aria-label="试用开始时间说明">
                  !
                  <em>试用开始：租户获得并开通应用时开始</em>
                </span>
                <header>
                  <strong>是否允许试用</strong>
                  <span>收费应用可选择不试用，也可声明试用周期。</span>
                </header>
                <div class="app-policy-rules">
                  <div v-for="rule in trialTransitionRules" :key="rule.title">
                    <strong>{{ rule.title }}</strong>
                    <span>{{ rule.content }}</span>
                  </div>
                </div>
                <div class="app-form app-form--embedded app-form--policy">
                  <div class="app-form__full">
                    <span class="app-field-title">使用策略</span>
                    <div class="app-choice-grid app-choice-grid--policy">
                      <button
                        v-for="item in trialPolicyOptions"
                        :key="item.value"
                        type="button"
                        class="app-choice-card"
                        :class="{ 'is-active': createDraft.trial_days === item.value, 'app-choice-card--custom': item.value === customTrialPolicyValue }"
                        @click="createDraft.trial_days = item.value"
                      >
                        <strong>{{ item.label }}</strong>
                        <input
                          v-if="item.value === customTrialPolicyValue && createDraft.trial_days === customTrialPolicyValue"
                          v-model="createDraft.trial_custom_time"
                          placeholder="45 天 / 3 个月"
                          @click.stop
                        />
                      </button>
                    </div>
                  </div>
                </div>
              </section>
            </div>
          </div>

          <div v-if="createForm.deployment_mode === 'STANDALONE'" class="app-create-panel">
            <div class="app-create-panel__head">
              <div>
                <h3>独立部署通讯</h3>
                <p>仅独立部署应用需要登记和底座之间的通讯方式。</p>
              </div>
            </div>
            <div class="app-create-checks app-create-checks--compact">
              <label v-for="mode in createDraft.communication_modes" :key="mode.key">
                <input v-model="mode.checked" type="checkbox" />
                <span>{{ mode.label }}</span>
              </label>
            </div>
          </div>

          <div id="create-section-clients" class="app-create-panel">
            <div class="app-create-panel__head">
              <div>
                <h3>客户端配置</h3>
                <p>声明应用支持的客户端形态，保存后进入应用客户端表，并作为后续租户权益联动依据。</p>
              </div>
            </div>
            <div class="app-create-checks">
              <label v-for="client in createDraft.clients" :key="client.key">
                <input v-model="client.checked" type="checkbox" />
                <span>{{ client.label }}</span>
              </label>
            </div>
          </div>

        </section>
      </div>
    </NeuroAgentDialog>

    <NeuroAgentDialog
      v-model="editDialogVisible"
      title="编辑应用"
      icon="📦"
      size="large"
      width="1040px"
      height="86vh"
      :loading="saving"
      :confirm-disabled="saving"
      confirm-text="保存"
      @confirm="saveEditDialog"
    >
      <div class="app-create-wizard">
        <aside class="app-create-steps">
          <button
            v-for="(step, index) in createSteps"
            :key="step.key"
            class="app-create-step"
            :class="{ 'is-active': editStep === step.key, 'is-complete': editStepCompletion[step.key] }"
            type="button"
            @click="setEditStep(step.key)"
          >
            <i>{{ editStepCompletion[step.key] ? '✓' : index + 1 }}</i>
            <span>
              <strong>{{ step.title }}</strong>
              <em>{{ step.hint }}</em>
            </span>
          </button>
        </aside>

        <section ref="editScrollRef" class="app-create-body" :class="{ 'is-scrolling': editBodyScrolling }" @scroll="handleEditScroll">
          <div id="edit-section-basic" class="app-create-panel">
            <div class="app-create-panel__head">
              <div>
                <h3>基础信息</h3>
                <p>维护应用主档。应用编码由注册或装载阶段生成，编辑时仅展示不修改。</p>
              </div>
              <span>保存到主档</span>
            </div>
            <div class="app-form">
              <label>
                <span>应用编码</span>
                <input :value="editAppSnapshot?.app_code || '—'" disabled />
              </label>
              <label>
                <span>应用名称</span>
                <input v-model="editForm.app_name" placeholder="请输入应用名称" />
              </label>
              <label class="app-form__full">
                <span>一句话介绍</span>
                <input v-model="editForm.description" maxlength="120" placeholder="用于应用列表卡片，例如：统一管理客户、商机与跟进任务" />
              </label>
              <label>
                <span>应用类型</span>
                <select v-model="editForm.app_type">
                  <option v-for="item in appTypeOptions" :key="item.value" :value="item.value">
                    {{ item.label }}
                  </option>
                </select>
              </label>
              <label>
                <span>部署方式</span>
                <select v-model="editForm.deployment_mode">
                  <option v-for="item in deploymentModeOptions" :key="item.value" :value="item.value">
                    {{ item.label }}
                  </option>
                </select>
              </label>
              <label>
                <span>负责人</span>
                <div class="app-owner-select" role="button" tabindex="0" @click="openOwnerPicker('edit')" @keydown.enter.prevent="openOwnerPicker('edit')">
                  <span v-if="editOwnerIds.length" class="app-owner-select__chips">
                    <button v-for="id in editOwnerIds" :key="id" type="button" @click.stop="removeOwner('edit', id)">
                      {{ ownerSelectLabel([id]) || id }} ×
                    </button>
                  </span>
                  <span v-else class="is-placeholder">选择成员，可多选</span>
                  <em>{{ ownerUsersLoading ? '加载中' : '选择' }}</em>
                </div>
              </label>
              <label>
                <span>图标</span>
                <input v-model="editForm.icon" placeholder="图标名或标识，选填" />
              </label>
              <label>
                <span>来源</span>
                <input :value="editAppSnapshot ? labelOf('app_source', editAppSnapshot.source) : '—'" disabled />
              </label>
              <label class="app-form__full">
                <span>应用详细介绍</span>
                <textarea v-model="editForm.detail_description" class="app-form__rich" placeholder="可填写较完整的应用背景、边界、核心能力和使用说明，支持后续接入富文本内容"></textarea>
              </label>
            </div>
          </div>

          <div id="edit-section-visibility" class="app-create-panel">
            <div class="app-create-panel__head">
              <div>
                <h3>可见范围</h3>
                <p>只维护哪些租户能看到并申请开通该应用，不承载收费和试用规则。</p>
              </div>
            </div>
            <div class="app-policy-layout">
              <section class="app-policy-block">
                <header>
                  <strong>可见范围</strong>
                  <span>控制租户是否能看到并申请开通该应用。</span>
                </header>
                <div class="app-choice-grid">
                  <button type="button" :class="{ 'is-active': editDraft.visibility_mode === 'ALL_TENANTS' }" @click="editDraft.visibility_mode = 'ALL_TENANTS'">
                    <strong>全部租户</strong>
                    <span>所有租户都可见，仍需按套餐或授权开通。</span>
                  </button>
                  <button type="button" :class="{ 'is-active': editDraft.visibility_mode === 'SPECIFIED_TENANTS' }" @click="editDraft.visibility_mode = 'SPECIFIED_TENANTS'">
                    <strong>指定租户</strong>
                    <span>只对指定租户可见，适合试点、定制或灰度。</span>
                  </button>
                </div>
                <label v-if="editDraft.visibility_mode === 'SPECIFIED_TENANTS'">
                  <span>指定租户</span>
                  <div class="app-owner-select" role="button" tabindex="0" @click="openTenantPicker('edit')" @keydown.enter.prevent="openTenantPicker('edit')">
                    <span v-if="editTenantIds.length" class="app-owner-select__chips">
                      <button v-for="id in editTenantIds" :key="id" type="button" @click.stop="removeTenant('edit', id)">
                        {{ tenantSelectLabel([id]) || id }} ×
                      </button>
                    </span>
                    <span v-else class="is-placeholder">选择租户，可多选</span>
                    <em>{{ tenantOptionsLoading ? '加载中' : '选择' }}</em>
                  </div>
                </label>
              </section>
            </div>
          </div>

          <div id="edit-section-commercial" class="app-create-panel">
            <div class="app-create-panel__head">
              <div>
                <h3>收费策略</h3>
                <p>先区分免费、收费和非售卖；收费应用再声明收费模式和试用规则。</p>
              </div>
            </div>
            <div class="app-policy-layout">

              <section class="app-policy-block">
                <header>
                  <strong>收费类型</strong>
                  <span>先判断应用是否售卖。免费和非售卖不配置试用期。</span>
                </header>
                <div class="app-choice-grid app-choice-grid--three">
                  <button
                    v-for="item in commercialKindOptions"
                    :key="item.value"
                    type="button"
                    :class="{ 'is-active': editDraft.commercial_kind === item.value }"
                    @click="chooseEditCommercialKind(item.value)"
                  >
                    <strong>{{ item.label }}</strong>
                    <span>{{ item.description }}</span>
                  </button>
                </div>
              </section>

              <section v-if="editDraft.commercial_kind === 'PAID'" class="app-policy-block">
                <header>
                  <strong>收费模式</strong>
                  <span>收费应用继续声明计费方式，价格、周期和计量项后续在套餐或计费配置中维护。</span>
                </header>
                <div class="app-choice-grid app-choice-grid--three">
                  <button
                    v-for="item in paidChargeModeOptions"
                    :key="item.value"
                    type="button"
                    :class="{ 'is-active': editDraft.paid_charge_mode === item.value }"
                    @click="editDraft.paid_charge_mode = item.value"
                  >
                    <strong>{{ item.label }}</strong>
                    <span>{{ item.description }}</span>
                  </button>
                </div>
              </section>

              <section v-if="editDraft.commercial_kind === 'PAID'" class="app-policy-block">
                <span class="app-policy-alert" tabindex="0" aria-label="试用开始时间说明">
                  !
                  <em>试用开始：租户获得并开通应用时开始</em>
                </span>
                <header>
                  <strong>是否允许试用</strong>
                  <span>收费应用可选择不试用，也可声明试用周期。</span>
                </header>
                <div class="app-policy-rules">
                  <div v-for="rule in trialTransitionRules" :key="rule.title">
                    <strong>{{ rule.title }}</strong>
                    <span>{{ rule.content }}</span>
                  </div>
                </div>
                <div class="app-form app-form--embedded app-form--policy">
                  <div class="app-form__full">
                    <span class="app-field-title">使用策略</span>
                    <div class="app-choice-grid app-choice-grid--policy">
                      <button
                        v-for="item in trialPolicyOptions"
                        :key="item.value"
                        type="button"
                        class="app-choice-card"
                        :class="{ 'is-active': editDraft.trial_days === item.value, 'app-choice-card--custom': item.value === customTrialPolicyValue }"
                        @click="editDraft.trial_days = item.value"
                      >
                        <strong>{{ item.label }}</strong>
                        <input
                          v-if="item.value === customTrialPolicyValue && editDraft.trial_days === customTrialPolicyValue"
                          v-model="editDraft.trial_custom_time"
                          placeholder="45 天 / 3 个月"
                          @click.stop
                        />
                      </button>
                    </div>
                  </div>
                </div>
              </section>
            </div>
          </div>

          <div v-if="editForm.deployment_mode === 'STANDALONE'" class="app-create-panel">
            <div class="app-create-panel__head">
              <div>
                <h3>独立部署通讯</h3>
                <p>仅独立部署应用需要登记和底座之间的通讯方式。</p>
              </div>
            </div>
            <div class="app-create-checks app-create-checks--compact">
              <label v-for="mode in editDraft.communication_modes" :key="mode.key">
                <input v-model="mode.checked" type="checkbox" />
                <span>{{ mode.label }}</span>
              </label>
            </div>
          </div>

          <div id="edit-section-clients" class="app-create-panel">
            <div class="app-create-panel__head">
              <div>
                <h3>客户端配置</h3>
                <p>维护应用支持的访问形态，保存后同步到应用客户端表，并作为后续租户权益联动依据。</p>
              </div>
            </div>
            <div class="app-create-checks">
              <label v-for="client in editDraft.clients" :key="client.key">
                <input v-model="client.checked" type="checkbox" />
                <span>{{ client.label }}</span>
              </label>
            </div>
          </div>

        </section>
      </div>
    </NeuroAgentDialog>

    <Teleport to="body">
      <div
        v-if="ownerPickerVisible"
        class="app-owner-picker-backdrop"
        @click.self="closeOwnerPicker"
        @wheel.prevent.stop
        @touchmove.prevent.stop
      >
        <section class="app-owner-picker" @click.stop @wheel.stop @touchmove.stop>
          <header class="app-owner-picker__head">
            <div>
              <h3>选择负责人</h3>
              <p>从成员管理中选择，可多选。</p>
            </div>
            <button type="button" aria-label="关闭" @click="closeOwnerPicker">×</button>
          </header>
          <div class="app-owner-picker__search">
            <Search class="app-owner-picker__search-icon" />
            <input v-model="ownerPickerKeyword" autofocus placeholder="搜索姓名、工号、手机号或邮箱" />
          </div>
          <div class="app-owner-picker__list" @scroll="handleOwnerPickerScroll">
            <button
              v-for="user in filteredOwnerUsers"
              :key="user.id"
              class="app-owner-picker__item"
              :class="{ 'is-selected': isOwnerDraftSelected(user.id) }"
              type="button"
              @click="toggleOwnerDraft(user.id)"
            >
              <i>{{ isOwnerDraftSelected(user.id) ? '✓' : '' }}</i>
              <span>
                <strong>{{ user.name }}</strong>
                <em>{{ user.employee_no }}{{ user.phone ? ` / ${user.phone}` : '' }}</em>
              </span>
            </button>
            <div v-if="ownerUsersLoading" class="app-owner-picker__empty">成员加载中...</div>
            <div v-else-if="filteredOwnerUsers.length === 0" class="app-owner-picker__empty">没有匹配的成员</div>
            <div v-else-if="ownerUsersHasMore" class="app-owner-picker__more">向下滚动加载更多</div>
          </div>
          <footer class="app-owner-picker__foot">
            <span>已选择 {{ ownerPickerDraftIds.length }} 人</span>
            <div>
              <button v-if="ownerPickerDraftIds.length" type="button" class="app-owner-picker__ghost" @click="ownerPickerDraftIds = []">清空</button>
              <button type="button" class="app-owner-picker__ghost" @click="closeOwnerPicker">取消</button>
              <button type="button" class="app-owner-picker__primary" @click="confirmOwnerPicker">确认选择</button>
            </div>
          </footer>
        </section>
      </div>
    </Teleport>

    <Teleport to="body">
      <div
        v-if="tenantPickerVisible"
        class="app-owner-picker-backdrop"
        @click.self="closeTenantPicker"
        @wheel.prevent.stop
        @touchmove.prevent.stop
      >
        <section class="app-owner-picker" @click.stop @wheel.stop @touchmove.stop>
          <header class="app-owner-picker__head">
            <div>
              <h3>选择租户</h3>
              <p>选择可见范围内的租户，可多选。</p>
            </div>
            <button type="button" aria-label="关闭" @click="closeTenantPicker">×</button>
          </header>
          <div class="app-owner-picker__search">
            <Search class="app-owner-picker__search-icon" />
            <input v-model="tenantPickerKeyword" autofocus placeholder="搜索租户名称、编码或联系人" />
          </div>
          <div class="app-owner-picker__list" @scroll="handleTenantPickerScroll">
            <button
              v-for="tenant in filteredTenantOptions"
              :key="tenant.id"
              class="app-owner-picker__item"
              :class="{ 'is-selected': isTenantDraftSelected(tenant.id) }"
              type="button"
              @click="toggleTenantDraft(tenant.id)"
            >
              <i>{{ isTenantDraftSelected(tenant.id) ? '✓' : '' }}</i>
              <span>
                <strong>{{ tenant.name }}</strong>
                <em>{{ tenant.code }}{{ tenant.plan_name ? ` / ${tenant.plan_name}` : '' }}</em>
              </span>
            </button>
            <div v-if="tenantOptionsLoading" class="app-owner-picker__empty">租户加载中...</div>
            <div v-else-if="filteredTenantOptions.length === 0" class="app-owner-picker__empty">没有匹配的租户</div>
            <div v-else-if="tenantOptionsHasMore" class="app-owner-picker__more">向下滚动加载更多</div>
          </div>
          <footer class="app-owner-picker__foot">
            <span>已选择 {{ tenantPickerDraftIds.length }} 个租户</span>
            <div>
              <button v-if="tenantPickerDraftIds.length" type="button" class="app-owner-picker__ghost" @click="tenantPickerDraftIds = []">清空</button>
              <button type="button" class="app-owner-picker__ghost" @click="closeTenantPicker">取消</button>
              <button type="button" class="app-owner-picker__primary" @click="confirmTenantPicker">确认选择</button>
            </div>
          </footer>
        </section>
      </div>
    </Teleport>

    <NeuroAgentDialog
      v-model="detailDialogVisible"
      title="应用详情"
      icon="📦"
      size="large"
      width="1040px"
      height="86vh"
      :loading="detailLoading"
      :show-confirm="false"
      cancel-text="关闭"
    >
      <div v-if="detailApp" class="app-detail">
        <section class="app-detail-hero">
          <div class="app-detail-hero__main">
            <div class="app-card__icon">
              <el-icon :size="24"><component :is="iconFor(detailApp)" /></el-icon>
            </div>
            <div>
              <h3>{{ detailApp.app_name }}</h3>
              <code>{{ detailApp.app_code }}</code>
              <p>{{ detailApp.description || '暂无一句话介绍，可在编辑应用中补充列表卡片说明。' }}</p>
            </div>
          </div>
          <div class="app-detail-hero__meta">
            <span class="app-status" :class="statusClass(detailApp.status)">
              {{ labelOf('app_status', detailApp.status) }}
            </span>
            <strong>{{ labelOf('app_type', detailApp.app_type) }}</strong>
            <span>{{ labelOf('app_source', detailApp.source) }}</span>
          </div>
        </section>

        <section class="app-detail-section">
          <div class="app-detail-section__head">
            <h4>应用主档</h4>
            <span>基础身份</span>
          </div>
          <div class="app-detail-fields">
            <div><span>应用编码</span><strong>{{ detailApp.app_code }}</strong></div>
            <div><span>应用名称</span><strong>{{ detailApp.app_name }}</strong></div>
            <div><span>应用类型</span><strong>{{ labelOf('app_type', detailApp.app_type) }}</strong></div>
            <div><span>部署方式</span><strong>{{ deploymentModeLabel(detailApp.deployment_mode) }}</strong></div>
            <div><span>通讯方式</span><strong>{{ communicationModesLabel(detailApp.communication_modes) }}</strong></div>
            <div><span>来源</span><strong>{{ labelOf('app_source', detailApp.source) }}</strong></div>
            <div><span>负责人</span><strong>{{ detailApp.owner || '—' }}</strong></div>
            <div><span>版本</span><strong>{{ detailApp.version || '—' }}</strong></div>
            <div><span>内置应用</span><strong>{{ detailApp.is_builtin ? '是' : '否' }}</strong></div>
            <div class="app-detail-fields__full">
              <span>应用详细介绍</span>
              <p class="app-detail-rich">{{ detailApp.detail_description || '—' }}</p>
            </div>
          </div>
        </section>

        <section class="app-detail-section">
          <div class="app-detail-section__head">
            <h4>可见、收费与试用</h4>
            <span>租户权益</span>
          </div>
          <div class="app-detail-fields">
            <div><span>当前可见范围</span><strong>{{ labelOf('app_visibility_scope', detailApp.visibility_scope) }}</strong></div>
            <div><span>V1 可见口径</span><strong>{{ detailApp.visibility_mode || (detailApp.visibility_scope === 'GLOBAL' ? '全部租户' : '指定租户 / 平台控制') }}</strong></div>
            <div><span>收费模型</span><strong>{{ labelOf('app_charge_mode', detailApp.charge_mode) }}</strong></div>
            <div><span>平台专属</span><strong>{{ detailApp.is_platform_only ? '是' : '否' }}</strong></div>
            <div><span>试用策略</span><strong>{{ trialPolicyLabel(detailApp.trial_policy) }}</strong></div>
            <div><span>试用开始</span><strong>{{ detailApp.trial_start_rule || (detailApp.trial_policy && detailApp.trial_policy !== '不支持试用' ? '获得并开通时开始' : '—') }}</strong></div>
          </div>
        </section>

        <section class="app-detail-section">
          <div class="app-detail-section__head">
            <h4>客户端</h4>
            <span>访问形态</span>
          </div>
          <div class="app-detail-tags">
            <span v-for="client in detailApp.clients || []" :key="client.client_code" :class="{ 'is-muted': client.enabled === false }">{{ client.client_name }}</span>
            <span v-if="!(detailApp.clients || []).length" class="is-muted">待配置客户端</span>
          </div>
        </section>

        <section class="app-detail-section">
          <div class="app-detail-section__head">
            <h4>入口、API 与权限</h4>
            <span>应用资产</span>
          </div>
          <div class="app-detail-assets">
            <div>
              <strong>入口</strong>
              <span>{{ detailApp.asset_config || '工作台入口、管理列表入口、详情入口' }}</span>
            </div>
            <div>
              <strong>API</strong>
              <span>查询 API、状态控制 API、主档维护 API</span>
            </div>
            <div>
              <strong>权限</strong>
              <span>查看应用、创建应用、编辑应用、启停应用</span>
            </div>
            <div>
              <strong>套餐资源</strong>
              <span>基础访问能力、客户端能力、试用与开通能力待配置</span>
            </div>
          </div>
        </section>

        <section class="app-detail-section">
          <div class="app-detail-section__head">
            <h4>项目同步信息</h4>
            <span>项目管理</span>
          </div>
          <div class="app-detail-fields">
            <div><span>同步版本</span><strong>{{ detailApp.version || '待项目管理同步' }}</strong></div>
            <div><span>发布渠道</span><strong>{{ detailApp.release_channel || '待项目管理同步' }}</strong></div>
            <div><span>文档入口</span><strong>{{ detailApp.doc_config || '待项目管理同步' }}</strong></div>
            <div><span>版本说明</span><strong>{{ detailApp.release_note || '待项目管理同步' }}</strong></div>
            <div><span>创建时间</span><strong>{{ formatDateTimeChina(detailApp.created_at) }}</strong></div>
            <div><span>更新时间</span><strong>{{ formatDateTimeChina(detailApp.updated_at) }}</strong></div>
          </div>
        </section>
      </div>
    </NeuroAgentDialog>
  </NeuroAgentPageShell>
</template>

<style scoped>
@import '@/styles/theme/neuro-theme.css';

.app-center-page :deep(.neuro-page-shell__panel-body) {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.app-center-layout {
  width: 100%;
}

.app-center-main {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.app-center-production-card p {
  margin: 0;
  color: var(--neuro-text-secondary);
}

.app-stat {
  width: 144px;
  min-height: 94px;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  justify-content: center;
  gap: 8px;
  padding: 14px 16px;
  border: 1px solid color-mix(in srgb, var(--neuro-primary) 30%, transparent);
  border-radius: var(--neuro-radius-md);
  background: color-mix(in srgb, var(--neuro-surface-2) 78%, transparent);
  color: var(--neuro-text);
  text-align: left;
  cursor: pointer;
  transition: border-color 0.18s ease, background 0.18s ease, transform 0.18s ease;
}

.app-stat strong {
  display: block;
  font-size: 30px;
  line-height: 1;
  color: var(--neuro-text);
}

.app-stat span {
  display: block;
  font-size: 13px;
  font-weight: 800;
  color: var(--neuro-text-secondary);
}

.app-stat em {
  color: var(--neuro-text-secondary);
  font-size: 12px;
  font-style: normal;
}

.app-stat:hover,
.app-stat--active {
  border-color: color-mix(in srgb, var(--neuro-primary) 72%, transparent);
  background: color-mix(in srgb, var(--neuro-primary) 12%, var(--neuro-surface-2));
  transform: translateY(-1px);
}

.app-stat--success strong {
  color: #65df98;
}

.app-stat--warning strong {
  color: #f2994a;
}

.app-stat--info strong {
  color: #72a6ff;
}

.app-stat--muted strong {
  color: #97a1b3;
}

.app-center-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 14px 20px 12px;
  border-radius: 0;
  background: linear-gradient(
    180deg,
    color-mix(in srgb, var(--neuro-background) 22%, transparent) 0%,
    transparent 72%
  );
}

.app-center-toolbar__filters,
.app-center-toolbar__actions {
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 10px;
}

.app-center-toolbar__filters {
  flex: 0 1 auto;
  flex-wrap: nowrap;
}

.app-center-toolbar__actions {
  flex: 0 0 auto;
  margin-left: auto;
}

.app-center-search,
.app-center-toolbar__filters select {
  height: 34px;
  min-width: 0;
  border: 1px solid rgba(45, 55, 72, 0.6);
  border-radius: var(--neuro-radius-md);
  background: rgba(255, 255, 255, 0.04);
  color: var(--neuro-text);
  font-size: 13px;
  box-shadow: none;
  transition: border-color 0.2s, box-shadow 0.2s, background-color 0.2s;
}

.app-center-search {
  flex: 0 0 360px;
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 12px;
}

.app-center-search:focus-within,
.app-center-toolbar__filters select:focus {
  border-color: rgba(0, 245, 212, 0.4);
  box-shadow: 0 0 0 3px rgba(0, 245, 212, 0.08);
  outline: none;
}

.app-center-search input {
  min-width: 0;
  width: 100%;
  border: 0;
  outline: 0;
  background: transparent;
  color: var(--neuro-text);
  font-size: 13px;
}

.app-center-search input::placeholder {
  color: var(--neuro-text-muted);
  font-size: 13px;
}

.app-center-toolbar__filters select {
  flex: 0 0 148px;
  padding: 0 28px 0 10px;
  appearance: none;
  background-image:
    linear-gradient(45deg, transparent 50%, #94a3b8 50%),
    linear-gradient(135deg, #94a3b8 50%, transparent 50%);
  background-position:
    calc(100% - 16px) 13px,
    calc(100% - 10px) 13px;
  background-size: 6px 6px, 6px 6px;
  background-repeat: no-repeat;
}

.app-btn {
  height: 34px;
  flex: 0 0 auto;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 0 14px;
  border: 1px solid color-mix(in srgb, var(--neuro-border) 90%, transparent);
  border-radius: var(--neuro-radius-md);
  background: color-mix(in srgb, var(--neuro-surface) 84%, transparent);
  color: var(--neuro-text);
  cursor: pointer;
  white-space: nowrap;
}

.app-btn--primary {
  border-color: transparent;
  background: linear-gradient(135deg, #00d4aa 0%, #7c3aed 100%);
  color: #051616;
  font-weight: 700;
}

.app-type-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.app-type-group {
  overflow: hidden;
  border: 1px solid color-mix(in srgb, var(--neuro-border) 84%, transparent);
  border-radius: var(--neuro-radius-lg);
  background: color-mix(in srgb, var(--neuro-surface) 78%, transparent);
}

.app-type-group__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 14px 16px;
  border-bottom: 1px solid color-mix(in srgb, var(--neuro-border) 72%, transparent);
  background:
    radial-gradient(circle at 18% 0%, color-mix(in srgb, var(--neuro-primary) 18%, transparent), transparent 38%),
    color-mix(in srgb, var(--neuro-surface-2) 82%, transparent);
}

.app-type-group__head > div {
  min-width: 0;
  display: flex;
  align-items: baseline;
  gap: 10px;
}

.app-type-group__head strong {
  color: var(--neuro-text);
  font-size: 18px;
  font-weight: 900;
}

.app-type-group__head span {
  color: var(--neuro-text-secondary);
  font-size: 12px;
  text-transform: uppercase;
}

.app-type-group__head em {
  min-width: 34px;
  height: 26px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 1px solid color-mix(in srgb, var(--neuro-primary) 34%, transparent);
  border-radius: 999px;
  background: color-mix(in srgb, var(--neuro-primary) 12%, transparent);
  color: var(--neuro-primary);
  font-style: normal;
  font-weight: 900;
}

.app-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  align-items: stretch;
  gap: 18px;
  padding: 18px;
}

.app-card {
  position: relative;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 22px;
  border: 1px solid color-mix(in srgb, var(--neuro-border) 88%, transparent);
  border-radius: var(--neuro-radius-lg);
  background:
    linear-gradient(180deg, color-mix(in srgb, var(--neuro-surface-2) 72%, transparent), transparent 68%),
    color-mix(in srgb, var(--neuro-surface) 90%, transparent);
  box-shadow: inset 0 1px 0 color-mix(in srgb, var(--neuro-primary) 38%, transparent);
  transition:
    border-color var(--shell-t-fast) var(--shell-ease-standard),
    background var(--shell-t-fast) var(--shell-ease-standard),
    box-shadow var(--shell-t-fast) var(--shell-ease-standard),
    transform var(--shell-t-fast) var(--shell-ease-standard);
}

.app-card:hover {
  border-color: color-mix(in srgb, var(--neuro-primary) 58%, transparent);
  background:
    linear-gradient(180deg, color-mix(in srgb, var(--neuro-primary) 10%, transparent), transparent 70%),
    linear-gradient(180deg, color-mix(in srgb, var(--neuro-surface-2) 84%, transparent), transparent 68%),
    color-mix(in srgb, var(--neuro-surface) 96%, transparent);
  box-shadow:
    inset 0 1px 0 color-mix(in srgb, var(--neuro-primary) 58%, transparent),
    0 18px 44px color-mix(in srgb, #000 18%, transparent),
    0 0 0 1px color-mix(in srgb, var(--neuro-primary) 12%, transparent);
  transform: translateY(-2px);
}

.app-card:focus-within {
  border-color: color-mix(in srgb, var(--neuro-primary) 62%, transparent);
  box-shadow:
    inset 0 1px 0 color-mix(in srgb, var(--neuro-primary) 58%, transparent),
    0 0 0 3px color-mix(in srgb, var(--neuro-primary) 10%, transparent);
}

.app-card--disabled {
  opacity: 0.58;
  filter: grayscale(0.35);
}

.app-card__head,
.app-card__foot {
  display: flex;
  align-items: center;
  gap: 10px;
}

.app-card__icon {
  width: 54px;
  height: 54px;
  flex: 0 0 54px;
  display: grid;
  place-items: center;
  border: 1px solid color-mix(in srgb, var(--neuro-primary) 30%, transparent);
  border-radius: 16px;
  background: color-mix(in srgb, var(--neuro-primary) 16%, transparent);
  color: var(--neuro-primary);
  transition:
    border-color var(--shell-t-fast) var(--shell-ease-standard),
    background var(--shell-t-fast) var(--shell-ease-standard),
    transform var(--shell-t-fast) var(--shell-ease-standard);
}

.app-card:hover .app-card__icon {
  border-color: color-mix(in srgb, var(--neuro-primary) 54%, transparent);
  background: color-mix(in srgb, var(--neuro-primary) 22%, transparent);
  transform: translateY(-1px);
}

.app-card__icon span {
  font-size: 20px;
  font-weight: 900;
}

.app-card__title {
  min-width: 0;
  flex: 1 1 auto;
}

.app-card__title h3 {
  margin: 0 0 4px;
  font-size: 20px;
  color: var(--neuro-text);
}

.app-card__title code {
  color: var(--neuro-text-secondary);
  font-size: 14px;
}

.app-status,
.app-card__badges span,
.app-card__channels span {
  display: inline-flex;
  align-items: center;
  min-height: 28px;
  padding: 0 12px;
  border-radius: 999px;
  font-size: 13px;
  font-weight: 800;
  white-space: nowrap;
}

.app-status {
  background: color-mix(in srgb, var(--neuro-surface-2) 92%, transparent);
  color: var(--neuro-text-secondary);
}

.app-status--online {
  background: color-mix(in srgb, #0f8f4d 22%, transparent);
  color: #22d36f;
}

.app-status--developing {
  background: color-mix(in srgb, #2f6fed 22%, transparent);
  color: #72a6ff;
}

.app-status--disabled {
  color: #8b95a7;
}

.app-status--draft {
  background: color-mix(in srgb, #b7791f 24%, transparent);
  color: #f2b95b;
}

.app-card__desc {
  margin: 0;
  color: var(--neuro-text-secondary);
  font-size: 14px;
  line-height: 1.6;
}

.app-card__badges,
.app-card__channels {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.app-card__badges span:nth-child(1) {
  border: 1px solid color-mix(in srgb, #5a9cff 40%, transparent);
  color: #78a9ff;
  background: color-mix(in srgb, #2f6fed 14%, transparent);
}

.app-card__badges span:nth-child(2) {
  border: 1px solid color-mix(in srgb, #f2994a 42%, transparent);
  color: #f2994a;
  background: color-mix(in srgb, #f2994a 12%, transparent);
}

.app-card__metrics {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 10px;
}

.app-card__metrics div {
  min-width: 0;
  padding: 10px 12px;
  border: 1px solid color-mix(in srgb, var(--neuro-border) 74%, transparent);
  border-radius: var(--neuro-radius-sm);
  background: color-mix(in srgb, var(--neuro-surface-2) 62%, transparent);
}

.app-card__metrics span,
.app-card__metrics strong {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.app-card__metrics span {
  color: var(--neuro-text-secondary);
  font-size: 12px;
  font-weight: 800;
}

.app-card__metrics strong {
  margin-top: 4px;
  color: var(--neuro-text);
  font-size: 14px;
}

.app-card__channels span {
  border: 1px solid color-mix(in srgb, #5a9cff 42%, transparent);
  color: #7fb0ff;
  background: color-mix(in srgb, #2f6fed 12%, transparent);
}

.app-card__foot {
  justify-content: space-between;
  padding-top: 10px;
  border-top: 1px solid color-mix(in srgb, var(--neuro-border) 70%, transparent);
  color: var(--neuro-text-secondary);
  font-size: 12px;
}

.app-card__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  padding-top: 14px;
  border-top: 1px solid color-mix(in srgb, var(--neuro-border) 70%, transparent);
}

.app-card__detail {
  height: 34px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  border: 1px solid color-mix(in srgb, var(--neuro-primary) 42%, transparent);
  border-radius: var(--neuro-radius-sm);
  background: color-mix(in srgb, var(--neuro-primary) 10%, transparent);
  color: var(--neuro-primary);
  cursor: pointer;
  font-weight: 700;
}

.app-card__detail:disabled {
  cursor: not-allowed;
  opacity: 0.52;
}

.app-card__detail--primary {
  min-width: 92px;
  border-color: transparent;
  background: var(--neuro-primary);
  color: #051616;
}

.app-center-empty {
  padding: 48px 16px;
  text-align: center;
  color: var(--neuro-text-secondary);
}

.app-center-production-card {
  min-height: 320px;
  display: grid;
  place-items: center;
  align-content: center;
  gap: 8px;
  padding: 42px;
  border: 1px solid color-mix(in srgb, var(--neuro-border) 84%, transparent);
  border-radius: var(--neuro-radius-lg);
  background: color-mix(in srgb, var(--neuro-surface) 88%, transparent);
  text-align: center;
}

.app-center-production-card strong {
  font-size: 54px;
  line-height: 1;
  color: var(--neuro-primary);
}

.app-center-production-card span {
  font-size: 20px;
  font-weight: 900;
  color: var(--neuro-text);
}

.app-create-wizard {
  height: 100%;
  min-height: 0;
  display: grid;
  grid-template-columns: 180px minmax(0, 1fr);
  gap: 20px;
  align-items: flex-start;
}

.app-create-steps {
  position: sticky;
  top: 0;
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 4px 0;
  opacity: 0.82;
}

.app-create-step {
  width: 100%;
  min-width: 0;
  display: grid;
  grid-template-columns: 22px minmax(0, 1fr);
  gap: 8px;
  align-items: flex-start;
  padding: 9px 8px;
  border: 1px solid transparent;
  border-radius: var(--neuro-radius-sm);
  background: transparent;
  color: var(--neuro-text-secondary);
  text-align: left;
  cursor: pointer;
}

.app-create-step i {
  width: 22px;
  height: 22px;
  display: inline-grid;
  place-items: center;
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--neuro-border) 76%, transparent);
  background: color-mix(in srgb, var(--neuro-surface) 86%, transparent);
  color: var(--neuro-text-secondary);
  font-style: normal;
  font-size: 12px;
  font-weight: 900;
}

.app-create-step strong,
.app-create-step em {
  display: block;
  min-width: 0;
}

.app-create-step strong {
  margin-bottom: 2px;
  color: var(--neuro-text);
  font-size: 13px;
  line-height: 1.35;
}

.app-create-step em {
  color: var(--neuro-text-secondary);
  font-style: normal;
  font-size: 12px;
  line-height: 1.45;
}

.app-create-step:hover {
  background: color-mix(in srgb, var(--neuro-surface) 74%, transparent);
}

.app-create-step.is-active {
  border-color: color-mix(in srgb, var(--neuro-primary) 28%, transparent);
  background: color-mix(in srgb, var(--neuro-primary) 7%, transparent);
}

.app-create-step.is-complete i {
  border-color: color-mix(in srgb, var(--neuro-primary) 46%, transparent);
  background: color-mix(in srgb, var(--neuro-primary) 18%, transparent);
  color: var(--neuro-primary);
}

.app-create-body {
  position: relative;
  width: 100%;
  min-width: 0;
  height: calc(86vh - 150px);
  overflow-y: auto;
  padding: 2px 8px 10px 0;
  scroll-behavior: smooth;
  scroll-padding-top: 4px;
  overscroll-behavior: contain;
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.app-create-body::-webkit-scrollbar {
  width: 8px;
}

.app-create-body::-webkit-scrollbar-track {
  background: transparent;
}

.app-create-body::-webkit-scrollbar-thumb {
  border-radius: 999px;
  background: transparent;
}

.app-create-body.is-scrolling::-webkit-scrollbar-thumb {
  background: color-mix(in srgb, var(--neuro-border) 80%, transparent);
}

.app-create-mode-panel {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 16px;
  border: 1px solid color-mix(in srgb, var(--neuro-primary) 24%, var(--neuro-border));
  border-radius: var(--neuro-radius-md);
  background: color-mix(in srgb, var(--neuro-surface-2) 40%, transparent);
}

.app-create-mode-panel__head {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  color: var(--neuro-text-secondary);
  font-size: 13px;
}

.app-create-mode-panel__head strong {
  color: var(--neuro-text);
  font-size: 15px;
  font-weight: 900;
}

.app-create-mode-switch {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.app-create-mode-switch button {
  min-width: 0;
  height: 42px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  border: 1px solid color-mix(in srgb, var(--neuro-border) 82%, transparent);
  border-radius: var(--neuro-radius-sm);
  background: var(--neuro-surface);
  color: var(--neuro-text-secondary);
  font-weight: 900;
  cursor: pointer;
}

.app-create-mode-switch button.is-active {
  border-color: color-mix(in srgb, var(--neuro-primary) 68%, transparent);
  background: color-mix(in srgb, var(--neuro-primary) 12%, var(--neuro-surface));
  color: var(--neuro-text);
}

.app-manifest-input {
  display: none;
}

.app-create-mode-panel__result {
  margin: 0;
  padding: 10px 12px;
  border-left: 2px solid color-mix(in srgb, var(--neuro-primary) 58%, transparent);
  color: var(--neuro-text-secondary);
  font-size: 13px;
  line-height: 1.5;
}

.app-create-mode-panel__result strong {
  display: block;
  color: var(--neuro-text);
}

.app-create-panel {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 18px;
  padding: 20px;
  border: 1px solid color-mix(in srgb, var(--neuro-border) 80%, transparent);
  border-radius: var(--neuro-radius-md);
  background:
    linear-gradient(180deg, color-mix(in srgb, var(--neuro-surface) 96%, transparent), color-mix(in srgb, var(--neuro-surface) 88%, transparent));
  box-shadow: 0 18px 44px color-mix(in srgb, #000 8%, transparent);
}

.app-create-panel--review {
  margin-bottom: 4px;
}

.app-create-panel__head {
  position: relative;
  display: flex;
  justify-content: space-between;
  gap: 14px;
  align-items: flex-start;
  margin: -20px -20px 2px;
  padding: 18px 20px 16px 24px;
  border-bottom: 1px solid color-mix(in srgb, var(--neuro-border) 88%, transparent);
  border-radius: var(--neuro-radius-md) var(--neuro-radius-md) 0 0;
  background:
    linear-gradient(180deg, color-mix(in srgb, var(--neuro-primary) 18%, var(--neuro-surface)), color-mix(in srgb, var(--neuro-primary) 10%, var(--neuro-bg)));
  box-shadow: inset 0 -1px 0 color-mix(in srgb, var(--neuro-primary) 10%, transparent);
}

.app-create-panel__head::before {
  content: "";
  position: absolute;
  left: 12px;
  top: 18px;
  bottom: 18px;
  width: 3px;
  border-radius: 999px;
  background: color-mix(in srgb, var(--neuro-primary) 78%, transparent);
}

.app-create-panel__head h3 {
  margin: 0 0 5px;
  color: var(--neuro-text);
  font-size: 18px;
  line-height: 1.3;
}

.app-create-panel__head p {
  margin: 0;
  color: var(--neuro-text-secondary);
  line-height: 1.65;
}

.app-create-panel__head > span {
  flex: 0 0 auto;
  padding: 5px 10px;
  border-radius: 999px;
  background: color-mix(in srgb, var(--neuro-primary) 14%, transparent);
  color: var(--neuro-primary);
  font-size: 12px;
  font-weight: 900;
}

.app-policy-layout {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.app-policy-block {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 14px;
  border: 1px solid color-mix(in srgb, var(--neuro-border) 76%, transparent);
  border-radius: var(--neuro-radius-sm);
  background: color-mix(in srgb, var(--neuro-surface-2) 34%, transparent);
}

.app-policy-block *,
.app-policy-block *::before,
.app-policy-block *::after {
  box-sizing: border-box;
}

.app-policy-block header {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.app-policy-block header strong {
  color: var(--neuro-text);
  font-size: 15px;
  font-weight: 900;
}

.app-policy-block header span {
  max-width: calc(100% - 44px);
  color: var(--neuro-text-secondary);
  line-height: 1.5;
}

.app-policy-alert {
  position: absolute;
  top: 18px;
  right: 18px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 22px;
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--neuro-primary) 48%, var(--neuro-border));
  background: color-mix(in srgb, var(--neuro-primary) 12%, var(--neuro-surface));
  color: var(--neuro-primary);
  font-size: 13px;
  font-weight: 900;
  line-height: 1;
  cursor: help;
}

.app-policy-alert:focus-visible {
  outline: none;
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--neuro-primary) 16%, transparent);
}

.app-policy-alert em {
  position: absolute;
  top: 50%;
  right: 30px;
  width: max-content;
  max-width: 280px;
  padding: 10px 12px;
  border: 1px solid color-mix(in srgb, var(--neuro-primary) 38%, var(--neuro-border));
  border-radius: var(--neuro-radius-sm);
  background: color-mix(in srgb, var(--neuro-elevated) 96%, var(--neuro-surface));
  box-shadow: var(--neuro-shadow-md);
  color: var(--neuro-text);
  font-size: 13px;
  font-style: normal;
  font-weight: 800;
  line-height: 1.5;
  opacity: 0;
  pointer-events: none;
  transform: translateY(-50%) translateX(4px);
  transition:
    opacity var(--shell-t-fast) var(--shell-ease-standard),
    transform var(--shell-t-fast) var(--shell-ease-standard);
  z-index: 5;
}

.app-policy-alert em::after {
  position: absolute;
  top: 50%;
  right: -6px;
  width: 10px;
  height: 10px;
  border-top: 1px solid color-mix(in srgb, var(--neuro-primary) 38%, var(--neuro-border));
  border-right: 1px solid color-mix(in srgb, var(--neuro-primary) 38%, var(--neuro-border));
  background: color-mix(in srgb, var(--neuro-elevated) 96%, var(--neuro-surface));
  content: '';
  transform: translateY(-50%) rotate(45deg);
}

.app-policy-alert:hover em,
.app-policy-alert:focus-visible em {
  opacity: 1;
  transform: translateY(-50%) translateX(0);
}

.app-policy-rules {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-top: -2px;
  padding: 8px 0 8px 12px;
  border-left: 2px solid color-mix(in srgb, var(--neuro-primary) 34%, var(--neuro-border));
}

.app-policy-rules div {
  min-width: 0;
  display: grid;
  grid-template-columns: 108px minmax(0, 1fr);
  gap: 10px;
  align-items: baseline;
}

.app-policy-rules strong {
  color: color-mix(in srgb, var(--neuro-text) 86%, var(--neuro-primary));
  font-size: 13px;
  font-weight: 900;
}

.app-policy-rules span {
  color: var(--neuro-text-secondary);
  font-size: 13px;
  line-height: 1.45;
}

.app-policy-block label {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-top: 2px;
}

.app-policy-block label > span {
  color: var(--neuro-text);
  font-size: 13px;
  font-weight: 800;
}

.app-policy-block input {
  width: 100%;
  min-width: 0;
  height: 40px;
  padding: 0 12px;
  border: 1px solid color-mix(in srgb, var(--neuro-border) 88%, transparent);
  border-radius: var(--neuro-radius-sm);
  background: var(--neuro-surface);
  color: var(--neuro-text);
  font: inherit;
}

.app-policy-block input:focus {
  outline: none;
  border-color: color-mix(in srgb, var(--neuro-primary) 54%, var(--neuro-border));
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--neuro-primary) 10%, transparent);
}

.app-choice-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.app-choice-grid--three {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.app-choice-grid--policy {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.app-choice-grid button {
  position: relative;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 5px;
  min-height: 72px;
  padding: 15px 14px;
  border: 1px solid color-mix(in srgb, var(--neuro-border) 82%, transparent);
  border-radius: var(--neuro-radius-sm);
  background: var(--neuro-surface);
  color: var(--neuro-text-secondary);
  text-align: left;
  cursor: pointer;
  transition:
    border-color var(--shell-t-fast) var(--shell-ease-standard),
    background var(--shell-t-fast) var(--shell-ease-standard),
    box-shadow var(--shell-t-fast) var(--shell-ease-standard),
    transform var(--shell-t-fast) var(--shell-ease-standard);
}

.app-choice-grid button:hover {
  border-color: color-mix(in srgb, var(--neuro-primary) 42%, var(--neuro-border));
  background: color-mix(in srgb, var(--neuro-primary) 5%, var(--neuro-surface));
  transform: translateY(-1px);
}

.app-choice-grid button:focus-visible {
  outline: none;
  border-color: color-mix(in srgb, var(--neuro-primary) 58%, var(--neuro-border));
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--neuro-primary) 10%, transparent);
}

.app-choice-grid button.is-active {
  border-color: color-mix(in srgb, var(--neuro-primary) 70%, transparent);
  background: color-mix(in srgb, var(--neuro-primary) 13%, var(--neuro-surface));
}

.app-choice-grid button.is-active:hover {
  border-color: color-mix(in srgb, var(--neuro-primary) 82%, transparent);
  background: color-mix(in srgb, var(--neuro-primary) 18%, var(--neuro-surface));
}

.app-choice-grid strong {
  color: var(--neuro-text);
  font-weight: 900;
}

.app-choice-grid span {
  line-height: 1.45;
}

.app-choice-card--custom input {
  width: 100%;
  height: 28px;
  min-width: 0;
  margin-top: 1px;
  padding: 0;
  border: 0;
  border-radius: 0;
  background: transparent;
  box-shadow: none;
  color: var(--neuro-text);
  font-size: 14px;
  font-weight: 800;
}

.app-choice-card--custom input::placeholder {
  color: var(--neuro-text-secondary);
  font-weight: 700;
}

.app-choice-card--custom input:focus {
  outline: none;
  border: 0;
  box-shadow: none;
}

.app-form--embedded {
  gap: 14px 18px;
}

.app-form--policy {
  grid-template-columns: 1fr;
}

.app-field-title {
  display: block;
  margin-bottom: 8px;
  color: var(--neuro-text);
  font-size: 13px;
  font-weight: 800;
}

.app-policy-help {
  display: block;
  color: var(--neuro-text-secondary);
  font-size: 13px;
  font-style: normal;
  line-height: 1.5;
}

.app-policy-fixed {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  min-height: 44px;
  padding: 10px 12px;
  border: 1px dashed color-mix(in srgb, var(--neuro-primary) 38%, var(--neuro-border));
  border-radius: var(--neuro-radius-sm);
  background: color-mix(in srgb, var(--neuro-primary) 8%, var(--neuro-surface));
}

.app-policy-fixed span {
  color: var(--neuro-text-secondary);
  font-size: 13px;
  font-weight: 800;
}

.app-policy-fixed strong {
  color: var(--neuro-text);
  font-size: 14px;
}

.app-create-checks {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
}

.app-form__label {
  display: block;
  margin-bottom: 8px;
  color: var(--neuro-text);
  font-size: 13px;
  font-weight: 800;
}

.app-create-checks--compact {
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.app-create-checks label {
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 10px;
  min-height: 44px;
  padding: 10px 12px;
  border: 1px solid color-mix(in srgb, var(--neuro-border) 80%, transparent);
  border-radius: var(--neuro-radius-sm);
  background: color-mix(in srgb, var(--neuro-surface) 88%, transparent);
  color: var(--neuro-text);
  font-weight: 800;
}

.app-create-checks input {
  width: 16px;
  height: 16px;
  accent-color: var(--neuro-primary);
}

.app-create-checks span {
  min-width: 0;
  line-height: 1.4;
}

.app-create-note {
  display: flex;
  gap: 8px;
  align-items: flex-start;
  padding: 12px 14px;
  border: 1px solid color-mix(in srgb, var(--neuro-primary) 24%, transparent);
  border-radius: var(--neuro-radius-sm);
  background: color-mix(in srgb, var(--neuro-primary) 8%, transparent);
  color: var(--neuro-text-secondary);
  line-height: 1.6;
}

.app-create-note strong {
  flex: 0 0 auto;
  color: var(--neuro-text);
}

.app-create-review {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.app-create-review > div {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 12px;
  border: 1px solid color-mix(in srgb, var(--neuro-border) 80%, transparent);
  border-radius: var(--neuro-radius-sm);
  background: color-mix(in srgb, var(--neuro-surface) 88%, transparent);
}

.app-create-review span {
  color: var(--neuro-text-secondary);
  font-size: 12px;
  font-weight: 800;
}

.app-create-review strong {
  min-width: 0;
  color: var(--neuro-text);
  line-height: 1.45;
  overflow-wrap: anywhere;
}

.app-form {
  width: 100%;
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
}

.app-form *,
.app-form *::before,
.app-form *::after {
  box-sizing: border-box;
}

.app-form label {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.app-form label > span {
  color: var(--neuro-text);
  font-weight: 800;
}

.app-form input,
.app-form select,
.app-form textarea {
  width: 100%;
  min-width: 0;
  max-width: 100%;
  border: 1px solid color-mix(in srgb, var(--neuro-border) 88%, transparent);
  border-radius: var(--neuro-radius-sm);
  background: var(--neuro-surface);
  color: var(--neuro-text);
  font: inherit;
  transition:
    border-color var(--shell-t-fast) var(--shell-ease-standard),
    background var(--shell-t-fast) var(--shell-ease-standard),
    box-shadow var(--shell-t-fast) var(--shell-ease-standard);
}

.app-form input::placeholder,
.app-form textarea::placeholder {
  color: color-mix(in srgb, var(--neuro-text-secondary) 76%, transparent);
}

.app-form input:hover,
.app-form select:hover,
.app-form textarea:hover {
  border-color: color-mix(in srgb, var(--neuro-primary) 34%, var(--neuro-border));
  background: color-mix(in srgb, var(--neuro-primary) 5%, var(--neuro-surface));
}

.app-form input:focus,
.app-form select:focus,
.app-form textarea:focus {
  outline: none;
  border-color: color-mix(in srgb, var(--neuro-primary) 54%, var(--neuro-border));
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--neuro-primary) 10%, transparent);
}

.app-form input:disabled {
  color: var(--neuro-text-secondary);
  background: color-mix(in srgb, var(--neuro-surface) 86%, var(--neuro-background));
}

.app-form input,
.app-form select {
  height: 40px;
  padding: 0 12px;
}

.app-form textarea {
  min-height: 92px;
  padding: 12px;
  resize: vertical;
}

.app-form textarea.app-form__rich {
  min-height: 150px;
}

.app-form__full {
  grid-column: 1 / -1;
}

.app-owner-select {
  box-sizing: border-box;
  width: 100%;
  min-width: 0;
  min-height: 40px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 6px 12px;
  border: 1px solid color-mix(in srgb, var(--neuro-border) 88%, transparent);
  border-radius: var(--neuro-radius-sm);
  background: var(--neuro-surface);
  color: var(--neuro-text);
  font: inherit;
  text-align: left;
  cursor: pointer;
  transition:
    border-color var(--shell-t-fast) var(--shell-ease-standard),
    background var(--shell-t-fast) var(--shell-ease-standard),
    box-shadow var(--shell-t-fast) var(--shell-ease-standard);
}

.app-owner-select:hover {
  border-color: color-mix(in srgb, var(--neuro-primary) 34%, var(--neuro-border));
  background: color-mix(in srgb, var(--neuro-primary) 5%, var(--neuro-surface));
}

.app-owner-select:focus-visible {
  outline: none;
  border-color: color-mix(in srgb, var(--neuro-primary) 54%, var(--neuro-border));
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--neuro-primary) 10%, transparent);
}

.app-owner-select > span {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.app-owner-select > span.is-placeholder {
  color: color-mix(in srgb, var(--neuro-text-secondary) 76%, transparent);
}

.app-owner-select > em {
  flex: 0 0 auto;
  color: var(--neuro-primary);
  font-size: 12px;
  font-style: normal;
  font-weight: 900;
}

.app-owner-select__chips {
  flex: 1 1 auto;
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;
  min-width: 0;
  overflow: visible;
  text-overflow: clip;
  white-space: normal;
}

.app-owner-select__chips button {
  max-width: 180px;
  min-width: 0;
  min-height: 28px;
  padding: 3px 10px;
  border: 1px solid color-mix(in srgb, var(--neuro-primary) 34%, transparent);
  border-radius: 999px;
  background: color-mix(in srgb, var(--neuro-primary) 12%, transparent);
  color: var(--neuro-primary);
  font-size: 12px;
  font-weight: 900;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  cursor: pointer;
}

.app-owner-select__chips button:hover {
  border-color: color-mix(in srgb, var(--neuro-primary) 62%, transparent);
  background: color-mix(in srgb, var(--neuro-primary) 18%, transparent);
}

.app-owner-picker-backdrop {
  position: fixed;
  inset: 0;
  z-index: 10030;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
  background: color-mix(in srgb, #020617 62%, transparent);
  backdrop-filter: blur(6px);
  overscroll-behavior: none;
  touch-action: none;
}

.app-owner-picker {
  width: min(560px, calc(100vw - 48px));
  max-height: min(680px, calc(100vh - 48px));
  display: flex;
  flex-direction: column;
  overflow: hidden;
  border: 1px solid color-mix(in srgb, var(--neuro-border) 90%, transparent);
  border-radius: var(--neuro-radius-md);
  background:
    linear-gradient(180deg, color-mix(in srgb, var(--neuro-primary) 9%, transparent), transparent 32%),
    var(--neuro-surface);
  box-shadow: 0 26px 80px color-mix(in srgb, #000 42%, transparent);
  overscroll-behavior: contain;
  touch-action: auto;
  font-size: 14px;
}

.app-owner-picker *,
.app-owner-picker *::before,
.app-owner-picker *::after {
  box-sizing: border-box;
}

.app-owner-picker__head {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  padding: 16px 20px 13px;
  border-bottom: 1px solid color-mix(in srgb, var(--neuro-border) 80%, transparent);
}

.app-owner-picker__head h3 {
  margin: 0 0 4px;
  color: var(--neuro-text);
  font-size: 17px;
  line-height: 1.35;
}

.app-owner-picker__head p {
  margin: 0;
  color: var(--neuro-text-secondary);
  font-size: 14px;
  line-height: 1.55;
}

.app-owner-picker__head button {
  width: 32px;
  height: 32px;
  border: 1px solid color-mix(in srgb, var(--neuro-border) 80%, transparent);
  border-radius: var(--neuro-radius-sm);
  background: color-mix(in srgb, var(--neuro-surface-2) 78%, transparent);
  color: var(--neuro-text-secondary);
  font-size: 22px;
  line-height: 1;
  cursor: pointer;
}

.app-owner-picker__search {
  position: relative;
  padding: 13px 20px 12px;
}

.app-owner-picker__search-icon {
  position: absolute;
  left: 34px;
  top: 50%;
  width: 16px;
  height: 16px;
  transform: translateY(-50%);
  color: var(--neuro-text-secondary);
}

.app-owner-picker__search input {
  width: 100%;
  max-width: 100%;
  height: 38px;
  padding: 0 12px 0 36px;
  border: 1px solid color-mix(in srgb, var(--neuro-border) 88%, transparent);
  border-radius: var(--neuro-radius-sm);
  background: var(--neuro-surface);
  color: var(--neuro-text);
  font: inherit;
  font-size: 14px;
}

.app-owner-picker__search input::placeholder {
  color: color-mix(in srgb, var(--neuro-text-secondary) 76%, transparent);
}

.app-owner-picker__search input:focus {
  outline: none;
  border-color: color-mix(in srgb, var(--neuro-primary) 54%, var(--neuro-border));
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--neuro-primary) 10%, transparent);
}

.app-owner-picker__list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  min-height: 180px;
  max-height: 360px;
  overflow-y: auto;
  overscroll-behavior: contain;
  padding: 0 20px 16px;
}

.app-owner-picker__list::-webkit-scrollbar {
  width: 8px;
}

.app-owner-picker__list::-webkit-scrollbar-track {
  background: transparent;
}

.app-owner-picker__list::-webkit-scrollbar-thumb {
  border-radius: 999px;
  background: transparent;
}

.app-owner-picker__list:hover::-webkit-scrollbar-thumb {
  background: color-mix(in srgb, var(--neuro-border) 80%, transparent);
}

.app-owner-picker__item {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 11px 12px;
  border: 1px solid color-mix(in srgb, var(--neuro-border) 80%, transparent);
  border-radius: var(--neuro-radius-sm);
  background: var(--neuro-surface);
  color: var(--neuro-text);
  text-align: left;
  cursor: pointer;
}

.app-owner-picker__item.is-selected {
  border-color: color-mix(in srgb, var(--neuro-primary) 62%, transparent);
  background: color-mix(in srgb, var(--neuro-primary) 14%, var(--neuro-surface));
}

.app-owner-picker__item i {
  flex: 0 0 auto;
  width: 21px;
  height: 21px;
  display: grid;
  place-items: center;
  border: 1px solid color-mix(in srgb, var(--neuro-primary) 38%, var(--neuro-border));
  border-radius: 999px;
  color: var(--neuro-primary);
  font-style: normal;
  font-size: 13px;
  font-weight: 900;
}

.app-owner-picker__item span {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.app-owner-picker__item strong,
.app-owner-picker__item em {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.app-owner-picker__item strong {
  color: var(--neuro-text);
  font-size: 14px;
  line-height: 1.35;
}

.app-owner-picker__item em {
  color: var(--neuro-text-secondary);
  font-style: normal;
  font-size: 13px;
  line-height: 1.35;
}

.app-owner-picker__empty {
  padding: 34px 12px;
  text-align: center;
  color: var(--neuro-text-secondary);
  font-size: 14px;
}

.app-owner-picker__more {
  padding: 4px 0 2px;
  text-align: center;
  color: var(--neuro-text-secondary);
  font-size: 12px;
}

.app-owner-picker__foot {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 13px 20px 16px;
  border-top: 1px solid color-mix(in srgb, var(--neuro-border) 80%, transparent);
}

.app-owner-picker__foot > span {
  color: var(--neuro-text-secondary);
  font-size: 13px;
}

.app-owner-picker__foot > div {
  display: flex;
  gap: 8px;
}

.app-owner-picker__ghost,
.app-owner-picker__primary {
  min-height: 34px;
  padding: 0 13px;
  border-radius: var(--neuro-radius-sm);
  font-size: 13px;
  font-weight: 900;
  cursor: pointer;
}

.app-owner-picker__ghost {
  border: 1px solid color-mix(in srgb, var(--neuro-border) 82%, transparent);
  background: color-mix(in srgb, var(--neuro-surface-2) 72%, transparent);
  color: var(--neuro-text-secondary);
}

.app-owner-picker__primary {
  border: 0;
  background: var(--neuro-gradient-primary);
  color: #04111d;
}

.app-detail {
  display: flex;
  flex-direction: column;
  gap: 18px;
  min-height: 0;
}

.app-detail-hero {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 18px;
  padding: 20px;
  border: 1px solid color-mix(in srgb, var(--neuro-primary) 24%, transparent);
  border-radius: var(--neuro-radius-md);
  background:
    linear-gradient(135deg, color-mix(in srgb, var(--neuro-primary) 10%, transparent), transparent 38%),
    color-mix(in srgb, var(--neuro-surface) 92%, transparent);
}

.app-detail-hero__main {
  min-width: 0;
  display: flex;
  gap: 14px;
  align-items: flex-start;
}

.app-detail-hero h3 {
  margin: 0 0 5px;
  color: var(--neuro-text);
  font-size: 22px;
  line-height: 1.25;
}

.app-detail-hero code,
.app-detail-hero p,
.app-detail-hero__meta span {
  color: var(--neuro-text-secondary);
}

.app-detail-hero p {
  max-width: 640px;
  margin: 10px 0 0;
  line-height: 1.7;
}

.app-detail-hero__meta {
  flex: 0 0 150px;
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 8px;
  text-align: right;
}

.app-detail-hero__meta strong {
  color: var(--neuro-text);
}

.app-detail-section {
  display: flex;
  flex-direction: column;
  gap: 14px;
  padding: 18px;
  border: 1px solid color-mix(in srgb, var(--neuro-border) 80%, transparent);
  border-radius: var(--neuro-radius-md);
  background: color-mix(in srgb, var(--neuro-surface) 90%, transparent);
}

.app-detail-section__head {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  align-items: center;
}

.app-detail-section__head h4 {
  margin: 0;
  color: var(--neuro-text);
  font-size: 16px;
}

.app-detail-section__head > span {
  padding: 4px 9px;
  border-radius: 999px;
  background: color-mix(in srgb, var(--neuro-primary) 10%, transparent);
  color: var(--neuro-primary);
  font-size: 12px;
  font-weight: 900;
}

.app-detail-fields {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.app-detail-fields > div,
.app-detail-assets > div {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 12px;
  border: 1px solid color-mix(in srgb, var(--neuro-border) 72%, transparent);
  border-radius: var(--neuro-radius-sm);
  background: color-mix(in srgb, var(--neuro-surface) 84%, transparent);
}

.app-detail-fields__full {
  grid-column: 1 / -1;
}

.app-detail-fields span,
.app-detail-assets span {
  color: var(--neuro-text-secondary);
  font-size: 12px;
  font-weight: 800;
}

.app-detail-fields strong,
.app-detail-assets strong {
  min-width: 0;
  color: var(--neuro-text);
  line-height: 1.45;
  overflow-wrap: anywhere;
}

.app-detail-rich {
  margin: 0;
  color: var(--neuro-text);
  line-height: 1.7;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
}

.app-detail-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.app-detail-tags span {
  padding: 8px 11px;
  border: 1px solid color-mix(in srgb, var(--neuro-primary) 22%, transparent);
  border-radius: 999px;
  background: color-mix(in srgb, var(--neuro-primary) 8%, transparent);
  color: var(--neuro-text);
  font-weight: 800;
}

.app-detail-tags span.is-muted {
  border-color: color-mix(in srgb, var(--neuro-border) 76%, transparent);
  background: color-mix(in srgb, var(--neuro-surface) 78%, transparent);
  color: var(--neuro-text-secondary);
}

.app-detail-assets {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

@media (max-width: 980px) {
  .app-center-toolbar {
    align-items: stretch;
    flex-wrap: wrap;
  }

  .app-center-toolbar__filters {
    flex-basis: 100%;
    flex-wrap: wrap;
  }

  .app-center-search {
    flex-basis: 100%;
  }

  .app-center-toolbar__filters select {
    flex: 1 1 160px;
  }

  .app-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .app-card__metrics {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .app-create-wizard {
    grid-template-columns: 1fr;
  }

  .app-create-steps {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .app-create-checks {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 640px) {
  .app-center-toolbar__filters > *,
  .app-center-toolbar__actions {
    flex-basis: 100%;
  }

  .app-center-toolbar__actions {
    justify-content: flex-end;
  }

  .app-type-group__head > div {
    flex-direction: column;
    gap: 4px;
    align-items: flex-start;
  }

  .app-card {
    padding: 16px;
  }

  .app-grid {
    grid-template-columns: 1fr;
  }

  .app-card__head {
    align-items: flex-start;
  }

  .app-card__metrics {
    grid-template-columns: 1fr;
  }

  .app-form {
    grid-template-columns: 1fr;
  }

  .app-create-steps,
  .app-create-checks,
  .app-create-review {
    grid-template-columns: 1fr;
  }

  .app-create-panel__head,
  .app-create-note {
    flex-direction: column;
  }

  .app-detail-hero,
  .app-detail-section__head {
    flex-direction: column;
    align-items: flex-start;
  }

  .app-detail-hero__meta {
    flex: 0 0 auto;
    align-items: flex-start;
    text-align: left;
  }

  .app-detail-fields,
  .app-detail-assets {
    grid-template-columns: 1fr;
  }
}
</style>
