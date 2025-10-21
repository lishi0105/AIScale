# 市场价格数据库表结构说明

## 概述

本文档说明如何使用MariaDB存储Excel市场价格数据。每个Excel sheet代表一个品类（category），如蔬菜类、水产类等。

## 数据库表结构

### 1. 核心表

#### `market_price_period` - 价格期次表
存储每个价格统计期次的基本信息。

**主要字段：**
- `title`: 期次标题（如：2025年9月上旬都匀市主要蔬菜类市场参考价）
- `period_year`: 年份（2025）
- `period_month`: 月份（1-12）
- `period_type`: 期次类型（上旬/中旬/下旬/月度）
- `category_id`: 品类ID（关联到蔬菜类、水产类等）
- `publish_date`: 发布日期

**唯一约束：**
同一组织、同一品类、同一年月期次唯一

#### `market_price_record` - 价格记录表
存储每个商品在某个期次的各个市场价格。

**主要字段：**
- `period_id`: 关联期次表
- `goods_id`: 关联商品表
- `goods_name`: 商品名称（冗余字段，便于查询）
- `spec_name`: 规格名称（如：新鲜）
- `unit_name`: 单位名称（如：斤）
- `seq_num`: Excel中的序号
- `ndrc_guide_price`: 发改委指导价
- `fuwanjia_price`: 富万家超市价格
- `yuying_market_price`: 育英巷菜市场价格
- `daxiangfa_price`: 大湘发价格
- `last_month_avg_price`: 上月均价
- `current_period_avg_price`: 本期均价
- `hupu_settlement_price`: 胡埔本期结算价（下浮12%）
- `guihai_settlement_price`: 贵海本期结算价（下浮14%）

**唯一约束：**
同一期次下同一商品唯一

#### `market_price_detail` - 价格明细表（可选）
提供更灵活的市场-价格关系存储方式。

**主要字段：**
- `period_id`: 关联期次表
- `goods_id`: 关联商品表
- `market_id`: 关联市场表
- `price`: 价格
- `price_type`: 价格类型（市场价/指导价/结算价等）

**使用场景：**
- 当市场数量和名称经常变化时
- 需要动态添加新市场时
- 需要记录同一商品在同一市场的多种价格类型时

### 2. 依赖表

#### `base_market` - 市场基础表
存储市场/超市的基本信息。

**示例数据：**
- 发改委
- 富万家超市
- 育英巷菜市场
- 大湘发
- 胡埔
- 贵海

#### `base_category` - 品类表
存储商品品类（对应Excel的每个sheet）。

**示例数据：**
- 蔬菜类
- 水产类
- 肉类
- 调味品
- 等等...

#### `base_goods` - 商品表
存储商品的基本信息（在`10_goods_domain.sql`中定义）。

## 数据导入流程

### 方案一：使用固定字段（market_price_record表）

适用于市场数量和名称相对固定的场景。

```sql
-- 1. 确保基础数据已存在
INSERT INTO base_market (id, name, org_id, code, sort) VALUES (UUID(), '富万家超市', ..., 'FUWANJIA', 1);
INSERT INTO base_category (id, name, org_id, code) VALUES (UUID(), '蔬菜类', ..., 'VEGETABLES');

-- 2. 创建期次
INSERT INTO market_price_period (
  id, title, period_year, period_month, period_type, category_id, org_id
) VALUES (
  UUID(), '2025年9月上旬都匀市主要蔬菜类市场参考价', 
  2025, 9, '上旬', 
  (SELECT id FROM base_category WHERE code = 'VEGETABLES'),
  (SELECT id FROM base_org LIMIT 1)
);

-- 3. 导入价格记录
INSERT INTO market_price_record (
  id, period_id, goods_id, goods_name, spec_name, unit_name, seq_num,
  fuwanjia_price, yuying_market_price, daxiangfa_price,
  last_month_avg_price, current_period_avg_price,
  hupu_settlement_price, guihai_settlement_price
) VALUES (
  UUID(),
  (SELECT id FROM market_price_period WHERE period_year = 2025 AND period_month = 9 AND period_type = '上旬'),
  (SELECT id FROM base_goods WHERE name = '四季豆'),
  '四季豆', '新鲜', '斤', 4,
  4.98, 5.00, 5.59,
  5.98, 5.19,
  4.57, 4.46
);
```

### 方案二：使用灵活字段（market_price_detail表）

适用于市场数量和名称经常变化的场景。

```sql
-- 1-2. 同方案一

-- 3. 导入价格明细（为每个市场创建单独的记录）
INSERT INTO market_price_detail (id, period_id, goods_id, market_id, price, price_type) VALUES
  (UUID(), [period_id], [goods_id], (SELECT id FROM base_market WHERE code = 'FUWANJIA'), 4.98, '市场价'),
  (UUID(), [period_id], [goods_id], (SELECT id FROM base_market WHERE code = 'YUYING'), 5.00, '市场价'),
  (UUID(), [period_id], [goods_id], (SELECT id FROM base_market WHERE code = 'HUPU'), 4.57, '结算价');
```

