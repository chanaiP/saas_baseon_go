<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'

import NeuroAgentPageShell from '@/views/components/NeuroAgentPageShell.vue'
import { fetchTenants } from '@/api/tenant'

import { createAiResource, fetchAiResource } from '../api'
import { emptyPage, includesKeyword, numberText, rowId, statusText, statusType, text, type AiRow } from './viewHelpers'
import AiJsonDialog from './AiJsonDialog.vue'
import AiResourceActions from './AiResourceActions.vue'
import './aiPrototype.css'

defineOptions({ name: 'AiStrategyView' })

const loading = ref(false)
const keyword = ref('')
const policies = ref(emptyPage())
const quotaRules = ref(emptyPage())
const rateRules = ref(emptyPage())
const routes = ref(emptyPage())
const tenants = ref(emptyPage())
const scenarios = ref(emptyPage())
const createVisible = ref(false)

const filtered = computed(() => policies.value.items.filter((row) => includesKeyword(row, keyword.value, ['policy_name', 'tenant_ids', 'app_code', 'ai_scenario_code', 'description'])))
const overrideCount = computed(() => policies.value.items.filter((row) => Boolean(row.override_base_route_id)).length)
const warningCount = computed(() => quotaRules.value.items.filter((row) => Number(row.used_amount ?? 0) / Math.max(Number(row.quota_limit ?? 1), 1) * 100 >= Number(row.warning_threshold ?? 80)).length)
const routeOptions = computed(() => routes.value.items.map((route) => ({
  label: `${text(route.route_name)} / ${text(route.route_code)}`,
  value: rowId(route),
})))
const tenantOptions = computed(() => tenants.value.items.map((tenant) => ({
  label: `${text(tenant.name || tenant.code || tenant.id)} / ${text(tenant.code || tenant.id)}`,
  value: String(tenant.id || tenant.code || ''),
})).filter((option) => option.value))
const appOptions = computed(() => {
  const rows = new Map<string, { label: string, value: string, fill: Record<string, unknown> }>()
  for (const scenario of scenarios.value.items) {
    const appCode = text(scenario.app_code, '')
    if (!appCode || rows.has(appCode)) continue
    rows.set(appCode, {
      label: `${text(scenario.app_name || appCode)} / ${appCode}`,
      value: appCode,
      fill: { app_name: text(scenario.app_name, '') },
    })
  }
  return Array.from(rows.values())
})
const scenarioOptions = computed(() => scenarios.value.items.map((scenario) => ({
  label: `${text(scenario.ai_scenario_name || scenario.ai_scenario_code)} / ${text(scenario.ai_scenario_code)} / ${text(scenario.app_name || scenario.app_code)}`,
  value: text(scenario.ai_scenario_code, ''),
  fill: {
    app_code: text(scenario.app_code, ''),
    app_name: text(scenario.app_name, ''),
    ai_scenario_name: text(scenario.ai_scenario_name, ''),
    default_base_route_id: text(scenario.default_base_route_id, ''),
  },
})).filter((option) => option.value))
const strategyDialogOptions = computed(() => ({
  tenant_ids: tenantOptions.value,
  app_code: appOptions.value,
  ai_scenario_code: scenarioOptions.value,
  default_base_route_id: routeOptions.value,
  override_base_route_id: routeOptions.value,
}))

function routeName(routeId: unknown) {
  const route = routes.value.items.find((row) => rowId(row) === String(routeId || ''))
  return route ? text(route.route_name) : '-'
}

function tenantScopeText(value: unknown) {
  const labels: Record<string, string> = { include: '指定租户', exclude: '排除租户', all: '全部租户' }
  return labels[text(value, '')] || text(value)
}

function tenantNames(value: unknown) {
  if (!Array.isArray(value) || !value.length) return '未选择'
  const names = value.map((id) => {
    const tenant = tenants.value.items.find((row) => String(row.id || row.code || '') === String(id))
    return tenant ? text(tenant.name || tenant.code || id) : String(id)
  })
  return names.join('、')
}

function appScenarioText(row: AiRow) {
  const app = text(row.app_name || row.app_code)
  const scenario = text(row.ai_scenario_name || row.ai_scenario_code)
  return `${app} / ${scenario}`
}

function quotaFor(policyId: unknown) {
  return quotaRules.value.items.filter((row) => String(row.policy_id || '') === String(policyId || '')).slice(0, 2)
}

function rateFor(policyId: unknown) {
  return rateRules.value.items.filter((row) => String(row.policy_id || '') === String(policyId || '')).slice(0, 2)
}

async function loadData() {
  loading.value = true
  try {
    const [policyPage, quotaPage, ratePage, routePage, tenantPage, scenarioPage] = await Promise.all([
      fetchAiResource('tenant-strategies', { skip: 0, limit: 200, keyword: keyword.value.trim() }),
      fetchAiResource('quota-rules', { skip: 0, limit: 200 }),
      fetchAiResource('rate-limit-rules', { skip: 0, limit: 200 }),
      fetchAiResource('base-routes', { skip: 0, limit: 200 }),
      fetchTenants(0, 500),
      fetchAiResource('scenarios', { skip: 0, limit: 500 }),
    ])
    policies.value = policyPage
    quotaRules.value = quotaPage
    rateRules.value = ratePage
    routes.value = routePage
    tenants.value = { ...tenantPage, items: tenantPage.items as unknown as AiRow[] }
    scenarios.value = scenarioPage
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '策略中心加载失败')
  } finally {
    loading.value = false
  }
}

