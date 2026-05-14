<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus, Refresh, Upload } from '@element-plus/icons-vue'

import NeuroAgentPageShell from '@/views/components/NeuroAgentPageShell.vue'

import { createAiResource, fetchAiResource, importAiScenarios } from '../api'
import { emptyPage, includesKeyword, rowId, statusType, text, type AiRow } from './viewHelpers'
import AiJsonDialog from './AiJsonDialog.vue'
import AiResourceActions from './AiResourceActions.vue'
import './aiPrototype.css'

defineOptions({ name: 'AiScenariosView' })

const loading = ref(false)
const keyword = ref('')
const activeApp = ref('all')
const scenarios = ref(emptyPage())
const capabilities = ref(emptyPage())
const routes = ref(emptyPage())
const strategies = ref(emptyPage())
const createVisible = ref(false)
const importVisible = ref(false)

const appGroups = computed(() => {
  const map = new Map<string, { app_code: string; app_name: string; total: number; active: number }>()
  for (const row of scenarios.value.items) {
    const code = String(row.app_code || '')
    if (!code) continue
    const group = map.get(code) ?? { app_code: code, app_name: String(row.app_name || code), total: 0, active: 0 }
    group.total += 1
    if (row.status === 'active') group.active += 1
    map.set(code, group)
  }
  return Array.from(map.values())
})
const filtered = computed(() => scenarios.value.items.filter((row) => {
  const matchesApp = activeApp.value === 'all' || row.app_code === activeApp.value
  return matchesApp && includesKeyword(row, keyword.value, ['app_code', 'app_name', 'ai_scenario_code', 'ai_scenario_name', 'capability_code', 'owner'])
}))

function routeName(routeId: unknown) {
  const route = routes.value.items.find((item) => rowId(item) === String(routeId || ''))
  return route ? `${text(route.route_name)} / ${text(route.route_code)}` : '-'
}

function capabilityName(code: unknown) {
  const capability = capabilities.value.items.find((item) => item.capability_code === code)
  return text(capability?.capability_name || code)
}

function strategyCount(row: AiRow) {
  return strategies.value.items.filter((item) => item.app_code === row.app_code && item.ai_scenario_code === row.ai_scenario_code).length
}

async function loadData() {
  loading.value = true
  try {
    const [scenarioPage, capabilityPage, routePage, strategyPage] = await Promise.all([
      fetchAiResource('scenarios', { skip: 0, limit: 200, keyword: keyword.value.trim() }),
      fetchAiResource('capabilities', { skip: 0, limit: 200 }),
      fetchAiResource('base-routes', { skip: 0, limit: 200 }),
      fetchAiResource('tenant-strategies', { skip: 0, limit: 200 }),
    ])
    scenarios.value = scenarioPage
    capabilities.value = capabilityPage
    routes.value = routePage
    strategies.value = strategyPage
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : 'AI 场景加载失败')
  } finally {
    loading.value = false
  }
}

async function createScenario(payload: AiRow) {
  await createAiResource('scenarios', payload)
  createVisible.value = false
  ElMessage.success('AI 场景已创建')
  await loadData()
}

async function importScenarios(payload: AiRow) {
  const result = await importAiScenarios(payload)
  importVisible.value = false
  ElMessage.success(`已导入场景 ${result.scenarios} 个`)
  await loadData()
}

onMounted(loadData)
</script>

