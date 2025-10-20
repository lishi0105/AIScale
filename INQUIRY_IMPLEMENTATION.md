# 询价记录 CRUD 操作实现总结

## 概述

已为 `base_price_inquiry` 表实现了完整的 CRUD 操作，遵循项目现有的代码架构模式。

## 实现的文件

### 1. Domain Model
- **文件**: `internal/domain/inquiry/model.go`
- **功能**: 定义询价记录的数据模型
- **特性**: 
  - 自动生成 UUID
  - BeforeCreate 钩子验证必填字段和时间范围

### 2. Repository Layer
- **接口**: `internal/repository/inquiry/repo.go`
- **实现**: `internal/repository/inquiry/repo_gorm.go`
- **功能**: 
  - CreateInquiry - 创建询价记录
  - GetInquiry - 获取单个询价记录
  - ListInquiries - 分页查询询价记录（支持关键词、日期范围过滤）
  - UpdateInquiry - 更新询价记录（支持部分字段更新）
  - SoftDeleteInquiry - 软删除
  - HardDeleteInquiry - 硬删除

### 3. Service Layer
- **文件**: `internal/service/inquiry/service.go`
- **功能**: 业务逻辑层，包含参数验证和数据规范化
- **特性**:
  - 字符串字段自动去空格
  - 必填字段验证
  - 时间范围验证

### 4. HTTP Handler
- **文件**: `internal/server/handler/inquiry.go`
- **功能**: HTTP 路由处理
- **端点**:
  - POST `/api/v1/inquiries/create_inquiry` - 创建
  - POST `/api/v1/inquiries/get_inquiry` - 获取单个
  - GET `/api/v1/inquiries/list_inquiries` - 列表查询
  - POST `/api/v1/inquiries/update_inquiry` - 更新
  - POST `/api/v1/inquiries/soft_delete_inquiry` - 软删除
  - POST `/api/v1/inquiries/hard_delete_inquiry` - 硬删除

### 5. 路由注册
- **文件**: `internal/server/server.go`
- **修改**: 
  - 添加了 inquiry 相关的导入
  - 新增 `registerInquiryRoutes` 函数
  - 在 `New` 函数中注册路由

### 6. SQL 修复
- **文件**: `sql/10_goods_domain.sql`
- **修改**: 修复外键引用从 `price_inquiry` 到 `base_price_inquiry`

## 数据库特性

### 表结构亮点

1. **软删除支持**: 使用 `is_deleted` 字段
2. **计算列**: `active_title` 仅对未删除记录生效
3. **唯一性约束**: `uk_org_active_title_date` 确保同组织下未删除记录的标题+日期唯一
4. **时间约束**: CHECK 约束确保结束时间晚于开始时间
5. **索引优化**: 
   - 组织+有效状态+日期的复合索引
   - 组织+标题的复合索引

## 权限控制

- **创建/更新/删除**: 仅管理员可操作
- **查询**: 所有已认证用户可操作
- **自动检查**: 账户是否已删除

## API 特性

### 列表查询支持
- 按组织ID过滤（必填）
- 按标题关键词模糊搜索
- 按日期范围过滤（可选）
- 分页支持（默认20条/页，最大1000条）
- 按日期降序排序

### 时间格式
- `inquiry_date`: YYYY-MM-DD（如: 2024-10-20）
- `inquiry_start_date` 和 `inquiry_end_date`: RFC3339（如: 2024-10-20T08:00:00Z）

### 更新特性
- 支持部分字段更新
- 可将可选字段设为 null 清空数据
- 自动验证时间范围约束

## 编译验证

✅ 代码已通过 Go 编译验证，无语法错误

## 文档

详细 API 文档请查看: `docs/inquiry_api.md`

## 与现有代码的一致性

本实现完全遵循项目现有模式：
- ✅ 使用相同的目录结构
- ✅ 遵循相同的命名规范
- ✅ 使用相同的错误处理方式
- ✅ 实现相同的权限控制逻辑
- ✅ 使用相同的响应格式

## 下一步

如需进一步功能：
1. 添加询价记录的批量操作
2. 添加询价记录的导出功能
3. 添加询价记录的统计报表
4. 为前端添加 TypeScript API 客户端
