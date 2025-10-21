import http from './http'

// ========== 询价单 (BasePriceInquiry) ==========

export interface InquiryListParams {
  org_id: string
  year?: number
  month?: number
  ten_day?: number
  keyword?: string
  page?: number
  page_size?: number
}

export interface InquiryCreatePayload {
  org_id: string
  inquiry_title: string
  inquiry_date: string // YYYY-MM-DD
}

export interface InquiryUpdatePayload {
  id: string
  inquiry_title?: string
  inquiry_date?: string // YYYY-MM-DD
}

export interface InquiryRow {
  ID: string
  OrgID: string
  InquiryTitle: string
  InquiryDate: string
  InquiryYear?: number
  InquiryMonth?: number
  InquiryTenDay?: number
  IsDeleted: number
  CreatedAt: string
  UpdatedAt: string
}

export const InquiryAPI = {
  create: (data: InquiryCreatePayload) => http.post('/inquiry/create_inquiry', data),
  get: (id: string) => http.post('/inquiry/get_inquiry', { id }),
  list: (params: InquiryListParams) => http.post('/inquiry/list_inquiries', null, { params }),
  update: (data: InquiryUpdatePayload) => http.post('/inquiry/update_inquiry', data),
  remove: (id: string) => http.post('/inquiry/soft_delete_inquiry', { id }),
}

// ========== 询价商品明细 (PriceInquiryItem) ==========

export interface InquiryItemListParams {
  inquiry_id: string
  category_id?: string
  page?: number
  page_size?: number
}

export interface InquiryItemCreatePayload {
  inquiry_id: string
  goods_id: string
  category_id: string
  spec_id?: string | null
  unit_id?: string | null
  goods_name_snap: string
  category_name_snap: string
  spec_name_snap?: string | null
  unit_name_snap?: string | null
  guide_price?: number | null
  last_month_avg_price?: number | null
  current_avg_price?: number | null
  sort?: number
}

export interface InquiryItemUpdatePayload {
  id: string
  goods_id?: string
  category_id?: string
  spec_id?: string | null
  unit_id?: string | null
  goods_name_snap?: string
  category_name_snap?: string
  spec_name_snap?: string | null
  unit_name_snap?: string | null
  guide_price?: number | null
  last_month_avg_price?: number | null
  current_avg_price?: number | null
  sort?: number
}

export interface InquiryItemRow {
  ID: string
  InquiryID: string
  GoodsID: string
  CategoryID: string
  SpecID?: string | null
  UnitID?: string | null
  GoodsNameSnap: string
  CategoryNameSnap: string
  SpecNameSnap?: string | null
  UnitNameSnap?: string | null
  GuidePrice?: number | null
  LastMonthAvgPrice?: number | null
  CurrentAvgPrice?: number | null
  Sort: number
  IsDeleted: number
  CreatedAt: string
  UpdatedAt: string
}

export const InquiryItemAPI = {
  create: (data: InquiryItemCreatePayload) => http.post('/inquiry_item/create_inquiry_item', data),
  get: (id: string) => http.post('/inquiry_item/get_inquiry_item', { id }),
  list: (params: InquiryItemListParams) => http.post('/inquiry_item/list_inquiry_items', null, { params }),
  update: (data: InquiryItemUpdatePayload) => http.post('/inquiry_item/update_inquiry_item', data),
  remove: (id: string) => http.post('/inquiry_item/soft_delete_inquiry_item', { id }),
}

// ========== 市场报价 (PriceMarketInquiry) ==========

export interface MarketInquiryListParams {
  inquiry_id?: string
  item_id?: string
  page?: number
  page_size?: number
}

export interface MarketInquiryCreatePayload {
  inquiry_id: string
  item_id: string
  market_id?: string | null
  market_name_snap: string
  price?: number | null
}

export interface MarketInquiryUpdatePayload {
  id: string
  market_id?: string | null
  market_name_snap?: string
  price?: number | null
}

export interface MarketInquiryRow {
  ID: string
  InquiryID: string
  ItemID: string
  MarketID?: string | null
  MarketNameSnap: string
  Price?: number | null
  IsDeleted: number
  CreatedAt: string
  UpdatedAt: string
}

export const MarketInquiryAPI = {
  create: (data: MarketInquiryCreatePayload) => http.post('/market_inquiry/create_market_inquiry', data),
  get: (id: string) => http.post('/market_inquiry/get_market_inquiry', { id }),
  list: (params: MarketInquiryListParams) => http.post('/market_inquiry/list_market_inquiries', null, { params }),
  update: (data: MarketInquiryUpdatePayload) => http.post('/market_inquiry/update_market_inquiry', data),
  remove: (id: string) => http.post('/market_inquiry/soft_delete_market_inquiry', { id }),
}

// ========== 市场主数据 (BaseMarket) ==========

export interface MarketListParams {
  org_id: string
  keyword?: string
  page?: number
  page_size?: number
}

export interface MarketCreatePayload {
  name: string
  org_id: string
  code?: string
  sort?: number
}

export interface MarketUpdatePayload {
  id: string
  name?: string
  code?: string
  sort?: number
}

export interface MarketRow {
  ID: string
  Name: string
  OrgID: string
  Code?: string | null
  Sort: number
  IsDeleted: number
  CreatedAt: string
  UpdatedAt: string
}

export const MarketAPI = {
  create: (data: MarketCreatePayload) => http.post('/market/create_market', data),
  get: (id: string) => http.post('/market/get_market', { id }),
  list: (params: MarketListParams) => http.post('/market/list_markets', null, { params }),
  update: (data: MarketUpdatePayload) => http.post('/market/update_market', data),
  remove: (id: string) => http.post('/market/soft_delete_market', { id }),
}

export default { InquiryAPI, InquiryItemAPI, MarketInquiryAPI, MarketAPI }
