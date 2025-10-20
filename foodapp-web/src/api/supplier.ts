import http from './http'

export interface SupplierListParams {
  org_id: string
  keyword?: string
  contact_name?: string
  contact_phone?: string
  contact_email?: string
  contact_address?: string
  status?: number
  page?: number
  page_size?: number
}

export interface SupplierCreatePayload {
  name: string
  org_id: string
  float_ratio: number
  description?: string
  code?: string | null
  pinyin?: string | null
  contact_name?: string | null
  contact_phone?: string | null
  contact_email?: string | null
  contact_address?: string | null
  status?: number
  start_time?: string | null
  end_time?: string | null
}

export interface SupplierUpdatePayload {
  id: string
  name?: string
  code?: string | null
  pinyin?: string | null
  sort?: number
  status?: number
  description?: string
  float_ratio?: number
  contact_name?: string | null
  contact_phone?: string | null
  contact_email?: string | null
  contact_address?: string | null
  start_time?: string | null
  end_time?: string | null
}

export interface SupplierRow {
  ID: string
  Name: string
  Code: string | null
  Sort: number
  Pinyin: string | null
  Status: number
  Description: string
  FloatRatio: number
  OrgID: string
  ContactName: string | null
  ContactPhone: string | null
  ContactEmail: string | null
  ContactAddress: string | null
  StartTime: string | null
  EndTime: string | null
  CreatedAt: string
  UpdatedAt: string
}

export const SupplierAPI = {
  create: (data: SupplierCreatePayload) => http.post('/supplier/create_supplier', data),
  list: (params: SupplierListParams) => http.post('/supplier/list_supplier', null, { params }),
  update: (data: SupplierUpdatePayload) => http.post('/supplier/update_supplier', data),
  remove: (id: string) => http.post('/supplier/soft_delete_supplier', { id }),
  get: (id: string) => http.post('/supplier/get_supplier', { id }),
}

export default SupplierAPI
