// src/api/acl.ts
import http from './http'

export const AccountAPI = {
  // 后端 /accounts/list 接受 JSON Body：{ username_like, is_deleted, role, limit, offset }
  list: (params: { username_like?: string; is_deleted?: number; role?: number; limit?: number; offset?: number }) =>
    http.post('/accounts/list', params || {}),

  get: (id: string) => http.post('/accounts/get', { id }),

  get_by_username: (username: string) => http.post('/accounts/get_by_username', { username }),

  // create: 后端要求 org_id 必填，description 可选
  create: (data: { username: string; password: string; org_id: string; role?: number; description?: string }) =>
    http.post('/accounts/create', data),

  // update: 更新用户信息（username, org_id, description, role）
  update: (data: { id: string; username?: string; org_id?: string; description?: string; role?: number }) =>
    http.post('/accounts/update', data),

  update_password: (data: { id: string; password: string }) =>
    http.post('/accounts/update_password', data),

  // 重置密码（普通用户本人）用 change_password，需要 old_password/new_password
  change_password: (data: { username: string; old_password: string; new_password: string }) =>
    http.post('/accounts/change_password', data),

  // 软删除
  remove: (id: string) => http.post('/accounts/delete', { id }),

  // 硬删除
  hardRemove: (id: string) => http.post('/accounts/hard_delete', { id }),
}

// 占位
export const PermAPI = {}
