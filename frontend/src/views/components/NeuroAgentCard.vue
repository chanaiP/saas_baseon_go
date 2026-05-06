<template>
  <div class="neuro-card" :class="cardClasses">
    <!-- 卡片头部 -->
    <div class="card-header" @click="toggleExpand">
      <div class="header-left">
        <div class="card-icon">
          <span class="icon">{{ icon }}</span>
        </div>
        <div class="header-content">
          <h3 class="card-title">{{ title }}</h3>
          <p v-if="subtitle" class="card-subtitle">{{ subtitle }}</p>
        </div>
      </div>
      <div class="header-right">
        <slot name="badge-area">
          <div class="card-badges">
            <span v-if="status" class="status-badge" :class="getStatusClass(status)">
              <slot name="status-badge">{{ getStatusText(status) }}</slot>
            </span>
            <span v-if="priority" class="priority-badge" :class="getPriorityClass(priority)">
              {{ getPriorityText(priority) }}
            </span>
            <span v-if="tags && tags.length > 0" class="tags-count">
              <span class="tags-icon">🏷️</span>
              <span class="tags-number">{{ tags.length }}</span>
            </span>
          </div>
        </slot>
        <div class="header-actions">
          <button class="expand-btn" :class="{ expanded: isExpanded }">
            <span class="expand-icon">{{ isExpanded ? '▼' : '▶' }}</span>
          </button>
        </div>
      </div>
    </div>

    <!-- 卡片内容 -->
    <div v-if="isExpanded || !collapsible" class="card-content">
      <!-- 主要内容 -->
      <div class="main-content">
        <slot name="default">
          <div v-if="description" class="card-description">
            <p>{{ description }}</p>
          </div>
        </slot>

        <!-- 进度条 -->
        <div v-if="progress !== undefined" class="progress-section">
          <div class="progress-header">
            <span class="progress-label">进度</span>
            <span class="progress-value">{{ progress }}%</span>
          </div>
          <div class="progress-bar">
            <div class="progress-fill" :style="{ width: progress + '%' }"></div>
            <div class="progress-glow"></div>
          </div>
        </div>

        <!-- 标签 -->
        <div v-if="tags && tags.length > 0" class="tags-section">
          <div class="tags-label">标签</div>
          <div class="tags-list">
            <span v-for="tag in tags" :key="tag" class="tag" @click="handleTagClick(tag)">
              {{ tag }}
            </span>
          </div>
        </div>

        <!-- 元数据 -->
        <div v-if="metadata && metadata.length > 0" class="metadata-section">
          <div class="metadata-grid">
            <div v-for="item in metadata" :key="item.key" class="metadata-item">
              <span class="metadata-label">{{ item.label }}:</span>
              <span class="metadata-value">{{ item.value }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- AI分析部分 -->
      <div v-if="aiAnalysisEnabled" class="ai-analysis-section">
        <div class="ai-section-header">
          <span class="ai-icon">🧠</span>
          <span class="ai-title">AI分析</span>
        </div>
        <div class="ai-insights">
          <div v-for="insight in aiInsights" :key="insight.id" class="ai-insight">
            <span class="insight-icon">{{ insight.icon }}</span>
            <span class="insight-text">{{ insight.text }}</span>
          </div>
        </div>
      </div>

      <!-- 语音交互面板 -->
      <div v-if="voiceInteraction" class="voice-interaction-panel">
        <div class="voice-header">
          <span class="voice-icon">🎤</span>
          <span class="voice-title">语音交互</span>
          <button class="close-voice" @click="toggleVoiceInteraction">✕</button>
        </div>

        <div class="voice-controls">
          <button
            class="voice-btn"
            :class="{ listening: isListening }"
            @click="isListening ? stopListening() : startListening()"
          >
            <span class="voice-btn-icon">{{ isListening ? '⏹️' : '🎤' }}</span>
            <span class="voice-btn-text">{{ isListening ? '停止聆听' : '开始语音' }}</span>
          </button>

          <button
            class="voice-speak-btn"
            @click="speakAIResponse"
            :disabled="!aiVoiceResponse"
          >
            <span class="voice-speak-icon">🔊</span>
            <span class="voice-speak-text">播放回复</span>
          </button>
        </div>

        <div class="voice-feedback">
          <div class="voice-command" v-if="voiceCommand">
            <span class="command-label">指令:</span>
            <span class="command-text">{{ voiceCommand }}</span>
          </div>

          <div class="ai-response" v-if="aiVoiceResponse">
            <span class="response-label">AI回复:</span>
            <span class="response-text">{{ aiVoiceResponse }}</span>
          </div>

          <div class="voice-hints" v-if="!voiceCommand">
            <p>尝试说:</p>
            <ul>
              <li>"显示项目详情"</li>
              <li>"查看进度状态"</li>
              <li>"分析项目风险"</li>
              <li>"生成报告摘要"</li>
            </ul>
          </div>
        </div>
      </div>

      <!-- 操作按钮 -->
      <div class="card-actions">
        <div class="primary-actions">
          <slot name="actions">
            <button
              v-for="action in primaryActions"
              :key="action.id"
              class="action-btn primary"
              @click="handleAction(action)"
              :title="action.description"
            >
              <span class="action-icon">{{ action.icon }}</span>
              <span class="action-label">{{ action.label }}</span>
            </button>
            <button
              class="action-btn primary voice-toggle"
              @click="toggleVoiceInteraction"
              :class="{ active: voiceInteraction }"
              title="语音交互"
            >
              <span class="action-icon">🎤</span>
              <span class="action-label">{{ voiceInteraction ? '关闭语音' : '语音交互' }}</span>
            </button>
          </slot>
        </div>
        <div class="secondary-actions">
          <div class="dropdown-container">
            <button class="more-actions-btn" @click="toggleMoreActions">
              <span class="more-icon">⋯</span>
            </button>
            <div v-if="showMoreActions" class="dropdown-menu">
              <div
                v-for="action in secondaryActions"
                :key="action.id"
                class="dropdown-item"
                @click="handleAction(action)"
              >
                <span class="dropdown-icon">{{ action.icon }}</span>
                <span class="dropdown-label">{{ action.label }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 卡片底部 -->
    <div v-if="$slots.footer" class="card-footer">
      <slot name="footer"></slot>
    </div>
    <div v-else class="card-footer">
      <div class="footer-left">
        <div v-if="createdAt" class="timestamp">
          <span class="time-icon">🕒</span>
          <span class="time-text">{{ formatTime(createdAt) }}</span>
        </div>
        <div v-if="updatedAt" class="timestamp">
          <span class="time-icon">🔄</span>
          <span class="time-text">{{ formatTime(updatedAt) }}</span>
        </div>
      </div>
      <div class="footer-right">
        <div class="interaction-stats">
          <span v-if="viewCount !== undefined" class="stat-item">
            <span class="stat-icon">👁️</span>
            <span class="stat-value">{{ viewCount }}</span>
          </span>
          <span v-if="likeCount !== undefined" class="stat-item">
            <span class="stat-icon">❤️</span>
            <span class="stat-value">{{ likeCount }}</span>
          </span>
          <span v-if="commentCount !== undefined" class="stat-item">
            <span class="stat-icon">💬</span>
            <span class="stat-value">{{ commentCount }}</span>
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'

// Props
interface Props {
  title: string
  subtitle?: string
  description?: string
  icon?: string
  status?: string
  priority?: string
  progress?: number
  tags?: string[]
  metadata?: MetadataItem[]
  createdAt?: Date | string
  updatedAt?: Date | string
  viewCount?: number
  likeCount?: number
  commentCount?: number
  highlight?: boolean
  collapsible?: boolean
  expanded?: boolean
  aiAnalysisEnabled?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  icon: '📋',
  highlight: false,
  collapsible: true,
  expanded: false,
  aiAnalysisEnabled: false
})

// Emits
const emit = defineEmits<{
  action: [action: CardAction]
  tagClick: [tag: string]
  expand: [expanded: boolean]
  click: []
}>()

// 类型定义
interface MetadataItem {
  key: string
  label: string
  value: string | number
}

interface CardAction {
  id: string
  label: string
  icon: string
  description?: string
  type: 'primary' | 'secondary'
}

interface AIInsight {
  id: string
  text: string
  icon: string
}

// 响应式数据
const isExpanded = ref(props.expanded)
const showMoreActions = ref(false)
const voiceInteraction = ref(false)
const isListening = ref(false)
const voiceCommand = ref('')
const aiVoiceResponse = ref('')

// AI分析数据
const aiInsights = ref<AIInsight[]>([
  { id: 'insight-1', text: '项目进度正常', icon: '✅' },
  { id: 'insight-2', text: '需要关注优先级', icon: '⚠️' },
  { id: 'insight-3', text: '建议添加更多标签', icon: '💡' },
  { id: 'insight-4', text: '最近有更新活动', icon: '🔄' }
])

// 操作按钮
const primaryActions = ref<CardAction[]>([
  { id: 'view', label: '查看', icon: '👁️', description: '查看详情', type: 'primary' },
  { id: 'edit', label: '编辑', icon: '✏️', description: '编辑卡片', type: 'primary' },
  { id: 'share', label: '分享', icon: '🔗', description: '分享卡片', type: 'primary' }
])

const secondaryActions = ref<CardAction[]>([
  { id: 'copy', label: '复制', icon: '📋', description: '复制卡片', type: 'secondary' },
  { id: 'archive', label: '归档', icon: '📦', description: '归档卡片', type: 'secondary' },
  { id: 'delete', label: '删除', icon: '🗑️', description: '删除卡片', type: 'secondary' },
  { id: 'analyze', label: '分析', icon: '📊', description: '分析数据', type: 'secondary' }
])

// 计算属性
const cardClasses = computed(() => ({
  'card-highlighted': props.highlight,
  'card-expanded': isExpanded.value,
  'card-collapsed': !isExpanded.value && props.collapsible,
  'ai-enabled': props.aiAnalysisEnabled
}))

// 方法
const toggleExpand = () => {
  if (props.collapsible) {
    isExpanded.value = !isExpanded.value
    emit('expand', isExpanded.value)
  }
}

const handleAction = (action: CardAction) => {
  emit('action', action)
}

const handleTagClick = (tag: string) => {
  emit('tagClick', tag)
}

const toggleMoreActions = () => {
  showMoreActions.value = !showMoreActions.value
}

const getStatusClass = (status: string) => {
  const statusClasses: Record<string, string> = {
    active: 'status-active',
    pending: 'status-pending',
    paused: 'status-paused',
    completed: 'status-completed',
    draft: 'status-draft',
    review: 'status-review',
    inactive: 'status-off',
    disabled: 'status-off',
  }
  return statusClasses[status] || 'status-default'
}

const getStatusText = (status: string) => {
  const statusTexts: Record<string, string> = {
    active: '进行中',
    pending: '待处理',
    paused: '已暂停',
    completed: '已完成',
    draft: '草稿',
    review: '审核中',
    inactive: '停用',
    disabled: '停用',
  }
  return statusTexts[status] || status
}

const getPriorityClass = (priority: string) => {
  const priorityClasses: Record<string, string> = {
    high: 'priority-high',
    medium: 'priority-medium',
    low: 'priority-low',
    critical: 'priority-critical'
  }
  return priorityClasses[priority] || 'priority-default'
}

const getPriorityText = (priority: string) => {
  const priorityTexts: Record<string, string> = {
    high: '高',
    medium: '中',
    low: '低',
    critical: '紧急'
  }
  return priorityTexts[priority] || priority
}

const formatTime = (time: Date | string) => {
  if (!time) return ''
  const date = new Date(time)
  const now = new Date()
  const diff = now.getTime() - date.getTime()

  if (diff < 60000) return '刚刚'
  if (diff < 3600000) return `${Math.floor(diff / 60000)}分钟前`
  if (diff < 86400000) return `${Math.floor(diff / 3600000)}小时前`
  if (diff < 604800000) return `${Math.floor(diff / 86400000)}天前`
  return date.toLocaleDateString('zh-CN')
}

// 语音交互功能
const toggleVoiceInteraction = () => {
  voiceInteraction.value = !voiceInteraction.value
  if (!voiceInteraction.value) {
    stopListening()
    voiceCommand.value = ''
    aiVoiceResponse.value = ''
  }
}

const startListening = () => {
  if (!('webkitSpeechRecognition' in window || 'SpeechRecognition' in window)) {
    aiVoiceResponse.value = '您的浏览器不支持语音识别功能'
    return
  }

  isListening.value = true
  voiceCommand.value = '正在聆听...'
  aiVoiceResponse.value = ''

  // 模拟语音识别
  setTimeout(() => {
    const commands = [
      '显示项目详情',
      '查看进度状态',
      '分析项目风险',
      '生成报告摘要',
      '分享项目信息',
      '设置提醒通知'
    ]

    const randomCommand = commands[Math.floor(Math.random() * commands.length)]
    voiceCommand.value = `识别到: "${randomCommand}"`
    processVoiceCommand(randomCommand)
    isListening.value = false
  }, 2000)
}

const stopListening = () => {
  isListening.value = false
  if (voiceCommand.value === '正在聆听...') {
    voiceCommand.value = '语音识别已取消'
  }
}

const processVoiceCommand = (command: string) => {
  const responses: Record<string, string> = {
    '显示项目详情': `正在显示"${props.title}"的详细信息。这是一个${props.priority}优先级项目，当前进度${props.progress ?? 0}%。`,
    '查看进度状态': `项目进度: ${props.progress ?? 0}%。${(props.progress ?? 0) < 30 ? '进度较慢，需要关注。' : (props.progress ?? 0) < 70 ? '进度正常，继续保持。' : '进度良好，接近完成。'}`,
    '分析项目风险': `风险分析: ${props.priority === '高' ? '高风险项目，需要密切监控。' : '风险可控，按计划进行。'}建议${(props.progress ?? 0) < 50 ? '加快进度' : '保持当前节奏'}。`,
    '生成报告摘要': `报告摘要: "${props.title}" - ${props.description?.substring(0, 100)}... 状态: ${props.status}，优先级: ${props.priority}，进度: ${props.progress ?? 0}%。`,
    '分享项目信息': `已准备分享"${props.title}"项目信息。包含进度、状态和关键指标。`,
    '设置提醒通知': `已为您设置项目提醒。将在重要里程碑时通知您。`
  }

  aiVoiceResponse.value = responses[command] || `已执行命令: ${command}`
}

// 模拟AI语音响应
const speakAIResponse = () => {
  if (!aiVoiceResponse.value) return

  if ('speechSynthesis' in window) {
    const utterance = new SpeechSynthesisUtterance(aiVoiceResponse.value)
    utterance.lang = 'zh-CN'
    utterance.rate = 1.0
    utterance.pitch = 1.0
    speechSynthesis.speak(utterance)
  }
}

// 暴露方法
defineExpose({
  toggleExpand,
  expand: () => { isExpanded.value = true },
  collapse: () => { isExpanded.value = false },
  toggleVoiceInteraction,
  startListening,
  stopListening
})
</script>

<style scoped>
@import '@/styles/theme/neuro-theme.css';

.neuro-card {
  position: relative;
  background: var(--neuro-surface);
  border: 2px solid var(--neuro-primary-30);
  border-radius: var(--neuro-radius-lg);
  padding: var(--neuro-spacing-lg);
  margin-bottom: var(--neuro-spacing-md);
  box-shadow: var(--neuro-shadow-lg);
  backdrop-filter: blur(10px);
  transition: border-color var(--neuro-transition-normal), box-shadow var(--neuro-transition-normal);
  overflow: hidden;
}

.neuro-card:hover {
  border-color: var(--neuro-primary);
  box-shadow: var(--neuro-shadow-xl);
}

.neuro-card.card-highlighted {
  border-color: var(--neuro-error);
  box-shadow: 0 12px 40px rgba(239, 68, 68, 0.25);
}

.neuro-card.ai-enabled {
  border-color: var(--neuro-info);
}

/* 卡片头部 */
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  cursor: pointer;
  user-select: none;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
}

