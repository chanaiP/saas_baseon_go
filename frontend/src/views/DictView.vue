<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'

import {
  createDictItem,
  createDictType,
  deleteDictItem,
  deleteDictType,
  fetchDictItems,
  fetchDictTypes,
  restoreDictItemDefault,
  updateDictItem,
  updateDictType,
} from '@/api/dict'
import { confirmArchiveAction } from '@/composables/useArchiveConfirm'
import type { DictItemRow, DictTypeRow } from '@/api/dict'
import { usePermissionStore } from '@/stores/permission'
import type { FilterField, TableColumn } from '@/views/components/NeuroAgentListPage.vue'
import NeuroAgentDialog from '@/views/components/NeuroAgentDialog.vue'
import NeuroAgentListPage from '@/views/components/NeuroAgentListPage.vue'

const loadingT = ref(false)
const permStore = usePermissionStore()
const isPlatformAdmin = computed(
  () => !!(permStore.profile?.is_platform_admin || permStore.profile?.tenant_is_platform),
)
const types = ref<DictTypeRow[]>([])
const totalT = ref(0)
const pageT = ref(1)
const limitT = ref(10)
const dictTypeNameOrCode = ref('')
const dictTypePlatformOnly = ref<boolean | undefined>(undefined)
const currentType = ref<DictTypeRow | null>(null)

const loadingI = ref(false)
const items = ref<DictItemRow[]>([])
const totalI = ref(0)
const pageI = ref(1)
const limitI = ref(10)

const dlgT = ref(false)
const dlgTEdit = ref(false)
const tForm = ref({ code: '', name: '', remark: '', scope: 'HYBRID', tenant_editable: true, is_platform_only: false })
const tEditId = ref<number | null>(null)

const dlgI = ref(false)
const dlgIEdit = ref(false)
const iForm = ref({ label: '', value: '', sort_order: 0, enabled: true })
const iEdit = ref<DictItemRow | null>(null)

const dictTypeFilterFields = computed<FilterField[]>(() => {
  const rows: FilterField[] = [
    { key: 'nameOrCode', label: '名称 / 编码', type: 'text', placeholder: '模糊匹配' },
  ]
  if (isPlatformAdmin.value) {
    rows.push({
      key: 'platformOnly',
      label: '平台专属',
      type: 'select',
      options: [
        { label: '全部', value: '' },
        { label: '是', value: '1' },
        { label: '否', value: '0' },
      ],
    })
  }
  return rows
})

async function loadTypes() {
  loadingT.value = true
  try {
    const skip = (pageT.value - 1) * limitT.value
    const res = await fetchDictTypes(skip, limitT.value, {
      keyword: dictTypeNameOrCode.value || undefined,
      platform_only: isPlatformAdmin.value ? dictTypePlatformOnly.value : undefined,
    })
    types.value = res.items
    totalT.value = res.total
    if (!res.items.length) {
      currentType.value = null
    } else if (!currentType.value) {
      currentType.value = res.items[0]
    } else if (currentType.value) {
      const still = res.items.find((x) => x.id === currentType.value!.id)
      if (!still && res.items.length) currentType.value = res.items[0]
      else if (still) currentType.value = still
    }
  } finally {
    loadingT.value = false
  }
}

/** 左侧选中类型变化时拉取右侧字典项（含首次 loadTypes 赋值默认选中） */
watch(
  currentType,
  () => {
    pageI.value = 1
    void loadItems()
  },
)

async function loadItems() {
  if (!currentType.value) {
    items.value = []
    totalI.value = 0
    return
  }
  loadingI.value = true
  try {
    const skip = (pageI.value - 1) * limitI.value
    const res = await fetchDictItems(currentType.value.id, skip, limitI.value)
    items.value = res.items
    totalI.value = res.total
  } finally {
    loadingI.value = false
  }
}

function onTypeRowChange(row: DictTypeRow | undefined) {
  if (row) currentType.value = row
}

function onTypePageChange(p: number) {
  pageT.value = p
  void loadTypes()
}

