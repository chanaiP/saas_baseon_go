<script setup lang="ts">
defineOptions({ name: 'PositionView' })
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
/** 岗位类型列表：名称或编码（与查询按钮联动） */
const typeKw = ref('')

const typeFilterFields: FilterField[] = [
  { key: 'typeHint', label: '名称 / 编码', type: 'text', placeholder: '模糊匹配' },
]

const posFilterFields: FilterField[] = [
  { key: 'nameOrCode', label: '名称 / 编码', type: 'text', placeholder: '模糊匹配' },
]

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
  if (!selectedType.value && res.items.length) {
    const first = res.items[0]
    selectedType.value = first.id
    selectedTypeMeta.value = { id: first.id, name: first.name, code: first.code }
  } else if (selectedType.value) {
    const hit = res.items.find((x) => x.id === selectedType.value)
    if (hit) {
      selectedTypeMeta.value = { id: hit.id, name: hit.name, code: hit.code }
    }
  }
}

async function loadPos() {
  if (!selectedType.value) {
    positions.value = []
    totalP.value = 0
    return
  }
  const skip = (pageP.value - 1) * limitP.value
  const res = await fetchPositions({
    position_type_id: selectedType.value,
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
  await createPositionType(tForm.value)
  dlgT.value = false
  await loadTypes()
  await loadPos()
}

async function saveTypeEdit() {
  if (!tEdit.value) return
  await updatePositionType(tEdit.value.id, { name: tForm.value.name, code: tForm.value.code })
  dlgTEdit.value = false
  await loadTypes()
}

async function removeType(row: PositionTypeRow) {
  await confirmArchiveAction({ name: row.name || row.code, title: '归档岗位类型', detail: '仅无岗位引用的类型可归档；历史配置仍会保留。' })
  await deletePositionType(row.id)
  if (selectedType.value === row.id) {
    selectedType.value = null
    selectedTypeMeta.value = null
  }
  await loadTypes()
  await loadPos()
}

async function openPosDlg() {
  if (!selectedType.value) return
  await loadTypeOptions()
  pEdit.value = null
  pForm.value = { name: '', code: '', position_type_id: selectedType.value }
  dlgP.value = true
}

async function openPosEdit(row: PositionRow) {
  await loadTypeOptions()
  pEdit.value = row
  pForm.value = { name: row.name, code: row.code, position_type_id: row.position_type_id }
  dlgPEdit.value = true
}

async function savePos() {
  if (!selectedType.value) return
  await createPosition({ position_type_id: selectedType.value, name: pForm.value.name, code: pForm.value.code })
  dlgP.value = false
  await loadPos()
  await loadTypes()
}

async function savePosEdit() {
  if (!pEdit.value) return
  await updatePosition(pEdit.value.id, {
    name: pForm.value.name,
    code: pForm.value.code,
    position_type_id: pForm.value.position_type_id,
  })
  dlgPEdit.value = false
  await loadPos()
  await loadTypes()
}

async function removePos(row: PositionRow) {
  await confirmArchiveAction({ name: row.name || row.code, title: '归档岗位' })
  await deletePosition(row.id)
  await loadPos()
  await loadTypes()
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
  { key: 'actions', title: '操作', width: 160, fixed: 'right', tooltip: false, hidden: !selectedType.value },
])

const positionSubtitle = computed(() => {
  if (!selectedTypeMeta.value) return '请先在左侧选择一个岗位类型，右侧将展示该类型下的岗位'
  return `当前岗位归属于：${selectedTypeMeta.value.name}（编码 ${selectedTypeMeta.value.code}）`
})

onMounted(async () => {
  await loadTypes()
  await loadPos()
})
</script>

<template>
  <div class="page">
    <el-row :gutter="16">
      <el-col :xs="24" :lg="10">
        <NeuroAgentListPage
          mode="el-table"
          title="岗位类型"
          subtitle="点击一行联动右侧岗位列表"
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
      <el-col :xs="24" :lg="14">
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
          @search="onPosSearch"
          @page-change="onPosPageChange"
          @page-size-change="onPosPageSizeChange"
        >
          <template #actions>
            <el-button v-permission="'pos:create'" class="btn-gradient" :disabled="!selectedType" @click="openPosDlg">新增</el-button>
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
            <el-input
              :model-value="
                selectedTypeMeta ? `${selectedTypeMeta.name}（${selectedTypeMeta.code}）` : '—'
              "
              disabled
            />
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

.nm-form-item--full {
  grid-column: 1 / -1;
}
</style>
