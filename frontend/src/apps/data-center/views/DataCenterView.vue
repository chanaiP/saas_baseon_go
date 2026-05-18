<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'

import { usePermissionStore } from '@/stores/permission'

import {
  analyzeAnomaly,
  closeTask,
  completeTask,
  confirmReview,
  createMetric,
  createRule,
  createTask,
  feedbackTask,
  fetchAnomalies,
  fetchDashboardRankings,
  fetchDashboardSummary,
  fetchDashboardTrends,
  fetchRawBatch,
  fetchRawErrors,
  fetchReview,
  fetchMetrics,
  fetchPipeline,
  fetchRawBatches,
  fetchReviews,
  fetchRule,
  fetchRules,
  fetchStandardData,
  fetchTask,
  fetchTasks,
  generateTaskFromAnomaly,
  reanalyzeAnomaly,
  reprocessRawBatch,
  scanAnomalies,
  setMetricEnabled,
  setRuleEnabled,
  startTask,
  testRule,
  updateAnomalyStatus,
  updateMetric,
  updateReview,
  updateRule,
  updateTask,
} from '../api'
import type {
  AnomalyRecord,
  AnomalyRule,
  DashboardSummary,
  MetricDefinition,
  RawBatch,
  RectificationReview,
  RectificationTask,
} from '../types'
import { useDataCenterFilters } from '../composables/useDataCenterFilters'
import './dataCenter.css'

defineOptions({ name: 'DataCenterView' })

const route = useRoute()
const router = useRouter()
const permissionStore = usePermissionStore()
const section = computed(() => String(route.params.section || 'overview'))
const loading = ref(false)
const error = ref('')
const { filters } = useDataCenterFilters()
const summary = ref<DashboardSummary | null>(null)
const trends = ref<Array<Record<string, unknown>>>([])
const rankings = ref<Record<string, Array<Record<string, unknown>>>>({})
const pipeline = ref<Array<Record<string, unknown>>>([])
const rawBatches = ref<RawBatch[]>([])
const standardRows = ref<Array<Record<string, unknown>>>([])
const standardType = ref('sales')
const metrics = ref<MetricDefinition[]>([])
const rules = ref<AnomalyRule[]>([])
const anomalies = ref<AnomalyRecord[]>([])
const tasks = ref<RectificationTask[]>([])
const reviews = ref<RectificationReview[]>([])
const selectedAnomaly = ref<AnomalyRecord | null>(null)
const detailDrawer = ref(false)
const detailTitle = ref('')
const detailData = ref<unknown>(null)
const rawErrors = ref<Array<Record<string, unknown>>>([])
const metricDialog = ref(false)
const ruleDialog = ref(false)
const taskDialog = ref(false)
const reviewDialog = ref(false)
const editingMetricId = ref<number | null>(null)
const editingRuleId = ref<number | null>(null)
const editingTaskId = ref<number | null>(null)
const editingReviewId = ref<number | null>(null)
const ruleTestResult = ref<Record<string, unknown> | null>(null)
const metricForm = ref<Record<string, unknown>>({})
const ruleForm = ref<Record<string, unknown>>({})
const taskForm = ref<Record<string, unknown>>({})
const reviewForm = ref<Record<string, unknown>>({})

const sections = [
  { key: 'dashboard', label: '经营看板' },
  { key: 'overview', label: '数据总览' },
  { key: 'raw', label: '原始数据' },
  { key: 'standard', label: '标准数据' },
  { key: 'metrics', label: '指标中心' },
  { key: 'anomalies', label: '异常分析' },
  { key: 'rules', label: '异常规则' },
  { key: 'tasks', label: '整改任务' },
  { key: 'reviews', label: '整改复盘' },
]

const currentTitle = computed(() => {
  if (!route.params.section) return '总览'
  return sections.find((item) => item.key === section.value)?.label ?? '经营看板'
})
const emptyText = computed(() => (error.value ? error.value : '暂无真实数据库记录'))
const trendMaxValue = computed(() => Math.max(1, ...trends.value.map((row) => Number(row.gmv ?? 0))))
const taskStatuses = ['待处理', '处理中', '已完成', '已逾期']
const rawTabs = ['全部数据', '订单原始数据', '广告原始数据', '库存原始数据', '门店销售原始数据', '退款原始数据']
const ruleTabs = ['全部规则', '销售异常', '投流异常', '退款异常', '库存异常', '门店异常']
const dashboardKpis = computed(() =>
  listOrEmpty(summary.value?.kpis).slice(0, 6).map((item) => ({
    label: item.label,
    value: valueText(item.value, item.unit),
    change: item.trend || '实时',
    trend: item.trend?.includes('-') ? 'down' : 'up',
  })),
)
const gmvTrendRows = computed(() => trends.value.map((row) => ({ date: shortDate(row.date), gmv: Number(row.gmv ?? 0) / 10000 })))
const roiTrendRows = computed(() => {
  const roi = Number(summary.value?.kpis?.find((item) => item.code === 'roi')?.value ?? 0)
  return trends.value.map((row) => ({ date: shortDate(row.date), roi: Number(row.roi ?? roi) }))
})
const channelRankRows = computed(() => {
  const source = listOrEmpty(rankings.value.channel_rank?.length ? rankings.value.channel_rank : rankings.value.brand_rank)
  const total = source.reduce((sum, item) => sum + Number(item.value ?? 0), 0) || 1
  return source.slice(0, 5).map((item) => ({
    name: String(item.name ?? item.code ?? '-'),
    value: Number(item.value ?? 0) / 10000,
    ratio: Math.max(5, Math.round((Number(item.value ?? 0) / total) * 100)),
  }))
})
const metricSummaryCards = computed(() => {
  const sales = metrics.value.filter((item) => item.metric_category?.includes('sales') || item.metric_category?.includes('销售')).length
  const ad = metrics.value.filter((item) => item.metric_category?.includes('ad') || item.metric_category?.includes('投流')).length
  const inventory = metrics.value.filter((item) => item.metric_category?.includes('inventory') || item.metric_category?.includes('库存')).length
  const anomaly = metrics.value.filter((item) => item.anomaly_enabled).length
  return [
    { label: '销售指标', value: sales, desc: 'GMV、净销售额、订单数等' },
    { label: '投流指标', value: ad, desc: '消耗、ROI、点击率、转化率等' },
    { label: '商品库存', value: inventory, desc: '动销率、可售天数、库存周转等' },
    { label: '参与异常判断', value: anomaly, desc: '已启用规则扫描' },
  ]
})
const anomalySummaryCards = computed(() => [
  { label: '高等级异常', value: anomalies.value.filter((item) => ['critical', 'high', '严重', '高'].includes(item.anomaly_level)).length, desc: '需优先整改', danger: true },
  { label: '待 AI 分析', value: anomalies.value.filter((item) => item.ai_status !== 'completed').length, desc: '等待生成原因和建议' },
  { label: '已生成任务', value: anomalies.value.filter((item) => item.task_status === 'generated').length, desc: '进入整改闭环' },
  { label: '待复盘', value: anomalies.value.filter((item) => item.review_status !== 'confirmed').length, desc: '等待验证结果' },
])
const reviewSummaryCards = computed(() => [
  { label: '已复盘', value: reviews.value.filter((item) => item.review_conclusion).length, desc: '近 30 天' },
  { label: '整改有效', value: reviews.value.filter((item) => item.review_conclusion === 'effective').length, desc: '持续沉淀经验' },
  { label: '无效整改', value: reviews.value.filter((item) => ['weak', 'ineffective'].includes(item.review_conclusion)).length, desc: '需继续跟进', danger: true },
  { label: '待复盘', value: tasks.value.filter((item) => item.review_status !== 'confirmed').length, desc: '到期自动提醒' },
])
const selectedReview = computed(() => reviews.value[0] ?? null)

