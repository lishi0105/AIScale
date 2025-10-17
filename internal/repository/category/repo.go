package category

import (
	"context"

	"gorm.io/gorm"
	category "hdzk.cn/foodapp/internal/domain/category"
)

type CategoryRepository interface {
	CreateCategory(ctx context.Context, m *category.Category) error
	GetCategory(ctx context.Context, id string) (*category.Category, error)
	ListCategories(ctx context.Context, keyword string, page, pageSize int) ([]category.Category, int64, error)
	UpdateCategory(ctx context.Context, id string, name string, code *string, pinyin *string, updateCode bool) error
	DeleteCategory(ctx context.Context, id string) error
}

func NewRepository(db *gorm.DB) CategoryRepository { return &categoryRepo{db: db} }