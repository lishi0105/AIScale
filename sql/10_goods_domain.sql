/* ======== 创建业务库（带注释） ======== */
CREATE DATABASE IF NOT EXISTS main
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci
  COMMENT = '商品库管理库（商品库、品类库、询价记录、均价明细、单价表和供货商等）';
USE main;

/* ---------- 品类：蔬菜/肉类/调味品等 ---------- */
CREATE TABLE IF NOT EXISTS base_category (
  id          CHAR(36)     NOT NULL COMMENT '主键UUID',
  name        VARCHAR(64)  NOT NULL COMMENT '品类名称（同一中队内唯一）',
  org_id      CHAR(36)     NOT NULL COMMENT '中队ID',
  code        VARCHAR(64)      NULL COMMENT '品类编码',
  pinyin      VARCHAR(64)      NULL COMMENT '拼音（可选，用于搜索）',
  sort        INT          NOT NULL DEFAULT 0 COMMENT '排序码',
  is_deleted  TINYINT(1)   NOT NULL DEFAULT 0 COMMENT '软删标记：0=有效,1=已删除',
  created_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  -- 同一中队下品类名称唯一
  UNIQUE KEY uq_category_org_name (org_id, name),
  -- 品类编码全局唯一（如果业务需要）
  UNIQUE KEY uq_category_code (code)
) ENGINE=InnoDB
  COMMENT='商品品类（如 蔬菜/肉类/调味品 等）';

/* ---------- 市场字典（用于记录 Excel 中的市场名称） ---------- */
CREATE TABLE IF NOT EXISTS base_market (
  id         CHAR(36)    NOT NULL COMMENT '主键UUID',
  name       VARCHAR(64) NOT NULL COMMENT '市场名称（全局唯一）',
  city       VARCHAR(64)     NULL COMMENT '城市（可选）',
  address    VARCHAR(255)    NULL COMMENT '地址（可选）',
  is_deleted TINYINT(1)  NOT NULL DEFAULT 0 COMMENT '软删：0=有效 1=删除',
  created_at DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uq_market_name (name),
  KEY idx_market_city (city),
  KEY idx_market_del (is_deleted)
) ENGINE=InnoDB COMMENT='市场字典（富万家、育英巷菜市场、大润发等）';

/* ---------- Base_商品库 ---------- */
/* 说明：
   - spec_id     → base_spec.id（规格字典）
   - category_id → base_category.id（品类字典）
   - org_id     → 多组织/中队隔离
*/
CREATE TABLE IF NOT EXISTS base_goods (
  id            CHAR(36)      NOT NULL COMMENT '主键UUID',
  name          VARCHAR(128)  NOT NULL COMMENT '商品名称',
  code          VARCHAR(64)   NOT NULL COMMENT 'SKU/条码',
  sort          INT           NOT NULL DEFAULT 0 COMMENT '排序码',
  pinyin        VARCHAR(128)      NULL COMMENT '商品拼音（检索用）',
  spec_id       CHAR(36)      NOT NULL COMMENT '规格ID（base_spec.id）',
  unit_id       CHAR(36)      NOT NULL COMMENT '单位ID（base_unit.id）',
  image_url     VARCHAR(512)      NULL COMMENT '商品图片URL',
  description   VARCHAR(512)  NULL COMMENT     '商品描述',
  category_id   CHAR(36)      NOT NULL COMMENT '商品品类ID（base_category.id）',
  org_id        CHAR(36)      NOT NULL COMMENT '中队ID',
  is_deleted    TINYINT(1)    NOT NULL DEFAULT 0 COMMENT '软删标记：0=有效 1=删除',
  created_at    DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at    DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),

  -- 常用检索索引
  KEY idx_goods_name_py   (name, pinyin),
  KEY idx_goods_category  (category_id),
  KEY idx_goods_spec      (spec_id),
  KEY idx_goods_unit      (unit_id),

  -- 可避免同一中队下名字+规格重复
  UNIQUE KEY uq_goods_code (code),
  UNIQUE KEY uq_goods_org_name_spec_unit (org_id, name, spec_id, unit_id),

  -- 外键（确保依赖表已创建）
  CONSTRAINT fk_goods_spec     FOREIGN KEY (spec_id)     REFERENCES base_spec(id),
  CONSTRAINT fk_goods_unit     FOREIGN KEY (unit_id)     REFERENCES base_unit(id),
  CONSTRAINT fk_goods_category FOREIGN KEY (category_id) REFERENCES base_category(id)
) ENGINE=InnoDB
  COMMENT='Base_商品库（基础商品主数据：名称/拼音/规格/SKU/图片/品类）';

