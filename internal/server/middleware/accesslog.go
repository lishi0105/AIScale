package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"hdzk.cn/foodapp/pkg/logger"
)

func AccessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 让下一个中间件/handler 继续执行
		c.Next()

		// 结束后记录一次访问日志
		logger.L().Info("access",
			zap.String("rid", c.GetString("rid")),
			zap.String("method", c.Request.Method),
			zap.String("path", c.FullPath()), // 路由模板，如 /api/v1/accounts/create
			zap.Int("status", c.Writer.Status()),
			zap.String("client_ip", c.ClientIP()),
			zap.String("ua", c.Request.UserAgent()),
			zap.Duration("cost", time.Since(start)),
		)

		// 如果你想把业务日志里的 request-id 自动带上，也可以考虑在这里：
		// logger.SetRequestIDToContext(c) → 然后 L().With(zap.String("rid", ...))
	}
}
