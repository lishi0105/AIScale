# Excel导入功能实现说明

## 功能概述

本项目实现了完整的Excel文件导入功能，用于导入市场价格询价数据。支持文件切片上传、MD5校验、Excel结构验证和数据库导入。

## 实现的功能

### 1. Excel文件校验 ✅

实现了完整的Excel文件结构校验，包括：

- ✅ **标题校验**: Excel必须包含标题（A1单元格），格式如"2025年9月上旬都匀市主要蔬菜类市场参考价"
- ✅ **Sheet校验**: Excel必须包含至少1个sheet，每个sheet代表一个品类
- ✅ **必需列校验**: 每个sheet必须包含"品名"、"规格标准"、"单位"、"本期均价"四个列
- ✅ **询价市场校验**: 每个sheet必须包含至少1个询价市场
- ✅ **供应商校验**: 每个sheet必须包含至少1个供应商，并解析浮动比例

### 2. 文件上传功能 ✅

实现了安全可靠的文件上传机制：

- ✅ **切片上传**: 支持大文件分片上传，避免超时
- ✅ **MD5校验**: 上传完成后验证文件完整性
- ✅ **临时存储**: 切片临时存储，合并后删除
- ✅ **路径安全**: 文件保存在指定目录，防止路径遍历

### 3. 数据导入功能 ✅

实现了完整的数据导入流程：

#### 品类处理
- ✅ 自动识别sheet名称作为品类名称
- ✅ 如果数据库不存在该品类，自动创建
- ✅ 关联到指定的组织（org_id）

#### 规格标准处理
- ✅ 自动提取"规格标准"列
- ✅ 如果数据库不存在该规格，自动创建
- ✅ 空值默认为"默认"

#### 单位处理
- ✅ 自动提取"单位"列
- ✅ 如果数据库不存在该单位，自动创建
- ✅ 空值默认为"个"

#### 商品处理
- ✅ 根据品名、品类、规格、单位组合判断是否存在
- ✅ 如果不存在，自动创建新商品
- ✅ 自动关联品类、规格、单位

#### 询价市场处理
- ✅ 自动提取询价市场列（如"富万家超市"、"育英巷菜市场"等）
- ✅ 如果数据库不存在该市场，自动创建
- ✅ 关联到指定的组织

#### 供应商处理
- ✅ 从列名中提取供应商名称和浮动比例
- ✅ 支持"下浮"和"上浮"两种方式
- ✅ 如果数据库不存在该供应商，自动创建
- ✅ 如果已存在但浮动比例不同，自动更新

#### 价格数据处理
- ✅ 上月均价存储到`price_inquiry_item.last_month_avg_price`
- ✅ 本期均价存储到`price_inquiry_item.current_avg_price`
- ✅ 市场报价存储到`price_market_inquiry`表
- ✅ 供应商结算价自动计算：`结算价 = 本期均价 × 浮动比例`

### 4. 事务处理 ✅

- ✅ 整个导入过程在一个事务中完成
- ✅ 任何错误都会回滚，确保数据一致性
- ✅ 成功后自动清理临时文件

## 项目结构

```
/workspace/
├── internal/
│   ├── service/
│   │   └── excel/
│   │       └── service.go          # Excel导入服务逻辑
│   ├── server/
│   │   ├── handler/
│   │   │   └── excel.go            # Excel API处理器
│   │   └── server.go               # 路由注册
│   └── domain/                      # 数据模型
├── docs/
│   └── excel_import_api.md         # API文档
├── uploads/                         # 文件上传目录
└── EXCEL_IMPORT_README.md          # 本文档
```

## API接口

实现了4个核心API接口：

1. **POST /api/v1/excel/upload_chunk** - 上传文件切片
2. **POST /api/v1/excel/merge_chunks** - 合并文件切片并校验MD5
3. **POST /api/v1/excel/validate** - 校验Excel文件结构
4. **POST /api/v1/excel/import** - 导入Excel数据到数据库

详细API文档请参考: [docs/excel_import_api.md](docs/excel_import_api.md)

## 数据库表关系

导入过程涉及以下数据库表：