/* ---------- Base_询价记录 ---------- */
CREATE TABLE IF NOT EXISTS base_price_inquiry (
  id                 CHAR(36)     NOT NULL COMMENT 'UUID',
  inquiry_title      VARCHAR(64)  NOT NULL COMMENT '询价单标题',
  inquiry_date       DATE         NOT NULL COMMENT '询价单日期（业务日）',

  market_1           VARCHAR(128)     NULL COMMENT '市场1',
  market_2           VARCHAR(128)     NULL COMMENT '市场2',
  market_3           VARCHAR(128)     NULL COMMENT '市场3',

  -- Excel 导入友好：期间标签（如：2025年9月上旬）
  period_label       VARCHAR(64)      NULL COMMENT '期间标签（可选，用于导入Excel）',

  org_id             CHAR(36)     NOT NULL COMMENT '中队ID',
  is_deleted         TINYINT(1)   NOT NULL DEFAULT 0 COMMENT '软删：0=有效 1=删除',

  created_at         DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at         DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

  -- 仅对未删除行生效的唯一：利用 NULL 不参与唯一的特性
  active_title VARCHAR(64) AS (CASE WHEN is_deleted = 0 THEN inquiry_title ELSE NULL END) STORED,

  PRIMARY KEY (id),

  -- 仅“有效行”的唯一：同 org + 标题 + 业务日期 不能重复
  UNIQUE KEY uk_org_active_title_date (org_id, active_title, inquiry_date),

  -- 常用检索：组织 + 有效 + 日期倒序（InnoDB 索引默认升序，配合 ORDER BY DESC 仍可用）
  KEY idx_org_valid_date (org_id, is_deleted, inquiry_date),

  -- 如果常按标题前缀搜，可加：
  KEY idx_org_title (org_id, inquiry_title),

  -- 你原来的索引如果确实需要也可保留（但注意不要与上面的重复）
  KEY idx_inquiry_date (inquiry_date),
  KEY idx_inquiry_org  (org_id),

  -- ,CONSTRAINT chk_date_match CHECK (inquiry_date = DATE(inquiry_start_date))
) ENGINE=InnoDB COMMENT='询价记录';

/* ---------- Base_商品均价明细 ---------- */
/* 说明：
   - goods_id   → base_goods.id（商品库）
   - inquiry_id → price_inquiry.id（询价抬头）
   - avg_price  按已填写的市场价自动算“非空项平均”，都为空则为 NULL
*/
CREATE TABLE IF NOT EXISTS base_goods_avg_detail (
  id              CHAR(36)      NOT NULL COMMENT '商品均价明细Id(UUID)',
  goods_id        CHAR(36)      NOT NULL COMMENT '商品Id（base_goods.id）',
  guide_price     DECIMAL(10,2)     NULL COMMENT '指导价',

  market1_price   DECIMAL(10,2)     NULL COMMENT '市场1价格',
  market2_price   DECIMAL(10,2)     NULL COMMENT '市场2价格',
  market3_price   DECIMAL(10,2)     NULL COMMENT '市场3价格',

  -- Excel 中“上月均价/上期均价”
  prev_avg_price  DECIMAL(10,2)     NULL COMMENT '上期/上月均价（导入Excel保留）',

  -- 非空平均：有几项填几项求平均；若都为空则为 NULL
  avg_price       DECIMAL(10,2)
    GENERATED ALWAYS AS (
      CASE
        WHEN NULLIF(
               (market1_price IS NOT NULL) +
               (market2_price IS NOT NULL) +
               (market3_price IS NOT NULL), 0
             ) IS NULL
        THEN NULL
        ELSE ROUND(
          (IFNULL(market1_price,0) + IFNULL(market2_price,0) + IFNULL(market3_price,0)) /
          ((market1_price IS NOT NULL) + (market2_price IS NOT NULL) + (market3_price IS NOT NULL))
        , 2)
      END
    ) STORED COMMENT '商品均价（自动按非空项求平均，保留2位）',

  inquiry_id      CHAR(36)      NOT NULL COMMENT '询价记录Id（base_price_inquiry.id）',
  org_id          CHAR(36)          NULL COMMENT '中队Id',
  is_deleted      TINYINT(1)    NOT NULL DEFAULT 0 COMMENT '软删标记：0=有效,1=已删除',
  created_at      DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at      DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

  PRIMARY KEY (id),

  -- 同一询价单内，同一商品仅一条均价明细（按需保留）
  UNIQUE KEY uq_gad_inquiry_goods (inquiry_id, goods_id),

  -- 常用检索索引
  KEY idx_gad_inquiry (inquiry_id),
  KEY idx_gad_goods   (goods_id),

  -- 外键
  CONSTRAINT fk_gad_goods   FOREIGN KEY (goods_id)   REFERENCES base_goods(id),
  CONSTRAINT fk_gad_inquiry FOREIGN KEY (inquiry_id) REFERENCES base_price_inquiry(id)
) ENGINE=InnoDB
  COMMENT='Base_商品均价明细（按询价记录保存各市场价并生成均价）';

