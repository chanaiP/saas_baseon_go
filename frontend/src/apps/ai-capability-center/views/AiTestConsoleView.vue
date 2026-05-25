<script setup lang="ts">
import { computed, onBeforeUnmount, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { ChatLineRound, Film, Picture, Promotion, Refresh } from '@element-plus/icons-vue'

import NeuroAgentPageShell from '@/views/components/NeuroAgentPageShell.vue'

import { fetchAiGatewayVideoTask, invokeAiGateway } from '../api'
import { statusText, statusType, text } from './viewHelpers'
import './aiPrototype.css'

defineOptions({ name: 'AiTestConsoleView' })

type TestMode = 'text' | 'kimi' | 'image' | 'video'

interface ConsoleMessage {
  id: number
  role: 'user' | 'assistant' | 'system'
  mode: TestMode
  content: string
  meta?: Record<string, unknown>
}

const mode = ref<TestMode>('text')
const prompt = ref('请只回答：链路已跑通')
const loading = ref(false)
const messages = ref<ConsoleMessage[]>([])
const lastResult = ref<Record<string, unknown> | null>(null)
const pollingTaskId = ref('')
const pollingAttempt = ref(0)
let pollingTimer: number | undefined

const modeOptions = [
  {
    value: 'text' as const,
    label: '文本聊天',
    icon: ChatLineRound,
    scenario: 'test_console_chat',
    description: '走默认文本聊天路由，验证当前可用供应商链路。',
    placeholder: '输入一段文本，验证聊天路由。',
  },
  {
    value: 'kimi' as const,
    label: 'Kimi 文本',
    icon: ChatLineRound,
    scenario: 'kimi_test_console_chat',
    description: '走 kimi-k2.6，验证 Moonshot 国内接口链路。',
    placeholder: '输入一段文本，验证 Kimi 聊天路由。',
  },
  {
    value: 'image' as const,
    label: '图片生成',
    icon: Picture,
    scenario: 'image_generation',
    description: '走默认图片生成路由，提交图片生成请求。',
    placeholder: '描述要生成的图片，例如：一张干净的 SaaS 控制台概念图。',
  },
  {
    value: 'video' as const,
    label: '视频任务',
    icon: Film,
    scenario: 'video_generation',
    description: '走视频生成异步任务，返回 task_id 后在调用日志追踪。',
    placeholder: '描述 3 秒视频，例如：一只小猫在月光下奔跑。',
  },
]

const activeMode = computed(() => modeOptions.find((item) => item.value === mode.value) ?? modeOptions[0])
const canSubmit = computed(() => prompt.value.trim().length > 0 && !loading.value)
const resultStatus = computed(() => text(lastResult.value?.status, '未调用'))

function selectMode(value: TestMode) {
  mode.value = value
  if (value === 'text') prompt.value = '请只回答：链路已跑通'
  if (value === 'kimi') prompt.value = '请只回答：Kimi 链路已跑通'
  if (value === 'image') prompt.value = '生成一张 16:9 的未来感 SaaS 数据驾驶舱，深色背景，清晰 UI 面板'
  if (value === 'video') prompt.value = '一只小猫在月光下奔跑，镜头缓慢跟随，电影感，3 秒'
}

function buildPayload() {
  const base: Record<string, unknown> = {
    tenant_id: '1',
    app_code: 'ai-capability-center',
    app_name: 'AI 能力中心',
    ai_scenario_code: activeMode.value.scenario,
    input: { prompt: prompt.value.trim() },
    params: {},
  }
  if (mode.value === 'text') {
    base.params = { max_tokens: 128, temperature: 0.2 }
  } else if (mode.value === 'kimi') {
    base.params = { max_tokens: 256, temperature: 0.6, thinking: { type: 'disabled' } }
  } else if (mode.value === 'image') {
    base.params = { n: 1, size: '1024*1024', prompt_extend: false, watermark: false }
  } else {
    base.params = { duration: 3, resolution: '720P', ratio: '16:9', audio: false, watermark: true, prompt_extend: false }
  }
  return base
}

function responseText(data: Record<string, unknown>) {
  const payload = data.data as Record<string, unknown> | undefined
  if (!payload) return '供应商已返回结果，但没有标准化内容。'
  if (mode.value === 'text' || mode.value === 'kimi') return text(payload.text, '已返回文本响应。')
  if (mode.value === 'image') {
    const urls = Array.isArray(payload.urls) ? payload.urls : []
    return urls.length ? `图片生成成功：${urls[0]}` : `图片任务已返回：${text(payload.image_count, '0')} 张`
  }
  return `视频任务已提交：${text(payload.task_id, '等待 task_id')}`
}

function stringValue(value: unknown) {
  return typeof value === 'string' ? value : ''
}

function stopVideoPolling() {
  if (pollingTimer !== undefined) {
    window.clearTimeout(pollingTimer)
    pollingTimer = undefined
  }
  pollingTaskId.value = ''
  pollingAttempt.value = 0
}

async function pollVideoTask(taskId: string, requestId: string, message: ConsoleMessage) {
  pollingTaskId.value = taskId
  pollingAttempt.value += 1
  try {
    const task = await fetchAiGatewayVideoTask(taskId, requestId)
    const status = text(task.task_status, 'UNKNOWN')
    message.content = status === 'SUCCEEDED' ? '视频生成完成' : `视频任务处理中：${status}`
    message.meta = {
      ...(message.meta ?? {}),
      task_status: status,
      task_id: task.task_id,
      video_url: task.video_url,
      usage: task.usage ?? message.meta?.usage,
      query_request_id: task.request_id,
      message: task.message,
    }
    if (status === 'SUCCEEDED' || status === 'FAILED' || status === 'CANCELED' || pollingAttempt.value >= 40) {
      if (status !== 'SUCCEEDED' && pollingAttempt.value >= 40) message.content = '视频任务仍在处理中，请稍后到调用日志继续查看。'
      stopVideoPolling()
      return
    }
    pollingTimer = window.setTimeout(() => pollVideoTask(taskId, requestId, message), 5000)
  } catch (error) {
    const messageText = error instanceof Error ? error.message : '视频任务查询失败'
    message.content = messageText
    message.meta = { ...(message.meta ?? {}), task_status: 'QUERY_FAILED' }
    stopVideoPolling()
  }
}

async function submit() {
  if (!canSubmit.value) return
  stopVideoPolling()
  const userMessage: ConsoleMessage = {
    id: Date.now(),
    role: 'user',
    mode: mode.value,
    content: prompt.value.trim(),
  }
  messages.value.push(userMessage)
  loading.value = true
  try {
    const result = await invokeAiGateway(buildPayload())
    lastResult.value = result
    const assistantMessage: ConsoleMessage = {
      id: Date.now() + 1,
      role: 'assistant',
      mode: mode.value,
      content: responseText(result),
      meta: {
        request_id: result.request_id,
        provider_request_id: result.provider_request_id,
        status: result.status,
        usage: result.usage,
      },
    }
    messages.value.push(assistantMessage)
    const data = result.data as Record<string, unknown> | undefined
    const taskId = stringValue(data?.task_id)
    if (mode.value === 'video' && result.status === 'success' && taskId && result.request_id) {
      assistantMessage.content = `视频任务已提交，正在查询生成状态：${taskId}`
      pollingAttempt.value = 0
      pollingTimer = window.setTimeout(() => pollVideoTask(taskId, text(result.request_id), assistantMessage), 3000)
    }
    ElMessage.success('调用成功，已写入调用日志')
  } catch (error) {
    const message = error instanceof Error ? error.message : '调用失败'
    messages.value.push({ id: Date.now() + 1, role: 'system', mode: mode.value, content: message })
    ElMessage.error(message)
  } finally {
    loading.value = false
  }
}

function resetConsole() {
  stopVideoPolling()
  messages.value = []
  lastResult.value = null
}

onBeforeUnmount(stopVideoPolling)
</script>

<template>
  <NeuroAgentPageShell class="ai-prototype" :show-hero="false">
    <template #title>测试窗口</template>
    <template #subtitle>用同一个 AI Gateway 入口验证文本、图片、视频场景的路由、供应商连通和调用日志沉淀。</template>

    <div class="ai-test-console">
      <aside class="ai-test-console__modes">
        <button
          v-for="item in modeOptions"
          :key="item.value"
          type="button"
          :class="{ active: mode === item.value }"
          @click="selectMode(item.value)"
        >
          <el-icon><component :is="item.icon" /></el-icon>
          <span><strong>{{ item.label }}</strong><small>{{ item.description }}</small></span>
        </button>
      </aside>

      <section class="ai-card ai-test-console__main">
        <header class="ai-card__header">
          <div>
            <h3>{{ activeMode.label }}</h3>
            <p class="ai-card__description">{{ activeMode.description }}</p>
          </div>
          <el-tag :type="statusType(resultStatus)">{{ statusText(resultStatus, resultStatus) }}</el-tag>
        </header>

        <div class="ai-test-chat">
          <div v-if="!messages.length" class="ai-test-empty">
            <strong>等待一次真实调用</strong>
            <span>点击发送后会经过场景、策略、路由、模型节点和供应商 API，并在调用日志生成一条记录。</span>
          </div>
          <article v-for="item in messages" :key="item.id" class="ai-test-message" :class="`is-${item.role}`">
            <strong>{{ item.role === 'user' ? '你' : item.role === 'assistant' ? 'AI Gateway' : '系统' }}</strong>
            <p>{{ item.content }}</p>
            <dl v-if="item.meta" class="ai-test-meta">
              <div v-for="(value, key) in item.meta" :key="key">
                <dt>{{ key }}</dt>
                <dd>{{ text(value) }}</dd>
              </div>
            </dl>
            <div v-if="stringValue(item.meta?.video_url)" class="ai-test-video-result">
              <video :src="stringValue(item.meta?.video_url)" controls playsinline preload="metadata" />
              <a :href="stringValue(item.meta?.video_url)" target="_blank" rel="noreferrer">打开视频临时链接</a>
            </div>
          </article>
        </div>

        <footer class="ai-test-input">
          <el-input
            v-model="prompt"
            type="textarea"
            :rows="4"
            resize="none"
            :placeholder="activeMode.placeholder"
            @keydown.meta.enter.prevent="submit"
            @keydown.ctrl.enter.prevent="submit"
          />
          <div class="ai-test-input__actions">
            <el-button :icon="Refresh" @click="resetConsole">清空</el-button>
            <el-button type="primary" :icon="Promotion" :loading="loading" :disabled="!canSubmit" @click="submit">发送并验证</el-button>
          </div>
        </footer>
      </section>
    </div>
  </NeuroAgentPageShell>
</template>
