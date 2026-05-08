<script setup lang="ts">
defineOptions({ name: 'PlanManagementView' })

import {
  Check,
  Close,
  Connection,
  CopyDocument,
  EditPen,
  Grid,
  Link,
  Plus,
  Refresh,
  SetUp,
  Tickets,
} from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { computed, onMounted, ref } from 'vue'

import {
  createFeature,
  createPlan,
  createQuota,
  copyPlan,
  deletePlan,
  fetchFeatures,
  fetchPlanMatrix,
  fetchPlanFeatures,
  fetchPlanQuotas,
  fetchQuotas,
  fetchTenantFeatureOverrides,
  fetchTenantQuotaOverrides,
  fetchTenantQuotaUsage,
  fetchTenantSubscription,
  savePlanFeatures,
  savePlanCapabilities,
  savePlanQuotas,
  saveTenantFeatureOverrides,
  saveTenantQuotaOverrides,
  saveTenantSubscription,
  updateFeature,
  updatePlan,
  updateQuota,
} from '@/api/plan'
import { archiveSuccessMessage, confirmArchiveAction } from '@/composables/useArchiveConfirm'
import type {
  Feature,
  FeaturePayload,
  Plan,
  PlanCapabilityMatrixData,
  PlanCapabilityNode,
  PlanCopyPayload,
  PlanPayload,
  PlanQuotaValue,
  Quota,
  QuotaPayload,
  TenantFeatureOverride,
  TenantQuotaOverride,
  TenantQuotaUsage,
  TenantSubscriptionPayload,
} from '@/api/plan'
import { fetchTenants } from '@/api/tenant'
import type { Tenant } from '@/api/tenant'
import { usePermissionStore } from '@/stores/permission'
import { useUiPreferencesStore } from '@/stores/uiPreferences'
import {
  collectFeatureIds,
  collectQuotaValues,
  patchPlanQuotaValues,
  quotaValuesFromMatrix,
  selectionFromMatrix,
  setNodeEnabled,
  type PlanFeatureSelection,
  type PlanQuotaValues,
} from '@/composables/usePlanCapabilityMatrix'
import { compareQuotaDisplayOrder, sortQuotasByDisplayOrder } from '@/utils/quotaDisplayOrder'
import PlanCapabilityDrawer from '@/views/components/PlanCapabilityDrawer.vue'
import PlanCapabilityMatrix from '@/views/components/PlanCapabilityMatrix.vue'

type WorkTab = 'features' | 'quotas' | 'tenants'

interface QuotaMatrixRow {
  id: string
  capability: string
  capabilityCode: string
  quota_id: number
  quota_code: string
  quota_name: string
  quota_type: string
  period_type?: string | null
  unit?: string | null
  values: Record<number, number>
}

const loading = ref(false)
const uiPrefs = useUiPreferencesStore()
const perm = usePermissionStore()
const theme = computed(() => uiPrefs.theme)
const plans = ref<Plan[]>([])
const totalPlans = ref(0)
const planMatrix = ref<PlanCapabilityMatrixData | null>(null)
const features = ref<Feature[]>([])
const quotas = ref<Quota[]>([])
const tenants = ref<Tenant[]>([])
const selectedPlanId = ref<number | null>(null)
const activeTab = ref<WorkTab>('features')
const keyword = ref('')

const selectedFeatureIds = ref<number[]>([])
const quotaValues = ref<Record<number, number>>({})
const matrixFeatureSelection = ref<PlanFeatureSelection>({})
const matrixQuotaValues = ref<PlanQuotaValues>({})
const savingFeatures = ref(false)
const savingQuotas = ref(false)
const savingMatrixPlanId = ref<number | null>(null)

const planDialog = ref(false)
const copyPlanDialog = ref(false)
const featureDialog = ref(false)
const quotaDialog = ref(false)
const subscriptionDialog = ref(false)
const featureOverrideDialog = ref(false)
const quotaOverrideDialog = ref(false)
const quotaUsageDialog = ref(false)
const capabilityDrawer = ref(false)
const editPlanId = ref<number | null>(null)
const copySourcePlan = ref<Plan | null>(null)
const editFeatureId = ref<number | null>(null)
const editQuotaId = ref<number | null>(null)
const subscriptionTenant = ref<Tenant | null>(null)
const overrideTenant = ref<Tenant | null>(null)
const usageTenant = ref<Tenant | null>(null)
const quotaUsageRows = ref<TenantQuotaUsage[]>([])
const capabilityDrawerPlan = ref<Plan | null>(null)
const capabilityDrawerNode = ref<PlanCapabilityNode | null>(null)

function blankPlan(): PlanPayload {
  return {
    plan_code: '',
    plan_name: '',
    plan_type: 'BASIC',
    billing_cycle: 'MONTH',
    price: 0,
    status: 1,
    is_default: false,
    sort_order: 0,
    description: '',
  }
}

function blankFeature(): FeaturePayload {
  return {
    feature_code: '',
    feature_name: '',
    feature_type: 'MENU',
    parent_id: 0,
    status: 1,
    description: '',
  }
}

function blankQuota(): QuotaPayload {
  return {
    quota_code: '',
    quota_name: '',
    quota_type: 'STATIC',
    period_type: 'NONE',
    unit: 'COUNT',
    status: 1,
    description: '',
  }
}

const planForm = ref<PlanPayload>(blankPlan())
const copyPlanForm = ref<PlanCopyPayload>({ plan_code: '', plan_name: '', description: '' })
const featureForm = ref<FeaturePayload>(blankFeature())
const quotaForm = ref<QuotaPayload>(blankQuota())
const featureOverrideStates = ref<Record<number, 'inherit' | 'enable' | 'disable'>>({})
const featureOverrideReasons = ref<Record<number, string>>({})
const quotaOverrideEnabled = ref<Record<number, boolean>>({})
const quotaOverrideValues = ref<Record<number, number>>({})
const quotaOverrideReasons = ref<Record<number, string>>({})
const quotaOverrideRows = ref<Quota[]>([])
const subscriptionForm = ref<TenantSubscriptionPayload>({
  plan_id: 0,
  subscription_status: 'ACTIVE',
  start_time: '',
  end_time: '',
  trial_end_time: '',
  auto_renew: false,
  frozen_reason: '',
})

const selectedPlan = computed(() => plans.value.find((item) => item.id === selectedPlanId.value) || null)
const enabledPlanCount = computed(() => plans.value.filter((item) => item.status === 1).length)
const activeFeatureCount = computed(() => (
  selectedPlan.value ? (matrixFeatureSelection.value[selectedPlan.value.id] || selectedFeatureIds.value).length : 0
))
const configuredQuotaCount = computed(() => quotaMatrixRows.value.length)
const enterprisePlan = computed(() => plans.value.find((item) => item.plan_code === 'ENTERPRISE'))
const visibleMatrix = computed<PlanCapabilityMatrixData | null>(() => {
  if (!planMatrix.value) return null
  const visiblePlanIds = new Set(plans.value.map((plan) => plan.id))
  return {
    ...planMatrix.value,
    plans: planMatrix.value.plans.filter((plan) => visiblePlanIds.has(plan.id)),
  }
})
const quotaMatrixColumns = computed(() => `460px repeat(${visibleMatrix.value?.plans.length || 0}, 190px)`)
const drawerEnabledFeatureIds = computed(() => (
  capabilityDrawerPlan.value ? (matrixFeatureSelection.value[capabilityDrawerPlan.value.id] || []) : []
))
const drawerQuotaValues = computed(() => (
  capabilityDrawerPlan.value ? (matrixQuotaValues.value[capabilityDrawerPlan.value.id] || {}) : {}
))

