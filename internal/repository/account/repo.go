package account

import (
	"context"

	domain "hdzk.cn/foodapp/internal/domain/account"
)

type Repository interface {
	// C
	Create(ctx context.Context, a *domain.Account) error

	// R
	GetByID(ctx context.Context, id string) (*domain.Account, error)
	GetByUsername(ctx context.Context, username string) (*domain.Account, error)
	List(ctx context.Context, q domain.ListQuery) ([]domain.Account, int64, error)

	// U
	UpdateEmail(ctx context.Context, id string, email string) error
	UpdatePasswordHash(ctx context.Context, id string, hash string) error
	UpdateStatus(ctx context.Context, id string, status int) error

	// D
	SoftDeleteByID(ctx context.Context, id string) error
	HardDeleteByID(ctx context.Context, id string) error
}
