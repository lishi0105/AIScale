/* ======== 价格询价数据插入示例 ======== */
/* 说明：
   本文件提供价格询价数据的插入示例
   展示如何记录一次完整的询价信息
*/

USE main;

/* 
假设已有以下基础数据：
- 品类：蔬菜类 (category_id = 'vegetable-category-id')
- 品类：水产类 (category_id = 'seafood-category-id')
- 商品：可豆、无筋豆、棒豆等 (base_goods)
- 市场：富万家超市、育英巷菜市场、大润发 (base_market)
- 供应商：胡坤、贵海 (supplier)
- 组织：某中队 (org_id = 'org-id-001')
*/

-- ==========================================
-- 示例1：插入一次蔬菜类询价记录
-- ==========================================

-- 1. 插入询价主表记录
INSERT INTO price_inquiry (
  id,
  title,
  year,
  month,
  period,
  category_id,
  org_id,
  inquiry_date,
  remark
) VALUES (
  UUID(), -- 或使用具体的UUID，如：'inquiry-2025-09-vegetable-001'
  '2025年9月上旬都匀市主要蔬菜类市场参考价',
  2025,
  9,
  '上旬',
  'vegetable-category-id', -- 替换为实际的品类ID
  'org-id-001',            -- 替换为实际的组织ID
  '2025-09-10',
  NULL
);

-- 2. 插入询价商品明细（以可豆为例）
INSERT INTO price_inquiry_detail (
  id,
  inquiry_id,
  goods_id,
  sequence,
  guide_price,
  last_month_avg_price,
  current_avg_price,
  remark
) VALUES (
  UUID(),
  'inquiry-2025-09-vegetable-001', -- 对应上面的询价主表ID
  'goods-kedou-001',               -- 可豆的商品ID
  1,                               -- 序号
  3.98,                            -- 发改委指导价
  5.79,                            -- 上月均价
  4.65,                            -- 本期均价
  NULL
);

-- 3. 插入市场价格明细（可豆在富万家超市的价格）
INSERT INTO price_inquiry_market_detail (
  id,
  inquiry_detail_id,
  market_id,
  price,
  remark
) VALUES (
  UUID(),
  'detail-kedou-001',    -- 对应上面的商品明细ID
  'market-fuwanjia-001', -- 富万家超市ID
  4.0,                   -- 市场价格
  NULL
);

-- 4. 插入市场价格明细（可豆在育英巷菜市场的价格）
INSERT INTO price_inquiry_market_detail (
  id,
  inquiry_detail_id,
  market_id,
  price,
  remark
) VALUES (
  UUID(),
  'detail-kedou-001',           -- 对应上面的商品明细ID
  'market-yuyingxiang-001',     -- 育英巷菜市场ID
  4.5,                          -- 市场价格
  NULL
);

-- 5. 插入市场价格明细（可豆在大润发的价格）
INSERT INTO price_inquiry_market_detail (
  id,
  inquiry_detail_id,
  market_id,
  price,
  remark
) VALUES (
  UUID(),
  'detail-kedou-001',    -- 对应上面的商品明细ID
  'market-darunfa-001',  -- 大润发ID
  5.98,                  -- 市场价格
  NULL
);

-- 6. 插入供应商结算价明细（胡坤供应商）
INSERT INTO price_inquiry_supplier_detail (
  id,
  inquiry_detail_id,
  supplier_id,
  settlement_price,
  float_ratio,
  remark
) VALUES (
  UUID(),
  'detail-kedou-001',    -- 对应上面的商品明细ID
  'supplier-hukun-001',  -- 胡坤供应商ID
  4.09,                  -- 结算价
  0.88,                  -- 下浮比例12%（即1-0.12=0.88）
  '胡坤本期结算价（下浮12%）'
);

-- 7. 插入供应商结算价明细（贵海供应商）
INSERT INTO price_inquiry_supplier_detail (
  id,
  inquiry_detail_id,
  supplier_id,
  settlement_price,
  float_ratio,
  remark
) VALUES (
  UUID(),
  'detail-kedou-001',     -- 对应上面的商品明细ID
  'supplier-guihai-001',  -- 贵海供应商ID
  4.00,                   -- 结算价
  0.86,                   -- 下浮比例14%（即1-0.14=0.86）
  '贵海本期结算价（下浮14%）'
);


-- ==========================================
-- 示例2：插入一次水产类询价记录
-- ==========================================

-- 1. 插入询价主表记录
INSERT INTO price_inquiry (
  id,
  title,
  year,
  month,
  period,
  category_id,
  org_id,
  inquiry_date,
  remark
) VALUES (
  'inquiry-2025-09-seafood-001',
  '2025年9月上旬都匀市主要水产类市场参考价',
  2025,
  9,
  '上旬',
  'seafood-category-id', -- 替换为实际的水产类品类ID
  'org-id-001',          -- 替换为实际的组织ID
  '2025-09-10',
  NULL
);

