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
  unit_id       CHAR(36)      NOT NULL COMMENT '单位ID（base_spec.id）',
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
