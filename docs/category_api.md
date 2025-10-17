# 商品品类 API 文档

## 概述

商品品类（Category）模块提供了完整的 CRUD 操作，用于管理商品品类信息（如蔬菜、肉类、调味品等）。

## 数据模型

```go
type Category struct {
    ID        string    // 主键UUID
    Name      string    // 品类名称（唯一）
    Code      *string   // 品类编码（可选，建议唯一）
    Pinyin    *string   // 拼音（可选，用于搜索）
    IsDeleted int       // 软删标记：0=有效,1=已删除
    CreatedAt time.Time // 创建时间
    UpdatedAt time.Time // 更新时间
}
```

## API 端点

### 1. 创建品类

**请求**
```http
POST /api/v1/category/create_category
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "蔬菜",
  "code": "VEG",
  "pinyin": "shucai"
}
```

**响应**
```json
{
  "ID": "550e8400-e29b-41d4-a716-446655440000",
  "Name": "蔬菜",
  "Code": "VEG",
  "Pinyin": "shucai",
  "IsDeleted": 0,
  "CreatedAt": "2025-10-17T10:30:00Z",
  "UpdatedAt": "2025-10-17T10:30:00Z"
}
```

**说明**
- 仅管理员可创建
- `name` 必填，最大64字符
- `code` 和 `pinyin` 可选，最大64字符

---

### 2. 获取品类详情

**请求**
```http
POST /api/v1/category/get_category
Authorization: Bearer <token>
Content-Type: application/json

{
  "id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**响应**
```json
{
  "ID": "550e8400-e29b-41d4-a716-446655440000",
  "Name": "蔬菜",
  "Code": "VEG",
  "Pinyin": "shucai",
  "IsDeleted": 0,
  "CreatedAt": "2025-10-17T10:30:00Z",
  "UpdatedAt": "2025-10-17T10:30:00Z"
}
```

---

### 3. 获取品类列表

**请求**
```http
POST /api/v1/category/list_category?page=1&page_size=20&keyword=蔬菜
Authorization: Bearer <token>
```

**参数**
- `page`: 页码（默认1）
- `page_size`: 每页数量（默认20，最大1000）
- `keyword`: 搜索关键词（可选，支持按名称、编码、拼音模糊搜索）

**响应**
```json
{
  "total": 100,
  "items": [
    {
      "ID": "550e8400-e29b-41d4-a716-446655440000",
      "Name": "蔬菜",
      "Code": "VEG",
      "Pinyin": "shucai",
      "IsDeleted": 0,
      "CreatedAt": "2025-10-17T10:30:00Z",
      "UpdatedAt": "2025-10-17T10:30:00Z"
    },
    ...
  ]
}
```

---

### 4. 更新品类

**请求**
```http
POST /api/v1/category/update_category
Authorization: Bearer <token>
Content-Type: application/json

{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "蔬菜类",
  "code": "VEG01",
  "pinyin": "shucailei"
}
```

**响应**
```http
HTTP/1.1 204 No Content
```

**说明**
- 仅管理员可更新
- `id` 必填，必须是有效的UUID
- `name` 必填，最大64字符
- `code` 和 `pinyin` 可选，最大64字符

---

### 5. 删除品类（软删除）

**请求**
```http
POST /api/v1/category/udelete_category
Authorization: Bearer <token>
Content-Type: application/json

{
  "id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**响应**
```http
HTTP/1.1 204 No Content
```

**说明**
- 仅管理员可删除
- 采用软删除方式，设置 `is_deleted = 1`
- 删除后的品类不会出现在列表中

---

## 错误响应

### 401 Unauthorized
```json
{
  "error": "未授权"
}
```

### 403 Forbidden
```json
{
  "title": "创建品类失败",
  "message": "仅管理员可新增品类"
}
```

### 400 Bad Request
```json
{
  "title": "创建品类失败",
  "message": "输入格式非法"
}
```

### 404 Not Found
```json
{
  "title": "获取品类失败",
  "message": "品类不存在"
}
```

### 409 Conflict
```json
{
  "title": "创建品类失败",
  "message": "添加品类失败: Error 1062: Duplicate entry '蔬菜' for key 'uq_category_name'"
}
```

---

## 使用示例

### cURL 示例

```bash
# 1. 创建品类
curl -X POST http://localhost:8080/api/v1/category/create_category \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "蔬菜",
    "code": "VEG",
    "pinyin": "shucai"
  }'

# 2. 获取品类列表
curl -X POST "http://localhost:8080/api/v1/category/list_category?page=1&page_size=20&keyword=蔬菜" \
  -H "Authorization: Bearer YOUR_TOKEN"

# 3. 获取品类详情
curl -X POST http://localhost:8080/api/v1/category/get_category \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "550e8400-e29b-41d4-a716-446655440000"
  }'

# 4. 更新品类
curl -X POST http://localhost:8080/api/v1/category/update_category \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "蔬菜类",
    "code": "VEG01",
    "pinyin": "shucailei"
  }'

# 5. 删除品类
curl -X POST http://localhost:8080/api/v1/category/udelete_category \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "550e8400-e29b-41d4-a716-446655440000"
  }'
```

---

## 注意事项

1. **权限要求**：所有操作都需要身份验证，创建、更新、删除操作仅限管理员
2. **唯一性约束**：`name` 和 `code` 字段都有唯一性约束
3. **软删除**：删除操作不会真正删除数据，只是标记 `is_deleted = 1`
4. **搜索功能**：列表查询支持按 `name`、`code`、`pinyin` 进行模糊搜索
5. **分页**：列表查询支持分页，最大每页1000条记录

---

## 技术实现

### 项目结构
```
internal/
├── domain/category/model.go          # 领域模型
├── repository/category/
│   ├── repo.go                       # 仓储接口
│   └── repo_gorm.go                  # GORM实现
├── service/category/service.go       # 业务逻辑层
└── server/handler/category.go        # HTTP处理层
```

### 数据库表结构
```sql
CREATE TABLE IF NOT EXISTS base_category (
  id          CHAR(36)     NOT NULL COMMENT '主键UUID',
  name        VARCHAR(64)  NOT NULL COMMENT '品类名称（唯一）',
  code        VARCHAR(64)      NULL COMMENT '品类编码（可选，建议唯一）',
  pinyin      VARCHAR(64)      NULL COMMENT '拼音（可选，用于搜索）',
  is_deleted  TINYINT(1)   NOT NULL DEFAULT 0 COMMENT '软删标记：0=有效,1=已删除',
  created_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uq_category_name (name),
  UNIQUE KEY uq_category_code (code)
) ENGINE=InnoDB COMMENT='商品品类（如 蔬菜/肉类/调味品 等）';
```
