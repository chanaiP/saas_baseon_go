<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Close, CopyDocument, Refresh, Search } from '@element-plus/icons-vue'

import { formatDateTimeChina } from '@/utils/datetime'
import NeuroAgentPageShell from '@/views/components/NeuroAgentPageShell.vue'
import { fetchTenants } from '@/api/tenant'

import { fetchAiResource } from '../api'
import { emptyPage, moneyText, numberText, rowId, statusText, statusType, text, type AiRow } from './viewHelpers'
import './aiPrototype.css'

defineOptions({ name: 'AiUsageLogsView' })

const loading = ref(false)
const keyword = ref('')
const startDate = ref(todayChinaDate())
const endDate = ref(todayChinaDate())
const scenarioKey = ref('')
const page = ref(1)
const pageSize = ref(20)
const records = ref(emptyPage())
const scenarios = ref(emptyPage())
const tenants = ref(emptyPage())
const selected = ref<AiRow | null>(null)

interface UsageTotals {
  calls: number
  cost: number
  billing: number
  usage: number
}

interface ContentPolicy {
  level: number | null
  title: string
  description: string
  tone: 'success' | 'warning' | 'danger' | 'info'
}

const summaryRows = computed(() => {
  const summary = records.value.summary as AiRow | AiRow[] | undefined
  if (Array.isArray(summary)) return summary
  return Array.isArray(summary?.usage_units) ? summary.usage_units as AiRow[] : []
})
const totals = computed<UsageTotals>(() => summaryRows.value.reduce<UsageTotals>((acc, row) => {
  acc.calls += Number(row.calls ?? 0)
  acc.cost += Number(row.cost_amount ?? 0)
  acc.billing += Number(row.billing_amount ?? 0)
  acc.usage += Number(row.usage_amount ?? 0)
  return acc
}, { calls: 0, cost: 0, billing: 0, usage: 0 }))
const successCount = computed(() => records.value.items.filter((row) => row.status === 'success').length)
const successRate = computed(() => {
  if (!records.value.items.length) return '0.00%'
  return `${((successCount.value / records.value.items.length) * 100).toFixed(2)}%`
})

const selectedContentPolicy = computed(() => selected.value ? contentRecordPolicy(selected.value) : null)
const selectedContentMode = computed(() => selectedContentPolicy.value?.title || '-')
const scenarioOptions = computed(() => scenarios.value.items.map((row) => {
  const appCode = text(row.app_code, '')
  const scenarioCode = text(row.ai_scenario_code, '')
  const appName = text(row.app_name || row.app_code, '')
  const scenarioName = text(row.ai_scenario_name || row.ai_scenario_code)
  return {
    key: `${appCode}::${scenarioCode}`,
    appCode,
    scenarioCode,
    appName,
    scenarioName,
    label: `${appName} - ${scenarioName}`,
  }
}).filter((row) => row.appCode && row.scenarioCode))
const selectedScenarioFilter = computed(() => {
  const [appCode = '', scenarioCode = ''] = scenarioKey.value.split('::')
  return { appCode, scenarioCode }
})

async function loadScenarios() {
  try {
    scenarios.value = await fetchAiResource('scenarios', { skip: 0, limit: 200 })
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : 'AI 场景加载失败')
  }
}

async function loadTenants() {
  try {
    const page = await fetchTenants(0, 500)
    tenants.value = { ...page, items: page.items as unknown as AiRow[] }
  } catch {
    tenants.value = emptyPage()
  }
}

async function loadData() {
  loading.value = true
  try {
    records.value = await fetchAiResource('usage-records', {
      skip: (page.value - 1) * pageSize.value,
      limit: pageSize.value,
      keyword: keyword.value.trim(),
      start_date: startDate.value,
      end_date: endDate.value,
      app_code: selectedScenarioFilter.value.appCode,
      ai_scenario_code: selectedScenarioFilter.value.scenarioCode,
    })
    if (!records.value.items.length && page.value > 1) {
      page.value -= 1
      await loadData()
    }
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '调用日志加载失败')
  } finally {
    loading.value = false
  }
}

function search() {
  page.value = 1
  loadData()
}

function reset() {
  keyword.value = ''
  scenarioKey.value = ''
  startDate.value = todayChinaDate()
  endDate.value = todayChinaDate()
  search()
}

function showLast7Days() {
  startDate.value = chinaDateOffset(-6)
  endDate.value = todayChinaDate()
  search()
}

