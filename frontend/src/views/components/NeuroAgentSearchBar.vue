<template>
  <div class="neuro-search-bar" :class="{ 'search-active': searchActive, 'ai-active': aiActive }">
    <!-- 搜索输入区域 -->
    <div class="search-input-container">
      <!-- 搜索图标 -->
      <div class="search-icon">
        <div class="neural-pulse"></div>
        <span class="icon">🔍</span>
      </div>

      <!-- 主搜索输入 -->
      <input
        ref="searchInput"
        v-model="searchQuery"
        type="text"
        class="search-input"
        :placeholder="placeholder"
        @focus="handleFocus"
        @blur="handleBlur"
        @input="handleInput"
        @keyup.enter="handleSearch"
      />

      <!-- 语音搜索按钮 -->
      <button
        class="voice-search-btn"
        :class="{ active: voiceActive }"
        @click="toggleVoiceSearch"
        :title="voiceActive ? '停止语音搜索' : '开始语音搜索'"
      >
        <span class="voice-icon">{{ voiceActive ? '🎤' : '🎤' }}</span>
        <div class="voice-pulse" v-if="voiceActive"></div>
      </button>

      <!-- AI增强按钮 -->
      <button
        class="ai-enhance-btn"
        :class="{ active: aiActive }"
        @click="toggleAI"
        :title="aiActive ? '关闭AI增强' : '开启AI增强'"
      >
        <span class="ai-icon">🧠</span>
        <span class="ai-text">{{ aiActive ? 'AI ON' : 'AI OFF' }}</span>
        <div class="ai-pulse" v-if="aiActive"></div>
      </button>

      <!-- 清除按钮 -->
      <button
        v-if="searchQuery"
        class="clear-btn"
        @click="clearSearch"
        title="清除搜索"
      >
        <span class="clear-icon">✕</span>
      </button>

      <!-- 搜索按钮 -->
      <button class="search-btn" @click="handleSearch" title="搜索">
        <span class="search-btn-icon">🚀</span>
        <span class="search-btn-text">搜索</span>
      </button>
    </div>

    <!-- AI建议面板 -->
    <div v-if="showSuggestions && suggestions.length > 0" class="ai-suggestions-panel">
      <div class="suggestions-header">
        <span class="ai-icon-small">🤖</span>
        <span class="suggestions-title">AI建议搜索</span>
        <button class="close-suggestions" @click="closeSuggestions">✕</button>
      </div>

      <div class="suggestions-list">
        <div
          v-for="(suggestion, index) in suggestions"
          :key="index"
          class="suggestion-item"
          :class="{ highlighted: index === highlightedIndex }"
          @click="selectSuggestion(suggestion)"
          @mouseenter="highlightedIndex = index"
        >
          <div class="suggestion-content">
            <span class="suggestion-icon">{{ suggestion.icon }}</span>
            <div class="suggestion-text">
              <div class="suggestion-main">{{ suggestion.text }}</div>
              <div class="suggestion-desc">{{ suggestion.description }}</div>
            </div>
          </div>
          <div class="suggestion-ai-tag" v-if="suggestion.aiGenerated">
            <span class="ai-tag-icon">⚡</span>
            <span class="ai-tag-text">AI生成</span>
          </div>
        </div>
      </div>

      <!-- AI洞察区域 -->
      <div v-if="aiInsights.length > 0" class="ai-insights-section">
        <div class="insights-header">
          <span class="insights-icon">💡</span>
          <span class="insights-title">AI洞察</span>
        </div>
        <div class="insights-list">
          <div v-for="(insight, index) in aiInsights" :key="index" class="insight-item">
            <span class="insight-icon">⚡</span>
            <span class="insight-text">{{ insight }}</span>
          </div>
        </div>
      </div>

      <!-- 搜索模式检测 -->
      <div v-if="searchPatterns.length > 0" class="search-patterns-section">
        <div class="patterns-header">
          <span class="patterns-icon">🔍</span>
          <span class="patterns-title">搜索模式检测</span>
        </div>
        <div class="patterns-list">
          <div v-for="(pattern, index) in searchPatterns" :key="index" class="pattern-item">
            <span class="pattern-icon">🎯</span>
            <span class="pattern-text">{{ pattern }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 搜索历史 -->
    <div v-if="showHistory && searchHistory.length > 0" class="search-history-panel">
      <div class="history-header">
        <span class="history-icon">🕒</span>
        <span class="history-title">搜索历史</span>
        <button class="clear-history" @click="clearHistory">清除</button>
      </div>

      <div class="history-list">
        <div
          v-for="(item, index) in searchHistory"
          :key="index"
          class="history-item"
          @click="selectHistory(item)"
        >
          <div class="history-content">
            <span class="history-query">{{ item.query }}</span>
            <div class="history-meta">
              <span class="history-time">{{ formatTime(item.timestamp) }}</span>
              <span v-if="item.aiEnhanced" class="history-ai-tag">
                <span class="ai-tag-icon">🤖</span>
                <span class="ai-tag-text">AI增强</span>
              </span>
              <span v-if="item.resultCount" class="history-count">
                <span class="count-icon">📊</span>
                <span class="count-text">{{ item.resultCount }}条</span>
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 搜索状态指示器 -->
    <div class="search-status" v-if="searchStatus">
      <div class="status-indicator" :class="searchStatus.type">
        <span class="status-icon">{{ searchStatus.icon }}</span>
        <span class="status-text">{{ searchStatus.message }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, nextTick } from 'vue'

// Props
interface Props {
  placeholder?: string
  initialQuery?: string
  showHistory?: boolean
  aiEnabled?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  placeholder: '输入关键词进行搜索...',
  initialQuery: '',
  showHistory: true,
  aiEnabled: true
})

