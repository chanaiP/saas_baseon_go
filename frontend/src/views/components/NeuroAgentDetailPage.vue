<template>
  <div class="neuro-agent-detail-page" :data-theme="theme">
    <!-- 全息3D背景层 -->
    <div class="hologram-3d-background">
      <div class="hologram-layer layer-1"></div>
      <div class="hologram-layer layer-2"></div>
      <div class="hologram-layer layer-3"></div>
      <div class="hologram-particle-field">
        <div v-for="n in 12" :key="n" class="hologram-particle" :style="{
          left: `${Math.random() * 100}%`,
          top: `${Math.random() * 100}%`,
          animationDelay: `${n * 0.3}s`
        }"></div>
      </div>
    </div>

    <!-- 页面头部 -->
    <header class="detail-header">
      <div class="header-back">
        <button class="back-button" @click="handleBack">
          <span class="back-icon">←</span>
          <span class="back-text">返回列表</span>
        </button>
      </div>

      <div class="header-title">
        <h1 class="detail-title">
          <span class="title-icon">{{ data.icon || '📋' }}</span>
          <span class="title-text">{{ data.title || '项目详情' }}</span>
          <span class="title-status" :class="'status-' + data.status">
            {{ getStatusText(data.status) }}
          </span>
        </h1>

        <div class="header-subtitle">
          <span class="subtitle-id">ID: {{ data.id || 'N/A' }}</span>
          <span class="subtitle-separator">•</span>
          <span class="subtitle-updated">最后更新: {{ formatDate(data.updatedAt) }}</span>
        </div>
      </div>

      <div class="header-actions">
        <button class="neuro-btn neuro-btn-secondary" @click="handleEdit">
          <span class="btn-icon">✏️</span>
          <span class="btn-label">编辑</span>
        </button>
        <button class="neuro-btn neuro-btn-primary" @click="handleAction('primary')">
          <span class="btn-icon">⚡</span>
          <span class="btn-label">{{ getPrimaryActionText() }}</span>
        </button>
        <div class="neuro-btn-dropdown">
          <button class="neuro-btn neuro-btn-ghost">
            <span class="btn-icon">⋯</span>
          </button>
          <div class="dropdown-menu">
            <button class="dropdown-item" @click="handleDuplicate">复制项目</button>
            <button class="dropdown-item" @click="handleExport">导出数据</button>
            <button class="dropdown-item" @click="handleArchive">归档</button>
            <div class="dropdown-divider"></div>
            <button class="dropdown-item dropdown-danger" @click="handleDelete">删除</button>
          </div>
        </div>
      </div>
    </header>

    <!-- 主内容区域 - 3D分层布局 -->
    <div class="detail-content-3d">
      <!-- 左侧：核心信息层 -->
      <div class="content-layer layer-core">
        <div class="layer-header">
          <h2 class="layer-title">
            <span class="title-icon">🧠</span>
            <span class="title-text">核心信息</span>
          </h2>
          <div class="layer-glow"></div>
        </div>

        <div class="core-info-cards">
          <!-- 基本信息卡片 -->
          <div class="info-card card-basic">
            <div class="card-header">
              <h3 class="card-title">基本信息</h3>
              <span class="card-icon">📝</span>
            </div>
            <div class="card-content">
              <div class="info-row">
                <span class="info-label">项目名称</span>
                <span class="info-value">{{ data.name || '未命名' }}</span>
              </div>
              <div class="info-row">
                <span class="info-label">负责人</span>
                <span class="info-value">{{ data.owner || '未分配' }}</span>
              </div>
              <div class="info-row">
                <span class="info-label">创建时间</span>
                <span class="info-value">{{ formatDate(data.createdAt) }}</span>
              </div>
              <div class="info-row">
                <span class="info-label">更新时间</span>
                <span class="info-value">{{ formatDate(data.updatedAt) }}</span>
              </div>
            </div>
          </div>

          <!-- 状态卡片 -->
          <div class="info-card card-status">
            <div class="card-header">
              <h3 class="card-title">项目状态</h3>
              <span class="card-icon">📊</span>
            </div>
            <div class="card-content">
              <div class="status-visual">
                <div class="status-progress">
                  <div class="progress-bar">
                    <div class="progress-fill" :style="{ width: data.progress + '%' }"></div>
                  </div>
                  <span class="progress-text">{{ data.progress || 0 }}%</span>
                </div>
                <div class="status-tags">
                  <span class="status-tag" :class="'priority-' + data.priority">
                    {{ getPriorityText(data.priority) }}
                  </span>
                  <span class="status-tag" :class="'status-' + data.status">
                    {{ getStatusText(data.status) }}
                  </span>
                </div>
              </div>
              <div class="status-metrics">
                <div class="metric-item">
                  <span class="metric-label">预计完成</span>
                  <span class="metric-value">{{ data.estimatedCompletion || '未设置' }}</span>
                </div>
                <div class="metric-item">
                  <span class="metric-label">剩余时间</span>
                  <span class="metric-value">{{ data.remainingTime || 'N/A' }}</span>
                </div>
              </div>
            </div>
          </div>

          <!-- 标签卡片 -->
          <div class="info-card card-tags" v-if="data.tags && data.tags.length > 0">
            <div class="card-header">
              <h3 class="card-title">项目标签</h3>
              <span class="card-icon">🏷️</span>
            </div>
            <div class="card-content">
              <div class="tags-cloud">
                <span
                  v-for="(tag, index) in data.tags"
                  :key="index"
                  class="tag-item"
                  :style="{ '--tag-index': index }"
                >
                  {{ tag }}
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 中间：详细描述层 -->
      <div class="content-layer layer-description">
        <div class="layer-header">
          <h2 class="layer-title">
            <span class="title-icon">📋</span>
            <span class="title-text">详细描述</span>
          </h2>
          <div class="layer-glow"></div>
        </div>

        <div class="description-content">
          <div class="description-text" v-if="data.description">
            {{ data.description }}
          </div>
          <div class="description-empty" v-else>
            <div class="empty-icon">📄</div>
            <p class="empty-text">暂无项目描述</p>
            <button class="neuro-btn neuro-btn-ghost" @click="handleAddDescription">
              添加描述
            </button>
          </div>

          <!-- 附件区域 -->
          <div class="description-attachments" v-if="data.attachments && data.attachments.length > 0">
            <h3 class="attachments-title">相关附件</h3>
            <div class="attachments-list">
              <div
                v-for="attachment in data.attachments"
                :key="attachment.id"
                class="attachment-item"
              >
                <span class="attachment-icon">{{ getAttachmentIcon(attachment.type) }}</span>
                <span class="attachment-name">{{ attachment.name }}</span>
                <span class="attachment-size">{{ attachment.size }}</span>
                <button class="attachment-action" @click="handleDownload(attachment)">
                  <span class="action-icon">⬇️</span>
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 右侧：AI分析层 -->
      <div class="content-layer layer-analysis">
        <div class="layer-header">
          <h2 class="layer-title">
            <span class="title-icon">🤖</span>
            <span class="title-text">AI分析</span>
          </h2>
          <div class="layer-glow"></div>
        </div>

        <div class="analysis-content">
          <!-- AI洞察 -->
          <div class="analysis-insights">
            <h3 class="insights-title">AI洞察</h3>
            <div class="insights-list">
              <div class="insight-item" v-for="insight in aiInsights" :key="insight.id">
                <span class="insight-icon">{{ insight.icon }}</span>
                <div class="insight-content">
                  <p class="insight-text">{{ insight.text }}</p>
                  <span class="insight-confidence">
                    置信度: <span class="confidence-value">{{ insight.confidence }}%</span>
                  </span>
                </div>
              </div>
            </div>
          </div>

          <!-- 预测分析 -->
          <div class="analysis-predictions">
            <h3 class="predictions-title">预测分析</h3>
            <div class="predictions-chart">
              <div class="chart-container">
                <div class="chart-bars">
                  <div
                    v-for="prediction in predictions"
                    :key="prediction.id"
                    class="chart-bar"
                    :style="{ height: prediction.value + '%' }"
                    :title="prediction.label + ': ' + prediction.value + '%'"
                  >
                    <span class="bar-label">{{ prediction.label }}</span>
                  </div>
                </div>
              </div>
              <div class="chart-legend">
                <div class="legend-item" v-for="prediction in predictions" :key="prediction.id">
                  <span class="legend-color" :style="{ backgroundColor: prediction.color }"></span>
                  <span class="legend-text">{{ prediction.label }}</span>
                </div>
              </div>
            </div>
          </div>

          <!-- 建议操作 -->
          <div class="analysis-suggestions">
            <h3 class="suggestions-title">建议操作</h3>
            <div class="suggestions-list">
              <button
                v-for="suggestion in aiSuggestions"
                :key="suggestion.id"
                class="suggestion-item"
                @click="handleSuggestion(suggestion)"
              >
                <span class="suggestion-icon">{{ suggestion.icon }}</span>
                <span class="suggestion-text">{{ suggestion.text }}</span>
                <span class="suggestion-priority" :class="'priority-' + suggestion.priority">
                  {{ suggestion.priority }}
                </span>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 底部：时间线层 -->
    <div class="detail-timeline">
      <div class="timeline-header">
        <h2 class="timeline-title">
          <span class="title-icon">🕒</span>
          <span class="title-text">时间线</span>
        </h2>
        <button class="timeline-filter" @click="toggleTimelineFilter">
          <span class="filter-icon">🔍</span>
          <span class="filter-text">筛选</span>
        </button>
      </div>

      <div class="timeline-container">
        <div class="timeline-track">
          <div
            v-for="event in timelineEvents"
            :key="event.id"
            class="timeline-event"
            :class="'event-' + event.type"
            :style="{ left: event.position + '%' }"
          >
            <div class="event-marker"></div>
            <div class="event-content">
              <div class="event-header">
                <span class="event-icon">{{ event.icon }}</span>
                <span class="event-title">{{ event.title }}</span>
                <span class="event-time">{{ event.time }}</span>
              </div>
              <p class="event-description" v-if="event.description">{{ event.description }}</p>
              <div class="event-user" v-if="event.user && typeof event.user === 'object'">
                <span class="user-avatar">{{ (event.user as any).avatar }}</span>
                <span class="user-name">{{ (event.user as any).name }}</span>
              </div>
            </div>
          </div>
        </div>

        <div class="timeline-scale">
          <span class="scale-label">1周前</span>
          <span class="scale-label">3天前</span>
          <span class="scale-label">昨天</span>
          <span class="scale-label">今天</span>
          <span class="scale-label">明天</span>
        </div>
      </div>
    </div>

    <!-- 全息导航点 -->
    <div class="hologram-nav-dots">
      <button
        v-for="dot in navDots"
        :key="dot.id"
        class="nav-dot"
        :class="{ 'dot-active': activeLayer === dot.layer }"
        @click="scrollToLayer(dot.layer)"
        :title="dot.title"
      >
        <span class="dot-glow"></span>
      </button>
    </div>

    <!-- AI助手浮窗 -->
    <div class="ai-assistant-float" :class="{ 'assistant-minimized': isAssistantMinimized }">
      <div class="assistant-header" @click="toggleAssistant">
        <span class="assistant-icon">🤖</span>
        <span class="assistant-title">详情助手</span>
        <span class="assistant-toggle">{{ isAssistantMinimized ? '▶' : '▼' }}</span>
      </div>

      <div class="assistant-content" v-if="!isAssistantMinimized">
        <div class="assistant-message">
          <p>我可以帮你分析这个项目的：</p>
          <ul class="assistant-features">
            <li>📈 进度趋势和风险</li>
            <li>🔍 关键问题和瓶颈</li>
            <li>💡 优化建议和策略</li>
            <li>📊 数据可视化和报告</li>
          </ul>
        </div>

        <div class="assistant-quick-questions">
          <button
            v-for="question in quickQuestions"
            :key="question.id"
            class="question-btn"
            @click="askAssistant(question.text)"
          >
            {{ question.text }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'

// 组件属性
interface Props {
  data?: any
  theme?: 'dark' | 'light'
}

const props = withDefaults(defineProps<Props>(), {
  data: () => ({
    id: 'project-001',
    title: 'AI智能项目管理平台',
    name: 'NeuroMatrix PM System',
    icon: '🧠',
    status: 'active',
    progress: 75,
    priority: 'high',
    owner: '张AI',
    description: '基于DevOS 神经美学的AI协作项目管理平台，融合未来感交互设计和智能数据分析功能。',
    createdAt: new Date('2024-01-15'),
    updatedAt: new Date(),
    estimatedCompletion: '2024-06-30',
    remainingTime: '2个月',
    tags: ['AI', '前端', '管理', '创新', '未来感'],
    attachments: [
      { id: 1, name: '项目需求文档.pdf', type: 'pdf', size: '2.4MB' },
      { id: 2, name: '设计原型.fig', type: 'design', size: '5.7MB' },
      { id: 3, name: '技术架构图.png', type: 'image', size: '1.2MB' }
    ]
  }),
  theme: 'dark'
})

// 响应式状态
const activeLayer = ref('core')
const isAssistantMinimized = ref(false)
const showTimelineFilter = ref(false)

// AI分析数据
const aiInsights = ref([
  { id: 1, icon: '📈', text: '项目进度正常，但代码审查环节存在延迟', confidence: 85 },
  { id: 2, icon: '⚠️', text: '前端组件测试覆盖率低于目标值', confidence: 92 },
  { id: 3, icon: '💡', text: 'AI助手功能用户满意度较高，可考虑扩展', confidence: 78 },
  { id: 4, icon: '🎯', text: '建议优先完成核心数据可视化模块', confidence: 88 }
])

const predictions = ref([
  { id: 1, label: '按时完成', value: 65, color: 'var(--neuro-primary)' },
  { id: 2, label: '轻微延迟', value: 25, color: 'var(--neuro-warning)' },
  { id: 3, label: '重大延迟', value: 8, color: 'var(--neuro-error)' },
  { id: 4, label: '提前完成', value: 2, color: '#22c55e' }
])

const aiSuggestions = ref([
  { id: 1, icon: '🚀', text: '分配更多资源到前端测试', priority: '高' },
  { id: 2, icon: '🔍', text: '安排代码审查会议', priority: '中' },
  { id: 3, icon: '📊', text: '生成详细进度报告', priority: '低' },
  { id: 4, icon: '🤝', text: '与设计团队同步UI规范', priority: '中' }
])

// 时间线数据
const timelineEvents = ref([
  { id: 1, type: 'create', icon: '🚀', title: '项目创建', time: '2024-01-15', position: 0, user: { avatar: '张', name: '张AI' } },
  { id: 2, type: 'milestone', icon: '🎯', title: '完成需求分析', time: '2024-01-25', position: 15, description: '完成所有用户需求收集和分析' },
  { id: 3, type: 'design', icon: '🎨', title: 'UI设计完成', time: '2024-02-10', position: 30, user: { avatar: '李', name: '李设计' } },
  { id: 4, type: 'development', icon: '💻', title: '核心功能开发', time: '2024-03-01', position: 45, description: '完成DevOS 神经布局和基础组件' },
  { id: 5, type: 'test', icon: '🧪', title: '第一阶段测试', time: '2024-03-20', position: 60, user: { avatar: '王', name: '王测试' } },
  { id: 6, type: 'update', icon: '🔄', title: 'AI助手集成', time: '2024-04-05', position: 75, description: '集成AI分析预测功能' },
  { id: 7, type: 'current', icon: '📍', title: '当前进度', time: '今天', position: 85, user: { avatar: '你', name: '当前用户' } }
])

// 导航点
const navDots = ref([
  { id: 1, layer: 'core', title: '核心信息' },
  { id: 2, layer: 'description', title: '详细描述' },
  { id: 3, layer: 'analysis', title: 'AI分析' },
  { id: 4, layer: 'timeline', title: '时间线' }
])

// 快速问题
const quickQuestions = ref([
  { id: 1, text: '项目的主要风险是什么？' },
  { id: 2, text: '如何加快项目进度？' },
  { id: 3, text: '需要哪些资源支持？' },
  { id: 4, text: '生成项目报告' }
])

// 计算属性
const theme = computed(() => props.theme)

// 方法
const handleBack = () => {
  emit('back')
}

const handleEdit = () => {
  emit('edit', props.data)
}

const handleAction = (actionType: string) => {
  console.log('执行操作:', actionType)
  emit('action', { type: actionType, data: props.data })
}

const getPrimaryActionText = () => {
  const status = props.data.status
  switch (status) {
    case 'active': return '继续项目'
    case 'pending': return '开始项目'
    case 'paused': return '恢复项目'
    case 'completed': return '重新打开'
    default: return '操作'
  }
}

const handleDuplicate = () => {
  emit('duplicate', props.data)
}

const handleExport = () => {
  emit('export', props.data)
}

const handleArchive = () => {
  emit('archive', props.data)
}

const handleDelete = () => {
  if (confirm('确定归档这个项目吗？归档后默认不再出现在业务列表中。')) {
    emit('delete', props.data)
  }
}

const handleAddDescription = () => {
  emit('add-description', props.data)
}

const handleDownload = (attachment: any) => {
  console.log('下载附件:', attachment)
  emit('download', attachment)
}

const getStatusText = (status: string) => {
  const statusMap: Record<string, string> = {
    'active': '进行中',
    'pending': '待开始',
    'paused': '已暂停',
    'completed': '已完成'
  }
  return statusMap[status] || status
}

const getPriorityText = (priority: string) => {
  const priorityMap: Record<string, string> = {
    'high': '高优先级',
    'medium': '中优先级',
    'low': '低优先级'
  }
  return priorityMap[priority] || priority
}

const formatDate = (date: Date | string) => {
  if (!date) return 'N/A'
  const d = new Date(date)
  return d.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit'
  })
}

