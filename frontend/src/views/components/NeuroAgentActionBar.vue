<template>
  <div class="neuro-action-bar" :class="{ 'has-selection': hasSelection, 'ai-enhanced': aiEnhanced }">
    <!-- 选择状态指示器 -->
    <div class="selection-indicator" v-if="hasSelection">
      <div class="selection-pulse"></div>
      <span class="selection-icon">⚡</span>
      <span class="selection-count">{{ selectedCount }} 项已选择</span>
      <button class="clear-selection" @click="clearSelection" title="清除选择">
        <span class="clear-icon">✕</span>
      </button>
    </div>

    <!-- 操作按钮组 -->
    <div class="action-buttons">
      <!-- 主要操作按钮 -->
      <div class="primary-actions">
        <button
          v-for="action in primaryActions"
          :key="action.id"
          class="action-btn primary"
          :class="{ disabled: action.requiresSelection && !hasSelection }"
          @click="handleAction(action)"
          :title="action.description"
        >
          <span class="action-icon">{{ action.icon }}</span>
          <span class="action-label">{{ action.label }}</span>
          <div class="action-glow" v-if="action.highlight"></div>
        </button>
      </div>

      <!-- 次要操作按钮（下拉菜单） -->
      <div class="secondary-actions">
        <div class="dropdown-container">
          <button class="more-actions-btn" @click="toggleMoreActions">
            <span class="more-icon">⋯</span>
            <span class="more-label">更多操作</span>
            <span class="dropdown-arrow" :class="{ open: showMoreActions }">▼</span>
          </button>

          <div v-if="showMoreActions" class="dropdown-menu">
            <div
              v-for="action in secondaryActions"
              :key="action.id"
              class="dropdown-item"
              :class="{ disabled: action.requiresSelection && !hasSelection }"
              @click="handleAction(action)"
            >
              <span class="dropdown-icon">{{ action.icon }}</span>
              <span class="dropdown-label">{{ action.label }}</span>
              <span class="dropdown-desc" v-if="action.description">{{ action.description }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- AI增强操作 -->
      <div class="ai-actions" v-if="aiEnabled">
        <button class="ai-action-btn" @click="toggleAIEnhance">
          <span class="ai-icon">🧠</span>
          <span class="ai-label">{{ aiEnhanced ? 'AI增强中' : 'AI增强' }}</span>
          <div class="ai-pulse" v-if="aiEnhanced"></div>
        </button>

        <div v-if="aiEnhanced" class="ai-suggestions">
          <div class="ai-suggestion" v-for="suggestion in aiSuggestions" :key="suggestion.id">
            <span class="suggestion-icon">{{ suggestion.icon }}</span>
            <span class="suggestion-text">{{ suggestion.text }}</span>
            <span class="suggestion-confidence" v-if="suggestion.confidence">
              {{ (suggestion.confidence * 100).toFixed(0) }}%
            </span>
            <button class="suggestion-action" @click="handleAISuggestion(suggestion)">
              <span class="action-icon">⚡</span>
            </button>
          </div>
        </div>

        <!-- AI批量模式控制 -->
        <div v-if="aiEnhanced" class="ai-batch-controls">
          <button class="ai-batch-toggle" @click="toggleAIBatchMode" :class="{ active: aiBatchMode }">
            <span class="batch-toggle-icon">{{ aiBatchMode ? '🤖' : '⚡' }}</span>
            <span class="batch-toggle-text">{{ aiBatchMode ? '批量模式' : 'AI批量' }}</span>
          </button>

          <button class="ai-history-btn" @click="viewAIActionHistory" title="查看AI操作历史">
            <span class="history-icon">📜</span>
          </button>
        </div>
      </div>
    </div>

    <!-- 批量操作面板 -->
    <div v-if="hasSelection && showBatchPanel" class="batch-panel">
      <div class="batch-header">
        <span class="batch-icon">📦</span>
        <span class="batch-title">批量操作</span>
        <button class="close-batch" @click="closeBatchPanel">✕</button>
      </div>

      <div class="batch-actions">
        <button
          v-for="action in batchActions"
          :key="action.id"
          class="batch-btn"
          @click="handleBatchAction(action)"
        >
          <span class="batch-icon">{{ action.icon }}</span>
          <span class="batch-label">{{ action.label }}</span>
        </button>
      </div>
    </div>

    <!-- AI批量操作面板 -->
    <div v-if="aiBatchMode && aiEnhanced" class="ai-batch-panel">
      <div class="ai-batch-header">
        <span class="ai-batch-icon">🤖</span>
        <span class="ai-batch-title">AI批量操作选择</span>
        <span class="ai-batch-count">{{ selectedAIActions.length }} 个操作已选择</span>
      </div>

      <div class="ai-batch-actions">
        <div class="ai-action-selection">
          <div class="selection-group">
            <div class="group-title">主要AI操作</div>
            <div class="action-checkboxes">
              <label v-for="action in primaryActions.filter(a => a.aiEnhanced)" :key="action.id" class="action-checkbox">
                <input type="checkbox" :value="action.id" v-model="selectedAIActions">
                <span class="checkbox-icon">{{ action.icon }}</span>
                <span class="checkbox-label">{{ action.label }}</span>
                <span class="checkbox-confidence" v-if="action.aiConfidence">
                  {{ (action.aiConfidence * 100).toFixed(0) }}%
                </span>
              </label>
            </div>
          </div>

          <div class="selection-group">
            <div class="group-title">次要AI操作</div>
            <div class="action-checkboxes">
              <label v-for="action in secondaryActions.filter(a => a.aiEnhanced)" :key="action.id" class="action-checkbox">
                <input type="checkbox" :value="action.id" v-model="selectedAIActions">
                <span class="checkbox-icon">{{ action.icon }}</span>
                <span class="checkbox-label">{{ action.label }}</span>
                <span class="checkbox-confidence" v-if="action.aiConfidence">
                  {{ (action.aiConfidence * 100).toFixed(0) }}%
                </span>
              </label>
            </div>
          </div>
        </div>

        <div class="ai-batch-execute">
          <button class="execute-btn" @click="executeSelectedAIActions" :disabled="!aiBatchActionsSelected">
            <span class="execute-icon">🚀</span>
            <span class="execute-text">执行选中AI操作</span>
            <span class="execute-count" v-if="aiBatchActionsSelected">({{ selectedAIActions.length }})</span>
          </button>
        </div>
      </div>
    </div>

    <!-- 操作状态指示器 -->
    <div v-if="operationStatus" class="operation-status">
      <div class="status-indicator" :class="operationStatus.type">
        <span class="status-icon">{{ operationStatus.icon }}</span>
        <span class="status-text">{{ operationStatus.message }}</span>
        <div class="status-progress" v-if="operationStatus.progress !== undefined && !Array.isArray(operationStatus.progress)">
          <div class="progress-bar" :style="{ width: operationStatus.progress + '%' }"></div>
        </div>
        <div v-if="operationStatus.details" class="status-details">
          <ul>
            <li v-for="(detail, index) in operationStatus.details" :key="index">{{ detail }}</li>
          </ul>
        </div>
      </div>
    </div>

    <!-- 操作确认对话框 -->
    <div v-if="showConfirmDialog" class="confirm-dialog-overlay">
      <div class="confirm-dialog">
        <div class="dialog-header">
          <span class="dialog-icon">{{ confirmAction?.icon }}</span>
          <span class="dialog-title">{{ confirmAction?.label }}</span>
        </div>
        <div class="dialog-content">
          <p>{{ confirmMessage }}</p>
          <div class="dialog-actions">
            <button class="dialog-btn cancel" @click="cancelConfirm">取消</button>
            <button class="dialog-btn confirm" @click="executeConfirm">确认</button>
          </div>
        </div>
      </div>
    </div>

    <!-- AI分析结果面板 -->
    <div v-if="showAIAnalysisPanel && aiAnalysisResults" class="ai-analysis-panel-overlay">
      <div class="ai-analysis-panel">
        <div class="analysis-header">
          <span class="analysis-icon">📊</span>
          <span class="analysis-title">AI分析报告</span>
          <button class="close-analysis" @click="closeAIAnalysisPanel">✕</button>
        </div>

        <div class="analysis-content">
          <div class="analysis-summary">
            <div class="summary-item">
              <span class="summary-icon">📋</span>
              <span class="summary-label">分析类型:</span>
              <span class="summary-value">{{ aiAnalysisResults.type }}</span>
            </div>
            <div class="summary-item">
              <span class="summary-icon">📊</span>
              <span class="summary-label">分析项目:</span>
              <span class="summary-value">{{ aiAnalysisResults.items.length }} 个</span>
            </div>
            <div class="summary-item">
              <span class="summary-icon">🕒</span>
              <span class="summary-label">分析时间:</span>
              <span class="summary-value">刚刚</span>
            </div>
          </div>

          <div class="analysis-sections">
            <div class="analysis-section">
              <div class="section-header">
                <span class="section-icon">💡</span>
                <span class="section-title">关键洞察</span>
              </div>
              <div class="section-content">
                <div v-for="(insight, index) in aiAnalysisResults.insights" :key="index" class="insight-item">
                  <span class="insight-icon">⚡</span>
                  <span class="insight-text">{{ insight }}</span>
                </div>
              </div>
            </div>

            <div class="analysis-section">
              <div class="section-header">
                <span class="section-icon">✅</span>
                <span class="section-title">优化建议</span>
              </div>
              <div class="section-content">
                <div v-for="(recommendation, index) in aiAnalysisResults.recommendations" :key="index" class="recommendation-item">
                  <span class="recommendation-icon">🎯</span>
                  <span class="recommendation-text">{{ recommendation }}</span>
                </div>
              </div>
            </div>

            <div class="analysis-section">
              <div class="section-header">
                <span class="section-icon">⚠️</span>
                <span class="section-title">风险提示</span>
              </div>
              <div class="section-content">
                <div v-for="(risk, index) in aiAnalysisResults.risks" :key="index" class="risk-item">
                  <span class="risk-icon">🔴</span>
                  <span class="risk-text">{{ risk }}</span>
                </div>
              </div>
            </div>
          </div>

          <div class="analysis-actions">
            <button class="analysis-action-btn export">
              <span class="action-icon">📤</span>
              <span class="action-text">导出报告</span>
            </button>
            <button class="analysis-action-btn share">
              <span class="action-icon">🔗</span>
              <span class="action-text">分享分析</span>
            </button>
            <button class="analysis-action-btn schedule">
              <span class="action-icon">📅</span>
              <span class="action-text">安排跟进</span>
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'

// Props
interface Props {
  selectedItems?: any[]
  aiEnabled?: boolean
  showBatchPanel?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  selectedItems: () => [],
  aiEnabled: true,
  showBatchPanel: true
})

