// src/utils/jwt.ts

export interface JwtPayload {
  sub: string      // user ID
  usr: string      // username
  role: number     // role
  del?: number | boolean // deletion flag (0/1 or boolean)
  organ_id?: string // organization / organ identifier
  iat: number      // issued at
  exp: number      // expire at
  iss: string      // issuer
}

/**
 * 从 Bearer Token 中解析 JWT payload
 * @param token Bearer xxx 或 xxx
 * @returns 解析后的 payload，失败返回 null
 */
export function parseJwt(token: string): JwtPayload | null {
  try {
    // 去掉 "Bearer " 前缀（如果存在）
    const cleanToken = token.replace(/^Bearer\s+/i, '')
    const payloadBase64 = cleanToken.split('.')[1]
    if (!payloadBase64) return null

    // base64url -> base64
    const base64 = payloadBase64.replace(/-/g, '+').replace(/_/g, '/')
    const jsonPayload = decodeURIComponent(
      atob(base64)
        .split('')
        .map(c => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2))
        .join('')
    )
    return JSON.parse(jsonPayload)
  } catch (e) {
    console.warn('Failed to parse JWT:', e)
    return null
  }
}