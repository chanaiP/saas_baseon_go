<template>
  <div class="neuro-agent-list-page" :data-theme="theme">
    <!-- 神经数据流背景 -->
    <div class="neural-data-stream">
      <div v-for="n in 8" :key="n" class="data-stream-line" :style="{
        left: `${(n - 1) * 12.5}%`,
        animationDelay: `${n * 0.2}s`
      }"></div>
    </div>

    <!-- el-table：第一层标题/统计/操作；第二层面板内为搜索 + 表格（与 NeuroAgentPageShell 约定一致） -->
    <div v-if="mode === 'el-table'" class="list-card list-card--layered">
      <div v-if="showCardTopbar" class="list-el-hero">
        <div class="card-topbar">
          <div class="topbar-left">
            <h2 v-if="title?.trim() || showTitleCountBadge" class="card-title">
              {{ title }}
              <span class="title-badge" v-if="showTitleCountBadge">
                <span class="badge-dot"></span>
                {{ dataCount }}
              </span>
            </h2>
            <p class="card-subtitle" v-if="subtitle">{{ subtitle }}</p>
          </div>
          <div v-if="hasTitleTrailingSlot" class="topbar-trailing">
            <slot name="title-trailing" />
          </div>
          <div class="topbar-actions">
            <slot name="actions">
              <button v-if="showCreate" class="card-btn-create" @click="emit('create')">
                <span class="btn-icon">⚡</span>
                {{ createLabel }}
              </button>
            </slot>
          </div>
        </div>
      </div>

      <div
        class="list-el-panel"
        :class="{ 'list-el-panel--flush-bottom': panelTableFlushBottom }"
      >
        <div
          class="card-search"
          v-if="cardSearchVisible"
        >
          <!-- 筛选项在可横向滚动区内；查询/重置紧跟其后，与控件同一行从左连续排列；右侧区仅 #search-trailing-actions -->
          <div class="search-row search-row--primary">
            <div v-if="searchLeadVisible" class="search-row__lead">
              <div class="search-row__lead-scroll">
                <div class="search-input-group" v-if="showKeywordSearch">
                <span class="search-icon">
                  <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2">
                    <circle cx="11" cy="11" r="7"/><path d="M21 21l-4.35-4.35"/>
                  </svg>
                </span>
                <input
                  class="search-input"
                  :placeholder="searchPlaceholder"
                  :value="localKeyword"
                  @input="onKeywordInput"
                  @keyup.enter="handleSearch"
                />
                <button class="search-clear" v-if="localKeyword" @click="clearKeyword" title="清除">
                  <svg viewBox="0 0 24 24" width="12" height="12" fill="none" stroke="currentColor" stroke-width="2.5">
                    <path d="M18 6L6 18M6 6l12 12"/>
                  </svg>
                </button>
              </div>
              <slot name="search"></slot>
              <template v-for="f in primaryFilterFields" :key="f.key">
                <div class="filter-item" :class="'filter-item--' + f.type">
                  <el-input
                    v-if="f.type === 'text'"
                    v-model="filterValues[f.key]"
                    :placeholder="filterControlHint(f)"
                    :aria-label="filterControlHint(f)"
                    clearable
                    size="default"
                    @keyup.enter="handleSearch"
                  />
                  <el-input-number
                    v-else-if="f.type === 'number'"
                    v-model="filterValues[f.key]"
                    :placeholder="filterControlHint(f)"
                    :aria-label="filterControlHint(f)"
                    size="default"
                  />
                  <el-select
                    v-else-if="f.type === 'select'"
                    v-model="filterValues[f.key]"
                    :placeholder="filterControlHint(f)"
                    :aria-label="filterControlHint(f)"
                    clearable
                    size="default"
                  >
                    <el-option v-for="opt in f.options" :key="f.key + '-' + String(opt.value) + '-' + opt.label" :label="opt.label" :value="opt.value" />
                  </el-select>
                  <el-date-picker
                    v-else-if="f.type === 'date'"
                    v-model="filterValues[f.key]"
                    type="date"
                    :placeholder="filterControlHint(f)"
                    :aria-label="filterControlHint(f)"
                    value-format="YYYY-MM-DD"
                    size="default"
                  />
                  <el-date-picker
                    v-else-if="f.type === 'daterange'"
                    v-model="filterValues[f.key]"
                    type="daterange"
                    range-separator="至"
                    :start-placeholder="filterDateRangeStartHint(f)"
                    :end-placeholder="filterDateRangeEndHint(f)"
                    :aria-label="filterControlHint(f)"
                    value-format="YYYY-MM-DD"
                    size="default"
                  />
                  <el-switch v-else-if="f.type === 'switch'" v-model="filterValues[f.key]" :aria-label="filterControlHint(f)" />
                </div>
              </template>
              <button
                v-if="showFiltersMoreToggle"
                type="button"
                class="search-filters-toggle"
                :aria-expanded="filtersExpanded"
                @click="filtersExpanded = !filtersExpanded"
              >
                <span>{{ filtersExpanded ? '收起筛选' : '更多筛选' }}</span>
                <svg
                  class="search-filters-toggle__icon"
                  viewBox="0 0 24 24"
                  width="14"
                  height="14"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                  :class="{ 'is-open': filtersExpanded }"
                >
                  <path d="M6 9l6 6 6-6"/>
                </svg>
              </button>
              </div>
              <div class="search-btns search-btns--primary">
                <button type="button" class="btn-search" @click.stop="handleSearch">查询</button>
                <button type="button" class="btn-reset" @click.stop="handleReset" v-if="hasActiveFilters">重置</button>
              </div>
            </div>
            <div
              v-if="hasSearchTrailingSlot"
              class="search-row__trailing"
              :class="{ 'search-row__trailing--solo': !searchLeadVisible }"
            >
              <slot name="search-trailing-actions" />
            </div>
          </div>
          <div v-show="filtersExpanded && extraFilterFields.length > 0" class="card-search-extra">
            <div class="search-row search-row--extra">
              <template v-for="f in extraFilterFields" :key="f.key">
                <div class="filter-item" :class="'filter-item--' + f.type">
                  <el-input
                    v-if="f.type === 'text'"
                    v-model="filterValues[f.key]"
                    :placeholder="filterControlHint(f)"
                    :aria-label="filterControlHint(f)"
                    clearable
                    size="default"
                    @keyup.enter="handleSearch"
                  />
                  <el-input-number
                    v-else-if="f.type === 'number'"
                    v-model="filterValues[f.key]"
                    :placeholder="filterControlHint(f)"
                    :aria-label="filterControlHint(f)"
                    size="default"
                  />
                  <el-select
                    v-else-if="f.type === 'select'"
                    v-model="filterValues[f.key]"
                    :placeholder="filterControlHint(f)"
                    :aria-label="filterControlHint(f)"
                    clearable
                    size="default"
                  >
                    <el-option v-for="opt in f.options" :key="f.key + '-' + String(opt.value) + '-' + opt.label" :label="opt.label" :value="opt.value" />
                  </el-select>
                  <el-date-picker
                    v-else-if="f.type === 'date'"
                    v-model="filterValues[f.key]"
                    type="date"
                    :placeholder="filterControlHint(f)"
                    :aria-label="filterControlHint(f)"
                    value-format="YYYY-MM-DD"
                    size="default"
                  />
                  <el-date-picker
                    v-else-if="f.type === 'daterange'"
                    v-model="filterValues[f.key]"
                    type="daterange"
                    range-separator="至"
                    :start-placeholder="filterDateRangeStartHint(f)"
                    :end-placeholder="filterDateRangeEndHint(f)"
                    :aria-label="filterControlHint(f)"
                    value-format="YYYY-MM-DD"
                    size="default"
                  />
                  <el-switch v-else-if="f.type === 'switch'" v-model="filterValues[f.key]" :aria-label="filterControlHint(f)" />
                </div>
              </template>
            </div>
          </div>
        </div>

        <div class="card-toolbar" v-if="$slots.toolbar || hasFilters">
          <slot name="toolbar"></slot>
        </div>

        <div ref="cardTableEl" class="card-table">
        <el-table
          ref="tableRef"
          v-loading="loading"
          :data="paginatedData"
          :row-key="rowKey"
          :tree-props="treeProps"
          :default-expand-all="defaultExpandAll"
          :expand-row-keys="expandedRowKeys"
          :row-class-name="rowClassName"
          highlight-current-row
          class="neuro-el-table"
          native-scrollbar
          @expand-change="onExpandChange"
          @row-click="onRowClick"
          @selection-change="onSelectionChange"
        >
          <el-table-column v-if="showSelection" type="selection" width="55" fixed="left" />
          <el-table-column v-if="showIndex" type="index" :label="indexLabel" width="60" align="center" />
          <template v-for="col in columns" :key="col.key">
            <el-table-column
              v-if="!col.hidden"
              :prop="col.key"
              :label="col.title"
              :width="col.width"
              :min-width="col.minWidth"
              :align="col.align || 'left'"
              :fixed="col.fixed"
              :sortable="col.sortable ? 'custom' : false"
              :show-overflow-tooltip="col.tooltip !== false"
            >
            <template #default="{ row }">
              <slot :name="`col-${col.key}`" :row="row" :value="row[col.key]">
                <template v-if="col.type === 'switch'">
                  <el-switch :model-value="row[col.key]" @change="(v: boolean) => emit('cell-change', col.key, row, v)" @click.stop size="small" />
                </template>
                <template v-else-if="col.type === 'tag'">
                  <template v-if="col.tagMap">
                    <el-tag :type="getTagMapType(col.tagMap[row[col.key]])" size="small" effect="plain">
                      {{ getTagMapLabel(col.tagMap[row[col.key]]) }}
                    </el-tag>
                  </template>
                  <template v-else>{{ row[col.key] }}</template>
                </template>
                <template v-else-if="col.type === 'status'">{{ getStatusText(row[col.key]) }}</template>
                <template v-else-if="col.type === 'date'">{{ formatDate(row[col.key]) }}</template>
                <template v-else-if="col.type === 'action'">
                  <div class="neuro-action-col">
                    <template v-for="act in getActionsForRow(row)" :key="act.key">
                      <el-button v-if="!act.hidden" :type="act.type === 'danger' ? 'danger' : 'default'" size="small" :disabled="act.disabled" @click.stop="act.onClick(row)">{{ act.label }}</el-button>
                    </template>
                  </div>
                </template>
                <template v-else>{{ col.formatter ? col.formatter(row[col.key], row) : row[col.key] ?? '' }}</template>
              </slot>
            </template>
          </el-table-column>
          </template>
          <template #empty>
            <div class="neuro-empty"><p>暂无数据</p></div>
          </template>
        </el-table>
      </div>

      <!-- 分页 -->
      <div class="card-pagination" v-if="showPagination && totalPages > 1">
        <div class="pagination-size" @click="sizeDropdownOpen = !sizeDropdownOpen" tabindex="0" @blur="sizeDropdownOpen = false">
          <span class="size-text">每页 {{ pageSize }} 条</span>
          <span class="size-arrow" :class="{ 'arrow-up': sizeDropdownOpen }">▾</span>
          <div class="size-dropdown" v-if="sizeDropdownOpen">
            <div v-for="size in pageSizes" :key="size" class="size-option" :class="{ 'option-active': size === pageSize }" @click="pageSize = size; currentPage = 1; sizeDropdownOpen = false">{{ size }} 条</div>
          </div>
        </div>
        <button class="pagination-btn pagination-prev" :disabled="currentPage === 1" @click="goToPage(currentPage - 1)">← 上一页</button>
        <div class="pagination-pages">
          <button v-for="page in visiblePages" :key="page" class="pagination-page" :class="{ 'page-active': page === currentPage }" @click="goToPage(page)">{{ page }}</button>
          <span class="pagination-ellipsis" v-if="showStartEllipsis">...</span>
          <span class="pagination-ellipsis" v-if="showEndEllipsis">...</span>
        </div>
        <button class="pagination-btn pagination-next" :disabled="currentPage === totalPages" @click="goToPage(currentPage + 1)">下一页 →</button>
        <div class="pagination-info"><span class="info-total">总计 {{ totalItems }} 条</span></div>
      </div>
      </div>
    </div>

    <!-- 原生神经表格模式（保持原有） -->
    <div v-else class="neuro-data-table">
      <div class="table-header" v-if="showSelection || selectedItems.length > 0">
        <div class="table-actions">
          <label class="neuro-checkbox" v-if="showSelection">
            <input type="checkbox" v-model="selectAll" @change="toggleSelectAll" />
            <span class="checkbox-custom"></span>
            <span class="checkbox-label">全选</span>
          </label>
          <div class="selected-actions" v-if="selectedItems.length > 0">
            <span class="selected-count">已选择 {{ selectedItems.length }} 项</span>
            <button class="action-btn action-delete" @click="handleSelectedDelete">删除</button>
            <button class="action-btn action-export" @click="handleSelectedExport">导出</button>
          </div>
        </div>
      </div>
      <div class="table-container">
        <div class="table-scroll">
          <table class="neuro-table">
            <thead>
              <tr>
                <th class="table-col-select" v-if="showSelection">
                  <label class="neuro-checkbox">
                    <input type="checkbox" v-model="selectAll" />
                    <span class="checkbox-custom"></span>
                  </label>
                </th>
                <th v-for="column in columns" :key="column.key" :class="['table-col-' + column.key, column.sortable ? 'sortable' : '']" @click="column.sortable ? sortByColumn(column.key) : null">
                  <span class="column-header">
                    <span class="column-title">{{ column.title }}</span>
                    <span class="sort-indicator" v-if="sortField === column.key">{{ sortDirection === 'asc' ? '↑' : '↓' }}</span>
                  </span>
                </th>
                <th class="table-col-actions">操作</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="item in paginatedData" :key="resolveRowKey(item)" :class="['table-row', { 'row-selected': selectedItems.includes(resolveRowKey(item)) }]">
                <td class="table-col-select" v-if="showSelection">
                  <label class="neuro-checkbox">
                    <input type="checkbox" :checked="selectedItems.includes(resolveRowKey(item))" @change="toggleSelectItem(resolveRowKey(item))" />
                    <span class="checkbox-custom"></span>
                  </label>
                </td>
                <td v-for="column in columns" :key="column.key" :class="'table-col-' + column.key">
                  <div class="cell-content">
                    <span v-if="column.type === 'text'" class="cell-text">{{ item[column.key] }}</span>
                    <span v-else-if="column.type === 'status'" class="cell-status">
                      <span class="status-dot" :class="'status-' + item[column.key]"></span>
                      <span class="status-text">{{ getStatusText(item[column.key]) }}</span>
                    </span>
                    <span v-else-if="column.type === 'date'" class="cell-date">{{ formatDate(item[column.key]) }}</span>
                    <span v-else-if="column.type === 'tags'" class="cell-tags">
                      <span v-for="tag in item[column.key]" :key="tag" class="data-tag">{{ tag }}</span>
                    </span>
                    <span v-else-if="column.type === 'progress'" class="cell-progress">
                      <div class="progress-bar"><div class="progress-fill" :style="{ width: item[column.key] + '%' }"></div></div>
                      <span class="progress-text">{{ item[column.key] }}%</span>
                    </span>
                    <span v-else class="cell-text">{{ item[column.key] }}</span>
                  </div>
                </td>
                <td class="table-col-actions">
                  <div class="action-buttons">
                    <button class="action-btn action-view" @click="handleView(item)"><span class="btn-label">查看</span></button>
                    <button class="action-btn action-edit" @click="handleEdit(item)"><span class="btn-label">编辑</span></button>
                    <button class="action-btn action-delete" @click="handleDelete(item)"><span class="btn-label">删除</span></button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        <div class="table-empty" v-if="paginatedData.length === 0">
          <div class="empty-neuron"><div class="neuron-animation"></div></div>
          <h3 class="empty-title">暂无数据</h3>
          <p class="empty-desc">点击"新建"按钮创建第一条记录</p>
          <button v-if="showCreate" class="neuro-btn neuro-btn-primary" @click="emit('create')">
            <span class="btn-icon">⚡</span>
            <span class="btn-label">{{ createLabel }}</span>
          </button>
        </div>
      </div>
      <div class="neuro-pagination" v-if="showPagination && totalPages > 1">
        <div class="pagination-size" @click="sizeDropdownOpen = !sizeDropdownOpen" tabindex="0" @blur="sizeDropdownOpen = false">
          <span class="size-text">每页 {{ pageSize }} 条</span>
          <span class="size-arrow" :class="{ 'arrow-up': sizeDropdownOpen }">▾</span>
          <div class="size-dropdown" v-if="sizeDropdownOpen">
            <div v-for="size in pageSizes" :key="size" class="size-option" :class="{ 'option-active': size === pageSize }" @click="pageSize = size; currentPage = 1; sizeDropdownOpen = false">{{ size }} 条</div>
          </div>
        </div>
        <button class="pagination-btn pagination-prev" :disabled="currentPage === 1" @click="goToPage(currentPage - 1)">← 上一页</button>
        <div class="pagination-pages">
          <button v-for="page in visiblePages" :key="page" class="pagination-page" :class="{ 'page-active': page === currentPage }" @click="goToPage(page)">{{ page }}</button>
          <span class="pagination-ellipsis" v-if="showStartEllipsis">...</span>
          <span class="pagination-ellipsis" v-if="showEndEllipsis">...</span>
        </div>
        <button class="pagination-btn pagination-next" :disabled="currentPage === totalPages" @click="goToPage(currentPage + 1)">下一页 →</button>
        <div class="pagination-info"><span class="info-total">总计 {{ totalItems }} 条</span></div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, reactive, watch, nextTick, useSlots } from 'vue'