// Emits
const emit = defineEmits<{
  action: [action: Action, items?: any[]]
  selectionClear: []
  aiToggle: [enabled: boolean]
  batchAction: [action: Action, items: any[]]
}>()

// 类型定义
interface Action {
  id: string
  label: string
  icon: string
  description?: string
  requiresSelection?: boolean
  requiresConfirmation?: boolean
  confirmationMessage?: string
  highlight?: boolean
  type: 'primary' | 'secondary' | 'batch' | 'ai' | 'ai-batch'
  aiEnhanced?: boolean
  aiConfidence?: number
}

interface AISuggestion {
  id: string
  text: string
  icon: string
  action: string
  confidence?: number
  category?: string
  aiGenerated?: boolean
}

interface OperationStatus {
  type: 'loading' | 'success' | 'error' | 'info' | 'warning'
  message: string
  icon: string
  progress?: number
  details?: string[]
}

interface AIAnalysisResult {
  type: string
  items: any[]
  insights: string[]
  recommendations: string[]
  risks: string[]
}

// 响应式数据
const selectedItems = ref<any[]>(props.selectedItems)
const aiEnhanced = ref(false)
const showMoreActions = ref(false)
const showBatchPanel = ref(props.showBatchPanel)
const operationStatus = ref<OperationStatus | null>(null)
const showConfirmDialog = ref(false)
const confirmAction = ref<Action | null>(null)
const confirmMessage = ref('')
const showAIAnalysisPanel = ref(false)
const aiAnalysisResults = ref<AIAnalysisResult | null>(null)
const isAnalyzing = ref(false)
const aiBatchMode = ref(false)
const selectedAIActions = ref<string[]>([])
const aiActionHistory = ref<any[]>([])

