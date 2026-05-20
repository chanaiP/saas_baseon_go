import axios from 'axios'
import type { AxiosError, InternalAxiosRequestConfig } from 'axios'

import type { ApiResponse } from './types'

/** 当前页是否在 loopback 上打开（未配 API_BASE 时无需区分） */
function pageHostIsLoopback(): boolean {
  if (typeof window === 'undefined') return true
  const h = window.location.hostname
  return h === 'localhost' || h === '127.0.0.1' || h === '[::1]'
}

function apiOriginHostIsLoopback(hostname: string): boolean {
  return hostname === 'localhost' || hostname === '127.0.0.1' || hostname === '[::1]'
}

/**
 * 开发时用局域网 IP 打开 Vite（如 http://192.168.x.x:5173），若 VITE_API_BASE 指向本机 loopback，
 * 浏览器会在「访问者电脑」上找 localhost:8081，导致 /api/* 全部失败（登录页 footer、phone-login-tenants 等）。
 * 此时改为走相对路径，由 Vite 把请求转到运行 dev 的那台机器上的后端。
 */
function ignoreConfiguredLoopbackBaseInDev(): boolean {
  return Boolean(import.meta.env.DEV && typeof window !== 'undefined' && !pageHostIsLoopback())
}

/** 绝对地址只保留 origin，避免误配成 …/api/users 导致 URL 拼错 */
function resolveApiBase(): string {
  const raw = (import.meta.env.VITE_API_BASE || '').trim()
  if (!raw) return ''
  if (/^https?:\/\//i.test(raw)) {
    try {
      const u = new URL(raw)
      if (ignoreConfiguredLoopbackBaseInDev() && apiOriginHostIsLoopback(u.hostname)) {
        return ''
      }
      return `${u.protocol}//${u.host}`
    } catch {
      return ''
    }
  }
  return raw.replace(/\/$/, '')
}

/**
 * 健康检查地址：须与 API 服务同源。
 * - 未配置 VITE_API_BASE 时用相对路径 `/health`（走 Vite 代理或与前端同域反代）。
 * - 配置绝对地址时请求 `{origin}/health`，避免前后端分离时误请求静态站点导致永远「后端不可用」。
 */
export function resolveHealthUrl(): string {
  const raw = (import.meta.env.VITE_API_BASE || '').trim()
  if (!raw) return '/health'
  if (/^https?:\/\//i.test(raw)) {
    try {
      const u = new URL(raw)
      if (ignoreConfiguredLoopbackBaseInDev() && apiOriginHostIsLoopback(u.hostname)) {
        return '/health'
      }
      return `${u.protocol}//${u.host}/health`
    } catch {
      return '/health'
    }
  }
  return '/health'
}

const http = axios.create({
  baseURL: resolveApiBase(),
  /** 慢网络或大体量 profile 时避免误当作「登录失效」 */
  timeout: Number(import.meta.env.VITE_HTTP_TIMEOUT_MS || 60000),
})

/** 解析请求 pathname（不含 query），兼容 baseURL 绝对地址与相对路径 */
function pathnameOnly(config: InternalAxiosRequestConfig): string {
  const url = config.url || ''
  const base = (config.baseURL || '').replace(/\/$/, '')
  const combined = /^https?:\/\//i.test(url)
    ? url
    : `${base}${url.startsWith('/') ? url : `/${url}`}`
  try {
    return new URL(combined, 'http://local.invalid').pathname
  } catch {
    const stripped = combined.replace(/^https?:\/\/[^/]+/i, '')
    return stripped.split('?')[0] || '/'
  }
}

/** 未登录态接口不应携带旧 Bearer（登录页残留 token 会导致链路怪异且不利于排查） */
function shouldOmitAuthorization(config: InternalAxiosRequestConfig): boolean {
  const path = pathnameOnly(config)
  const method = (config.method || 'get').toLowerCase()
  if (path.startsWith('/api/public/')) return true
  if (method === 'post' && path === '/api/auth/login') return true
  if (method === 'post' && path === '/api/auth/register') return true
  if (method === 'get' && path === '/api/auth/captcha') return true
  if (method === 'get' && path === '/api/auth/phone-login-tenants') return true
  return false
}

http.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  if (shouldOmitAuthorization(config)) {
    config.headers.delete('Authorization')
    return config
  }
  const token = localStorage.getItem('access_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

type ErrorPayload = Partial<ApiResponse<unknown>> & {
  detail?: { message?: string } | string | Array<{ msg?: string }>
}

const QUOTA_EXCEEDED_CODE = 200429
const QUOTA_EXCEEDED_MESSAGE = '配额不足，请联系管理员'

function isQuotaExceeded(payload: ErrorPayload | undefined, status: number | undefined, message: string): boolean {
  if (payload?.code === QUOTA_EXCEEDED_CODE) return true
  return status === 429 && /配额|套餐/.test(message)
}

http.interceptors.response.use(
  (res) => res,
  (err: AxiosError<ErrorPayload>) => {
    if (!err.response) {
      const hint =
        err.code === 'ERR_NETWORK' || err.message === 'Network Error'
          ? '无法连接后端：请确认已执行 docker compose up -d，并启动 Go API：DB_AUTO_MIGRATE=false go run ./cmd/api（开发时前端需 npm run dev 以启用 Vite 代理）。'
          : err.message
      return Promise.reject(new Error(hint))
    }
    const payload = err.response?.data
    const d = payload?.detail as unknown
    let msg = err.message
    if (payload?.message) {
      msg = payload.message
    } else if (Array.isArray(d)) {
      const parts = d
        .map((x: { msg?: string }) => (typeof x?.msg === 'string' ? x.msg : ''))
        .filter(Boolean)
      if (parts.length) msg = parts.join('；')
    } else if (typeof d === 'object' && d !== null && 'message' in d) {
      msg = String((d as { message?: string }).message ?? msg)
    } else if (typeof d === 'string') msg = d
    const st = err.response?.status
    if (isQuotaExceeded(payload, st, msg)) {
      return Promise.reject(new Error(QUOTA_EXCEEDED_MESSAGE))
    }
    if (st != null) msg = `${msg}（HTTP ${st}）`
    const out = new Error(msg) as Error & { status?: number }
    if (typeof st === 'number') out.status = st
    return Promise.reject(out)
  },
)

export async function unwrap<T>(p: Promise<{ data: ApiResponse<T> }>): Promise<T> {
  const { data } = await p
  if (data.code !== 0) {
    if (data.code === QUOTA_EXCEEDED_CODE) throw new Error(QUOTA_EXCEEDED_MESSAGE)
    throw new Error(data.message || '请求失败')
  }
  return data.data as T
}

export default http
