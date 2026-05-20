import http, { unwrap } from '@/api/http'
import type { ApiResponse } from '@/api/types'

export interface AiGeoOverview {
  project_count: number
  monitored_query_count: number
  average_visibility_score: number
  citation_count: number
}

export interface AiGeoPage<T> {
  items: T[]
  total: number
  skip: number
  limit: number
}

export async function fetchAiGeoOverview() {
  return unwrap(http.get<ApiResponse<AiGeoOverview>>('/api/ai-geo/overview'))
}