function listOrEmpty<T>(value: T[] | null | undefined): T[] {
  return Array.isArray(value) ? value : []
}

function pageItems<T>(value: { items?: T[] | null } | null | undefined): T[] {
  return listOrEmpty(value?.items)
}

function canAction(code: string) {
  return permissionStore.canUseAction(code)
}

function trendBarHeight(value: unknown) {
  const n = Number(value ?? 0)
  if (!Number.isFinite(n) || n <= 0) return 8
  return Math.max(8, Math.round((n / trendMaxValue.value) * 168))
}

function shortDate(value: unknown) {
  const text = String(value ?? '')
  return text.length > 10 ? text.slice(5, 10) : text
}

function dateText(value: unknown) {
  return String(value ?? '-').slice(0, 10)
}

function percentBar(value: unknown, max: number) {
  const n = Number(value ?? 0)
  if (!Number.isFinite(n) || max <= 0) return 18
  return Math.max(18, Math.round((n / max) * 190))
}

function statusText(status?: string) {
  const map: Record<string, string> = {
    success: '成功',
    warning: '需关注',
    failed: '失败',
    pending: '待处理',
    processing: '处理中',
    completed: '已完成',
    overdue: '已逾期',
    closed: '已关闭',
    generated: '已生成',
    confirmed: '已确认',
    ignored: '已忽略',
    effective: '整改有效',
    weak: '效果不明显',
    ineffective: '整改无效',
    follow_up: '需继续跟进',
  }
  return map[String(status ?? '')] ?? String(status ?? '-')
}

function statusClass(status?: string) {
  if (['success', 'completed', 'confirmed', 'effective'].includes(String(status))) return 'success'
  if (['failed', 'overdue', 'critical', 'high', 'ineffective'].includes(String(status))) return 'danger'
  if (['warning', 'processing', 'generated', 'weak', 'follow_up'].includes(String(status))) return 'warning'
  return 'normal'
}

function levelText(level?: string) {
  const map: Record<string, string> = { critical: '严重', high: '高', medium: '中', low: '低' }
  return map[String(level ?? '')] ?? String(level ?? '-')
}

function taskStatusText(status?: string) {
  const map: Record<string, string> = { pending: '待处理', processing: '处理中', completed: '已完成', overdue: '已逾期' }
  return map[String(status ?? '')] ?? statusText(status)
}

function tasksByStatus(status: string) {
  return tasks.value.filter((item) => taskStatusText(item.status) === status)
}

function parseLooseJSON<T>(value: unknown, fallback: T): T {
  if (!value) return fallback
  if (typeof value !== 'string') return value as T
  try {
    return JSON.parse(value) as T
  } catch {
    return fallback
  }
}

function ruleConditions(rule: AnomalyRule) {
  const source = parseLooseJSON<Record<string, unknown>>((rule as unknown as Record<string, unknown>).metric_conditions_json ?? (rule as unknown as Record<string, unknown>).metric_conditions, {})
  const entries = Object.entries(source)
  if (!entries.length) return ['指标阈值由规则配置决定']
  return entries.map(([key, value]) => `${key}: ${typeof value === 'object' ? JSON.stringify(value) : String(value)}`)
}