const featureGroups = computed(() => {
  const groups: Record<string, Feature[]> = {}
  for (const feature of features.value) {
    const key = feature.feature_type || 'OTHER'
    if (!groups[key]) groups[key] = []
    groups[key].push(feature)
  }
  return groups
})

const quotaRows = computed(() => {
  if (!selectedPlan.value || !planMatrix.value) return []
  const planId = selectedPlan.value.id
  const selected = new Set(matrixFeatureSelection.value[planId] || selectedFeatureIds.value)
  const byQuota = new Map<number, Quota & { quota_value: number }>()
  const quotaById = new Map(quotas.value.map((quota) => [quota.id, quota]))
  function walk(node: PlanCapabilityNode) {
    if (node.feature_id != null && selected.has(node.feature_id)) {
      for (const quota of collectQuotaValues(node, planId)) {
        byQuota.set(quota.quota_id, {
          ...(quotaById.get(quota.quota_id) || {
            id: quota.quota_id,
            quota_code: quota.quota_code,
            quota_name: quota.quota_name,
            quota_type: 'STATIC',
            period_type: quota.period_type,
            unit: quota.unit,
            status: 1,
            description: null,
          }),
          quota_value: matrixQuotaValues.value[planId]?.[quota.quota_id] ?? quota.quota_value,
        })
      }
    }
    for (const child of node.children || []) walk(child)
  }
  for (const node of planMatrix.value.nodes) walk(node)
  return [...byQuota.values()].sort(compareQuotaDisplayOrder)
})

const quotaMatrixRows = computed<QuotaMatrixRow[]>(() => {
  if (!visibleMatrix.value) return []
  const rows = new Map<string, QuotaMatrixRow>()
  const quotaById = new Map(quotas.value.map((quota) => [quota.id, quota]))
  function walk(node: PlanCapabilityNode) {
    if (node.feature_id != null) {
      for (const plan of visibleMatrix.value?.plans || []) {
        const cell = node.cells.find((item) => item.plan_id === plan.id)
        for (const quota of cell?.quota_values || []) {
          const key = `${node.id}-${quota.quota_id}`
          const meta = quotaById.get(quota.quota_id)
          if (!rows.has(key)) {
            rows.set(key, {
              id: key,
              capability: node.label,
              capabilityCode: node.feature_code || '',
              quota_id: quota.quota_id,
              quota_code: quota.quota_code,
              quota_name: quota.quota_name,
              quota_type: meta?.quota_type || 'STATIC',
              period_type: quota.period_type,
              unit: quota.unit,
              values: {},
            })
          }
          rows.get(key)!.values[plan.id] = matrixQuotaValues.value[plan.id]?.[quota.quota_id] ?? quota.quota_value
        }
      }
    }
    for (const child of node.children || []) walk(child)
  }
  for (const node of visibleMatrix.value.nodes) walk(node)
  return [...rows.values()].sort(compareQuotaDisplayOrder)
})

function typeLabel(value: string) {
  const map: Record<string, string> = {
    TRIAL: '试用',
    BASIC: '基础',
    PRO: '专业',
    ENTERPRISE: '企业',
    CUSTOM: '定制',
    MENU: '菜单',
    BUTTON: '按钮',
    API: '接口',
    SERVICE: '服务',
    CONFIG: '配置',
    STATIC: '静态',
    DYNAMIC: '动态',
  }
  return map[value] || value
}

function quotaText(value: number) {
  if (value === -1) return '无限制'
  if (value === 0) return '不可用'
  return String(value)
}

function dateTimeStart(date: string) {
  return date ? `${date}T00:00:00` : ''
}

function dateTimeEnd(date: string) {
  return date ? `${date}T23:59:59` : ''
}

async function loadAll() {
  loading.value = true
  try {
    const [matrixRes, featureRes, quotaRes, tenantRes] = await Promise.all([
      fetchPlanMatrix(),
      fetchFeatures(0, 300),
      fetchQuotas(0, 300),
      fetchTenants(0, 100),
    ])
    planMatrix.value = matrixRes
    matrixFeatureSelection.value = selectionFromMatrix(matrixRes)
    matrixQuotaValues.value = quotaValuesFromMatrix(matrixRes)
    const kw = keyword.value.trim().toLowerCase()
    plans.value = kw
      ? matrixRes.plans.filter((plan) => `${plan.plan_code} ${plan.plan_name}`.toLowerCase().includes(kw))
      : matrixRes.plans
    totalPlans.value = matrixRes.plans.length
    features.value = featureRes.items
    quotas.value = sortQuotasByDisplayOrder(quotaRes.items)
    tenants.value = tenantRes.items
    if (!selectedPlanId.value || !plans.value.some((item) => item.id === selectedPlanId.value)) {
      selectedPlanId.value = plans.value[0]?.id ?? null
    }
    if (selectedPlanId.value) await loadPlanConfig(selectedPlanId.value)
  } finally {
    loading.value = false
  }
}

async function loadPlanConfig(planId: number) {
  const [featureData, quotaData] = await Promise.all([
    fetchPlanFeatures(planId),
    fetchPlanQuotas(planId),
  ])
  selectedFeatureIds.value = featureData.feature_ids
  quotaValues.value = Object.fromEntries(quotaData.quotas.map((item: PlanQuotaValue) => [item.quota_id, item.quota_value]))
}

async function choosePlanById(planId: number) {
  selectedPlanId.value = planId
  await loadPlanConfig(planId)
}

async function focusPlan(plan: Plan) {
  if (selectedPlanId.value === plan.id) return
  await choosePlanById(plan.id)
}

function openPlan(row?: Plan) {
  editPlanId.value = row?.id ?? null
  planForm.value = row
    ? {
        plan_code: row.plan_code,
        plan_name: row.plan_name,
        plan_type: row.plan_type,
        billing_cycle: row.billing_cycle,
        price: Number(row.price || 0),
        status: row.status,
        is_default: row.is_default,
        sort_order: row.sort_order,
        description: row.description || '',
      }
    : blankPlan()
  planDialog.value = true
}

async function submitPlan() {
  if (!planForm.value.plan_code || !planForm.value.plan_name) return ElMessage.warning('请填写套餐编码和名称')
  if (editPlanId.value) await updatePlan(editPlanId.value, planForm.value)
  else await createPlan(planForm.value)
  planDialog.value = false
  ElMessage.success('套餐已保存')
  await loadAll()
}

async function togglePlan(row: Plan) {
  await updatePlan(row.id, { status: row.status === 1 ? 0 : 1 })
  ElMessage.success(row.status === 1 ? '套餐已停用' : '套餐已启用')
  await loadAll()
}

function openCopyPlan(row: Plan) {
  copySourcePlan.value = row
  copyPlanForm.value = {
    plan_code: `${row.plan_code}_COPY`,
    plan_name: `${row.plan_name} 副本`,
    description: row.description || '',
  }
  copyPlanDialog.value = true
}

async function submitCopyPlan() {
  if (!copySourcePlan.value) return
  if (!copyPlanForm.value.plan_code || !copyPlanForm.value.plan_name) return ElMessage.warning('请填写新套餐编码和名称')
  const created = await copyPlan(copySourcePlan.value.id, copyPlanForm.value)
  copyPlanDialog.value = false
  ElMessage.success('套餐已复制，新套餐默认停用')
  await loadAll()
  selectedPlanId.value = created.id
  await loadPlanConfig(created.id)
}

