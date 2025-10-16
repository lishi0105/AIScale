package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

/************* 常量 & 上下文键 *************/
const (
	// gin.Context keys
	ContextUserIDKey   = "uid"
	ContextUsernameKey = "usr"
	ContextRoleKey     = "role"
	ContextDeletedKey  = "deleted"

	// roles（与 account.Account.Role 一致：0=用户 1=管理员）
	RoleUser  = 0
	RoleAdmin = 1

	// deleted（与 account.Account.IsDeleted 一致：0=未删除 1=已删除/停用）
	DeletedNo  = 0
	DeletedYes = 1
)

/************* 数据结构 *************/
type Actor struct {
	ID       string
	Username string
	Role     int
	Deleted  int
}

// AccountLookup：按 uid 实时查询 role / deleted（可选，传 nil 则不查库）
type AccountLookup func(ctx context.Context, uid string) (role int, deleted int, err error)

/************* 中间件 *************/

// RequireAuth：校验 Bearer JWT，注入 uid/usr/role/deleted/actor；可选查库刷新 role/deleted
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
			deleted       = DeletedNo
		)

		// 从 claims 读 sub/usr/role/deleted（兼容旧 status）
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
			// 新字段：deleted（0/1）
			if v, ok := claims["deleted"]; ok {
				switch vv := v.(type) {
				case float64:
					deleted = int(vv)
				case int:
					deleted = vv
				}
			} else if v, ok := claims["status"]; ok {
				// 兼容旧 token：status(0=禁用,1=启用) -> deleted(1/0)
				switch vv := v.(type) {
				case float64:
					if int(vv) == 0 {
						deleted = DeletedYes
					} else {
						deleted = DeletedNo
					}
				case int:
					if vv == 0 {
						deleted = DeletedYes
					} else {
						deleted = DeletedNo
					}
				}
			}
		}

		// 可选：每次请求到库里刷新 role/deleted（即刻生效）
		if lookup != nil && uid != "" {
			if r, d, err := lookup(c.Request.Context(), uid); err == nil {
				role, deleted = r, d
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
		c.Set(ContextDeletedKey, deleted)
		c.Set("actor", &Actor{ID: uid, Username: username, Role: role, Deleted: deleted})

		c.Next()
	}
}

// ActiveGuard：若账户已删除/停用则拦截（需放在 RequireAuth 之后）
func ActiveGuard() gin.HandlerFunc {
	return func(c *gin.Context) {
		if v, ok := c.Get(ContextDeletedKey); ok {
			if del, ok := v.(int); ok && del == DeletedYes {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"error":   "禁止操作",
					"details": "账户已停用或删除",
				})
				return
			}
		}
		c.Next()
	}
}

/************* 便捷函数 *************/

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
	if v, ok := c.Get(ContextDeletedKey); ok {
		if i, ok := v.(int); ok {
			a.Deleted = i
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
