# 询价记录 API 文档

## 概述

询价记录模块提供了完整的CRUD操作，支持询价单的创建、查询、更新和删除功能。

## 数据模型

### BasePriceInquiry 字段说明

| 字段名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | string(UUID) | 是 | 主键，自动生成 |
| inquiry_title | string | 是 | 询价单标题，最大64字符 |
| inquiry_date | date | 是 | 询价单日期（业务日） |
| market_1 | string | 否 | 市场1，最大128字符 |
| market_2 | string | 否 | 市场2，最大128字符 |
| market_3 | string | 否 | 市场3，最大128字符 |
| org_id | string(UUID) | 是 | 中队ID |
| inquiry_start_date | datetime | 是 | 开始时间 |
| inquiry_end_date | datetime | 是 | 结束时间 |
| created_at | datetime | 是 | 创建时间，自动生成 |
| updated_at | datetime | 是 | 更新时间，自动生成 |

### 业务约束

1. **唯一性约束**：同一组织下，标题和业务日期不能重复
2. **时间约束**：结束时间必须晚于开始时间
3. **软删除**：使用 `is_deleted` 字段标记删除状态

## API 接口

### 1. 创建询价记录

**POST** `/api/v1/inquiry/create_inquiry`

**权限要求**：管理员

**请求体**：
```json
{
  "inquiry_title": "2024年第一季度询价",
  "inquiry_date": "2024-01-15",
  "market_1": "北京市场",
  "market_2": "上海市场",
  "market_3": "广州市场",
  "org_id": "123e4567-e89b-12d3-a456-426614174000",
  "inquiry_start_date": "2024-01-15T09:00:00Z",
  "inquiry_end_date": "2024-01-15T18:00:00Z"
}
```

**响应**：
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174001",
  "inquiry_title": "2024年第一季度询价",
  "inquiry_date": "2024-01-15T00:00:00Z",
  "market_1": "北京市场",
  "market_2": "上海市场",
  "market_3": "广州市场",
  "org_id": "123e4567-e89b-12d3-a456-426614174000",
  "is_deleted": 0,
  "inquiry_start_date": "2024-01-15T09:00:00Z",
  "inquiry_end_date": "2024-01-15T18:00:00Z",
  "created_at": "2024-01-15T10:00:00Z",
  "updated_at": "2024-01-15T10:00:00Z",
  "active_title": "2024年第一季度询价"
}
```

### 2. 获取询价记录

**POST** `/api/v1/inquiry/get_inquiry`

**权限要求**：已认证用户

**请求体**：
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174001"
}
```

**响应**：返回完整的询价记录对象

### 3. 查询询价记录列表

**POST** `/api/v1/inquiry/list_inquiries`

**权限要求**：已认证用户

**请求体**（JSON方式）：
```json
{
  "keyword": "第一季度",
  "org_id": "123e4567-e89b-12d3-a456-426614174000",
  "start_date": "2024-01-01",
  "end_date": "2024-03-31",
  "market_1": "北京市场",
  "page": 1,
  "page_size": 20
}
```

**请求参数**（Query方式）：
- `org_id`: 组织ID（必填）
- `keyword`: 关键词搜索（可选）
- `start_date`: 开始日期（可选，格式：YYYY-MM-DD）
- `end_date`: 结束日期（可选，格式：YYYY-MM-DD）
- `market_1`: 市场1过滤（可选）
- `market_2`: 市场2过滤（可选）
- `market_3`: 市场3过滤（可选）
- `page`: 页码（可选，默认1）
- `page_size`: 每页大小（可选，默认20，最大100）

**响应**：
```json
{
  "total": 100,
  "items": [
    {
      "id": "123e4567-e89b-12d3-a456-426614174001",
      "inquiry_title": "2024年第一季度询价",
      "inquiry_date": "2024-01-15T00:00:00Z",
      "market_1": "北京市场",
      "market_2": "上海市场",
      "market_3": "广州市场",
      "org_id": "123e4567-e89b-12d3-a456-426614174000",
      "is_deleted": 0,
      "inquiry_start_date": "2024-01-15T09:00:00Z",
      "inquiry_end_date": "2024-01-15T18:00:00Z",
      "created_at": "2024-01-15T10:00:00Z",
      "updated_at": "2024-01-15T10:00:00Z",
      "active_title": "2024年第一季度询价"
    }
  ]
}
```

### 4. 更新询价记录

**POST** `/api/v1/inquiry/update_inquiry`

**权限要求**：管理员

**请求体**：
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174001",
  "inquiry_title": "2024年第一季度询价（更新）",
  "market_1": "北京市场（更新）",
  "inquiry_end_date": "2024-01-15T19:00:00Z"
}
```

**响应**：HTTP 204 No Content

### 5. 软删除询价记录

**POST** `/api/v1/inquiry/soft_delete_inquiry`

**权限要求**：管理员

**请求体**：
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174001"
}
```

**响应**：
```json
{
  "ok": true
}
```

### 6. 硬删除询价记录

**POST** `/api/v1/inquiry/hard_delete_inquiry`

**权限要求**：管理员

**请求体**：
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174001"
}
```

**响应**：HTTP 204 No Content

## 错误处理

所有接口都遵循统一的错误响应格式：

```json
{
  "error": "错误描述",
  "details": "详细错误信息"
}
```

### 常见错误码

- `400 Bad Request`: 请求参数错误
- `401 Unauthorized`: 未认证
- `403 Forbidden`: 权限不足
- `404 Not Found`: 资源不存在
- `409 Conflict`: 业务冲突（如唯一性约束）
- `500 Internal Server Error`: 服务器内部错误

## 使用示例

### 创建询价记录
```bash
curl -X POST http://localhost:8080/api/v1/inquiry/create_inquiry \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "inquiry_title": "2024年第一季度询价",
    "inquiry_date": "2024-01-15",
    "org_id": "123e4567-e89b-12d3-a456-426614174000",
    "inquiry_start_date": "2024-01-15T09:00:00Z",
    "inquiry_end_date": "2024-01-15T18:00:00Z"
  }'
```

### 查询询价记录列表
```bash
curl -X POST http://localhost:8080/api/v1/inquiry/list_inquiries \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "org_id": "123e4567-e89b-12d3-a456-426614174000",
    "keyword": "第一季度",
    "page": 1,
    "page_size": 20
  }'
```

## 注意事项

1. 所有时间字段都使用 ISO 8601 格式
2. 软删除的记录不会在查询结果中显示
3. 更新操作会同时更新 `active_title` 字段
4. 分页查询默认按日期倒序排列
5. 关键词搜索仅针对标题字段