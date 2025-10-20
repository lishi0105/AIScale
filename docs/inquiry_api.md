# 询价记录 API 文档

## 概述

询价记录模块提供了完整的 CRUD 操作，用于管理询价单信息。

## API 端点

所有端点都需要身份验证，并且必须具有管理员权限（创建、更新、删除操作）。

基础路径: `/api/v1/inquiries`

### 1. 创建询价记录

**端点**: `POST /api/v1/inquiries/create_inquiry`

**权限**: 仅管理员

**请求体**:
```json
{
  "inquiry_title": "2024年10月询价单",
  "inquiry_date": "2024-10-20",
  "market_1": "北京新发地市场",
  "market_2": "北京大洋路市场",
  "market_3": "北京顺义市场",
  "org_id": "550e8400-e29b-41d4-a716-446655440000",
  "inquiry_start_date": "2024-10-20T08:00:00Z",
  "inquiry_end_date": "2024-10-20T18:00:00Z"
}
```

**字段说明**:
- `inquiry_title` (必填): 询价单标题，最大64字符
- `inquiry_date` (必填): 询价单日期（业务日），格式: YYYY-MM-DD
- `market_1` (可选): 市场1名称，最大128字符
- `market_2` (可选): 市场2名称，最大128字符
- `market_3` (可选): 市场3名称，最大128字符
- `org_id` (必填): 中队ID，UUID格式
- `inquiry_start_date` (必填): 开始时间，RFC3339格式
- `inquiry_end_date` (必填): 结束时间，RFC3339格式（必须晚于开始时间）

**响应**: 
- 成功: HTTP 201 Created，返回创建的询价记录对象
- 失败: HTTP 4xx/5xx，返回错误信息

**约束**:
- 同一组织内，相同标题和业务日期的询价单不能重复（软删除的记录不受此限制）
- 结束时间必须晚于开始时间

---

### 2. 获取单个询价记录

**端点**: `POST /api/v1/inquiries/get_inquiry`

**权限**: 所有已认证用户

**请求体**:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**响应**:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "inquiry_title": "2024年10月询价单",
  "inquiry_date": "2024-10-20T00:00:00Z",
  "market_1": "北京新发地市场",
  "market_2": "北京大洋路市场",
  "market_3": "北京顺义市场",
  "org_id": "550e8400-e29b-41d4-a716-446655440000",
  "is_deleted": 0,
  "inquiry_start_date": "2024-10-20T08:00:00Z",
  "inquiry_end_date": "2024-10-20T18:00:00Z",
  "created_at": "2024-10-20T10:30:00Z",
  "updated_at": "2024-10-20T10:30:00Z"
}
```

---

### 3. 获取询价记录列表

**端点**: `GET /api/v1/inquiries/list_inquiries`

**权限**: 所有已认证用户

**查询参数**:
- `org_id` (必填): 组织ID
- `keyword` (可选): 搜索关键词（按标题模糊匹配）
- `start_date` (可选): 开始日期，格式: YYYY-MM-DD
- `end_date` (可选): 结束日期，格式: YYYY-MM-DD
- `page` (可选): 页码，默认为1
- `page_size` (可选): 每页数量，默认为20，最大1000

**示例**:
```
GET /api/v1/inquiries/list_inquiries?org_id=550e8400-e29b-41d4-a716-446655440000&keyword=10月&start_date=2024-10-01&end_date=2024-10-31&page=1&page_size=20
```

**响应**:
```json
{
  "total": 42,
  "items": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "inquiry_title": "2024年10月询价单",
      "inquiry_date": "2024-10-20T00:00:00Z",
      "market_1": "北京新发地市场",
      "market_2": "北京大洋路市场",
      "market_3": "北京顺义市场",
      "org_id": "550e8400-e29b-41d4-a716-446655440000",
      "is_deleted": 0,
      "inquiry_start_date": "2024-10-20T08:00:00Z",
      "inquiry_end_date": "2024-10-20T18:00:00Z",
      "created_at": "2024-10-20T10:30:00Z",
      "updated_at": "2024-10-20T10:30:00Z"
    }
    // ... 更多记录
  ]
}
```

**排序规则**:
- 按 `inquiry_date` 降序（最新的在前）
- 相同日期按 `inquiry_title` 升序

---

### 4. 更新询价记录

**端点**: `POST /api/v1/inquiries/update_inquiry`

**权限**: 仅管理员

**请求体** (所有字段都是可选的，只更新提供的字段):
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "inquiry_title": "2024年10月询价单（已更新）",
  "inquiry_date": "2024-10-21",
  "market_1": "北京新发地市场（更新）",
  "market_2": null,
  "market_3": "北京顺义市场",
  "inquiry_start_date": "2024-10-21T08:00:00Z",
  "inquiry_end_date": "2024-10-21T18:00:00Z"
}
```

