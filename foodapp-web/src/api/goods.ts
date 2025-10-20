import http from './http'

export interface GoodsListParams {
  org_id: string
  keyword?: string
  category_id?: string
  spec_id?: string
  page?: number
  page_size?: number
}

export interface GoodsCreatePayload {
  name: string
  code: string
  org_id: string
  spec_id: string
  category_id: string
  sort?: number
  pinyin?: string
  image_url?: string
  acceptance_standard?: string
}

export interface GoodsUpdatePayload {
  id: string
  name?: string
  code?: string
  sort?: number
  spec_id?: string
  category_id?: string
  pinyin?: string
  image_url?: string
  acceptance_standard?: string
}

export interface GoodsRow {
  ID: string
  Name: string
  Code: string
  Sort: number
  Pinyin: string | null
  SpecID: string
  ImageURL: string | null
  AcceptanceStandard: string | null
  CategoryID: string
  OrgID: string
  IsDeleted: number
  CreatedAt: string
  UpdatedAt: string
}

export interface GoodsListResponse {
  total: number
  items: GoodsRow[]
}

export const GoodsAPI = {
  create: (data: GoodsCreatePayload) => http.post('/goods/create_goods', data),
  get: (id: string) => http.post('/goods/get_goods', { id }),
  list: (params: GoodsListParams) => 
    http.post<GoodsListResponse>('/goods/list_goods', null, { params }),
  update: (data: GoodsUpdatePayload) => http.post('/goods/update_goods', data),
  softDelete: (id: string) => http.post('/goods/soft_delete_goods', { id }),
  hardDelete: (id: string) => http.post('/goods/hard_delete_goods', { id }),
}

export default GoodsAPI
