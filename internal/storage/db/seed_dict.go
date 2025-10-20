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
	{"支"}, {"条"}, {"瓶"}, {"本"}, {"台"},
	{"箱"}, {"卷"}, {"袋"}, {"只"}, {"份"},
}

// ② 商品规格（根据你的截图示例，后续可再补）
var defaultSpecs = []struct {
	Name string
}{
	// —— 状态 / 温度 —— //
	{"新鲜"}, {"新鲜散装"}, {"新鲜不杀"},
	{"冷鲜"}, {"冰鲜"}, {"冷冻"}, {"鲜活"},
	{"腌制"}, {"熟制"}, {"半成品"},

	// —— 包装 / 形态 —— //
	{"散装"}, {"袋装"}, {"盒装"}, {"瓶装"}, {"罐装"}, {"听装"}, {"桶装"},
	{"真空装"}, {"托盘保鲜膜"}, {"带托盒装"},
	{"整只"}, {"半只"}, {"整条"}, {"整块"},
	{"切块"}, {"切片"}, {"切丝"}, {"切丁"}, {"段"},
	{"带骨"}, {"去骨"}, {"去皮"},
	{"带壳"}, {"去壳"}, {"去头去壳"}, {"去刺"}, {"去内脏"}, {"开背"},

	// —— 重量（g/kg） —— //
	{"100g"}, {"150g"}, {"180g"}, {"200g"}, {"250g"}, {"300g"},
	{"350g"}, {"400g"}, {"450g"}, {"500g"}, {"600g"}, {"700g"},
	{"750g"}, {"800g"}, {"900g"}, {"1kg"}, {"1.5kg"}, {"2kg"},
	{"2.5kg"}, {"3kg"}, {"5kg"},

	// —— 体积（ml/L） —— //
	{"180ml"}, {"200ml"}, {"250ml"}, {"330ml"}, {"350ml"}, {"500ml"},
	{"1L"}, {"1.25L"}, {"1.5L"}, {"1.8L"}, {"2L"}, {"5L"},

	// —— 计件 / 分级 —— //
	{"5枚装"}, {"6枚装"}, {"10枚装"}, {"12枚装"}, {"15枚装"}, {"30枚/板"},
	{"小个"}, {"中个"}, {"大个"}, {"特大个"},
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
