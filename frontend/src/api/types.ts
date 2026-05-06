export interface ApiResponse<T> {
  code: number
  message: string
  data: T | null
}

export interface Paginated<T> {
  items: T[]
  total: number
  skip: number
  limit: number
}
