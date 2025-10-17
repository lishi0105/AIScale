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
	{"蔬菜"}, {"水产海鲜"}, {"水果"}, {"肉、禽、蛋"},
	{"牛奶饮品"}, {"干货调料"}, {"冻品、豆制品"},
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
		if err := upsertCategoryByName(ctx, gdb, &category.Category{Name: it.Name}); err != nil {
			return err
		}
	}

	logger.L().Info("Category seeded/ensured",
		zap.Int("category", len(defaultCategory)),
	)
	return nil
}