export interface TableColumn {
  key: string
  title: string
  type?: 'text' | 'status' | 'date' | 'tags' | 'progress' | 'switch' | 'tag' | 'action'
  width?: number | string
  minWidth?: number | string
  align?: 'left' | 'center' | 'right'
  fixed?: boolean | 'left' | 'right'
  sortable?: boolean
  tooltip?: boolean
  hidden?: boolean
  formatter?: (value: any, row: any) => string
  tagMap?: Record<string, string | { label: string; type: string }>
  actions?: ActionConfig[]
  icon?: string
}

export interface ActionConfig {
  key: string
  label: string
  type?: 'primary' | 'success' | 'warning' | 'danger' | 'info'
  disabled?: boolean
  hidden?: boolean | ((row: any) => boolean)
  onClick: (row: any) => void
}

export interface FilterField {
  key: string
  /** 不在左侧单独展示；与 `placeholder` 合并写入控件内提示（占位/范围起止/aria-label） */
  label: string
  type: 'text' | 'number' | 'select' | 'date' | 'daterange' | 'switch'
  placeholder?: string
  options?: { label: string; value: any }[]
}

/** 主行/更多筛选：无独立标签，将字段名并入控件描述 */
function filterControlHint(f: FilterField): string {
  const label = (f.label || '').trim()
  const ph = (f.placeholder || '').trim()
  if (!label) return ph
  if (!ph) {
    if (f.type === 'select' || f.type === 'date' || f.type === 'daterange') return `请选择${label}`
    if (f.type === 'switch') return label
    return `请输入${label}`
  }
  return `${label}（${ph}）`
}

