/* ======== 创建业务库（带注释） ======== */
CREATE DATABASE IF NOT EXISTS main
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci
  COMMENT = '食品/物资/用户 管理业务库（字典、主数据、用户、设备与AI模型等）';

USE main;

/* =======================================================================
   字典：计量单位 / 规格 / 餐次
   ======================================================================= */
/* ---------- 计量单位（如 斤/公斤/包/瓶/袋/条） ---------- */
CREATE TABLE IF NOT EXISTS base_unit (
  id          CHAR(36)     NOT NULL COMMENT '主键UUID',
  name        VARCHAR(32)  NOT NULL COMMENT '单位名称',
  code        VARCHAR(32)      NULL COMMENT '单位编码（可选）',
  sort        INT          NOT NULL DEFAULT 0 COMMENT '排序码',
  is_deleted  TINYINT(1)   NOT NULL DEFAULT 0 COMMENT '是否已删除：0=否 1=是',
  created_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_unit_name (name),
  UNIQUE KEY uk_unit_code (code),
  KEY idx_unit_sort (sort),
  KEY idx_unit_del  (is_deleted)
) ENGINE=InnoDB
  COMMENT='单位字典';

/* ---------- 规格字典（如 新鲜/500g/180g 等） ---------- */
CREATE TABLE IF NOT EXISTS base_spec (
  id          CHAR(36)     NOT NULL COMMENT '主键UUID',
  name        VARCHAR(32)  NOT NULL COMMENT '规格名称',
  code        VARCHAR(32)      NULL COMMENT '规格编码（可选）',
  sort        INT          NOT NULL DEFAULT 0 COMMENT '排序码',
  is_deleted  TINYINT(1)   NOT NULL DEFAULT 0 COMMENT '是否已删除：0=否 1=是',
  created_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_spec_name (name),
  UNIQUE KEY uk_spec_code (code),
  KEY idx_spec_sort (sort),
  KEY idx_spec_del  (is_deleted)
) ENGINE=InnoDB
  COMMENT='规格字典';

/* ---------- 餐次字典（早餐/午餐/晚餐 等） ---------- */
CREATE TABLE IF NOT EXISTS menu_meal (
  id          CHAR(36)     NOT NULL COMMENT '主键UUID',
  name        VARCHAR(32)  NOT NULL COMMENT '餐次名称',
  code        VARCHAR(32)      NULL COMMENT '餐次编码（可选）',
  sort        INT          NOT NULL DEFAULT 0 COMMENT '排序码',
  is_deleted  TINYINT(1)   NOT NULL DEFAULT 0 COMMENT '软删：0=有效 1=已删除',
  created_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_meal_name (name),
  UNIQUE KEY uk_meal_code (code),
  KEY idx_meal_sort (sort),
  KEY idx_meal_del  (is_deleted)
) ENGINE=InnoDB
  COMMENT='餐次字典';

/* =======================================================================
   组织 / 设备 / AI 模型 / 用户
   ======================================================================= */

/* ---------- 组织架构表 ---------- */
CREATE TABLE IF NOT EXISTS base_org (
  id          CHAR(36)     NOT NULL COMMENT '组织机构Id(UUID)',
  name        VARCHAR(128) NOT NULL COMMENT '组织机构名称',
  code        VARCHAR(64)  NOT NULL COMMENT '组织机构编码',
  sort        INT          NOT NULL DEFAULT 0 COMMENT '排序码',
  is_deleted  TINYINT(1)   NOT NULL DEFAULT 0 COMMENT '是否删除：0=否 1=是',
  created_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_org_code (code),
  KEY idx_org_sort (sort),
  KEY idx_org_del  (is_deleted)
) ENGINE=InnoDB
  COMMENT='Base_组织机构表';

/* ---------- 智能秤表 ---------- */
CREATE TABLE IF NOT EXISTS base_smart_scale (
  id             CHAR(36)     NOT NULL COMMENT '设备Id(UUID)',
  mac_addr       VARCHAR(17)  NOT NULL COMMENT '设备MAC地址(AA:BB:CC:DD:EE:FF)',
  ip_addr        VARCHAR(45)      NULL COMMENT '设备IP(IPv4/IPv6)',
  org_id         CHAR(36)     NOT NULL COMMENT '组织机构Id（base_org.id）',
  org_code_snap  VARCHAR(64)      NULL COMMENT '组织编码快照（便于对账）',
  is_deleted     TINYINT(1)   NOT NULL DEFAULT 0 COMMENT '是否删除：0=否 1=是',
  created_at     DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at     DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_scale_mac (mac_addr),
  KEY idx_scale_org (org_id),
  CONSTRAINT fk_scale_org FOREIGN KEY (org_id) REFERENCES base_org(id)
) ENGINE=InnoDB
  COMMENT='Base_智能秤表';

/* ---------- AI 模型表 ---------- */
CREATE TABLE IF NOT EXISTS base_ai_model (
  id            CHAR(36)     NOT NULL COMMENT '模型Id(UUID)',
  org_id        CHAR(36)     NOT NULL COMMENT '组织Id（base_org.id）',
  scale_id      CHAR(36)         NULL COMMENT '秤Id（base_smart_scale.id）',
  train_epochs  INT          NOT NULL DEFAULT 0 COMMENT '学习次数/训练轮数',
  model_url     VARCHAR(512)     NULL COMMENT '模型URL地址（对象存储/文件服务）',
  is_deleted    TINYINT(1)   NOT NULL DEFAULT 0 COMMENT '是否删除：0=否 1=是',
  created_at    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  KEY idx_model_org (org_id),
  KEY idx_model_scale (scale_id),
  CONSTRAINT fk_model_org   FOREIGN KEY (org_id)   REFERENCES base_org(id),
  CONSTRAINT fk_model_scale FOREIGN KEY (scale_id) REFERENCES base_smart_scale(id)
) ENGINE=InnoDB
  COMMENT='Base_AI模型表';

/* -------------- 用户 模型表 ------------------ */
CREATE TABLE IF NOT EXISTS base_user (
  id             CHAR(36)     NOT NULL COMMENT '用户_id(UUID)',
  org_id         CHAR(36)         NULL COMMENT '组织机构id（base_org.id）',
  role           TINYINT NOT NULL DEFAULT 1 COMMENT '角色 0管理员 1用户',
  username       VARCHAR(64)  NOT NULL COMMENT '用户名',
  password_hash  VARCHAR(255) NOT NULL COMMENT '用户密码Hash（建议BCrypt/Argon2）',
  is_deleted     TINYINT(1)   NOT NULL DEFAULT 0 COMMENT '是否删除标记：0=否 1=是',
  created_at     DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  last_login_at  DATETIME         NULL COMMENT '登录时间',
  login_ip       VARCHAR(45)      NULL COMMENT '登录ip（支持IPv4/IPv6）',
  updated_at     DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

  PRIMARY KEY (id),
  UNIQUE KEY uk_user_org_username (org_id, username),
  KEY idx_user_org (org_id),
  KEY idx_user_login (last_login_at),
  CONSTRAINT fk_user_org FOREIGN KEY (org_id) REFERENCES base_org(id)
) ENGINE=InnoDB
  COMMENT='Base_用户表';