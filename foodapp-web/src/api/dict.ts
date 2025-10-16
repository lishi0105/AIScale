import http from './http'


// 与后端 /dict/* 路由对齐
export const DictAPI = {
    // Unit
    createUnit: (data: { Name: string; Sort?: number; Code?: string }) =>
        http.post('/dict/create_unit', data),
    getUnit: (ID: string) => http.post('/dict/get_unit', { ID }),
    listUnits: (params: { keyword?: string; page?: number; page_size?: number }) =>
        http.post('/dict/list_unit', null, { params }),
    updateUnit: (data: { ID: string; Name: string; Sort?: number; Code?: string }) =>
        http.post('/dict/update_unit', data),
    deleteUnit: (ID: string) => http.post('/dict/delete_unit', { ID }), // 需后端开放


    // Spec
    createSpec: (data: { Name: string; Sort?: number; Code?: string }) =>
        http.post('/dict/create_spec', data),
    getSpec: (ID: string) => http.post('/dict/get_spec', { ID }),
    listSpecs: (params: { keyword?: string; page?: number; page_size?: number }) =>
        http.post('/dict/list_spec', null, { params }),
    updateSpec: (data: { ID: string; Name: string; Sort?: number; Code?: string }) =>
        http.post('/dict/update_spec', data),
    deleteSpec: (ID: string) => http.post('/dict/delete_spec', { ID }), // 需后端开放


    // MealTime
    createMealTime: (data: { Name: string; Sort?: number; Code?: string }) =>
        http.post('/dict/create_mealTime', data),
    getMealTime: (ID: string) => http.post('/dict/get_mealTime', { ID }),
    listMealTimes: (params: { keyword?: string; page?: number; page_size?: number }) =>
        http.post('/dict/list_mealTime', null, { params }),
    updateMealTime: (data: { ID: string; Name: string; Sort?: number; Code?: string }) =>
        http.post('/dict/update_mealTime', data),
    deleteMealTime: (ID: string) => http.post('/dict/delete_mealTime', { ID }), // 需后端开放
}
