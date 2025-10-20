package goods

import (
	"context"

	"gorm.io/gorm"
	domain "hdzk.cn/foodapp/internal/domain/goods"
)

type UpdateParams struct {
	ID                string
	Name              *string
	Code              *string
	Pinyin            *string
	Sort              *int
	SpecID            *string
	UnitID            *string
	CategoryID        *string
	ImageURL          *string
	UpdatePinyin      bool
	UpdateImageURL    bool
	Description       *string
	UpdateDescription bool
}

type GoodsRepository interface {
	CreateGoods(ctx context.Context, m *domain.Goods) error
	GetGoods(ctx context.Context, id string) (*domain.Goods, error)
	ListGoods(ctx context.Context, keyword string, orgID string, categoryID, specID, unitID *string, page, pageSize int) ([]domain.Goods, int64, error)
	UpdateGoods(ctx context.Context, params UpdateParams) error
	SoftDeleteGoods(ctx context.Context, id string) error
	HardDeleteGoods(ctx context.Context, id string) error
}

func NewRepository(db *gorm.DB) GoodsRepository { return &goodsRepo{db: db} }
