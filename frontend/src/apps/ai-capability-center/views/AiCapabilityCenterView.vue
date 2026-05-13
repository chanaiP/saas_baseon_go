<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Delete, Edit, Plus, Refresh, Search, Upload } from '@element-plus/icons-vue'

import NeuroAgentPageShell from '@/views/components/NeuroAgentPageShell.vue'

import { createAiResource, deleteAiResource, fetchAiOverview, fetchAiResource, importAiModels, importAiProviders, importAiRoutes, importAiScenarios, updateAiResource } from '../api'
import type { AiModelImportPayload, AiOverview, AiPage, AiProviderImportPayload, AiResource, AiRouteImportPayload, AiScenarioImportPayload, AiSectionConfig } from '../types'

defineOptions({ name: 'AiCapabilityCenterView' })

const route = useRoute()
const router = useRouter()

type FieldConfig = {
  key: string
  label: string
  type?: 'text' | 'number' | 'textarea' | 'select' | 'tags' | 'json' | 'switch'
  options?: Array<{ label: string; value: string }>
  required?: boolean
}

function toOptions(values: string[]) {
  return values.map((value) => ({ label: value, value }))
}

const sections: AiSectionConfig[] = [
  { key: 'dashboard', title: '总览', route: '/ai-capability-center', description: '平台调用、成本、成功率、租户排行榜和健康检查。', columns: [] },
  { key: 'providers', title: '供应商', route: '/ai-capability-center/providers', resource: 'providers', description: '维护供应商、接入资源、预算、区域和负责人。', writable: true, columns: [
    { key: 'name', label: '供应商', width: 180 }, { key: 'code', label: '编码', width: 160 }, { key: 'type', label: '类型', width: 120 }, { key: 'base_url', label: 'Endpoint' }, { key: 'auth_type', label: '鉴权', width: 120 }, { key: 'qps_limit', label: 'QPS', width: 90 }, { key: 'monthly_budget', label: '月预算', width: 120 }, { key: 'owner', label: '负责人', width: 120 }, { key: 'status', label: '状态', width: 100 },
  ] },
  { key: 'models', title: '模型目录', route: '/ai-capability-center/models', resource: 'models', description: '按供应商维护模型、能力标签、质量指标和默认场景。', writable: true, columns: [
    { key: 'model_name', label: '模型', width: 200 }, { key: 'model_code', label: '模型编码', width: 180 }, { key: 'model_type', label: '类型', width: 120 }, { key: 'capabilities', label: '能力' }, { key: 'context_window', label: '上下文', width: 110 }, { key: 'success_rate', label: '成功率', width: 100 }, { key: 'latency_p95', label: 'P95', width: 100 }, { key: 'status', label: '状态', width: 100 },
  ] },
  { key: 'scenarios', title: 'AI 场景', route: '/ai-capability-center/scenarios', resource: 'scenarios', description: '强制注册 app_code + ai_scenario_code，并绑定默认基础路由。', writable: true, columns: [
    { key: 'ai_scenario_name', label: '场景', width: 200 }, { key: 'ai_scenario_code', label: '场景编码', width: 200 }, { key: 'app_name', label: '应用', width: 160 }, { key: 'app_code', label: '应用编码', width: 160 }, { key: 'scenario_type', label: '场景类型', width: 120 }, { key: 'capability_code', label: '能力', width: 160 }, { key: 'default_base_route_id', label: '默认基础路由', width: 240 }, { key: 'owner', label: '负责人', width: 120 }, { key: 'status', label: '状态', width: 100 },
  ] },
  { key: 'routes', title: '基础路由', route: '/ai-capability-center/routes', resource: 'base-routes', description: '维护可复用模型池和路由策略，不绑定租户和具体场景。', writable: true, columns: [
    { key: 'route_name', label: '路由', width: 220 }, { key: 'route_code', label: '路由编码', width: 200 }, { key: 'capability_code', label: '能力', width: 160 }, { key: 'model_type', label: '模型类型', width: 120 }, { key: 'strategy', label: '策略', width: 140 }, { key: 'timeout_ms', label: '超时', width: 100 }, { key: 'max_retry', label: '重试', width: 90 }, { key: 'status', label: '状态', width: 100 },
  ] },
  { key: 'strategy', title: '策略中心', route: '/ai-capability-center/strategy', resource: 'tenant-strategies', description: '一条租户策略统一表达路由覆盖、多维配额、多维限流和超限动作。', writable: true, columns: [
    { key: 'policy_name', label: '策略', width: 220 }, { key: 'tenant_scope', label: '租户范围', width: 130 }, { key: 'tenant_ids', label: '租户', width: 180 }, { key: 'app_code', label: '应用', width: 150 }, { key: 'ai_scenario_code', label: 'AI 场景', width: 190 }, { key: 'override_base_route_id', label: '覆盖路由', width: 240 }, { key: 'status', label: '状态', width: 100 },
  ] },
  { key: 'settings', title: '系统设置', route: '/ai-capability-center/settings', resource: 'settings', description: '维护网关参数、能力字典基线和安全归属。', writable: true, columns: [
    { key: 'setting_key', label: '配置项', width: 220 }, { key: 'setting_value', label: '配置值' }, { key: 'description', label: '说明', width: 300 }, { key: 'status', label: '状态', width: 100 },
  ] },
]

const activeSection = computed(() => sections.find((item) => item.route === route.path) ?? sections[0])
const overview = ref<AiOverview | null>(null)
const page = ref<AiPage<Record<string, unknown>>>({ items: [], total: 0, skip: 0, limit: 20 })
const allProvidersPage = ref<AiPage<Record<string, unknown>>>({ items: [], total: 0, skip: 0, limit: 200 })
const providerAccountsPage = ref<AiPage<Record<string, unknown>>>({ items: [], total: 0, skip: 0, limit: 200 })
const providerApisPage = ref<AiPage<Record<string, unknown>>>({ items: [], total: 0, skip: 0, limit: 200 })
const pricePoliciesPage = ref<AiPage<Record<string, unknown>>>({ items: [], total: 0, skip: 0, limit: 200 })
const priceTiersPage = ref<AiPage<Record<string, unknown>>>({ items: [], total: 0, skip: 0, limit: 200 })
const capabilitiesPage = ref<AiPage<Record<string, unknown>>>({ items: [], total: 0, skip: 0, limit: 200 })
const baseRoutesPage = ref<AiPage<Record<string, unknown>>>({ items: [], total: 0, skip: 0, limit: 200 })
const routeModelsPage = ref<AiPage<Record<string, unknown>>>({ items: [], total: 0, skip: 0, limit: 200 })
const allModelsPage = ref<AiPage<Record<string, unknown>>>({ items: [], total: 0, skip: 0, limit: 200 })
const scenariosPage = ref<AiPage<Record<string, unknown>>>({ items: [], total: 0, skip: 0, limit: 200 })
const tenantStrategiesPage = ref<AiPage<Record<string, unknown>>>({ items: [], total: 0, skip: 0, limit: 200 })
const currentPage = computed(() => Math.floor(page.value.skip / Math.max(page.value.limit, 1)) + 1)
const keyword = ref('')
const loading = ref(false)
const errorText = ref('')
const editorVisible = ref(false)
const editorMode = ref<'create' | 'edit'>('create')
const editorResource = ref<AiResource | undefined>()
const editorJson = ref('{}')
const formModel = ref<Record<string, unknown>>({})
const editingId = ref('')
const selectedProviderId = ref('')
const selectedAccountId = ref('')
const selectedModelId = ref('')
const selectedPolicyId = ref('')
const selectedAppCode = ref('')
const selectedBaseRouteId = ref('')
const importVisible = ref(false)
const modelImportVisible = ref(false)
const scenarioImportVisible = ref(false)
const routeImportVisible = ref(false)
const importJson = ref(JSON.stringify({
  providers: [{
    name: 'OpenAI',
    code: 'openai',
    type: 'public_cloud',
    base_url: 'https://api.openai.com',
    auth_type: 'api_key',
    accounts: [{
      account_name: 'prod',
      endpoint: 'https://api.openai.com',
      key_alias: 'OPENAI_API_KEY',
      encrypted_api_key: 'ciphertext',
      apis: [{
        api_name: 'chat.completions',
        api_path: '/v1/chat/completions',
        api_type: 'chat',
        capabilities: ['chat_completion'],
        timeout_ms: 30000,
      }],
    }],
  }],
}, null, 2))
const modelImportJson = ref(JSON.stringify({
  models: [{
    provider_code: 'openai',
    model_code: 'gpt-4.1',
    model_name: 'GPT 4.1',
    model_type: 'text',
    capabilities: ['chat_completion'],
    context_window: 128000,
    unit: 'tokens',
    price_policies: [{
      feature_key: 'chat_tokens',
      feature_name: '对话 Token',
      model_type: 'text',
      capability_code: 'chat_completion',
      billing_mode: 'tiered',
      billing_unit: 'tokens',
      platform_unit: 'tokens',
      base_cost_price: 0.01,
      base_sale_price: 0.02,
      tiers: [{
        tier_name: 'standard',
        mode: 'sync',
        cost_price: 0.01,
        sale_price: 0.02,
        platform_amount: 0.01,
        enabled: true,
      }],
    }],
  }],
}, null, 2))
const scenarioImportJson = ref(JSON.stringify({
  scenarios: [{
    app_code: 'product_center',
    app_name: '商品中心',
    ai_scenario_code: 'product_copy_generate',
    ai_scenario_name: '商品文案生成',
    scenario_type: 'text',
    capability_code: 'chat_completion',
    model_type: 'text',
    default_base_route_id: 'route-uuid',
    owner: '商品平台组',
    version: 'v1.0',
  }],
}, null, 2))
const routeImportJson = ref(JSON.stringify({
  base_routes: [{
    route_code: 'chat-default',
    route_name: '对话默认路由',
    capability_code: 'chat_completion',
    model_type: 'text',
    strategy: 'fallback',
    timeout_ms: 30000,
    max_retry: 2,
    route_models: [{
      provider_code: 'openai',
      model_code: 'gpt-4.1',
      role: 'primary',
      priority: 1,
      weight: 100,
      max_retry: 1,
      timeout_ms: 25000,
    }],
  }],
}, null, 2))