.card-icon {
  position: relative;
}

.icon {
  font-size: 24px;
  color: var(--neuro-primary);
}

.icon-glow {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 40px;
  height: 40px;
  background: radial-gradient(circle, var(--neuro-primary-30), transparent 70%);
  border-radius: 50%;
  display: none;
}

@keyframes pulse {
  0% { opacity: 0.3; }
  50% { opacity: 0.6; }
  100% { opacity: 0.3; }
}

.card-title {
  color: var(--neuro-text);
  margin: 0;
  font-size: 18px;
  font-weight: bold;
}

.card-subtitle {
  color: var(--neuro-text-secondary);
  margin: 4px 0 0 0;
  font-size: 14px;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.card-badges {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.status-badge,
.priority-badge {
  padding: 4px 10px;
  border-radius: var(--neuro-radius-lg);
  font-size: 11px;
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

.status-off {
  background: color-mix(in srgb, var(--nm-bg-elevated) 72%, var(--nm-bg-deep));
  color: var(--nm-text-muted);
  border: 1px solid color-mix(in srgb, var(--nm-border) 90%, transparent);
}

.status-completed {
  background: var(--neuro-info-20);
  color: var(--neuro-info);
  border: 1px solid var(--neuro-info-30);
}

.priority-high {
  background: var(--neuro-error-20);
  color: var(--neuro-error);
  border: 1px solid var(--neuro-error-30);
}

.priority-medium {
  background: var(--neuro-warning-20);
  color: var(--neuro-warning);
  border: 1px solid var(--neuro-warning-30);
}

.priority-low {
  background: var(--neuro-success-20);
  color: var(--neuro-success);
  border: 1px solid var(--neuro-success-30);
}

.tags-count {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  background: var(--neuro-primary-10);
  border: 1px solid var(--neuro-primary-30);
  border-radius: var(--neuro-radius-lg);
  color: var(--neuro-primary);
  font-size: 11px;
}

.tags-icon {
  font-size: 12px;
}

.expand-btn {
  background: transparent;
  border: none;
  color: rgba(255, 255, 255, 0.5);
  cursor: pointer;
  padding: var(--neuro-spacing-xs);
  transition: all 0.2s ease;
}

.expand-btn:hover {
  color: var(--neuro-primary);
}

.expand-btn.expanded .expand-icon {
  transform: rotate(90deg);
}

/* 卡片内容 */
.card-content {
  margin: 16px 0;
}

.card-description {
  color: rgba(255, 255, 255, 0.8);
  line-height: 1.6;
  margin-bottom: 20px;
}

.progress-section {
  margin-bottom: 20px;
}

.progress-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.progress-label {
  color: var(--neuro-text-secondary);
  font-size: 14px;
}

.progress-value {
  color: var(--neuro-primary);
  font-weight: bold;
  font-size: 14px;
}

.progress-bar {
  position: relative;
  height: 8px;
  background: var(--neuro-surface-90);
  border-radius: var(--neuro-radius-sm);
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: linear-gradient(90deg, var(--neuro-primary), var(--neuro-secondary));
  border-radius: var(--neuro-radius-sm);
  transition: width 0.5s ease;
}

.progress-glow {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(90deg, transparent, var(--neuro-surface-90), transparent);
  display: none;
}

@keyframes shimmer {
  0% { transform: translateX(-100%); }
  100% { transform: translateX(100%); }
}

.tags-section {
  margin-bottom: 20px;
}

.tags-label {
  color: var(--neuro-text-secondary);
  font-size: 14px;
  margin-bottom: 8px;
}

.tags-list {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.tag {
  display: inline-block;
  padding: 4px 10px;
  background: var(--neuro-primary-10);
  border: 1px solid var(--neuro-primary-30);
  border-radius: var(--neuro-radius-lg);
  color: var(--neuro-primary);
  font-size: 12px;
  cursor: pointer;
  transition: background 0.2s ease;
}

.tag:hover {
  background: var(--neuro-primary-20);
}

.metadata-section {
  margin-bottom: 20px;
}

.metadata-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
  gap: 12px;
}

.metadata-item {
  display: flex;
  flex-direction: column;
}

.metadata-label {
  color: var(--neuro-text-secondary);
  font-size: 12px;
  margin-bottom: 2px;
}

.metadata-value {
  color: var(--neuro-text);
  font-size: 14px;
  font-weight: 500;
}

/* AI分析部分 */
.ai-analysis-section {
  margin: 20px 0;
  padding: var(--neuro-spacing-md);
  background: var(--neuro-info-10);
  border: 1px solid var(--neuro-info-30);
  border-radius: var(--neuro-radius-lg);
}

.ai-section-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
}

.ai-icon {
  font-size: 18px;
  color: #0096ff;
}

.ai-title {
  color: #0096ff;
  font-weight: bold;
  font-size: 16px;
}

.ai-insights {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.ai-insight {
  display: flex;
  align-items: center;
  gap: 8px;
}

.insight-icon {
  font-size: 16px;
  color: #0096ff;
}

.insight-text {
  color: rgba(255, 255, 255, 0.8);
  font-size: 14px;
}

/* 操作按钮 */
.card-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 20px;
  padding-top: 16px;
  border-top: 1px solid var(--neuro-surface-90);
}

.primary-actions {
  display: flex;
  gap: 8px;
}

.action-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border: none;
  border-radius: var(--neuro-radius-lg);
  font-weight: bold;
  cursor: pointer;
  transition: all var(--neuro-transition-normal);
}

.action-btn.primary {
  background: linear-gradient(135deg, var(--neuro-primary-20), var(--neuro-primary-10));
  border: 1px solid var(--neuro-primary-30);
  color: var(--neuro-primary);
}

.action-btn.primary:hover {
  background: linear-gradient(135deg, var(--neuro-primary-30), var(--neuro-primary-20));
  box-shadow: 0 8px 20px var(--neuro-primary-30);
}

.secondary-actions {
  position: relative;
}

.more-actions-btn {
  background: var(--neuro-surface-90);
  border: 1px solid var(--neuro-surface-90);
  color: var(--neuro-text);
  padding: 8px 12px;
  border-radius: var(--neuro-radius-lg);
  cursor: pointer;
  transition: all var(--neuro-transition-normal);
}

.more-actions-btn:hover {
  background: var(--neuro-surface-90);
}

.dropdown-menu {
  position: absolute;
  top: calc(100% + 8px);
  right: 0;
  min-width: 150px;
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
  gap: 8px;
  padding: 10px 12px;
  border-radius: var(--neuro-radius-md);
  cursor: pointer;
  transition: all 0.2s ease;
}

.dropdown-item:hover {
  background: var(--neuro-primary-10);
}

.dropdown-icon {
  font-size: 16px;
  color: var(--neuro-primary);
}

.dropdown-label {
  color: var(--neuro-text);
  font-size: 14px;
}

/* 卡片底部 */
.card-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid var(--neuro-surface-90);
}

