package organ

import (
	"context"

	domain "hdzk.cn/foodapp/internal/domain/organ"
)

type Repository interface {
	Create(ctx context.Context, organ *domain.Organ) error
	Update(ctx context.Context, organ *domain.Organ) error
	SoftDelete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*domain.Organ, error)
	List(ctx context.Context, query domain.ListQuery) ([]domain.Organ, int64, error)
}