// 主要操作
const primaryActions = ref<Action[]>([
  { id: 'create', label: '新建', icon: '➕', description: '创建新项目', type: 'primary', highlight: true },
  { id: 'edit', label: '编辑', icon: '✏️', description: '编辑选中项目', requiresSelection: true, type: 'primary' },
  { id: 'delete', label: '删除', icon: '🗑️', description: '删除选中项目', requiresSelection: true, requiresConfirmation: true, confirmationMessage: '确定要删除选中的项目吗？', type: 'primary' },
  { id: 'export', label: '导出', icon: '📤', description: '导出选中项目', requiresSelection: true, type: 'primary' },
  { id: 'import', label: '导入', icon: '📥', description: '导入项目数据', type: 'primary' },
  { id: 'ai-analyze', label: 'AI分析', icon: '🧠', description: 'AI分析选中项目', requiresSelection: true, type: 'primary', aiEnhanced: true, aiConfidence: 0.92 }
])

// 次要操作
const secondaryActions = ref<Action[]>([
  { id: 'copy', label: '复制', icon: '📋', description: '复制选中项目', requiresSelection: true, type: 'secondary' },
  { id: 'move', label: '移动', icon: '🚚', description: '移动选中项目', requiresSelection: true, type: 'secondary' },
  { id: 'archive', label: '归档', icon: '📦', description: '归档选中项目', requiresSelection: true, type: 'secondary' },
  { id: 'share', label: '分享', icon: '🔗', description: '分享项目链接', requiresSelection: true, type: 'secondary' },
  { id: 'analyze', label: '分析', icon: '📊', description: '分析项目数据', requiresSelection: true, type: 'secondary' },
  { id: 'schedule', label: '排期', icon: '📅', description: '设置项目排期', requiresSelection: true, type: 'secondary' },
  { id: 'ai-optimize', label: 'AI优化', icon: '⚡', description: 'AI优化项目设置', requiresSelection: true, type: 'secondary', aiEnhanced: true, aiConfidence: 0.88 },
  { id: 'ai-predict', label: 'AI预测', icon: '🔮', description: 'AI预测项目趋势', requiresSelection: true, type: 'secondary', aiEnhanced: true, aiConfidence: 0.85 }
])

// 批量操作
const batchActions = ref<Action[]>([
  { id: 'batch-delete', label: '批量删除', icon: '🗑️', description: '删除所有选中项目', type: 'batch' },
  { id: 'batch-export', label: '批量导出', icon: '📤', description: '导出所有选中项目', type: 'batch' },
  { id: 'batch-tag', label: '批量标签', icon: '🏷️', description: '为选中项目添加标签', type: 'batch' },
  { id: 'batch-status', label: '批量状态', icon: '🔄', description: '更新选中项目状态', type: 'batch' },
  { id: 'batch-priority', label: '批量优先级', icon: '🚀', description: '设置选中项目优先级', type: 'batch' },
  { id: 'ai-batch-analyze', label: 'AI批量分析', icon: '🧠', description: 'AI批量分析所有项目', type: 'batch', aiEnhanced: true, aiConfidence: 0.95 },
  { id: 'ai-batch-optimize', label: 'AI批量优化', icon: '⚡', description: 'AI批量优化项目设置', type: 'batch', aiEnhanced: true, aiConfidence: 0.90 }
])

// AI建议
const aiSuggestions = ref<AISuggestion[]>([
  { id: 'ai-analyze', text: 'AI分析项目', icon: '🧠', action: 'analyze', confidence: 0.92, category: '分析', aiGenerated: true },
  { id: 'ai-optimize', text: 'AI优化建议', icon: '⚡', action: 'optimize', confidence: 0.88, category: '优化', aiGenerated: true },
  { id: 'ai-predict', text: 'AI预测风险', icon: '🔮', action: 'predict', confidence: 0.85, category: '预测', aiGenerated: true },
  { id: 'ai-summarize', text: 'AI生成摘要', icon: '📝', action: 'summarize', confidence: 0.95, category: '总结', aiGenerated: true },
  { id: 'ai-cluster', text: 'AI智能分组', icon: '📊', action: 'cluster', confidence: 0.90, category: '分组', aiGenerated: true },
  { id: 'ai-prioritize', text: 'AI智能排序', icon: '🚀', action: 'prioritize', confidence: 0.87, category: '排序', aiGenerated: true }
])

// 计算属性
const hasSelection = computed(() => selectedItems.value.length > 0)
const selectedCount = computed(() => selectedItems.value.length)
const aiActionCount = computed(() => primaryActions.value.filter(a => a.aiEnhanced).length + secondaryActions.value.filter(a => a.aiEnhanced).length)
const hasAIAnalysis = computed(() => aiAnalysisResults.value !== null)
const aiBatchActionsSelected = computed(() => selectedAIActions.value.length > 0)

// 方法
const handleAction = (action: Action) => {
  if (action.requiresSelection && !hasSelection.value) {
    showOperationStatus('error', '请先选择项目', '❌')
    return
  }

  if (action.requiresConfirmation) {
    confirmAction.value = action
    confirmMessage.value = action.confirmationMessage || `确定要执行"${action.label}"操作吗？`
    showConfirmDialog.value = true
    return
  }

  executeAction(action)
}

const executeAction = (action: Action) => {
  // 记录AI操作历史
  if (action.aiEnhanced) {
    aiActionHistory.value.unshift({
      action: action.label,
      timestamp: Date.now(),
      items: selectedItems.value.length,
      confidence: action.aiConfidence
    })

    // 限制历史记录数量
    if (aiActionHistory.value.length > 10) {
      aiActionHistory.value = aiActionHistory.value.slice(0, 10)
    }
  }

  showOperationStatus('loading',
    action.aiEnhanced ? `AI正在${action.label}...` : `正在${action.label}...`,
    action.aiEnhanced ? '🧠' : '⏳',
    action.aiEnhanced ? 0 : undefined
  )

  // 模拟操作延迟
  setTimeout(() => {
    emit('action', action, selectedItems.value)

    if (action.aiEnhanced) {
      showOperationStatus('success',
        `AI${action.label}完成 (置信度: ${(action.aiConfidence! * 100).toFixed(1)}%)`,
        '✅',
        action.aiEnhanced ? ['AI分析完成', '结果已优化', '建议已生成'] : undefined
      )
    } else {
      showOperationStatus('success', `${action.label}操作完成`, '✅')
    }

    // 3秒后清除状态
    setTimeout(() => {
      operationStatus.value = null
    }, 3000)
  }, action.aiEnhanced ? 1200 : 800)

  // AI增强操作的特殊处理
  if (action.aiEnhanced) {
    simulateAIAnalysis(action)
  }
}

