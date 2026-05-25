import http, { unwrap } from '@/api/http'
import type { ApiResponse } from '@/api/types'

export interface AiGeoPage<T> {
  items: T[]
  total: number
  skip: number
  limit: number
}

export interface AiGeoOverview {
  brand_count: number
  product_count: number
  sku_count: number
  channel_count: number
  channel_account_count: number
  draft_count_today: number
  pending_draft_count: number
  channel_content_count: number
  publish_plan_today: number
  average_completeness: number
  pending_tasks: Array<{ tag: string; title: string }>
  quota_usage: Record<string, unknown>
}

export interface AiGeoBrand {
  id: number
  tenant_id: number
  brand_code: string
  brand_name: string
  positioning?: string | null
  target_audience?: string | null
  price_band?: string | null
  tone?: string | null
  keywords: string
  completeness: number
  status: string
}

export interface AiGeoProduct {
  id: number
  tenant_id: number
  brand_id: number
  product_code: string
  product_name: string
  category_name?: string | null
  selling_points: string
  faq: string
  content_angles: string
  completeness: number
  status: string
}

export interface AiGeoSKU {
  id: number
  tenant_id: number
  product_id: number
  sku_code: string
  sku_name: string
  attributes: string
  price: number
  image_url?: string | null
  stock_status: string
  status: string
}

export interface AiGeoCompetitor {
  id: number
  tenant_id: number
  product_id: number
  brand_name: string
  product_name: string
  price_text?: string | null
  point?: string | null
  difference?: string | null
  angle?: string | null
  link_url?: string | null
  status: string
}

export interface AiGeoKeyword {
  id: number
  tenant_id: number
  brand_id?: number | null
  product_id?: number | null
  keyword_group: string
  keyword: string
  intent?: string | null
  source: string
  weight: number
  status: string
}

export interface AiGeoMaterialAsset {
  id: number
  tenant_id: number
  brand_id?: number | null
  product_id?: number | null
  asset_type: string
  asset_name: string
  url?: string | null
  metadata: string
  status: string
}

export interface AiGeoHotspot {
  id: number
  tenant_id: number
  source_id?: number | null
  platform: string
  title: string
  heat_score: number
  source_url?: string | null
  captured_at: string
  metadata?: string | Record<string, unknown> | null
  status: string
}

export interface AiGeoExternalSource {
  id: number
  source_type: string
  source_url: string
  source_site?: string | null
  source_title?: string | null
  raw_text: string
  clean_text: string
  content_hash: string
  extracted_meta?: string | Record<string, unknown> | null
  extraction_status: string
  extraction_error?: string | null
  captured_at: string
  status: string
}

export interface AiGeoStyleTemplate {
  id: number
  source_id?: number | null
  template_code: string
  template_name: string
  description?: string | null
  content_type?: string | null
  platform?: string | null
  tone_profile: string | Record<string, unknown>
  structure_profile: string | Record<string, unknown>
  technique_profile: string | Record<string, unknown>
  style_keywords: string | string[]
  prompt_fragment: string
  negative_rules: string | string[]
  extraction_summary: string | Record<string, unknown>
  status: string
  created_at?: string
  updated_at?: string
}

export interface AiGeoExternalExtractResult {
  source: AiGeoExternalSource
  style_template?: AiGeoStyleTemplate
  hotspot_draft?: Record<string, unknown>
}

export interface AiGeoChannel {
  id: number
  tenant_id: number
  channel_code: string
  channel_name: string
  channel_type: string
  entry_url?: string | null
  content_forms: string
  support_modes: string
  default_publish_mode: string
  status: string
}

export interface AiGeoDraft {
  id: number
  tenant_id: number
  draft_code: string
  brand_id?: number | null
  product_id?: number | null
  title: string
  summary?: string | null
  body: string
  keywords: string
  conversation?: string
  source_snapshot?: string
  source: string
  audit_status: string
  channel_status: string
  status: string
  created_at?: string
  updated_at?: string
}

export interface AiGeoGenerateDraftPayload {
  brand_id?: number
  product_id?: number
  skill?: string
  content_type?: string
  hotspot_id?: number
  style_template_id?: number
  prompt: string
  conversation?: Array<Record<string, unknown>>
  source_snapshot?: Record<string, unknown>
}

export interface AiGeoChannelContent {
  id: number
  tenant_id: number
  draft_id: number
  channel_id: number
  title: string
  body: string
  audit_status: string
  publish_status: string
  status: string
  created_at?: string
  updated_at?: string
}

