<template>
  <div class="card-list-view" :data-theme="theme" :class="{ 'is-empty': filteredCards.length === 0 }">
    <!-- 搜索与过滤区：主行默认单行；扩展筛选用 #toolbar-filters 并折叠 -->
    <div
      class="card-list-toolbar-wrap"
      v-if="showSearch || showFilter || $slots.toolbar || hasToolbarFiltersSlot"
    >
      <div class="card-list-toolbar card-list-toolbar--primary">
        <!-- 搜索框 -->
        <div v-if="showSearch" class="search-wrapper">
          <svg class="search-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="11" cy="11" r="8" />
            <path d="m21 21-4.35-4.35" />
          </svg>
          <input
            v-model="searchText"
            type="text"
            class="search-input"
            :placeholder="searchPlaceholder"
            @input="onSearch"
          />
          <button v-if="searchText" class="search-clear" @click="searchText = ''">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M18 6 6 18M6 6l12 12" />
            </svg>
          </button>
        </div>

        <!-- 状态过滤器 -->
        <div v-if="showFilter" class="filter-wrapper">
          <button
            v-for="opt in filterOptions"
            :key="opt.value"
            class="filter-chip"
            :class="{ 'chip-active': activeFilter === opt.value }"
            @click="activeFilter = opt.value"
          >
            {{ opt.label }}
            <span v-if="opt.count !== undefined" class="chip-count">{{ opt.count }}</span>
          </button>
        </div>

        <div class="toolbar-spacer"></div>

        <!-- 视图切换 -->
        <div class="toolbar-right">
          <div class="view-switcher">
            <button
              class="view-btn"
              :class="{ 'view-btn--active': layout === 'grid' }"
              @click="layout = 'grid'"
              title="网格视图"
            >
              <svg viewBox="0 0 24 24" fill="currentColor" width="18" height="18">
                <rect x="3" y="3" width="7" height="7" rx="1.5" />
                <rect x="14" y="3" width="7" height="7" rx="1.5" />
                <rect x="3" y="14" width="7" height="7" rx="1.5" />
                <rect x="14" y="14" width="7" height="7" rx="1.5" />
              </svg>
            </button>
            <button
              class="view-btn"
              :class="{ 'view-btn--active': layout === 'list' }"
              @click="layout = 'list'"
              title="列表视图"
            >
              <svg viewBox="0 0 24 24" fill="currentColor" width="18" height="18">
                <rect x="3" y="4" width="18" height="4" rx="1.5" />
                <rect x="3" y="10" width="18" height="4" rx="1.5" />
                <rect x="3" y="16" width="18" height="4" rx="1.5" />
              </svg>
            </button>
          </div>

          <slot name="toolbar-extra"></slot>
        </div>

        <!-- 与主行同行的轻量工具（展开入口放在主行末尾） -->
        <slot name="toolbar"></slot>

        <button
          v-if="hasToolbarFiltersSlot"
          type="button"
          class="toolbar-filters-toggle"
          :aria-expanded="filtersExpanded"
          @click="filtersExpanded = !filtersExpanded"
        >
          <span>{{ filtersExpanded ? '收起筛选' : '更多筛选' }}</span>
          <svg
            class="toolbar-filters-toggle__icon"
            viewBox="0 0 24 24"
            width="14"
            height="14"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            :class="{ 'is-open': filtersExpanded }"
          >
            <path d="M6 9l6 6 6-6" />
          </svg>
        </button>
      </div>

      <div v-show="filtersExpanded && hasToolbarFiltersSlot" class="card-list-toolbar card-list-toolbar--filters">
        <slot name="toolbar-filters" />
      </div>
    </div>

    <!-- 结果统计 -->
    <div v-if="filteredCards.length !== cards.length" class="result-count">
      找到 <span class="count-num">{{ filteredCards.length }}</span> 条结果
    </div>

    <!-- 卡片网格 -->
    <div
      v-if="filteredCards.length > 0"
      class="card-grid"
      :class="[`card-grid--${layout}`, `card-grid--${colCount}`]"
    >
      <div
        v-for="(card, index) in filteredCards"
        :key="card.id"
        class="card-grid-item"
        :style="{ animationDelay: `${index * 60}ms` }"
      >
        <slot :card="card" :index="index">
          <!-- 默认使用 NeuroAgentCard -->
          <NeuroAgentCard
            :title="card.title"
            :subtitle="card.subtitle"
            :description="card.description"
            :icon="card.icon"
            :status="card.status"
            :priority="card.priority"
            :progress="card.progress"
            :tags="card.tags"
            :metadata="card.metadata"
            :created-at="card.createdAt"
            :updated-at="card.updatedAt"
            :view-count="card.viewCount"
            :like-count="card.likeCount"
            :comment-count="card.commentCount"
            :highlight="card.highlight"
            :collapsible="card.collapsible"
            :ai-analysis-enabled="card.aiAnalysisEnabled"
            @action="(action) => emit('card-action', action, card)"
            @tag-click="(tag) => emit('tag-click', tag, card)"
            @click="emit('card-click', card)"
          />
        </slot>
      </div>
    </div>

    <!-- 空状态 -->
    <div v-else class="card-list-empty">
      <div class="empty-visual">
        <div class="empty-circle">
          <svg viewBox="0 0 120 120" fill="none" xmlns="http://www.w3.org/2000/svg">
            <circle cx="60" cy="60" r="40" stroke="var(--neuro-primary-30)" stroke-width="2" stroke-dasharray="8 4">
              <animateTransform
                attributeName="transform"
                type="rotate"
                from="0 60 60"
                to="360 60 60"
                dur="20s"
                repeatCount="indefinite"
              />
            </circle>
            <circle cx="60" cy="60" r="25" stroke="var(--neuro-primary-20)" stroke-width="1.5" stroke-dasharray="4 6">
              <animateTransform
                attributeName="transform"
                type="rotate"
                from="360 60 60"
                to="0 60 60"
                dur="15s"
                repeatCount="indefinite"
              />
            </circle>
            <circle cx="60" cy="60" r="4" fill="var(--neuro-primary)" opacity="0.6" />
          </svg>
        </div>
      </div>
      <h3 class="empty-title">{{ emptyTitle }}</h3>
      <p class="empty-desc">{{ emptyDescription }}</p>
      <slot name="empty-action">
        <button v-if="showCreate" class="empty-btn" @click="emit('create')">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="18" height="18">
            <path d="M12 5v14M5 12h14" />
          </svg>
          {{ createLabel }}
        </button>
      </slot>
    </div>

    <!-- 加载状态 -->
    <div v-if="loading" class="card-list-loading">
      <div class="loading-spinner">
        <div class="spinner-ring"></div>
        <div class="spinner-ring"></div>
        <div class="spinner-ring"></div>
      </div>
      <span class="loading-text">{{ loadingText }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, useSlots } from 'vue'