const handleBatchAction = (action: Action) => {
  if (action.aiEnhanced) {
    // AI批量操作
    executeAIBatchAction(action)
  } else {
    // 普通批量操作
    emit('batchAction', action, selectedItems.value)
    showOperationStatus('loading', `正在批量${action.label}...`, '⏳')

    // 模拟批量操作
    setTimeout(() => {
      showOperationStatus('success', `批量${action.label}完成`, '✅')
      clearSelection()
    }, 1200)
  }
}

const clearSelection = () => {
  selectedItems.value = []
  emit('selectionClear')
  showOperationStatus('info', '选择已清除', '🗑️')
}

const toggleAIEnhance = () => {
  aiEnhanced.value = !aiEnhanced.value
  emit('aiToggle', aiEnhanced.value)

  if (aiEnhanced.value) {
    showOperationStatus('info', 'AI增强已开启', '🧠', ['智能建议已激活', 'AI分析已就绪', '批量优化可用'])
    // 自动生成AI建议
    generateAISuggestions()
  } else {
    showOperationStatus('info', 'AI增强已关闭', '🔌')
    aiBatchMode.value = false
    selectedAIActions.value = []
  }
}

const toggleMoreActions = () => {
  showMoreActions.value = !showMoreActions.value
}

const closeBatchPanel = () => {
  showBatchPanel.value = false
}

const handleAISuggestion = (suggestion: AISuggestion) => {
  showOperationStatus('loading', `AI正在${suggestion.text}...`, '🧠', 0)

  // 模拟AI处理
  const progressInterval = setInterval(() => {
    if (operationStatus.value?.progress !== undefined) {
      operationStatus.value.progress += 10
      if (operationStatus.value.progress >= 100) {
        clearInterval(progressInterval)
        showOperationStatus('success',
          `AI${suggestion.text}完成 (置信度: ${(suggestion.confidence! * 100).toFixed(1)}%)`,
          '✅',
          [`${suggestion.category}分析完成`, '结果已优化', '建议已生成']
        )
      }
    }
  }, 150)

  // 记录AI操作
  aiActionHistory.value.unshift({
    action: suggestion.text,
    timestamp: Date.now(),
    items: selectedItems.value.length,
    confidence: suggestion.confidence,
    category: suggestion.category
  })
}

const showOperationStatus = (type: OperationStatus['type'], message: string, icon: string, progress?: number | string[]) => {
  if (Array.isArray(progress)) {
    operationStatus.value = { type, message, icon, details: progress }
  } else {
    operationStatus.value = { type, message, icon, progress }
  }
}

// 生成AI建议
const generateAISuggestions = () => {
  if (!aiEnhanced.value) return

  // 根据选择的项目生成个性化建议
  if (hasSelection.value) {
    const count = selectedItems.value.length
    const newSuggestions: AISuggestion[] = []

    if (count > 5) {
      newSuggestions.push({
        id: 'ai-bulk-analyze',
        text: 'AI批量深度分析',
        icon: '🔍',
        action: 'bulk-analyze',
        confidence: 0.94,
        category: '批量分析',
        aiGenerated: true
      })
    }

    if (count > 10) {
      newSuggestions.push({
        id: 'ai-cluster-groups',
        text: 'AI智能分组',
        icon: '📊',
        action: 'cluster-groups',
        confidence: 0.89,
        category: '智能分组',
        aiGenerated: true
      })
    }

    // 添加通用建议
    newSuggestions.push({
      id: 'ai-priority-sort',
      text: 'AI智能优先级排序',
      icon: '🚀',
      action: 'priority-sort',
      confidence: 0.91,
      category: '智能排序',
      aiGenerated: true
    })

    // 合并建议
    aiSuggestions.value = [...newSuggestions, ...aiSuggestions.value.slice(0, 4)]
  }
}

// 模拟AI分析
const simulateAIAnalysis = (action: Action) => {
  if (!operationStatus.value || operationStatus.value.type !== 'loading') return

  const analysisSteps = [
    '🔍 分析项目数据...',
    '🧠 应用机器学习模型...',
    '📊 计算优化方案...',
    '⚡ 生成智能建议...',
    '✅ 完成AI处理...'
  ]

  let step = 0
  const analysisInterval = setInterval(() => {
    if (step < analysisSteps.length) {
      if (operationStatus.value) {
        operationStatus.value.message = analysisSteps[step]
        if (operationStatus.value.progress !== undefined) {
          operationStatus.value.progress = Math.min(100, (step + 1) * 20)
        }
      }
      step++
    } else {
      clearInterval(analysisInterval)
    }
  }, 300)
}

// 执行AI批量操作
const executeAIBatchAction = (action: Action) => {
  showOperationStatus('loading', `AI正在批量${action.label}...`, '🧠', 0)

  // 模拟AI批量处理
  const progressInterval = setInterval(() => {
    if (operationStatus.value?.progress !== undefined) {
      operationStatus.value.progress += 5
      if (operationStatus.value.progress >= 100) {
        clearInterval(progressInterval)

        // 生成分析结果
        const analysisResult: AIAnalysisResult = {
          type: action.label,
          items: selectedItems.value,
          insights: [
            `分析了 ${selectedItems.value.length} 个项目`,
            '检测到3个高优先级项目',
            '发现2个项目存在风险',
            '生成5个优化建议'
          ],
          recommendations: [
            '建议优先处理高风险项目',
            '优化项目资源配置',
            '调整项目时间线',
            '加强团队协作',
            '增加监控频率'
          ],
          risks: [
            '项目A存在延期风险',
            '项目B资源不足',
            '项目C依赖关系复杂'
          ]
        }

        aiAnalysisResults.value = analysisResult
        showAIAnalysisPanel.value = true

        showOperationStatus('success',
          `AI批量${action.label}完成`,
          '✅',
          [`处理了 ${selectedItems.value.length} 个项目`, '生成详细分析报告', '优化建议已就绪']
        )

        emit('batchAction', action, selectedItems.value)
      }
    }
  }, 100)
}

// 切换AI批量模式
const toggleAIBatchMode = () => {
  aiBatchMode.value = !aiBatchMode.value
  if (aiBatchMode.value) {
    showOperationStatus('info', 'AI批量模式已激活', '🤖', ['可选择多个AI操作', '批量执行AI任务', '智能优化工作流'])
  } else {
    selectedAIActions.value = []
    showOperationStatus('info', 'AI批量模式已关闭', '🔌')
  }
}

