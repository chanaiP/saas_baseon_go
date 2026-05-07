<template>
  <div class="showcase-content-wrapper" :data-theme="theme">
    <!-- 左侧：组件名称列表 -->
    <aside class="showcase-sidebar">
      <div class="sidebar-header">
        <h2 class="sidebar-title">组件库</h2>
        <p class="sidebar-count">目录 {{ filteredComponents.length }} / 共 {{ components.length }} 个</p>
      </div>
      <nav class="component-list">
        <button
          v-for="comp in filteredComponents"
          :key="comp.id"
          class="component-item"
          :class="{ 'component-active': activeComponent === comp.id }"
          @click="activeComponent = comp.id"
        >
          <span class="comp-icon">{{ comp.icon }}</span>
          <div class="comp-info">
            <span class="comp-name">{{ comp.name }}</span>
            <span class="comp-status" :class="'status-' + comp.status">{{ comp.statusText }}</span>
          </div>
        </button>
      </nav>
    </aside>

    <!-- 右侧：与业务页一致的两层骨架（NeuroAgentPageShell）+ 搜索条 + 演示区 -->
    <main class="showcase-main">
      <NeuroAgentPageShell v-if="currentComp" class="showcase-page-shell">
        <template #title>
          <span class="showcase-hero-title">
            <span class="showcase-hero-icon" aria-hidden="true">{{ currentComp.icon }}</span>
            {{ currentComp.fullName }}
          </span>
        </template>
        <template #subtitle>{{ currentComp.description }}</template>
        <template #meta>
          <span class="showcase-status-pill" :class="'showcase-status--' + currentComp.status">{{ currentComp.statusText }}</span>
          <span v-for="tag in currentComp.tags" :key="tag" class="showcase-meta-tag">{{ tag }}</span>
        </template>
        <template #actions>
          <button type="button" class="showcase-copy-btn" @click="copyComponentName">复制组件名</button>
        </template>

        <div class="showcase-workbench">
          <!-- 列表类演示自带搜索区，避免在「业务第一层」与「搜索+表」之间插入目录搜索条 -->
          <template v-if="showWorkbenchDirectorySearch">
            <div class="showcase-search-strip">
              <div class="showcase-search-wrap">
                <svg class="showcase-search-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
                  <circle cx="11" cy="11" r="8" />
                  <path d="m21 21-4.35-4.35" />
                </svg>
                <input
                  v-model="compSearch"
                  type="search"
                  class="showcase-search-input"
                  placeholder="搜索左侧目录中的组件…"
                  autocomplete="off"
                />
                <button v-if="compSearch" type="button" class="showcase-search-clear" title="清除" @click="compSearch = ''">
                  <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M18 6 6 18M6 6l12 12" />
                  </svg>
                </button>
              </div>
              <button
                type="button"
                class="showcase-filter-toggle"
                :aria-expanded="filtersExpanded"
                @click="filtersExpanded = !filtersExpanded"
              >
                <span>{{ filtersExpanded ? '收起筛选' : '更多筛选' }}</span>
                <svg
                  class="showcase-filter-toggle__chev"
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

            <div v-show="filtersExpanded" class="showcase-filters-extra">
              <span class="showcase-filter-label">目录状态</span>
              <div class="showcase-filter-chips">
                <button
                  type="button"
                  class="showcase-filter-chip"
                  :class="{ active: statusSidebarFilter === 'all' }"
                  @click="statusSidebarFilter = 'all'"
                >全部</button>
                <button
                  type="button"
                  class="showcase-filter-chip"
                  :class="{ active: statusSidebarFilter === 'ready' }"
                  @click="statusSidebarFilter = 'ready'"
                >已就绪</button>
                <button
                  type="button"
                  class="showcase-filter-chip"
                  :class="{ active: statusSidebarFilter === 'developing' }"
                  @click="statusSidebarFilter = 'developing'"
                >开发中</button>
              </div>
            </div>
          </template>

          <div class="demo-area">
          <!-- 智能列表页 -->
          <div v-if="activeComponent === 'list'" class="demo-section">
            <div class="demo-wrapper">
              <NeuroAgentListPage
                :title="'AI项目管理'"
                :columns="listColumns"
                :data="listData"
                :pageSize="8"
                show-keyword-search
                search-placeholder="搜索项目名称…"
                :filter-fields="listDemoFilterFields"
                @create="handleListCreate"
                @view="handleListView"
                @edit="handleListEdit"
                @delete="handleListDelete"
              >
                <template #search-trailing-actions>
                  <div class="na-demo-search-trailing" role="toolbar" aria-label="列表展示与排序">
                    <div class="na-demo-search-trailing__main">
                      <el-select
                        v-model="listDemoSortSelect"
                        class="na-demo-sort-select"
                        placeholder="请选择排序"
                        clearable
                        aria-label="排序"
                      >
                        <el-option
                          v-for="o in listDemoSortSelectOptions"
                          :key="o.value"
                          :label="o.label"
                          :value="o.value"
                        />
                      </el-select>
                    </div>
                    <div class="na-demo-search-trailing__view">
                      <div class="na-demo-view-switcher" role="group" aria-label="视图">
                        <button
                          type="button"
                          class="na-demo-view-btn"
                          :class="{ 'na-demo-view-btn--active': listDemoViewLayout === 'grid' }"
                          title="网格 / 卡片"
                          @click="listDemoViewLayout = 'grid'"
                        >
                          <svg viewBox="0 0 24 24" fill="currentColor" width="18" height="18" aria-hidden="true">
                            <rect x="3" y="3" width="7" height="7" rx="1.5" />
                            <rect x="14" y="3" width="7" height="7" rx="1.5" />
                            <rect x="3" y="14" width="7" height="7" rx="1.5" />
                            <rect x="14" y="14" width="7" height="7" rx="1.5" />
                          </svg>
                        </button>
                        <button
                          type="button"
                          class="na-demo-view-btn"
                          :class="{ 'na-demo-view-btn--active': listDemoViewLayout === 'list' }"
                          title="列表"
                          @click="listDemoViewLayout = 'list'"
                        >
                          <svg viewBox="0 0 24 24" fill="currentColor" width="18" height="18" aria-hidden="true">
                            <rect x="3" y="4" width="18" height="4" rx="1.5" />
                            <rect x="3" y="10" width="18" height="4" rx="1.5" />
                            <rect x="3" y="16" width="18" height="4" rx="1.5" />
                          </svg>
                        </button>
                      </div>
                    </div>
                  </div>
                </template>
              </NeuroAgentListPage>
            </div>
          </div>

          <!-- el-table 标准模式 -->
          <div v-else-if="activeComponent === 'list-el'" class="demo-section">
            <div class="demo-wrapper">
              <NeuroAgentListPage
                mode="el-table"
                :title="'用户管理 — el-table 模式'"
                :subtitle="'支持分页、排序、批量操作'"
                :columns="elTableColumns"
                :data="elTableData"
                row-key="id"
                :show-selection="true"
                :show-index="true"
                :show-create="true"
                create-label="新增用户"
                :status-map="elTableStatusMap"
                :page-size="8"
                show-keyword-search
                search-placeholder="搜索姓名或邮箱…"
                :filter-fields="elTableDemoFilterFields"
                @create="handleElTableCreate"
                @view="handleElTableView"
                @edit="handleElTableEdit"
                @delete="handleElTableDelete"
                @batch-delete="handleElTableBatchDelete"
              >
                <template #search-trailing-actions>
                  <div class="na-demo-search-trailing" role="toolbar" aria-label="用户列表展示与排序">
                    <div class="na-demo-search-trailing__main">
                      <el-select
                        v-model="elTableDemoSortSelect"
                        class="na-demo-sort-select"
                        placeholder="请选择排序"
                        clearable
                        aria-label="排序"
                      >
                        <el-option
                          v-for="o in elTableDemoSortSelectOptions"
                          :key="o.value"
                          :label="o.label"
                          :value="o.value"
                        />
                      </el-select>
                    </div>
                    <div class="na-demo-search-trailing__view">
                      <div class="na-demo-view-switcher" role="group" aria-label="视图">
                        <button
                          type="button"
                          class="na-demo-view-btn"
                          :class="{ 'na-demo-view-btn--active': elTableDemoViewLayout === 'grid' }"
                          title="网格 / 卡片"
                          @click="elTableDemoViewLayout = 'grid'"
                        >
                          <svg viewBox="0 0 24 24" fill="currentColor" width="18" height="18" aria-hidden="true">
                            <rect x="3" y="3" width="7" height="7" rx="1.5" />
                            <rect x="14" y="3" width="7" height="7" rx="1.5" />
                            <rect x="3" y="14" width="7" height="7" rx="1.5" />
                            <rect x="14" y="14" width="7" height="7" rx="1.5" />
                          </svg>
                        </button>
                        <button
                          type="button"
                          class="na-demo-view-btn"
                          :class="{ 'na-demo-view-btn--active': elTableDemoViewLayout === 'list' }"
                          title="列表"
                          @click="elTableDemoViewLayout = 'list'"
                        >
                          <svg viewBox="0 0 24 24" fill="currentColor" width="18" height="18" aria-hidden="true">
                            <rect x="3" y="4" width="18" height="4" rx="1.5" />
                            <rect x="3" y="10" width="18" height="4" rx="1.5" />
                            <rect x="3" y="16" width="18" height="4" rx="1.5" />
                          </svg>
                        </button>
                      </div>
                    </div>
                  </div>
                </template>
              </NeuroAgentListPage>
            </div>
          </div>

          <!-- el-table 树形模式（菜单管理） -->
          <div v-else-if="activeComponent === 'list-tree'" class="demo-section">
            <div class="demo-wrapper">
              <NeuroAgentListPage
                :key="'tree-expand-' + String(treeDemoTreeExpanded)"
                mode="el-table"
                :title="'菜单管理 — 树形模式'"
                :subtitle="'目录 / 菜单 / 按钮 三级结构'"
                :columns="treeTableColumns"
                :data="treeTableData"
                row-key="id"
                :tree-props="{ children: 'children', hasChildren: 'hasChildren' }"
                :default-expand-all="treeDemoTreeExpanded"
                :status-map="treeStatusMap"
                show-keyword-search
                search-placeholder="搜索菜单名称…"
                :filter-fields="treeDemoFilterFields"
                @cell-change="handleTreeCellChange"
                @edit="handleTreeEdit"
              >
                <template #search-trailing-actions>
                  <div class="na-demo-search-trailing" role="toolbar" aria-label="菜单树展示">
                    <div class="na-demo-search-trailing__main">
                      <el-select
                        v-model="treeDemoSortSelect"
                        class="na-demo-sort-select"
                        placeholder="请选择排序"
                        clearable
                        aria-label="排序"
                      >
                        <el-option
                          v-for="o in treeDemoSortSelectOptions"
                          :key="o.value"
                          :label="o.label"
                          :value="o.value"
                        />
                      </el-select>
                      <button
                        type="button"
                        class="na-demo-hierarchy-toggle"
                        :title="treeDemoTreeExpanded ? '全部收起' : '全部展开'"
                        :aria-label="treeDemoTreeExpanded ? '全部收起' : '全部展开'"
                        @click="treeDemoTreeExpanded = !treeDemoTreeExpanded"
                      >
                        <span class="na-demo-hierarchy-toggle__text">{{
                          treeDemoTreeExpanded ? '全部收起' : '全部展开'
                        }}</span>
                        <svg
                          v-if="treeDemoTreeExpanded"
                          class="na-demo-hierarchy-toggle__chev"
                          viewBox="0 0 24 24"
                          fill="none"
                          stroke="currentColor"
                          stroke-width="2"
                          stroke-linecap="round"
                          stroke-linejoin="round"
                          width="16"
                          height="16"
                          aria-hidden="true"
                        >
                          <path d="M18 15l-6-6-6 6" />
                        </svg>
                        <svg
                          v-else
                          class="na-demo-hierarchy-toggle__chev"
                          viewBox="0 0 24 24"
                          fill="none"
                          stroke="currentColor"
                          stroke-width="2"
                          stroke-linecap="round"
                          stroke-linejoin="round"
                          width="16"
                          height="16"
                          aria-hidden="true"
                        >
                          <path d="M6 9l6 6 6-6" />
                        </svg>
                      </button>
                    </div>
                    <div class="na-demo-search-trailing__view">
                      <div class="na-demo-view-switcher" role="group" aria-label="视图">
                        <button
                          type="button"
                          class="na-demo-view-btn"
                          :class="{ 'na-demo-view-btn--active': treeDemoViewLayout === 'grid' }"
                          title="网格 / 卡片"
                          @click="treeDemoViewLayout = 'grid'"
                        >
                          <svg viewBox="0 0 24 24" fill="currentColor" width="18" height="18" aria-hidden="true">
                            <rect x="3" y="3" width="7" height="7" rx="1.5" />
                            <rect x="14" y="3" width="7" height="7" rx="1.5" />
                            <rect x="3" y="14" width="7" height="7" rx="1.5" />
                            <rect x="14" y="14" width="7" height="7" rx="1.5" />
                          </svg>
                        </button>
                        <button
                          type="button"
                          class="na-demo-view-btn"
                          :class="{ 'na-demo-view-btn--active': treeDemoViewLayout === 'list' }"
                          title="列表"
                          @click="treeDemoViewLayout = 'list'"
                        >
                          <svg viewBox="0 0 24 24" fill="currentColor" width="18" height="18" aria-hidden="true">
                            <rect x="3" y="4" width="18" height="4" rx="1.5" />
                            <rect x="3" y="10" width="18" height="4" rx="1.5" />
                            <rect x="3" y="16" width="18" height="4" rx="1.5" />
                          </svg>
                        </button>
                      </div>
                    </div>
                  </div>
                </template>
                <template #col-showInSidebar="{ row }">
                  <el-switch
                    v-model="row.showInSidebar"
                    size="small"
                    @click.stop
                    @change="handleTreeCellChange('showInSidebar', row)"
                  />
                </template>
              </NeuroAgentListPage>
            </div>
          </div>

          <!-- 全息详情页 -->
          <div v-else-if="activeComponent === 'detail'" class="demo-section">
            <div class="demo-wrapper">
              <NeuroAgentDetailPage
                :data="detailData"
                theme="dark"
                @back="() => {}"
                @edit="() => {}"
                @action="() => {}"
              />
            </div>
          </div>

          <!-- 智能表单页 -->
          <div v-else-if="activeComponent === 'form'" class="demo-section">
            <div class="demo-wrapper">
              <NeuroAgentFormPage
                :config="formConfig"
                :fields="formFields"
                :steps="formSteps"
                theme="dark"
                @back="() => {}"
                @submit="handleFormSubmit"
                @reset="() => {}"
              />
            </div>
          </div>

          <!-- 全息弹窗 -->
          <div v-else-if="activeComponent === 'dialog'" class="demo-section">
            <div class="dialog-demo-actions">
              <button class="demo-btn" @click="openDialog('small')">小尺寸</button>
              <button class="demo-btn" @click="openDialog('medium')">中尺寸</button>
              <button class="demo-btn" @click="openDialog('large')">大尺寸</button>
              <button class="demo-btn" @click="openDialog('fullscreen')">全屏</button>
              <button class="demo-btn demo-btn--loading" @click="openLoadingDialog">加载中</button>
            </div>
            <NeuroAgentDialog
              v-model="dialogVisible"
              :title="dialogTitle"
              :icon="dialogIcon"
              :size="dialogSize"
              :loading="dialogLoading"
              @close="dialogVisible = false"
              @cancel="dialogVisible = false"
              @confirm="dialogVisible = false"
            >
              <div class="dialog-demo-body">
                <p>这是一个弹窗的内容区域示例。</p>
                <ul>
                  <li>支持四种尺寸：small / medium / large / fullscreen</li>
                  <li>加载状态覆盖层</li>
                  <li>ESC 关闭 / 点击遮罩关闭</li>
                  <li>渐变顶部装饰线</li>
                </ul>
              </div>
            </NeuroAgentDialog>
          </div>

          <!-- 智能搜索栏 -->
          <div v-else-if="activeComponent === 'search'" class="demo-section">
            <div class="demo-wrapper demo-compact">
              <NeuroAgentSearchBar
                placeholder="输入关键词搜索项目、任务或文档..."
                @search="handleSearch"
              />
            </div>
            <div class="search-hint">
              <p>自然语言 AI 搜索栏，支持模糊匹配和语义理解。</p>
            </div>
          </div>

          <!-- 智能数据表 -->
          <div v-else-if="activeComponent === 'table'" class="demo-section">
            <div class="demo-wrapper">
              <NeuroAgentTable
                :columns="tableColumns"
                :data="tableData"
              />
            </div>
          </div>

          <!-- 神经数据卡片 -->
          <div v-else-if="activeComponent === 'card'" class="demo-section">
            <div class="card-grid">
              <NeuroAgentCard
                v-for="card in cardData"
                :key="card.id"
                :title="card.title"
                :description="card.description"
                :icon="card.icon"
                :status="card.status"
              />
            </div>
          </div>

          <!-- 卡片列表容器 -->
          <div v-else-if="activeComponent === 'card-list'" class="demo-section">
            <CardListView
              :cards="cardListData"
              :filter-options="cardListFilters"
              :col-count="3"
              default-layout="grid"
              search-placeholder="搜索项目名称、描述..."
              @card-click="handleCardClick"
            >
              <template #toolbar-filters>
                <div class="card-list-demo-extra">
                  <span class="card-list-demo-extra__hint">以下为 CardListView 的「更多筛选」折叠区示例（可多放日期、下拉等字段）。</span>
                </div>
              </template>
            </CardListView>
          </div>

          <!-- 神经操作栏 -->
          <div v-else-if="activeComponent === 'action'" class="demo-section">
            <p class="demo-label">基于选中项的上下文操作栏：</p>
            <div class="demo-wrapper demo-compact">
              <NeuroAgentActionBar
                :selected-items="actionSelectedItems"
                :ai-enabled="true"
                :show-batch-panel="true"
                @action="handleAction"
              />
            </div>
            <div class="action-demo-hint">
              <p>勾选上方列表中的项目，操作栏会显示批量操作选项。</p>
            </div>
          </div>
          </div>
        </div>
      </NeuroAgentPageShell>

      <div v-else class="showcase-empty-main">
        <p>无匹配组件，请调整搜索或「更多筛选」中的目录状态。</p>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { silentDebug } from '@/utils/debug'