import NeuroAgentCard from './NeuroAgentCard.vue'

// === 类型定义 ===
export interface CardListItem {
  id: string | number
  title: string
  subtitle?: string
  description?: string
  icon?: string
  status?: string
  priority?: string
  progress?: number
  tags?: string[]
  metadata?: Array<{ key: string; label: string; value: string | number }>
  createdAt?: Date | string
  updatedAt?: Date | string
  viewCount?: number
  likeCount?: number
  commentCount?: number
  highlight?: boolean
  collapsible?: boolean
  aiAnalysisEnabled?: boolean
  [key: string]: any
}

export interface FilterOption {
  label: string
  value: string
  count?: number
}

// === Props ===
interface Props {
  cards: CardListItem[]
  /** 搜索字段，默认搜索 title, description, subtitle */
  searchFields?: string[]
  showSearch?: boolean
  searchPlaceholder?: string
  /** 过滤选项 */
  filterOptions?: FilterOption[]
  /** 过滤字段，默认按 status 过滤 */
  filterField?: string
  showFilter?: boolean
  /** 默认布局 */
  defaultLayout?: 'grid' | 'list'
  /** 网格列数 */
  colCount?: 1 | 2 | 3 | 4 | 5
  loading?: boolean
  loadingText?: string
  emptyTitle?: string
  emptyDescription?: string
  showCreate?: boolean
  createLabel?: string
}