function onTypePageSizeChange(s: number) {
  limitT.value = s
  pageT.value = 1
  void loadTypes()
}

function onDictTypeSearch(payload: { keyword: string; filters: Record<string, unknown> }) {
  dictTypeNameOrCode.value = String(payload.filters?.nameOrCode ?? '').trim()
  const po = String(payload.filters?.platformOnly ?? '')
  if (po === '1') dictTypePlatformOnly.value = true
  else if (po === '0') dictTypePlatformOnly.value = false
  else dictTypePlatformOnly.value = undefined
  pageT.value = 1
  void loadTypes()
}

function onItemPageChange(p: number) {
  pageI.value = p
  void loadItems()
}

function onItemPageSizeChange(s: number) {
  limitI.value = s
  pageI.value = 1
  void loadItems()
}

function openTypeDlg() {
  tEditId.value = null
  tForm.value = { code: '', name: '', remark: '', scope: 'HYBRID', tenant_editable: true, is_platform_only: false }
  dlgT.value = true
}

function openTypeEdit(row: DictTypeRow) {
  tEditId.value = row.id
  tForm.value = {
    code: row.code,
    name: row.name,
    remark: row.remark || '',
    scope: row.scope || 'HYBRID',
    tenant_editable: row.tenant_editable !== false,
    is_platform_only: !!row.is_platform_only,
  }
  dlgTEdit.value = true
}

async function saveType() {
  await createDictType({
    code: tForm.value.code,
    name: tForm.value.name,
    remark: tForm.value.remark || undefined,
    scope: tForm.value.scope,
    tenant_editable: tForm.value.tenant_editable,
    is_platform_only: tForm.value.is_platform_only,
  })
  dlgT.value = false
  await loadTypes()
}

async function saveTypeEdit() {
  if (!tEditId.value) return
  await updateDictType(tEditId.value, {
    name: tForm.value.name,
    remark: tForm.value.remark || null,
    scope: tForm.value.scope,
    tenant_editable: tForm.value.tenant_editable,
    is_platform_only: tForm.value.is_platform_only,
  })
  dlgTEdit.value = false
  await loadTypes()
}

async function removeType(row: DictTypeRow) {
  await confirmArchiveAction({ name: row.name || row.code, title: '归档字典类型', detail: '仅无字典项的类型可归档；历史定义仍会保留。' })
  await deleteDictType(row.id)
  if (currentType.value?.id === row.id) currentType.value = null
  await loadTypes()
}

function openItemDlg() {
  if (!currentType.value) return
  iEdit.value = null
  iForm.value = { label: '', value: '', sort_order: 0, enabled: true }
  dlgI.value = true
}

function openItemEdit(row: DictItemRow) {
  iEdit.value = row
  iForm.value = { label: row.label, value: row.value, sort_order: row.sort_order, enabled: row.enabled !== false }
  dlgIEdit.value = true
}

async function saveItem() {
  if (!currentType.value) return
  await createDictItem({
    dict_type_id: currentType.value.id,
    label: iForm.value.label,
    value: iForm.value.value,
    sort_order: iForm.value.sort_order,
    enabled: iForm.value.enabled,
  })
  dlgI.value = false
  await loadItems()
}

async function saveItemEdit() {
  if (!iEdit.value) return
  await updateDictItem(iEdit.value.id, {
    label: iForm.value.label,
    value: iForm.value.value,
    sort_order: iForm.value.sort_order,
    enabled: iForm.value.enabled,
  })
  dlgIEdit.value = false
  await loadItems()
}

async function removeItem(row: DictItemRow) {
  await confirmArchiveAction({ name: row.label || row.value, title: '归档字典项' })
  await deleteDictItem(row.id)
  await loadItems()
}

async function restoreItem(row: DictItemRow) {
  await restoreDictItemDefault(row.id)
  await loadItems()
}

