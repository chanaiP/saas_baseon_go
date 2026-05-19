<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'

import { deleteAiResource, updateAiResource } from '../api'
import type { AiResource } from '../types'
import { rowId, type AiRow } from './viewHelpers'
import AiJsonDialog from './AiJsonDialog.vue'

type FieldOption = { label: string; value: string | number; fill?: Record<string, unknown> }

const props = defineProps<{
  resource: AiResource
  row: AiRow
  options?: Record<string, FieldOption[]>
}>()

const emit = defineEmits<{
  saved: []
}>()

const editVisible = ref(false)

async function save(payload: AiRow) {
  await updateAiResource(props.resource, rowId(props.row), payload)
  editVisible.value = false
  ElMessage.success('已更新')
  emit('saved')
}

async function remove() {
  await ElMessageBox.confirm('确认删除该配置？删除会写入底座操作日志。', '删除确认', { type: 'warning' })
  await deleteAiResource(props.resource, rowId(props.row))
  ElMessage.success('已删除')
  emit('saved')
}
</script>

<template>
  <div class="ai-row-actions">
    <el-button link type="primary" @click.stop="editVisible = true">编辑</el-button>
    <el-button link type="danger" @click.stop="remove">删除</el-button>
    <AiJsonDialog
      v-model="editVisible"
      title="编辑配置"
      :sample="row"
      :options="options"
      @submit="save"
    />
  </div>
</template>