async function removePlan(row: Plan) {
  await confirmArchiveAction({
    name: row.plan_name,
    title: '归档套餐',
    detail: '已绑定租户的套餐不会被允许归档；未绑定套餐归档后仍保留历史配置。',
  })
  await deletePlan(row.id)
  ElMessage.success(archiveSuccessMessage(row.plan_name))
  if (selectedPlanId.value === row.id) selectedPlanId.value = null
  await loadAll()
}

function openFeature(row?: Feature) {
  editFeatureId.value = row?.id ?? null
  featureForm.value = row
    ? {
        feature_code: row.feature_code,
        feature_name: row.feature_name,
        feature_type: row.feature_type,
        parent_id: row.parent_id,
        menu_id: row.menu_id,
        api_method: row.api_method,
        api_path: row.api_path,
        service_key: row.service_key,
        status: row.status,
        description: row.description || '',
      }
    : blankFeature()
  featureDialog.value = true
}

async function submitFeature() {
  if (!featureForm.value.feature_code || !featureForm.value.feature_name) return ElMessage.warning('请填写功能编码和名称')
  if (editFeatureId.value) await updateFeature(editFeatureId.value, featureForm.value)
  else await createFeature(featureForm.value)
  featureDialog.value = false
  ElMessage.success('功能点已保存')
  await loadAll()
}

function openQuota(row?: Quota) {
  editQuotaId.value = row?.id ?? null
  quotaForm.value = row
    ? {
        quota_code: row.quota_code,
        quota_name: row.quota_name,
        quota_type: row.quota_type,
        period_type: row.period_type || 'NONE',
        unit: row.unit || 'COUNT',
        status: row.status,
        description: row.description || '',
      }
    : blankQuota()
  quotaDialog.value = true
}

function openQuotaFromMatrix(row: QuotaMatrixRow) {
  const quota = quotas.value.find((item) => item.id === row.quota_id)
  if (quota) openQuota(quota)
}

async function submitQuota() {
  if (!quotaForm.value.quota_code || !quotaForm.value.quota_name) return ElMessage.warning('请填写配额编码和名称')
  if (editQuotaId.value) await updateQuota(editQuotaId.value, quotaForm.value)
  else await createQuota(quotaForm.value)
  quotaDialog.value = false
  ElMessage.success('配额已保存')
  await loadAll()
}

async function submitFeatures() {
  if (!selectedPlan.value) return
  savingFeatures.value = true
  try {
    await savePlanFeatures(selectedPlan.value.id, selectedFeatureIds.value)
    await loadPlanConfig(selectedPlan.value.id)
    ElMessage.success('套餐功能已保存')
  } finally {
    savingFeatures.value = false
  }
}

async function submitQuotas() {
  if (!selectedPlan.value) return
  savingQuotas.value = true
  try {
    const values = quotaRows.value.map((quota) => ({ quota_id: quota.id, quota_value: quotaValues.value[quota.id] ?? 0 }))
    await savePlanQuotas(selectedPlan.value.id, values)
    ElMessage.success('套餐配额已保存')
  } finally {
    savingQuotas.value = false
  }
}

function openCapabilityDetail(payload: { plan: Plan; node: PlanCapabilityNode }) {
  capabilityDrawerPlan.value = payload.plan
  capabilityDrawerNode.value = payload.node
  capabilityDrawer.value = true
}

function toggleCapability(payload: { plan: Plan; node: PlanCapabilityNode; enabled: boolean }) {
  matrixFeatureSelection.value = setNodeEnabled(
    matrixFeatureSelection.value,
    payload.node,
    payload.plan,
    payload.enabled,
  )
}

function applyCapabilityDetail(payload: { featureIds: number[]; quotaValues: Record<number, number> }) {
  if (!capabilityDrawerPlan.value || !capabilityDrawerNode.value) return
  const planId = capabilityDrawerPlan.value.id
  const scopedIds = new Set(collectFeatureIds(capabilityDrawerNode.value))
  const selected = new Set(matrixFeatureSelection.value[planId] || [])
  for (const id of scopedIds) selected.delete(id)
  for (const id of payload.featureIds) {
    if (scopedIds.has(id)) selected.add(id)
  }
  matrixFeatureSelection.value = {
    ...matrixFeatureSelection.value,
    [planId]: [...selected].sort((a, b) => a - b),
  }
  matrixQuotaValues.value = patchPlanQuotaValues(matrixQuotaValues.value, planId, payload.quotaValues)
}

function matrixQuotaPayload(plan: Plan) {
  if (!planMatrix.value) return []
  const selected = new Set(matrixFeatureSelection.value[plan.id] || [])
  const byQuota = new Map<number, number>()
  function walk(node: PlanCapabilityNode) {
    if (node.feature_id != null && selected.has(node.feature_id)) {
      for (const quota of collectQuotaValues(node, plan.id)) {
        byQuota.set(quota.quota_id, matrixQuotaValues.value[plan.id]?.[quota.quota_id] ?? quota.quota_value)
      }
    }
    for (const child of node.children || []) walk(child)
  }
  for (const node of planMatrix.value.nodes) walk(node)
  return [...byQuota.entries()].map(([quota_id, quota_value]) => ({ quota_id, quota_value }))
}

function updateQuotaMatrixValue(planId: number, quotaId: number, value: number | undefined) {
  matrixQuotaValues.value = patchPlanQuotaValues(matrixQuotaValues.value, planId, {
    [quotaId]: value ?? 0,
  })
}

async function saveMatrixPlan(plan: Plan) {
  savingMatrixPlanId.value = plan.id
  try {
    await savePlanCapabilities(plan.id, {
      feature_ids: matrixFeatureSelection.value[plan.id] || [],
      quotas: matrixQuotaPayload(plan),
    })
    ElMessage.success(`${plan.plan_name} 能力矩阵已保存`)
    try {
      await loadAll()
      if (selectedPlanId.value === plan.id) await loadPlanConfig(plan.id)
    } catch (e) {
      ElMessage.error(e instanceof Error ? e.message : '刷新列表失败，请点「刷新」重试')
    }
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '保存失败')
  } finally {
    savingMatrixPlanId.value = null
  }
}

async function handleMatrixPlanAction(payload: { plan: Plan; action: 'edit' | 'copy' | 'toggle' | 'delete' | 'detail' }) {
  await focusPlan(payload.plan)
  if (payload.action === 'detail') {
    activeTab.value = 'quotas'
    return
  }
  if (payload.action === 'edit') {
    openPlan(payload.plan)
    return
  }
  if (payload.action === 'copy') {
    openCopyPlan(payload.plan)
    return
  }
  if (payload.action === 'toggle') {
    await togglePlan(payload.plan)
    return
  }
  await removePlan(payload.plan)
}

async function openSubscription(row: Tenant) {
  if (!selectedPlan.value) return
  subscriptionTenant.value = row
  const today = new Date().toISOString().slice(0, 10)
  const nextYear = new Date(Date.now() + 365 * 86400000).toISOString().slice(0, 10)
  subscriptionForm.value = {
    plan_id: selectedPlan.value.id,
    subscription_status: 'ACTIVE',
    start_time: dateTimeStart(today),
    end_time: dateTimeEnd(nextYear),
    trial_end_time: null,
    auto_renew: false,
    frozen_reason: '',
  }
  try {
    const current = await fetchTenantSubscription(row.id)
    subscriptionForm.value = {
      plan_id: current.plan_id,
      subscription_status: current.subscription_status,
      start_time: current.start_time,
      end_time: current.end_time,
      trial_end_time: current.trial_end_time,
      auto_renew: current.auto_renew,
      frozen_reason: current.frozen_reason || '',
    }
  } catch {
    // 未绑定套餐时使用默认表单。
  }
  subscriptionDialog.value = true
}

