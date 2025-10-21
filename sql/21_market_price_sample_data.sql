/* ======== 市场价格示例数据 ======== */
/* 说明：此文件展示如何将Excel数据导入到market_price相关表中 */

USE main;

/* ---------- 示例：插入市场数据 ---------- */
-- 插入市场基础数据（对应Excel中的各个市场/超市）
INSERT INTO base_market (id, name, org_id, code, sort) VALUES
  (UUID(), '发改委', (SELECT id FROM base_org LIMIT 1), 'NDRC', 1),
  (UUID(), '富万家超市', (SELECT id FROM base_org LIMIT 1), 'FUWANJIA', 2),
  (UUID(), '育英巷菜市场', (SELECT id FROM base_org LIMIT 1), 'YUYING', 3),
  (UUID(), '大湘发', (SELECT id FROM base_org LIMIT 1), 'DAXIANGFA', 4),
  (UUID(), '胡埔', (SELECT id FROM base_org LIMIT 1), 'HUPU', 5),
  (UUID(), '贵海', (SELECT id FROM base_org LIMIT 1), 'GUIHAI', 6)
ON DUPLICATE KEY UPDATE name=VALUES(name);


/* ---------- 示例：插入品类数据 ---------- */
-- 插入品类（对应Excel的每个sheet）
INSERT INTO base_category (id, name, org_id, code, sort) VALUES
  (UUID(), '蔬菜类', (SELECT id FROM base_org LIMIT 1), 'VEGETABLES', 1),
  (UUID(), '水产类', (SELECT id FROM base_org LIMIT 1), 'AQUATIC', 2)
ON DUPLICATE KEY UPDATE name=VALUES(name);


/* ---------- 示例：插入期次数据 ---------- */
-- 插入2025年9月上旬的蔬菜类价格期次
INSERT INTO market_price_period (
  id, title, period_year, period_month, period_type, 
  category_id, org_id, publish_date
) VALUES (
  UUID(),
  '2025年9月上旬都匀市主要蔬菜类市场参考价',
  2025,
  9,
  '上旬',
  (SELECT id FROM base_category WHERE code = 'VEGETABLES' LIMIT 1),
  (SELECT id FROM base_org LIMIT 1),
  '2025-09-10'
);

-- 插入2025年9月上旬的水产类价格期次
INSERT INTO market_price_period (
  id, title, period_year, period_month, period_type, 
  category_id, org_id, publish_date
) VALUES (
  UUID(),
  '2025年9月上旬都匀市主要水产类市场参考价',
  2025,
  9,
  '上旬',
  (SELECT id FROM base_category WHERE code = 'AQUATIC' LIMIT 1),
  (SELECT id FROM base_org LIMIT 1),
  '2025-09-10'
);


/* ---------- 示例：插入蔬菜类价格记录 ---------- */
-- 根据Excel第一个sheet（蔬菜类）的数据示例
-- 注意：实际使用时需要先确保对应的商品已经在base_goods表中存在

-- 示例1：四季豆（序号4）
INSERT INTO market_price_record (
  id, period_id, goods_id, goods_name, spec_name, unit_name, seq_num,
  ndrc_guide_price, fuwanjia_price, yuying_market_price, daxiangfa_price,
  last_month_avg_price, current_period_avg_price,
  hupu_settlement_price, guihai_settlement_price
) VALUES (
  UUID(),
  (SELECT id FROM market_price_period WHERE title LIKE '%蔬菜类%' AND period_year = 2025 AND period_month = 9 LIMIT 1),
  (SELECT id FROM base_goods WHERE name = '四季豆' LIMIT 1),
  '四季豆',
  '新鲜',
  '斤',
  4,
  NULL,      -- 发改委指导价（Excel中为空）
  4.98,      -- 富万家超市
  5.00,      -- 育英巷菜市场
  5.59,      -- 大湘发
  5.98,      -- 上月均价
  5.19,      -- 本期均价
  4.57,      -- 胡埔结算价（下浮12%）
  4.46       -- 贵海结算价（下浮14%）
);

-- 示例2：水果玉米棒（序号5）
INSERT INTO market_price_record (
  id, period_id, goods_id, goods_name, spec_name, unit_name, seq_num,
  ndrc_guide_price, fuwanjia_price, yuying_market_price, daxiangfa_price,
  last_month_avg_price, current_period_avg_price,
  hupu_settlement_price, guihai_settlement_price
) VALUES (
  UUID(),
  (SELECT id FROM market_price_period WHERE title LIKE '%蔬菜类%' AND period_year = 2025 AND period_month = 9 LIMIT 1),
  (SELECT id FROM base_goods WHERE name = '水果玉米棒' LIMIT 1),
  '水果玉米棒',
  '新鲜',
  '斤',
  5,
  NULL,      -- 发改委指导价（Excel中为空）
  4.58,      -- 富万家超市
  6.00,      -- 育英巷菜市场
  6.00,      -- 大湘发
  6.59,      -- 上月均价
  5.53,      -- 本期均价
  4.86,      -- 胡埔结算价（下浮12%）
  4.75       -- 贵海结算价（下浮14%）
);