const canWrite = computed(() => Boolean(activeSection.value.resource && activeSection.value.writable))
const formFields = computed<FieldConfig[]>(() => fieldsForResource(editorResource.value ?? activeSection.value.resource))
const providerRows = computed(() => activeSection.value.key === 'providers' ? page.value.items : allProvidersPage.value.items)
const selectedProviderAccounts = computed(() => providerAccountsPage.value.items.filter((item) => String(item.provider_id || '') === selectedProviderId.value))
const selectedAccountApis = computed(() => providerApisPage.value.items.filter((item) => {
  const matchesProvider = String(item.provider_id || '') === selectedProviderId.value
  const matchesAccount = !selectedAccountId.value || String(item.account_id || '') === selectedAccountId.value
  return matchesProvider && matchesAccount
}))
const selectedProviderModels = computed(() => page.value.items.filter((item) => String(item.provider_id || '') === selectedProviderId.value))
const selectedModelPolicies = computed(() => pricePoliciesPage.value.items.filter((item) => String(item.model_id || '') === selectedModelId.value))
const selectedPolicyTiers = computed(() => priceTiersPage.value.items.filter((item) => String(item.price_policy_id || '') === selectedPolicyId.value))
const activeProviderName = computed(() => providerRows.value.find((item) => String(item.id || '') === selectedProviderId.value)?.name || '全部供应商')
const capabilityOptions = computed(() => capabilitiesPage.value.items.map((item) => ({ label: String(item.capability_name || item.capability_code), value: String(item.capability_code || '') })).filter((item) => item.value))
const baseRouteOptions = computed(() => baseRoutesPage.value.items.map((item) => ({ label: `${item.route_name || item.route_code} · ${item.capability_code || '-'}`, value: String(item.id || '') })).filter((item) => item.value))
const modelOptions = computed(() => allModelsPage.value.items.map((item) => ({ label: `${item.model_name || item.model_code} · ${item.model_type || '-'}`, value: String(item.id || '') })).filter((item) => item.value))
const selectedBaseRoute = computed(() => page.value.items.find((item) => String(item.id || '') === selectedBaseRouteId.value))
const selectedRouteModels = computed(() => routeModelsPage.value.items.filter((item) => String(item.base_route_id || '') === selectedBaseRouteId.value))
const appGroups = computed(() => {
  const groups = new Map<string, { app_code: string; app_name: string; total: number; active: number }>()
  for (const item of page.value.items) {
    const appCode = String(item.app_code || '')
    if (!appCode) continue
    const group = groups.get(appCode) || { app_code: appCode, app_name: String(item.app_name || appCode), total: 0, active: 0 }
    group.total++
    if (item.status === 'active') group.active++
    groups.set(appCode, group)
  }
  return Array.from(groups.values()).sort((a, b) => a.app_code.localeCompare(b.app_code))
})
const selectedAppScenarios = computed(() => page.value.items.filter((item) => String(item.app_code || '') === selectedAppCode.value))

function displayCell(row: Record<string, unknown>, key: string) {
  const value = row[key]
  if (Array.isArray(value)) return value.join('、') || '-'
  if (value === null || value === undefined || value === '') return '-'
  if (typeof value === 'object') return JSON.stringify(value)
  return String(value)
}

function statusType(value: unknown) {
  if (value === 'active' || value === 'success') return 'success'
  if (value === 'warning') return 'warning'
  if (value === 'error' || value === 'failed' || value === 'timeout') return 'danger'
  return 'info'
}

function numberText(value: unknown) {
  const numeric = Number(value ?? 0)
  return Number.isFinite(numeric) ? numeric.toLocaleString('zh-CN') : '0'
}

function moneyText(value: unknown) {
  const numeric = Number(value ?? 0)
  return Number.isFinite(numeric) ? `¥${numeric.toFixed(2)}` : '¥0.00'
}

function strategyCount(row: Record<string, unknown>) {
  const appCode = String(row.app_code || '')
  const scenarioCode = String(row.ai_scenario_code || '')
  return tenantStrategiesPage.value.items.filter((item) => String(item.app_code || '') === appCode && String(item.ai_scenario_code || '') === scenarioCode).length
}

function modelName(modelId: unknown) {
  const model = allModelsPage.value.items.find((item) => String(item.id || '') === String(modelId || ''))
  return String(model?.model_name || model?.model_code || modelId || '-')
}

function routeReferenceCount(routeId: unknown) {
  const id = String(routeId || '')
  const scenarios = scenariosPage.value.items.filter((item) => String(item.default_base_route_id || '') === id).length
  const strategies = tenantStrategiesPage.value.items.filter((item) => String(item.default_base_route_id || '') === id || String(item.override_base_route_id || '') === id).length
  return scenarios + strategies
}

async function loadData() {
  loading.value = true
  errorText.value = ''
  try {
    if (!activeSection.value.resource) {
      overview.value = await fetchAiOverview()
      return
    }
    page.value = await fetchAiResource(activeSection.value.resource, {
      skip: page.value.skip,
      limit: page.value.limit,
      keyword: keyword.value.trim(),
    })
    if (activeSection.value.key === 'providers') {
      providerAccountsPage.value = await fetchAiResource('accounts', { skip: 0, limit: 200 })
      providerApisPage.value = await fetchAiResource('apis', { skip: 0, limit: 200 })
      reconcileProviderSelection()
    }
    if (activeSection.value.key === 'models') {
      allProvidersPage.value = await fetchAiResource('providers', { skip: 0, limit: 200 })
      pricePoliciesPage.value = await fetchAiResource('price-policies', { skip: 0, limit: 200 })
      priceTiersPage.value = await fetchAiResource('price-tiers', { skip: 0, limit: 200 })
      reconcileModelSelection()
    }
    if (activeSection.value.key === 'scenarios') {
      scenariosPage.value = page.value
      capabilitiesPage.value = await fetchAiResource('capabilities', { skip: 0, limit: 200 })
      baseRoutesPage.value = await fetchAiResource('base-routes', { skip: 0, limit: 200 })
      tenantStrategiesPage.value = await fetchAiResource('tenant-strategies', { skip: 0, limit: 200 })
      reconcileScenarioSelection()
    }
    if (activeSection.value.key === 'routes') {
      baseRoutesPage.value = page.value
      capabilitiesPage.value = await fetchAiResource('capabilities', { skip: 0, limit: 200 })
      routeModelsPage.value = await fetchAiResource('route-models', { skip: 0, limit: 200 })
      allModelsPage.value = await fetchAiResource('models', { skip: 0, limit: 200 })
      scenariosPage.value = await fetchAiResource('scenarios', { skip: 0, limit: 200 })
      tenantStrategiesPage.value = await fetchAiResource('tenant-strategies', { skip: 0, limit: 200 })
      reconcileRouteSelection()
    }
  } catch (error) {
    errorText.value = error instanceof Error ? error.message : '数据加载失败'
  } finally {
    loading.value = false
  }
}

function switchSection(path: string) {
  if (path !== route.path) router.push(path)
}

function openCreate() {
  if (!activeSection.value.resource) return
  editorMode.value = 'create'
  editorResource.value = activeSection.value.resource
  editingId.value = ''
  formModel.value = normalizeEditorRow(defaultPayload(activeSection.value.resource))
  editorJson.value = JSON.stringify(formModel.value, null, 2)
  editorVisible.value = true
}

function openProviderCreate(resource: AiResource) {
  editorMode.value = 'create'
  editorResource.value = resource
  editingId.value = ''
  const payload = defaultPayload(resource)
  if (resource === 'accounts') payload.provider_id = selectedProviderId.value
  if (resource === 'apis') {
    payload.provider_id = selectedProviderId.value
    payload.account_id = selectedAccountId.value
  }
  formModel.value = normalizeEditorRow(payload)
  editorJson.value = JSON.stringify(formModel.value, null, 2)
  editorVisible.value = true
}

