<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh } from '@element-plus/icons-vue'

import NeuroAgentPageShell from '@/views/components/NeuroAgentPageShell.vue'

import { fetchAiOverview } from '../api'
import type { AiOverview } from '../types'
import { moneyText, numberText, percentText, statusType, text } from './viewHelpers'
import './aiPrototype.css'

defineOptions({ name: 'AiDashboardView' })

const loading = ref(false)
const overview = ref<AiOverview | null>(null)

const costMax = computed(() => Math.max(...(overview.value?.model_cost_share ?? []).map((item) => Number(item.cost_amount ?? 0)), 1))
const chartWidth = 720
const chartHeight = 260
const chartPadding = 36
const trendPoints = computed(() => {
  const data = overview.value?.usage_trend ?? []
  if (!data.length) return []
  const calls = data.map((item) => Number(item.calls ?? 0))
  const max = Math.max(...calls)
  const min = Math.min(...calls)
  const range = max - min
  const innerWidth = chartWidth - chartPadding * 2
  const innerHeight = chartHeight - chartPadding * 2
  return data.map((item, index) => {
    const x = chartPadding + (index * innerWidth) / Math.max(data.length - 1, 1)
    const ratio = range === 0 ? 0.5 : (Number(item.calls ?? 0) - min) / range
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

onMounted(loadData)
</script>

<template>
  <NeuroAgentPageShell class="ai-prototype" :show-hero="false">
    <template #title>AI 能力中心</template>
    <template #subtitle>平台调用、成本、成功率、租户排行和健康检查。</template>
    <template #actions>
      <el-button :icon="Refresh" :loading="loading" @click="loadData">刷新</el-button>
    </template>

    <div class="ai-hero-card">
      <div>
        <span class="ai-tag">AI Gateway 正常运行</span>
        <h2>统一管理模型、供应商、基础路由、场景策略、用量与价格</h2>
        <p>业务中心只传租户、应用和 AI 场景；平台完成鉴权、配额校验、模型路由、降级、用量沉淀，配置操作写入底座操作日志。</p>
      </div>
      <div class="ai-actions">
        <el-button :icon="Refresh" :loading="loading" @click="loadData">刷新</el-button>
        <el-button type="primary">新增接入</el-button>
        <el-button>查看用量明细</el-button>
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
            <h3>调用趋势</h3>
            <p class="ai-card__description">近 7 天调用量与成本走势</p>
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
                <text :x="point.x" :y="Math.max(18, point.y - 10)" text-anchor="middle" class="ai-chart-value">{{ numberText(point.item.calls) }}</text>
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
            <span>{{ text(item.model_type) }}</span>
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
        <el-table :data="overview?.tenant_ranking ?? []" border>
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
            <p class="ai-card__description">供应商、密钥、限流和底座操作日志写入状态</p>
          </div>
        </header>
        <div class="ai-card__body ai-health-list">
          <div v-for="item in overview?.health_checks ?? []" :key="item.name">
            <el-tag :type="statusType(item.status)">{{ item.status }}</el-tag>
            <strong> {{ item.name }}</strong>
            <span class="ai-muted"> {{ item.message }}</span>
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
          <el-table :data="overview?.core_base_routes ?? []" border>
            <el-table-column label="基础路由" min-width="180"><template #default="{ row }"><strong>{{ text(row.route_name) }}</strong></template></el-table-column>
            <el-table-column label="能力" min-width="140"><template #default="{ row }">{{ text(row.capability_code) }}</template></el-table-column>
            <el-table-column label="策略" width="130"><template #default="{ row }"><el-tag>{{ text(row.strategy) }}</el-tag></template></el-table-column>
            <el-table-column label="状态" width="100"><template #default="{ row }"><el-tag :type="statusType(row.status)">{{ text(row.status) }}</el-tag></template></el-table-column>
          </el-table>
        </div>
      </section>
    </div>
  </NeuroAgentPageShell>
</template>