// 选择AI操作
const toggleAIActionSelection = (actionId: string) => {
  const index = selectedAIActions.value.indexOf(actionId)
  if (index > -1) {
    selectedAIActions.value.splice(index, 1)
  } else {
    selectedAIActions.value.push(actionId)
  }
}

// 执行选中的AI操作
const executeSelectedAIActions = () => {
  if (selectedAIActions.value.length === 0) {
    showOperationStatus('error', '请先选择AI操作', '❌')
    return
  }

  showOperationStatus('loading', `正在执行 ${selectedAIActions.value.length} 个AI操作...`, '🧠', 0)

  const selectedActions = [...primaryActions.value, ...secondaryActions.value]
    .filter(action => selectedAIActions.value.includes(action.id))

  let completed = 0
  const total = selectedActions.length

  const executeNext = () => {
    if (completed >= total) {
      showOperationStatus('success', `已完成 ${total} 个AI操作`, '✅', ['所有AI任务完成', '结果已汇总', '报告已生成'])
      selectedAIActions.value = []
      aiBatchMode.value = false
      return
    }

    const action = selectedActions[completed]
    showOperationStatus('loading', `正在执行: ${action.label} (${completed + 1}/${total})`, '🧠', Math.floor((completed / total) * 100))

    setTimeout(() => {
      completed++
      if (operationStatus.value) {
        operationStatus.value.progress = Math.floor((completed / total) * 100)
      }
      executeNext()
    }, 800)
  }

  executeNext()
}

// 关闭AI分析面板
const closeAIAnalysisPanel = () => {
  showAIAnalysisPanel.value = false
  aiAnalysisResults.value = null
}

// 查看AI操作历史
const viewAIActionHistory = () => {
  if (aiActionHistory.value.length === 0) {
    showOperationStatus('info', '暂无AI操作历史', '📜')
    return
  }

  const historyDetails = aiActionHistory.value.slice(0, 5).map((item, index) =>
    `${index + 1}. ${item.action} (${item.items}项, ${(item.confidence * 100).toFixed(1)}%)`
  )

  showOperationStatus('info', '最近AI操作历史', '📜', historyDetails)
}

const cancelConfirm = () => {
  showConfirmDialog.value = false
  confirmAction.value = null
  confirmMessage.value = ''
}

const executeConfirm = () => {
  if (confirmAction.value) {
    executeAction(confirmAction.value)
  }
  cancelConfirm()
}

// 监听props变化
watch(() => props.selectedItems, (newItems) => {
  selectedItems.value = newItems
}, { deep: true })

// 模拟进度更新（用于长时间操作）
const simulateProgress = () => {
  if (operationStatus.value?.type === 'loading') {
    let progress = 0
    const interval = setInterval(() => {
      progress += 10
      if (operationStatus.value) {
        operationStatus.value.progress = progress
      }
      if (progress >= 100) {
        clearInterval(interval)
      }
    }, 200)
  }
}

// 生命周期
onMounted(() => {
  // 初始化AI建议
  if (props.aiEnabled) {
    generateAISuggestions()
  }
})

// 监听选择变化
watch(() => selectedItems.value, () => {
  if (aiEnhanced.value && hasSelection.value) {
    generateAISuggestions()
  }
}, { deep: true })

// 暴露方法
defineExpose({
  clearSelection,
  toggleAIEnhance,
  showOperationStatus,
  generateAISuggestions,
  toggleAIBatchMode,
  executeSelectedAIActions,
  viewAIActionHistory
})
</script>

<style scoped>
@import '@/styles/theme/neuro-theme.css';
.neuro-action-bar {
  position: relative;
  background: linear-gradient(135deg, rgba(26, 26, 46, 0.95), rgba(40, 40, 60, 0.95));
  border: 2px solid var(--neuro-primary-30);
  border-radius: var(--neuro-radius-xl);
  padding: 16px 24px;
  margin: 20px 0;
  box-shadow: 0 8px 32px rgba(0, 245, 212, 0.15);
  backdrop-filter: blur(10px);
  transition: all var(--neuro-transition-normal);
}

.neuro-action-bar.has-selection {
  border-color: rgba(0, 245, 212, 0.6);
  box-shadow: 0 12px 40px rgba(0, 245, 212, 0.25);
}

.neuro-action-bar.ai-enhanced {
  border-color: rgba(0, 150, 255, 0.6);
  box-shadow: 0 12px 40px rgba(0, 150, 255, 0.25);
}

/* 选择状态指示器 */
.selection-indicator {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 16px;
  background: var(--neuro-primary-10);
  border-radius: var(--neuro-radius-lg);
  margin-bottom: 16px;
  position: relative;
  overflow: hidden;
}

.selection-pulse {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: radial-gradient(circle at center, var(--neuro-primary-20), transparent 70%);
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0% { opacity: 0.3; }
  50% { opacity: 0.6; }
  100% { opacity: 0.3; }
}

.selection-icon {
  font-size: 20px;
  color: var(--neuro-primary);
}

.selection-count {
  flex: 1;
  color: var(--neuro-text);
  font-weight: bold;
  font-size: 14px;
}

.clear-selection {
  background: transparent;
  border: none;
  color: rgba(255, 255, 255, 0.5);
  cursor: pointer;
  padding: var(--neuro-spacing-xs);
  transition: all 0.2s ease;
}

.clear-selection:hover {
  color: #ff6b6b;
  transform: scale(1.1);
}

.clear-icon {
  font-size: 16px;
  font-weight: bold;
}

/* 操作按钮组 */
.action-buttons {
  display: flex;
  align-items: center;
  gap: 16px;
  flex-wrap: wrap;
}

.primary-actions {
  display: flex;
  gap: 12px;
  flex: 1;
  flex-wrap: wrap;
}

.action-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 20px;
  border: none;
  border-radius: var(--neuro-radius-lg);
  font-weight: bold;
  cursor: pointer;
  transition: all var(--neuro-transition-normal);
  position: relative;
  overflow: hidden;
}

.action-btn.primary {
  background: linear-gradient(135deg, var(--neuro-primary-20), var(--neuro-primary-10));
  border: 1px solid var(--neuro-primary-30);
  color: var(--neuro-primary);
}

