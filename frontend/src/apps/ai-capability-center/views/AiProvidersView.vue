<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'

import NeuroAgentPageShell from '@/views/components/NeuroAgentPageShell.vue'

import { checkAiProviderAPIConnectivity, createAiResource, fetchAiResource } from '../api'
import { capabilityTagList, emptyPage, includesKeyword, moneyText, numberText, rowId, statusText, statusType, text, type AiRow } from './viewHelpers'
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
const accountScopeAll = ref(true)
const createVisible = ref(false)
const connectivityChecking = ref(false)
const accountPage = ref(1)
const accountPageSize = ref(10)
const apiPage = ref(1)
const apiPageSize = ref(10)

const filteredProviders = computed(() => providers.value.items.filter((row) => includesKeyword(row, keyword.value, ['name', 'code', 'base_url', 'owner'])))
const selectedProvider = computed(() => providers.value.items.find((row) => rowId(row) === selectedProviderId.value))
const currentAccounts = computed(() => accounts.value.items)
const currentApis = computed(() => apis.value.items)
const providerListStats = computed(() => summaryRecord(providers.value.summary).provider_stats as Record<string, AiRow> | undefined)
const activeApiQps = computed(() => Number(summaryRecord(apis.value.summary).qps_total ?? 0))
const activeAccountCount = computed(() => Number(summaryRecord(accounts.value.summary).active_count ?? 0))
const activeApiCount = computed(() => Number(summaryRecord(apis.value.summary).active_count ?? 0))
const accountQuotaTotal = computed(() => Number(summaryRecord(accounts.value.summary).quota_total ?? 0))
const selectedAccount = computed(() => currentAccounts.value.find((row) => rowId(row) === selectedAccountId.value))
const accountProviderFilter = computed(() => (!accountScopeAll.value && selectedProviderId.value ? selectedProviderId.value : undefined))
const apiProviderFilter = computed(() => (!accountScopeAll.value && selectedProviderId.value ? selectedProviderId.value : undefined))
const apiAccountFilter = computed(() => selectedAccountId.value || undefined)
const accountScopeTitle = computed(() => {
  if (accountScopeAll.value || !selectedProvider.value) return '所有接入账号'
  return `${text(selectedProvider.value.name)}的接入账号`
})
const accountScopeDescription = computed(() => {
  if (!selectedProvider.value) return '默认显示所有账号；选择左侧供应商后，右侧自动切换为该供应商账号。'
  return `当前只显示 ${text(selectedProvider.value.name)} 下的账号；点击全部账号可回到全量视图。`
})
const providerAccountScopeText = computed(() => {
  if (selectedProvider.value) return `${text(selectedProvider.value.name)}账号`
  return '所有接入账号'
})
const apiScopeTitle = computed(() => {
  if (selectedAccount.value) return `${text(selectedAccount.value.account_name)} 的 API`
  if (!accountScopeAll.value && selectedProvider.value) return `${text(selectedProvider.value.name)}的 API`
  return '所有 API'
})
const providerOptions = computed(() => providers.value.items.map((provider) => ({
  label: `${text(provider.name)} / ${text(provider.code)}`,
  value: rowId(provider),
})))
const accountOptions = computed(() => accounts.value.items.map((account) => ({
  label: `${text(account.account_name)} / ${text(account.key_alias)}`,
  value: rowId(account),
})))

function providerStats(providerId: string) {
  const stat = providerListStats.value?.[providerId]
  return {
    accountCount: Number(stat?.account_count ?? 0),
    apiCount: Number(stat?.api_count ?? 0),
    activeApiCount: Number(stat?.active_api_count ?? 0),
  }
}

function summaryRecord(value: unknown) {
  return value && typeof value === 'object' ? value as AiRow : {}
}

function loginMethodText(value: unknown) {
  const method = text(value, '')
  const labels: Record<string, string> = {
    email: '邮箱',
    phone: '手机',
    oauth: '第三方登录',
    cloud_console: '云控制台',
    service_account: '服务账号',
    team_account: '团队账号',
  }
  return method ? labels[method] || method : '未登记'
}

function normalizedStatus(value: unknown) {
  return typeof value === 'string' ? value.toLowerCase() : ''
}

function apiRunnable(row: AiRow) {
  const status = normalizedStatus(row.status)
  const health = normalizedStatus(row.health_status)
  return ['active', 'enabled', 'online'].includes(status) && ['active', 'success', 'online'].includes(health)
}

function apiAvailabilityText(row: AiRow) {
  const status = normalizedStatus(row.status)
  const health = normalizedStatus(row.health_status)
  if (!['active', 'enabled', 'online'].includes(status)) return '已停用'
  if (apiRunnable(row)) return '可执行'
  if (!health || health === 'unknown') return '未验证'
  return '不可执行'
}

