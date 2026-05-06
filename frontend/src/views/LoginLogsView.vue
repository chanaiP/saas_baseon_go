<script setup lang="ts">
defineOptions({ name: 'LoginLogsView' })
import { computed, onMounted, ref } from 'vue'

import { fetchLoginLogs } from '@/api/logs'
import type { LoginLogRow } from '@/api/logs'
import { formatDateTimeChina } from '@/utils/datetime'
import type { FilterField, TableColumn } from '@/views/components/NeuroAgentListPage.vue'
import NeuroAgentListPage from '@/views/components/NeuroAgentListPage.vue'
import { usePageAction } from '@/composables/usePageAction'
import { usePermissionStore } from '@/stores/permission'

const { pageAction } = usePageAction()
const perm = usePermissionStore()

const loading = ref(false)
const items = ref<LoginLogRow[]>([])
const total = ref(0)
const page = ref(1)
const limit = ref(20)
const successFilter = ref<boolean | undefined>(undefined)
const appliedAccount = ref('')
const appliedIp = ref('')
const appliedTenantNameHint = ref('')
const appliedDateFrom = ref('')
const appliedDateTo = ref('')

/** 与操作日志一致：平台管理员或平台主体租户内用户可跨主体筛选 */
const showTenantScopeFilters = computed(
  () => !!(perm.profile?.is_platform_admin || perm.profile?.tenant_is_platform),
)

async function load() {
  if (!pageAction('login:view')) return
  loading.value = true
  try {
    const skip = (page.value - 1) * limit.value
    const res = await fetchLoginLogs(skip, limit.value, {
      success: successFilter.value,
      account: appliedAccount.value || undefined,
      ip: appliedIp.value || undefined,
      tenant_name_hint: showTenantScopeFilters.value ? appliedTenantNameHint.value || undefined : undefined,
      date_from: appliedDateFrom.value || undefined,
      date_to: appliedDateTo.value || undefined,
    })
    items.value = res.items
    total.value = res.total
  } finally {
    loading.value = false
  }
}

function onSearch(v: { keyword: string; filters: Record<string, any> }) {
  const s = v.filters?.success
  successFilter.value = s === '' || s === undefined || s === null ? undefined : Boolean(s)
  appliedAccount.value = String(v.filters?.account ?? '').trim()
  appliedIp.value = String(v.filters?.ip ?? '').trim()
  appliedTenantNameHint.value = String(v.filters?.tenantName ?? '').trim()
  const dr = v.filters?.loginDate as string[] | undefined
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
  successFilter.value = undefined
  appliedAccount.value = ''
  appliedIp.value = ''
  appliedTenantNameHint.value = ''
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
  void load()
})

const columns = computed<TableColumn[]>(() => [
  { key: 'id', title: 'ID', width: 72 },
  { key: 'account', title: '账号', width: 140 },
  { key: 'result', title: '结果', width: 88 },
  { key: 'message', title: '说明', minWidth: 160 },
  { key: 'ip', title: 'IP', width: 130 },
  {
    key: 'tenant_id',
    title: '主体',
    width: 160,
    hidden: !showTenantScopeFilters.value,
    formatter: (_v: unknown, row: LoginLogRow) => {
      const n = row.tenant_name?.trim()
      if (n) return n
      if (row.tenant_id != null) return `#${row.tenant_id}`
      return '—'
    },
  },
  {
    key: 'created_at',
    title: '时间',
    width: 196,
    tooltip: false,
    formatter: (v: unknown) => formatDateTimeChina(v as string | Date | null | undefined),
  },
])

const loginFilterFields = computed<FilterField[]>(() => {
  const rows: FilterField[] = [
    { key: 'account', label: '账号', type: 'text', placeholder: '模糊匹配' },
    { key: 'ip', label: 'IP', type: 'text', placeholder: '模糊匹配' },
  ]
  if (showTenantScopeFilters.value) {
    rows.push({ key: 'tenantName', label: '主体', type: 'text', placeholder: '名称或编码模糊' })
  }
  rows.push(
    { key: 'loginDate', label: '登录日期', type: 'daterange' },
    {
      key: 'success',
      label: '结果',
      type: 'select',
      placeholder: '全部',
      options: [
        { label: '全部', value: '' },
        { label: '成功', value: true },
        { label: '失败', value: false },
      ],
    },
  )
  return rows
})
</script>

<template>
  <div class="page">
    <el-empty v-if="!pageAction('login:view')" description="无查看权限" />
    <NeuroAgentListPage
      v-else
      mode="el-table"
      title="登录日志"
      :columns="columns"
      :data="items"
      :loading="loading"
      :total="total"
      :page-sizes="[10, 20, 50]"
      :show-create="false"
      :show-selection="false"
      :filter-fields="loginFilterFields"
      @search="onSearch"
      @reset="onReset"
      @page-change="onPageChange"
      @page-size-change="onPageSizeChange"
    >
      <template #col-result="{ row }">
        <el-tag effect="plain" size="small" :type="row.success ? 'success' : 'danger'">
          {{ row.success ? '成功' : '失败' }}
        </el-tag>
      </template>
      <template #col-created_at="{ row }">
        {{ formatDateTimeChina(row.created_at) }}
      </template>
    </NeuroAgentListPage>
  </div>
</template>

<style scoped>
.page {
  padding: 0;
}
</style>