async function submitSubscription() {
  if (!subscriptionTenant.value) return
  const body = {
    ...subscriptionForm.value,
    end_time: subscriptionForm.value.end_time || null,
    trial_end_time: subscriptionForm.value.trial_end_time || null,
    frozen_reason: subscriptionForm.value.frozen_reason || null,
  }
  await saveTenantSubscription(subscriptionTenant.value.id, body)
  if (perm.profile?.tenant_id === subscriptionTenant.value.id) {
    await perm.load({ force: true })
  }
  subscriptionDialog.value = false
  ElMessage.success('租户套餐已更新')
}

async function openFeatureOverrides(row: Tenant) {
  overrideTenant.value = row
  const data = await fetchTenantFeatureOverrides(row.id)
  const byFeature = new Map<number, TenantFeatureOverride>(data.overrides.map((item) => [item.feature_id, item]))
  featureOverrideStates.value = {}
  featureOverrideReasons.value = {}
  for (const feature of features.value) {
    const override = byFeature.get(feature.id)
    featureOverrideStates.value[feature.id] = override ? (override.enabled ? 'enable' : 'disable') : 'inherit'
    featureOverrideReasons.value[feature.id] = override?.reason || ''
  }
  featureOverrideDialog.value = true
}

async function submitFeatureOverrides() {
  if (!overrideTenant.value) return
  const overrides = features.value
    .filter((feature) => featureOverrideStates.value[feature.id] !== 'inherit')
    .map((feature) => ({
      feature_id: feature.id,
      enabled: featureOverrideStates.value[feature.id] === 'enable',
      reason: featureOverrideReasons.value[feature.id] || null,
    }))
  await saveTenantFeatureOverrides(overrideTenant.value.id, overrides)
  featureOverrideDialog.value = false
  ElMessage.success('功能覆盖已保存')
}

async function openQuotaOverrides(row: Tenant) {
  overrideTenant.value = row
  const [overrideData, featureOverrideData] = await Promise.all([
    fetchTenantQuotaOverrides(row.id),
    fetchTenantFeatureOverrides(row.id),
  ])
  let planFeatureIds: number[] = []
  let subscriptionPlanId: number | null = null
  try {
    const subscription = await fetchTenantSubscription(row.id)
    subscriptionPlanId = subscription.plan_id
    const featureData = await fetchPlanFeatures(subscription.plan_id)
    planFeatureIds = featureData.feature_ids
  } catch {
    planFeatureIds = []
  }
  const effectiveFeatureIds = new Set(planFeatureIds)
  for (const item of featureOverrideData.overrides) {
    if (item.enabled) effectiveFeatureIds.add(item.feature_id)
    else effectiveFeatureIds.delete(item.feature_id)
  }
  const quotaIds = new Set<number>()
  function walkMatrix(node: PlanCapabilityNode) {
    if (subscriptionPlanId != null && node.feature_id != null && effectiveFeatureIds.has(node.feature_id)) {
      for (const quota of collectQuotaValues(node, subscriptionPlanId)) quotaIds.add(quota.quota_id)
    }
    for (const child of node.children || []) walkMatrix(child)
  }
  for (const node of planMatrix.value?.nodes || []) walkMatrix(node)
  quotaOverrideRows.value = sortQuotasByDisplayOrder(quotas.value.filter((quota) => quotaIds.has(quota.id)))

  const data = overrideData
  const byQuota = new Map<number, TenantQuotaOverride>(data.overrides.map((item) => [item.quota_id, item]))
  quotaOverrideEnabled.value = {}
  quotaOverrideValues.value = {}
  quotaOverrideReasons.value = {}
  for (const quota of quotaOverrideRows.value) {
    const override = byQuota.get(quota.id)
    quotaOverrideEnabled.value[quota.id] = !!override
    quotaOverrideValues.value[quota.id] = override?.quota_value ?? (subscriptionPlanId != null ? (matrixQuotaValues.value[subscriptionPlanId]?.[quota.id] ?? 0) : 0)
    quotaOverrideReasons.value[quota.id] = override?.reason || ''
  }
  quotaOverrideDialog.value = true
}

async function submitQuotaOverrides() {
  if (!overrideTenant.value) return
  const overrides = quotaOverrideRows.value
    .filter((quota) => quotaOverrideEnabled.value[quota.id])
    .map((quota) => ({
      quota_id: quota.id,
      quota_value: quotaOverrideValues.value[quota.id] ?? 0,
      reason: quotaOverrideReasons.value[quota.id] || null,
    }))
  await saveTenantQuotaOverrides(overrideTenant.value.id, overrides)
  quotaOverrideDialog.value = false
  ElMessage.success('配额覆盖已保存')
}

async function openQuotaUsage(row: Tenant) {
  usageTenant.value = row
  const data = await fetchTenantQuotaUsage(row.id)
  quotaUsageRows.value = data.usages
  quotaUsageDialog.value = true
}

onMounted(loadAll)
</script>

