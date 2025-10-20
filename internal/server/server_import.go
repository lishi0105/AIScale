package server

import (
    "hdzk.cn/foodapp/configs"
    handler "hdzk.cn/foodapp/internal/server/handler"
    "hdzk.cn/foodapp/internal/server/middleware"
    importrepo "hdzk.cn/foodapp/internal/repository/inquiry"
    importsvc "hdzk.cn/foodapp/internal/service/inquiry"
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)

func registerInquiryImportRoutes(r *gin.Engine, gdb *gorm.DB, authCfg configs.AuthConfig) {
    repo := importrepo.NewImportRepository(gdb)
    svc := importsvc.NewImportService(repo)
    h := handler.NewInquiryImportHandler(svc)

    v1 := r.Group("/api/v1")
    protected := v1.Group("/")
    protected.Use(
        middleware.RequireAuth(authCfg.JWTSecret, nil),
        middleware.ActiveGuard(),
    )
    h.Register(protected)
}