export interface AiGeoPublishPlan {
  id: number
  tenant_id: number
  plan_code: string
  channel_content_id: number
  channel_id: number
  scheduled_at: string
  publish_method: string
  automation_level: string
  status: string
  published_url?: string | null
  fail_reason?: string | null
  created_at?: string
  updated_at?: string
}

export interface AiGeoPublishPlanCalendarDay {
  date: string
  total: number
  scheduled: number
  publishing: number
  published: number
  failed: number
  cancelled: number
  items: AiGeoPublishPlan[]
}

export interface AiGeoPublishPlanCalendar {
  start_date: string
  end_date: string
  days: AiGeoPublishPlanCalendarDay[]
}

export interface AiGeoAuditSuggestion {
  id: number
  object_type?: string
  object_id?: number
  object_code?: string | null
  scenario_code?: string
  risk_level?: string
  passed?: boolean
  summary?: string | null
  suggestion_json?: string
  model_code?: string | null
  status?: string
  error_message?: string | null
  generated_at?: string
}

export interface AiGeoListParams {
  skip?: number
  limit?: number
  keyword?: string
  status?: string
  brand_id?: number
  product_id?: number
  channel_id?: number
  audit_status?: string
  start_date?: string
  end_date?: string
}

export async function fetchAiGeoOverview() {
  return unwrap(http.get<ApiResponse<AiGeoOverview>>('/api/ai-geo/overview'))
}

export async function fetchAiGeoBrands(params: AiGeoListParams = {}) {
  return unwrap(http.get<ApiResponse<AiGeoPage<AiGeoBrand>>>('/api/ai-geo/materials/brands', { params }))
}

export async function createAiGeoBrand(payload: Partial<AiGeoBrand> & { keywords?: string[] }) {
  return unwrap(http.post<ApiResponse<AiGeoBrand>>('/api/ai-geo/materials/brands', payload))
}

export async function updateAiGeoBrand(id: number, payload: Partial<AiGeoBrand> & { keywords?: string[] }) {
  return unwrap(http.put<ApiResponse<AiGeoBrand>>(`/api/ai-geo/materials/brands/${id}`, payload))
}

export async function archiveAiGeoBrand(id: number) {
  return unwrap(http.delete<ApiResponse<AiGeoBrand>>(`/api/ai-geo/materials/brands/${id}`))
}

export async function fetchAiGeoProducts(params: AiGeoListParams = {}) {
  return unwrap(http.get<ApiResponse<AiGeoPage<AiGeoProduct>>>('/api/ai-geo/materials/products', { params }))
}

export async function createAiGeoProduct(payload: Record<string, unknown>) {
  return unwrap(http.post<ApiResponse<AiGeoProduct>>('/api/ai-geo/materials/products', payload))
}

export async function updateAiGeoProduct(id: number, payload: Record<string, unknown>) {
  return unwrap(http.put<ApiResponse<AiGeoProduct>>(`/api/ai-geo/materials/products/${id}`, payload))
}

export async function archiveAiGeoProduct(id: number) {
  return unwrap(http.delete<ApiResponse<AiGeoProduct>>(`/api/ai-geo/materials/products/${id}`))
}

export async function fetchAiGeoSKUs(params: AiGeoListParams = {}) {
  return unwrap(http.get<ApiResponse<AiGeoPage<AiGeoSKU>>>('/api/ai-geo/materials/skus', { params }))
}

export async function createAiGeoSKU(payload: Record<string, unknown>) {
  return unwrap(http.post<ApiResponse<AiGeoSKU>>('/api/ai-geo/materials/skus', payload))
}

export async function updateAiGeoSKU(id: number, payload: Record<string, unknown>) {
  return unwrap(http.put<ApiResponse<AiGeoSKU>>(`/api/ai-geo/materials/skus/${id}`, payload))
}

export async function archiveAiGeoSKU(id: number) {
  return unwrap(http.delete<ApiResponse<AiGeoSKU>>(`/api/ai-geo/materials/skus/${id}`))
}

export async function fetchAiGeoCompetitors(params: AiGeoListParams = {}) {
  return unwrap(http.get<ApiResponse<AiGeoPage<AiGeoCompetitor>>>('/api/ai-geo/materials/competitors', { params }))
}

