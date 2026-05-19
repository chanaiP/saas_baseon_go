import http, { unwrap } from '@/api/http'
import type { ApiResponse } from '@/api/types'

import type {
  AnomalyRecord,
  AnomalyRule,
  DashboardSummary,
  DataCenterQuery,
  MetricDefinition,
  PageResponse,
  RawBatch,
  RectificationReview,
  RectificationTask,
} from './types'

function toSnakeKey(key: string) {
  return key
    .replace(/ID/g, 'Id')
    .replace(/JSON/g, 'Json')
    .replace(/AI/g, 'Ai')
    .replace(/([a-z0-9])([A-Z])/g, '$1_$2')
    .toLowerCase()
}

function normalizeKeys<T>(value: T): T {
  if (Array.isArray(value)) return value.map((item) => normalizeKeys(item)) as T
  if (!value || typeof value !== 'object') return value
  const next: Record<string, unknown> = {}
  for (const [key, item] of Object.entries(value as Record<string, unknown>)) {
    next[toSnakeKey(key)] = normalizeKeys(item)
  }
  return next as T
}

async function unwrapNormalized<T>(p: Promise<{ data: ApiResponse<T> }>) {
  return normalizeKeys(await unwrap<T>(p))
}

export function fetchDashboardSummary(params?: DataCenterQuery) {
  return unwrapNormalized<DashboardSummary>(http.get('/api/data-center/dashboard/summary', { params }))
}

export function fetchDashboardTrends(params?: DataCenterQuery) {
  return unwrapNormalized<Array<Record<string, unknown>>>(http.get('/api/data-center/dashboard/trends', { params }))
}

export function fetchDashboardRankings(params?: DataCenterQuery) {
  return unwrapNormalized<Record<string, Array<Record<string, unknown>>>>(http.get('/api/data-center/dashboard/rankings', { params }))
}

export function fetchPipeline() {
  return unwrapNormalized<{ items: Array<Record<string, unknown>> }>(http.get('/api/data-center/overview/pipeline'))
}

export function fetchRawBatches(params?: DataCenterQuery) {
  return unwrapNormalized<PageResponse<RawBatch>>(http.get('/api/data-center/raw/batches', { params }))
}

export function fetchRawBatch(id: number) {
  return unwrapNormalized<RawBatch>(http.get(`/api/data-center/raw/batches/${encodeURIComponent(String(id))}`))
}

export function fetchRawErrors(id: number, params?: DataCenterQuery) {
  return unwrapNormalized<PageResponse<Record<string, unknown>>>(http.get(`/api/data-center/raw/batches/${encodeURIComponent(String(id))}/errors`, { params }))
}

export function reprocessRawBatch(id: number) {
  return unwrapNormalized<RawBatch>(http.post(`/api/data-center/raw/batches/${encodeURIComponent(String(id))}/reprocess`))
}

export function fetchStandardData<T = Record<string, unknown>>(dataType: string, params?: DataCenterQuery) {
  return unwrapNormalized<PageResponse<T>>(http.get(`/api/data-center/standard/${encodeURIComponent(dataType)}`, { params }))
}

export function fetchStandardDataDetail<T = Record<string, unknown>>(dataType: string, id: number) {
  return unwrapNormalized<T>(http.get(`/api/data-center/standard/${encodeURIComponent(dataType)}/${encodeURIComponent(String(id))}`))
}

export function fetchMetrics(params?: DataCenterQuery) {
  return unwrapNormalized<PageResponse<MetricDefinition>>(http.get('/api/data-center/metrics', { params }))
}

export function fetchMetric(id: number) {
  return unwrapNormalized<MetricDefinition>(http.get(`/api/data-center/metrics/${encodeURIComponent(String(id))}`))
}

export function createMetric(payload: Record<string, unknown>) {
  return unwrapNormalized<MetricDefinition>(http.post('/api/data-center/metrics', payload))
}

export function updateMetric(id: number, payload: Record<string, unknown>) {
  return unwrapNormalized<MetricDefinition>(http.put(`/api/data-center/metrics/${encodeURIComponent(String(id))}`, payload))
}

export function setMetricEnabled(id: number, enabled: boolean) {
  return unwrapNormalized<Record<string, unknown>>(http.post(`/api/data-center/metrics/${encodeURIComponent(String(id))}/${enabled ? 'enable' : 'disable'}`))
}

export function fetchRules(params?: DataCenterQuery) {
  return unwrapNormalized<PageResponse<AnomalyRule>>(http.get('/api/data-center/anomaly-rules', { params }))
}

export function fetchRule(id: number) {
  return unwrapNormalized<AnomalyRule>(http.get(`/api/data-center/anomaly-rules/${encodeURIComponent(String(id))}`))
}

