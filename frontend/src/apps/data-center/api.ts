import http, { unwrap } from '@/api/http'

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

export function fetchDashboardSummary(params?: DataCenterQuery) {
  return unwrap<DashboardSummary>(http.get('/api/data-center/dashboard/summary', { params }))
}

export function fetchDashboardTrends(params?: DataCenterQuery) {
  return unwrap<Array<Record<string, unknown>>>(http.get('/api/data-center/dashboard/trends', { params }))
}

export function fetchDashboardRankings(params?: DataCenterQuery) {
  return unwrap<Record<string, Array<Record<string, unknown>>>>(http.get('/api/data-center/dashboard/rankings', { params }))
}

export function fetchPipeline() {
  return unwrap<{ items: Array<Record<string, unknown>> }>(http.get('/api/data-center/overview/pipeline'))
}

export function fetchRawBatches(params?: DataCenterQuery) {
  return unwrap<PageResponse<RawBatch>>(http.get('/api/data-center/raw/batches', { params }))
}

export function fetchRawBatch(id: number) {
  return unwrap<RawBatch>(http.get(`/api/data-center/raw/batches/${encodeURIComponent(String(id))}`))
}

export function fetchRawErrors(id: number, params?: DataCenterQuery) {
  return unwrap<PageResponse<Record<string, unknown>>>(http.get(`/api/data-center/raw/batches/${encodeURIComponent(String(id))}/errors`, { params }))
}

export function reprocessRawBatch(id: number) {
  return unwrap<RawBatch>(http.post(`/api/data-center/raw/batches/${encodeURIComponent(String(id))}/reprocess`))
}

export function fetchStandardData<T = Record<string, unknown>>(dataType: string, params?: DataCenterQuery) {
  return unwrap<PageResponse<T>>(http.get(`/api/data-center/standard/${encodeURIComponent(dataType)}`, { params }))
}

export function fetchStandardDataDetail<T = Record<string, unknown>>(dataType: string, id: number) {
  return unwrap<T>(http.get(`/api/data-center/standard/${encodeURIComponent(dataType)}/${encodeURIComponent(String(id))}`))
}

export function fetchMetrics(params?: DataCenterQuery) {
  return unwrap<PageResponse<MetricDefinition>>(http.get('/api/data-center/metrics', { params }))
}

export function fetchMetric(id: number) {
  return unwrap<MetricDefinition>(http.get(`/api/data-center/metrics/${encodeURIComponent(String(id))}`))
}

export function createMetric(payload: Record<string, unknown>) {
  return unwrap<MetricDefinition>(http.post('/api/data-center/metrics', payload))
}

export function updateMetric(id: number, payload: Record<string, unknown>) {
  return unwrap<MetricDefinition>(http.put(`/api/data-center/metrics/${encodeURIComponent(String(id))}`, payload))
}

export function setMetricEnabled(id: number, enabled: boolean) {
  return unwrap<Record<string, unknown>>(http.post(`/api/data-center/metrics/${encodeURIComponent(String(id))}/${enabled ? 'enable' : 'disable'}`))
}

export function fetchRules(params?: DataCenterQuery) {
  return unwrap<PageResponse<AnomalyRule>>(http.get('/api/data-center/anomaly-rules', { params }))
}

export function fetchRule(id: number) {
  return unwrap<AnomalyRule>(http.get(`/api/data-center/anomaly-rules/${encodeURIComponent(String(id))}`))
}

export function createRule(payload: Record<string, unknown>) {
  return unwrap<AnomalyRule>(http.post('/api/data-center/anomaly-rules', payload))
}

export function updateRule(id: number, payload: Record<string, unknown>) {
  return unwrap<AnomalyRule>(http.put(`/api/data-center/anomaly-rules/${encodeURIComponent(String(id))}`, payload))
}

export function setRuleEnabled(id: number, enabled: boolean) {
  return unwrap<Record<string, unknown>>(http.post(`/api/data-center/anomaly-rules/${encodeURIComponent(String(id))}/${enabled ? 'enable' : 'disable'}`))
}