export async function createAiGeoCompetitor(payload: Record<string, unknown>) {
  return unwrap(http.post<ApiResponse<AiGeoCompetitor>>('/api/ai-geo/materials/competitors', payload))
}

export async function updateAiGeoCompetitor(id: number, payload: Record<string, unknown>) {
  return unwrap(http.put<ApiResponse<AiGeoCompetitor>>(`/api/ai-geo/materials/competitors/${id}`, payload))
}

export async function archiveAiGeoCompetitor(id: number) {
  return unwrap(http.delete<ApiResponse<AiGeoCompetitor>>(`/api/ai-geo/materials/competitors/${id}`))
}

export async function fetchAiGeoKeywords(params: AiGeoListParams = {}) {
  return unwrap(http.get<ApiResponse<AiGeoPage<AiGeoKeyword>>>('/api/ai-geo/materials/keywords', { params }))
}

export async function fetchAiGeoKeyword(id: number) {
  return unwrap(http.get<ApiResponse<AiGeoKeyword>>(`/api/ai-geo/materials/keywords/${id}`))
}

export async function createAiGeoKeyword(payload: Record<string, unknown>) {
  return unwrap(http.post<ApiResponse<AiGeoKeyword>>('/api/ai-geo/materials/keywords', payload))
}

export async function updateAiGeoKeyword(id: number, payload: Record<string, unknown>) {
  return unwrap(http.put<ApiResponse<AiGeoKeyword>>(`/api/ai-geo/materials/keywords/${id}`, payload))
}

export async function archiveAiGeoKeyword(id: number) {
  return unwrap(http.delete<ApiResponse<AiGeoKeyword>>(`/api/ai-geo/materials/keywords/${id}`))
}

export async function fetchAiGeoMaterialAssets(params: AiGeoListParams = {}) {
  return unwrap(http.get<ApiResponse<AiGeoPage<AiGeoMaterialAsset>>>('/api/ai-geo/materials/assets', { params }))
}

export async function fetchAiGeoMaterialAsset(id: number) {
  return unwrap(http.get<ApiResponse<AiGeoMaterialAsset>>(`/api/ai-geo/materials/assets/${id}`))
}

export async function createAiGeoMaterialAsset(payload: Record<string, unknown>) {
  return unwrap(http.post<ApiResponse<AiGeoMaterialAsset>>('/api/ai-geo/materials/assets', payload))
}

export async function updateAiGeoMaterialAsset(id: number, payload: Record<string, unknown>) {
  return unwrap(http.put<ApiResponse<AiGeoMaterialAsset>>(`/api/ai-geo/materials/assets/${id}`, payload))
}

export async function archiveAiGeoMaterialAsset(id: number) {
  return unwrap(http.delete<ApiResponse<AiGeoMaterialAsset>>(`/api/ai-geo/materials/assets/${id}`))
}

export async function fetchAiGeoHotspots(params: AiGeoListParams = {}) {
  return unwrap(http.get<ApiResponse<AiGeoPage<AiGeoHotspot>>>('/api/ai-geo/materials/hotspots', { params }))
}

export async function fetchAiGeoHotspot(id: number) {
  return unwrap(http.get<ApiResponse<AiGeoHotspot>>(`/api/ai-geo/materials/hotspots/${id}`))
}

export async function createAiGeoHotspot(payload: Record<string, unknown>) {
  return unwrap(http.post<ApiResponse<AiGeoHotspot>>('/api/ai-geo/materials/hotspots', payload))
}

export async function updateAiGeoHotspot(id: number, payload: Record<string, unknown>) {
  return unwrap(http.put<ApiResponse<AiGeoHotspot>>(`/api/ai-geo/materials/hotspots/${id}`, payload))
}

export async function archiveAiGeoHotspot(id: number) {
  return unwrap(http.delete<ApiResponse<AiGeoHotspot>>(`/api/ai-geo/materials/hotspots/${id}`))
}

export async function extractAiGeoExternalSource(payload: Record<string, unknown>) {
  return unwrap(http.post<ApiResponse<AiGeoExternalExtractResult>>('/api/ai-geo/external-sources/extract', payload))
}

export async function fetchAiGeoStyleTemplates(params: AiGeoListParams = {}) {
  return unwrap(http.get<ApiResponse<AiGeoPage<AiGeoStyleTemplate>>>('/api/ai-geo/style-templates', { params }))
}

