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

func New(r repo.Repository) *Service { return &Service{r: r} }

func (s *Service) Authenticate(ctx context.Context, username, plain string) (string, int, error) {
	u, err := s.r.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", int(middleware.RoleUser), errors.New("用户名不存在")
		}
		return "", int(middleware.RoleUser), err
	}
	if !crypto.VerifyPassword(u.PasswordHash, plain) {
		return "", int(middleware.RoleUser), errors.New("用户名或密码错误")
	}
	role := int(u.Role)
	return u.ID, role, nil
}

func (s *Service) Create(ctx context.Context, a *domain.Account) error { return s.r.Create(ctx, a) }
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
func (s *Service) UpdateStatus(ctx context.Context, id string, status int) error {
	return s.r.UpdateStatus(ctx, id, status)
}
func (s *Service) SoftDeleteByID(ctx context.Context, id string) error {
	return s.r.SoftDeleteByID(ctx, id)
}

func (s *Service) HardDeleteByID(ctx context.Context, id string) error {
	return s.r.HardDeleteByID(ctx, id)
}

func (s *Service) ChangePassword(ctx context.Context, username, oldPlain, newPlain string) error {
	u, err := s.r.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		return err
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
