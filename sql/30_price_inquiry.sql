/* ======== 价格询价记录管理 ======== */
/* 说明：
   本文件定义价格询价相关的表结构，用于记录每次询价的详细信息
   - price_inquiry: 询价主表，记录每次询价的基本信息（标题、日期、品类等）
   - price_inquiry_detail: 询价商品明细表，记录每个商品的价格信息
   - price_inquiry_market_detail: 市场价格明细表，记录商品在各市场的具体价格
   - price_inquiry_supplier_detail: 供应商结算价明细表，记录供应商的结算价格
*/

USE main;

/* ---------- 价格询价主表 ---------- */
/* 场景：每月上中下旬都会对商品库的所有商品随机选择几个市场做一次询价
   示例：2025年9月上旬都匀市主要蔬菜类市场参考价
*/
CREATE TABLE IF NOT EXISTS price_inquiry (
  id              CHAR(36)     NOT NULL COMMENT '主键UUID',
  title           VARCHAR(255) NOT NULL COMMENT '询价标题（如：2025年9月上旬都匀市主要蔬菜类市场参考价）',
  year            INT          NOT NULL COMMENT '年份（如：2025）',
  month           INT          NOT NULL COMMENT '月份（1-12）',
  period          VARCHAR(16)  NOT NULL COMMENT '旬次：上旬/中旬/下旬',
  category_id     CHAR(36)     NOT NULL COMMENT '品类ID（关联base_category.id）',
  org_id          CHAR(36)     NOT NULL COMMENT '中队ID',
  inquiry_date    DATE         NOT NULL COMMENT '询价日期',
  remark          TEXT             NULL COMMENT '备注信息',
  is_deleted      TINYINT(1)   NOT NULL DEFAULT 0 COMMENT '软删标记：0=有效,1=已删除',
  created_at      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  
  PRIMARY KEY (id),
  
  -- 索引：支持按日期、品类、标题模糊搜索
  KEY idx_inquiry_date (year, month, period),
  KEY idx_inquiry_category (category_id),
  KEY idx_inquiry_org (org_id),
  KEY idx_inquiry_title (title),
  KEY idx_inquiry_inquiry_date (inquiry_date),
  
  -- 外键约束
  CONSTRAINT fk_inquiry_category FOREIGN KEY (category_id) REFERENCES base_category(id),
  
  -- 检查约束
  CONSTRAINT ck_inquiry_month CHECK (month BETWEEN 1 AND 12),
  CONSTRAINT ck_inquiry_period CHECK (period IN ('上旬', '中旬', '下旬'))
) ENGINE=InnoDB
  COMMENT='价格询价主表（记录每次询价的基本信息：标题、日期、品类等）';


/* ---------- 价格询价商品明细表 ---------- */
/* 说明：
   - 每一行对应Excel中的一个商品
   - 记录商品的发改委指导价、上月均价、本期均价等信息
*/
CREATE TABLE IF NOT EXISTS price_inquiry_detail (
  id                    CHAR(36)      NOT NULL COMMENT '主键UUID',
  inquiry_id            CHAR(36)      NOT NULL COMMENT '询价主表ID（关联price_inquiry.id）',
  goods_id              CHAR(36)      NOT NULL COMMENT '商品ID（关联base_goods.id）',
  sequence              INT           NOT NULL COMMENT '序号（对应Excel中的序号）',
  guide_price           DECIMAL(10,2)     NULL COMMENT '发改委指导价',
  last_month_avg_price  DECIMAL(10,2)     NULL COMMENT '上月均价',
  current_avg_price     DECIMAL(10,2)     NULL COMMENT '本期均价',
  remark                VARCHAR(512)      NULL COMMENT '备注',
  is_deleted            TINYINT(1)    NOT NULL DEFAULT 0 COMMENT '软删标记：0=有效,1=已删除',
  created_at            DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at            DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  
  PRIMARY KEY (id),
  
  -- 索引
  KEY idx_detail_inquiry (inquiry_id),
  KEY idx_detail_goods (goods_id),
  KEY idx_detail_sequence (inquiry_id, sequence),
  
  -- 唯一约束：同一次询价中，同一商品只能出现一次
  UNIQUE KEY uq_detail_inquiry_goods (inquiry_id, goods_id),
  
  -- 外键约束
  CONSTRAINT fk_detail_inquiry FOREIGN KEY (inquiry_id) REFERENCES price_inquiry(id) ON DELETE CASCADE,
  CONSTRAINT fk_detail_goods FOREIGN KEY (goods_id) REFERENCES base_goods(id),
  
  -- 检查约束：价格必须为正数
  CONSTRAINT ck_detail_guide_price CHECK (guide_price IS NULL OR guide_price >= 0),
  CONSTRAINT ck_detail_last_avg_price CHECK (last_month_avg_price IS NULL OR last_month_avg_price >= 0),
  CONSTRAINT ck_detail_current_avg_price CHECK (current_avg_price IS NULL OR current_avg_price >= 0)
) ENGINE=InnoDB
  COMMENT='价格询价商品明细表（记录每个商品的价格信息：指导价、上月均价、本期均价等）';


