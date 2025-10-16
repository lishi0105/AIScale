package service

import (
	"context"
	"strings"

	"github.com/google/uuid"
	domain "hdzk.cn/foodapp/internal/domain/dict"
	repo "hdzk.cn/foodapp/internal/repository/dict"
)

type Service struct {
	r repo.DictRepository
}

func NewService(r repo.DictRepository) *Service { return &Service{r: r} }

// Unit
func (s *Service) CreateUnit(ctx context.Context, name string, code *string, sort int) (*domain.Unit, error) {
	normalizedCode, _ := normalizeCode(code)
	m := &domain.Unit{ID: uuid.NewString(), Name: name, Sort: sort, Code: normalizedCode}
	return m, s.r.CreateUnit(ctx, m)
}
func (s *Service) GetUnit(ctx context.Context, id string) (*domain.Unit, error) {
	return s.r.GetUnit(ctx, id)
}
func (s *Service) ListUnits(ctx context.Context, keyword string, page, pageSize int) ([]domain.Unit, int64, error) {
	return s.r.ListUnits(ctx, keyword, page, pageSize)
}
func (s *Service) UpdateUnit(ctx context.Context, id, name string, code *string, sort int) error {
	normalizedCode, updateCode := normalizeCode(code)
	return s.r.UpdateUnit(ctx, id, name, normalizedCode, sort, updateCode)
}
func (s *Service) DeleteUnit(ctx context.Context, id string) error {
	return s.r.DeleteUnit(ctx, id)
}

// Spec
func (s *Service) CreateSpec(ctx context.Context, name string, code *string, sort int) (*domain.Spec, error) {
	normalizedCode, _ := normalizeCode(code)
	m := &domain.Spec{ID: uuid.NewString(), Name: name, Sort: sort, Code: normalizedCode}
	return m, s.r.CreateSpec(ctx, m)
}
func (s *Service) GetSpec(ctx context.Context, id string) (*domain.Spec, error) {
	return s.r.GetSpec(ctx, id)
}
func (s *Service) ListSpecs(ctx context.Context, keyword string, page, pageSize int) ([]domain.Spec, int64, error) {
	return s.r.ListSpecs(ctx, keyword, page, pageSize)
}
func (s *Service) UpdateSpec(ctx context.Context, id, name string, code *string, sort int) error {
	normalizedCode, updateCode := normalizeCode(code)
	return s.r.UpdateSpec(ctx, id, name, normalizedCode, sort, updateCode)
}
func (s *Service) DeleteSpec(ctx context.Context, id string) error {
	return s.r.DeleteSpec(ctx, id)
}

// MealTime
func (s *Service) CreateMealTime(ctx context.Context, name string, code *string, sort int) (*domain.MealTime, error) {
	normalizedCode, _ := normalizeCode(code)
	m := &domain.MealTime{ID: uuid.NewString(), Name: name, Sort: sort, Code: normalizedCode}
	return m, s.r.CreateMealTime(ctx, m)
}
func (s *Service) GetMealTime(ctx context.Context, id string) (*domain.MealTime, error) {
	return s.r.GetMealTime(ctx, id)
}
func (s *Service) ListMealTimes(ctx context.Context, keyword string, page, pageSize int) ([]domain.MealTime, int64, error) {
	return s.r.ListMealTimes(ctx, keyword, page, pageSize)
}
func (s *Service) UpdateMealTime(ctx context.Context, id, name string, code *string, sort int) error {
	normalizedCode, updateCode := normalizeCode(code)
	return s.r.UpdateMealTime(ctx, id, name, normalizedCode, sort, updateCode)
}
func (s *Service) DeleteMealTime(ctx context.Context, id string) error {
	return s.r.DeleteMealTime(ctx, id)
}

func normalizeCode(code *string) (*string, bool) {
	if code == nil {
		return nil, false
	}
	trimmed := strings.TrimSpace(*code)
	if trimmed == "" {
		return nil, true
	}
	normalized := trimmed
	return &normalized, true
}
