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
	{"个", 1}, {"包", 2}, {"件", 3}, {"盒", 4}, {"KG", 5},
	{"支", 6}, {"条", 7}, {"瓶", 8}, {"本", 9}, {"台", 10},
	{"箱", 11}, {"卷", 12}, {"袋", 13}, {"只", 14}, {"份", 15},
}

// ② 商品规格（根据你的截图示例，后续可再补）
var defaultSpecs = []struct {
	Name string
	Sort int
}{
	{"新鲜", 1}, {"新鲜散装", 2}, {"新鲜不杀", 3},
	{"500g", 4}, {"180g", 5}, {"250g", 6}, {"1.25L", 7},
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
			Columns:   []clause.Column{{Name: "name"}},                  // 要求表上有 UNIQUE(name)
			DoUpdates: clause.Assignments(map[string]any{"sort": sort}), // 冲突时更新 sort
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
