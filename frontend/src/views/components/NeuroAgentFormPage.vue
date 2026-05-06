<template>
  <div class="neuro-agent-form-page" :data-theme="theme">
    <!-- 智能表单背景 -->
    <div class="form-intelligence-background">
      <div class="intelligence-grid"></div>
      <div class="intelligence-particles">
        <div v-for="n in 6" :key="n" class="intelligence-particle" :style="{
          left: `${Math.random() * 100}%`,
          top: `${Math.random() * 100}%`,
          animationDelay: `${n * 0.5}s`
        }"></div>
      </div>
    </div>

    <!-- 表单容器 -->
    <div class="form-container">
      <!-- 表单头部 -->
      <header class="form-header">
        <div class="header-left">
          <button class="back-button" @click="handleBack" v-if="showBackButton">
            <span class="back-icon">←</span>
            <span class="back-text">返回</span>
          </button>
        </div>

        <div class="header-center">
          <h1 class="form-title">
            <span class="title-icon">{{ config.icon || '📝' }}</span>
            <span class="title-text">{{ config.title || '智能表单' }}</span>
          </h1>
          <p class="form-subtitle" v-if="config.subtitle">{{ config.subtitle }}</p>
        </div>

        <div class="header-right">
          <div class="form-progress" v-if="showSteps">
            <span class="progress-text">步骤 {{ currentStep }} / {{ totalSteps }}</span>
            <div class="progress-bar">
              <div class="progress-fill" :style="{ width: stepProgress + '%' }"></div>
            </div>
          </div>
        </div>
      </header>

      <!-- 步骤导航 -->
      <div class="form-steps" v-if="showSteps && steps.length > 1">
        <div class="steps-track">
          <div
            v-for="(step, index) in steps"
            :key="step.id"
            class="step-item"
            :class="{
              'step-completed': index < currentStep - 1,
              'step-active': index === currentStep - 1,
              'step-pending': index > currentStep - 1
            }"
            @click="goToStep(index + 1)"
          >
            <div class="step-marker">
              <span class="marker-number" v-if="!step.completed && !step.active">{{ index + 1 }}</span>
              <span class="marker-icon" v-if="step.completed">✓</span>
            </div>
            <span class="step-label">{{ step.label }}</span>
            <div class="step-connection" v-if="index < steps.length - 1"></div>
          </div>
        </div>
      </div>

      <!-- 表单内容 -->
      <div class="form-content">
        <!-- AI助手提示 -->
        <div class="ai-form-assistant" v-if="showAiAssistant">
          <div class="assistant-header">
            <span class="assistant-icon">🤖</span>
            <span class="assistant-title">AI表单助手</span>

            <!-- AI控制面板 -->
            <div class="ai-controls">
              <button
                class="ai-control-btn"
                :class="{ active: aiAutoFillEnabled }"
                @click="toggleAIAutoFill"
                :title="aiAutoFillEnabled ? '禁用AI自动填充' : '启用AI自动填充'"
              >
                <span class="control-icon">{{ aiAutoFillEnabled ? '✅' : '❌' }}</span>
                <span class="control-text">自动填充</span>
              </button>

              <button
                class="ai-control-btn"
                @click="viewAIFillHistory"
                :disabled="!hasAIFillHistory"
                title="查看AI填充历史"
              >
                <span class="control-icon">📜</span>
                <span class="control-text">历史</span>
                <span class="control-badge" v-if="hasAIFillHistory">{{ aiFillHistory.length }}</span>
              </button>

              <button
                class="ai-control-btn"
                @click="clearAIFill()"
                :disabled="!hasAIFillHistory"
                title="清除所有AI填充"
              >
                <span class="control-icon">🗑️</span>
                <span class="control-text">清除</span>
              </button>

              <button class="assistant-toggle" @click="toggleAssistant">
                {{ showAssistantDetails ? '收起' : '展开' }}
              </button>
            </div>
          </div>

          <!-- AI状态指示器 -->
          <div class="ai-status-indicator" v-if="showAssistantDetails">
            <div class="status-item">
              <span class="status-label">AI填充率:</span>
              <div class="status-bar">
                <div class="status-fill" :style="{ width: `${aiFillPercentage}%` }"></div>
              </div>
              <span class="status-value">{{ aiFillPercentage }}%</span>
            </div>
            <div class="status-item">
              <span class="status-label">平均置信度:</span>
              <span class="status-value">{{ (averageAIConfidence * 100).toFixed(1) }}%</span>
            </div>
            <div class="status-item">
              <span class="status-label">AI填充字段:</span>
              <span class="status-value">{{ aiFilledCount }} / {{ props.fields.length }}</span>
            </div>
            <div class="status-item" v-if="isAIAnalyzing">
              <span class="status-label">分析状态:</span>
              <span class="status-value analyzing">分析中...</span>
            </div>
          </div>

          <div class="assistant-suggestions" v-if="showAssistantDetails && aiSuggestions.length > 0">
            <div
              v-for="suggestion in aiSuggestions"
              :key="suggestion.id"
              class="suggestion-item"
              @click="applySuggestion(suggestion)"
            >
              <div class="suggestion-header">
                <span class="suggestion-icon">{{ suggestion.icon }}</span>
                <span class="suggestion-text">{{ suggestion.text }}</span>
                <span class="suggestion-category">{{ suggestion.category }}</span>
              </div>
              <div class="suggestion-footer">
                <span class="suggestion-confidence">{{ (suggestion.confidence * 100).toFixed(0) }}%</span>
                <span class="suggestion-action">应用</span>
              </div>
            </div>
          </div>

          <!-- AI字段洞察 -->
          <div class="ai-field-insights" v-if="showAssistantDetails && aiInsightCount > 0">
            <div class="insights-header">
              <span class="insights-icon">💡</span>
              <span class="insights-title">AI字段洞察</span>
            </div>
            <div class="insights-list">
              <div
                v-for="(insights, fieldId) in aiFieldInsights"
                :key="fieldId"
                class="insight-item"
              >
                <span class="insight-field">{{ getFieldLabel(fieldId) }}:</span>
                <span class="insight-text">{{ insights[0] }}</span>
                <span class="insight-confidence">{{ (aiConfidenceScores[fieldId] * 100).toFixed(0) }}%</span>
                <button class="insight-clear" @click="clearAIFill(fieldId)" title="清除此字段的AI填充">
                  <span class="clear-icon">×</span>
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- 动态表单字段 -->
        <div class="form-fields">
          <div
            v-for="field in visibleFields"
            :key="field.id"
            class="form-field"
            :class="{
              'field-required': field.required,
              'field-error': fieldErrors[field.id],
              'field-focused': focusedField === field.id
            }"
          >
            <!-- 字段标签 -->
            <label class="field-label" :for="field.id">
              <span class="label-text">{{ field.label }}</span>
              <span class="label-required" v-if="field.required">*</span>
              <span class="label-ai" v-if="field.aiRecommended">
                <span class="ai-icon">🤖</span>
                <span class="ai-text">AI推荐</span>
              </span>
            </label>

            <!-- 字段描述 -->
            <p class="field-description" v-if="field.description">{{ field.description }}</p>

            <!-- 输入控件 -->
            <div class="field-control">
              <!-- 文本输入 -->
              <input
                v-if="field.type === 'text' || field.type === 'email' || field.type === 'number'"
                :type="field.type"
                :id="field.id"
                v-model="formData[field.id]"
                :placeholder="field.placeholder || `请输入${field.label}`"
                :required="field.required"
                :disabled="field.disabled"
                class="neuro-input"
                @focus="handleFieldFocus(field.id)"
                @blur="handleFieldBlur(field.id)"
                @input="handleFieldInput(field.id)"
              />

              <!-- 文本域 -->
              <textarea
                v-else-if="field.type === 'textarea'"
                :id="field.id"
                v-model="formData[field.id]"
                :placeholder="field.placeholder || `请输入${field.label}`"
                :required="field.required"
                :disabled="field.disabled"
                :rows="field.rows || 4"
                class="neuro-textarea"
                @focus="handleFieldFocus(field.id)"
                @blur="handleFieldBlur(field.id)"
                @input="handleFieldInput(field.id)"
              ></textarea>

              <!-- 选择器 -->
              <select
                v-else-if="field.type === 'select'"
                :id="field.id"
                v-model="formData[field.id]"
                :required="field.required"
                :disabled="field.disabled"
                class="neuro-select"
                @focus="handleFieldFocus(field.id)"
                @blur="handleFieldBlur(field.id)"
                @change="handleFieldChange(field.id)"
              >
                <option value="" disabled>{{ field.placeholder || `请选择${field.label}` }}</option>
                <option
                  v-for="option in field.options"
                  :key="option.value"
                  :value="option.value"
                >
                  {{ option.label }}
                </option>
              </select>

              <!-- 多选框组 -->
              <div v-else-if="field.type === 'checkbox-group'" class="checkbox-group">
                <label
                  v-for="option in field.options"
                  :key="option.value"
                  class="checkbox-option"
                >
                  <input
                    type="checkbox"
                    :value="option.value"
                    v-model="formData[field.id]"
                    :disabled="field.disabled"
                    class="neuro-checkbox"
                    @change="handleFieldChange(field.id)"
                  />
                  <span class="checkbox-label">{{ option.label }}</span>
                </label>
              </div>

              <!-- 单选框组 -->
              <div v-else-if="field.type === 'radio-group'" class="radio-group">
                <label
                  v-for="option in field.options"
                  :key="option.value"
                  class="radio-option"
                >
                  <input
                    type="radio"
                    :name="field.id"
                    :value="option.value"
                    v-model="formData[field.id]"
                    :disabled="field.disabled"
                    class="neuro-radio"
                    @change="handleFieldChange(field.id)"
                  />
                  <span class="radio-label">{{ option.label }}</span>
                </label>
              </div>

              <!-- 开关 -->
              <label v-else-if="field.type === 'switch'" class="switch-field">
                <input
                  type="checkbox"
                  :id="field.id"
                  v-model="formData[field.id]"
                  :disabled="field.disabled"
                  class="neuro-switch"
                  @change="handleFieldChange(field.id)"
                />
                <span class="switch-slider"></span>
                <span class="switch-label">{{ field.switchLabel || field.label }}</span>
              </label>

              <!-- 日期选择器 -->
              <div v-else-if="field.type === 'date'" class="date-field">
                <input
                  type="date"
                  :id="field.id"
                  v-model="formData[field.id]"
                  :required="field.required"
                  :disabled="field.disabled"
                  class="neuro-date"
                  @focus="handleFieldFocus(field.id)"
                  @blur="handleFieldBlur(field.id)"
                  @change="handleFieldChange(field.id)"
                />
                <button class="date-today" type="button" @click="setToday(field.id)">
                  今天
                </button>
              </div>

              <!-- 文件上传 -->
              <div v-else-if="field.type === 'file'" class="file-field">
                <input
                  type="file"
                  :id="field.id"
                  @change="handleFileUpload(field.id, $event)"
                  :required="field.required"
                  :disabled="field.disabled"
                  :accept="field.accept"
                  class="neuro-file"
                  hidden
                />
                <label :for="field.id" class="file-label">
                  <span class="file-icon">📎</span>
                  <span class="file-text">
                    {{ formData[field.id] ? formData[field.id].name : field.placeholder || '选择文件' }}
                  </span>
                  <span class="file-browse">浏览</span>
                </label>
                <div class="file-preview" v-if="formData[field.id]">
                  <span class="preview-name">{{ formData[field.id].name }}</span>
                  <span class="preview-size">{{ formatFileSize(formData[field.id].size) }}</span>
                  <button class="preview-remove" @click="removeFile(field.id)">×</button>
                </div>
              </div>

              <!-- 标签输入 -->
              <div v-else-if="field.type === 'tags'" class="tags-field">
                <div class="tags-input-wrapper">
                  <input
                    type="text"
                    :id="field.id"
                    v-model="tagInputs[field.id]"
                    :placeholder="field.placeholder || '输入标签后按Enter'"
                    class="neuro-tags-input"
                    @keydown.enter="addTag(field.id)"
                    @focus="handleFieldFocus(field.id)"
                    @blur="handleFieldBlur(field.id)"
                  />
                  <button class="tags-add" @click="addTag(field.id)">+</button>
                </div>
                <div class="tags-list" v-if="formData[field.id] && formData[field.id].length > 0">
                  <span
                    v-for="(tag, index) in formData[field.id]"
                    :key="index"
                    class="tag-item"
                  >
                    {{ tag }}
                    <button class="tag-remove" @click="removeTag(field.id, index as number)">×</button>
                  </span>
                </div>
              </div>
            </div>

            <!-- 字段帮助 -->
            <div class="field-help" v-if="field.helpText">
              <span class="help-icon">💡</span>
              <span class="help-text">{{ field.helpText }}</span>
            </div>

            <!-- AI填充标记 -->
            <div class="field-ai-fill-marker" v-if="aiFilledFields[field.id]">
              <span class="ai-fill-icon">🤖</span>
              <span class="ai-fill-text">AI填充</span>
              <span class="ai-fill-confidence" v-if="aiConfidenceScores[field.id]">
                {{ Math.round(aiConfidenceScores[field.id] * 100) }}% 置信度
              </span>
            </div>

            <!-- 字段错误 -->
            <div class="field-error-message" v-if="fieldErrors[field.id]">
              <span class="error-icon">⚠️</span>
              <span class="error-text">{{ fieldErrors[field.id] }}</span>
            </div>

            <!-- AI自动完成 -->
            <div class="field-autocomplete" v-if="showAutocomplete(field.id) && autocompleteSuggestions[field.id]">
              <div
                v-for="suggestion in autocompleteSuggestions[field.id]"
                :key="suggestion"
                class="autocomplete-item"
                @click="applyAutocomplete(field.id, suggestion)"
              >
                {{ suggestion }}
              </div>
            </div>
          </div>
        </div>

        <!-- 表单验证总结 -->
        <div class="form-validation-summary" v-if="Object.keys(fieldErrors).length > 0">
          <div class="validation-header">
            <span class="validation-icon">⚠️</span>
            <span class="validation-title">表单验证错误</span>
            <span class="validation-count">{{ Object.keys(fieldErrors).length }} 个错误</span>
          </div>
          <ul class="validation-errors">
            <li v-for="(error, fieldId) in fieldErrors" :key="fieldId">
              {{ getFieldLabel(fieldId) }}: {{ error }}
            </li>
          </ul>
        </div>

        <!-- 表单操作 -->
        <div class="form-actions">
          <div class="actions-left">
            <button
              class="neuro-btn neuro-btn-ghost"
              @click="handleReset"
              type="button"
              v-if="showResetButton"
            >
              <span class="btn-icon">🔄</span>
              <span class="btn-label">重置</span>
            </button>

            <button
              class="neuro-btn neuro-btn-secondary"
              @click="handleSaveDraft"
              type="button"
              v-if="showSaveDraft"
            >
              <span class="btn-icon">💾</span>
              <span class="btn-label">保存草稿</span>
            </button>
          </div>

          <div class="actions-center">
            <button
              class="neuro-btn neuro-btn-secondary"
              @click="goToPrevStep"
              type="button"
              v-if="showSteps && currentStep > 1"
            >
              <span class="btn-icon">←</span>
              <span class="btn-label">上一步</span>
            </button>
          </div>

          <div class="actions-right">
            <button
              class="neuro-btn neuro-btn-secondary"
              @click="handlePreview"
              type="button"
              v-if="showPreviewButton"
            >
              <span class="btn-icon">👁️</span>
              <span class="btn-label">预览</span>
            </button>

            <button
              class="neuro-btn neuro-btn-primary"
              @click="handleSubmit"
              type="button"
              :disabled="isSubmitting || !isFormValid"
            >
              <span class="btn-icon" v-if="isSubmitting">⏳</span>
              <span class="btn-icon" v-else>⚡</span>
              <span class="btn-label">
                {{ isSubmitting ? '提交中...' : submitButtonText }}
              </span>
            </button>
          </div>
        </div>
      </div>

      <!-- 表单底部信息 -->
      <footer class="form-footer" v-if="config.footerText">
        <div class="footer-content">
          <span class="footer-icon">ℹ️</span>
          <span class="footer-text">{{ config.footerText }}</span>
        </div>
      </footer>
    </div>

    <!-- 表单成功状态 -->
    <div class="form-success-overlay" v-if="showSuccess">
      <div class="success-content">
        <div class="success-icon">🎉</div>
        <h2 class="success-title">{{ successTitle }}</h2>
        <p class="success-message">{{ successMessage }}</p>
        <div class="success-actions">
          <button class="neuro-btn neuro-btn-primary" @click="handleSuccessAction('continue')">
            继续创建
          </button>
          <button class="neuro-btn neuro-btn-secondary" @click="handleSuccessAction('view')">
            查看详情
          </button>
          <button class="neuro-btn neuro-btn-ghost" @click="handleSuccessAction('close')">
            关闭
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, reactive, watch, onMounted } from 'vue'

