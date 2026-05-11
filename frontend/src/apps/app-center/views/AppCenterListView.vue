<script setup lang="ts">
defineOptions({ name: 'AppCenterListView' })

import { Box, Grid, Monitor, Plus, Refresh, Search, Setting } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { computed, onMounted, ref, watch } from 'vue'
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
  createDialogVisible.value = true
}

async function saveCreateDialog() {
  const payload: AppCenterCreatePayload = {
    ...createForm.value,
    app_code: createForm.value.app_code.trim(),
    app_name: createForm.value.app_name.trim(),
    icon: trimNullable(createForm.value.icon),
    owner: trimNullable(createForm.value.owner),
    version: trimNullable(createForm.value.version),
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
      :loading="saving"
      :confirm-disabled="saving"
      confirm-text="保存"
      @confirm="saveCreateDialog"
    >
      <div class="app-form">
        <label>
          <span>应用编码</span>
          <input v-model="createForm.app_code" placeholder="例如 crm-suite，只允许小写字母、数字、中横线" />
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
          <span>状态</span>
          <select v-model="createForm.status">
            <option v-for="item in dictOptions('app_status')" :key="item.value" :value="item.value">
              {{ item.label }}
            </option>
          </select>
        </label>
        <label>
          <span>计费模式</span>
          <select v-model="createForm.charge_mode">
            <option v-for="item in dictOptions('app_charge_mode')" :key="item.value" :value="item.value">
              {{ item.label }}
            </option>
          </select>
        </label>
        <label>
          <span>可见范围</span>
          <select v-model="createForm.visibility_scope">
            <option v-for="item in dictOptions('app_visibility_scope')" :key="item.value" :value="item.value">
              {{ item.label }}
            </option>
          </select>
        </label>
        <label>
          <span>图标</span>
          <input v-model="createForm.icon" placeholder="图标名或标识，选填" />
        </label>
        <label>
          <span>负责人</span>
          <input v-model="createForm.owner" placeholder="负责人，选填" />
        </label>
        <label>
          <span>版本</span>
          <input v-model="createForm.version" placeholder="例如 0.1.0" />
        </label>
        <label>
          <span>排序</span>
          <input v-model.number="createForm.sort_order" type="number" min="0" step="1" />
        </label>
        <label class="app-form__full">
          <span>说明</span>
          <textarea v-model="createForm.description" placeholder="请输入应用说明，选填"></textarea>
        </label>
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
}
</style>
