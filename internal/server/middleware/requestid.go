package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const HeaderRequestID = "X-Request-ID"

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader(HeaderRequestID)
		if rid == "" {
			rid = uuid.NewString()
		}
		// 对客户端回显，便于串联日志与调用方
		c.Header(HeaderRequestID, rid)
		// 存到上下文，供后续 handler/日志使用
		c.Set("rid", rid)

		c.Next()
	}
}