function openModelCreate(resource: AiResource) {
  editorMode.value = 'create'
  editorResource.value = resource
  editingId.value = ''
  const payload = defaultPayload(resource)
  if (resource === 'models') payload.provider_id = selectedProviderId.value
  if (resource === 'price-policies') payload.model_id = selectedModelId.value
  if (resource === 'price-tiers') payload.price_policy_id = selectedPolicyId.value
  formModel.value = normalizeEditorRow(payload)
  editorJson.value = JSON.stringify(formModel.value, null, 2)
  editorVisible.value = true
}

function openScenarioCreate() {
  editorMode.value = 'create'
  editorResource.value = 'scenarios'
  editingId.value = ''
  const payload = defaultPayload('scenarios')
  const group = appGroups.value.find((item) => item.app_code === selectedAppCode.value)
  if (group) {
    payload.app_code = group.app_code
    payload.app_name = group.app_name
  }
  payload.capability_code = capabilityOptions.value[0]?.value || 'chat_completion'
  payload.default_base_route_id = baseRouteOptions.value[0]?.value || ''
  formModel.value = normalizeEditorRow(payload)
  editorJson.value = JSON.stringify(formModel.value, null, 2)
  editorVisible.value = true
}

function openRouteCreate(resource: AiResource) {
  editorMode.value = 'create'
  editorResource.value = resource
  editingId.value = ''
  const payload = defaultPayload(resource)
  if (resource === 'base-routes') {
    payload.capability_code = capabilityOptions.value[0]?.value || 'chat_completion'
  }
  if (resource === 'route-models') {
    payload.base_route_id = selectedBaseRouteId.value
    payload.model_id = modelOptions.value[0]?.value || ''
  }
  formModel.value = normalizeEditorRow(payload)
  editorJson.value = JSON.stringify(formModel.value, null, 2)
  editorVisible.value = true
}

function openEdit(row: Record<string, unknown>, resource = activeSection.value.resource) {
  if (!resource) return
  editorMode.value = 'edit'
  editorResource.value = resource
  editingId.value = String(row.id || '')
  formModel.value = normalizeEditorRow(row)
  editorJson.value = JSON.stringify(row, null, 2)
  editorVisible.value = true
}

async function saveEditor() {
  const resource = editorResource.value ?? activeSection.value.resource
  if (!resource) return
  let payload: Record<string, unknown>
  try {
    payload = formFields.value.length ? buildEditorPayload() : JSON.parse(editorJson.value) as Record<string, unknown>
  } catch {
    ElMessage.error('配置内容格式不合法')
    return
  }
  if (editorMode.value === 'edit' && editingId.value) {
    await updateAiResource(resource, editingId.value, payload)
    ElMessage.success('已更新')
  } else {
    await createAiResource(resource, payload)
    ElMessage.success('已创建')
  }
  editorVisible.value = false
  await loadData()
}

async function removeRow(row: Record<string, unknown>, resource = activeSection.value.resource) {
  if (!resource) return
  const id = String(row.id || '')
  if (!id) return
  await ElMessageBox.confirm('确认删除该配置？删除会写入底座操作日志。', '删除确认', { type: 'warning' })
  await deleteAiResource(resource, id)
  ElMessage.success('已删除')
  await loadData()
}

function selectProvider(row: Record<string, unknown>) {
  selectedProviderId.value = String(row.id || '')
  const firstAccount = selectedProviderAccounts.value[0]
  selectedAccountId.value = firstAccount ? String(firstAccount.id || '') : ''
}

function selectAccount(row: Record<string, unknown>) {
  selectedAccountId.value = String(row.id || '')
}

function selectModel(row: Record<string, unknown>) {
  selectedModelId.value = String(row.id || '')
  const firstPolicy = selectedModelPolicies.value[0]
  selectedPolicyId.value = firstPolicy ? String(firstPolicy.id || '') : ''
}

function selectPolicy(row: Record<string, unknown>) {
  selectedPolicyId.value = String(row.id || '')
}

function reconcileProviderSelection() {
  if (!providerRows.value.some((item) => String(item.id || '') === selectedProviderId.value)) {
    selectedProviderId.value = providerRows.value[0] ? String(providerRows.value[0].id || '') : ''
  }
  if (!selectedProviderAccounts.value.some((item) => String(item.id || '') === selectedAccountId.value)) {
    const firstAccount = selectedProviderAccounts.value[0]
    selectedAccountId.value = firstAccount ? String(firstAccount.id || '') : ''
  }
}

function reconcileModelSelection() {
  if (!providerRows.value.some((item) => String(item.id || '') === selectedProviderId.value)) {
    selectedProviderId.value = providerRows.value[0] ? String(providerRows.value[0].id || '') : ''
  }
  if (!selectedProviderModels.value.some((item) => String(item.id || '') === selectedModelId.value)) {
    const firstModel = selectedProviderModels.value[0]
    selectedModelId.value = firstModel ? String(firstModel.id || '') : ''
  }
  if (!selectedModelPolicies.value.some((item) => String(item.id || '') === selectedPolicyId.value)) {
    const firstPolicy = selectedModelPolicies.value[0]
    selectedPolicyId.value = firstPolicy ? String(firstPolicy.id || '') : ''
  }
}

function reconcileScenarioSelection() {
  if (!appGroups.value.some((item) => item.app_code === selectedAppCode.value)) {
    selectedAppCode.value = appGroups.value[0]?.app_code || ''
  }
}

function reconcileRouteSelection() {
  if (!page.value.items.some((item) => String(item.id || '') === selectedBaseRouteId.value)) {
    selectedBaseRouteId.value = page.value.items[0] ? String(page.value.items[0].id || '') : ''
  }
}

async function submitProviderImport() {
  let payload: AiProviderImportPayload
  try {
    payload = JSON.parse(importJson.value) as AiProviderImportPayload
  } catch {
    ElMessage.error('导入内容不是合法 JSON')
    return
  }
  const result = await importAiProviders(payload)
  importVisible.value = false
  ElMessage.success(`导入完成：供应商 ${result.providers}，账号 ${result.accounts}，API ${result.apis}`)
  await loadData()
}

async function submitModelImport() {
  let payload: AiModelImportPayload
  try {
    payload = JSON.parse(modelImportJson.value) as AiModelImportPayload
  } catch {
    ElMessage.error('导入内容不是合法 JSON')
    return
  }
  const result = await importAiModels(payload)
  modelImportVisible.value = false
  ElMessage.success(`导入完成：模型 ${result.models}，价格策略 ${result.price_policies}，分档 ${result.price_tiers}`)
  await loadData()
}

async function submitScenarioImport() {
  let payload: AiScenarioImportPayload
  try {
    payload = JSON.parse(scenarioImportJson.value) as AiScenarioImportPayload
  } catch {
    ElMessage.error('导入内容不是合法 JSON')
    return
  }
  const result = await importAiScenarios(payload)
  scenarioImportVisible.value = false
  ElMessage.success(`导入完成：AI 场景 ${result.scenarios}`)
  await loadData()
}

async function submitRouteImport() {
  let payload: AiRouteImportPayload
  try {
    payload = JSON.parse(routeImportJson.value) as AiRouteImportPayload
  } catch {
    ElMessage.error('导入内容不是合法 JSON')
    return
  }
  const result = await importAiRoutes(payload)
  routeImportVisible.value = false
  ElMessage.success(`导入完成：基础路由 ${result.base_routes}，模型池 ${result.route_models}`)
  await loadData()
}

function defaultPayload(resource?: AiResource): Record<string, unknown> {
  const status = 'active'
  if (resource === 'providers') return { name: '', code: '', type: 'public_cloud', base_url: '', auth_type: 'api_key', status, priority: 80, region: 'CN', qps_limit: 100, monthly_budget: 10000, owner: '' }
  if (resource === 'accounts') return { provider_id: '', account_name: '', endpoint: '', key_alias: '', encrypted_api_key: '', encrypted_secret: '', quota_limit: 0, used_quota: 0, status }
  if (resource === 'apis') return { provider_id: '', account_id: '', api_name: '', api_path: '', api_type: 'chat', capabilities: ['chat_completion'], auth_type: 'api_key', qps_limit: 100, timeout_ms: 30000, status }
  if (resource === 'models') return { provider_id: '', model_code: '', model_name: '', model_type: 'text', capabilities: ['text_generation'], context_window: 32000, unit: 'tokens', latency_p95: 0, success_rate: 0, status, default_for: [] }
  if (resource === 'price-policies') return { model_id: '', feature_key: '', feature_name: '', model_type: 'text', capability_code: 'chat_completion', billing_mode: 'tiered', billing_unit: 'tokens', platform_unit: 'tokens', base_cost_price: 0, base_sale_price: 0, base_platform_amount: 0, currency: 'CNY', status }
  if (resource === 'price-tiers') return { price_policy_id: '', tier_name: '', mode: 'sync', resolution: '', quality: '', duration_seconds: 0, aspect_ratio: '', cost_price: 0, sale_price: 0, platform_amount: 0, enabled: true, sort_order: 0 }
  if (resource === 'scenarios') return { app_code: '', app_name: '', ai_scenario_code: '', ai_scenario_name: '', scenario_type: 'text', capability_code: 'text_generation', model_type: 'text', default_base_route_id: '', owner: '', version: 'v1.0', status }
  if (resource === 'base-routes') return { route_code: '', route_name: '', capability_code: 'text_generation', model_type: 'text', strategy: 'fallback', timeout_ms: 30000, max_retry: 2, status }
  if (resource === 'route-models') return { base_route_id: '', model_id: '', role: 'candidate', priority: 1, weight: 100, max_retry: 0, timeout_ms: 30000, status }
  if (resource === 'tenant-strategies') return { policy_name: '', tenant_scope: 'include', tenant_ids: [], app_code: '', app_name: '', ai_scenario_code: '', ai_scenario_name: '', default_base_route_id: '', override_base_route_id: '', status }
  if (resource === 'settings') return { setting_key: '', setting_value: {}, description: '', status }
  return { status }
}

