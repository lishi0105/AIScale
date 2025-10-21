# 数据库改进说明

## 概述

根据您提供的Excel表格（2025年9月上旬都匀市主要蔬菜类/水产类/水果类市场参考价），对数据库进行了全面的完善和扩展。

## 您原始SQL中的问题

### 1. 语法错误
```sql
-- ❌ 错误示例（缺少逗号）
KEY idx_price_inquiry_goods_spec_unit (goods_id, goods_spec_id, goods_unit_id)
UNIQUE KEY uq_price_inquiry_goods_spec_unit (goods_id, goods_spec_id, goods_unit_id)
-- ✅ 应该是：
KEY idx_price_inquiry_goods_spec_unit (goods_id, goods_spec_id, goods_unit_id),
UNIQUE KEY uq_price_inquiry_goods_spec_unit (goods_id, goods_spec_id, goods_unit_id)
```

### 2. 缺少必要的表
- ❌ `base_org` 表未定义，但在多处被引用
- ❌ `base_market` 表定义不完整
- ❌ 缺少价格时期管理表

### 3. 设计不足
- ❌ 无法记录多个市场的价格（Excel中有发改委、富万家、育英巷、大润发等）
- ❌ 无法记录不同时期的价格对比（上月均价、本期均价）
- ❌ 供应商结算价与市场价混在一起，不便于管理

## 改进方案

### 1. 修复并完善了表结构

#### 新增表：

**`base_market`** - 市场/渠道字典
- 用途：管理不同的价格来源（政府、超市、菜市场等）
- 字段：市场名称、市场类型、组织ID等
- 支持类型：政府指导、超市、菜市场、批发市场、其他

**`base_price_period`** - 价格时期字典
- 用途：管理价格采集的时间段
- 字段：时期名称、开始日期、结束日期、时期类型等
- 支持类型：上旬、中旬、下旬、月度、季度、年度

**`base_market_price`** - 市场价格记录表
- 用途：记录商品在不同市场、不同时期的价格
- 字段：商品ID、市场ID、时期ID、价格、价格类型等
- 价格类型：市场价、指导价、上月均价、本期均价
- 唯一约束：同一商品+市场+时期+价格类型只允许一条记录

**`base_supplier_price`** - 供应商结算价格表
- 用途：记录供应商的结算价格
- 字段：商品ID、供应商ID、时期ID、参考价格、浮动比例、结算价格（自动计算）
- 计算公式：结算价格 = 参考价格 × 浮动比例
- 示例：本期均价5.00元，下浮12%（0.88），结算价4.40元

#### 新增视图：

**`v_comprehensive_price`** - 综合价格视图
- 用途：便于查询和对比各类价格
- 字段：商品信息、指导价、本期均价、上月均价等

### 2. Excel数据导入功能

创建了完整的Python导入脚本 `import_market_prices.py`：

**功能特点**：
- ✅ 自动读取Excel文件（支持xlsx格式）
- ✅ 自动创建商品、市场、供应商等基础数据
- ✅ 批量导入多品类、多市场的价格数据
- ✅ 支持重复导入（自动更新现有数据）
- ✅ 详细的日志输出

**支持的Excel格式**：
```
Sheet名称：蔬菜类、水产类、水果类
列名：品名、规格标准、单位、发改委指导价、富万家超市、育英巷菜市场、
     大润发、上月均价、本期均价、胡埗本期结算价、黄海本期结算价
```

**使用示例**：
```bash
python import_market_prices.py \
  --file 都匀市主要商品市场参考价.xlsx \
  --period "2025年9月上旬" \
  --start-date "2025-09-01" \
  --end-date "2025-09-10"
```

### 3. 部署和测试工具

**`quick_start.sh`** - 快速部署脚本
- 一键部署整个数据库系统
- 自动执行所有SQL脚本
- 可选择是否插入示例数据

**`12_sample_data.sql`** - 示例数据
- 包含7个示例商品
- 包含6个市场
- 包含2个供应商
- 包含24条市场价格记录
- 包含8条供应商结算价格记录

**`QUICK_TEST.md`** - 5分钟快速测试指南
- 详细的测试步骤
- 预期结果说明
- 常见问题解决方案

### 4. 完善的文档

**`README.md`** - 项目总览
- 文件结构说明
- 快速开始指南
- 核心功能介绍
- 常见问题解答

**`README_market_price.md`** - 市场价格系统详细使用说明
- 表结构详细说明
- 数据库安装步骤
- Excel导入详细说明
- 丰富的查询示例
- 性能优化建议

## 数据映射关系

### Excel → 数据库映射