<template>
  <div class="plan-page" :data-theme="theme" v-loading="loading">
    <section class="plan-hero">
      <div class="hero-copy">
        <span class="hero-kicker">SaaS Commercial Control</span>
        <h1>套餐中心</h1>
        <p>用套餐管控租户购买能力，用权限继续管控用户操作，用配额限制资源消耗。</p>
      </div>
      <div class="hero-metrics">
        <div class="metric-block">
          <span>{{ totalPlans }}</span>
          <small>套餐</small>
        </div>
        <div class="metric-block">
          <span>{{ enabledPlanCount }}</span>
          <small>启用</small>
        </div>
        <div class="metric-block">
          <span>{{ features.length }}</span>
          <small>功能点</small>
        </div>
        <div class="metric-block">
          <span>{{ quotas.length }}</span>
          <small>配额项</small>
        </div>
      </div>
    </section>

    <section class="plan-toolbar">
      <el-input v-model="keyword" class="toolbar-search" placeholder="搜索套餐编码或名称" clearable @keyup.enter="loadAll" />
      <el-button :icon="Refresh" @click="loadAll">刷新</el-button>
      <el-button type="primary" :icon="Plus" @click="openPlan()">新增套餐</el-button>
      <el-button :icon="Grid" @click="openFeature()">功能点</el-button>
      <el-button :icon="Tickets" @click="openQuota()">配额项</el-button>
    </section>

    <section class="plan-workbench">
      <main class="config-surface" v-if="selectedPlan">
        <div class="control-strip">
          <button :class="{ active: activeTab === 'features' }" @click="activeTab = 'features'">
            <el-icon><SetUp /></el-icon>
            能力矩阵
            <b>{{ activeFeatureCount }}</b>
          </button>
          <button :class="{ active: activeTab === 'quotas' }" @click="activeTab = 'quotas'">
            <el-icon><Tickets /></el-icon>
            套餐详情
            <b>{{ configuredQuotaCount }}</b>
          </button>
          <button :class="{ active: activeTab === 'tenants' }" @click="activeTab = 'tenants'">
            <el-icon><Connection /></el-icon>
            租户权益
            <b>{{ tenants.length }}</b>
          </button>
        </div>

        <div v-if="activeTab === 'features'" class="capability-workspace">
          <PlanCapabilityMatrix
            v-if="visibleMatrix"
            :matrix="visibleMatrix"
            :feature-selection="matrixFeatureSelection"
            :quota-values="matrixQuotaValues"
            :saving-plan-id="savingMatrixPlanId"
            @toggle="toggleCapability"
            @detail="openCapabilityDetail"
            @save-plan="saveMatrixPlan"
            @plan-action="handleMatrixPlanAction"
          />
          <el-empty v-else description="暂无套餐能力矩阵" />
        </div>

        <div v-else-if="activeTab === 'quotas'" class="quota-board">
          <div class="detail-brief">
            <span>Quota Matrix</span>
            <strong>套餐配额矩阵</strong>
            <p>按能力查看关联配额，横向维护多个套餐的配额值；每列保存对应套餐。</p>
            <p class="quota-tip">填写说明：<b>-1</b> 表示不限，<b>0</b> 表示不可用。</p>
          </div>
          <div class="quota-matrix-shell">
            <div
              class="quota-matrix-row quota-matrix-row--head"
              :style="{ gridTemplateColumns: quotaMatrixColumns }"
            >
              <div class="quota-matrix-head">能力 / 配额</div>
              <div v-for="plan in visibleMatrix?.plans || []" :key="plan.id" class="quota-plan-head">
                <strong>{{ plan.plan_name }}</strong>
                <small>{{ plan.plan_code }}</small>
                <button
                  type="button"
                  class="quota-save-link"
                  :disabled="savingMatrixPlanId === plan.id"
                  @click.stop="saveMatrixPlan(plan)"
                >
                  {{ savingMatrixPlanId === plan.id ? '保存中' : '保存本列' }}
                </button>
              </div>
            </div>

            <div
              v-for="row in quotaMatrixRows"
              :key="row.id"
              class="quota-matrix-row"
              :style="{ gridTemplateColumns: quotaMatrixColumns }"
            >
              <div class="quota-capability-cell">
                <span class="quota-type-badge">{{ typeLabel(row.quota_type) }}</span>
                <div class="quota-name">
                  <strong>{{ row.quota_name }}</strong>
                  <small>{{ row.capability }} · {{ row.quota_code }} · {{ row.period_type || 'NONE' }} · {{ row.unit || 'COUNT' }}</small>
                </div>
                <el-button link :icon="EditPen" @click="openQuotaFromMatrix(row)">编辑</el-button>
              </div>
              <div v-for="plan in visibleMatrix?.plans || []" :key="plan.id" class="quota-value-cell">
                <el-input-number
                  :model-value="row.values[plan.id] ?? 0"
                  :min="-1"
                  controls-position="right"
                  @update:model-value="(value: number | undefined) => updateQuotaMatrixValue(plan.id, row.quota_id, value)"
                />
              </div>
            </div>

            <el-empty v-if="!quotaMatrixRows.length" description="当前能力矩阵暂无关联配额" />
          </div>
        </div>

        <div v-else class="tenant-bindings">
          <div v-for="tenant in tenants" :key="tenant.id" class="tenant-row">
            <div>
              <strong>{{ tenant.name }}</strong>
              <small>{{ tenant.code }}</small>
            </div>
            <el-tag
              size="small"
              round
              effect="plain"
              :type="tenant.status === 1 ? 'success' : undefined"
              :class="{ 'nm-status-pill--inactive': tenant.status !== 1 }"
            >
              {{ tenant.status === 1 ? '启用' : '停用' }}
            </el-tag>
            <el-button :icon="Link" @click="openSubscription(tenant)">绑定/更换套餐</el-button>
            <el-button :icon="SetUp" @click="openFeatureOverrides(tenant)">功能覆盖</el-button>
            <el-button :icon="Tickets" @click="openQuotaOverrides(tenant)">配额覆盖</el-button>
            <el-button :icon="Grid" @click="openQuotaUsage(tenant)">使用量</el-button>
          </div>
        </div>
      </main>

      <main v-else class="config-empty">
        <el-empty description="暂无套餐，请先新增套餐" />
      </main>
    </section>

    <PlanCapabilityDrawer
      v-model="capabilityDrawer"
      :node="capabilityDrawerNode"
      :plan="capabilityDrawerPlan"
      :enabled-feature-ids="drawerEnabledFeatureIds"
      :quota-values="drawerQuotaValues"
      @apply="applyCapabilityDetail"
    />

    <el-dialog v-model="planDialog" class="plan-mgmt-dialog" :title="editPlanId ? '编辑套餐' : '新增套餐'" width="620px">
      <el-form label-width="96px" class="dense-form">
        <el-form-item label="套餐编码"><el-input v-model="planForm.plan_code" :disabled="!!editPlanId" /></el-form-item>
        <el-form-item label="套餐名称"><el-input v-model="planForm.plan_name" /></el-form-item>
        <el-form-item label="套餐类型">
          <el-select v-model="planForm.plan_type">
            <el-option v-for="item in ['TRIAL', 'BASIC', 'PRO', 'ENTERPRISE', 'CUSTOM']" :key="item" :label="typeLabel(item)" :value="item" />
          </el-select>
        </el-form-item>
        <el-form-item label="计费周期">
          <el-select v-model="planForm.billing_cycle">
            <el-option label="月付" value="MONTH" />
            <el-option label="年付" value="YEAR" />
            <el-option label="自定义" value="CUSTOM" />
          </el-select>
        </el-form-item>
        <el-form-item label="价格"><el-input-number v-model="planForm.price" :min="0" /></el-form-item>
        <el-form-item label="状态"><el-switch v-model="planForm.status" :active-value="1" :inactive-value="0" /></el-form-item>
        <el-form-item label="默认套餐"><el-switch v-model="planForm.is_default" /></el-form-item>
        <el-form-item label="排序"><el-input-number v-model="planForm.sort_order" /></el-form-item>
        <el-form-item label="说明"><el-input v-model="planForm.description" type="textarea" :rows="3" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button :icon="Close" @click="planDialog = false">取消</el-button>
        <el-button type="primary" :icon="Check" @click="submitPlan">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="copyPlanDialog" class="plan-mgmt-dialog" title="复制套餐" width="560px">
      <el-form label-width="104px" class="dense-form">
        <el-form-item label="来源套餐"><el-input :model-value="copySourcePlan?.plan_name" disabled /></el-form-item>
        <el-form-item label="新套餐编码"><el-input v-model="copyPlanForm.plan_code" /></el-form-item>
        <el-form-item label="新套餐名称"><el-input v-model="copyPlanForm.plan_name" /></el-form-item>
        <el-form-item label="说明"><el-input v-model="copyPlanForm.description" type="textarea" :rows="3" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button :icon="Close" @click="copyPlanDialog = false">取消</el-button>
        <el-button type="primary" :icon="CopyDocument" @click="submitCopyPlan">复制</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="featureDialog" class="plan-mgmt-dialog" :title="editFeatureId ? '编辑功能点' : '新增功能点'" width="620px">
      <el-form label-width="104px" class="dense-form">
        <el-form-item label="功能编码"><el-input v-model="featureForm.feature_code" :disabled="!!editFeatureId" /></el-form-item>
        <el-form-item label="功能名称"><el-input v-model="featureForm.feature_name" /></el-form-item>
        <el-form-item label="功能类型">
          <el-select v-model="featureForm.feature_type">
            <el-option v-for="item in ['MENU', 'BUTTON', 'API', 'SERVICE', 'CONFIG']" :key="item" :label="typeLabel(item)" :value="item" />
          </el-select>
        </el-form-item>
        <el-form-item label="API 方法"><el-input v-model="featureForm.api_method" /></el-form-item>
        <el-form-item label="API 路径"><el-input v-model="featureForm.api_path" /></el-form-item>
        <el-form-item label="服务标识"><el-input v-model="featureForm.service_key" /></el-form-item>
        <el-form-item label="状态"><el-switch v-model="featureForm.status" :active-value="1" :inactive-value="0" /></el-form-item>
        <el-form-item label="说明"><el-input v-model="featureForm.description" type="textarea" :rows="3" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button :icon="Close" @click="featureDialog = false">取消</el-button>
        <el-button type="primary" :icon="Check" @click="submitFeature">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="quotaDialog" class="plan-mgmt-dialog" :title="editQuotaId ? '编辑配额' : '新增配额'" width="620px">
      <el-form label-width="96px" class="dense-form">
        <el-form-item label="配额编码"><el-input v-model="quotaForm.quota_code" :disabled="!!editQuotaId" /></el-form-item>
        <el-form-item label="配额名称"><el-input v-model="quotaForm.quota_name" /></el-form-item>
        <el-form-item label="配额类型">
          <el-select v-model="quotaForm.quota_type">
            <el-option label="静态配额" value="STATIC" />
            <el-option label="动态配额" value="DYNAMIC" />
          </el-select>
        </el-form-item>
        <el-form-item label="周期"><el-input v-model="quotaForm.period_type" /></el-form-item>
        <el-form-item label="单位"><el-input v-model="quotaForm.unit" /></el-form-item>
        <el-form-item label="状态"><el-switch v-model="quotaForm.status" :active-value="1" :inactive-value="0" /></el-form-item>
        <el-form-item label="说明"><el-input v-model="quotaForm.description" type="textarea" :rows="3" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button :icon="Close" @click="quotaDialog = false">取消</el-button>
        <el-button type="primary" :icon="Check" @click="submitQuota">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="featureOverrideDialog" class="plan-mgmt-dialog" title="租户功能覆盖" width="860px">
      <div class="override-title">
        <strong>{{ overrideTenant?.name }}</strong>
        <small>继承表示使用当前订阅套餐配置；启用/禁用会覆盖套餐功能矩阵。</small>
      </div>
      <el-table :data="features" class="quota-table override-table" row-key="id" max-height="520">
        <el-table-column prop="feature_name" label="功能" min-width="180">
          <template #default="{ row }">
            <div class="quota-name">
              <strong>{{ row.feature_name }}</strong>
              <small>{{ row.feature_code }}</small>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="覆盖策略" width="170">
          <template #default="{ row }">
            <el-select v-model="featureOverrideStates[row.id]">
              <el-option label="继承套餐" value="inherit" />
              <el-option label="强制启用" value="enable" />
              <el-option label="强制禁用" value="disable" />
            </el-select>
          </template>
        </el-table-column>
        <el-table-column label="原因" min-width="220">
          <template #default="{ row }">
            <el-input v-model="featureOverrideReasons[row.id]" :disabled="featureOverrideStates[row.id] === 'inherit'" />
          </template>
        </el-table-column>
      </el-table>
      <template #footer>
        <el-button :icon="Close" @click="featureOverrideDialog = false">取消</el-button>
        <el-button type="primary" :icon="Check" @click="submitFeatureOverrides">保存覆盖</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="quotaOverrideDialog" class="plan-mgmt-dialog" title="租户配额覆盖" width="940px">
      <div class="override-title">
        <strong>{{ overrideTenant?.name }}</strong>
        <small>仅展示当前套餐功能矩阵实际启用后关联的配额；-1 表示无限制，0 表示不可用。</small>
      </div>
      <el-table :data="quotaOverrideRows" class="quota-table override-table" row-key="id" max-height="520">
        <el-table-column label="覆盖" width="80">
          <template #default="{ row }">
            <el-switch v-model="quotaOverrideEnabled[row.id]" />
          </template>
        </el-table-column>
        <el-table-column prop="quota_name" label="配额" min-width="180">
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
            <el-input-number v-model="quotaOverrideValues[row.id]" :disabled="!quotaOverrideEnabled[row.id]" :min="-1" controls-position="right" />
          </template>
        </el-table-column>
        <el-table-column label="原因" min-width="220">
          <template #default="{ row }">
            <el-input v-model="quotaOverrideReasons[row.id]" :disabled="!quotaOverrideEnabled[row.id]" />
          </template>
        </el-table-column>
      </el-table>
      <template #footer>
        <el-button :icon="Close" @click="quotaOverrideDialog = false">取消</el-button>
        <el-button type="primary" :icon="Check" @click="submitQuotaOverrides">保存覆盖</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="quotaUsageDialog" class="plan-mgmt-dialog" title="租户配额使用" width="940px">
      <div class="override-title">
        <strong>{{ usageTenant?.name }}</strong>
        <small>展示当前套餐与租户覆盖后的限制值、已用值和剩余额度。</small>
      </div>
      <el-table :data="quotaUsageRows" class="quota-table override-table" row-key="quota_id" max-height="520">
        <el-table-column prop="quota_name" label="配额" min-width="190">
          <template #default="{ row }">
            <div class="quota-name">
              <strong>{{ row.quota_name }}</strong>
              <small>{{ row.quota_code }}</small>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="period_key" label="周期" width="120" />
        <el-table-column prop="unit" label="单位" width="90" />
        <el-table-column label="限制" width="110">
          <template #default="{ row }">{{ quotaText(row.limit_value) }}</template>
        </el-table-column>
        <el-table-column prop="used_value" label="已用" width="110" />
        <el-table-column label="剩余" width="110">
          <template #default="{ row }">{{ row.remaining_value === null ? '无限制' : row.remaining_value }}</template>
        </el-table-column>
      </el-table>
      <template #footer>
        <el-button :icon="Close" @click="quotaUsageDialog = false">关闭</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="subscriptionDialog" class="plan-mgmt-dialog" title="租户套餐绑定" width="620px">
      <el-form label-width="104px" class="dense-form">
        <el-form-item label="租户"><el-input :model-value="subscriptionTenant?.name" disabled /></el-form-item>
        <el-form-item label="套餐">
          <el-select v-model="subscriptionForm.plan_id">
            <el-option v-for="plan in plans" :key="plan.id" :label="`${plan.plan_name} / ${plan.plan_code}`" :value="plan.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="订阅状态">
          <el-select v-model="subscriptionForm.subscription_status">
            <el-option v-for="item in ['TRIAL', 'ACTIVE', 'OVERDUE', 'FROZEN', 'EXPIRED', 'CANCELLED']" :key="item" :label="item" :value="item" />
          </el-select>
        </el-form-item>
        <el-form-item label="开始时间"><el-date-picker v-model="subscriptionForm.start_time" type="datetime" value-format="YYYY-MM-DDTHH:mm:ss" /></el-form-item>
        <el-form-item label="结束时间"><el-date-picker v-model="subscriptionForm.end_time" type="datetime" value-format="YYYY-MM-DDTHH:mm:ss" /></el-form-item>
        <el-form-item label="自动续费"><el-switch v-model="subscriptionForm.auto_renew" /></el-form-item>
        <el-form-item label="冻结原因"><el-input v-model="subscriptionForm.frozen_reason" type="textarea" :rows="2" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button :icon="Close" @click="subscriptionDialog = false">取消</el-button>
        <el-button type="primary" :icon="Check" @click="submitSubscription">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.plan-page {
  /* 底层与主布局 --nm-bg-surface 一致，避免比内容区更暗的「垫底黑块」（原 min-height:100% + 深色纯色造成） */
  --plan-page-bg: var(--nm-bg-surface, #111827);
  --plan-panel-bg: rgba(15, 27, 43, 0.92);
  --plan-panel-strong: rgba(20, 37, 55, 0.96);
  --plan-panel-soft: rgba(18, 35, 51, 0.78);
  --plan-panel-grad: linear-gradient(180deg, rgba(11, 28, 45, 0.88), rgba(12, 24, 39, 0.92));
  --plan-panel-soft-grad: linear-gradient(180deg, rgba(16, 36, 54, 0.76), rgba(15, 30, 46, 0.8));
  --plan-text: #e7f3f7;
  --plan-text-strong: #f8fdff;
  --plan-text-muted: #92a6b4;
  --plan-border: rgba(0, 245, 212, 0.16);
  --plan-border-strong: rgba(0, 245, 212, 0.34);
  --plan-accent: #00f5d4;
  --plan-accent-strong: #22d3ee;
  --plan-warning: #f59e0b;
  /* 与其他业务页（主体管理等）对齐的圆角刻度 */
  --plan-radius-sm: var(--neuro-radius-md, 8px);
  --plan-radius-md: var(--neuro-radius-lg, 12px);
  --plan-radius-lg: var(--neuro-radius-xl, 16px);
  box-sizing: border-box;
  width: 100%;
  min-height: 100%;
  padding: 8px 0 24px;
  color: var(--plan-text);
  background: var(--plan-page-bg);
  /* 与全局 body / admin 主内容一致，勿单独使用西文展示字体栈 */
  font-family: var(--nm-font-body);
  overflow-x: hidden;
}

.plan-page[data-theme="light"] {
  --plan-page-bg: var(--nm-bg-surface, #ffffff);
  --plan-panel-bg: rgba(255, 255, 255, 0.88);
  --plan-panel-strong: rgba(255, 255, 255, 0.96);
  --plan-panel-soft: #f8faf6;
  --plan-panel-grad: linear-gradient(180deg, rgba(255, 255, 255, 0.96), rgba(249, 252, 255, 0.92));
  --plan-panel-soft-grad: linear-gradient(180deg, rgba(250, 252, 248, 0.92), rgba(246, 250, 245, 0.9));
  --plan-text: #162029;
  --plan-text-strong: #14222b;
  --plan-text-muted: #58646d;
  --plan-border: rgba(35, 49, 59, 0.12);
  --plan-border-strong: rgba(49, 125, 134, 0.55);
  --plan-accent: #1eb7a6;
  --plan-accent-strong: #317d86;
  --plan-warning: #d88319;
}

.plan-page *,
.plan-page *::before,
.plan-page *::after {
  box-sizing: border-box;
}

.plan-hero,
.plan-toolbar,
.plan-workbench {
  width: min(100%, 100%);
  max-width: none;
  margin: 0 auto 14px;
}

.plan-hero {
  display: flex;
  align-items: stretch;
  justify-content: space-between;
  gap: 20px;
  padding: 24px 28px;
  border: 1px solid var(--plan-border);
  border-radius: var(--plan-radius-lg);
  background: var(--plan-panel-grad);
  overflow: hidden;
}

.hero-kicker {
  display: inline-block;
  margin-bottom: 8px;
  color: var(--plan-warning);
  font-size: 12px;
  font-weight: 800;
  letter-spacing: 0;
  text-transform: uppercase;
}

.hero-copy h1 {
  margin: 0;
  color: var(--plan-text-strong);
  font-size: 34px;
  line-height: 1.1;
}

.hero-copy p {
  max-width: 620px;
  margin: 10px 0 0;
  color: var(--plan-text-muted);
}

.hero-metrics {
  display: grid;
  grid-template-columns: repeat(4, minmax(82px, 1fr));
  gap: 10px;
  width: min(440px, 40%);
  min-width: 360px;
  flex-shrink: 0;
}

.metric-block {
  display: grid;
  place-items: center;
  min-height: 92px;
  border: 1px solid var(--plan-border);
  border-radius: var(--plan-radius-md);
  background: var(--plan-panel-soft-grad);
}

.metric-block span {
  font-size: 28px;
  font-weight: 900;
}

.metric-block small {
  color: var(--plan-text-muted);
}

.plan-toolbar {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px;
  border: 1px solid var(--plan-border);
  border-radius: var(--plan-radius-lg);
  background: var(--plan-panel-grad);
}

.toolbar-search {
  max-width: 360px;
}

.plan-toolbar :deep(.el-button) {
  min-height: 38px;
  border-radius: var(--plan-radius-sm);
  border-color: var(--plan-border);
  background: var(--plan-panel-bg);
  color: var(--plan-text-strong);
  font-weight: 700;
}

.plan-toolbar :deep(.el-button--primary) {
  border-radius: var(--plan-radius-sm);
  border-color: var(--plan-accent);
  background: var(--plan-accent);
  color: #06211e;
}

.plan-toolbar :deep(.el-input__wrapper) {
  border-radius: var(--plan-radius-sm);
}

.plan-workbench {
  display: block;
  min-width: 0;
}

.surface-label {
  color: var(--plan-text-muted);
  font-size: 12px;
  font-weight: 800;
  letter-spacing: 0;
}

.config-surface,
.config-empty {
  min-height: 640px;
  border: 1px solid var(--plan-border);
  border-radius: var(--plan-radius-lg);
  background: var(--plan-panel-grad);
  min-width: 0;
  overflow: hidden;
}

.surface-head {
  display: flex;
  justify-content: space-between;
  gap: 18px;
  padding: 24px;
  border-bottom: 1px solid var(--plan-border);
}

.surface-head h2 {
  margin: 4px 0;
  color: var(--plan-text-strong);
  font-size: 28px;
}

.surface-head p {
  margin: 0;
  color: var(--plan-text-muted);
}

.tenant-row :deep(.el-button),
.save-bar :deep(.el-button) {
  border-radius: var(--plan-radius-sm);
  border-color: var(--plan-border);
  background: var(--plan-panel-strong);
  color: var(--plan-text-strong);
  font-weight: 700;
}

.save-bar :deep(.el-button--primary) {
  border-radius: var(--plan-radius-sm);
  border-color: var(--plan-accent);
  background: var(--plan-accent);
  color: #06211e;
}

.control-strip {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  border-bottom: 1px solid var(--plan-border);
}

.control-strip button {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  min-height: 58px;
  border: 0;
  border-right: 1px solid var(--plan-border);
  background: var(--plan-panel-soft);
  /* 未选中须明显高于面板底色，避免与 muted 灰糊成一团 */
  color: var(--plan-text);
  font-weight: 800;
  cursor: pointer;
  transition: background 0.15s ease, color 0.15s ease;
}

.control-strip button :deep(.el-icon) {
  color: inherit;
  font-size: 18px;
}

.control-strip button:hover:not(.active) {
  background: var(--plan-panel-strong);
  color: var(--plan-text-strong);
}

.control-strip button.active {
  background: var(--plan-accent);
  color: #06211e;
}

.control-strip b {
  min-width: 26px;
  padding: 2px 8px;
  border-radius: var(--plan-radius-sm);
  font-weight: 900;
  line-height: 1.2;
}

.control-strip button:not(.active) b {
  color: var(--plan-warning);
  background: rgba(245, 158, 11, 0.14);
  border: 1px solid rgba(245, 158, 11, 0.35);
}

.control-strip button.active b {
  color: #06211e;
  background: rgba(6, 33, 30, 0.18);
  border: 1px solid rgba(6, 33, 30, 0.28);
}

.feature-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px;
  padding: 18px;
  min-width: 0;
}

.capability-workspace {
  padding: 18px;
  min-width: 0;
}

.feature-section {
  border: 1px solid var(--plan-border);
  border-radius: var(--plan-radius-md);
  background: var(--plan-panel-soft);
  overflow: hidden;
}

.section-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 14px;
  border-bottom: 1px solid var(--plan-border);
  color: var(--plan-text-strong);
  font-weight: 900;
}