async function createPolicy(payload: AiRow) {
  await createAiResource('tenant-strategies', payload)
  createVisible.value = false
  ElMessage.success('策略已创建')
  await loadData()
}

onMounted(loadData)
</script>

<template>
  <NeuroAgentPageShell class="ai-prototype" :show-hero="false">
    <template #title>策略中心</template>
    <template #subtitle>一条租户策略统一表达路由覆盖、多维配额、多维限流和超限动作。</template>
    <template #actions>
      <el-button type="primary" :icon="Plus" @click="createVisible = true">新增策略</el-button>
    </template>

    <section class="ai-card ai-table">
      <header class="ai-card__header">
        <div>
          <h3>策略配置中心</h3>
          <p class="ai-card__description">以租户策略为主对象，不再把配额/限流拆成独立主页面。</p>
        </div>
        <div class="ai-actions">
          <el-button type="primary" :icon="Plus" @click="createVisible = true">新增策略</el-button>
        </div>
      </header>
      <div class="ai-card__body">
        <div class="ai-metric-grid">
          <div class="ai-mini-stat"><span>租户策略</span><strong>{{ policies.total }}</strong></div>
          <div class="ai-mini-stat"><span>配额规则</span><strong>{{ quotaRules.total }}</strong></div>
          <div class="ai-mini-stat"><span>限流规则</span><strong>{{ rateRules.total }}</strong></div>
          <div class="ai-mini-stat"><span>路由覆盖 / 预警</span><strong>{{ overrideCount }} / {{ warningCount }}</strong></div>
        </div>
        <div class="ai-toolbar">
          <el-input v-model="keyword" clearable placeholder="搜索租户、应用、场景、路由、配额、限流" />
          <span class="ai-tag">策略 = 租户 + AI 场景 + 路由覆盖 + 配额限流</span>
        </div>
        <el-table :data="filtered" border v-loading="loading">
          <el-table-column label="策略" min-width="220"><template #default="{ row }"><span class="ai-table-cell-main"><strong>{{ text(row.policy_name) }}</strong><small>{{ text(row.id) }}</small></span></template></el-table-column>
          <el-table-column label="租户范围" min-width="190"><template #default="{ row }"><span class="ai-table-cell-main"><strong>{{ tenantScopeText(row.tenant_scope) }}</strong><small>{{ tenantNames(row.tenant_ids) }}</small></span></template></el-table-column>
          <el-table-column label="应用 / AI 场景" min-width="230"><template #default="{ row }"><span class="ai-table-cell-main"><strong>{{ appScenarioText(row) }}</strong><small>{{ text(row.app_code) }} / {{ text(row.ai_scenario_code) }}</small></span></template></el-table-column>
          <el-table-column label="路由策略" min-width="180"><template #default="{ row }">{{ routeName(row.override_base_route_id || row.default_base_route_id) }}</template></el-table-column>
          <el-table-column label="配额规则" min-width="220">
            <template #default="{ row }">
              <span class="ai-table-cell-main"><small v-for="rule in quotaFor(row.id)" :key="rowId(rule)">{{ text(rule.dimension) }} {{ numberText(rule.quota_limit) }} {{ text(rule.usage_unit) }}/{{ text(rule.period) }}</small></span>
            </template>
          </el-table-column>
          <el-table-column label="限流规则" min-width="190">
            <template #default="{ row }">
              <span class="ai-table-cell-main"><small v-for="rule in rateFor(row.id)" :key="rowId(rule)">QPS {{ numberText(rule.qps) }} · 并发 {{ numberText(rule.concurrency) }}</small></span>
            </template>
          </el-table-column>
          <el-table-column label="状态" width="100"><template #default="{ row }"><el-tag :type="statusType(row.status)">{{ statusText(row.status) }}</el-tag></template></el-table-column>
          <el-table-column label="操作" width="140"><template #default="{ row }"><AiResourceActions resource="tenant-strategies" :row="row" :options="strategyDialogOptions" @saved="loadData" /></template></el-table-column>
        </el-table>
      </div>
    </section>

    <div class="ai-rule-grid">
      <section class="ai-card">
        <header class="ai-card__header"><div><h3>策略优先级</h3><p class="ai-card__description">没有策略时使用 AI 场景绑定的默认基础路由。</p></div></header>
        <div class="ai-card__body"><pre class="ai-code-block">租户 + 场景策略
  > 租户 + 应用策略
  > 全部租户 + 场景策略
  > AI 场景 default_base_route</pre></div>
      </section>
      <section class="ai-card">
        <header class="ai-card__header"><div><h3>多维度控制</h3><p class="ai-card__description">列表显示规则结果，编辑入口维护规则明细。</p></div></header>
        <div class="ai-card__body ai-health-list">
          <span>租户总额度</span><span>应用额度</span><span>AI 场景额度</span><span>模型额度</span><span>供应商账号限流</span><span>用户级防刷</span>
        </div>
      </section>
    </div>

    <AiJsonDialog
      v-model="createVisible"
      title="新增租户策略"
      :sample="{ policy_name: '重点租户商品文案策略', tenant_scope: 'include', tenant_ids: [], ai_scenario_code: text(scenarios.items[0]?.ai_scenario_code, ''), app_code: text(scenarios.items[0]?.app_code, ''), app_name: text(scenarios.items[0]?.app_name, ''), ai_scenario_name: text(scenarios.items[0]?.ai_scenario_name, ''), default_base_route_id: rowId(routes.items[0]), override_base_route_id: '', status: 'active' }"
      :options="strategyDialogOptions"
      @submit="createPolicy"
    />
  </NeuroAgentPageShell>
</template>