// Emits
const emit = defineEmits<{
  search: [query: string, aiEnhanced: boolean]
  clear: []
  aiToggle: [enabled: boolean]
  suggestionSelect: [suggestion: Suggestion]
}>()

// 类型定义
interface Suggestion {
  text: string
  description: string
  icon: string
  aiGenerated: boolean
  confidence?: number
  category?: string
}

interface SearchHistoryItem {
  query: string
  timestamp: number
  aiEnhanced?: boolean
  resultCount?: number
}

interface SearchStatus {
  type: 'loading' | 'success' | 'error' | 'info' | 'warning'
  message: string
  icon: string
}

interface AISearchResult {
  query: string
  enhancedQuery: string
  confidence: number
  categories: string[]
  suggestions: string[]
  predictedResults: number
}

// 响应式数据
const searchQuery = ref(props.initialQuery)
const searchActive = ref(false)
const aiActive = ref(props.aiEnabled)
const showSuggestions = ref(false)
const showHistory = ref(false)
const highlightedIndex = ref(-1)
const searchInput = ref<HTMLInputElement | null>(null)
const voiceActive = ref(false)
const isAnalyzing = ref(false)
const aiInsights = ref<string[]>([])
const searchPatterns = ref<string[]>([])

// 搜索历史
const searchHistory = ref<SearchHistoryItem[]>([
  { query: 'AI项目管理', timestamp: Date.now() - 3600000, aiEnhanced: true, resultCount: 42 },
  { query: 'DevOS 神经', timestamp: Date.now() - 7200000, aiEnhanced: true, resultCount: 28 },
  { query: '智能分析', timestamp: Date.now() - 10800000, aiEnhanced: false, resultCount: 15 },
  { query: '自动化测试', timestamp: Date.now() - 14400000, aiEnhanced: true, resultCount: 37 }
])

// AI建议
const suggestions = ref<Suggestion[]>([
  { text: 'AI项目管理', description: '查找所有AI相关的项目', icon: '📊', aiGenerated: true, confidence: 0.92, category: '项目管理' },
  { text: 'DevOS 神经分析', description: '进行DevOS 神经数据分析', icon: '🧠', aiGenerated: true, confidence: 0.88, category: '数据分析' },
  { text: '智能搜索优化', description: '优化搜索算法和结果', icon: '🔍', aiGenerated: false, confidence: 0.75, category: '技术优化' },
  { text: '自动化测试报告', description: '查看自动化测试结果', icon: '⚡', aiGenerated: false, confidence: 0.81, category: '测试' },
  { text: '数据可视化', description: '创建数据可视化图表', icon: '📈', aiGenerated: true, confidence: 0.95, category: '数据展示' }
])

