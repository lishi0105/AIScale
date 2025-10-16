// internal/service/account/service.go
package account

import (
	"context"
	"errors"

	"gorm.io/gorm"
	domain "hdzk.cn/foodapp/internal/domain/account"
	repo "hdzk.cn/foodapp/internal/repository/account"
	middleware "hdzk.cn/foodapp/internal/server/middleware"
	"hdzk.cn/foodapp/pkg/crypto"
)

type Service struct{ r repo.Repository }

func NewService(r repo.Repository) *Service { return &Service{r: r} }

func (s *Service) Authenticate(ctx context.Context, username, plain string) (string, int, error) {
	u, err := s.r.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", int(middleware.RoleUser), errors.New("用户名或密码错误")
		}
		return "", int(middleware.RoleUser), err
	}
	if !crypto.VerifyPassword(u.PasswordHash, plain) {
		return "", int(middleware.RoleUser), errors.New("用户名或密码错误")
	}
	if u.IsDeleted == 1 {
		return "", int(u.Role), errors.New("账户已停用")
	}
	return u.ID, int(u.Role), nil
}

func (s *Service) Create(ctx context.Context, a *domain.Account) error {
	// 这里默认 handler 已经校验了必填项并完成了密码哈希
	if a.Username == "" || a.PasswordHash == "" || a.OrgID == "" {
		return errors.New("username/password/org_id 为必填")
	}
	return s.r.Create(ctx, a)
}

func (s *Service) GetByID(ctx context.Context, id string) (*domain.Account, error) {
	return s.r.GetByID(ctx, id)
}
func (s *Service) GetByUsername(ctx context.Context, username string) (*domain.Account, error) {
	return s.r.GetByUsername(ctx, username)
}
func (s *Service) List(ctx context.Context, q domain.ListQuery) ([]domain.Account, int64, error) {
	return s.r.List(ctx, q)
}
func (s *Service) UpdatePasswordHash(ctx context.Context, id string, hash string) error {
	return s.r.UpdatePasswordHash(ctx, id, hash)
}
func (s *Service) SoftDelete(ctx context.Context, id string) error { return s.r.SoftDelete(ctx, id) }
func (s *Service) HardDelete(ctx context.Context, id string) error { return s.r.HardDelete(ctx, id) }

func (s *Service) ChangePassword(ctx context.Context, username, oldPlain, newPlain string) error {
	u, err := s.r.GetByUsername(ctx, username)
	if err != nil {
		return err
	}
	if u.IsDeleted == 1 {
		return errors.New("账户已停用")
	}
	if !crypto.VerifyPassword(u.PasswordHash, oldPlain) {
		return errors.New("旧密码错误")
	}
	hash, err := crypto.HashPassword(newPlain)
	if err != nil {
		return err
	}
	return s.r.UpdatePasswordHash(ctx, u.ID, hash)
}

// ✅ 新增：通用字段更新（Username/OrgID/Description/Role）
type UpdateInput struct {
	ID          string
	Username    *string
	OrgID       *string
	Description *string
	Role        *int // 权限在 handler 控制：仅管理员可传入有效值
}

func (s *Service) Update(ctx context.Context, in UpdateInput) error {
	if in.ID == "" {
		return errors.New("ID 不能为空")
	}
	fields := map[string]any{}
	if in.Username != nil {
		if *in.Username == "" {
			return errors.New("用户名不能为空")
		}
		fields["username"] = *in.Username
	}
	if in.OrgID != nil {
		if *in.OrgID == "" {
			return errors.New("org_id 不能为空")
		}
		fields["org_id"] = *in.OrgID
	}
	if in.Description != nil {
		fields["description"] = in.Description // *string -> 可置 NULL
	}
	if in.Role != nil {
		fields["role"] = *in.Role
	}
	if len(fields) == 0 {
		return nil
	}
	return s.r.UpdateFields(ctx, in.ID, fields)
}