import NeuroAgentListPage, { type FilterField, type TableColumn } from '@/views/components/NeuroAgentListPage.vue'
import NeuroAgentDetailPage from '@/views/components/NeuroAgentDetailPage.vue'
import NeuroAgentFormPage, { type FormConfig, type FormField } from '@/views/components/NeuroAgentFormPage.vue'
import NeuroAgentDialog from '@/views/components/NeuroAgentDialog.vue'
import NeuroAgentSearchBar from '@/views/components/NeuroAgentSearchBar.vue'
import NeuroAgentTable from '@/views/components/NeuroAgentTable.vue'
import NeuroAgentCard from '@/views/components/NeuroAgentCard.vue'
import NeuroAgentActionBar from '@/views/components/NeuroAgentActionBar.vue'
import CardListView from '@/views/components/CardListView.vue'
import NeuroAgentPageShell from '@/views/components/NeuroAgentPageShell.vue'
import { ref, computed, watch } from 'vue'

defineProps<{ theme: 'dark' | 'light' }>()

interface Action {
  id: string
  label: string
  icon: string
  description?: string
}

interface ComponentInfo {
  id: string
  name: string
  fullName: string
  icon: string
  description: string
  tags: string[]
  status: 'ready' | 'developing'
  statusText: string
}

const components: ComponentInfo[] = [
  { id: 'list', name: '智能列表页', fullName: 'NeuroAgentListPage', icon: '📊', description: '具有神经数据流背景的智能数据表格，支持 AI 搜索、筛选和批量操作', tags: ['数据表格', 'AI 搜索', '批量操作'], status: 'ready', statusText: '已就绪' },
  { id: 'list-el', name: 'el-table 标准模式', fullName: 'NeuroAgentListPage (el-table)', icon: '📋', description: '封装 Element Plus el-table，支持多选、索引列、状态映射、批量操作', tags: ['el-table', '多选', '批量操作'], status: 'ready', statusText: '已就绪' },
  { id: 'list-tree', name: 'el-table 树形模式', fullName: 'NeuroAgentListPage (树形)', icon: '🌳', description: '树形表格展示，适用于菜单管理、组织架构等层级数据，支持自定义插槽', tags: ['树形表格', '层级数据', '自定义插槽'], status: 'ready', statusText: '已就绪' },
  { id: 'detail', name: '全息详情页', fullName: 'NeuroAgentDetailPage', icon: '🔍', description: '3D 分层信息展示，AI 分析面板，时间线可视化', tags: ['详情展示', 'AI 分析', '时间线'], status: 'developing', statusText: '开发中' },
  { id: 'form', name: '智能表单页', fullName: 'NeuroAgentFormPage', icon: '📝', description: '自适应表单布局，AI 辅助填写，实时验证反馈', tags: ['表单', 'AI 辅助', '验证'], status: 'developing', statusText: '开发中' },
  { id: 'dialog', name: '全息弹窗', fullName: 'NeuroAgentDialog', icon: '💬', description: '极简磨砂弹窗，渐变装饰线，流畅动画', tags: ['弹窗', '多尺寸', '极简'], status: 'ready', statusText: '已就绪' },
  { id: 'search', name: '智能搜索栏', fullName: 'NeuroAgentSearchBar', icon: '🔎', description: '自然语言 AI 搜索，模糊匹配和语义理解', tags: ['搜索', 'NLP', '语义'], status: 'developing', statusText: '开发中' },
  { id: 'table', name: '智能数据表', fullName: 'NeuroAgentTable', icon: '📋', description: '动态列配置，AI 高亮关键数据', tags: ['表格', '动态列', '高亮'], status: 'developing', statusText: '开发中' },
  { id: 'card', name: '神经数据卡片', fullName: 'NeuroAgentCard', icon: '🃏', description: '悬浮卡片数据展示，神经脉冲动画', tags: ['卡片', '悬浮', '脉冲'], status: 'developing', statusText: '开发中' },
  { id: 'card-list', name: '卡片列表容器', fullName: 'CardListView', icon: '🗂️', description: '响应式网格卡片列表，支持搜索过滤、网格/列表切换、空状态', tags: ['卡片列表', '响应式', '搜索过滤'], status: 'ready', statusText: '已就绪' },
  { id: 'action', name: '神经操作栏', fullName: 'NeuroAgentActionBar', icon: '⚡', description: '上下文感知操作栏，智能推荐的操作', tags: ['操作栏', '上下文', '智能推荐'], status: 'developing', statusText: '开发中' },
]

