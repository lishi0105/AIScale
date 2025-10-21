# Excel导入API文档

## 概述

该API提供Excel文件的上传、校验和导入功能，支持切片上传以确保大文件的可靠传输。

## API端点

### 1. 上传文件切片

**端点**: `POST /api/v1/excel/upload_chunk`

**描述**: 上传文件的一个切片。适用于大文件分片上传。

**请求参数**:
- `filename` (string, 必需): 文件名
- `chunk_index` (int, 必需): 切片索引（从0开始）
- `file` (file, 必需): 文件切片数据

**请求头**:
- `Authorization`: `Bearer <token>`

**示例请求**:
```bash
curl -X POST http://localhost:8080/api/v1/excel/upload_chunk \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -F "filename=市场价格.xlsx" \
  -F "chunk_index=0" \
  -F "file=@/path/to/chunk0"
```

**响应**:
```json
{
  "ok": true,
  "chunk_index": 0,
  "message": "切片上传成功"
}
```

---

### 2. 合并文件切片

**端点**: `POST /api/v1/excel/merge_chunks`

**描述**: 合并所有已上传的文件切片并验证MD5。

**请求体**:
```json
{
  "filename": "市场价格.xlsx",
  "total_chunks": 3,
  "md5": "d41d8cd98f00b204e9800998ecf8427e"
}
```

**请求头**:
- `Authorization`: `Bearer <token>`
- `Content-Type`: `application/json`

**示例请求**:
```bash
curl -X POST http://localhost:8080/api/v1/excel/merge_chunks \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "filename": "市场价格.xlsx",
    "total_chunks": 3,
    "md5": "d41d8cd98f00b204e9800998ecf8427e"
  }'
```

**响应**:
```json
{
  "ok": true,
  "filepath": "./uploads/市场价格.xlsx",
  "message": "文件合并成功"
}
```

---

### 3. 校验Excel文件

**端点**: `POST /api/v1/excel/validate`

**描述**: 校验Excel文件结构是否符合要求。

**请求体**:
```json
{
  "filepath": "./uploads/市场价格.xlsx"
}
```

**请求头**:
- `Authorization`: `Bearer <token>`
- `Content-Type`: `application/json`

**示例请求**:
```bash
curl -X POST http://localhost:8080/api/v1/excel/validate \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "filepath": "./uploads/市场价格.xlsx"
  }'
```

**响应**:
```json
{
  "ok": true,
  "title": "2025年9月上旬都匀市主要蔬菜类市场参考价",
  "date": "2025-09-05",
  "stats": {
    "sheets": 7,
    "markets": 3,
    "suppliers": 2
  },
  "message": "Excel文件校验通过"
}
```

---

### 4. 导入Excel数据

**端点**: `POST /api/v1/excel/import`

**描述**: 将Excel数据导入到数据库中。

**请求体**:
```json
{
  "filepath": "./uploads/市场价格.xlsx",
  "org_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**请求头**:
- `Authorization`: `Bearer <token>`
- `Content-Type`: `application/json`

**示例请求**:
```bash
curl -X POST http://localhost:8080/api/v1/excel/import \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "filepath": "./uploads/市场价格.xlsx",
    "org_id": "550e8400-e29b-41d4-a716-446655440000"
  }'
