package foodDB

import (
	"context"

	dict "hdzk.cn/foodapp/internal/domain/dict"
	"hdzk.cn/foodapp/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ① 商品单位
var defaultUnits = []struct {
	Name string
	Sort int
}{
	{"个", 1}, {"包", 2}, {"件", 3}, {"盒", 4}, {"KG", 5}, {"斤", 6},
	{"支", 7}, {"条", 8}, {"瓶", 9}, {"本", 10}, {"台", 11},
	{"箱", 12}, {"卷", 13}, {"袋", 14}, {"只", 15}, {"份", 16},
}

// ② 商品规格（根据你的截图示例，后续可再补）
var defaultSpecs = []struct {
	Name string
	Sort int
}{
	{"新鲜", 1}, {"新鲜散装", 2}, {"新鲜不杀", 3},
	{"500g", 4}, {"180g", 5}, {"250g", 6},
	{"700g", 7}, {"1.25L", 8},
}

// ③ 餐次
var defaultMealTimes = []struct {
	Name string
	Sort int
}{
	{"早餐", 1}, {"午餐", 2}, {"晚餐", 3},
}

// 共用的 upsert（按 name 唯一冲突更新 sort）
func upsertByName(ctx context.Context, db *gorm.DB, row any, sort int) error {
	return db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "name"}}, // 要求表上有 UNIQUE(name)
			DoUpdates: clause.Assignments(map[string]any{
				"sort":       sort,
				"is_deleted": 0,
			}),
		}).
		Create(row).Error
}

func EnsureDefaultDicts(ctx context.Context, gdb *gorm.DB) error {
	// units
	for _, it := range defaultUnits {
		if err := upsertByName(ctx, gdb, &dict.Unit{Name: it.Name, Sort: it.Sort}, it.Sort); err != nil {
			return err
		}
	}

	// specs
	for _, it := range defaultSpecs {
		if err := upsertByName(ctx, gdb, &dict.Spec{Name: it.Name, Sort: it.Sort}, it.Sort); err != nil {
			return err
		}
	}

	// meal times
	for _, it := range defaultMealTimes {
		if err := upsertByName(ctx, gdb, &dict.MealTime{Name: it.Name, Sort: it.Sort}, it.Sort); err != nil {
			return err
		}
	}

	logger.L().Info("dicts seeded/ensured",
		zap.Int("units", len(defaultUnits)),
		zap.Int("specs", len(defaultSpecs)),
		zap.Int("mealtimes", len(defaultMealTimes)),
	)
	return nil
}