## Excel数据映射关系

### Sheet 1: 蔬菜类

| Excel列名 | 数据库字段 | 说明 |
|---------|---------|-----|
| 序号 | `seq_num` | 商品在Excel中的序号 |
| 品名 | `goods_name` | 商品名称 |
| 规格标准 | `spec_name` | 规格（如：新鲜） |
| 单位 | `unit_name` | 计量单位（如：斤） |
| 发改委指导价 | `ndrc_guide_price` | 发改委指导价 |
| 富万家超市 | `fuwanjia_price` | 富万家超市价格 |
| 育英巷菜市场 | `yuying_market_price` | 育英巷菜市场价格 |
| 大湘发 | `daxiangfa_price` | 大湘发价格 |
| 上月均价 | `last_month_avg_price` | 上月平均价格 |
| 本期均价 | `current_period_avg_price` | 本期平均价格 |
| 胡埔本期结算价(下浮12%) | `hupu_settlement_price` | 胡埔结算价 |
| 贵海本期结算价(下浮14%) | `guihai_settlement_price` | 贵海结算价 |

### Sheet 2: 水产类

数据结构与蔬菜类相同，通过`category_id`区分。

## 常用查询示例

### 查询某个期次的所有价格记录

```sql
SELECT 
  mpr.seq_num AS 序号,
  mpr.goods_name AS 品名,
  mpr.spec_name AS 规格标准,
  mpr.unit_name AS 单位,
  mpr.fuwanjia_price AS 富万家超市,
  mpr.yuying_market_price AS 育英巷菜市场,
  mpr.daxiangfa_price AS 大湘发,
  mpr.current_period_avg_price AS 本期均价,
  bc.name AS 品类
FROM market_price_record mpr
JOIN market_price_period mpp ON mpr.period_id = mpp.id
JOIN base_category bc ON mpp.category_id = bc.id
WHERE mpp.period_year = 2025 
  AND mpp.period_month = 9 
  AND mpp.period_type = '上旬'
  AND bc.code = 'VEGETABLES'
ORDER BY mpr.seq_num;
```

### 查询某个商品的价格历史

```sql
SELECT 
  mpp.period_year AS 年份,
  mpp.period_month AS 月份,
  mpp.period_type AS 期次,
  mpr.current_period_avg_price AS 本期均价,
  mpr.fuwanjia_price AS 富万家超市,
  bc.name AS 品类
FROM market_price_record mpr
JOIN market_price_period mpp ON mpr.period_id = mpp.id
JOIN base_category bc ON mpp.category_id = bc.id
WHERE mpr.goods_name = '四季豆'
ORDER BY mpp.period_year DESC, mpp.period_month DESC;
```

### 查询价格涨跌幅

```sql
SELECT 
  mpr.goods_name AS 品名,
  mpr.last_month_avg_price AS 上月均价,
  mpr.current_period_avg_price AS 本期均价,
  ROUND(
    (mpr.current_period_avg_price - mpr.last_month_avg_price) / mpr.last_month_avg_price * 100, 
    2
  ) AS 涨跌幅_百分比
FROM market_price_record mpr
JOIN market_price_period mpp ON mpr.period_id = mpp.id
WHERE mpp.period_year = 2025 
  AND mpp.period_month = 9 
  AND mpp.period_type = '上旬'
ORDER BY 涨跌幅_百分比 DESC;
```

## 表结构扩展建议

### 如果需要添加新的市场

**方案一（使用market_price_record）：**
```sql
ALTER TABLE market_price_record 
ADD COLUMN new_market_price DECIMAL(10,2) NULL COMMENT '新市场价格';
```

**方案二（使用market_price_detail）：**
```sql
-- 只需在base_market中添加新市场即可，无需修改表结构
INSERT INTO base_market (id, name, org_id, code) 
VALUES (UUID(), '新市场名称', [org_id], 'NEW_MARKET');
```

### 如果需要记录更多价格类型

建议使用`market_price_detail`表，它支持灵活的价格类型：
- 市场价
- 指导价
- 结算价
- 批发价
- 零售价
- 等等...

## 注意事项

1. **数据完整性**：
   - 导入价格数据前，确保相关的商品已在`base_goods`表中存在
   - 确保品类已在`base_category`表中存在
   - 确保市场已在`base_market`表中存在

2. **数据一致性**：
   - 同一期次、同一商品只能有一条价格记录
   - 期次的唯一性由（组织、品类、年、月、期次类型）共同决定

3. **性能优化**：
   - 为常用查询字段创建了索引
   - 冗余存储了商品名称、规格、单位，避免频繁JOIN

4. **软删除**：
   - 所有表都支持软删除（`is_deleted`字段）
   - 删除数据时建议使用软删除而非物理删除

5. **时间戳**：
   - 所有表都自动记录创建时间和更新时间
   - 便于数据审计和追溯

## 文件清单

- `20_market_price.sql` - 市场价格表结构定义
- `21_market_price_sample_data.sql` - 示例数据和插入脚本
- `README_market_price.md` - 本说明文档