function evidenceCards(row: AnomalyRecord | null) {
  const evidence = parseLooseJSON<Record<string, unknown>>(row?.evidence_json, {})
  return Object.entries(evidence).slice(0, 4).map(([key, value]) => ({
    label: key,
    value: typeof value === 'number' ? valueText(value) : String(value),
    desc: '指标证据',
  }))
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    if (section.value === 'dashboard') {
      const [summaryData, trendData, rankingData, anomalyData, taskData] = await Promise.all([
        fetchDashboardSummary(filters),
        fetchDashboardTrends(filters),
        fetchDashboardRankings(filters),
        fetchAnomalies({ ...filters, limit: 5 }),
        fetchTasks({ ...filters, limit: 5 }),
      ])
      summary.value = summaryData
      trends.value = listOrEmpty(trendData)
      rankings.value = {
        brand_rank: listOrEmpty(rankingData?.brand_rank),
        channel_rank: listOrEmpty(rankingData?.channel_rank),
      }
      anomalies.value = pageItems(anomalyData)
      tasks.value = pageItems(taskData)
    } else if (section.value === 'overview') {
      const [pipeData, batchData] = await Promise.all([fetchPipeline(), fetchRawBatches(filters)])
      pipeline.value = pageItems(pipeData)
      rawBatches.value = pageItems(batchData)
    } else if (section.value === 'raw') {
      rawBatches.value = pageItems(await fetchRawBatches(filters))
    } else if (section.value === 'standard') {
      standardRows.value = pageItems(await fetchStandardData(standardType.value, filters))
    } else if (section.value === 'metrics') {
      metrics.value = pageItems(await fetchMetrics(filters))
    } else if (section.value === 'rules') {
      rules.value = pageItems(await fetchRules(filters))
    } else if (section.value === 'anomalies') {
      anomalies.value = pageItems(await fetchAnomalies(filters))
      selectedAnomaly.value = anomalies.value[0] ?? null
    } else if (section.value === 'tasks') {
      tasks.value = pageItems(await fetchTasks(filters))
    } else if (section.value === 'reviews') {
      reviews.value = pageItems(await fetchReviews(filters))
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : '数据加载失败'
  } finally {
    loading.value = false
  }
}

async function runAnalyze(row: AnomalyRecord) {
  await ElMessageBox.confirm(`确认对异常「${row.title}」发起 AI 分析？`, '操作确认', { type: 'warning' })
  await analyzeAnomaly(row.id)
  ElMessage.success('AI 分析已生成')
  await load()
}

async function runGenerateTask(row: AnomalyRecord) {
  await ElMessageBox.confirm(`确认从异常「${row.title}」生成整改任务？`, '操作确认', { type: 'warning' })
  await generateTaskFromAnomaly(row.id)
  ElMessage.success('整改任务已生成')
  await load()
}

async function runTask(action: 'start' | 'complete', row: RectificationTask) {
  const actionText = action === 'start' ? '开始处理' : '完成'
  await ElMessageBox.confirm(`确认${actionText}任务「${row.title}」？`, '操作确认', { type: 'warning' })
  if (action === 'start') await startTask(row.id)
  else await completeTask(row.id)
  ElMessage.success('任务状态已更新')
  await load()
}

async function openRawDetail(row: RawBatch) {
  const [detail, errorsPage] = await Promise.all([fetchRawBatch(row.id), fetchRawErrors(row.id, { limit: 50 })])
  detailTitle.value = `原始批次 ${row.batch_code}`
  detailData.value = detail
  rawErrors.value = pageItems(errorsPage)
  detailDrawer.value = true
}

async function runReprocess(row: RawBatch) {
  await ElMessageBox.confirm(`确认重新处理批次「${row.batch_code}」？`, '操作确认', { type: 'warning' })
  await reprocessRawBatch(row.id)
  ElMessage.success('批次已重新处理')
  await load()
}

function openMetricForm(row?: MetricDefinition) {
  editingMetricId.value = row?.id ?? null
  metricForm.value = row ? { ...row } : { metric_code: '', metric_name: '', metric_category: 'sales', enabled: true, anomaly_enabled: true }
  metricDialog.value = true
}

async function saveMetric() {
  if (editingMetricId.value) await updateMetric(editingMetricId.value, metricForm.value)
  else await createMetric(metricForm.value)
  metricDialog.value = false
  ElMessage.success('指标已保存')
  await load()
}

async function toggleMetric(row: MetricDefinition) {
  await setMetricEnabled(row.id, !row.enabled)
  ElMessage.success('指标状态已更新')
  await load()
}

function openRuleForm(row?: AnomalyRule) {
  editingRuleId.value = row?.id ?? null
  ruleTestResult.value = null
  ruleForm.value = row
    ? { ...row }
    : { rule_code: '', rule_name: '', business_domain: '销售异常', target_object_type: 'brand', metric_conditions: {}, enabled: true, ai_enabled: true, task_enabled: true }
  ruleDialog.value = true
}

async function saveRule() {
  const payload = parseJSONFields(ruleForm.value, ['scope', 'metric_conditions', 'level_config', 'confidence_config'])
  if (editingRuleId.value) await updateRule(editingRuleId.value, payload)
  else await createRule(payload)
  ruleDialog.value = false
  ElMessage.success('规则已保存')
  await load()
}

async function openRuleDetail(row: AnomalyRule) {
  detailTitle.value = `规则 ${row.rule_code}`
  detailData.value = await fetchRule(row.id)
  detailDrawer.value = true
}

async function toggleRule(row: AnomalyRule) {
  await setRuleEnabled(row.id, !row.enabled)
  ElMessage.success('规则状态已更新')
  await load()
}

async function runRuleTest(row: AnomalyRule) {
  ruleTestResult.value = await testRule(row.id, { metrics: { gmv: 500, order_count: 5, ad_cost: 150, ad_roi: 1.2, available_days: 45, sales_7d: 2 } })
  detailTitle.value = `规则测试 ${row.rule_code}`
  detailData.value = ruleTestResult.value
  detailDrawer.value = true
}

async function runScan() {
  await scanAnomalies(filters)
  ElMessage.success('异常扫描完成')
  await load()
}

async function runReanalyze(row: AnomalyRecord) {
  await ElMessageBox.confirm(`确认重新分析异常「${row.title}」？`, '操作确认', { type: 'warning' })
  await reanalyzeAnomaly(row.id)
  ElMessage.success('AI 重新分析已生成')
  await load()
}