function requestParamsPreview(row: AiRow) {
  const raw = text(row.request_params, '')
  if (!raw) return '-'
  try {
    return JSON.stringify(translateContentRecord(JSON.parse(raw)), null, 2)
  } catch {
    return raw
  }
}

function shortId(value: unknown) {
  const raw = text(value, '')
  return raw.length > 16 ? `${raw.slice(0, 8)}...${raw.slice(-6)}` : raw || '-'
}

function techId(value: unknown) {
  const raw = text(value, '')
  if (!raw) return '-'
  return raw.length > 18 ? `${raw.slice(0, 10)} · ${raw.slice(-8)}` : raw
}

function dateTime(value: unknown) {
  return formatDateTimeChina(value as string | Date | number | null | undefined)
}

function httpStatusText(value: unknown) {
  const status = Number(value ?? 0)
  return status > 0 ? String(status) : '未请求'
}

async function copyText(value: unknown) {
  const raw = text(value, '')
  if (!raw) return
  try {
    await navigator.clipboard.writeText(raw)
    ElMessage.success('已复制')
  } catch {
    ElMessage.error('复制失败')
  }
}

function todayChinaDate() {
  return chinaDateOffset(0)
}

function chinaDateOffset(offsetDays: number) {
  const date = new Date()
  date.setDate(date.getDate() + offsetDays)
  const parts = new Intl.DateTimeFormat('en-CA', {
    timeZone: 'Asia/Shanghai',
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  }).formatToParts(date)
  const get = (type: Intl.DateTimeFormatPart['type']) => parts.find((part) => part.type === type)?.value ?? ''
  return `${get('year')}-${get('month')}-${get('day')}`
}

function usageUnitText(value: unknown, fallback = '-') {
  const raw = text(value, '').toLowerCase()
  const labels: Record<string, string> = {
    tokens: 'tokens',
    token: 'tokens',
    images: '张图片',
    image: '张图片',
    calls: '次调用',
    requests: '次请求',
    characters: '字符',
    seconds: '秒',
    minutes: '分钟',
  }
  return raw ? labels[raw] || raw : fallback
}

function usageAmountText(amount: unknown, unit: unknown) {
  return `${numberText(amount)} ${usageUnitText(unit, '')}`.trim()
}

function tenantDisplayName(row: AiRow) {
  const named = text(row.tenant_name, '')
  if (named) return named
  const tenantId = String(row.tenant_id || '')
  const tenant = tenants.value.items.find((item) => String(item.id || '') === tenantId)
  return text(tenant?.name || tenant?.code || (tenantId ? `租户 #${tenantId}` : '未知租户'))
}

function appDisplayName(row: AiRow) {
  return text(row.app_name || row.app_code, '未知应用')
}

function scenarioDisplayName(row: AiRow) {
  return text(row.ai_scenario_name || row.ai_scenario_code, '未知场景')
}

function contentRecordMode(row: AiRow) {
  return contentRecordPolicy(row).title
}

function contentRecordPolicy(row: AiRow): ContentPolicy {
  const raw = text(row.request_params, '')
  if (!raw || raw === '{}') {
    return {
      level: 0,
      title: '不记录内容',
      description: '本次调用未保存请求参数、输入内容或 AI 输出内容，只保留业务、计费、状态和链路追踪字段。',
      tone: 'info',
    }
  }
  try {
    const parsed = JSON.parse(raw) as AiRow
    const level = Number(parsed.content_record_level)
    const policies: Record<number, ContentPolicy> = {
      0: {
        level: 0,
        title: '不记录内容',
        description: '系统参数设置为 0，本次调用不保存参数、输入或 AI 输出内容。',
        tone: 'info',
      },
      1: {
        level: 1,
        title: '仅记录概要',
        description: '系统参数设置为 1，仅保存字段列表、字段数、字节数和哈希，便于排查但不展示原文。',
        tone: 'success',
      },
      2: {
        level: 2,
        title: '记录脱敏内容',
        description: '系统参数设置为 2，保存脱敏后的参数和输入内容，手机号、邮箱、证件、银行卡、密钥、授权头、Cookie 等会被替换。',
        tone: 'warning',
      },
      3: {
        level: 3,
        title: '记录完整内容',
        description: '系统参数设置为 3，本次调用可能保存完整参数、输入或输出内容，应仅用于受控排障并配合权限、保留周期和审计复核。',
        tone: 'danger',
      },
    }
    if (Number.isFinite(level)) {
      return policies[level] || {
        level,
        title: `记录级别 ${level}`,
        description: '该记录使用了未知内容记录级别，请核对系统参数 ai.gateway.content_record_level。',
        tone: 'warning',
      }
    }
    return {
      level: null,
      title: '历史记录',
      description: '该记录没有显式内容记录级别，可能来自旧版本或历史导入数据。',
      tone: 'info',
    }
  } catch {
    return {
      level: null,
      title: '原始文本',
      description: '请求参数不是标准 JSON，以下内容按原始文本展示。',
      tone: 'warning',
    }
  }
}

