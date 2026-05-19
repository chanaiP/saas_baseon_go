<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'

import NeuroAgentDialog from '@/views/components/NeuroAgentDialog.vue'

type FieldOption = { label: string; value: string | number; fill?: Record<string, unknown> }

const props = defineProps<{
  modelValue: boolean
  title: string
  tip?: string
  sample: Record<string, unknown> | Record<string, unknown>[]
  options?: Record<string, FieldOption[]>
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  submit: [value: Record<string, unknown>]
}>()

const text = ref('')
const mode = ref<'form' | 'json'>('form')
const formData = reactive<Record<string, unknown>>({})

const sampleObject = computed(() => (Array.isArray(props.sample) ? {} : props.sample))
const hiddenFormKeys = new Set(['id', 'createdat', 'updatedat', 'deletedat', 'created_at', 'updated_at', 'deleted_at'])
const fields = computed(() => Object.entries(sampleObject.value)
  .filter(([key]) => !hiddenFormKeys.has(key.toLowerCase()))
  .map(([key, value]) => ({
    key,
    label: fieldLabel(key),
    arrayObject: isArrayOfObjects(value),
    arrayColumns: arrayObjectColumns(value),
    options: fieldOptions(key),
    multiple: isMultipleField(key),
    complex: isComplex(value) && !fieldOptions(key).length && !isArrayOfObjects(value),
    inputType: typeof value === 'number' ? 'number' : 'text',
  })))

watch(
  () => props.modelValue,
  (visible) => {
    if (!visible) return
    mode.value = 'form'
    text.value = JSON.stringify(props.sample, null, 2)
    resetFormData()
  },
  { immediate: true },
)

function resetFormData() {
  for (const key of Object.keys(formData)) delete formData[key]
  for (const [key, value] of Object.entries(sampleObject.value)) {
    const options = fieldOptions(key)
    if (isArrayOfObjects(value)) formData[key] = cloneRows(value)
    else if (options.length && Array.isArray(value)) formData[key] = value
    else formData[key] = isComplex(value) ? JSON.stringify(value, null, 2) : value
  }
}

function fieldLabel(key: string) {
  const labels: Record<string, string> = {
    account_name: '账号名称',
    ai_scenario_code: 'AI 场景编码',
    ai_scenario_name: 'AI 场景名称',
    api_name: 'API 名称',
    api_path: 'API 路径',
    api_type: 'API 类型',
    app_code: '应用编码',
    app_name: '应用名称',
    auth_type: '鉴权方式',
    base_route_id: '基础路由 ID',
    base_url: '基础地址',
    capability_code: '所需能力',
    capabilities: '能力标签',
    code: '编码',
    context_window: '上下文窗口',
    description: '说明',
    default_base_route_id: '默认基础路由',
    default_for: '默认用途',
    endpoint: 'Endpoint',
    encrypted_api_key: 'API Key',
    latency_p95: 'P95 延迟',
    login_account: '登录账号',
    login_method: '登录方式',
    maintainer: '维护人',
    maintainer_contact: '维护人联系方式',
    max_retry: '失败重试',
    model_code: '模型 ID',
    model_name: '模型名称',
    model_type: '模型类型',
    monthly_budget: '月预算',
    name: '名称',
    owner: '负责人',
    override_base_route_id: '覆盖基础路由',
    policy_name: '策略名称',
    priority: '优先级',
    provider_id: '供应商 ID',
    price_policy_id: '价格策略',
    qps_limit: 'QPS 限制',
    region: '区域',
    route_code: '路由编码',
    route_name: '路由名称',
    scenario_type: '场景类型',
    status: '状态',
    strategy: '路由策略',
    success_rate: '成功率',
    tenant_scope: '租户范围',
    tenant_ids: '租户',
    timeout_ms: '超时时间',
    type: '类型',
    unit: '计费单位',
    usage_unit: '用量单位',
    version: '版本',
  }
  return labels[key] || key
}

function isMultipleField(key: string) {
  return key === 'capabilities' || key === 'default_for' || key === 'tenant_ids'
}

function isComplex(value: unknown) {
  return Array.isArray(value) || (value !== null && typeof value === 'object')
}

function isPlainObject(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === 'object' && !Array.isArray(value)
}

function isArrayOfObjects(value: unknown) {
  return Array.isArray(value) && value.length > 0 && value.every((item) => isPlainObject(item))
}

function cloneRows(value: unknown) {
  if (!Array.isArray(value)) return []
  return value.map((item) => ({ ...(isPlainObject(item) ? item : {}) }))
}

function arrayObjectColumns(value: unknown) {
  if (!isArrayOfObjects(value)) return []
  const rows = value as Record<string, unknown>[]
  const keys = new Set<string>()
  for (const item of rows) {
    for (const key of Object.keys(item)) keys.add(key)
  }
  const first = rows[0] ?? {}
  return Array.from(keys).map((key) => ({
    key,
    label: fieldLabel(key),
    options: fieldOptions(key),
    multiple: isMultipleField(key),
    complex: isComplex(first[key]) && !fieldOptions(key).length,
    inputType: typeof first[key] === 'number' ? 'number' : 'text',
  }))
}

