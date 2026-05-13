<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Delete, Edit, Plus, Refresh, Search } from '@element-plus/icons-vue'

import NeuroAgentPageShell from '@/views/components/NeuroAgentPageShell.vue'

import { createAiResource, deleteAiResource, fetchAiOverview, fetchAiResource, updateAiResource } from '../api'
import type { AiOverview, AiPage, AiResource, AiSectionConfig } from '../types'

defineOptions({ name: 'AiCapabilityCenterView' })

const route = useRoute()
const router = useRouter()

type FieldConfig = {
  key: string
  label: string
  type?: 'text' | 'number' | 'textarea' | 'select' | 'tags' | 'json'
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
  { key: 'usage', title: '用量统计', route: '/ai-capability-center/usage', resource: 'usage-records', description: '按租户、应用、场景、模型、供应商统计用量、成本和销售额。', columns: [
    { key: 'called_at', label: '调用时间', width: 180 }, { key: 'tenant_name', label: '租户', width: 160 }, { key: 'app_name', label: '应用', width: 140 }, { key: 'ai_scenario_name', label: '场景', width: 180 }, { key: 'usage_amount', label: '用量', width: 100 }, { key: 'usage_unit', label: '单位', width: 120 }, { key: 'calls', label: '调用', width: 90 }, { key: 'cost_amount', label: '成本', width: 100 }, { key: 'billing_amount', label: '销售额', width: 110 }, { key: 'status', label: '状态', width: 100 },
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
const currentPage = computed(() => Math.floor(page.value.skip / Math.max(page.value.limit, 1)) + 1)
const keyword = ref('')
const usageDateRange = ref<[string, string] | []>([])
const loading = ref(false)
const errorText = ref('')
const editorVisible = ref(false)
const editorMode = ref<'create' | 'edit'>('create')
const editorJson = ref('{}')
const formModel = ref<Record<string, unknown>>({})
const editingId = ref('')

const canWrite = computed(() => Boolean(activeSection.value.resource && activeSection.value.writable))
const formFields = computed<FieldConfig[]>(() => fieldsForResource(activeSection.value.resource))
const usageSummary = computed(() => Array.isArray(page.value.summary) ? page.value.summary as Array<Record<string, unknown>> : [])

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
      start_date: activeSection.value.resource === 'usage-records' ? usageDateRange.value[0] : undefined,
      end_date: activeSection.value.resource === 'usage-records' ? usageDateRange.value[1] : undefined,
    })
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
  editorMode.value = 'create'
  editingId.value = ''
  formModel.value = normalizeEditorRow(defaultPayload(activeSection.value.resource))
  editorJson.value = JSON.stringify(formModel.value, null, 2)
  editorVisible.value = true
}

function openEdit(row: Record<string, unknown>) {
  editorMode.value = 'edit'
  editingId.value = String(row.id || '')
  formModel.value = normalizeEditorRow(row)
  editorJson.value = JSON.stringify(row, null, 2)
  editorVisible.value = true
}

async function saveEditor() {
  if (!activeSection.value.resource) return
  let payload: Record<string, unknown>
  try {
    payload = formFields.value.length ? buildEditorPayload() : JSON.parse(editorJson.value) as Record<string, unknown>
  } catch {
    ElMessage.error('配置内容格式不合法')
    return
  }
  if (editorMode.value === 'edit' && editingId.value) {
    await updateAiResource(activeSection.value.resource, editingId.value, payload)
    ElMessage.success('已更新')
  } else {
    await createAiResource(activeSection.value.resource, payload)
    ElMessage.success('已创建')
  }
  editorVisible.value = false
  await loadData()
}

async function removeRow(row: Record<string, unknown>) {
  if (!activeSection.value.resource) return
  const id = String(row.id || '')
  if (!id) return
  await ElMessageBox.confirm('确认删除该配置？删除会写入底座操作日志。', '删除确认', { type: 'warning' })
  await deleteAiResource(activeSection.value.resource, id)
  ElMessage.success('已删除')
  await loadData()
}

