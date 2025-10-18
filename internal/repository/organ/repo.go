package organ

import (
	"context"

	"gorm.io/gorm"
	domain "hdzk.cn/foodapp/internal/domain/organ"
)

type Repository interface {
	Create(ctx context.Context, m *domain.Organ) error
	GetByID(ctx context.Context, id string) (*domain.Organ, error)
	List(ctx context.Context, NameLike string, Deleted, Role *int, page, page_size int) ([]domain.Organ, int64, error)
	UpdateFields(ctx context.Context, id string, fields map[string]any) error
	SoftDelete(ctx context.Context, id string) error
	HardDelete(ctx context.Context, id string) error
}

func NewRepository(db *gorm.DB) Repository {
	return &gormRepo{db: db}
}
