// src/utils/dict.ts

/**
 * 角色类型定义（与后端保持一致）
 */
export const ROLE_USER = 0
export const ROLE_ADMIN = 1

/**
 * 角色标签映射
 */
export const ROLE_LABELS: Record<number, string> = {
  [ROLE_ADMIN]: '管理员',
  [ROLE_USER]: '用户',
}

/**
 * 获取角色显示文本
 */
export function roleLabel(role: number): string {
  return ROLE_LABELS[role] || `未知角色(${role})`
}

// ====== 其他字典（可选扩展）======

// 删除状态（IsDeleted）
export const DELETED_NO = 0
export const DELETED_YES = 1

export const DELETED_LABELS: Record<number, string> = {
  [DELETED_NO]: '正常',
  [DELETED_YES]: '已删除',
}

export function deletedLabel(isDeleted: number): string {
  return DELETED_LABELS[isDeleted] || `未知(${isDeleted})`
}

// 保留旧的状态定义以保持向后兼容（如果其他地方还在使用）
export const STATUS_DISABLED = 0
export const STATUS_ENABLED = 1

export const STATUS_LABELS: Record<number, string> = {
  [STATUS_ENABLED]: '启用',
  [STATUS_DISABLED]: '停用',
}

export function statusLabel(status: number): string {
  return STATUS_LABELS[status] || `未知状态(${status})`
}

export function isAdmin(role: number): boolean {
    return role === ROLE_ADMIN
}