function filterDateRangeStartHint(f: FilterField): string {
  const t = (f.label || '').trim()
  return t ? `${t}开始` : '开始'
}

function filterDateRangeEndHint(f: FilterField): string {
  const t = (f.label || '').trim()
  return t ? `${t}结束` : '结束'
}

interface Props {
  mode?: 'neuro' | 'el-table'
  title?: string
  subtitle?: string
  columns?: TableColumn[]
  data?: any[]
  rowKey?: string | ((row: any) => string)
  pageSize?: number
  pageSizes?: number[]
  showPagination?: boolean
  showSelection?: boolean
  showIndex?: boolean
  showCreate?: boolean
  createLabel?: string
  indexLabel?: string
  treeProps?: { children?: string; hasChildren?: string }
  defaultExpandAll?: boolean
  expandedRowKeys?: string[]
  rowClassName?: (data: { row: any; rowIndex: number }) => string
  statusMap?: Record<string, string>
  loading?: boolean
  hasFilters?: boolean
  total?: number
  showSearch?: boolean
  filterFields?: FilterField[]
  searchPlaceholder?: string
  keyword?: string
  /** 服务端分页当前页；父组件重置页码（如查询）时传入以同步底部分页器。 */
  page?: number
  showKeywordSearch?: boolean
  /** 与 `rowKey` 对应的主键值；设置后与表格数据联动高亮当前行（用于主从联动）。 */
  currentRowId?: number | string | null
  /** 为 true 时不对 `data` 做客户端排序（树形/预排序数据可避免无效 sort 与卡顿）。 */
  skipClientSort?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  mode: 'el-table',
  title: '数据列表',
  subtitle: '',
  columns: () => [],
  data: () => [],
  rowKey: 'id',
  pageSize: 10,
  pageSizes: () => [10, 20, 50],
  showPagination: true,
  showSelection: false,
  showIndex: false,
  showCreate: true,
  createLabel: '新建',
  indexLabel: '#',
  treeProps: () => ({ children: 'children' }),
  defaultExpandAll: false,
  expandedRowKeys: () => [],
  statusMap: () => ({
    active: '进行中', enabled: '启用', disabled: '停用',
    pending: '待审核', paused: '已暂停', completed: '已完成',
  }),
  loading: false,
  hasFilters: false,
  showSearch: true,
  filterFields: () => [],
  searchPlaceholder: '搜索关键词...',
  keyword: '',
  showKeywordSearch: false,
  currentRowId: undefined,
  skipClientSort: false,
})

const slots = useSlots()

/** 右侧操作区：仅 #search-trailing-actions（排序、视图、展开收起等，由业务按需实现） */
const hasSearchTrailingSlot = computed(() => typeof slots['search-trailing-actions'] === 'function')

/** 标题与主操作按钮之间：树表展开收起、辅助操作等 */
const hasTitleTrailingSlot = computed(() => typeof slots['title-trailing'] === 'function')

const cardSearchVisible = computed(
  () =>
    hasSearchTrailingSlot.value ||
    (props.showSearch &&
      (Boolean(slots.search) || props.filterFields.length > 0 || props.showKeywordSearch)),
)

/** 左侧搜索带：关键词、#search、主行筛选项、更多筛选、查询/重置 */
const searchLeadVisible = computed(
  () =>
    props.showSearch &&
    (props.showKeywordSearch || Boolean(slots.search) || props.filterFields.length > 0),
)

/** ElTable expose：文档滚动模式下树展开/收起后需强制同步布局与滚动容器高度，否则会残留「内层纵轴滚动条」。 */
type ElTableListExpose = {
  setCurrentRow: (row: any) => void
  doLayout?: () => void
  /** useScrollbar 的 ref，在部分版本上为 Ref，需解包后再调子组件 `update` */
  scrollBarRef?: { update?: () => void } | { value?: { update?: () => void } }
}

const tableRef = ref<ElTableListExpose | null>(null)
/** 表格格子容器（用于复位误产生的 scrollTop / 外层 scrollport）。 */
const cardTableEl = ref<HTMLElement | null>(null)

function resetNeuroElTableSubtreeScrollLocks() {
  const wrap = cardTableEl.value
  if (wrap) wrap.scrollTop = 0

  const inst = tableRef.value as (ElTableListExpose & { $el?: HTMLElement }) | null
  const root = inst?.$el
  if (!(root instanceof HTMLElement)) return
  root.querySelectorAll<HTMLElement>('.el-scrollbar__wrap, .el-table__body-wrapper').forEach((el) => {
    el.scrollTop = 0
  })
}

