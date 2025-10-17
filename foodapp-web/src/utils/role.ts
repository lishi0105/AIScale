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

export const DELETED_NO = 0
export const DELETED_YES = 1

export const DELETED_LABELS: Record<number, string> = {
  [DELETED_NO]: '启用',
  [DELETED_YES]: '停用',
}

export function DeletedLabel(deleted: number): string {
  return DELETED_LABELS[deleted] || `未知状态(${deleted})`
}

export function isAdmin(role: number): boolean {
    return role === ROLE_ADMIN
}