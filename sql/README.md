# 食品/物资管理系统数据库

## 项目概述

本项目提供了一套完整的食品/物资管理系统数据库方案，包括：
- 基础数据管理（组织、用户、单位、规格等）
- 商品管理（商品库、品类、供应商等）
- 市场价格管理（多市场、多时期价格记录）
- 供应商结算价格管理

## 文件结构

```
sql/
├── README.md                      # 本文件 - 项目总览
├── README_market_price.md         # 市场价格系统详细使用说明
├── requirements.txt               # Python依赖
├── quick_start.sh                 # 快速部署脚本 ⭐
├── import_market_prices.py        # Excel数据导入脚本 ⭐
│
├── 00_db_users.sql                # 数据库和用户创建
├── 01_base_sys.sql                # 系统基础表（组织、用户、设备等）
├── 10_goods_domain.sql            # 商品相关表（商品、品类、供应商等）
├── 11_market_price_system.sql     # 市场价格管理表 ⭐ 新增
└── 12_sample_data.sql             # 示例数据 ⭐ 新增
```

## 快速开始

### 方法一：使用快速部署脚本（推荐）

```bash
# 1. 进入sql目录
cd sql/

# 2. 运行快速部署脚本
./quick_start.sh

# 如果数据库有密码，使用：
./quick_start.sh --root-pass your_password

# 查看帮助
./quick_start.sh --help
```

### 方法二：手动执行SQL脚本

```bash
# 1. 创建数据库和用户
mysql -u root -p < 00_db_users.sql

# 2. 创建系统基础表
mysql -u food_user -p main < 01_base_sys.sql

# 3. 创建商品相关表
mysql -u food_user -p main < 10_goods_domain.sql

# 4. 创建市场价格管理表
mysql -u food_user -p main < 11_market_price_system.sql

# 5. 插入示例数据（可选）
mysql -u food_user -p main < 12_sample_data.sql
```

## 核心功能

### 1. 基础数据管理

**相关表**：
- `base_org`: 组织机构
- `base_user`: 用户
- `base_unit`: 计量单位
- `base_spec`: 规格
- `menu_meal`: 餐次

### 2. 商品管理

**相关表**：
- `base_category`: 商品品类
- `base_goods`: 商品库
- `supplier`: 供应商

### 3. 市场价格管理 ⭐ 新增

**相关表**：
- `base_market`: 市场/渠道字典
- `base_price_period`: 价格时期字典
- `base_market_price`: 市场价格记录表
- `base_supplier_price`: 供应商结算价格表

**功能特点**：
✅ 支持多市场价格记录（政府指导价、超市、菜市场等）  
✅ 支持多时期价格对比（上旬、中旬、下旬、月度等）  
✅ 支持供应商结算价格管理（带浮动比例）  
✅ 提供综合价格视图便于查询分析  

### 4. Excel数据导入 ⭐ 新增

**功能**：从Excel表格批量导入市场价格数据

**支持的数据格式**：
- 蔬菜类、水产类、水果类等多品类
- 多市场价格（发改委指导价、富万家超市、育英巷菜市场、大润发等）
- 上月均价、本期均价
- 供应商结算价（支持不同浮动比例）

## Excel数据导入示例

### 1. 安装Python依赖

```bash
pip install -r requirements.txt
```

### 2. 准备Excel文件

Excel文件格式示例：

| 序号 | 品名 | 规格标准 | 单位 | 发改委指导价 | 富万家超市 | 育英巷菜市场 | 大润发 | 上月均价 | 本期均价 | 胡埗本期结算价(下浮12%) | 黄海本期结算价(下浮14%) |
|------|------|----------|------|--------------|-----------|--------------|--------|----------|----------|------------------------|------------------------|
| 1    | 四季豆 | 新鲜 | 斤 | 4.98 | 5 | 5.59 | 5.98 | 5.19 | 4.57 | 4.46 |

Excel文件应包含以下sheet：
- 蔬菜类
- 水产类
- 水果类

### 3. 执行导入

```bash
# 基本用法（使用默认配置）
python import_market_prices.py --file your_excel_file.xlsx

# 自定义配置
python import_market_prices.py \
  --host localhost \
  --port 3306 \
  --user food_user \
  --password StrongPassw0rd! \
  --database main \
  --file your_excel_file.xlsx \
  --period "2025年9月上旬" \
  --start-date "2025-09-01" \
  --end-date "2025-09-10"

# 查看帮助
python import_market_prices.py --help
```

## 数据查询示例

### 查询商品价格信息

```sql
-- 查询"四季豆"在"2025年9月上旬"的所有价格
SELECT 
  g.name AS 商品名,
  m.name AS 市场名称,
  CASE mp.price_type
    WHEN 1 THEN '市场价'
    WHEN 2 THEN '指导价'
    WHEN 3 THEN '上月均价'
    WHEN 4 THEN '本期均价'
  END AS 价格类型,
  mp.price AS 价格
FROM base_market_price mp
JOIN base_goods g ON mp.goods_id = g.id
JOIN base_market m ON mp.market_id = m.id
JOIN base_price_period pp ON mp.period_id = pp.id
WHERE g.name = '四季豆'
  AND pp.name = '2025年9月上旬'
  AND mp.is_deleted = 0
ORDER BY mp.price_type, m.name;
```

