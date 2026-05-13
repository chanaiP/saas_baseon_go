<script setup lang="ts">
defineOptions({ name: 'AuditLogsView' })
import { computed, onMounted, ref } from 'vue'

import { fetchAppCenterApps } from '@/apps/app-center/api'
import { fetchAuditLogs } from '@/api/logs'
import type { AuditLogRow } from '@/api/logs'
import { formatDateTimeChina } from '@/utils/datetime'
import type { FilterField, TableColumn } from '@/views/components/NeuroAgentListPage.vue'
import NeuroAgentListPage from '@/views/components/NeuroAgentListPage.vue'
import { usePageAction } from '@/composables/usePageAction'
import { usePermissionStore } from '@/stores/permission'

const { pageMenu } = usePageAction()
const perm = usePermissionStore()

const loading = ref(false)
const items = ref<AuditLogRow[]>([])
const total = ref(0)
const page = ref(1)
const limit = ref(10)
const appLoading = ref(false)
const apps = ref<{ app_code: string; app_name: string }[]>([])
const selectedAppCode = ref('')

const appliedKeyword = ref('')
const appliedTenantNameHint = ref('')
const appliedAccount = ref('')
const appliedIp = ref('')
const appliedDateFrom = ref('')
const appliedDateTo = ref('')

/** 与登录日志一致：平台管理员或平台主体租户内用户可跨主体筛选 */
const showTenantScopeFilters = computed(
  () => !!(perm.profile?.is_platform_admin || perm.profile?.tenant_is_platform),
)

const auditFilterFields = computed<FilterField[]>(() => {
  const rows: FilterField[] = [
    { key: 'keyword', label: '模糊关键字', type: 'text', placeholder: '模块/动作/摘要/详情' },
  ]
  if (showTenantScopeFilters.value) {
    rows.push({ key: 'tenantName', label: '主体', type: 'text', placeholder: '名称或编码模糊' })
  }
  rows.push(
    { key: 'account', label: '账号', type: 'text', placeholder: '工号/姓名/手机/邮箱' },
    { key: 'ip', label: 'IP', type: 'text', placeholder: '模糊匹配' },
    { key: 'opTime', label: '操作时间', type: 'daterange' },
  )
  return rows
})

const appFilterOptions = computed(() => [
  { app_code: '', app_name: '所有日志' },
  ...apps.value,
])

const selectedAppName = computed(() => {
  return appFilterOptions.value.find((item) => item.app_code === selectedAppCode.value)?.app_name || '所有日志'
})

async function load() {
  if (!pageMenu()) return
  loading.value = true
  try {
    const skip = (page.value - 1) * limit.value
    const res = await fetchAuditLogs(skip, limit.value, {
      keyword: appliedKeyword.value || undefined,
      app_code: selectedAppCode.value || undefined,
      tenant_name_hint: showTenantScopeFilters.value ? appliedTenantNameHint.value || undefined : undefined,
      account: appliedAccount.value || undefined,
      ip: appliedIp.value || undefined,
      date_from: appliedDateFrom.value || undefined,
      date_to: appliedDateTo.value || undefined,
    })
    items.value = res.items
    total.value = res.total
  } finally {
    loading.value = false
  }
}

async function loadApps() {
  appLoading.value = true
  try {
    const res = await fetchAppCenterApps({ skip: 0, limit: 100 })
    apps.value = res.items.map((item) => ({
      app_code: item.app_code,
      app_name: item.app_name,
    }))
  } catch {
    apps.value = [
      { app_code: 'app-center', app_name: '应用中心' },
      { app_code: 'system-management', app_name: '系统管理' },
      { app_code: 'system-monitor', app_name: '系统监控' },
      { app_code: 'model-manager', app_name: '模型管理' },
    ]
  } finally {
    appLoading.value = false
  }
}

function selectApp(appCode: string) {
  if (selectedAppCode.value === appCode) return
  selectedAppCode.value = appCode
  page.value = 1
  void load()
}