const getAttachmentIcon = (type: string) => {
  const iconMap: Record<string, string> = {
    'pdf': '📄',
    'doc': '📝',
    'image': '🖼️',
    'design': '🎨',
    'code': '💻',
    'data': '📊'
  }
  return iconMap[type] || '📎'
}

const handleSuggestion = (suggestion: any) => {
  console.log('执行AI建议:', suggestion)
  emit('suggestion', suggestion)
}

const toggleTimelineFilter = () => {
  showTimelineFilter.value = !showTimelineFilter.value
}

const scrollToLayer = (layer: string) => {
  activeLayer.value = layer
  const element = document.querySelector(`.layer-${layer}`)
  if (element) {
    element.scrollIntoView({ behavior: 'smooth', block: 'start' })
  }
}

const toggleAssistant = () => {
  isAssistantMinimized.value = !isAssistantMinimized.value
}

const askAssistant = (question: string) => {
  console.log('询问AI助手:', question)
  emit('assistant-ask', question)
}

// 事件
const emit = defineEmits<{
  back: []
  edit: [data: any]
  action: [action: any]
  duplicate: [data: any]
  export: [data: any]
  archive: [data: any]
  delete: [data: any]
  'add-description': [data: any]
  download: [attachment: any]
  suggestion: [suggestion: any]
  'assistant-ask': [question: string]
}>()

