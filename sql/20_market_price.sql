/* ======== 创建业务库（带注释） ======== */
CREATE DATABASE IF NOT EXISTS main
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci
  COMMENT = '市场价格管理库（市场库、询价记录、均价明细、单价表和供货商等）';
USE main;

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

/* =====================================================================
   询价与报价数据表
   - 支持任意数量的市场与供应商参与同一轮询价
   - 可按 年/月/上中下旬、品类、标题模糊 搜索
   ===================================================================== */

/* ---------- 询价单（表头） ---------- */
CREATE TABLE IF NOT EXISTS base_price_inquiry (
  id              CHAR(36)     NOT NULL COMMENT '主键UUID',
  org_id          CHAR(36)     NOT NULL COMMENT '中队ID（base_org.id）',
  inquiry_title   VARCHAR(64)  NOT NULL COMMENT '询价单标题（如：2025年9月上旬都匀市主要水产类市场参考价）',
  inquiry_date    DATE         NOT NULL COMMENT '业务日期（取当月任意一天）',

  -- 便于检索的冗余/生成列（MySQL/MariaDB 支持）
  inquiry_year    SMALLINT     AS (YEAR(inquiry_date)) STORED,
  inquiry_month   TINYINT      AS (MONTH(inquiry_date)) STORED,
  inquiry_ten_day TINYINT      AS (CASE
                                   WHEN DAYOFMONTH(inquiry_date) <= 10 THEN 1
                                   WHEN DAYOFMONTH(inquiry_date) <= 20 THEN 2
                                   ELSE 3 END) STORED COMMENT '旬：1=上旬 2=中旬 3=下旬',

  is_deleted      TINYINT(1)   NOT NULL DEFAULT 0 COMMENT '软删：0=有效 1=删除',
  created_at      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_inquiry_org (org_id),
  KEY idx_inquiry_date (inquiry_date),
  KEY idx_inquiry_period (org_id, inquiry_year, inquiry_month, inquiry_ten_day),
  KEY idx_inquiry_title (inquiry_title),
  KEY idx_inquiry_markets (market_1, market_2, market_3),
  CONSTRAINT fk_inquiry_org FOREIGN KEY (org_id) REFERENCES base_org(id)
) ENGINE=InnoDB
  COMMENT='询价单（表头）';

/* ---------- 询价商品明细（每行一个商品） ---------- */
CREATE TABLE IF NOT EXISTS price_inquiry_item (
  id                   CHAR(36)      NOT NULL COMMENT '主键UUID',
  inquiry_id           CHAR(36)      NOT NULL COMMENT 'base_price_inquiry.id',
  goods_id             CHAR(36)      NOT NULL COMMENT 'base_goods.id',
  category_id          CHAR(36)      NOT NULL COMMENT 'base_category.id',
  spec_id              CHAR(36)          NULL COMMENT 'base_spec.id（快照）',
  unit_id              CHAR(36)          NULL COMMENT 'base_unit.id（快照）',

  -- 名称/规格/单位快照，防止后续被修改影响历史
  goods_name_snap      VARCHAR(128)  NOT NULL COMMENT '商品名称快照',
  category_name_snap   VARCHAR(64)   NOT NULL COMMENT '品类名称快照',
  spec_name_snap       VARCHAR(32)       NULL COMMENT '规格名称快照（如：新鲜/500g）',
  unit_name_snap       VARCHAR(32)       NULL COMMENT '单位名称快照（如：斤/公斤）',

  guide_price          DECIMAL(12,2)     NULL COMMENT '发改委指导价',
  last_month_avg_price DECIMAL(12,2)     NULL COMMENT '上月均价',
  current_avg_price    DECIMAL(12,2)     NULL COMMENT '本期均价（可由市场报价汇总后回填）',

  sort                 INT           NOT NULL DEFAULT 0 COMMENT '排序码（保留与导出一致顺序）',
  is_deleted           TINYINT(1)    NOT NULL DEFAULT 0 COMMENT '软删：0=有效 1=删除',
  created_at           DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at           DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  PRIMARY KEY (id),
  KEY idx_item_inquiry (inquiry_id),
  KEY idx_item_category (category_id),
  KEY idx_item_goods (goods_id),

  CONSTRAINT fk_item_inquiry   FOREIGN KEY (inquiry_id)  REFERENCES base_price_inquiry(id) ON DELETE CASCADE,
  CONSTRAINT fk_item_goods     FOREIGN KEY (goods_id)    REFERENCES base_goods(id),
  CONSTRAINT fk_item_category  FOREIGN KEY (category_id) REFERENCES base_category(id),
  CONSTRAINT fk_item_spec      FOREIGN KEY (spec_id)     REFERENCES base_spec(id),
  CONSTRAINT fk_item_unit      FOREIGN KEY (unit_id)     REFERENCES base_unit(id)
) ENGINE=InnoDB
  COMMENT='询价商品明细（每行一个商品）';

