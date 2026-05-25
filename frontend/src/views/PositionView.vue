<script setup lang="ts">
defineOptions({ name: 'PositionView' })
import { ElMessage } from 'element-plus'
import { computed, onMounted, ref } from 'vue'

import {
  createPosition,
  createPositionType,
  deletePosition,
  deletePositionType,
  fetchPositions,
  fetchPositionTypes,
  updatePosition,
  updatePositionType,
} from '@/api/position'
import type { PositionRow, PositionTypeRow } from '@/api/position'
import { confirmArchiveAction } from '@/composables/useArchiveConfirm'
import type { FilterField, TableColumn } from '@/views/components/NeuroAgentListPage.vue'
import NeuroAgentDialog from '@/views/components/NeuroAgentDialog.vue'
import NeuroAgentListPage from '@/views/components/NeuroAgentListPage.vue'

const types = ref<PositionTypeRow[]>([])
const totalT = ref(0)
const pageT = ref(1)
const limitT = ref(10)
const selectedType = ref<number | null>(null)
/** 左侧选中类型的展示信息；分页导致当前页不含选中行时仍用于右侧标题与表单说明 */
const selectedTypeMeta = ref<{ id: number; name: string; code: string } | null>(null)
const typeOptions = ref<PositionTypeRow[]>([])

const positions = ref<PositionRow[]>([])
const totalP = ref(0)
const pageP = ref(1)
const limitP = ref(10)
/** 岗位列表：名称或编码合一筛选（与查询按钮联动） */
const posNameOrCode = ref('')
/** 岗位列表：岗位类型筛选；为空表示全部岗位 */
const posTypeFilter = ref<number | ''>('')
/** 岗位类型列表：名称或编码（与查询按钮联动） */
const typeKw = ref('')

const typeFilterFields: FilterField[] = [
  { key: 'typeHint', label: '名称 / 编码', type: 'text', placeholder: '模糊匹配' },
]

const posFilterFields = computed<FilterField[]>(() => [
  { key: 'nameOrCode', label: '名称 / 编码', type: 'text', placeholder: '模糊匹配', defaultValue: posNameOrCode.value },
  {
    key: 'positionTypeId',
    label: '岗位类型',
    type: 'select',
    placeholder: '全部岗位类型',
    defaultValue: posTypeFilter.value,
    options: typeOptions.value.map((item) => ({ label: `${item.name}（${item.code}）`, value: item.id })),
  },
])

const dlgT = ref(false)
const dlgTEdit = ref(false)
const tForm = ref({ name: '', code: '' })
const tEdit = ref<PositionTypeRow | null>(null)

const dlgP = ref(false)
const dlgPEdit = ref(false)
const pForm = ref({ name: '', code: '', position_type_id: 0 as number })
const pEdit = ref<PositionRow | null>(null)

async function loadTypeOptions() {
  const res = await fetchPositionTypes(0, 500)
  typeOptions.value = res.items
}

async function loadTypes() {
  const skip = (pageT.value - 1) * limitT.value
  const res = await fetchPositionTypes(skip, limitT.value, typeKw.value || undefined)
  types.value = res.items
  totalT.value = res.total
  if (selectedType.value) {
    const hit = res.items.find((x) => x.id === selectedType.value)
    if (hit) {
      selectedTypeMeta.value = { id: hit.id, name: hit.name, code: hit.code }
    }
  }
}

async function loadPos() {
  const skip = (pageP.value - 1) * limitP.value
  const res = await fetchPositions({
    position_type_id: posTypeFilter.value === '' ? undefined : Number(posTypeFilter.value),
    keyword: posNameOrCode.value || undefined,
    skip,
    limit: limitP.value,
  })
  positions.value = res.items
  totalP.value = res.total
}

function onTypeRowChange(row: PositionTypeRow | undefined) {
  if (!row) return
  selectedType.value = row.id
  selectedTypeMeta.value = { id: row.id, name: row.name, code: row.code }
  posTypeFilter.value = row.id
  pageP.value = 1
  void loadPos()
}

function onTypePageChange(p: number) {
  pageT.value = p
  void loadTypes().then(() => loadPos())
}

function onTypePageSizeChange(s: number) {
  limitT.value = s
  pageT.value = 1
  void loadTypes().then(() => loadPos())
}

function onTypeSearch(payload: { keyword: string; filters: Record<string, any> }) {
  typeKw.value = String(payload.filters?.typeHint ?? '').trim()
  pageT.value = 1
  void loadTypes()
}

