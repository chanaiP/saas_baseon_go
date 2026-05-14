<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus, Upload } from '@element-plus/icons-vue'

import NeuroAgentPageShell from '@/views/components/NeuroAgentPageShell.vue'

import { createAiResource, fetchAiResource, importAiProviders } from '../api'
import { emptyPage, includesKeyword, moneyText, numberText, rowId, statusType, text, type AiRow } from './viewHelpers'
import AiJsonDialog from './AiJsonDialog.vue'
import AiResourceActions from './AiResourceActions.vue'
import './aiPrototype.css'

defineOptions({ name: 'AiProvidersView' })

const loading = ref(false)
const keyword = ref('')
const providers = ref(emptyPage())
const accounts = ref(emptyPage())
const apis = ref(emptyPage())
const selectedProviderId = ref('')
const selectedAccountId = ref('')
const importVisible = ref(false)
const createVisible = ref(false)

const filteredProviders = computed(() => providers.value.items.filter((row) => includesKeyword(row, keyword.value, ['name', 'code', 'base_url', 'owner'])))
const selectedProvider = computed(() => filteredProviders.value.find((row) => rowId(row) === selectedProviderId.value) ?? filteredProviders.value[0])
const currentAccounts = computed(() => accounts.value.items.filter((row) => String(row.provider_id || '') === rowId(selectedProvider.value)))
const currentApis = computed(() => apis.value.items.filter((row) => {
  const matchesProvider = String(row.provider_id || '') === rowId(selectedProvider.value)
  const matchesAccount = !selectedAccountId.value || String(row.account_id || '') === selectedAccountId.value
  return matchesProvider && matchesAccount
}))
const activeApiQps = computed(() => currentApis.value.reduce((sum, row) => sum + Number(row.qps_limit ?? 0), 0))

function providerStats(providerId: string) {
  const accountCount = accounts.value.items.filter((row) => String(row.provider_id || '') === providerId).length
  const apiRows = apis.value.items.filter((row) => String(row.provider_id || '') === providerId)
  return { accountCount, apiCount: apiRows.length, activeApiCount: apiRows.filter((row) => row.status === 'active').length }
}

async function loadData() {
  loading.value = true
  try {
    const [providerPage, accountPage, apiPage] = await Promise.all([
      fetchAiResource('providers', { skip: 0, limit: 200, keyword: keyword.value.trim() }),
      fetchAiResource('accounts', { skip: 0, limit: 200 }),
      fetchAiResource('apis', { skip: 0, limit: 200 }),
    ])
    providers.value = providerPage
    accounts.value = accountPage
    apis.value = apiPage
    selectedProviderId.value ||= rowId(providerPage.items[0])
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '供应商加载失败')
  } finally {
    loading.value = false
  }
}

async function createProvider(payload: AiRow) {
  await createAiResource('providers', payload)
  createVisible.value = false
  ElMessage.success('供应商已创建')
  await loadData()
}

async function importProviders(payload: AiRow) {
  const result = await importAiProviders(payload)
  importVisible.value = false
  ElMessage.success(`已导入供应商 ${result.providers} 个、账号 ${result.accounts} 个、API ${result.apis} 个`)
  await loadData()
}

onMounted(loadData)
</script>