function defaultPayload(resource?: AiResource): Record<string, unknown> {
  const status = 'active'
  if (resource === 'providers') return { name: '', code: '', type: 'public_cloud', base_url: '', auth_type: 'api_key', status, priority: 80, region: 'CN', qps_limit: 100, monthly_budget: 10000, owner: '' }
  if (resource === 'models') return { provider_id: '', model_code: '', model_name: '', model_type: 'text', capabilities: ['text_generation'], context_window: 32000, unit: 'tokens', latency_p95: 0, success_rate: 0, status, default_for: [] }
  if (resource === 'scenarios') return { app_code: '', app_name: '', ai_scenario_code: '', ai_scenario_name: '', scenario_type: 'text', capability_code: 'text_generation', model_type: 'text', default_base_route_id: '', owner: '', version: 'v1.0', status }
  if (resource === 'base-routes') return { route_code: '', route_name: '', capability_code: 'text_generation', model_type: 'text', strategy: 'fallback', timeout_ms: 30000, max_retry: 2, status }
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
  if (resource === 'scenarios') return [
    { key: 'app_name', label: '应用名称', required: true },
    { key: 'app_code', label: '应用编码', required: true },
    { key: 'ai_scenario_name', label: 'AI 场景名称', required: true },
    { key: 'ai_scenario_code', label: 'AI 场景编码', required: true },
    { key: 'scenario_type', label: '场景类型', type: 'select', options: toOptions(['text', 'embedding', 'image', 'audio']) },
    { key: 'capability_code', label: '能力编码', required: true },
    { key: 'model_type', label: '模型类型', type: 'select', options: toOptions(['text', 'embedding', 'image', 'audio', 'rerank']) },
    { key: 'default_base_route_id', label: '默认基础路由 ID' },
    { key: 'owner', label: '负责人' },
    { key: 'version', label: '版本' },
    status,
  ]
  if (resource === 'base-routes') return [
    { key: 'route_name', label: '路由名称', required: true },
    { key: 'route_code', label: '路由编码', required: true },
    { key: 'capability_code', label: '能力编码', required: true },
    { key: 'model_type', label: '模型类型', type: 'select', options: toOptions(['text', 'embedding', 'image', 'audio', 'rerank']) },
    { key: 'strategy', label: '策略', type: 'select', options: toOptions(['fallback', 'weighted', 'priority', 'cost_first', 'latency_first']) },
    { key: 'timeout_ms', label: '超时时间(ms)', type: 'number' },
    { key: 'max_retry', label: '最大重试', type: 'number' },
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
  for (const field of fieldsForResource(activeSection.value.resource)) {
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
  usageDateRange.value = []
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
      <el-button v-if="canWrite" v-permission="'ai_capability_center:manage'" type="primary" :icon="Plus" @click="openCreate">新增配置</el-button>
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

    <div v-else class="ai-resource">
      <div class="ai-toolbar">
        <el-input v-model="keyword" clearable placeholder="搜索当前页面数据" @keyup.enter="loadData">
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
        <el-date-picker
          v-if="activeSection.resource === 'usage-records'"
          v-model="usageDateRange"
          type="daterange"
          value-format="YYYY-MM-DD"
          start-placeholder="开始日期"
          end-placeholder="结束日期"
          unlink-panels
        />
        <el-button :icon="Search" @click="loadData">查询</el-button>
      </div>
      <div v-if="activeSection.resource === 'usage-records' && usageSummary.length" class="ai-usage-summary">
        <article v-for="item in usageSummary" :key="String(item.usage_unit)">
          <span>{{ item.usage_unit }}</span>
          <strong>{{ numberText(item.usage_amount) }}</strong>
          <small>{{ numberText(item.calls) }} 次调用 / {{ moneyText(item.billing_amount) }}</small>
        </article>
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

    <el-dialog v-model="editorVisible" :title="editorMode === 'create' ? '新增配置' : '编辑配置'" width="720px">
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
          <el-input v-else v-model="formModel[field.key]" />
        </el-form-item>
      </el-form>
      <el-input v-else v-model="editorJson" class="json-editor" type="textarea" :rows="18" spellcheck="false" />
      <template #footer>
        <el-button @click="editorVisible = false">取消</el-button>
        <el-button type="primary" @click="saveEditor">保存</el-button>
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

.ai-toolbar :deep(.el-date-editor) {
  max-width: 320px;
}

.ai-usage-summary {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 10px;
}

.ai-usage-summary article {
  border: 1px solid var(--neuro-border);
  background: var(--neuro-surface);
  border-radius: 8px;
  padding: 12px;
}

.ai-usage-summary span,
.ai-usage-summary small {
  display: block;
  color: var(--neuro-text-muted);
}

.ai-usage-summary strong {
  display: block;
  margin: 6px 0;
  font-size: 20px;
}

.ai-table {
  width: 100%;
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

  .ai-toolbar {
    align-items: stretch;
    flex-direction: column;
  }

  .ai-toolbar .el-input {
    max-width: none;
  }

  .ai-toolbar :deep(.el-date-editor) {
    max-width: none;
    width: 100%;
  }

  .ai-usage-summary {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .ai-form {
    grid-template-columns: 1fr;
  }
}
</style>