const props = withDefaults(defineProps<Props>(), {
  searchFields: () => ['title', 'description', 'subtitle'],
  searchPlaceholder: '搜索...',
  filterField: 'status',
  defaultLayout: 'grid',
  colCount: 3,
  loadingText: '加载中...',
  emptyTitle: '暂无数据',
  emptyDescription: '点击"新建"按钮创建第一条记录',
  showSearch: true,
  showFilter: true,
  showCreate: false,
  createLabel: '新建',
})

// === Emits ===
const emit = defineEmits<{
  search: [value: string]
  filter: [value: string]
  'card-click': [card: CardListItem]
  'card-action': [action: any, card: CardListItem]
  'tag-click': [tag: string, card: CardListItem]
  create: []
  'layout-change': [layout: 'grid' | 'list']
}>()

const slots = useSlots()

/** 多行扩展筛选：使用 #toolbar-filters，默认折叠 */
const filtersExpanded = ref(false)
const hasToolbarFiltersSlot = computed(() => typeof slots['toolbar-filters'] === 'function')

// === 状态 ===
const searchText = ref('')
const activeFilter = ref('')
const layout = ref<'grid' | 'list'>(props.defaultLayout)

// === 过滤逻辑 ===
const filteredCards = computed(() => {
  let result = [...props.cards]

  // 文本搜索
  if (searchText.value.trim()) {
    const q = searchText.value.trim().toLowerCase()
    result = result.filter((card) =>
      props.searchFields.some((field) => {
        const val = card[field]
        return val && String(val).toLowerCase().includes(q)
      })
    )
  }

  // 状态过滤
  if (activeFilter.value) {
    result = result.filter((card) => card[props.filterField] === activeFilter.value)
  }

  return result
})

// === 事件 ===
let searchTimer: ReturnType<typeof setTimeout>
function onSearch() {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => emit('search', searchText.value), 300)
}

const theme = ref<'dark' | 'light'>('dark')

onMounted(() => {
  const layout = document.querySelector('.neuro-command-layout')
  if (layout) {
    theme.value = layout.getAttribute('data-theme') as 'dark' | 'light' || 'dark'
  }
})

watch(layout, (val) => emit('layout-change', val))
</script>

<style scoped>
@import '@/styles/theme/neuro-theme.css';

.card-list-view {
  position: relative;
  width: 100%;
  min-height: 200px;
}

/* === 工具栏：外包一层边框；主行单行（必要时横向滚动）=== */
.card-list-toolbar-wrap {
  margin-bottom: 24px;
  border: 1px solid var(--neuro-primary-10);
  border-radius: var(--neuro-radius-lg);
  background: var(--neuro-surface);
  backdrop-filter: blur(16px);
  overflow: hidden;
}

.card-list-toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  box-sizing: border-box;
}

.card-list-toolbar--primary {
  flex-wrap: nowrap;
  overflow-x: auto;
  scrollbar-width: thin;
  scrollbar-color: rgba(0, 245, 212, 0.2) transparent;
}

.card-list-toolbar--primary::-webkit-scrollbar {
  height: 6px;
}

.card-list-toolbar--primary::-webkit-scrollbar-thumb {
  background: rgba(0, 245, 212, 0.2);
  border-radius: 3px;
}

.card-list-toolbar--filters {
  flex-wrap: wrap;
  align-items: flex-start;
  border-top: 1px solid var(--neuro-primary-10);
  background: color-mix(in srgb, var(--neuro-surface) 88%, var(--neuro-primary-10));
  padding-top: 14px;
  padding-bottom: 14px;
}

.toolbar-filters-toggle {
  display: inline-flex;
  flex-shrink: 0;
  align-items: center;
  gap: 6px;
  padding: 8px 12px;
  margin-left: 4px;
  border: 1px solid var(--neuro-border);
  border-radius: var(--neuro-radius-full);
  background: var(--neuro-surface);
  color: var(--neuro-text-secondary);
  font-size: 13px;
  font-family: var(--neuro-font-sans);
  font-weight: 600;
  cursor: pointer;
  transition: all var(--neuro-transition-fast);
  white-space: nowrap;
}

.toolbar-filters-toggle:hover {
  border-color: var(--neuro-primary-30);
  color: var(--neuro-primary);
}

.toolbar-filters-toggle__icon {
  flex-shrink: 0;
  transition: transform 0.2s ease;
}

