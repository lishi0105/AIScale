/* ========================================================================
   示例数据
   用于测试市场价格管理系统
   ======================================================================== */

USE main;

-- 注意：执行前请确保已运行 00_db_users.sql, 01_base_sys.sql, 10_goods_domain.sql, 11_market_price_system.sql

-- ========== 1. 创建测试组织 ==========
SET @org_id = UUID();
INSERT INTO base_org (id, name, code, sort, parent_id, description, is_deleted)
VALUES (@org_id, '都匀市', 'DUYUN', 0, @org_id, '都匀市测试组织', 0);

-- ========== 2. 创建单位 ==========
SET @unit_jin = UUID();
INSERT INTO base_unit (id, name, code, sort, is_deleted)
VALUES (@unit_jin, '斤', 'JIN', 1, 0);

SET @unit_kg = UUID();
INSERT INTO base_unit (id, name, code, sort, is_deleted)
VALUES (@unit_kg, '公斤', 'KG', 2, 0);

-- ========== 3. 创建规格 ==========
SET @spec_fresh = UUID();
INSERT INTO base_spec (id, name, code, sort, is_deleted)
VALUES (@spec_fresh, '新鲜', 'FRESH', 1, 0);

SET @spec_fresh_alive = UUID();
INSERT INTO base_spec (id, name, code, sort, is_deleted)
VALUES (@spec_fresh_alive, '新鲜不杀', 'FRESH_ALIVE', 2, 0);

-- ========== 4. 创建品类 ==========
SET @category_veg = UUID();
INSERT INTO base_category (id, name, org_id, code, sort, is_deleted)
VALUES (@category_veg, '蔬菜类', @org_id, 'CAT_VEG', 1, 0);

SET @category_fish = UUID();
INSERT INTO base_category (id, name, org_id, code, sort, is_deleted)
VALUES (@category_fish, '水产类', @org_id, 'CAT_FISH', 2, 0);

SET @category_fruit = UUID();
INSERT INTO base_category (id, name, org_id, code, sort, is_deleted)
VALUES (@category_fruit, '水果类', @org_id, 'CAT_FRUIT', 3, 0);

-- ========== 5. 创建商品 ==========

-- 蔬菜类商品
SET @goods_sijidou = UUID();
INSERT INTO base_goods (id, name, code, sort, spec_id, unit_id, category_id, org_id, is_deleted)
VALUES (@goods_sijidou, '四季豆', 'SKU_VEG_001', 1, @spec_fresh, @unit_jin, @category_veg, @org_id, 0);

SET @goods_corn = UUID();
INSERT INTO base_goods (id, name, code, sort, spec_id, unit_id, category_id, org_id, is_deleted)
VALUES (@goods_corn, '水果玉米棒', 'SKU_VEG_002', 2, @spec_fresh, @unit_jin, @category_veg, @org_id, 0);

SET @goods_cabbage = UUID();
INSERT INTO base_goods (id, name, code, sort, spec_id, unit_id, category_id, org_id, is_deleted)
VALUES (@goods_cabbage, '大白菜', 'SKU_VEG_003', 3, @spec_fresh, @unit_jin, @category_veg, @org_id, 0);

-- 水产类商品
SET @goods_tilapia = UUID();
INSERT INTO base_goods (id, name, code, sort, spec_id, unit_id, category_id, org_id, is_deleted)
VALUES (@goods_tilapia, '罗非鱼', 'SKU_FISH_001', 1, @spec_fresh_alive, @unit_jin, @category_fish, @org_id, 0);

SET @goods_carp = UUID();
INSERT INTO base_goods (id, name, code, sort, spec_id, unit_id, category_id, org_id, is_deleted)
VALUES (@goods_carp, '甲鱼', 'SKU_FISH_002', 2, @spec_fresh_alive, @unit_jin, @category_fish, @org_id, 0);

-- 水果类商品
SET @goods_melon = UUID();
INSERT INTO base_goods (id, name, code, sort, spec_id, unit_id, category_id, org_id, is_deleted)
VALUES (@goods_melon, '哈密瓜', 'SKU_FRUIT_001', 1, @spec_fresh, @unit_jin, @category_fruit, @org_id, 0);

SET @goods_banana = UUID();
INSERT INTO base_goods (id, name, code, sort, spec_id, unit_id, category_id, org_id, is_deleted)
VALUES (@goods_banana, '香蕉', 'SKU_FRUIT_002', 2, @spec_fresh, @unit_jin, @category_fruit, @org_id, 0);

