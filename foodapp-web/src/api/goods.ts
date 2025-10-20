import http from './http'

export interface GoodsListParams {
  org_id: string
  category_id?: string
  spec_id?: string
  keyword?: string
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
  pinyin?: string | null
  image_url?: string | null
  acceptance_standard?: string | null
}

export interface GoodsUpdatePayload {
  id: string
  name?: string
  code?: string
  sort?: number
  spec_id?: string
  category_id?: string
  pinyin?: string | null
  image_url?: string | null
  acceptance_standard?: string | null
}

export interface GoodsRow {
  ID: string
  Name: string
  Code: string
  Sort: number
  Pinyin: string | null
  SpecID: string
  CategoryID: string
  OrgID: string
  ImageURL: string | null
  AcceptanceStandard: string | null
  CreatedAt: string
  UpdatedAt: string
}

export const GoodsAPI = {
  create: (data: GoodsCreatePayload) => http.post('/goods/create_goods', data),
  get: (id: string) => http.post('/goods/get_goods', { id }),
  list: (params: GoodsListParams) => http.post('/goods/list_goods', null, { params }),
  update: (data: GoodsUpdatePayload) => http.post('/goods/update_goods', data),
  remove: (id: string) => http.post('/goods/soft_delete_goods', { id }),
}

export default GoodsAPI