const activeComponent = ref('list')
const compSearch = ref('')
const filtersExpanded = ref(false)
const statusSidebarFilter = ref<'all' | 'ready' | 'developing'>('all')

const showcaseListDemoIds = new Set(['list', 'list-el', 'list-tree'])
const showWorkbenchDirectorySearch = computed(() => !showcaseListDemoIds.has(activeComponent.value))

const filteredComponents = computed(() => {
  let list = [...components]
  if (statusSidebarFilter.value === 'ready') list = list.filter((c) => c.status === 'ready')
  if (statusSidebarFilter.value === 'developing') list = list.filter((c) => c.status === 'developing')
  const q = compSearch.value.trim().toLowerCase()
  if (q) {
    list = list.filter(
      (c) =>
        c.name.toLowerCase().includes(q)
        || c.fullName.toLowerCase().includes(q)
        || c.description.toLowerCase().includes(q)
        || c.tags.some((t) => t.toLowerCase().includes(q)),
    )
  }
  return list
})

const currentComp = computed(() => filteredComponents.value.find((c) => c.id === activeComponent.value))

watch(
  filteredComponents,
  (list) => {
    if (!list.length) return
    if (!list.some((c) => c.id === activeComponent.value)) {
      activeComponent.value = list[0].id
    }
  },
  { immediate: true },
)

