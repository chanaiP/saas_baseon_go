<script setup lang="ts">
defineOptions({ name: 'ParamView' })
import { computed, onMounted, ref } from 'vue'

import { createSysParam, deleteSysParam, fetchSysParams, restoreSysParamDefault, updateSysParam } from '@/api/param'
import type { SysParamRow } from '@/api/param'
import { confirmArchiveAction } from '@/composables/useArchiveConfirm'
import type { FilterField, TableColumn } from '@/views/components/NeuroAgentListPage.vue'
import NeuroAgentDialog from '@/views/components/NeuroAgentDialog.vue'
import NeuroAgentListPage from '@/views/components/NeuroAgentListPage.vue'
import { usePermissionStore } from '@/stores/permission'

const loading = ref(false)
const items = ref<SysParamRow[]>([])
const total = ref(0)
const page = ref(1)
const limit = ref(10)
const appliedParamKeyKeyword = ref('')

const paramFilterFields: FilterField[] = [
  { key: 'paramKey', label: '参数键名', type: 'text', placeholder: '模糊匹配' },
]

const dlg = ref(false)
const dlgEdit = ref(false)
const form = ref({ param_key: '', param_value: '', remark: '', value_type: 'STRING', tenant_editable: true, is_platform_only: false })
const editRow = ref<SysParamRow | null>(null)

async function load() {
  loading.value = true
  try {
    const skip = (page.value - 1) * limit.value
    const res = await fetchSysParams(skip, limit.value, {
      keyword: appliedParamKeyKeyword.value || undefined,
    })
    items.value = res.items
    total.value = res.total
  } finally {
    loading.value = false
  }
}

function openDlg() {
  form.value = { param_key: '', param_value: '', remark: '', value_type: 'STRING', tenant_editable: true, is_platform_only: false }
  dlg.value = true
}

function openEdit(row: SysParamRow) {
  editRow.value = row
  form.value = {
    param_key: row.param_key,
    param_value: row.param_value || '',
    remark: row.remark || '',
    value_type: row.value_type || 'STRING',
    tenant_editable: row.tenant_editable !== false,
    is_platform_only: !!row.is_platform_only,
  }
  dlgEdit.value = true
}

async function save() {
  await createSysParam({
    param_key: form.value.param_key,
    param_value: form.value.param_value || undefined,
    remark: form.value.remark || undefined,
    value_type: form.value.value_type,
    tenant_editable: form.value.tenant_editable,
    is_platform_only: form.value.is_platform_only,
  })
  dlg.value = false
  await load()
}

async function saveEdit() {
  if (!editRow.value) return
  await updateSysParam(editRow.value.id, {
    param_value: form.value.param_value || null,
    remark: form.value.remark || null,
    value_type: form.value.value_type,
    tenant_editable: form.value.tenant_editable,
    is_platform_only: form.value.is_platform_only,
  })
  dlgEdit.value = false
  await load()
}

async function remove(row: SysParamRow) {
  await confirmArchiveAction({ name: row.param_key, title: '归档系统参数' })
  await deleteSysParam(row.id)
  await load()
}

async function restoreDefault(row: SysParamRow) {
  await restoreSysParamDefault(row.id)
  await load()
}

function onPageChange(p: number) {
  page.value = p
  load()
}

function onPageSizeChange(s: number) {
  limit.value = s
  page.value = 1
  load()
}

function onParamSearch(payload: { keyword: string; filters: Record<string, unknown> }) {
  appliedParamKeyKeyword.value = String(payload.filters?.paramKey ?? '').trim()
  page.value = 1
  void load()
}

const permStore = usePermissionStore()
const isPlatformAdmin = computed(
  () => !!(permStore.profile?.is_platform_admin || permStore.profile?.tenant_is_platform),
)
const hasParamActionColumn = computed(
  () =>
    permStore.can('param:create') || permStore.can('param:edit') || permStore.can('param:delete'),
)
const showParamCreate = computed(() => permStore.can('param:create'))
function canDeleteParam(row: SysParamRow) {
  return permStore.can('param:delete') && (isPlatformAdmin.value || row.is_tenant_owned === true)
}

const columns = computed<TableColumn[]>(() => [
  { key: 'param_key', title: '参数键名', minWidth: 280, width: 360 },
  { key: 'param_value', title: isPlatformAdmin.value ? '默认值' : '当前值' },
  { key: 'value_type', title: '类型', width: 100 },
  { key: 'tenant_editable', title: '租户覆盖', width: 110 },
  { key: 'is_override', title: '覆盖', width: 80, hidden: isPlatformAdmin.value },
  { key: 'remark', title: '备注' },
  { key: 'actions', title: '操作', width: isPlatformAdmin.value ? 160 : 190, fixed: 'right', tooltip: false, hidden: !hasParamActionColumn.value },
])

onMounted(load)
</script>

