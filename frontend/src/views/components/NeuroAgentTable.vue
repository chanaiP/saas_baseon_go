<template>
  <div class="neuro-table">
    <div class="table-header">
      <h2>{{ title }}</h2>
      <div class="table-controls">
        <button class="refresh-btn" @click="refreshData">
          <span class="refresh-icon">🔄</span>
          <span class="refresh-text">刷新</span>
        </button>
        <button class="realtime-btn" :class="{ active: realTimeUpdates }" @click="toggleRealTimeUpdates">
          <span class="realtime-icon">⚡</span>
          <span class="realtime-text">{{ realTimeUpdates ? '实时更新中' : '实时更新' }}</span>
          <div class="realtime-pulse" v-if="realTimeUpdates"></div>
        </button>
        <button class="ai-analyze-btn" @click="toggleAIAnalysis">
          <span class="ai-icon">🧠</span>
          <span class="ai-text">{{ aiAnalysisEnabled ? 'AI分析中' : 'AI分析' }}</span>
        </button>
      </div>
    </div>

    <div class="table-container">
      <table class="data-table">
        <thead>
          <tr>
            <th v-for="column in columns" :key="column.key" :class="column.className">
              <div class="column-header">
                <span class="column-icon">{{ column.icon }}</span>
                <span class="column-title">{{ column.title }}</span>
                <button v-if="column.sortable" class="sort-btn" @click="sortBy(column.key)">
                  <span class="sort-icon">{{ getSortIcon(column.key) }}</span>
                </button>
              </div>
            </th>
            <th class="actions-column">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in paginatedData" :key="item.id" :class="{ selected: selectedItems.includes(item.id) }">
            <td v-for="column in columns" :key="column.key" :data-label="column.title">
              <div class="cell-content">
                <span v-if="column.type === 'status'" class="status-badge" :class="getStatusClass(item[column.key])">
                  {{ getStatusText(item[column.key]) }}
                </span>
                <span v-else-if="column.type === 'progress'">
                  <div class="progress-bar">
                    <div class="progress-fill" :style="{ width: item[column.key] + '%' }"></div>
                    <span class="progress-text">{{ item[column.key] }}%</span>
                  </div>
                </span>
                <span v-else-if="column.type === 'tags'">
                  <span v-for="tag in item[column.key]" :key="tag" class="tag">{{ tag }}</span>
                </span>
                <span v-else-if="column.type === 'date'">
                  {{ formatDate(item[column.key]) }}
                </span>
                <span v-else>
                  {{ item[column.key] }}
                </span>
              </div>
            </td>
            <td class="actions-cell" data-label="操作">
              <div class="action-buttons">
                <button class="action-btn view" @click="handleView(item)" title="查看">
                  <span class="action-icon">👁️</span>
                </button>
                <button class="action-btn edit" @click="handleEdit(item)" title="编辑">
                  <span class="action-icon">✏️</span>
                </button>
                <button class="action-btn delete" @click="handleDelete(item)" title="删除">
                  <span class="action-icon">🗑️</span>
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 分页控件 -->
    <div class="pagination">
      <div class="pagination-info">
        显示 {{ startIndex + 1 }}-{{ endIndex }} 条，共 {{ filteredData.length }} 条
      </div>
      <div class="pagination-controls">
        <button class="page-btn" :disabled="currentPage === 1" @click="goToPage(1)">
          <span class="page-icon">⏮️</span>
        </button>
        <button class="page-btn" :disabled="currentPage === 1" @click="goToPage(currentPage - 1)">
          <span class="page-icon">◀️</span>
        </button>
        <span class="page-numbers">
          <button
            v-for="page in visiblePages"
            :key="page"
            class="page-number"
            :class="{ active: page === currentPage }"
            @click="goToPage(page)"
          >
            {{ page }}
          </button>
        </span>
        <button class="page-btn" :disabled="currentPage === totalPages" @click="goToPage(currentPage + 1)">
          <span class="page-icon">▶️</span>
        </button>
        <button class="page-btn" :disabled="currentPage === totalPages" @click="goToPage(totalPages)">
          <span class="page-icon">⏭️</span>
        </button>
      </div>
      <div class="page-size-selector">
        <span class="page-size-label">每页显示:</span>
        <select v-model="pageSize" @change="handlePageSizeChange">
          <option value="10">10</option>
          <option value="20">20</option>
          <option value="50">50</option>
          <option value="100">100</option>
        </select>
      </div>
    </div>

    <!-- AI分析面板 -->
    <div v-if="aiAnalysisEnabled && showAIPanel" class="ai-analysis-panel">
      <div class="ai-panel-header">
        <span class="ai-icon">🧠</span>
        <span class="ai-title">AI数据分析</span>
        <button class="close-ai-panel" @click="closeAIPanel">✕</button>
      </div>
      <div class="ai-panel-content">
        <div class="ai-insight">
          <h3>数据洞察</h3>
          <p>• 共 {{ filteredData.length }} 条数据</p>
          <p>• 平均进度: {{ calculateAverageProgress() }}%</p>
          <p>• 状态分布: {{ getStatusDistribution() }}</p>
          <p>• 实时更新: {{ realTimeUpdates ? '启用' : '禁用' }}</p>
        </div>
        <div class="ai-recommendation">
          <h3>AI预测与建议</h3>
          <div v-if="Object.keys(aiPredictions).length > 0">
            <p>• 预测完成率: {{ calculatePredictedCompletion() }}%</p>
            <p>• 高风险项目: {{ countHighRiskProjects() }} 个</p>
            <p>• 预计完成时间: {{ getEstimatedCompletion() }}</p>
          </div>
          <div v-else>
            <p>• 建议优先处理进度低于30%的项目</p>
            <p>• 高优先级项目需要关注</p>
            <p>• 考虑批量更新状态为"进行中"的项目</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'

