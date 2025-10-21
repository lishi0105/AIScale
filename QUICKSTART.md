# Excel导入功能快速入门

## 🚀 5分钟快速上手

### 前提条件

1. 已部署并运行foodapp服务
2. 拥有管理员账户和JWT token
3. 准备好符合格式的Excel文件

---

## 📝 步骤1: 准备Excel文件

### Excel文件格式要求

您的Excel文件需要包含：

1. **标题** (A1单元格): `2025年9月上旬都匀市主要蔬菜类市场参考价`
2. **Sheet**: 每个sheet代表一个品类（如：蔬菜类、水产海鲜）
3. **必需列**:
   - 品名
   - 规格标准
   - 单位
   - 本期均价
4. **询价市场列**: 如 `富万家超市`、`育英巷菜市场`
5. **供应商列**: 如 `胡坤本期结算价（下浮12%）`

### 示例表格

```
标题: 2025年9月上旬都匀市主要蔬菜类市场参考价

| 序号 | 品名 | 规格标准 | 单位 | 富万家超市 | 育英巷菜市场 | 上月均价 | 本期均价 | 胡坤本期结算价（下浮12%） |
|------|------|----------|------|------------|--------------|----------|----------|---------------------------|
| 1    | 豇豆 | 新鲜     | 斤   | 3.98       | 4            | 5.79     | 4.65     | 4.09                      |
```

---

## 📝 步骤2: 获取认证Token

```bash
# 登录获取token
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "your_password"
  }'

# 响应示例
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

保存`access_token`，后续API调用需要使用。

---

## 📝 步骤3: 上传Excel文件

### 方法1: 使用测试脚本（推荐）

```bash
# 简单一行命令完成整个流程
./test_excel_import.sh market_price.xlsx <org_id> <your_token>
```

### 方法2: 手动调用API

#### 3.1 计算MD5

```bash
# Linux/Mac
md5sum market_price.xlsx

# 或
md5 -q market_price.xlsx
```

#### 3.2 上传文件（小文件）

如果文件小于1MB，可以直接上传：

```bash
curl -X POST http://localhost:8080/api/v1/excel/upload_chunk \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -F "filename=market_price.xlsx" \
  -F "chunk_index=0" \
  -F "file=@market_price.xlsx"
```

#### 3.3 合并文件

```bash
curl -X POST http://localhost:8080/api/v1/excel/merge_chunks \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "filename": "market_price.xlsx",
    "total_chunks": 1,
    "md5": "YOUR_FILE_MD5"
  }'
```

#### 3.4 导入数据

```bash
curl -X POST http://localhost:8080/api/v1/excel/import \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "filepath": "./uploads/market_price.xlsx",
    "org_id": "YOUR_ORG_ID"
  }'
```

---

## ✅ 步骤4: 验证导入结果

### 查看询价单

```bash
curl -X POST http://localhost:8080/api/v1/inquiry/list_inquiries \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"org_id": "YOUR_ORG_ID"}'
```

### 查看商品

```bash
curl -X POST http://localhost:8080/api/v1/goods/list_goods \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"org_id": "YOUR_ORG_ID"}'
```

### 查看市场

```bash
curl -X POST http://localhost:8080/api/v1/market/list_markets \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"org_id": "YOUR_ORG_ID"}'
```

---

## 🔍 常见问题

### Q1: 上传失败怎么办？

**A**: 检查以下几点：
1. Token是否有效（未过期）
2. 用户是否是管理员角色
3. 文件是否存在
4. 网络连接是否正常

### Q2: 校验失败怎么办？

**A**: 检查Excel文件格式：
1. A1单元格是否包含标题
2. 是否包含必需列：品名、规格标准、单位、本期均价
3. 是否至少有1个询价市场
4. 是否至少有1个供应商（含浮动比例）

### Q3: 导入失败怎么办？

**A**: 查看错误信息，可能原因：
1. 数据库连接失败
2. org_id不存在
3. 数据格式错误
4. 权限不足

### Q4: 如何处理大文件？

**A**: 对于大文件（>5MB），建议：
1. 使用切片上传
2. 每个切片1MB
3. 使用测试脚本自动处理

### Q5: 导入后数据在哪里？

**A**: 数据分布在以下表中：
- 品类: `base_category`
- 规格: `base_spec`
- 单位: `base_unit`
- 商品: `base_goods`
- 市场: `base_market`
- 供应商: `supplier`
- 询价单: `base_price_inquiry`
- 询价明细: `price_inquiry_item`
- 市场报价: `price_market_inquiry`
- 供应商结算: `price_supplier_settlement`

---

## 💻 JavaScript示例

```javascript
// 完整的上传流程
async function importExcel(file, orgId, token) {
  const CHUNK_SIZE = 1024 * 1024; // 1MB
  const totalChunks = Math.ceil(file.size / CHUNK_SIZE);
  const md5 = await calculateMD5(file);
  
  // 1. 上传切片
  for (let i = 0; i < totalChunks; i++) {
    const chunk = file.slice(i * CHUNK_SIZE, (i + 1) * CHUNK_SIZE);
    const formData = new FormData();
    formData.append('filename', file.name);
    formData.append('chunk_index', i);
    formData.append('file', chunk);
    
    await fetch('/api/v1/excel/upload_chunk', {
      method: 'POST',
      headers: { 'Authorization': `Bearer ${token}` },
      body: formData
    });
  }
  
  // 2. 合并切片
  const mergeRes = await fetch('/api/v1/excel/merge_chunks', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      filename: file.name,
      total_chunks: totalChunks,
      md5: md5
    })
  });
  
  const { filepath } = await mergeRes.json();
  
  // 3. 导入数据
  const importRes = await fetch('/api/v1/excel/import', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ filepath, org_id: orgId })
  });
  
  return await importRes.json();
}

// MD5计算（需要SparkMD5库）
async function calculateMD5(file) {
  return new Promise((resolve, reject) => {
    const spark = new SparkMD5.ArrayBuffer();
    const reader = new FileReader();
    
    reader.onload = (e) => {
      spark.append(e.target.result);
      resolve(spark.end());
    };
    
    reader.onerror = reject;
    reader.readAsArrayBuffer(file);
  });
}
```

---

## 📚 更多文档

- **API详细文档**: [docs/excel_import_api.md](docs/excel_import_api.md)
- **功能说明**: [EXCEL_IMPORT_README.md](EXCEL_IMPORT_README.md)
- **实现总结**: [IMPLEMENTATION_SUMMARY.md](IMPLEMENTATION_SUMMARY.md)
- **完成报告**: [COMPLETION_REPORT.md](COMPLETION_REPORT.md)

---

## 🆘 获取帮助

如果遇到问题：

1. 查看API文档了解详细用法
2. 使用测试脚本验证流程
3. 检查服务器日志
4. 联系技术支持

---

## 🎉 开始使用

现在您已经了解了基本用法，可以开始导入您的Excel数据了！

```bash
# 使用测试脚本快速开始
./test_excel_import.sh your_file.xlsx your_org_id your_token

# 祝使用愉快！
```