// 搜索状态
const searchStatus = ref<SearchStatus | null>(null)

// 计算属性
const hasSuggestions = computed(() => suggestions.value.length > 0)
const hasHistory = computed(() => searchHistory.value.length > 0)
const aiInsightCount = computed(() => aiInsights.value.length)
const searchPatternCount = computed(() => searchPatterns.value.length)

// 方法
const handleFocus = () => {
  searchActive.value = true
  if (searchQuery.value) {
    showSuggestions.value = true
  } else {
    showHistory.value = true
  }
}

const handleBlur = () => {
  setTimeout(() => {
    searchActive.value = false
    showSuggestions.value = false
    showHistory.value = false
  }, 200)
}

const handleInput = () => {
  if (searchQuery.value) {
    showSuggestions.value = true
    showHistory.value = false

    // 实时AI分析
    if (aiActive.value) {
      analyzeSearchQuery()
      detectSearchPatterns()
    }

    generateAISuggestions()
  } else {
    showSuggestions.value = false
    showHistory.value = true
    aiInsights.value = []
    searchPatterns.value = []
  }
}

const handleSearch = () => {
  if (!searchQuery.value.trim()) return

  // AI增强搜索
  const enhancedQuery = aiActive.value ? enhanceSearchQuery(searchQuery.value) : searchQuery.value

  // 添加到历史记录
  addToHistory(searchQuery.value, aiActive.value)

  // 触发搜索事件
  emit('search', enhancedQuery, aiActive.value)

  // 显示搜索状态
  showSearchStatus('loading', aiActive.value ? 'AI正在优化搜索...' : '正在搜索中...', '⏳')

  // 模拟AI分析过程
  if (aiActive.value) {
    isAnalyzing.value = true
    simulateAIAnalysis()
  }

  // 模拟搜索延迟
  setTimeout(() => {
    const resultCount = aiActive.value
      ? Math.floor(Math.random() * 150) + 50  // AI增强搜索找到更多结果
      : Math.floor(Math.random() * 80) + 20   // 普通搜索

    showSearchStatus('success',
      aiActive.value
        ? `AI找到 ${resultCount} 条相关结果 (增强模式)`
        : `找到 ${resultCount} 条结果`,
      aiActive.value ? '🤖' : '✅'
    )

    // 3秒后清除状态
    setTimeout(() => {
      searchStatus.value = null
      isAnalyzing.value = false
    }, 3000)
  }, aiActive.value ? 1200 : 800)

  // 关闭建议面板
  showSuggestions.value = false
  showHistory.value = false
}

const clearSearch = () => {
  searchQuery.value = ''
  emit('clear')
  searchInput.value?.focus()
}

const toggleAI = () => {
  aiActive.value = !aiActive.value
  emit('aiToggle', aiActive.value)

  if (aiActive.value) {
    showSearchStatus('info', 'AI增强已开启', '🧠')
  } else {
    showSearchStatus('info', 'AI增强已关闭', '🔌')
  }
}

const selectSuggestion = (suggestion: Suggestion) => {
  searchQuery.value = suggestion.text
  emit('suggestionSelect', suggestion)
  handleSearch()
}

const selectHistory = (item: SearchHistoryItem) => {
  searchQuery.value = item.query
  handleSearch()
}

const closeSuggestions = () => {
  showSuggestions.value = false
}

const clearHistory = () => {
  searchHistory.value = []
  showSearchStatus('info', '搜索历史已清除', '🗑️')
}

const addToHistory = (query: string, aiEnhanced: boolean = false) => {
  // 避免重复
  const existingIndex = searchHistory.value.findIndex(item => item.query === query)
  if (existingIndex > -1) {
    searchHistory.value.splice(existingIndex, 1)
  }

  // 添加到开头
  searchHistory.value.unshift({
    query,
    timestamp: Date.now(),
    aiEnhanced,
    resultCount: aiEnhanced ? Math.floor(Math.random() * 150) + 50 : Math.floor(Math.random() * 80) + 20
  })

  // 限制历史记录数量
  if (searchHistory.value.length > 10) {
    searchHistory.value = searchHistory.value.slice(0, 10)
  }
}

