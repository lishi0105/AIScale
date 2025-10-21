package server

import (
	"context"
	"net/http"
	"path/filepath"
	"strings"

	"hdzk.cn/foodapp/configs"
	accrepo "hdzk.cn/foodapp/internal/repository/account"
	categoryrepo "hdzk.cn/foodapp/internal/repository/category"
	dictrepo "hdzk.cn/foodapp/internal/repository/dict"
	goodsrepo "hdzk.cn/foodapp/internal/repository/goods"
	inquiryitemrepo "hdzk.cn/foodapp/internal/repository/inquiry_item"
	marketrepo "hdzk.cn/foodapp/internal/repository/market"
	marketinquiryrepo "hdzk.cn/foodapp/internal/repository/market_inquiry"
	organrepo "hdzk.cn/foodapp/internal/repository/organ"
	priceinquiryrepo "hdzk.cn/foodapp/internal/repository/price_inquiry"
	supplierrepo "hdzk.cn/foodapp/internal/repository/supplier"
	suppliersettlementrepo "hdzk.cn/foodapp/internal/repository/supplier_settlement"
	handler "hdzk.cn/foodapp/internal/server/handler"
	"hdzk.cn/foodapp/internal/server/middleware"
	accsvc "hdzk.cn/foodapp/internal/service/account"
	categorysvc "hdzk.cn/foodapp/internal/service/category"
	dictsvc "hdzk.cn/foodapp/internal/service/dict"
	goodssvc "hdzk.cn/foodapp/internal/service/goods"
	inquiryitemsvc "hdzk.cn/foodapp/internal/service/inquiry_item"
	marketsvc "hdzk.cn/foodapp/internal/service/market"
	marketinquirysvc "hdzk.cn/foodapp/internal/service/market_inquiry"
	organsvc "hdzk.cn/foodapp/internal/service/organ"
	priceinquirysvc "hdzk.cn/foodapp/internal/service/price_inquiry"
	suppliersvc "hdzk.cn/foodapp/internal/service/supplier"
	suppliersettlementsvc "hdzk.cn/foodapp/internal/service/supplier_settlement"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func registerAccountRoutes(r *gin.Engine, gdb *gorm.DB, authCfg configs.AuthConfig) {
	accService := accsvc.NewService(accrepo.NewRepository(gdb))
	accH := handler.NewAccountHandler(accService)
	authH := handler.NewAuthHandler(accService, authCfg.JWTSecret, authCfg.AccessTokenTTLMinute)

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
	dictRepo := dictrepo.NewRepository(gdb)
	dictSvc := dictsvc.NewService(dictRepo)
	dictH := handler.NewDictHandler(dictSvc)

	v1 := r.Group("/api/v1")
	protected := v1.Group("/")
	protected.Use(
		middleware.RequireAuth(authCfg.JWTSecret, nil), // 字典不强制每次刷新
		middleware.ActiveGuard(),
	)
	dictH.Register(protected)
}

func registerOrganRoutes(r *gin.Engine, gdb *gorm.DB, authCfg configs.AuthConfig) {
	organRepo := organrepo.NewRepository(gdb)
	organSvc := organsvc.NewService(organRepo)
	organH := handler.NewOrganHandler(organSvc)
	v1 := r.Group("/api/v1")
	protected := v1.Group("/")
	protected.Use(
		middleware.RequireAuth(authCfg.JWTSecret, nil), // 字典不强制每次刷新
		middleware.ActiveGuard(),
	)
	organH.Register(protected)
}

func registerCategoryRoutes(r *gin.Engine, gdb *gorm.DB, authCfg configs.AuthConfig) {
	categoryRepo := categoryrepo.NewRepository(gdb)
	categorySvc := categorysvc.NewService(categoryRepo)
	categoryH := handler.NewCategoryHandler(categorySvc)
	v1 := r.Group("/api/v1")
	protected := v1.Group("/")
	protected.Use(
		middleware.RequireAuth(authCfg.JWTSecret, nil), // 品类不强制每次刷新
		middleware.ActiveGuard(),
	)
	categoryH.Register(protected)
}

func registerGoodsRoutes(r *gin.Engine, gdb *gorm.DB, authCfg configs.AuthConfig) {
	goodsRepo := goodsrepo.NewRepository(gdb)
	goodsSvc := goodssvc.NewService(goodsRepo)
	goodsH := handler.NewGoodsHandler(goodsSvc)

	v1 := r.Group("/api/v1")
	protected := v1.Group("/")
	protected.Use(
		middleware.RequireAuth(authCfg.JWTSecret, nil), // 商品不强制每次刷新
		middleware.ActiveGuard(),
	)
	goodsH.Register(protected)
}

func registerSupplierRoutes(r *gin.Engine, gdb *gorm.DB, authCfg configs.AuthConfig) {
	supplierRepo := supplierrepo.NewRepository(gdb)
	supplierSvc := suppliersvc.NewService(supplierRepo)
	supplierH := handler.NewSupplierHandler(supplierSvc)
	v1 := r.Group("/api/v1")
	protected := v1.Group("/")
	protected.Use(
		middleware.RequireAuth(authCfg.JWTSecret, nil), // 供应商不强制每次刷新
		middleware.ActiveGuard(),
	)
	supplierH.Register(protected)
}

func registerMarketRoutes(r *gin.Engine, gdb *gorm.DB, authCfg configs.AuthConfig) {
	marketRepo := marketrepo.NewRepository(gdb)
	marketSvc := marketsvc.NewService(marketRepo)
	marketH := handler.NewMarketHandler(marketSvc)
	v1 := r.Group("/api/v1")
	protected := v1.Group("/")
	protected.Use(
		middleware.RequireAuth(authCfg.JWTSecret, nil),
		middleware.ActiveGuard(),
	)
	marketH.Register(protected)
}

func registerPriceInquiryRoutes(r *gin.Engine, gdb *gorm.DB, authCfg configs.AuthConfig) {
	priceInquiryRepo := priceinquiryrepo.NewRepository(gdb)
	priceInquirySvc := priceinquirysvc.NewService(priceInquiryRepo)
	priceInquiryH := handler.NewPriceInquiryHandler(priceInquirySvc)
	v1 := r.Group("/api/v1")
	protected := v1.Group("/")
	protected.Use(
		middleware.RequireAuth(authCfg.JWTSecret, nil),
		middleware.ActiveGuard(),
	)
	priceInquiryH.Register(protected)
}

func registerInquiryItemRoutes(r *gin.Engine, gdb *gorm.DB, authCfg configs.AuthConfig) {
	inquiryItemRepo := inquiryitemrepo.NewRepository(gdb)
	inquiryItemSvc := inquiryitemsvc.NewService(inquiryItemRepo)
	inquiryItemH := handler.NewInquiryItemHandler(inquiryItemSvc)
	v1 := r.Group("/api/v1")
	protected := v1.Group("/")
	protected.Use(
		middleware.RequireAuth(authCfg.JWTSecret, nil),
		middleware.ActiveGuard(),
	)
	inquiryItemH.Register(protected)
}

func registerMarketInquiryRoutes(r *gin.Engine, gdb *gorm.DB, authCfg configs.AuthConfig) {
	marketInquiryRepo := marketinquiryrepo.NewRepository(gdb)
	marketInquirySvc := marketinquirysvc.NewService(marketInquiryRepo)
	marketInquiryH := handler.NewMarketInquiryHandler(marketInquirySvc)
	v1 := r.Group("/api/v1")
	protected := v1.Group("/")
	protected.Use(
		middleware.RequireAuth(authCfg.JWTSecret, nil),
		middleware.ActiveGuard(),
	)
	marketInquiryH.Register(protected)
}

func registerSupplierSettlementRoutes(r *gin.Engine, gdb *gorm.DB, authCfg configs.AuthConfig) {
	supplierSettlementRepo := suppliersettlementrepo.NewRepository(gdb)
	supplierSettlementSvc := suppliersettlementsvc.NewService(supplierSettlementRepo)
	supplierSettlementH := handler.NewSupplierSettlementHandler(supplierSettlementSvc)
	v1 := r.Group("/api/v1")
	protected := v1.Group("/")
	protected.Use(
		middleware.RequireAuth(authCfg.JWTSecret, nil),
		middleware.ActiveGuard(),
	)
	supplierSettlementH.Register(protected)
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
	registerOrganRoutes(r, gdb, authCfg)
	registerCategoryRoutes(r, gdb, authCfg)
	registerSupplierRoutes(r, gdb, authCfg)
	registerGoodsRoutes(r, gdb, authCfg)
	registerMarketRoutes(r, gdb, authCfg)
	registerPriceInquiryRoutes(r, gdb, authCfg)
	registerInquiryItemRoutes(r, gdb, authCfg)
	registerMarketInquiryRoutes(r, gdb, authCfg)
	registerSupplierSettlementRoutes(r, gdb, authCfg)

	return r
}