.action-btn.primary:hover:not(.disabled) {
  background: linear-gradient(135deg, var(--neuro-primary-30), var(--neuro-primary-20));
  transform: translateY(-2px);
  box-shadow: 0 8px 20px var(--neuro-primary-30);
}

.action-btn.primary.disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.action-glow {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: radial-gradient(circle at center, rgba(0, 245, 212, 0.4), transparent 70%);
  animation: glow 1.5s infinite alternate;
}

@keyframes glow {
  from { opacity: 0.3; }
  to { opacity: 0.6; }
}

.action-icon {
  font-size: 18px;
}

.action-label {
  font-size: 14px;
  letter-spacing: 0.5px;
}

/* 更多操作按钮 */
.secondary-actions {
  position: relative;
}

.more-actions-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 20px;
  background: var(--neuro-surface-90);
  border: 1px solid var(--neuro-surface-90);
  border-radius: var(--neuro-radius-lg);
  color: var(--neuro-text);
  cursor: pointer;
  transition: all var(--neuro-transition-normal);
}

.more-actions-btn:hover {
  background: var(--neuro-surface-90);
  transform: translateY(-2px);
}

.more-icon {
  font-size: 20px;
}

.more-label {
  font-size: 14px;
}

.dropdown-arrow {
  font-size: 12px;
  transition: transform 0.3s ease;
}

.dropdown-arrow.open {
  transform: rotate(180deg);
}

.dropdown-menu {
  position: absolute;
  top: calc(100% + 8px);
  right: 0;
  min-width: 200px;
  background: var(--neuro-surface);
  border: 2px solid var(--neuro-primary-30);
  border-radius: var(--neuro-radius-lg);
  padding: var(--neuro-spacing-sm);
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(20px);
  z-index: 1000;
  animation: slideDown 0.3s ease;
}

@keyframes slideDown {
  from { opacity: 0; transform: translateY(-10px); }
  to { opacity: 1; transform: translateY(0); }
}

.dropdown-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  border-radius: var(--neuro-radius-md);
  cursor: pointer;
  transition: all 0.2s ease;
}

.dropdown-item:hover:not(.disabled) {
  background: var(--neuro-primary-10);
}

.dropdown-item.disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.dropdown-icon {
  font-size: 18px;
  color: var(--neuro-primary);
}

.dropdown-label {
  flex: 1;
  color: var(--neuro-text);
  font-size: 14px;
}

.dropdown-desc {
  color: rgba(255, 255, 255, 0.5);
  font-size: 12px;
}

/* AI操作 */
.ai-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.ai-action-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 20px;
  background: linear-gradient(135deg, rgba(0, 150, 255, 0.2), rgba(0, 150, 255, 0.1));
  border: 1px solid var(--neuro-info-30);
  border-radius: var(--neuro-radius-lg);
  color: #0096ff;
  font-weight: bold;
  cursor: pointer;
  transition: all var(--neuro-transition-normal);
  position: relative;
  overflow: hidden;
}

.ai-action-btn:hover {
  background: linear-gradient(135deg, rgba(0, 150, 255, 0.3), rgba(0, 150, 255, 0.2));
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(0, 150, 255, 0.3);
}

.ai-pulse {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: radial-gradient(circle at center, rgba(0, 150, 255, 0.3), transparent 70%);
  animation: aiPulse 1.5s infinite;
}

@keyframes aiPulse {
  0% { opacity: 0.3; }
  50% { opacity: 0.6; }
  100% { opacity: 0.3; }
}

.ai-icon {
  font-size: 18px;
}

.ai-label {
  font-size: 14px;
}

.ai-suggestions {
  display: flex;
  gap: 8px;
}

.ai-suggestion {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  background: var(--neuro-info-10);
  border: 1px solid rgba(0, 150, 255, 0.2);
  border-radius: var(--neuro-radius-md);
  color: #0096ff;
  font-size: 12px;
  transition: all 0.2s ease;
}

.ai-suggestion:hover {
  background: rgba(0, 150, 255, 0.2);
  transform: translateY(-1px);
}

.suggestion-icon {
  font-size: 14px;
}

.suggestion-text {
  flex: 1;
}

.suggestion-action {
  background: transparent;
  border: none;
  color: #9333ea;
  cursor: pointer;
  padding: 2px;
  transition: all 0.2s ease;
}

.suggestion-action:hover {
  transform: scale(1.1);
}

/* 批量操作面板 */
.batch-panel {
  margin-top: 16px;
  padding: var(--neuro-spacing-md);
  background: rgba(0, 245, 212, 0.05);
  border: 1px solid var(--neuro-primary-20);
  border-radius: var(--neuro-radius-lg);
  animation: slideDown 0.3s ease;
}

.batch-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 12px;
}

.batch-icon {
  font-size: 18px;
  color: var(--neuro-primary);
}

.batch-title {
  flex: 1;
  color: var(--neuro-text);
  font-weight: bold;
  font-size: 14px;
}

.close-batch {
  background: transparent;
  border: none;
  color: rgba(255, 255, 255, 0.5);
  cursor: pointer;
  padding: var(--neuro-spacing-xs);
  transition: all 0.2s ease;
}

.close-batch:hover {
  color: #ff6b6b;
}

.batch-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.batch-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background: var(--neuro-primary-10);
  border: 1px solid var(--neuro-primary-30);
  border-radius: var(--neuro-radius-md);
  color: var(--neuro-primary);
  cursor: pointer;
  transition: all 0.2s ease;
}

.batch-btn:hover {
  background: var(--neuro-primary-20);
  transform: translateY(-1px);
}

.batch-icon {
  font-size: 16px;
}

.batch-label {
  font-size: 13px;
}

/* 操作状态指示器 */
.operation-status {
  margin-top: 16px;
}

.status-indicator {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 16px;
  border-radius: var(--neuro-radius-lg);
  animation: slideDown 0.3s ease;
}

.status-indicator.loading {
  background: var(--neuro-primary-10);
  border: 1px solid var(--neuro-primary-30);
  color: var(--neuro-primary);
}

.status-indicator.success {
  background: rgba(46, 204, 113, 0.1);
  border: 1px solid var(--neuro-success-30);
  color: var(--neuro-success);
}

.status-indicator.error {
  background: rgba(231, 76, 60, 0.1);
  border: 1px solid var(--neuro-error-30);
  color: var(--neuro-error);
}

.status-indicator.info {
  background: rgba(52, 152, 219, 0.1);
  border: 1px solid var(--neuro-info-30);
  color: var(--neuro-info);
}