const typeColumns = computed<TableColumn[]>(() => [
  { key: 'name_code', title: '名称 / 编码', minWidth: 196, tooltip: true },
  { key: 'scope', title: '作用域', minWidth: 100, width: 108 },
  { key: 'tenant_editable', title: '租户覆盖', minWidth: 112, width: 120 },
  { key: 'is_platform_only', title: '平台专属', minWidth: 100, width: 108, hidden: !isPlatformAdmin.value },
  { key: 'actions', title: '操作', width: 156, minWidth: 156, tooltip: false, hidden: !isPlatformAdmin.value },
])

const itemColumns = computed<TableColumn[]>(() => [
  { key: 'default_label', title: '默认标签', minWidth: 100, hidden: isPlatformAdmin.value },
  { key: 'label', title: isPlatformAdmin.value ? '默认标签' : '当前标签', minWidth: isPlatformAdmin.value ? 100 : 130 },
  { key: 'default_value', title: '默认值', minWidth: 100, hidden: isPlatformAdmin.value },
  { key: 'value', title: isPlatformAdmin.value ? '默认值' : '当前值', minWidth: isPlatformAdmin.value ? 100 : 130 },
  { key: 'sort_order', title: '排序', minWidth: 96, width: 96, align: 'center' },
  { key: 'enabled', title: '启用', minWidth: 96, width: 96, align: 'center' },
  { key: 'is_override', title: '覆盖', minWidth: 96, width: 96, align: 'center', hidden: isPlatformAdmin.value },
  { key: 'actions', title: '操作', width: isPlatformAdmin.value ? 156 : 228, minWidth: isPlatformAdmin.value ? 156 : 228, tooltip: false },
])

onMounted(() => void loadTypes())
</script>

