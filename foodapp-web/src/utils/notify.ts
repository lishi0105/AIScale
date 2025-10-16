// src/utils/notify.ts
import type { AxiosError } from 'axios'
import { ElMessage } from 'element-plus'

type BackendErr = { error?: string; details?: string; [k: string]: any }

function getBackendMessage(payload: any): string | null {
  if (!payload || typeof payload !== 'object') return null
  const { error, details } = payload as BackendErr
  const msg = [error, details].filter(Boolean).join('：')
  return msg || null
}

// 简单的防抖去重，避免 1 秒内重复弹同一条
let lastMsg = ''
let lastTs = 0
function shouldShow(msg: string): boolean {
  const now = Date.now()
  if (msg === lastMsg && now - lastTs < 1000) return false
  lastMsg = msg
  lastTs = now
  return true
}

/**
 * 统一错误提示：直接从后端 {error, details} 读取
 * - 401：静默（交给 http 拦截器做跳转）
 * - 其它：弹出后端拼好的文案；兜底为 AxiosError.message 或通用文案
 */
export function notifyError(e: unknown): void {
  const ax = e as AxiosError
  const status = ax?.response?.status
  if (status === 401) return // 交给拦截器处理跳转

  const payload = ax?.response?.data
  const msg =
    getBackendMessage(payload) ||
    ax?.message ||
    (status ? `请求失败（${status}）` : '网络异常，请检查连接')

  if (msg && shouldShow(msg)) {
    ElMessage.error(msg)
  }
}