.status-icon {
  font-size: 18px;
}

.status-text {
  flex: 1;
  font-size: 14px;
  font-weight: 500;
}

.status-progress {
  width: 100px;
  height: 4px;
  background: var(--neuro-surface-90);
  border-radius: 2px;
  overflow: hidden;
}

.progress-bar {
  height: 100%;
  background: linear-gradient(90deg, var(--neuro-primary), var(--neuro-secondary));
  transition: width 0.3s ease;
}

/* 确认对话框 */
.confirm-dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2000;
  backdrop-filter: blur(5px);
  animation: fadeIn 0.3s ease;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

.confirm-dialog {
  background: linear-gradient(135deg, rgba(26, 26, 46, 0.98), rgba(40, 40, 60, 0.98));
  border: 2px solid var(--neuro-primary-30);
  border-radius: var(--neuro-radius-xl);
  padding: var(--neuro-spacing-lg);
  max-width: 400px;
  width: 90%;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
  animation: slideUp 0.3s ease;
}

@keyframes slideUp {
  from { opacity: 0; transform: translateY(20px); }
  to { opacity: 1; transform: translateY(0); }
}

.dialog-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 20px;
}

.dialog-icon {
  font-size: 24px;
  color: var(--neuro-primary);
}

.dialog-title {
  color: var(--neuro-text);
  font-size: 18px;
  font-weight: bold;
}

.dialog-content {
  color: rgba(255, 255, 255, 0.8);
  line-height: 1.6;
  margin-bottom: 24px;
}

.dialog-actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
}

.dialog-btn {
  padding: 10px 24px;
  border: none;
  border-radius: var(--neuro-radius-lg);
  font-weight: bold;
  cursor: pointer;
  transition: all var(--neuro-transition-normal);
}

.dialog-btn.cancel {
  background: var(--neuro-surface-90);
  border: 1px solid var(--neuro-surface-90);
  color: var(--neuro-text);
}

.dialog-btn.cancel:hover {
  background: var(--neuro-surface-90);
}

.dialog-btn.confirm {
  background: linear-gradient(135deg, var(--neuro-primary), #00b8a9);
  color: #1a1a2e;
}

.dialog-btn.confirm:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(0, 245, 212, 0.4);
}

/* AI批量模式控制 */
.ai-batch-controls {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-left: 12px;
}

.ai-batch-toggle {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background: var(--neuro-info-10);
  border: 1px solid var(--neuro-info-30);
  border-radius: var(--neuro-radius-lg);
  color: #0096ff;
  font-weight: bold;
  cursor: pointer;
  transition: all var(--neuro-transition-normal);
}

.ai-batch-toggle:hover {
  background: rgba(0, 150, 255, 0.2);
  transform: translateY(-1px);
}

.ai-batch-toggle.active {
  background: rgba(0, 150, 255, 0.3);
  border-color: #0096ff;
  box-shadow: 0 0 20px rgba(0, 150, 255, 0.3);
}

.batch-toggle-icon {
  font-size: 16px;
}

.batch-toggle-text {
  font-size: 13px;
}

.ai-history-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  background: var(--neuro-surface-90);
  border: 1px solid var(--neuro-surface-90);
  border-radius: 10px;
  color: var(--neuro-text);
  cursor: pointer;
  transition: all var(--neuro-transition-normal);
}

.ai-history-btn:hover {
  background: var(--neuro-surface-90);
  transform: translateY(-1px);
}

.history-icon {
  font-size: 16px;
}

/* AI建议置信度 */
.suggestion-confidence {
  padding: 2px 6px;
  background: var(--neuro-info-10);
  border-radius: 6px;
  font-size: 11px;
  font-weight: bold;
  color: #0096ff;
}

/* AI批量操作面板 */
.ai-batch-panel {
  margin-top: 16px;
  padding: 20px;
  background: rgba(0, 150, 255, 0.05);
  border: 2px solid rgba(0, 150, 255, 0.3);
  border-radius: var(--neuro-radius-xl);
  animation: slideDown 0.3s ease;
}

.ai-batch-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 20px;
  padding-bottom: 12px;
  border-bottom: 1px solid rgba(0, 150, 255, 0.2);
}

.ai-batch-icon {
  font-size: 24px;
  color: #0096ff;
}

.ai-batch-title {
  flex: 1;
  color: var(--neuro-text);
  font-size: 18px;
  font-weight: bold;
}

.ai-batch-count {
  padding: 6px 12px;
  background: rgba(0, 150, 255, 0.2);
  border-radius: var(--neuro-radius-md);
  color: #0096ff;
  font-size: 14px;
  font-weight: bold;
}

.ai-batch-actions {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.ai-action-selection {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 24px;
}

.selection-group {
  background: rgba(255, 255, 255, 0.05);
  border-radius: var(--neuro-radius-lg);
  padding: var(--neuro-spacing-md);
}

.group-title {
  color: var(--neuro-primary);
  font-size: 16px;
  font-weight: bold;
  margin-bottom: 12px;
  padding-bottom: 8px;
  border-bottom: 1px solid var(--neuro-primary-20);
}

.action-checkboxes {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.action-checkbox {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  background: rgba(255, 255, 255, 0.05);
  border-radius: var(--neuro-radius-md);
  cursor: pointer;
  transition: all 0.2s ease;
}

.action-checkbox:hover {
  background: var(--neuro-info-10);
}

.action-checkbox input[type="checkbox"] {
  width: 16px;
  height: 16px;
  accent-color: #0096ff;
}

.checkbox-icon {
  font-size: 18px;
  color: #0096ff;
}

.checkbox-label {
  flex: 1;
  color: var(--neuro-text);
  font-size: 14px;
}

.checkbox-confidence {
  padding: 2px 8px;
  background: rgba(0, 150, 255, 0.2);
  border-radius: 6px;
  font-size: 12px;
  font-weight: bold;
  color: #0096ff;
}

.ai-batch-execute {
  display: flex;
  justify-content: center;
  padding-top: 16px;
  border-top: 1px solid rgba(0, 150, 255, 0.2);
}

.execute-btn {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 14px 32px;
  background: linear-gradient(135deg, #0096ff, #0077cc);
  border: none;
  border-radius: var(--neuro-radius-lg);
  color: var(--neuro-text);
  font-size: 16px;
  font-weight: bold;
  cursor: pointer;
  transition: all var(--neuro-transition-normal);
}

.execute-btn:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(0, 150, 255, 0.4);
}

.execute-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.execute-icon {
  font-size: 20px;
}

.execute-count {
  background: var(--neuro-surface-90);
  padding: 2px 8px;
  border-radius: 6px;
  font-size: 14px;
}

/* AI分析结果面板 */
.ai-analysis-panel-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.8);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 3000;
  backdrop-filter: blur(10px);
  animation: fadeIn 0.3s ease;
}