const generateAISuggestions = () => {
  if (!aiActive.value) return

  // 模拟AI生成建议
  const query = searchQuery.value.toLowerCase()
  const newSuggestions: Suggestion[] = []

  if (query.includes('ai') || query.includes('智能')) {
    newSuggestions.push({
      text: 'AI智能分析报告',
      description: '生成AI分析报告',
      icon: '📊',
      aiGenerated: true
    })
  }

  if (query.includes('项目') || query.includes('管理')) {
    newSuggestions.push({
      text: '项目管理仪表板',
      description: '查看项目管理仪表板',
      icon: '📈',
      aiGenerated: true
    })
  }

  if (query.includes('数据') || query.includes('分析')) {
    newSuggestions.push({
      text: '数据可视化分析',
      description: '进行数据可视化分析',
      icon: '📊',
      aiGenerated: true
    })
  }

  if (query.includes('测试') || query.includes('自动化')) {
    newSuggestions.push({
      text: '自动化测试套件',
      description: '运行自动化测试',
      icon: '⚡',
      aiGenerated: true
    })
  }

  // 合并现有建议
  suggestions.value = [...newSuggestions, ...suggestions.value.slice(0, 3)]
}

// AI增强搜索查询
const enhanceSearchQuery = (query: string): string => {
  const enhancements: Record<string, string[]> = {
    'ai': ['人工智能', '机器学习', '深度学习'],
    '项目': ['任务', '计划', '方案'],
    '管理': ['控制', '协调', '组织'],
    '数据': ['信息', '资料', '统计'],
    '分析': ['解析', '研究', '评估'],
    '测试': ['验证', '检查', '调试'],
    '优化': ['改进', '提升', '完善'],
    '系统': ['平台', '框架', '架构']
  }

  let enhanced = query
  Object.entries(enhancements).forEach(([key, synonyms]) => {
    if (query.toLowerCase().includes(key)) {
      const randomSynonym = synonyms[Math.floor(Math.random() * synonyms.length)]
      enhanced = enhanced + ' ' + randomSynonym
    }
  })

  return enhanced.trim()
}

// 分析搜索查询
const analyzeSearchQuery = () => {
  aiInsights.value = []
  const query = searchQuery.value.toLowerCase()

  if (query.includes('ai') || query.includes('智能')) {
    aiInsights.value.push('🔍 检测到AI相关搜索，正在优化机器学习模型匹配...')
    aiInsights.value.push('🧠 建议：尝试搜索"深度学习项目"或"神经网络分析"')
  }

  if (query.includes('项目') || query.includes('管理')) {
    aiInsights.value.push('📊 检测到项目管理搜索，正在加载项目模板...')
    aiInsights.value.push('⚡ 建议：查看"敏捷项目管理"或"项目进度跟踪"')
  }

  if (query.includes('数据') || query.includes('分析')) {
    aiInsights.value.push('📈 检测到数据分析搜索，正在准备可视化工具...')
    aiInsights.value.push('🔬 建议：搜索"数据挖掘"或"统计分析报告"')
  }

  if (query.includes('测试') || query.includes('自动化')) {
    aiInsights.value.push('⚡ 检测到测试相关搜索，正在加载测试用例...')
    aiInsights.value.push('✅ 建议：查看"自动化测试框架"或"性能测试报告"')
  }

  // 通用AI洞察
  if (aiInsights.value.length === 0 && query.length > 3) {
    aiInsights.value.push('🤖 AI正在分析您的搜索意图...')
    aiInsights.value.push('💡 提示：尝试更具体的搜索词以获得更准确的结果')
  }
}

// 检测搜索模式
const detectSearchPatterns = () => {
  searchPatterns.value = []
  const query = searchQuery.value

  // 检测问题模式
  if (query.includes('?') || query.includes('如何') || query.includes('怎么')) {
    searchPatterns.value.push('❓ 问题模式：检测到疑问句，正在准备解答...')
  }

  // 检测比较模式
  if (query.includes('vs') || query.includes('对比') || query.includes('比较')) {
    searchPatterns.value.push('⚖️ 比较模式：检测到对比需求，正在准备对比分析...')
  }

  // 检测技术术语
  if (query.includes('api') || query.includes('sdk') || query.includes('框架')) {
    searchPatterns.value.push('💻 技术模式：检测到技术术语，正在加载技术文档...')
  }

  // 检测时间相关
  if (query.includes('最新') || query.includes('最近') || query.includes('202')) {
    searchPatterns.value.push('🕒 时效模式：检测到时间敏感搜索，正在筛选最新内容...')
  }
}