function onSearch(v: { keyword: string; filters: Record<string, unknown> }) {
  appliedKeyword.value = String(v.filters?.keyword ?? '').trim()
  appliedTenantNameHint.value = String(v.filters?.tenantName ?? '').trim()
  appliedAccount.value = String(v.filters?.account ?? '').trim()
  appliedIp.value = String(v.filters?.ip ?? '').trim()
  const dr = v.filters?.opTime as string[] | undefined
  if (Array.isArray(dr) && dr.length === 2 && dr[0] && dr[1]) {
    appliedDateFrom.value = String(dr[0]).trim()
    appliedDateTo.value = String(dr[1]).trim()
  } else {
    appliedDateFrom.value = ''
    appliedDateTo.value = ''
  }
  page.value = 1
  void load()
}

function onReset() {
  appliedKeyword.value = ''
  appliedTenantNameHint.value = ''
  appliedAccount.value = ''
  appliedIp.value = ''
  appliedDateFrom.value = ''
  appliedDateTo.value = ''
  page.value = 1
  void load()
}

function onPageChange(p: number) {
  page.value = p
  void load()
}

function onPageSizeChange(s: number) {
  limit.value = s
  page.value = 1
  void load()
}

onMounted(() => {
  void loadApps()
  void load()
})

const columns = computed<TableColumn[]>(() => [
  {
    key: 'operator',
    title: '操作方',
    width: 150,
  },
  { key: 'app_name', title: '应用', width: 108, formatter: (_v, row: AuditLogRow) => row.app_name || row.app_code || '—' },
  { key: 'action', title: '动作', width: 92 },
  {
    key: 'detail',
    title: '内容',
    minWidth: 180,
    formatter: (v) => (v != null && String(v).trim() !== '' ? String(v) : '—'),
  },
  {
    key: 'created_at',
    title: '操作时间',
    width: 142,
    tooltip: false,
    formatter: (v: unknown) => formatDateTimeChina(v as string | Date | null | undefined),
  },
])
</script>

<template>
  <div class="page">
    <el-empty v-if="!pageMenu()" description="无查看权限" />
    <div v-else class="audit-layout">
      <aside class="audit-app-filter">
        <div class="audit-app-filter__head">
          <strong>应用</strong>
          <span>{{ appFilterOptions.length - 1 }}</span>
        </div>
        <div v-loading="appLoading" class="audit-app-filter__list">
          <button
            v-for="app in appFilterOptions"
            :key="app.app_code || 'all'"
            type="button"
            :class="['audit-app-filter__item', { 'is-active': selectedAppCode === app.app_code }]"
            @click="selectApp(app.app_code)"
          >
            <span>{{ app.app_name }}</span>
            <code v-if="app.app_code">{{ app.app_code }}</code>
            <code v-else>ALL</code>
          </button>
        </div>
      </aside>
      <NeuroAgentListPage
        mode="el-table"
        :title="`操作日志 · ${selectedAppName}`"
        :columns="columns"
        :data="items"
        :loading="loading"
        :total="total"
        :page="page"
        :page-size="limit"
        :page-sizes="[10, 20, 50]"
        :show-create="false"
        :show-selection="false"
        :filter-fields="auditFilterFields"
        @search="onSearch"
        @reset="onReset"
        @page-change="onPageChange"
        @page-size-change="onPageSizeChange"
      >
        <template #col-operator="{ row }">
          <div class="audit-operator-cell">
            <strong>{{ row.tenant_name || (row.tenant_id != null ? `#${row.tenant_id}` : '未知主体') }}</strong>
            <span>
              {{ row.user_name || row.user_account || row.user_employee_no || (row.user_id != null ? `#${row.user_id}` : '系统') }}
              <template v-if="row.user_employee_no"> / {{ row.user_employee_no }}</template>
            </span>
          </div>
        </template>
        <template #col-detail="{ row }">
          <div class="audit-detail-cell">
            <strong>{{ row.summary || row.detail || '—' }}</strong>
            <span v-if="row.summary && row.detail && row.summary !== row.detail">{{ row.detail }}</span>
            <span>
              {{ row.module || '—' }}
              <template v-if="row.ip"> / IP {{ row.ip }}</template>
              <template v-if="row.id != null"> / #{{ row.id }}</template>
            </span>
          </div>
        </template>
        <!-- 插槽渲染避免 el-table 在部分场景仍用 prop 原始 ISO 串（带 T） -->
        <template #col-created_at="{ row }">
          {{ formatDateTimeChina(row.created_at) }}
        </template>
      </NeuroAgentListPage>
    </div>
  </div>
