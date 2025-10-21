# 市场价格管理 API 文档

本文档描述了市场价格管理相关的 API 接口，包括市场、询价单、询价商品明细、市场报价和供应商结算。

## 基础信息

- Base URL: `/api/v1`
- 所有接口都需要 JWT 认证
- 创建、更新、删除操作仅限管理员

---

## 1. 市场管理 (Market)

### 1.1 创建市场
**POST** `/market/create_market`

**请求体：**
```json
{
  "name": "富万家市场",
  "org_id": "uuid",
  "code": "FWJ001",  // 可选，不提供则自动生成
  "sort": 100        // 可选，不提供则自动生成
}
```

**响应：**
```json
{
  "id": "uuid",
  "name": "富万家市场",
  "org_id": "uuid",
  "code": "FWJ001",
  "sort": 100,
  "is_deleted": 0,
  "created_at": "2025-01-01T00:00:00Z",
  "updated_at": "2025-01-01T00:00:00Z"
}
```

### 1.2 获取市场详情
**POST** `/market/get_market`

**请求体：**
```json
{
  "id": "uuid"
}
```

### 1.3 获取市场列表
**POST** `/market/list_markets`

**查询参数：**
- `org_id` (必需): 组织ID
- `keyword` (可选): 搜索关键词（名称、编码）
- `page` (可选，默认1): 页码
- `page_size` (可选，默认20): 每页数量

**响应：**
```json
{
  "total": 100,
  "items": [...]
}
```

### 1.4 更新市场
**POST** `/market/update_market`

**请求体：**
```json
{
  "id": "uuid",
  "name": "新市场名称",  // 可选
  "code": "NEW001",     // 可选
  "sort": 200,          // 可选
  "org_id": "uuid"      // 可选
}
```

### 1.5 软删除市场
**POST** `/market/soft_delete_market`

### 1.6 硬删除市场
**POST** `/market/hard_delete_market`

---

## 2. 询价单管理 (Price Inquiry)

### 2.1 创建询价单
**POST** `/price_inquiry/create_price_inquiry`

**请求体：**
```json
{
  "org_id": "uuid",
  "inquiry_title": "2025年9月上旬都匀市主要水产类市场参考价",
  "inquiry_date": "2025-09-05"  // YYYY-MM-DD 格式
}
```

**响应：**
```json
{
  "id": "uuid",
  "org_id": "uuid",
  "inquiry_title": "...",
  "inquiry_date": "2025-09-05T00:00:00Z",
  "inquiry_year": 2025,
  "inquiry_month": 9,
  "inquiry_ten_day": 1,  // 1=上旬, 2=中旬, 3=下旬
  "is_deleted": 0,
  "created_at": "...",
  "updated_at": "..."
}
```

### 2.2 获取询价单详情
**POST** `/price_inquiry/get_price_inquiry`

### 2.3 获取询价单列表
**POST** `/price_inquiry/list_price_inquiries`

**查询参数：**
- `org_id` (必需): 组织ID
- `keyword` (可选): 搜索关键词（标题）
- `year` (可选): 年份
- `month` (可选): 月份
- `ten_day` (可选): 旬 (1/2/3)
- `page` (可选，默认1): 页码
- `page_size` (可选，默认20): 每页数量

### 2.4 更新询价单
**POST** `/price_inquiry/update_price_inquiry`

**请求体：**
```json
{
  "id": "uuid",
  "org_id": "uuid",           // 可选
  "inquiry_title": "新标题",  // 可选
  "inquiry_date": "2025-09-10" // 可选
}
```

### 2.5 软删除询价单
**POST** `/price_inquiry/soft_delete_price_inquiry`

### 2.6 硬删除询价单
**POST** `/price_inquiry/hard_delete_price_inquiry`

---

## 3. 询价商品明细 (Inquiry Item)

### 3.1 创建询价商品明细
**POST** `/inquiry_item/create_inquiry_item`

**请求体：**
```json
{
  "inquiry_id": "uuid",
  "goods_id": "uuid",
  "category_id": "uuid",
  "spec_id": "uuid",           // 可选
  "unit_id": "uuid",           // 可选
  "goods_name_snap": "鲜鱼",
  "category_name_snap": "水产",
  "spec_name_snap": "新鲜",    // 可选
  "unit_name_snap": "斤",      // 可选
  "guide_price": 25.50,        // 可选，发改委指导价
  "last_month_avg_price": 24.00, // 可选，上月均价
  "current_avg_price": 25.00,  // 可选，本期均价
  "sort": 100                  // 可选
}
```

