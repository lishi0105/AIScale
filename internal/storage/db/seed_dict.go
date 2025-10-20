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
}{
	{"个"}, {"包"}, {"件"}, {"盒"}, {"KG"}, {"斤"},
	{"支"}, {"条"}, {"瓶"}, {"本"}, {"台"}, {"板"},
	{"箱"}, {"卷"}, {"袋"}, {"只"}, {"份"},
}

// ② 商品规格（根据你的截图示例，后续可再补）
var defaultSpecs = []struct {
	Name string
}{
	// —— 状态 / 温度 —— //
	{"新鲜"}, {"新鲜散装"}, {"新鲜不杀"},
	{"鲜活"}, {"冰鲜"},

	// —— 包装形态 —— //
	{"散装"}, {"盒装"},

	// —— 重量规格 —— //
	{"180g"},
	{"250g"},
	{"500g"},
	{"700g"},
	{"1.25L"},
}

// ③ 餐次
var defaultMealTimes = []struct {
	Name string
}{
	{"早餐"}, {"午餐"}, {"晚餐"},
}

func upsertDictByName(ctx context.Context, db *gorm.DB, row any) error {
	return db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "name"}}, // 要求表上有 UNIQUE(name)
			DoUpdates: clause.Assignments(map[string]any{
				"is_deleted": 0,
			}),
		}).
		Create(row).Error
}

// 共用的 upsert（按 name 唯一冲突更新 sort）
func EnsureDefaultDicts(ctx context.Context, gdb *gorm.DB) error {
	// units
	for _, it := range defaultUnits {
		if err := upsertDictByName(ctx, gdb, &dict.Unit{Name: it.Name}); err != nil {
			return err
		}
	}

	// specs
	for _, it := range defaultSpecs {
		if err := upsertDictByName(ctx, gdb, &dict.Spec{Name: it.Name}); err != nil {
			return err
		}
	}

	// meal times
	for _, it := range defaultMealTimes {
		if err := upsertDictByName(ctx, gdb, &dict.MealTime{Name: it.Name}); err != nil {
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