function translateContentRecord(value: unknown): unknown {
  if (Array.isArray(value)) return value.map((item) => translateContentRecord(item))
  if (value && typeof value === 'object') {
    return Object.fromEntries(Object.entries(value as AiRow).map(([key, item]) => [
      contentKeyText(key),
      contentValueText(key, translateContentRecord(item)),
    ]))
  }
  if (typeof value === 'boolean') return value ? '是' : '否'
  return value
}

function contentKeyText(key: string) {
  const labels: Record<string, string> = {
    channel: '来源渠道',
    dry_run: '试运行',
    content_record_level: '内容记录级别',
    params: '请求参数',
    input: '输入内容',
    params_summary: '参数概要',
    input_summary: '输入概要',
    keys: '字段列表',
    field_count: '字段数量',
    bytes: '内容字节数',
    hash: '内容哈希',
    params_hash: '参数哈希',
    input_hash: '输入哈希',
    content_record_policy: '内容记录策略',
    mode: '记录模式',
    scope: '记录范围',
    audit_required: '需要审计',
    retention_days: '建议保留天数',
    warning: '风险提示',
  }
  return labels[key] || key
}

function contentValueText(key: string, value: unknown) {
  if (key === 'channel' && value === 'demo-seed') return '演示种子数据'
  if (key === 'channel' && value === 'demo-history-seed') return '演示历史数据'
  if (key === 'content_record_level') {
    const level = Number(value)
    const labels: Record<number, string> = {
      0: '0 - 不记录',
      1: '1 - 记录概要',
      2: '2 - 记录脱敏内容',
      3: '3 - 记录完整内容',
    }
    return labels[level] || value
  }
  if (key === 'mode') {
    const labels: Record<string, string> = {
      summary: '仅概要',
      redacted: '脱敏内容',
      full: '完整内容',
    }
    return labels[String(value)] || value
  }
  if (key === 'scope') {
    const labels: Record<string, string> = {
      metadata_only: '仅元数据',
      params_and_input: '请求参数与输入内容',
    }
    return labels[String(value)] || value
  }
  return value
}

onMounted(() => {
  loadTenants()
  loadScenarios()
  loadData()
})
</script>