function onPosSearch(v: { keyword: string; filters: Record<string, any> }) {
  posNameOrCode.value = String(v.filters?.nameOrCode ?? '').trim()
  const nextTypeID = v.filters?.positionTypeId
  posTypeFilter.value = nextTypeID === '' || nextTypeID == null ? '' : Number(nextTypeID)
  if (posTypeFilter.value === '') {
    selectedType.value = null
    selectedTypeMeta.value = null
  } else {
    selectedType.value = Number(posTypeFilter.value)
    const hit = typeOptions.value.find((x) => x.id === selectedType.value) || types.value.find((x) => x.id === selectedType.value)
    selectedTypeMeta.value = hit ? { id: hit.id, name: hit.name, code: hit.code } : selectedTypeMeta.value
  }
  pageP.value = 1
  void loadPos()
}

function onPosReset() {
  posNameOrCode.value = ''
  posTypeFilter.value = ''
  selectedType.value = null
  selectedTypeMeta.value = null
  pageP.value = 1
  void loadPos()
}

function onPosPageChange(p: number) {
  pageP.value = p
  void loadPos()
}

function onPosPageSizeChange(s: number) {
  limitP.value = s
  pageP.value = 1
  void loadPos()
}

function positionTypeName(id: number) {
  const hit = typeOptions.value.find((item) => item.id === id) || types.value.find((item) => item.id === id)
  return hit ? `${hit.name}（${hit.code}）` : `#${id}`
}

function openTypeDlg() {
  tEdit.value = null
  tForm.value = { name: '', code: '' }
  dlgT.value = true
}

function openTypeEdit(row: PositionTypeRow) {
  tEdit.value = row
  tForm.value = { name: row.name, code: row.code }
  dlgTEdit.value = true
}

async function saveType() {
  try {
    await createPositionType(tForm.value)
    dlgT.value = false
    ElMessage.success('已保存')
    await loadTypes()
    await loadPos()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '保存失败')
  }
}

async function saveTypeEdit() {
  if (!tEdit.value) return
  try {
    await updatePositionType(tEdit.value.id, { name: tForm.value.name, code: tForm.value.code })
    dlgTEdit.value = false
    ElMessage.success('已保存')
    await loadTypes()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '保存失败')
  }
}

async function removeType(row: PositionTypeRow) {
  await confirmArchiveAction({ name: row.name || row.code, title: '归档岗位类型', detail: '仅无岗位引用的类型可归档；历史配置仍会保留。' })
  try {
    await deletePositionType(row.id)
    types.value = types.value.filter((item) => item.id !== row.id)
    typeOptions.value = typeOptions.value.filter((item) => item.id !== row.id)
    totalT.value = Math.max(0, totalT.value - 1)
    if (types.value.length === 0 && pageT.value > 1) {
      pageT.value -= 1
    }
    if (selectedType.value === row.id) {
      selectedType.value = null
      selectedTypeMeta.value = null
      posTypeFilter.value = ''
      pageP.value = 1
    }
    ElMessage.success('已归档')
    await Promise.all([loadTypeOptions(), loadTypes(), loadPos()])
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '归档失败')
  }
}

async function openPosDlg() {
  await loadTypeOptions()
  pEdit.value = null
  const defaultTypeID = posTypeFilter.value === '' ? 0 : Number(posTypeFilter.value)
  pForm.value = { name: '', code: '', position_type_id: defaultTypeID }
  dlgP.value = true
}

async function openPosEdit(row: PositionRow) {
  await loadTypeOptions()
  pEdit.value = row
  pForm.value = { name: row.name, code: row.code, position_type_id: row.position_type_id }
  dlgPEdit.value = true
}

async function savePos() {
  if (!pForm.value.position_type_id) return
  const nextTypeId = pForm.value.position_type_id
  try {
    await createPosition({ position_type_id: nextTypeId, name: pForm.value.name, code: pForm.value.code })
    dlgP.value = false
    const nextType = typeOptions.value.find((x) => x.id === nextTypeId)
    if (nextType) selectedTypeMeta.value = { id: nextType.id, name: nextType.name, code: nextType.code }
    selectedType.value = nextTypeId
    posTypeFilter.value = nextTypeId
    pageP.value = 1
    ElMessage.success('已保存')
    await loadTypes()
    await loadPos()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '保存失败')
  }
}

async function savePosEdit() {
  if (!pEdit.value) return
  const nextTypeId = pForm.value.position_type_id
  try {
    await updatePosition(pEdit.value.id, {
      name: pForm.value.name,
      code: pForm.value.code,
      position_type_id: nextTypeId,
    })
    dlgPEdit.value = false
    const nextType = typeOptions.value.find((x) => x.id === nextTypeId)
    if (nextType) selectedTypeMeta.value = { id: nextType.id, name: nextType.name, code: nextType.code }
    selectedType.value = nextTypeId
    posTypeFilter.value = nextTypeId
    pageP.value = 1
    ElMessage.success('已保存')
    await loadTypes()
    await loadPos()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '保存失败')
  }
}

