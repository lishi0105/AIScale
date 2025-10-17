// src/api/category.ts
import http from './http'

export interface CategoryCreatePayload {
  name: string
  team_id: string
  code?: string
  pinyin?: string
}

export interface CategoryUpdatePayload {
  id: string
  name: string
  code?: string
  pinyin?: string
}

export interface CategoryListParams {
  team_id: string
  keyword?: string
  page?: number
  page_size?: number
}

export const CategoryAPI = {
  create: (data: CategoryCreatePayload) => http.post('/category/create_category', data),
  get: (id: string) => http.post('/category/get_category', { id }),
  list: (params: CategoryListParams) => http.get('/category/list_category', { params }),
  update: (data: CategoryUpdatePayload) => http.post('/category/update_category', data),
  delete: (id: string) => http.post('/category/udelete_category', { id }),
}

export type CategoryRow = {
  ID: string
  Name: string
  Code: string | null
  Pinyin: string | null
  Sort: number
  TeamId: string
  IsDeleted: number
  CreatedAt: string
  UpdatedAt: string
}

export default CategoryAPI
