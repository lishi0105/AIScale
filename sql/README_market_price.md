# 市场价格管理系统使用说明

## 概述

本系统用于记录和管理商品在不同市场、不同时期的价格信息，支持：
- 多市场价格记录（政府指导价、超市、菜市场等）
- 多时期价格对比（上旬、中旬、下旬、月度等）
- 供应商结算价格管理（支持浮动比例）
- Excel批量导入功能

## 数据库表结构

### 1. 市场/渠道字典 (base_market)
记录不同的市场来源和渠道。

**市场类型**：
- 1 = 政府指导（如发改委）
- 2 = 超市（如富万家、大润发）
- 3 = 菜市场（如育英巷菜市场）
- 4 = 批发市场
- 5 = 其他

### 2. 价格时期字典 (base_price_period)
记录价格采集的时间段。

**时期类型**：
- 1 = 上旬
- 2 = 中旬
- 3 = 下旬
- 4 = 月度
- 5 = 季度
- 6 = 年度

### 3. 市场价格记录表 (base_market_price)
记录商品在不同市场、不同时期的价格。

**价格类型**：
- 1 = 市场价
- 2 = 指导价
- 3 = 上月均价
- 4 = 本期均价

**特点**：
- 同一商品+市场+时期+价格类型只允许一条记录
- 支持价格查询和对比分析

### 4. 供应商结算价格表 (base_supplier_price)
记录供应商的结算价格，支持浮动比例。

**计算公式**：
```
结算价格 = 参考价格 × 浮动比例
```

**示例**：
- 本期均价 = 5.00元
- 浮动比例 = 0.88（下浮12%）
- 结算价格 = 5.00 × 0.88 = 4.40元

## 数据库安装

### 1. 执行SQL脚本

按顺序执行以下脚本：

```bash
# 1. 创建数据库和用户
mysql -u root -p < sql/00_db_users.sql

# 2. 创建系统基础表（组织、用户等）
mysql -u root -p < sql/01_base_sys.sql

# 3. 创建商品相关表
mysql -u root -p < sql/10_goods_domain.sql

# 4. 创建市场价格管理表
mysql -u root -p < sql/11_market_price_system.sql
```

### 2. 验证安装

```sql
USE main;

-- 查看所有表
SHOW TABLES;

-- 验证市场价格相关表
DESC base_market;
DESC base_price_period;
DESC base_market_price;
DESC base_supplier_price;
```

## Excel数据导入

### 1. 准备Excel文件

Excel文件应包含以下sheet（每个sheet对应一个品类）：
- 蔬菜类
- 水产类
- 水果类

**表格格式**：

| 序号 | 品名 | 规格标准 | 单位 | 发改委指导价 | 富万家超市 | 育英巷菜市场 | 大润发 | 上月均价 | 本期均价 | 胡埗本期结算价(下浮12%) | 黄海本期结算价(下浮14%) |
|------|------|----------|------|--------------|-----------|--------------|--------|----------|----------|------------------------|------------------------|
| 1    | 四季豆 | 新鲜 | 斤 | 4.98 | 5 | 5.59 | 5.98 | 5.19 | 4.57 | 4.46 |
| 2    | 水果玉米棒 | 新鲜 | 斤 | 4.58 | 6 | 6 | 6.59 | 5.53 | 4.86 | 4.75 |

**注意事项**：
- 第一行是标题行（包含日期等信息）
- 第二行是列名
- 从第三行开始是数据行
- 价格字段可以为空

### 2. 安装Python依赖

```bash
pip install pandas openpyxl pymysql
```

### 3. 执行导入脚本

```bash
# 基本用法
python sql/import_market_prices.py \
  --file your_excel_file.xlsx \
  --period "2025年9月上旬" \
  --start-date "2025-09-01" \
  --end-date "2025-09-10"

# 指定数据库连接参数
python sql/import_market_prices.py \
  --host localhost \
  --port 3306 \
  --user food_user \
  --password StrongPassw0rd! \
  --database main \
  --file your_excel_file.xlsx \
  --period "2025年9月上旬" \
  --start-date "2025-09-01" \
  --end-date "2025-09-10"
```

