# 价格询价数据表设计说明

## 概述

本文档描述价格询价系统的数据库表结构设计，用于记录每月上中下旬对商品库的商品进行市场询价的数据。

## 业务场景

每月上中下旬都会对商品库的所有商品随机选择几个市场做一次询价并做记录：
- **询价频率**：每月3次（上旬、中旬、下旬）
- **询价对象**：商品库中的商品（按品类组织）
- **询价市场**：随机选择的市场（如：富万家超市、育英巷菜市场、大润发等）
- **供应商**：记录供应商的结算价格和下浮比例（如：胡坤、贵海等）
- **灵活性**：每次询价的市场数量、供应商数量和名称都是变化的

## 数据表结构

### 1. price_inquiry（价格询价主表）

记录每次询价的基本信息。

**主要字段**：
- `id`: 主键UUID
- `title`: 询价标题（如："2025年9月上旬都匀市主要蔬菜类市场参考价"）
- `year`: 年份（如：2025）
- `month`: 月份（1-12）
- `period`: 旬次（上旬/中旬/下旬）
- `category_id`: 品类ID（关联`base_category`表）
- `org_id`: 中队ID
- `inquiry_date`: 询价日期
- `remark`: 备注信息

**索引**：
- 按日期查询：`idx_inquiry_date (year, month, period)`
- 按品类查询：`idx_inquiry_category (category_id)`
- 按标题模糊搜索：`idx_inquiry_title (title)`

**示例数据**：
```sql
INSERT INTO price_inquiry (id, title, year, month, period, category_id, org_id, inquiry_date)
VALUES ('inquiry-001', '2025年9月上旬都匀市主要蔬菜类市场参考价', 2025, 9, '上旬', 'category-001', 'org-001', '2025-09-10');
```

---

### 2. price_inquiry_detail（价格询价商品明细表）

记录每个商品的价格信息（对应Excel中的每一行）。

**主要字段**：
- `id`: 主键UUID
- `inquiry_id`: 询价主表ID（关联`price_inquiry`）
- `goods_id`: 商品ID（关联`base_goods`）
- `sequence`: 序号（对应Excel中的序号）
- `guide_price`: 发改委指导价
- `last_month_avg_price`: 上月均价
- `current_avg_price`: 本期均价

**约束**：
- 同一次询价中，同一商品只能出现一次：`uq_detail_inquiry_goods (inquiry_id, goods_id)`
- 价格必须为正数或NULL

**示例数据**：
```sql
INSERT INTO price_inquiry_detail (id, inquiry_id, goods_id, sequence, guide_price, last_month_avg_price, current_avg_price)
VALUES ('detail-001', 'inquiry-001', 'goods-kedou-001', 1, 3.98, 5.79, 4.65);
```

---

### 3. price_inquiry_market_detail（市场价格明细表）

记录每个商品在各个市场的具体价格。

**主要字段**：
- `id`: 主键UUID
- `inquiry_detail_id`: 询价商品明细ID（关联`price_inquiry_detail`）
- `market_id`: 市场ID（关联`base_market`）
- `price`: 市场价格

**约束**：
- 同一商品在同一市场只能有一个价格：`uq_market_detail_inquiry_market (inquiry_detail_id, market_id)`

**示例数据**：
```sql
-- 可豆在富万家超市的价格
INSERT INTO price_inquiry_market_detail (id, inquiry_detail_id, market_id, price)
VALUES ('market-detail-001', 'detail-001', 'market-fuwanjia-001', 4.0);

-- 可豆在育英巷菜市场的价格
INSERT INTO price_inquiry_market_detail (id, inquiry_detail_id, market_id, price)
VALUES ('market-detail-002', 'detail-001', 'market-yuyingxiang-001', 4.5);
```

---

### 4. price_inquiry_supplier_detail（供应商结算价明细表）

记录供应商的结算价格和下浮比例。

**主要字段**：
- `id`: 主键UUID
- `inquiry_detail_id`: 询价商品明细ID（关联`price_inquiry_detail`）
- `supplier_id`: 供应商ID（关联`supplier`）
- `settlement_price`: 结算价格
- `float_ratio`: 下浮比例（如：0.88表示下浮12%，0.86表示下浮14%）

**约束**：
- 同一商品的同一供应商只能有一个结算价：`uq_supplier_detail_inquiry_supplier (inquiry_detail_id, supplier_id)`
- 下浮比例必须在0-1之间

