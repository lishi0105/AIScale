package service

import (
	"context"

	"github.com/google/uuid"
	domain "hdzk.cn/foodapp/internal/domain/dict"
	repo "hdzk.cn/foodapp/internal/repository/dict"
)

type DictService struct {
	r repo.DictRepo
}

func New(r repo.DictRepo) *DictService { return &DictService{r: r} }

// Unit
func (s *DictService) CreateUnit(ctx context.Context, name string, sort int) (*domain.Unit, error) {
	m := &domain.Unit{ID: uuid.NewString(), Name: name, Sort: sort}
	return m, s.r.CreateUnit(ctx, m)
}
func (s *DictService) GetUnit(ctx context.Context, id string) (*domain.Unit, error) {
	return s.r.GetUnit(ctx, id)
}
func (s *DictService) ListUnits(ctx context.Context, keyword string, page, pageSize int) ([]domain.Unit, int64, error) {
	return s.r.ListUnits(ctx, keyword, page, pageSize)
}
func (s *DictService) UpdateUnit(ctx context.Context, id, name string, sort int) error {
	return s.r.UpdateUnit(ctx, id, name, sort)
}
func (s *DictService) DeleteUnit(ctx context.Context, id string) error {
	return s.r.DeleteUnit(ctx, id)
}

// Spec
func (s *DictService) CreateSpec(ctx context.Context, name string, sort int) (*domain.Spec, error) {
	m := &domain.Spec{ID: uuid.NewString(), Name: name, Sort: sort}
	return m, s.r.CreateSpec(ctx, m)
}
func (s *DictService) GetSpec(ctx context.Context, id string) (*domain.Spec, error) {
	return s.r.GetSpec(ctx, id)
}
func (s *DictService) ListSpecs(ctx context.Context, keyword string, page, pageSize int) ([]domain.Spec, int64, error) {
	return s.r.ListSpecs(ctx, keyword, page, pageSize)
}
func (s *DictService) UpdateSpec(ctx context.Context, id, name string, sort int) error {
	return s.r.UpdateSpec(ctx, id, name, sort)
}
func (s *DictService) DeleteSpec(ctx context.Context, id string) error {
	return s.r.DeleteSpec(ctx, id)
}

// MealTime
func (s *DictService) CreateMealTime(ctx context.Context, name string, sort int) (*domain.MealTime, error) {
	m := &domain.MealTime{ID: uuid.NewString(), Name: name, Sort: sort}
	return m, s.r.CreateMealTime(ctx, m)
}
func (s *DictService) GetMealTime(ctx context.Context, id string) (*domain.MealTime, error) {
	return s.r.GetMealTime(ctx, id)
}
func (s *DictService) ListMealTimes(ctx context.Context, keyword string, page, pageSize int) ([]domain.MealTime, int64, error) {
	return s.r.ListMealTimes(ctx, keyword, page, pageSize)
}
func (s *DictService) UpdateMealTime(ctx context.Context, id, name string, sort int) error {
	return s.r.UpdateMealTime(ctx, id, name, sort)
}
func (s *DictService) DeleteMealTime(ctx context.Context, id string) error {
	return s.r.DeleteMealTime(ctx, id)
}