-- 2. 插入询价商品明细（以罗非鱼为例）
INSERT INTO price_inquiry_detail (
  id,
  inquiry_id,
  goods_id,
  sequence,
  guide_price,
  last_month_avg_price,
  current_avg_price,
  remark
) VALUES (
  'detail-luofeiyu-001',
  'inquiry-2025-09-seafood-001', -- 对应上面的询价主表ID
  'goods-luofeiyu-001',          -- 罗非鱼的商品ID
  9,                             -- 序号
  15.9,                          -- 发改委指导价
  16.57,                         -- 上月均价
  15.90,                         -- 本期均价
  NULL
);

-- 3. 插入市场价格明细（罗非鱼在富万家超市的价格）
INSERT INTO price_inquiry_market_detail (
  id,
  inquiry_detail_id,
  market_id,
  price,
  remark
) VALUES (
  UUID(),
  'detail-luofeiyu-001', -- 对应上面的商品明细ID
  'market-fuwanjia-001', -- 富万家超市ID
  15.0,                  -- 市场价格
  NULL
);

-- 4. 插入市场价格明细（罗非鱼在大润发的价格）
INSERT INTO price_inquiry_market_detail (
  id,
  inquiry_detail_id,
  market_id,
  price,
  remark
) VALUES (
  UUID(),
  'detail-luofeiyu-001', -- 对应上面的商品明细ID
  'market-darunfa-001',  -- 大润发ID
  16.8,                  -- 市场价格
  NULL
);

-- 5. 插入供应商结算价明细（胡坤供应商）
INSERT INTO price_inquiry_supplier_detail (
  id,
  inquiry_detail_id,
  supplier_id,
  settlement_price,
  float_ratio,
  remark
) VALUES (
  UUID(),
  'detail-luofeiyu-001', -- 对应上面的商品明细ID
  'supplier-hukun-001',  -- 胡坤供应商ID
  13.99,                 -- 结算价
  0.88,                  -- 下浮比例12%
  '胡坤本期结算价（下浮12%）'
);

-- 6. 插入供应商结算价明细（贵海供应商）
INSERT INTO price_inquiry_supplier_detail (
  id,
  inquiry_detail_id,
  supplier_id,
  settlement_price,
  float_ratio,
  remark
) VALUES (
  UUID(),
  'detail-luofeiyu-001',  -- 对应上面的商品明细ID
  'supplier-guihai-001',  -- 贵海供应商ID
  13.67,                  -- 结算价
  0.86,                   -- 下浮比例14%
  '贵海本期结算价（下浮14%）'
);


-- ==========================================
-- 常用查询示例
-- ==========================================

-- 查询1：按年月旬次查询询价记录
SELECT 
  id,
  title,
  year,
  month,
  period,
  inquiry_date,
  created_at
FROM price_inquiry 
WHERE year = 2025 
  AND month = 9 
  AND period = '上旬'
  AND is_deleted = 0
ORDER BY inquiry_date DESC;


-- 查询2：按品类查询询价记录
SELECT 
  pi.id,
  pi.title,
  bc.name AS category_name,
  pi.inquiry_date
FROM price_inquiry pi
INNER JOIN base_category bc ON pi.category_id = bc.id
WHERE pi.category_id = 'vegetable-category-id'
  AND pi.is_deleted = 0
ORDER BY pi.inquiry_date DESC;


-- 查询3：模糊搜索询价标题（支持搜索"蔬菜类"或"水产类"）
SELECT 
  id,
  title,
  year,
  month,
  period,
  inquiry_date
FROM price_inquiry 
WHERE (title LIKE '%蔬菜类%' OR title LIKE '%水产类%')
  AND is_deleted = 0
ORDER BY inquiry_date DESC;


-- 查询4：查询某次询价的完整信息（包括市场价格和供应商结算价）
SELECT 
  pid.sequence AS 序号,
  bg.name AS 品名,
  bs.name AS 规格标准,
  bu.name AS 单位,
  pid.guide_price AS 发改委指导价,
  GROUP_CONCAT(DISTINCT CONCAT(bm.name, ':', pimd.price) ORDER BY bm.name SEPARATOR ', ') AS 市场价格,
  pid.last_month_avg_price AS 上月均价,
  pid.current_avg_price AS 本期均价,
  GROUP_CONCAT(DISTINCT CONCAT(s.name, '本期结算价（下浮', ROUND((1-pisd.float_ratio)*100, 0), '%）:', pisd.settlement_price) ORDER BY s.name SEPARATOR ', ') AS 供应商结算价
FROM price_inquiry pi
INNER JOIN price_inquiry_detail pid ON pi.id = pid.inquiry_id
INNER JOIN base_goods bg ON pid.goods_id = bg.id
INNER JOIN base_spec bs ON bg.spec_id = bs.id
INNER JOIN base_unit bu ON bg.unit_id = bu.id
LEFT JOIN price_inquiry_market_detail pimd ON pid.id = pimd.inquiry_detail_id
LEFT JOIN base_market bm ON pimd.market_id = bm.id
LEFT JOIN price_inquiry_supplier_detail pisd ON pid.id = pisd.inquiry_detail_id
LEFT JOIN supplier s ON pisd.supplier_id = s.id
WHERE pi.id = 'inquiry-2025-09-vegetable-001'
  AND pi.is_deleted = 0
  AND pid.is_deleted = 0
