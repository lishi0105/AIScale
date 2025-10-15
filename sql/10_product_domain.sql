USE main;

CREATE TABLE IF NOT EXISTS base_product (
  id CHAR(36) PRIMARY KEY,
  product_code VARCHAR(64) NULL,
  name VARCHAR(128) NOT NULL,
  pinyin VARCHAR(128) NULL,
  spec_id CHAR(36) NULL,
  spec_text VARCHAR(128) NULL,
  default_unit_id CHAR(36) NULL,
  image_url VARCHAR(512) NULL,
  category_id CHAR(36) NOT NULL,
  guide_price DECIMAL(10,2) NULL,
  team_id CHAR(36) NULL,
  is_deleted TINYINT(1) NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  CONSTRAINT fk_prod_cat  FOREIGN KEY (category_id)     REFERENCES base_category(id),
  CONSTRAINT fk_prod_unit FOREIGN KEY (default_unit_id) REFERENCES base_unit(id),
  CONSTRAINT fk_prod_spec FOREIGN KEY (spec_id)         REFERENCES base_spec(id)
) ENGINE=InnoDB;
CREATE INDEX idx_product_name ON base_product(name);

CREATE TABLE IF NOT EXISTS supplier (
  id CHAR(36) PRIMARY KEY,
  name VARCHAR(128) NOT NULL,
  active_from DATE NULL,
  active_to DATE NULL,
  status TINYINT NOT NULL DEFAULT 1,
  team_id CHAR(36) NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uq_supplier_name (name)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS supplier_float_ratio (
  id CHAR(36) PRIMARY KEY,
  supplier_id CHAR(36) NOT NULL,
  ratio DECIMAL(6,4) NOT NULL DEFAULT 0.0000,
  effective_from DATE NOT NULL,
  effective_to DATE NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  CONSTRAINT fk_float_supplier FOREIGN KEY (supplier_id) REFERENCES supplier(id)
) ENGINE=InnoDB;
CREATE INDEX idx_float_supplier ON supplier_float_ratio(supplier_id, effective_from, effective_to);

CREATE TABLE IF NOT EXISTS price_inquiry (
  id CHAR(36) PRIMARY KEY,
  sheet_no VARCHAR(64) NOT NULL,
  sheet_date DATE NOT NULL,
  source VARCHAR(128) NULL,
  team_id CHAR(36) NULL,
  remark VARCHAR(512) NULL,
  created_by VARCHAR(64) NULL,
  is_deleted TINYINT(1) NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uq_inquiry_no (sheet_no)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS price_inquiry_item (
  id CHAR(36) PRIMARY KEY,
  inquiry_id CHAR(36) NOT NULL,
  product_id CHAR(36) NOT NULL,
  supplier_id CHAR(36) NOT NULL,
  unit_id CHAR(36) NOT NULL,
  spec_id CHAR(36) NULL,
  spec_text VARCHAR(128) NULL,
  unit_price DECIMAL(10,2) NOT NULL,
  order_qty DECIMAL(12,4) NULL,
  accept_qty DECIMAL(12,4) NULL,
  amount DECIMAL(12,2)
    GENERATED ALWAYS AS (ROUND(COALESCE(accept_qty,0)*unit_price,2)) STORED,
  meal_id CHAR(36) NULL,
  category_note VARCHAR(64) NULL,
  float_ratio_used DECIMAL(6,4) NULL,
  team_id CHAR(36) NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  CONSTRAINT fk_item_inquiry  FOREIGN KEY (inquiry_id)  REFERENCES price_inquiry(id),
  CONSTRAINT fk_item_product  FOREIGN KEY (product_id)  REFERENCES base_product(id),
  CONSTRAINT fk_item_supplier FOREIGN KEY (supplier_id) REFERENCES supplier(id),
  CONSTRAINT fk_item_unit     FOREIGN KEY (unit_id)     REFERENCES base_unit(id),
  CONSTRAINT fk_item_spec     FOREIGN KEY (spec_id)     REFERENCES base_spec(id),
  CONSTRAINT fk_item_meal     FOREIGN KEY (meal_id)     REFERENCES menu_meal(id)
) ENGINE=InnoDB;

CREATE INDEX idx_item_inquiry  ON price_inquiry_item(inquiry_id);
CREATE INDEX idx_item_product  ON price_inquiry_item(product_id);
CREATE INDEX idx_item_supplier ON price_inquiry_item(supplier_id);