function builtInOptions(key: string): FieldOption[] {
  const options: Record<string, FieldOption[]> = {
    api_type: [
      { label: '对话', value: 'chat' },
      { label: '图像', value: 'image' },
      { label: '视频', value: 'video' },
      { label: '向量', value: 'embedding' },
      { label: '重排', value: 'rerank' },
    ],
    auth_type: [
      { label: 'API Key', value: 'api_key' },
      { label: 'DashScope', value: 'dashscope' },
      { label: 'OAuth', value: 'oauth' },
      { label: '无鉴权', value: 'none' },
    ],
    capabilities: [
      { label: '对话生成', value: 'chat_completion' },
      { label: '文本生成', value: 'text_generation' },
      { label: '多模态', value: 'multimodal' },
      { label: '视觉理解', value: 'vision' },
      { label: '图像生成', value: 'image_generation' },
      { label: '视频生成', value: 'video_generation' },
      { label: '推理增强', value: 'reasoning' },
      { label: '长上下文', value: 'long_context' },
      { label: '向量生成', value: 'embedding' },
      { label: '重排', value: 'rerank' },
    ],
    capability_code: [
      { label: '对话生成', value: 'chat_completion' },
      { label: '文本生成', value: 'text_generation' },
      { label: '多模态', value: 'multimodal' },
      { label: '视觉理解', value: 'vision' },
      { label: '图像生成', value: 'image_generation' },
      { label: '视频生成', value: 'video_generation' },
      { label: '推理增强', value: 'reasoning' },
      { label: '长上下文', value: 'long_context' },
      { label: '向量生成', value: 'embedding' },
      { label: '重排', value: 'rerank' },
    ],
    login_method: [
      { label: '邮箱', value: 'email' },
      { label: '手机号', value: 'phone' },
      { label: '控制台', value: 'console' },
      { label: '未登记', value: 'none' },
    ],
    model_type: [
      { label: '文本模型', value: 'text' },
      { label: '多模态模型', value: 'multimodal' },
      { label: '图像模型', value: 'image' },
      { label: '视频模型', value: 'video' },
      { label: '向量模型', value: 'embedding' },
      { label: '重排模型', value: 'rerank' },
    ],
    scenario_type: [
      { label: '文本对话', value: 'chat' },
      { label: '文本生成', value: 'text' },
      { label: '图片生成', value: 'image' },
      { label: '视频生成', value: 'video' },
      { label: '多模态', value: 'multimodal' },
      { label: 'Agent 场景', value: 'agent' },
    ],
    status: [
      { label: '正常', value: 'active' },
      { label: '停用', value: 'inactive' },
      { label: '草稿', value: 'draft' },
    ],
    strategy: [
      { label: '优先级路由', value: 'priority' },
      { label: '故障降级', value: 'fallback' },
      { label: '权重分配', value: 'weighted' },
      { label: '轮询分配', value: 'round_robin' },
    ],
    tenant_scope: [
      { label: '指定租户', value: 'include' },
      { label: '排除租户', value: 'exclude' },
      { label: '全部租户', value: 'all' },
    ],
    type: [
      { label: '公有云', value: 'public_cloud' },
      { label: '私有化', value: 'private_cloud' },
      { label: '自建网关', value: 'self_hosted' },
    ],
    unit: [
      { label: 'tokens', value: 'tokens' },
      { label: '1K tokens', value: '1K tokens' },
      { label: '图片', value: 'image' },
      { label: '秒', value: 'seconds' },
      { label: '次', value: 'request' },
    ],
  }
  return options[key] || []
}

function fieldOptions(key: string) {
  return props.options?.[key] || builtInOptions(key)
}

function formValue(key: string) {
  const value = formData[key]
  return value === null || value === undefined ? '' : String(value)
}

function selectValue(key: string) {
  return formData[key]
}

function updateFormValue(key: string, value: unknown) {
  formData[key] = value
  const selected = fieldOptions(key).find((option) => option.value === value)
  if (selected?.fill) {
    for (const [fillKey, fillValue] of Object.entries(selected.fill)) {
      formData[fillKey] = fillValue
    }
  }
}

function arrayRows(key: string) {
  const value = formData[key]
  return Array.isArray(value) ? value as Record<string, unknown>[] : []
}

function rowInputValue(row: Record<string, unknown>, key: string) {
  const value = row[key]
  return value === null || value === undefined ? '' : String(value)
}

function updateArrayCell(parentKey: string, index: number, childKey: string, value: unknown) {
  const rows = arrayRows(parentKey)
  if (!rows[index]) return
  rows[index][childKey] = value
}

function addArrayRow(parentKey: string, columns: { key: string; inputType: string; multiple: boolean }[]) {
  const rows = arrayRows(parentKey)
  const next: Record<string, unknown> = {}
  for (const column of columns) next[column.key] = column.multiple ? [] : column.inputType === 'number' ? 0 : ''
  rows.push(next)
}

function removeArrayRow(parentKey: string, index: number) {
  arrayRows(parentKey).splice(index, 1)
}

