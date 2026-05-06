import http, { unwrap } from './http'
import type { ApiResponse } from './types'
import type { Paginated } from './types'

export interface PositionTypeRow {
  id: number
  name: string
  code: string
  position_count: number
}

export interface PositionRow {
  id: number
  position_type_id: number
  name: string
  code: string
}

export async function fetchPositionTypes(skip = 0, limit = 50, keyword?: string) {
  const params: Record<string, unknown> = { skip, limit }
  if (keyword != null && String(keyword).trim() !== '') params.keyword = String(keyword).trim()
  return unwrap(http.get<ApiResponse<Paginated<PositionTypeRow>>>('/api/position-types', { params }))
}

export async function createPositionType(body: { name: string; code: string }) {
  return unwrap(http.post<ApiResponse<{ id: number }>>('/api/position-types', body))
}

export async function updatePositionType(id: number, body: Partial<{ name: string; code: string }>) {
  return unwrap(http.put<ApiResponse<{ id: number }>>(`/api/position-types/${id}`, body))
}

export async function deletePositionType(id: number) {
  return unwrap(http.delete<ApiResponse<{ deleted: number }>>(`/api/position-types/${id}`))
}

export async function fetchPositions(params: {
  skip?: number
  limit?: number
  position_type_id?: number
  keyword?: string
}) {
  const p: Record<string, unknown> = { ...params }
  if (p.keyword != null && String(p.keyword).trim() === '') delete p.keyword
  return unwrap(http.get<ApiResponse<Paginated<PositionRow>>>('/api/positions', { params: p }))
}

export async function createPosition(body: { position_type_id: number; name: string; code: string }) {
  return unwrap(http.post<ApiResponse<{ id: number }>>('/api/positions', body))
}

export async function updatePosition(id: number, body: Partial<{ name: string; code: string; position_type_id: number }>) {
  return unwrap(http.put<ApiResponse<{ id: number }>>(`/api/positions/${id}`, body))
}

export async function deletePosition(id: number) {
  return unwrap(http.delete<ApiResponse<{ deleted: number }>>(`/api/positions/${id}`))
}