.footer-left {
  display: flex;
  gap: 16px;
}

.timestamp {
  display: flex;
  align-items: center;
  gap: 6px;
  color: rgba(255, 255, 255, 0.5);
  font-size: 12px;
}

.time-icon {
  font-size: 14px;
}

.interaction-stats {
  display: flex;
  gap: 12px;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 4px;
  color: var(--neuro-text-secondary);
  font-size: 12px;
}

.stat-icon {
  font-size: 14px;
}

/* 神经连接效果 */
.neural-connections {
  display: none;
}

.connection-line {
  display: none;
}

@keyframes flow {
  0% { transform: translateX(-100%); }
  100% { transform: translateX(100%); }
}

/* 悬停效果 */
.hover-effect {
  display: none;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .neuro-card {
    padding: var(--neuro-spacing-md);
  }

  .card-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }

  .header-right {
    width: 100%;
    justify-content: space-between;
  }

  .card-badges {
    order: 2;
  }

  .header-actions {
    order: 1;
  }

  .metadata-grid {
    grid-template-columns: 1fr;
    gap: 8px;
  }

  .card-actions {
    flex-direction: column;
    gap: 12px;
  }

  .primary-actions {
    width: 100%;
    justify-content: space-between;
  }

  .action-btn.primary {
    flex: 1;
    justify-content: center;
  }

  .secondary-actions {
    width: 100%;
    display: flex;
    justify-content: center;
  }

  .card-footer {
    flex-direction: column;
    gap: 12px;
    align-items: flex-start;
  }

  .footer-left {
    width: 100%;
    justify-content: space-between;
  }

  .footer-right {
    width: 100%;
    justify-content: space-between;
  }

  .ai-panel-content {
    grid-template-columns: 1fr;
    gap: 12px;
  }
}

