package dict

import (
	"context"

	dict "hdzk.cn/foodapp/internal/domain/dict"
)

type DictRepo interface {
	// Unit
	CreateUnit(ctx context.Context, m *dict.Unit) error
	GetUnit(ctx context.Context, id string) (*dict.Unit, error)
	ListUnits(ctx context.Context, keyword string, page, pageSize int) ([]dict.Unit, int64, error)
	UpdateUnit(ctx context.Context, id string, name string, code *string, sort int, updateCode bool) error
	DeleteUnit(ctx context.Context, id string) error

	// Spec
	CreateSpec(ctx context.Context, m *dict.Spec) error
	GetSpec(ctx context.Context, id string) (*dict.Spec, error)
	ListSpecs(ctx context.Context, keyword string, page, pageSize int) ([]dict.Spec, int64, error)
	UpdateSpec(ctx context.Context, id string, name string, code *string, sort int, updateCode bool) error
	DeleteSpec(ctx context.Context, id string) error

	// MealTime
	CreateMealTime(ctx context.Context, m *dict.MealTime) error
	GetMealTime(ctx context.Context, id string) (*dict.MealTime, error)
	ListMealTimes(ctx context.Context, keyword string, page, pageSize int) ([]dict.MealTime, int64, error)
	UpdateMealTime(ctx context.Context, id string, name string, code *string, sort int, updateCode bool) error
	DeleteMealTime(ctx context.Context, id string) error
}