// 初始化
onMounted(() => {
  // 检测主题
  const layout = document.querySelector('.neuro-command-layout')
  if (layout) {
    const layoutTheme = layout.getAttribute('data-theme') as 'dark' | 'light'
    if (layoutTheme) {
      // 这里可以设置主题，但props是只读的
      // 实际应用中可能需要使用provide/inject或store
    }
  }
})
</script>

<style scoped>
@import '@/styles/theme/neuro-theme.css';
/* 基础样式 */
.neuro-agent-detail-page {
  position: relative;
  width: 100%;
  min-height: 100vh;
  background: rgba(15, 23, 42, 0.95);
  font-family: 'IBM Plex Mono', monospace;
  color: var(--neuro-text);
}

/* 头部样式 */
.detail-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 24px 32px;
  background: rgba(15, 23, 42, 0.9);
  backdrop-filter: blur(20px);
  border-bottom: 1px solid var(--neuro-primary-20);
}

.header-back .back-button {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 16px;
  background: rgba(15, 23, 42, 0.8);
  border: 1px solid var(--neuro-primary-30);
  border-radius: var(--neuro-radius-lg);
  color: var(--neuro-primary);
  font-family: 'IBM Plex Mono', monospace;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.back-button:hover {
  background: var(--neuro-primary-10);
  transform: translateX(-4px);
}

.header-title {
  flex: 1;
  text-align: center;
}

.detail-title {
  font-family: 'Orbitron', monospace;
  font-size: 28px;
  font-weight: 700;
  background: linear-gradient(135deg, var(--neuro-primary), #9d4edd);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  margin: 0 0 8px 0;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
}

.title-status {
  padding: 4px 12px;
  background: rgba(34, 197, 94, 0.1);
  border: 1px solid rgba(34, 197, 94, 0.3);
  border-radius: var(--neuro-radius-xl);
  font-size: 12px;
  color: #22c55e;
  font-weight: 500;
}

.header-subtitle {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  font-size: 13px;
  color: var(--neuro-text-secondary);
}

/* 内容区域 */
.detail-content-3d {
  display: grid;
  grid-template-columns: 1fr 2fr 1fr;
  gap: 24px;
  padding: var(--neuro-spacing-xl);
}

.content-layer {
  background: rgba(15, 23, 42, 0.7);
  backdrop-filter: blur(20px);
  border: 1px solid var(--neuro-primary-20);
  border-radius: var(--neuro-radius-xl);
  padding: var(--neuro-spacing-lg);
}

.layer-header {
  margin-bottom: 24px;
  padding-bottom: 16px;
  border-bottom: 1px solid var(--neuro-primary-10);
}

.layer-title {
  font-family: 'Orbitron', monospace;
  font-size: 18px;
  font-weight: 600;
  color: var(--neuro-primary);
  margin: 0;
  display: flex;
  align-items: center;
  gap: 8px;
}

/* 信息卡片 */
.info-card {
  background: rgba(15, 23, 42, 0.8);
  border: 1px solid var(--neuro-primary-10);
  border-radius: var(--neuro-radius-lg);
  padding: 20px;
  margin-bottom: 16px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.card-title {
  font-family: 'Orbitron', monospace;
  font-size: 16px;
  font-weight: 600;
  color: var(--neuro-text);
  margin: 0;
}

.card-icon {
  font-size: 18px;
  opacity: 0.8;
}

.info-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  padding-bottom: 12px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
}

.info-label {
  font-size: 13px;
  color: var(--neuro-text-secondary);
}

.info-value {
  font-size: 14px;
  font-weight: 500;
  color: var(--neuro-text);
}

/* 进度条 */
.progress-bar {
  height: 8px;
  background: rgba(15, 23, 42, 0.8);
  border-radius: var(--neuro-radius-sm);
  overflow: hidden;
  margin: 12px 0;
}

.progress-fill {
  height: 100%;
  background: linear-gradient(90deg, var(--neuro-primary), #9d4edd);
  border-radius: var(--neuro-radius-sm);
  transition: width 0.5s ease;
}

/* 标签云 */
.tags-cloud {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.tag-item {
  padding: 6px 12px;
  background: var(--neuro-primary-10);
  border: 1px solid var(--neuro-primary-30);
  border-radius: var(--neuro-radius-xl);
  font-size: 12px;
  color: var(--neuro-primary);
  animation: tag-float 3s infinite;
  animation-delay: calc(var(--tag-index) * 0.3s);
}

@keyframes tag-float {
  0%, 100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-4px);
  }
}

/* AI分析 */
.insight-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 12px;
  background: rgba(157, 78, 221, 0.1);
  border: 1px solid rgba(157, 78, 221, 0.2);
  border-radius: var(--neuro-radius-lg);
  margin-bottom: 12px;
}

.insight-icon {
  font-size: 16px;
}

.insight-text {
  flex: 1;
  font-size: 13px;
  color: var(--neuro-text);
  margin: 0 0 4px 0;
}

.insight-confidence {
  font-size: 11px;
  color: #9d4edd;
}

/* 时间线 */
.detail-timeline {
  padding: var(--neuro-spacing-xl);
  background: rgba(15, 23, 42, 0.7);
  backdrop-filter: blur(20px);
  border-top: 1px solid var(--neuro-primary-20);
}

.timeline-container {
  position: relative;
  height: 120px;
  margin-top: 24px;
}

.timeline-track {
  position: relative;
  height: 4px;
  background: var(--neuro-primary-20);
  border-radius: 2px;
  margin: 0 40px;
}

.timeline-event {
  position: absolute;
  top: -8px;
  transform: translateX(-50%);
}

.event-marker {
  width: 16px;
  height: 16px;
  background: var(--neuro-primary);
  border-radius: 50%;
  border: 2px solid rgba(15, 23, 42, 0.9);
  box-shadow: 0 0 8px rgba(0, 245, 212, 0.5);
}

/* AI助手 */
.ai-assistant-float {
  position: fixed;
  right: 24px;
  bottom: 24px;
  width: 300px;
  background: rgba(15, 23, 42, 0.95);
  backdrop-filter: blur(20px);
  border: 1px solid rgba(157, 78, 221, 0.3);
  border-radius: var(--neuro-radius-xl);
  overflow: hidden;
  z-index: 1000;
}

.assistant-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: var(--neuro-spacing-md);
  background: rgba(157, 78, 221, 0.1);
  cursor: pointer;
}

