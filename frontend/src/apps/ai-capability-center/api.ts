import http, { unwrap } from '@/api/http'
import type { ApiResponse } from '@/api/types'

import type {
  AiModelImportPayload,
  AiModelImportResult,
  AiOverview,
  AiPage,
  AiProviderImportPayload,
  AiProviderImportResult,
  AiResource,
  AiScenarioImportPayload,
  AiScenarioImportResult,
} from './types'

export async function fetchAiOverview() {
  return unwrap(http.get<ApiResponse<AiOverview>>('/api/ai-capability-center/overview'))
}

export async function fetchAiResource<T = Record<string, unknown>>(
  resource: AiResource,
  query: Record<string, string | number | undefined> = {},
) {
  const params: Record<string, string | number> = {
    skip: query.skip ?? 0,
    limit: query.limit ?? 20,
  }
  for (const [key, value] of Object.entries(query)) {
    if (value !== undefined && value !== '') params[key] = value
  }
  return unwrap(http.get<ApiResponse<AiPage<T>>>(`/api/ai-capability-center/${resource}`, { params }))
}

export async function createAiResource<T = Record<string, unknown>>(resource: AiResource, payload: Record<string, unknown>) {
  return unwrap(http.post<ApiResponse<T>>(`/api/ai-capability-center/${resource}`, payload))
}

export async function updateAiResource<T = Record<string, unknown>>(resource: AiResource, id: string, payload: Record<string, unknown>) {
  return unwrap(http.put<ApiResponse<T>>(`/api/ai-capability-center/${resource}/${id}`, payload))
}

export async function deleteAiResource(resource: AiResource, id: string) {
  return unwrap(http.delete<ApiResponse<{ deleted: boolean }>>(`/api/ai-capability-center/${resource}/${id}`))
}

export async function importAiProviders(payload: AiProviderImportPayload) {
  return unwrap(http.post<ApiResponse<AiProviderImportResult>>('/api/ai-capability-center/providers/import', payload))
}

export async function importAiModels(payload: AiModelImportPayload) {
  return unwrap(http.post<ApiResponse<AiModelImportResult>>('/api/ai-capability-center/models/import', payload))
}

export async function importAiScenarios(payload: AiScenarioImportPayload) {
  return unwrap(http.post<ApiResponse<AiScenarioImportResult>>('/api/ai-capability-center/scenarios/import', payload))
}
