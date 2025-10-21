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
  KEY idx_market_type (market_type),
  -- 外键
  CONSTRAINT fk_market_org FOREIGN KEY (org_id) REFERENCES base_org(id)
) ENGINE=InnoDB
  COMMENT='Base_市场（基础市场主数据：名称）';
