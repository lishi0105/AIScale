# 快速测试指南

## 5分钟快速测试

### 步骤1: 部署数据库（1分钟）

```bash
cd /workspace/sql

# 方式A: 使用快速部署脚本（推荐）
./quick_start.sh

# 方式B: 手动执行
mysql -u root -p < 00_db_users.sql
mysql -u food_user -pStrongPassw0rd! main < 01_base_sys.sql
mysql -u food_user -pStrongPassw0rd! main < 10_goods_domain.sql
mysql -u food_user -pStrongPassw0rd! main < 11_market_price_system.sql
mysql -u food_user -pStrongPassw0rd! main < 12_sample_data.sql
```

### 步骤2: 验证表结构（30秒）

```bash
mysql -u food_user -pStrongPassw0rd! main -e "
SELECT table_name, table_comment 
FROM information_schema.tables 
WHERE table_schema='main' 
  AND table_name LIKE 'base_%' 
ORDER BY table_name;
"
```

**预期结果**：应该看到以下表（包含新增的价格管理表）
- base_market (市场/渠道字典)
- base_price_period (价格时期字典)
- base_market_price (市场价格记录表)
- base_supplier_price (供应商结算价格表)
- 以及其他基础表...

### 步骤3: 查询示例数据（1分钟）

```bash
# 查询所有商品
mysql -u food_user -pStrongPassw0rd! main -e "
SELECT g.name AS 商品名, c.name AS 品类, s.name AS 规格, u.name AS 单位
FROM base_goods g
JOIN base_category c ON g.category_id = c.id
JOIN base_spec s ON g.spec_id = s.id
JOIN base_unit u ON g.unit_id = u.id
WHERE g.is_deleted = 0;
"
```

**预期结果**：应该看到7个示例商品（四季豆、水果玉米棒、大白菜、罗非鱼、甲鱼、哈密瓜、香蕉）

```bash
# 查询四季豆的所有价格
mysql -u food_user -pStrongPassw0rd! main -e "
SELECT 
  m.name AS 市场,
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
WHERE g.name = '四季豆'
ORDER BY mp.price_type, m.name;
"
```

**预期结果**：应该看到四季豆在不同市场的6个价格记录

```bash
# 查询供应商结算价
mysql -u food_user -pStrongPassw0rd! main -e "
SELECT 
  g.name AS 商品名,
  s.name AS 供应商,
  sp.reference_price AS 参考价,
  sp.float_ratio AS 浮动比例,
  sp.settlement_price AS 结算价
FROM base_supplier_price sp
JOIN base_goods g ON sp.goods_id = g.id
JOIN supplier s ON sp.supplier_id = s.id
ORDER BY g.name, s.name
LIMIT 10;
"
```

**预期结果**：应该看到8个供应商结算价记录（4个商品 × 2个供应商）

### 步骤4: 测试Excel导入（2分钟）

```bash
# 安装Python依赖
pip install -r requirements.txt

# 查看导入脚本帮助
python import_market_prices.py --help

# 如果有Excel文件，可以测试导入：
# python import_market_prices.py --file your_file.xlsx
```

### 步骤5: 使用综合价格视图（30秒）

```bash
mysql -u food_user -pStrongPassw0rd! main -e "
SELECT 
  goods_name AS 商品名,
  category_name AS 品类,
  guide_price AS 指导价,
  current_avg_price AS 本期均价,
  last_month_avg_price AS 上月均价
FROM v_comprehensive_price
WHERE period_name = '2025年9月上旬'
LIMIT 10;
"
```

**预期结果**：应该看到商品的价格汇总信息

## 功能测试清单

- [ ] 数据库创建成功
- [ ] 所有表创建成功（至少包含4个新增的价格管理表）
- [ ] 示例数据插入成功
- [ ] 可以查询商品信息
- [ ] 可以查询市场价格
- [ ] 可以查询供应商结算价
- [ ] 综合价格视图工作正常
- [ ] Python导入脚本可以运行

## 测试数据说明

示例数据包括：
- 1个组织（都匀市）
- 3个品类（蔬菜类、水产类、水果类）
- 7个商品（四季豆、水果玉米棒、大白菜、罗非鱼、甲鱼、哈密瓜、香蕉）
- 6个市场（发改委、富万家超市、大润发、育英巷菜市场、上月均价、本期均价）
- 2个供应商（胡埗、黄海）
- 1个价格时期（2025年9月上旬）
- 24条市场价格记录
- 8条供应商结算价格记录

## 常见问题

### Q: 提示"Access denied for user 'food_user'"？
**A**: 确保已执行 `00_db_users.sql` 创建用户，或使用 root 用户。

### Q: 提示"外键约束失败"？
**A**: 确保按顺序执行SQL脚本：00 → 01 → 10 → 11 → 12

### Q: Python导入脚本报错？
**A**: 
1. 检查是否安装依赖：`pip install -r requirements.txt`
2. 检查数据库连接参数是否正确
3. 查看脚本输出的详细错误信息

## 性能测试（可选）

```bash
# 测试插入性能
time mysql -u food_user -pStrongPassw0rd! main -e "
INSERT INTO base_market_price (id, goods_id, market_id, period_id, price, price_type, org_id)
SELECT 
  UUID(),
  (SELECT id FROM base_goods LIMIT 1),
  (SELECT id FROM base_market LIMIT 1),
  (SELECT id FROM base_price_period LIMIT 1),
  RAND() * 100,
  1,
  (SELECT id FROM base_org LIMIT 1)
FROM 
  (SELECT 1 UNION SELECT 2 UNION SELECT 3 UNION SELECT 4 UNION SELECT 5) t1,
  (SELECT 1 UNION SELECT 2 UNION SELECT 3 UNION SELECT 4 UNION SELECT 5) t2,
  (SELECT 1 UNION SELECT 2 UNION SELECT 3 UNION SELECT 4 UNION SELECT 5) t3;
"

# 测试查询性能
time mysql -u food_user -pStrongPassw0rd! main -e "
SELECT COUNT(*) FROM base_market_price;
"
```

## 清理测试数据

```bash
# 清空所有业务数据（保留表结构）
mysql -u food_user -pStrongPassw0rd! main -e "
SET FOREIGN_KEY_CHECKS=0;
TRUNCATE TABLE base_market_price;
TRUNCATE TABLE base_supplier_price;
TRUNCATE TABLE base_goods;
TRUNCATE TABLE base_category;
TRUNCATE TABLE supplier;
TRUNCATE TABLE base_market;
TRUNCATE TABLE base_price_period;
SET FOREIGN_KEY_CHECKS=1;
"

# 重新插入示例数据
mysql -u food_user -pStrongPassw0rd! main < 12_sample_data.sql
```

## 下一步

✅ 测试通过后，可以：
1. 准备实际的Excel数据文件
2. 使用导入脚本批量导入数据
3. 根据业务需求调整表结构
4. 开发应用程序接口

更多详细信息，请参阅：
- [README.md](README.md) - 项目总览
- [README_market_price.md](README_market_price.md) - 详细使用说明