</template>

<style scoped>
.page {
  padding: 0;
}

.audit-layout {
  display: grid;
  grid-template-columns: 252px minmax(0, 1fr);
  gap: 20px;
  align-items: start;
  min-width: 0;
}

.audit-app-filter {
  margin-top: 12px;
  overflow: hidden;
  border: 1px solid color-mix(in srgb, var(--neuro-primary, #00f5d4) 24%, transparent);
  border-radius: 16px;
  background-color: color-mix(in srgb, var(--neuro-surface, #1e293b) 88%, var(--neuro-background, #0f172a));
  background-image:
    radial-gradient(120% 90% at 0% 0%, color-mix(in srgb, var(--neuro-primary, #00f5d4) 12%, transparent), transparent 58%),
    linear-gradient(168deg, color-mix(in srgb, var(--neuro-primary, #00f5d4) 5%, transparent) 0%, transparent 44%);
  box-shadow: inset 0 1px 0 color-mix(in srgb, #fff 5%, transparent);
}

.audit-layout :deep(.neuro-agent-list-page) {
  min-width: 0;
}

.audit-app-filter__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 18px;
  border-bottom: 1px solid color-mix(in srgb, var(--neuro-text, #f8fafc) 10%, transparent);
  color: var(--neuro-text, #f8fafc);
  font-size: 18px;
}

.audit-app-filter__head span {
  color: var(--neuro-text-secondary, #94a3b8);
  font-size: 12px;
}

.audit-app-filter__list {
  display: grid;
  gap: 8px;
  min-height: 120px;
  max-height: calc(100vh - 240px);
  overflow-y: auto;
  padding: 12px;
}

.audit-app-filter__item {
  display: grid;
  gap: 5px;
  width: 100%;
  padding: 12px 14px;
  border: 1px solid color-mix(in srgb, var(--neuro-text, #f8fafc) 12%, transparent);
  border-radius: 8px;
  background: color-mix(in srgb, var(--neuro-surface, #1e293b) 86%, var(--neuro-background, #0f172a));
  color: var(--neuro-text, #f8fafc);
  text-align: left;
  cursor: pointer;
  transition: border-color 0.18s ease, background-color 0.18s ease, box-shadow 0.18s ease, transform 0.18s ease;
}

.audit-app-filter__item:hover,
.audit-app-filter__item.is-active {
  border-color: var(--neuro-primary, #00f5d4);
  background: color-mix(in srgb, var(--neuro-primary, #00f5d4) 18%, var(--neuro-surface, #1e293b));
  box-shadow: 0 12px 28px color-mix(in srgb, var(--neuro-primary, #00f5d4) 10%, transparent);
}

.audit-app-filter__item:hover {
  transform: translateY(-1px);
}

.audit-app-filter__item code {
  overflow: hidden;
  color: var(--neuro-text-secondary, #94a3b8);
  font-size: 12px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.audit-operator-cell {
  display: grid;
  gap: 4px;
  min-width: 0;
}

.audit-operator-cell strong,
.audit-operator-cell span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.audit-operator-cell span {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.audit-detail-cell {
  display: grid;
  gap: 4px;
  min-width: 0;
}

.audit-detail-cell strong,
.audit-detail-cell span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.audit-detail-cell span {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.audit-layout :deep(.list-card),
.audit-layout :deep(.list-el-panel),
.audit-layout :deep(.card-search),
.audit-layout :deep(.card-table),
.audit-layout :deep(.el-table),
.audit-layout :deep(.el-table__inner-wrapper) {
  min-width: 0;
}

@media (max-width: 980px) {
  .audit-layout {
    grid-template-columns: 1fr;
  }

  .audit-app-filter__list {
    grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
    max-height: none;
  }
}
</style>