<template>
  <div class="page">
    <el-row :gutter="20">
      <el-col :xs="24" :lg="10" class="dict-data-table">
        <NeuroAgentListPage
          mode="el-table"
          title="字典类型"
          :columns="typeColumns"
          :data="types"
          :loading="loadingT"
          :total="totalT"
          :page="pageT"
          :page-size="limitT"
          :page-sizes="[10, 20, 50]"
          :show-create="isPlatformAdmin"
          :show-selection="false"
          :current-row-id="currentType?.id ?? undefined"
          :filter-fields="dictTypeFilterFields"
          @create="openTypeDlg"
          @row-click="onTypeRowChange"
          @search="onDictTypeSearch"
          @page-change="onTypePageChange"
          @page-size-change="onTypePageSizeChange"
        >
          <template #actions>
            <el-button v-if="isPlatformAdmin" v-permission="'dict_type:create'" class="btn-gradient" @click="openTypeDlg">新增</el-button>
          </template>
          <template #col-name_code="{ row }">
            <div class="dict-type-title-cell">
              <div class="dict-type-title-cell__name">{{ row.name }}</div>
              <code class="dict-type-title-cell__code">{{ row.code }}</code>
            </div>
          </template>
          <template #col-tenant_editable="{ row }">
            <el-tag size="small" :type="row.tenant_editable === false ? 'info' : 'success'">
              {{ row.tenant_editable === false ? '不允许' : '允许' }}
            </el-tag>
          </template>
          <template #col-is_platform_only="{ row }">
            <el-tag size="small" :type="row.is_platform_only ? 'warning' : 'info'">
              {{ row.is_platform_only ? '是' : '否' }}
            </el-tag>
          </template>
          <template #col-actions="{ row }">
            <span class="op-btns">
              <el-button v-permission="'dict_type:edit'" size="small" @click.stop="openTypeEdit(row)">编辑</el-button>
              <el-button v-permission="'dict_type:delete'" type="danger" size="small" @click.stop="removeType(row)">删除</el-button>
            </span>
          </template>
        </NeuroAgentListPage>
      </el-col>
      <el-col :xs="24" :lg="14" class="dict-data-table">
        <el-empty v-if="!currentType" :description="`请选择字典类型`" />
        <NeuroAgentListPage
          v-else
          mode="el-table"
          :title="`字典项 · ${currentType.name}`"
          :columns="itemColumns"
          :data="items"
          :loading="loadingI"
          :total="totalI"
          :page="pageI"
          :page-size="limitI"
          :page-sizes="[10, 20, 50]"
          :show-create="isPlatformAdmin"
          :show-selection="false"
          @create="openItemDlg"
          @page-change="onItemPageChange"
          @page-size-change="onItemPageSizeChange"
        >
          <template #actions>
            <el-button v-if="isPlatformAdmin" v-permission="'dict_item:create'" class="btn-gradient" @click="openItemDlg">新增</el-button>
          </template>
          <template #col-enabled="{ row }">
            <el-tag
              size="small"
              effect="plain"
              round
              :type="row.enabled === false ? undefined : 'success'"
              :class="{ 'nm-status-pill--inactive': row.enabled === false }"
            >
              {{ row.enabled === false ? '停用' : '启用' }}
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
                v-permission="'dict_item:edit'"
                :disabled="!isPlatformAdmin && currentType?.tenant_editable === false"
                @click="openItemEdit(row)"
              >
                {{ isPlatformAdmin ? '编辑' : '覆盖' }}
              </el-button>
              <el-button v-if="!isPlatformAdmin" :disabled="!row.is_override" @click="restoreItem(row)">恢复默认</el-button>
              <el-button v-if="isPlatformAdmin" v-permission="'dict_item:delete'" type="danger" @click="removeItem(row)">删除</el-button>
            </span>
          </template>
        </NeuroAgentListPage>
      </el-col>
    </el-row>

    <NeuroAgentDialog v-model="dlgT" title="新增类型" icon="📁" size="small">
      <div class="nm-form">
        <div class="nm-form-row">
          <div class="nm-form-item">
            <label class="nm-form-label">编码</label>
            <el-input v-model="tForm.code" />
          </div>
          <div class="nm-form-item">
            <label class="nm-form-label">名称</label>
            <el-input v-model="tForm.name" />
          </div>
        </div>
        <div class="nm-form-item">
          <label class="nm-form-label">备注</label>
          <el-input v-model="tForm.remark" />
        </div>
        <div class="nm-form-row">
          <div class="nm-form-item">
            <label class="nm-form-label">作用域</label>
            <el-select v-model="tForm.scope">
              <el-option label="混合覆盖" value="HYBRID" />
              <el-option label="租户数据" value="TENANT" />
              <el-option label="平台定义" value="PLATFORM" />
            </el-select>
          </div>
          <div class="nm-form-item">
            <label class="nm-form-label">租户覆盖</label>
            <el-switch v-model="tForm.tenant_editable" />
          </div>
          <div class="nm-form-item">
            <label class="nm-form-label">平台专属</label>
            <el-switch v-model="tForm.is_platform_only" />
          </div>
        </div>
      </div>
      <template #footer-right>
        <button class="nm-btn nm-btn--primary" @click="saveType">保存</button>
      </template>
    </NeuroAgentDialog>

    <NeuroAgentDialog v-model="dlgTEdit" title="编辑类型" icon="📁" size="small">
      <div class="nm-form">
        <div class="nm-form-row">
          <div class="nm-form-item">
            <label class="nm-form-label">编码</label>
            <el-input v-model="tForm.code" disabled />
          </div>
          <div class="nm-form-item">
            <label class="nm-form-label">名称</label>
            <el-input v-model="tForm.name" />
          </div>
        </div>
        <div class="nm-form-item">
          <label class="nm-form-label">备注</label>
          <el-input v-model="tForm.remark" />
        </div>
        <div class="nm-form-row">
          <div class="nm-form-item">
            <label class="nm-form-label">作用域</label>
            <el-select v-model="tForm.scope">
              <el-option label="混合覆盖" value="HYBRID" />
              <el-option label="租户数据" value="TENANT" />
              <el-option label="平台定义" value="PLATFORM" />
            </el-select>
          </div>
          <div class="nm-form-item">
            <label class="nm-form-label">租户覆盖</label>
            <el-switch v-model="tForm.tenant_editable" />
          </div>
          <div class="nm-form-item">
            <label class="nm-form-label">平台专属</label>
            <el-switch v-model="tForm.is_platform_only" />
          </div>
        </div>
      </div>
      <template #footer-right>
        <button class="nm-btn nm-btn--primary" @click="saveTypeEdit">保存</button>
      </template>
    </NeuroAgentDialog>

    <NeuroAgentDialog v-model="dlgI" title="新增字典项" icon="📋" size="small">
      <div class="nm-form">
        <div class="nm-form-row">
          <div class="nm-form-item">
            <label class="nm-form-label">标签</label>
            <el-input v-model="iForm.label" />
          </div>
          <div class="nm-form-item">
            <label class="nm-form-label">值</label>
            <el-input v-model="iForm.value" />
          </div>
        </div>
        <div class="nm-form-item">
          <label class="nm-form-label">排序</label>
          <el-input-number v-model="iForm.sort_order" />
        </div>
        <div class="nm-form-item">
          <label class="nm-form-label">启用</label>
          <el-switch v-model="iForm.enabled" />
        </div>
      </div>
      <template #footer-right>
        <button class="nm-btn nm-btn--primary" @click="saveItem">保存</button>
      </template>
    </NeuroAgentDialog>

    <NeuroAgentDialog v-model="dlgIEdit" title="编辑字典项" icon="📋" size="small">
      <div class="nm-form">
        <div class="nm-form-row">
          <div class="nm-form-item">
            <label class="nm-form-label">标签</label>
            <el-input v-model="iForm.label" />
          </div>
          <div class="nm-form-item">
            <label class="nm-form-label">值</label>
            <el-input v-model="iForm.value" />
          </div>
        </div>
        <div class="nm-form-item">
          <label class="nm-form-label">排序</label>
          <el-input-number v-model="iForm.sort_order" />
        </div>
        <div class="nm-form-item">
          <label class="nm-form-label">启用</label>
          <el-switch v-model="iForm.enabled" />
        </div>
      </div>
      <template #footer-right>
        <button class="nm-btn nm-btn--primary" @click="saveItemEdit">保存</button>
      </template>
    </NeuroAgentDialog>
  </div>