/* ---------- 询价结算策略（名称 + 下浮比例） ---------- */
/* 对应 Excel 中“胡坎本期结算价(下浮12%) / 贵海本期结算价(下浮14%)”。 */
CREATE TABLE IF NOT EXISTS base_inquiry_settlement (
  id            CHAR(36)     NOT NULL COMMENT '主键UUID',
  inquiry_id    CHAR(36)     NOT NULL COMMENT '询价记录Id（base_price_inquiry.id）',
  label         VARCHAR(64)  NOT NULL COMMENT '策略名称（如：胡坎/贵海）',
  down_ratio    DECIMAL(6,4) NOT NULL COMMENT '下浮比例（0.1200 表示下浮12%）',
  is_deleted    TINYINT(1)   NOT NULL DEFAULT 0 COMMENT '软删：0=有效 1=删除',
  created_at    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uq_settle_inquiry_label (inquiry_id, label),
  KEY idx_settle_inquiry (inquiry_id),
  CONSTRAINT ck_settle_ratio_range CHECK (down_ratio >= 0 AND down_ratio < 1),
  CONSTRAINT fk_settle_inquiry FOREIGN KEY (inquiry_id) REFERENCES base_price_inquiry(id)
) ENGINE=InnoDB COMMENT='询价结算策略（按名称+下浮比例）';

/* 视图：计算结算价 = 均价 × (1 - 下浮比例) */
CREATE OR REPLACE VIEW v_inquiry_goods_settlement AS
SELECT
  d.inquiry_id,
  d.goods_id,
  s.label      AS settlement_label,
  s.down_ratio AS down_ratio,
  d.avg_price,
  ROUND(d.avg_price * (1 - s.down_ratio), 2) AS settlement_price
FROM base_goods_avg_detail d
JOIN base_inquiry_settlement s ON s.inquiry_id = d.inquiry_id
WHERE d.is_deleted = 0 AND s.is_deleted = 0;

CREATE TABLE IF NOT EXISTS supplier (
  id              CHAR(36)     NOT NULL COMMENT '主键UUID',
  name            VARCHAR(128) NOT NULL COMMENT '供货商名称',
  code            VARCHAR(64)      NULL COMMENT '供货商编码',
  sort            INT          NOT NULL DEFAULT 0 COMMENT '排序：越小越前',
  pinyin          VARCHAR(64)      NULL COMMENT '拼音（可选，用于搜索）',
  status          TINYINT      NOT NULL DEFAULT 1 COMMENT '状态：1=正常,2=禁用',
  description     TEXT         NOT NULL COMMENT '供应商描述',
  contact_name    VARCHAR(64)      NULL COMMENT '联系人姓名',
  contact_phone   VARCHAR(32)      NULL COMMENT '联系电话（手机/固话）',
  contact_email   VARCHAR(128)     NULL COMMENT '联系邮箱',
  contact_address VARCHAR(255)     NULL COMMENT '联系地址',
  float_ratio     DECIMAL(6,4) NOT NULL DEFAULT 1.0000 COMMENT '浮动比例：结算价=合同价*float_ratio',
  org_id          CHAR(36)     NOT NULL COMMENT '中队ID（必填）',
  start_time      DATETIME         NULL COMMENT '开始时间', 
  end_time        DATETIME         NULL COMMENT '结束时间',
  is_deleted      TINYINT(1)   NOT NULL DEFAULT 0 COMMENT '软删标记：0=有效,1=已删除',
  created_at      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间', 
  updated_at      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  
  PRIMARY KEY (id),
  
  -- ✅ 同一组织内供货商名称必须唯一
  UNIQUE KEY uq_supplier_org_name (org_id, name),
  
  -- ✅ 同一组织内联系方式组合唯一（防止重复录入同一联系人）
  -- 如果业务允许同一供应商有多个联系人，可删除此索引
  UNIQUE KEY uq_supplier_org_contact (org_id, contact_name, contact_phone, contact_email, contact_address),
  
  -- 按组织查询很常见，建议加索引
  KEY idx_supplier_org_id (org_id),
  
  -- 时间范围查询索引
  KEY idx_supplier_active (start_time, end_time),
  
  -- 约束
  CONSTRAINT ck_supplier_ratio_pos CHECK (float_ratio > 0),
  CONSTRAINT ck_supplier_active_range CHECK (start_time IS NULL OR end_time IS NULL OR end_time >= start_time)
) ENGINE=InnoDB
  COMMENT='供货商';

