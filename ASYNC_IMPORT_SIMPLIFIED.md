# 询价单异步导入功能说明（简化版）

## 更新日期
2025-10-22

## 功能概述

将询价单导入接口从同步改为异步，解决大文件导入时HTTP超时的问题。

## 核心改动

### 1. 使用内存存储任务状态
- **无需数据库表**：使用 `sync.Map` 在内存中存储导入任务状态
- **轻量级方案**：不需要创建额外的数据库表和迁移脚本
- **并发安全**：使用Go标准库的并发安全Map

### 2. 修改的文件

#### Service层 (`internal/service/inquiry/import.go`)
- 新增 `ImportTask` 结构体（内存存储）
- 修改 `InquiryImportService` 添加 `tasks sync.Map` 字段
- 新增方法：
  - `CreateImportTask()` - 创建任务（返回任务ID）
  - `GetImportTask()` - 查询任务状态
  - `UpdateTaskStatus()` - 更新进度
  - `UpdateTaskError()` - 记录错误
  - `UpdateTaskSuccess()` - 标记成功
  - `ImportExcelDataAsync()` - 异步导入（带进度更新）
- **Bug修复**：修正了供应商浮动比例计算错误
  - 下浮12%：从错误的 `0.12` 修正为 `0.88`
  - 上浮12%：从错误的 `-0.12` 修正为 `1.12`

#### Handler层 (`internal/server/handler/inquiry.go`)
- 修改 `importInquiry()` - 改为异步处理
  - 校验通过后立即返回 HTTP 202 和任务ID
  - 启动goroutine在后台执行导入
- 新增 `getImportStatus()` - 查询导入进度
- 新增路由 `POST /inquiry_import/import_status`

### 3. API接口

#### 提交导入（异步）
```http
POST /inquiry_import/import_inquiry
Content-Type: multipart/form-data

参数:
- org_id: 组织ID
- file: Excel文件

响应 (HTTP 202 Accepted):
{
  "ok": true,
  "message": "Excel文件校验通过，开始异步导入",
  "task_id": "550e8400-e29b-41d4-a716-446655440000",
  "stats": {
    "title": "...",
    "sheets": 3,
    "markets": 5,
    "suppliers": 2
  }
}
```

#### 查询进度
```http
POST /inquiry_import/import_status
Content-Type: application/json

请求:
{
  "task_id": "550e8400-e29b-41d4-a716-446655440000"
}

响应 (HTTP 200):
{
  "ok": true,
  "task_id": "...",
  "status": "processing",  // pending/processing/success/failed
  "progress": 45,
  "total_sheets": 3,
  "processed_sheets": 1,
  "file_name": "...",
  "message": "正在导入中...",
  "created_at": "...",
  "updated_at": "..."
}
```

## 使用流程

1. **上传文件** → 返回 task_id (立即返回，不等待导入完成)
2. **轮询查询** → 根据 task_id 查询进度，直到 status 为 success 或 failed

## 优点

✅ **简单轻量**：不需要数据库表，零迁移成本  
✅ **快速响应**：HTTP请求立即返回，不会超时  
✅ **实时进度**：支持查询导入进度（0-100%）  
✅ **并发安全**：使用sync.Map保证并发访问安全  

## 局限性

⚠️ **内存存储**：服务重启后任务状态会丢失  
⚠️ **单实例**：不支持多实例部署时的任务共享  
⚠️ **无历史**：不保存历史导入记录  

> 如果需要历史记录追溯、多实例部署或持久化任务状态，可以考虑引入Redis或数据库存储。

## 部署步骤

1. **更新代码**
   ```bash
   git pull
   go build
   ```

2. **重启服务**
   ```bash
   systemctl restart foodapp
   ```

3. **前端适配**
   - 修改导入接口调用，处理新的响应格式（202 + task_id）
   - 实现进度查询轮询逻辑（每2秒查询一次）

## 前端示例

```javascript
// 1. 提交导入
async function importFile(orgId, file) {
  const formData = new FormData();
  formData.append('org_id', orgId);
  formData.append('file', file);

  const response = await fetch('/inquiry_import/import_inquiry', {
    method: 'POST',
    body: formData
  });

  if (response.status === 202) {
    const result = await response.json();
    pollProgress(result.task_id);
  }
}

// 2. 轮询进度
async function pollProgress(taskId) {
  const interval = setInterval(async () => {
    const response = await fetch('/inquiry_import/import_status', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ task_id: taskId })
    });

    const result = await response.json();
    console.log(`进度: ${result.progress}%`);

    if (result.status === 'success') {
      clearInterval(interval);
      console.log('导入成功!', result.inquiry_id);
    } else if (result.status === 'failed') {
      clearInterval(interval);
      console.error('导入失败:', result.error_message);
    }
  }, 2000); // 每2秒查询一次
}
```

## 重要Bug修复

### 供应商浮动比例计算错误

**问题**：原代码中浮动比例计算错误，导致结算价严重偏差

**修复前**：
- 下浮12% → ratio = 0.12 → 本期均价100元 × 0.12 = **12元** ❌
- 上浮12% → ratio = -0.12 → 本期均价100元 × -0.12 = **-12元** ❌

**修复后**：
- 下浮12% → ratio = 0.88 → 本期均价100元 × 0.88 = **88元** ✅
- 上浮12% → ratio = 1.12 → 本期均价100元 × 1.12 = **112元** ✅

**影响**：所有通过导入功能创建的供应商结算价都可能有问题，建议检查并重新导入。

## 测试清单

- [ ] 小文件导入（1-2个sheet）
- [ ] 大文件导入（5+个sheet）
- [ ] 进度查询功能
- [ ] 重复导入错误提示
- [ ] 格式错误文件处理
- [ ] 浮动比例计算验证
- [ ] 前端轮询逻辑
- [ ] 服务重启后的任务状态（应该丢失，这是预期行为）

## 后续优化建议

如果需要更强大的功能，可以考虑：

1. **Redis存储**：替换 sync.Map，支持多实例和持久化
2. **任务队列**：使用消息队列（RabbitMQ/Kafka）管理导入任务
3. **WebSocket推送**：主动推送进度，无需前端轮询
4. **历史记录**：将完成的任务存入数据库，支持审计和追溯

但对于当前解决HTTP超时的需求，内存方案已经足够。