-- ========== 6. 创建市场 ==========

-- 政府指导
SET @market_fgw = UUID();
INSERT INTO base_market (id, name, code, market_type, org_id, is_deleted)
VALUES (@market_fgw, '发改委', 'MKT_FGW', 1, @org_id, 0);

-- 超市
SET @market_fwj = UUID();
INSERT INTO base_market (id, name, code, market_type, org_id, is_deleted)
VALUES (@market_fwj, '富万家超市', 'MKT_FWJ', 2, @org_id, 0);

SET @market_drf = UUID();
INSERT INTO base_market (id, name, code, market_type, org_id, is_deleted)
VALUES (@market_drf, '大润发', 'MKT_DRF', 2, @org_id, 0);

-- 菜市场
SET @market_yyx = UUID();
INSERT INTO base_market (id, name, code, market_type, org_id, is_deleted)
VALUES (@market_yyx, '育英巷菜市场', 'MKT_YYX', 3, @org_id, 0);

-- 均价（虚拟市场）
SET @market_avg_last = UUID();
INSERT INTO base_market (id, name, code, market_type, org_id, is_deleted)
VALUES (@market_avg_last, '上月均价', 'MKT_AVG_LAST', 5, @org_id, 0);

SET @market_avg_cur = UUID();
INSERT INTO base_market (id, name, code, market_type, org_id, is_deleted)
VALUES (@market_avg_cur, '本期均价', 'MKT_AVG_CUR', 5, @org_id, 0);

-- ========== 7. 创建价格时期 ==========

SET @period_202509_early = UUID();
INSERT INTO base_price_period (id, name, code, start_date, end_date, period_type, org_id, is_deleted)
VALUES (@period_202509_early, '2025年9月上旬', '2025-09-01', '2025-09-01', '2025-09-10', 1, @org_id, 0);

SET @period_202508 = UUID();
INSERT INTO base_price_period (id, name, code, start_date, end_date, period_type, org_id, is_deleted)
VALUES (@period_202508, '2025年8月', '2025-08-01', '2025-08-01', '2025-08-31', 4, @org_id, 0);

-- ========== 8. 创建供应商 ==========

SET @supplier_hupu = UUID();
INSERT INTO supplier (id, name, code, sort, status, description, float_ratio, org_id, is_deleted)
VALUES (@supplier_hupu, '胡埗', 'SUP_HUPU', 1, 1, '胡埗供应商（下浮12%）', 0.88, @org_id, 0);

SET @supplier_huanghai = UUID();
INSERT INTO supplier (id, name, code, sort, status, description, float_ratio, org_id, is_deleted)
VALUES (@supplier_huanghai, '黄海', 'SUP_HUANGHAI', 2, 1, '黄海供应商（下浮14%）', 0.86, @org_id, 0);

-- ========== 9. 插入市场价格数据 ==========

-- 四季豆的价格数据（2025年9月上旬）
-- 指导价（发改委）
INSERT INTO base_market_price (id, goods_id, market_id, period_id, price, price_type, org_id, is_deleted)
VALUES (UUID(), @goods_sijidou, @market_fgw, @period_202509_early, 4.98, 2, @org_id, 0);

-- 市场价（富万家超市）
INSERT INTO base_market_price (id, goods_id, market_id, period_id, price, price_type, org_id, is_deleted)
VALUES (UUID(), @goods_sijidou, @market_fwj, @period_202509_early, 5.00, 1, @org_id, 0);

-- 市场价（育英巷菜市场）
INSERT INTO base_market_price (id, goods_id, market_id, period_id, price, price_type, org_id, is_deleted)
VALUES (UUID(), @goods_sijidou, @market_yyx, @period_202509_early, 5.59, 1, @org_id, 0);

-- 市场价（大润发）
INSERT INTO base_market_price (id, goods_id, market_id, period_id, price, price_type, org_id, is_deleted)
VALUES (UUID(), @goods_sijidou, @market_drf, @period_202509_early, 5.98, 1, @org_id, 0);

-- 上月均价
INSERT INTO base_market_price (id, goods_id, market_id, period_id, price, price_type, org_id, is_deleted)
VALUES (UUID(), @goods_sijidou, @market_avg_last, @period_202509_early, 5.19, 3, @org_id, 0);

-- 本期均价
INSERT INTO base_market_price (id, goods_id, market_id, period_id, price, price_type, org_id, is_deleted)
VALUES (UUID(), @goods_sijidou, @market_avg_cur, @period_202509_early, 4.57, 4, @org_id, 0);