.assistant-icon {
  font-size: 20px;
}

.assistant-title {
  flex: 1;
  font-family: 'Orbitron', monospace;
  font-size: 16px;
  font-weight: 600;
  color: #9d4edd;
}

/* 按钮样式 */
.neuro-btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 10px 16px;
  border: none;
  border-radius: var(--neuro-radius-lg);
  font-family: 'IBM Plex Mono', monospace;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}

.neuro-btn-primary {
  background: linear-gradient(135deg, var(--neuro-primary), #9d4edd);
  color: var(--neuro-background);
}

.neuro-btn-secondary {
  background: rgba(15, 23, 42, 0.8);
  border: 1px solid var(--neuro-primary-30);
  color: var(--neuro-primary);
}

.neuro-btn-ghost {
  background: transparent;
  border: 1px solid rgba(100, 116, 139, 0.3);
  color: var(--neuro-text-secondary);
}

/* 响应式设计 */
@media (max-width: 1200px) {
  .detail-content-3d {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 768px) {
  .detail-header {
    flex-direction: column;
    gap: 16px;
    text-align: center;
  }

  .header-actions {
    width: 100%;
    justify-content: center;
  }

  .ai-assistant-float {
    width: 280px;
    right: 16px;
    bottom: 16px;
  }
}

/* 浅色模式 */
.neuro-agent-detail-page[data-theme="light"] {
  background: rgba(248, 250, 252, 0.95);
  color: var(--neuro-background);
}

.neuro-agent-detail-page[data-theme="light"] .detail-header,
.neuro-agent-detail-page[data-theme="light"] .content-layer,
.neuro-agent-detail-page[data-theme="light"] .info-card,
.neuro-agent-detail-page[data-theme="light"] .detail-timeline,
.neuro-agent-detail-page[data-theme="light"] .ai-assistant-float {
  background: rgba(248, 250, 252, 0.9);
  border-color: rgba(14, 165, 233, 0.2);
}

.neuro-agent-detail-page[data-theme="light"] .detail-title {
  background: linear-gradient(135deg, #0ea5e9, #8b5cf6);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

.neuro-agent-detail-page[data-theme="light"] .info-label,
.neuro-agent-detail-page[data-theme="light"] .header-subtitle {
  color: #64748b;
}

.neuro-agent-detail-page[data-theme="light"] .info-value,
.neuro-agent-detail-page[data-theme="light"] .card-title {
  color: var(--neuro-background);
}

.neuro-agent-detail-page[data-theme="light"] .tag-item {
  background: rgba(14, 165, 233, 0.1);
  border-color: rgba(14, 165, 233, 0.3);
  color: #0ea5e9;
}

.neuro-agent-detail-page[data-theme="light"] .progress-fill {
  background: linear-gradient(90deg, #0ea5e9, #8b5cf6);
}
</style>