GROUP BY pid.id, pid.sequence, bg.name, bs.name, bu.name, pid.guide_price, pid.last_month_avg_price, pid.current_avg_price
ORDER BY pid.sequence ASC;


-- 查询5：查询某个商品在某个时间段内的价格变化趋势
SELECT 
  pi.year AS 年份,
  pi.month AS 月份,
  pi.period AS 旬次,
  pi.inquiry_date AS 询价日期,
  bg.name AS 商品名称,
  pid.current_avg_price AS 本期均价,
  bm.name AS 市场名称,
  pimd.price AS 市场价格
FROM price_inquiry pi
INNER JOIN price_inquiry_detail pid ON pi.id = pid.inquiry_id
INNER JOIN base_goods bg ON pid.goods_id = bg.id
LEFT JOIN price_inquiry_market_detail pimd ON pid.id = pimd.inquiry_detail_id
LEFT JOIN base_market bm ON pimd.market_id = bm.id
WHERE bg.id = 'goods-kedou-001' 
  AND pi.inquiry_date BETWEEN '2025-01-01' AND '2025-12-31'
  AND pi.is_deleted = 0
  AND pid.is_deleted = 0
ORDER BY pi.inquiry_date ASC, bm.name ASC;


-- 查询6：查询某次询价中所有商品的市场均价和供应商结算价对比
SELECT 
  pid.sequence AS 序号,
  bg.name AS 商品名称,
  pid.current_avg_price AS 市场均价,
  s.name AS 供应商名称,
  pisd.settlement_price AS 结算价,
  pisd.float_ratio AS 下浮比例,
  ROUND((1 - pisd.float_ratio) * 100, 2) AS 下浮百分比,
  ROUND((pid.current_avg_price - pisd.settlement_price), 2) AS 价格差,
  ROUND((pid.current_avg_price - pisd.settlement_price) / pid.current_avg_price * 100, 2) AS 折扣百分比
FROM price_inquiry pi
INNER JOIN price_inquiry_detail pid ON pi.id = pid.inquiry_id
INNER JOIN base_goods bg ON pid.goods_id = bg.id
LEFT JOIN price_inquiry_supplier_detail pisd ON pid.id = pisd.inquiry_detail_id
LEFT JOIN supplier s ON pisd.supplier_id = s.id
WHERE pi.id = 'inquiry-2025-09-vegetable-001'
  AND pi.is_deleted = 0
  AND pid.is_deleted = 0
ORDER BY pid.sequence ASC, s.name ASC;


-- 查询7：统计某个品类在某个时间段内的平均价格走势
SELECT 
  pi.year AS 年份,
  pi.month AS 月份,
  pi.period AS 旬次,
  bc.name AS 品类名称,
  COUNT(DISTINCT pid.goods_id) AS 询价商品数量,
  ROUND(AVG(pid.current_avg_price), 2) AS 平均价格,
  ROUND(MIN(pid.current_avg_price), 2) AS 最低价格,
  ROUND(MAX(pid.current_avg_price), 2) AS 最高价格
FROM price_inquiry pi
INNER JOIN base_category bc ON pi.category_id = bc.id
INNER JOIN price_inquiry_detail pid ON pi.id = pid.inquiry_id
WHERE pi.category_id = 'vegetable-category-id'
  AND pi.inquiry_date BETWEEN '2025-01-01' AND '2025-12-31'
  AND pi.is_deleted = 0
  AND pid.is_deleted = 0
GROUP BY pi.year, pi.month, pi.period, bc.name
ORDER BY pi.year ASC, pi.month ASC, 
  CASE pi.period 
    WHEN '上旬' THEN 1 
    WHEN '中旬' THEN 2 
    WHEN '下旬' THEN 3 
  END ASC;


-- 查询8：查询某个市场在某次询价中所有商品的价格
SELECT 
  pid.sequence AS 序号,
  bg.name AS 商品名称,
  bs.name AS 规格,
  bu.name AS 单位,
  bm.name AS 市场名称,
  pimd.price AS 市场价格,
  pid.current_avg_price AS 本期均价,
  ROUND((pimd.price - pid.current_avg_price), 2) AS 与均价差,
  ROUND((pimd.price - pid.current_avg_price) / pid.current_avg_price * 100, 2) AS 差异百分比
FROM price_inquiry pi
INNER JOIN price_inquiry_detail pid ON pi.id = pid.inquiry_id
INNER JOIN base_goods bg ON pid.goods_id = bg.id
INNER JOIN base_spec bs ON bg.spec_id = bs.id
INNER JOIN base_unit bu ON bg.unit_id = bu.id
INNER JOIN price_inquiry_market_detail pimd ON pid.id = pimd.inquiry_detail_id
INNER JOIN base_market bm ON pimd.market_id = bm.id
WHERE pi.id = 'inquiry-2025-09-vegetable-001'
  AND bm.id = 'market-fuwanjia-001'
  AND pi.is_deleted = 0
  AND pid.is_deleted = 0
ORDER BY pid.sequence ASC;