// Props
interface Props {
  title?: string
  columns: TableColumn[]
  data: any[]
  pageSize?: number
}

const props = withDefaults(defineProps<Props>(), {
  title: '智能数据表',
  pageSize: 10
})

// Emits
const emit = defineEmits<{
  view: [item: any]
  edit: [item: any]
  delete: [item: any]
  refresh: []
  sort: [key: string, direction: 'asc' | 'desc']
}>()

// 类型定义
interface TableColumn {
  key: string
  title: string
  icon: string
  type?: 'text' | 'status' | 'progress' | 'tags' | 'date'
  sortable?: boolean
  className?: string
}

// 响应式数据
const currentPage = ref(1)
const pageSize = ref(props.pageSize)
const sortKey = ref<string>('')
const sortDirection = ref<'asc' | 'desc'>('asc')
const selectedItems = ref<string[]>([])
const aiAnalysisEnabled = ref(false)
const showAIPanel = ref(false)
const realTimeUpdates = ref(false)
const aiPredictions = ref<Record<string, any>>({})
const dataStream = ref<any[]>([]) // 实时数据流

// 计算属性
const filteredData = computed(() => {
  let data = [...props.data]

  // 排序
  if (sortKey.value) {
    data.sort((a, b) => {
      const aVal = a[sortKey.value]
      const bVal = b[sortKey.value]

      if (sortDirection.value === 'asc') {
        return aVal > bVal ? 1 : -1
      } else {
        return aVal < bVal ? 1 : -1
      }
    })
  }

  return data
})

const totalPages = computed(() => Math.ceil(filteredData.value.length / pageSize.value))
const startIndex = computed(() => (currentPage.value - 1) * pageSize.value)
const endIndex = computed(() => Math.min(startIndex.value + pageSize.value, filteredData.value.length))
const paginatedData = computed(() => filteredData.value.slice(startIndex.value, endIndex.value))

const visiblePages = computed(() => {
  const pages: number[] = []
  const maxVisible = 5
  let start = Math.max(1, currentPage.value - Math.floor(maxVisible / 2))
  let end = Math.min(totalPages.value, start + maxVisible - 1)

  if (end - start + 1 < maxVisible) {
    start = Math.max(1, end - maxVisible + 1)
  }

  for (let i = start; i <= end; i++) {
    pages.push(i)
  }

  return pages
})