**参数说明**：
- `--host`: 数据库主机地址（默认：localhost）
- `--port`: 数据库端口（默认：3306）
- `--user`: 数据库用户名（默认：food_user）
- `--password`: 数据库密码（默认：StrongPassw0rd!）
- `--database`: 数据库名（默认：main）
- `--file`: Excel文件路径（必填）
- `--period`: 价格时期名称（默认：2025年9月上旬）
- `--start-date`: 开始日期，格式YYYY-MM-DD（默认：2025-09-01）
- `--end-date`: 结束日期，格式YYYY-MM-DD（默认：2025-09-10）

### 4. 导入流程

脚本会自动：
1. 创建或获取组织（都匀市）
2. 创建或获取品类（蔬菜类、水产类、水果类）
3. 创建或获取商品（含规格、单位）
4. 创建或获取市场（发改委、富万家超市等）
5. 创建或获取价格时期
6. 创建或获取供应商（胡埗、黄海）
7. 导入各类价格数据

## 数据查询示例

### 1. 查询某个商品的所有价格信息

```sql
-- 查询"四季豆"在"2025年9月上旬"的所有价格
SELECT 
  g.name AS 商品名,
  s.name AS 规格,
  u.name AS 单位,
  m.name AS 市场名称,
  CASE mp.price_type
    WHEN 1 THEN '市场价'
    WHEN 2 THEN '指导价'
    WHEN 3 THEN '上月均价'
    WHEN 4 THEN '本期均价'
  END AS 价格类型,
  mp.price AS 价格,
  pp.name AS 时期
FROM base_market_price mp
JOIN base_goods g ON mp.goods_id = g.id
JOIN base_spec s ON g.spec_id = s.id
JOIN base_unit u ON g.unit_id = u.id
JOIN base_market m ON mp.market_id = m.id
JOIN base_price_period pp ON mp.period_id = pp.id
WHERE g.name = '四季豆'
  AND pp.name = '2025年9月上旬'
  AND mp.is_deleted = 0
ORDER BY mp.price_type, m.name;
```

### 2. 查询某个品类的价格对比

```sql
-- 查询蔬菜类商品的本期均价
SELECT 
  g.name AS 商品名,
  s.name AS 规格,
  u.name AS 单位,
  mp.price AS 本期均价,
  pp.name AS 时期
FROM base_market_price mp
JOIN base_goods g ON mp.goods_id = g.id
JOIN base_spec s ON g.spec_id = s.id
JOIN base_unit u ON g.unit_id = u.id
JOIN base_category c ON g.category_id = c.id
JOIN base_price_period pp ON mp.period_id = pp.id
WHERE c.name = '蔬菜类'
  AND mp.price_type = 4  -- 本期均价
  AND pp.name = '2025年9月上旬'
  AND mp.is_deleted = 0
ORDER BY g.name;
```

### 3. 查询供应商结算价

```sql
-- 查询胡埗供应商的结算价
SELECT 
  g.name AS 商品名,
  s.name AS 规格,
  u.name AS 单位,
  sup.name AS 供应商,
  sp.reference_price AS 参考价,
  sp.float_ratio AS 浮动比例,
  sp.settlement_price AS 结算价,
  pp.name AS 时期
FROM base_supplier_price sp
JOIN base_goods g ON sp.goods_id = g.id
JOIN base_spec s ON g.spec_id = s.id
JOIN base_unit u ON g.unit_id = u.id
JOIN supplier sup ON sp.supplier_id = sup.id
JOIN base_price_period pp ON sp.period_id = pp.id
WHERE sup.name = '胡埗'
  AND pp.name = '2025年9月上旬'
  AND sp.is_deleted = 0
ORDER BY g.name;
```

### 4. 使用综合价格视图

```sql
-- 查询所有商品的价格概览
SELECT 
  goods_name AS 商品名,
  spec_name AS 规格,
  unit_name AS 单位,
  category_name AS 品类,
  period_name AS 时期,
  guide_price AS 指导价,
  current_avg_price AS 本期均价,
  last_month_avg_price AS 上月均价
FROM v_comprehensive_price
WHERE period_name = '2025年9月上旬'
ORDER BY category_name, goods_name;
```

### 5. 价格变化趋势分析

