# 询价单导入功能说明

## 功能概述

实现了询价单Excel文件的异步导入功能，包含：
- ✅ 异步上传和导入，避免HTTP超时
- ✅ 实时进度显示（0-100%）
- ✅ 重复检测和确认覆盖功能
- ✅ 友好的错误提示

## 已完成的修改

### 后端修改

#### 1. Service层 (`internal/service/inquiry/import.go`)

**新增功能**：
- `DuplicateInquiryError` - 重复错误类型
- `checkDuplicateInquiry()` - 检查重复并返回详细错误
- `deleteInquiryCompletely()` - 完全删除询价单及所有关联数据
- 修改 `ImportExcelDataAsync()` - 支持 `forceDelete` 参数

**关键逻辑**：
```go
// 检测到重复时
if dupErr, ok := err.(*DuplicateInquiryError); ok && forceDelete {
    // 删除旧记录
    s.deleteInquiryCompletely(tx, dupErr.InquiryID)
    // 继续导入新数据
}
```

#### 2. Handler层 (`internal/server/handler/inquiry.go`)

**修改**：
- `importInquiry()` - 接收 `force_delete` 表单参数
- 传递 `forceDelete` 给异步导入方法

### 前端修改

#### 1. API文件 (`src/api/inquiry.ts`)

```typescript
// 询价单API
export const InquiryAPI = {
  // CRUD操作
  create: (data) => { ... },
  get: (id) => { ... },
  list: (params) => { ... },
  update: (data) => { ... },
  remove: (id) => { ... },
  
  // 导入功能（新增）
  importInquiry: (formData: FormData) => { ... },
  getImportStatus: (taskId: string) => { ... }
}

// 还包含：
// - InquiryItemAPI: 询价商品明细
// - MarketInquiryAPI: 市场报价
// - MarketAPI: 市场主数据
```

#### 2. 导入页面 (`src/pages/InquiryImport.vue`)

**功能特性**：
- 📤 拖拽上传Excel文件
- 📊 实时进度显示（进度条 + 百分比）
- ⚠️ 重复检测和确认对话框
- ✅ 成功/失败状态显示
- 🔄 自动轮询任务状态（每2秒）

**核心流程**：
```
1. 用户上传文件
   ↓
2. 提交到后端（force_delete=false）
   ↓
3. 如果检测到重复
   → 弹出确认对话框
   → 用户确认后重新提交（force_delete=true）
   ↓
4. 返回task_id，开始轮询
   ↓
5. 每2秒查询一次进度
   ↓
6. 完成或失败时停止轮询
```

#### 3. 路由配置 (`src/router/index.ts`)

新增路由：
```typescript
{
  path: '/inquiry/import',
  component: InquiryImport,
  meta: { 
    requiresAuth: true, 
    section: '询价管理', 
    title: '询价单导入' 
  }
}
```

## 使用说明

### 访问页面

访问路径: `http://your-domain/inquiry/import`

### 操作步骤

1. **输入组织ID**
   - 在"组织ID"输入框中填写目标组织的ID

2. **上传Excel文件**
   - 拖拽文件到上传区域，或点击上传
   - 仅支持 `.xlsx` 或 `.xls` 格式

3. **开始导入**
   - 点击"开始导入"按钮
   - 系统会先校验Excel格式（同步）
   - 校验通过后立即返回任务ID

4. **查看进度**
   - 页面自动显示导入进度
   - 实时更新百分比和已处理sheet数
   - 进度条颜色会根据状态变化：
     - 🔵 蓝色：处理中
     - 🟢 绿色：成功
     - 🔴 红色：失败

5. **处理重复**
   - 如果检测到相同标题或日期的询价单
   - 系统会弹出确认对话框
   - 选择"确定覆盖"会删除旧记录并导入新数据
   - 选择"取消"则停止导入

### 状态说明

| 状态 | 说明 | 颜色 |
|------|------|------|
| pending | 等待处理 | 灰色 |
| processing | 正在导入 | 橙色 |
| success | 导入成功 | 绿色 |
| failed | 导入失败 | 红色 |

## API接口

### 1. 导入接口

**请求**：
```http
POST /api/v1/inquiry_import/import_inquiry
Content-Type: multipart/form-data

参数：
- org_id: string (必填)
- file: File (必填)
- force_delete: string (可选, "true" 表示强制覆盖)
```

**响应**：
```json
{
  "ok": true,
  "message": "Excel文件校验通过，开始异步导入",
  "task_id": "550e8400-e29b-41d4-a716-446655440000",
  "stats": {
    "title": "2025年9月上旬询价单",
    "sheets": 3,
    "markets": 5,
    "suppliers": 2
  }
}
```

### 2. 查询进度接口

**请求**：
```http
POST /api/v1/inquiry_import/import_status
Content-Type: application/json

{
  "task_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**响应**：
```json
{
  "ok": true,
  "task_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "processing",
  "progress": 45,
  "total_sheets": 3,
  "processed_sheets": 1,
  "file_name": "询价单.xlsx",
  "message": "正在导入中...",
  "created_at": "2025-10-22T10:00:00Z",
  "updated_at": "2025-10-22T10:00:15Z"
}
```

## 技术细节

### 后端

1. **重复检测**：
   - 检查相同组织下是否存在相同标题或日期的询价单
   - 返回 `DuplicateInquiryError` 类型错误

2. **强制删除**：
   - 删除询价单 (`base_price_inquiry`)
   - 删除商品明细 (`price_inquiry_item`)
   - 删除市场报价 (`price_market_inquiry`)
   - 删除供应商结算 (`price_supplier_settlement`)
   - 全部在同一个事务中执行

3. **异步处理**：
   - 使用 goroutine 在后台执行导入
   - 使用 `sync.Map` 存储任务状态（内存）
   - 实时更新进度到内存

### 前端

1. **轮询机制**：
   - 上传成功后立即查询一次
   - 之后每2秒查询一次
   - 完成或失败时自动停止
   - 组件卸载时清理定时器

2. **错误处理**：
   - 检测错误消息中的关键字
   - 识别重复错误并弹出确认框
   - 其他错误直接显示提示

3. **用户体验**：
   - 实时进度反馈
   - 清晰的状态提示
   - 友好的错误信息
   - 支持拖拽上传

## 注意事项

⚠️ **重要**：
1. 强制覆盖会**永久删除**旧的询价单及所有关联数据
2. 删除操作**不可恢复**
3. 建议在覆盖前确认旧数据不再需要
4. 服务重启会丢失进行中的任务状态（使用内存存储）

## 测试清单

部署前请测试：
- [ ] 正常导入流程
- [ ] 重复检测功能
- [ ] 确认覆盖功能
- [ ] 取消覆盖功能
- [ ] 进度显示准确性
- [ ] 成功/失败状态显示
- [ ] 错误提示友好性
- [ ] 文件格式校验
- [ ] 大文件导入（测试异步效果）

## 相关文档

后端详细说明: `../ASYNC_IMPORT_SIMPLIFIED.md`

## 后续优化建议

1. **历史记录**：添加导入历史记录页面
2. **批量操作**：支持批量导入多个文件
3. **导出功能**：支持导出询价单为Excel
4. **模板下载**：提供标准模板下载
5. **预览功能**：导入前预览数据