// 方法
const sortBy = (key: string) => {
  if (sortKey.value === key) {
    sortDirection.value = sortDirection.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortKey.value = key
    sortDirection.value = 'asc'
  }

  emit('sort', key, sortDirection.value)
  currentPage.value = 1
}

const getSortIcon = (key: string) => {
  if (sortKey.value !== key) return '↕️'
  return sortDirection.value === 'asc' ? '⬆️' : '⬇️'
}

const getStatusClass = (status: string) => {
  const statusClasses: Record<string, string> = {
    active: 'status-active',
    pending: 'status-pending',
    paused: 'status-paused',
    completed: 'status-completed'
  }
  return statusClasses[status] || 'status-default'
}

const getStatusText = (status: string) => {
  const statusTexts: Record<string, string> = {
    active: '进行中',
    pending: '待处理',
    paused: '已暂停',
    completed: '已完成'
  }
  return statusTexts[status] || status
}

const formatDate = (date: Date | string) => {
  if (!date) return ''
  const d = new Date(date)
  return d.toLocaleDateString('zh-CN')
}

const handleView = (item: any) => {
  emit('view', item)
}

const handleEdit = (item: any) => {
  emit('edit', item)
}

const handleDelete = (item: any) => {
  if (confirm(`确定归档 "${item.name || item.id}" 吗？归档后默认不再出现在业务列表中。`)) {
    emit('delete', item)
  }
}

const refreshData = () => {
  emit('refresh')
  showOperationStatus('正在刷新数据...', 'info')
}

const toggleAIAnalysis = () => {
  aiAnalysisEnabled.value = !aiAnalysisEnabled.value
  if (aiAnalysisEnabled.value) {
    showAIPanel.value = true
    generateAIPredictions()
    showOperationStatus('AI分析已开启', 'success')
  } else {
    showAIPanel.value = false
    aiPredictions.value = {}
    showOperationStatus('AI分析已关闭', 'info')
  }
}

const toggleRealTimeUpdates = () => {
  if (realTimeUpdates.value) {
    stopRealTimeUpdates()
    showOperationStatus('实时更新已停止', 'info')
  } else {
    startRealTimeUpdates()
    showOperationStatus('实时更新已开启', 'success')
  }
}

const closeAIPanel = () => {
  showAIPanel.value = false
}

const goToPage = (page: number) => {
  if (page >= 1 && page <= totalPages.value) {
    currentPage.value = page
  }
}

const handlePageSizeChange = () => {
  currentPage.value = 1
}

const calculateAverageProgress = () => {
  const progressColumn = props.columns.find(col => col.type === 'progress')
  if (!progressColumn) return 0

  const progressData = filteredData.value
    .map(item => item[progressColumn.key])
    .filter(val => typeof val === 'number')

  if (progressData.length === 0) return 0

  const sum = progressData.reduce((a, b) => a + b, 0)
  return Math.round(sum / progressData.length)
}

const getStatusDistribution = () => {
  const statusColumn = props.columns.find(col => col.type === 'status')
  if (!statusColumn) return '无状态数据'

  const statusCounts: Record<string, number> = {}
  filteredData.value.forEach(item => {
    const status = item[statusColumn.key]
    statusCounts[status] = (statusCounts[status] || 0) + 1
  })

  return Object.entries(statusCounts)
    .map(([status, count]) => `${getStatusText(status)}: ${count}`)
    .join(', ')
}

const calculatePredictedCompletion = () => {
  if (Object.keys(aiPredictions.value).length === 0) return 0

  let totalPredicted = 0
  let count = 0

  Object.values(aiPredictions.value).forEach(prediction => {
    if (prediction.predictedProgress !== undefined) {
      totalPredicted += prediction.predictedProgress
      count++
    }
  })

  return count > 0 ? Math.round(totalPredicted / count) : 0
}

const countHighRiskProjects = () => {
  return Object.values(aiPredictions.value).filter(prediction =>
    prediction.riskLevel === 'high'
  ).length
}