**字段说明**:
- `id` (必填): 询价记录ID
- 其他字段都是可选的
- 设置为 `null` 会清空该字段（仅对可选字段有效，如 market_1/2/3）

**响应**: 
- 成功: HTTP 204 No Content
- 失败: HTTP 4xx/5xx，返回错误信息

**约束**:
- 如果更新时间范围，结束时间必须晚于开始时间
- 更新标题或日期时，仍需满足唯一性约束

---

### 5. 软删除询价记录

**端点**: `POST /api/v1/inquiries/soft_delete_inquiry`

**权限**: 仅管理员

**请求体**:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**响应**:
```json
{
  "ok": true
}
```

**说明**: 软删除只是将 `is_deleted` 标记为1，数据仍然保留在数据库中，但不会在列表中显示。

---

### 6. 硬删除询价记录

**端点**: `POST /api/v1/inquiries/hard_delete_inquiry`

**权限**: 仅管理员

**请求体**:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**响应**: HTTP 204 No Content

**警告**: 硬删除会从数据库中永久删除记录，无法恢复！

---

## 数据库结构

### 表名: `base_price_inquiry`

| 字段 | 类型 | 说明 |
|------|------|------|
| id | CHAR(36) | UUID主键 |
| inquiry_title | VARCHAR(64) | 询价单标题 |
| inquiry_date | DATE | 询价单日期（业务日） |
| market_1 | VARCHAR(128) | 市场1 |
| market_2 | VARCHAR(128) | 市场2 |
| market_3 | VARCHAR(128) | 市场3 |
| org_id | CHAR(36) | 中队ID |
| is_deleted | TINYINT(1) | 软删标记：0=有效 1=删除 |
| inquiry_start_date | DATETIME | 开始时间 |
| inquiry_end_date | DATETIME | 结束时间 |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |
| active_title | VARCHAR(64) | 计算列（仅未删除记录的标题） |

### 索引

- **主键**: `id`
- **唯一索引**: `uk_org_active_title_date` (org_id, active_title, inquiry_date) - 保证同组织下未删除记录的标题+日期唯一
- **普通索引**:
  - `idx_org_valid_date` (org_id, is_deleted, inquiry_date)
  - `idx_org_title` (org_id, inquiry_title)
  - `idx_inquiry_date` (inquiry_date)
  - `idx_inquiry_org` (org_id)

### 约束

- `chk_time_order`: 结束时间必须晚于开始时间

---

## 错误处理

所有 API 在出错时会返回统一的错误格式：

```json
{
  "error": "错误标题",
  "message": "详细错误信息"
}
```

常见 HTTP 状态码：
- `400 Bad Request`: 请求参数错误
- `401 Unauthorized`: 未认证
- `403 Forbidden`: 无权限
- `404 Not Found`: 资源不存在
- `409 Conflict`: 数据冲突（如重复的标题+日期）
- `500 Internal Server Error`: 服务器内部错误

---

## 使用示例

### 使用 curl 创建询价记录

```bash
curl -X POST http://localhost:8080/api/v1/inquiries/create_inquiry \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "inquiry_title": "2024年10月询价单",
    "inquiry_date": "2024-10-20",
    "market_1": "北京新发地市场",
    "org_id": "550e8400-e29b-41d4-a716-446655440000",
    "inquiry_start_date": "2024-10-20T08:00:00Z",
    "inquiry_end_date": "2024-10-20T18:00:00Z"
  }'
```

### 使用 curl 查询列表

```bash
curl -X GET "http://localhost:8080/api/v1/inquiries/list_inquiries?org_id=550e8400-e29b-41d4-a716-446655440000&page=1&page_size=20" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

---

## 注意事项

1. **时间格式**:
   - `inquiry_date` 使用 `YYYY-MM-DD` 格式（如: 2024-10-20）
   - `inquiry_start_date` 和 `inquiry_end_date` 使用 RFC3339 格式（如: 2024-10-20T08:00:00Z）

2. **唯一性约束**:
   - 同一组织内，未删除的记录中，相同的标题和业务日期组合必须唯一
   - 软删除的记录不参与唯一性检查

3. **权限控制**:
   - 创建、更新、删除操作仅限管理员
   - 查询操作对所有已认证用户开放

4. **分页**:
   - 默认每页20条记录
   - 最大每页1000条记录
   - 结果按日期降序排列（最新的在前）