async function runAnomalyStatus(row: AnomalyRecord, action: 'confirm' | 'ignore' | 'close') {
  await updateAnomalyStatus(row.id, action)
  ElMessage.success('异常状态已更新')
  await load()
}

async function openTaskDetail(row: RectificationTask) {
  detailTitle.value = `任务 ${row.task_code}`
  detailData.value = await fetchTask(row.id)
  detailDrawer.value = true
}

function openTaskForm(row?: RectificationTask) {
  editingTaskId.value = row?.id ?? null
  taskForm.value = row ? { ...row } : { title: '', task_type: 'manual', priority: 'medium', progress: 0, target_desc: '' }
  taskDialog.value = true
}

async function saveTask() {
  if (editingTaskId.value) await updateTask(editingTaskId.value, taskForm.value)
  else await createTask(taskForm.value)
  taskDialog.value = false
  ElMessage.success('任务已保存')
  await load()
}

async function runFeedback(row: RectificationTask) {
  await feedbackTask(row.id, { content: row.title, progress: Math.min(100, Number(row.progress ?? 0) + 10) })
  ElMessage.success('反馈已提交')
  await load()
}

async function runCloseTask(row: RectificationTask) {
  await closeTask(row.id)
  ElMessage.success('任务已关闭')
  await load()
}

async function openReviewDetail(row: RectificationReview) {
  detailTitle.value = `复盘 ${row.review_code}`
  detailData.value = await fetchReview(row.id)
  detailDrawer.value = true
}

function openReviewForm(row?: RectificationReview) {
  editingReviewId.value = row?.id ?? null
  reviewForm.value = row ? { ...row } : { task_id: tasks.value[0]?.id, review_conclusion: 'follow_up', experience_summary: '' }
  reviewDialog.value = true
}

async function saveReview(confirm = false) {
  if (!editingReviewId.value) return
  if (confirm) await confirmReview(editingReviewId.value, reviewForm.value)
  else await updateReview(editingReviewId.value, reviewForm.value)
  reviewDialog.value = false
  ElMessage.success('复盘已保存')
  await load()
}

function parseJSONFields(source: Record<string, unknown>, fields: string[]) {
  const payload: Record<string, unknown> = { ...source }
  for (const field of fields) {
    if (typeof payload[field] === 'string' && String(payload[field]).trim()) {
      payload[field] = JSON.parse(String(payload[field]))
    }
  }
  return payload
}

function valueText(value: unknown, unit?: string) {
  const num = Number(value ?? 0)
  if (unit === 'CNY') return `¥${num.toLocaleString()}`
  if (unit === 'PERCENT') return `${num.toFixed(2)}%`
  if (unit === 'RATIO') return num.toFixed(2)
  return Number.isFinite(num) ? num.toLocaleString() : String(value ?? '-')
}

watch([section, standardType], load)
onMounted(load)
</script>

