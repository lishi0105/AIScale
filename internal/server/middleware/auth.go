package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	// context keys
	ContextUserIDKey   = "uid"
	ContextUsernameKey = "usr"
	ContextRoleKey     = "role"
	ContextStatusKey   = "status"

	// roles
	RoleAdmin = 0
	RoleUser  = 1

	// status（与数据库 account.status 保持一致：0=禁用 1=启用）
	StatusDisabled = 0
	StatusEnabled  = 1
)

// Actor 注入到 gin.Context 的登录用户快照
type Actor struct {
	ID       string
	Username string
	Role     int
	Status   int
}

// AccountLookup 按 uid 实时查询 role/status（可选，传 nil 则不查库）
type AccountLookup func(ctx context.Context, uid string) (role int, status int, err error)

// RequireAuth 校验 Bearer JWT，注入 uid/usr/role/status/actor；可选查库刷新 role/status
func RequireAuth(secret string, lookup AccountLookup) gin.HandlerFunc {
	sec := []byte(secret)

	return func(c *gin.Context) {
		ah := c.GetHeader("Authorization")
		if !strings.HasPrefix(strings.ToLower(ah), "bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "请登录后操作"})
			return
		}
		tokenString := strings.TrimSpace(ah[len("Bearer "):])

		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrTokenSignatureInvalid
			}
			return sec, nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "非法token"})
			return
		}

		// 缺省值
		var (
			uid, username string
			role          = RoleUser
			status        = StatusEnabled
		)

		// 从 claims 读 sub/usr/role/status
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if v, ok := claims["sub"].(string); ok {
				uid = v
			}
			if v, ok := claims["usr"].(string); ok {
				username = v
			}
			if v, ok := claims["role"]; ok {
				switch vv := v.(type) {
				case float64:
					role = int(vv)
				case int:
					role = vv
				}
			}
			if v, ok := claims["status"]; ok {
				switch vv := v.(type) {
				case float64:
					status = int(vv)
				case int:
					status = vv
				}
			}
		}

		// 可选：每次请求实时刷新 role/status，停用立刻生效
		if lookup != nil && uid != "" {
			if r, s, err := lookup(c.Request.Context(), uid); err == nil {
				role, status = r, s
			}
		}

		// 注入上下文
		if uid != "" {
			c.Set(ContextUserIDKey, uid)
		}
		if username != "" {
			c.Set(ContextUsernameKey, username)
		}
		c.Set(ContextRoleKey, role)
		c.Set(ContextStatusKey, status)
		c.Set("actor", &Actor{ID: uid, Username: username, Role: role, Status: status})

		c.Next()
	}
}

// ActiveGuard 停用账户一律拦截（需放在 RequireAuth 之后）
func ActiveGuard() gin.HandlerFunc {
	return func(c *gin.Context) {
		if v, ok := c.Get(ContextStatusKey); ok {
			if st, ok := v.(int); ok && st == StatusDisabled {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"error":   "禁止操作",
					"details": "账户已停用",
				})
				return
			}
		}
		c.Next()
	}
}

// —— 便捷取值（可给 handler 使用）——

func GetActor(c *gin.Context) *Actor {
	a := &Actor{}
	if v, ok := c.Get(ContextUserIDKey); ok {
		if s, ok := v.(string); ok {
			a.ID = s
		}
	}
	if v, ok := c.Get(ContextUsernameKey); ok {
		if s, ok := v.(string); ok {
			a.Username = s
		}
	}
	if v, ok := c.Get(ContextRoleKey); ok {
		if i, ok := v.(int); ok {
			a.Role = i
		}
	}
	if v, ok := c.Get(ContextStatusKey); ok {
		if i, ok := v.(int); ok {
			a.Status = i
		}
	}
	return a
}

func IsAdmin(c *gin.Context) bool {
	if v, ok := c.Get(ContextRoleKey); ok {
		if i, ok := v.(int); ok {
			return i == RoleAdmin
		}
	}
	return false
}
