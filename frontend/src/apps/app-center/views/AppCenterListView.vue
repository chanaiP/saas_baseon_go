<script setup lang="ts">
defineOptions({ name: 'AppCenterListView' })

import { Box, Grid, Monitor, Plus, Refresh, Search, Setting } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'

import { fetchDictItemsByCode, type DictItemRow } from '@/api/dict'
import NeuroAgentDialog from '@/views/components/NeuroAgentDialog.vue'
import NeuroAgentPageShell from '@/views/components/NeuroAgentPageShell.vue'

import { createAppCenterApp, fetchAppCenterApp, fetchAppCenterApps, fetchAppCenterStats, updateAppCenterApp, updateAppCenterAppStatus } from '../api'
import type { AppCenterApp, AppCenterCreatePayload, AppCenterStats, AppCenterUpdatePayload } from '../types'

const loading = ref(false)
const apps = ref<AppCenterApp[]>([])
const total = ref(0)
const keyword = ref('')
const typeFilter = ref('')
const statusFilter = ref('')
const sourceFilter = ref('')
const activeSection = ref('apps')
const route = useRoute()
const createDialogVisible = ref(false)
const editDialogVisible = ref(false)
const detailDialogVisible = ref(false)
const saving = ref(false)
const detailLoading = ref(false)
const detailApp = ref<AppCenterApp | null>(null)
const createForm = ref<AppCenterCreatePayload>(newCreateForm())
const createStep = ref('basic')
const createScrollRef = ref<HTMLElement | null>(null)
const createDraft = ref(newCreateDraft())
const editAppId = ref<number | null>(null)
const editForm = ref<AppCenterUpdatePayload>(newEditForm())
const emptyStats: AppCenterStats = {
  total: 0,
  online: 0,
  beta: 0,
  developing: 0,
  builtin: 0,
  disabled: 0,
  categories: 0,
  client_apps: 0,
  tenant_openings: 0,
  trial_invites: 0,
  manifest_loads: 0,
  audit_logs: 0,
}
const stats = ref<AppCenterStats>({ ...emptyStats })

const dicts = ref<Record<string, DictItemRow[]>>({
  app_type: [],
  app_status: [],
  app_source: [],
  app_charge_mode: [],
  app_visibility_scope: [],
})

const fallbackLabels: Record<string, Record<string, string>> = {
  app_type: {
    SYSTEM_APP: '系统内置型',
    ABILITY_APP: '业务中台型',
    BUSINESS_APP: '独立业务型',
    SUITE_APP: '组合套件型',
    CONNECTOR_APP: '连接器型',
    CLIENT_APP: '客户端型',
    AI_APP: 'AI / Agent 型',
    API_APP: 'API 能力型',
  },
  app_status: {
    DRAFT: '草稿',
    PLANNED: '规划中',
    DEVELOPING: '开发中',
    BETA: 'Beta',
    ONLINE: '已上线',
    DISABLED: '已停用',
    ARCHIVED: '已归档',
  },
  app_source: {
    BUILTIN: '系统内置',
    MANUAL: '手工创建',
    MANIFEST: '声明文件装载',
  },
  app_charge_mode: {
    FREE: '免费',
    SUBSCRIPTION: '订阅制',
    BUYOUT: '买断制',
    USAGE_BASED: '按量收费',
    MIXED: '组合计费',
    NON_SELLABLE: '非售卖',
  },
  app_visibility_scope: {
    PLATFORM_ONLY: '仅平台',
    TENANT: '租户可用',
    GLOBAL: '全局可见',
  },
}

const statCards = computed(() => [
  { label: '所有应用', value: stats.value.total, status: '', tone: 'default', hint: '全部应用' },
  { label: '已上线', value: stats.value.online, status: 'ONLINE', tone: 'success', hint: '正式可用' },
  { label: 'Beta 版', value: stats.value.beta, status: 'BETA', tone: 'warning', hint: '测试验证中' },
  { label: '规划开发中', value: stats.value.developing, status: 'PLANNED,DEVELOPING', tone: 'info', hint: '建设中' },
  { label: '已停用', value: stats.value.disabled, status: 'DISABLED', tone: 'muted', hint: '不可选用' },
])

const statusFilterOptions = computed(() => [
  { label: '规划开发中', value: 'PLANNED,DEVELOPING' },
  ...dictOptions('app_status'),
])

const createSteps = [
  { key: 'basic', title: '基础信息', hint: 'app_code、类型、状态' },
  { key: 'access', title: '可见与开通', hint: '范围、收费、试用' },
  { key: 'clients', title: '客户端', hint: 'PC Web、API、移动端' },
  { key: 'assets', title: '入口/API/权限', hint: '应用资产草稿' },
  { key: 'docs', title: '文档与版本', hint: '知识资产骨架' },
  { key: 'review', title: '确认创建', hint: '生成应用主档' },
]

const selectedCreateClients = computed(() => {
  return createDraft.value.clients.filter((item) => item.checked).map((item) => item.label)
})
const selectedCreateAssets = computed(() => {
  return createDraft.value.assets.filter((item) => item.checked).map((item) => item.label)
})
const selectedCreateDocs = computed(() => {
  return createDraft.value.docs.filter((item) => item.checked).map((item) => item.label)
})
const createStepCompletion = computed<Record<string, boolean>>(() => ({
  basic: Boolean(createForm.value.app_code.trim() && createForm.value.app_name.trim()),
  access: Boolean(createForm.value.visibility_scope && createForm.value.charge_mode && createDraft.value.visibility_mode),
  clients: selectedCreateClients.value.length > 0,
  assets: selectedCreateAssets.value.length > 0,
  docs: Boolean(createDraft.value.version_no.trim() && createDraft.value.release_note.trim() && selectedCreateDocs.value.length > 0),
  review: Boolean(createForm.value.app_code.trim() && createForm.value.app_name.trim()),
}))

const groupedApps = computed(() => {
  const order = new Map(dictOptions('app_type').map((item, index) => [item.value, index]))
  const groups = new Map<string, AppCenterApp[]>()
  apps.value.forEach((app) => {
    const key = app.app_type || 'UNKNOWN'
    const rows = groups.get(key) || []
    rows.push(app)
    groups.set(key, rows)
  })
  return Array.from(groups.entries())
    .sort(([a], [b]) => {
      const aOrder = order.get(a) ?? Number.MAX_SAFE_INTEGER
      const bOrder = order.get(b) ?? Number.MAX_SAFE_INTEGER
      if (aOrder !== bOrder) return aOrder - bOrder
      return labelOf('app_type', a).localeCompare(labelOf('app_type', b), 'zh-Hans-CN')
    })
    .map(([type, items]) => ({
      type,
      label: labelOf('app_type', type),
      items,
    }))
})

const hasActiveFilters = computed(() => {
  return Boolean(keyword.value.trim() || typeFilter.value || statusFilter.value || sourceFilter.value)
})

