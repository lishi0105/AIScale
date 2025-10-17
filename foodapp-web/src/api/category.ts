import http from './http'

// 与后端 /category/* 路由对齐
export const CategoryAPI = {
  // 创建品类
  createCategory: (data: { name: string; code?: string; pinyin?: string }) =>
    http.post('/category/create_category', data),

  // 获取单个品类
  getCategory: (id: string) => http.post('/category/get_category', { id }),

  // 获取品类列表
  listCategories: (params: { keyword?: string; page?: number; page_size?: number }) =>
    http.post('/category/list_category', null, { params }),

  // 更新品类
  updateCategory: (data: { id: string; name: string; code?: string; pinyin?: string }) =>
    http.post('/category/update_category', data),

  // 删除品类
  deleteCategory: (id: string) => http.post('/category/udelete_category', { id }),
}
