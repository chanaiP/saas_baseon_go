import http, { unwrap } from './http'
import type { ApiResponse, Paginated } from './types'

export interface Plan {
  id: number
  plan_code: string
  plan_name: string
  plan_type: string
  billing_cycle: string
  price: number | string
  status: number
  is_default: boolean
  sort_order: number
  description?: string | null
  created_at?: string | null
  updated_at?: string | null
}

export interface PlanPayload {
  plan_code: string
  plan_name: string
  plan_type: string
  billing_cycle: string
  price: number
  status: number
  is_default: boolean
  sort_order: number
  description?: string | null
}

export interface PlanCopyPayload {
  plan_code: string
  plan_name: string
  description?: string | null
}

export interface Feature {
  id: number
  feature_code: string
  feature_name: string
  feature_type: string
  app_code?: string | null
  parent_id: number
  menu_id?: number | null
  api_method?: string | null
  api_path?: string | null
  service_key?: string | null
  status: number
  description?: string | null
}

export interface FeaturePayload {
  feature_code: string
  feature_name: string
  feature_type: string
  app_code?: string | null
  parent_id: number
  menu_id?: number | null
  api_method?: string | null
  api_path?: string | null
  service_key?: string | null
  status: number
  description?: string | null
}

export interface Quota {
  id: number
  quota_code: string
  quota_name: string
  quota_type: string
  period_type?: string | null
  unit?: string | null
  status: number
  description?: string | null
}

export interface QuotaPayload {
  quota_code: string
  quota_name: string
  quota_type: string
  period_type?: string | null
  unit?: string | null
  status: number
  description?: string | null
}

export interface PlanFeatureData {
  plan_id: number
  feature_ids: number[]
}

export interface PlanQuotaValue {
  quota_id: number
  quota_code: string
  quota_name: string
  quota_value: number
  period_type?: string | null
  unit?: string | null
}

export interface PlanQuotaData {
  plan_id: number
  quotas: PlanQuotaValue[]
}

export interface TenantSubscriptionPayload {
  plan_id: number
  subscription_status: string
  start_time: string
  end_time?: string | null
  trial_end_time?: string | null
  auto_renew: boolean
  frozen_reason?: string | null
}

export interface TenantSubscription extends TenantSubscriptionPayload {
  id: number
  tenant_id: number
  created_at?: string | null
  updated_at?: string | null
}

export interface TenantFeatureOverride {
  id: number
  tenant_id: number
  feature_id: number
  feature_code: string
  feature_name: string
  enabled: boolean
  reason?: string | null
  start_time?: string | null
  end_time?: string | null
}

export interface TenantQuotaOverride {
  id: number
  tenant_id: number
  quota_id: number
  quota_code: string
  quota_name: string
  quota_value: number
  period_type?: string | null
  unit?: string | null
  reason?: string | null
  start_time?: string | null
  end_time?: string | null
}

export interface TenantQuotaUsage {
  quota_id: number
  quota_code: string
  quota_name: string
  quota_type: string
  period_type?: string | null
  period_key: string
  unit?: string | null
  used_value: number
  limit_value: number
  remaining_value?: number | null
}

export interface CapabilityQuotaValue {
  quota_id: number
  quota_code: string
  quota_name: string
  quota_value: number
  period_type?: string | null
  unit?: string | null
}

export type CapabilityCellState = 'enabled' | 'partial' | 'disabled'

export interface PlanCapabilityCell {
  plan_id: number
  plan_code: string
  enabled: boolean
  state: CapabilityCellState
  feature_ids: number[]
  quota_values: CapabilityQuotaValue[]
}

export interface PlanCapabilityNode {
  id: string
  label: string
  node_type: 'app' | 'domain' | 'group' | 'feature'
  feature_id?: number | null
  feature_code?: string | null
  feature_type?: string | null
  app_code?: string | null
  description?: string | null
  children: PlanCapabilityNode[]
  cells: PlanCapabilityCell[]
}

export interface PlanCapabilityMatrixData {
  plans: Plan[]
  nodes: PlanCapabilityNode[]
}

export interface PlanCapabilitySaveData {
  plan_id: number
  feature_ids: number[]
  quotas: PlanQuotaValue[]
}

export async function fetchPlans(skip = 0, limit = 50, keyword = '') {
  return unwrap(http.get<ApiResponse<Paginated<Plan>>>('/api/plans', { params: { skip, limit, keyword } }))
}

export async function fetchPlanMatrix() {
  return unwrap(http.get<ApiResponse<PlanCapabilityMatrixData>>('/api/plans/matrix'))
}

export async function createPlan(body: PlanPayload) {
  return unwrap(http.post<ApiResponse<Plan>>('/api/plans', body))
}