function apiAvailabilityType(row: AiRow) {
  if (apiRunnable(row)) return 'success'
  const status = normalizedStatus(row.status)
  const health = normalizedStatus(row.health_status)
  if (!['active', 'enabled', 'online'].includes(status) || ['error', 'failed', 'timeout', 'offline', 'inactive', 'disabled'].includes(health)) return 'danger'
  return 'warning'
}

function apiHealthSummary(row: AiRow) {
  const raw = text(row.health_message, '')
  const status = normalizedStatus(row.health_status)
  if (!raw) {
    if (!status || status === 'unknown') return '尚未检测'
    if (['active', 'success', 'online'].includes(status)) return '探测通过'
    return '待重新检测'
  }

  const lower = raw.toLowerCase()
  if (lower.includes('http 200')) return '探测通过，未触发模型生成'
  if (lower.includes('http 400') && (lower.includes('invalidparameter') || lower.includes('鉴权已通过') || lower.includes('url error'))) {
    return '鉴权可用，接口参数需校验'
  }
  if (lower.includes('http 404')) return '接口路径不可用，请检查 API 地址'
  if (lower.includes('401') || lower.includes('unauthorized')) return '鉴权失败，请检查密钥'
  if (lower.includes('403') || lower.includes('forbidden')) return '权限不足，请检查账号授权'
  if (lower.includes('timeout') || lower.includes('超时')) return '探测超时，请检查网络或超时设置'
  if (lower.includes('no response') || lower.includes('无响应体')) return '服务无响应，请检查协议或路径'
  if (['active', 'success', 'online'].includes(status)) return '探测通过'
  if (['warning', 'degraded'].includes(status)) return '探测异常，请复查配置'
  if (['error', 'failed', 'timeout', 'offline'].includes(status)) return '探测失败，请检查配置'
  return raw.length > 42 ? `${raw.slice(0, 42)}...` : raw
}

async function loadData() {
  loading.value = true
  try {
    const [providerPage, accountResult, apiResult] = await Promise.all([
      fetchAiResource('providers', { skip: 0, limit: 200, keyword: keyword.value.trim() }),
      fetchAiResource('accounts', { skip: (accountPage.value - 1) * accountPageSize.value, limit: accountPageSize.value, provider_id: accountProviderFilter.value }),
      fetchAiResource('apis', { skip: (apiPage.value - 1) * apiPageSize.value, limit: apiPageSize.value, provider_id: apiProviderFilter.value, account_id: apiAccountFilter.value }),
    ])
    providers.value = providerPage
    accounts.value = accountResult
    apis.value = apiResult
    if (selectedProviderId.value && !providerPage.items.some((row) => rowId(row) === selectedProviderId.value)) {
      selectedProviderId.value = ''
      selectedAccountId.value = ''
      accountScopeAll.value = true
    }
    if (selectedAccountId.value && !currentAccounts.value.some((row) => rowId(row) === selectedAccountId.value)) {
      selectedAccountId.value = ''
    }
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '供应商加载失败')
  } finally {
    loading.value = false
  }
}

function selectProvider(providerId: string) {
  const selectingSameProvider = selectedProviderId.value === providerId
  selectedProviderId.value = selectingSameProvider ? '' : providerId
  accountScopeAll.value = selectingSameProvider
  selectedAccountId.value = ''
  resetAccountPaging()
  void loadData()
}

function showAllAccounts() {
  selectedProviderId.value = ''
  accountScopeAll.value = true
  selectedAccountId.value = ''
  resetAccountPaging()
  void loadData()
}

function showProviderAccounts() {
  if (!selectedProviderId.value) return
  accountScopeAll.value = false
  selectedAccountId.value = ''
  resetAccountPaging()
  void loadData()
}

function selectAccount(accountId: string) {
  selectedAccountId.value = selectedAccountId.value === accountId ? '' : accountId
  apiPage.value = 1
  void loadData()
}

function showAllApis() {
  selectedAccountId.value = ''
  apiPage.value = 1
  void loadData()
}

function accountRowClassName({ row }: { row: AiRow }) {
  return selectedAccountId.value === rowId(row) ? 'ai-selectable-row is-selected' : 'ai-selectable-row'
}

async function createProvider(payload: AiRow) {
  await createAiResource('providers', payload)
  createVisible.value = false
  ElMessage.success('供应商已创建')
  await loadData()
}

async function checkConnectivity() {
  const currentApiIds = currentApis.value.map((row) => rowId(row)).filter(Boolean)
  if (!currentApiIds.length) {
    ElMessage.warning('当前列表没有可检测的 API')
    return
  }
  connectivityChecking.value = true
  try {
    const result = await checkAiProviderAPIConnectivity({
      provider_id: !accountScopeAll.value && selectedProviderId.value ? selectedProviderId.value : undefined,
      account_id: selectedAccountId.value || undefined,
    }, { api_ids: currentApiIds })
    ElMessage.success(`连通性检测完成：本次检测 ${result.total} 个 API，正常 ${result.active} 个，告警 ${result.warning} 个，异常 ${result.error} 个`)
    await loadData()
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '连通性检测失败')
  } finally {
    connectivityChecking.value = false
  }
}

