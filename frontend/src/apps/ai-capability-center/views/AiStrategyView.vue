<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus, Refresh, Upload } from '@element-plus/icons-vue'

import NeuroAgentPageShell from '@/views/components/NeuroAgentPageShell.vue'

import { createAiResource, fetchAiResource, importAiTenantStrategies } from '../api'
import { emptyPage, includesKeyword, numberText, rowId, statusType, text, type AiRow } from './viewHelpers'
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
const createVisible = ref(false)
const importVisible = ref(false)

const filtered = computed(() => policies.value.items.filter((row) => includesKeyword(row, keyword.value, ['policy_name', 'tenant_ids', 'app_code', 'ai_scenario_code', 'description'])))
const overrideCount = computed(() => policies.value.items.filter((row) => Boolean(row.override_base_route_id)).length)
const warningCount = computed(() => quotaRules.value.items.filter((row) => Number(row.used_amount ?? 0) / Math.max(Number(row.quota_limit ?? 1), 1) * 100 >= Number(row.warning_threshold ?? 80)).length)

function routeName(routeId: unknown) {
  const route = routes.value.items.find((row) => rowId(row) === String(routeId || ''))
  return route ? text(route.route_name) : '-'
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
    const [policyPage, quotaPage, ratePage, routePage] = await Promise.all([
      fetchAiResource('tenant-strategies', { skip: 0, limit: 200, keyword: keyword.value.trim() }),
      fetchAiResource('quota-rules', { skip: 0, limit: 200 }),
      fetchAiResource('rate-limit-rules', { skip: 0, limit: 200 }),
      fetchAiResource('base-routes', { skip: 0, limit: 200 }),
    ])
    policies.value = policyPage
    quotaRules.value = quotaPage
    rateRules.value = ratePage
    routes.value = routePage
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

async function importPolicies(payload: AiRow) {
  const result = await importAiTenantStrategies(payload)
  importVisible.value = false
  ElMessage.success(`已导入策略 ${result.policies} 条、配额 ${result.quota_rules} 条、限流 ${result.rate_limit_rules} 条`)
  await loadData()
}

onMounted(loadData)
</script>

<template>
  <NeuroAgentPageShell class="ai-prototype">
    <template #title>策略中心</template>
    <template #subtitle>一条租户策略统一表达路由覆盖、多维配额、多维限流和超限动作。</template>
    <template #actions>
      <el-button :icon="Refresh" :loading="loading" @click="loadData">刷新</el-button>
      <el-button :icon="Upload" @click="importVisible = true">导入策略</el-button>
      <el-button type="primary" :icon="Plus" @click="createVisible = true">新增策略</el-button>
    </template>

    <section class="ai-card ai-table">
      <header class="ai-card__header">
        <div>
          <h3>策略配置中心</h3>
          <p class="ai-card__description">以租户策略为主对象，不再把配额/限流拆成独立主页面。</p>
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
          <el-table-column label="租户范围" min-width="170"><template #default="{ row }"><span class="ai-table-cell-main"><strong>{{ text(row.tenant_scope) }}</strong><small>{{ text(row.tenant_ids) }}</small></span></template></el-table-column>
          <el-table-column label="应用 / AI 场景" min-width="210"><template #default="{ row }">{{ text(row.app_code) }} / {{ text(row.ai_scenario_code) }}</template></el-table-column>
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
          <el-table-column label="状态" width="100"><template #default="{ row }"><el-tag :type="statusType(row.status)">{{ text(row.status) }}</el-tag></template></el-table-column>
          <el-table-column label="操作" width="140"><template #default="{ row }"><AiResourceActions resource="tenant-strategies" :row="row" @saved="loadData" /></template></el-table-column>
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
      :sample="{ policy_name: '重点租户商品文案策略', tenant_scope: 'include', tenant_ids: ['tenant-a'], app_code: 'product_center', app_name: '商品中心', ai_scenario_code: 'product_copy_generate', ai_scenario_name: '商品文案生成', default_base_route_id: rowId(routes.items[0]), override_base_route_id: '', status: 'active' }"
      @submit="createPolicy"
    />
    <AiJsonDialog
      v-model="importVisible"
      title="导入策略"
      :sample="{ policies: [{ policy_name: '重点租户商品文案策略', tenant_scope: 'include', tenant_ids: ['tenant-a'], app_code: 'product_center', ai_scenario_code: 'product_copy_generate', default_base_route_id: rowId(routes.items[0]), status: 'active' }] }"
      @submit="importPolicies"
    />
  </NeuroAgentPageShell>
</template>
