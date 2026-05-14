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

export function numberText(value: unknown) {
  const numeric = Number(value ?? 0)
  return Number.isFinite(numeric) ? numeric.toLocaleString('zh-CN') : '0'
}

export function moneyText(value: unknown) {
  const numeric = Number(value ?? 0)
  return Number.isFinite(numeric) ? `¥${numeric.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}` : '¥0.00'
}

export function percentText(value: unknown) {
  const numeric = Number(value ?? 0)
  if (!Number.isFinite(numeric)) return '0.00%'
  return `${numeric.toFixed(2)}%`
}

export function statusType(value: unknown) {
  if (value === true || value === 'active' || value === 'success') return 'success'
  if (value === 'warning' || value === 'degraded') return 'warning'
  if (value === false || value === 'error' || value === 'failed' || value === 'timeout' || value === 'inactive') return 'danger'
  return 'info'
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