export async function fetchAiGeoStyleTemplate(id: number) {
  return unwrap(http.get<ApiResponse<AiGeoStyleTemplate>>(`/api/ai-geo/style-templates/${id}`))
}

export async function createAiGeoStyleTemplate(payload: Record<string, unknown>) {
  return unwrap(http.post<ApiResponse<AiGeoStyleTemplate>>('/api/ai-geo/style-templates', payload))
}

export async function updateAiGeoStyleTemplate(id: number, payload: Record<string, unknown>) {
  return unwrap(http.put<ApiResponse<AiGeoStyleTemplate>>(`/api/ai-geo/style-templates/${id}`, payload))
}

export async function archiveAiGeoStyleTemplate(id: number) {
  return unwrap(http.delete<ApiResponse<AiGeoStyleTemplate>>(`/api/ai-geo/style-templates/${id}`))
}

export async function importAiGeoMaterials(payload: Record<string, unknown>) {
  return unwrap(http.post<ApiResponse<Record<string, unknown>>>('/api/ai-geo/materials/imports', payload))
}

export async function fetchAiGeoImportErrors(id: number, params: AiGeoListParams = {}) {
  return unwrap(http.get<ApiResponse<AiGeoPage<Record<string, unknown>>>>(`/api/ai-geo/materials/imports/${id}/errors`, { params }))
}

export async function generateAiGeoDraft(payload: AiGeoGenerateDraftPayload) {
  return unwrap(http.post<ApiResponse<AiGeoDraft>>('/api/ai-geo/workbench/drafts/generate', payload))
}

export type AiGatewayStreamEvent = {
  type: 'meta' | 'delta' | 'final' | 'error' | string
  delta?: string
  text?: string
  status?: string
  error_code?: string
  error_message?: string
  data?: Record<string, unknown>
}

function normalizeAiGatewayStreamError(error: unknown) {
  const message = error instanceof Error ? error.message : String(error || '')
  if (/(load|fetch|network).{0,20}(fail|failed|failure|error)|(fail|failed|failure|error).{0,20}(load|fetch|network)|networkerror|aborterror/i.test(message)) {
    return new Error('AI 专家服务连接失败，请检查网络或稍后重试')
  }
  return error instanceof Error ? error : new Error(message || 'AI Gateway 流式调用失败')
}