// 模拟AI分析过程
const simulateAIAnalysis = () => {
  const analysisSteps = [
    '🔍 解析搜索意图...',
    '🧠 应用机器学习模型...',
    '📊 分析相关数据模式...',
    '⚡ 优化搜索结果排序...',
    '✅ 生成增强查询...'
  ]

  let step = 0
  const analysisInterval = setInterval(() => {
    if (step < analysisSteps.length) {
      showSearchStatus('loading', analysisSteps[step], '⏳')
      step++
    } else {
      clearInterval(analysisInterval)
    }
  }, 200)
}

// 语音搜索功能
const toggleVoiceSearch = () => {
  if (voiceActive.value) {
    stopVoiceSearch()
  } else {
    startVoiceSearch()
  }
}

const startVoiceSearch = () => {
  voiceActive.value = true
  showSearchStatus('info', '🎤 正在监听语音输入...', '🎤')

  // 模拟语音识别
  setTimeout(() => {
    const voiceCommands = [
      '搜索AI项目管理',
      '查找DevOS 神经分析',
      '显示数据可视化',
      '打开自动化测试报告'
    ]

    const randomCommand = voiceCommands[Math.floor(Math.random() * voiceCommands.length)]
    searchQuery.value = randomCommand.replace('搜索', '').replace('查找', '').replace('显示', '').replace('打开', '').trim()

    showSearchStatus('success', `🎤 识别到: "${randomCommand}"`, '✅')

    setTimeout(() => {
      voiceActive.value = false
      handleSearch()
    }, 1500)
  }, 2000)
}

const stopVoiceSearch = () => {
  voiceActive.value = false
  showSearchStatus('info', '🎤 语音搜索已停止', '🔇')
}

const showSearchStatus = (type: SearchStatus['type'], message: string, icon: string) => {
  searchStatus.value = { type, message, icon }
}

const formatTime = (timestamp: number) => {
  const now = Date.now()
  const diff = now - timestamp

  if (diff < 60000) return '刚刚'
  if (diff < 3600000) return `${Math.floor(diff / 60000)}分钟前`
  if (diff < 86400000) return `${Math.floor(diff / 3600000)}小时前`
  return `${Math.floor(diff / 86400000)}天前`
}

// 键盘导航
const handleKeyDown = (event: KeyboardEvent) => {
  if (!showSuggestions.value || suggestions.value.length === 0) return

  switch (event.key) {
    case 'ArrowDown':
      event.preventDefault()
      highlightedIndex.value = (highlightedIndex.value + 1) % suggestions.value.length
      break
    case 'ArrowUp':
      event.preventDefault()
      highlightedIndex.value = highlightedIndex.value <= 0
        ? suggestions.value.length - 1
        : highlightedIndex.value - 1
      break
    case 'Enter':
      if (highlightedIndex.value >= 0) {
        event.preventDefault()
        selectSuggestion(suggestions.value[highlightedIndex.value])
      }
      break
    case 'Escape':
      showSuggestions.value = false
      break
  }
}

// 生命周期
onMounted(() => {
  document.addEventListener('keydown', handleKeyDown)
})

// 清理
import { onUnmounted } from 'vue'
onUnmounted(() => {
  document.removeEventListener('keydown', handleKeyDown)
})
</script>

<style scoped>
@import '@/styles/theme/neuro-theme.css';
.neuro-search-bar {
  position: relative;
  width: 100%;
  max-width: 800px;
  margin: 0 auto;
  transition: all var(--neuro-transition-normal);
}

.neuro-search-bar.search-active {
  transform: translateY(-2px);
}

.search-input-container {
  display: flex;
  align-items: center;
  background: linear-gradient(135deg, rgba(26, 26, 46, 0.95), rgba(40, 40, 60, 0.95));
  border: 2px solid var(--neuro-primary-30);
  border-radius: var(--neuro-radius-xl);
  padding: 8px 16px;
  box-shadow: 0 8px 32px rgba(0, 245, 212, 0.15);
  backdrop-filter: blur(10px);
  transition: all var(--neuro-transition-normal);
}

