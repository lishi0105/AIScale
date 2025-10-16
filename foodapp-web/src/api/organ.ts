// src/api/organ.ts
import http from './http'

export interface Organ {
  ID: string
  Name: string
  Code?: string
  Parent: string
  Description: string
  Sort: number
  IsDeleted: number
  CreatedAt: string
  UpdatedAt: string
}

export interface OrganCreateReq {
  name: string
  parent?: string
  code?: string
  description?: string
  sort?: number
}

export interface OrganUpdateReq {
  id: string
  name?: string
  parent?: string
  code?: string
  description?: string
}

export interface OrganListReq {
  name_like?: string
  is_deleted?: number
  limit?: number
  offset?: number
}

export const OrganAPI = {
  // 创建组织
  create: (data: OrganCreateReq) =>
    http.post('/orgs/create', data),

  // 获取单个组织
  get: (id: string) => 
    http.post('/orgs/get', { id }),

  // 获取组织列表
  list: (params: OrganListReq) =>
    http.post('/orgs/list', params || {}),

  // 更新组织
  update: (data: OrganUpdateReq) =>
    http.post('/orgs/update', data),

  // 软删除组织
  remove: (id: string) =>
    http.post('/orgs/delete', { id }),

  // 硬删除组织
  hardRemove: (id: string) =>
    http.post('/orgs/hard_delete', { id }),
}