.toolbar-filters-toggle__icon.is-open {
  transform: rotate(180deg);
}

.toolbar-spacer {
  flex: 1;
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

/* 搜索框：主行内弹性宽度，避免轻易折行 */
.search-wrapper {
  position: relative;
  flex: 1 1 280px;
  min-width: 160px;
  max-width: 420px;
}

.search-icon {
  position: absolute;
  left: 14px;
  top: 50%;
  transform: translateY(-50%);
  width: 18px;
  height: 18px;
  color: var(--neuro-text-secondary);
  pointer-events: none;
  transition: color var(--neuro-transition-fast);
}

.search-input {
  width: 100%;
  height: 40px;
  padding: 0 40px 0 42px;
  background: var(--neuro-surface);
  border: 1px solid var(--neuro-border);
  border-radius: var(--neuro-radius-lg);
  color: var(--neuro-text);
  font-size: 14px;
  font-family: var(--neuro-font-sans);
  transition: all var(--neuro-transition-fast);
  box-sizing: border-box;
}

.search-input::placeholder {
  color: var(--neuro-text-secondary);
  opacity: 1;
}

.search-input:focus {
  outline: none;
  border-color: var(--neuro-primary);
  box-shadow: 0 0 0 3px var(--neuro-primary-10), 0 0 20px var(--neuro-primary-10);
}

.search-input:focus ~ .search-icon {
  color: var(--neuro-primary);
}

.search-clear {
  position: absolute;
  right: 8px;
  top: 50%;
  transform: translateY(-50%);
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: none;
  color: var(--neuro-text-secondary);
  cursor: pointer;
  border-radius: 50%;
  transition: all var(--neuro-transition-fast);
}

.search-clear:hover {
  color: var(--neuro-text);
  background: var(--neuro-primary-10);
}

.search-clear svg {
  width: 14px;
  height: 14px;
}

/* 过滤芯片：主行内单行展示，超出横向滚动 */
.filter-wrapper {
  display: flex;
  flex-wrap: nowrap;
  flex-shrink: 0;
  align-items: center;
  gap: 6px;
  max-width: min(520px, 42vw);
  overflow-x: auto;
  scrollbar-width: thin;
}

.filter-chip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 14px;
  background: var(--neuro-surface);
  border: 1px solid var(--neuro-border);
  border-radius: var(--neuro-radius-full);
  color: var(--neuro-text-secondary);
  font-size: 13px;
  font-family: var(--neuro-font-sans);
  cursor: pointer;
  transition: all var(--neuro-transition-fast);
  white-space: nowrap;
}

.filter-chip:hover {
  border-color: var(--neuro-primary-30);
  color: var(--neuro-primary);
}

.filter-chip.chip-active {
  background: var(--neuro-primary-20);
  border-color: var(--neuro-primary);
  color: var(--neuro-primary);
  font-weight: 600;
}

.chip-count {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 18px;
  height: 18px;
  padding: 0 5px;
  background: rgba(255, 255, 255, 0.1);
  border-radius: var(--neuro-radius-full);
  font-size: 11px;
  font-weight: 600;
}

.chip-active .chip-count {
  background: rgba(0, 245, 212, 0.2);
}

/* 视图切换 */
.view-switcher {
  display: flex;
  background: var(--neuro-surface);
  border: 1px solid var(--neuro-border);
  border-radius: var(--neuro-radius-md);
  overflow: hidden;
}

.view-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 34px;
  background: transparent;
  border: none;
  color: var(--neuro-text-secondary);
  cursor: pointer;
  transition: all var(--neuro-transition-fast);
}

.view-btn:hover {
  color: var(--neuro-primary);
}

.view-btn.view-btn--active {
  background: var(--neuro-primary-20);
  color: var(--neuro-primary);
}

.view-btn + .view-btn {
  border-left: 1px solid var(--neuro-border);
}

/* 结果统计 */
.result-count {
  margin-bottom: 16px;
  font-size: 13px;
  color: var(--neuro-text-secondary);
}

.count-num {
  color: var(--neuro-primary);
  font-weight: 700;
  font-size: 15px;
}

/* === 网格布局 === */
.card-grid {
  display: grid;
  gap: 16px;
}

