// src/api/acl.ts
import http from './http'

export const AccountAPI = {
  // 后端 /accounts/list 接受 JSON Body：{ username_like, limit, offset, is_deleted, role }
  list: (params: { username_like?: string; limit?: number; offset?: number; is_deleted?: number; role?: number }) =>
    http.post('/accounts/list', params || {}),

  get: (id: string) => http.post('/accounts/get', { id }),

  get_by_username: (username: string) => http.post('/accounts/get_by_username', { username }),

  // create: 后端要求字段：username / password / org_id / role / description
  create: (data: { username: string; password: string; org_id: string; role?: number; description?: string }) =>
    http.post('/accounts/create', data),

  // update: 后端要求字段：id / username / org_id / description / role
  update: (data: { id: string; username?: string; org_id?: string; description?: string; role?: number }) =>
    http.post('/accounts/update', data),

  update_password: (data: { id: string; password: string }) =>
    http.post('/accounts/update_password', data),

  // 重置密码（普通用户本人）用 change_password，需要 old_password/new_password
  change_password: (data: { username: string; old_password: string; new_password: string }) =>
    http.post('/accounts/change_password', data),

  // 删除：用软删
  remove: (id: string) => http.post('/accounts/delete', { id }),

  // 硬删
  hardRemove: (id: string) => http.post('/accounts/hard_delete', { id }),
}

// 占位
export const PermAPI = {}