</template>

<style scoped>
.page {
  padding: 16px;
}

.dict-data-table {
  min-width: 0;
}

/* 字典页表格单元格略增左右留白，避免列内容贴在分隔线上 */
.page :deep(.dict-data-table .neuro-el-table .cell) {
  padding-left: 10px;
  padding-right: 10px;
}

.page :deep(.dict-data-table .op-btns) {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
  max-width: 100%;
}

.page :deep(.dict-data-table .op-btns .el-button) {
  margin-left: 0;
  min-width: 64px;
}

.page :deep(.dict-data-table .card-table),
.page :deep(.dict-data-table .el-table__inner-wrapper),
.page :deep(.dict-data-table .el-table__body-wrapper),
.page :deep(.dict-data-table .el-scrollbar),
.page :deep(.dict-data-table .el-scrollbar__wrap),
.page :deep(.dict-data-table .el-scrollbar__view) {
  height: auto !important;
  max-height: none !important;
  overflow-y: visible !important;
}

.page :deep(.dict-data-table .el-scrollbar__bar) {
  display: none !important;
}

.page :deep(.dict-data-table .el-table__body-wrapper) {
  scrollbar-width: none;
}

.page :deep(.dict-data-table .el-table__body-wrapper::-webkit-scrollbar) {
  width: 0;
  height: 0;
}

.dict-type-title-cell {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 6px;
  min-width: 0;
  padding: 2px 0;
}

.dict-type-title-cell__name {
  font-size: 13px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  line-height: 1.35;
  word-break: break-word;
}

.dict-type-title-cell__code {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 12px;
  font-weight: 500;
  color: var(--el-text-color-regular);
  background: var(--el-fill-color-light);
  padding: 3px 8px;
  border-radius: 6px;
  max-width: 100%;
  word-break: break-all;
}
</style>