const directoryItems = computed(() => [
  {
    key: 'apps',
    label: '应用列表',
    count: stats.value.total,
    description: '平台已登记的内置应用、业务应用与装载应用。',
  },
  {
    key: 'clients',
    label: '客户端中心',
    count: stats.value.client_apps,
    description: '类型为客户端型的应用数量。',
  },
  {
    key: 'openings',
    label: '租户开通总览',
    count: stats.value.tenant_openings,
    description: '租户套餐订阅与开通记录总数。',
  },
  {
    key: 'invites',
    label: '体验邀请总览',
    count: stats.value.trial_invites,
    description: '试用订阅与体验邀请相关记录。',
  },
  {
    key: 'manifest',
    label: 'Manifest 装载记录',
    count: stats.value.manifest_loads,
    description: '通过 Manifest 声明文件装载的应用。',
  },
  {
    key: 'audit',
    label: '应用审计日志',
    count: stats.value.audit_logs,
    description: '应用中心相关审计事件。',
  },
])

const activeDirectoryItem = computed(() => directoryItems.value.find((item) => item.key === activeSection.value) || directoryItems.value[0])

const sectionRouteMap: Record<string, string> = {
  apps: '/apps',
  clients: '/apps/clients',
  openings: '/apps/tenant-openings',
  invites: '/apps/trial-invites',
  manifest: '/apps/manifests',
  audit: '/apps/audit-logs',
}

function sectionFromPath(path: string) {
  const hit = Object.entries(sectionRouteMap).find(([, routePath]) => routePath === path)
  return hit?.[0] || 'apps'
}

function dictOptions(code: keyof typeof dicts.value) {
  const rows = dicts.value[code] || []
  if (rows.length > 0) return rows
  return Object.entries(fallbackLabels[code] || {}).map(([value, label], index) => ({
    id: index,
    dict_type_id: 0,
    label,
    value,
    sort_order: index + 1,
    enabled: true,
  }))
}

function labelOf(code: keyof typeof dicts.value, value: string | null | undefined) {
  if (!value) return '—'
  const row = dictOptions(code).find((item) => item.value === value)
  return row?.label || fallbackLabels[code]?.[value] || value
}

function statusClass(status: string) {
  return {
    'app-status--online': status === 'ONLINE',
    'app-status--developing': status === 'DEVELOPING' || status === 'BETA',
    'app-status--disabled': status === 'DISABLED' || status === 'ARCHIVED',
    'app-status--draft': status === 'DRAFT' || status === 'PLANNED',
  }
}

function iconFor(app: AppCenterApp) {
  if (app.app_code === 'system-management') return Setting
  if (app.app_code === 'system-monitor') return Monitor
  if (app.app_code === 'app-center') return Grid
  return Box
}

function appInitial(app: AppCenterApp) {
  return (app.app_name || app.app_code || '应用').trim().slice(0, 1)
}

function newCreateForm(): AppCenterCreatePayload {
  return {
    app_code: '',
    app_name: '',
    icon: '',
    app_type: 'BUSINESS_APP',
    status: 'DRAFT',
    charge_mode: 'SUBSCRIPTION',
    visibility_scope: 'PLATFORM_ONLY',
    owner: '',
    version: '0.1.0',
    description: '',
    sort_order: 0,
  }
}

function newCreateDraft() {
  return {
    app_position: 'BUSINESS_SCENARIO',
    visibility_mode: 'SPECIFIED_TENANTS',
    visible_tenants: '租户 A、租户 B、演示主体',
    open_method: 'INVITE_CODE,ADMIN_GRANT',
    trial_days: '不限期',
    trial_start_rule: '首次安装时开始',
    clients: [
      { key: 'PC_WEB', label: 'PC Web', checked: true },
      { key: 'API_ONLY', label: 'API Only', checked: true },
      { key: 'H5', label: 'H5', checked: false },
      { key: 'IOS', label: 'iOS', checked: false },
      { key: 'ANDROID', label: 'Android', checked: false },
      { key: 'WINDOWS', label: 'Windows', checked: false },
      { key: 'MACOS', label: 'macOS', checked: false },
      { key: 'MINIAPP', label: '小程序', checked: false },
      { key: 'WEWORK_DINGTALK', label: '企微 / 钉钉 / 飞书', checked: false },
    ],
    assets: [
      { key: 'entry_dashboard', label: '创建应用工作台入口', checked: true },
      { key: 'entry_manage', label: '创建管理列表入口', checked: true },
      { key: 'api_query', label: '创建查询 API 草稿', checked: true },
      { key: 'permission_view', label: '创建查看权限点', checked: true },
      { key: 'permission_manage', label: '创建管理权限点', checked: false },
      { key: 'package_resource', label: '生成套餐资源草稿', checked: true },
    ],
    version_no: '0.1.0',
    release_channel: 'DEV',
    release_note: '创建应用草稿，建立基础信息、客户端、入口、权限、文档与初始版本骨架。',
    docs: [
      { key: 'prd', label: '需求文档占位', checked: true },
      { key: 'design', label: '技术设计占位', checked: true },
      { key: 'api_doc', label: 'API 文档占位', checked: true },
      { key: 'manual', label: '操作手册占位', checked: true },
      { key: 'ai_trace', label: 'AI 生成来源记录', checked: true },
      { key: 'release_note', label: '版本发布说明', checked: true },
    ],
  }
}

function newEditForm(): AppCenterUpdatePayload {
  return {
    app_name: '',
    icon: '',
    app_type: 'BUSINESS_APP',
    charge_mode: 'SUBSCRIPTION',
    visibility_scope: 'PLATFORM_ONLY',
    owner: '',
    version: '',
    description: '',
    sort_order: 0,
  }
}

function appToEditForm(app: AppCenterApp): AppCenterUpdatePayload {
  return {
    app_name: app.app_name,
    icon: app.icon || '',
    app_type: app.app_type,
    charge_mode: app.charge_mode,
    visibility_scope: app.visibility_scope,
    owner: app.owner || '',
    version: app.version || '',
    description: app.description || '',
    sort_order: app.sort_order,
  }
}

function trimNullable(value: string | null | undefined) {
  const trimmed = String(value ?? '').trim()
  return trimmed || null
}

function normalizeUpdatePayload(value: AppCenterUpdatePayload): AppCenterUpdatePayload {
  return {
    ...value,
    app_name: value.app_name.trim(),
    icon: trimNullable(value.icon),
    owner: trimNullable(value.owner),
    version: trimNullable(value.version),
    description: trimNullable(value.description),
    sort_order: Number(value.sort_order || 0),
  }
}

function openCreateDialog() {
  createForm.value = newCreateForm()
  createDraft.value = newCreateDraft()
  createStep.value = 'basic'
  createDialogVisible.value = true
  nextTick(() => {
    createScrollRef.value?.scrollTo({ top: 0 })
  })
}

async function setCreateStep(step: string) {
  createStep.value = step
  await nextTick()
  const target = document.getElementById(`create-section-${step}`)
  if (!target || !createScrollRef.value) return
  createScrollRef.value.scrollTo({
    top: target.offsetTop - createScrollRef.value.offsetTop,
    behavior: 'smooth',
  })
}

