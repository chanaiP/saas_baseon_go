<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
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
const permissionStore = usePermissionStore()
const section = computed(() => String(route.params.section || 'dashboard'))
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

const currentTitle = computed(() => sections.find((item) => item.key === section.value)?.label ?? '经营看板')
const emptyText = computed(() => (error.value ? error.value : '暂无真实数据库记录'))

function listOrEmpty<T>(value: T[] | null | undefined): T[] {
  return Array.isArray(value) ? value : []
}

function pageItems<T>(value: { items?: T[] | null } | null | undefined): T[] {
  return listOrEmpty(value?.items)
}

function canAction(code: string) {
  return permissionStore.canUseAction(code)
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
      <header class="dc-header">
        <div>
          <p>Data Center</p>
          <h1>{{ currentTitle }}</h1>
        </div>
        <div class="dc-filters">
          <select v-model="filters.time_range" @change="load">
            <option value="today">今日</option>
            <option value="last_7_days">近 7 天</option>
            <option value="last_30_days">近 30 天</option>
            <option value="this_month">本月</option>
            <option value="last_month">上月</option>
          </select>
          <input v-model="filters.brand_code" placeholder="品牌 code" @keyup.enter="load" />
          <input v-model="filters.platform_code" placeholder="平台 code" @keyup.enter="load" />
          <button @click="load">刷新</button>
        </div>
      </header>

      <div v-if="error" class="dc-state error">{{ error }}</div>

      <section v-if="section === 'dashboard'" class="dc-grid">
        <article class="dc-hero">
          <span>经营健康分</span>
          <strong>{{ summary?.health_score ?? 0 }}</strong>
          <p>来自真实订单、投流、异常和任务数据的聚合结果。</p>
        </article>
        <article v-for="item in summary?.kpis ?? []" :key="item.code" class="dc-card">
          <span>{{ item.label }}</span>
          <strong>{{ valueText(item.value, item.unit) }}</strong>
        </article>
        <section class="dc-panel wide">
          <h3>趋势</h3>
          <div v-if="!trends.length" class="dc-empty">{{ emptyText }}</div>
          <div v-else class="dc-bars">
            <div v-for="row in trends" :key="String(row.date)">
              <i :style="{ height: `${Math.max(8, Number(row.gmv ?? 0) / 100)}px` }"></i>
              <span>{{ row.date }}</span>
            </div>
          </div>
        </section>
        <section class="dc-panel">
          <h3>品牌排行</h3>
          <p v-for="row in rankings.brand_rank ?? []" :key="String(row.code)">{{ row.code }} · {{ valueText(row.value, 'CNY') }}</p>
          <div v-if="!(rankings.brand_rank ?? []).length" class="dc-empty">{{ emptyText }}</div>
        </section>
      </section>

      <section v-else-if="section === 'overview'" class="dc-grid">
        <article v-for="item in pipeline" :key="String(item.code)" class="dc-card">
          <span>{{ item.name }}</span>
          <strong>{{ item.value }}</strong>
          <small>{{ item.status }}</small>
        </article>
      </section>

      <section v-else-if="section === 'raw'" class="dc-panel">
        <h3>原始数据批次</h3>
        <el-table :data="rawBatches" border :empty-text="emptyText">
          <el-table-column prop="batch_code" label="批次号" min-width="180" />
          <el-table-column prop="data_type" label="数据类型" />
          <el-table-column prop="platform_code" label="平台" />
          <el-table-column prop="record_count" label="记录数" />
          <el-table-column prop="failed_count" label="失败" />
          <el-table-column prop="status" label="状态" />
          <el-table-column label="操作" width="190">
            <template #default="{ row }">
              <button @click="openRawDetail(row)">详情</button>
              <button v-if="canAction('data_center:batch_retry')" @click="runReprocess(row)">重处理</button>
            </template>
          </el-table-column>
        </el-table>
      </section>

      <section v-else-if="section === 'standard'" class="dc-panel">
        <div class="dc-tabs">
          <button :class="{ active: standardType === 'sales' }" @click="standardType = 'sales'">销售标准数据</button>
          <button :class="{ active: standardType === 'ad' }" @click="standardType = 'ad'">投流标准数据</button>
          <button :class="{ active: standardType === 'inventory' }" @click="standardType = 'inventory'">库存标准数据</button>
          <button :class="{ active: standardType === 'refund' }" @click="standardType = 'refund'">退款标准数据</button>
          <button :class="{ active: standardType === 'product' }" @click="standardType = 'product'">商品标准数据</button>
          <button :class="{ active: standardType === 'store-sales' }" @click="standardType = 'store-sales'">门店销售数据</button>
        </div>
        <el-table :data="standardRows" border :empty-text="emptyText">
          <el-table-column v-for="key in Object.keys(standardRows[0] ?? { id: '' })" :key="key" :prop="key" :label="key" min-width="140" />
        </el-table>
      </section>

      <section v-else-if="section === 'metrics'" class="dc-panel">
        <div class="dc-panel-head">
          <h3>指标定义</h3>
          <button v-if="canAction('data_center:metric_manage')" @click="openMetricForm()">新建指标</button>
        </div>
        <el-table :data="metrics" border :empty-text="emptyText">
          <el-table-column prop="metric_code" label="指标 code" />
          <el-table-column prop="metric_name" label="指标名称" />
          <el-table-column prop="metric_category" label="分类" />
          <el-table-column prop="formula" label="公式" min-width="220" />
          <el-table-column prop="enabled" label="启用" />
          <el-table-column label="操作" width="190">
            <template #default="{ row }">
              <button v-if="canAction('data_center:metric_manage')" @click="openMetricForm(row)">编辑</button>
              <button v-if="canAction('data_center:metric_manage')" @click="toggleMetric(row)">{{ row.enabled ? '禁用' : '启用' }}</button>
            </template>
          </el-table-column>
        </el-table>
      </section>

      <section v-else-if="section === 'rules'" class="dc-panel">
        <div class="dc-panel-head">
          <h3>异常规则</h3>
          <button v-if="canAction('data_center:rule_manage')" @click="openRuleForm()">新建规则</button>
        </div>
        <el-table :data="rules" border :empty-text="emptyText">
          <el-table-column prop="rule_code" label="规则 code" />
          <el-table-column prop="rule_name" label="规则名称" />
          <el-table-column prop="business_domain" label="业务域" />
          <el-table-column prop="target_object_type" label="适用对象" />
          <el-table-column prop="enabled" label="启用" />
          <el-table-column label="操作" width="260">
            <template #default="{ row }">
              <button @click="openRuleDetail(row)">详情</button>
              <button v-if="canAction('data_center:rule_manage')" @click="openRuleForm(row)">编辑</button>
              <button v-if="canAction('data_center:rule_manage')" @click="runRuleTest(row)">测试</button>
              <button v-if="canAction('data_center:rule_manage')" @click="toggleRule(row)">{{ row.enabled ? '禁用' : '启用' }}</button>
            </template>
          </el-table-column>
        </el-table>
      </section>

      <section v-else-if="section === 'anomalies'" class="dc-split">
        <div class="dc-panel">
          <div class="dc-panel-head">
            <h3>异常列表</h3>
            <button v-if="canAction('data_center:scan')" @click="runScan">扫描</button>
          </div>
          <div v-if="!anomalies.length" class="dc-empty">{{ emptyText }}</div>
          <article v-for="row in anomalies" :key="row.id" class="dc-list-row" @click="selectedAnomaly = row">
            <b>{{ row.title }}</b>
            <span>{{ row.anomaly_level }} · {{ row.confidence_score }}% · {{ row.status }}</span>
          </article>
        </div>
        <aside class="dc-panel">
          <h3>异常详情</h3>
          <template v-if="selectedAnomaly">
            <h4>{{ selectedAnomaly.title }}</h4>
            <p>影响金额：{{ valueText(selectedAnomaly.impact_amount, 'CNY') }}</p>
            <p>AI：{{ selectedAnomaly.ai_status }} · 任务：{{ selectedAnomaly.task_status }} · 复盘：{{ selectedAnomaly.review_status }}</p>
            <pre>{{ selectedAnomaly.evidence_json }}</pre>
            <button v-if="canAction('data_center:ai_analyze')" @click="runAnalyze(selectedAnomaly)">触发 AI 分析</button>
            <button v-if="canAction('data_center:ai_analyze')" @click="runReanalyze(selectedAnomaly)">重新分析</button>
            <button v-if="canAction('data_center:task_generate')" @click="runGenerateTask(selectedAnomaly)">生成整改任务</button>
            <button v-if="canAction('data_center:task_generate')" @click="runAnomalyStatus(selectedAnomaly, 'confirm')">确认</button>
            <button v-if="canAction('data_center:task_generate')" @click="runAnomalyStatus(selectedAnomaly, 'ignore')">忽略</button>
            <button v-if="canAction('data_center:task_generate')" @click="runAnomalyStatus(selectedAnomaly, 'close')">关闭</button>
          </template>
          <div v-else class="dc-empty">{{ emptyText }}</div>
        </aside>
      </section>

      <section v-else-if="section === 'tasks'" class="dc-panel">
        <div class="dc-panel-head">
          <h3>整改任务</h3>
          <button v-if="canAction('data_center:task_flow')" @click="openTaskForm()">新建任务</button>
        </div>
        <el-table :data="tasks" border :empty-text="emptyText">
          <el-table-column prop="task_code" label="任务编号" />
          <el-table-column prop="title" label="任务标题" min-width="220" />
          <el-table-column prop="priority" label="优先级" />
          <el-table-column prop="progress" label="进度" />
          <el-table-column prop="status" label="状态" />
          <el-table-column label="操作" width="300">
            <template #default="{ row }">
              <button @click="openTaskDetail(row)">详情</button>
              <button v-if="canAction('data_center:task_flow')" @click="openTaskForm(row)">编辑</button>
              <button v-if="canAction('data_center:task_flow')" @click="runTask('start', row)">开始</button>
              <button v-if="canAction('data_center:task_flow')" @click="runFeedback(row)">反馈</button>
              <button v-if="canAction('data_center:task_flow')" @click="runTask('complete', row)">完成</button>
              <button v-if="canAction('data_center:task_flow')" @click="runCloseTask(row)">关闭</button>
            </template>
          </el-table-column>
        </el-table>
      </section>

      <section v-else-if="section === 'reviews'" class="dc-panel">
        <h3>整改复盘</h3>
        <el-table :data="reviews" border :empty-text="emptyText">
          <el-table-column prop="review_code" label="复盘编号" />
          <el-table-column prop="task_code" label="任务编号" />
          <el-table-column prop="improvement_result" label="改善幅度" />
          <el-table-column prop="review_conclusion" label="结论" />
          <el-table-column prop="reviewed_at" label="复盘时间" />
          <el-table-column label="操作" width="210">
            <template #default="{ row }">
              <button @click="openReviewDetail(row)">详情</button>
              <button v-if="canAction('data_center:review_confirm')" @click="openReviewForm(row)">编辑</button>
            </template>
          </el-table-column>
        </el-table>
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
