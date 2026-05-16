<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'

import NeuroAgentPageShell from '@/views/components/NeuroAgentPageShell.vue'

import { fetchAiOverview } from '../api'
import type { AiOverview } from '../types'
import { modelTypeText, moneyText, numberText, percentText, statusText, statusType, text } from './viewHelpers'
import './aiPrototype.css'

defineOptions({ name: 'AiDashboardView' })

const loading = ref(false)
const overview = ref<AiOverview | null>(null)
const trendMode = ref<'calls' | 'cost' | 'success'>('calls')

const trendModeOptions = [
  { value: 'calls', label: '调用', title: '调用趋势', description: '近 7 天真实调用量走势', unit: '' },
  { value: 'cost', label: '成本', title: '成本趋势', description: '近 7 天真实成本走势', unit: '¥' },
  { value: 'success', label: '成功率', title: '成功率趋势', description: '近 7 天真实调用成功率走势', unit: '%' },
] as const
const activeTrendMode = computed(() => trendModeOptions.find((item) => item.value === trendMode.value) ?? trendModeOptions[0])

const costMax = computed(() => Math.max(...(overview.value?.model_cost_share ?? []).map((item) => Number(item.cost_amount ?? 0)), 1))
const chartWidth = 720
const chartHeight = 260
const chartPadding = 36
const trendPoints = computed(() => {
  const data = overview.value?.usage_trend ?? []
  if (!data.length) return []
  const values = data.map((item) => trendValue(item))
  const max = trendMode.value === 'success' ? 100 : Math.max(...values)
  const min = trendMode.value === 'success' ? 0 : Math.min(...values)
  const range = max - min
  const innerWidth = chartWidth - chartPadding * 2
  const innerHeight = chartHeight - chartPadding * 2
  return data.map((item, index) => {
    const x = chartPadding + (index * innerWidth) / Math.max(data.length - 1, 1)
    const ratio = range === 0 ? 0.5 : (trendValue(item) - min) / range
    const y = chartHeight - chartPadding - ratio * innerHeight
    return { x, y, item }
  })
})
const trendPath = computed(() => trendPoints.value.map((point, index) => `${index === 0 ? 'M' : 'L'} ${point.x} ${point.y}`).join(' '))
const trendAreaPath = computed(() => {
  const points = trendPoints.value
  if (!points.length) return ''
  return `${trendPath.value} L ${points[points.length - 1].x} ${chartHeight - chartPadding} L ${points[0].x} ${chartHeight - chartPadding} Z`
})
const gatewayChecks = computed(() => (overview.value?.health_checks ?? []).filter((item) => item.name !== '底座操作日志'))
const gatewayReady = computed(() => gatewayChecks.value.length > 0 && gatewayChecks.value.every((item) => item.status === 'active' || item.status === 'success'))
const gatewayStatusText = computed(() => {
  if (!overview.value) return 'AI Gateway 检查中'
  return gatewayReady.value ? 'AI Gateway 正常运行' : 'AI Gateway 待处理'
})
const gatewayHealthStats = computed(() => {
  const result = { normal: 0, warning: 0, error: 0, total: gatewayChecks.value.length }
  for (const item of gatewayChecks.value) {
    const status = String(item.status || '').toLowerCase()
    if (['active', 'success', 'enabled', 'online'].includes(status)) result.normal += 1
    else if (['warning', 'degraded', 'pending', 'draft'].includes(status)) result.warning += 1
    else if (['error', 'failed', 'timeout', 'inactive', 'disabled', 'offline'].includes(status)) result.error += 1
  }
  return result
})

async function loadData() {
  loading.value = true
  try {
    overview.value = await fetchAiOverview()
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '总览加载失败')
  } finally {
    loading.value = false
  }
}

function trendValue(item: Record<string, unknown>) {
  if (trendMode.value === 'cost') return Number(item.cost_amount ?? 0)
  if (trendMode.value === 'success') return Number(item.success_rate ?? 0)
  return Number(item.calls ?? 0)
}