**示例数据**：
```sql
-- 胡坤供应商的结算价（下浮12%）
INSERT INTO price_inquiry_supplier_detail (id, inquiry_detail_id, supplier_id, settlement_price, float_ratio, remark)
VALUES ('supplier-detail-001', 'detail-001', 'supplier-hukun-001', 4.09, 0.88, '胡坤本期结算价（下浮12%）');

-- 贵海供应商的结算价（下浮14%）
INSERT INTO price_inquiry_supplier_detail (id, inquiry_detail_id, supplier_id, settlement_price, float_ratio, remark)
VALUES ('supplier-detail-002', 'detail-001', 'supplier-guihai-001', 4.00, 0.86, '贵海本期结算价（下浮14%）');
```

---

## 视图

### v_price_inquiry_full（完整询价信息视图）

将询价主表、商品明细、市场价格、供应商结算价等信息关联在一起，方便查询和展示。

**使用示例**：
```sql
-- 查询某次询价的完整信息
SELECT * FROM v_price_inquiry_full WHERE inquiry_id = 'inquiry-001';
```

---

## 数据关系图

```
price_inquiry (询价主表)
    ↓ (1:N)
price_inquiry_detail (商品明细)
    ↓ (1:N)                    ↓ (1:N)
price_inquiry_market_detail   price_inquiry_supplier_detail
(市场价格明细)                 (供应商结算价明细)
    ↓                             ↓
base_market (市场)            supplier (供应商)
```

---

## 常用查询场景

### 1. 按日期查询询价记录

```sql
-- 查询2025年9月上旬的所有询价记录
SELECT * FROM price_inquiry 
WHERE year = 2025 AND month = 9 AND period = '上旬'
ORDER BY inquiry_date DESC;
```

### 2. 按品类查询询价记录

```sql
-- 查询蔬菜类的所有询价记录
SELECT pi.*, bc.name AS category_name
FROM price_inquiry pi
INNER JOIN base_category bc ON pi.category_id = bc.id
WHERE bc.name = '蔬菜类'
ORDER BY pi.inquiry_date DESC;
```

### 3. 模糊搜索询价标题

```sql
-- 搜索标题中包含"蔬菜类"或"水产类"的询价记录
SELECT * FROM price_inquiry 
WHERE title LIKE '%蔬菜类%' OR title LIKE '%水产类%'
ORDER BY inquiry_date DESC;
```

### 4. 查询某次询价的完整信息

```sql
-- 查询某次询价的所有商品、市场价格和供应商结算价
SELECT 
  pid.sequence AS 序号,
  bg.name AS 品名,
  bs.name AS 规格标准,
  bu.name AS 单位,
  pid.guide_price AS 发改委指导价,
  bm.name AS 市场名称,
  pimd.price AS 市场价格,
  pid.current_avg_price AS 本期均价,
  s.name AS 供应商名称,
  pisd.settlement_price AS 结算价,
  pisd.float_ratio AS 下浮比例
FROM price_inquiry pi
INNER JOIN price_inquiry_detail pid ON pi.id = pid.inquiry_id
INNER JOIN base_goods bg ON pid.goods_id = bg.id
INNER JOIN base_spec bs ON bg.spec_id = bs.id
INNER JOIN base_unit bu ON bg.unit_id = bu.id
LEFT JOIN price_inquiry_market_detail pimd ON pid.id = pimd.inquiry_detail_id
LEFT JOIN base_market bm ON pimd.market_id = bm.id
LEFT JOIN price_inquiry_supplier_detail pisd ON pid.id = pisd.inquiry_detail_id
LEFT JOIN supplier s ON pisd.supplier_id = s.id
WHERE pi.id = 'inquiry-001'
ORDER BY pid.sequence ASC;
```

### 5. 查询某个商品的价格变化趋势

```sql
-- 查询可豆在2025年的价格变化趋势
SELECT 
  pi.year, pi.month, pi.period,
  pid.current_avg_price AS 本期均价,
  bm.name AS 市场名称,
  pimd.price AS 市场价格
FROM price_inquiry pi
INNER JOIN price_inquiry_detail pid ON pi.id = pid.inquiry_id
INNER JOIN base_goods bg ON pid.goods_id = bg.id
LEFT JOIN price_inquiry_market_detail pimd ON pid.id = pimd.inquiry_detail_id
LEFT JOIN base_market bm ON pimd.market_id = bm.id
WHERE bg.name = '可豆' AND pi.year = 2025
ORDER BY pi.month ASC, 
  CASE pi.period 
    WHEN '上旬' THEN 1 
    WHEN '中旬' THEN 2 
    WHEN '下旬' THEN 3 
  END ASC;
```