async function removePos(row: PositionRow) {
  await confirmArchiveAction({ name: row.name || row.code, title: '归档岗位' })
  try {
    await deletePosition(row.id)
    positions.value = positions.value.filter((item) => item.id !== row.id)
    totalP.value = Math.max(0, totalP.value - 1)
    if (positions.value.length === 0 && pageP.value > 1) {
      pageP.value -= 1
    }
    ElMessage.success('已归档')
    await Promise.all([loadPos(), loadTypes()])
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '归档失败')
  }
}

const typeColumns: TableColumn[] = [
  { key: 'name', title: '名称' },
  { key: 'code', title: '编码', width: 100 },
  { key: 'position_count', title: '岗位数', width: 80 },
  { key: 'actions', title: '操作', width: 160, fixed: 'right', tooltip: false },
]

const posColumns = computed<TableColumn[]>(() => [
  { key: 'name', title: '名称' },
  { key: 'code', title: '编码' },
  { key: 'position_type', title: '岗位类型', minWidth: 140 },
  { key: 'actions', title: '操作', width: 160, fixed: 'right', tooltip: false },
])

const positionSubtitle = computed(() => {
  if (!selectedTypeMeta.value) return '默认展示全部岗位；点击左侧岗位类型或使用筛选框可查看指定类型'
  return `当前筛选岗位类型：${selectedTypeMeta.value.name}（编码 ${selectedTypeMeta.value.code}）`
})

onMounted(async () => {
  await loadTypeOptions()
  await loadTypes()
  await loadPos()
})
</script>

<template>
  <div class="page position-page">
    <el-row :gutter="16">
      <el-col :xs="24" :lg="10" class="position-panel-col">
        <NeuroAgentListPage
          mode="el-table"
          title="岗位类型"
          subtitle="默认展示全部类型；点击一行筛选右侧岗位"
          :columns="typeColumns"
          :data="types"
          :total="totalT"
          :page="pageT"
          :page-size="limitT"
          :page-sizes="[10, 20, 50]"
          :show-create="true"
          :show-selection="false"
          :current-row-id="selectedType"
          :filter-fields="typeFilterFields"
          @create="openTypeDlg"
          @row-click="onTypeRowChange"
          @search="onTypeSearch"
          @page-change="onTypePageChange"
          @page-size-change="onTypePageSizeChange"
        >
          <template #actions>
            <el-button v-permission="'pos:create'" class="btn-gradient" @click="openTypeDlg">新增</el-button>
          </template>
          <template #col-actions="{ row }">
            <span class="op-btns">
              <el-button v-permission="'pos:edit'" size="small" @click.stop="openTypeEdit(row)">编辑</el-button>
              <el-button v-permission="'pos:delete'" type="danger" size="small" @click.stop="removeType(row)">删除</el-button>
            </span>
          </template>
        </NeuroAgentListPage>
      </el-col>
      <el-col :xs="24" :lg="14" class="position-panel-col">
        <NeuroAgentListPage
          mode="el-table"
          title="岗位"
          :subtitle="positionSubtitle"
          :columns="posColumns"
          :data="positions"
          :total="totalP"
          :page="pageP"
          :page-size="limitP"
          :page-sizes="[10, 20, 50]"
          :show-create="true"
          :show-selection="false"
          :filter-fields="posFilterFields"
          @create="openPosDlg"
          @reset="onPosReset"
          @search="onPosSearch"
          @page-change="onPosPageChange"
          @page-size-change="onPosPageSizeChange"
        >
          <template #actions>
            <el-button v-permission="'pos:create'" class="btn-gradient" @click="openPosDlg">新增</el-button>
          </template>
          <template #col-position_type="{ row }">
            <span>{{ row.position_type_name || positionTypeName(row.position_type_id) }}</span>
          </template>
          <template #col-actions="{ row }">
            <span class="op-btns">
              <el-button v-permission="'pos:edit'" @click="openPosEdit(row)">编辑</el-button>
              <el-button v-permission="'pos:delete'" type="danger" @click="removePos(row)">删除</el-button>
            </span>
          </template>
        </NeuroAgentListPage>
      </el-col>
    </el-row>

    <NeuroAgentDialog v-model="dlgT" title="新增类型" icon="📁" size="small">
      <div class="nm-form">
        <div class="nm-form-row">
          <div class="nm-form-item">
            <label class="nm-form-label">名称</label>
            <el-input v-model="tForm.name" />
          </div>
          <div class="nm-form-item">
            <label class="nm-form-label">编码</label>
            <el-input v-model="tForm.code" />
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
            <label class="nm-form-label">名称</label>
            <el-input v-model="tForm.name" />
          </div>
          <div class="nm-form-item">
            <label class="nm-form-label">编码</label>
            <el-input v-model="tForm.code" />
          </div>
        </div>
      </div>
      <template #footer-right>
        <button class="nm-btn nm-btn--primary" @click="saveTypeEdit">保存</button>
      </template>
    </NeuroAgentDialog>

    <NeuroAgentDialog v-model="dlgP" title="新增岗位" icon="💼" size="small">
      <div class="nm-form">
        <div class="nm-form-row">
          <div class="nm-form-item nm-form-item--full">
            <label class="nm-form-label">所属岗位类型</label>
            <el-select v-model="pForm.position_type_id" placeholder="选择类型" filterable style="width: 100%">
              <el-option
                v-for="t in typeOptions"
                :key="t.id"
                :label="`${t.name}（${t.code}）`"
                :value="t.id"
              />
            </el-select>
          </div>
          <div class="nm-form-item">
            <label class="nm-form-label">名称</label>
            <el-input v-model="pForm.name" placeholder="岗位名称" />
          </div>
          <div class="nm-form-item">
            <label class="nm-form-label">编码</label>
            <el-input v-model="pForm.code" placeholder="岗位编码" />
          </div>
        </div>
      </div>
      <template #footer-right>
        <button class="nm-btn nm-btn--primary" @click="savePos">保存</button>
      </template>
    </NeuroAgentDialog>

    <NeuroAgentDialog v-model="dlgPEdit" title="编辑岗位" icon="💼" size="small">
      <div class="nm-form">
        <div class="nm-form-row">
          <div class="nm-form-item nm-form-item--full">
            <label class="nm-form-label">所属岗位类型</label>
            <p class="nm-form-tip">修改后岗位会移动到新的岗位类型，保存后右侧列表将自动切换过去。</p>
            <el-select v-model="pForm.position_type_id" placeholder="选择类型" filterable style="width: 100%">
              <el-option
                v-for="t in typeOptions"
                :key="t.id"
                :label="`${t.name}（${t.code}）`"
                :value="t.id"
              />
            </el-select>
          </div>
          <div class="nm-form-item">
            <label class="nm-form-label">名称</label>
            <el-input v-model="pForm.name" />
          </div>
          <div class="nm-form-item">
            <label class="nm-form-label">编码</label>
            <el-input v-model="pForm.code" />
          </div>
        </div>
      </div>
      <template #footer-right>
        <button class="nm-btn nm-btn--primary" @click="savePosEdit">保存</button>
      </template>
    </NeuroAgentDialog>
  </div>