function handleCreateScroll() {
  const container = createScrollRef.value
  if (!container) return
  let current = createSteps[0]?.key || 'basic'
  createSteps.forEach((step) => {
    const target = document.getElementById(`create-section-${step.key}`)
    if (target && target.offsetTop - container.offsetTop - container.scrollTop <= 28) {
      current = step.key
    }
  })
  createStep.value = current
}

async function saveCreateDialog() {
  const payload: AppCenterCreatePayload = {
    ...createForm.value,
    app_code: createForm.value.app_code.trim(),
    app_name: createForm.value.app_name.trim(),
    icon: trimNullable(createForm.value.icon),
    owner: trimNullable(createForm.value.owner),
    version: trimNullable(createDraft.value.version_no) || trimNullable(createForm.value.version),
    description: trimNullable(createForm.value.description),
    sort_order: Number(createForm.value.sort_order || 0),
  }
  if (!payload.app_code || !payload.app_name) {
    ElMessage.error('请填写应用编码和应用名称')
    return
  }
  saving.value = true
  try {
    const created = await createAppCenterApp(payload)
    ElMessage.success('应用已创建')
    createDialogVisible.value = false
    await loadApps()
    await openDetailDialog(created)
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '创建应用失败')
  } finally {
    saving.value = false
  }
}

function openEditDialog(app: AppCenterApp) {
  editAppId.value = app.id
  editForm.value = appToEditForm(app)
  editDialogVisible.value = true
}

async function saveEditDialog() {
  if (!editAppId.value) return
  const payload = normalizeUpdatePayload(editForm.value)
  if (!payload.app_name) {
    ElMessage.error('请填写应用名称')
    return
  }
  saving.value = true
  try {
    const updated = await updateAppCenterApp(editAppId.value, payload)
    ElMessage.success('应用已更新')
    editDialogVisible.value = false
    await loadApps()
    detailApp.value = updated
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '更新应用失败')
  } finally {
    saving.value = false
  }
}

async function toggleAppStatus(app: AppCenterApp) {
  const nextStatus = app.status === 'DISABLED' ? 'ONLINE' : 'DISABLED'
  const actionText = nextStatus === 'DISABLED' ? '停用' : '启用'
  try {
    await ElMessageBox.confirm(`确认${actionText}「${app.app_name}」？`, `${actionText}应用`, {
      confirmButtonText: actionText,
      cancelButtonText: '取消',
      type: nextStatus === 'DISABLED' ? 'warning' : 'info',
    })
    const updated = await updateAppCenterAppStatus(app.id, nextStatus)
    ElMessage.success(`应用已${actionText}`)
    await loadApps()
    if (detailApp.value?.id === updated.id) detailApp.value = updated
  } catch (error) {
    if (error === 'cancel' || error === 'close') return
    ElMessage.error(error instanceof Error ? error.message : `${actionText}应用失败`)
  }
}

async function openDetailDialog(app: AppCenterApp) {
  detailDialogVisible.value = true
  detailLoading.value = true
  detailApp.value = null
  try {
    detailApp.value = await fetchAppCenterApp(app.id)
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '读取应用详情失败')
    detailDialogVisible.value = false
  } finally {
    detailLoading.value = false
  }
}

async function loadDictionaries() {
  const codes = Object.keys(dicts.value) as Array<keyof typeof dicts.value>
  const results = await Promise.allSettled(codes.map((code) => fetchDictItemsByCode(code)))
  results.forEach((result, index) => {
    if (result.status === 'fulfilled') {
      dicts.value[codes[index]] = result.value.items.filter((item) => item.enabled !== false)
    }
  })
}

async function loadApps() {
  loading.value = true
  try {
    const [list, summary] = await Promise.all([
      fetchAppCenterApps({
        skip: 0,
        limit: 100,
        keyword: keyword.value,
        type: typeFilter.value,
        status: statusFilter.value,
        source: sourceFilter.value,
      }),
      fetchAppCenterStats(),
    ])
    apps.value = list.items
    total.value = list.total
    stats.value = summary
  } finally {
    loading.value = false
  }
}

function resetFilters() {
  keyword.value = ''
  typeFilter.value = ''
  statusFilter.value = ''
  sourceFilter.value = ''
  void loadApps()
}

function applyStatusFilter(status: string) {
  statusFilter.value = status
  void loadApps()
}

onMounted(async () => {
  activeSection.value = sectionFromPath(route.path)
  await loadDictionaries()
  await loadApps()
})

watch(
  () => route.path,
  (path) => {
    activeSection.value = sectionFromPath(path)
  },
)
</script>

