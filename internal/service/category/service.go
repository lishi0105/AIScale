package service

import (
	"context"
	"strings"

	"github.com/google/uuid"
	domain "hdzk.cn/foodapp/internal/domain/category"
	repo "hdzk.cn/foodapp/internal/repository/category"
)

type Service struct {
	r repo.CategoryRepository
}

func NewService(r repo.CategoryRepository) *Service { return &Service{r: r} }

func (s *Service) CreateCategory(ctx context.Context, name string, team_id string, code *string, pinyin *string) (*domain.Category, error) {
	normalizedCode, _ := normalizeString(code)
	normalizedPinyin, _ := normalizeString(pinyin)
	m := &domain.Category{
		ID:     uuid.NewString(),
		Name:   name,
		TeamID: team_id,
		Code:   normalizedCode,
		Pinyin: normalizedPinyin,
	}
	return m, s.r.CreateCategory(ctx, m)
}

func (s *Service) GetCategory(ctx context.Context, id string) (*domain.Category, error) {
	return s.r.GetCategory(ctx, id)
}

func (s *Service) ListCategories(ctx context.Context, keyword string, team_id string, page, pageSize int) ([]domain.Category, int64, error) {
	return s.r.ListCategories(ctx, keyword, team_id, page, pageSize)
}

func (s *Service) UpdateCategory(ctx context.Context, id, name string, code *string, pinyin *string, sort *int) error {
	normalizedCode, updateCode := normalizeString(code)
	normalizedPinyin, updatePinyin := normalizeString(pinyin)
	updateSort := sort != nil
	return s.r.UpdateCategory(ctx, id, name, normalizedCode, normalizedPinyin, sort, updateCode, updatePinyin, updateSort)
}

func (s *Service) DeleteCategory(ctx context.Context, id string) error {
	return s.r.DeleteCategory(ctx, id)
}

func normalizeString(str *string) (*string, bool) {
	if str == nil {
		return nil, false
	}
	trimmed := strings.TrimSpace(*str)
	if trimmed == "" {
		return nil, true
	}
	normalized := trimmed
	return &normalized, true
}