<template>
  <NeuroAgentPageShell class="ai-prototype">
    <template #title>AI 场景</template>
    <template #subtitle>业务使用模型能力前，必须注册 app_code + ai_scenario_code，并绑定默认基础路由。</template>
    <template #actions>
      <el-button :icon="Refresh" :loading="loading" @click="loadData">刷新</el-button>
      <el-button :icon="Upload" @click="importVisible = true">导入场景</el-button>
      <el-button type="primary" :icon="Plus" @click="createVisible = true">新增场景</el-button>
    </template>

    <section class="ai-card">
      <header class="ai-card__header">
        <div>
          <h3>AI 场景注册中心</h3>
          <p class="ai-card__description">Agent 只作为场景类型，工具编排、人工确认和审批由 Agent 工厂维护。</p>
        </div>
      </header>
      <div class="ai-card__body">
        <div class="ai-metric-grid">
          <div class="ai-mini-stat"><span>场景定义</span><strong>{{ scenarios.total }}</strong></div>
          <div class="ai-mini-stat"><span>已启用</span><strong>{{ scenarios.items.filter((row) => row.status === 'active').length }}</strong></div>
          <div class="ai-mini-stat"><span>Agent 场景</span><strong>{{ scenarios.items.filter((row) => row.scenario_type === 'agent').length }}</strong></div>
          <div class="ai-mini-stat"><span>租户覆盖</span><strong>{{ strategies.total }}</strong></div>
        </div>

        <div class="ai-scenario-layout">
          <aside class="ai-provider-rail">
            <button class="ai-provider-item" :class="{ active: activeApp === 'all' }" @click="activeApp = 'all'">
              <strong>全部应用</strong>
              <span>{{ scenarios.total }} 个场景</span>
            </button>
            <button v-for="app in appGroups" :key="app.app_code" class="ai-provider-item" :class="{ active: activeApp === app.app_code }" @click="activeApp = app.app_code">
              <strong>{{ app.app_name }}</strong>
              <small>{{ app.app_code }}</small>
              <span>{{ app.total }} 个场景 · {{ app.active }} 启用</span>
            </button>
          </aside>
          <section class="ai-table">
            <div class="ai-toolbar">
              <el-input v-model="keyword" clearable placeholder="搜索应用、场景 code、基础路由、所需能力、负责人" />
              <span class="ai-tag">app_code + ai_scenario_code + capability + default_base_route</span>
            </div>
            <el-table :data="filtered" border v-loading="loading">
              <el-table-column label="AI 场景" min-width="220"><template #default="{ row }"><span class="ai-table-cell-main"><strong>{{ text(row.ai_scenario_name) }}</strong><small>{{ text(row.ai_scenario_code) }}</small></span></template></el-table-column>
              <el-table-column label="应用" min-width="170"><template #default="{ row }"><span class="ai-table-cell-main"><strong>{{ text(row.app_name) }}</strong><small>{{ text(row.app_code) }}</small></span></template></el-table-column>
              <el-table-column label="场景类型" width="120"><template #default="{ row }"><el-tag>{{ text(row.scenario_type) }}</el-tag></template></el-table-column>
              <el-table-column label="所需能力" min-width="170"><template #default="{ row }">{{ capabilityName(row.capability_code) }}</template></el-table-column>
              <el-table-column label="默认基础路由" min-width="230"><template #default="{ row }">{{ routeName(row.default_base_route_id) }}</template></el-table-column>
              <el-table-column label="租户覆盖" width="110"><template #default="{ row }"><strong>{{ strategyCount(row) }}</strong></template></el-table-column>
              <el-table-column label="状态" width="100"><template #default="{ row }"><el-tag :type="statusType(row.status)">{{ text(row.status) }}</el-tag></template></el-table-column>
              <el-table-column label="操作" width="140"><template #default="{ row }"><AiResourceActions resource="scenarios" :row="row" @saved="loadData" /></template></el-table-column>
            </el-table>
          </section>
        </div>
      </div>
    </section>

    <section class="ai-card">
      <header class="ai-card__header">
        <div>
          <h3>调用标准</h3>
          <p class="ai-card__description">业务侧不传模型 ID，也不传路由 ID；只传租户、应用和场景，平台内部决策最终路由。</p>
        </div>
      </header>
      <div class="ai-card__body">
        <pre class="ai-code-block">POST /api/ai-gateway/v1/chat
{
  "tenantId": "tenant-hd-flagship",
  "appCode": "product_center",
  "aiScenarioCode": "product_copy_generate",
  "input": { "productName": "夏季防晒衣" }
}</pre>
      </div>
    </section>

    <AiJsonDialog
      v-model="createVisible"
      title="新增 AI 场景"
      :sample="{ app_code: 'product_center', app_name: '商品中心', ai_scenario_code: 'product_copy_generate', ai_scenario_name: '商品文案生成', scenario_type: 'text', capability_code: 'chat_completion', model_type: 'text', default_base_route_id: rowId(routes.items[0]), owner: '商品平台组', version: 'v1.0', status: 'active' }"
      @submit="createScenario"
    />
    <AiJsonDialog
      v-model="importVisible"
      title="导入 AI 场景"
      :sample="{ scenarios: [{ app_code: 'product_center', app_name: '商品中心', ai_scenario_code: 'product_copy_generate', ai_scenario_name: '商品文案生成', capability_code: 'chat_completion', model_type: 'text', default_base_route_id: rowId(routes.items[0]), owner: '商品平台组', status: 'active' }] }"
      @submit="importScenarios"
    />
  </NeuroAgentPageShell>
</template>
