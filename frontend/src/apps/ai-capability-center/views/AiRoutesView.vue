<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus, Upload } from '@element-plus/icons-vue'

import NeuroAgentPageShell from '@/views/components/NeuroAgentPageShell.vue'

import { createAiResource, fetchAiResource, importAiRoutes } from '../api'
import { emptyPage, includesKeyword, modelTypeText, numberText, rowId, statusType, text, type AiRow } from './viewHelpers'
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
const importVisible = ref(false)

const strategyOptions = computed(() => Array.from(new Set(routes.value.items.map((row) => String(row.strategy || '')).filter(Boolean))))
const filtered = computed(() => routes.value.items.filter((row) => {
  const matchesStrategy = strategyFilter.value === 'all' || row.strategy === strategyFilter.value
  return matchesStrategy && includesKeyword(row, keyword.value, ['route_name', 'route_code', 'capability_code', 'model_type', 'strategy'])
}))

function routePool(routeId: unknown) {
  return routeModels.value.items.filter((row) => String(row.base_route_id || '') === String(routeId || ''))
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
  if (account === '自动选择账号' && api === '自动选择 API') return '执行端点：按健康、能力和账号状态自动选择'
  return `执行端点：${account} / ${api}`
}

function referenceCount(routeId: unknown) {
  const id = String(routeId || '')
  return scenarios.value.items.filter((row) => String(row.default_base_route_id || '') === id).length
    + strategies.value.items.filter((row) => String(row.default_base_route_id || '') === id || String(row.override_base_route_id || '') === id).length
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

async function importRoutes(payload: AiRow) {
  const result = await importAiRoutes(payload)
  importVisible.value = false
  ElMessage.success(`已导入基础路由 ${result.base_routes} 条、模型节点 ${result.route_models} 个`)
  await loadData()
}

onMounted(loadData)
</script>

<template>
  <NeuroAgentPageShell class="ai-prototype" :show-hero="false">
    <template #title>基础路由</template>
    <template #subtitle>基础路由只定义可复用模型策略，不绑定租户、不绑定具体 AI 场景。</template>
    <template #actions>
      <el-button :icon="Upload" @click="importVisible = true">导入基础路由</el-button>
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
          <el-button :icon="Upload" @click="importVisible = true">导入基础路由</el-button>
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
            <el-option v-for="strategy in strategyOptions" :key="strategy" :label="strategy" :value="strategy" />
          </el-select>
        </div>
        <el-table :data="filtered" border v-loading="loading">
          <el-table-column label="基础路由" min-width="220"><template #default="{ row }"><span class="ai-table-cell-main"><strong>{{ text(row.route_name) }}</strong><small>{{ text(row.route_code) }}</small></span></template></el-table-column>
          <el-table-column label="能力 / 类型" min-width="180"><template #default="{ row }">{{ text(row.capability_code) }} / {{ modelTypeText(row.model_type) }}</template></el-table-column>
          <el-table-column label="策略" width="130"><template #default="{ row }"><el-tag>{{ text(row.strategy) }}</el-tag></template></el-table-column>
          <el-table-column label="模型池" min-width="280">
            <template #default="{ row }">
              <span class="ai-table-cell-main">
                <small v-for="node in routePool(row.id)" :key="rowId(node)">{{ text(node.role) }} · {{ modelName(node.model_id) }} · P{{ text(node.priority) }} · W{{ text(node.weight) }}</small>
              </span>
            </template>
          </el-table-column>
          <el-table-column label="执行端点" min-width="260">
            <template #default="{ row }">
              <span class="ai-table-cell-main">
                <small v-for="node in routePool(row.id)" :key="`endpoint-${rowId(node)}`">{{ endpointLabel(node) }}</small>
              </span>
            </template>
          </el-table-column>
          <el-table-column label="引用" width="90"><template #default="{ row }"><strong>{{ referenceCount(row.id) }}</strong></template></el-table-column>
          <el-table-column label="超时 / 重试" width="140"><template #default="{ row }">{{ numberText(row.timeout_ms) }}ms / {{ numberText(row.max_retry) }}</template></el-table-column>
          <el-table-column label="状态" width="100"><template #default="{ row }"><el-tag :type="statusType(row.status)">{{ text(row.status) }}</el-tag></template></el-table-column>
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
      @submit="createRoute"
    />
    <AiJsonDialog
      v-model="importVisible"
      title="导入基础路由"
      :sample="{ base_routes: [{ route_code: 'route_text_low_cost', route_name: '文本低成本基础路由', capability_code: 'chat_completion', model_type: 'text', strategy: 'fallback', route_models: [{ model_id: rowId(models.items[0]), provider_account_id: '', provider_api_id: '', role: 'primary', priority: 1, weight: 100, max_retry: 1, timeout_ms: 25000 }] }] }"
      @submit="importRoutes"
    />
  </NeuroAgentPageShell>
</template>
