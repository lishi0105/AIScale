USE main;

/* ---------- 品类：蔬菜/肉类/调味品等 ---------- */
CREATE TABLE IF NOT EXISTS base_category (
  id          CHAR(36)     NOT NULL COMMENT '主键UUID',
  name        VARCHAR(64)  NOT NULL COMMENT '品类名称（唯一）',
  code        VARCHAR(64)      NULL COMMENT '品类编码（可选，建议唯一）',
  pinyin      VARCHAR(64)      NULL COMMENT '拼音（可选，用于搜索）',
  is_deleted  TINYINT(1)   NOT NULL DEFAULT 0 COMMENT '软删标记：0=有效,1=已删除',
  created_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uq_category_name (name),
  UNIQUE KEY uq_category_code (code)
) ENGINE=InnoDB
  COMMENT='商品品类（如 蔬菜/肉类/调味品 等）';

CREATE TABLE IF NOT EXISTS supplier (
  id            CHAR(36)     NOT NULL COMMENT '主键UUID',
  name          VARCHAR(128) NOT NULL COMMENT '供货商名称',
  active_start  DATE             NULL COMMENT '开始日期（可空）',
  active_end    DATE             NULL COMMENT '结束日期（可空）',
  status        TINYINT      NOT NULL DEFAULT 1 COMMENT '状态：1=正常,2=禁用',
  float_ratio   DECIMAL(6,4) NOT NULL DEFAULT 0.0000 COMMENT '浮动比例：结算价=合同价*float_ratio',
  team_id       CHAR(36)         NULL COMMENT '中队ID',
  created_at    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间', 
  updated_at    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uq_supplier_name (name),
  KEY idx_supplier_active (active_start, active_end)
) ENGINE=InnoDB
  COMMENT='供货商';

CREATE TABLE IF NOT EXISTS price_inquiry (
  id            CHAR(36)    NOT NULL PRIMARY KEY COMMENT 'UUID',
  created_by    VARCHAR(64)     NULL COMMENT '询价人',
  inquiry_title VARCHAR(64) NOT NULL COMMENT '询价单标题',
  inquiry_date  DATE        NOT NULL COMMENT '询价单日期',
  market_1      VARCHAR(128)    NULL COMMENT '市场1',
  market_2      VARCHAR(128)    NULL COMMENT '市场2',
  market_3      VARCHAR(128)    NULL COMMENT '市场3',
  market_4      VARCHAR(128)    NULL COMMENT '市场4',
  market_5      VARCHAR(128)    NULL COMMENT '市场5',
  team_id       CHAR(36)   NOT NULL COMMENT '中队ID',
  is_deleted    TINYINT(1) NOT NULL DEFAULT 0 COMMENT '软删标记',
  created_at    DATETIME   NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at    DATETIME   NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  KEY idx_inquiry_date (inquiry_date),
  KEY idx_inquiry_team (team_id)
) ENGINE=InnoDB
  COMMENT='询价记录（抬头）';

CREATE TABLE IF NOT EXISTS base_product (
  id             CHAR(36)     NOT NULL COMMENT '明细Id(UUID)',
  product_id     CHAR(36)         NULL COMMENT '商品主数据ID（可空，若只有名称）',
  product_name   VARCHAR(128) NOT NULL COMMENT '商品名称',
  image_url      VARCHAR(512)     NULL COMMENT '商品图片URL',
  pinyin         VARCHAR(128)     NULL COMMENT '商品拼音（检索用）',
  sku            VARCHAR(64)      NULL COMMENT '商品SKU/编码',
  category_id    CHAR(36)         NULL COMMENT '商品品类Id（base_category.id）',
  guide_price    DECIMAL(10,2)    NULL COMMENT '指导价',

  market1_price  DECIMAL(10,2)    NULL COMMENT '市场1价格',
  market2_price  DECIMAL(10,2)    NULL COMMENT '市场2价格',
  market3_price  DECIMAL(10,2)    NULL COMMENT '市场3价格',

  /* 非空平均：有几项填几项求平均；若都为空则为 NULL */
  avg_price      DECIMAL(10,2)
    GENERATED ALWAYS AS (
      CASE
        WHEN NULLIF(
               (market1_price IS NOT NULL) +
               (market2_price IS NOT NULL) +
               (market3_price IS NOT NULL), 0
             ) IS NULL
        THEN NULL
        ELSE ROUND(
          (IFNULL(market1_price,0)+IFNULL(market2_price,0)+IFNULL(market3_price,0)) /
          ((market1_price IS NOT NULL)+(market2_price IS NOT NULL)+(market3_price IS NOT NULL))
        , 2)
      END
    ) STORED COMMENT '商品均价（自动按非空项求平均，保留2位）',

  inquiry_id     CHAR(36)     NOT NULL COMMENT '询价记录Id（price_inquiry.id）',
  team_id        CHAR(36)         NULL COMMENT '中队Id',
  is_deleted     TINYINT(1)   NOT NULL DEFAULT 0 COMMENT '软删标记：0=有效,1=已删除',
  created_at     DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at     DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

  PRIMARY KEY (id),

  -- 关联 & 检索索引
  KEY idx_bp_inquiry   (inquiry_id),
  KEY idx_bp_category  (category_id),
  KEY idx_bp_name_py   (product_name, pinyin),
  UNIQUE KEY uq_bp_inquiry_name (inquiry_id, product_name)  -- 同一询价单避免重复名称

  ,CONSTRAINT fk_bp_category FOREIGN KEY (category_id) REFERENCES base_category(id)
  ,CONSTRAINT fk_bp_inquiry  FOREIGN KEY (inquiry_id)  REFERENCES price_inquiry(id)
) ENGINE=InnoDB
  COMMENT='Base_商品均价明细（按单据行记录市场价与计算均价）';