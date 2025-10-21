/* ======== 创建业务库（带注释） ======== */
CREATE DATABASE IF NOT EXISTS main
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci
  COMMENT = '市场价格管理库（基础市场、询价单与明细）';
USE main;

/* =======================================================================
   基础表：市场
   ======================================================================= */
CREATE TABLE IF NOT EXISTS base_market (
  id            CHAR(36)     NOT NULL COMMENT 'UUID',
  name          VARCHAR(64)  NOT NULL COMMENT '市场名称',
  org_id        CHAR(36)     NOT NULL COMMENT '中队ID（base_org.id）',
  code          VARCHAR(64)      NULL COMMENT '市场编码（可选，建议唯一）',
  market_type   TINYINT      NOT NULL DEFAULT 0 COMMENT '市场类型：0=未知 1=农贸 2=超市 3=电商 4=其他',
  sort          INT          NOT NULL DEFAULT 0 COMMENT '排序码',
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
  COMMENT='Base_市场（基础市场主数据）';

/* =======================================================================
   询价单（Excel 工作簿/业务日）
   - 每次导入一份 Excel 可生成一条询价单头记录
   - 每个 sheet 代表一个品类；明细行在 base_price_inquiry_item 中记录
   ======================================================================= */
CREATE TABLE IF NOT EXISTS base_price_inquiry (
  id                 CHAR(36)     NOT NULL COMMENT '主键UUID',
  org_id             CHAR(36)     NOT NULL COMMENT '中队ID（base_org.id）',
  inquiry_title      VARCHAR(64)  NOT NULL COMMENT '询价单标题（如 2025年9月上旬均价）',
  inquiry_date       DATE         NOT NULL COMMENT '询价单日期（业务日）',

  -- Excel 里展示的三个市场名称（可选）
  market_1           VARCHAR(128)     NULL COMMENT '市场1名称（如 富万家超市）',
  market_2           VARCHAR(128)     NULL COMMENT '市场2名称（如 育英巷菜市场）',
  market_3           VARCHAR(128)     NULL COMMENT '市场3名称（如 大润发）',

  -- 计算结算价用（可选）：例如 下浮 12%、14%
  supplier_1         VARCHAR(128)     NULL COMMENT '供应商/结算主体1（如 胡坎）',
  supplier_1_ratio   DECIMAL(5,2)     NULL COMMENT '结算折减百分比：12.00 表示下浮12%（相对于本期均价）',
  supplier_2         VARCHAR(128)     NULL COMMENT '供应商/结算主体2（如 贵海）',
  supplier_2_ratio   DECIMAL(5,2)     NULL COMMENT '结算折减百分比：14.00 表示下浮14%（相对于本期均价）',

  is_deleted         TINYINT(1)   NOT NULL DEFAULT 0 COMMENT '软删：0=有效 1=删除',
  created_at         DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at         DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

  PRIMARY KEY (id),
  KEY idx_inquiry_org_date (org_id, inquiry_date),
  CONSTRAINT fk_price_inquiry_org FOREIGN KEY (org_id) REFERENCES base_org(id)
) ENGINE=InnoDB COMMENT='询价单头（一次 Excel 导入/一次业务日）';

/* =======================================================================
   询价明细行（Excel 行）
   - 每个 sheet 是一个品类；用 category_id 记录，也保存快照字段
   - 行内的多个市场价格、上月均价、本期均价、结算价等按列存储
   ======================================================================= */
CREATE TABLE IF NOT EXISTS base_price_inquiry_item (
  id                   CHAR(36)      NOT NULL COMMENT '主键UUID',
  inquiry_id           CHAR(36)      NOT NULL COMMENT '询价单头ID（base_price_inquiry.id）',
  ordinal              INT           NOT NULL DEFAULT 0 COMMENT '序号（Excel行号）',

  -- 品类（sheet）与商品
  category_id          CHAR(36)          NULL COMMENT '品类ID（base_category.id）',
  category_name_snap   VARCHAR(64)       NULL COMMENT '品类名称快照（sheet名称）',
  goods_id             CHAR(36)          NULL COMMENT '商品ID（可选，若已在 base_goods 建档）',
  goods_name           VARCHAR(128)  NOT NULL COMMENT '品名（Excel 原值）',

  -- 规格/单位（允许留空，仅保存快照文本）
  spec_id              CHAR(36)          NULL COMMENT '规格ID（base_spec.id）',
  spec_name_snap       VARCHAR(32)       NULL COMMENT '规格快照（如 新鲜/500g 等）',
  unit_id              CHAR(36)          NULL COMMENT '单位ID（base_unit.id）',
  unit_name_snap       VARCHAR(32)       NULL COMMENT '单位快照（如 斤/千克/包）',

  -- 价格列
  guidance_price       DECIMAL(10,2)     NULL COMMENT '发改委指导价',
  price_market_1       DECIMAL(10,2)     NULL COMMENT '市场1价格',
  price_market_2       DECIMAL(10,2)     NULL COMMENT '市场2价格',
  price_market_3       DECIMAL(10,2)     NULL COMMENT '市场3价格',
  price_avg_last_month DECIMAL(10,2)     NULL COMMENT '上月均价',
  price_avg_current    DECIMAL(10,2)     NULL COMMENT '本期均价',
  settle_price_1       DECIMAL(10,2)     NULL COMMENT '供应商1本期结算价',
  settle_price_2       DECIMAL(10,2)     NULL COMMENT '供应商2本期结算价',

  remark               VARCHAR(255)      NULL COMMENT '备注',
  is_deleted           TINYINT(1)    NOT NULL DEFAULT 0 COMMENT '软删：0=有效 1=删除',
  created_at           DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at           DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

  PRIMARY KEY (id),
  UNIQUE KEY uq_inquiry_goods_spec_unit (inquiry_id, goods_name, spec_name_snap, unit_name_snap),
  KEY idx_item_inquiry (inquiry_id),
  KEY idx_item_category (category_id),

  CONSTRAINT fk_inquiry_item_header   FOREIGN KEY (inquiry_id) REFERENCES base_price_inquiry(id) ON DELETE CASCADE,
  CONSTRAINT fk_inquiry_item_category FOREIGN KEY (category_id) REFERENCES base_category(id),
  CONSTRAINT fk_inquiry_item_goods    FOREIGN KEY (goods_id)    REFERENCES base_goods(id),
  CONSTRAINT fk_inquiry_item_spec     FOREIGN KEY (spec_id)     REFERENCES base_spec(id),
  CONSTRAINT fk_inquiry_item_unit     FOREIGN KEY (unit_id)     REFERENCES base_unit(id)
) ENGINE=InnoDB COMMENT='询价明细行（Excel 行）';

/* 可选：一个便捷视图，按 header 的下浮比例计算结算价（若明细已给值则优先用明细） */
CREATE OR REPLACE VIEW v_price_inquiry_item AS
SELECT
  i.*, h.supplier_1, h.supplier_1_ratio, h.supplier_2, h.supplier_2_ratio,
  /* 计算列：当存在比例时，基于本期均价计算结算价 */
  CASE
    WHEN i.settle_price_1 IS NOT NULL THEN i.settle_price_1
    WHEN h.supplier_1_ratio IS NOT NULL AND i.price_avg_current IS NOT NULL
      THEN ROUND(i.price_avg_current * (1 - h.supplier_1_ratio/100), 2)
    ELSE NULL
  END AS settle_price_1_calc,
  CASE
    WHEN i.settle_price_2 IS NOT NULL THEN i.settle_price_2
    WHEN h.supplier_2_ratio IS NOT NULL AND i.price_avg_current IS NOT NULL
      THEN ROUND(i.price_avg_current * (1 - h.supplier_2_ratio/100), 2)
    ELSE NULL
  END AS settle_price_2_calc
FROM base_price_inquiry_item i
JOIN base_price_inquiry h ON h.id = i.inquiry_id;
