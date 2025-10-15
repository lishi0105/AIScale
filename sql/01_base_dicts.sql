/* ======== 创建业务库（带注释） ======== */
CREATE DATABASE IF NOT EXISTS main
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci
  COMMENT = '食品/物资管理业务库（字典、主数据、单据等）';

USE main;

/* ---------- 单位字典：斤/公斤/包/瓶/袋/条等 ---------- */
CREATE TABLE IF NOT EXISTS base_unit (
  id          CHAR(36)     NOT NULL COMMENT '主键UUID',
  name        VARCHAR(32)  NOT NULL COMMENT '单位',
  sort        INT          NOT NULL DEFAULT 0 COMMENT '排序码',
  is_deleted  TINYINT(1)   NOT NULL DEFAULT 0 COMMENT '是否已删除',
  created_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_unit_name (name),
  INDEX idx_sort (sort),
  INDEX idx_is_deleted (is_deleted)
) ENGINE=InnoDB
  COMMENT='单位字典（如 斤/公斤/包/瓶/袋/条）';

/* ---------- 规格字典：可选带换算关系 ---------- */
CREATE TABLE IF NOT EXISTS base_spec (
  id          CHAR(36)     NOT NULL COMMENT '主键UUID',
  name        VARCHAR(32)  NOT NULL COMMENT '规格',
  sort        INT          NOT NULL DEFAULT 0 COMMENT '排序码',
  is_deleted  TINYINT(1)   NOT NULL DEFAULT 0 COMMENT '是否已删除',
  created_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_spec_name (name),
  INDEX idx_sort (sort),
  INDEX idx_is_deleted (is_deleted)
) ENGINE=InnoDB
  COMMENT='规格字典（如 新鲜散装/新鲜/新鲜不傻/500g/180g）';

/* ---------- 餐次字典：早餐/午餐/晚餐等 ---------- */
CREATE TABLE IF NOT EXISTS menu_meal (
  id          CHAR(36)   NOT NULL COMMENT '主键UUID',
  name        VARCHAR(32) NOT NULL COMMENT '餐次名称（早餐/午餐/晚餐等）',
  sort        INT          NOT NULL DEFAULT 0 COMMENT '排序码',
  is_deleted  TINYINT(1) NOT NULL DEFAULT 0 COMMENT '软删标记：0=有效,1=已删除',
  created_at  DATETIME   NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at  DATETIME   NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_menu_meal_name (name),
  INDEX idx_sort (sort),
  INDEX idx_is_deleted (is_deleted)
) ENGINE=InnoDB
  COMMENT='餐次字典（如 早餐/午餐/晚餐）';