export function testRule(id: number, payload: Record<string, unknown>) {
  return unwrap<Record<string, unknown>>(http.post(`/api/data-center/anomaly-rules/${encodeURIComponent(String(id))}/test`, payload))
}

export function fetchAnomalies(params?: DataCenterQuery) {
  return unwrap<PageResponse<AnomalyRecord>>(http.get('/api/data-center/anomalies', { params }))
}

export function fetchAnomaly(id: number) {
  return unwrap<AnomalyRecord>(http.get(`/api/data-center/anomalies/${encodeURIComponent(String(id))}`))
}

export function analyzeAnomaly(id: number) {
  return unwrap<Record<string, unknown>>(http.post(`/api/data-center/anomalies/${encodeURIComponent(String(id))}/analyze`))
}

export function reanalyzeAnomaly(id: number) {
  return unwrap<Record<string, unknown>>(http.post(`/api/data-center/anomalies/${encodeURIComponent(String(id))}/reanalyze`))
}

export function generateTaskFromAnomaly(id: number) {
  return unwrap<RectificationTask>(http.post(`/api/data-center/anomalies/${encodeURIComponent(String(id))}/generate-task`))
}

export function scanAnomalies(params?: DataCenterQuery) {
  return unwrap<Record<string, unknown>>(http.post('/api/data-center/anomalies/scan', null, { params }))
}

export function updateAnomalyStatus(id: number, action: 'confirm' | 'ignore' | 'close') {
  return unwrap<Record<string, unknown>>(http.post(`/api/data-center/anomalies/${encodeURIComponent(String(id))}/${action}`))
}

export function fetchTasks(params?: DataCenterQuery) {
  return unwrap<PageResponse<RectificationTask>>(http.get('/api/data-center/tasks', { params }))
}

export function fetchTask(id: number) {
  return unwrap<RectificationTask>(http.get(`/api/data-center/tasks/${encodeURIComponent(String(id))}`))
}

export function createTask(payload: Record<string, unknown>) {
  return unwrap<RectificationTask>(http.post('/api/data-center/tasks', payload))
}

export function updateTask(id: number, payload: Record<string, unknown>) {
  return unwrap<RectificationTask>(http.put(`/api/data-center/tasks/${encodeURIComponent(String(id))}`, payload))
}

export function startTask(id: number) {
  return unwrap<RectificationTask>(http.post(`/api/data-center/tasks/${encodeURIComponent(String(id))}/start`))
}

export function completeTask(id: number) {
  return unwrap<RectificationTask>(http.post(`/api/data-center/tasks/${encodeURIComponent(String(id))}/complete`))
}

export function feedbackTask(id: number, payload: Record<string, unknown>) {
  return unwrap<RectificationTask>(http.post(`/api/data-center/tasks/${encodeURIComponent(String(id))}/feedback`, payload))
}

export function closeTask(id: number) {
  return unwrap<RectificationTask>(http.post(`/api/data-center/tasks/${encodeURIComponent(String(id))}/close`))
}

export function fetchReviews(params?: DataCenterQuery) {
  return unwrap<PageResponse<RectificationReview>>(http.get('/api/data-center/reviews', { params }))
}

export function fetchReview(id: number) {
  return unwrap<RectificationReview>(http.get(`/api/data-center/reviews/${encodeURIComponent(String(id))}`))
}

export function generateReview(payload: Record<string, unknown>) {
  return unwrap<RectificationReview>(http.post('/api/data-center/reviews/generate', payload))
}

export function confirmReview(id: number, payload: Record<string, unknown>) {
  return unwrap<RectificationReview>(http.post(`/api/data-center/reviews/${encodeURIComponent(String(id))}/confirm`, payload))
}

export function updateReview(id: number, payload: Record<string, unknown>) {
  return unwrap<RectificationReview>(http.put(`/api/data-center/reviews/${encodeURIComponent(String(id))}`, payload))
}