// 组件属性
interface FormField {
  id: string
  label: string
  type: 'text' | 'email' | 'number' | 'textarea' | 'select' | 'checkbox-group' | 'radio-group' | 'switch' | 'date' | 'file' | 'tags'
  required?: boolean
  placeholder?: string
  description?: string
  helpText?: string
  defaultValue?: any
  options?: Array<{ value: string; label: string }>
  validation?: {
    pattern?: RegExp
    min?: number
    max?: number
    minLength?: number
    maxLength?: number
    custom?: (value: any) => string | null
  }
  aiRecommended?: boolean
  disabled?: boolean
  visible?: boolean
  dependsOn?: string
  dependsValue?: any
  rows?: number
  switchLabel?: string
  accept?: string
}

interface FormStep {
  id: string
  label: string
  fields: string[] // 字段ID数组
  completed?: boolean
  active?: boolean
}

interface FormConfig {
  title: string
  subtitle?: string
  icon?: string
  mode?: 'create' | 'edit' | 'view'
  showSteps?: boolean
  showAiAssistant?: boolean
  showBackButton?: boolean
  showResetButton?: boolean
  showSaveDraft?: boolean
  showPreviewButton?: boolean
  submitButtonText?: string
  footerText?: string
}

interface Props {
  config: FormConfig
  fields: FormField[]
  steps?: FormStep[]
  initialData?: Record<string, any>
  theme?: 'dark' | 'light'
}

