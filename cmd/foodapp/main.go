package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"hdzk.cn/foodapp/configs"
	"hdzk.cn/foodapp/internal/server"
	foodDB "hdzk.cn/foodapp/internal/storage/db"
	"hdzk.cn/foodapp/pkg/logger"
)

func main() {
	// 1. 加载配置文件
	cfg, created, err := configs.LoadConfig("configs/config.json")
	if err != nil {
		fmt.Println("加载配置错误:", err)
	}
	if created {
		fmt.Println("首次运行：已写入默认配置")
	}
	// 2. 初始化日志
	log := logger.Init(cfg.Log)
	defer log.Sync()

	log.Info("✅ 日志系统初始化完成",
		zap.String("level", cfg.Log.Level),
		zap.String("path", cfg.Log.Dir),
	)

	// 3. HTTP 服务配置
	log.Info("🌐 启动 HTTP 服务",
		zap.Int("port", cfg.Server.Port),
	)
	food_db, err := foodDB.OpenFromConfig(cfg.DB)
	if err != nil {
		log.Fatal("✅ 无法连接到数据库",
			zap.Error(err),
		)
	}
	log.Info("✅ 数据库连接成功")

	defer func() {
		if cerr := foodDB.Close(food_db); cerr != nil {
			logger.L().Warn("close db failed", zap.Error(cerr))
		}
	}()

	// 3.1 迁移（可选：如果你已实现）
	if err := foodDB.AutoMigrate(food_db); err != nil {
		log.Fatal("auto migrate failed", zap.Error(err))
	}

	// 3.2 确保默认管理员（admin/admin123456）
	if err := foodDB.EnsureDefaultAccount(context.Background(), food_db); err != nil {
		log.Fatal("ensure default admin failed", zap.Error(err))
	}

	if err := foodDB.EnsureDefaultDicts(context.Background(), food_db); err != nil {
		log.Fatal("seed dicts failed", zap.Error(err))
	}

	if err := foodDB.EnsureDefaultOrganization(context.Background(), food_db); err != nil {
		log.Fatal("ensure default org failed", zap.Error(err))
	}

	engine := server.New(food_db, cfg.Auth, cfg.Server.WebRoot)
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	srv := &http.Server{
		Addr:              addr,
		Handler:           engine,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// 5) 启动 + 优雅关停
	go func() {
		log.Info("http server starting", zap.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("http server error", zap.Error(err))
		}
	}()

	// 等待信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("shutting down ...")

	// 给一些时间优雅关停
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Warn("http shutdown forced", zap.Error(err))
	} else {
		log.Info("http server stopped")
	}
}
