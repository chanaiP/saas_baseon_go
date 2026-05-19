import type { AiPage } from '../types'

export type AiRow = Record<string, unknown>

export const emptyPage = <T = AiRow>(): AiPage<T> => ({ items: [], total: 0, skip: 0, limit: 200 })

export function text(value: unknown, fallback = '-') {
  if (Array.isArray(value)) return value.length ? value.join('、') : fallback
  if (value === null || value === undefined || value === '') return fallback
  if (typeof value === 'object') return JSON.stringify(value)
  return String(value)
}

const modelTypeLabels: Record<string, string> = {
  text: '文本模型',
  image: '图像模型',
  embedding: '向量模型',
  audio: '音频模型',
  video: '视频模型',
  multimodal: '多模态模型',
  rerank: '重排模型',
  unknown: '未归类模型',
}

export function modelTypeText(value: unknown, fallback = '-') {
  const raw = text(value, '')
  return raw ? modelTypeLabels[raw] || raw : fallback
}

const sceneTypeLabels: Record<string, string> = {
  agent: 'Agent 场景',
  chat: '文本对话',
  image: '图片生成',
  image_generation: '图片生成',
  multimodal: '多模态',
  text: '文本生成',
  video: '视频生成',
  video_generation: '视频生成',
  workflow: '工作流',
}

export function sceneTypeText(value: unknown, fallback = '-') {
  const raw = text(value, '').toLowerCase()
  return raw ? sceneTypeLabels[raw] || raw : fallback
}

const capabilityLabels: Record<string, string> = {
  agent_run: 'Agent 执行',
  chat_completion: '对话生成',
  embedding: '向量生成',
  image_generation: '图像生成',
  long_context: '长上下文',
  multimodal: '多模态',
  reasoning: '推理增强',
  rerank: '重排',
  text_generation: '文本生成',
  tool_calling: '工具调用',
  video_generation: '视频生成',
  vision: '视觉理解',
}

function normalizeList(value: unknown) {
  if (Array.isArray(value)) return value.map((item) => text(item, '')).filter(Boolean)
  const raw = text(value, '')
  if (!raw) return []
  if (raw.startsWith('[')) {
    try {
      const parsed = JSON.parse(raw) as unknown
      if (Array.isArray(parsed)) return parsed.map((item) => text(item, '')).filter(Boolean)
    } catch {
      // Fall through to delimiter parsing for legacy strings.
    }
  }
  return raw.split(/[,\s、，]+/).map((item) => item.trim()).filter(Boolean)
}

export function capabilityTagList(value: unknown) {
  return normalizeList(value).map((code) => ({ code, label: capabilityLabels[code] || code }))
}

export function capabilityText(value: unknown, fallback = '-') {
  const labels = capabilityTagList(value).map((item) => item.label)
  return labels.length ? labels.join('、') : fallback
}

export function numberText(value: unknown) {
  const numeric = Number(value ?? 0)
  return Number.isFinite(numeric) ? numeric.toLocaleString('zh-CN') : '0'
}

export function moneyText(value: unknown) {
  const numeric = Number(value ?? 0)
  if (!Number.isFinite(numeric)) return '¥0.00'
  const abs = Math.abs(numeric)
  const fractionDigits = abs > 0 && abs < 0.01 ? 6 : 2
  return `¥${numeric.toLocaleString('zh-CN', { minimumFractionDigits: fractionDigits, maximumFractionDigits: fractionDigits })}`
}

export function priceText(value: unknown) {
  const numeric = Number(value ?? 0)
  if (!Number.isFinite(numeric)) return '¥0'
  return `¥${numeric.toLocaleString('zh-CN', { minimumFractionDigits: 4, maximumFractionDigits: 4 })}`
}

export function billingModeText(value: unknown, fallback = '-') {
  const raw = text(value, '').toLowerCase()
  const labels: Record<string, string> = {
    fixed: '固定计费',
    tiered: '分档计费',
    usage: '按量计费',
  }
  return raw ? labels[raw] || raw : fallback
}

export function billingUnitText(value: unknown, fallback = '-') {
  const raw = text(value, '')
  const labels: Record<string, string> = {
    agent_runs: 'Agent 执行',
    image: '图片',
    images: '图片',
    request: '次',
    requests: '次',
    second: '秒',
    seconds: '秒',
    tokens: 'tokens',
    video_seconds: '视频秒',
  }
  return raw ? labels[raw] || raw : fallback
}

export function percentText(value: unknown) {
  const numeric = Number(value ?? 0)
  if (!Number.isFinite(numeric)) return '0.00%'
  return `${numeric.toFixed(2)}%`
}

export function statusType(value: unknown) {
  const raw = typeof value === 'string' ? value.toLowerCase() : value
  if (raw === true || raw === 'active' || raw === 'success' || raw === 'enabled' || raw === 'online') return 'success'
  if (raw === 'warning' || raw === 'degraded' || raw === 'pending' || raw === 'draft' || raw === 'provider_error') return 'warning'
  if (raw === false || raw === 'error' || raw === 'failed' || raw === 'timeout' || raw === 'inactive' || raw === 'disabled' || raw === 'offline') return 'danger'
  return 'info'
}

export function statusText(value: unknown, fallback = '-') {
  const raw = text(value, '').toLowerCase()
  const labels: Record<string, string> = {
    active: '正常',
    success: '成功',
    enabled: '启用',
    online: '在线',
    warning: '告警',
    degraded: '降级',
    pending: '待处理',
    draft: '草稿',
    error: '异常',
    failed: '失败',
    timeout: '超时',
    inactive: '停用',
    disabled: '停用',
    offline: '离线',
    unknown: '未检测',
    provider_error: '供应商异常',
  }
  return raw ? labels[raw] || raw : fallback
}

export function settingJSON(row?: AiRow) {
  const value = row?.setting_value
  if (!value) return {} as AiRow
  if (typeof value === 'object') return value as AiRow
  try {
    return JSON.parse(String(value)) as AiRow
  } catch {
    return {} as AiRow
  }
}

export function rowId(row?: AiRow) {
  return String(row?.id || '')
}

export function includesKeyword(row: AiRow, keyword: string, keys: string[]) {
  const normalized = keyword.trim().toLowerCase()
  if (!normalized) return true
  return keys.some((key) => text(row[key], '').toLowerCase().includes(normalized))
}

export function compactJoin(values: unknown[]) {
  return values.map((value) => text(value, '')).filter(Boolean).join(' · ') || '-'
}