function fieldsForResource(resource?: AiResource): FieldConfig[] {
  const status = { key: 'status', label: '状态', type: 'select', options: toOptions(['active', 'warning', 'inactive', 'draft']), required: true } satisfies FieldConfig
  if (resource === 'providers') return [
    { key: 'name', label: '供应商名称', required: true },
    { key: 'code', label: '供应商编码', required: true },
    { key: 'type', label: '类型', type: 'select', options: toOptions(['public_cloud', 'private_cloud', 'local', 'proxy']) },
    { key: 'base_url', label: 'Endpoint', required: true },
    { key: 'auth_type', label: '鉴权方式', type: 'select', options: toOptions(['api_key', 'oauth2', 'aksk', 'none']) },
    { key: 'region', label: '区域' },
    { key: 'owner', label: '负责人' },
    { key: 'priority', label: '优先级', type: 'number' },
    { key: 'qps_limit', label: 'QPS 限制', type: 'number' },
    { key: 'monthly_budget', label: '月预算', type: 'number' },
    status,
  ]
  if (resource === 'accounts') return [
    { key: 'provider_id', label: '供应商 ID', required: true },
    { key: 'account_name', label: '账号名称', required: true },
    { key: 'endpoint', label: '接入 Endpoint' },
    { key: 'key_alias', label: '密钥别名', required: true },
    { key: 'encrypted_api_key', label: '加密 API Key', type: 'textarea' },
    { key: 'encrypted_secret', label: '加密 Secret', type: 'textarea' },
    { key: 'quota_limit', label: '配额上限', type: 'number' },
    { key: 'used_quota', label: '已用配额', type: 'number' },
    status,
  ]
  if (resource === 'apis') return [
    { key: 'provider_id', label: '供应商 ID', required: true },
    { key: 'account_id', label: '账号 ID', required: true },
    { key: 'api_name', label: 'API 名称', required: true },
    { key: 'api_path', label: 'API 路径', required: true },
    { key: 'api_type', label: 'API 类型', type: 'select', options: toOptions(['chat', 'embedding', 'image', 'audio', 'rerank']) },
    { key: 'capabilities', label: '能力标签', type: 'tags' },
    { key: 'auth_type', label: '鉴权方式', type: 'select', options: toOptions(['api_key', 'oauth2', 'aksk', 'none']) },
    { key: 'qps_limit', label: 'QPS 限制', type: 'number' },
    { key: 'timeout_ms', label: '超时时间(ms)', type: 'number' },
    status,
  ]
  if (resource === 'models') return [
    { key: 'provider_id', label: '供应商 ID', required: true },
    { key: 'model_name', label: '模型名称', required: true },
    { key: 'model_code', label: '模型编码', required: true },
    { key: 'model_type', label: '模型类型', type: 'select', options: toOptions(['text', 'embedding', 'image', 'audio', 'rerank']) },
    { key: 'capabilities', label: '能力标签', type: 'tags' },
    { key: 'context_window', label: '上下文窗口', type: 'number' },
    { key: 'unit', label: '计量单位', type: 'select', options: toOptions(['tokens', 'characters', 'images', 'seconds', 'requests']) },
    { key: 'latency_p95', label: 'P95 延迟', type: 'number' },
    { key: 'success_rate', label: '成功率', type: 'number' },
    { key: 'default_for', label: '默认场景', type: 'tags' },
    status,
  ]
  if (resource === 'price-policies') return [
    { key: 'model_id', label: '模型 ID', required: true },
    { key: 'feature_key', label: '功能 Key', required: true },
    { key: 'feature_name', label: '功能名称', required: true },
    { key: 'model_type', label: '模型类型', type: 'select', options: toOptions(['text', 'embedding', 'image', 'audio', 'rerank']) },
    { key: 'capability_code', label: '能力编码', required: true },
    { key: 'billing_mode', label: '计费模式', type: 'select', options: toOptions(['per_unit', 'tiered', 'fixed']) },
    { key: 'billing_unit', label: '计费单位', type: 'select', options: toOptions(['tokens', 'characters', 'images', 'seconds', 'requests']) },
    { key: 'platform_unit', label: '平台单位', type: 'select', options: toOptions(['tokens', 'characters', 'images', 'seconds', 'requests']) },
    { key: 'base_cost_price', label: '基础成本价', type: 'number' },
    { key: 'base_sale_price', label: '基础销售价', type: 'number' },
    { key: 'base_platform_amount', label: '平台用量基数', type: 'number' },
    { key: 'currency', label: '币种', type: 'select', options: toOptions(['CNY', 'USD']) },
    status,
  ]
  if (resource === 'price-tiers') return [
    { key: 'price_policy_id', label: '价格策略 ID', required: true },
    { key: 'tier_name', label: '分档名称', required: true },
    { key: 'mode', label: '模式', type: 'select', options: toOptions(['sync', 'async', 'stream']) },
    { key: 'resolution', label: '分辨率' },
    { key: 'quality', label: '质量' },
    { key: 'duration_seconds', label: '时长(秒)', type: 'number' },
    { key: 'aspect_ratio', label: '画幅' },
    { key: 'cost_price', label: '成本价', type: 'number' },
    { key: 'sale_price', label: '销售价', type: 'number' },
    { key: 'platform_amount', label: '平台金额', type: 'number' },
    { key: 'enabled', label: '启用', type: 'switch' },
    { key: 'sort_order', label: '排序', type: 'number' },
  ]
  if (resource === 'scenarios') return [
    { key: 'app_name', label: '应用名称', required: true },
    { key: 'app_code', label: '应用编码', required: true },
    { key: 'ai_scenario_name', label: 'AI 场景名称', required: true },
    { key: 'ai_scenario_code', label: 'AI 场景编码', required: true },
    { key: 'scenario_type', label: '场景类型', type: 'select', options: toOptions(['text', 'embedding', 'image', 'audio']) },
    { key: 'capability_code', label: '能力编码', type: 'select', options: capabilityOptions.value, required: true },
    { key: 'model_type', label: '模型类型', type: 'select', options: toOptions(['text', 'embedding', 'image', 'audio', 'rerank']) },
    { key: 'default_base_route_id', label: '默认基础路由', type: 'select', options: baseRouteOptions.value, required: true },
    { key: 'owner', label: '负责人' },
    { key: 'version', label: '版本' },
    status,
  ]
  if (resource === 'base-routes') return [
    { key: 'route_name', label: '路由名称', required: true },
    { key: 'route_code', label: '路由编码', required: true },
    { key: 'capability_code', label: '能力编码', type: 'select', options: capabilityOptions.value, required: true },
    { key: 'model_type', label: '模型类型', type: 'select', options: toOptions(['text', 'embedding', 'image', 'audio', 'rerank']) },
    { key: 'strategy', label: '策略', type: 'select', options: toOptions(['fixed', 'fallback', 'priority', 'load_balance', 'cost_first', 'quality_first', 'latency_first', 'quota_aware', 'tenant_custom', 'capability_match']) },
    { key: 'timeout_ms', label: '超时时间(ms)', type: 'number' },
    { key: 'max_retry', label: '最大重试', type: 'number' },
    status,
  ]
  if (resource === 'route-models') return [
    { key: 'base_route_id', label: '基础路由', type: 'select', options: baseRouteOptions.value, required: true },
    { key: 'model_id', label: '模型', type: 'select', options: modelOptions.value, required: true },
    { key: 'role', label: '角色', type: 'select', options: toOptions(['primary', 'fallback', 'candidate']), required: true },
    { key: 'priority', label: '优先级', type: 'number' },
    { key: 'weight', label: '权重', type: 'number' },
    { key: 'max_retry', label: '最大重试', type: 'number' },
    { key: 'timeout_ms', label: '超时时间(ms)', type: 'number' },
    status,
  ]
  if (resource === 'tenant-strategies') return [
    { key: 'policy_name', label: '策略名称', required: true },
    { key: 'tenant_scope', label: '租户范围', type: 'select', options: toOptions(['include', 'exclude', 'all']) },
    { key: 'tenant_ids', label: '租户 ID', type: 'tags' },
    { key: 'app_name', label: '应用名称' },
    { key: 'app_code', label: '应用编码', required: true },
    { key: 'ai_scenario_name', label: 'AI 场景名称' },
    { key: 'ai_scenario_code', label: 'AI 场景编码', required: true },
    { key: 'override_base_route_id', label: '覆盖基础路由 ID' },
    { key: 'extra_config', label: '扩展配置 JSON', type: 'json' },
    status,
  ]
  if (resource === 'settings') return [
    { key: 'setting_key', label: '配置项', required: true },
    { key: 'setting_value', label: '配置值 JSON', type: 'json', required: true },
    { key: 'description', label: '说明', type: 'textarea' },
    status,
  ]
  return []
}

