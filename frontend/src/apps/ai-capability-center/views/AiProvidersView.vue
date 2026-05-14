<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus, Upload } from '@element-plus/icons-vue'

import NeuroAgentPageShell from '@/views/components/NeuroAgentPageShell.vue'

import { checkAiProviderAPIConnectivity, createAiResource, fetchAiResource, importAiProviders } from '../api'
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
const connectivityChecking = ref(false)

const filteredProviders = computed(() => providers.value.items.filter((row) => includesKeyword(row, keyword.value, ['name', 'code', 'base_url', 'owner'])))
const selectedProvider = computed(() => providers.value.items.find((row) => rowId(row) === selectedProviderId.value))
const currentAccounts = computed(() => {
  if (!selectedProviderId.value) return accounts.value.items
  return accounts.value.items.filter((row) => String(row.provider_id || '') === selectedProviderId.value)
})
const currentApis = computed(() => apis.value.items.filter((row) => {
  const matchesProvider = !selectedProviderId.value || String(row.provider_id || '') === selectedProviderId.value
  const matchesAccount = !selectedAccountId.value || String(row.account_id || '') === selectedAccountId.value
  return matchesProvider && matchesAccount
}))
const activeApiQps = computed(() => currentApis.value.reduce((sum, row) => sum + Number(row.qps_limit ?? 0), 0))
const activeAccountCount = computed(() => currentAccounts.value.filter((row) => row.status === 'active').length)
const activeApiCount = computed(() => currentApis.value.filter((row) => row.status === 'active').length)
const accountQuotaTotal = computed(() => currentAccounts.value.reduce((sum, row) => sum + Number(row.quota_limit ?? 0), 0))
const selectedAccount = computed(() => currentAccounts.value.find((row) => rowId(row) === selectedAccountId.value))
const accountScopeTitle = computed(() => {
  if (selectedProvider.value) return `${text(selectedProvider.value.name)}的接入账号`
  return '所有接入账号'
})
const apiScopeTitle = computed(() => {
  if (selectedAccount.value) return `${text(selectedAccount.value.account_name)} 的 API`
  if (selectedProvider.value) return `${text(selectedProvider.value.name)}的 API`
  return '所有 API'
})

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
    if (selectedProviderId.value && !providerPage.items.some((row) => rowId(row) === selectedProviderId.value)) {
      selectedProviderId.value = ''
      selectedAccountId.value = ''
    }
    if (selectedAccountId.value && !accountPage.items.some((row) => rowId(row) === selectedAccountId.value)) {
      selectedAccountId.value = ''
    }
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '供应商加载失败')
  } finally {
    loading.value = false
  }
}

function selectProvider(providerId: string) {
  selectedProviderId.value = selectedProviderId.value === providerId ? '' : providerId
  selectedAccountId.value = ''
}

function clearProvider() {
  selectedProviderId.value = ''
  selectedAccountId.value = ''
}