```sql
-- 对比上月和本期的价格变化
SELECT 
  g.name AS 商品名,
  c.name AS 品类,
  last_month.price AS 上月价格,
  current.price AS 本期价格,
  ROUND(current.price - last_month.price, 2) AS 价格差,
  CONCAT(ROUND((current.price - last_month.price) / last_month.price * 100, 2), '%') AS 涨跌幅
FROM base_goods g
JOIN base_category c ON g.category_id = c.id
LEFT JOIN base_market_price last_month 
  ON g.id = last_month.goods_id 
  AND last_month.price_type = 3  -- 上月均价
LEFT JOIN base_market_price current 
  ON g.id = current.goods_id 
  AND current.price_type = 4  -- 本期均价
  AND current.period_id = last_month.period_id
WHERE last_month.is_deleted = 0 
  AND current.is_deleted = 0
  AND last_month.price IS NOT NULL
  AND current.price IS NOT NULL
ORDER BY 涨跌幅 DESC;
```

## 常见问题

### Q1: 导入时提示"外键约束失败"？
**A**: 确保已按顺序执行所有SQL脚本，特别是01_base_sys.sql（创建base_org等基础表）和10_goods_domain.sql。

### Q2: 如何处理重复数据？
**A**: 脚本使用`ON DUPLICATE KEY UPDATE`，相同的商品+市场+时期会自动更新价格，不会产生重复记录。

### Q3: 如何删除某个时期的所有数据？
**A**: 
```sql
-- 软删除（推荐）
UPDATE base_market_price SET is_deleted = 1 
WHERE period_id = '时期ID';

UPDATE base_supplier_price SET is_deleted = 1 
WHERE period_id = '时期ID';

-- 物理删除（慎用）
DELETE FROM base_market_price WHERE period_id = '时期ID';
DELETE FROM base_supplier_price WHERE period_id = '时期ID';
DELETE FROM base_price_period WHERE id = '时期ID';
```

### Q4: 如何修改供应商的浮动比例？
**A**:
```sql
-- 修改supplier表中的float_ratio
UPDATE supplier 
SET float_ratio = 0.90  -- 下浮10%
WHERE name = '供应商名称';

-- 已有的结算价记录需要单独更新
UPDATE base_supplier_price 
SET float_ratio = 0.90
WHERE supplier_id = '供应商ID';
```

### Q5: Excel表格格式不匹配怎么办？
**A**: 可以修改导入脚本中的`column_mapping`字典，映射实际的列名。

## 扩展功能

### 1. 添加新的市场类型

```sql
-- 修改base_market表，添加新市场
INSERT INTO base_market (id, name, code, market_type, org_id)
VALUES (UUID(), '新市场名称', 'MKT_NEW', 2, '组织ID');
```

### 2. 添加新的价格类型

如需要添加新的价格类型（如促销价、批发价等），可以扩展`price_type`字段的取值范围，并修改相关代码注释。

### 3. 数据导出

```bash
# 导出某个时期的所有价格数据
mysql -u food_user -p main -e "
SELECT * FROM v_comprehensive_price 
WHERE period_name = '2025年9月上旬'
" > prices_export.txt

# 导出为CSV格式
mysql -u food_user -p main -e "
SELECT * FROM v_comprehensive_price 
WHERE period_name = '2025年9月上旬'
" | sed 's/\t/,/g' > prices_export.csv
```

## 性能优化建议

1. **定期清理软删除数据**：
```sql
-- 物理删除超过1年的软删除记录
DELETE FROM base_market_price 
WHERE is_deleted = 1 
  AND updated_at < DATE_SUB(NOW(), INTERVAL 1 YEAR);
```

2. **添加适当的索引**：
已在表结构中添加常用索引，如需优化特定查询，可以根据实际情况添加组合索引。

3. **使用分区表**（大数据量时）：
```sql
-- 按时期分区
ALTER TABLE base_market_price
PARTITION BY RANGE (YEAR(created_at) * 100 + MONTH(created_at)) (
    PARTITION p202501 VALUES LESS THAN (202502),
    PARTITION p202502 VALUES LESS THAN (202503),
    ...
);
```

## 技术支持

如有问题，请查看：
1. 数据库日志：`/var/log/mysql/error.log`
2. 脚本执行日志：导入脚本会输出详细的执行信息
3. 联系开发团队获取支持

---

**版本**: 1.0  
**更新日期**: 2025-10-21  
**维护者**: 开发团队
