<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus, Refresh, Upload } from '@element-plus/icons-vue'

import NeuroAgentPageShell from '@/views/components/NeuroAgentPageShell.vue'

import { createAiResource, fetchAiResource, importAiModels } from '../api'
import { compactJoin, emptyPage, includesKeyword, modelTypeText, moneyText, numberText, rowId, statusType, text, type AiRow } from './viewHelpers'
import AiJsonDialog from './AiJsonDialog.vue'
import AiResourceActions from './AiResourceActions.vue'
import './aiPrototype.css'

defineOptions({ name: 'AiModelsView' })

const loading = ref(false)
const keyword = ref('')
const typeFilter = ref('all')
const providers = ref(emptyPage())
const models = ref(emptyPage())
const policies = ref(emptyPage())
const tiers = ref(emptyPage())
const selectedProviderId = ref('')
const selectedModelId = ref('')
const createVisible = ref(false)
const importVisible = ref(false)

const selectedProvider = computed(() => providers.value.items.find((row) => rowId(row) === selectedProviderId.value) ?? providers.value.items[0])
const providerModels = computed(() => models.value.items.filter((row) => String(row.provider_id || '') === rowId(selectedProvider.value)))
const currentModels = computed(() => providerModels.value.filter((row) => {
  const matchesType = typeFilter.value === 'all' || row.model_type === typeFilter.value
  return matchesType && includesKeyword(row, keyword.value, ['model_name', 'model_code', 'model_type', 'capabilities'])
}))
const selectedModel = computed(() => currentModels.value.find((row) => rowId(row) === selectedModelId.value) ?? currentModels.value[0])
const selectedPolicies = computed(() => policies.value.items.filter((row) => String(row.model_id || '') === rowId(selectedModel.value)))
const selectedTiers = computed(() => tiers.value.items.filter((row) => selectedPolicies.value.some((policy) => rowId(policy) === String(row.price_policy_id || ''))))
const modelTypes = computed(() => Array.from(new Set(models.value.items.map((row) => String(row.model_type || '')).filter(Boolean))))
const avgLatency = computed(() => {
  if (!providerModels.value.length) return 0
  return Math.round(providerModels.value.reduce((sum, row) => sum + Number(row.latency_p95 ?? 0), 0) / providerModels.value.length)
})

function providerStats(providerId: string) {
  const rows = models.value.items.filter((row) => String(row.provider_id || '') === providerId)
  return { total: rows.length, active: rows.filter((row) => row.status === 'active').length, types: new Set(rows.map((row) => row.model_type)).size }
}

async function loadData() {
  loading.value = true
  try {
    const [providerPage, modelPage, policyPage, tierPage] = await Promise.all([
      fetchAiResource('providers', { skip: 0, limit: 200 }),
      fetchAiResource('models', { skip: 0, limit: 200 }),
      fetchAiResource('price-policies', { skip: 0, limit: 200 }),
      fetchAiResource('price-tiers', { skip: 0, limit: 200 }),
    ])
    providers.value = providerPage
    models.value = modelPage
    policies.value = policyPage
    tiers.value = tierPage
    selectedProviderId.value ||= rowId(providerPage.items[0])
    selectedModelId.value ||= rowId(modelPage.items[0])
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '模型目录加载失败')
  } finally {
    loading.value = false
  }
}

async function createModel(payload: AiRow) {
  await createAiResource('models', { ...payload, provider_id: payload.provider_id || rowId(selectedProvider.value) })
  createVisible.value = false
  ElMessage.success('模型已创建')
  await loadData()
}

async function importModels(payload: AiRow) {
  const result = await importAiModels(payload)
  importVisible.value = false
  ElMessage.success(`已导入模型 ${result.models} 个、价格策略 ${result.price_policies} 个、分档 ${result.price_tiers} 个`)
  await loadData()
}

onMounted(loadData)
</script>

