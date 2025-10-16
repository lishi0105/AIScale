package organ

import (
	"context"

	"gorm.io/gorm"
	domain "hdzk.cn/foodapp/internal/domain/organ"
)

type Repository interface {
	Create(ctx context.Context, m *domain.Organ) error
	GetByID(ctx context.Context, id string) (*domain.Organ, error)
	List(ctx context.Context, q domain.ListQuery) ([]*domain.Organ, int64, error)
	UpdateFields(ctx context.Context, id string, fields map[string]any) error
	SoftDelete(ctx context.Context, id string) error
	HardDelete(ctx context.Context, id string) error
}

func NewRepository(db *gorm.DB) Repository {
	return &gormRepo{db: db}
}