.neuro-search-bar.search-active .search-input-container {
  border-color: rgba(0, 245, 212, 0.6);
  box-shadow: 0 12px 40px rgba(0, 245, 212, 0.25);
}

.search-icon {
  position: relative;
  margin-right: 12px;
}

.neural-pulse {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: var(--neuro-primary-20);
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0% { transform: translate(-50%, -50%) scale(0.8); opacity: 0.5; }
  50% { transform: translate(-50%, -50%) scale(1.2); opacity: 0.8; }
  100% { transform: translate(-50%, -50%) scale(0.8); opacity: 0.5; }
}

.search-icon .icon {
  font-size: 20px;
  color: var(--neuro-primary);
}

.search-input {
  flex: 1;
  background: transparent;
  border: none;
  color: var(--neuro-text);
  font-size: 16px;
  padding: 12px 0;
  outline: none;
  font-family: 'Courier New', monospace;
}

.search-input::placeholder {
  color: rgba(255, 255, 255, 0.5);
}

.ai-enhance-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background: var(--neuro-primary-10);
  border: 1px solid var(--neuro-primary-30);
  border-radius: var(--neuro-radius-lg);
  color: var(--neuro-primary);
  cursor: pointer;
  transition: all var(--neuro-transition-normal);
  position: relative;
  overflow: hidden;
}

.ai-enhance-btn:hover {
  background: var(--neuro-primary-20);
  transform: translateY(-1px);
}

.ai-enhance-btn.active {
  background: var(--neuro-primary-30);
  border-color: var(--neuro-primary);
  box-shadow: 0 0 20px var(--neuro-primary-30);
}

.ai-pulse {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: radial-gradient(circle at center, var(--neuro-primary-30), transparent 70%);
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

.ai-text {
  font-size: 12px;
  font-weight: bold;
  letter-spacing: 1px;
}

.clear-btn {
  background: transparent;
  border: none;
  color: rgba(255, 255, 255, 0.5);
  cursor: pointer;
  padding: var(--neuro-spacing-sm);
  margin: 0 8px;
  transition: all 0.2s ease;
}

.clear-btn:hover {
  color: #ff6b6b;
  transform: scale(1.1);
}

.clear-icon {
  font-size: 18px;
  font-weight: bold;
}

.search-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 24px;
  background: linear-gradient(135deg, var(--neuro-primary), #00b8a9);
  border: none;
  border-radius: var(--neuro-radius-lg);
  color: #1a1a2e;
  font-weight: bold;
  cursor: pointer;
  transition: all var(--neuro-transition-normal);
  margin-left: 8px;
}

.search-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(0, 245, 212, 0.4);
}

.search-btn:active {
  transform: translateY(0);
}

.search-btn-icon {
  font-size: 18px;
}

.search-btn-text {
  font-size: 14px;
  letter-spacing: 1px;
}

/* AI建议面板 */
.ai-suggestions-panel {
  position: absolute;
  top: calc(100% + 8px);
  left: 0;
  right: 0;
  background: var(--neuro-surface);
  border: 2px solid var(--neuro-primary-30);
  border-radius: var(--neuro-radius-xl);
  padding: var(--neuro-spacing-md);
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(20px);
  z-index: 1000;
  animation: slideDown 0.3s ease;
}

@keyframes slideDown {
  from { opacity: 0; transform: translateY(-10px); }
  to { opacity: 1; transform: translateY(0); }
}

.suggestions-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--neuro-surface-90);
}

.ai-icon-small {
  font-size: 18px;
}

.suggestions-title {
  flex: 1;
  color: var(--neuro-primary);
  font-weight: bold;
  font-size: 14px;
  letter-spacing: 1px;
}

.close-suggestions {
  background: transparent;
  border: none;
  color: rgba(255, 255, 255, 0.5);
  cursor: pointer;
  padding: var(--neuro-spacing-xs);
  transition: all 0.2s ease;
}

.close-suggestions:hover {
  color: #ff6b6b;
}

.suggestions-list {
  max-height: 300px;
  overflow-y: auto;
}

.suggestion-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-radius: var(--neuro-radius-lg);
  margin-bottom: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
  background: rgba(255, 255, 255, 0.05);
}