-- 水果玉米棒的价格数据（2025年9月上旬）
INSERT INTO base_market_price (id, goods_id, market_id, period_id, price, price_type, org_id, is_deleted)
VALUES (UUID(), @goods_corn, @market_fgw, @period_202509_early, 4.58, 2, @org_id, 0);

INSERT INTO base_market_price (id, goods_id, market_id, period_id, price, price_type, org_id, is_deleted)
VALUES (UUID(), @goods_corn, @market_fwj, @period_202509_early, 6.00, 1, @org_id, 0);

INSERT INTO base_market_price (id, goods_id, market_id, period_id, price, price_type, org_id, is_deleted)
VALUES (UUID(), @goods_corn, @market_yyx, @period_202509_early, 6.00, 1, @org_id, 0);

INSERT INTO base_market_price (id, goods_id, market_id, period_id, price, price_type, org_id, is_deleted)
VALUES (UUID(), @goods_corn, @market_drf, @period_202509_early, 6.59, 1, @org_id, 0);

INSERT INTO base_market_price (id, goods_id, market_id, period_id, price, price_type, org_id, is_deleted)
VALUES (UUID(), @goods_corn, @market_avg_last, @period_202509_early, 5.53, 3, @org_id, 0);

INSERT INTO base_market_price (id, goods_id, market_id, period_id, price, price_type, org_id, is_deleted)
VALUES (UUID(), @goods_corn, @market_avg_cur, @period_202509_early, 4.86, 4, @org_id, 0);

-- 罗非鱼的价格数据（2025年9月上旬）
INSERT INTO base_market_price (id, goods_id, market_id, period_id, price, price_type, org_id, is_deleted)
VALUES (UUID(), @goods_tilapia, @market_fgw, @period_202509_early, 15.90, 2, @org_id, 0);

INSERT INTO base_market_price (id, goods_id, market_id, period_id, price, price_type, org_id, is_deleted)
VALUES (UUID(), @goods_tilapia, @market_fwj, @period_202509_early, 15.00, 1, @org_id, 0);

INSERT INTO base_market_price (id, goods_id, market_id, period_id, price, price_type, org_id, is_deleted)
VALUES (UUID(), @goods_tilapia, @market_yyx, @period_202509_early, 16.80, 1, @org_id, 0);

INSERT INTO base_market_price (id, goods_id, market_id, period_id, price, price_type, org_id, is_deleted)
VALUES (UUID(), @goods_tilapia, @market_drf, @period_202509_early, 16.57, 1, @org_id, 0);

INSERT INTO base_market_price (id, goods_id, market_id, period_id, price, price_type, org_id, is_deleted)
VALUES (UUID(), @goods_tilapia, @market_avg_last, @period_202509_early, 15.90, 3, @org_id, 0);

INSERT INTO base_market_price (id, goods_id, market_id, period_id, price, price_type, org_id, is_deleted)
VALUES (UUID(), @goods_tilapia, @market_avg_cur, @period_202509_early, 13.99, 4, @org_id, 0);

-- 哈密瓜的价格数据（2025年9月上旬）
INSERT INTO base_market_price (id, goods_id, market_id, period_id, price, price_type, org_id, is_deleted)
VALUES (UUID(), @goods_melon, @market_fgw, @period_202509_early, 6.98, 2, @org_id, 0);

INSERT INTO base_market_price (id, goods_id, market_id, period_id, price, price_type, org_id, is_deleted)
VALUES (UUID(), @goods_melon, @market_fwj, @period_202509_early, 6.00, 1, @org_id, 0);

INSERT INTO base_market_price (id, goods_id, market_id, period_id, price, price_type, org_id, is_deleted)
VALUES (UUID(), @goods_melon, @market_yyx, @period_202509_early, 6.99, 1, @org_id, 0);

INSERT INTO base_market_price (id, goods_id, market_id, period_id, price, price_type, org_id, is_deleted)
VALUES (UUID(), @goods_melon, @market_drf, @period_202509_early, 7.52, 1, @org_id, 0);

INSERT INTO base_market_price (id, goods_id, market_id, period_id, price, price_type, org_id, is_deleted)
VALUES (UUID(), @goods_melon, @market_avg_last, @period_202509_early, 6.66, 3, @org_id, 0);

INSERT INTO base_market_price (id, goods_id, market_id, period_id, price, price_type, org_id, is_deleted)
VALUES (UUID(), @goods_melon, @market_avg_cur, @period_202509_early, 5.86, 4, @org_id, 0);

