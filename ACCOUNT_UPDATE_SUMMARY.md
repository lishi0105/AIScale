# 账户管理前端适配总结

## 概述
根据后端 Account 模型和 API 的变更，已完成前端代码的全面更新，确保前后端接口对接正确。

## 后端变更
### Account Model 变化
- **删除字段**: `Status` (启用/停用状态)
- **新增字段**:
  - `OrgID`: 所属机构ID (必填，UUID)
  - `Description`: 描述信息 (可选，text)
  - `IsDeleted`: 删除标记 (0=未删除, 1=已删除)

### API 接口变化
1. **POST /accounts/create**
   - 新增必填: `org_id` (UUID)
   - 新增可选: `description` (string)
   - 删除: `status` 参数

2. **POST /accounts/list**
   - 新增筛选: `is_deleted` (0/1)
   - 新增筛选: `role` (0/1)
   - 删除: `status` 参数

3. **POST /accounts/update**
   - 替代原 `update_status` 接口
   - 支持更新: `username`, `org_id`, `description`, `role`

## 前端修改详情

### 1. API 层修改 (`foodapp-web/src/api/acl.ts`)
```typescript
// 更新接口定义
export const AccountAPI = {
  // list: 添加 is_deleted 和 role 筛选参数
  list: (params: { 
    username_like?: string; 
    is_deleted?: number; 
    role?: number; 
    limit?: number; 
    offset?: number 
  }) => http.post('/accounts/list', params || {}),

  // create: org_id 必填，添加 description
  create: (data: { 
    username: string; 
    password: string; 
    org_id: string; 
    role?: number; 
    description?: string 
  }) => http.post('/accounts/create', data),

  // update: 新增，替代 update_status
  update: (data: { 
    id: string; 
    username?: string; 
    org_id?: string; 
    description?: string; 
    role?: number 
  }) => http.post('/accounts/update', data),

  // 删除了 update_status 接口
}
```

### 2. 工具类更新 (`foodapp-web/src/utils/role.ts`)
```typescript
// 新增删除状态相关常量和函数
export const DELETED_NO = 0
export const DELETED_YES = 1

export const DELETED_LABELS: Record<number, string> = {
  [DELETED_NO]: '正常',
  [DELETED_YES]: '已删除',
}

export function deletedLabel(isDeleted: number): string {
  return DELETED_LABELS[isDeleted] || `未知(${isDeleted})`
}

// 保留了旧的 STATUS 相关定义以保持向后兼容
```

### 3. 页面组件更新 (`foodapp-web/src/pages/Accounts.vue`)

#### 3.1 类型定义更新
```typescript
interface Row {
  ID: string
  Username: string
  OrgID: string          // 新增：所属机构ID
  Description?: string   // 新增：描述
  Role: number           // 角色
  IsDeleted: number      // 新增：删除标记
  LastLoginAt?: string | null
  CreatedAt?: string
  UpdatedAt?: string
}
```

#### 3.2 表格列更新
- **删除**: "状态" 列 (启用/停用)
- **新增**:
  - "所属机构" 列：显示机构名称（自动映射 OrgID）
  - "描述" 列：显示账户描述
  - "删除状态" 列：显示正常/已删除状态

#### 3.3 表单字段更新
**新增账户表单**:
- 用户名 (必填，创建后不可修改)
- 密码 (必填，创建时)
- 确认密码 (必填，创建时)
- **所属机构** (必填，下拉选择)：从机构列表中选择
- **描述** (可选，多行文本)
- 角色 (下拉选择)

**编辑账户表单**:
- 用户名 (显示，不可修改)
- **所属机构** (可修改，下拉选择)
- **描述** (可修改，多行文本)
- 角色 (可修改，仅管理员)

#### 3.4 新增功能
1. **机构选择器**
   - 自动从后端加载机构列表
   - 支持搜索和筛选
   - 表格中显示机构名称而非ID

2. **高级筛选**
   - 按角色筛选 (管理员/用户)
   - 按删除状态筛选 (正常/已删除)
   - 按用户名搜索

3. **表单验证增强**
   - 创建账户时验证机构ID必填
   - 更新账户时支持部分字段更新

#### 3.5 提交逻辑更新
```typescript
// 创建账户
await AccountAPI.create({
  username: Username.trim(),
  password,
  org_id: OrgID,  // 必填
  role: Number(form.value.Role ?? ROLE_USER),
  description: Description || undefined
})

// 更新账户
await AccountAPI.update({
  id,
  username: Username?.trim() || undefined,
  org_id: OrgID || undefined,
  description: Description || undefined,
  role: Role !== undefined ? Number(Role) : undefined
})
```

## 功能对比

| 功能 | 修改前 | 修改后 |
|------|--------|--------|
| 账户状态 | 启用/停用 | 删除/未删除 |
| 机构关联 | 无 | 必须关联机构 (OrgID) |
| 账户描述 | 无 | 支持描述信息 |
| 机构选择 | - | 下拉选择，显示名称 |
| 列表筛选 | 用户名 | 用户名 + 角色 + 删除状态 |
| 更新接口 | update_status | update (更全面) |

## 向后兼容性
- 保留了 `STATUS_*` 相关常量定义，以防其他模块仍在使用
- 所有修改都向后兼容，不影响其他功能模块

## 测试建议
1. **创建账户测试**
   - 验证机构ID必填
   - 验证描述字段可选
   - 验证机构选择器正常工作

2. **更新账户测试**
   - 验证可以修改机构
   - 验证可以修改描述
   - 验证非管理员不能修改角色

3. **列表筛选测试**
   - 验证角色筛选功能
   - 验证删除状态筛选功能
   - 验证组合筛选功能

4. **显示测试**
   - 验证表格中正确显示机构名称
   - 验证删除状态标签显示正确

## 注意事项
1. 创建账户时必须选择机构，确保数据库中有可用的机构数据
2. 删除账户是软删除，不会真正从数据库删除
3. 管理员账户不允许删除（后端限制）
4. 用户只能修改自己的部分信息，管理员可以修改所有用户信息

## 文件清单
修改的文件：
- `foodapp-web/src/api/acl.ts` - API 接口定义
- `foodapp-web/src/utils/role.ts` - 工具类和常量定义
- `foodapp-web/src/pages/Accounts.vue` - 账户管理页面组件

依赖的文件（未修改）：
- `foodapp-web/src/api/organ.ts` - 机构 API（用于加载机构列表）
- `foodapp-web/src/api/http.ts` - HTTP 客户端
- `foodapp-web/src/utils/jwt.ts` - JWT 解析工具

## 总结
本次更新完整适配了后端 Account 模型和 API 的变更，增强了账户管理功能，提升了用户体验。主要改进包括：
- 支持机构关联和管理
- 增强的筛选和搜索功能
- 更友好的表单交互（机构选择器、描述字段）
- 符合后端新接口规范
