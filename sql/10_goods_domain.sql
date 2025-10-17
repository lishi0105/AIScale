/* ======== 创建业务库（带注释） ======== */
CREATE DATABASE IF NOT EXISTS main
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci
  COMMENT = '商品库管理库（商品库、品类库、询价记录、均价明细、单价表和供货商等）';
USE main;

/* ---------- 品类：蔬菜/肉类/调味品等 ---------- */
CREATE TABLE IF NOT EXISTS base_category (
  id          CHAR(36)     NOT NULL COMMENT '主键UUID',
  name        VARCHAR(64)  NOT NULL COMMENT '品类名称（唯一）',
  code        VARCHAR(64)      NULL COMMENT '品类编码（可选，建议唯一）',
  pinyin      VARCHAR(64)      NULL COMMENT '拼音（可选，用于搜索）',
  team_id     CHAR(36)     NOT NULL COMMENT '中队ID',
  is_deleted  TINYINT(1)   NOT NULL DEFAULT 0 COMMENT '软删标记：0=有效,1=已删除',
  created_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  -- 可避免同一中队下名字+规格重复
  UNIQUE KEY uq_goods_team_name_spec (team_id, name, spec_id),
  UNIQUE KEY uq_category_name (name),
  UNIQUE KEY uq_category_code (code)
) ENGINE=InnoDB
  COMMENT='商品品类（如 蔬菜/肉类/调味品 等）';


/* ---------- Base_询价记录 ---------- */
CREATE TABLE IF NOT EXISTS price_inquiry (
  id            CHAR(36)    NOT NULL PRIMARY KEY COMMENT 'UUID',
  created_by    VARCHAR(64)     NULL COMMENT '询价人',
  inquiry_title VARCHAR(64) NOT NULL COMMENT '询价单标题',
  inquiry_date  DATE        NOT NULL COMMENT '询价单日期',
  market_1      VARCHAR(128)    NULL COMMENT '市场1',
  market_2      VARCHAR(128)    NULL COMMENT '市场2',
  market_3      VARCHAR(128)    NULL COMMENT '市场3',
  team_id       CHAR(36)   NOT NULL COMMENT '中队ID',
  is_deleted    TINYINT(1) NOT NULL DEFAULT 0 COMMENT '软删标记',
  created_at    DATETIME   NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at    DATETIME   NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  KEY idx_inquiry_date (inquiry_date),
  KEY idx_inquiry_team (team_id)
) ENGINE=InnoDB
  COMMENT='询价记录（抬头）';

/* ---------- Base_商品库 ---------- */
/* 说明：
   - spec_id     → base_spec.id（规格字典）
   - category_id → base_category.id（品类字典）
   - team_id     → 多组织/中队隔离
*/
CREATE TABLE IF NOT EXISTS base_goods (
  id            CHAR(36)      NOT NULL COMMENT '主键UUID',
  name          VARCHAR(128)  NOT NULL COMMENT '商品名称',
  pinyin        VARCHAR(128)      NULL COMMENT '商品拼音（检索用）',
  spec_id       CHAR(36)          NULL COMMENT '规格ID（base_spec.id）',
  sku           VARCHAR(64)       NULL COMMENT 'SKU/条码',
  image_url     VARCHAR(512)      NULL COMMENT '商品图片URL',
  category_id   CHAR(36)          NULL COMMENT '商品品类ID（base_category.id）',
  team_id       CHAR(36)          NULL COMMENT '中队ID',
  is_deleted    TINYINT(1)    NOT NULL DEFAULT 0 COMMENT '软删标记：0=有效 1=删除',
  created_at    DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at    DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),

  -- 常用检索索引
  KEY idx_goods_name_py   (name, pinyin),
  KEY idx_goods_category  (category_id),
  KEY idx_goods_spec      (spec_id),

  -- 可避免同一中队下名字+规格重复
  UNIQUE KEY uq_goods_team_name_spec (team_id, name, spec_id),

  -- 外键（确保依赖表已创建）
  CONSTRAINT fk_goods_spec     FOREIGN KEY (spec_id)     REFERENCES base_spec(id),
  CONSTRAINT fk_goods_category FOREIGN KEY (category_id) REFERENCES base_category(id)
) ENGINE=InnoDB
  COMMENT='Base_商品库（基础商品主数据：名称/拼音/规格/SKU/图片/品类）';

USE main;

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

  inquiry_id      CHAR(36)      NOT NULL COMMENT '询价记录Id（price_inquiry.id）',
  team_id         CHAR(36)          NULL COMMENT '中队Id',
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
  CONSTRAINT fk_gad_inquiry FOREIGN KEY (inquiry_id) REFERENCES price_inquiry(id)
) ENGINE=InnoDB
  COMMENT='Base_商品均价明细（按询价记录保存各市场价并生成均价）';

CREATE TABLE IF NOT EXISTS supplier (
  id            CHAR(36)     NOT NULL COMMENT '主键UUID',
  name          VARCHAR(128) NOT NULL COMMENT '供货商名称',
  active_start  DATE             NULL COMMENT '开始日期（可空）',
  active_end    DATE             NULL COMMENT '结束日期（可空）',
  status        TINYINT      NOT NULL DEFAULT 1 COMMENT '状态：1=正常,2=禁用',
  float_ratio   DECIMAL(6,4) NOT NULL DEFAULT 1.0000 COMMENT '浮动比例：结算价=合同价*float_ratio',
  team_id       CHAR(36)         NULL COMMENT '中队ID',
  created_at    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间', 
  updated_at    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uq_supplier_name (name),
  KEY idx_supplier_active (active_start, active_end),
  CONSTRAINT ck_supplier_ratio_pos CHECK (float_ratio > 0),
  CONSTRAINT ck_supplier_active_range CHECK (active_end IS NULL OR active_start IS NULL OR active_end >= active_start)
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
  inquiry_id      CHAR(36)      NOT NULL COMMENT '询价记录ID（price_inquiry.id）',

  unit_price      DECIMAL(10,2) NOT NULL COMMENT '商品单价（本次报价）',
  float_ratio     DECIMAL(6,4)  NOT NULL DEFAULT 1.0000 COMMENT '浮动比例快照（来自 supplier.float_ratio）',

  team_id         CHAR(36)          NULL COMMENT '中队ID',
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
  CONSTRAINT fk_bgp_inquiry FOREIGN KEY (inquiry_id) REFERENCES price_inquiry(id),

  -- 业务约束（可选）
  CONSTRAINT ck_bgp_ratio_pos CHECK (float_ratio > 0)
) ENGINE=InnoDB
  COMMENT='Base_商品单价：按 询价×供应商×商品 的报价记录';