### 3.2 获取询价商品明细详情
**POST** `/inquiry_item/get_inquiry_item`

### 3.3 获取询价商品明细列表
**POST** `/inquiry_item/list_inquiry_items`

**查询参数：**
- `inquiry_id` (必需): 询价单ID
- `category_id` (可选): 品类ID
- `page` (可选，默认1): 页码
- `page_size` (可选，默认20): 每页数量

### 3.4 更新询价商品明细
**POST** `/inquiry_item/update_inquiry_item`

### 3.5 软删除询价商品明细
**POST** `/inquiry_item/soft_delete_inquiry_item`

### 3.6 硬删除询价商品明细
**POST** `/inquiry_item/hard_delete_inquiry_item`

---

## 4. 市场报价 (Market Inquiry)

### 4.1 创建市场报价
**POST** `/market_inquiry/create_market_inquiry`

**请求体：**
```json
{
  "inquiry_id": "uuid",
  "item_id": "uuid",
  "market_id": "uuid",         // 可选
  "market_name_snap": "富万家市场",
  "price": 26.50
}
```

**说明：**
- 同一个询价商品明细可以有多个市场的报价
- `market_name_snap` 和 `item_id` 组合必须唯一

### 4.2 获取市场报价详情
**POST** `/market_inquiry/get_market_inquiry`

### 4.3 获取市场报价列表
**POST** `/market_inquiry/list_market_inquiries`

**查询参数：**
- `inquiry_id` (可选): 询价单ID
- `item_id` (可选): 询价明细ID
- `page` (可选，默认1): 页码
- `page_size` (可选，默认20): 每页数量

### 4.4 更新市场报价
**POST** `/market_inquiry/update_market_inquiry`

### 4.5 删除市场报价
**POST** `/market_inquiry/delete_market_inquiry`

**说明：** 市场报价使用物理删除（硬删除）

---

## 5. 供应商结算 (Supplier Settlement)

### 5.1 创建供应商结算
**POST** `/supplier_settlement/create_supplier_settlement`

**请求体：**
```json
{
  "inquiry_id": "uuid",
  "item_id": "uuid",
  "supplier_id": "uuid",       // 可选
  "supplier_name_snap": "胡坤供应商",
  "float_ratio_snap": 0.88,    // 浮动比例，0.88表示下浮12%
  "settlement_price": 22.00    // 结算价
}
```

**说明：**
- 同一个询价商品明细可以有多个供应商的结算价
- `supplier_name_snap` 和 `item_id` 组合必须唯一

### 5.2 获取供应商结算详情
**POST** `/supplier_settlement/get_supplier_settlement`

### 5.3 获取供应商结算列表
**POST** `/supplier_settlement/list_supplier_settlements`

**查询参数：**
- `inquiry_id` (可选): 询价单ID
- `item_id` (可选): 询价明细ID
- `page` (可选，默认1): 页码
- `page_size` (可选，默认20): 每页数量

### 5.4 更新供应商结算
**POST** `/supplier_settlement/update_supplier_settlement`

### 5.5 删除供应商结算
**POST** `/supplier_settlement/delete_supplier_settlement`

**说明：** 供应商结算使用物理删除（硬删除）

---

## 数据模型关系

```
base_price_inquiry (询价单)
    ↓ 一对多
price_inquiry_item (询价商品明细)
    ↓ 一对多
    ├─ price_market_inquiry (市场报价)
    └─ price_supplier_settlement (供应商结算)
```

## 典型业务流程

1. **创建询价单**：创建一个询价单，指定标题和日期
2. **添加商品明细**：为询价单添加需要询价的商品
3. **录入市场报价**：为每个商品明细添加多个市场的报价
4. **计算均价**：根据市场报价计算本期均价（可手动或自动）
5. **生成供应商结算**：基于均价和浮动比例生成供应商结算价

## 注意事项

1. 所有快照字段（`*_snap`）用于保存历史数据，防止后续修改影响
2. 询价单的年月旬字段为数据库生成列，会自动计算
3. 市场报价和供应商结算采用物理删除（硬删除）
4. 询价单和询价明细采用软删除，支持恢复
5. 市场基础数据支持软删除和硬删除两种方式