const getEstimatedCompletion = () => {
  const predictions = Object.values(aiPredictions.value)
  if (predictions.length === 0) return '无预测数据'

  const earliestDate = predictions.reduce((earliest, prediction) => {
    if (!prediction.estimatedCompletion) return earliest
    const date = new Date(prediction.estimatedCompletion)
    return !earliest || date < earliest ? date : earliest
  }, null as Date | null)

  if (!earliestDate) return '无法预测'

  const now = new Date()
  const diffDays = Math.ceil((earliestDate.getTime() - now.getTime()) / (1000 * 60 * 60 * 24))

  if (diffDays <= 0) return '即将完成'
  if (diffDays === 1) return '1天内'
  if (diffDays <= 7) return `${diffDays}天内`
  if (diffDays <= 30) return `${Math.ceil(diffDays / 7)}周内`
  return `${Math.ceil(diffDays / 30)}月内`
}

const showOperationStatus = (message: string, type: 'success' | 'error' | 'info' | 'warning') => {
  console.log(`[${type.toUpperCase()}] ${message}`)
  // 这里可以添加更复杂的状态显示逻辑
}

// AI预测功能
const generateAIPredictions = () => {
  if (!aiAnalysisEnabled.value) return

  const predictions: Record<string, any> = {}

  filteredData.value.forEach(item => {
    if (item.progress !== undefined) {
      // 模拟AI预测
      const currentProgress = item.progress
      let predictedProgress = currentProgress
      let predictionConfidence = 0.7
      let riskLevel = 'low'

      if (currentProgress < 30) {
        predictedProgress = currentProgress + Math.random() * 20
        riskLevel = 'high'
      } else if (currentProgress < 70) {
        predictedProgress = currentProgress + Math.random() * 15
        riskLevel = 'medium'
      } else {
        predictedProgress = currentProgress + Math.random() * 10
        riskLevel = 'low'
      }

      predictions[item.id] = {
        predictedProgress: Math.min(100, Math.round(predictedProgress)),
        confidence: predictionConfidence,
        riskLevel,
        estimatedCompletion: new Date(Date.now() + (100 - currentProgress) * 24 * 60 * 60 * 1000),
        recommendations: getAIRecommendations(item)
      }
    }
  })

  aiPredictions.value = predictions
}

const getAIRecommendations = (item: any): string[] => {
  const recommendations: string[] = []

  if (item.progress < 30) {
    recommendations.push('需要立即关注，进度滞后')
    recommendations.push('建议分配更多资源')
  } else if (item.progress < 70) {
    recommendations.push('进度正常，继续保持')
    recommendations.push('建议定期检查里程碑')
  } else {
    recommendations.push('接近完成，准备验收')
    recommendations.push('建议开始文档整理')
  }

  if (item.status === 'pending') {
    recommendations.push('项目尚未启动，需要尽快开始')
  } else if (item.status === 'paused') {
    recommendations.push('项目已暂停，需要重新评估')
  }

  return recommendations.slice(0, 3) // 最多返回3条建议
}

// 实时数据更新
const startRealTimeUpdates = () => {
  if (realTimeUpdates.value) return

  realTimeUpdates.value = true
  dataStream.value = [...props.data]

  // 模拟实时数据更新
  const updateInterval = setInterval(() => {
    if (!realTimeUpdates.value) {
      clearInterval(updateInterval)
      return
    }

    // 随机更新一些数据
    const updatedData = [...dataStream.value]
    const randomIndex = Math.floor(Math.random() * updatedData.length)

    if (updatedData[randomIndex] && updatedData[randomIndex].progress !== undefined) {
      const change = Math.random() > 0.5 ? 1 : -1
      const newProgress = Math.max(0, Math.min(100, updatedData[randomIndex].progress + change))
      updatedData[randomIndex].progress = newProgress

      // 更新状态
      if (newProgress === 100) {
        updatedData[randomIndex].status = 'completed'
      } else if (newProgress > 0 && updatedData[randomIndex].status === 'pending') {
        updatedData[randomIndex].status = 'active'
      }

      dataStream.value = updatedData

      // 重新生成AI预测
      if (aiAnalysisEnabled.value) {
        generateAIPredictions()
      }
    }
  }, 5000) // 每5秒更新一次

  // 清理函数
  onUnmounted(() => {
    realTimeUpdates.value = false
    clearInterval(updateInterval)
  })
}