-- ========== 10. 插入供应商结算价格数据 ==========

-- 四季豆的供应商结算价
INSERT INTO base_supplier_price (id, goods_id, supplier_id, period_id, reference_price, float_ratio, org_id, is_deleted)
VALUES (UUID(), @goods_sijidou, @supplier_hupu, @period_202509_early, 4.57, 0.88, @org_id, 0);

INSERT INTO base_supplier_price (id, goods_id, supplier_id, period_id, reference_price, float_ratio, org_id, is_deleted)
VALUES (UUID(), @goods_sijidou, @supplier_huanghai, @period_202509_early, 4.57, 0.86, @org_id, 0);

-- 水果玉米棒的供应商结算价
INSERT INTO base_supplier_price (id, goods_id, supplier_id, period_id, reference_price, float_ratio, org_id, is_deleted)
VALUES (UUID(), @goods_corn, @supplier_hupu, @period_202509_early, 4.86, 0.88, @org_id, 0);

INSERT INTO base_supplier_price (id, goods_id, supplier_id, period_id, reference_price, float_ratio, org_id, is_deleted)
VALUES (UUID(), @goods_corn, @supplier_huanghai, @period_202509_early, 4.86, 0.86, @org_id, 0);

-- 罗非鱼的供应商结算价
INSERT INTO base_supplier_price (id, goods_id, supplier_id, period_id, reference_price, float_ratio, org_id, is_deleted)
VALUES (UUID(), @goods_tilapia, @supplier_hupu, @period_202509_early, 13.99, 0.88, @org_id, 0);

INSERT INTO base_supplier_price (id, goods_id, supplier_id, period_id, reference_price, float_ratio, org_id, is_deleted)
VALUES (UUID(), @goods_tilapia, @supplier_huanghai, @period_202509_early, 13.99, 0.86, @org_id, 0);

-- 哈密瓜的供应商结算价
INSERT INTO base_supplier_price (id, goods_id, supplier_id, period_id, reference_price, float_ratio, org_id, is_deleted)
VALUES (UUID(), @goods_melon, @supplier_hupu, @period_202509_early, 5.86, 0.88, @org_id, 0);

INSERT INTO base_supplier_price (id, goods_id, supplier_id, period_id, reference_price, float_ratio, org_id, is_deleted)
VALUES (UUID(), @goods_melon, @supplier_huanghai, @period_202509_early, 5.86, 0.86, @org_id, 0);

-- ========== 验证数据 ==========

SELECT '========== 数据插入完成 ==========' AS '';

-- 统计
SELECT '组织数量' AS '类型', COUNT(*) AS '数量' FROM base_org WHERE is_deleted = 0
UNION ALL
SELECT '商品数量', COUNT(*) FROM base_goods WHERE is_deleted = 0
UNION ALL
SELECT '市场数量', COUNT(*) FROM base_market WHERE is_deleted = 0
UNION ALL
SELECT '价格记录数', COUNT(*) FROM base_market_price WHERE is_deleted = 0
UNION ALL
SELECT '供应商数量', COUNT(*) FROM supplier WHERE is_deleted = 0
UNION ALL
SELECT '供应商价格数', COUNT(*) FROM base_supplier_price WHERE is_deleted = 0;

-- 查看示例商品的价格
SELECT '========== 四季豆价格明细 ==========' AS '';
SELECT 
  g.name AS 商品名,
  m.name AS 市场,
  CASE mp.price_type
    WHEN 1 THEN '市场价'
    WHEN 2 THEN '指导价'
    WHEN 3 THEN '上月均价'
    WHEN 4 THEN '本期均价'
  END AS 价格类型,
  mp.price AS 价格
FROM base_market_price mp
JOIN base_goods g ON mp.goods_id = g.id
JOIN base_market m ON mp.market_id = m.id
WHERE g.name = '四季豆'
  AND mp.is_deleted = 0
ORDER BY mp.price_type, m.name;

-- 查看供应商结算价
SELECT '========== 供应商结算价 ==========' AS '';
SELECT 
  g.name AS 商品名,
  s.name AS 供应商,
  sp.reference_price AS 参考价,
  CONCAT(ROUND((1 - sp.float_ratio) * 100, 2), '%') AS 下浮比例,
  sp.settlement_price AS 结算价
FROM base_supplier_price sp
JOIN base_goods g ON sp.goods_id = g.id
JOIN supplier s ON sp.supplier_id = s.id
WHERE sp.is_deleted = 0
ORDER BY g.name, s.name;