export function createRule(payload: Record<string, unknown>) {
  return unwrapNormalized<AnomalyRule>(http.post('/api/data-center/anomaly-rules', payload))
}

export function updateRule(id: number, payload: Record<string, unknown>) {
  return unwrapNormalized<AnomalyRule>(http.put(`/api/data-center/anomaly-rules/${encodeURIComponent(String(id))}`, payload))
}

export function setRuleEnabled(id: number, enabled: boolean) {
  return unwrapNormalized<Record<string, unknown>>(http.post(`/api/data-center/anomaly-rules/${encodeURIComponent(String(id))}/${enabled ? 'enable' : 'disable'}`))
}

export function testRule(id: number, payload: Record<string, unknown>) {
  return unwrapNormalized<Record<string, unknown>>(http.post(`/api/data-center/anomaly-rules/${encodeURIComponent(String(id))}/test`, payload))
}

export function fetchAnomalies(params?: DataCenterQuery) {
  return unwrapNormalized<PageResponse<AnomalyRecord>>(http.get('/api/data-center/anomalies', { params }))
}

export function fetchAnomaly(id: number) {
  return unwrapNormalized<AnomalyRecord>(http.get(`/api/data-center/anomalies/${encodeURIComponent(String(id))}`))
}

export function analyzeAnomaly(id: number) {
  return unwrapNormalized<Record<string, unknown>>(http.post(`/api/data-center/anomalies/${encodeURIComponent(String(id))}/analyze`))
}

export function reanalyzeAnomaly(id: number) {
  return unwrapNormalized<Record<string, unknown>>(http.post(`/api/data-center/anomalies/${encodeURIComponent(String(id))}/reanalyze`))
}

export function generateTaskFromAnomaly(id: number) {
  return unwrapNormalized<RectificationTask>(http.post(`/api/data-center/anomalies/${encodeURIComponent(String(id))}/generate-task`))
}

export function scanAnomalies(params?: DataCenterQuery) {
  return unwrapNormalized<Record<string, unknown>>(http.post('/api/data-center/anomalies/scan', null, { params }))
}

export function updateAnomalyStatus(id: number, action: 'confirm' | 'ignore' | 'close') {
  return unwrapNormalized<Record<string, unknown>>(http.post(`/api/data-center/anomalies/${encodeURIComponent(String(id))}/${action}`))
}

export function fetchTasks(params?: DataCenterQuery) {
  return unwrapNormalized<PageResponse<RectificationTask>>(http.get('/api/data-center/tasks', { params }))
}

export function fetchTask(id: number) {
  return unwrapNormalized<RectificationTask>(http.get(`/api/data-center/tasks/${encodeURIComponent(String(id))}`))
}

export function createTask(payload: Record<string, unknown>) {
  return unwrapNormalized<RectificationTask>(http.post('/api/data-center/tasks', payload))
}

export function updateTask(id: number, payload: Record<string, unknown>) {
  return unwrapNormalized<RectificationTask>(http.put(`/api/data-center/tasks/${encodeURIComponent(String(id))}`, payload))
}

export function startTask(id: number) {
  return unwrapNormalized<RectificationTask>(http.post(`/api/data-center/tasks/${encodeURIComponent(String(id))}/start`))
}

export function completeTask(id: number) {
  return unwrapNormalized<RectificationTask>(http.post(`/api/data-center/tasks/${encodeURIComponent(String(id))}/complete`))
}

export function feedbackTask(id: number, payload: Record<string, unknown>) {
  return unwrapNormalized<RectificationTask>(http.post(`/api/data-center/tasks/${encodeURIComponent(String(id))}/feedback`, payload))
}

export function closeTask(id: number) {
  return unwrapNormalized<RectificationTask>(http.post(`/api/data-center/tasks/${encodeURIComponent(String(id))}/close`))
}

export function fetchReviews(params?: DataCenterQuery) {
  return unwrapNormalized<PageResponse<RectificationReview>>(http.get('/api/data-center/reviews', { params }))
}

export function fetchReview(id: number) {
  return unwrapNormalized<RectificationReview>(http.get(`/api/data-center/reviews/${encodeURIComponent(String(id))}`))
}

export function generateReview(payload: Record<string, unknown>) {
  return unwrapNormalized<RectificationReview>(http.post('/api/data-center/reviews/generate', payload))
}

export function confirmReview(id: number, payload: Record<string, unknown>) {
  return unwrapNormalized<RectificationReview>(http.post(`/api/data-center/reviews/${encodeURIComponent(String(id))}/confirm`, payload))
}

export function updateReview(id: number, payload: Record<string, unknown>) {
  return unwrapNormalized<RectificationReview>(http.put(`/api/data-center/reviews/${encodeURIComponent(String(id))}`, payload))
}
