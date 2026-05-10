<script setup lang="ts">
defineOptions({ name: 'RoleView' })
import { ElMessage } from 'element-plus'
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import { createRole, deleteRole, fetchRoles, updateRole } from '@/api/role'
import type { RoleRow } from '@/api/role'
import { archiveSuccessMessage, confirmArchiveAction } from '@/composables/useArchiveConfirm'
import type { TableColumn } from '@/views/components/NeuroAgentListPage.vue'
import NeuroAgentDialog from '@/views/components/NeuroAgentDialog.vue'
import NeuroAgentListPage from '@/views/components/NeuroAgentListPage.vue'
const router = useRouter()

const loading = ref(false)
const items = ref<RoleRow[]>([])
const total = ref(0)
const page = ref(1)
const limit = ref(10)
const kw = ref('')

const dlg = ref(false)
const dlgCreate = ref(true)
const editRow = ref<RoleRow | null>(null)
const form = ref({ code: '', name: '', description: '' })

async function load() {
  loading.value = true
  try {
    const skip = (page.value - 1) * limit.value
    const res = await fetchRoles(skip, limit.value, kw.value || undefined)
    items.value = res.items
    total.value = res.total
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editRow.value = null
  dlgCreate.value = true
  form.value = { code: '', name: '', description: '' }
  dlg.value = true
}

function openEditInfo(row: RoleRow) {
  editRow.value = row
  dlgCreate.value = false
  form.value = { code: row.code, name: row.name, description: row.description || '' }
  dlg.value = true
}

function openPermissionConfig(row: RoleRow) {
  void router.push({ name: 'RolePermissionConfigView', params: { roleId: String(row.id) } })
}

async function saveDialog() {
  try {
    if (editRow.value) {
      await updateRole(editRow.value.id, {
        name: form.value.name,
        description: form.value.description,
      })
    } else {
      await createRole({ code: form.value.code, name: form.value.name, description: form.value.description })
    }
    dlg.value = false
    ElMessage.success('已保存')
    await load()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '保存失败')
  }
}

async function remove(row: RoleRow) {
  await confirmArchiveAction({ name: row.name || row.code, title: '归档角色' })
  try {
    await deleteRole(row.id)
    ElMessage.success(archiveSuccessMessage(row.name || row.code))
    await load()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '删除失败')
  }
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

function onSearch(v: { keyword: string; filters: Record<string, any> }) {
  kw.value = v.keyword
  page.value = 1
  load()
}

function onReset() {
  kw.value = ''
  page.value = 1
  load()
}

const columns: TableColumn[] = [
  { key: 'code', title: '编码', width: 140 },
  { key: 'name', title: '名称' },
  { key: 'description', title: '说明' },
  { key: 'actions', title: '操作', width: 260, fixed: 'right', tooltip: false },
]

onMounted(load)
</script>

<template>
  <div class="page">
    <NeuroAgentListPage
      mode="el-table"
      title="角色权限"
      :columns="columns"
      :data="items"
      :loading="loading"
      :total="total"
      :page="page"
      :page-size="limit"
      :page-sizes="[10, 20, 50]"
      :show-create="true"
      :show-selection="false"
      show-keyword-search
      search-placeholder="搜索角色名称、编码..."
      @create="openCreate"
      @search="onSearch"
      @reset="onReset"
      @page-change="onPageChange"
      @page-size-change="onPageSizeChange"
    >
      <template #actions>
        <el-button v-permission="'role:create'" class="btn-gradient" @click="openCreate">新增角色</el-button>
      </template>
      <template #col-actions="{ row }">
        <span class="op-btns">
          <el-button v-permission="'role:permission'" @click="openPermissionConfig(row)">权限</el-button>
          <el-button v-permission="'role:edit'" @click="openEditInfo(row)">编辑</el-button>
          <el-button v-permission="'role:delete'" type="danger" @click="remove(row)">删除</el-button>
        </span>
      </template>
    </NeuroAgentListPage>

    <NeuroAgentDialog v-model="dlg" :title="dlgCreate ? '新增角色' : '编辑角色信息'" icon="🛡️" size="medium">
      <div class="nm-form">
        <div class="nm-form-item">
          <label class="nm-form-label">编码</label>
          <el-input
            v-model="form.code"
            :disabled="!dlgCreate"
            placeholder="请输入角色编码（如 product_manager）"
          />
        </div>
        <div class="nm-form-item">
          <label class="nm-form-label">名称</label>
          <el-input v-model="form.name" placeholder="请输入角色名称（如 产品经理）" />
        </div>
        <div class="nm-form-item">
          <label class="nm-form-label">说明</label>
          <el-input v-model="form.description" placeholder="请输入角色说明（选填）" />
        </div>
      </div>
      <template #footer-right>
        <button class="nm-btn nm-btn--primary" @click="saveDialog">保存</button>
      </template>
    </NeuroAgentDialog>
  </div>
</template>

<style scoped>
.page {
  padding: 16px;
}
</style>
