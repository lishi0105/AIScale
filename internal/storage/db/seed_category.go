package foodDB

import (
	"context"

	category "hdzk.cn/foodapp/internal/domain/category"
	"hdzk.cn/foodapp/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ① 商品单位
var defaultCategory = []struct {
	Name string
}{
	{"肉类"}, {"干货"}, {"牛奶"}, {"豆类淀粉制品"},
	{"调料"}, {"果蔬"}, {"鲜（冻）水产品"},
	{"蛋类"}, {"豆制品"}, {"面筋制品"},
}

// 共用的 upsert（按 name 唯一冲突更新 sort）
func upsertCategoryByName(ctx context.Context, db *gorm.DB, row any) error {
	return db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "name"}}, // 要求表上有 UNIQUE(name)
			DoUpdates: clause.Assignments(map[string]any{
				"is_deleted": 0,
			}),
		}).
		Create(row).Error
}

func EnsureDefaultCategory(ctx context.Context, gdb *gorm.DB) error {
	// category
	for _, it := range defaultCategory {
		if err := upsertCategoryByName(ctx, gdb, &category.Category{Name: it.Name, OrgID: DefaultOrgID}); err != nil {
			return err
		}
	}

	logger.L().Info("Category seeded/ensured",
		zap.Int("category", len(defaultCategory)),
	)
	return nil
}