function callScrollbarUpdate(inst: ElTableListExpose | null) {
  if (!inst?.scrollBarRef) return
  const raw = inst.scrollBarRef as { update?: () => void; value?: { update?: () => void } }
  const comp = raw && typeof raw === 'object' && 'value' in raw ? raw.value : raw
  comp?.update?.()
}

/** 树表展开/收起后 EP 内置 scrollbar / layout 一拍延迟后仍会认为 body 溢出，需在 DOM 稳定后重做布局与 bar 测算。 */
function syncElTableLayoutAfterTreeToggle() {
  nextTick(() => {
    requestAnimationFrame(() => {
      requestAnimationFrame(() => {
        resetNeuroElTableSubtreeScrollLocks()
        const inst = tableRef.value
        inst?.doLayout?.()
        callScrollbarUpdate(inst)
        resetNeuroElTableSubtreeScrollLocks()
        requestAnimationFrame(() => {
          resetNeuroElTableSubtreeScrollLocks()
          inst?.doLayout?.()
          callScrollbarUpdate(inst)
        })
      })
    })
  })
}

const resolveRowKey = (row: any): string => {
  if (typeof props.rowKey === 'function') {
    return String(props.rowKey(row))
  }
  const key = props.rowKey as string
  const v = row[key]
  return v != null ? String(v) : ''
}

watch(
  () => [props.currentRowId, props.data] as const,
  async ([id, data]) => {
    await nextTick()
    const t = tableRef.value
    if (!t?.setCurrentRow) return
    if (id === undefined || id === null || id === '') {
      t.setCurrentRow(undefined)
      return
    }
    const row = (data || []).find((r: any) => String(resolveRowKey(r)) === String(id))
    t.setCurrentRow(row ?? undefined)
  },
  { flush: 'post' },
)

watch(
  () => (props.expandedRowKeys?.length ? JSON.stringify([...props.expandedRowKeys].sort()) : ''),
  () => syncElTableLayoutAfterTreeToggle(),
  { flush: 'post' },
)

const emit = defineEmits<{
  create: []
  view: [item: any]
  edit: [item: any]
  delete: [item: any]
  export: []
  import: []
  'batch-delete': [ids: string[]]
  'selected-export': [ids: string[]]
  'cell-change': [key: string, row: any, value: any]
  'expand-change': [row: any, expandedRows: any[]]
  'row-click': [row: any]
  'sort-change': [field: string, direction: string]
  'page-size-change': [size: number]
  'page-change': [page: number]
  search: [values: { keyword: string; filters: Record<string, any> }]
  reset: []
}>()

const theme = ref<'dark' | 'light'>('dark')
const sortField = ref('createdAt')
const sortDirection = ref<'asc' | 'desc'>('desc')
const selectAll = ref(false)
const selectedItems = ref<string[]>([])
const currentPage = ref(1)
watch(
  () => props.page,
  (p) => {
    if (p == null || !Number.isFinite(Number(p))) return
    const n = Number(p)
    if (n >= 1 && n !== currentPage.value) currentPage.value = n
  },
)
const localPageSize = ref(props.pageSize)
const pageSize = computed({ get: () => localPageSize.value, set: (v) => { localPageSize.value = v; emit('page-size-change', v) } })
const sizeDropdownOpen = ref(false)
const isServerPagination = computed(() => props.total !== undefined)
const totalItems = computed(() => isServerPagination.value ? (props.total ?? 0) : filteredData.value.length)
const totalPages = computed(() => Math.ceil(totalItems.value / pageSize.value))

/** 底部分页未出现时，表格区域为面板最底层，需与 .list-el-panel 大圆角衔接，避免底边/固定列在圆角处断裂 */
const panelTableFlushBottom = computed(() => !props.showPagination || totalPages.value <= 1)

const filteredData = computed(() => {
  if (props.skipClientSort) {
    return props.data ?? []
  }
  const result = [...(props.data ?? [])]
  result.sort((a, b) => {
    const aVal = a[sortField.value]
    const bVal = b[sortField.value]
    if (sortDirection.value === 'asc') return aVal < bVal ? -1 : aVal > bVal ? 1 : 0
    return aVal > bVal ? -1 : aVal < bVal ? 1 : 0
  })
  return result
})

const paginatedData = computed(() => {
  if (isServerPagination.value) return props.data
  if (!props.showPagination) return filteredData.value
  const start = (currentPage.value - 1) * pageSize.value
  return filteredData.value.slice(start, start + pageSize.value)
})

const visiblePages = computed(() => {
  const pages: number[] = []
  const maxVisible = 5
  if (totalPages.value <= maxVisible) {
    for (let i = 1; i <= totalPages.value; i++) pages.push(i)
  } else {
    let start = Math.max(1, currentPage.value - 2)
    let end = Math.min(totalPages.value, start + maxVisible - 1)
    if (end - start + 1 < maxVisible) start = Math.max(1, end - maxVisible + 1)
    for (let i = start; i <= end; i++) pages.push(i)
  }
  return pages
})
const showStartEllipsis = computed(() => visiblePages.value[0] > 1)
const showEndEllipsis = computed(() => visiblePages.value[visiblePages.value.length - 1] < totalPages.value)
const dataCount = computed(() => props.data.length)
/** 无标题时不展示仅含数字的角标（避免界面只剩「• 1」） */
const showTitleCountBadge = computed(() => dataCount.value > 0 && Boolean(props.title?.trim()))
/** 标题区与操作区皆空时不渲染顶栏，避免留白一条 */
const showCardTopbar = computed(
  () =>
    Boolean(props.title?.trim()) ||
    Boolean(props.subtitle?.trim()) ||
    showTitleCountBadge.value ||
    props.showCreate ||
    Boolean(slots.actions) ||
    hasTitleTrailingSlot.value,
)

// === 查询面板：主行默认至多 2 个控件（关键词占 1 格），其余在「更多筛选」中展开 ===
const PRIMARY_SEARCH_CONTROLS = 2
const filtersExpanded = ref(false)
const keywordSlotCount = computed(() => (props.showKeywordSearch ? 1 : 0))
const primaryFilterFieldCount = computed(() =>
  Math.max(0, PRIMARY_SEARCH_CONTROLS - keywordSlotCount.value),
)
const primaryFilterFields = computed(() =>
  props.filterFields.slice(0, primaryFilterFieldCount.value),
)
const extraFilterFields = computed(() =>
  props.filterFields.slice(primaryFilterFieldCount.value),
)
const showFiltersMoreToggle = computed(() => extraFilterFields.value.length > 0)

const localKeyword = ref(props.keyword)
const filterValues = reactive<Record<string, any>>({})
props.filterFields.forEach(f => {
  if (filterValues[f.key] === undefined) filterValues[f.key] = f.type === 'switch' ? false : ''
})

const hasActiveFilters = computed(() => {
  if (localKeyword.value) return true
  return props.filterFields.some(f => {
    const v = filterValues[f.key]
    return v !== undefined && v !== '' && v !== null && v !== false
  })
})

function onKeywordInput(e: Event) {
  localKeyword.value = (e.target as HTMLInputElement).value
}
function clearKeyword() { localKeyword.value = '' }

function handleSearch() {
  const topKw = String(localKeyword.value ?? '').trim()
  const filterKw =
    filterValues.keyword !== undefined && filterValues.keyword !== null
      ? String(filterValues.keyword).trim()
      : ''
  emit('search', { keyword: topKw || filterKw, filters: { ...filterValues } })
}

function handleReset() {
  localKeyword.value = ''
  props.filterFields.forEach(f => { filterValues[f.key] = f.type === 'switch' ? false : '' })
  emit('reset')
  emit('search', { keyword: '', filters: { ...filterValues } })
}

const getTagMapLabel = (entry: any) => {
  if (typeof entry === 'string') return entry
  if (entry && typeof entry === 'object') return entry.label ?? entry
  return entry
}
const getTagMapType = (entry: any) => {
  if (entry && typeof entry === 'object' && entry.type) return entry.type
  return ''
}
const getActionsForRow = (row: any) => {
  const actions = props.columns.find(c => c.type === 'action')?.actions || []
  return actions.map(act => ({
    ...act,
    hidden: typeof act.hidden === 'function' ? act.hidden(row) : act.hidden,
  }))
}
const getStatusText = (status: string) => props.statusMap[status] || status
const formatDate = (date: string | Date) => {
  if (!date) return ''
  return new Date(date).toLocaleDateString('zh-CN')
}