const props = withDefaults(defineProps<Props>(), {
  config: () => ({
    title: '智能表单',
    mode: 'create',
    showSteps: true,
    showAiAssistant: true,
    showBackButton: true,
    showResetButton: true,
    showSaveDraft: true,
    showPreviewButton: true,
    submitButtonText: '提交'
  }),
  fields: () => [],
  steps: () => [],
  initialData: () => ({}),
  theme: 'dark'
})

// 响应式状态
const formData = reactive<Record<string, any>>({})
const fieldErrors = reactive<Record<string, string>>({})
const tagInputs = reactive<Record<string, string>>({})
const focusedField = ref<string | null>(null)
const currentStep = ref(1)
const isSubmitting = ref(false)
const showSuccess = ref(false)
const showAssistantDetails = ref(true)
const autocompleteSuggestions = reactive<Record<string, string[]>>({})

// AI增强状态
const aiAutoFillEnabled = ref(true)
const isAIAnalyzing = ref(false)
const aiFillHistory = ref<any[]>([])
const aiFieldInsights = reactive<Record<string, string[]>>({})
const aiFormPatterns = ref<string[]>([])
const aiConfidenceScores = reactive<Record<string, number>>({})
const aiFilledFields = reactive<Record<string, boolean>>({}) // 记录哪些字段被AI填充

// AI建议
const aiSuggestions = ref([
  { id: 1, icon: '💡', text: 'AI自动填充表单', category: '填充', confidence: 0.92 },
  { id: 2, icon: '📊', text: '基于历史数据智能推荐', category: '推荐', confidence: 0.88 },
  { id: 3, icon: '🔍', text: 'AI表单完整性检查', category: '验证', confidence: 0.95 },
  { id: 4, icon: '⚡', text: '智能字段顺序优化', category: '优化', confidence: 0.85 },
  { id: 5, icon: '🧠', text: 'AI预测表单模式', category: '预测', confidence: 0.90 },
  { id: 6, icon: '🎯', text: '个性化表单建议', category: '个性化', confidence: 0.87 }
])

// 成功状态
const successTitle = ref('提交成功！')
const successMessage = ref('表单已成功提交，系统正在处理您的请求。')

// 计算属性
const theme = computed(() => props.theme)
const showSteps = computed(() => props.config.showSteps && props.steps.length > 1)
const totalSteps = computed(() => props.steps.length)
const showAiAssistant = computed(() => props.config.showAiAssistant)
const showBackButton = computed(() => props.config.showBackButton)
const showResetButton = computed(() => props.config.showResetButton)
const showSaveDraft = computed(() => props.config.showSaveDraft)
const showPreviewButton = computed(() => props.config.showPreviewButton)
const submitButtonText = computed(() => props.config.submitButtonText || '提交')

const stepProgress = computed(() => {
  if (!showSteps.value) return 100
  return ((currentStep.value - 1) / (totalSteps.value - 1)) * 100
})

const visibleFields = computed(() => {
  if (showSteps.value && props.steps.length > 0) {
    const currentStepFields = props.steps[currentStep.value - 1]?.fields || []
    return props.fields.filter(field =>
      currentStepFields.includes(field.id) && isFieldVisible(field)
    )
  }
  return props.fields.filter(field => isFieldVisible(field))
})

const isFormValid = computed(() => {
  return Object.keys(fieldErrors).length === 0 &&
    visibleFields.value.every(field =>
      !field.required || (formData[field.id] !== undefined && formData[field.id] !== '' && formData[field.id] !== null)
    )
})

// AI增强计算属性
const aiEnabled = computed(() => props.config.showAiAssistant)
const hasAIFillHistory = computed(() => aiFillHistory.value.length > 0)
const aiInsightCount = computed(() => Object.keys(aiFieldInsights).length)
const averageAIConfidence = computed(() => {
  const scores = Object.values(aiConfidenceScores)
  if (scores.length === 0) return 0
  return scores.reduce((a, b) => a + b, 0) / scores.length
})
const aiFillPercentage = computed(() => {
  const filledByAI = Object.keys(aiConfidenceScores).length
  const totalFields = props.fields.length
  return totalFields > 0 ? Math.round((filledByAI / totalFields) * 100) : 0
})
const aiFilledCount = computed(() => Object.keys(aiFilledFields).length) // AI填充字段数量

// 方法
const initializeFormData = () => {
  props.fields.forEach(field => {
    if (field.defaultValue !== undefined) {
      formData[field.id] = field.defaultValue
    } else if (props.initialData[field.id] !== undefined) {
      formData[field.id] = props.initialData[field.id]
    } else {
      // 设置默认值
      switch (field.type) {
        case 'checkbox-group':
          formData[field.id] = []
          break
        case 'tags':
          formData[field.id] = []
          break
        case 'switch':
          formData[field.id] = false
          break
        default:
          formData[field.id] = ''
      }
    }
  })
}

const isFieldVisible = (field: FormField) => {
  if (field.visible === false) return false
  if (field.dependsOn) {
    const dependsValue = formData[field.dependsOn]
    return dependsValue === field.dependsValue
  }
  return true
}

const validateField = (fieldId: string) => {
  const field = props.fields.find(f => f.id === fieldId)
  if (!field) return

  const value = formData[fieldId]
  let error: string | null = null

  // 必填验证
  if (field.required && (value === undefined || value === '' || value === null || (Array.isArray(value) && value.length === 0))) {
    error = '此字段为必填项'
  }

  // 类型特定验证
  if (!error && value && field.validation) {
    const validation = field.validation

    if (validation.pattern && !validation.pattern.test(String(value))) {
      error = '格式不正确'
    }

    if (validation.min !== undefined && Number(value) < validation.min) {
      error = `最小值不能小于 ${validation.min}`
    }

    if (validation.max !== undefined && Number(value) > validation.max) {
      error = `最大值不能超过 ${validation.max}`
    }

    if (validation.minLength !== undefined && String(value).length < validation.minLength) {
      error = `长度不能少于 ${validation.minLength} 个字符`
    }

    if (validation.maxLength !== undefined && String(value).length > validation.maxLength) {
      error = `长度不能超过 ${validation.maxLength} 个字符`
    }

    if (validation.custom) {
      const customError = validation.custom(value)
      if (customError) error = customError
    }
  }

  if (error) {
    fieldErrors[fieldId] = error
  } else {
    delete fieldErrors[fieldId]
  }
}

const validateAllFields = () => {
  visibleFields.value.forEach(field => {
    validateField(field.id)
  })
}

const handleFieldFocus = (fieldId: string) => {
  focusedField.value = fieldId
  // 触发AI自动完成
  generateAutocompleteSuggestions(fieldId)
}

const handleFieldBlur = (fieldId: string) => {
  focusedField.value = null
  validateField(fieldId)
}

const handleFieldInput = (fieldId: string) => {
  // 实时验证
  validateField(fieldId)
  // 更新依赖字段的可见性
  updateDependentFields()
  // 如果用户手动修改字段，清除AI填充标记
  if (aiFilledFields[fieldId]) {
    delete aiFilledFields[fieldId]
  }
}

const handleFieldChange = (fieldId: string) => {
  validateField(fieldId)
  updateDependentFields()
  // 如果用户手动修改字段，清除AI填充标记
  if (aiFilledFields[fieldId]) {
    delete aiFilledFields[fieldId]
  }
}

const handleFileUpload = (fieldId: string, event: Event) => {
  const input = event.target as HTMLInputElement
  if (input.files && input.files[0]) {
    formData[fieldId] = input.files[0]
  }
  validateField(fieldId)
}

const removeFile = (fieldId: string) => {
  formData[fieldId] = null
  const input = document.getElementById(fieldId) as HTMLInputElement
  if (input) input.value = ''
  validateField(fieldId)
}

const addTag = (fieldId: string) => {
  const input = tagInputs[fieldId]?.trim()
  if (input) {
    if (!formData[fieldId]) formData[fieldId] = []
    if (!formData[fieldId].includes(input)) {
      formData[fieldId].push(input)
    }
    tagInputs[fieldId] = ''
    validateField(fieldId)
  }
}