<template>
  <NeuroAgentPageShell class="app-center-page" :show-hero="activeSection === 'apps'">
    <template #title>应用中心</template>
    <template #subtitle>
      管理平台内置应用与后续应用装载入口；当前仅开放平台应用列表。
    </template>
    <template #meta>
      <span>内置应用必须声明 app_code、工程目录、菜单权限与套餐接入边界。</span>
    </template>
    <template #actions>
      <button
        v-for="card in statCards"
        :key="card.label"
        class="app-stat"
        :class="[`app-stat--${card.tone}`, { 'app-stat--active': statusFilter === card.status }]"
        type="button"
        @click="applyStatusFilter(card.status)"
      >
        <span>{{ card.label }}</span>
        <strong>{{ card.value }}</strong>
        <em>{{ card.hint }}</em>
      </button>
    </template>

    <div class="app-center-layout">
      <section class="app-center-main">
        <template v-if="activeSection === 'apps'">
          <div class="app-center-toolbar">
            <div class="app-center-toolbar__filters">
              <div class="app-center-search">
                <el-icon :size="16"><Search /></el-icon>
                <input v-model="keyword" type="text" placeholder="搜索应用编码、名称或说明" @keyup.enter="loadApps" />
              </div>
              <select v-model="typeFilter" aria-label="应用类型">
                <option value="">全部类型</option>
                <option v-for="item in dictOptions('app_type')" :key="item.value" :value="item.value">
                  {{ item.label }}
                </option>
              </select>
              <select v-model="statusFilter" aria-label="应用状态">
                <option value="">全部状态</option>
                <option v-for="item in statusFilterOptions" :key="item.value" :value="item.value">
                  {{ item.label }}
                </option>
              </select>
              <select v-model="sourceFilter" aria-label="应用来源">
                <option value="">全部来源</option>
                <option v-for="item in dictOptions('app_source')" :key="item.value" :value="item.value">
                  {{ item.label }}
                </option>
              </select>
              <button class="app-btn app-btn--primary" type="button" @click="loadApps">
                查询
              </button>
              <button v-if="hasActiveFilters" class="app-btn" type="button" @click="resetFilters">
                <el-icon :size="15"><Refresh /></el-icon>
                重置
              </button>
            </div>
            <div class="app-center-toolbar__actions">
              <button class="app-btn app-btn--primary" type="button" @click="openCreateDialog">
                <el-icon :size="15"><Plus /></el-icon>
                新增应用
              </button>
            </div>
          </div>

          <div v-if="loading" class="app-center-empty">加载中...</div>
          <div v-else-if="apps.length === 0" class="app-center-empty">暂无应用</div>
          <div v-else class="app-type-list">
            <section v-for="group in groupedApps" :key="group.type" class="app-type-group">
              <div class="app-type-group__head">
                <div>
                  <strong>{{ group.label }}</strong>
                  <span>{{ group.type }}</span>
                </div>
                <em>{{ group.items.length }}</em>
              </div>
              <div class="app-grid">
                <article
                  v-for="app in group.items"
                  :key="app.app_code"
                  class="app-card"
                  :class="{ 'app-card--disabled': app.status === 'DISABLED' || app.status === 'ARCHIVED' }"
                >
                  <div class="app-card__head">
                    <div class="app-card__icon">
                      <span>{{ appInitial(app) }}</span>
                    </div>
                    <div class="app-card__title">
                      <h3>{{ app.app_name }}</h3>
                      <code>{{ app.app_code }}</code>
                    </div>
                    <span class="app-status" :class="statusClass(app.status)">
                      {{ labelOf('app_status', app.status) }}
                    </span>
                  </div>
                  <div class="app-card__badges">
                    <span>{{ labelOf('app_source', app.source) }}</span>
                    <span>{{ labelOf('app_charge_mode', app.charge_mode) }}</span>
                  </div>
                  <p class="app-card__desc">{{ app.description || '暂无说明' }}</p>
                  <div class="app-card__metrics">
                    <div>
                      <span>负责人</span>
                      <strong>{{ app.owner || '—' }}</strong>
                    </div>
                    <div>
                      <span>当前版本</span>
                      <strong>{{ app.version || '—' }}</strong>
                    </div>
                    <div>
                      <span>来源</span>
                      <strong>{{ labelOf('app_source', app.source) }}</strong>
                    </div>
                    <div>
                      <span>可见范围</span>
                      <strong>{{ labelOf('app_visibility_scope', app.visibility_scope) }}</strong>
                    </div>
                  </div>
                  <div class="app-card__channels">
                    <span>PC Web</span>
                  </div>
                  <div class="app-card__actions">
                    <button class="app-card__detail app-card__detail--primary" type="button" @click="openDetailDialog(app)">
                      进入详情
                    </button>
                    <button class="app-card__detail" type="button" @click="openEditDialog(app)">编辑</button>
                    <button class="app-card__detail" type="button" :disabled="app.is_builtin" @click="toggleAppStatus(app)">
                      {{ app.status === 'DISABLED' ? '启用' : '停用' }}
                    </button>
                    <button class="app-card__detail" type="button" disabled>安装记录</button>
                    <button class="app-card__detail" type="button" disabled>邀请体验</button>
                  </div>
                </article>
              </div>
            </section>
          </div>

        </template>

        <div v-else class="app-center-production-card">
          <strong>{{ activeDirectoryItem.count }}</strong>
          <span>{{ activeDirectoryItem.label }}</span>
          <p>{{ activeDirectoryItem.description }}本页已接入生产统计数据，详细管理视图将在对应模块落地后展开。</p>
        </div>
      </section>
    </div>

    <NeuroAgentDialog
      v-model="createDialogVisible"
      title="新增应用"
      icon="📦"
      size="large"
      width="1040px"
      height="86vh"
      :loading="saving"
      :confirm-disabled="saving"
      confirm-text="创建应用主档"
      @confirm="saveCreateDialog"
    >
      <div class="app-create-wizard">
        <aside class="app-create-steps">
          <button
            v-for="(step, index) in createSteps"
            :key="step.key"
            class="app-create-step"
            :class="{ 'is-active': createStep === step.key, 'is-complete': createStepCompletion[step.key] }"
            type="button"
            @click="setCreateStep(step.key)"
          >
            <i>{{ createStepCompletion[step.key] ? '✓' : index + 1 }}</i>
            <span>
              <strong>{{ step.title }}</strong>
              <em>{{ step.hint }}</em>
            </span>
          </button>
        </aside>

        <section ref="createScrollRef" class="app-create-body" @scroll="handleCreateScroll">
          <div id="create-section-basic" class="app-create-panel">
            <div class="app-create-panel__head">
              <div>
                <h3>基础信息</h3>
                <p>创建应用主档，明确应用长期身份。应用编码创建后不建议修改。</p>
              </div>
              <span>必填</span>
            </div>
            <div class="app-form">
              <label>
                <span>应用编码</span>
                <input v-model="createForm.app_code" placeholder="例如 ai-project-manager，只允许小写字母、数字、中横线" />
              </label>
              <label>
                <span>应用名称</span>
                <input v-model="createForm.app_name" placeholder="请输入应用名称" />
              </label>
              <label>
                <span>应用类型</span>
                <select v-model="createForm.app_type">
                  <option v-for="item in dictOptions('app_type')" :key="item.value" :value="item.value">
                    {{ item.label }}
                  </option>
                </select>
              </label>
              <label>
                <span>应用定位</span>
                <select v-model="createDraft.app_position">
                  <option value="PLATFORM_BUILTIN">平台内置</option>
                  <option value="COMMON_CAPABILITY">公共能力 / 业务中台</option>
                  <option value="BUSINESS_SCENARIO">业务场景</option>
                  <option value="AI_AGENT">AI / Agent</option>
                  <option value="INTEGRATION_CONNECTOR">集成连接</option>
                </select>
              </label>
              <label>
                <span>状态</span>
                <select v-model="createForm.status">
                  <option v-for="item in dictOptions('app_status')" :key="item.value" :value="item.value">
                    {{ item.label }}
                  </option>
                </select>
              </label>
              <label>
                <span>负责人</span>
                <input v-model="createForm.owner" placeholder="负责人，选填" />
              </label>
              <label>
                <span>图标</span>
                <input v-model="createForm.icon" placeholder="图标名或标识，选填" />
              </label>
              <label>
                <span>排序</span>
                <input v-model.number="createForm.sort_order" type="number" min="0" step="1" />
              </label>
              <label class="app-form__full">
                <span>应用介绍</span>
                <textarea v-model="createForm.description" placeholder="说明应用定位、边界、主要用户和核心能力"></textarea>
              </label>
            </div>
          </div>

          <div id="create-section-access" class="app-create-panel">
            <div class="app-create-panel__head">
              <div>
                <h3>可见范围、开通与收费</h3>
                <p>可见范围最终只分为全部租户和指定租户。当前主档保存仍兼容现有后端可见范围字段。</p>
              </div>
            </div>
            <div class="app-form">
              <label>
                <span>当前后端可见范围</span>
                <select v-model="createForm.visibility_scope">
                  <option v-for="item in dictOptions('app_visibility_scope')" :key="item.value" :value="item.value">
                    {{ item.label }}
                  </option>
                </select>
              </label>
              <label>
                <span>V1 可见范围草稿</span>
                <select v-model="createDraft.visibility_mode">
                  <option value="ALL_TENANTS">全部租户</option>
                  <option value="SPECIFIED_TENANTS">指定租户</option>
                </select>
              </label>
              <label class="app-form__full">
                <span>指定租户草稿</span>
                <input v-model="createDraft.visible_tenants" placeholder="选择或记录指定租户，后续落库到可见租户关系表" />
              </label>
              <label>
                <span>开通方式草稿</span>
                <select v-model="createDraft.open_method">
                  <option value="PACKAGE,DIRECT_SUBSCRIBE">套餐开通 / 套餐外订阅</option>
                  <option value="INVITE_CODE,ADMIN_GRANT">邀请码 / 平台授权</option>
                  <option value="TRIAL">免费试用</option>
                  <option value="BUYOUT">一次性买断</option>
                </select>
              </label>
              <label>
                <span>收费方式</span>
                <select v-model="createForm.charge_mode">
                  <option v-for="item in dictOptions('app_charge_mode')" :key="item.value" :value="item.value">
                    {{ item.label }}
                  </option>
                </select>
              </label>
              <label>
                <span>免费使用时间草稿</span>
                <select v-model="createDraft.trial_days">
                  <option>不支持试用</option>
                  <option>7 天</option>
                  <option>15 天</option>
                  <option>30 天</option>
                  <option>不限期</option>
                </select>
              </label>
              <label>
                <span>试用开始规则草稿</span>
                <select v-model="createDraft.trial_start_rule">
                  <option>首次安装时开始</option>
                  <option>首次启用时开始</option>
                  <option>邀请码接受时开始</option>
                  <option>平台管理员手动开始</option>
                </select>
              </label>
            </div>
          </div>

          <div id="create-section-clients" class="app-create-panel">
            <div class="app-create-panel__head">
              <div>
                <h3>客户端草稿</h3>
                <p>先明确应用支持的客户端形态。后续实现会保存到应用客户端表，并和租户权益联动。</p>
              </div>
            </div>
            <div class="app-create-checks">
              <label v-for="client in createDraft.clients" :key="client.key">
                <input v-model="client.checked" type="checkbox" />
                <span>{{ client.label }}</span>
              </label>
            </div>
          </div>

          <div id="create-section-assets" class="app-create-panel">
            <div class="app-create-panel__head">
              <div>
                <h3>入口、API、权限草稿</h3>
                <p>手工创建阶段先生成资产草稿。真实代码应用后续通过 Manifest 装载补齐。</p>
              </div>
            </div>
            <div class="app-create-checks">
              <label v-for="asset in createDraft.assets" :key="asset.key">
                <input v-model="asset.checked" type="checkbox" />
                <span>{{ asset.label }}</span>
              </label>
            </div>
            <div class="app-create-note">
              <strong>建议默认生成：</strong>
              <span>应用访问权限、PC Web 工作台入口、查询 API 草稿和套餐资源草稿。</span>
            </div>
          </div>

          <div id="create-section-docs" class="app-create-panel">
            <div class="app-create-panel__head">
              <div>
                <h3>文档池与初始版本</h3>
                <p>创建应用时就建立知识资产骨架，避免只有菜单没有需求和设计依据。</p>
              </div>
            </div>
            <div class="app-form">
              <label>
                <span>初始版本</span>
                <input v-model="createDraft.version_no" placeholder="例如 0.1.0" />
              </label>
              <label>
                <span>版本渠道</span>
                <select v-model="createDraft.release_channel">
                  <option value="DEV">开发版 DEV</option>
                  <option value="INTERNAL_TEST">内测版 INTERNAL_TEST</option>
                  <option value="BETA">Beta BETA</option>
                  <option value="STABLE">正式版 STABLE</option>
                </select>
              </label>
              <label class="app-form__full">
                <span>初始发布说明</span>
                <textarea v-model="createDraft.release_note" placeholder="记录创建原因、初始范围和后续补齐计划"></textarea>
              </label>
            </div>
            <div class="app-create-checks">
              <label v-for="doc in createDraft.docs" :key="doc.key">
                <input v-model="doc.checked" type="checkbox" />
                <span>{{ doc.label }}</span>
              </label>
            </div>
          </div>

          <div id="create-section-review" class="app-create-panel app-create-panel--review">
            <div class="app-create-panel__head">
              <div>
                <h3>确认创建</h3>
                <p>确认后当前版本先保存应用主档，其他应用资产作为后续实现和详情页完善项。</p>
              </div>
              <span>草稿</span>
            </div>
            <div class="app-create-review">
              <div>
                <span>应用</span>
                <strong>{{ createForm.app_name || '未填写' }}</strong>
              </div>
              <div>
                <span>app_code</span>
                <strong>{{ createForm.app_code || '未填写' }}</strong>
              </div>
              <div>
                <span>可见范围</span>
                <strong>{{ createDraft.visibility_mode === 'ALL_TENANTS' ? '全部租户' : '指定租户' }}</strong>
              </div>
              <div>
                <span>收费方式</span>
                <strong>{{ labelOf('app_charge_mode', createForm.charge_mode) }}</strong>
              </div>
              <div>
                <span>客户端草稿</span>
                <strong>{{ selectedCreateClients.join(' / ') || '未选择' }}</strong>
              </div>
              <div>
                <span>资产草稿</span>
                <strong>{{ selectedCreateAssets.length }} 项</strong>
              </div>
              <div>
                <span>文档草稿</span>
                <strong>{{ selectedCreateDocs.length }} 项</strong>
              </div>
              <div>
                <span>初始版本</span>
                <strong>{{ createDraft.version_no }} / {{ createDraft.release_channel }}</strong>
              </div>
            </div>
          </div>
        </section>
      </div>
    </NeuroAgentDialog>

    <NeuroAgentDialog
      v-model="editDialogVisible"
      title="编辑应用"
      icon="📦"
      size="large"
      :loading="saving"
      :confirm-disabled="saving"
      confirm-text="保存"
      @confirm="saveEditDialog"
    >
      <div class="app-form">
        <label>
          <span>应用名称</span>
          <input v-model="editForm.app_name" placeholder="请输入应用名称" />
        </label>
        <label>
          <span>应用类型</span>
          <select v-model="editForm.app_type">
            <option v-for="item in dictOptions('app_type')" :key="item.value" :value="item.value">
              {{ item.label }}
            </option>
          </select>
        </label>
        <label>
          <span>计费模式</span>
          <select v-model="editForm.charge_mode">
            <option v-for="item in dictOptions('app_charge_mode')" :key="item.value" :value="item.value">
              {{ item.label }}
            </option>
          </select>
        </label>
        <label>
          <span>可见范围</span>
          <select v-model="editForm.visibility_scope">
            <option v-for="item in dictOptions('app_visibility_scope')" :key="item.value" :value="item.value">
              {{ item.label }}
            </option>
          </select>
        </label>
        <label>
          <span>图标</span>
          <input v-model="editForm.icon" placeholder="图标名或标识，选填" />
        </label>
        <label>
          <span>负责人</span>
          <input v-model="editForm.owner" placeholder="负责人，选填" />
        </label>
        <label>
          <span>版本</span>
          <input v-model="editForm.version" placeholder="例如 0.1.0" />
        </label>
        <label>
          <span>排序</span>
          <input v-model.number="editForm.sort_order" type="number" min="0" step="1" />
        </label>
        <label class="app-form__full">
          <span>说明</span>
          <textarea v-model="editForm.description" placeholder="请输入应用说明，选填"></textarea>
        </label>
      </div>
    </NeuroAgentDialog>

    <NeuroAgentDialog
      v-model="detailDialogVisible"
      title="应用详情"
      icon="📦"
      size="large"
      :loading="detailLoading"
      :show-confirm="false"
      cancel-text="关闭"
    >
      <div v-if="detailApp" class="app-detail">
        <div class="app-detail__head">
          <div class="app-card__icon">
            <el-icon :size="24"><component :is="iconFor(detailApp)" /></el-icon>
          </div>
          <div>
            <h3>{{ detailApp.app_name }}</h3>
            <code>{{ detailApp.app_code }}</code>
          </div>
          <span class="app-status" :class="statusClass(detailApp.status)">
            {{ labelOf('app_status', detailApp.status) }}
          </span>
        </div>
        <div class="app-detail__grid">
          <span>应用类型</span><strong>{{ labelOf('app_type', detailApp.app_type) }}</strong>
          <span>来源</span><strong>{{ labelOf('app_source', detailApp.source) }}</strong>
          <span>计费模式</span><strong>{{ labelOf('app_charge_mode', detailApp.charge_mode) }}</strong>
          <span>可见范围</span><strong>{{ labelOf('app_visibility_scope', detailApp.visibility_scope) }}</strong>
          <span>负责人</span><strong>{{ detailApp.owner || '—' }}</strong>
          <span>版本</span><strong>{{ detailApp.version || '—' }}</strong>
          <span>排序</span><strong>{{ detailApp.sort_order }}</strong>
          <span>内置应用</span><strong>{{ detailApp.is_builtin ? '是' : '否' }}</strong>
        </div>
        <p class="app-detail__desc">{{ detailApp.description || '暂无说明' }}</p>
      </div>
    </NeuroAgentDialog>
  </NeuroAgentPageShell>