| Excel列 | 数据库表 | 字段/说明 |
|---------|----------|-----------|
| 品名 | base_goods | name（商品名称） |
| 规格标准 | base_spec | name（规格名称） |
| 单位 | base_unit | name（单位名称） |
| 发改委指导价 | base_market_price | market_id=发改委, price_type=2(指导价) |
| 富万家超市 | base_market_price | market_id=富万家超市, price_type=1(市场价) |
| 育英巷菜市场 | base_market_price | market_id=育英巷菜市场, price_type=1(市场价) |
| 大润发 | base_market_price | market_id=大润发, price_type=1(市场价) |
| 上月均价 | base_market_price | price_type=3(上月均价) |
| 本期均价 | base_market_price | price_type=4(本期均价) |
| 胡埗本期结算价 | base_supplier_price | supplier_id=胡埗, float_ratio=0.88 |
| 黄海本期结算价 | base_supplier_price | supplier_id=黄海, float_ratio=0.86 |

## 使用流程

### 1. 部署数据库
```bash
cd /workspace/sql
./quick_start.sh
```

### 2. 准备Excel文件
- 格式参考：您提供的都匀市市场参考价表格
- 需要包含：蔬菜类、水产类、水果类等sheet

### 3. 安装Python依赖
```bash
pip install -r requirements.txt
```

### 4. 导入数据
```bash
python import_market_prices.py --file your_excel_file.xlsx
```

### 5. 查询数据
```sql
-- 查询四季豆的所有价格
SELECT * FROM v_comprehensive_price 
WHERE goods_name = '四季豆';

-- 查询供应商结算价
SELECT g.name, s.name, sp.settlement_price
FROM base_supplier_price sp
JOIN base_goods g ON sp.goods_id = g.id
JOIN supplier s ON sp.supplier_id = s.id;
```

## 技术特点

### 1. 数据完整性
- ✅ 所有外键约束正确
- ✅ 唯一约束防止重复数据
- ✅ CHECK约束确保数据合法性

### 2. 查询性能
- ✅ 在高频查询字段上建立索引
- ✅ 组合索引优化复杂查询
- ✅ 视图简化常用查询

### 3. 扩展性
- ✅ 支持多组织（中队）隔离
- ✅ 支持多时期价格对比
- ✅ 支持多市场价格采集
- ✅ 灵活的价格类型扩展

### 4. 易用性
- ✅ 一键部署脚本
- ✅ Excel批量导入
- ✅ 丰富的查询示例
- ✅ 详细的文档说明

## 文件清单

### SQL脚本
- ✅ `00_db_users.sql` - 数据库和用户创建
- ✅ `01_base_sys.sql` - 系统基础表
- ✅ `10_goods_domain.sql` - 商品相关表
- ✅ `11_market_price_system.sql` - ⭐ 市场价格管理表（新增）
- ✅ `12_sample_data.sql` - ⭐ 示例数据（新增）

### 脚本和工具
- ✅ `import_market_prices.py` - ⭐ Excel导入脚本（新增）
- ✅ `quick_start.sh` - ⭐ 快速部署脚本（新增）
- ✅ `requirements.txt` - ⭐ Python依赖（新增）

### 文档
- ✅ `README.md` - ⭐ 项目总览（新增）
- ✅ `README_market_price.md` - ⭐ 详细使用说明（新增）
- ✅ `QUICK_TEST.md` - ⭐ 快速测试指南（新增）
- ✅ `CHANGES.md` - ⭐ 改进说明（本文件）

## 对比优势

### 原方案 vs 改进方案

| 功能 | 原方案 | 改进方案 |
|------|--------|----------|
| 语法正确性 | ❌ 有错误 | ✅ 完全正确 |
| 表结构完整性 | ❌ 缺少关键表 | ✅ 完整设计 |
| 多市场支持 | ❌ 不支持 | ✅ 完全支持 |
| 多时期对比 | ❌ 不支持 | ✅ 完全支持 |
| Excel导入 | ❌ 无 | ✅ 全自动 |
| 供应商结算 | ❌ 混乱 | ✅ 独立管理 |
| 部署便利性 | ❌ 需手动 | ✅ 一键部署 |
| 文档完善度 | ❌ 无 | ✅ 详细文档 |
| 示例数据 | ❌ 无 | ✅ 完整示例 |
| 查询便利性 | ❌ 复杂 | ✅ 提供视图 |

## 后续建议

### 1. 数据录入
- 使用提供的Excel导入脚本批量导入历史数据
- 或手动录入小批量数据进行测试

### 2. 应用开发
- 基于数据库开发Web应用或API
- 实现价格查询、对比、分析等功能

### 3. 数据分析
- 价格趋势分析（环比、同比）
- 市场价格对比
- 供应商成本分析

### 4. 性能优化
- 根据实际数据量调整索引
- 考虑使用分区表（大数据量）
- 定期清理历史数据

## 联系支持

如有问题或需要进一步的帮助，请查看：
1. `README.md` - 项目总览
2. `README_market_price.md` - 详细使用说明
3. `QUICK_TEST.md` - 快速测试指南

---

**改进完成时间**: 2025-10-21  
**版本**: 1.0  
**状态**: ✅ 已完成所有改进