function resetAccountPaging() {
  accountPage.value = 1
  apiPage.value = 1
}

function handleAccountPageChange(page: number) {
  accountPage.value = page
  selectedAccountId.value = ''
  apiPage.value = 1
  void loadData()
}

function handleAccountSizeChange(size: number) {
  accountPageSize.value = size
  handleAccountPageChange(1)
}

function handleApiPageChange(page: number) {
  apiPage.value = page
  void loadData()
}

function handleApiSizeChange(size: number) {
  apiPageSize.value = size
  handleApiPageChange(1)
}

onMounted(loadData)
</script>

<template>
  <NeuroAgentPageShell class="ai-prototype" :show-hero="false">
    <template #title>供应商</template>
    <template #subtitle>统一维护公有云、私有化和自建模型供应商，并关联接入账号与 API。</template>
    <template #actions>
      <el-button type="primary" :icon="Plus" @click="createVisible = true">新增供应商</el-button>
    </template>

    <div class="ai-provider-workspace">
      <aside class="ai-provider-sidebar ai-card">
        <header class="ai-card__header">
          <div>
            <h3>供应商</h3>
            <p class="ai-card__description">左侧维护供应商，右侧联动账号和 API。</p>
          </div>
          <div class="ai-card__header-actions">
            <button class="ai-icon-action is-primary" type="button" title="新增供应商" @click="createVisible = true">
              <el-icon><Plus /></el-icon>
            </button>
          </div>
        </header>
        <div class="ai-card__body">
          <div class="ai-toolbar ai-provider-toolbar">
            <el-input v-model="keyword" clearable placeholder="搜索供应商、code、endpoint" @change="loadData" />
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
                <el-tag :type="statusType(provider.status)" size="small">{{ statusText(provider.status) }}</el-tag>
              </span>
              <span class="ai-provider-card__endpoint">{{ text(provider.base_url) }}</span>
              <span class="ai-provider-card__meta">
                <span>{{ providerStats(rowId(provider)).accountCount }} 账号</span>
                <span>{{ providerStats(rowId(provider)).apiCount }} API</span>
                <span>{{ providerStats(rowId(provider)).activeApiCount }} 启用</span>
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
              <p class="ai-card__description">{{ accountScopeDescription }}</p>
            </div>
            <div class="ai-segmented">
              <button :class="{ active: accountScopeAll }" @click="showAllAccounts">全部账号</button>
              <button v-if="selectedProviderId" :class="{ active: !accountScopeAll }" @click="showProviderAccounts">{{ providerAccountScopeText }}</button>
            </div>
          </header>
          <div class="ai-card__body">
            <div class="ai-metric-grid ai-provider-summary-grid">
              <div class="ai-mini-stat"><span>账号总数</span><strong>{{ accounts.total }}</strong></div>
              <div class="ai-mini-stat"><span>启用账号</span><strong>{{ activeAccountCount }}</strong></div>
              <div class="ai-mini-stat"><span>API 数量</span><strong>{{ apis.total }}</strong></div>
              <div class="ai-mini-stat"><span>额度合计</span><strong>{{ moneyText(accountQuotaTotal) }}</strong></div>
            </div>
            <el-table
              :data="currentAccounts"
              border
              v-loading="loading"
              row-key="id"
              :row-class-name="accountRowClassName"
              @row-click="(row: AiRow) => selectAccount(rowId(row))"
            >
              <el-table-column label="账号" min-width="180">
                <template #default="{ row }">
                  <span class="ai-table-cell-main">
                    <strong>{{ text(row.account_name) }}</strong>
                    <small>{{ text(row.key_alias) }}</small>
                  </span>
                </template>
              </el-table-column>
              <el-table-column label="注册账号" min-width="220">
                <template #default="{ row }">
                  <span class="ai-table-cell-main">
                    <strong>{{ text(row.login_account, '未登记') }}</strong>
                    <small>{{ loginMethodText(row.login_method) }} · {{ text(row.maintainer, '未指定维护人') }}</small>
                  </span>
                </template>
              </el-table-column>
              <el-table-column label="维护联系" min-width="160"><template #default="{ row }">{{ text(row.maintainer_contact, '未登记') }}</template></el-table-column>
              <el-table-column label="Endpoint" min-width="220"><template #default="{ row }">{{ text(row.endpoint) }}</template></el-table-column>
              <el-table-column label="额度使用" width="180"><template #default="{ row }">{{ numberText(row.used_quota) }} / {{ numberText(row.quota_limit) }}</template></el-table-column>
              <el-table-column label="状态" width="110"><template #default="{ row }"><el-tag :type="statusType(row.status)">{{ statusText(row.status) }}</el-tag></template></el-table-column>
              <el-table-column label="操作" width="190">
                <template #default="{ row }">
                  <div class="ai-row-actions">
                    <AiResourceActions resource="accounts" :row="row" :options="{ provider_id: providerOptions }" @saved="loadData" />
                  </div>
                </template>
              </el-table-column>
            </el-table>
            <el-pagination
              v-if="accounts.total > accountPageSize"
              v-model:current-page="accountPage"
              v-model:page-size="accountPageSize"
              class="ai-pagination"
              background
              layout="total, sizes, prev, pager, next"
              :page-sizes="[10, 20, 50, 100]"
              :total="accounts.total"
              @current-change="handleAccountPageChange"
              @size-change="handleAccountSizeChange"
            />
          </div>
        </section>

        <section class="ai-card ai-table">
          <header class="ai-card__header">
            <div>
              <h3>{{ apiScopeTitle }}</h3>
              <p class="ai-card__description">默认显示所有 API；选择账号后，只显示该账号挂载的 API。</p>
            </div>
            <div class="ai-segmented">
              <button :disabled="connectivityChecking" @click="checkConnectivity">{{ connectivityChecking ? '检测中' : '检测当前列表' }}</button>
              <button :class="{ active: !selectedAccountId }" @click="showAllApis">全部 API</button>
            </div>
          </header>
          <div class="ai-card__body">
            <div class="ai-metric-grid ai-provider-summary-grid">
              <div class="ai-mini-stat"><span>API 总数</span><strong>{{ apis.total }}</strong></div>
              <div class="ai-mini-stat"><span>启用 API</span><strong>{{ activeApiCount }}</strong></div>
              <div class="ai-mini-stat"><span>聚合 QPS</span><strong>{{ activeApiQps }}</strong></div>
              <div class="ai-mini-stat"><span>当前账号</span><strong>{{ selectedAccount ? text(selectedAccount.account_name) : '全部' }}</strong></div>
            </div>
            <el-table :data="currentApis" border v-loading="loading">
              <el-table-column label="API" min-width="260" :show-overflow-tooltip="false"><template #default="{ row }"><span class="ai-table-cell-main"><strong>{{ text(row.api_name) }}</strong><small>{{ text(row.api_path) }}</small></span></template></el-table-column>
              <el-table-column label="类型" width="120" :show-overflow-tooltip="false"><template #default="{ row }"><el-tag>{{ text(row.api_type) }}</el-tag></template></el-table-column>
              <el-table-column label="能力" min-width="210" :show-overflow-tooltip="false">
                <template #default="{ row }">
                  <span class="ai-capability-tags">
                    <el-tag v-for="capability in capabilityTagList(row.capabilities)" :key="capability.code" effect="plain">{{ capability.label }}</el-tag>
                  </span>
                </template>
              </el-table-column>
              <el-table-column label="QPS / 超时" width="150"><template #default="{ row }">{{ numberText(row.qps_limit) }} / {{ numberText(row.timeout_ms) }}ms</template></el-table-column>
              <el-table-column label="连通性" width="190" :show-overflow-tooltip="false">
                <template #default="{ row }">
                  <span class="ai-table-cell-main ai-health-cell">
                    <el-tag :type="statusType(row.health_status)">{{ statusText(row.health_status, '未检测') }}</el-tag>
                    <small :title="text(row.health_message, '')">{{ apiHealthSummary(row) }}</small>
                  </span>
                </template>
              </el-table-column>
              <el-table-column label="可执行性" width="118">
                <template #default="{ row }">
                  <el-tag :type="apiAvailabilityType(row)">{{ apiAvailabilityText(row) }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column label="状态" width="110"><template #default="{ row }"><el-tag :type="statusType(row.status)">{{ statusText(row.status) }}</el-tag></template></el-table-column>
              <el-table-column label="操作" width="128"><template #default="{ row }"><AiResourceActions resource="apis" :row="row" :options="{ provider_id: providerOptions, account_id: accountOptions, provider_account_id: accountOptions }" @saved="loadData" /></template></el-table-column>
            </el-table>
            <el-pagination
              v-if="apis.total > apiPageSize"
              v-model:current-page="apiPage"
              v-model:page-size="apiPageSize"
              class="ai-pagination"
              background
              layout="total, sizes, prev, pager, next"
              :page-sizes="[10, 20, 50, 100]"
              :total="apis.total"
              @current-change="handleApiPageChange"
              @size-change="handleApiSizeChange"
            />
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
  </NeuroAgentPageShell>
</template>