</template>

<style scoped>
@import '@/styles/theme/neuro-theme.css';

.app-center-page :deep(.neuro-page-shell__panel-body) {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.app-center-layout {
  width: 100%;
}

.app-center-main {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.app-center-production-card p {
  margin: 0;
  color: var(--neuro-text-secondary);
}

.app-stat {
  width: 144px;
  min-height: 94px;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  justify-content: center;
  gap: 8px;
  padding: 14px 16px;
  border: 1px solid color-mix(in srgb, var(--neuro-primary) 30%, transparent);
  border-radius: var(--neuro-radius-md);
  background: color-mix(in srgb, var(--neuro-surface-2) 78%, transparent);
  color: var(--neuro-text);
  text-align: left;
  cursor: pointer;
  transition: border-color 0.18s ease, background 0.18s ease, transform 0.18s ease;
}

.app-stat strong {
  display: block;
  font-size: 30px;
  line-height: 1;
  color: var(--neuro-text);
}

.app-stat span {
  display: block;
  font-size: 13px;
  font-weight: 800;
  color: var(--neuro-text-secondary);
}

.app-stat em {
  color: var(--neuro-text-secondary);
  font-size: 12px;
  font-style: normal;
}

.app-stat:hover,
.app-stat--active {
  border-color: color-mix(in srgb, var(--neuro-primary) 72%, transparent);
  background: color-mix(in srgb, var(--neuro-primary) 12%, var(--neuro-surface-2));
  transform: translateY(-1px);
}

.app-stat--success strong {
  color: #65df98;
}

.app-stat--warning strong {
  color: #f2994a;
}

.app-stat--info strong {
  color: #72a6ff;
}

.app-stat--muted strong {
  color: #97a1b3;
}

.app-center-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 14px;
  border-radius: var(--neuro-radius-lg);
  background: color-mix(in srgb, var(--neuro-surface-2) 72%, transparent);
}

.app-center-toolbar__filters,
.app-center-toolbar__actions {
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 10px;
}

.app-center-toolbar__filters {
  flex: 0 1 auto;
  flex-wrap: nowrap;
}

.app-center-toolbar__actions {
  flex: 0 0 auto;
  margin-left: auto;
}

.app-center-search,
.app-center-toolbar__filters select {
  height: 36px;
  min-width: 0;
  border: 1px solid color-mix(in srgb, var(--neuro-border) 84%, transparent);
  border-radius: var(--neuro-radius-sm);
  background: color-mix(in srgb, var(--neuro-surface) 92%, transparent);
  color: var(--neuro-text);
  font-size: 14px;
}

.app-center-search {
  flex: 0 0 210px;
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 12px;
}

.app-center-search input {
  min-width: 0;
  width: 100%;
  border: 0;
  outline: 0;
  background: transparent;
  color: var(--neuro-text);
  font-size: 14px;
}

.app-center-search input::placeholder {
  color: var(--neuro-text-muted);
  font-size: 14px;
}

.app-center-toolbar__filters select {
  flex: 0 0 142px;
  padding: 0 10px;
}

.app-btn {
  height: 36px;
  flex: 0 0 auto;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 0 14px;
  border: 1px solid color-mix(in srgb, var(--neuro-border) 90%, transparent);
  border-radius: var(--neuro-radius-sm);
  background: color-mix(in srgb, var(--neuro-surface) 84%, transparent);
  color: var(--neuro-text);
  cursor: pointer;
  white-space: nowrap;
}

.app-btn--primary {
  border-color: transparent;
  background: linear-gradient(135deg, var(--neuro-primary), var(--neuro-accent));
  color: #051616;
  font-weight: 700;
}

.app-type-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.app-type-group {
  overflow: hidden;
  border: 1px solid color-mix(in srgb, var(--neuro-border) 84%, transparent);
  border-radius: var(--neuro-radius-lg);
  background: color-mix(in srgb, var(--neuro-surface) 78%, transparent);
}

.app-type-group__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 14px 16px;
  border-bottom: 1px solid color-mix(in srgb, var(--neuro-border) 72%, transparent);
  background:
    radial-gradient(circle at 18% 0%, color-mix(in srgb, var(--neuro-primary) 18%, transparent), transparent 38%),
    color-mix(in srgb, var(--neuro-surface-2) 82%, transparent);
}