const toggleSelectAll = () => {
  selectedItems.value = selectAll.value ? paginatedData.value.map(item => resolveRowKey(item)) : []
}
const toggleSelectItem = (itemId: string) => {
  const index = selectedItems.value.indexOf(itemId)
  if (index > -1) selectedItems.value.splice(index, 1)
  else selectedItems.value.push(itemId)
  selectAll.value = selectedItems.value.length === paginatedData.value.length
}

const sortByColumn = (columnKey: string) => {
  if (sortField.value === columnKey) sortDirection.value = sortDirection.value === 'asc' ? 'desc' : 'asc'
  else { sortField.value = columnKey; sortDirection.value = 'asc' }
  emit('sort-change', sortField.value, sortDirection.value)
}

const goToPage = (page: number) => {
  if (page >= 1 && page <= totalPages.value) { currentPage.value = page; emit('page-change', page) }
}

const onExpandChange = (row: any, expandedRows: any[]) => {
  emit('expand-change', row, expandedRows)
  syncElTableLayoutAfterTreeToggle()
}
const onRowClick = (row: any) => emit('row-click', row)
const onSelectionChange = (selection: any[]) => { selectedItems.value = selection.map((s: any) => resolveRowKey(s)) }

const handleView = (item: any) => emit('view', item)
const handleEdit = (item: any) => emit('edit', item)
const handleDelete = (item: any) => emit('delete', item)
const handleSelectedDelete = () => emit('batch-delete', selectedItems.value)
const handleSelectedExport = () => emit('selected-export', selectedItems.value)

onMounted(() => {
  const layout = document.querySelector('.neuro-command-layout')
  if (layout) theme.value = layout.getAttribute('data-theme') as 'dark' | 'light' || 'dark'
})
</script>

<style scoped>
@import '@/styles/theme/neuro-theme.css';

.neuro-agent-list-page {
  position: relative;
  width: 100%;
  flex: 1 1 auto;
  min-height: 0;
  display: flex;
  flex-direction: column;
  /* 勿用 100vh：会与 .content-main 双重撑高，滚动时整块业务区被「拖」出可视范围 */
  overflow-x: hidden;
}

.neural-data-stream {
  position: absolute;
  top: 0; left: 0; right: 0; bottom: 0;
  pointer-events: none;
  z-index: 0;
  opacity: 0.15;
}