const removeTag = (fieldId: string, index: number) => {
  formData[fieldId].splice(index, 1)
  validateField(fieldId)
}

const setToday = (fieldId: string) => {
  const today = new Date().toISOString().split('T')[0]
  formData[fieldId] = today
  validateField(fieldId)
}

const updateDependentFields = () => {
  // 重新计算字段可见性
  // Vue的响应式系统会自动处理
}

const generateAutocompleteSuggestions = (fieldId: string) => {
  // 模拟AI自动完成
  const field = props.fields.find(f => f.id === fieldId)
  if (!field) return

  const suggestions: string[] = []
  const value = String(formData[fieldId] || '').toLowerCase()

  if (field.type === 'text' || field.type === 'textarea') {
    if (field.label.includes('名称') || field.label.includes('name')) {
      suggestions.push('AI项目管理平台', 'DevOS 神经系统', '智能协作工具')
    } else if (field.label.includes('描述') || field.label.includes('description')) {
      suggestions.push('基于神经美学的AI协作平台', '未来感交互设计系统', '智能数据分析工具')
    }
  } else if (field.type === 'select') {
    // 选择器不需要自动完成
    return
  }

  // 过滤已输入的文本
  const filteredSuggestions = suggestions.filter(s =>
    s.toLowerCase().includes(value) && s.toLowerCase() !== value
  )

  if (filteredSuggestions.length > 0) {
    autocompleteSuggestions[fieldId] = filteredSuggestions.slice(0, 3)
  } else {
    delete autocompleteSuggestions[fieldId]
  }
}

const showAutocomplete = (fieldId: string) => {
  return focusedField.value === fieldId &&
    (props.fields.find(f => f.id === fieldId)?.type === 'text' ||
     props.fields.find(f => f.id === fieldId)?.type === 'textarea')
}

const applyAutocomplete = (fieldId: string, suggestion: string) => {
  formData[fieldId] = suggestion
  delete autocompleteSuggestions[fieldId]
  validateField(fieldId)

  // 记录AI填充历史
  if (aiAutoFillEnabled.value) {
    recordAIFill(fieldId, suggestion, 'autocomplete')
  }
}

const applySuggestion = (suggestion: any) => {
  console.log('应用AI建议:', suggestion)
  emit('ai-suggestion', suggestion)

  // 根据建议类型执行不同的AI操作
  switch (suggestion.id) {
    case 1: // AI自动填充表单
      executeAIAutoFill()
      break
    case 2: // 基于历史数据智能推荐
      generateHistoricalRecommendations()
      break
    case 3: // AI表单完整性检查
      performAIFormValidation()
      break
    case 4: // 智能字段顺序优化
      optimizeFieldOrder()
      break
    case 5: // AI预测表单模式
      predictFormPatterns()
      break
    case 6: // 个性化表单建议
      generatePersonalizedSuggestions()
      break
  }
}

// 执行AI自动填充
const executeAIAutoFill = () => {
  if (!aiAutoFillEnabled.value) {
    alert('AI自动填充功能未启用')
    return
  }

  isAIAnalyzing.value = true
  aiSuggestions.value.push({
    id: Date.now(),
    icon: '⏳',
    text: 'AI正在分析表单结构...',
    category: '分析',
    confidence: 0.95
  })

  // 模拟AI分析过程
  setTimeout(() => {
    const fieldsToFill = props.fields.filter(field => {
      const currentValue = formData[field.id]
      return !currentValue ||
             (typeof currentValue === 'string' && currentValue.trim() === '') ||
             (Array.isArray(currentValue) && currentValue.length === 0)
    })

    fieldsToFill.forEach(field => {
      const aiValue = generateAIFieldValue(field)
      if (aiValue !== null) {
        formData[field.id] = aiValue.value
        aiConfidenceScores[field.id] = aiValue.confidence
        aiFilledFields[field.id] = true // 标记字段为AI填充

        // 添加字段洞察
        if (aiValue.insight) {
          aiFieldInsights[field.id] = aiFieldInsights[field.id] || []
          aiFieldInsights[field.id].push(aiValue.insight)
        }

        // 记录填充历史
        recordAIFill(field.id, aiValue.value, 'auto-fill', aiValue.confidence)
      }
    })

    isAIAnalyzing.value = false
    aiSuggestions.value.push({
      id: Date.now(),
      icon: '✅',
      text: `AI自动填充完成 (${fieldsToFill.length}个字段)`,
      category: '完成',
      confidence: averageAIConfidence.value
    })

    validateAllFields()
  }, 1500)
}

// 生成AI字段值
const generateAIFieldValue = (field: FormField): { value: any, confidence: number, insight?: string } | null => {
  const fieldLabel = field.label.toLowerCase()
  const fieldType = field.type

  // 基于字段标签和类型的智能填充
  if (fieldLabel.includes('名称') || fieldLabel.includes('name')) {
    const names = [
      'DevOS 神经AI平台',
      'Cyber-Organic智能系统',
      '未来感项目管理工具',
      'AI协作神经中枢',
      '智能数据分析平台'
    ]
    const randomName = names[Math.floor(Math.random() * names.length)]
    return {
      value: randomName,
      confidence: 0.92,
      insight: '基于项目命名模式生成'
    }
  }

  if (fieldLabel.includes('描述') || fieldLabel.includes('description')) {
    const descriptions = [
      '基于Cyber-Organic Neural Interface的未来感项目管理工具，集成AI智能分析和神经美学设计',
      '采用DevOS 神经架构的智能协作平台，支持实时数据分析和预测性决策',
      '融合AI增强交互和神经连接技术的下一代项目管理解决方案',
      '具备自我学习和优化能力的智能系统，提供个性化项目建议和风险预测'
    ]
    const randomDesc = descriptions[Math.floor(Math.random() * descriptions.length)]
    return {
      value: randomDesc,
      confidence: 0.88,
      insight: '智能描述生成，包含技术关键词'
    }
  }

  if (fieldLabel.includes('类型') || fieldLabel.includes('type')) {
    if (field.options && field.options.length > 0) {
      const recommendedOption = field.options.find(opt =>
        opt.label.includes('AI') || opt.label.includes('智能')
      ) || field.options[0]
      return {
        value: recommendedOption.value,
        confidence: 0.85,
        insight: '基于项目类型趋势推荐'
      }
    }
  }

  if (fieldLabel.includes('优先级') || fieldLabel.includes('priority')) {
    return {
      value: '高',
      confidence: 0.90,
      insight: '新项目通常需要高优先级关注'
    }
  }

  if (fieldLabel.includes('标签') || fieldLabel.includes('tags')) {
    const tags = ['AI', '项目管理', 'DevOS 神经', '智能分析', '未来科技', '协作工具']
    const selectedTags = tags.slice(0, 3 + Math.floor(Math.random() * 3))
    return {
      value: selectedTags,
      confidence: 0.87,
      insight: '基于项目领域推荐标签'
    }
  }

  if (fieldType === 'date') {
    const today = new Date()
    const futureDate = new Date(today)
    futureDate.setDate(today.getDate() + 30) // 30天后
    return {
      value: futureDate.toISOString().split('T')[0],
      confidence: 0.80,
      insight: '设置合理的时间范围'
    }
  }

  if (fieldType === 'number') {
    if (fieldLabel.includes('预算') || fieldLabel.includes('budget')) {
      return {
        value: 50000 + Math.floor(Math.random() * 100000),
        confidence: 0.75,
        insight: '基于类似项目预算估算'
      }
    }
  }

  return null
}

// 记录AI填充历史
const recordAIFill = (fieldId: string, value: any, method: string, confidence?: number) => {
  const field = props.fields.find(f => f.id === fieldId)
  aiFillHistory.value.unshift({
    timestamp: Date.now(),
    fieldId,
    fieldLabel: field?.label || fieldId,
    value,
    method,
    confidence: confidence || 0.8
  })

  // 限制历史记录数量
  if (aiFillHistory.value.length > 20) {
    aiFillHistory.value = aiFillHistory.value.slice(0, 20)
  }
}

// 生成历史数据推荐
const generateHistoricalRecommendations = () => {
  isAIAnalyzing.value = true
  aiSuggestions.value.push({
    id: Date.now(),
    icon: '📊',
    text: '正在分析历史数据模式...',
    category: '分析',
    confidence: 0.90
  })

  setTimeout(() => {
    // 模拟历史数据分析
    const patterns = [
      '检测到项目名称通常包含"AI"或"智能"关键词',
      '历史数据显示80%的项目预算在50,000-150,000之间',
      '类似项目通常设置30-90天的完成期限',
      '高优先级项目更倾向于使用敏捷开发方法'
    ]

    aiFormPatterns.value = patterns

    // 生成基于历史的建议
    patterns.forEach(pattern => {
      aiSuggestions.value.push({
        id: Date.now() + Math.random(),
        icon: '🎯',
        text: pattern,
        category: '历史模式',
        confidence: 0.85
      })
    })

    isAIAnalyzing.value = false
  }, 1200)
}

