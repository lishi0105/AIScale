import http from './http'

export interface CategoryListParams {
  org_id: string
  keyword?: string
  page?: number
  page_size?: number
}

export interface CategoryCreatePayload {
  name: string
  org_id: string
  code?: string
  pinyin?: string
}

export interface CategoryUpdatePayload {
  id: string
  name: string
  code?: string | null
  pinyin?: string | null
  sort?: number
}

export interface CategoryRow {
  ID: string
  Name: string
  Code: string | null
  Pinyin: string | null
  Sort: number
  OrganID: string
  IsDeleted: number
  CreatedAt: string
  UpdatedAt: string
}

export const CategoryAPI = {
  create: (data: CategoryCreatePayload) => http.post('/category/create_category', data),
  list: (params: CategoryListParams) => http.post('/category/list_category', null, { params }),
  update: (data: CategoryUpdatePayload) => http.post('/category/update_category', data),
  remove: (id: string) => http.post('/category/udelete_category', { id }),
}

export default CategoryAPI