.app-type-group__head > div {
  min-width: 0;
  display: flex;
  align-items: baseline;
  gap: 10px;
}

.app-type-group__head strong {
  color: var(--neuro-text);
  font-size: 18px;
  font-weight: 900;
}

.app-type-group__head span {
  color: var(--neuro-text-secondary);
  font-size: 12px;
  text-transform: uppercase;
}

.app-type-group__head em {
  min-width: 34px;
  height: 26px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 1px solid color-mix(in srgb, var(--neuro-primary) 34%, transparent);
  border-radius: 999px;
  background: color-mix(in srgb, var(--neuro-primary) 12%, transparent);
  color: var(--neuro-primary);
  font-style: normal;
  font-weight: 900;
}

.app-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  align-items: stretch;
  gap: 18px;
  padding: 18px;
}

.app-card {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 22px;
  border: 1px solid color-mix(in srgb, var(--neuro-border) 88%, transparent);
  border-radius: var(--neuro-radius-lg);
  background:
    linear-gradient(180deg, color-mix(in srgb, var(--neuro-surface-2) 72%, transparent), transparent 68%),
    color-mix(in srgb, var(--neuro-surface) 90%, transparent);
  box-shadow: inset 0 1px 0 color-mix(in srgb, var(--neuro-primary) 38%, transparent);
}