.ai-analysis-panel {
  background: linear-gradient(135deg, rgba(26, 26, 46, 0.98), rgba(40, 40, 60, 0.98));
  border: 3px solid rgba(0, 150, 255, 0.4);
  border-radius: 20px;
  width: 90%;
  max-width: 800px;
  max-height: 90vh;
  overflow-y: auto;
  box-shadow: 0 30px 80px rgba(0, 150, 255, 0.3);
  animation: slideUp 0.4s ease;
}

.analysis-header {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: var(--neuro-spacing-lg);
  background: var(--neuro-info-10);
  border-bottom: 2px solid rgba(0, 150, 255, 0.3);
}

.analysis-icon {
  font-size: 28px;
  color: #0096ff;
}

.analysis-title {
  flex: 1;
  color: var(--neuro-text);
  font-size: 24px;
  font-weight: bold;
}

.close-analysis {
  background: transparent;
  border: none;
  color: rgba(255, 255, 255, 0.5);
  cursor: pointer;
  padding: var(--neuro-spacing-sm);
  font-size: 20px;
  transition: all 0.2s ease;
}

.close-analysis:hover {
  color: #ff6b6b;
  transform: scale(1.1);
}

.analysis-content {
  padding: var(--neuro-spacing-lg);
}

.analysis-summary {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
  margin-bottom: 32px;
  padding: 20px;
  background: rgba(0, 150, 255, 0.05);
  border-radius: var(--neuro-radius-lg);
}

.summary-item {
  display: flex;
  align-items: center;
  gap: 12px;
}

.summary-icon {
  font-size: 20px;
  color: #0096ff;
}

.summary-label {
  color: rgba(255, 255, 255, 0.7);
  font-size: 14px;
}

.summary-value {
  color: var(--neuro-text);
  font-weight: bold;
  font-size: 16px;
}

.analysis-sections {
  display: flex;
  flex-direction: column;
  gap: 24px;
  margin-bottom: 32px;
}

.analysis-section {
  background: rgba(255, 255, 255, 0.05);
  border-radius: var(--neuro-radius-lg);
  padding: 20px;
  border-left: 4px solid;
}

.analysis-section:nth-child(1) {
  border-left-color: var(--neuro-primary);
}

.analysis-section:nth-child(2) {
  border-left-color: var(--neuro-success);
}

.analysis-section:nth-child(3) {
  border-left-color: var(--neuro-error);
}

.section-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--neuro-surface-90);
}

.section-icon {
  font-size: 20px;
}

.section-title {
  color: var(--neuro-text);
  font-size: 18px;
  font-weight: bold;
}

.section-content {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.insight-item,
.recommendation-item,
.risk-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 12px;
  background: rgba(255, 255, 255, 0.03);
  border-radius: var(--neuro-radius-md);
}

.insight-icon {
  color: var(--neuro-primary);
  font-size: 16px;
  margin-top: 2px;
}

.recommendation-icon {
  color: var(--neuro-success);
  font-size: 16px;
  margin-top: 2px;
}

.risk-icon {
  color: var(--neuro-error);
  font-size: 16px;
  margin-top: 2px;
}

.insight-text,
.recommendation-text,
.risk-text {
  color: rgba(255, 255, 255, 0.9);
  font-size: 14px;
  line-height: 1.5;
  flex: 1;
}

.analysis-actions {
  display: flex;
  gap: 16px;
  justify-content: center;
  padding-top: 24px;
  border-top: 1px solid var(--neuro-surface-90);
}

.analysis-action-btn {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 24px;
  border: none;
  border-radius: var(--neuro-radius-lg);
  font-weight: bold;
  cursor: pointer;
  transition: all var(--neuro-transition-normal);
}

.analysis-action-btn.export {
  background: linear-gradient(135deg, var(--neuro-primary), #00b8a9);
  color: #1a1a2e;
}

.analysis-action-btn.share {
  background: linear-gradient(135deg, #9333ea, #7c3aed);
  color: var(--neuro-text);
}

.analysis-action-btn.schedule {
  background: linear-gradient(135deg, var(--neuro-warning), #f39c12);
  color: #1a1a2e;
}

.analysis-action-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(0, 0, 0, 0.3);
}

.action-icon {
  font-size: 18px;
}

.action-text {
  font-size: 14px;
}

/* 操作状态详情 */
.status-indicator .status-details {
  margin-top: 8px;
  padding-left: 24px;
}

.status-details ul {
  list-style: none;
  margin: 0;
  padding: 0;
}

.status-details li {
  color: rgba(255, 255, 255, 0.8);
  font-size: 12px;
  margin-bottom: 4px;
  padding-left: 16px;
  position: relative;
}

.status-details li:before {
  content: '•';
  position: absolute;
  left: 0;
  color: var(--neuro-primary);
}

/* 响应式设计 */
@media (max-width: 768px) {
  .ai-batch-controls {
    margin-left: 0;
    margin-top: 12px;
    width: 100%;
    justify-content: center;
  }

  .ai-action-selection {
    grid-template-columns: 1fr;
  }

  .analysis-summary {
    grid-template-columns: 1fr;
  }

  .analysis-actions {
    flex-direction: column;
  }

  .analysis-action-btn {
    width: 100%;
    justify-content: center;
  }

  .ai-batch-panel {
    padding: var(--neuro-spacing-md);
  }

  .ai-analysis-panel {
    width: 95%;
    max-height: 85vh;
  }
}

@media (max-width: 480px) {
  .ai-batch-toggle .batch-toggle-text {
    display: none;
  }

  .ai-batch-toggle {
    padding: var(--neuro-spacing-sm);
  }

  .action-checkbox {
    padding: var(--neuro-spacing-sm);
  }

  .checkbox-label {
    font-size: 13px;
  }

  .analysis-header {
    padding: var(--neuro-spacing-md);
  }

  .analysis-title {
    font-size: 20px;
  }

  .analysis-content {
    padding: var(--neuro-spacing-md);
  }
}
</style>