/* ---------- 市场报价（N 个市场可变） ---------- */
CREATE TABLE IF NOT EXISTS price_market_quote (
  id               CHAR(36)     NOT NULL COMMENT '主键UUID',
  inquiry_id       CHAR(36)     NOT NULL COMMENT 'base_price_inquiry.id（冗余便于查询）',
  item_id          CHAR(36)     NOT NULL COMMENT 'price_inquiry_item.id',
  market_id        CHAR(36)         NULL COMMENT 'base_market.id（可为空，仅保存名称）',
  market_name_snap VARCHAR(64)  NOT NULL COMMENT '市场名称快照（如：富万家/育英巷/大润发）',
  price            DECIMAL(12,2) NOT NULL COMMENT '该市场的单价',
  created_at       DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at       DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uq_quote_item_market (item_id, market_name_snap),
  KEY idx_quote_inquiry (inquiry_id),
  KEY idx_quote_item (item_id),
  KEY idx_quote_market (market_id),
  CONSTRAINT fk_quote_inquiry FOREIGN KEY (inquiry_id) REFERENCES base_price_inquiry(id) ON DELETE CASCADE,
  CONSTRAINT fk_quote_item    FOREIGN KEY (item_id)    REFERENCES price_inquiry_item(id) ON DELETE CASCADE,
  CONSTRAINT fk_quote_market  FOREIGN KEY (market_id)  REFERENCES base_market(id)
) ENGINE=InnoDB
  COMMENT='市场报价（询价明细在多个市场的报价记录）';

/* ---------- 供应商结算（N 个供应商可变） ---------- */
CREATE TABLE IF NOT EXISTS price_supplier_settlement (
  id                    CHAR(36)     NOT NULL COMMENT '主键UUID',
  inquiry_id            CHAR(36)     NOT NULL COMMENT 'base_price_inquiry.id（冗余便于查询）',
  item_id               CHAR(36)     NOT NULL COMMENT 'price_inquiry_item.id',
  supplier_id           CHAR(36)         NULL COMMENT 'supplier.id（可为空，仅保存名称与比例）',
  supplier_name_snap    VARCHAR(128) NOT NULL COMMENT '供应商名称快照（如：胡坤/贵海）',
  float_ratio_snap      DECIMAL(6,4) NOT NULL COMMENT '浮动比例快照（如：0.88 表示下浮12%）',
  settlement_price      DECIMAL(12,2) NOT NULL COMMENT '本期结算价',
  created_at            DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at            DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uq_settle_item_supplier (item_id, supplier_name_snap),
  KEY idx_settle_inquiry (inquiry_id),
  KEY idx_settle_item (item_id),
  KEY idx_settle_supplier (supplier_id),
  CONSTRAINT fk_settle_inquiry  FOREIGN KEY (inquiry_id) REFERENCES base_price_inquiry(id) ON DELETE CASCADE,
  CONSTRAINT fk_settle_item     FOREIGN KEY (item_id)    REFERENCES price_inquiry_item(id) ON DELETE CASCADE,
  CONSTRAINT fk_settle_supplier FOREIGN KEY (supplier_id) REFERENCES supplier(id)
) ENGINE=InnoDB
  COMMENT='供应商结算（按供应商下浮比例计算的结算价）';

/* ---------- 汇总/搜索视图：便于按日期/品类/标题检索 ---------- */
CREATE OR REPLACE VIEW vw_price_inquiry_item_search AS
SELECT
  h.id                AS inquiry_id,
  h.org_id            AS org_id,
  h.inquiry_title     AS inquiry_title,
  h.inquiry_date      AS inquiry_date,
  h.inquiry_year      AS inquiry_year,
  h.inquiry_month     AS inquiry_month,
  h.inquiry_ten_day   AS inquiry_ten_day,
  i.id                AS item_id,
  i.goods_id          AS goods_id,
  i.goods_name_snap   AS goods_name,
  i.category_id       AS category_id,
  i.category_name_snap AS category_name,
  i.spec_name_snap    AS spec_name,
  i.unit_name_snap    AS unit_name,
  i.guide_price,
  i.last_month_avg_price,
  -- 如果未回填 current_avg_price，则以市场报价实算
  COALESCE(i.current_avg_price,
           (SELECT AVG(q.price) FROM price_market_quote q WHERE q.item_id = i.id)) AS current_avg_price
FROM base_price_inquiry h
JOIN price_inquiry_item i ON i.inquiry_id = h.id
WHERE h.is_deleted = 0 AND i.is_deleted = 0;
