// src/api/organ.ts
import http from './http'

export interface OrganListParams {
  name_like?: string
  is_deleted?: number
  limit?: number
  offset?: number
}

export interface OrganCreatePayload {
  name: string
  parent?: string
  code?: string
  description?: string
  sort?: number
}

export interface OrganUpdatePayload {
  id: string
  name?: string
  parent?: string
  code?: string
  description?: string
}

export const OrganAPI = {
  create: (data: OrganCreatePayload) => http.post('/orgs/create_organ', data),
  get: (id: string) => http.post('/orgs/get_organ', { id }),
  list: (params: OrganListParams) => http.post('/orgs/list_organ', params || {}),
  update: (data: OrganUpdatePayload) => http.post('/orgs/update_organ', data),
  softDelete: (id: string) => http.post('/orgs/soft_delete_organ', { id }),
  hardDelete: (id: string) => http.post('/orgs/hard_delete_organ', { id }),
}

export type OrganRow = {
  ID: string
  Name: string
  Code: string | null
  Parent: string
  Description: string
  Sort: number
  IsDeleted: number
  CreatedAt: string
  UpdatedAt: string
}

export default OrganAPI