/* ======== 创建业务库（带注释） ======== */
CREATE DATABASE IF NOT EXISTS main
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci
  COMMENT = '市场价格管理库（市场库、询价记录、均价明细、单价表和供货商等）';
USE main;

/* ---------- 市场基础表 ---------- */
CREATE TABLE IF NOT EXISTS base_market (
  id            CHAR(36)     NOT NULL COMMENT 'UUID',
  name          VARCHAR(64)  NOT NULL COMMENT '市场名称',
  org_id        CHAR(36)     NOT NULL COMMENT '中队ID',
  code          VARCHAR(64)      NULL COMMENT '市场编码',
  sort          INT          NOT NULL DEFAULT 0 COMMENT '排序码',
  is_deleted    TINYINT(1)   NOT NULL DEFAULT 0 COMMENT '软删标记：0=有效 1=已删除',
  created_at    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  -- 同一组织内市场名称唯一
  UNIQUE KEY uq_market_org_name (org_id, name),
  UNIQUE KEY uq_market_code (code),
  KEY idx_market_org (org_id),
  -- 外键
  CONSTRAINT fk_market_org FOREIGN KEY (org_id) REFERENCES base_org(id)
) ENGINE=InnoDB
  COMMENT='Base_市场（基础市场主数据：名称）';


/* ---------- 市场价格期次表 ---------- */
/* 说明：
   - 用于记录每个价格统计期次，如"2025年9月上旬"
   - 一个期次对应多条商品价格记录
*/
CREATE TABLE IF NOT EXISTS market_price_period (
  id            CHAR(36)     NOT NULL COMMENT '主键UUID',
  title         VARCHAR(128) NOT NULL COMMENT '期次标题（如：2025年9月上旬都匀市主要蔬菜类市场参考价）',
  period_year   INT          NOT NULL COMMENT '年份（如：2025）',
  period_month  INT          NOT NULL COMMENT '月份（1-12）',
  period_type   VARCHAR(32)  NOT NULL COMMENT '期次类型（上旬/中旬/下旬/月度）',
  category_id   CHAR(36)     NOT NULL COMMENT '品类ID（base_category.id）',
  org_id        CHAR(36)     NOT NULL COMMENT '组织ID（base_org.id）',
  publish_date  DATE             NULL COMMENT '发布日期',
  is_deleted    TINYINT(1)   NOT NULL DEFAULT 0 COMMENT '软删标记：0=有效 1=已删除',
  created_at    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  -- 同一组织、同一品类、同一年月期次唯一
  UNIQUE KEY uq_period_org_cat_time (org_id, category_id, period_year, period_month, period_type),
  KEY idx_period_org (org_id),
  KEY idx_period_category (category_id),
  KEY idx_period_date (period_year, period_month),
  -- 外键
  CONSTRAINT fk_period_org FOREIGN KEY (org_id) REFERENCES base_org(id),
  CONSTRAINT fk_period_category FOREIGN KEY (category_id) REFERENCES base_category(id)
) ENGINE=InnoDB
  COMMENT='市场价格期次表（如2025年9月上旬等）';