```

**响应**:
```json
{
  "ok": true,
  "message": "Excel数据导入成功",
  "stats": {
    "title": "2025年9月上旬都匀市主要蔬菜类市场参考价",
    "sheets": 7,
    "markets": 3,
    "suppliers": 2
  }
}
```

---

## Excel文件格式要求

### 1. 文件结构要求

- **标题**: Excel文件的第一个sheet的A1单元格必须包含标题，格式如：`2025年9月上旬都匀市主要蔬菜类市场参考价`
- **Sheet**: 至少包含1个sheet，每个sheet代表一个商品品类（如：蔬菜类、水产海鲜等）

### 2. Sheet结构要求

每个sheet必须包含以下列：

#### 必需列：
- `品名`: 商品名称
- `规格标准`: 商品规格
- `单位`: 商品单位
- `本期均价`: 本期平均价格

#### 可选列：
- `上月均价`: 上月平均价格
- `发改委指导价`: 发改委指导价

#### 询价市场列：
- 如：`富万家超市`、`育英巷菜市场`、`大润发`等
- 至少包含1个询价市场

#### 供应商结算列：
- 格式：`供应商名称本期结算价（下浮XX%）`或`供应商名称本期结算价（上浮XX%）`
- 示例：
  - `胡坤本期结算价（下浮12%）` -> 供应商名称：胡坤，浮动比例：0.88
  - `贵海本期结算价（下浮14%）` -> 供应商名称：贵海，浮动比例：0.86
- 至少包含1个供应商

### 3. 示例表格结构

```
| 序号 | 品名 | 规格标准 | 单位 | 发改委指导价 | 富万家超市 | 育英巷菜市场 | 大润发 | 上月均价 | 本期均价 | 胡坤本期结算价（下浮12%） | 贵海本期结算价（下浮14%） |
|------|------|----------|------|--------------|------------|--------------|--------|----------|----------|---------------------------|---------------------------|
| 1    | 豇豆 | 新鲜     | 斤   |              | 3.98       | 4            | 5.98   | 5.79     | 4.65     | 4.09                      | 4.00                      |
| 2    | 无筋豆| 新鲜    | 斤   |              | 5.58       | 4.5          | 6.5    | 7.95     | 5.53     | 4.86                      | 4.75                      |
```

---

## 完整上传流程示例

### 使用JavaScript实现切片上传

```javascript
async function uploadExcel(file, orgId, token) {
  const CHUNK_SIZE = 1024 * 1024; // 1MB per chunk
  const totalChunks = Math.ceil(file.size / CHUNK_SIZE);
  
  // 1. 计算文件MD5
  const md5 = await calculateMD5(file);
  
  // 2. 上传所有切片
  for (let i = 0; i < totalChunks; i++) {
    const start = i * CHUNK_SIZE;
    const end = Math.min(start + CHUNK_SIZE, file.size);
    const chunk = file.slice(start, end);
    
    const formData = new FormData();
    formData.append('filename', file.name);
    formData.append('chunk_index', i.toString());
    formData.append('file', chunk);
    
    const response = await fetch('/api/v1/excel/upload_chunk', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`
      },
      body: formData
    });
    
    if (!response.ok) {
      throw new Error(`上传切片 ${i} 失败`);
    }
  }
  
  // 3. 合并切片
  const mergeResponse = await fetch('/api/v1/excel/merge_chunks', {
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
  
  const mergeData = await mergeResponse.json();
  const filepath = mergeData.filepath;
  
  // 4. 校验Excel
  const validateResponse = await fetch('/api/v1/excel/validate', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ filepath })
  });
  
  if (!validateResponse.ok) {
    throw new Error('Excel文件校验失败');
  }
  
  // 5. 导入数据
  const importResponse = await fetch('/api/v1/excel/import', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      filepath,
      org_id: orgId
    })
  });
  
  return await importResponse.json();
}

// MD5计算函数（需要引入SparkMD5库）
async function calculateMD5(file) {
  return new Promise((resolve, reject) => {
    const spark = new SparkMD5.ArrayBuffer();
    const fileReader = new FileReader();
    
    fileReader.onload = (e) => {
      spark.append(e.target.result);
      resolve(spark.end());
    };
    
    fileReader.onerror = reject;
    fileReader.readAsArrayBuffer(file);
  });
}
```

---

## 错误处理

### 常见错误

#### 1. 校验错误 (400 Bad Request)
```json
{
  "title": "校验Excel文件失败",
  "message": "品类名: 缺少必需列: 品名"
}
```

#### 2. MD5校验失败 (409 Conflict)
```json
{
  "title": "合并文件切片失败",
  "message": "文件MD5校验失败: 期望 xxx, 实际 yyy"
}
```

#### 3. 文件不存在 (404 Not Found)
```json
{
  "title": "导入Excel失败",
  "message": "文件不存在"
}
```

#### 4. 权限错误 (403 Forbidden)
```json
{
  "title": "导入Excel失败",
  "message": "仅管理员可导入数据"
}
```

---

## 数据处理规则

### 1. 品类处理
- 如果数据库中不存在该品类名称，自动创建新品类
- 品类名称取自sheet名称

### 2. 规格/单位处理
- 如果数据库中不存在该规格/单位，自动创建
- 如果为空，规格默认为"默认"，单位默认为"个"

### 3. 商品处理
- 根据商品名称、品类、规格、单位组合判断是否存在
- 如果不存在，自动创建新商品

### 4. 市场处理
- 如果数据库中不存在该市场名称，自动创建新市场

### 5. 供应商处理
- 如果数据库中不存在该供应商，自动创建
- 如果已存在但浮动比例不同，更新浮动比例

### 6. 价格处理
- 上月均价和本期均价直接存储
- 供应商结算价自动计算：`结算价 = 本期均价 × 浮动比例`
- 示例：本期均价4.65元，浮动比例0.88（下浮12%），结算价 = 4.65 × 0.88 = 4.09元

---

## 安全性

1. **认证**: 所有API都需要JWT token认证
2. **授权**: 仅管理员角色可以上传和导入Excel
3. **文件校验**: 通过MD5校验确保文件完整性
4. **路径安全**: 文件保存在指定目录，防止路径遍历攻击
5. **数据事务**: 所有数据库操作在事务中进行，确保原子性

---

## 性能优化

1. **切片上传**: 大文件分片上传，避免超时
2. **批量插入**: 使用GORM的批量插入优化性能
3. **事务处理**: 整个导入过程在一个事务中完成
4. **自动清理**: 导入成功后自动删除临时文件
