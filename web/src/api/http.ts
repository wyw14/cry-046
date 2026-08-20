import axios, { AxiosError, AxiosInstance } from 'axios'
import type { ErrorEnvelope } from '@/types'

const baseURL = '/api/v1'

// http 是项目内的 axios 单例, 统一处理 Authorization, X-Tenant-ID,
// X-Request-Id 与错误包络.
export const http: AxiosInstance = axios.create({
  baseURL,
  timeout: 15000,
})

http.interceptors.request.use((cfg) => {
  const token = localStorage.getItem('wsr.token') || ''
  const tenant = localStorage.getItem('wsr.tenant') || 'default'
  if (token) {
    cfg.headers.Authorization = `Bearer ${token}`
  }
  cfg.headers['X-Tenant-ID'] = tenant
  cfg.headers['X-Request-Id'] =
    cfg.headers['X-Request-Id'] || cryptoRandomId()
  return cfg
})

http.interceptors.response.use(
  (resp) => resp,
  (err: AxiosError<ErrorEnvelope>) => {
    const env = err.response?.data
    const message = env?.message || err.message || '请求失败'
    return Promise.reject(new ApiError(message, env?.code || 'UNKNOWN', env?.fields || [], err))
  },
)

export class ApiError extends Error {
  constructor(
    message: string,
    public code: string,
    public fields: { field: string; message: string }[],
    public cause: unknown,
  ) {
    super(message)
    this.name = 'ApiError'
  }
}

function cryptoRandomId(): string {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return crypto.randomUUID()
  }
  return 'rid-' + Math.random().toString(36).slice(2) + Date.now().toString(36)
}

export function setToken(token: string) {
  localStorage.setItem('wsr.token', token)
}

export function setTenant(tenant: string) {
  localStorage.setItem('wsr.tenant', tenant)
}

export function getToken(): string {
  return localStorage.getItem('wsr.token') || ''
}

export function getTenant(): string {
  return localStorage.getItem('wsr.tenant') || 'default'
}
