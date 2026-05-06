<script setup lang="ts">
defineOptions({ name: 'OrgUserPicker' })
import { User, OfficeBuilding } from '@element-plus/icons-vue'
import { computed, nextTick, ref, watch } from 'vue'
import type { OrgNode } from '@/api/organization'

export interface PickerUser {
  id: number
  name: string
  employee_no: string
  company_id: number | null
  department_id: number | null
}

const props = defineProps<{
  orgTree: OrgNode[]
  users: PickerUser[]
  modelValue: string[]
  bundleKey: string
}>()

const emit = defineEmits<{
  'update:modelValue': [keys: string[]]
}>()

/* ---------- 合并树：组织节点 + 用户叶子 ---------- */

interface TreeNode {
  key: string
  label: string
  isUser?: boolean
  children?: TreeNode[]
}

function buildMergedTree(orgNodes: OrgNode[], users: PickerUser[]): TreeNode[] {
  const byDept = new Map<number, PickerUser[]>()
  const byCompanyOnly = new Map<number, PickerUser[]>()
  for (const u of users) {
    if (u.department_id != null) {
      const arr = byDept.get(u.department_id) ?? []
      arr.push(u)
      byDept.set(u.department_id, arr)
    } else if (u.company_id != null) {
      const arr = byCompanyOnly.get(u.company_id) ?? []
      arr.push(u)
      byCompanyOnly.set(u.company_id, arr)
    }
  }

  function userNodes(list: PickerUser[]): TreeNode[] {
    return list.map((u) => ({
      key: `u_${u.id}`,
      label: `${u.name}（${u.employee_no}）`,
      isUser: true,
    }))
  }

  function convert(n: OrgNode): TreeNode {
    const key = `${n.node_type === 'company' ? 'c' : 'd'}_${n.id}`
    const orgChildren = n.children?.length ? n.children.map(convert) : []
    const uChildren =
      n.node_type === 'department'
        ? userNodes(byDept.get(n.id) ?? [])
        : userNodes(byCompanyOnly.get(n.id) ?? [])
    const children = [...orgChildren, ...uChildren]
    return { key, label: n.name, children: children.length ? children : undefined }
  }

  return orgNodes.map(convert)
}

const mergedTree = computed(() => buildMergedTree(props.orgTree, props.users))

/* ---------- 左侧树 ---------- */

const treeRef = ref<any>()

function emitCheckedKeys() {
  const keys = (treeRef.value?.getCheckedKeys(false) as string[]) ?? []
  emit('update:modelValue', keys)
}

function onTreeCheck() {
  emitCheckedKeys()
}

watch(
  [() => props.bundleKey, mergedTree],
  () => {
    nextTick(() => {
      treeRef.value?.setCheckedKeys(props.modelValue, false)
    })
  },
  { immediate: true },
)

/* ---------- 右侧：用户快捷面板 ---------- */

const highlightedNode = ref<TreeNode | null>(null)

function onNodeClick(data: TreeNode) {
  if (data.isUser) return
  highlightedNode.value = highlightedNode.value?.key === data.key ? null : data
}

function collectOrgIds(node: TreeNode): { companyIds: number[]; deptIds: number[] } {
  const cIds: number[] = []
  const dIds: number[] = []
  function walk(n: TreeNode) {
    if (n.isUser) return
    if (n.key.startsWith('c_')) cIds.push(Number(n.key.slice(2)))
    else if (n.key.startsWith('d_')) dIds.push(Number(n.key.slice(2)))
    n.children?.forEach(walk)
  }
  walk(node)
  return { companyIds: cIds, deptIds: dIds }
}

const searchText = ref('')
const currentPage = ref(1)
const pageSize = 15

watch([searchText, highlightedNode], () => { currentPage.value = 1 })

const filteredUsers = computed(() => {
  let list = props.users
  const node = highlightedNode.value
  if (node) {
    const { companyIds, deptIds } = collectOrgIds(node)
    const cSet = new Set(companyIds)
    const dSet = new Set(deptIds)
    list = list.filter(
      (u) =>
        (u.department_id != null && dSet.has(u.department_id)) ||
        (u.company_id != null && cSet.has(u.company_id)),
    )
  }
  if (searchText.value) {
    const q = searchText.value.toLowerCase()
    list = list.filter(
      (u) => u.name.toLowerCase().includes(q) || u.employee_no.toLowerCase().includes(q),
    )
  }
  return list
})

const pagedUsers = computed(() => {
  const start = (currentPage.value - 1) * pageSize
  return filteredUsers.value.slice(start, start + pageSize)
})

const checkedSet = computed(() => new Set(props.modelValue))