/* ---------- 市场价格明细表 ---------- */
/* 说明：
   - 记录每个商品在各个市场的具体价格
   - 市场数量和名称是变化的（如：富万家、育英巷菜市场、大润发等）
*/
CREATE TABLE IF NOT EXISTS price_inquiry_market_detail (
  id                CHAR(36)      NOT NULL COMMENT '主键UUID',
  inquiry_detail_id CHAR(36)      NOT NULL COMMENT '询价商品明细ID（关联price_inquiry_detail.id）',
  market_id         CHAR(36)      NOT NULL COMMENT '市场ID（关联base_market.id）',
  price             DECIMAL(10,2)     NULL COMMENT '市场价格',
  remark            VARCHAR(255)      NULL COMMENT '备注',
  is_deleted        TINYINT(1)    NOT NULL DEFAULT 0 COMMENT '软删标记：0=有效,1=已删除',
  created_at        DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at        DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  
  PRIMARY KEY (id),
  
  -- 索引
  KEY idx_market_detail_inquiry (inquiry_detail_id),
  KEY idx_market_detail_market (market_id),
  
  -- 唯一约束：同一商品在同一市场只能有一个价格
  UNIQUE KEY uq_market_detail_inquiry_market (inquiry_detail_id, market_id),
  
  -- 外键约束
  CONSTRAINT fk_market_detail_inquiry FOREIGN KEY (inquiry_detail_id) REFERENCES price_inquiry_detail(id) ON DELETE CASCADE,
  CONSTRAINT fk_market_detail_market FOREIGN KEY (market_id) REFERENCES base_market(id),
  
  -- 检查约束：价格必须为正数
  CONSTRAINT ck_market_detail_price CHECK (price IS NULL OR price >= 0)
) ENGINE=InnoDB
  COMMENT='市场价格明细表（记录商品在各个市场的具体价格）';


/* ---------- 供应商结算价明细表 ---------- */
/* 说明：
   - 记录供应商的结算价格和下浮比例
   - 供应商数量和名称是变化的（如：胡坤、贵海等）
   - 下浮比例对应supplier表的float_ratio字段
*/
CREATE TABLE IF NOT EXISTS price_inquiry_supplier_detail (
  id                CHAR(36)      NOT NULL COMMENT '主键UUID',
  inquiry_detail_id CHAR(36)      NOT NULL COMMENT '询价商品明细ID（关联price_inquiry_detail.id）',
  supplier_id       CHAR(36)      NOT NULL COMMENT '供应商ID（关联supplier.id）',
  settlement_price  DECIMAL(10,2)     NULL COMMENT '结算价格',
  float_ratio       DECIMAL(6,4)      NULL COMMENT '下浮比例（如：0.88表示下浮12%，0.86表示下浮14%）',
  remark            VARCHAR(255)      NULL COMMENT '备注',
  is_deleted        TINYINT(1)    NOT NULL DEFAULT 0 COMMENT '软删标记：0=有效,1=已删除',
  created_at        DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at        DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  
  PRIMARY KEY (id),
  
  -- 索引
  KEY idx_supplier_detail_inquiry (inquiry_detail_id),
  KEY idx_supplier_detail_supplier (supplier_id),
  
  -- 唯一约束：同一商品的同一供应商只能有一个结算价
  UNIQUE KEY uq_supplier_detail_inquiry_supplier (inquiry_detail_id, supplier_id),
  
  -- 外键约束
  CONSTRAINT fk_supplier_detail_inquiry FOREIGN KEY (inquiry_detail_id) REFERENCES price_inquiry_detail(id) ON DELETE CASCADE,
  CONSTRAINT fk_supplier_detail_supplier FOREIGN KEY (supplier_id) REFERENCES supplier(id),
  
  -- 检查约束：价格和比例必须为正数
  CONSTRAINT ck_supplier_detail_price CHECK (settlement_price IS NULL OR settlement_price >= 0),
  CONSTRAINT ck_supplier_detail_ratio CHECK (float_ratio IS NULL OR (float_ratio > 0 AND float_ratio <= 1))
) ENGINE=InnoDB
  COMMENT='供应商结算价明细表（记录供应商的结算价格和下浮比例）';


