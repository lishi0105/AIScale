// src/api/auth.ts
import http, { setAuth } from './http'

// 登录：POST /api/v1/auth/login
export async function login(username: string, password: string) {
  // 也可以直接用完整 URL：'http://172.16.66.35:7380/api/v1/auth/login'
  const { data } = await http.post('/auth/login', { username, password })
  // 期望后端返回：{ expires_in, token, token_type }
  setAuth(data.token, data.expires_in)
  return data
}