@media (max-width: 480px) {
  .neuro-card {
    padding: 12px;
  }

  .header-left {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }

  .card-icon {
    align-self: flex-start;
  }

  .card-badges {
    flex-wrap: wrap;
    justify-content: flex-start;
  }

  .primary-actions {
    flex-direction: column;
  }

  .action-btn.primary {
    width: 100%;
  }

  .tags-list {
    justify-content: flex-start;
  }

  .tag {
    font-size: 11px;
    padding: 3px 8px;
  }

  .interaction-stats {
    width: 100%;
    justify-content: space-between;
  }
}

/* 触摸设备优化 */
@media (hover: none) and (pointer: coarse) {
  .neuro-card:hover {
    border-color: var(--neuro-primary-30);
  }

  .action-btn.primary:hover,
  .tag:hover,
  .more-actions-btn:hover {
    transform: none;
  }

  .dropdown-item:hover {
    background: transparent;
  }

  .dropdown-item:active {
    background: var(--neuro-primary-10);
  }
}

/* 语音交互面板 */
.voice-interaction-panel {
  margin: 20px 0;
  padding: var(--neuro-spacing-md);
  background: var(--neuro-accent-10);
  border: 1px solid var(--neuro-accent-30);
  border-radius: var(--neuro-radius-lg);
  animation: slideDown 0.3s ease;
}

