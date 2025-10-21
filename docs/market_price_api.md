# 市场价格管理 API 文档

本文档描述了市场价格管理系统的 CRUD API 接口。

## 1. 基础市场管理 (BaseMarket)

### 1.1 创建市场
- **URL**: `POST /api/v1/market/create_market`
- **请求体**:
```json
{
  "name": "富万家市场",
  "org_id": "uuid",
  "code": "optional",
  "sort": 0
}
```

### 1.2 获取市场
- **URL**: `POST /api/v1/market/get_market`
- **请求体**:
```json
{
  "id": "uuid"
}
```

### 1.3 列出市场
- **URL**: `POST /api/v1/market/list_markets`
- **查询参数**: 
  - `org_id` (必需)
  - `keyword` (可选)
  - `page` (默认: 1)
  - `page_size` (默认: 20)

### 1.4 更新市场
- **URL**: `POST /api/v1/market/update_market`
- **请求体**:
```json
{
  "id": "uuid",
  "name": "optional",
  "code": "optional",
  "sort": 0
}
```

### 1.5 软删除市场
- **URL**: `POST /api/v1/market/soft_delete_market`

### 1.6 硬删除市场
- **URL**: `POST /api/v1/market/hard_delete_market`

---

## 2. 询价单管理 (BasePriceInquiry)

### 2.1 创建询价单
- **URL**: `POST /api/v1/inquiry/create_inquiry`
- **请求体**:
```json
{
  "org_id": "uuid",
  "inquiry_title": "2025年9月上旬都匀市主要水产类市场参考价",
  "inquiry_date": "2025-09-05"
}
```

### 2.2 获取询价单
- **URL**: `POST /api/v1/inquiry/get_inquiry`

### 2.3 列出询价单
- **URL**: `POST /api/v1/inquiry/list_inquiries`
- **查询参数**:
  - `org_id` (必需)
  - `year` (可选)
  - `month` (可选)
  - `ten_day` (可选: 1=上旬, 2=中旬, 3=下旬)
  - `keyword` (可选)
  - `page`, `page_size`

### 2.4 更新询价单
- **URL**: `POST /api/v1/inquiry/update_inquiry`
- **请求体**:
```json
{
  "id": "uuid",
  "inquiry_title": "optional",
  "inquiry_date": "optional (YYYY-MM-DD)"
}
```

### 2.5 软删除询价单
- **URL**: `POST /api/v1/inquiry/soft_delete_inquiry`

### 2.6 硬删除询价单
- **URL**: `POST /api/v1/inquiry/hard_delete_inquiry`

---

## 3. 询价商品明细管理 (PriceInquiryItem)

### 3.1 创建询价商品明细
- **URL**: `POST /api/v1/inquiry_item/create_inquiry_item`
- **请求体**:
```json
{
  "inquiry_id": "uuid",
  "goods_id": "uuid",
  "category_id": "uuid",
  "spec_id": "optional uuid",
  "unit_id": "optional uuid",
  "goods_name_snap": "鲤鱼",
  "category_name_snap": "水产类",
  "spec_name_snap": "新鲜",
  "unit_name_snap": "斤",
  "guide_price": 12.5,
  "last_month_avg_price": 12.0,
  "current_avg_price": 12.3,
  "sort": 0
}
```

### 3.2 获取询价商品明细
- **URL**: `POST /api/v1/inquiry_item/get_inquiry_item`

### 3.3 列出询价商品明细
- **URL**: `POST /api/v1/inquiry_item/list_inquiry_items`
- **查询参数**:
  - `inquiry_id` (必需)
  - `category_id` (可选)
  - `page`, `page_size`

### 3.4 更新询价商品明细
- **URL**: `POST /api/v1/inquiry_item/update_inquiry_item`

### 3.5 软删除询价商品明细
- **URL**: `POST /api/v1/inquiry_item/soft_delete_inquiry_item`

### 3.6 硬删除询价商品明细
- **URL**: `POST /api/v1/inquiry_item/hard_delete_inquiry_item`

---

## 4. 市场报价管理 (PriceMarketInquiry)

### 4.1 创建市场报价
- **URL**: `POST /api/v1/market_inquiry/create_market_inquiry`
- **请求体**:
```json
{
  "inquiry_id": "uuid",
  "item_id": "uuid",
  "market_id": "optional uuid",
  "market_name_snap": "富万家",
  "price": 12.5
}
```

### 4.2 获取市场报价
- **URL**: `POST /api/v1/market_inquiry/get_market_inquiry`

### 4.3 列出市场报价
- **URL**: `POST /api/v1/market_inquiry/list_market_inquiries`
- **查询参数**:
  - `inquiry_id` (可选)
  - `item_id` (可选)
  - `page`, `page_size`

### 4.4 更新市场报价
- **URL**: `POST /api/v1/market_inquiry/update_market_inquiry`

### 4.5 软删除市场报价
- **URL**: `POST /api/v1/market_inquiry/soft_delete_market_inquiry`

### 4.6 硬删除市场报价
- **URL**: `POST /api/v1/market_inquiry/hard_delete_market_inquiry`

---

## 5. 供应商结算管理 (PriceSupplierSettlement)

### 5.1 创建供应商结算
- **URL**: `POST /api/v1/supplier_settlement/create_supplier_settlement`
- **请求体**:
```json
{
  "inquiry_id": "uuid",
  "item_id": "uuid",
  "supplier_id": "optional uuid",
  "supplier_name_snap": "胡坤",
  "float_ratio_snap": 0.88,
  "settlement_price": 10.8
}
```

### 5.2 获取供应商结算
- **URL**: `POST /api/v1/supplier_settlement/get_supplier_settlement`

### 5.3 列出供应商结算
- **URL**: `POST /api/v1/supplier_settlement/list_supplier_settlements`
- **查询参数**:
  - `inquiry_id` (可选)
  - `item_id` (可选)
  - `page`, `page_size`

### 5.4 更新供应商结算
- **URL**: `POST /api/v1/supplier_settlement/update_supplier_settlement`

### 5.5 软删除供应商结算
- **URL**: `POST /api/v1/supplier_settlement/soft_delete_supplier_settlement`

### 5.6 硬删除供应商结算
- **URL**: `POST /api/v1/supplier_settlement/hard_delete_supplier_settlement`

---

## 权限说明

- 所有接口都需要登录认证（JWT Token）
- 创建、更新、删除操作仅限管理员（RoleAdmin）
- 查询操作对所有已登录用户开放

## 数据模型特点

1. **BaseMarket**: 支持自动生成 code 和 sort 字段
2. **BasePriceInquiry**: 包含自动生成的年/月/旬字段，便于检索
3. **PriceInquiryItem**: 支持商品信息快照，防止历史数据被修改影响
4. **PriceMarketInquiry**: 支持市场名称快照，允许临时市场报价
5. **PriceSupplierSettlement**: 支持浮动比例和结算价计算

## 实现文件结构

```
internal/
├── domain/market/          # 数据模型定义
│   └── model.go
├── repository/market/      # 数据库访问层
│   ├── repo.go            # 接口定义
│   └── repo_gorm.go       # GORM 实现
├── service/market/         # 业务逻辑层
│   └── service.go
└── server/
    ├── handler/           # HTTP 处理器
    │   └── market.go
    └── server.go          # 路由注册
```
