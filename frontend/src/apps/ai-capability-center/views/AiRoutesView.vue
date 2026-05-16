<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'

import NeuroAgentPageShell from '@/views/components/NeuroAgentPageShell.vue'

import { createAiResource, fetchAiResource } from '../api'
import { capabilityText, emptyPage, includesKeyword, modelTypeText, numberText, rowId, statusText, statusType, text, type AiRow } from './viewHelpers'
import AiJsonDialog from './AiJsonDialog.vue'
import AiResourceActions from './AiResourceActions.vue'
import './aiPrototype.css'

defineOptions({ name: 'AiRoutesView' })

const loading = ref(false)
const keyword = ref('')
const strategyFilter = ref('all')
const routes = ref(emptyPage())
const routeModels = ref(emptyPage())
const models = ref(emptyPage())
const accounts = ref(emptyPage())
const apis = ref(emptyPage())
const scenarios = ref(emptyPage())
const strategies = ref(emptyPage())
const createVisible = ref(false)

const strategyOptions = computed(() => Array.from(new Set(routes.value.items.map((row) => String(row.strategy || '')).filter(Boolean))))
const routeDialogOptions = {
  capability_code: [
    { label: '对话生成', value: 'chat_completion' },
    { label: '文本生成', value: 'text_generation' },
    { label: '图像生成', value: 'image_generation' },
    { label: '视频生成', value: 'video_generation' },
    { label: '向量生成', value: 'embedding' },
    { label: '重排', value: 'rerank' },
  ],
  strategy: [
    { label: '优先级路由', value: 'priority' },
    { label: '故障降级', value: 'fallback' },
    { label: '权重分配', value: 'weighted' },
    { label: '轮询分配', value: 'round_robin' },
  ],
}
const filtered = computed(() => routes.value.items.filter((row) => {
  const matchesStrategy = strategyFilter.value === 'all' || row.strategy === strategyFilter.value
  return matchesStrategy && includesKeyword(row, keyword.value, ['route_name', 'route_code', 'capability_code', 'model_type', 'strategy'])
}))

function routePool(routeId: unknown) {
  return routeModels.value.items
    .filter((row) => String(row.base_route_id || '') === String(routeId || ''))
    .sort((a, b) => Number(a.priority ?? 999) - Number(b.priority ?? 999))
}

function modelName(modelId: unknown) {
  const model = models.value.items.find((row) => rowId(row) === String(modelId || ''))
  return text(model?.model_name || model?.model_code || modelId)
}

function accountName(accountId: unknown) {
  const id = String(accountId || '')
  if (!id) return '自动选择账号'
  const account = accounts.value.items.find((row) => rowId(row) === id)
  return text(account?.account_name || account?.key_alias || accountId)
}

function apiName(apiId: unknown) {
  const id = String(apiId || '')
  if (!id) return '自动选择 API'
  const api = apis.value.items.find((row) => rowId(row) === id)
  return text(api?.api_name || api?.api_path || apiId)
}

function endpointLabel(node: AiRow) {
  const account = accountName(node.provider_account_id)
  const api = apiName(node.provider_api_id)
  if (account === '自动选择账号' && api === '自动选择 API') return '自动选择可用端点'
  return `${account} / ${api}`
}

function referenceCount(routeId: unknown) {
  const id = String(routeId || '')
  return scenarios.value.items.filter((row) => String(row.default_base_route_id || '') === id).length
    + strategies.value.items.filter((row) => String(row.default_base_route_id || '') === id || String(row.override_base_route_id || '') === id).length
}

function strategyText(value: unknown) {
  const raw = text(value, '').toLowerCase()
  const labels: Record<string, string> = {
    fallback: '故障降级',
    priority: '优先级路由',
    round_robin: '轮询分配',
    weighted: '权重分配',
  }
  return raw ? labels[raw] || raw : '未配置'
}

function roleText(value: unknown) {
  const raw = text(value, '').toLowerCase()
  const labels: Record<string, string> = {
    primary: '主模型',
    fallback: '备用模型',
    backup: '备用模型',
  }
  return raw ? labels[raw] || raw : '模型'
}