// 执行AI表单验证
const performAIFormValidation = () => {
  isAIAnalyzing.value = true
  aiSuggestions.value.push({
    id: Date.now(),
    icon: '🔍',
    text: 'AI正在深度验证表单...',
    category: '验证',
    confidence: 0.95
  })

  setTimeout(() => {
    const validationResults = []

    // 检查必填字段
    const missingRequired = visibleFields.value.filter(field =>
      field.required && (!formData[field.id] ||
        (typeof formData[field.id] === 'string' && formData[field.id].trim() === '') ||
        (Array.isArray(formData[field.id]) && formData[field.id].length === 0))
    )

    if (missingRequired.length > 0) {
      validationResults.push(`发现 ${missingRequired.length} 个必填字段未填写`)
    }

    // 检查数据一致性
    const nameField = props.fields.find(f => f.label.includes('名称'))
    const descField = props.fields.find(f => f.label.includes('描述'))

    if (nameField && descField && formData[nameField.id] && formData[descField.id]) {
      const name = String(formData[nameField.id])
      const desc = String(formData[descField.id])

      if (desc.includes(name)) {
        validationResults.push('描述中包含了项目名称，建议增加更多细节')
      }
    }

    // 检查标签数量
    const tagsFields = props.fields.filter(f => f.type === 'tags')
    tagsFields.forEach(field => {
      if (formData[field.id] && Array.isArray(formData[field.id])) {
        const tagCount = formData[field.id].length
        if (tagCount < 2) {
          validationResults.push(`${field.label} 标签数量较少，建议添加更多关键词`)
        } else if (tagCount > 8) {
          validationResults.push(`${field.label} 标签数量过多，建议精简到5-8个`)
        }
      }
    })

    // 显示验证结果
    if (validationResults.length > 0) {
      validationResults.forEach(result => {
        aiSuggestions.value.push({
          id: Date.now() + Math.random(),
          icon: '⚠️',
          text: result,
          category: '验证建议',
          confidence: 0.88
        })
      })
    } else {
      aiSuggestions.value.push({
        id: Date.now(),
        icon: '✅',
        text: 'AI验证通过！表单数据完整且一致',
        category: '验证完成',
        confidence: 0.95
      })
    }

    isAIAnalyzing.value = false
  }, 1800)
}

// 优化字段顺序
const optimizeFieldOrder = () => {
  aiSuggestions.value.push({
    id: Date.now(),
    icon: '⚡',
    text: '正在优化字段顺序...',
    category: '优化',
    confidence: 0.90
  })

  setTimeout(() => {
    // 模拟字段顺序优化建议
    const optimizationTips = [
      '建议将"项目名称"和"项目描述"放在表单开头',
      '相关字段（如预算和时间）应该分组显示',
      '复杂字段（如文件上传）建议放在表单末尾',
      '必填字段应该用更明显的视觉标识'
    ]

    optimizationTips.forEach(tip => {
      aiSuggestions.value.push({
        id: Date.now() + Math.random(),
        icon: '💡',
        text: tip,
        category: '优化建议',
        confidence: 0.85
      })
    })
  }, 1000)
}

// 预测表单模式
const predictFormPatterns = () => {
  isAIAnalyzing.value = true
  aiSuggestions.value.push({
    id: Date.now(),
    icon: '🧠',
    text: 'AI正在预测表单模式...',
    category: '预测',
    confidence: 0.92
  })

  setTimeout(() => {
    const predictions = [
      '基于当前数据，预测项目完成时间可能需要45-60天',
      '检测到技术类项目模式，建议增加技术栈字段',
      '预测预算使用率可能在70-85%之间',
      '类似项目通常需要3-5人的团队规模'
    ]

    predictions.forEach(prediction => {
      aiSuggestions.value.push({
        id: Date.now() + Math.random(),
        icon: '🔮',
        text: prediction,
        category: 'AI预测',
        confidence: 0.80
      })
    })

    isAIAnalyzing.value = false
  }, 1500)
}

// 生成个性化建议
const generatePersonalizedSuggestions = () => {
  aiSuggestions.value.push({
    id: Date.now(),
    icon: '🎯',
    text: '正在生成个性化建议...',
    category: '个性化',
    confidence: 0.87
  })

  setTimeout(() => {
    const personalizedTips = [
      '基于您的历史项目，建议增加风险管理字段',
      '检测到您偏好使用敏捷方法，建议相关字段',
      '根据团队规模，建议调整项目时间线',
      '基于行业趋势，建议关注AI集成功能'
    ]

    personalizedTips.forEach(tip => {
      aiSuggestions.value.push({
        id: Date.now() + Math.random(),
        icon: '🌟',
        text: tip,
        category: '个性化建议',
        confidence: 0.75
      })
    })
  }, 1200)
}

// 切换AI自动填充
const toggleAIAutoFill = () => {
  aiAutoFillEnabled.value = !aiAutoFillEnabled.value
  aiSuggestions.value.push({
    id: Date.now(),
    icon: aiAutoFillEnabled.value ? '✅' : '❌',
    text: `AI自动填充 ${aiAutoFillEnabled.value ? '已启用' : '已禁用'}`,
    category: '设置',
    confidence: 1.0
  })
}

// 查看AI填充历史
const viewAIFillHistory = () => {
  if (aiFillHistory.value.length === 0) {
    aiSuggestions.value.push({
      id: Date.now(),
      icon: '📜',
      text: '暂无AI填充历史',
      category: '历史',
      confidence: 1.0
    })
    return
  }

  const recentHistory = aiFillHistory.value.slice(0, 5)
  recentHistory.forEach(record => {
    aiSuggestions.value.push({
      id: record.timestamp,
      icon: '🕒',
      text: `${record.fieldLabel}: ${record.value} (${(record.confidence * 100).toFixed(0)}%)`,
      category: '历史记录',
      confidence: record.confidence
    })
  })
}

// 清除AI填充
const clearAIFill = (fieldId?: string) => {
  if (fieldId) {
    // 清除特定字段的AI填充
    delete aiConfidenceScores[fieldId]
    delete aiFieldInsights[fieldId]
    delete aiFilledFields[fieldId] // 清除AI填充标记
    aiFillHistory.value = aiFillHistory.value.filter(record => record.fieldId !== fieldId)

    aiSuggestions.value.push({
      id: Date.now(),
      icon: '🗑️',
      text: `已清除 ${getFieldLabel(fieldId)} 的AI填充`,
      category: '清除',
      confidence: 1.0
    })
  } else {
    // 清除所有AI填充
    Object.keys(aiConfidenceScores).forEach(key => delete aiConfidenceScores[key])
    Object.keys(aiFieldInsights).forEach(key => delete aiFieldInsights[key])
    Object.keys(aiFilledFields).forEach(key => delete aiFilledFields[key]) // 清除所有AI填充标记
    aiFillHistory.value = []

    aiSuggestions.value.push({
      id: Date.now(),
      icon: '🗑️',
      text: '已清除所有AI填充数据',
      category: '清除',
      confidence: 1.0
    })
  }
}

const toggleAssistant = () => {
  showAssistantDetails.value = !showAssistantDetails.value
}

const goToStep = (step: number) => {
  if (step >= 1 && step <= totalSteps.value) {
    // 验证当前步骤
    validateAllFields()
    if (Object.keys(fieldErrors).length === 0) {
      currentStep.value = step
    }
  }
}

const goToPrevStep = () => {
  if (currentStep.value > 1) {
    currentStep.value--
  }
}

const goToNextStep = () => {
  if (currentStep.value < totalSteps.value) {
    // 验证当前步骤
    validateAllFields()
    if (Object.keys(fieldErrors).length === 0) {
      currentStep.value++
    }
  }
}

const handleBack = () => {
  emit('back')
}

const handleReset = () => {
  if (confirm('确定要重置表单吗？所有输入的数据将会丢失。')) {
    Object.keys(formData).forEach(key => {
      const field = props.fields.find(f => f.id === key)
      if (field && field.defaultValue !== undefined) {
        formData[key] = field.defaultValue
      } else {
        delete formData[key]
      }
    })
    Object.keys(fieldErrors).forEach(key => delete fieldErrors[key])
    Object.keys(tagInputs).forEach(key => tagInputs[key] = '')
    currentStep.value = 1
    emit('reset')
  }
}

const handleSaveDraft = () => {
  console.log('保存草稿:', formData)
  emit('save-draft', formData)
  alert('草稿已保存')
}

const handlePreview = () => {
  console.log('预览表单:', formData)
  emit('preview', formData)
}