.app-card--disabled {
  opacity: 0.58;
  filter: grayscale(0.35);
}

.app-card__head,
.app-card__foot {
  display: flex;
  align-items: center;
  gap: 10px;
}

.app-card__icon {
  width: 54px;
  height: 54px;
  flex: 0 0 54px;
  display: grid;
  place-items: center;
  border: 1px solid color-mix(in srgb, var(--neuro-primary) 30%, transparent);
  border-radius: 16px;
  background: color-mix(in srgb, var(--neuro-primary) 16%, transparent);
  color: var(--neuro-primary);
}

.app-card__icon span {
  font-size: 20px;
  font-weight: 900;
}

.app-card__title {
  min-width: 0;
  flex: 1 1 auto;
}

.app-card__title h3 {
  margin: 0 0 4px;
  font-size: 20px;
  color: var(--neuro-text);
}

.app-card__title code {
  color: var(--neuro-text-secondary);
  font-size: 14px;
}

.app-status,
.app-card__badges span,
.app-card__channels span {
  display: inline-flex;
  align-items: center;
  min-height: 28px;
  padding: 0 12px;
  border-radius: 999px;
  font-size: 13px;
  font-weight: 800;
  white-space: nowrap;
}

.app-status {
  background: color-mix(in srgb, var(--neuro-surface-2) 92%, transparent);
  color: var(--neuro-text-secondary);
}