const stopRealTimeUpdates = () => {
  realTimeUpdates.value = false
}

// 监听数据变化
watch(() => props.data, () => {
  currentPage.value = 1
  dataStream.value = [...props.data]
}, { deep: true })

// 监听AI分析状态
watch(aiAnalysisEnabled, (enabled) => {
  if (enabled) {
    generateAIPredictions()
    showOperationStatus('AI分析已启用，正在生成预测...', 'info')
  } else {
    aiPredictions.value = {}
    showOperationStatus('AI分析已禁用', 'info')
  }
})

// 初始化
onMounted(() => {
  dataStream.value = [...props.data]
  if (aiAnalysisEnabled.value) {
    generateAIPredictions()
  }
})
</script>

<style scoped>
@import '@/styles/theme/neuro-theme.css';
.neuro-table {
  background: linear-gradient(135deg, rgba(26, 26, 46, 0.95), rgba(40, 40, 60, 0.95));
  border: 2px solid var(--neuro-primary-30);
  border-radius: var(--neuro-radius-xl);
  padding: var(--neuro-spacing-lg);
  box-shadow: 0 8px 32px rgba(0, 245, 212, 0.15);
  backdrop-filter: blur(10px);
}

.table-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.table-header h2 {
  color: var(--neuro-primary);
  margin: 0;
  font-size: 20px;
}

.table-controls {
  display: flex;
  gap: 12px;
}

.refresh-btn,
.ai-analyze-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  border: none;
  border-radius: var(--neuro-radius-lg);
  font-weight: bold;
  cursor: pointer;
  transition: all var(--neuro-transition-normal);
}

.refresh-btn {
  background: var(--neuro-primary-10);
  border: 1px solid var(--neuro-primary-30);
  color: var(--neuro-primary);
}

.refresh-btn:hover {
  background: var(--neuro-primary-20);
  transform: translateY(-2px);
}

.realtime-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background: rgba(255, 193, 7, 0.1);
  border: 1px solid rgba(255, 193, 7, 0.3);
  border-radius: var(--neuro-radius-lg);
  color: #ffc107;
  font-weight: bold;
  cursor: pointer;
  transition: all var(--neuro-transition-normal);
  position: relative;
  overflow: hidden;
}

.realtime-btn:hover {
  background: rgba(255, 193, 7, 0.2);
  transform: translateY(-2px);
}

.realtime-btn.active {
  background: rgba(255, 193, 7, 0.3);
  border-color: #ffc107;
  box-shadow: 0 0 20px rgba(255, 193, 7, 0.3);
}

.realtime-pulse {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: radial-gradient(circle at center, rgba(255, 193, 7, 0.3), transparent 70%);
  animation: pulse 1s infinite;
}

.ai-analyze-btn {
  background: var(--neuro-info-10);
  border: 1px solid var(--neuro-info-30);
  color: #0096ff;
}

.ai-analyze-btn:hover {
  background: rgba(0, 150, 255, 0.2);
  transform: translateY(-2px);
}

.table-container {
  overflow-x: auto;
  margin-bottom: 20px;
}

.data-table {
  width: 100%;
  border-collapse: separate;
  border-spacing: 0;
}

.data-table th {
  background: var(--neuro-primary-10);
  color: var(--neuro-primary);
  font-weight: bold;
  padding: var(--neuro-spacing-md);
  text-align: left;
  border-bottom: 2px solid var(--neuro-primary-30);
}

.column-header {
  display: flex;
  align-items: center;
  gap: 8px;
}

.column-icon {
  font-size: 16px;
}

.column-title {
  flex: 1;
}

.sort-btn {
  background: transparent;
  border: none;
  color: rgba(0, 245, 212, 0.6);
  cursor: pointer;
  padding: var(--neuro-spacing-xs);
}

.sort-btn:hover {
  color: var(--neuro-primary);
}

.data-table td {
  padding: 12px 16px;
  border-bottom: 1px solid var(--neuro-surface-90);
  color: var(--neuro-text);
}