.data-stream-line {
  position: absolute;
  top: 0; bottom: 0;
  width: 1px;
  background: linear-gradient(to bottom, transparent, var(--nm-primary-glow, #00f5d4), transparent);
  animation: data-stream-flow 8s infinite linear;
}

@keyframes data-stream-flow {
  0% { opacity: 0; transform: translateY(-100%); }
  10%, 90% { opacity: 0.6; }
  100% { opacity: 0; transform: translateY(100vh); }
}

/* ===== el-table 双层：与 NeuroAgentPageShell 对齐（hero → panel 内搜索+表） ===== */
.list-card.list-card--layered {
  position: relative;
  z-index: 5;
  margin-top: 12px;
  flex: 1 1 auto;
  min-height: 0;
  display: flex;
  flex-direction: column;
  gap: 12px;
  background: transparent;
  border: none;
  border-radius: 0;
  overflow: visible;
  backdrop-filter: none;
  box-shadow: none;
}

/* 第一、二层：无边框，用渐变 + 极淡斜纹底纹区分块面 */
.list-el-hero {
  flex-shrink: 0;
  border: none;
  border-radius: var(--neuro-radius-xl, 16px);
  overflow: hidden;
  backdrop-filter: blur(16px);
  background-color: color-mix(in srgb, var(--neuro-surface) 93%, var(--neuro-background));
  background-image:
    radial-gradient(120% 90% at 0% 0%, color-mix(in srgb, var(--neuro-primary) 16%, transparent), transparent 55%),
    radial-gradient(90% 70% at 100% 0%, color-mix(in srgb, var(--neuro-accent) 12%, transparent), transparent 50%);
}

.list-el-hero .card-topbar {
  border-bottom: none;
}

.list-el-panel {
  position: relative;
  flex: 1 1 auto;
  min-height: 0;
  display: flex;
  flex-direction: column;
  border: none;
  border-radius: var(--neuro-radius-xl, 16px);
  /* 与「单根主滚动条」一致：勿在此层裁切或形成第二滚动条；圆角由内层 card-table 处理 */
  overflow: visible;
  backdrop-filter: blur(16px);
  background-color: color-mix(in srgb, var(--neuro-surface) 86%, var(--neuro-background));
  background-image: linear-gradient(
    168deg,
    color-mix(in srgb, var(--neuro-primary) 6%, transparent) 0%,
    transparent 42%
  );
}

/* 仅在 list-el-panel 已 flex + min-height:0 的链路上让表区吃剩余高度并自行滚动 */
.list-el-panel > .card-search,
.list-el-panel > .card-toolbar,
.list-el-panel > .card-pagination {
  flex-shrink: 0;
}

/* 禁止 overflow-x:auto 与 overflow-y:visible 组合：规范会将纵向算作 auto，.card-table 变成第二纵向滚动容器（树收起后易残留滚动条占位）。
 * 横向溢出由表格内层 scrollbar（EP scrollX）处理；此处仅纵向随文档流动。 */
.list-el-panel .card-table {
  min-width: 0;
  flex: 0 1 auto;
  min-height: 0;
  overflow: visible;
}

/*
 * 无分页贴底面板：不要用 overflow:hidden 形成 scrollport（会逼出表体内层纵轴滚动）。
 * 底圆角改用 clip-path 裁切视觉，不改变 overflow 语义。
 */
.list-el-panel--flush-bottom .card-table {
  position: relative;
  z-index: 1;
  overflow: visible;
  clip-path: inset(0 round var(--neuro-radius-xl, 16px));
  border-bottom-left-radius: var(--neuro-radius-xl, 16px);
  border-bottom-right-radius: var(--neuro-radius-xl, 16px);
  border-bottom: 1px solid color-mix(in srgb, var(--neuro-primary) 20%, transparent);
}

.neuro-agent-list-page[data-theme="light"] .list-el-panel--flush-bottom .card-table {
  border-bottom-color: color-mix(in srgb, var(--neuro-primary) 28%, rgba(30, 41, 59, 0.22));
}

.list-el-panel--flush-bottom .neuro-el-table :deep(.el-table__body tr:last-child > td) {
  border-bottom: none !important;
}

.list-el-panel--flush-bottom .neuro-el-table :deep(.el-table__fixed-right .el-table__body tr:last-child > td),
.list-el-panel--flush-bottom .neuro-el-table :deep(.el-table__fixed-left .el-table__body tr:last-child > td) {
  border-bottom: none !important;
}


/* 顶部栏 */
.card-topbar {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px 20px;
  border-bottom: 1px solid var(--neuro-primary-10);
}

.topbar-left {
  flex: 1;
  min-width: 0;
}

.topbar-trailing {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
  flex-shrink: 0;
}

.card-title {
  margin: 0;
  font-size: 18px;
  font-weight: 700;
  color: var(--nm-text-primary, #e2e8f0);
  display: flex;
  align-items: center;
  gap: 10px;
}

.card-subtitle {
  margin: 2px 0 0;
  font-size: 12px;
  color: var(--nm-text-secondary, #94a3b8);
}

.title-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 2px 8px;
  background: rgba(0, 245, 212, 0.1);
  border: 1px solid rgba(0, 245, 212, 0.2);
  border-radius: 12px;
  font-size: 11px;
  font-weight: 500;
  color: var(--nm-primary, #00f5d4);
}

.badge-dot {
  width: 5px; height: 5px;
  background: var(--nm-primary, #00f5d4);
  border-radius: 50%;
}

.topbar-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.card-btn-create {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 18px;
  background: linear-gradient(135deg, #00d4aa 0%, #7c3aed 100%);
  border: none;
  border-radius: 10px;
  color: #fff;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  box-shadow: 0 2px 12px rgba(0, 212, 170, 0.25);
  white-space: nowrap;
}

.card-btn-create:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 20px rgba(0, 212, 170, 0.35);
}

/* 搜索栏：不用硬分割线，仅靠与面板一致的浅底过渡 */
.card-search {
  padding: 14px 20px 12px;
  border-bottom: none;
  background: linear-gradient(
    180deg,
    color-mix(in srgb, var(--neuro-background) 22%, transparent) 0%,
    transparent 72%
  );
}

.search-row {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 10px 12px;
}

.search-row--primary {
  display: flex;
  align-items: center;
  flex-wrap: nowrap;
  gap: 14px;
  justify-content: flex-start;
  min-width: 0;
}

.search-row__lead {
  display: flex;
  flex-wrap: nowrap;
  align-items: center;
  gap: 10px 12px;
  flex: 1 1 0;
  min-width: 0;
  overflow: visible;
}

/* 仅筛选项横向滚动；按钮在层外、紧跟控件 */
.search-row__lead-scroll {
  display: flex;
  flex-wrap: nowrap;
  align-items: center;
  gap: 10px 12px;
  flex: 0 1 auto;
  min-width: 0;
  max-width: 100%;
  overflow-x: auto;
}

/* 仅 #search-trailing-actions：排序、视图、展开收起等 */
.search-row__trailing {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: flex-end;
  gap: 8px 10px;
  flex: 0 0 auto;
  max-width: min(560px, 46%);
  margin-left: auto;
}

.search-row__trailing--solo {
  flex: 1 1 100%;
  max-width: none;
  margin-left: 0;
}

/* 查询/重置放在 overflow-x 容器外，避免可滚动区内命中/层叠导致点击无效 */
.search-btns--primary {
  flex-shrink: 0;
  position: relative;
  z-index: 2;
}

.search-row--extra {
  align-items: flex-start;
}

/* 与主搜索区同一底纹：展开后不再单独铺一层 */
.card-search-extra {
  margin-top: 0;
  padding-top: 10px;
  padding-bottom: 4px;
  border-top: none;
  background: transparent;
}

.search-filters-toggle {
  display: inline-flex;
  flex-shrink: 0;
  align-items: center;
  gap: 6px;
  padding: 8px 12px;
  border: 1px solid var(--neuro-border, rgba(148, 163, 184, 0.35));
  border-radius: var(--neuro-radius-full, 999px);
  background: var(--neuro-surface);
  color: var(--neuro-text-secondary, #94a3b8);
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: border-color 0.2s, color 0.2s;
  white-space: nowrap;
}

.search-filters-toggle:hover {
  border-color: var(--neuro-primary-30, rgba(0, 245, 212, 0.35));
  color: var(--neuro-primary, #00f5d4);
}

.search-filters-toggle__icon {
  flex-shrink: 0;
  transition: transform 0.2s ease;
}

.search-filters-toggle__icon.is-open {
  transform: rotate(180deg);
}

.search-input-group {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 12px;
  background: rgba(0, 0, 0, 0.03);
  border: 1px solid rgba(0, 0, 0, 0.1);
  border-radius: 8px;
  transition: border-color 0.2s, box-shadow 0.2s;
  max-width: 360px;
}

.dark .search-input-group {
  background: rgba(255, 255, 255, 0.04);
  border-color: rgba(45, 55, 72, 0.6);
}

.search-input-group:focus-within {
  border-color: rgba(0, 245, 212, 0.4);
  box-shadow: 0 0 0 3px rgba(0, 245, 212, 0.08);
}

.search-icon { color: #94a3b8; flex-shrink: 0; display: flex; align-items: center; }

.search-input {
  flex: 1;
  border: none;
  background: transparent;
  padding: 8px 0;
  font-size: 13px;
  color: #1e293b;
  outline: none;
  min-width: 0;
}

.dark .search-input { color: #e2e8f0; }
.search-input::placeholder { color: #94a3b8; }

.search-clear {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 20px; height: 20px;
  border: none;
  background: rgba(0, 0, 0, 0.05);
  border-radius: 50%;
  color: #64748b;
  cursor: pointer;
  transition: all 0.15s;
  flex-shrink: 0;
}

.dark .search-clear { background: rgba(255, 255, 255, 0.05); color: #94a3b8; }
.search-clear:hover { background: rgba(239, 68, 68, 0.1); color: #ef4444; }

.search-btns {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-shrink: 0;
}

.btn-search {
  padding: 7px 16px;
  background: linear-gradient(135deg, #00d4aa 0%, #7c3aed 100%);
  border: none;
  border-radius: 8px;
  color: #fff;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
}

.btn-search:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 16px rgba(0, 212, 170, 0.3);
}

.btn-reset {
  padding: 7px 12px;
  background: transparent;
  border: 1px solid rgba(0, 0, 0, 0.1);
  border-radius: 8px;
  color: #64748b;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
}

.dark .btn-reset { border-color: rgba(45, 55, 72, 0.5); color: #94a3b8; }
.btn-reset:hover { border-color: rgba(239, 68, 68, 0.3); color: #ef4444; background: rgba(239, 68, 68, 0.05); }

/* 筛选字段（与关键词、按钮同一行，按需换行） */
.filter-item {
  display: inline-flex;
  flex-direction: row;
  align-items: center;
  gap: 0;
  flex: 0 0 auto;
  max-width: 100%;
}

.filter-item--select :deep(.el-select) {
  width: 148px;
}

.filter-item--text :deep(.el-input) {
  width: 200px;
}

.filter-item--number :deep(.el-input-number) {
  width: 140px;
}

.filter-item--date :deep(.el-date-editor) {
  width: 160px;
}

.filter-item--daterange :deep(.el-date-editor) {
  width: 260px;
  max-width: min(260px, 100%);
}

.card-search .filter-item :deep(.el-input__wrapper) {
  background: rgba(0, 0, 0, 0.03) !important;
  border: 1px solid rgba(0, 0, 0, 0.1) !important;
  border-radius: 8px !important;
  box-shadow: none !important;
  padding: 0 10px !important;
}

.dark .card-search .filter-item :deep(.el-input__wrapper) {
  background: rgba(255, 255, 255, 0.04) !important;
  border-color: rgba(45, 55, 72, 0.6) !important;
}

.card-search .filter-item :deep(.el-input__inner) { color: #1e293b !important; }
.dark .card-search .filter-item :deep(.el-input__inner) { color: #e2e8f0 !important; }
.card-search .filter-item :deep(.el-input__inner::placeholder) { color: #94a3b8 !important; }

.card-search .filter-item :deep(.el-select__wrapper) {
  background: rgba(0, 0, 0, 0.03) !important;
  border: 1px solid rgba(0, 0, 0, 0.1) !important;
  border-radius: 8px !important;
  box-shadow: none !important;
  min-height: 34px;
}

.dark .card-search .filter-item :deep(.el-select__wrapper) {
  background: rgba(255, 255, 255, 0.04) !important;
  border-color: rgba(45, 55, 72, 0.6) !important;
}

/* 工具栏 */
.card-toolbar {
  padding: 12px 20px;
  border-bottom: none;
  background: color-mix(in srgb, var(--neuro-surface) 94%, var(--neuro-primary-10));
}

.card-table {
  padding: 0;
}

.neuro-el-table {
  --el-table-bg-color: transparent;
  --el-table-tr-bg-color: transparent;
  --el-table-header-bg-color: transparent;
  --el-table-row-hover-bg-color: var(--neuro-primary-10);
  --el-table-border-color: color-mix(in srgb, var(--neuro-primary) 8%, transparent);
  --el-table-text-color: var(--neuro-text);
  --el-table-header-text-color: var(--neuro-primary);
  border-radius: 0 !important;
}

.neuro-el-table :deep(.el-table__inner-wrapper::before),
.neuro-el-table :deep(.el-table__inner-wrapper),
.neuro-el-table :deep(.el-table__header-wrapper),
.neuro-el-table :deep(.el-table__header),
.neuro-el-table :deep(.el-table__header thead),
.neuro-el-table :deep(.el-table__body-wrapper),
.neuro-el-table :deep(.el-table__body),
.neuro-el-table :deep(.el-table__footer-wrapper) {
  border-radius: 0 !important;
}

.neuro-el-table :deep(.el-table__header-wrapper),
.neuro-el-table :deep(.el-table__header),
.neuro-el-table :deep(.el-table__header th) {
  background: rgba(0, 0, 0, 0.04) !important;
}

.dark .neuro-el-table :deep(.el-table__header-wrapper),
.dark .neuro-el-table :deep(.el-table__header),
.dark .neuro-el-table :deep(.el-table__header th) {
  background: rgba(0, 245, 212, 0.04) !important;
}

.neuro-el-table :deep(.el-table__header th) {
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 0.5px;
  color: #475569 !important;
  border-bottom: 1px solid color-mix(in srgb, var(--neuro-text) 8%, transparent) !important;
}

.dark .neuro-el-table :deep(.el-table__header th) {
  color: #94a3b8 !important;
  border-bottom-color: color-mix(in srgb, var(--neuro-text) 6%, transparent) !important;
}

.neuro-el-table :deep(.el-table__body td) {
  font-size: 13px;
  padding: 12px 0;
  color: #1e293b !important;
}

.dark .neuro-el-table :deep(.el-table__body td) {
  color: #e2e8f0 !important;
}

.neuro-el-table :deep(.el-table-column--selection .cell),
.neuro-el-table :deep(.el-table__header .el-table-column--selection .cell) {
  display: flex !important;
  justify-content: center !important;
}

.neuro-el-table :deep(.el-table__row) { transition: background 0.2s ease; }
.neuro-action-col { display: flex; gap: 4px; flex-wrap: wrap; }
.neuro-empty { padding: 40px 20px; text-align: center; color: var(--nm-text-secondary, #94a3b8); font-size: 14px; }

/*
 * 单一纵向滚动源（文档）：EP 默认 inner-wrapper 为 height:100% + body-wrapper 为 flex:1 + overflow:hidden，
 * 在固定列或树表节点展开后会在表内再出一根纵轴滚动条（菜单、组织架构等）。
 */
.neuro-el-table :deep(.el-table__inner-wrapper) {
  height: auto !important;
}

.neuro-el-table :deep(.el-table__body-wrapper) {
  flex: none !important;
  overflow: visible !important;
}

/* 后台统一使用浏览器系统滚动条；表格不可再生成自己的纵向滚动条。 */
.neuro-el-table :deep(.el-table__body-wrapper),
.neuro-el-table :deep(.el-scrollbar),
.neuro-el-table :deep(.el-scrollbar__wrap),
.neuro-el-table :deep(.el-scrollbar__view) {
  height: auto !important;
  max-height: none !important;
}

.neuro-el-table :deep(.el-table__body-wrapper),
.neuro-el-table :deep(.el-scrollbar__wrap) {
  overflow-y: visible !important;
}

.neuro-el-table :deep(.el-scrollbar__bar) {
  display: none !important;
}

.neuro-el-table :deep(.el-table__body-wrapper) {
  scrollbar-width: none;
}

.neuro-el-table :deep(.el-table__body-wrapper)::-webkit-scrollbar {
  width: 0;
  height: 0;
}

/* 分页 */
.card-pagination {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 12px;
  padding: 16px 20px 14px;
  margin-top: 2px;
  border-top: none;
  background: linear-gradient(
    180deg,
    transparent 0%,
    color-mix(in srgb, var(--neuro-background) 28%, transparent) 100%
  );
}

.pagination-size {
  position: relative;
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  background: var(--neuro-surface);
  border: 1px solid var(--neuro-border);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
  user-select: none;
}

.pagination-size:hover { border-color: rgba(0, 245, 212, 0.5); }
.size-text { font-size: 12px; color: var(--nm-text-secondary, #94a3b8); }
.size-arrow { font-size: 10px; color: var(--nm-text-secondary, #94a3b8); transition: transform 0.2s ease; }
.size-arrow.arrow-up { transform: rotate(180deg); }

.size-dropdown {
  position: absolute;
  bottom: calc(100% + 6px);
  left: 0;
  background: var(--neuro-surface-90);
  border: 1px solid var(--neuro-primary-20);
  border-radius: 8px;
  padding: 4px 0;
  min-width: 90px;
  z-index: 100;
  box-shadow: var(--neuro-shadow-md);
}

.size-option {
  padding: 6px 14px;
  font-size: 12px;
  color: var(--nm-text-secondary, #94a3b8);
  white-space: nowrap;
  transition: all 0.15s ease;
}

.size-option:hover { background: rgba(0, 245, 212, 0.1); color: var(--nm-primary, #00f5d4); }
.size-option.option-active { color: var(--nm-primary, #00f5d4); font-weight: 600; }

.pagination-btn {
  padding: 6px 14px;
  background: transparent;
  border: 1px solid transparent;
  border-radius: 8px;
  color: var(--nm-text-secondary, #94a3b8);
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s ease;
  font-family: inherit;
}

.pagination-btn:hover:not(:disabled) { border-color: rgba(0, 245, 212, 0.3); color: var(--nm-primary, #00f5d4); }
.pagination-btn:disabled { opacity: 0.4; cursor: not-allowed; }

.pagination-pages { display: flex; gap: 4px; }

.pagination-page {
  width: 30px; height: 30px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: 1px solid transparent;
  border-radius: 8px;
  color: var(--nm-text-secondary, #94a3b8);
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s ease;
  font-family: inherit;
}

.pagination-page:hover { border-color: rgba(0, 245, 212, 0.3); color: var(--nm-primary, #00f5d4); }
.pagination-page.page-active { background: var(--nm-primary, #00f5d4); color: var(--nm-bg, #0a0c14); font-weight: 600; }
.pagination-ellipsis { color: var(--nm-text-secondary, #94a3b8); font-size: 12px; padding: 0 4px; }
.pagination-info { font-size: 12px; color: var(--nm-text-secondary, #94a3b8); }

/* ===== 原生神经表格（未变） ===== */
.neuro-data-table {
  position: relative;
  z-index: 5;
  margin-top: 12px;
  overflow: visible;
  border: 1px solid rgba(0, 245, 212, 0.12);
  border-radius: 12px;
}

.table-header {
  display: flex;
  align-items: center;
  padding: 12px 20px;
  background: rgba(15, 23, 42, 0.95);
}

.table-actions { display: flex; align-items: center; gap: 12px; }

.neuro-checkbox { display: flex; align-items: center; gap: 6px; cursor: pointer; user-select: none; }
.neuro-checkbox input { display: none; }

.checkbox-custom {
  width: 14px; height: 14px;
  border: 2px solid rgba(0, 245, 212, 0.5);
  border-radius: 4px;
  position: relative;
  transition: all 0.2s ease;
}

.neuro-checkbox input:checked + .checkbox-custom {
  background: var(--nm-primary, #00f5d4);
  border-color: var(--nm-primary, #00f5d4);
}

.neuro-checkbox input:checked + .checkbox-custom::after {
  content: '✓';
  position: absolute;
  top: 50%; left: 50%;
  transform: translate(-50%, -50%);
  color: var(--nm-bg, #0a0c14);
  font-size: 9px;
  font-weight: bold;
}

.checkbox-label { font-size: 12px; color: var(--nm-text-secondary, #94a3b8); }
.selected-actions { display: flex; align-items: center; gap: 8px; }
.selected-count { font-size: 12px; color: var(--nm-primary, #00f5d4); font-weight: 500; }

.action-btn {
  padding: 5px 10px;
  background: rgba(15, 23, 42, 0.8);
  border: 1px solid rgba(100, 116, 139, 0.3);
  border-radius: 8px;
  color: var(--nm-text-secondary, #94a3b8);
  font-size: 11px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.action-btn:hover { border-color: rgba(0, 245, 212, 0.5); color: var(--nm-primary, #00f5d4); }
.action-delete:hover { border-color: rgba(239, 68, 68, 0.5); color: #ef4444; }
.action-export:hover { border-color: rgba(34, 197, 94, 0.5); color: #22c55e; }

.table-scroll { overflow-x: auto; }
.neuro-table { width: 100%; border-collapse: collapse; }

.neuro-table thead {
  background: rgba(15, 23, 42, 0.95);
  border-bottom: 2px solid rgba(0, 245, 212, 0.2);
}

.neuro-table th {
  padding: 0 16px;
  height: 65px;
  text-align: left;
  font-size: 12px;
  font-weight: 600;
  color: var(--nm-primary, #00f5d4);
  white-space: nowrap;
  letter-spacing: 0.5px;
  vertical-align: middle;
}

.neuro-table th.sortable { cursor: pointer; }
.neuro-table th.sortable:hover { background: rgba(0, 245, 212, 0.05); }

.column-header { display: flex; align-items: center; gap: 6px; }
.sort-indicator { font-size: 10px; }

.neuro-table td {
  padding: 12px 16px;
  border-bottom: 1px solid rgba(0, 245, 212, 0.08);
  color: var(--nm-text-primary, #e2e8f0);
  font-size: 13px;
}

.neuro-table .table-col-select { width: 50px; min-width: 50px; text-align: center; }
.neuro-table .table-col-select .neuro-checkbox { justify-content: center; }
.neuro-table tbody tr { background: transparent; transition: background 0.2s ease; }
.neuro-table tbody tr:hover { background: rgba(0, 245, 212, 0.03); }
.neuro-table tbody tr.row-selected { background: rgba(0, 245, 212, 0.08); }
.cell-content { display: flex; align-items: center; gap: 6px; }
.cell-tags { display: flex; flex-wrap: wrap; gap: 4px; align-items: center; }

.status-dot {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 18px;
  height: 18px;
  border-radius: 50%;
  font-size: 10px;
  font-weight: bold;
  line-height: 1;
}

.status-active,
.status-enabled {
  background: rgba(34, 197, 94, 0.15);
  color: #22c55e;
}

.status-completed {
  background: rgba(34, 197, 94, 0.15);
  color: #22c55e;
}

.status-disabled {
  background: color-mix(in srgb, var(--nm-bg-elevated) 72%, var(--nm-bg-deep));
  color: var(--nm-text-muted);
}

.status-error {
  background: rgba(239, 68, 68, 0.15);
  color: #ef4444;
}

.status-pending,
.status-warning {
  background: rgba(245, 158, 11, 0.15);
  color: #f59e0b;
}

.status-paused,
.status-inactive {
  background: color-mix(in srgb, var(--nm-bg-elevated) 72%, var(--nm-bg-deep));
  color: var(--nm-text-muted);
}

.dark .status-active,
.dark .status-enabled,
.dark .status-completed {
  background: rgba(34, 197, 94, 0.2);
  color: #4ade80;
}

.dark .status-disabled {
  background: color-mix(in srgb, var(--nm-bg-elevated) 78%, var(--nm-bg-deep));
  color: var(--nm-text-muted);
}

.dark .status-error {
  background: rgba(239, 68, 68, 0.2);
  color: #f87171;
}

.dark .status-pending,
.dark .status-warning {
  background: rgba(245, 158, 11, 0.2);
  color: #fbbf24;
}

.dark .status-paused,
.dark .status-inactive {
  background: color-mix(in srgb, var(--nm-bg-elevated) 78%, var(--nm-bg-deep));
  color: var(--nm-text-muted);
}

.data-tag {
  padding: 2px 8px;
  background: rgba(0, 245, 212, 0.08);
  border: 1px solid rgba(0, 245, 212, 0.2);
  border-radius: 8px;
  font-size: 11px;
  color: var(--nm-primary, #00f5d4);
}

.cell-progress { display: flex; align-items: center; gap: 8px; }
.progress-bar { flex: 1; height: 4px; background: rgba(15, 23, 42, 0.8); border-radius: 2px; overflow: hidden; }
.progress-fill { height: 100%; background: linear-gradient(90deg, var(--nm-primary, #00f5d4), #9d4edd); border-radius: 2px; }
.progress-text { font-size: 11px; color: var(--nm-text-secondary, #94a3b8); min-width: 35px; }

.table-col-actions { width: 200px; min-width: 200px; }
.action-buttons { display: flex; gap: 8px; }
.action-buttons .action-btn { padding: 6px 16px; font-size: 12px; }
.action-view:hover { border-color: rgba(59, 130, 246, 0.5); color: #3b82f6; }
.action-edit:hover { border-color: rgba(245, 158, 11, 0.5); color: #f59e0b; }

.table-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  text-align: center;
}

.empty-neuron { width: 80px; height: 80px; margin-bottom: 24px; }
.neuron-animation {
  width: 100%; height: 100%;
  background: radial-gradient(circle, rgba(0, 245, 212, 0.15) 0%, transparent 70%);
  border: 2px solid rgba(0, 245, 212, 0.2);
  border-radius: 50%;
  animation: empty-pulse 4s infinite;
}

@keyframes empty-pulse {
  0%, 100% { box-shadow: 0 0 30px rgba(0, 245, 212, 0.05); }
  50% { box-shadow: 0 0 50px rgba(0, 245, 212, 0.15); }
}

.empty-title {
  font-size: 18px;
  font-weight: 600;
  color: var(--nm-primary, #00f5d4);
  margin: 0 0 8px;
}

.empty-desc { font-size: 13px; color: var(--nm-text-secondary, #94a3b8); margin-bottom: 20px; }

.neuro-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 10px 18px;
  border: none;
  border-radius: 10px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}

.neuro-btn-primary {
  background: linear-gradient(135deg, var(--nm-primary, #00f5d4), #9d4edd);
  color: var(--nm-bg, #0a0c14);
}

.neuro-btn-primary:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(0, 245, 212, 0.3);
}

/* 分页 */
.neuro-pagination {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 12px;
  padding: 16px;
  border-top: 1px solid var(--neuro-primary-10);
  overflow: visible;
}

/* 响应式 */
@media (max-width: 768px) {
  .search-row--primary {
    flex-direction: column;
    align-items: stretch;
    flex-wrap: wrap;
    gap: 10px;
  }
  .search-row__lead {
    flex: none;
    width: 100%;
    flex-wrap: wrap;
  }
  .search-row__lead-scroll {
    flex: 1 1 auto;
    width: 100%;
    max-width: none;
    overflow-x: visible;
    flex-wrap: wrap;
  }
  .search-row__trailing {
    flex: none;
    width: 100%;
    max-width: none;
    margin-left: 0;
    justify-content: flex-end;
  }
  .search-row__trailing--solo {
    flex: none;
    width: 100%;
    max-width: none;
    margin-left: 0;
    justify-content: flex-end;
  }
  .search-row--extra {
    flex-direction: column;
    align-items: stretch;
  }
  .search-input-group { max-width: 100%; }
  .search-btns {
    margin-left: 0;
    width: auto;
    justify-content: flex-start;
  }
  .search-filters-toggle {
    width: auto;
    justify-content: center;
  }
  .filter-item { width: 100%; }
  .filter-item--select :deep(.el-select) {
    flex: 1;
    width: auto !important;
    min-width: 0;
  }
  .filter-item--text :deep(.el-input),
  .filter-item--number :deep(.el-input-number),
  .filter-item--date :deep(.el-date-editor),
  .filter-item--daterange :deep(.el-date-editor) {
    flex: 1;
    width: auto !important;
    max-width: none !important;
    min-width: 0;
  }
  .card-topbar {
    flex-direction: column;
    align-items: stretch;
    gap: 12px;
  }
  .topbar-trailing,
  .topbar-actions {
    justify-content: flex-end;
    flex-wrap: wrap;
  }
  .advanced-grid { grid-template-columns: 1fr; }
}
</style>