/* 网格视图列数 */
.card-grid--grid.card-grid--1 { grid-template-columns: 1fr; }
.card-grid--grid.card-grid--2 { grid-template-columns: repeat(2, 1fr); }
.card-grid--grid.card-grid--3 { grid-template-columns: repeat(3, 1fr); }
.card-grid--grid.card-grid--4 { grid-template-columns: repeat(4, 1fr); }
.card-grid--grid.card-grid--5 { grid-template-columns: repeat(5, 1fr); }

/* 列表视图 */
.card-grid--list {
  grid-template-columns: 1fr;
  gap: 12px;
}

.card-grid--list .card-grid-item :deep(.neuro-card) {
  margin-bottom: 0;
}

/* 卡片入场动画 */
.card-grid-item {
  opacity: 0;
  transform: translateY(16px);
  animation: card-enter 400ms ease forwards;
}

@keyframes card-enter {
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* === 空状态 === */
.card-list-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 80px 20px;
  text-align: center;
}

.empty-visual {
  margin-bottom: 32px;
}

.empty-circle {
  width: 120px;
  height: 120px;
}

.empty-title {
  margin: 0 0 8px;
  font-size: 20px;
  font-weight: 700;
  color: var(--neuro-text);
  font-family: var(--neuro-font-sans);
}

.empty-desc {
  margin: 0 0 24px;
  font-size: 14px;
  color: var(--neuro-text-secondary);
}

.empty-btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 10px 24px;
  background: var(--neuro-gradient-primary);
  border: none;
  border-radius: var(--neuro-radius-lg);
  color: var(--neuro-background);
  font-size: 14px;
  font-weight: 600;
  font-family: var(--neuro-font-sans);
  cursor: pointer;
  transition: all var(--neuro-transition-normal);
}

.empty-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 24px var(--neuro-primary-30);
}

.empty-btn svg {
  width: 18px;
  height: 18px;
}

/* === 加载状态 === */
.card-list-loading {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: var(--neuro-surface-90);
  backdrop-filter: blur(8px);
  border-radius: var(--neuro-radius-lg);
  z-index: 10;
}

.loading-spinner {
  position: relative;
  width: 48px;
  height: 48px;
  margin-bottom: 16px;
}

.spinner-ring {
  position: absolute;
  inset: 0;
  border: 3px solid transparent;
  border-radius: 50%;
}

.spinner-ring:nth-child(1) {
  border-top-color: var(--neuro-primary);
  animation: spin 1s linear infinite;
}

.spinner-ring:nth-child(2) {
  inset: 6px;
  border-right-color: var(--neuro-accent);
  animation: spin 1.5s linear infinite reverse;
}

.spinner-ring:nth-child(3) {
  inset: 12px;
  border-bottom-color: var(--neuro-secondary);
  animation: spin 2s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.loading-text {
  font-size: 13px;
  color: var(--neuro-text-secondary);
}

/* === 响应式 === */
@media (max-width: 1400px) {
  .card-grid--grid.card-grid--5 { grid-template-columns: repeat(4, 1fr); }
}

@media (max-width: 1200px) {
  .card-grid--grid.card-grid--4,
  .card-grid--grid.card-grid--5 { grid-template-columns: repeat(3, 1fr); }
}

@media (max-width: 960px) {
  .card-grid--grid.card-grid--3,
  .card-grid--grid.card-grid--4,
  .card-grid--grid.card-grid--5 { grid-template-columns: repeat(2, 1fr); }

  .card-list-toolbar {
    flex-direction: column;
    align-items: stretch;
  }

  .toolbar-right {
    justify-content: flex-end;
  }
}

@media (max-width: 640px) {
  .card-grid--grid.card-grid--1,
  .card-grid--grid.card-grid--2,
  .card-grid--grid.card-grid--3,
  .card-grid--grid.card-grid--4,
  .card-grid--grid.card-grid--5 { grid-template-columns: 1fr; }

  .card-list-toolbar {
    padding: 12px 16px;
  }

  .search-wrapper {
    max-width: 100%;
  }
}

/* 浅色模式 */
.card-list-view[data-theme="light"] .chip-count {
  background: rgba(0, 0, 0, 0.06);
}
.card-list-view[data-theme="light"] .chip-active .chip-count {
  background: rgba(14, 165, 233, 0.15);
}
</style>