/* ---------- Base_商品单价 ----------
   同一询价(inquiry) × 同一供应商 × 同一商品 只允许一条报价
   采购明细从这里取“商品单价”，再结合 supplier.float_ratio 计算结算价/金额
   - goods_id     -> base_goods.id
   - supplier_id  -> supplier.id
   - inquiry_id   -> price_inquiry.id
*/
CREATE TABLE IF NOT EXISTS base_goods_price (
  id              CHAR(36)      NOT NULL COMMENT '主键UUID',
  goods_id        CHAR(36)      NOT NULL COMMENT '商品ID（base_goods.id）',
  supplier_id     CHAR(36)      NOT NULL COMMENT '供应商ID（supplier.id）',
  inquiry_id      CHAR(36)      NOT NULL COMMENT '询价记录ID（base_price_inquiry.id）',

  unit_price      DECIMAL(10,2) NOT NULL COMMENT '商品单价（本次报价）',
  float_ratio     DECIMAL(6,4)  NOT NULL DEFAULT 1.0000 COMMENT '浮动比例快照（来自 supplier.float_ratio）',

  org_id         CHAR(36)          NULL COMMENT '中队ID',
  is_deleted      TINYINT(1)    NOT NULL DEFAULT 0 COMMENT '软删：0=有效 1=删除',
  created_at      DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at      DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

  PRIMARY KEY (id),

  -- 同一询价+供应商+商品 只允许一条报价
  UNIQUE KEY uq_bgp_inquiry_supplier_goods (inquiry_id, supplier_id, goods_id),

  -- 常用检索
  KEY idx_bgp_goods    (goods_id),
  KEY idx_bgp_supplier (supplier_id),
  KEY idx_bgp_inquiry  (inquiry_id),

  -- 外键
  CONSTRAINT fk_bgp_goods FOREIGN KEY (goods_id) REFERENCES base_goods(id),
  CONSTRAINT fk_bgp_supplier FOREIGN KEY (supplier_id) REFERENCES supplier(id),
  CONSTRAINT fk_bgp_inquiry FOREIGN KEY (inquiry_id) REFERENCES base_price_inquiry(id),

  -- 业务约束（可选）
  CONSTRAINT ck_bgp_ratio_pos CHECK (float_ratio > 0)
) ENGINE=InnoDB
  COMMENT='Base_商品单价：按 询价×供应商×商品 的报价记录';

/* =======================================================================
   Excel 导入暂存表（结构贴合 Excel 列，便于一次性导入）
   ======================================================================= */
CREATE TABLE IF NOT EXISTS stg_market_price_excel_row (
  id               BIGINT       NOT NULL AUTO_INCREMENT COMMENT '自增Id',
  file_name        VARCHAR(128)     NULL COMMENT '来源文件名',
  sheet_name       VARCHAR(64)      NULL COMMENT 'Sheet 名（如：蔬菜类/水产海鲜/水果）',
  period_label     VARCHAR(64)      NULL COMMENT '期间标签（2025年9月上旬等）',
  seq_no           INT              NULL COMMENT '序号',
  goods_name       VARCHAR(128) NOT NULL COMMENT '品名',
  spec_name        VARCHAR(32)  NOT NULL COMMENT '规格标准（如：新鲜/500g）',
  unit_name        VARCHAR(32)  NOT NULL COMMENT '单位（斤/公斤/袋等）',
  guide_price      DECIMAL(10,2)    NULL COMMENT '发改委指导价',
  market1_price    DECIMAL(10,2)    NULL COMMENT '市场1价格',
  market2_price    DECIMAL(10,2)    NULL COMMENT '市场2价格',
  market3_price    DECIMAL(10,2)    NULL COMMENT '市场3价格',
  prev_avg_price   DECIMAL(10,2)    NULL COMMENT '上月均价',
  curr_avg_price   DECIMAL(10,2)    NULL COMMENT '本期均价（Excel计算/给定）',
  settle1_label    VARCHAR(64)      NULL COMMENT '结算1名称（如：胡坎）',
  settle1_price    DECIMAL(10,2)    NULL COMMENT '结算1价格',
  settle2_label    VARCHAR(64)      NULL COMMENT '结算2名称（如：贵海）',
  settle2_price    DECIMAL(10,2)    NULL COMMENT '结算2价格',
  created_at       DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_stg_file_sheet (file_name, sheet_name)
) ENGINE=InnoDB COMMENT='Excel 导入暂存（与 Excel 列一一对应）';