function copyComponentName() {
  const c = currentComp.value
  if (!c) return
  const text = c.fullName
  void navigator.clipboard.writeText(text).then(
    () => alert(`已复制：${text}`),
    () => alert(`组件名：${text}`),
  )
}

// --- 弹窗 ---
const dialogVisible = ref(false)
const dialogSize = ref<'small' | 'medium' | 'large' | 'fullscreen'>('medium')
const dialogTitle = ref('弹窗')
const dialogIcon = ref('💬')
const dialogLoading = ref(false)

const openDialog = (size: typeof dialogSize.value) => {
  dialogSize.value = size
  const sizeLabels: Record<string, string> = { small: '小尺寸', medium: '中尺寸', large: '大尺寸', fullscreen: '全屏' }
  dialogTitle.value = `${sizeLabels[size] || size}`
  dialogIcon.value = '💬'
  dialogLoading.value = false
  dialogVisible.value = true
}
const openLoadingDialog = () => {
  dialogSize.value = 'medium'
  dialogTitle.value = '加载中'
  dialogIcon.value = '⏳'
  dialogLoading.value = true
  dialogVisible.value = true
}

// --- 列表 ---
/** 列表演示：主行固定 2 格（关键词 + 1 个筛选项），其余在「更多筛选」 */
const listDemoFilterFields: FilterField[] = [
  {
    key: 'status',
    label: '状态',
    type: 'select',
    placeholder: '全部',
    options: [
      { label: '全部', value: '' },
      { label: '进行中', value: 'active' },
      { label: '待审核', value: 'pending' },
      { label: '已暂停', value: 'paused' },
      { label: '已完成', value: 'completed' },
    ],
  },
  {
    key: 'priority',
    label: '优先级',
    type: 'select',
    placeholder: '全部',
    options: [
      { label: '全部', value: '' },
      { label: '高', value: '高' },
      { label: '中', value: '中' },
      { label: '低', value: '低' },
    ],
  },
  { key: 'nameHint', label: '项目名称', type: 'text', placeholder: '模糊匹配' },
]

const elTableDemoFilterFields: FilterField[] = [
  {
    key: 'status',
    label: '状态',
    type: 'select',
    placeholder: '全部',
    options: [
      { label: '全部', value: '' },
      { label: '启用', value: 'active' },
      { label: '禁用', value: 'inactive' },
      { label: '待审核', value: 'pending' },
    ],
  },
  {
    key: 'role',
    label: '角色',
    type: 'select',
    placeholder: '全部',
    options: [
      { label: '全部', value: '' },
      { label: '管理员', value: 'admin' },
      { label: '编辑', value: 'editor' },
      { label: '只读', value: 'viewer' },
    ],
  },
  { key: 'emailHint', label: '邮箱', type: 'text', placeholder: '邮箱片段' },
]

const treeDemoFilterFields: FilterField[] = [
  {
    key: 'type',
    label: '类型',
    type: 'select',
    placeholder: '全部',
    options: [
      { label: '全部', value: '' },
      { label: '目录', value: 'directory' },
      { label: '菜单', value: 'menu' },
      { label: '按钮', value: 'button' },
    ],
  },
  { key: 'pathHint', label: '路由', type: 'text', placeholder: '路径关键词' },
  { key: 'nameHint', label: '名称', type: 'text', placeholder: '菜单名称' },
]

/** 右侧：排序下拉（无标签，占位「请选择排序」）；视图与 CardListView 同款图标；树为单按钮展开/收起 */
const listDemoSortSelect = ref<string | undefined>(undefined)
const listDemoSortSelectOptions = [
  { label: '按创建时间（新→旧）', value: 'created_desc' },
  { label: '按创建时间（旧→新）', value: 'created_asc' },
  { label: '按项目名称 A-Z', value: 'name_asc' },
  { label: '按项目名称 Z-A', value: 'name_desc' },
  { label: '按优先级（高→低）', value: 'priority_desc' },
  { label: '按进度（高→低）', value: 'progress_desc' },
]
const listDemoViewLayout = ref<'grid' | 'list'>('list')

const elTableDemoSortSelect = ref<string | undefined>(undefined)
const elTableDemoSortSelectOptions = [
  { label: '按创建时间（新→旧）', value: 'created_desc' },
  { label: '按创建时间（旧→新）', value: 'created_asc' },
  { label: '按姓名 A-Z', value: 'name_asc' },
  { label: '按邮箱 A-Z', value: 'email_asc' },
  { label: '按角色', value: 'role' },
  { label: '按状态', value: 'status' },
]
const elTableDemoViewLayout = ref<'grid' | 'list'>('list')

const treeDemoSortSelect = ref<string | undefined>(undefined)
const treeDemoSortSelectOptions = [
  { label: '按菜单名称 A-Z', value: 'name_asc' },
  { label: '按菜单名称 Z-A', value: 'name_desc' },
  { label: '按排序号（小→大）', value: 'sort_asc' },
  { label: '按排序号（大→小）', value: 'sort_desc' },
  { label: '按类型（目录→按钮）', value: 'type_order' },
  { label: '按路由路径 A-Z', value: 'path_asc' },
]
const treeDemoViewLayout = ref<'grid' | 'list'>('list')
/** 单图标切换：true 表示当前为「全部展开」态，点击后图标切为「展开」语义即收起 */
const treeDemoTreeExpanded = ref(true)