function normalizeEditorRow(row: Record<string, unknown>) {
  const next = { ...row }
  for (const field of fieldsForResource(editorResource.value ?? activeSection.value.resource)) {
    if (field.type === 'tags' && !Array.isArray(next[field.key])) {
      next[field.key] = next[field.key] ? [String(next[field.key])] : []
    }
    if (field.type === 'json') {
      const value = next[field.key]
      next[field.key] = typeof value === 'string' ? value : JSON.stringify(value ?? {}, null, 2)
    }
  }
  return next
}

function buildEditorPayload() {
  const payload: Record<string, unknown> = {}
  for (const field of formFields.value) {
    payload[field.key] = formModel.value[field.key]
    if (field.type === 'json') {
      const raw = String(payload[field.key] ?? '').trim()
      payload[field.key] = raw ? JSON.parse(raw) : {}
    }
  }
  return payload
}

watch(() => route.path, () => {
  page.value.skip = 0
  keyword.value = ''
  loadData()
})

onMounted(loadData)
</script>

<template>
  <NeuroAgentPageShell class="ai-center-page">
    <template #title>AI 能力中心</template>
    <template #subtitle>{{ activeSection.description }}</template>
    <template #actions>
      <el-button :icon="Refresh" @click="loadData">刷新</el-button>
      <el-button v-if="activeSection.key === 'providers'" v-permission="'ai_capability_center:manage'" :icon="Upload" @click="importVisible = true">整体导入</el-button>
      <el-button v-if="activeSection.key === 'models'" v-permission="'ai_capability_center:manage'" :icon="Upload" @click="modelImportVisible = true">模型导入</el-button>
      <el-button v-if="activeSection.key === 'scenarios'" v-permission="'ai_capability_center:manage'" :icon="Upload" @click="scenarioImportVisible = true">场景导入</el-button>
      <el-button v-if="activeSection.key === 'routes'" v-permission="'ai_capability_center:manage'" :icon="Upload" @click="routeImportVisible = true">路由导入</el-button>
      <el-button v-if="canWrite && !['providers', 'models', 'scenarios', 'routes'].includes(activeSection.key)" v-permission="'ai_capability_center:manage'" type="primary" :icon="Plus" @click="openCreate">新增配置</el-button>
    </template>

    <div class="ai-center-tabs">
      <button v-for="item in sections" :key="item.key" :class="{ active: item.route === route.path }" @click="switchSection(item.route)">
        {{ item.title }}
      </button>
    </div>

    <div v-if="activeSection.key === 'dashboard'" class="ai-dashboard" v-loading="loading">
      <el-alert v-if="errorText" :title="errorText" type="error" show-icon />
      <div class="ai-metrics">
        <article v-for="metric in overview?.metrics || []" :key="metric.label">
          <span>{{ metric.label }}</span>
          <strong>{{ metric.value }}</strong>
          <small>{{ metric.trend }}</small>
        </article>
      </div>
      <section class="ai-panel">
        <header><strong>7 天用量趋势</strong><span>调用量、成本和销售额聚合</span></header>
        <el-table :data="overview?.usage_trend || []" border>
          <el-table-column prop="date" label="日期" min-width="130" />
          <el-table-column label="调用量" width="120">
            <template #default="{ row }">{{ numberText(row.calls) }}</template>
          </el-table-column>
          <el-table-column label="成本" width="130">
            <template #default="{ row }">{{ moneyText(row.cost_amount) }}</template>
          </el-table-column>
          <el-table-column label="销售额" width="130">
            <template #default="{ row }">{{ moneyText(row.billing_amount) }}</template>
          </el-table-column>
        </el-table>
      </section>
      <section class="ai-dashboard-grid">
        <div class="ai-panel">
          <header><strong>成本结构</strong><span>按模型类型汇总</span></header>
          <div class="ai-share-list">
            <div v-for="item in overview?.model_cost_share || []" :key="String(item.model_type)">
              <span>{{ item.model_type || 'unknown' }}</span>
              <strong>{{ moneyText(item.cost_amount) }}</strong>
            </div>
            <el-empty v-if="!(overview?.model_cost_share || []).length" :image-size="72" description="暂无成本数据" />
          </div>
        </div>
        <div class="ai-panel">
          <header><strong>租户排行榜</strong><span>近 7 天销售额排序</span></header>
          <el-table :data="overview?.tenant_ranking || []" border>
            <el-table-column prop="tenant_name" label="租户" min-width="140" />
            <el-table-column label="调用" width="100">
              <template #default="{ row }">{{ numberText(row.calls) }}</template>
            </el-table-column>
            <el-table-column label="利润" width="110">
              <template #default="{ row }">{{ moneyText(row.profit_amount) }}</template>
            </el-table-column>
          </el-table>
        </div>
      </section>
      <section class="ai-panel">
        <header><strong>平台健康检查</strong><span>供应商、路由、租户策略和日志写入状态</span></header>
        <div class="ai-health">
          <el-tag v-for="item in overview?.health_checks || []" :key="item.name" :type="statusType(item.status)">
            {{ item.name }}：{{ item.message }}
          </el-tag>
        </div>
      </section>
      <section class="ai-panel">
        <header><strong>核心基础路由</strong><span>AI 场景默认绑定，租户策略可覆盖</span></header>
        <el-table :data="overview?.core_base_routes || []" border>
          <el-table-column prop="route_name" label="路由" min-width="180" />
          <el-table-column prop="route_code" label="编码" min-width="180" />
          <el-table-column prop="capability_code" label="能力" width="160" />
          <el-table-column prop="strategy" label="策略" width="140" />
          <el-table-column prop="status" label="状态" width="110" />
        </el-table>
      </section>
    </div>

    <div v-else-if="activeSection.key === 'providers'" class="ai-resource ai-provider-workbench">
      <div class="ai-toolbar">
        <el-input v-model="keyword" clearable placeholder="搜索供应商名称、编码、负责人" @keyup.enter="loadData">
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
        <el-button :icon="Search" @click="loadData">查询</el-button>
        <el-button v-permission="'ai_capability_center:manage'" type="primary" :icon="Plus" @click="openCreate">新增供应商</el-button>
      </div>
      <el-alert v-if="errorText" :title="errorText" type="error" show-icon />

      <section class="ai-panel">
        <header><strong>供应商</strong><span>平台统一维护，不进入租户后台和套餐售卖</span></header>
        <el-table v-loading="loading" :data="providerRows" border class="ai-table" empty-text="暂无供应商" highlight-current-row @row-click="selectProvider">
          <el-table-column prop="name" label="供应商" min-width="160" />
          <el-table-column prop="code" label="编码" min-width="140" />
          <el-table-column prop="type" label="类型" width="120" />
          <el-table-column prop="base_url" label="Endpoint" min-width="220" />
          <el-table-column prop="owner" label="负责人" width="120" />
          <el-table-column label="状态" width="110">
            <template #default="{ row }"><el-tag :type="statusType(row.status)">{{ displayCell(row, 'status') }}</el-tag></template>
          </el-table-column>
          <el-table-column fixed="right" label="操作" width="150">
            <template #default="{ row }">
              <el-button v-permission="'ai_capability_center:manage'" link type="primary" :icon="Edit" @click.stop="openEdit(row, 'providers')">编辑</el-button>
              <el-button v-permission="'ai_capability_center:manage'" link type="danger" :icon="Delete" @click.stop="removeRow(row, 'providers')">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
        <el-pagination
          :current-page="currentPage"
          class="ai-pagination"
          layout="total, prev, pager, next"
          :page-size="page.limit"
          :total="page.total"
          @current-change="(pageNo: number) => { page.skip = (pageNo - 1) * page.limit; loadData() }"
        />
      </section>

      <section class="ai-provider-grid">
        <div class="ai-panel">
          <header>
            <strong>接入账号</strong>
            <el-button v-permission="'ai_capability_center:manage'" size="small" type="primary" :icon="Plus" :disabled="!selectedProviderId" @click="openProviderCreate('accounts')">新增账号</el-button>
          </header>
          <el-table :data="selectedProviderAccounts" border class="ai-table" empty-text="请选择供应商或新增账号" highlight-current-row @row-click="selectAccount">
            <el-table-column prop="account_name" label="账号" min-width="130" />
            <el-table-column prop="endpoint" label="Endpoint" min-width="190" />
            <el-table-column prop="key_alias" label="密钥别名" min-width="130" />
            <el-table-column label="配额" width="120">
              <template #default="{ row }">{{ numberText(row.used_quota) }} / {{ numberText(row.quota_limit) }}</template>
            </el-table-column>
            <el-table-column label="状态" width="100">
              <template #default="{ row }"><el-tag :type="statusType(row.status)">{{ displayCell(row, 'status') }}</el-tag></template>
            </el-table-column>
            <el-table-column fixed="right" label="操作" width="150">
              <template #default="{ row }">
                <el-button v-permission="'ai_capability_center:manage'" link type="primary" :icon="Edit" @click.stop="openEdit(row, 'accounts')">编辑</el-button>
                <el-button v-permission="'ai_capability_center:manage'" link type="danger" :icon="Delete" @click.stop="removeRow(row, 'accounts')">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </div>

        <div class="ai-panel">
          <header>
            <strong>API 配置</strong>
            <el-button v-permission="'ai_capability_center:manage'" size="small" type="primary" :icon="Plus" :disabled="!selectedProviderId || !selectedAccountId" @click="openProviderCreate('apis')">新增 API</el-button>
          </header>
          <el-table :data="selectedAccountApis" border class="ai-table" empty-text="请选择账号或新增 API">
            <el-table-column prop="api_name" label="API" min-width="150" />
            <el-table-column prop="api_path" label="路径" min-width="180" />
            <el-table-column prop="api_type" label="类型" width="110" />
            <el-table-column label="能力" min-width="160">
              <template #default="{ row }">{{ displayCell(row, 'capabilities') }}</template>
            </el-table-column>
            <el-table-column prop="timeout_ms" label="超时(ms)" width="110" />
            <el-table-column label="状态" width="100">
              <template #default="{ row }"><el-tag :type="statusType(row.status)">{{ displayCell(row, 'status') }}</el-tag></template>
            </el-table-column>
            <el-table-column fixed="right" label="操作" width="150">
              <template #default="{ row }">
                <el-button v-permission="'ai_capability_center:manage'" link type="primary" :icon="Edit" @click="openEdit(row, 'apis')">编辑</el-button>
                <el-button v-permission="'ai_capability_center:manage'" link type="danger" :icon="Delete" @click="removeRow(row, 'apis')">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </section>
    </div>

    <div v-else-if="activeSection.key === 'models'" class="ai-resource ai-model-workbench">
      <div class="ai-toolbar">
        <el-input v-model="keyword" clearable placeholder="搜索模型名称、编码、能力标签" @keyup.enter="loadData">
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
        <el-button :icon="Search" @click="loadData">查询</el-button>
        <el-button v-permission="'ai_capability_center:manage'" type="primary" :icon="Plus" :disabled="!selectedProviderId" @click="openModelCreate('models')">新增模型</el-button>
      </div>
      <el-alert v-if="errorText" :title="errorText" type="error" show-icon />

      <section class="ai-model-shell">
        <aside class="ai-provider-rail">
          <header>
            <strong>供应商</strong>
            <span>{{ providerRows.length }} 个接入源</span>
          </header>
          <button
            v-for="provider in providerRows"
            :key="String(provider.id)"
            :class="{ active: String(provider.id || '') === selectedProviderId }"
            @click="selectedProviderId = String(provider.id || ''); reconcileModelSelection()"
          >
            <strong>{{ provider.name }}</strong>
            <span>{{ provider.code }} · {{ provider.type || 'provider' }}</span>
          </button>
          <el-empty v-if="!providerRows.length" :image-size="72" description="暂无供应商" />
        </aside>

        <div class="ai-model-main">
          <section class="ai-panel ai-model-hero">
            <header><strong>{{ activeProviderName }}</strong><span>模型目录、价格策略和分档价格统一在平台侧维护</span></header>
            <div class="ai-model-stats">
              <div><span>模型</span><strong>{{ selectedProviderModels.length }}</strong></div>
              <div><span>价格策略</span><strong>{{ selectedModelPolicies.length }}</strong></div>
              <div><span>分档</span><strong>{{ selectedPolicyTiers.length }}</strong></div>
            </div>
          </section>

          <section class="ai-panel">
            <header><strong>模型列表</strong><span>选择模型后维护价格策略</span></header>
            <el-table v-loading="loading" :data="selectedProviderModels" border class="ai-table" empty-text="当前供应商暂无模型" highlight-current-row @row-click="selectModel">
              <el-table-column prop="model_name" label="模型" min-width="160" />
              <el-table-column prop="model_code" label="编码" min-width="150" />
              <el-table-column prop="model_type" label="类型" width="110" />
              <el-table-column label="能力" min-width="180">
                <template #default="{ row }">{{ displayCell(row, 'capabilities') }}</template>
              </el-table-column>
              <el-table-column label="上下文" width="110">
                <template #default="{ row }">{{ numberText(row.context_window) }}</template>
              </el-table-column>
              <el-table-column label="状态" width="100">
                <template #default="{ row }"><el-tag :type="statusType(row.status)">{{ displayCell(row, 'status') }}</el-tag></template>
              </el-table-column>
              <el-table-column fixed="right" label="操作" width="150">
                <template #default="{ row }">
                  <el-button v-permission="'ai_capability_center:manage'" link type="primary" :icon="Edit" @click.stop="openEdit(row, 'models')">编辑</el-button>
                  <el-button v-permission="'ai_capability_center:manage'" link type="danger" :icon="Delete" @click.stop="removeRow(row, 'models')">删除</el-button>
                </template>
              </el-table-column>
            </el-table>
            <el-pagination
              :current-page="currentPage"
              class="ai-pagination"
              layout="total, prev, pager, next"
              :page-size="page.limit"
              :total="page.total"
              @current-change="(pageNo: number) => { page.skip = (pageNo - 1) * page.limit; loadData() }"
            />
          </section>

          <section class="ai-model-price-grid">
            <div class="ai-panel">
              <header>
                <strong>价格策略</strong>
                <el-button v-permission="'ai_capability_center:manage'" size="small" type="primary" :icon="Plus" :disabled="!selectedModelId" @click="openModelCreate('price-policies')">新增策略</el-button>
              </header>
              <el-table :data="selectedModelPolicies" border class="ai-table" empty-text="请选择模型或新增价格策略" highlight-current-row @row-click="selectPolicy">
                <el-table-column prop="feature_name" label="功能" min-width="150" />
                <el-table-column prop="feature_key" label="Key" min-width="140" />
                <el-table-column prop="billing_mode" label="计费" width="100" />
                <el-table-column label="基础价" width="150">
                  <template #default="{ row }">{{ moneyText(row.base_cost_price) }} / {{ moneyText(row.base_sale_price) }}</template>
                </el-table-column>
                <el-table-column fixed="right" label="操作" width="150">
                  <template #default="{ row }">
                    <el-button v-permission="'ai_capability_center:manage'" link type="primary" :icon="Edit" @click.stop="openEdit(row, 'price-policies')">编辑</el-button>
                    <el-button v-permission="'ai_capability_center:manage'" link type="danger" :icon="Delete" @click.stop="removeRow(row, 'price-policies')">删除</el-button>
                  </template>
                </el-table-column>
              </el-table>
            </div>

            <div class="ai-panel">
              <header>
                <strong>分档价格</strong>
                <el-button v-permission="'ai_capability_center:manage'" size="small" type="primary" :icon="Plus" :disabled="!selectedPolicyId" @click="openModelCreate('price-tiers')">新增分档</el-button>
              </header>
              <el-table :data="selectedPolicyTiers" border class="ai-table" empty-text="请选择价格策略或新增分档">
                <el-table-column prop="tier_name" label="分档" min-width="130" />
                <el-table-column prop="mode" label="模式" width="90" />
                <el-table-column label="成本/销售" width="150">
                  <template #default="{ row }">{{ moneyText(row.cost_price) }} / {{ moneyText(row.sale_price) }}</template>
                </el-table-column>
                <el-table-column prop="sort_order" label="排序" width="80" />
                <el-table-column fixed="right" label="操作" width="150">
                  <template #default="{ row }">
                    <el-button v-permission="'ai_capability_center:manage'" link type="primary" :icon="Edit" @click="openEdit(row, 'price-tiers')">编辑</el-button>
                    <el-button v-permission="'ai_capability_center:manage'" link type="danger" :icon="Delete" @click="removeRow(row, 'price-tiers')">删除</el-button>
                  </template>
                </el-table-column>
              </el-table>
            </div>
          </section>
        </div>
      </section>
    </div>

    <div v-else-if="activeSection.key === 'scenarios'" class="ai-resource ai-scenario-workbench">
      <div class="ai-toolbar">
        <el-input v-model="keyword" clearable placeholder="搜索应用、场景、负责人" @keyup.enter="loadData">
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
        <el-button :icon="Search" @click="loadData">查询</el-button>
        <el-button v-permission="'ai_capability_center:manage'" type="primary" :icon="Plus" @click="openScenarioCreate">新增场景</el-button>
      </div>
      <el-alert v-if="errorText" :title="errorText" type="error" show-icon />

      <section class="ai-model-shell">
        <aside class="ai-provider-rail ai-app-rail">
          <header>
            <strong>应用分组</strong>
            <span>{{ appGroups.length }} 个应用</span>
          </header>
          <button
            v-for="app in appGroups"
            :key="app.app_code"
            :class="{ active: app.app_code === selectedAppCode }"
            @click="selectedAppCode = app.app_code"
          >
            <strong>{{ app.app_name }}</strong>
            <span>{{ app.app_code }} · {{ app.active }}/{{ app.total }} active</span>
          </button>
          <el-empty v-if="!appGroups.length" :image-size="72" description="暂无 AI 场景" />
        </aside>

        <div class="ai-model-main">
          <section class="ai-panel ai-scenario-hero">
            <header><strong>{{ appGroups.find(item => item.app_code === selectedAppCode)?.app_name || 'AI 场景' }}</strong><span>场景是业务调用 AI Gateway 的入口契约</span></header>
            <div class="ai-model-stats">
              <div><span>场景数</span><strong>{{ selectedAppScenarios.length }}</strong></div>
              <div><span>能力字典</span><strong>{{ capabilityOptions.length }}</strong></div>
              <div><span>基础路由</span><strong>{{ baseRouteOptions.length }}</strong></div>
            </div>
          </section>

          <section class="ai-panel">
            <header><strong>场景注册表</strong><span>app_code + ai_scenario_code 保持唯一</span></header>
            <el-table v-loading="loading" :data="selectedAppScenarios" border class="ai-table" empty-text="当前应用暂无场景">
              <el-table-column prop="ai_scenario_name" label="场景" min-width="170" />
              <el-table-column prop="ai_scenario_code" label="场景编码" min-width="180" />
              <el-table-column prop="capability_code" label="能力" width="150" />
              <el-table-column prop="default_base_route_id" label="默认路由" min-width="220" />
              <el-table-column label="策略覆盖" width="100">
                <template #default="{ row }">{{ strategyCount(row) }}</template>
              </el-table-column>
              <el-table-column prop="owner" label="负责人" width="120" />
              <el-table-column label="状态" width="100">
                <template #default="{ row }"><el-tag :type="statusType(row.status)">{{ displayCell(row, 'status') }}</el-tag></template>
              </el-table-column>
              <el-table-column fixed="right" label="操作" width="150">
                <template #default="{ row }">
                  <el-button v-permission="'ai_capability_center:manage'" link type="primary" :icon="Edit" @click="openEdit(row, 'scenarios')">编辑</el-button>
                  <el-button v-permission="'ai_capability_center:manage'" link type="danger" :icon="Delete" @click="removeRow(row, 'scenarios')">删除</el-button>
                </template>
              </el-table-column>
            </el-table>
            <el-pagination
              :current-page="currentPage"
              class="ai-pagination"
              layout="total, prev, pager, next"
              :page-size="page.limit"
              :total="page.total"
              @current-change="(pageNo: number) => { page.skip = (pageNo - 1) * page.limit; loadData() }"
            />
          </section>
        </div>
      </section>
    </div>

    <div v-else-if="activeSection.key === 'routes'" class="ai-resource ai-route-workbench">
      <div class="ai-toolbar">
        <el-input v-model="keyword" clearable placeholder="搜索路由、能力、策略" @keyup.enter="loadData">
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
        <el-button :icon="Search" @click="loadData">查询</el-button>
        <el-button v-permission="'ai_capability_center:manage'" type="primary" :icon="Plus" @click="openRouteCreate('base-routes')">新增路由</el-button>
      </div>
      <el-alert v-if="errorText" :title="errorText" type="error" show-icon />

      <section class="ai-model-shell">
        <aside class="ai-provider-rail ai-route-rail">
          <header>
            <strong>基础路由</strong>
            <span>{{ page.total }} 条</span>
          </header>
          <button
            v-for="item in page.items"
            :key="String(item.id)"
            :class="{ active: String(item.id || '') === selectedBaseRouteId }"
            @click="selectedBaseRouteId = String(item.id || '')"
          >
            <strong>{{ item.route_name }}</strong>
            <span>{{ item.route_code }} · {{ item.strategy }}</span>
          </button>
          <el-empty v-if="!page.items.length" :image-size="72" description="暂无基础路由" />
        </aside>

        <div class="ai-model-main">
          <section class="ai-panel ai-route-hero">
            <header><strong>{{ selectedBaseRoute?.route_name || '基础路由' }}</strong><span>路由只维护模型池与策略，不绑定租户和场景</span></header>
            <div class="ai-model-stats">
              <div><span>模型池节点</span><strong>{{ selectedRouteModels.length }}</strong></div>
              <div><span>引用数</span><strong>{{ routeReferenceCount(selectedBaseRouteId) }}</strong></div>
              <div><span>策略</span><strong>{{ selectedBaseRoute?.strategy || '-' }}</strong></div>
            </div>
          </section>

          <section class="ai-panel">
            <header>
              <strong>路由配置</strong>
              <span>{{ selectedBaseRoute?.capability_code || '-' }} · {{ selectedBaseRoute?.model_type || '-' }}</span>
            </header>
            <el-table v-loading="loading" :data="selectedBaseRoute ? [selectedBaseRoute] : []" border class="ai-table" empty-text="请选择基础路由">
              <el-table-column prop="route_code" label="路由编码" min-width="180" />
              <el-table-column prop="capability_code" label="能力" width="150" />
              <el-table-column prop="strategy" label="策略" width="150" />
              <el-table-column prop="timeout_ms" label="超时(ms)" width="110" />
              <el-table-column prop="max_retry" label="重试" width="90" />
              <el-table-column label="状态" width="100">
                <template #default="{ row }"><el-tag :type="statusType(row.status)">{{ displayCell(row, 'status') }}</el-tag></template>
              </el-table-column>
              <el-table-column fixed="right" label="操作" width="150">
                <template #default="{ row }">
                  <el-button v-permission="'ai_capability_center:manage'" link type="primary" :icon="Edit" @click="openEdit(row, 'base-routes')">编辑</el-button>
                  <el-button v-permission="'ai_capability_center:manage'" link type="danger" :icon="Delete" @click="removeRow(row, 'base-routes')">删除</el-button>
                </template>
              </el-table-column>
            </el-table>
          </section>

          <section class="ai-panel">
            <header>
              <strong>模型池</strong>
              <el-button v-permission="'ai_capability_center:manage'" size="small" type="primary" :icon="Plus" :disabled="!selectedBaseRouteId" @click="openRouteCreate('route-models')">新增节点</el-button>
            </header>
            <el-table :data="selectedRouteModels" border class="ai-table" empty-text="当前路由暂无模型池节点">
              <el-table-column label="模型" min-width="180">
                <template #default="{ row }">{{ modelName(row.model_id) }}</template>
              </el-table-column>
              <el-table-column prop="role" label="角色" width="110" />
              <el-table-column prop="priority" label="优先级" width="90" />
              <el-table-column prop="weight" label="权重" width="90" />
              <el-table-column prop="max_retry" label="重试" width="90" />
              <el-table-column prop="timeout_ms" label="超时(ms)" width="110" />
              <el-table-column label="状态" width="100">
                <template #default="{ row }"><el-tag :type="statusType(row.status)">{{ displayCell(row, 'status') }}</el-tag></template>
              </el-table-column>
              <el-table-column fixed="right" label="操作" width="150">
                <template #default="{ row }">
                  <el-button v-permission="'ai_capability_center:manage'" link type="primary" :icon="Edit" @click="openEdit(row, 'route-models')">编辑</el-button>
                  <el-button v-permission="'ai_capability_center:manage'" link type="danger" :icon="Delete" @click="removeRow(row, 'route-models')">删除</el-button>
                </template>
              </el-table-column>
            </el-table>
          </section>
        </div>
      </section>
    </div>

    <div v-else class="ai-resource">
      <div class="ai-toolbar">
        <el-input v-model="keyword" clearable placeholder="搜索当前页面数据" @keyup.enter="loadData">
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
        <el-button :icon="Search" @click="loadData">查询</el-button>
      </div>
      <el-alert v-if="errorText" :title="errorText" type="error" show-icon />
      <el-table v-loading="loading" :data="page.items" border class="ai-table" empty-text="暂无数据">
        <el-table-column v-for="column in activeSection.columns" :key="column.key" :label="column.label" :min-width="column.width || 160">
          <template #default="{ row }">
            <el-tag v-if="column.key === 'status'" :type="statusType(row[column.key])">{{ displayCell(row, column.key) }}</el-tag>
            <span v-else>{{ displayCell(row, column.key) }}</span>
          </template>
        </el-table-column>
        <el-table-column v-if="canWrite" fixed="right" label="操作" width="150">
          <template #default="{ row }">
            <el-button v-permission="'ai_capability_center:manage'" link type="primary" :icon="Edit" @click="openEdit(row)">编辑</el-button>
            <el-button v-permission="'ai_capability_center:manage'" link type="danger" :icon="Delete" @click="removeRow(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-pagination
        :current-page="currentPage"
        class="ai-pagination"
        layout="total, prev, pager, next"
        :page-size="page.limit"
        :total="page.total"
        @current-change="(pageNo: number) => { page.skip = (pageNo - 1) * page.limit; loadData() }"
      />
    </div>

    <el-drawer v-model="editorVisible" :title="editorMode === 'create' ? '新增配置' : '编辑配置'" size="720px">
      <el-alert title="按字段提交配置；保存后会写入底座操作日志。" type="info" show-icon />
      <el-form v-if="formFields.length" class="ai-form" label-position="top">
        <el-form-item v-for="field in formFields" :key="field.key" :label="field.label" :required="field.required">
          <el-select v-if="field.type === 'select'" v-model="formModel[field.key]" filterable>
            <el-option v-for="option in field.options || []" :key="option.value" :label="option.label" :value="option.value" />
          </el-select>
          <el-input-number v-else-if="field.type === 'number'" v-model="formModel[field.key]" :min="0" controls-position="right" />
          <el-input v-else-if="field.type === 'textarea'" v-model="formModel[field.key]" type="textarea" :rows="3" />
          <el-select v-else-if="field.type === 'tags'" v-model="formModel[field.key]" multiple filterable allow-create default-first-option />
          <el-input v-else-if="field.type === 'json'" v-model="formModel[field.key]" class="json-editor" type="textarea" :rows="5" spellcheck="false" />
          <el-switch v-else-if="field.type === 'switch'" v-model="formModel[field.key]" />
          <el-input v-else v-model="formModel[field.key]" />
        </el-form-item>
      </el-form>
      <el-input v-else v-model="editorJson" class="json-editor" type="textarea" :rows="18" spellcheck="false" />
      <template #footer>
        <el-button @click="editorVisible = false">取消</el-button>
        <el-button type="primary" @click="saveEditor">保存</el-button>
      </template>
    </el-drawer>

    <el-dialog v-model="importVisible" title="整体导入供应商" width="780px">
      <el-alert title="导入会在一个事务内 upsert 供应商、账号和 API；任一引用无效会整体回滚。" type="info" show-icon />
      <el-input v-model="importJson" class="json-editor" type="textarea" :rows="18" spellcheck="false" />
      <template #footer>
        <el-button @click="importVisible = false">取消</el-button>
        <el-button type="primary" @click="submitProviderImport">导入</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="modelImportVisible" title="整体导入模型和价格" width="780px">
      <el-alert title="导入会在一个事务内 upsert 模型、价格策略和分档；任一引用无效会整体回滚。" type="info" show-icon />
      <el-input v-model="modelImportJson" class="json-editor" type="textarea" :rows="18" spellcheck="false" />
      <template #footer>
        <el-button @click="modelImportVisible = false">取消</el-button>
        <el-button type="primary" @click="submitModelImport">导入</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="scenarioImportVisible" title="批量导入 AI 场景" width="780px">
      <el-alert title="导入会按 app_code + ai_scenario_code upsert；能力字典或默认路由引用无效会整体回滚。" type="info" show-icon />
      <el-input v-model="scenarioImportJson" class="json-editor" type="textarea" :rows="18" spellcheck="false" />
      <template #footer>
        <el-button @click="scenarioImportVisible = false">取消</el-button>
        <el-button type="primary" @click="submitScenarioImport">导入</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="routeImportVisible" title="整体导入基础路由和模型池" width="780px">
      <el-alert title="导入会按 route_code 和 base_route + model + role upsert；能力、模型、权重、优先级无效会整体回滚。" type="info" show-icon />
      <el-input v-model="routeImportJson" class="json-editor" type="textarea" :rows="18" spellcheck="false" />
      <template #footer>
        <el-button @click="routeImportVisible = false">取消</el-button>
        <el-button type="primary" @click="submitRouteImport">导入</el-button>
      </template>
    </el-dialog>
  </NeuroAgentPageShell>