</template>

<style scoped>
.page {
  padding: 16px;
}

.position-page :deep(.position-panel-col),
.position-page :deep(.position-panel-col .list-card),
.position-page :deep(.position-panel-col .list-el-panel),
.position-page :deep(.position-panel-col .card-table),
.position-page :deep(.position-panel-col .el-table),
.position-page :deep(.position-panel-col .el-table__inner-wrapper),
.position-page :deep(.position-panel-col .el-table__body-wrapper),
.position-page :deep(.position-panel-col .el-table__fixed-body-wrapper),
.position-page :deep(.position-panel-col .el-table__fixed-right),
.position-page :deep(.position-panel-col .el-table__fixed-left),
.position-page :deep(.position-panel-col .el-scrollbar),
.position-page :deep(.position-panel-col .el-scrollbar__wrap),
.position-page :deep(.position-panel-col .el-scrollbar__view) {
  height: auto !important;
  max-height: none !important;
}

.position-page :deep(.position-panel-col .list-card),
.position-page :deep(.position-panel-col .list-el-panel),
.position-page :deep(.position-panel-col .card-table),
.position-page :deep(.position-panel-col .el-table__inner-wrapper),
.position-page :deep(.position-panel-col .el-table__body-wrapper),
.position-page :deep(.position-panel-col .el-table__fixed-body-wrapper),
.position-page :deep(.position-panel-col .el-scrollbar__wrap) {
  overflow-y: visible !important;
}

.position-page :deep(.position-panel-col .el-scrollbar__bar.is-vertical),
.position-page :deep(.position-panel-col .el-table__body-wrapper::-webkit-scrollbar),
.position-page :deep(.position-panel-col .el-scrollbar__wrap::-webkit-scrollbar) {
  display: none !important;
}

.position-page :deep(.position-panel-col .el-table__body-wrapper),
.position-page :deep(.position-panel-col .el-scrollbar__wrap) {
  scrollbar-width: none;
}

.nm-form-item--full {
  grid-column: 1 / -1;
}
</style>
