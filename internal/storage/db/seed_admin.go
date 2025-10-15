package foodDB

import (
	"context"
	"errors"

	domain "hdzk.cn/foodapp/internal/domain/account"
	accrepo "hdzk.cn/foodapp/internal/repository/account"
	middleware "hdzk.cn/foodapp/internal/server/middleware"
	"hdzk.cn/foodapp/pkg/crypto"
	"hdzk.cn/foodapp/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	defaultAdminUsername = "admin"
	defaultAdminPassword = "admin123456"
	defaultAdminEmail    = "admin@example.com"
	defaultAdminRole     = middleware.RoleAdmin
)

// EnsureDefaultAdmin 若不存在 admin 用户则创建（密码为默认值，bcrypt 加密）
func EnsureDefaultAdmin(ctx context.Context, gdb *gorm.DB) error {
	repo := accrepo.NewGorm(gdb)

	// 是否已存在
	_, err := repo.GetByUsername(ctx, defaultAdminUsername)
	if err == nil {
		logger.L().Info("default admin exists", zap.String("username", defaultAdminUsername))
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) && err != nil {
		return err
	}

	// 创建
	hash, err := crypto.HashPassword(defaultAdminPassword)
	if err != nil {
		return err
	}
	a := &domain.Account{
		Username:     defaultAdminUsername,
		Email:        defaultAdminEmail,
		PasswordHash: hash,
		Status:       0,
		Role:         int(defaultAdminRole),
	}
	if err := repo.Create(ctx, a); err != nil {
		return err
	}
	logger.L().Info("created default admin user",
		zap.String("username", defaultAdminUsername),
		zap.String("id", a.ID),
	)
	return nil
}