</template>

<style scoped>
.ai-center-page :deep(.neuro-page-shell__panel-body) {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.ai-center-tabs {
  display: flex;
  gap: 8px;
  overflow-x: auto;
  padding-bottom: 2px;
}

.ai-center-tabs button {
  border: 1px solid var(--neuro-border);
  background: var(--neuro-surface);
  color: var(--neuro-text);
  border-radius: 8px;
  padding: 8px 12px;
  white-space: nowrap;
  cursor: pointer;
}

.ai-center-tabs button.active {
  border-color: var(--neuro-primary);
  color: var(--neuro-primary);
  background: color-mix(in srgb, var(--neuro-primary) 10%, var(--neuro-surface));
}

.ai-dashboard,
.ai-resource {
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-width: 0;
}

.ai-metrics {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
}

.ai-metrics article,
.ai-panel {
  border: 1px solid var(--neuro-border);
  background: var(--neuro-surface);
  border-radius: 8px;
  padding: 14px;
}

.ai-metrics span,
.ai-metrics small,
.ai-panel header span {
  display: block;
  color: var(--neuro-text-muted);
}

.ai-metrics strong {
  display: block;
  margin: 8px 0;
  font-size: 24px;
}

.ai-dashboard-grid {
  display: grid;
  grid-template-columns: minmax(0, 0.85fr) minmax(0, 1.15fr);
  gap: 12px;
}

.ai-panel header {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
}

.ai-share-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.ai-share-list > div {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  border-bottom: 1px solid var(--neuro-border);
  padding-bottom: 8px;
}

.ai-share-list > div:last-child {
  border-bottom: 0;
  padding-bottom: 0;
}

.ai-health {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.ai-toolbar {
  display: flex;
  gap: 8px;
  align-items: center;
}

.ai-toolbar .el-input {
  max-width: 360px;
}

.ai-table {
  width: 100%;
}

.ai-provider-workbench .ai-panel {
  min-width: 0;
}

.ai-provider-grid {
  display: grid;
  grid-template-columns: minmax(0, 0.95fr) minmax(0, 1.05fr);
  gap: 12px;
}

.ai-model-shell {
  display: grid;
  grid-template-columns: 260px minmax(0, 1fr);
  gap: 12px;
  min-width: 0;
}

.ai-provider-rail {
  border: 1px solid var(--neuro-border);
  background:
    linear-gradient(180deg, color-mix(in srgb, var(--neuro-primary) 9%, transparent), transparent 42%),
    var(--neuro-surface);
  border-radius: 8px;
  padding: 12px;
  min-width: 0;
}

.ai-provider-rail header {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 10px;
  margin-bottom: 10px;
}

.ai-provider-rail header span,
.ai-model-stats span {
  color: var(--neuro-text-muted);
  font-size: 12px;
}

.ai-provider-rail button {
  width: 100%;
  border: 1px solid transparent;
  border-radius: 8px;
  background: transparent;
  color: var(--neuro-text);
  cursor: pointer;
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 10px;
  text-align: left;
}

.ai-provider-rail button + button {
  margin-top: 6px;
}

.ai-provider-rail button span {
  color: var(--neuro-text-muted);
  font-size: 12px;
}

.ai-provider-rail button.active {
  border-color: color-mix(in srgb, var(--neuro-primary) 42%, var(--neuro-border));
  background: color-mix(in srgb, var(--neuro-primary) 10%, var(--neuro-surface));
}

.ai-model-main {
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-width: 0;
}

.ai-model-hero {
  background:
    linear-gradient(135deg, color-mix(in srgb, var(--neuro-primary) 14%, transparent), transparent 56%),
    var(--neuro-surface);
}

.ai-scenario-hero {
  background:
    linear-gradient(135deg, color-mix(in srgb, #0f766e 16%, transparent), transparent 58%),
    var(--neuro-surface);
}

.ai-route-hero {
  background:
    linear-gradient(135deg, color-mix(in srgb, #92400e 14%, transparent), transparent 58%),
    var(--neuro-surface);
}

.ai-model-stats {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
}

.ai-model-stats div {
  border: 1px solid var(--neuro-border);
  border-radius: 8px;
  padding: 10px;
}

.ai-model-stats strong {
  display: block;
  font-size: 22px;
  margin-top: 4px;
}

.ai-model-price-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
  gap: 12px;
}

.ai-pagination {
  justify-content: flex-end;
}

.ai-form {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 2px 14px;
  margin-top: 12px;
}

.ai-form :deep(.el-form-item) {
  margin-bottom: 12px;
}

.ai-form :deep(.el-select),
.ai-form :deep(.el-input-number) {
  width: 100%;
}

.ai-form :deep(.el-textarea),
.ai-form :deep(.el-form-item:has(.json-editor)) {
  grid-column: 1 / -1;
}

.json-editor {
  margin-top: 12px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
}

@media (max-width: 900px) {
  .ai-metrics {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .ai-dashboard-grid {
    grid-template-columns: 1fr;
  }

  .ai-provider-grid {
    grid-template-columns: 1fr;
  }

  .ai-model-shell,
  .ai-model-price-grid {
    grid-template-columns: 1fr;
  }

  .ai-toolbar {
    align-items: stretch;
    flex-direction: column;
  }

  .ai-toolbar .el-input {
    max-width: none;
  }

  .ai-form {
    grid-template-columns: 1fr;
  }
}
</style>