<template>
  <NeuroAgentPageShell class="ai-prototype" :show-hero="false">
    <template #title>模型目录</template>
    <template #subtitle>以供应商为单位维护模型；业务侧只消费模型能力，不直接维护供应商资源。</template>
    <template #actions>
      <el-button :icon="Refresh" :loading="loading" @click="loadData">刷新</el-button>
      <el-button :icon="Upload" @click="importVisible = true">导入模型</el-button>
      <el-button type="primary" :icon="Plus" @click="createVisible = true">新增模型</el-button>
    </template>

    <section class="ai-card">
      <header class="ai-card__header">
        <div>
          <h3>模型目录</h3>
          <p class="ai-card__description">按供应商、模型类型、能力标签和质量指标管理可路由模型。</p>
        </div>
        <div class="ai-actions">
          <el-button :icon="Refresh" :loading="loading" @click="loadData">刷新</el-button>
          <el-button :icon="Upload" @click="importVisible = true">导入模型</el-button>
          <el-button type="primary" :icon="Plus" @click="createVisible = true">新增模型</el-button>
        </div>
      </header>
      <div class="ai-card__body">
        <div class="ai-provider-layout">
          <aside class="ai-provider-rail">
            <button
              v-for="provider in providers.items"
              :key="rowId(provider)"
              class="ai-provider-item"
              :class="{ active: rowId(provider) === rowId(selectedProvider) }"
              @click="selectedProviderId = rowId(provider); typeFilter = 'all'"
            >
              <strong>{{ text(provider.name) }}</strong>
              <small>{{ text(provider.code) }}</small>
              <span>{{ providerStats(rowId(provider)).total }} 个模型 · {{ providerStats(rowId(provider)).active }} 启用</span>
            </button>
          </aside>

          <section>
            <div class="ai-metric-grid">
              <div class="ai-mini-stat"><span>当前供应商</span><strong>{{ text(selectedProvider?.name) }}</strong></div>
              <div class="ai-mini-stat"><span>模型数量</span><strong>{{ providerModels.length }}</strong></div>
              <div class="ai-mini-stat"><span>启用模型</span><strong>{{ providerModels.filter((row) => row.status === 'active').length }}</strong></div>
              <div class="ai-mini-stat"><span>平均 P95</span><strong>{{ avgLatency }}ms</strong></div>
            </div>

            <div class="ai-toolbar">
              <el-input v-model="keyword" clearable placeholder="搜索当前供应商下的模型、code" />
              <div class="ai-segmented">
                <button :class="{ active: typeFilter === 'all' }" @click="typeFilter = 'all'">全部</button>
                <button v-for="modelType in modelTypes" :key="modelType" :class="{ active: typeFilter === modelType }" @click="typeFilter = modelType">{{ modelTypeText(modelType) }}</button>
              </div>
            </div>

            <div class="ai-table">
              <el-table :data="currentModels" border v-loading="loading" @row-click="(row: AiRow) => selectedModelId = rowId(row)">
                <el-table-column label="模型" min-width="210"><template #default="{ row }"><span class="ai-table-cell-main"><strong>{{ text(row.model_name) }}</strong><small>{{ text(row.model_code) }}</small></span></template></el-table-column>
                <el-table-column label="类型" width="120"><template #default="{ row }"><el-tag>{{ modelTypeText(row.model_type) }}</el-tag></template></el-table-column>
                <el-table-column label="能力标签" min-width="180"><template #default="{ row }">{{ text(row.capabilities) }}</template></el-table-column>
                <el-table-column label="上下文" width="120"><template #default="{ row }">{{ numberText(row.context_window) }}</template></el-table-column>
                <el-table-column label="质量" width="140"><template #default="{ row }">{{ text(row.success_rate) }}% / {{ text(row.latency_p95) }}ms</template></el-table-column>
                <el-table-column label="状态" width="110"><template #default="{ row }"><el-tag :type="statusType(row.status)">{{ text(row.status) }}</el-tag></template></el-table-column>
                <el-table-column label="操作" width="140"><template #default="{ row }"><AiResourceActions resource="models" :row="row" @saved="loadData" /></template></el-table-column>
              </el-table>
            </div>
          </section>
        </div>
      </div>
    </section>

    <div class="ai-rule-grid">
      <section class="ai-card ai-table">
        <header class="ai-card__header">
          <div>
            <h3>价格策略</h3>
            <p class="ai-card__description">{{ text(selectedModel?.model_name) }} 的成本、销售和平台消费单位。</p>
          </div>
        </header>
        <div class="ai-card__body">
          <el-table :data="selectedPolicies" border>
            <el-table-column label="计费项" min-width="180"><template #default="{ row }"><strong>{{ text(row.feature_name || row.feature_key) }}</strong></template></el-table-column>
            <el-table-column label="模式 / 单位" min-width="160"><template #default="{ row }">{{ compactJoin([row.billing_mode, row.billing_unit, row.platform_unit]) }}</template></el-table-column>
            <el-table-column label="成本 / 销售" width="150"><template #default="{ row }">{{ moneyText(row.base_cost_price) }} / {{ moneyText(row.base_sale_price) }}</template></el-table-column>
            <el-table-column label="操作" width="140"><template #default="{ row }"><AiResourceActions resource="price-policies" :row="row" @saved="loadData" /></template></el-table-column>
          </el-table>
        </div>
      </section>

      <section class="ai-card ai-table">
        <header class="ai-card__header">
          <div>
            <h3>分档价格</h3>
            <p class="ai-card__description">同步、异步、质量档位和平台点数换算。</p>
          </div>
        </header>
        <div class="ai-card__body">
          <el-table :data="selectedTiers" border>
            <el-table-column label="档位" min-width="150"><template #default="{ row }"><strong>{{ text(row.tier_name) }}</strong></template></el-table-column>
            <el-table-column label="模式 / 区间" min-width="180"><template #default="{ row }">{{ compactJoin([row.mode, row.min_quantity, row.max_quantity]) }}</template></el-table-column>
            <el-table-column label="成本 / 销售" width="150"><template #default="{ row }">{{ moneyText(row.cost_price) }} / {{ moneyText(row.sale_price) }}</template></el-table-column>
            <el-table-column label="操作" width="140"><template #default="{ row }"><AiResourceActions resource="price-tiers" :row="row" @saved="loadData" /></template></el-table-column>
          </el-table>
        </div>
      </section>
    </div>

    <AiJsonDialog
      v-model="createVisible"
      title="新增模型"
      :sample="{ provider_id: rowId(selectedProvider), model_code: 'gpt-4.1', model_name: 'GPT 4.1', model_type: 'text', capabilities: ['chat_completion'], context_window: 128000, unit: 'tokens', status: 'active', latency_p95: 900, success_rate: 99.5 }"
      @submit="createModel"
    />
    <AiJsonDialog
      v-model="importVisible"
      title="导入模型和价格"
      tip="支持 models/price_policies/price_tiers 一次性导入。"
      :sample="{ models: [{ provider_id: rowId(selectedProvider), model_code: 'gpt-4.1', model_name: 'GPT 4.1', model_type: 'text', capabilities: ['chat_completion'], context_window: 128000, status: 'active' }] }"
      @submit="importModels"
    />
  </NeuroAgentPageShell>
</template>