const handleSubmit = async () => {
  // 验证所有字段
  validateAllFields()

  if (Object.keys(fieldErrors).length > 0) {
    alert('请先修正表单错误')
    return
  }

  if (!isFormValid.value) {
    alert('请填写所有必填字段')
    return
  }

  isSubmitting.value = true

  try {
    console.log('提交表单数据:', formData)
    // 模拟API调用
    await new Promise(resolve => setTimeout(resolve, 1500))

    // 提交成功
    showSuccess.value = true
    successTitle.value = '提交成功！'
    successMessage.value = '您的表单已成功提交，系统正在处理您的请求。'

    emit('submit', formData)
  } catch (error) {
    console.error('提交失败:', error)
    alert('提交失败，请稍后重试')
  } finally {
    isSubmitting.value = false
  }
}

const handleSuccessAction = (action: string) => {
  switch (action) {
    case 'continue':
      handleReset()
      showSuccess.value = false
      break
    case 'view':
      emit('success-view', formData)
      break
    case 'close':
      showSuccess.value = false
      emit('success-close')
      break
  }
}

const formatFileSize = (bytes: number) => {
  if (bytes === 0) return '0 Bytes'
  const k = 1024
  const sizes = ['Bytes', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const getFieldLabel = (fieldId: string) => {
  const field = props.fields.find(f => f.id === fieldId)
  return field?.label || fieldId
}

// 事件
const emit = defineEmits<{
  back: []
  reset: []
  'save-draft': [data: any]
  preview: [data: any]
  submit: [data: any]
  'ai-suggestion': [suggestion: any]
  'success-view': [data: any]
  'success-close': []
  'field-change': [fieldId: string, value: any]
}>()

// 导出类型
export type { FormField, FormStep, FormConfig }

// 初始化
onMounted(() => {
  initializeFormData()

  // 监听表单数据变化
  watch(formData, () => {
    validateAllFields()
  }, { deep: true })

  // 监听步骤变化
  watch(currentStep, () => {
    validateAllFields()
  })
})
</script>

<style scoped>
@import '@/styles/theme/neuro-theme.css';
/* 神经智能表单页样式 */
.neuro-agent-form-page {
  position: relative;
  min-height: 100vh;
  font-family: 'IBM Plex Mono', monospace;
}

/* 智能表单背景 */
.form-intelligence-background {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: -1;
  overflow: hidden;
}

.intelligence-grid {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background:
    linear-gradient(90deg, rgba(0, 245, 212, 0.05) 1px, transparent 1px) 0 0 / 40px 40px,
    linear-gradient(rgba(0, 245, 212, 0.05) 1px, transparent 1px) 0 0 / 40px 40px;
  animation: grid-move 20s linear infinite;
}

@keyframes grid-move {
  0% { transform: translate(0, 0); }
  100% { transform: translate(40px, 40px); }
}

.intelligence-particles {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
}

.intelligence-particle {
  position: absolute;
  width: 2px;
  height: 2px;
  background: var(--neuro-primary);
  border-radius: 50%;
  filter: blur(1px);
  animation: particle-float 8s infinite ease-in-out;
}

@keyframes particle-float {
  0%, 100% {
    transform: translateY(0) scale(1);
    opacity: 0.3;
  }
  50% {
    transform: translateY(-20px) scale(1.2);
    opacity: 0.8;
  }
}

/* 表单容器 */
.form-container {
  position: relative;
  max-width: 1200px;
  margin: 0 auto;
  padding: 40px 24px;
  z-index: 1;
}

/* 表单头部 */
.form-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 32px;
  padding: var(--neuro-spacing-lg);
  background: rgba(15, 23, 42, 0.8);
  backdrop-filter: blur(20px);
  border: 1px solid var(--neuro-primary-20);
  border-radius: 20px;
}

.header-left,
.header-center,
.header-right {
  flex: 1;
}

.header-center {
  text-align: center;
}

.header-right {
  display: flex;
  justify-content: flex-end;
}

.back-button {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 16px;
  background: var(--neuro-primary-10);
  border: 1px solid var(--neuro-primary-30);
  border-radius: var(--neuro-radius-lg);
  color: var(--neuro-primary);
  font-family: 'IBM Plex Mono', monospace;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.back-button:hover {
  background: var(--neuro-primary-20);
  transform: translateX(-4px);
}

.back-icon {
  font-size: 16px;
}

.form-title {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  margin: 0 0 8px 0;
  font-family: 'Orbitron', monospace;
  font-size: 28px;
  font-weight: 700;
  background: linear-gradient(135deg, var(--neuro-primary), #9d4edd);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.title-icon {
  font-size: 32px;
}

.form-subtitle {
  margin: 0;
  color: var(--neuro-text-secondary);
  font-size: 14px;
  text-align: center;
}

.form-progress {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 8px;
  min-width: 200px;
}

.progress-text {
  color: var(--neuro-primary);
  font-size: 12px;
  font-weight: 500;
}

.progress-bar {
  width: 100%;
  height: 6px;
  background: var(--neuro-primary-10);
  border-radius: 3px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: linear-gradient(90deg, var(--neuro-primary), #9d4edd);
  border-radius: 3px;
  transition: width 0.3s ease;
}

/* 步骤导航 */
.form-steps {
  margin-bottom: 32px;
  padding: var(--neuro-spacing-lg);
  background: rgba(15, 23, 42, 0.8);
  backdrop-filter: blur(20px);
  border: 1px solid var(--neuro-primary-20);
  border-radius: 20px;
}

.steps-track {
  display: flex;
  justify-content: space-between;
  position: relative;
}

.step-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  position: relative;
  z-index: 2;
  cursor: pointer;
  flex: 1;
}

.step-marker {
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(15, 23, 42, 0.9);
  border: 2px solid rgba(100, 116, 139, 0.5);
  border-radius: 50%;
  margin-bottom: 12px;
  transition: all var(--neuro-transition-normal);
}

.step-completed .step-marker {
  background: var(--neuro-primary-10);
  border-color: var(--neuro-primary);
}

.step-active .step-marker {
  background: var(--neuro-primary-20);
  border-color: var(--neuro-primary);
  box-shadow: 0 0 20px var(--neuro-primary-30);
}

.marker-number,
.marker-icon {
  color: var(--neuro-text-secondary);
  font-size: 14px;
  font-weight: 600;
}

.step-completed .marker-icon {
  color: var(--neuro-primary);
}

.step-active .marker-number {
  color: var(--neuro-primary);
}

.step-label {
  color: var(--neuro-text-secondary);
  font-size: 12px;
  font-weight: 500;
  text-align: center;
  transition: color 0.3s ease;
}

.step-completed .step-label,
.step-active .step-label {
  color: var(--neuro-primary);
}

.step-connection {
  position: absolute;
  top: 20px;
  left: 50%;
  right: -50%;
  height: 2px;
  background: rgba(100, 116, 139, 0.3);
  z-index: 1;
}

.step-completed .step-connection {
  background: linear-gradient(90deg, var(--neuro-primary), var(--neuro-primary-30));
}

/* 表单内容 */
.form-content {
  background: rgba(15, 23, 42, 0.8);
  backdrop-filter: blur(20px);
  border: 1px solid var(--neuro-primary-20);
  border-radius: 20px;
  padding: var(--neuro-spacing-xl);
}

/* AI助手提示 */
.ai-form-assistant {
  margin-bottom: 32px;
  padding: 20px;
  background: rgba(0, 245, 212, 0.05);
  border: 1px solid var(--neuro-primary-20);
  border-radius: var(--neuro-radius-xl);
}

.assistant-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}

.assistant-icon {
  font-size: 20px;
}

.assistant-title {
  color: var(--neuro-primary);
  font-family: 'Orbitron', monospace;
  font-size: 16px;
  font-weight: 600;
  flex: 1;
}

.assistant-toggle {
  padding: 6px 12px;
  background: var(--neuro-primary-10);
  border: 1px solid var(--neuro-primary-30);
  border-radius: var(--neuro-radius-md);
  color: var(--neuro-primary);
  font-family: 'IBM Plex Mono', monospace;
  font-size: 11px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.assistant-toggle:hover {
  background: var(--neuro-primary-20);
}

.assistant-suggestions {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 12px;
}

.suggestion-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  background: rgba(15, 23, 42, 0.8);
  border: 1px solid var(--neuro-primary-10);
  border-radius: var(--neuro-radius-lg);
  cursor: pointer;
  transition: all 0.2s ease;
}

.suggestion-item:hover {
  background: var(--neuro-primary-10);
  border-color: var(--neuro-primary-30);
  transform: translateY(-2px);
}

.suggestion-icon {
  font-size: 16px;
}

.suggestion-text {
  flex: 1;
  color: var(--neuro-text-secondary);
  font-size: 13px;
}

.suggestion-action {
  color: var(--neuro-primary);
  font-size: 11px;
  font-weight: 600;
  opacity: 0;
  transition: opacity 0.2s ease;
}

.suggestion-item:hover .suggestion-action {
  opacity: 1;
}

/* 表单字段 */
.form-fields {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 24px;
  margin-bottom: 32px;
}

.form-field {
  position: relative;
}

.field-label {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
  color: #e2e8f0;
  font-size: 14px;
  font-weight: 500;
}

.label-text {
  flex: 1;
}

.label-required {
  color: #f87171;
  font-size: 16px;
}

.label-ai {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 2px 6px;
  background: var(--neuro-primary-10);
  border: 1px solid var(--neuro-primary-30);
  border-radius: var(--neuro-radius-md);
  color: var(--neuro-primary);
  font-size: 10px;
}

.ai-icon {
  font-size: 10px;
}

.field-description {
  margin: 0 0 12px 0;
  color: var(--neuro-text-secondary);
  font-size: 12px;
  line-height: 1.4;
}

/* 输入控件通用样式 */
.neuro-input,
.neuro-textarea,
.neuro-select,
.neuro-date {
  width: 100%;
  padding: 12px 16px;
  background: rgba(15, 23, 42, 0.9);
  border: 1px solid rgba(100, 116, 139, 0.5);
  border-radius: var(--neuro-radius-lg);
  color: #e2e8f0;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 13px;
  transition: all 0.2s ease;
}

.neuro-input:focus,
.neuro-textarea:focus,
.neuro-select:focus,
.neuro-date:focus {
  outline: none;
  border-color: var(--neuro-primary);
  box-shadow: 0 0 0 3px var(--neuro-primary-10);
}

.neuro-input::placeholder,
.neuro-textarea::placeholder {
  color: #64748b;
}

.neuro-textarea {
  resize: vertical;
  min-height: 80px;
}

.neuro-select {
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='16' height='16' viewBox='0 0 24 24' fill='none' stroke='%2300f5d4' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpolyline points='6 9 12 15 18 9'%3E%3C/polyline%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 12px center;
  background-size: 16px;
  padding-right: 40px;
}

/* 字段聚焦状态 */
.field-focused .neuro-input,
.field-focused .neuro-textarea,
.field-focused .neuro-select,
.field-focused .neuro-date {
  border-color: var(--neuro-primary);
  background: rgba(0, 245, 212, 0.05);
}

/* 字段错误状态 */
.field-error .neuro-input,
.field-error .neuro-textarea,
.field-error .neuro-select,
.field-error .neuro-date {
  border-color: #f87171;
  background: rgba(248, 113, 113, 0.05);
}

.field-error .field-label {
  color: #f87171;
}

/* 多选框组和单选框组 */
.checkbox-group,
.radio-group {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.checkbox-option,
.radio-option {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
}

.neuro-checkbox,
.neuro-radio {
  width: 18px;
  height: 18px;
  accent-color: var(--neuro-primary);
  cursor: pointer;
}

.checkbox-label,
.radio-label {
  color: var(--neuro-text-secondary);
  font-size: 13px;
}

/* 开关 */
.switch-field {
  display: flex;
  align-items: center;
  gap: 12px;
  cursor: pointer;
}

.neuro-switch {
  position: relative;
  width: 44px;
  height: 24px;
  appearance: none;
  background: rgba(100, 116, 139, 0.5);
  border-radius: var(--neuro-radius-lg);
  cursor: pointer;
  transition: all var(--neuro-transition-normal);
}

.neuro-switch:checked {
  background: var(--neuro-primary);
}

.switch-slider {
  position: absolute;
  top: 2px;
  left: 2px;
  width: 20px;
  height: 20px;
  background: var(--neuro-text);
  border-radius: 50%;
  transition: transform 0.3s ease;
}

.neuro-switch:checked + .switch-slider {
  transform: translateX(20px);
}

.switch-label {
  color: #e2e8f0;
  font-size: 14px;
}

/* 日期字段 */
.date-field {
  display: flex;
  gap: 12px;
}

.date-today {
  padding: 12px 16px;
  background: var(--neuro-primary-10);
  border: 1px solid var(--neuro-primary-30);
  border-radius: var(--neuro-radius-lg);
  color: var(--neuro-primary);
  font-family: 'IBM Plex Mono', monospace;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s ease;
  var(--neuro-text)-space: nowrap;
}

.date-today:hover {
  background: var(--neuro-primary-20);
}

/* 文件上传 */
.file-field {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.file-label {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  background: rgba(15, 23, 42, 0.9);
  border: 2px dashed rgba(100, 116, 139, 0.5);
  border-radius: var(--neuro-radius-lg);
  color: var(--neuro-text-secondary);
  font-family: 'IBM Plex Mono', monospace;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.file-label:hover {
  border-color: var(--neuro-primary);
  color: var(--neuro-primary);
}

.file-icon {
  font-size: 16px;
}

.file-text {
  flex: 1;
}

.file-browse {
  color: var(--neuro-primary);
  font-weight: 600;
}

.file-preview {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  background: rgba(0, 245, 212, 0.05);
  border: 1px solid var(--neuro-primary-20);
  border-radius: var(--neuro-radius-lg);
}

.preview-name {
  flex: 1;
  color: var(--neuro-primary);
  font-size: 13px;
  font-weight: 500;
}

.preview-size {
  color: var(--neuro-text-secondary);
  font-size: 12px;
}

.preview-remove {
  width: 20px;
  height: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(248, 113, 113, 0.1);
  border: 1px solid rgba(248, 113, 113, 0.3);
  border-radius: 50%;
  color: #f87171;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.preview-remove:hover {
  background: rgba(248, 113, 113, 0.2);
}

/* 标签输入 */
.tags-field {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.tags-input-wrapper {
  display: flex;
  gap: 8px;
}

.neuro-tags-input {
  flex: 1;
}

.tags-add {
  padding: 12px 16px;
  background: var(--neuro-primary-10);
  border: 1px solid var(--neuro-primary-30);
  border-radius: var(--neuro-radius-lg);
  color: var(--neuro-primary);
  font-family: 'IBM Plex Mono', monospace;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s ease;
  var(--neuro-text)-space: nowrap;
}

.tags-add:hover {
  background: var(--neuro-primary-20);
}

.tags-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.tag-item {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  background: var(--neuro-primary-10);
  border: 1px solid var(--neuro-primary-30);
  border-radius: 20px;
  color: var(--neuro-primary);
  font-size: 12px;
}

.tag-remove {
  width: 16px;
  height: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(248, 113, 113, 0.1);
  border: 1px solid rgba(248, 113, 113, 0.3);
  border-radius: 50%;
  color: #f87171;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.tag-remove:hover {
  background: rgba(248, 113, 113, 0.2);
}

/* 字段帮助和错误 */
.field-help {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 8px;
  color: var(--neuro-text-secondary);
  font-size: 11px;
}

.help-icon {
  font-size: 12px;
}

.field-ai-fill-marker {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 8px;
  padding: 4px 8px;
  background: linear-gradient(135deg, var(--neuro-primary-10), rgba(0, 184, 169, 0.1));
  border: 1px solid var(--neuro-primary-30);
  border-radius: 6px;
  color: var(--neuro-primary);
  font-size: 11px;
  animation: ai-fill-pulse 2s infinite;
}

.ai-fill-icon {
  font-size: 12px;
}

.ai-fill-text {
  font-weight: 600;
}

.ai-fill-confidence {
  margin-left: auto;
  font-size: 10px;
  opacity: 0.8;
}

@keyframes ai-fill-pulse {
  0%, 100% {
    box-shadow: 0 0 0 0 rgba(0, 245, 212, 0.4);
  }
  50% {
    box-shadow: 0 0 0 4px rgba(0, 245, 212, 0);
  }
}

.field-error-message {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 8px;
  color: #f87171;
  font-size: 11px;
}

.error-icon {
  font-size: 12px;
}

/* AI自动完成 */
.field-autocomplete {
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  margin-top: 4px;
  background: rgba(15, 23, 42, 0.95);
  border: 1px solid var(--neuro-primary-30);
  border-radius: var(--neuro-radius-lg);
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.3);
  z-index: 10;
  overflow: hidden;
}

.autocomplete-item {
  padding: 12px 16px;
  color: var(--neuro-text-secondary);
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.autocomplete-item:hover {
  background: var(--neuro-primary-10);
  color: var(--neuro-primary);
}

/* 表单验证总结 */
.form-validation-summary {
  margin-bottom: 32px;
  padding: 20px;
  background: rgba(248, 113, 113, 0.05);
  border: 1px solid rgba(248, 113, 113, 0.2);
  border-radius: var(--neuro-radius-xl);
}

.validation-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}

.validation-icon {
  font-size: 18px;
}

.validation-title {
  color: #f87171;
  font-family: 'Orbitron', monospace;
  font-size: 16px;
  font-weight: 600;
  flex: 1;
}

.validation-count {
  color: #f87171;
  font-size: 12px;
  font-weight: 500;
}

.validation-errors {
  margin: 0;
  padding-left: 20px;
  color: #f87171;
  font-size: 13px;
  line-height: 1.6;
}

.validation-errors li {
  margin-bottom: 4px;
}

/* 表单操作 */
.form-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-top: 32px;
  border-top: 1px solid rgba(100, 116, 139, 0.3);
}

.actions-left,
.actions-center,
.actions-right {
  display: flex;
  gap: 12px;
}

/* 神经按钮样式 */
.neuro-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 24px;
  border: 1px solid;
  border-radius: var(--neuro-radius-lg);
  font-family: 'IBM Plex Mono', monospace;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.neuro-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.neuro-btn-primary {
  background: linear-gradient(135deg, var(--neuro-primary), #9d4edd);
  border-color: transparent;
  color: var(--neuro-text);
}

.neuro-btn-primary:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 10px 20px var(--neuro-primary-20);
}

.neuro-btn-secondary {
  background: var(--neuro-primary-10);
  border-color: var(--neuro-primary-30);
  color: var(--neuro-primary);
}

.neuro-btn-secondary:hover:not(:disabled) {
  background: var(--neuro-primary-20);
  transform: translateY(-2px);
}

.neuro-btn-ghost {
  background: transparent;
  border-color: rgba(100, 116, 139, 0.5);
  color: var(--neuro-text-secondary);
}

.neuro-btn-ghost:hover:not(:disabled) {
  border-color: var(--neuro-primary);
  color: var(--neuro-primary);
  transform: translateY(-2px);
}

.btn-icon {
  font-size: 14px;
}

/* 表单底部信息 */
.form-footer {
  margin-top: 32px;
  padding: 20px;
  background: rgba(0, 245, 212, 0.05);
  border: 1px solid var(--neuro-primary-20);
  border-radius: var(--neuro-radius-xl);
}

.footer-content {
  display: flex;
  align-items: center;
  gap: 12px;
}

.footer-icon {
  font-size: 16px;
}

.footer-text {
  color: var(--neuro-text-secondary);
  font-size: 13px;
  line-height: 1.5;
}

/* 表单成功状态 */
.form-success-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(15, 23, 42, 0.95);
  backdrop-filter: blur(20px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  animation: fadeIn 0.3s ease;
}

.success-content {
  max-width: 500px;
  padding: 40px;
  background: rgba(15, 23, 42, 0.9);
  border: 1px solid var(--neuro-primary-30);
  border-radius: var(--neuro-radius-2xl);
  text-align: center;
  animation: slideUp 0.4s ease;
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.success-icon {
  font-size: 64px;
  margin-bottom: 24px;
  animation: bounce 1s ease infinite;
}

@keyframes bounce {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-10px); }
}

.success-title {
  margin: 0 0 12px 0;
  font-family: 'Orbitron', monospace;
  font-size: 28px;
  font-weight: 700;
  background: linear-gradient(135deg, var(--neuro-primary), #9d4edd);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.success-message {
  margin: 0 0 32px 0;
  color: var(--neuro-text-secondary);
  font-size: 16px;
  line-height: 1.6;
}

.success-actions {
  display: flex;
  gap: 12px;
  justify-content: center;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .form-container {
    padding: 24px 16px;
  }

  .form-header {
    flex-direction: column;
    gap: 20px;
    text-align: center;
  }

  .header-left,
  .header-center,
  .header-right {
    width: 100%;
  }

  .header-right {
    justify-content: center;
  }

  .form-progress {
    align-items: center;
  }

  .steps-track {
    flex-wrap: wrap;
    gap: 20px;
  }

  .step-item {
    flex: none;
    width: calc(50% - 10px);
  }

  .form-fields {
    grid-template-columns: 1fr;
  }

  .form-actions {
    flex-direction: column;
    gap: 16px;
  }

  .actions-left,
  .actions-center,
  .actions-right {
    width: 100%;
    justify-content: center;
  }

  .success-actions {
    flex-direction: column;
  }
}

/* 浅色模式适配 */
.neuro-agent-form-page[data-theme="light"] .form-intelligence-background {
  background: linear-gradient(135deg, #f0f9ff 0%, #e0f2fe 100%);
}

.neuro-agent-form-page[data-theme="light"] .intelligence-grid {
  background:
    linear-gradient(90deg, rgba(14, 165, 233, 0.05) 1px, transparent 1px) 0 0 / 40px 40px,
    linear-gradient(rgba(14, 165, 233, 0.05) 1px, transparent 1px) 0 0 / 40px 40px;
}

.neuro-agent-form-page[data-theme="light"] .intelligence-particle {
  background: #0ea5e9;
}

.neuro-agent-form-page[data-theme="light"] .form-header,
.neuro-agent-form-page[data-theme="light"] .form-steps,
.neuro-agent-form-page[data-theme="light"] .form-content,
.neuro-agent-form-page[data-theme="light"] .ai-form-assistant,
.neuro-agent-form-page[data-theme="light"] .form-validation-summary,
.neuro-agent-form-page[data-theme="light"] .form-footer {
  background: rgba(248, 250, 252, 0.9);
  border-color: rgba(14, 165, 233, 0.2);
}

.neuro-agent-form-page[data-theme="light"] .form-title {
  background: linear-gradient(135deg, #0ea5e9, #8b5cf6);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

.neuro-agent-form-page[data-theme="light"] .form-subtitle,
.neuro-agent-form-page[data-theme="light"] .step-label,
.neuro-agent-form-page[data-theme="light"] .suggestion-text,
.neuro-agent-form-page[data-theme="light"] .field-description,
.neuro-agent-form-page[data-theme="light"] .checkbox-label,
.neuro-agent-form-page[data-theme="light"] .radio-label,
.neuro-agent-form-page[data-theme="light"] .preview-size,
.neuro-agent-form-page[data-theme="light"] .footer-text {
  color: #64748b;
}

.neuro-agent-form-page[data-theme="light"] .field-label,
.neuro-agent-form-page[data-theme="light"] .switch-label {
  color: var(--neuro-surface);
}

.neuro-agent-form-page[data-theme="light"] .neuro-input,
.neuro-agent-form-page[data-theme="light"] .neuro-textarea,
.neuro-agent-form-page[data-theme="light"] .neuro-select,
.neuro-agent-form-page[data-theme="light"] .neuro-date,
.neuro-agent-form-page[data-theme="light"] .file-label,
.neuro-agent-form-page[data-theme="light"] .suggestion-item {
  background: rgba(248, 250, 252, 0.9);
  border-color: rgba(100, 116, 139, 0.3);
  color: var(--neuro-surface);
}

.neuro-agent-form-page[data-theme="light"] .back-button,
.neuro-agent-form-page[data-theme="light"] .assistant-toggle,
.neuro-agent-form-page[data-theme="light"] .date-today,
.neuro-agent-form-page[data-theme="light"] .tags-add,
.neuro-agent-form-page[data-theme="light"] .neuro-btn-secondary {
  background: rgba(14, 165, 233, 0.1);
  border-color: rgba(14, 165, 233, 0.3);
  color: #0ea5e9;
}

.neuro-agent-form-page[data-theme="light"] .progress-fill,
.neuro-agent-form-page[data-theme="light"] .neuro-btn-primary {
  background: linear-gradient(135deg, #0ea5e9, #8b5cf6);
}

.neuro-agent-form-page[data-theme="light"] .step-completed .step-marker,
.neuro-agent-form-page[data-theme="light"] .step-active .step-marker {
  background: rgba(14, 165, 233, 0.1);
  border-color: #0ea5e9;
}

.neuro-agent-form-page[data-theme="light"] .step-completed .marker-icon,
.neuro-agent-form-page[data-theme="light"] .step-active .marker-number,
.neuro-agent-form-page[data-theme="light"] .step-completed .step-label,
.neuro-agent-form-page[data-theme="light"] .step-active .step-label,
.neuro-agent-form-page[data-theme="light"] .progress-text,
.neuro-agent-form-page[data-theme="light"] .assistant-title,
.neuro-agent-form-page[data-theme="light"] .label-ai,
.neuro-agent-form-page[data-theme="light"] .preview-name,
.neuro-agent-form-page[data-theme="light"] .tag-item {
  color: #0ea5e9;
}

.neuro-agent-form-page[data-theme="light"] .field-error .neuro-input,
.neuro-agent-form-page[data-theme="light"] .field-error .neuro-textarea,
.neuro-agent.form-page[data-theme="light"] .field-error .neuro-select,
.neuro-agent-form-page[data-theme="light"] .field-error .neuro-date {
  border-color: #dc2626;
  background: rgba(220, 38, 38, 0.05);
}

.neuro-agent-form-page[data-theme="light"] .field-error .field-label,
.neuro-agent-form-page[data-theme="light"] .validation-title,
.neuro-agent-form-page[data-theme="light"] .validation-count,
.neuro-agent-form-page[data-theme="light"] .validation-errors {
  color: #dc2626;
}
</style>