1. **base_category** - 品类表
2. **base_spec** - 规格表
3. **base_unit** - 单位表
4. **base_goods** - 商品表
5. **base_market** - 市场表
6. **supplier** - 供应商表
7. **base_price_inquiry** - 询价单表头
8. **price_inquiry_item** - 询价商品明细
9. **price_market_inquiry** - 市场报价
10. **price_supplier_settlement** - 供应商结算

## Excel示例格式

### 标题（A1单元格）
```
2025年9月上旬都匀市主要蔬菜类市场参考价
```

### 表格结构
```
| 序号 | 品名   | 规格标准 | 单位 | 发改委指导价 | 富万家超市 | 育英巷菜市场 | 大润发 | 上月均价 | 本期均价 | 胡坤本期结算价（下浮12%） | 贵海本期结算价（下浮14%） |
|------|--------|----------|------|--------------|------------|--------------|--------|----------|----------|---------------------------|---------------------------|
| 1    | 豇豆   | 新鲜     | 斤   |              | 3.98       | 4            | 5.98   | 5.79     | 4.65     | 4.09                      | 4.00                      |
| 2    | 无筋豆 | 新鲜     | 斤   |              | 5.58       | 4.5          | 6.5    | 7.95     | 5.53     | 4.86                      | 4.75                      |
```

## 依赖库

新增依赖：
- `github.com/xuri/excelize/v2` - Excel文件读写库

## 安全性

1. **认证授权**: 所有API需要JWT token，仅管理员可操作
2. **文件校验**: MD5校验确保文件完整性
3. **路径安全**: 文件保存在指定目录
4. **事务处理**: 保证数据一致性
5. **输入验证**: 严格的参数校验

## 使用示例

### 1. 上传Excel文件

```javascript
// 1. 分片上传
for (let i = 0; i < totalChunks; i++) {
  await uploadChunk(file, i);
}

// 2. 合并切片
const { filepath } = await mergeChunks(filename, totalChunks, md5);

// 3. 校验Excel
await validateExcel(filepath);

// 4. 导入数据
await importExcel(filepath, orgId);
```

### 2. 使用curl上传（小文件）

```bash
# 上传切片
curl -X POST http://localhost:8080/api/v1/excel/upload_chunk \
  -H "Authorization: Bearer TOKEN" \
  -F "filename=test.xlsx" \
  -F "chunk_index=0" \
  -F "file=@test.xlsx"

# 合并并校验
curl -X POST http://localhost:8080/api/v1/excel/merge_chunks \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"filename":"test.xlsx","total_chunks":1,"md5":"xxx"}'

# 导入数据
curl -X POST http://localhost:8080/api/v1/excel/import \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"filepath":"./uploads/test.xlsx","org_id":"org-uuid"}'
```

## 错误处理

所有错误都会返回详细的错误信息：

- **400 Bad Request**: 参数错误、Excel格式错误
- **403 Forbidden**: 权限不足
- **404 Not Found**: 文件不存在
- **409 Conflict**: MD5校验失败
- **500 Internal Server Error**: 服务器内部错误

## 性能优化

1. **切片上传**: 避免大文件上传超时
2. **批量插入**: 优化数据库写入性能
3. **事务处理**: 减少数据库往返次数
4. **自动清理**: 及时释放存储空间

## 测试建议

### 1. 单元测试
- 测试Excel解析逻辑
- 测试供应商浮动比例解析
- 测试日期提取

### 2. 集成测试
- 测试完整上传流程
- 测试MD5校验
- 测试数据库导入

### 3. 边界测试
- 空Excel文件
- 缺少必需列
- 格式错误的供应商信息
- 无效的日期格式

## 未来改进方向

1. **进度反馈**: 实现导入进度实时反馈
2. **批量导入**: 支持多个Excel文件批量导入
3. **导入历史**: 记录导入历史和统计
4. **数据验证**: 增强价格数据的合理性验证
5. **导出功能**: 实现数据导出为Excel

## 技术栈

- **语言**: Go 1.24
- **框架**: Gin
- **ORM**: GORM
- **Excel库**: excelize/v2
- **数据库**: MySQL

## 许可证

根据项目主许可证