function submitForm() {
  const payload: Record<string, unknown> = {}
  try {
    for (const field of fields.value) {
      const value = formData[field.key]
      if (field.arrayObject) {
        payload[field.key] = Array.isArray(value) ? value : []
      } else if (field.options.length) {
        payload[field.key] = field.multiple ? (Array.isArray(value) ? value : []) : value
      } else if (field.complex) {
        payload[field.key] = value ? JSON.parse(String(value)) : null
      } else if (field.inputType === 'number') {
        payload[field.key] = Number(value ?? 0)
      } else {
        payload[field.key] = value
      }
    }
    emit('submit', payload)
  } catch {
    ElMessage.error('表单中的数组或对象格式错误，请检查后再提交')
  }
}

function submit() {
  if (mode.value === 'form') {
    submitForm()
    return
  }
  try {
    emit('submit', JSON.parse(text.value) as Record<string, unknown>)
  } catch {
    ElMessage.error('JSON 格式错误，请检查后再提交')
  }
}
</script>

<template>
  <NeuroAgentDialog
    :model-value="modelValue"
    :title="title"
    icon="⚙️"
    width="760px"
    :show-confirm="false"
    cancel-text="取消"
    @update:model-value="emit('update:modelValue', $event)"
    @cancel="emit('update:modelValue', false)"
    @close="emit('update:modelValue', false)"
  >
    <p v-if="tip" class="ai-dialog-tip">{{ tip }}</p>
    <el-tabs v-model="mode" class="ai-dialog-mode-tabs">
      <el-tab-pane label="表单填写" name="form">
        <div class="nm-form ai-dialog-form">
          <div
            v-for="field in fields"
            :key="field.key"
            class="nm-form-item"
            :class="{ 'is-wide': field.complex || field.arrayObject }"
          >
            <label class="nm-form-label">{{ field.label }}</label>
            <div v-if="field.arrayObject" class="ai-array-form">
              <section v-for="(row, rowIndex) in arrayRows(field.key)" :key="rowIndex" class="ai-array-form__row">
                <header>
                  <strong>{{ field.label }} {{ rowIndex + 1 }}</strong>
                  <button type="button" @click="removeArrayRow(field.key, rowIndex)">删除</button>
                </header>
                <div class="ai-array-form__grid">
                  <div
                    v-for="column in field.arrayColumns"
                    :key="column.key"
                    class="nm-form-item"
                    :class="{ 'is-wide': column.complex }"
                  >
                    <label class="nm-form-label">{{ column.label }}</label>
                    <el-select
                      v-if="column.options.length"
                      :model-value="row[column.key]"
                      :multiple="column.multiple"
                      clearable
                      filterable
                      collapse-tags
                      collapse-tags-tooltip
                      placeholder="请选择"
                      @update:model-value="updateArrayCell(field.key, rowIndex, column.key, $event)"
                    >
                      <el-option
                        v-for="option in column.options"
                        :key="option.value"
                        :label="option.label"
                        :value="option.value"
                      />
                    </el-select>
                    <el-input
                      v-else-if="column.complex"
                      :model-value="rowInputValue(row, column.key)"
                      type="textarea"
                      :rows="4"
                      class="ai-json-editor"
                      @update:model-value="updateArrayCell(field.key, rowIndex, column.key, $event)"
                    />
                    <el-input
                      v-else
                      :model-value="rowInputValue(row, column.key)"
                      :type="column.inputType"
                      clearable
                      @update:model-value="updateArrayCell(field.key, rowIndex, column.key, $event)"
                    />
                  </div>
                </div>
              </section>
              <button class="ai-array-form__add" type="button" @click="addArrayRow(field.key, field.arrayColumns)">新增一行</button>
            </div>
            <el-select
              v-else-if="field.options.length"
              :model-value="selectValue(field.key)"
              :multiple="field.multiple"
              clearable
              filterable
              collapse-tags
              collapse-tags-tooltip
              placeholder="请选择"
              @update:model-value="updateFormValue(field.key, $event)"
            >
              <el-option
                v-for="option in field.options"
                :key="option.value"
                :label="option.label"
                :value="option.value"
              />
            </el-select>
            <el-input
              v-else-if="field.complex"
              :model-value="formValue(field.key)"
              type="textarea"
              :rows="5"
              class="ai-json-editor"
              @update:model-value="updateFormValue(field.key, $event)"
            />
            <el-input
              v-else
              :model-value="formValue(field.key)"
              :type="field.inputType"
              clearable
              @update:model-value="updateFormValue(field.key, $event)"
            />
          </div>
        </div>
      </el-tab-pane>
      <el-tab-pane label="JSON 高级模式" name="json">
        <div class="nm-form">
          <el-input v-model="text" type="textarea" :rows="16" class="ai-json-editor" />
        </div>
      </el-tab-pane>
    </el-tabs>
    <template #footer-right>
      <button class="nm-btn nm-btn--ghost" type="button" @click="emit('update:modelValue', false)">取消</button>
      <button class="nm-btn nm-btn--primary" type="button" @click="submit">提交</button>
    </template>
  </NeuroAgentDialog>
</template>