function isUserChecked(uid: number) {
  return checkedSet.value.has(`u_${uid}`)
}

function toggleUserFromPanel(uid: number) {
  const key = `u_${uid}`
  treeRef.value?.setChecked(key, !isUserChecked(uid), false)
  nextTick(emitCheckedKeys)
}

function togglePage(checked: boolean) {
  for (const u of filteredUsers.value) {
    if (checked !== isUserChecked(u.id)) {
      treeRef.value?.setChecked(`u_${u.id}`, checked, false)
    }
  }
  nextTick(emitCheckedKeys)
}

const pageAllChecked = computed(
  () =>
    filteredUsers.value.length > 0 && filteredUsers.value.every((u) => isUserChecked(u.id)),
)
const pageIndeterminate = computed(
  () => !pageAllChecked.value && filteredUsers.value.some((u) => isUserChecked(u.id)),
)

const selectedUserCount = computed(
  () => props.modelValue.filter((k) => k.startsWith('u_')).length,
)
const headerLabel = computed(() => highlightedNode.value?.label ?? '全部用户')
</script>

<template>
  <div class="org-user-picker">
    <!-- 左：组织 + 用户 合并树 -->
    <div class="picker-left">
      <div class="panel-hd">
        <el-icon :size="14"><OfficeBuilding /></el-icon>
        组织架构
      </div>
      <el-scrollbar max-height="300px">
        <el-tree
          ref="treeRef"
          :data="mergedTree"
          :props="{ children: 'children', label: 'label' }"
          node-key="key"
          show-checkbox
          default-expand-all
          highlight-current
          :expand-on-click-node="false"
          @check="onTreeCheck"
          @node-click="onNodeClick"
        >
          <template #default="{ data }">
            <span :class="data.isUser ? 'node-user' : 'node-org'">
              <el-icon :size="13" class="node-icon">
                <User v-if="data.isUser" />
                <OfficeBuilding v-else />
              </el-icon>
              {{ data.label }}
            </span>
          </template>
        </el-tree>
      </el-scrollbar>
    </div>

    <!-- 右：用户快捷面板 -->
    <div class="picker-right">
      <div class="panel-hd">
        <el-checkbox
          :model-value="pageAllChecked"
          :indeterminate="pageIndeterminate"
          @change="(v: boolean | string | number) => togglePage(!!v)"
        >{{ headerLabel }}</el-checkbox>
        <el-tag v-if="selectedUserCount" size="small" type="info" class="count-tag">
          已选 {{ selectedUserCount }} 人
        </el-tag>
      </div>

      <el-input
        v-model="searchText"
        placeholder="搜索姓名 / 工号"
        size="small"
        clearable
        class="search-input"
      />

      <el-scrollbar max-height="220px">
        <div v-for="u in pagedUsers" :key="u.id" class="user-row">
          <el-checkbox
            :model-value="isUserChecked(u.id)"
            @change="() => toggleUserFromPanel(u.id)"
          >{{ u.name }}（{{ u.employee_no }}）</el-checkbox>
        </div>
        <div v-if="!filteredUsers.length" class="picker-empty">
          {{ searchText ? '无匹配用户' : '该部门暂无用户' }}
        </div>
      </el-scrollbar>

      <el-pagination
        v-if="filteredUsers.length > pageSize"
        v-model:current-page="currentPage"
        :page-size="pageSize"
        :total="filteredUsers.length"
        layout="total, prev, pager, next"
        small
        class="user-pager"
      />
    </div>
  </div>
</template>

<style scoped>
.org-user-picker {
  display: flex;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 6px;
  overflow: hidden;
  width: 100%;
}
.picker-left {
  width: 46%;
  border-right: 1px solid var(--el-border-color-lighter);
  flex-shrink: 0;
}
.picker-right {
  flex: 1;
  min-width: 0;
  padding: 0 8px 8px;
}
.panel-hd {
  padding: 8px 10px;
  font-size: 13px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  background: var(--el-fill-color-lighter);
  display: flex;
  align-items: center;
  gap: 6px;
}
.count-tag { margin-left: auto; }
.search-input { margin: 8px 0 4px; }
.user-row { padding: 4px 2px; font-size: 13px; }
.user-row:hover { background: var(--el-fill-color-light); border-radius: 4px; }
.picker-empty {
  padding: 20px 0;
  text-align: center;
  color: var(--el-text-color-placeholder);
  font-size: 13px;
}

.user-pager {
  margin-top: 6px;
  justify-content: center;
}

/* 树节点样式 */
.node-org { font-weight: 500; }
.node-user { color: var(--el-text-color-regular); font-size: 13px; }
.node-icon { margin-right: 4px; vertical-align: -2px; }
</style>