.suggestion-item:hover {
  background: var(--neuro-primary-10);
  transform: translateX(4px);
}

.suggestion-item.highlighted {
  background: rgba(0, 245, 212, 0.15);
  border-left: 4px solid var(--neuro-primary);
}

.suggestion-content {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
}

.suggestion-icon {
  font-size: 20px;
}

.suggestion-text {
  flex: 1;
}

.suggestion-main {
  color: var(--neuro-text);
  font-weight: 500;
  margin-bottom: 4px;
}

.suggestion-desc {
  color: var(--neuro-text-secondary);
  font-size: 12px;
}

.suggestion-ai-tag {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  background: var(--neuro-primary-10);
  border-radius: var(--neuro-radius-md);
  font-size: 11px;
  color: var(--neuro-primary);
}

.ai-tag-icon {
  font-size: 12px;
}

/* AI洞察区域 */
.ai-insights-section {
  margin-top: 20px;
  padding-top: 16px;
  border-top: 1px solid var(--neuro-surface-90);
}

.insights-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
}

.insights-icon {
  font-size: 16px;
  color: var(--neuro-primary);
}

.insights-title {
  color: var(--neuro-primary);
  font-weight: bold;
  font-size: 14px;
  letter-spacing: 1px;
}

.insights-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.insight-item {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 10px 12px;
  background: rgba(0, 245, 212, 0.05);
  border-radius: var(--neuro-radius-md);
  border-left: 3px solid var(--neuro-primary);
  animation: fadeIn 0.3s ease;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateX(-10px); }
  to { opacity: 1; transform: translateX(0); }
}

.insight-icon {
  font-size: 14px;
  color: var(--neuro-primary);
  margin-top: 2px;
}

.insight-text {
  color: rgba(255, 255, 255, 0.8);
  font-size: 13px;
  line-height: 1.4;
  flex: 1;
}

/* 搜索模式检测 */
.search-patterns-section {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid var(--neuro-surface-90);
}

.patterns-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
}

.patterns-icon {
  font-size: 16px;
  color: #0096ff;
}

.patterns-title {
  color: #0096ff;
  font-weight: bold;
  font-size: 14px;
  letter-spacing: 1px;
}

.patterns-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.pattern-item {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 10px 12px;
  background: rgba(0, 150, 255, 0.05);
  border-radius: var(--neuro-radius-md);
  border-left: 3px solid #0096ff;
  animation: fadeIn 0.3s ease 0.1s both;
}

.pattern-icon {
  font-size: 14px;
  color: #0096ff;
  margin-top: 2px;
}

.pattern-text {
  color: rgba(255, 255, 255, 0.8);
  font-size: 13px;
  line-height: 1.4;
  flex: 1;
}

/* 搜索历史面板 */
.search-history-panel {
  position: absolute;
  top: calc(100% + 8px);
  left: 0;
  right: 0;
  background: var(--neuro-surface);
  border: 2px solid var(--neuro-primary-30);
  border-radius: var(--neuro-radius-xl);
  padding: var(--neuro-spacing-md);
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(20px);
  z-index: 1000;
  animation: slideDown 0.3s ease;
}

.history-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--neuro-surface-90);
}

.history-icon {
  font-size: 18px;
  color: var(--neuro-primary);
}

.history-title {
  flex: 1;
  color: var(--neuro-text);
  font-weight: bold;
  font-size: 14px;
}

