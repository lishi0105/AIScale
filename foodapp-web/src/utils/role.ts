// src/utils/dict.ts

/**
 * 角色类型定义（与后端保持一致）
 * 0=用户 1=管理员
 */
export const ROLE_USER = 0
export const ROLE_ADMIN = 1

/**
 * 角色标签映射
 */
export const ROLE_LABELS: Record<number, string> = {
  [ROLE_USER]: '用户',
  [ROLE_ADMIN]: '管理员',
}

/**
 * 获取角色显示文本
 */
export function roleLabel(role: number): string {
  return ROLE_LABELS[role] || `未知角色(${role})`
}

// ====== 删除状态定义（与后端保持一致）======

export const DELETED_NO = 0  // 未删除
export const DELETED_YES = 1 // 已删除

export const DELETED_LABELS: Record<number, string> = {
  [DELETED_NO]: '正常',
  [DELETED_YES]: '已删除',
}

export function deletedLabel(isDeleted: number): string {
  return DELETED_LABELS[isDeleted] || `未知状态(${isDeleted})`
}

export function isAdmin(role: number): boolean {
    return role === ROLE_ADMIN
}