<template>
  <NeuroAgentPageShell class="ai-prototype" :show-hero="false">
    <template #title>供应商</template>
    <template #subtitle>统一维护公有云、私有化和自建模型供应商，并关联接入账号与 API。</template>
    <template #actions>
      <el-button :icon="Upload" @click="importVisible = true">整体导入</el-button>
      <el-button type="primary" :icon="Plus" @click="createVisible = true">新增供应商</el-button>
    </template>

    <section class="ai-card">
      <header class="ai-card__header">
        <div>
          <h3>供应商管理</h3>
          <p class="ai-card__description">统一维护公有云、私有化和自建模型供应商，并关联接入账号与 API。</p>
        </div>
        <div class="ai-actions">
          <el-button :icon="Upload" @click="importVisible = true">整体导入</el-button>
          <el-button type="primary" :icon="Plus" @click="createVisible = true">新增供应商</el-button>
        </div>
      </header>
      <div class="ai-card__body">
        <div class="ai-toolbar">
          <el-input v-model="keyword" clearable placeholder="搜索供应商、code、endpoint" @change="loadData" />
        </div>
        <div class="ai-table">
          <el-table :data="filteredProviders" border v-loading="loading" @row-click="(row: AiRow) => { selectedProviderId = rowId(row); selectedAccountId = '' }">
            <el-table-column label="供应商" min-width="190">
              <template #default="{ row }">
                <span class="ai-table-cell-main">
                  <strong>{{ text(row.name) }}</strong>
                  <small>{{ text(row.code) }}</small>
                </span>
              </template>
            </el-table-column>
            <el-table-column label="Endpoint" min-width="240"><template #default="{ row }">{{ text(row.base_url) }}</template></el-table-column>
            <el-table-column label="鉴权" width="110"><template #default="{ row }"><el-tag>{{ text(row.auth_type) }}</el-tag></template></el-table-column>
            <el-table-column label="账号 / API" width="150">
              <template #default="{ row }">{{ providerStats(rowId(row)).accountCount }} 个账号 / {{ providerStats(rowId(row)).apiCount }} 个 API</template>
            </el-table-column>
            <el-table-column label="QPS / 月预算" width="145">
              <template #default="{ row }">
                <span class="ai-table-cell-main"><strong>{{ numberText(row.qps_limit) }} QPS</strong><small>{{ moneyText(row.monthly_budget) }}</small></span>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="128" fixed="right"><template #default="{ row }"><AiResourceActions resource="providers" :row="row" @saved="loadData" /></template></el-table-column>
          </el-table>
        </div>
      </div>
    </section>

    <section class="ai-card">
      <header class="ai-card__header">
        <div>
          <h3>当前供应商：{{ text(selectedProvider?.name) }}</h3>
          <p class="ai-card__description">先选供应商，再维护该供应商下的接入账号和 API。</p>
        </div>
      </header>
      <div class="ai-card__body">
        <div class="ai-metric-grid">
          <div class="ai-mini-stat"><span>接入账号</span><strong>{{ currentAccounts.length }}</strong></div>
          <div class="ai-mini-stat"><span>API 数量</span><strong>{{ currentApis.length }}</strong></div>
          <div class="ai-mini-stat"><span>API 聚合 QPS</span><strong>{{ activeApiQps }}</strong></div>
          <div class="ai-mini-stat"><span>月预算</span><strong>{{ moneyText(selectedProvider?.monthly_budget) }}</strong></div>
        </div>
      </div>
    </section>

    <section class="ai-card ai-table">
      <header class="ai-card__header">
        <div>
          <h3>接入账号</h3>
          <p class="ai-card__description">一个供应商可配置多个账号，用于不同租户、环境、区域或备用 Key。</p>
        </div>
      </header>
      <div class="ai-card__body">
        <el-table :data="currentAccounts" border v-loading="loading">
          <el-table-column label="账号" min-width="180"><template #default="{ row }"><strong>{{ text(row.account_name) }}</strong></template></el-table-column>
          <el-table-column label="Endpoint" min-width="220"><template #default="{ row }">{{ text(row.endpoint) }}</template></el-table-column>
          <el-table-column label="密钥别名" min-width="160"><template #default="{ row }">{{ text(row.key_alias) }}</template></el-table-column>
          <el-table-column label="额度使用" width="180"><template #default="{ row }">{{ numberText(row.used_quota) }} / {{ numberText(row.quota_limit) }}</template></el-table-column>
          <el-table-column label="状态" width="110"><template #default="{ row }"><el-tag :type="statusType(row.status)">{{ text(row.status) }}</el-tag></template></el-table-column>
          <el-table-column label="操作" width="190" fixed="right">
            <template #default="{ row }">
              <div class="ai-row-actions">
                <el-button link type="primary" @click="selectedAccountId = rowId(row)">筛选 API</el-button>
                <AiResourceActions resource="accounts" :row="row" @saved="loadData" />
              </div>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </section>

    <section class="ai-card ai-table">
      <header class="ai-card__header">
        <div>
          <h3>API 配置</h3>
          <p class="ai-card__description">一个接入账号可挂载多个 API，例如文本、图片、向量、视频、Agent Runtime。</p>
        </div>
        <div class="ai-segmented">
          <button :class="{ active: !selectedAccountId }" @click="selectedAccountId = ''">全部账号</button>
          <button v-for="account in currentAccounts" :key="rowId(account)" :class="{ active: selectedAccountId === rowId(account) }" @click="selectedAccountId = rowId(account)">
            {{ text(account.account_name) }}
          </button>
        </div>
      </header>
      <div class="ai-card__body">
        <el-table :data="currentApis" border v-loading="loading">
          <el-table-column label="API" min-width="220"><template #default="{ row }"><span class="ai-table-cell-main"><strong>{{ text(row.api_name) }}</strong><small>{{ text(row.api_path) }}</small></span></template></el-table-column>
          <el-table-column label="类型" width="120"><template #default="{ row }"><el-tag>{{ text(row.api_type) }}</el-tag></template></el-table-column>
          <el-table-column label="能力" min-width="180"><template #default="{ row }">{{ text(row.capabilities) }}</template></el-table-column>
          <el-table-column label="QPS / 超时" width="150"><template #default="{ row }">{{ numberText(row.qps_limit) }} / {{ numberText(row.timeout_ms) }}ms</template></el-table-column>
          <el-table-column label="状态" width="110"><template #default="{ row }"><el-tag :type="statusType(row.status)">{{ text(row.status) }}</el-tag></template></el-table-column>
          <el-table-column label="操作" width="128" fixed="right"><template #default="{ row }"><AiResourceActions resource="apis" :row="row" @saved="loadData" /></template></el-table-column>
        </el-table>
      </div>
    </section>

    <AiJsonDialog
      v-model="createVisible"
      title="新增供应商"
      :sample="{ name: 'OpenAI', code: 'openai', type: 'public_cloud', base_url: 'https://api.openai.com', auth_type: 'api_key', status: 'active', priority: 80, region: 'US', qps_limit: 100, monthly_budget: 10000, owner: '平台组' }"
      @submit="createProvider"
    />
    <AiJsonDialog
      v-model="importVisible"
      title="整体导入供应商"
      tip="支持 providers/accounts/apis 一次性导入。"
      :sample="{ providers: [{ name: 'OpenAI', code: 'openai', type: 'public_cloud', base_url: 'https://api.openai.com', auth_type: 'api_key', status: 'active', accounts: [{ account_name: 'prod', endpoint: 'https://api.openai.com', key_alias: 'OPENAI_API_KEY', encrypted_api_key: 'ciphertext', apis: [{ api_name: 'chat.completions', api_path: '/v1/chat/completions', api_type: 'chat', capabilities: ['chat_completion'] }] }] }] }"
      @submit="importProviders"
    />
  </NeuroAgentPageShell>
</template>