/* ----------------------------------------------------------------------
   以下为“从暂存表写入标准表”的示例（按需执行/可在ETL或应用层实现）
   需替换其中的 ORG_ID、INQUIRY_ID 为实际值；示例仅用于演示。
---------------------------------------------------------------------- */
/*
-- 1) 规格字典补齐
INSERT INTO base_spec (id, name, code, sort, is_deleted)
SELECT UUID(), t.spec_name, NULL, 0, 0
FROM stg_market_price_excel_row t
LEFT JOIN base_spec s ON s.name = t.spec_name
WHERE s.id IS NULL
GROUP BY t.spec_name;

-- 2) 单位字典补齐
INSERT INTO base_unit (id, name, code, sort, is_deleted)
SELECT UUID(), t.unit_name, NULL, 0, 0
FROM stg_market_price_excel_row t
LEFT JOIN base_unit u ON u.name = t.unit_name
WHERE u.id IS NULL
GROUP BY t.unit_name;

-- 3) 商品补齐（以 org+名称+规格+单位 唯一）
INSERT INTO base_goods (id, name, code, sort, pinyin, spec_id, unit_id, image_url, description, category_id, org_id, is_deleted)
SELECT UUID(), t.goods_name, CONCAT('AUTO-', UUID()), 0, NULL,
       (SELECT id FROM base_spec WHERE name = t.spec_name LIMIT 1),
       (SELECT id FROM base_unit WHERE name = t.unit_name LIMIT 1),
       NULL, NULL,
       (SELECT id FROM base_category WHERE name = t.sheet_name LIMIT 1),
       '00000000-0000-0000-0000-000000000000',
       0
FROM stg_market_price_excel_row t
LEFT JOIN base_goods g
  ON g.name = t.goods_name
 AND g.spec_id = (SELECT id FROM base_spec WHERE name = t.spec_name LIMIT 1)
 AND g.unit_id = (SELECT id FROM base_unit WHERE name = t.unit_name LIMIT 1)
 AND g.org_id = '00000000-0000-0000-0000-000000000000'
WHERE g.id IS NULL
GROUP BY t.goods_name, t.spec_name, t.unit_name;

-- 4) 将市场价落到均价明细（avg_price 自动生成）
INSERT INTO base_goods_avg_detail (id, goods_id, guide_price, market1_price, market2_price, market3_price, prev_avg_price, inquiry_id, org_id, is_deleted)
SELECT UUID(),
       g.id,
       t.guide_price,
       t.market1_price,
       t.market2_price,
       t.market3_price,
       t.prev_avg_price,
       '00000000-0000-0000-0000-000000000000', -- INQUIRY_ID
       '00000000-0000-0000-0000-000000000000', -- ORG_ID
       0
FROM stg_market_price_excel_row t
JOIN base_goods g ON g.name = t.goods_name
  AND g.spec_id = (SELECT id FROM base_spec WHERE name = t.spec_name LIMIT 1)
  AND g.unit_id = (SELECT id FROM base_unit WHERE name = t.unit_name LIMIT 1)
  AND g.org_id = '00000000-0000-0000-0000-000000000000';

-- 5) 结算策略（示例：插入 Excel 中出现过的两个标签；下浮比例请在应用层计算）
INSERT INTO base_inquiry_settlement (id, inquiry_id, label, down_ratio)
SELECT UUID(), '00000000-0000-0000-0000-000000000000', x.label, 0.1200
FROM (
  SELECT settle1_label AS label FROM stg_market_price_excel_row WHERE settle1_label IS NOT NULL
  UNION
  SELECT settle2_label AS label FROM stg_market_price_excel_row WHERE settle2_label IS NOT NULL
) x
LEFT JOIN base_inquiry_settlement s
  ON s.inquiry_id = '00000000-0000-0000-0000-000000000000' AND s.label = x.label
WHERE x.label IS NOT NULL AND s.id IS NULL
GROUP BY x.label;
*/

