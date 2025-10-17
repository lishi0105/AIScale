// src/api/acl.ts
import http from './http'

export const AccountAPI = {
  // 后端 /accounts/list_account 接受 JSON Body：{ username_like, limit, offset, ... }
  list: (params: { username_like?: string; is_deleted?: number; role?: number; limit?: number; offset?: number }) =>
    http.post('/accounts/list_account', params || {}),

  get: (id: string) => http.post('/accounts/get_account', { id }),

  get_by_username: (username: string) => http.post('/accounts/get_account_by_username', { username }),

  // create: 后端要求字段全小写：username / password / org_id / role? / description?
  create: (data: { username: string; password: string; org_id: string; role?: number; description?: string | null }) =>
    http.post('/accounts/create_account', data),

  // update: 更新通用字段（部分可选）
  update: (data: { id: string; username?: string; org_id?: string; description?: string | null; role?: number }) =>
    http.post('/accounts/update_account', data),

  update_password: (data: { id: string; password: string }) =>
    http.post('/accounts/update_account_password', data),

  // 重置密码（普通用户本人）用 change_password，需要 old_password/new_password
  change_password: (data: { username: string; old_password: string; new_password: string }) =>
    http.post('/accounts/change_account_password', data),

  // 删除：用软删
  remove: (id: string) => http.post('/accounts/soft_delete_account', { id }),

  // 如需硬删，可暴露：hardRemove: (id: string) => http.post('/accounts/hard_delete_account', { id }),
}

// 占位
export const PermAPI = {}
