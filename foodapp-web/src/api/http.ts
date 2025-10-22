// src/api/http.ts
// src/api/http.ts
import axios, { AxiosError } from 'axios'
import type { AxiosInstance, AxiosResponse } from 'axios'

// 所有业务请求都以 /api/v1 开头
const http: AxiosInstance = axios.create({ baseURL: '/api/v1', timeout: 15000 })

// 读取/保存 token 的工具
const TOKEN_KEY = 'auth_token'
const EXP_AT_KEY = 'auth_exp_at' // 毫秒时间戳

export function setAuth(token: string, expiresInSec: number) {
  localStorage.setItem(TOKEN_KEY, token)
  const expAt = Date.now() + expiresInSec * 1000 - 5000 // 提前5秒过期
  localStorage.setItem(EXP_AT_KEY, String(expAt))
}

export function getToken(): string | null {
  const t = localStorage.getItem(TOKEN_KEY)
  const exp = Number(localStorage.getItem(EXP_AT_KEY) || 0)
  if (!t || !exp || Date.now() >= exp) {
    // 过期/不存在
    localStorage.removeItem(TOKEN_KEY)
    localStorage.removeItem(EXP_AT_KEY)
    return null
  }
  return t
}

export function clearAuth() {
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(EXP_AT_KEY)
}

// ---- 工具：抽取后端 {error, details} 并拼装消息 ----
type BackendErr = { error?: string; details?: string; [k: string]: any }
function extractBackendMessage(payload: any): string | null {
  if (!payload || typeof payload !== 'object') return null
  const { error, details } = payload as BackendErr
  const msg = [error, details].filter(Boolean).join('：')
  return msg || null
}

// 请求拦截：自动加 Authorization
http.interceptors.request.use((cfg) => {
  const t = getToken()
  if (t) {
    cfg.headers = cfg.headers || {}
    cfg.headers['Authorization'] = `Bearer ${t}`
  }
  return cfg
})

// 避免多次 401 重复跳转
let redirecting = false

// 响应拦截
http.interceptors.response.use(
  // 成功分支：解包统一成功结构 { ok, data }；若是业务错误 { ok:false, error, details } 则转为失败并保留原始响应
  (resp: AxiosResponse) => {
    const payload = resp.data as any
    if (payload && typeof payload === 'object' && 'ok' in payload) {
      if (payload.ok === true) {
        // 解包 data，使调用方直接使用 response.data
        resp.data = payload.data
        return resp
      }
      // ok=false 视为业务失败，构造 AxiosError 保留 response 供调用方读取 details
      const msg = extractBackendMessage(payload) || '业务处理失败'
      const err = new AxiosError(msg, undefined, resp.config, resp.request, resp)
      return Promise.reject(err)
    }
    return resp
  },
  // 失败分支：保留原始 AxiosError，只增强 message，便于上层读取 error.response?.data 细节
  (error: AxiosError) => {
    const status = error?.response?.status
    const respData = error?.response?.data
    const msg =
      extractBackendMessage(respData) ||
      error.message ||
      (status ? `请求失败（${status}）` : '网络异常，请检查连接')

    // 401 未认证：清 token 并跳登录（带 redirect）
    if (status === 401) {
      clearAuth()
      if (!redirecting && !location.pathname.startsWith('/login')) {
        redirecting = true
        const here = encodeURIComponent(location.pathname + location.search + location.hash)
        location.href = `/login?redirect=${here}`
      }
    }

    // 不丢失原始响应，便于业务侧读取 details
    error.message = msg
    return Promise.reject(error)
  }
)

export default http