/* ---------- 市场价格记录表 ---------- */
/* 说明：
   - 存储每个商品在某个期次的各个市场价格
   - 包含发改委指导价、各市场价格、均价、结算价等
*/
CREATE TABLE IF NOT EXISTS market_price_record (
  id                       CHAR(36)       NOT NULL COMMENT '主键UUID',
  period_id                CHAR(36)       NOT NULL COMMENT '期次ID（market_price_period.id）',
  goods_id                 CHAR(36)       NOT NULL COMMENT '商品ID（base_goods.id）',
  goods_name               VARCHAR(128)   NOT NULL COMMENT '商品名称（冗余字段，便于查询）',
  spec_name                VARCHAR(32)        NULL COMMENT '规格名称（冗余字段）',
  unit_name                VARCHAR(32)        NULL COMMENT '单位名称（冗余字段）',
  seq_num                  INT            NOT NULL DEFAULT 0 COMMENT '序号（Excel中的序号）',
  
  -- 发改委指导价
  ndrc_guide_price         DECIMAL(10,2)      NULL COMMENT '发改委指导价',
  
  -- 各市场价格（根据Excel动态添加）
  fuwanjia_price           DECIMAL(10,2)      NULL COMMENT '富万家超市价格',
  yuying_market_price      DECIMAL(10,2)      NULL COMMENT '育英巷菜市场价格',
  daxiangfa_price          DECIMAL(10,2)      NULL COMMENT '大湘发价格',
  
  -- 均价
  last_month_avg_price     DECIMAL(10,2)      NULL COMMENT '上月均价',
  current_period_avg_price DECIMAL(10,2)      NULL COMMENT '本期均价',
  
  -- 结算价
  hupu_settlement_price    DECIMAL(10,2)      NULL COMMENT '胡埔本期结算价（下浮12%）',
  guihai_settlement_price  DECIMAL(10,2)      NULL COMMENT '贵海本期结算价（下浮14%）',
  
  -- 备注和状态
  remark                   TEXT               NULL COMMENT '备注',
  is_deleted               TINYINT(1)     NOT NULL DEFAULT 0 COMMENT '软删标记：0=有效 1=已删除',
  created_at               DATETIME       NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at               DATETIME       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  
  PRIMARY KEY (id),
  -- 同一期次下同一商品唯一
  UNIQUE KEY uq_record_period_goods (period_id, goods_id),
  KEY idx_record_period (period_id),
  KEY idx_record_goods (goods_id),
  KEY idx_record_seq (seq_num),
  -- 外键
  CONSTRAINT fk_record_period FOREIGN KEY (period_id) REFERENCES market_price_period(id),
  CONSTRAINT fk_record_goods FOREIGN KEY (goods_id) REFERENCES base_goods(id)
) ENGINE=InnoDB
  COMMENT='市场价格记录表（存储每个商品在各市场的价格）';


/* ---------- 市场价格明细表（可选，如需记录更灵活的市场-价格关系） ---------- */
/* 说明：
   - 如果市场数量和名称经常变化，可以用此表代替market_price_record中的固定字段
   - 采用 period_id + goods_id + market_id 的组合存储价格
*/
CREATE TABLE IF NOT EXISTS market_price_detail (
  id            CHAR(36)       NOT NULL COMMENT '主键UUID',
  period_id     CHAR(36)       NOT NULL COMMENT '期次ID（market_price_period.id）',
  goods_id      CHAR(36)       NOT NULL COMMENT '商品ID（base_goods.id）',
  market_id     CHAR(36)       NOT NULL COMMENT '市场ID（base_market.id）',
  price         DECIMAL(10,2)  NOT NULL COMMENT '价格',
  price_type    VARCHAR(32)    NOT NULL COMMENT '价格类型（市场价/指导价/结算价等）',
  is_deleted    TINYINT(1)     NOT NULL DEFAULT 0 COMMENT '软删标记：0=有效 1=已删除',
  created_at    DATETIME       NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at    DATETIME       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  
  PRIMARY KEY (id),
  -- 同一期次、同一商品、同一市场、同一价格类型唯一
  UNIQUE KEY uq_detail_period_goods_market_type (period_id, goods_id, market_id, price_type),
  KEY idx_detail_period (period_id),
  KEY idx_detail_goods (goods_id),
  KEY idx_detail_market (market_id),
  -- 外键
  CONSTRAINT fk_detail_period FOREIGN KEY (period_id) REFERENCES market_price_period(id),
  CONSTRAINT fk_detail_goods FOREIGN KEY (goods_id) REFERENCES base_goods(id),
  CONSTRAINT fk_detail_market FOREIGN KEY (market_id) REFERENCES base_market(id)
) ENGINE=InnoDB
  COMMENT='市场价格明细表（灵活存储商品在各市场的价格）';