const listColumns = ref<TableColumn[]>([
  { key: 'name', title: '项目名称', type: 'text', icon: '📝', sortable: true },
  { key: 'status', title: '状态', type: 'status', icon: '🔍', sortable: true },
  { key: 'progress', title: '进度', type: 'progress', icon: '📊', sortable: true },
  { key: 'priority', title: '优先级', type: 'text', icon: '🚀' },
  { key: 'tags', title: '标签', type: 'tags', icon: '🏷️' },
  { key: 'createdAt', title: '创建时间', type: 'date', icon: '🕒', sortable: true },
])
const listData = ref(Array.from({ length: 12 }, (_, i) => ({
  id: `project-${i + 1}`,
  name: `AI项目 ${i + 1}`,
  status: (['active', 'pending', 'paused', 'completed'] as const)[i % 4],
  progress: Math.floor(Math.random() * 100),
  priority: (['高', '中', '低'] as const)[i % 3],
  tags: ['AI', '前端', '后端', '设计'].slice(0, (i % 3) + 1),
  createdAt: new Date(Date.now() - i * 86400000),
})))

const handleListCreate = () => alert('创建')
const handleListView = (item: any) => alert(`查看: ${item.name}`)
const handleListEdit = (item: any) => alert(`编辑: ${item.name}`)
const handleListDelete = (item: any) => {
  if (confirm(`归档 "${item.name}"?`)) {
    listData.value = listData.value.filter(d => d.id !== item.id)
  }
}

// --- el-table 标准模式 ---
const elTableColumns = ref<TableColumn[]>([
  { key: 'name', title: '姓名', type: 'text', sortable: true },
  { key: 'email', title: '邮箱', type: 'text' },
  { key: 'role', title: '角色', type: 'tag', tagMap: {
    admin: { label: '管理员', type: 'danger' },
    editor: { label: '编辑', type: 'warning' },
    viewer: { label: '只读', type: 'info' },
  }},
  { key: 'status', title: '状态', type: 'status' },
  { key: 'createdAt', title: '创建时间', type: 'date', sortable: true },
])
const elTableStatusMap = { active: '启用', inactive: '禁用', pending: '待审核' }
const elTableData = ref(Array.from({ length: 15 }, (_, i) => ({
  id: `user-${i + 1}`,
  name: `用户 ${i + 1}`,
  email: `user${i + 1}@example.com`,
  role: (['admin', 'editor', 'viewer'] as const)[i % 3],
  status: (['active', 'inactive', 'pending'] as const)[i % 3],
  createdAt: new Date(Date.now() - i * 86400000),
})))

const handleElTableCreate = () => alert('新增用户')
const handleElTableView = (item: any) => alert(`查看: ${item.name}`)
const handleElTableEdit = (item: any) => alert(`编辑: ${item.name}`)
const handleElTableDelete = (item: any) => {
  if (confirm(`归档 "${item.name}"?`)) {
    elTableData.value = elTableData.value.filter(d => d.id !== item.id)
  }
}
const handleElTableBatchDelete = (items: any[]) => {
  if (confirm(`批量归档 ${items.length} 条记录?`)) {
    const ids = new Set(items.map(i => i.id))
    elTableData.value = elTableData.value.filter(d => !ids.has(d.id))
  }
}

// --- el-table 树形模式 ---
const treeTableColumns = ref<TableColumn[]>([
  { key: 'name', title: '菜单名称', type: 'text', sortable: true },
  { key: 'icon', title: '图标', type: 'text' },
  { key: 'path', title: '路由/权限', type: 'text' },
  { key: 'type', title: '类型', type: 'tag', tagMap: {
    directory: { label: '目录', type: '' },
    menu: { label: '菜单', type: 'success' },
    button: { label: '按钮', type: 'warning' },
  }},
  { key: 'showInSidebar', title: '侧栏', type: 'switch' },
  { key: 'sort', title: '排序', type: 'text', sortable: true },
])
const treeStatusMap = {}
const treeTableData = ref([
  {
    id: 'sys', name: '系统管理', icon: '⚙️', path: '/system', type: 'directory',
    showInSidebar: true, sort: 1,
    children: [
      {
        id: 'sys-user', name: '用户管理', icon: '👤', path: '/system/users', type: 'menu',
        showInSidebar: true, sort: 1,
        children: [
          { id: 'sys-user-add', name: '新增用户', icon: '', path: 'user:add', type: 'button', showInSidebar: false, sort: 1 },
          { id: 'sys-user-edit', name: '编辑用户', icon: '', path: 'user:edit', type: 'button', showInSidebar: false, sort: 2 },
          { id: 'sys-user-del', name: '删除用户', icon: '', path: 'user:delete', type: 'button', showInSidebar: false, sort: 3 },
        ],
      },
      {
        id: 'sys-role', name: '角色管理', icon: '🛡️', path: '/system/roles', type: 'menu',
        showInSidebar: true, sort: 2,
        children: [
          { id: 'sys-role-add', name: '新增角色', icon: '', path: 'role:add', type: 'button', showInSidebar: false, sort: 1 },
          { id: 'sys-role-edit', name: '编辑角色', icon: '', path: 'role:edit', type: 'button', showInSidebar: false, sort: 2 },
        ],
      },
      {
        id: 'sys-menu', name: '菜单管理', icon: '📋', path: '/system/menus', type: 'menu',
        showInSidebar: true, sort: 3,
        children: [
          { id: 'sys-menu-add', name: '新增菜单', icon: '', path: 'menu:add', type: 'button', showInSidebar: false, sort: 1 },
        ],
      },
    ],
  },
  {
    id: 'monitor', name: '系统监控', icon: '📊', path: '/monitor', type: 'directory',
    showInSidebar: true, sort: 2,
    children: [
      { id: 'monitor-online', name: '在线用户', icon: '👥', path: '/monitor/online', type: 'menu', showInSidebar: true, sort: 1 },
      { id: 'monitor-login', name: '登录日志', icon: '📝', path: '/monitor/login-logs', type: 'menu', showInSidebar: true, sort: 2 },
    ],
  },
])

const handleTreeCellChange = (_key: string, row: any) => silentDebug('tree cell change', row)
const handleTreeEdit = (item: any) => alert(`编辑菜单: ${item.name}`)

// --- 详情 ---
const detailData = ref({
  id: 'project-001',
  title: '神经矩阵指挥中心',
  name: 'NeuroMatrix Command Center',
  icon: '🧠',
  status: 'active',
  progress: 85,
  priority: 'high',
  owner: 'AI架构师',
  description: '基于 Cyber-Organic Neural Interface 美学的 AI 协作平台。',
  createdAt: new Date('2024-01-10'),
  updatedAt: new Date(),
  tags: ['AI', '神经美学', '3D可视化'],
  attachments: [
    { id: 1, name: '架构设计文档.pdf', type: 'pdf', size: '3.2MB' },
  ],
})

// --- 表单 ---
const formConfig = ref<FormConfig>({
  title: '创建AI项目',
  subtitle: '基于神经美学的AI协作项目管理',
  icon: '🧠',
  mode: 'create',
  showSteps: true,
  showAiAssistant: true,
  showBackButton: true,
  showResetButton: true,
  showSaveDraft: false,
  showPreviewButton: false,
  submitButtonText: '创建项目',
  footerText: '所有字段均为AI推荐的最佳实践配置',
})
const formFields = ref<FormField[]>([
  { id: 'name', label: '项目名称', type: 'text', required: true, placeholder: '请输入项目名称' },
  { id: 'desc', label: '项目描述', type: 'textarea', required: true, placeholder: '请输入描述' },
  { id: 'type', label: '项目类型', type: 'select', required: true, options: [
    { value: 'ai-research', label: 'AI研究' },
    { value: 'product', label: '产品开发' },
    { value: 'infra', label: '基础设施' },
  ]},
  { id: 'priority', label: '优先级', type: 'radio-group', required: true, options: [
    { value: 'high', label: '高' }, { value: 'medium', label: '中' }, { value: 'low', label: '低' },
  ]},
  { id: 'startDate', label: '开始日期', type: 'date', required: true },
  { id: 'endDate', label: '结束日期', type: 'date', required: true },
])
const formSteps = ref([
  { id: 'basic', label: '基本信息', fields: ['name', 'desc', 'type'] },
  { id: 'settings', label: '高级设置', fields: ['priority', 'startDate', 'endDate'] },
])
const handleFormSubmit = (data: any) => alert(`创建: ${data.name}`)