function trendValueText(item: Record<string, unknown>) {
  if (trendMode.value === 'cost') return moneyText(item.cost_amount)
  if (trendMode.value === 'success') return percentText(item.success_rate)
  return numberText(item.calls)
}

onMounted(loadData)
</script>

<template>
  <NeuroAgentPageShell class="ai-prototype" :show-hero="false">
    <template #title>AI 能力中心</template>
    <template #subtitle>今日真实调用、近 7 天真实趋势、租户排行和 AI Gateway 健康检查。</template>

    <div class="ai-hero-card">
      <div class="ai-hero-copy">
        <div class="ai-hero-status">
          <span class="ai-tag" :class="{ 'is-warning': overview && !gatewayReady }">{{ gatewayStatusText }}</span>
        </div>
        <h2>统一管理模型、供应商、基础路由、场景策略、用量与价格</h2>
        <p>业务中心只传租户、应用和 AI 场景；平台完成鉴权、配额校验、模型路由、降级、用量沉淀，配置操作写入底座操作日志。</p>
      </div>
      <div class="ai-gateway-panel">
        <div class="ai-gateway-panel__stats">
          <span class="is-normal"><strong>{{ gatewayHealthStats.normal }}</strong><small>正常</small></span>
          <span class="is-error"><strong>{{ gatewayHealthStats.error }}</strong><small>异常</small></span>
          <span class="is-warning"><strong>{{ gatewayHealthStats.warning }}</strong><small>告警</small></span>
          <span class="is-total"><strong>{{ gatewayHealthStats.total }}</strong><small>检查项</small></span>
        </div>
      </div>
    </div>

    <div class="ai-metric-grid" v-loading="loading">
      <div v-for="metric in overview?.metrics ?? []" :key="metric.label" class="ai-mini-stat">
        <span>{{ metric.label }}</span>
        <strong>{{ metric.value }}</strong>
        <small class="ai-muted">{{ metric.trend }}</small>
      </div>
    </div>

    <div class="ai-dashboard-grid">
      <section class="ai-card">
        <header class="ai-card__header">
          <div>
            <h3>{{ activeTrendMode.title }}</h3>
            <p class="ai-card__description">{{ activeTrendMode.description }}</p>
          </div>
          <div class="ai-trend-switch" role="tablist" aria-label="趋势指标切换">
            <button
              v-for="option in trendModeOptions"
              :key="option.value"
              type="button"
              :class="{ active: trendMode === option.value }"
              @click="trendMode = option.value"
            >
              {{ option.label }}
            </button>
          </div>
        </header>
        <div class="ai-card__body">
          <div class="ai-line-chart" role="img" aria-label="调用趋势折线图">
            <svg :viewBox="`0 0 ${chartWidth} ${chartHeight}`">
              <defs>
                <linearGradient id="aiTrendArea" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stop-color="currentColor" stop-opacity="0.2" />
                  <stop offset="100%" stop-color="currentColor" stop-opacity="0" />
                </linearGradient>
              </defs>
              <line
                v-for="tick in [0, 1, 2, 3]"
                :key="tick"
                :x1="chartPadding"
                :x2="chartWidth - chartPadding"
                :y1="chartPadding + (tick * (chartHeight - chartPadding * 2)) / 3"
                :y2="chartPadding + (tick * (chartHeight - chartPadding * 2)) / 3"
                class="ai-chart-grid"
              />
              <path v-if="trendAreaPath" :d="trendAreaPath" class="ai-chart-area" />
              <path v-if="trendPath" :d="trendPath" class="ai-chart-line" />
              <g v-for="point in trendPoints" :key="String(point.item.date)">
                <circle :cx="point.x" :cy="point.y" r="4" class="ai-chart-dot" />
                <text :x="point.x" :y="chartHeight - 10" text-anchor="middle" class="ai-chart-label">{{ text(point.item.date).slice(5) }}</text>
                <text :x="point.x" :y="Math.max(18, point.y - 10)" text-anchor="middle" class="ai-chart-value">{{ trendValueText(point.item) }}</text>
              </g>
            </svg>
          </div>
        </div>
      </section>

      <section class="ai-card">
        <header class="ai-card__header">
          <div>
            <h3>成本结构</h3>
            <p class="ai-card__description">按模型类型拆分</p>
          </div>
        </header>
        <div class="ai-card__body ai-chart-list">
          <div v-for="item in overview?.model_cost_share ?? []" :key="String(item.model_type)" class="ai-bar-row">
            <span>{{ modelTypeText(item.model_type) }}</span>
            <span class="ai-bar"><i :style="{ width: `${Math.max(4, (Number(item.cost_amount ?? 0) / costMax) * 100)}%` }" /></span>
            <strong>{{ moneyText(item.cost_amount) }}</strong>
          </div>
        </div>
      </section>
    </div>

    <section class="ai-card ai-table">
      <header class="ai-card__header">
        <div>
          <h3>租户排行榜</h3>
          <p class="ai-card__description">按销售额排序，展示租户调用量、成本、毛利和成功率。</p>
        </div>
      </header>
      <div class="ai-card__body">
        <el-table :data="overview?.tenant_ranking ?? []" border empty-text="暂无近 7 天真实租户调用。">
          <el-table-column label="租户" min-width="180">
            <template #default="{ row }">
              <span class="ai-table-cell-main"><strong>{{ text(row.tenant_name) }}</strong><small>{{ numberText(row.scenario_count) }} 个 AI 场景</small></span>
            </template>
          </el-table-column>
          <el-table-column label="调用量" width="130"><template #default="{ row }">{{ numberText(row.calls) }}</template></el-table-column>
          <el-table-column label="销售额" width="130"><template #default="{ row }">{{ moneyText(row.billing_amount) }}</template></el-table-column>
          <el-table-column label="成本" width="130"><template #default="{ row }">{{ moneyText(row.cost_amount) }}</template></el-table-column>
          <el-table-column label="毛利" width="130"><template #default="{ row }">{{ moneyText(row.profit_amount) }}</template></el-table-column>
          <el-table-column label="成功率" width="120"><template #default="{ row }">{{ percentText(row.success_rate) }}</template></el-table-column>
        </el-table>
      </div>
    </section>

    <div class="ai-insight-grid">
      <section class="ai-card">
        <header class="ai-card__header">
          <div>
            <h3>平台健康检查</h3>
          <p class="ai-card__description">供应商、密钥、限流和可执行路由状态</p>
          </div>
        </header>
        <div class="ai-card__body ai-health-list">
          <div v-for="item in overview?.health_checks ?? []" :key="item.name" class="ai-health-row">
            <el-tag :type="statusType(item.status)" class="ai-health-row__status">{{ statusText(item.status) }}</el-tag>
            <strong class="ai-health-row__name">{{ item.name }}</strong>
            <span class="ai-health-row__message ai-muted">{{ item.message }}</span>
          </div>
        </div>
      </section>

      <section class="ai-card ai-table">
        <header class="ai-card__header">
          <div>
            <h3>核心基础路由</h3>
            <p class="ai-card__description">可复用路由策略，供 AI 场景绑定</p>
          </div>
        </header>
        <div class="ai-card__body">
          <el-table :data="overview?.core_base_routes ?? []" border empty-text="暂无可展示的核心基础路由。">
            <el-table-column label="基础路由" min-width="180"><template #default="{ row }"><strong>{{ text(row.route_name) }}</strong></template></el-table-column>
            <el-table-column label="能力" min-width="140"><template #default="{ row }">{{ text(row.capability_code) }}</template></el-table-column>
            <el-table-column label="策略" width="130"><template #default="{ row }"><el-tag>{{ text(row.strategy) }}</el-tag></template></el-table-column>
            <el-table-column label="状态" width="100"><template #default="{ row }"><el-tag :type="statusType(row.status)">{{ statusText(row.status) }}</el-tag></template></el-table-column>
          </el-table>
        </div>
      </section>
    </div>
  </NeuroAgentPageShell>
</template>