function primaryNode(routeId: unknown) {
  const pool = routePool(routeId)
  return pool.find((node) => text(node.role, '').toLowerCase() === 'primary') || pool[0]
}

function fallbackNodes(routeId: unknown) {
  const primary = primaryNode(routeId)
  return routePool(routeId).filter((node) => rowId(node) !== rowId(primary))
}

function modelPoolTitle(routeId: unknown) {
  return routePool(routeId).map((node) => `${roleText(node.role)}：${modelName(node.model_id)}`).join('\n')
}

function endpointSummary(routeId: unknown) {
  const endpoints = Array.from(new Set(routePool(routeId).map(endpointLabel)))
  if (!endpoints.length) return { main: '未配置端点', extra: '' }
  if (endpoints.length === 1) return { main: endpoints[0], extra: '' }
  return { main: endpoints[0], extra: `另有 ${endpoints.length - 1} 个端点` }
}

function endpointTitle(routeId: unknown) {
  return Array.from(new Set(routePool(routeId).map(endpointLabel))).join('\n')
}

function timeoutRetryText(row: AiRow) {
  const seconds = Number(row.timeout_ms ?? 0) / 1000
  const retry = Number(row.max_retry ?? 0)
  const timeout = seconds > 0 ? `${seconds.toLocaleString('zh-CN')} 秒` : '未配置超时'
  return `${timeout} / ${retry > 0 ? `重试 ${retry} 次` : '不重试'}`
}

async function loadData() {
  loading.value = true
  try {
    const [routePage, routeModelPage, modelPage, accountPage, apiPage, scenarioPage, strategyPage] = await Promise.all([
      fetchAiResource('base-routes', { skip: 0, limit: 200, keyword: keyword.value.trim() }),
      fetchAiResource('route-models', { skip: 0, limit: 200 }),
      fetchAiResource('models', { skip: 0, limit: 200 }),
      fetchAiResource('accounts', { skip: 0, limit: 500 }),
      fetchAiResource('apis', { skip: 0, limit: 500 }),
      fetchAiResource('scenarios', { skip: 0, limit: 200 }),
      fetchAiResource('tenant-strategies', { skip: 0, limit: 200 }),
    ])
    routes.value = routePage
    routeModels.value = routeModelPage
    models.value = modelPage
    accounts.value = accountPage
    apis.value = apiPage
    scenarios.value = scenarioPage
    strategies.value = strategyPage
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '基础路由加载失败')
  } finally {
    loading.value = false
  }
}

async function createRoute(payload: AiRow) {
  await createAiResource('base-routes', payload)
  createVisible.value = false
  ElMessage.success('基础路由已创建')
  await loadData()
}

onMounted(loadData)
</script>

