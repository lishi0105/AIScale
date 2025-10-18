// src/api/acl.ts
import http from './http'
export interface AccountListParams {
  keyword?: string
  is_deleted?: number
  role?:number
  page?: number
  page_size?: number
}

export interface AccountCreatePayload {
  username: string
  password: string
  org_id: string
  role?: number
  description?: string | null
}

export interface AccountUpdatePayload {
  id: string
  username?: string
  org_id?: string
  description?: string | null
  role?: number
}

export interface AccountUpdatePasswdPayload {
  id: string
  password?: string
}

export interface AccountChangePasswdPayload {
  id: string
  old_password: string
  new_password: string
}

export const AccountAPI = {
  list: (params: AccountListParams) => http.post('/accounts/list_account', null, { params }),
  get: (id: string) => http.post('/accounts/get_account', { id }),
  get_by_username: (username: string) => http.post('/accounts/get_account_by_username', { username }),
  create: (payload: AccountCreatePayload) => http.post('/accounts/create_account', payload),
  update: (payload: AccountUpdatePayload) => http.post('/accounts/update_account', payload),
  update_password: (payload: AccountUpdatePasswdPayload) => http.post('/accounts/update_account_password', payload),
  change_password: (payload: AccountChangePasswdPayload) => http.post('/accounts/change_account_password', payload),
  remove: (id: string) => http.post('/accounts/soft_delete_account', { id }),
  delete: (id: string) => http.post('/accounts/hard_delete_account', { id }),
}

// 占位
export const PermAPI = {}
