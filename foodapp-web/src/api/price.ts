import http from './http'

export interface InquiryRow {
  ID: string
  OrgID: string
  InquiryTitle: string
  InquiryDate: string
  InquiryYear?: number | null
  InquiryMonth?: number | null
  InquiryTenDay?: number | null
  CreatedAt: string
  UpdatedAt: string
}

export const PriceAPI = {
  inquiryCreate: (data: { org_id: string; inquiry_title: string; inquiry_date: string }) =>
    http.post('/inquiry/create_inquiry', data),
  inquiryList: (params: {
    org_id: string
    year?: number
    month?: number
    ten_day?: number
    keyword?: string
    page?: number
    page_size?: number
  }) => http.post('/inquiry/list_inquiries', null, { params }),
  inquiryItemCreate: (data: {
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
    sort?: number | null
  }) => http.post('/inquiry_item/create_inquiry_item', data),
}

export default PriceAPI