<template>
  <NeuroAgentPageShell class="ai-prototype" :show-hero="false">
    <template #title>基础路由</template>
    <template #subtitle>基础路由只定义可复用模型策略，不绑定租户、不绑定具体 AI 场景。</template>
    <template #actions>
      <el-button type="primary" :icon="Plus" @click="createVisible = true">新增基础路由</el-button>
    </template>

    <div class="ai-flow-panel">
      <div>基础路由</div><span>→</span><div>AI 场景绑定</div><span>→</span><div>租户策略覆盖</div><span>→</span><div>配额限流校验</div><span>→</span><div>模型与端点池</div>
    </div>

    <section class="ai-card ai-table">
      <header class="ai-card__header">
        <div>
          <h3>基础路由管理</h3>
          <p class="ai-card__description">AI 场景选择默认基础路由，租户策略中心可按场景覆盖。</p>
        </div>
        <div class="ai-actions">
          <el-button type="primary" :icon="Plus" @click="createVisible = true">新增基础路由</el-button>
        </div>
      </header>
      <div class="ai-card__body">
        <div class="ai-metric-grid">
          <div class="ai-mini-stat"><span>基础路由</span><strong>{{ routes.total }}</strong></div>
          <div class="ai-mini-stat"><span>启用路由</span><strong>{{ routes.items.filter((row) => row.status === 'active').length }}</strong></div>
          <div class="ai-mini-stat"><span>模型节点</span><strong>{{ routeModels.total }}</strong></div>
          <div class="ai-mini-stat"><span>策略种类</span><strong>{{ strategyOptions.length }}</strong></div>
        </div>
        <div class="ai-toolbar">
          <el-input v-model="keyword" clearable placeholder="搜索基础路由、能力、模型、策略" />
          <el-select v-model="strategyFilter" style="width: 180px">
            <el-option label="全部策略" value="all" />
            <el-option v-for="strategy in strategyOptions" :key="strategy" :label="strategyText(strategy)" :value="strategy" />
          </el-select>
        </div>
        <el-table :data="filtered" border v-loading="loading">
          <el-table-column label="基础路由" min-width="220"><template #default="{ row }"><span class="ai-table-cell-main"><strong>{{ text(row.route_name) }}</strong><small>{{ text(row.route_code) }}</small></span></template></el-table-column>
          <el-table-column label="能力 / 类型" min-width="180"><template #default="{ row }">{{ capabilityText(row.capability_code) }} / {{ modelTypeText(row.model_type) }}</template></el-table-column>
          <el-table-column label="策略" width="130"><template #default="{ row }"><el-tag>{{ strategyText(row.strategy) }}</el-tag></template></el-table-column>
          <el-table-column label="模型池" min-width="260">
            <template #default="{ row }">
              <span class="ai-table-cell-main" :title="modelPoolTitle(row.id)">
                <strong>{{ primaryNode(row.id) ? modelName(primaryNode(row.id)?.model_id) : '未配置主模型' }}</strong>
                <small v-if="fallbackNodes(row.id).length">备用 {{ fallbackNodes(row.id).length }} 个：{{ fallbackNodes(row.id).map((node) => modelName(node.model_id)).join('、') }}</small>
                <small v-else>暂无备用模型</small>
              </span>
            </template>
          </el-table-column>
          <el-table-column label="执行端点" min-width="230">
            <template #default="{ row }">
              <span class="ai-table-cell-main" :title="endpointTitle(row.id)">
                <strong>{{ endpointSummary(row.id).main }}</strong>
                <small v-if="endpointSummary(row.id).extra">{{ endpointSummary(row.id).extra }}</small>
              </span>
            </template>
          </el-table-column>
          <el-table-column label="场景引用" width="110"><template #default="{ row }"><strong>{{ referenceCount(row.id) }}</strong></template></el-table-column>
          <el-table-column label="超时 / 重试" width="170"><template #default="{ row }">{{ timeoutRetryText(row) }}</template></el-table-column>
          <el-table-column label="状态" width="100"><template #default="{ row }"><el-tag :type="statusType(row.status)">{{ statusText(row.status) }}</el-tag></template></el-table-column>
          <el-table-column label="操作" width="140"><template #default="{ row }"><AiResourceActions resource="base-routes" :row="row" @saved="loadData" /></template></el-table-column>
        </el-table>
      </div>
    </section>

    <section class="ai-card">
      <header class="ai-card__header">
        <div>
          <h3>调用命中逻辑</h3>
          <p class="ai-card__description">业务调用只传租户、应用和已注册 AI 场景；最终路由由场景默认路由和租户策略共同决定。</p>
        </div>
      </header>
      <div class="ai-card__body">
        <pre class="ai-code-block">tenant_id + app_code + ai_scenario_code
  → 校验 AI 场景已注册
  → 读取场景 default_base_route
  → 查询租户策略 override_base_route
  → 校验多维配额/限流
  → 执行最终基础路由的模型池
  → 按节点绑定或自动策略选择供应商账号/API
  → 跳过停用账号、停用 API、连通性异常 API</pre>
      </div>
    </section>

    <AiJsonDialog
      v-model="createVisible"
      title="新增基础路由"
      :sample="{ route_code: 'route_text_low_cost', route_name: '文本低成本基础路由', capability_code: 'chat_completion', model_type: 'text', strategy: 'fallback', timeout_ms: 30000, max_retry: 2, status: 'active' }"
      :options="routeDialogOptions"
      @submit="createRoute"
    />
  </NeuroAgentPageShell>
</template>