.data-table tr:hover {
  background: rgba(0, 245, 212, 0.05);
}

.data-table tr.selected {
  background: var(--neuro-primary-10);
}

.cell-content {
  display: flex;
  align-items: center;
  gap: 8px;
}

.status-badge {
  padding: 4px 12px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: bold;
}

.status-active {
  background: var(--neuro-success-20);
  color: var(--neuro-success);
  border: 1px solid var(--neuro-success-30);
}

.status-pending {
  background: var(--neuro-warning-20);
  color: var(--neuro-warning);
  border: 1px solid var(--neuro-warning-30);
}

.status-paused {
  background: var(--neuro-error-20);
  color: var(--neuro-error);
  border: 1px solid var(--neuro-error-30);
}

.status-completed {
  background: var(--neuro-info-20);
  color: var(--neuro-info);
  border: 1px solid var(--neuro-info-30);
}

.progress-bar {
  position: relative;
  width: 100px;
  height: 20px;
  background: var(--neuro-surface-90);
  border-radius: 10px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: linear-gradient(90deg, var(--neuro-primary), var(--neuro-secondary));
  border-radius: 10px;
  transition: width 0.3s ease;
}

.progress-text {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--neuro-text);
  font-size: 11px;
  font-weight: bold;
}

.tag {
  display: inline-block;
  padding: 2px 8px;
  margin: 2px;
  background: var(--neuro-primary-10);
  border: 1px solid var(--neuro-primary-30);
  border-radius: var(--neuro-radius-lg);
  font-size: 11px;
  color: var(--neuro-primary);
}

.actions-cell {
  width: 120px;
}

.action-buttons {
  display: flex;
  gap: 8px;
}