.voice-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 16px;
}

.voice-icon {
  font-size: 20px;
  color: #9c27b0;
}

.voice-title {
  flex: 1;
  color: #9c27b0;
  font-weight: bold;
  font-size: 16px;
}

.close-voice {
  background: transparent;
  border: none;
  color: rgba(255, 255, 255, 0.5);
  cursor: pointer;
  padding: var(--neuro-spacing-xs);
}

.close-voice:hover {
  color: #ff6b6b;
}

.voice-controls {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
}

.voice-btn {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 12px;
  background: rgba(156, 39, 176, 0.2);
  border: 1px solid var(--neuro-accent-30);
  border-radius: var(--neuro-radius-lg);
  color: #9c27b0;
  font-weight: bold;
  cursor: pointer;
  transition: all var(--neuro-transition-normal);
}

.voice-btn:hover {
  background: rgba(156, 39, 176, 0.3);
}

.voice-btn.listening {
  background: rgba(244, 67, 54, 0.2);
  border-color: rgba(244, 67, 54, 0.3);
  color: #f44336;
}

.voice-speak-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 12px 20px;
  background: rgba(33, 150, 243, 0.2);
  border: 1px solid rgba(33, 150, 243, 0.3);
  border-radius: var(--neuro-radius-lg);
  color: #2196f3;
  font-weight: bold;
  cursor: pointer;
  transition: all var(--neuro-transition-normal);
}

.voice-speak-btn:hover:not(:disabled) {
  background: rgba(33, 150, 243, 0.3);
}

.voice-speak-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.voice-feedback {
  background: rgba(0, 0, 0, 0.2);
  border-radius: var(--neuro-radius-md);
  padding: 12px;
}

.voice-command,
.ai-response {
  margin-bottom: 12px;
}

.command-label,
.response-label {
  display: block;
  color: var(--neuro-text-secondary);
  font-size: 12px;
  margin-bottom: 4px;
}

.command-text,
.response-text {
  color: var(--neuro-text);
  font-size: 14px;
  line-height: 1.4;
}

.voice-hints {
  color: rgba(255, 255, 255, 0.7);
  font-size: 13px;
}

.voice-hints p {
  margin: 0 0 8px 0;
  font-weight: bold;
}

.voice-hints ul {
  margin: 0;
  padding-left: 20px;
}

.voice-hints li {
  margin-bottom: 4px;
}

.action-btn.primary.voice-toggle.active {
  background: rgba(156, 39, 176, 0.3);
  border-color: #9c27b0;
  color: #9c27b0;
}
</style>