// --- 表格 ---
const tableColumns = ref([
  { key: 'name', title: '名称', icon: '📝' },
  { key: 'status', title: '状态', icon: '🔍' },
  { key: 'progress', title: '进度', icon: '📊' },
  { key: 'owner', title: '负责人', icon: '👤' },
])
const tableData = ref([
  { name: '项目 Alpha', status: '进行中', progress: '72%', owner: '张三' },
  { name: '项目 Beta', status: '待启动', progress: '0%', owner: '李四' },
  { name: '项目 Gamma', status: '已完成', progress: '100%', owner: '王五' },
])

// --- 搜索 ---
const handleSearch = (query: string) => alert(`搜索: ${query}`)

// --- 卡片 ---
const cardData = ref([
  { id: 'card-1', title: '项目 Alpha', description: 'AI 驱动的数据平台', icon: '🧠', status: 'active' },
  { id: 'card-2', title: '项目 Beta', description: '神经美学设计系统', icon: '🎨', status: 'pending' },
  { id: 'card-3', title: '项目 Gamma', description: '实时协作引擎', icon: '⚡', status: 'completed' },
])

// --- 卡片列表 ---
const cardListData = ref([
  { id: 1, title: '数据中台', subtitle: '核心业务', description: '统一数据采集、清洗、建模，为业务提供标准数据服务', icon: '📊', status: 'active', priority: 'high', progress: 72, tags: ['数据', 'ETL'], createdAt: '2025-11-20', updatedAt: new Date() },
  { id: 2, title: '用户增长平台', subtitle: '增长团队', description: '基于行为分析的自动化用户分层与精准触达', icon: '🚀', status: 'active', priority: 'high', progress: 58, tags: ['增长', '分析'], createdAt: '2025-12-01', viewCount: 1230 },
  { id: 3, title: 'AI 智能客服', subtitle: '创新实验室', description: '大语言模型驱动的多渠道智能客服系统', icon: '🤖', status: 'pending', priority: 'medium', progress: 34, tags: ['AI', 'NLP', '客服'], createdAt: '2026-01-10', likeCount: 56 },
  { id: 4, title: '权限中心 2.0', subtitle: '基础架构', description: 'RBAC + ABAC 混合权限模型，支持细粒度数据权限', icon: '🔐', status: 'active', priority: 'critical', progress: 88, tags: ['权限', '安全'], createdAt: '2025-09-15', commentCount: 23 },
  { id: 5, title: '监控告警平台', subtitle: '运维', description: '全链路监控与智能告警，支持钉钉、企业微信通知', icon: '📡', status: 'completed', priority: 'low', progress: 100, tags: ['监控', '运维'], createdAt: '2025-06-20', updatedAt: '2026-03-01' },
  { id: 6, title: '移动端 H5 重构', subtitle: '前端团队', description: '从 Vue 2 + Vuex 迁移至 Vue 3 + Pinia，性能提升 40%', icon: '📱', status: 'paused', priority: 'medium', progress: 45, tags: ['前端', '移动端'], createdAt: '2026-02-28', viewCount: 890 },
  { id: 7, title: '财务对账系统', subtitle: '财务', description: '多支付渠道自动对账，异常订单智能匹配', icon: '💰', status: 'draft', priority: 'medium', progress: 12, tags: ['财务', '自动化'], createdAt: '2026-04-01' },
  { id: 8, title: '内容管理中台', subtitle: '运营', description: '统一 CMS 能力，支持多站点、多语言、多渠道发布', icon: '📝', status: 'active', priority: 'low', progress: 65, tags: ['CMS', '运营'], createdAt: '2026-01-15', commentCount: 8 },
])

const cardListFilters = ref([
  { label: '全部', value: '' },
  { label: '进行中', value: 'active', count: 4 },
  { label: '待处理', value: 'pending', count: 1 },
  { label: '已完成', value: 'completed', count: 1 },
  { label: '已暂停', value: 'paused', count: 1 },
  { label: '草稿', value: 'draft', count: 1 },
])

const handleCardClick = (card: any) => silentDebug('clicked card:', card)

// --- 操作栏 ---
const actionSelectedItems = ref(['project-1', 'project-2'])
const handleAction = (_action: Action, _items?: any[]) => {}
</script>