<template>
  <div class="page">
    <NeuroAgentListPage
      mode="el-table"
      title="系统参数"
      :columns="columns"
      :data="items"
      :loading="loading"
      :total="total"
      :page="page"
      :page-size="limit"
      :page-sizes="[10, 20, 50]"
      :show-create="showParamCreate"
      :show-selection="false"
      :filter-fields="paramFilterFields"
      @create="openDlg"
      @search="onParamSearch"
      @page-change="onPageChange"
      @page-size-change="onPageSizeChange"
    >
      <template #actions>
        <el-button v-permission="'param:create'" class="btn-gradient" @click="openDlg">新增</el-button>
      </template>
      <template #col-tenant_editable="{ row }">
        <el-tag size="small" :type="row.tenant_editable === false ? 'info' : 'success'">
          {{ row.tenant_editable === false ? '不允许' : '允许' }}
        </el-tag>
      </template>
      <template #col-is_override="{ row }">
        <el-tag size="small" :type="row.is_override ? 'warning' : 'info'">
          {{ row.is_override ? '已覆盖' : '默认' }}
        </el-tag>
      </template>
      <template #col-actions="{ row }">
        <span class="op-btns">
          <el-button
            v-permission="'param:edit'"
            :disabled="!isPlatformAdmin && row.tenant_editable === false"
            @click="openEdit(row)"
          >
            {{ isPlatformAdmin ? '编辑' : '覆盖' }}
          </el-button>
          <el-button v-if="!isPlatformAdmin" v-permission="'param:edit'" :disabled="!row.is_override" @click="restoreDefault(row)">恢复默认</el-button>
          <el-button v-if="canDeleteParam(row)" v-permission="'param:delete'" type="danger" @click="remove(row)">删除</el-button>
        </span>
      </template>
    </NeuroAgentListPage>

    <NeuroAgentDialog v-model="dlg" title="新增参数" icon="⚙️" size="medium">
      <div class="nm-form">
        <div class="nm-form-item">
          <label class="nm-form-label">键</label>
          <el-input v-model="form.param_key" />
        </div>
        <div class="nm-form-item">
          <label class="nm-form-label">值</label>
          <el-input v-model="form.param_value" />
        </div>
        <div class="nm-form-item">
          <label class="nm-form-label">备注</label>
          <el-input v-model="form.remark" />
        </div>
        <div class="nm-form-row">
          <div class="nm-form-item">
            <label class="nm-form-label">类型</label>
            <el-select v-model="form.value_type">
              <el-option label="字符串" value="STRING" />
              <el-option label="数字" value="NUMBER" />
              <el-option label="布尔" value="BOOLEAN" />
              <el-option label="JSON" value="JSON" />
            </el-select>
          </div>
          <div class="nm-form-item">
            <label class="nm-form-label">租户覆盖</label>
            <el-switch v-model="form.tenant_editable" />
          </div>
          <div class="nm-form-item">
            <label class="nm-form-label">平台专属</label>
            <el-switch v-model="form.is_platform_only" :disabled="!isPlatformAdmin" />
          </div>
        </div>
      </div>
      <template #footer-right>
        <button class="nm-btn nm-btn--primary" @click="save">保存</button>
      </template>
    </NeuroAgentDialog>

    <NeuroAgentDialog v-model="dlgEdit" title="编辑参数" icon="⚙️" size="medium">
      <div class="nm-form">
        <div class="nm-form-item">
          <label class="nm-form-label">键</label>
          <el-input v-model="form.param_key" disabled />
        </div>
        <div class="nm-form-item">
          <label class="nm-form-label">值</label>
          <el-input v-model="form.param_value" />
        </div>
        <div class="nm-form-item">
          <label class="nm-form-label">备注</label>
          <el-input v-model="form.remark" :disabled="!isPlatformAdmin" />
        </div>
        <div v-if="isPlatformAdmin" class="nm-form-row">
          <div class="nm-form-item">
            <label class="nm-form-label">类型</label>
            <el-select v-model="form.value_type">
              <el-option label="字符串" value="STRING" />
              <el-option label="数字" value="NUMBER" />
              <el-option label="布尔" value="BOOLEAN" />
              <el-option label="JSON" value="JSON" />
            </el-select>
          </div>
          <div class="nm-form-item">
            <label class="nm-form-label">租户覆盖</label>
            <el-switch v-model="form.tenant_editable" />
          </div>
          <div class="nm-form-item">
            <label class="nm-form-label">平台专属</label>
            <el-switch v-model="form.is_platform_only" />
          </div>
        </div>
      </div>
      <template #footer-right>
        <button class="nm-btn nm-btn--primary" @click="saveEdit">保存</button>
      </template>
    </NeuroAgentDialog>
  </div>
</template>

<style scoped>
.page {
  padding: 16px;
}
</style>