/* ---------- 示例：插入水产类价格记录 ---------- */
-- 根据Excel第二个sheet（水产类）的数据示例

-- 示例1：罗非鱼（序号9）
INSERT INTO market_price_record (
  id, period_id, goods_id, goods_name, spec_name, unit_name, seq_num,
  ndrc_guide_price, fuwanjia_price, yuying_market_price, daxiangfa_price,
  last_month_avg_price, current_period_avg_price,
  hupu_settlement_price, guihai_settlement_price
) VALUES (
  UUID(),
  (SELECT id FROM market_price_period WHERE title LIKE '%水产类%' AND period_year = 2025 AND period_month = 9 LIMIT 1),
  (SELECT id FROM base_goods WHERE name = '罗非鱼' LIMIT 1),
  '罗非鱼',
  '新鲜不杀',
  '斤',
  9,
  NULL,      -- 发改委指导价（Excel中为空）
  15.90,     -- 富万家超市
  15.00,     -- 育英巷菜市场
  16.80,     -- 大湘发
  16.57,     -- 上月均价
  15.90,     -- 本期均价
  13.99,     -- 胡埔结算价（下浮12%）
  13.67      -- 贵海结算价（下浮14%）
);

-- 示例2：甲鱼（序号10）
INSERT INTO market_price_record (
  id, period_id, goods_id, goods_name, spec_name, unit_name, seq_num,
  ndrc_guide_price, fuwanjia_price, yuying_market_price, daxiangfa_price,
  last_month_avg_price, current_period_avg_price,
  hupu_settlement_price, guihai_settlement_price
) VALUES (
  UUID(),
  (SELECT id FROM market_price_period WHERE title LIKE '%水产类%' AND period_year = 2025 AND period_month = 9 LIMIT 1),
  (SELECT id FROM base_goods WHERE name = '甲鱼' LIMIT 1),
  '甲鱼',
  '新鲜不杀',
  '斤',
  10,
  NULL,      -- 发改委指导价（Excel中为空）
  49.90,     -- 富万家超市
  45.00,     -- 育英巷菜市场
  55.50,     -- 大湘发
  51.80,     -- 上月均价
  50.13,     -- 本期均价
  44.12,     -- 胡埔结算价（下浮12%）
  43.11      -- 贵海结算价（下浮14%）
);


/* ---------- 使用market_price_detail表的替代方案（可选） ---------- */
-- 如果使用更灵活的market_price_detail表，可以这样插入数据：

-- 示例：四季豆在各个市场的价格
/*
INSERT INTO market_price_detail (id, period_id, goods_id, market_id, price, price_type) VALUES
  (UUID(), 
   (SELECT id FROM market_price_period WHERE title LIKE '%蔬菜类%' AND period_year = 2025 AND period_month = 9 LIMIT 1),
   (SELECT id FROM base_goods WHERE name = '四季豆' LIMIT 1),
   (SELECT id FROM base_market WHERE code = 'FUWANJIA' LIMIT 1),
   4.98,
   '市场价'),
  (UUID(), 
   (SELECT id FROM market_price_period WHERE title LIKE '%蔬菜类%' AND period_year = 2025 AND period_month = 9 LIMIT 1),
   (SELECT id FROM base_goods WHERE name = '四季豆' LIMIT 1),
   (SELECT id FROM base_market WHERE code = 'YUYING' LIMIT 1),
   5.00,
   '市场价'),
  (UUID(), 
   (SELECT id FROM market_price_period WHERE title LIKE '%蔬菜类%' AND period_year = 2025 AND period_month = 9 LIMIT 1),
   (SELECT id FROM base_goods WHERE name = '四季豆' LIMIT 1),
   (SELECT id FROM base_market WHERE code = 'HUPU' LIMIT 1),
   4.57,
   '结算价');
*/

/* ---------- 查询示例 ---------- */
-- 查询某个期次的所有价格记录
/*
SELECT 
  mpr.seq_num AS 序号,
  mpr.goods_name AS 品名,
  mpr.spec_name AS 规格标准,
  mpr.unit_name AS 单位,
  mpr.ndrc_guide_price AS 发改委指导价,
  mpr.fuwanjia_price AS 富万家超市,
  mpr.yuying_market_price AS 育英巷菜市场,
  mpr.daxiangfa_price AS 大湘发,
  mpr.last_month_avg_price AS 上月均价,
  mpr.current_period_avg_price AS 本期均价,
  mpr.hupu_settlement_price AS '胡埔本期结算价(下浮12%)',
  mpr.guihai_settlement_price AS '贵海本期结算价(下浮14%)',
  mpp.title AS 期次标题
FROM market_price_record mpr
JOIN market_price_period mpp ON mpr.period_id = mpp.id
WHERE mpp.period_year = 2025 
  AND mpp.period_month = 9 
  AND mpp.period_type = '上旬'
ORDER BY mpr.seq_num;
*/