<template>
  <div class="dc-page">
    <main class="dc-main" v-loading="loading">
      <div v-if="error" class="dc-state error">{{ error }}</div>

      <section v-if="section === 'dashboard'" class="page">
        <div class="hero-panel">
          <div>
            <div class="eyebrow">经营看板</div>
            <h1>集团经营健康分 {{ summary?.health_score ?? 0 }}</h1>
            <p>系统基于销售、投流、退款、库存和整改进度综合计算，当前最大风险来自投流 ROI 下滑和部分门店 GMV 下滑。</p>
          </div>
          <div class="hero-actions">
            <button class="primary-btn" @click="router.push('/data-center/anomalies')">查看高风险异常</button>
            <button class="secondary-btn" @click="load">生成会议包</button>
          </div>
        </div>

        <div class="metric-grid">
          <article v-for="item in dashboardKpis" :key="item.label" class="metric-card">
            <div class="metric-label">{{ item.label }}</div>
            <div class="metric-value">{{ item.value }}</div>
            <div class="metric-row"><span class="change" :class="item.trend">{{ item.change }}</span><span>真实 API 汇总</span></div>
          </article>
        </div>

        <div class="content-grid two-col">
          <section class="chart-card">
            <div class="section-head"><div><h3>GMV 趋势</h3><p>近 7 天全渠道 GMV</p></div><span class="pill">万元</span></div>
            <div v-if="!gmvTrendRows.length" class="dc-empty">{{ emptyText }}</div>
            <div v-else class="bars">
              <div v-for="row in gmvTrendRows" :key="row.date" class="bar-col">
                <span class="bar" :style="{ height: `${percentBar(row.gmv, Math.max(...gmvTrendRows.map((item) => item.gmv), 1))}px` }"></span>
                <span>{{ row.date }}</span>
              </div>
            </div>
          </section>
          <section class="chart-card">
            <div class="section-head"><div><h3>ROI 趋势</h3><p>投流综合 ROI 波动</p></div><span class="pill">ROI</span></div>
            <div v-if="!roiTrendRows.length" class="dc-empty">{{ emptyText }}</div>
            <div v-else class="bars">
              <div v-for="row in roiTrendRows" :key="row.date" class="bar-col">
                <span class="bar" :style="{ height: `${percentBar(row.roi, Math.max(...roiTrendRows.map((item) => item.roi), 1))}px` }"></span>
                <span>{{ row.date }}</span>
              </div>
            </div>
          </section>
        </div>

        <div class="content-grid three-col">
          <div class="panel-card span-2">
            <div class="section-head">
              <div><h3>重点异常</h3><p>规则引擎识别，AI 已完成原因分析</p></div>
              <button class="ghost-btn" @click="router.push('/data-center/anomalies')">全部异常</button>
            </div>
            <div v-if="!anomalies.length" class="dc-empty">{{ emptyText }}</div>
            <div v-else class="anomaly-list compact">
              <div v-for="item in anomalies.slice(0, 3)" :key="item.id" class="anomaly-row">
                <div>
                  <span class="level" :class="levelText(item.anomaly_level)">{{ levelText(item.anomaly_level) }}</span>
                  <b>{{ item.title }}</b>
                  <p>{{ valueText(item.impact_amount, 'CNY') }} · 置信度 {{ item.confidence_score }}%</p>
                </div>
                <button class="secondary-btn" @click="selectedAnomaly = item; router.push('/data-center/anomalies')">查看分析</button>
              </div>
            </div>
          </div>
          <div class="panel-card">
            <div class="section-head"><div><h3>渠道贡献</h3><p>GMV 占比</p></div></div>
            <div v-if="!channelRankRows.length" class="dc-empty">{{ emptyText }}</div>
            <div v-else class="rank-list">
              <div v-for="item in channelRankRows" :key="item.name" class="rank-row">
                <div class="rank-top"><b>{{ item.name }}</b><span>{{ item.ratio }}%</span></div>
                <div class="rank-bar"><span :style="{ width: item.ratio + '%' }"></span></div>
                <p>{{ item.value.toFixed(2) }} 万</p>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section v-else-if="section === 'overview'" class="page">
        <div class="page-head">
          <div>
            <div class="eyebrow">数据总览</div>
            <h1>数据链路运行状态</h1>
            <p>查看原始数据接入、标准化清洗、指标计算、异常扫描、AI 分析和任务生成的链路状态。</p>
          </div>
          <button class="primary-btn" @click="load">重新执行链路检查</button>
        </div>

        <div class="pipeline">
          <div v-for="(item, index) in pipeline" :key="String(item.code)" class="pipeline-card" :class="statusClass(String(item.status))">
            <div class="pipeline-index">{{ index + 1 }}</div>
            <h3>{{ item.name }}</h3>
            <b>{{ item.value }}</b>
            <p>{{ statusText(String(item.status)) }}</p>
          </div>
        </div>

        <div class="panel-card">
          <div class="section-head">
            <div><h3>最近任务批次</h3><p>同步、清洗、指标、异常扫描任务执行情况</p></div>
            <button class="ghost-btn" @click="router.push('/data-center/raw')">查看日志</button>
          </div>
          <table class="data-table">
            <thead><tr><th>任务编号</th><th>类型</th><th>来源</th><th>执行时间</th><th>记录数</th><th>成功</th><th>失败</th><th>状态</th></tr></thead>
            <tbody>
              <tr v-for="job in rawBatches" :key="job.id">
                <td>{{ job.batch_code }}</td><td>{{ job.data_type }}</td><td>{{ job.platform_code || '-' }}</td><td>{{ dateText(job.sync_time) }}</td><td>{{ job.record_count.toLocaleString() }}</td><td>{{ job.success_count.toLocaleString() }}</td><td>{{ job.failed_count.toLocaleString() }}</td>
                <td><span class="status-badge" :class="statusClass(job.status)">{{ statusText(job.status) }}</span></td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      <section v-else-if="section === 'raw'" class="page">
        <div class="page-head">
          <div>
            <div class="eyebrow">原始数据</div>
            <h1>外部平台同步批次</h1>
            <p>保留外部平台原始数据批次，用于追溯、错误排查和重新清洗。</p>
          </div>
          <button class="primary-btn" @click="load">新增同步任务</button>
        </div>
        <div class="split-layout">
          <aside class="left-tabs">
            <button v-for="tab in rawTabs" :key="tab" :class="{ active: tab === '全部数据' }">{{ tab }}</button>
          </aside>
          <div class="panel-card fill">
            <div class="section-head">
              <div><h3>同步批次列表</h3><p>原始数据不直接参与分析，必须先进入标准数据。</p></div>
              <button class="ghost-btn" @click="load">重新清洗失败批次</button>
            </div>
            <table class="data-table">
              <thead><tr><th>批次号</th><th>数据类型</th><th>平台</th><th>连接实例</th><th>记录数</th><th>成功</th><th>失败</th><th>时间</th><th>状态</th><th>操作</th></tr></thead>
              <tbody>
                <tr v-for="item in rawBatches" :key="item.id">
                  <td>{{ item.batch_code }}</td><td>{{ item.data_type }}</td><td>{{ item.platform_code || '-' }}</td><td>{{ item.connection_code || '-' }}</td><td>{{ item.record_count.toLocaleString() }}</td><td>{{ item.success_count.toLocaleString() }}</td><td>{{ item.failed_count.toLocaleString() }}</td><td>{{ dateText(item.sync_time) }}</td>
                  <td><span class="status-badge" :class="statusClass(item.status)">{{ statusText(item.status) }}</span></td>
                  <td><button class="link-btn" @click="openRawDetail(item)">查看</button><button v-if="canAction('data_center:batch_retry')" class="link-btn" @click="runReprocess(item)">重清洗</button></td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </section>

      <section v-else-if="section === 'standard'" class="page">
        <div class="page-head">
          <div>
            <div class="eyebrow">标准数据</div>
            <h1>清洗后的统一业务数据</h1>
            <p>不同平台字段差异被统一到标准模型，后续指标计算只依赖标准数据。</p>
          </div>
          <button class="primary-btn" @click="load">查看清洗规则</button>
        </div>
        <div class="tabbar">
          <button :class="{ active: standardType === 'sales' }" @click="standardType = 'sales'">销售标准数据</button>
          <button :class="{ active: standardType === 'ad' }" @click="standardType = 'ad'">投流标准数据</button>
          <button :class="{ active: standardType === 'inventory' }" @click="standardType = 'inventory'">库存标准数据</button>
          <button :class="{ active: standardType === 'product' }" @click="standardType = 'product'">商品标准数据</button>
          <button :class="{ active: standardType === 'refund' }" @click="standardType = 'refund'">退款标准数据</button>
        </div>
        <div class="panel-card">
          <div class="section-head"><div><h3>{{ currentTitle }}</h3><p>数据来自数据库标准化结果，表头跟随当前标准模型。</p></div></div>
          <table class="data-table">
            <thead><tr><th v-for="key in Object.keys(standardRows[0] ?? { id: '' })" :key="key">{{ key }}</th></tr></thead>
            <tbody>
              <tr v-for="(row, index) in standardRows" :key="index">
                <td v-for="key in Object.keys(standardRows[0] ?? { id: '' })" :key="key">{{ row[key] }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      <section v-else-if="section === 'metrics'" class="page">
        <div class="page-head">
          <div>
            <div class="eyebrow">指标中心</div>
            <h1>经营指标定义与口径</h1>
            <p>指标是异常规则的输入，必须定义清楚计算口径、统计周期和可分析维度。</p>
          </div>
          <button v-if="canAction('data_center:metric_manage')" class="primary-btn" @click="openMetricForm()">新增指标</button>
        </div>
        <div class="content-grid four-col">
          <div v-for="item in metricSummaryCards" :key="item.label" class="summary-card"><span>{{ item.label }}</span><b>{{ item.value }}</b><p>{{ item.desc }}</p></div>
        </div>
        <div class="panel-card">
          <div class="section-head">
            <div><h3>指标定义</h3><p>第一版建议只维护核心经营指标，避免指标膨胀。</p></div>
            <button class="ghost-btn" @click="load">导出口径文档</button>
          </div>
          <table class="data-table">
            <thead><tr><th>指标 code</th><th>指标名称</th><th>分类</th><th>计算公式 / 口径</th><th>周期</th><th>统计维度</th><th>数据来源</th><th>异常判断</th><th>状态</th><th>操作</th></tr></thead>
            <tbody>
              <tr v-for="item in metrics" :key="item.id">
                <td>{{ item.metric_code }}</td><td><b>{{ item.metric_name }}</b></td><td>{{ item.metric_category }}</td><td>{{ item.formula || '-' }}</td><td>{{ item.statistic_period || '-' }}</td><td>-</td><td>{{ item.data_source || '-' }}</td><td>{{ item.anomaly_enabled ? '是' : '否' }}</td><td><span class="status-badge" :class="statusClass(item.enabled ? 'success' : 'warning')">{{ item.enabled ? '启用' : '禁用' }}</span></td>
                <td><button v-if="canAction('data_center:metric_manage')" class="link-btn" @click="openMetricForm(item)">编辑</button><button v-if="canAction('data_center:metric_manage')" class="link-btn" @click="toggleMetric(item)">{{ item.enabled ? '禁用' : '启用' }}</button></td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      <section v-else-if="section === 'rules'" class="page">
        <div class="page-head">
          <div>
            <div class="eyebrow">异常规则</div>
            <h1>定义什么是异常</h1>
            <p>异常必须由规则引擎基于指标判断，AI 不负责决定异常是否成立。</p>
          </div>
          <button v-if="canAction('data_center:rule_manage')" class="primary-btn" @click="openRuleForm()">新增异常规则</button>
        </div>
        <div class="split-layout">
          <aside class="left-tabs">
            <button v-for="tab in ruleTabs" :key="tab" :class="{ active: tab === '全部规则' }">{{ tab }}</button>
          </aside>
          <div class="panel-card fill">
            <div class="section-head">
              <div><h3>规则列表</h3><p>规则配置包括适用对象、触发条件、异常等级、置信度、任务和复盘。</p></div>
              <button class="ghost-btn" @click="rules[0] && runRuleTest(rules[0])">规则测试</button>
            </div>
            <div class="rule-list">
              <article v-for="rule in rules" :key="rule.id" class="rule-card">
                <div class="rule-head">
                  <div>
                    <div class="rule-title"><b>{{ rule.rule_name }}</b><span class="status-badge" :class="statusClass(rule.enabled ? 'success' : 'warning')">{{ rule.enabled ? '启用' : '禁用' }}</span></div>
                    <p>{{ rule.business_domain }} · 适用对象：{{ rule.target_object_type }} · 最近触发 {{ rule.priority || 0 }} 次</p>
                  </div>
                  <button v-if="canAction('data_center:rule_manage')" class="secondary-btn" @click="openRuleForm(rule)">编辑规则</button>
                </div>
                <div class="condition-box">
                  <span v-for="condition in ruleConditions(rule)" :key="condition">{{ condition }}</span>
                </div>
                <div class="rule-foot">
                  <span>异常等级：{{ rule.priority >= 3 ? '高' : '中' }}</span>
                  <span>AI 分析：{{ rule.ai_enabled ? '开启' : '关闭' }}</span>
                  <span>生成任务：{{ rule.task_enabled ? '开启' : '关闭' }}</span>
                  <span>默认责任：运营负责人</span>
                  <span>复盘：任务完成后自动进入</span>
                  <button class="link-btn" @click="openRuleDetail(rule)">详情</button>
                  <button v-if="canAction('data_center:rule_manage')" class="link-btn" @click="toggleRule(rule)">{{ rule.enabled ? '禁用' : '启用' }}</button>
                </div>
              </article>
            </div>
          </div>
        </div>
      </section>

      <section v-else-if="section === 'anomalies'" class="page drawer-page">
        <div class="page-main">
          <div class="page-head">
            <div>
              <div class="eyebrow">异常分析</div>
              <h1>规则引擎识别的经营异常</h1>
              <p>异常由指标规则确定，AI 只对异常证据做原因分析和整改建议。</p>
            </div>
            <button v-if="canAction('data_center:scan')" class="primary-btn" @click="runScan">扫描最新异常</button>
          </div>
          <div class="content-grid four-col">
            <div v-for="item in anomalySummaryCards" :key="item.label" class="summary-card" :class="{ danger: item.danger }"><span>{{ item.label }}</span><b>{{ item.value }}</b><p>{{ item.desc }}</p></div>
          </div>
          <div class="panel-card">
            <div class="section-head">
              <div><h3>异常列表</h3><p>每条异常都必须有触发规则、指标证据、AI 分析、任务状态。</p></div>
              <div class="filter-chips"><span>全部</span><span>高</span><span>投流异常</span><span>销售异常</span></div>
            </div>
            <div v-if="!anomalies.length" class="dc-empty">{{ emptyText }}</div>
            <div v-else class="anomaly-list">
              <div v-for="item in anomalies" :key="item.id" class="anomaly-item" @click="selectedAnomaly = item">
                <div class="anomaly-top">
                  <div>
                    <span class="level" :class="levelText(item.anomaly_level)">{{ levelText(item.anomaly_level) }}</span>
                    <b>{{ item.title }}</b>
                  </div>
                  <span class="confidence">{{ item.confidence_score }}%</span>
                </div>
                <p>{{ item.business_domain }} · {{ item.object_name || item.object_code }} · {{ dateText(item.occurred_at) }}</p>
                <div class="anomaly-meta">
                  <span>{{ valueText(item.impact_amount, 'CNY') }}</span>
                  <span>AI：{{ statusText(item.ai_status) }}</span>
                  <span>任务：{{ statusText(item.task_status) }}</span>
                  <span>复盘：{{ statusText(item.review_status) }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
        <aside class="detail-drawer">
          <template v-if="selectedAnomaly">
            <div class="drawer-head">
              <div><span class="level" :class="levelText(selectedAnomaly.anomaly_level)">{{ levelText(selectedAnomaly.anomaly_level) }}</span><h2>{{ selectedAnomaly.title }}</h2></div>
            </div>
            <div class="detail-grid two">
              <div class="detail-item"><span>影响金额</span><b>{{ valueText(selectedAnomaly.impact_amount, 'CNY') }}</b></div>
              <div class="detail-item"><span>置信度</span><b>{{ selectedAnomaly.confidence_score }}%</b></div>
              <div class="detail-item"><span>对象</span><b>{{ selectedAnomaly.object_name || selectedAnomaly.object_code }}</b></div>
              <div class="detail-item"><span>状态</span><b>{{ statusText(selectedAnomaly.status) }}</b></div>
            </div>
            <div class="drawer-section">
              <h3>指标证据</h3>
              <div class="evidence-grid">
                <div v-for="item in evidenceCards(selectedAnomaly)" :key="item.label" class="evidence-card"><span>{{ item.label }}</span><b>{{ item.value }}</b><p>{{ item.desc }}</p></div>
              </div>
            </div>
            <div class="drawer-section">
              <h3>AI 分析摘要</h3>
              <ol class="reason-list"><li>基于异常证据与历史趋势识别主要波动原因。</li><li>建议生成整改任务，并在任务完成后进入复盘。</li></ol>
            </div>
            <div class="drawer-actions">
              <button v-if="canAction('data_center:ai_analyze')" class="secondary-btn" @click="runAnalyze(selectedAnomaly)">触发 AI 分析</button>
              <button v-if="canAction('data_center:ai_analyze')" class="secondary-btn" @click="runReanalyze(selectedAnomaly)">重新分析</button>
              <button v-if="canAction('data_center:task_generate')" class="primary-btn" @click="runGenerateTask(selectedAnomaly)">生成整改任务</button>
            </div>
          </template>
          <div v-else class="dc-empty">{{ emptyText }}</div>
        </aside>
      </section>

      <section v-else-if="section === 'tasks'" class="page">
        <div class="page-head">
          <div>
            <div class="eyebrow">整改任务</div>
            <h1>异常生成的整改闭环</h1>
            <p>任务必须关联来源异常、AI 建议、责任人、整改目标和复盘指标。</p>
          </div>
          <button v-if="canAction('data_center:task_flow')" class="primary-btn" @click="openTaskForm()">新建整改任务</button>
        </div>
        <div class="kanban">
          <div v-for="status in taskStatuses" :key="status" class="kanban-col">
            <div class="kanban-head"><b>{{ status }}</b><span>{{ tasksByStatus(status).length }}</span></div>
            <article v-for="task in tasksByStatus(status)" :key="task.id" class="task-card">
              <div class="task-top"><b>{{ task.title }}</b><span :class="['priority', levelText(task.priority)]">{{ levelText(task.priority) }}</span></div>
              <p>来源异常：{{ task.anomaly_code || '-' }}</p>
              <p>责任人：{{ task.owner_role || '运营负责人' }}</p>
              <p>协同人：跨部门协同</p>
              <div class="progress"><span :style="{ width: `${Number(task.progress ?? 0)}%` }"></span></div>
              <div class="task-foot"><span>截止：{{ dateText(task.deadline) }}</span><button class="link-btn" @click="openTaskDetail(task)">详情</button></div>
            </article>
          </div>
        </div>
        <div class="panel-card">
          <div class="section-head"><div><h3>任务列表</h3><p>用于批量查看、导出和会议复盘。</p></div></div>
          <table class="data-table">
            <thead><tr><th>任务编号</th><th>任务标题</th><th>来源异常</th><th>责任人</th><th>优先级</th><th>截止时间</th><th>整改目标</th><th>状态</th><th>操作</th></tr></thead>
            <tbody>
              <tr v-for="task in tasks" :key="task.id">
                <td>{{ task.task_code }}</td><td><b>{{ task.title }}</b></td><td>{{ task.anomaly_code || '-' }}</td><td>{{ task.owner_role || '-' }}</td><td>{{ levelText(task.priority) }}</td><td>{{ dateText(task.deadline) }}</td><td>{{ task.target_desc || '-' }}</td><td>{{ taskStatusText(task.status) }}</td>
                <td><button class="link-btn" @click="openTaskDetail(task)">详情</button><button v-if="canAction('data_center:task_flow')" class="link-btn" @click="openTaskForm(task)">编辑</button><button v-if="canAction('data_center:task_flow')" class="link-btn" @click="runTask('start', task)">开始</button><button v-if="canAction('data_center:task_flow')" class="link-btn" @click="runTask('complete', task)">完成</button></td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      <section v-else-if="section === 'reviews'" class="page">
        <div class="page-head">
          <div>
            <div class="eyebrow">整改复盘</div>
            <h1>验证整改是否有效</h1>
            <p>复盘对比整改前后指标，判断任务是否真的解决异常，并沉淀经验。</p>
          </div>
          <button class="primary-btn" @click="load">生成复盘报告</button>
        </div>
        <div class="content-grid four-col">
          <div v-for="item in reviewSummaryCards" :key="item.label" class="summary-card" :class="{ danger: item.danger }"><span>{{ item.label }}</span><b>{{ item.value }}</b><p>{{ item.desc }}</p></div>
        </div>
        <div class="content-grid two-col">
          <div class="panel-card">
            <div class="section-head"><div><h3>复盘记录</h3><p>每条复盘都关联任务和来源异常。</p></div></div>
            <div class="review-list">
              <div v-for="item in reviews" :key="item.id" class="review-item" @click="openReviewDetail(item)">
                <div><b>{{ item.task_code }}</b><p>{{ item.review_code }} · {{ item.anomaly_code || '-' }} · {{ dateText(item.reviewed_at) }}</p></div>
                <span class="review-result" :class="item.review_conclusion === 'effective' ? 'good' : 'bad'">{{ statusText(item.review_conclusion) }}</span>
              </div>
            </div>
          </div>
          <div class="panel-card">
            <div class="section-head"><div><h3>复盘详情示例</h3><p>整改前后指标对比</p></div></div>
            <div v-if="selectedReview" class="review-detail">
              <h2>{{ selectedReview.task_code }}</h2>
              <div class="compare-box">
                <div><span>整改前</span><b>异常触发</b></div>
                <div><span>整改后</span><b>{{ statusText(selectedReview.review_conclusion) }}</b></div>
                <div><span>改善幅度</span><b>{{ selectedReview.improvement_result || '-' }}</b></div>
              </div>
              <div class="ai-summary">
                <h3>AI 复盘总结</h3>
                <p>整改动作执行后，核心指标已重新验证。建议将有效动作沉淀为标准经验，并纳入后续门店巡检清单。</p>
              </div>
              <button v-if="canAction('data_center:review_confirm')" class="secondary-btn" @click="openReviewForm(selectedReview)">沉淀为经验</button>
            </div>
          </div>
        </div>
      </section>

      <el-drawer v-model="detailDrawer" :title="detailTitle" size="46%">
        <section v-if="rawErrors.length" class="dc-drawer-block">
          <h4>错误明细</h4>
          <el-table :data="rawErrors" border>
            <el-table-column v-for="key in Object.keys(rawErrors[0] ?? {})" :key="key" :prop="key" :label="key" min-width="140" />
          </el-table>
        </section>
        <pre>{{ JSON.stringify(detailData, null, 2) }}</pre>
      </el-drawer>

      <el-dialog v-model="metricDialog" title="指标配置" width="560px">
        <el-form label-width="110px">
          <el-form-item label="指标 code"><el-input v-model="metricForm.metric_code" /></el-form-item>
          <el-form-item label="指标名称"><el-input v-model="metricForm.metric_name" /></el-form-item>
          <el-form-item label="分类"><el-input v-model="metricForm.metric_category" /></el-form-item>
          <el-form-item label="公式"><el-input v-model="metricForm.formula" type="textarea" /></el-form-item>
          <el-form-item label="启用"><el-switch v-model="metricForm.enabled" /></el-form-item>
        </el-form>
        <template #footer><button @click="saveMetric">保存</button></template>
      </el-dialog>

      <el-dialog v-model="ruleDialog" title="异常规则" width="640px">
        <el-form label-width="120px">
          <el-form-item label="规则 code"><el-input v-model="ruleForm.rule_code" /></el-form-item>
          <el-form-item label="规则名称"><el-input v-model="ruleForm.rule_name" /></el-form-item>
          <el-form-item label="业务域"><el-input v-model="ruleForm.business_domain" /></el-form-item>
          <el-form-item label="对象类型"><el-input v-model="ruleForm.target_object_type" /></el-form-item>
          <el-form-item label="条件 JSON"><el-input v-model="ruleForm.metric_conditions" type="textarea" :rows="5" /></el-form-item>
          <el-form-item label="启用"><el-switch v-model="ruleForm.enabled" /></el-form-item>
        </el-form>
        <template #footer><button @click="saveRule">保存</button></template>
      </el-dialog>

      <el-dialog v-model="taskDialog" title="整改任务" width="560px">
        <el-form label-width="110px">
          <el-form-item label="标题"><el-input v-model="taskForm.title" /></el-form-item>
          <el-form-item label="优先级"><el-input v-model="taskForm.priority" /></el-form-item>
          <el-form-item label="进度"><el-input-number v-model="taskForm.progress" :min="0" :max="100" /></el-form-item>
          <el-form-item label="目标"><el-input v-model="taskForm.target_desc" type="textarea" /></el-form-item>
        </el-form>
        <template #footer><button @click="saveTask">保存</button></template>
      </el-dialog>

      <el-dialog v-model="reviewDialog" title="整改复盘" width="620px">
        <el-form label-width="120px">
          <el-form-item label="结论"><el-input v-model="reviewForm.review_conclusion" /></el-form-item>
          <el-form-item label="人工总结"><el-input v-model="reviewForm.manual_review_summary" type="textarea" /></el-form-item>
          <el-form-item label="经验沉淀"><el-input v-model="reviewForm.experience_summary" type="textarea" /></el-form-item>
        </el-form>
        <template #footer>
          <button @click="saveReview(false)">保存</button>
          <button @click="saveReview(true)">确认结论</button>
        </template>
      </el-dialog>
    </main>
  </div>
</template>
