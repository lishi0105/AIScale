// storage/db/seed_account.go
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
	defaultAdminPassword = "admin123"
	defaultAdminRole     = middleware.RoleAdmin
	defaultAdminSort     = -1

	// 默认组织固定 UUID（你刚刚要的那一个）
	defaultOrgID = DefaultOrgID
)

// EnsureDefaultAdminWithOrg 若 admin 不存在则创建并绑定默认组织；
// 若已存在：不重置密码，仅在 OrgID 为空时补齐，并确保角色=管理员、未删除。
func EnsureDefaultAccount(ctx context.Context, gdb *gorm.DB) error {
	repo := accrepo.NewRepository(gdb) // 若你的构造函数是 NewRepository，请替换

	// 查询是否已存在
	existing, err := repo.GetByUsername(ctx, defaultAdminUsername)
	switch {
	case err == nil && existing != nil:
		// 已存在：仅修正必要字段（不动密码）
		patch := map[string]any{}
		if existing.OrgID == "" {
			patch["org_id"] = defaultOrgID
		}
		if existing.Role != int(defaultAdminRole) {
			patch["role"] = int(defaultAdminRole)
		}
		if existing.IsDeleted != middleware.DeletedNo {
			patch["is_deleted"] = middleware.DeletedNo
		}
		if len(patch) == 0 {
			logger.L().Info("default admin exists and healthy",
				zap.String("username", defaultAdminUsername),
				zap.String("id", existing.ID),
				zap.String("org_id", existing.OrgID),
			)
			return nil
		}
		if err := repo.UpdateFields(ctx, existing.ID, patch); err != nil {
			return err
		}
		logger.L().Info("patched existing default admin",
			zap.String("username", defaultAdminUsername),
			zap.String("id", existing.ID),
			zap.Any("patch", patch),
		)
		return nil

	case errors.Is(err, gorm.ErrRecordNotFound):
		// 不存在：创建
		hash, herr := crypto.HashPassword(defaultAdminPassword)
		if herr != nil {
			return herr
		}
		a := &domain.Account{
			Username:     defaultAdminUsername,
			PasswordHash: hash,
			OrgID:        defaultOrgID, // 绑定默认组织
			Sort:         defaultAdminSort,
			IsDeleted:    middleware.DeletedNo,
			Role:         int(defaultAdminRole),
		}
		logger.L().Info("created default admin user",
			zap.String("username", defaultAdminUsername),
			zap.String("id", a.ID),
			zap.String("org_id", a.OrgID),
			zap.Int("role", a.Role),
		)
		if cerr := repo.Create(ctx, a); cerr != nil {
			return cerr
		}
		return nil

	default:
		// 其他错误
		return err
	}
}