export async function streamAiGeoGatewayInvoke(
  payload: Record<string, unknown>,
  handlers: {
    onMeta?: (event: AiGatewayStreamEvent) => void
    onDelta?: (delta: string, event: AiGatewayStreamEvent) => void
    onFinal?: (event: AiGatewayStreamEvent) => void
    onError?: (event: AiGatewayStreamEvent) => void
  } = {},
) {
  const baseURL = String(http.defaults.baseURL || '').replace(/\/$/, '')
  const token = localStorage.getItem('access_token')
  const scenarioCode = String(payload.ai_scenario_code || '')
  const endpointByScenario: Record<string, string> = {
    ai_geo_draft_generation: '/api/ai-geo/workbench/ai-stream',
    ai_geo_channel_content_editor: '/api/ai-geo/channel-contents/ai-editor-stream',
  }
  const endpoint = endpointByScenario[scenarioCode]
  if (!endpoint) {
    throw new Error('AI GEO 场景未纳入应用权限边界')
  }
  try {
    const response = await fetch(`${baseURL}${endpoint}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
      body: JSON.stringify(payload),
    })
    if (!response.ok || !response.body) {
      throw new Error(`AI Gateway 流式调用失败（HTTP ${response.status}）`)
    }
    const reader = response.body.getReader()
    const decoder = new TextDecoder('utf-8')
    let buffer = ''
    const consumeEvent = (raw: string) => {
      const dataLine = raw.split('\n').find(line => line.startsWith('data:'))
      if (!dataLine) return
      const data = dataLine.replace(/^data:\s*/, '').trim()
      if (!data || data === '[DONE]') return
      let event: AiGatewayStreamEvent
      try {
        event = JSON.parse(data) as AiGatewayStreamEvent
      } catch {
        handlers.onError?.({
          type: 'error',
          error_code: 'STREAM_PARSE_ERROR',
          error_message: 'AI 专家服务返回数据解析失败，请稍后重试',
        })
        return
      }
      if (event.type === 'meta') handlers.onMeta?.(event)
      if (event.type === 'delta' && event.delta) handlers.onDelta?.(event.delta, event)
      if (event.type === 'final') handlers.onFinal?.(event)
      if (event.type === 'error') handlers.onError?.(event)
    }
    while (true) {
      const { value, done } = await reader.read()
      buffer += decoder.decode(value || new Uint8Array(), { stream: !done })
      const parts = buffer.split('\n\n')
      buffer = parts.pop() || ''
      parts.forEach(consumeEvent)
      if (done) break
    }
    if (buffer.trim()) consumeEvent(buffer)
  } catch (error) {
    throw normalizeAiGatewayStreamError(error)
  }
}

export async function fetchAiGeoDrafts(params: AiGeoListParams = {}) {
  return unwrap(http.get<ApiResponse<AiGeoPage<AiGeoDraft>>>('/api/ai-geo/drafts', { params }))
}

export async function createAiGeoDraft(payload: Record<string, unknown>) {
  return unwrap(http.post<ApiResponse<AiGeoDraft>>('/api/ai-geo/drafts', payload))
}

export async function updateAiGeoDraft(id: number, payload: Record<string, unknown>) {
  return unwrap(http.put<ApiResponse<AiGeoDraft>>(`/api/ai-geo/drafts/${id}`, payload))
}

export async function generateAiGeoChannelContent(id: number, payload: Record<string, unknown>) {
  return unwrap(http.post<ApiResponse<AiGeoChannelContent>>(`/api/ai-geo/drafts/${id}/channel-contents`, payload))
}

export async function fetchAiGeoChannelContents(params: AiGeoListParams = {}) {
  return unwrap(http.get<ApiResponse<AiGeoPage<AiGeoChannelContent>>>('/api/ai-geo/channel-contents', { params }))
}

export async function updateAiGeoChannelContent(id: number, payload: Record<string, unknown>) {
  return unwrap(http.put<ApiResponse<AiGeoChannelContent>>(`/api/ai-geo/channel-contents/${id}`, payload))
}

export async function fetchAiGeoPublishPlans(params: AiGeoListParams = {}) {
  return unwrap(http.get<ApiResponse<AiGeoPage<AiGeoPublishPlan>>>('/api/ai-geo/publish-plans', { params }))
}

export async function fetchAiGeoPublishPlanCalendar(params: Pick<AiGeoListParams, 'start_date' | 'end_date'> = {}) {
  return unwrap(http.get<ApiResponse<AiGeoPublishPlanCalendar>>('/api/ai-geo/publish-plans/calendar', { params }))
}

export async function createAiGeoPublishPlan(payload: Record<string, unknown>) {
  return unwrap(http.post<ApiResponse<AiGeoPublishPlan>>('/api/ai-geo/publish-plans', payload))
}

export async function updateAiGeoPublishPlan(id: number, payload: Record<string, unknown>) {
  return unwrap(http.put<ApiResponse<AiGeoPublishPlan>>(`/api/ai-geo/publish-plans/${id}`, payload))
}

export async function updateAiGeoPublishPlanStatus(id: number, payload: Record<string, unknown>) {
  return unwrap(http.patch<ApiResponse<AiGeoPublishPlan>>(`/api/ai-geo/publish-plans/${id}/status`, payload))
}

export async function fetchAiGeoChannels(params: AiGeoListParams = {}) {
  return unwrap(http.get<ApiResponse<AiGeoPage<AiGeoChannel>>>('/api/ai-geo/channels', { params }))
}

export async function createAiGeoChannel(payload: Record<string, unknown>) {
  return unwrap(http.post<ApiResponse<AiGeoChannel>>('/api/ai-geo/channels', payload))
}

export async function updateAiGeoChannel(id: number, payload: Record<string, unknown>) {
  return unwrap(http.put<ApiResponse<AiGeoChannel>>(`/api/ai-geo/channels/${id}`, payload))
}

export async function testAiGeoChannel(id: number) {
  return unwrap(http.post<ApiResponse<Record<string, unknown>>>(`/api/ai-geo/channels/${id}/test`, {}))
}

export async function fetchAiGeoChannelAccounts(params: AiGeoListParams = {}) {
  return unwrap(http.get<ApiResponse<AiGeoPage<Record<string, unknown>>>>('/api/ai-geo/channel-accounts', { params }))
}

export async function createAiGeoChannelAccount(payload: Record<string, unknown>) {
  return unwrap(http.post<ApiResponse<Record<string, unknown>>>('/api/ai-geo/channel-accounts', payload))
}