function selectAccount(accountId: string) {
  selectedAccountId.value = selectedAccountId.value === accountId ? '' : accountId
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

async function checkConnectivity() {
  connectivityChecking.value = true
  try {
    const result = await checkAiProviderAPIConnectivity()
    ElMessage.success(`连通性检测完成：正常 ${result.active} 个，告警 ${result.warning} 个，异常 ${result.error} 个`)
    await loadData()
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '连通性检测失败')
  } finally {
    connectivityChecking.value = false
  }
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

    <div class="ai-provider-workspace">
      <aside class="ai-provider-sidebar ai-card">
        <header class="ai-card__header">
          <div>
            <h3>供应商</h3>
            <p class="ai-card__description">左侧维护供应商，右侧联动账号和 API。</p>
          </div>
        </header>
        <div class="ai-card__body">
          <div class="ai-toolbar ai-provider-toolbar">
            <el-input v-model="keyword" clearable placeholder="搜索供应商、code、endpoint" @change="loadData" />
            <button class="ai-ghost-button" :class="{ active: !selectedProviderId }" @click="clearProvider">全部</button>
          </div>
          <div v-loading="loading" class="ai-provider-card-list">
            <div
              v-for="provider in filteredProviders"
              :key="rowId(provider)"
              class="ai-provider-card"
              :class="{ active: selectedProviderId === rowId(provider) }"
              role="button"
              tabindex="0"
              @click="selectProvider(rowId(provider))"
              @keydown.enter.prevent="selectProvider(rowId(provider))"
              @keydown.space.prevent="selectProvider(rowId(provider))"
            >
              <span class="ai-provider-card__top">
                <span class="ai-provider-card__name">
                  <strong>{{ text(provider.name) }}</strong>
                  <small>{{ text(provider.code) }}</small>
                </span>
                <el-tag :type="statusType(provider.status)" size="small">{{ text(provider.status) }}</el-tag>
              </span>
              <span class="ai-provider-card__endpoint">{{ text(provider.base_url) }}</span>
              <span class="ai-provider-card__meta">
                <span>{{ providerStats(rowId(provider)).accountCount }} 账号</span>
                <span>{{ providerStats(rowId(provider)).apiCount }} API</span>
                <span>{{ providerStats(rowId(provider)).activeApiCount }} 可用</span>
              </span>
              <span class="ai-provider-card__bottom">
                <span>{{ numberText(provider.qps_limit) }} QPS</span>
                <span>{{ moneyText(provider.monthly_budget) }}</span>
              </span>
              <span class="ai-provider-card__actions" @click.stop>
                <AiResourceActions resource="providers" :row="provider" @saved="loadData" />
              </span>
            </div>
            <el-empty v-if="!loading && !filteredProviders.length" description="暂无供应商" />
          </div>
        </div>
      </aside>

      <main class="ai-provider-main">
        <section class="ai-card ai-table">
          <header class="ai-card__header">
            <div>
              <h3>{{ accountScopeTitle }}</h3>
              <p class="ai-card__description">默认显示所有账号；选择左侧供应商后，只显示该供应商账号。</p>
            </div>
            <button v-if="selectedProviderId" class="ai-ghost-button" @click="clearProvider">查看全部供应商</button>
          </header>
          <div class="ai-card__body">
            <div class="ai-metric-grid ai-provider-summary-grid">
              <div class="ai-mini-stat"><span>账号总数</span><strong>{{ currentAccounts.length }}</strong></div>
              <div class="ai-mini-stat"><span>启用账号</span><strong>{{ activeAccountCount }}</strong></div>
              <div class="ai-mini-stat"><span>API 数量</span><strong>{{ currentApis.length }}</strong></div>
              <div class="ai-mini-stat"><span>额度合计</span><strong>{{ moneyText(accountQuotaTotal) }}</strong></div>
            </div>
            <el-table :data="currentAccounts" border v-loading="loading" row-key="id" @row-click="(row: AiRow) => selectAccount(rowId(row))">
              <el-table-column label="账号" min-width="180">
                <template #default="{ row }">
                  <span class="ai-table-cell-main">
                    <strong>{{ text(row.account_name) }}</strong>
                    <small>{{ text(row.key_alias) }}</small>
                  </span>
                </template>
              </el-table-column>
              <el-table-column label="Endpoint" min-width="220"><template #default="{ row }">{{ text(row.endpoint) }}</template></el-table-column>
              <el-table-column label="额度使用" width="180"><template #default="{ row }">{{ numberText(row.used_quota) }} / {{ numberText(row.quota_limit) }}</template></el-table-column>
              <el-table-column label="状态" width="110"><template #default="{ row }"><el-tag :type="statusType(row.status)">{{ text(row.status) }}</el-tag></template></el-table-column>
              <el-table-column label="操作" width="190" fixed="right">
                <template #default="{ row }">
                  <div class="ai-row-actions">
                    <el-button link type="primary" @click.stop="selectAccount(rowId(row))">{{ selectedAccountId === rowId(row) ? '取消筛选' : '筛选 API' }}</el-button>
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
              <h3>{{ apiScopeTitle }}</h3>
              <p class="ai-card__description">默认显示所有 API；选择账号后，只显示该账号挂载的 API。</p>
            </div>
            <div class="ai-segmented">
              <button :disabled="connectivityChecking" @click="checkConnectivity">{{ connectivityChecking ? '检测中' : '检测连通性' }}</button>
              <button :class="{ active: !selectedAccountId }" @click="selectedAccountId = ''">全部账号</button>
              <button v-for="account in currentAccounts" :key="rowId(account)" :class="{ active: selectedAccountId === rowId(account) }" @click="selectAccount(rowId(account))">
                {{ text(account.account_name) }}
              </button>
            </div>
          </header>
          <div class="ai-card__body">
            <div class="ai-metric-grid ai-provider-summary-grid">
              <div class="ai-mini-stat"><span>API 总数</span><strong>{{ currentApis.length }}</strong></div>
              <div class="ai-mini-stat"><span>启用 API</span><strong>{{ activeApiCount }}</strong></div>
              <div class="ai-mini-stat"><span>聚合 QPS</span><strong>{{ activeApiQps }}</strong></div>
              <div class="ai-mini-stat"><span>当前账号</span><strong>{{ selectedAccount ? text(selectedAccount.account_name) : '全部' }}</strong></div>
            </div>
            <el-table :data="currentApis" border v-loading="loading">
              <el-table-column label="API" min-width="220"><template #default="{ row }"><span class="ai-table-cell-main"><strong>{{ text(row.api_name) }}</strong><small>{{ text(row.api_path) }}</small></span></template></el-table-column>
              <el-table-column label="类型" width="120"><template #default="{ row }"><el-tag>{{ text(row.api_type) }}</el-tag></template></el-table-column>
              <el-table-column label="能力" min-width="180"><template #default="{ row }">{{ text(row.capabilities) }}</template></el-table-column>
              <el-table-column label="QPS / 超时" width="150"><template #default="{ row }">{{ numberText(row.qps_limit) }} / {{ numberText(row.timeout_ms) }}ms</template></el-table-column>
              <el-table-column label="连通性" width="180">
                <template #default="{ row }">
                  <span class="ai-table-cell-main">
                    <el-tag :type="statusType(row.health_status)">{{ text(row.health_status, 'unknown') }}</el-tag>
                    <small>{{ text(row.health_message, '未检测') }}</small>
                  </span>
                </template>
              </el-table-column>
              <el-table-column label="状态" width="110"><template #default="{ row }"><el-tag :type="statusType(row.status)">{{ text(row.status) }}</el-tag></template></el-table-column>
              <el-table-column label="操作" width="128" fixed="right"><template #default="{ row }"><AiResourceActions resource="apis" :row="row" @saved="loadData" /></template></el-table-column>
            </el-table>
          </div>
        </section>
      </main>
    </div>

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
      :sample="{ providers: [{ name: 'DeepSeek 官方账号', code: 'deepseek', type: 'public_cloud', base_url: 'https://api.deepseek.com', auth_type: 'api_key', status: 'active', region: 'CN', qps_limit: 500, monthly_budget: 50000, accounts: [{ account_name: 'prod-main', endpoint: 'https://api.deepseek.com', key_alias: 'DEEPSEEK_API_KEY', encrypted_api_key: '', apis: [{ api_name: 'chat.completions', api_path: '/v1/chat/completions', api_type: 'chat', capabilities: ['chat_completion', 'text_generation', 'reasoning'], qps_limit: 260, timeout_ms: 45000 }] }] }, { name: '通义千问 DashScope', code: 'dashscope', type: 'public_cloud', base_url: 'https://dashscope.aliyuncs.com', auth_type: 'dashscope', status: 'active', region: 'CN', qps_limit: 500, monthly_budget: 60000, accounts: [{ account_name: 'prod-main', endpoint: 'https://dashscope.aliyuncs.com', key_alias: 'DASHSCOPE_API_KEY', encrypted_api_key: '', apis: [{ api_name: 'generation', api_path: '/api/v1/services/aigc/text-generation/generation', api_type: 'chat', capabilities: ['chat_completion', 'text_generation'], qps_limit: 260, timeout_ms: 30000 }, { api_name: 'embeddings', api_path: '/api/v1/services/embeddings/text-embedding/text-embedding', api_type: 'embedding', capabilities: ['embedding'], qps_limit: 260, timeout_ms: 15000 }] }] }] }"
      @submit="importProviders"
    />
  </NeuroAgentPageShell>
</template>