<template>
  <NeuroAgentPageShell class="ai-prototype" :show-hero="false">
    <template #title>调用日志</template>
    <template #subtitle>按请求追踪 AI Gateway 的租户、场景、路由、成本、状态和内容记录级别。</template>

    <section class="ai-card ai-table">
      <header class="ai-card__header">
        <div>
          <h3>调用明细</h3>
          <p class="ai-card__description">默认展示今日调用并按调用时间倒序排列；内容字段受系统参数 ai.gateway.content_record_level 控制。</p>
        </div>
      </header>
      <div class="ai-card__body">
        <div class="ai-toolbar ai-usage-filter-toolbar">
          <el-input v-model="keyword" class="ai-filter-keyword" clearable placeholder="搜索租户、应用、用户、状态" @keyup.enter="search">
            <template #prefix><el-icon><Search /></el-icon></template>
          </el-input>
          <el-select v-model="scenarioKey" class="ai-filter-scenario" clearable filterable placeholder="应用 - AI 场景" :teleported="false" @change="search">
            <el-option v-for="row in scenarioOptions" :key="row.key" :label="row.label" :value="row.key">
              <span class="ai-select-option"><strong>{{ row.appName }} - {{ row.scenarioName }}</strong><small>{{ row.appCode }} · {{ row.scenarioCode }}</small></span>
            </el-option>
          </el-select>
          <el-date-picker v-model="startDate" class="ai-filter-date" type="date" value-format="YYYY-MM-DD" placeholder="开始日期" :teleported="false" />
          <el-date-picker v-model="endDate" class="ai-filter-date" type="date" value-format="YYYY-MM-DD" placeholder="结束日期" :teleported="false" />
          <div class="ai-filter-actions">
            <el-button :icon="Search" type="primary" @click="search">查询</el-button>
            <el-button @click="showLast7Days">近 7 天</el-button>
            <el-button :icon="Refresh" @click="reset">重置</el-button>
          </div>
        </div>

        <div class="ai-metric-grid ai-usage-log-stats">
          <div class="ai-mini-stat"><span>调用次数</span><strong>{{ numberText(totals.calls) }}</strong><small class="ai-muted">SUM(calls)</small></div>
          <div class="ai-mini-stat"><span>记录数</span><strong>{{ numberText(records.total) }}</strong><small class="ai-muted">当前筛选</small></div>
          <div class="ai-mini-stat"><span>成功率</span><strong>{{ successRate }}</strong><small class="ai-muted">当前页记录</small></div>
          <div class="ai-mini-stat"><span>销售额 / 成本</span><strong>{{ moneyText(totals.billing) }}</strong><small class="ai-muted">成本 {{ moneyText(totals.cost) }}</small></div>
        </div>

        <el-table v-loading="loading" :data="records.items" border empty-text="暂无今日真实调用，可切换近 7 天或检查 AI Gateway 接入配置。">
          <el-table-column label="调用时间" width="180">
            <template #default="{ row }">{{ dateTime(row.called_at) }}</template>
          </el-table-column>
          <el-table-column label="请求 ID" min-width="150">
            <template #default="{ row }"><code>{{ shortId(row.request_id) }}</code></template>
          </el-table-column>
          <el-table-column label="租户 / 应用" min-width="190">
            <template #default="{ row }">
              <span class="ai-table-cell-main"><strong>{{ tenantDisplayName(row) }}</strong><small>{{ appDisplayName(row) }}</small></span>
            </template>
          </el-table-column>
          <el-table-column label="AI 场景" min-width="190">
            <template #default="{ row }">
              <span class="ai-table-cell-main"><strong>{{ scenarioDisplayName(row) }}</strong><small>场景编码：{{ text(row.ai_scenario_code) }}</small></span>
            </template>
          </el-table-column>
          <el-table-column label="调用 / 用量" width="130">
            <template #default="{ row }">
              <span class="ai-table-cell-main">
                <strong>调用 {{ numberText(row.calls) }} 次</strong>
                <small>用量 {{ usageAmountText(row.usage_amount, row.usage_unit) }}</small>
              </span>
            </template>
          </el-table-column>
          <el-table-column label="销售额 / 成本" width="140">
            <template #default="{ row }"><span class="ai-table-cell-main"><strong>{{ moneyText(row.billing_amount) }}</strong><small>成本 {{ moneyText(row.cost_amount) }}</small></span></template>
          </el-table-column>
          <el-table-column label="延迟" width="90">
            <template #default="{ row }">{{ numberText(row.latency_ms) }}ms</template>
          </el-table-column>
          <el-table-column label="状态" width="100">
            <template #default="{ row }"><el-tag :type="statusType(row.status)">{{ statusText(row.status) }}</el-tag></template>
          </el-table-column>
          <el-table-column label="操作" width="90" fixed="right">
            <template #default="{ row }"><el-button link type="primary" @click="selected = row">详情</el-button></template>
          </el-table-column>
        </el-table>

        <div class="ai-pagination">
          <el-pagination
            v-model:current-page="page"
            v-model:page-size="pageSize"
            layout="total, sizes, prev, pager, next"
            :page-sizes="[10, 20, 50, 100]"
            :total="records.total"
            @current-change="loadData"
            @size-change="search"
          />
        </div>
      </div>
    </section>

    <el-drawer
      :model-value="Boolean(selected)"
      :with-header="false"
      class="ai-usage-drawer"
      size="min(720px, calc(100vw - 24px))"
      @close="selected = null"
    >
      <div v-if="selected" class="ai-detail-stack">
        <header class="ai-detail-hero">
          <div>
            <span class="ai-detail-kicker">调用详情</span>
            <h3>{{ text(selected.ai_scenario_name || selected.ai_scenario_code) }}</h3>
            <p>{{ tenantDisplayName(selected) }} · {{ appDisplayName(selected) }} · {{ dateTime(selected.called_at) }}</p>
          </div>
          <button class="ai-icon-button" type="button" aria-label="关闭调用详情" @click="selected = null">
            <el-icon><Close /></el-icon>
          </button>
        </header>

        <div class="ai-detail-summary">
          <span>
            <small>状态</small>
            <el-tag :type="statusType(selected.status)">{{ statusText(selected.status) }}</el-tag>
          </span>
          <span>
            <small>调用次数</small>
            <strong>{{ numberText(selected.calls) }}</strong>
          </span>
          <span>
            <small>用量</small>
            <strong>{{ usageAmountText(selected.usage_amount, selected.usage_unit) }}</strong>
          </span>
          <span>
            <small>内容策略</small>
            <strong>{{ selectedContentMode }}</strong>
          </span>
        </div>

        <section class="ai-detail-section">
          <h4>业务上下文</h4>
          <dl>
            <div><dt>租户</dt><dd>{{ tenantDisplayName(selected) }}</dd></div>
            <div><dt>应用</dt><dd>{{ appDisplayName(selected) }}</dd></div>
            <div><dt>AI 场景</dt><dd>{{ scenarioDisplayName(selected) }}</dd></div>
            <div><dt>调用用户</dt><dd>{{ text(selected.user_name || selected.user_id) }}</dd></div>
            <div><dt>请求 ID</dt><dd><code :title="text(selected.request_id)">{{ text(selected.request_id) }}</code></dd></div>
          </dl>
        </section>

        <section class="ai-detail-section">
          <h4>技术链路</h4>
          <div class="ai-id-grid">
            <span><small>供应商</small><code :title="text(selected.provider_id)">{{ techId(selected.provider_id) }}</code><button type="button" aria-label="复制供应商 ID" @click="copyText(selected.provider_id)"><el-icon><CopyDocument /></el-icon></button></span>
            <span><small>供应商账号</small><code :title="text(selected.provider_account_id)">{{ techId(selected.provider_account_id) }}</code><button type="button" aria-label="复制供应商账号 ID" @click="copyText(selected.provider_account_id)"><el-icon><CopyDocument /></el-icon></button></span>
            <span><small>API</small><code :title="text(selected.provider_api_id)">{{ techId(selected.provider_api_id) }}</code><button type="button" aria-label="复制 API ID" @click="copyText(selected.provider_api_id)"><el-icon><CopyDocument /></el-icon></button></span>
            <span><small>模型</small><code :title="text(selected.model_id)">{{ techId(selected.model_id) }}</code><button type="button" aria-label="复制模型 ID" @click="copyText(selected.model_id)"><el-icon><CopyDocument /></el-icon></button></span>
            <span><small>基础路由</small><code :title="text(selected.base_route_id)">{{ techId(selected.base_route_id) }}</code><button type="button" aria-label="复制基础路由 ID" @click="copyText(selected.base_route_id)"><el-icon><CopyDocument /></el-icon></button></span>
            <span><small>租户策略</small><code :title="text(selected.tenant_strategy_id)">{{ techId(selected.tenant_strategy_id) }}</code><button type="button" aria-label="复制租户策略 ID" @click="copyText(selected.tenant_strategy_id)"><el-icon><CopyDocument /></el-icon></button></span>
          </div>
        </section>

        <section class="ai-detail-section">
          <h4>计费与结果</h4>
          <dl>
            <div><dt>成本</dt><dd>{{ moneyText(selected.cost_amount) }}</dd></div>
            <div><dt>销售额</dt><dd>{{ moneyText(selected.billing_amount) }}</dd></div>
            <div><dt>平台用量</dt><dd>{{ usageAmountText(selected.platform_amount, selected.platform_unit || selected.usage_unit) }}</dd></div>
            <div><dt>延迟</dt><dd>{{ numberText(selected.latency_ms) }}ms</dd></div>
            <div><dt>HTTP 状态</dt><dd>{{ httpStatusText(selected.provider_http_status) }}</dd></div>
            <div><dt>重试次数</dt><dd>{{ numberText(selected.retry_count) }}</dd></div>
            <div><dt>错误</dt><dd>{{ text(selected.error_message || selected.error_code, '无') }}</dd></div>
          </dl>
        </section>

        <section class="ai-detail-section">
          <h4>供应商追踪</h4>
          <dl>
            <div><dt>供应商请求 ID</dt><dd><code :title="text(selected.provider_request_id)">{{ text(selected.provider_request_id, '无') }}</code></dd></div>
            <div><dt>开始时间</dt><dd>{{ dateTime(selected.started_at) }}</dd></div>
            <div><dt>结束时间</dt><dd>{{ dateTime(selected.finished_at) }}</dd></div>
          </dl>
        </section>

        <section class="ai-detail-section">
          <h4>内容记录</h4>
          <div v-if="selectedContentPolicy" class="ai-content-policy" :class="`is-${selectedContentPolicy.tone}`">
            <strong>{{ selectedContentPolicy.title }}</strong>
            <span>{{ selectedContentPolicy.description }}</span>
          </div>
          <pre class="ai-code-block">{{ requestParamsPreview(selected) }}</pre>
        </section>
      </div>
    </el-drawer>
  </NeuroAgentPageShell>
</template>