<style scoped>
.showcase-content-wrapper {
  display: flex;
  gap: 20px;
  padding: 20px;
  min-height: calc(100vh - 90px);
  background: var(--showcase-bg, #0a0c14);
  color: var(--showcase-text, #e2e8f0);
}

/* 左侧边栏 */
.showcase-sidebar {
  width: 240px;
  flex-shrink: 0;
  background: var(--nm-bg-card, rgba(15, 23, 42, 0.7));
  backdrop-filter: blur(20px);
  border: 1px solid var(--nm-border-glow, rgba(0, 245, 212, 0.2));
  border-radius: 16px;
  padding: 16px;
  align-self: flex-start;
  position: sticky;
  top: 110px;
  max-height: calc(100vh - 130px);
  display: flex;
  flex-direction: column;
}

.sidebar-header {
  padding: 8px 8px 16px;
  border-bottom: 1px solid var(--nm-border, rgba(100, 116, 139, 0.2));
  margin-bottom: 8px;
}

.sidebar-title {
  font-family: 'Orbitron', 'JetBrains Mono', monospace;
  font-size: 18px;
  font-weight: 700;
  background: linear-gradient(135deg, var(--nm-primary, #00f5d4), var(--nm-secondary, #9d4edd));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  margin: 0;
}

.sidebar-count {
  color: var(--nm-text-muted, #64748b);
  font-size: 12px;
  margin: 4px 0 0;
}

.component-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
  overflow-y: auto;
  flex: 1;
}

.component-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  background: transparent;
  border: 1px solid transparent;
  border-radius: 10px;
  color: var(--nm-text-secondary, #94a3b8);
  cursor: pointer;
  transition: all 0.2s ease;
  text-align: left;
  width: 100%;
  font-family: inherit;
}

.component-item:hover {
  background: var(--nm-primary-glow, rgba(0, 245, 212, 0.05));
  border-color: rgba(0, 245, 212, 0.15);
}

.component-active {
  background: rgba(0, 245, 212, 0.1);
  border-color: rgba(0, 245, 212, 0.3);
}

.comp-icon { font-size: 18px; flex-shrink: 0; }

.comp-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.comp-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--nm-text-primary, #e2e8f0);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.component-active .comp-name {
  color: var(--nm-primary, #00f5d4);
}

.comp-status {
  font-size: 10px;
  padding: 1px 6px;
  border-radius: 6px;
  width: fit-content;
}

.status-ready {
  color: #10b981;
  background: rgba(16, 185, 129, 0.1);
}

.status-developing {
  color: #f59e0b;
  background: rgba(245, 158, 11, 0.1);
}

/* 右侧内容 */
.showcase-main {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.showcase-page-shell {
  flex: 1 1 0;
  min-height: 0;
  min-width: 0;
}

.showcase-hero-title {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  font-weight: inherit;
  font-size: inherit;
}

.showcase-hero-icon {
  font-size: 1.15em;
  line-height: 1;
}

.showcase-status-pill {
  display: inline-flex;
  align-items: center;
  padding: 2px 10px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 600;
}

.showcase-status--ready {
  color: #10b981;
  background: rgba(16, 185, 129, 0.12);
  border: 1px solid rgba(16, 185, 129, 0.35);
}

.showcase-status--developing {
  color: #f59e0b;
  background: rgba(245, 158, 11, 0.12);
  border: 1px solid rgba(245, 158, 11, 0.35);
}

.showcase-meta-tag {
  font-size: 11px;
  padding: 3px 10px;
  background: rgba(0, 245, 212, 0.08);
  border: 1px solid rgba(0, 245, 212, 0.2);
  border-radius: 12px;
  color: var(--nm-primary, #00f5d4);
}

.showcase-copy-btn {
  padding: 8px 14px;
  border-radius: var(--neuro-radius-md, 8px);
  border: 1px solid var(--neuro-border, rgba(100, 116, 139, 0.35));
  background: var(--neuro-surface, rgba(15, 23, 42, 0.6));
  color: var(--neuro-text, #e2e8f0);
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: border-color 0.15s ease, color 0.15s ease;
}

.showcase-copy-btn:hover {
  border-color: var(--neuro-primary, #00f5d4);
  color: var(--neuro-primary, #00f5d4);
}

.showcase-workbench {
  flex: 1 1 0;
  min-height: 0;
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 12px 16px 16px;
  box-sizing: border-box;
}

.showcase-search-strip {
  display: flex;
  flex-wrap: nowrap;
  align-items: center;
  gap: 10px;
  flex-shrink: 0;
}

.showcase-search-wrap {
  position: relative;
  flex: 1 1 280px;
  min-width: 120px;
  max-width: 520px;
}

.showcase-search-icon {
  position: absolute;
  left: 12px;
  top: 50%;
  transform: translateY(-50%);
  width: 18px;
  height: 18px;
  color: var(--neuro-text-secondary, #94a3b8);
  pointer-events: none;
}

.showcase-search-input {
  width: 100%;
  height: 40px;
  padding: 0 36px 0 40px;
  box-sizing: border-box;
  border: 1px solid var(--neuro-border, rgba(100, 116, 139, 0.35));
  border-radius: var(--neuro-radius-lg, 12px);
  background: var(--neuro-surface, rgba(15, 23, 42, 0.75));
  color: var(--neuro-text, #e2e8f0);
  font-size: 14px;
  font-family: var(--neuro-font-sans, system-ui, sans-serif);
  outline: none;
  transition: border-color 0.15s ease, box-shadow 0.15s ease;
}

.showcase-search-input:focus {
  border-color: var(--neuro-primary, #00f5d4);
  box-shadow: 0 0 0 2px rgba(0, 245, 212, 0.12);
}

.showcase-search-input::placeholder {
  color: var(--neuro-text-secondary, #64748b);
}

.showcase-search-clear {
  position: absolute;
  right: 8px;
  top: 50%;
  transform: translateY(-50%);
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  padding: 0;
  border: none;
  border-radius: 50%;
  background: transparent;
  color: var(--neuro-text-secondary, #94a3b8);
  cursor: pointer;
}

.showcase-search-clear:hover {
  color: var(--neuro-text, #e2e8f0);
  background: rgba(0, 245, 212, 0.08);
}

.showcase-filter-toggle {
  display: inline-flex;
  flex-shrink: 0;
  align-items: center;
  gap: 6px;
  padding: 8px 12px;
  border-radius: 999px;
  border: 1px solid var(--neuro-border, rgba(100, 116, 139, 0.35));
  background: var(--neuro-surface, rgba(15, 23, 42, 0.6));
  color: var(--neuro-text-secondary, #94a3b8);
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s ease;
}

.showcase-filter-toggle:hover {
  border-color: var(--neuro-primary-30, rgba(0, 245, 212, 0.35));
  color: var(--neuro-primary, #00f5d4);
}

.showcase-filter-toggle__chev {
  transition: transform 0.2s ease;
}

.showcase-filter-toggle__chev.is-open {
  transform: rotate(180deg);
}

.showcase-filters-extra {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 10px 14px;
  padding: 10px 12px;
  border-radius: var(--neuro-radius-lg, 12px);
  border: 1px solid var(--neuro-primary-10, rgba(0, 245, 212, 0.12));
  background: color-mix(in srgb, var(--neuro-surface, #0f172a) 90%, var(--neuro-primary-10, rgba(0, 245, 212, 0.08)));
}

.showcase-filter-label {
  font-size: 12px;
  font-weight: 700;
  color: var(--neuro-text-secondary, #94a3b8);
}

.showcase-filter-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.showcase-filter-chip {
  padding: 6px 12px;
  border-radius: 999px;
  border: 1px solid var(--neuro-border, rgba(100, 116, 139, 0.35));
  background: transparent;
  color: var(--neuro-text-secondary, #94a3b8);
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s ease;
}

.showcase-filter-chip:hover {
  border-color: var(--neuro-primary-30, rgba(0, 245, 212, 0.35));
  color: var(--neuro-text, #e2e8f0);
}

.showcase-filter-chip.active {
  border-color: var(--neuro-primary, #00f5d4);
  color: var(--neuro-primary, #00f5d4);
  background: rgba(0, 245, 212, 0.1);
}

.showcase-empty-main {
  flex: 1;
  display: grid;
  place-items: center;
  padding: 48px 24px;
  color: var(--neuro-text-secondary, #94a3b8);
  font-size: 14px;
}

.card-list-demo-extra {
  width: 100%;
}

.card-list-demo-extra__hint {
  font-size: 12px;
  color: var(--neuro-text-secondary, #94a3b8);
  line-height: 1.45;
}

/* 演示区域 */
.demo-area {
  flex: 1 1 0;
  min-height: 0;
  overflow: auto;
  background: var(--nm-bg-card, rgba(15, 23, 42, 0.5));
  border: 1px solid var(--nm-border-glow, rgba(0, 245, 212, 0.15));
  border-radius: 16px;
  padding: 24px;
}

.demo-section { animation: fadeIn 0.3s ease; }

.demo-wrapper {
  position: relative;
  overflow-y: auto;
  overflow-x: hidden;
  max-height: 700px;
  border: none;
  border-radius: 0;
}

.demo-wrapper :deep(.neuro-data-stream),
.demo-wrapper :deep(.neural-data-stream),
.demo-wrapper :deep(.neuro-particle-field),
.demo-wrapper :deep(.neuro-connection-grid),
.demo-wrapper :deep(.neuro-bg-layer) {
  position: absolute !important;
  height: 100% !important;
}

.demo-wrapper :deep(.neuro-agent-list-page),
.demo-wrapper :deep(.neuro-agent-detail-page),
.demo-wrapper :deep(.neuro-agent-form-page),
.demo-wrapper :deep(.neuro-agent-table),
.demo-wrapper :deep(.neuro-table) {
  min-height: 0 !important;
  height: 100% !important;
  position: relative !important;
}

.demo-wrapper :deep(.neuro-agent-list-page .neuro-bg-layer),
.demo-wrapper :deep(.neuro-agent-detail-page .neuro-bg-layer),
.demo-wrapper :deep(.neuro-agent-form-page .neuro-bg-layer),
.demo-wrapper :deep(.neuro-agent-table .neuro-bg-layer),
.demo-wrapper :deep(.neuro-table .neuro-bg-layer) {
  position: absolute !important;
}

.demo-wrapper :deep(.neuro-assistant-panel) {
  position: absolute !important;
  right: 8px !important;
  bottom: 8px !important;
  width: 260px !important;
}

.demo-wrapper :deep(.ai-assistant-float) {
  position: absolute !important;
  right: 8px !important;
  bottom: 8px !important;
  width: 260px !important;
}

.demo-wrapper :deep(.form-intelligence-background) { position: absolute !important; }
.demo-wrapper :deep(.form-success-overlay) { position: absolute !important; }
.demo-wrapper :deep(.confirm-dialog-overlay) { position: absolute !important; }
.demo-wrapper :deep(.ai-analysis-panel-overlay) { position: absolute !important; }
.demo-wrapper :deep(.neuro-dialog-overlay) { position: absolute !important; }

.demo-compact {
  max-height: none !important;
  border: none !important;
  border-radius: 0 !important;
  overflow: visible !important;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(8px); }
  to { opacity: 1; transform: translateY(0); }
}

.dialog-demo-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-bottom: 20px;
}

.demo-btn {
  padding: 8px 18px;
  background: rgba(0, 245, 212, 0.1);
  border: 1px solid rgba(0, 245, 212, 0.3);
  border-radius: 10px;
  color: var(--nm-primary, #00f5d4);
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s ease;
  font-family: inherit;
}

.demo-btn:hover {
  background: rgba(0, 245, 212, 0.2);
  transform: translateY(-1px);
}

.dialog-demo-body {
  padding: 20px;
  color: var(--nm-text-secondary, #94a3b8);
  line-height: 1.8;
}

.dialog-demo-body ul { padding-left: 20px; margin: 12px 0 0; }

/* NeuroAgentListPage：搜索行右侧演示（排序下拉 / 层级展开 / 与 CardListView 一致的视图图标） */
.na-demo-search-trailing {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 12px;
  width: 100%;
}

.na-demo-search-trailing__main {
  display: flex;
  align-items: center;
  gap: 10px;
  flex: 1 1 auto;
  min-width: 0;
}

.na-demo-search-trailing__view {
  flex: 0 0 auto;
  margin-left: auto;
}

.na-demo-sort-select {
  flex: 1 1 160px;
  max-width: 280px;
  min-width: 140px;
}

.na-demo-hierarchy-toggle {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  min-height: 34px;
  padding: 0 12px;
  flex-shrink: 0;
  border-radius: var(--neuro-radius-md, 8px);
  border: 1px solid var(--neuro-border, #334155);
  background: var(--neuro-surface, #1e293b);
  color: var(--neuro-text-secondary, #94a3b8);
  font-size: 13px;
  font-weight: 600;
  font-family: inherit;
  cursor: pointer;
  transition: color 0.15s, border-color 0.15s, background 0.15s;
}

.na-demo-hierarchy-toggle__text {
  white-space: nowrap;
}

.na-demo-hierarchy-toggle__chev {
  flex-shrink: 0;
}

.na-demo-hierarchy-toggle:hover {
  color: var(--neuro-primary, #00f5d4);
  border-color: color-mix(in srgb, var(--neuro-primary, #00f5d4) 45%, transparent);
}

.na-demo-view-switcher {
  display: flex;
  background: var(--neuro-surface, #1e293b);
  border: 1px solid var(--neuro-border, #334155);
  border-radius: var(--neuro-radius-md, 8px);
  overflow: hidden;
}

.na-demo-view-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 34px;
  background: transparent;
  border: none;
  color: var(--neuro-text-secondary, #94a3b8);
  cursor: pointer;
  transition: color 0.15s, background 0.15s;
}

.na-demo-view-btn:hover {
  color: var(--neuro-primary, #00f5d4);
}

.na-demo-view-btn.na-demo-view-btn--active {
  background: color-mix(in srgb, var(--neuro-primary, #00f5d4) 20%, transparent);
  color: var(--neuro-primary, #00f5d4);
}

.na-demo-view-btn + .na-demo-view-btn {
  border-left: 1px solid var(--neuro-border, #334155);
}

.search-hint {
  margin-top: 20px;
  padding: 16px;
  background: rgba(0, 245, 212, 0.05);
  border: 1px solid rgba(0, 245, 212, 0.1);
  border-radius: 10px;
}

.search-hint p { color: var(--nm-text-secondary, #94a3b8); font-size: 13px; margin: 0; }

.card-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 16px;
}

/* 响应式 */
@media (max-width: 900px) {
  .showcase-content-wrapper { flex-direction: column; }
  .showcase-sidebar { width: 100%; position: static; max-height: none; }
  .component-list { flex-direction: row; flex-wrap: wrap; }
  .component-item { width: auto; }
}

/* 浅色模式 */
.showcase-content-wrapper[data-theme="light"] {
  --showcase-bg: #f8fafc;
  --showcase-text: #1e293b;
  --showcase-text-secondary: #475569;
  --showcase-text-muted: #64748b;
  --nm-bg-card: rgba(248, 250, 252, 0.9);
  --nm-border: rgba(14, 165, 233, 0.15);
  --nm-border-glow: rgba(14, 165, 233, 0.2);
  --nm-primary: #0ea5e9;
  --nm-primary-glow: rgba(14, 165, 233, 0.05);
  --nm-text-primary: #1e293b;
  --nm-text-secondary: #475569;
  --nm-text-muted: #64748b;
}

.showcase-content-wrapper[data-theme="light"] .sidebar-title {
  background: linear-gradient(135deg, #0ea5e9, #8b5cf6);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

.showcase-content-wrapper[data-theme="light"] .component-item:hover {
  background: rgba(14, 165, 233, 0.05);
  border-color: rgba(14, 165, 233, 0.15);
}

.showcase-content-wrapper[data-theme="light"] .component-active {
  background: rgba(14, 165, 233, 0.1);
  border-color: rgba(14, 165, 233, 0.3);
}

.showcase-content-wrapper[data-theme="light"] .comp-name { color: #1e293b; }
.showcase-content-wrapper[data-theme="light"] .component-active .comp-name { color: #0ea5e9; }
.showcase-content-wrapper[data-theme="light"] .panel-title { color: #0ea5e9; }
.showcase-content-wrapper[data-theme="light"] .panel-desc { color: #475569; }
.showcase-content-wrapper[data-theme="light"] .tag {
  color: #0ea5e9;
  background: rgba(14, 165, 233, 0.08);
  border-color: rgba(14, 165, 233, 0.2);
}

.showcase-content-wrapper[data-theme="light"] .demo-area {
  background: rgba(248, 250, 252, 0.7);
  border-color: rgba(14, 165, 233, 0.15);
}

.showcase-content-wrapper[data-theme="light"] .demo-btn {
  color: #0ea5e9;
  background: rgba(14, 165, 233, 0.1);
  border-color: rgba(14, 165, 233, 0.3);
}

.showcase-content-wrapper[data-theme="light"] .sidebar-count { color: #64748b; }
.showcase-content-wrapper[data-theme="light"] .search-hint p { color: #475569; }
.showcase-content-wrapper[data-theme="light"] .dialog-demo-body { color: #475569; }
</style>