.action-btn {
  background: transparent;
  border: none;
  padding: 6px;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.action-btn.view {
  color: var(--neuro-info);
}

.action-btn.edit {
  color: var(--neuro-warning);
}

.action-btn.delete {
  color: var(--neuro-error);
}

.action-btn:hover {
  transform: scale(1.1);
  background: var(--neuro-surface-90);
}

.pagination {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 20px;
  padding-top: 20px;
  border-top: 1px solid var(--neuro-surface-90);
}

.pagination-info {
  color: var(--neuro-text-secondary);
  font-size: 14px;
}

.pagination-controls {
  display: flex;
  align-items: center;
  gap: 8px;
}

.page-btn {
  background: var(--neuro-surface-90);
  border: 1px solid var(--neuro-surface-90);
  color: var(--neuro-text);
  padding: 6px 12px;
  border-radius: var(--neuro-radius-md);
  cursor: pointer;
  transition: all 0.2s ease;
}

.page-btn:hover:not(:disabled) {
  background: var(--neuro-primary-20);
  border-color: var(--neuro-primary-30);
}

.page-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.page-numbers {
  display: flex;
  gap: 4px;
}

.page-number {
  min-width: 32px;
  height: 32px;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid var(--neuro-surface-90);
  color: var(--neuro-text);
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.page-number:hover {
  background: var(--neuro-primary-10);
  border-color: var(--neuro-primary-30);
}

.page-number.active {
  background: var(--neuro-primary-20);
  border-color: var(--neuro-primary);
  color: var(--neuro-primary);
  font-weight: bold;
}

.page-size-selector {
  display: flex;
  align-items: center;
  gap: 8px;
}

.page-size-label {
  color: var(--neuro-text-secondary);
  font-size: 14px;
}

.page-size-selector select {
  background: var(--neuro-surface-90);
  border: 1px solid var(--neuro-surface-90);
  color: var(--neuro-text);
  padding: 6px 12px;
  border-radius: 6px;
  outline: none;
}

.page-size-selector select:focus {
  border-color: var(--neuro-primary);
}

.ai-analysis-panel {
  margin-top: 20px;
  padding: 20px;
  background: var(--neuro-info-10);
  border: 1px solid var(--neuro-info-30);
  border-radius: var(--neuro-radius-lg);
  animation: slideDown 0.3s ease;
}

@keyframes slideDown {
  from { opacity: 0; transform: translateY(-10px); }
  to { opacity: 1; transform: translateY(0); }
}

.ai-panel-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 16px;
}

.ai-icon {
  font-size: 20px;
  color: #0096ff;
}

.ai-title {
  flex: 1;
  color: #0096ff;
  font-weight: bold;
  font-size: 16px;
}

.close-ai-panel {
  background: transparent;
  border: none;
  color: rgba(255, 255, 255, 0.5);
  cursor: pointer;
  padding: var(--neuro-spacing-xs);
}

.close-ai-panel:hover {
  color: #ff6b6b;
}

.ai-panel-content {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
}

.ai-insight,
.ai-recommendation {
  padding: var(--neuro-spacing-md);
  background: rgba(0, 0, 0, 0.2);
  border-radius: var(--neuro-radius-md);
}

.ai-insight h3,
.ai-recommendation h3 {
  color: var(--neuro-primary);
  margin: 0 0 12px 0;
  font-size: 16px;
}

.ai-insight p,
.ai-recommendation p {
  color: rgba(255, 255, 255, 0.8);
  margin: 8px 0;
  font-size: 14px;
  line-height: 1.4;
}

/* 滚动条样式 */
.table-container::-webkit-scrollbar {
  height: 8px;
}

.table-container::-webkit-scrollbar-track {
  background: rgba(255, 255, 255, 0.05);
  border-radius: var(--neuro-radius-sm);
}

.table-container::-webkit-scrollbar-thumb {
  background: var(--neuro-primary-30);
  border-radius: var(--neuro-radius-sm);
}

.table-container::-webkit-scrollbar-thumb:hover {
  background: rgba(0, 245, 212, 0.5);
}

/* 响应式设计 */
@media (max-width: 768px) {
  .neuro-table {
    padding: var(--neuro-spacing-md);
  }

  .table-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 16px;
  }

  .table-controls {
    width: 100%;
    justify-content: space-between;
  }

  .data-table th,
  .data-table td {
    padding: 10px 12px;
    font-size: 14px;
  }

  .column-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 4px;
  }

  .column-icon {
    font-size: 14px;
  }

  .column-title {
    font-size: 12px;
  }

  .status-badge {
    padding: 3px 8px;
    font-size: 11px;
  }

  .progress-bar {
    width: 80px;
  }

  .action-buttons {
    flex-direction: column;
    gap: 4px;
  }

  .action-btn {
    padding: var(--neuro-spacing-xs);
  }

  .pagination {
    flex-direction: column;
    gap: 16px;
    align-items: stretch;
  }

  .pagination-controls {
    order: 1;
  }

  .pagination-info {
    order: 2;
    text-align: center;
  }

  .page-size-selector {
    order: 3;
    justify-content: center;
  }

  .ai-panel-content {
    grid-template-columns: 1fr;
    gap: 16px;
  }
}

@media (max-width: 480px) {
  .neuro-table {
    padding: 12px;
  }

  .table-controls {
    flex-direction: column;
    gap: 8px;
  }

  .refresh-btn,
  .ai-analyze-btn {
    width: 100%;
    justify-content: center;
  }

  .data-table {
    display: block;
  }

  .data-table thead {
    display: none;
  }

  .data-table tbody {
    display: block;
  }

  .data-table tr {
    display: block;
    margin-bottom: 16px;
    background: rgba(255, 255, 255, 0.05);
    border-radius: var(--neuro-radius-lg);
    padding: 12px;
  }

  .data-table td {
    display: block;
    border: none;
    padding: 8px 0;
    border-bottom: 1px solid var(--neuro-surface-90);
  }

  .data-table td:last-child {
    border-bottom: none;
  }

  .data-table td::before {
    content: attr(data-label);
    display: block;
    font-weight: bold;
    color: var(--neuro-primary);
    font-size: 12px;
    margin-bottom: 4px;
  }

  .cell-content {
    justify-content: space-between;
  }

  .page-numbers {
    display: none;
  }

  .ai-analysis-panel {
    padding: var(--neuro-spacing-md);
  }
}
</style>