/* ---------- 创建视图：完整询价信息视图 ---------- */
/* 说明：
   将询价主表、商品明细、市场价格、供应商结算价等信息关联在一起
   方便查询和展示
*/
CREATE OR REPLACE VIEW v_price_inquiry_full AS
SELECT 
  pi.id AS inquiry_id,
  pi.title AS inquiry_title,
  pi.year,
  pi.month,
  pi.period,
  pi.inquiry_date,
  bc.id AS category_id,
  bc.name AS category_name,
  pid.id AS detail_id,
  pid.sequence,
  bg.id AS goods_id,
  bg.name AS goods_name,
  bs.name AS spec_name,
  bu.name AS unit_name,
  pid.guide_price,
  pid.last_month_avg_price,
  pid.current_avg_price,
  pimd.market_id,
  bm.name AS market_name,
  pimd.price AS market_price,
  pisd.supplier_id,
  s.name AS supplier_name,
  pisd.settlement_price,
  pisd.float_ratio,
  pi.created_at AS inquiry_created_at
FROM price_inquiry pi
INNER JOIN base_category bc ON pi.category_id = bc.id
INNER JOIN price_inquiry_detail pid ON pi.id = pid.inquiry_id
INNER JOIN base_goods bg ON pid.goods_id = bg.id
INNER JOIN base_spec bs ON bg.spec_id = bs.id
INNER JOIN base_unit bu ON bg.unit_id = bu.id
LEFT JOIN price_inquiry_market_detail pimd ON pid.id = pimd.inquiry_detail_id
LEFT JOIN base_market bm ON pimd.market_id = bm.id
LEFT JOIN price_inquiry_supplier_detail pisd ON pid.id = pisd.inquiry_detail_id
LEFT JOIN supplier s ON pisd.supplier_id = s.id
WHERE pi.is_deleted = 0 
  AND pid.is_deleted = 0
ORDER BY pi.inquiry_date DESC, pid.sequence ASC;


/* ---------- 常用查询示例 ---------- */

-- 1. 按年月旬次查询询价记录
-- SELECT * FROM price_inquiry 
-- WHERE year = 2025 AND month = 9 AND period = '上旬';

-- 2. 按品类查询询价记录
-- SELECT * FROM price_inquiry 
-- WHERE category_id = 'xxx-category-id';

-- 3. 模糊搜索询价标题
-- SELECT * FROM price_inquiry 
-- WHERE title LIKE '%蔬菜类%' OR title LIKE '%水产类%';

-- 4. 查询某次询价的完整信息（包括市场价格和供应商结算价）
-- SELECT * FROM v_price_inquiry_full 
-- WHERE inquiry_id = 'xxx-inquiry-id';

-- 5. 查询某个商品在某个时间段内的价格变化
-- SELECT 
--   pi.inquiry_date,
--   pi.period,
--   bg.name AS goods_name,
--   pid.current_avg_price,
--   bm.name AS market_name,
--   pimd.price AS market_price
-- FROM price_inquiry pi
-- INNER JOIN price_inquiry_detail pid ON pi.id = pid.inquiry_id
-- INNER JOIN base_goods bg ON pid.goods_id = bg.id
-- LEFT JOIN price_inquiry_market_detail pimd ON pid.id = pimd.inquiry_detail_id
-- LEFT JOIN base_market bm ON pimd.market_id = bm.id
-- WHERE bg.id = 'xxx-goods-id' 
--   AND pi.inquiry_date BETWEEN '2025-01-01' AND '2025-12-31'
-- ORDER BY pi.inquiry_date ASC;

-- 6. 查询某次询价中所有商品的市场均价和供应商结算价对比
-- SELECT 
--   pid.sequence,
--   bg.name AS goods_name,
--   pid.current_avg_price AS market_avg_price,
--   s.name AS supplier_name,
--   pisd.settlement_price,
--   pisd.float_ratio,
--   ROUND((pid.current_avg_price - pisd.settlement_price) / pid.current_avg_price * 100, 2) AS discount_percent
-- FROM price_inquiry pi
-- INNER JOIN price_inquiry_detail pid ON pi.id = pid.inquiry_id
-- INNER JOIN base_goods bg ON pid.goods_id = bg.id
-- LEFT JOIN price_inquiry_supplier_detail pisd ON pid.id = pisd.inquiry_detail_id
-- LEFT JOIN supplier s ON pisd.supplier_id = s.id
-- WHERE pi.id = 'xxx-inquiry-id'
-- ORDER BY pid.sequence ASC;
