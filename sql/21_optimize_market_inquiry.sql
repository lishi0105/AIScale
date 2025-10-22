-- 优化 price_market_inquiry 表设计
-- 添加询价日期和来源追踪字段

USE main;

-- 添加询价日期字段（便于按日期查询和统计）
ALTER TABLE price_market_inquiry 
ADD COLUMN inquiry_date DATE NOT NULL DEFAULT (CURDATE()) COMMENT '询价日期' AFTER market_name_snap;

-- 添加数据来源字段（便于追踪数据来源）
ALTER TABLE price_market_inquiry 
ADD COLUMN source VARCHAR(32) NOT NULL DEFAULT 'manual' COMMENT '数据来源：manual=人工录入, import=Excel导入, api=API接口' AFTER inquiry_date;

-- 添加索引优化查询性能
ALTER TABLE price_market_inquiry 
ADD KEY idx_inquiry_date (inquiry_date);

ALTER TABLE price_market_inquiry 
ADD KEY idx_source (source);

-- 添加复合索引优化按商品和日期查询
ALTER TABLE price_market_inquiry 
ADD KEY idx_goods_date (goods_id, inquiry_date);

-- 更新现有数据（如果需要）
-- UPDATE price_market_inquiry SET inquiry_date = (SELECT inquiry_date FROM base_price_inquiry WHERE id = inquiry_id);
-- UPDATE price_market_inquiry SET source = 'import' WHERE created_at > '2024-01-01';

-- 创建市场报价统计视图
CREATE OR REPLACE VIEW vw_market_inquiry_stats AS
SELECT 
    inquiry_id,
    item_id,
    goods_id,
    goods_name_snap,
    market_name_snap,
    COUNT(*) as market_count,
    AVG(price) as avg_price,
    MIN(price) as min_price,
    MAX(price) as max_price,
    COUNT(price) as valid_price_count,
    COUNT(*) - COUNT(price) as null_price_count
FROM price_market_inquiry 
WHERE is_deleted = 0
GROUP BY inquiry_id, item_id, goods_id, goods_name_snap, market_name_snap;

-- 创建按日期范围查询的视图
CREATE OR REPLACE VIEW vw_market_inquiry_by_date AS
SELECT 
    m.id,
    m.goods_id,
    m.inquiry_id,
    m.item_id,
    m.market_id,
    m.market_name_snap,
    m.price,
    m.inquiry_date,
    m.source,
    m.created_at,
    m.updated_at,
    i.goods_name_snap,
    i.category_name_snap,
    i.spec_name_snap,
    i.unit_name_snap,
    h.inquiry_title,
    h.org_id
FROM price_market_inquiry m
JOIN price_inquiry_item i ON i.id = m.item_id
JOIN base_price_inquiry h ON h.id = m.inquiry_id
WHERE m.is_deleted = 0 AND i.is_deleted = 0 AND h.is_deleted = 0;