.clear-history {
  background: transparent;
  border: 1px solid var(--neuro-surface-90);
  border-radius: var(--neuro-radius-md);
  color: var(--neuro-text-secondary);
  padding: 4px 12px;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.clear-history:hover {
  color: #ff6b6b;
  border-color: #ff6b6b;
}

.history-list {
  max-height: 200px;
  overflow-y: auto;
}

.history-item {
  padding: 10px 12px;
  border-radius: var(--neuro-radius-md);
  margin-bottom: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
  background: rgba(255, 255, 255, 0.05);
}

.history-item:hover {
  background: var(--neuro-primary-10);
}

.history-content {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.history-query {
  color: var(--neuro-text);
  font-size: 14px;
  font-weight: 500;
}

.history-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 12px;
}

.history-time {
  color: rgba(255, 255, 255, 0.5);
}

.history-ai-tag {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 2px 6px;
  background: var(--neuro-primary-10);
  border-radius: 6px;
  color: var(--neuro-primary);
}

.ai-tag-icon {
  font-size: 10px;
}

.ai-tag-text {
  font-size: 10px;
  font-weight: bold;
}

.history-count {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 2px 6px;
  background: var(--neuro-info-10);
  border-radius: 6px;
  color: #0096ff;
}

.count-icon {
  font-size: 10px;
}

.count-text {
  font-size: 10px;
  font-weight: bold;
}

/* 语音搜索按钮 */
.voice-search-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  background: var(--neuro-info-10);
  border: 1px solid var(--neuro-info-30);
  border-radius: var(--neuro-radius-lg);
  color: #0096ff;
  cursor: pointer;
  transition: all var(--neuro-transition-normal);
  position: relative;
  overflow: hidden;
  margin-right: 8px;
}

.voice-search-btn:hover {
  background: rgba(0, 150, 255, 0.2);
  transform: translateY(-1px);
}

.voice-search-btn.active {
  background: rgba(0, 150, 255, 0.3);
  border-color: #0096ff;
  box-shadow: 0 0 20px rgba(0, 150, 255, 0.3);
}

.voice-pulse {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: radial-gradient(circle at center, rgba(0, 150, 255, 0.3), transparent 70%);
  animation: voicePulse 1s infinite;
}

@keyframes voicePulse {
  0% { opacity: 0.3; transform: scale(0.9); }
  50% { opacity: 0.6; transform: scale(1.1); }
  100% { opacity: 0.3; transform: scale(0.9); }
}

.voice-icon {
  font-size: 18px;
}

/* 搜索状态 */
.search-status {
  position: absolute;
  top: calc(100% + 8px);
  left: 0;
  right: 0;
  z-index: 1001;
}

.status-indicator {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  border-radius: var(--neuro-radius-lg);
  margin-top: 8px;
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

.status-indicator.warning {
  background: rgba(241, 196, 15, 0.1);
  border: 1px solid var(--neuro-warning-30);
  color: var(--neuro-warning);
}

.status-icon {
  font-size: 16px;
}

.status-text {
  font-size: 14px;
  font-weight: 500;
}

/* 滚动条样式 */
::-webkit-scrollbar {
  width: 6px;
}

::-webkit-scrollbar-track {
  background: rgba(255, 255, 255, 0.05);
  border-radius: 3px;
}

::-webkit-scrollbar-thumb {
  background: var(--neuro-primary-30);
  border-radius: 3px;
}

::-webkit-scrollbar-thumb:hover {
  background: rgba(0, 245, 212, 0.5);
}

/* 响应式设计 */
@media (max-width: 768px) {
  .neuro-search-bar {
    max-width: 100%;
  }

  .search-input-container {
    padding: 6px 12px;
  }

  .search-input {
    font-size: 14px;
    padding: 10px 0;
  }

  .voice-search-btn,
  .ai-enhance-btn {
    width: 36px;
    height: 36px;
    padding: 6px;
  }

  .ai-text {
    display: none;
  }

  .search-btn {
    padding: 8px 16px;
  }

  .search-btn-text {
    display: none;
  }

  .ai-suggestions-panel,
  .search-history-panel {
    position: fixed;
    top: 60px;
    left: 10px;
    right: 10px;
    max-height: 70vh;
    overflow-y: auto;
  }

  .history-meta {
    flex-wrap: wrap;
    gap: 6px;
  }

  .insight-item,
  .pattern-item {
    padding: 8px 10px;
    font-size: 12px;
  }
}

@media (max-width: 480px) {
  .search-input-container {
    flex-wrap: wrap;
  }

  .search-icon {
    margin-right: 8px;
  }

  .neural-pulse {
    width: 24px;
    height: 24px;
  }

  .search-icon .icon {
    font-size: 16px;
  }

  .voice-search-btn,
  .ai-enhance-btn {
    width: 32px;
    height: 32px;
    margin-right: 4px;
  }

  .clear-btn {
    padding: 6px;
    margin: 0 4px;
  }

  .search-btn {
    padding: 6px 12px;
    margin-left: 4px;
  }

  .search-btn-icon {
    font-size: 16px;
  }
}
</style>