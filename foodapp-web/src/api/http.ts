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
  // 成功分支：如果后端在 200 中也塞了 {error, details}，按业务失败处理
  (resp: AxiosResponse) => {
    const msg = extractBackendMessage(resp.data)
    if (msg) {
      const err = new AxiosError(msg, undefined, resp.config, resp.request, resp)
      return Promise.reject(err)
    }
    return resp
  },
  // 失败分支：把 message 统一为可读文案，并处理 401 跳转
  (error: AxiosError) => {
    const status = error?.response?.status
    const respData = error?.response?.data
    let msg =
      extractBackendMessage(respData) ||
      error.message ||
      (status ? `请求失败（${status}）` : '网络异常，请检查连接')

    // 401 未认证：清 token 并跳登录（带 redirect）
    if (status === 401) {
      clearAuth()
      if (!redirecting && !location.pathname.startsWith('/login')) {
        redirecting = true
        const here = encodeURIComponent(location.pathname + location.search + location.hash)
        location.href = `/login?redirect=${here}` // ✨ 过期后登录回来仍在原页
      }
    }

    // 用 Error 包装，message 已是友好文案
    return Promise.reject(new Error(msg))
  }
)

export default http