.app-status--online {
  background: color-mix(in srgb, #0f8f4d 22%, transparent);
  color: #22d36f;
}

.app-status--developing {
  background: color-mix(in srgb, #2f6fed 22%, transparent);
  color: #72a6ff;
}

.app-status--disabled {
  color: #8b95a7;
}

.app-status--draft {
  background: color-mix(in srgb, #b7791f 24%, transparent);
  color: #f2b95b;
}

.app-card__desc {
  margin: 0;
  color: var(--neuro-text-secondary);
  font-size: 14px;
  line-height: 1.6;
}

.app-card__badges,
.app-card__channels {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.app-card__badges span:nth-child(1) {
  border: 1px solid color-mix(in srgb, #5a9cff 40%, transparent);
  color: #78a9ff;
  background: color-mix(in srgb, #2f6fed 14%, transparent);
}

.app-card__badges span:nth-child(2) {
  border: 1px solid color-mix(in srgb, #f2994a 42%, transparent);
  color: #f2994a;
  background: color-mix(in srgb, #f2994a 12%, transparent);
}

.app-card__metrics {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 10px;
}

.app-card__metrics div {
  min-width: 0;
  padding: 10px 12px;
  border: 1px solid color-mix(in srgb, var(--neuro-border) 74%, transparent);
  border-radius: var(--neuro-radius-sm);
  background: color-mix(in srgb, var(--neuro-surface-2) 62%, transparent);
}

.app-card__metrics span,
.app-card__metrics strong {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.app-card__metrics span {
  color: var(--neuro-text-secondary);
  font-size: 12px;
  font-weight: 800;
}

.app-card__metrics strong {
  margin-top: 4px;
  color: var(--neuro-text);
  font-size: 14px;
}

.app-card__channels span {
  border: 1px solid color-mix(in srgb, #5a9cff 42%, transparent);
  color: #7fb0ff;
  background: color-mix(in srgb, #2f6fed 12%, transparent);
}

.app-card__foot {
  justify-content: space-between;
  padding-top: 10px;
  border-top: 1px solid color-mix(in srgb, var(--neuro-border) 70%, transparent);
  color: var(--neuro-text-secondary);
  font-size: 12px;
}

.app-card__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  padding-top: 14px;
  border-top: 1px solid color-mix(in srgb, var(--neuro-border) 70%, transparent);
}

.app-card__detail {
  height: 34px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  border: 1px solid color-mix(in srgb, var(--neuro-primary) 42%, transparent);
  border-radius: var(--neuro-radius-sm);
  background: color-mix(in srgb, var(--neuro-primary) 10%, transparent);
  color: var(--neuro-primary);
  cursor: pointer;
  font-weight: 700;
}

.app-card__detail:disabled {
  cursor: not-allowed;
  opacity: 0.52;
}

.app-card__detail--primary {
  min-width: 92px;
  border-color: transparent;
  background: var(--neuro-primary);
  color: #051616;
}

.app-center-empty {
  padding: 48px 16px;
  text-align: center;
  color: var(--neuro-text-secondary);
}

.app-center-production-card {
  min-height: 320px;
  display: grid;
  place-items: center;
  align-content: center;
  gap: 8px;
  padding: 42px;
  border: 1px solid color-mix(in srgb, var(--neuro-border) 84%, transparent);
  border-radius: var(--neuro-radius-lg);
  background: color-mix(in srgb, var(--neuro-surface) 88%, transparent);
  text-align: center;
}

.app-center-production-card strong {
  font-size: 54px;
  line-height: 1;
  color: var(--neuro-primary);
}

.app-center-production-card span {
  font-size: 20px;
  font-weight: 900;
  color: var(--neuro-text);
}

.app-create-wizard {
  height: 100%;
  min-height: 0;
  display: grid;
  grid-template-columns: 180px minmax(0, 1fr);
  gap: 20px;
  align-items: flex-start;
}

.app-create-steps {
  position: sticky;
  top: 0;
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 4px 0;
  opacity: 0.82;
}

.app-create-step {
  width: 100%;
  min-width: 0;
  display: grid;
  grid-template-columns: 22px minmax(0, 1fr);
  gap: 8px;
  align-items: flex-start;
  padding: 9px 8px;
  border: 1px solid transparent;
  border-radius: var(--neuro-radius-sm);
  background: transparent;
  color: var(--neuro-text-secondary);
  text-align: left;
  cursor: pointer;
}

.app-create-step i {
  width: 22px;
  height: 22px;
  display: inline-grid;
  place-items: center;
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--neuro-border) 76%, transparent);
  background: color-mix(in srgb, var(--neuro-surface) 86%, transparent);
  color: var(--neuro-text-secondary);
  font-style: normal;
  font-size: 12px;
  font-weight: 900;
}

.app-create-step strong,
.app-create-step em {
  display: block;
  min-width: 0;
}

.app-create-step strong {
  margin-bottom: 2px;
  color: var(--neuro-text);
  font-size: 13px;
  line-height: 1.35;
}

.app-create-step em {
  color: var(--neuro-text-secondary);
  font-style: normal;
  font-size: 12px;
  line-height: 1.45;
}

.app-create-step:hover {
  background: color-mix(in srgb, var(--neuro-surface) 74%, transparent);
}

.app-create-step.is-active {
  border-color: color-mix(in srgb, var(--neuro-primary) 28%, transparent);
  background: color-mix(in srgb, var(--neuro-primary) 7%, transparent);
}

.app-create-step.is-complete i {
  border-color: color-mix(in srgb, var(--neuro-primary) 46%, transparent);
  background: color-mix(in srgb, var(--neuro-primary) 18%, transparent);
  color: var(--neuro-primary);
}

.app-create-body {
  position: relative;
  width: 100%;
  min-width: 0;
  height: calc(86vh - 150px);
  overflow-y: auto;
  padding: 2px 8px 10px 0;
  scroll-behavior: smooth;
  scroll-padding-top: 4px;
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.app-create-body::-webkit-scrollbar {
  width: 8px;
}

.app-create-body::-webkit-scrollbar-track {
  background: transparent;
}

.app-create-body::-webkit-scrollbar-thumb {
  border-radius: 999px;
  background: color-mix(in srgb, var(--neuro-border) 80%, transparent);
}

.app-create-panel {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 20px;
  border: 1px solid color-mix(in srgb, var(--neuro-border) 80%, transparent);
  border-radius: var(--neuro-radius-md);
  background:
    linear-gradient(180deg, color-mix(in srgb, var(--neuro-surface) 96%, transparent), color-mix(in srgb, var(--neuro-surface) 88%, transparent));
  box-shadow: 0 18px 44px color-mix(in srgb, #000 8%, transparent);
}

.app-create-panel--review {
  margin-bottom: 4px;
}

.app-create-panel__head {
  display: flex;
  justify-content: space-between;
  gap: 14px;
  align-items: flex-start;
}

.app-create-panel__head h3 {
  margin: 0 0 5px;
  color: var(--neuro-text);
  font-size: 18px;
  line-height: 1.3;
}

.app-create-panel__head p {
  margin: 0;
  color: var(--neuro-text-secondary);
  line-height: 1.65;
}

.app-create-panel__head > span {
  flex: 0 0 auto;
  padding: 5px 10px;
  border-radius: 999px;
  background: color-mix(in srgb, var(--neuro-primary) 14%, transparent);
  color: var(--neuro-primary);
  font-size: 12px;
  font-weight: 900;
}

.app-create-checks {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
}

.app-create-checks label {
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 10px;
  min-height: 44px;
  padding: 10px 12px;
  border: 1px solid color-mix(in srgb, var(--neuro-border) 80%, transparent);
  border-radius: var(--neuro-radius-sm);
  background: color-mix(in srgb, var(--neuro-surface) 88%, transparent);
  color: var(--neuro-text);
  font-weight: 800;
}

.app-create-checks input {
  width: 16px;
  height: 16px;
  accent-color: var(--neuro-primary);
}

.app-create-checks span {
  min-width: 0;
  line-height: 1.4;
}

.app-create-note {
  display: flex;
  gap: 8px;
  align-items: flex-start;
  padding: 12px 14px;
  border: 1px solid color-mix(in srgb, var(--neuro-primary) 24%, transparent);
  border-radius: var(--neuro-radius-sm);
  background: color-mix(in srgb, var(--neuro-primary) 8%, transparent);
  color: var(--neuro-text-secondary);
  line-height: 1.6;
}

.app-create-note strong {
  flex: 0 0 auto;
  color: var(--neuro-text);
}

.app-create-review {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.app-create-review > div {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 12px;
  border: 1px solid color-mix(in srgb, var(--neuro-border) 80%, transparent);
  border-radius: var(--neuro-radius-sm);
  background: color-mix(in srgb, var(--neuro-surface) 88%, transparent);
}

.app-create-review span {
  color: var(--neuro-text-secondary);
  font-size: 12px;
  font-weight: 800;
}

.app-create-review strong {
  min-width: 0;
  color: var(--neuro-text);
  line-height: 1.45;
  overflow-wrap: anywhere;
}

.app-form {
  width: 100%;
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
}

.app-form *,
.app-form *::before,
.app-form *::after {
  box-sizing: border-box;
}

.app-form label {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.app-form label > span {
  color: var(--neuro-text);
  font-weight: 800;
}

.app-form input,
.app-form select,
.app-form textarea {
  width: 100%;
  min-width: 0;
  max-width: 100%;
  border: 1px solid color-mix(in srgb, var(--neuro-border) 84%, transparent);
  border-radius: var(--neuro-radius-sm);
  background: color-mix(in srgb, var(--neuro-surface) 92%, transparent);
  color: var(--neuro-text);
  font: inherit;
}

.app-form input,
.app-form select {
  height: 40px;
  padding: 0 12px;
}

.app-form textarea {
  min-height: 92px;
  padding: 12px;
  resize: vertical;
}

.app-form__full {
  grid-column: 1 / -1;
}

.app-detail {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.app-detail__head {
  display: flex;
  align-items: center;
  gap: 12px;
}

.app-detail__head h3 {
  margin: 0 0 4px;
  color: var(--neuro-text);
}

.app-detail__head code,
.app-detail__desc {
  color: var(--neuro-text-secondary);
}

.app-detail__grid {
  display: grid;
  grid-template-columns: 120px minmax(0, 1fr);
  gap: 10px 14px;
  padding: 14px;
  border: 1px solid color-mix(in srgb, var(--neuro-border) 80%, transparent);
  border-radius: var(--neuro-radius-md);
  background: color-mix(in srgb, var(--neuro-surface) 82%, transparent);
}

.app-detail__grid span {
  color: var(--neuro-text-secondary);
}

.app-detail__grid strong {
  min-width: 0;
  color: var(--neuro-text);
}

.app-detail__desc {
  margin: 0;
  line-height: 1.7;
}

@media (max-width: 980px) {
  .app-center-toolbar {
    align-items: stretch;
    flex-wrap: wrap;
  }

  .app-center-toolbar__filters {
    flex-basis: 100%;
    flex-wrap: wrap;
  }

  .app-center-search {
    flex-basis: 100%;
  }

  .app-center-toolbar__filters select {
    flex: 1 1 160px;
  }

  .app-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .app-card__metrics {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .app-create-wizard {
    grid-template-columns: 1fr;
  }

  .app-create-steps {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .app-create-checks {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 640px) {
  .app-center-toolbar__filters > *,
  .app-center-toolbar__actions {
    flex-basis: 100%;
  }

  .app-center-toolbar__actions {
    justify-content: flex-end;
  }

  .app-type-group__head > div {
    flex-direction: column;
    gap: 4px;
    align-items: flex-start;
  }

  .app-card {
    padding: 16px;
  }

  .app-grid {
    grid-template-columns: 1fr;
  }

  .app-card__head {
    align-items: flex-start;
  }

  .app-card__metrics {
    grid-template-columns: 1fr;
  }

  .app-form {
    grid-template-columns: 1fr;
  }

  .app-create-steps,
  .app-create-checks,
  .app-create-review {
    grid-template-columns: 1fr;
  }

  .app-create-panel__head,
  .app-create-note {
    flex-direction: column;
  }
}
</style>
