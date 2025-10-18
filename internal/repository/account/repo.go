package account

import (
	"context"

	"gorm.io/gorm"
	domain "hdzk.cn/foodapp/internal/domain/account"
)

type Repository interface {
	// C
	Create(ctx context.Context, a *domain.Account) error

	// R
	GetByID(ctx context.Context, id string) (*domain.Account, error)
	GetByUsername(ctx context.Context, username string) (*domain.Account, error)
	List(ctx context.Context, NameLike string, Deleted, Role *int, page, page_size int) ([]domain.Account, int64, error)

	// U
	UpdatePasswordHash(ctx context.Context, id string, hash string) error
	UpdateFields(ctx context.Context, id string, fields map[string]any) error

	// D
	SoftDelete(ctx context.Context, id string) error
	HardDelete(ctx context.Context, id string) error
}

func NewRepository(db *gorm.DB) *GormRepo { return &GormRepo{db: db} }