### 6. 市场均价与供应商结算价对比

```sql
-- 查询某次询价中所有商品的市场均价和供应商结算价对比
SELECT 
  pid.sequence AS 序号,
  bg.name AS 商品名称,
  pid.current_avg_price AS 市场均价,
  s.name AS 供应商名称,
  pisd.settlement_price AS 结算价,
  ROUND((1 - pisd.float_ratio) * 100, 2) AS 下浮百分比,
  ROUND((pid.current_avg_price - pisd.settlement_price) / pid.current_avg_price * 100, 2) AS 折扣百分比
FROM price_inquiry pi
INNER JOIN price_inquiry_detail pid ON pi.id = pid.inquiry_id
INNER JOIN base_goods bg ON pid.goods_id = bg.id
LEFT JOIN price_inquiry_supplier_detail pisd ON pid.id = pisd.inquiry_detail_id
LEFT JOIN supplier s ON pisd.supplier_id = s.id
WHERE pi.id = 'inquiry-001'
ORDER BY pid.sequence ASC, s.name ASC;
```

---

## 数据导入流程

从Excel导入询价数据的步骤：

1. **准备基础数据**：
   - 确保`base_category`（品类）、`base_goods`（商品）、`base_market`（市场）、`supplier`（供应商）等基础数据已存在

2. **插入询价主表**：
   - 根据Excel的标题（如："2025年9月上旬都匀市主要蔬菜类市场参考价"）解析年、月、旬次、品类等信息
   - 插入一条`price_inquiry`记录

3. **插入商品明细**：
   - 遍历Excel的每一行（每个商品）
   - 插入`price_inquiry_detail`记录，包括序号、发改委指导价、上月均价、本期均价

4. **插入市场价格明细**：
   - 对于每个商品，遍历Excel中的市场列（如：富万家、育英巷、大润发）
   - 为每个有价格的市场插入`price_inquiry_market_detail`记录

5. **插入供应商结算价明细**：
   - 对于每个商品，遍历Excel中的供应商列（如：胡坤、贵海）
   - 为每个供应商插入`price_inquiry_supplier_detail`记录，包括结算价和下浮比例

---

## 注意事项

1. **UUID生成**：所有表的主键都使用UUID，可以使用`UUID()`函数生成

2. **级联删除**：
   - 删除询价主表记录时，会级联删除相关的商品明细、市场价格明细、供应商结算价明细
   - 使用软删除（`is_deleted`字段）而非物理删除

3. **价格精度**：所有价格字段使用`DECIMAL(10,2)`类型，保留2位小数

4. **下浮比例**：
   - `float_ratio`表示结算价相对于市场价的比例
   - 例如：下浮12%表示`float_ratio = 0.88`（即1 - 0.12）
   - 例如：下浮14%表示`float_ratio = 0.86`（即1 - 0.14）

5. **灵活性**：
   - 市场和供应商的数量不固定，每次询价可以选择不同的市场和供应商
   - 不是每个商品都必须有所有市场的价格或所有供应商的结算价

6. **查询性能**：
   - 已为常用查询场景创建索引
   - 对于复杂查询，建议使用`v_price_inquiry_full`视图

---

## 文件清单

- `30_price_inquiry.sql`: 价格询价表结构定义
- `31_price_inquiry_examples.sql`: 示例数据和常用查询
- `README_PRICE_INQUIRY.md`: 本说明文档

---

## 依赖关系

本模块依赖以下基础表：
- `base_category` (品类表) - 定义在 `10_goods_domain.sql`
- `base_goods` (商品表) - 定义在 `10_goods_domain.sql`
- `base_spec` (规格表) - 定义在 `01_base_sys.sql`
- `base_unit` (单位表) - 定义在 `01_base_sys.sql`
- `base_market` (市场表) - 定义在 `20_market_price.sql`
- `supplier` (供应商表) - 定义在 `10_goods_domain.sql`
- `base_org` (组织表) - 定义在 `01_base_sys.sql`

执行顺序：
1. `00_db_users.sql`
2. `01_base_sys.sql`
3. `10_goods_domain.sql`
4. `20_market_price.sql`
5. `30_price_inquiry.sql` ← 本模块
6. `31_price_inquiry_examples.sql` ← 示例数据（可选）
