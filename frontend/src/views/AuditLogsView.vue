<script setup lang="ts">
defineOptions({ name: 'AuditLogsView' })
import { computed, onMounted, ref } from 'vue'

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

async function load() {
  if (!pageMenu()) return
  loading.value = true
  try {
    const skip = (page.value - 1) * limit.value
    const res = await fetchAuditLogs(skip, limit.value, {
      keyword: appliedKeyword.value || undefined,
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
  void load()
})

const columns = computed<TableColumn[]>(() => [
  { key: 'id', title: 'ID', width: 72 },
  {
    key: 'tenant_id',
    title: '主体',
    width: 160,
    hidden: !showTenantScopeFilters.value,
    formatter: (_v: unknown, row: AuditLogRow) => {
      const n = row.tenant_name?.trim()
      if (n) return n
      if (row.tenant_id != null) return `#${row.tenant_id}`
      return '—'
    },
  },
  { key: 'module', title: '模块', width: 100 },
  { key: 'action', title: '动作', width: 88 },
  /** 靠前展示；用大 minWidth 参与分配剩余宽度，避免摘要独占摘要与 IP 之间的空隙 */
  {
    key: 'detail',
    title: '详情',
    minWidth: 420,
    formatter: (v) => (v != null && String(v).trim() !== '' ? String(v) : '—'),
  },
  { key: 'summary', title: '摘要', minWidth: 160 },
  { key: 'ip', title: 'IP', width: 130 },
  { key: 'user_id', title: '用户', width: 80 },
  {
    key: 'created_at',
    title: '操作时间',
    width: 196,
    tooltip: false,
    formatter: (v: unknown) => formatDateTimeChina(v as string | Date | null | undefined),
  },
])
</script>

<template>
  <div class="page">
    <el-empty v-if="!pageMenu()" description="无查看权限" />
    <NeuroAgentListPage
      v-else
      mode="el-table"
      title="操作日志"
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
      <!-- 插槽渲染避免 el-table 在部分场景仍用 prop 原始 ISO 串（带 T） -->
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
