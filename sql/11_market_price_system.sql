/* ========================================================================
   市场价格管理系统
   用于记录商品在不同市场渠道的价格信息，支持多时期价格对比
   ======================================================================== */

USE main;

/* ---------- 市场/渠道字典 ---------- */
/* 说明：记录不同的市场来源，如发改委、富万家超市、育英巷菜市场、大润发等 */
CREATE TABLE IF NOT EXISTS base_market (
  id            CHAR(36)     NOT NULL COMMENT 'UUID',
  name          VARCHAR(64)  NOT NULL COMMENT '市场/渠道名称',
  code          VARCHAR(64)      NULL COMMENT '市场编码',
  market_type   TINYINT      NOT NULL DEFAULT 1 COMMENT '市场类型：1=政府指导 2=超市 3=菜市场 4=批发市场 5=其他',
  sort          INT          NOT NULL DEFAULT 0 COMMENT '排序码',
  org_id        CHAR(36)     NOT NULL COMMENT '中队ID',
  description   VARCHAR(255)     NULL COMMENT '市场描述',
  is_deleted    TINYINT(1)   NOT NULL DEFAULT 0 COMMENT '软删标记：0=有效 1=已删除',
  created_at    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  -- 同一组织内市场名称唯一
  UNIQUE KEY uq_market_org_name (org_id, name),
  UNIQUE KEY uq_market_code (code),
  KEY idx_market_org (org_id),
  KEY idx_market_type (market_type),
  -- 外键
  CONSTRAINT fk_market_org FOREIGN KEY (org_id) REFERENCES base_org(id)
) ENGINE=InnoDB
  COMMENT='市场/渠道字典（发改委、超市、菜市场等）';

/* ---------- 价格时期字典 ---------- */
/* 说明：记录价格采集的时期，如"2025年9月上旬"、"2025年9月中旬"等 */
CREATE TABLE IF NOT EXISTS base_price_period (
  id            CHAR(36)     NOT NULL COMMENT 'UUID',
  name          VARCHAR(64)  NOT NULL COMMENT '时期名称（如：2025年9月上旬）',
  code          VARCHAR(64)      NULL COMMENT '时期编码（如：2025-09-01）',
  start_date    DATE         NOT NULL COMMENT '开始日期',
  end_date      DATE         NOT NULL COMMENT '结束日期',
  period_type   TINYINT      NOT NULL DEFAULT 1 COMMENT '时期类型：1=上旬 2=中旬 3=下旬 4=月度 5=季度 6=年度',
  org_id        CHAR(36)     NOT NULL COMMENT '中队ID',
  is_deleted    TINYINT(1)   NOT NULL DEFAULT 0 COMMENT '软删标记：0=有效 1=已删除',
  created_at    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  -- 同一组织内时期名称唯一
  UNIQUE KEY uq_period_org_name (org_id, name),
  UNIQUE KEY uq_period_code (code),
  KEY idx_period_org (org_id),
  KEY idx_period_date (start_date, end_date),
  -- 外键
  CONSTRAINT fk_period_org FOREIGN KEY (org_id) REFERENCES base_org(id),
  -- 约束
  CONSTRAINT ck_period_date_range CHECK (end_date >= start_date)
) ENGINE=InnoDB
  COMMENT='价格时期字典（用于记录价格采集的时间段）';

/* ---------- 市场价格记录表 ---------- */
/* 说明：
   - 记录商品在不同市场、不同时期的价格
   - 支持记录指导价、市场价、上月均价、本期均价等
   - 同一商品+市场+时期只允许一条记录
*/
CREATE TABLE IF NOT EXISTS base_market_price (
  id              CHAR(36)       NOT NULL COMMENT 'UUID',
  goods_id        CHAR(36)       NOT NULL COMMENT '商品ID（base_goods.id）',
  market_id       CHAR(36)       NOT NULL COMMENT '市场ID（base_market.id）',
  period_id       CHAR(36)       NOT NULL COMMENT '时期ID（base_price_period.id）',
  
  price           DECIMAL(10,2)  NOT NULL COMMENT '价格',
  price_type      TINYINT        NOT NULL DEFAULT 1 COMMENT '价格类型：1=市场价 2=指导价 3=上月均价 4=本期均价',
  
  org_id          CHAR(36)       NOT NULL COMMENT '中队ID',
  remark          VARCHAR(255)       NULL COMMENT '备注',
  is_deleted      TINYINT(1)     NOT NULL DEFAULT 0 COMMENT '软删标记：0=有效 1=已删除',
  created_at      DATETIME       NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at      DATETIME       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  
  PRIMARY KEY (id),
  
  -- 同一商品+市场+时期+价格类型只允许一条记录
  UNIQUE KEY uq_mp_goods_market_period_type (goods_id, market_id, period_id, price_type),
  
  -- 常用检索索引
  KEY idx_mp_goods (goods_id),
  KEY idx_mp_market (market_id),
  KEY idx_mp_period (period_id),
  KEY idx_mp_org (org_id),
  KEY idx_mp_org_period (org_id, period_id),
  
  -- 外键
  CONSTRAINT fk_mp_goods FOREIGN KEY (goods_id) REFERENCES base_goods(id),
  CONSTRAINT fk_mp_market FOREIGN KEY (market_id) REFERENCES base_market(id),
  CONSTRAINT fk_mp_period FOREIGN KEY (period_id) REFERENCES base_price_period(id),
  CONSTRAINT fk_mp_org FOREIGN KEY (org_id) REFERENCES base_org(id),
  
  -- 约束
  CONSTRAINT ck_mp_price_positive CHECK (price >= 0)
) ENGINE=InnoDB
  COMMENT='市场价格记录表（记录商品在不同市场、不同时期的价格）';

