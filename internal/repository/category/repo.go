package category

import (
	"context"

	"gorm.io/gorm"
	category "hdzk.cn/foodapp/internal/domain/category"
)

type CategoryRepository interface {
	Create(ctx context.Context, m *category.Category) error
	Get(ctx context.Context, id string) (*category.Category, error)
	List(ctx context.Context, keyword string, org_id string, page, pageSize int) ([]category.Category, int64, error)
	Update(ctx context.Context, id string, name string, code *string, pinyin *string, sort *int, updateCode bool, updatePinyin bool, updateSort bool) error
	SoftDelete(ctx context.Context, id string) error
	HardDelete(ctx context.Context, id string) error
	FindByName(ctx context.Context, name string, orgID string) (*category.Category, error)
}

func NewRepository(db *gorm.DB) CategoryRepository { return &categoryRepo{db: db} }
