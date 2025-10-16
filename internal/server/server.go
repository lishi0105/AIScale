package server

import (
	"context"
	"net/http"
	"path/filepath"
	"strings"

	"hdzk.cn/foodapp/configs"
	accrepo "hdzk.cn/foodapp/internal/repository/account"
	dictrepo "hdzk.cn/foodapp/internal/repository/dict"
	acchandler "hdzk.cn/foodapp/internal/server/handler"
	dicthandler "hdzk.cn/foodapp/internal/server/handler"
	"hdzk.cn/foodapp/internal/server/middleware"
	accsvc "hdzk.cn/foodapp/internal/service/account"
	dictsvc "hdzk.cn/foodapp/internal/service/dict"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func registerAccountRoutes(r *gin.Engine, gdb *gorm.DB, authCfg configs.AuthConfig) {
	accService := accsvc.New(accrepo.NewGorm(gdb))
	accH := acchandler.NewAccountHandler(accService)
	authH := acchandler.NewAuthHandler(accService, authCfg.JWTSecret, authCfg.AccessTokenTTLMinute)

	v1 := r.Group("/api/v1")

	// —— 公开路由（登录/刷新等）——
	authH.Register(v1)

	// 查库回调：按 uid 刷新操作者 role/status（停用立刻生效）
	lookup := func(ctx context.Context, uid string) (int, int, error) {
		a, err := accService.GetByID(ctx, uid)
		if err != nil {
			return middleware.RoleUser, middleware.DeletedYes, nil
		}
		return a.Role, a.IsDeleted, nil
	}

	// —— 受保护路由：一次挂载（鉴权 + 停用拦截）——
	protected := v1.Group("/")
	protected.Use(
		middleware.RequireAuth(authCfg.JWTSecret, lookup),
		middleware.ActiveGuard(),
	)
	accH.Register(protected)
}

func registerDictRoutes(r *gin.Engine, gdb *gorm.DB, authCfg configs.AuthConfig) {
	dictRepo := dictrepo.New(gdb)
	dictSvc := dictsvc.New(dictRepo)
	dictH := dicthandler.New(dictSvc)

	v1 := r.Group("/api/v1")
	protected := v1.Group("/")
	protected.Use(
		middleware.RequireAuth(authCfg.JWTSecret, nil), // 字典不强制每次刷新
		middleware.ActiveGuard(),
	)
	dictH.Register(protected)
}

func New(gdb *gorm.DB, authCfg configs.AuthConfig, webDir string) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery(), middleware.RequestID(), middleware.AccessLog())

	// 健康探针
	r.GET("/healthz", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	// 静态资源
	if webDir == "" {
		webDir = "./web"
	}
	r.StaticFile("/", filepath.Join(webDir, "index.html"))
	r.Static("/assets", filepath.Join(webDir, "assets"))
	r.StaticFile("/vite.svg", filepath.Join(webDir, "vite.svg"))

	// SPA 回退
	r.NoRoute(func(c *gin.Context) {
		p := c.Request.URL.Path
		if strings.HasPrefix(p, "/api/") || p == "/healthz" {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.File(filepath.Join(webDir, "index.html"))
	})

	// API
	registerAccountRoutes(r, gdb, authCfg)
	registerDictRoutes(r, gdb, authCfg)

	return r
}