/* ---------- 供应商结算价格表 ---------- */
/* 说明：
   - 记录供应商在某个时期对某个商品的结算价格
   - 结算价 = 参考价格 × 浮动比例
   - 支持记录浮动比例（如下浮12%、下浮14%）
*/
CREATE TABLE IF NOT EXISTS base_supplier_price (
  id                CHAR(36)       NOT NULL COMMENT 'UUID',
  goods_id          CHAR(36)       NOT NULL COMMENT '商品ID（base_goods.id）',
  supplier_id       CHAR(36)       NOT NULL COMMENT '供应商ID（supplier.id）',
  period_id         CHAR(36)       NOT NULL COMMENT '时期ID（base_price_period.id）',
  
  reference_price   DECIMAL(10,2)  NOT NULL COMMENT '参考价格（如本期均价）',
  float_ratio       DECIMAL(6,4)   NOT NULL DEFAULT 1.0000 COMMENT '浮动比例（如0.88表示下浮12%）',
  settlement_price  DECIMAL(10,2)  GENERATED ALWAYS AS (ROUND(reference_price * float_ratio, 2)) STORED COMMENT '结算价格（自动计算）',
  
  org_id            CHAR(36)       NOT NULL COMMENT '中队ID',
  remark            VARCHAR(255)       NULL COMMENT '备注',
  is_deleted        TINYINT(1)     NOT NULL DEFAULT 0 COMMENT '软删标记：0=有效 1=已删除',
  created_at        DATETIME       NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at        DATETIME       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  
  PRIMARY KEY (id),
  
  -- 同一供应商+商品+时期只允许一条记录
  UNIQUE KEY uq_sp_supplier_goods_period (supplier_id, goods_id, period_id),
  
  -- 常用检索索引
  KEY idx_sp_goods (goods_id),
  KEY idx_sp_supplier (supplier_id),
  KEY idx_sp_period (period_id),
  KEY idx_sp_org (org_id),
  
  -- 外键
  CONSTRAINT fk_sp_goods FOREIGN KEY (goods_id) REFERENCES base_goods(id),
  CONSTRAINT fk_sp_supplier FOREIGN KEY (supplier_id) REFERENCES supplier(id),
  CONSTRAINT fk_sp_period FOREIGN KEY (period_id) REFERENCES base_price_period(id),
  CONSTRAINT fk_sp_org FOREIGN KEY (org_id) REFERENCES base_org(id),
  
  -- 约束
  CONSTRAINT ck_sp_price_positive CHECK (reference_price >= 0),
  CONSTRAINT ck_sp_ratio_positive CHECK (float_ratio > 0)
) ENGINE=InnoDB
  COMMENT='供应商结算价格表（记录供应商结算价及浮动比例）';

/* ---------- 综合价格视图 ---------- */
/* 说明：综合查询商品的各类价格信息，便于价格对比分析 */
CREATE OR REPLACE VIEW v_comprehensive_price AS
SELECT 
  g.id AS goods_id,
  g.name AS goods_name,
  g.code AS goods_code,
  c.name AS category_name,
  s.name AS spec_name,
  u.name AS unit_name,
  p.name AS period_name,
  p.start_date AS period_start,
  p.end_date AS period_end,
  
  -- 指导价
  (SELECT mp.price 
   FROM base_market_price mp 
   JOIN base_market m ON mp.market_id = m.id 
   WHERE mp.goods_id = g.id 
     AND mp.period_id = p.id 
     AND m.market_type = 1 
     AND mp.price_type = 2
     AND mp.is_deleted = 0
   LIMIT 1) AS guide_price,
  
  -- 本期均价
  (SELECT mp.price 
   FROM base_market_price mp 
   WHERE mp.goods_id = g.id 
     AND mp.period_id = p.id 
     AND mp.price_type = 4
     AND mp.is_deleted = 0
   LIMIT 1) AS current_avg_price,
  
  -- 上月均价
  (SELECT mp.price 
   FROM base_market_price mp 
   WHERE mp.goods_id = g.id 
     AND mp.period_id = p.id 
     AND mp.price_type = 3
     AND mp.is_deleted = 0
   LIMIT 1) AS last_month_avg_price,
  
  g.org_id,
  p.id AS period_id
FROM base_goods g
JOIN base_category c ON g.category_id = c.id
JOIN base_spec s ON g.spec_id = s.id
JOIN base_unit u ON g.unit_id = u.id
CROSS JOIN base_price_period p
WHERE g.is_deleted = 0 
  AND p.is_deleted = 0;

/* ---------- 索引优化建议 ---------- */
/* 
  1. base_market_price表的查询性能优化：
     - 已创建组合索引 idx_mp_org_period(org_id, period_id)
     - 已创建唯一索引 uq_mp_goods_market_period_type
  
  2. base_supplier_price表的查询性能优化：
     - 已创建唯一索引 uq_sp_supplier_goods_period
     - 建议根据实际查询模式添加其他组合索引
*/