export async function updatePlan(id: number, body: Partial<PlanPayload>) {
  return unwrap(http.put<ApiResponse<Plan>>(`/api/plans/${id}`, body))
}

export async function copyPlan(id: number, body: PlanCopyPayload) {
  return unwrap(http.post<ApiResponse<Plan>>(`/api/plans/${id}/copy`, body))
}

export async function deletePlan(id: number) {
  return unwrap(http.delete<ApiResponse<{ id: number }>>(`/api/plans/${id}`))
}

export async function fetchFeatures(skip = 0, limit = 200, keyword = '') {
  return unwrap(http.get<ApiResponse<Paginated<Feature>>>('/api/plans/features', { params: { skip, limit, keyword } }))
}

export async function createFeature(body: FeaturePayload) {
  return unwrap(http.post<ApiResponse<Feature>>('/api/plans/features', body))
}

export async function updateFeature(id: number, body: Partial<FeaturePayload>) {
  return unwrap(http.put<ApiResponse<Feature>>(`/api/plans/features/${id}`, body))
}

export async function fetchPlanFeatures(planId: number) {
  return unwrap(http.get<ApiResponse<PlanFeatureData>>(`/api/plans/${planId}/features`))
}

export async function savePlanFeatures(planId: number, featureIds: number[]) {
  return unwrap(http.put<ApiResponse<PlanFeatureData>>(`/api/plans/${planId}/features`, { feature_ids: featureIds }))
}

export async function savePlanCapabilities(
  planId: number,
  body: { feature_ids: number[]; quotas: Array<{ quota_id: number; quota_value: number }> },
) {
  return unwrap(http.put<ApiResponse<PlanCapabilitySaveData>>(`/api/plans/${planId}/capabilities`, body))
}

export async function fetchQuotas(skip = 0, limit = 200, keyword = '') {
  return unwrap(http.get<ApiResponse<Paginated<Quota>>>('/api/plans/quotas', { params: { skip, limit, keyword } }))
}

export async function createQuota(body: QuotaPayload) {
  return unwrap(http.post<ApiResponse<Quota>>('/api/plans/quotas', body))
}

export async function updateQuota(id: number, body: Partial<QuotaPayload>) {
  return unwrap(http.put<ApiResponse<Quota>>(`/api/plans/quotas/${id}`, body))
}

export async function fetchPlanQuotas(planId: number) {
  return unwrap(http.get<ApiResponse<PlanQuotaData>>(`/api/plans/${planId}/quotas`))
}

export async function savePlanQuotas(planId: number, quotas: Array<{ quota_id: number; quota_value: number }>) {
  return unwrap(http.put<ApiResponse<PlanQuotaData>>(`/api/plans/${planId}/quotas`, { quotas }))
}

export async function fetchTenantSubscription(tenantId: number) {
  return unwrap(http.get<ApiResponse<TenantSubscription>>(`/api/tenants/${tenantId}/subscription`))
}

export async function saveTenantSubscription(tenantId: number, body: TenantSubscriptionPayload) {
  return unwrap(http.put<ApiResponse<TenantSubscription>>(`/api/tenants/${tenantId}/subscription`, body))
}

export async function fetchTenantFeatureOverrides(tenantId: number) {
  return unwrap(http.get<ApiResponse<{ tenant_id: number; overrides: TenantFeatureOverride[] }>>(`/api/tenants/${tenantId}/feature-overrides`))
}

export async function saveTenantFeatureOverrides(
  tenantId: number,
  overrides: Array<{ feature_id: number; enabled: boolean; reason?: string | null }>,
) {
  return unwrap(http.put<ApiResponse<{ tenant_id: number; overrides: TenantFeatureOverride[] }>>(
    `/api/tenants/${tenantId}/feature-overrides`,
    { overrides },
  ))
}

export async function fetchTenantQuotaOverrides(tenantId: number) {
  return unwrap(http.get<ApiResponse<{ tenant_id: number; overrides: TenantQuotaOverride[] }>>(`/api/tenants/${tenantId}/quota-overrides`))
}

export async function saveTenantQuotaOverrides(
  tenantId: number,
  overrides: Array<{ quota_id: number; quota_value: number; reason?: string | null }>,
) {
  return unwrap(http.put<ApiResponse<{ tenant_id: number; overrides: TenantQuotaOverride[] }>>(
    `/api/tenants/${tenantId}/quota-overrides`,
    { overrides },
  ))
}

export async function fetchTenantQuotaUsage(tenantId: number) {
  return unwrap(http.get<ApiResponse<{ tenant_id: number; usages: TenantQuotaUsage[] }>>(`/api/tenants/${tenantId}/quota-usage`))
}
