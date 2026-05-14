<script setup lang="ts">
import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'

const props = defineProps<{
  modelValue: boolean
  title: string
  tip?: string
  sample: Record<string, unknown> | Record<string, unknown>[]
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  submit: [value: Record<string, unknown>]
}>()

const text = ref('')

watch(
  () => props.modelValue,
  (visible) => {
    if (visible) text.value = JSON.stringify(props.sample, null, 2)
  },
  { immediate: true },
)

function submit() {
  try {
    emit('submit', JSON.parse(text.value) as Record<string, unknown>)
  } catch {
    ElMessage.error('JSON 格式错误，请检查后再提交')
  }
}
</script>

<template>
  <el-dialog
    :model-value="modelValue"
    :title="title"
    width="760px"
    @update:model-value="emit('update:modelValue', $event)"
  >
    <p v-if="tip" class="ai-dialog-tip">{{ tip }}</p>
    <el-input v-model="text" type="textarea" :rows="16" class="ai-json-editor" />
    <template #footer>
      <el-button @click="emit('update:modelValue', false)">取消</el-button>
      <el-button type="primary" @click="submit">提交</el-button>
    </template>
  </el-dialog>
</template>
