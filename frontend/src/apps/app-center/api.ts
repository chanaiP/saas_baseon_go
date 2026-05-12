import http, { unwrap } from '@/api/http'
import type { ApiResponse } from '@/api/types'

import type { AppCenterApp, AppCenterCreatePayload, AppCenterListQuery, AppCenterListResponse, AppCenterStats, AppCenterUpdatePayload, AppManifestParseResult } from './types'

export async function fetchAppCenterApps(query: AppCenterListQuery = {}) {
  const params: Record<string, string | number> = {
    skip: query.skip ?? 0,
    limit: query.limit ?? 20,
  }
  const keyword = (query.keyword ?? '').trim()
  if (keyword) params.keyword = keyword
  if (query.type) params.type = query.type
  if (query.status) params.status = query.status
  if (query.source) params.source = query.source
  return unwrap(http.get<ApiResponse<AppCenterListResponse>>('/api/apps', { params }))
}

export async function fetchAppCenterStats() {
  return unwrap(http.get<ApiResponse<AppCenterStats>>('/api/apps/stats'))
}

export async function fetchAppCenterApp(id: number) {
  return unwrap(http.get<ApiResponse<AppCenterApp>>(`/api/apps/${id}`))
}

export async function createAppCenterApp(payload: AppCenterCreatePayload) {
  return unwrap(http.post<ApiResponse<AppCenterApp>>('/api/apps', payload))
}

export async function updateAppCenterApp(id: number, payload: AppCenterUpdatePayload) {
  return unwrap(http.put<ApiResponse<AppCenterApp>>(`/api/apps/${id}`, payload))
}

export async function updateAppCenterAppStatus(id: number, status: string) {
  return unwrap(http.patch<ApiResponse<AppCenterApp>>(`/api/apps/${id}/status`, { status }))
}

export async function downloadAppManifestTemplate() {
  const response = await http.get('/api/apps/manifest/template', { responseType: 'blob' })
  return response.data as Blob
}

export async function parseAppManifestFile(file: File) {
  const form = new FormData()
  form.append('file', file)
  return unwrap(http.post<ApiResponse<AppManifestParseResult>>('/api/apps/manifest/parse', form, {
    headers: { 'Content-Type': 'multipart/form-data' },
  }))
}