### 查询供应商结算价

```sql
-- 查询胡埗供应商的结算价
SELECT 
  g.name AS 商品名,
  sup.name AS 供应商,
  sp.reference_price AS 参考价,
  CONCAT(ROUND((1 - sp.float_ratio) * 100, 2), '%') AS 下浮比例,
  sp.settlement_price AS 结算价
FROM base_supplier_price sp
JOIN base_goods g ON sp.goods_id = g.id
JOIN supplier sup ON sp.supplier_id = sup.id
WHERE sup.name = '胡埗'
  AND sp.is_deleted = 0
ORDER BY g.name;
```

### 使用综合价格视图

```sql
-- 查询所有商品的价格概览
SELECT 
  goods_name AS 商品名,
  category_name AS 品类,
  spec_name AS 规格,
  unit_name AS 单位,
  guide_price AS 指导价,
  current_avg_price AS 本期均价,
  last_month_avg_price AS 上月均价
FROM v_comprehensive_price
WHERE period_name = '2025年9月上旬'
ORDER BY category_name, goods_name;
```

## 数据库架构

### 核心表关系

```
base_org (组织)
    ↓
base_category (品类)
    ↓
base_goods (商品) ← base_spec (规格)
    ↓              ← base_unit (单位)
    ├─→ base_market_price (市场价格) ← base_market (市场)
    │                                 ← base_price_period (时期)
    └─→ base_supplier_price (供应商结算价) ← supplier (供应商)
                                            ← base_price_period (时期)
```

### 主要数据流

1. **商品基础数据录入**：
   - 创建组织 → 创建品类 → 创建规格/单位 → 创建商品

2. **市场价格录入**：
   - 创建市场 → 创建价格时期 → 录入市场价格

3. **供应商结算**：
   - 创建供应商（设置浮动比例）→ 录入结算价格

4. **Excel批量导入**：
   - 准备Excel → 执行导入脚本 → 自动创建所有关联数据

## 主要改进 ⭐

相比原有的SQL片段，本方案提供了以下改进：

### 1. 修复了SQL语法错误
- ✅ 修复了`base_inquiry`表中缺少逗号的问题
- ✅ 修复了外键引用错误
- ✅ 统一了表命名规范

### 2. 完善了表结构设计
- ✅ 添加了`base_market`（市场字典表）
- ✅ 添加了`base_price_period`（价格时期表）
- ✅ 重新设计了`base_market_price`（市场价格表）
- ✅ 添加了`base_supplier_price`（供应商结算价格表）
- ✅ 添加了综合价格视图

### 3. 支持Excel数据导入
- ✅ 提供了完整的Python导入脚本
- ✅ 支持批量导入多品类、多市场的价格数据
- ✅ 自动创建商品、市场、供应商等关联数据
- ✅ 支持重复导入（自动更新）

### 4. 提供了完整的文档和工具
- ✅ 详细的使用说明文档
- ✅ 快速部署脚本
- ✅ 示例数据
- ✅ 查询示例

## 数据库用户

| 用户名 | 密码 | 权限 |
|--------|------|------|
| food_user | StrongPassw0rd! | SELECT, INSERT, UPDATE, DELETE, CREATE, ALTER, INDEX, REFERENCES |

⚠️ **安全提示**：生产环境请修改默认密码！

## 技术栈

- **数据库**: MySQL 5.7+ / MariaDB 10.3+
- **Python**: 3.8+
- **依赖库**: pandas, openpyxl, pymysql

## 常见问题

### Q1: 执行SQL脚本时提示"外键约束失败"？
**A**: 请按顺序执行SQL脚本：00 → 01 → 10 → 11

### Q2: Excel导入失败？
**A**: 
1. 检查Excel文件格式是否正确
2. 确保已安装Python依赖：`pip install -r requirements.txt`
3. 确认数据库连接参数正确
4. 查看脚本输出的错误信息

### Q3: 如何清空测试数据？
**A**:
```sql
-- 清空价格数据（保留基础数据）
DELETE FROM base_market_price;
DELETE FROM base_supplier_price;

-- 清空所有业务数据（慎用！）
DELETE FROM base_market_price;
DELETE FROM base_supplier_price;
DELETE FROM base_goods;
DELETE FROM base_category;
DELETE FROM supplier;
DELETE FROM base_market;
DELETE FROM base_price_period;
```

## 详细文档

- 📖 [市场价格系统使用说明](README_market_price.md) - 详细的功能说明和使用指南
- 📊 [示例数据](12_sample_data.sql) - 测试数据脚本
- 🔧 [快速部署](quick_start.sh) - 一键部署脚本
- 📥 [数据导入](import_market_prices.py) - Excel导入工具

## 下一步

1. ✅ 执行数据库脚本创建表结构
2. ✅ 运行示例数据脚本测试功能
3. ✅ 准备Excel文件并导入实际数据
4. ✅ 根据业务需求调整和扩展

## 更新日志

### v1.0 (2025-10-21)
- ✨ 新增市场价格管理系统
- ✨ 新增Excel数据导入功能
- 🐛 修复原有SQL语法错误
- 📝 完善文档和示例

## 技术支持

如有问题或建议，请联系开发团队。

---

**License**: MIT  
**Version**: 1.0  
**Last Updated**: 2025-10-21