.section-title small {
  color: var(--plan-text-muted);
}

.feature-list {
  display: grid;
}

.feature-row {
  display: grid;
  grid-template-columns: 24px minmax(160px, 1fr) 70px 52px;
  align-items: center;
  gap: 10px;
  padding: 12px 14px;
  border-bottom: 1px solid var(--plan-border);
  color: var(--plan-text);
}

.feature-row:last-child {
  border-bottom: 0;
}

.feature-main {
  display: grid;
  min-width: 0;
  color: var(--plan-text);
  line-height: 1.35;
}

.feature-main strong {
  display: block;
  font-size: 14px;
}

.feature-main small {
  display: block;
  margin-top: 2px;
  font-size: 12px;
}

.feature-main strong,
.quota-name strong,
.tenant-row strong {
  color: var(--plan-text);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.feature-main small,
.quota-name small,
.tenant-row small {
  color: var(--plan-text-muted);
}

.feature-row :deep(.el-checkbox) {
  height: 18px;
}

.feature-row :deep(.el-checkbox__label) {
  display: none;
}

.quota-board,
.tenant-bindings {
  padding: 18px;
  min-width: 0;
}

.detail-brief {
  display: grid;
  gap: 4px;
  margin-bottom: 14px;
  padding: 18px;
  border: 1px solid var(--plan-border);
  border-radius: var(--plan-radius-md);
  background:
    radial-gradient(circle at 10% 0, color-mix(in srgb, var(--plan-accent) 18%, transparent), transparent 34%),
    var(--plan-panel-soft);
}

.detail-brief span {
  color: var(--plan-warning);
  font-size: 12px;
  font-weight: 900;
  letter-spacing: 0;
  text-transform: uppercase;
}

.detail-brief strong {
  color: var(--plan-text-strong);
  font-size: 22px;
}

.detail-brief p {
  margin: 0;
  color: var(--plan-text-muted);
}

.quota-tip {
  font-size: 12px;
}

.quota-table {
  --el-table-bg-color: var(--plan-panel-bg);
  --el-table-tr-bg-color: var(--plan-panel-bg);
  --el-table-header-bg-color: var(--plan-panel-soft);
  --el-table-border-color: var(--plan-border);
  --el-table-text-color: var(--plan-text);
  --el-table-header-text-color: var(--plan-text-strong);
  --el-table-row-hover-bg-color: color-mix(in srgb, var(--plan-accent) 10%, transparent);
  border: 1px solid var(--plan-border);
  border-radius: var(--plan-radius-md);
  background: var(--plan-panel-bg);
  color: var(--plan-text);
  overflow: hidden;
}

.quota-matrix-shell {
  overflow: auto;
  border: 1px solid var(--plan-border);
  border-radius: var(--plan-radius-md);
  background: var(--plan-panel-bg);
}

.quota-matrix-row {
  display: grid;
  width: max-content;
  min-width: 100%;
  border-bottom: 1px solid var(--plan-border);
}

.quota-matrix-row:last-child {
  border-bottom: none;
}

.quota-matrix-row--head {
  position: sticky;
  top: 0;
  z-index: 2;
  background: var(--plan-panel-strong);
}

.quota-matrix-head,
.quota-plan-head,
.quota-capability-cell,
.quota-value-cell {
  padding: 14px 16px;
  border-right: 1px solid var(--plan-border);
}

.quota-plan-head {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.quota-plan-head small,
.quota-name small {
  color: var(--plan-text-muted);
}

.quota-save-link {
  position: relative;
  z-index: 4;
  align-self: flex-start;
  border: none;
  background: none;
  color: var(--plan-accent);
  cursor: pointer;
  font-weight: 800;
  padding: 0;
}

.quota-save-link:disabled {
  color: var(--plan-text-muted);
  cursor: default;
}

.quota-capability-cell {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  align-items: center;
  gap: 10px;
}

.quota-capability-cell :deep(.el-button) {
  grid-column: 2;
  justify-self: start;
  padding: 0;
}

.quota-type-badge {
  padding: 4px 8px;
  border-radius: 999px;
  background: color-mix(in srgb, var(--plan-accent) 14%, transparent);
  color: var(--plan-accent);
  font-size: 12px;
  font-weight: 900;
}

.quota-value-cell {
  display: flex;
  align-items: center;
  justify-content: center;
}

.quota-value-cell :deep(.el-input-number) {
  width: 128px;
}

.quota-value-cell :deep(.el-input__wrapper) {
  border-radius: var(--plan-radius-sm);
}

.quota-table :deep(.el-table__inner-wrapper::before),
.quota-table :deep(.el-table__border-left-patch) {
  background: var(--plan-border);
}

.quota-table :deep(.el-table__header th) {
  background: var(--plan-panel-soft);
  color: var(--plan-text-strong);
}

.quota-table :deep(.el-table__body tr),
.quota-table :deep(.el-table__body td) {
  background: var(--plan-panel-bg);
  color: var(--plan-text);
}

.quota-table :deep(.el-table__body tr:hover > td) {
  background: color-mix(in srgb, var(--plan-accent) 10%, var(--plan-panel-bg));
}

.quota-name {
  display: grid;
}

.quota-input {
  display: flex;
  align-items: center;
  gap: 10px;
}

.quota-input span {
  min-width: 64px;
  color: var(--plan-warning);
  font-weight: 800;
}

.tenant-bindings {
  display: grid;
  gap: 10px;
}

.tenant-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 86px 150px 120px 120px 104px;
  align-items: center;
  gap: 14px;
  padding: 14px;
  border: 1px solid var(--plan-border);
  border-radius: var(--plan-radius-md);
  background: var(--plan-panel-soft);
}

.tenant-row > div {
  display: grid;
}

.override-title {
  display: grid;
  gap: 4px;
  margin-bottom: 12px;
  color: var(--plan-text);
}

.override-title small {
  color: var(--plan-text-muted);
}

.override-table {
  width: 100%;
}

.save-bar {
  grid-column: 1 / -1;
  display: flex;
  justify-content: flex-end;
  padding: 10px 0 0;
}

.config-empty {
  display: grid;
  place-items: center;
}

.plan-page :deep(.el-input__wrapper),
.plan-page :deep(.el-textarea__inner),
.plan-page :deep(.el-select__wrapper) {
  border-radius: var(--plan-radius-sm);
  background: var(--plan-panel-strong);
  box-shadow: 0 0 0 1px var(--plan-border) inset;
}

.plan-page :deep(.el-input__inner),
.plan-page :deep(.el-textarea__inner) {
  color: var(--plan-text);
}

.plan-page :deep(.el-checkbox__label) {
  color: var(--plan-text);
}

.dense-form :deep(.el-select),
.dense-form :deep(.el-date-editor.el-input),
.dense-form :deep(.el-input-number) {
  width: 100%;
}

@media (max-width: 1180px) {
  .plan-hero {
    flex-direction: column;
  }
  .hero-metrics {
    min-width: 0;
  }
}

@media (max-width: 760px) {
  .plan-page {
    padding: 12px 0;
  }
  .hero-metrics,
  .feature-grid {
    grid-template-columns: 1fr;
  }
  .plan-toolbar,
  .surface-head {
    flex-direction: column;
    align-items: stretch;
  }
  .control-strip {
    grid-template-columns: 1fr;
  }
  .feature-row,
  .tenant-row {
    grid-template-columns: 1fr;
  }
}
</style>

<style>
/* Teleported `el-dialog`：scoped 无法命中，与后台通用 Neuro 圆角对齐 */
html.admin-shell .plan-mgmt-dialog.el-dialog {
  border-radius: var(--neuro-radius-xl, 16px);
  overflow: hidden;
}
</style>
