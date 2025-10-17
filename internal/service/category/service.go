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

func (s *Service) CreateCategory(ctx context.Context, name string, code *string, pinyin *string) (*domain.Category, error) {
	normalizedCode, _ := normalizeCode(code)
	normalizedPinyin, _ := normalizePinyin(pinyin)
	m := &domain.Category{
		ID:     uuid.NewString(),
		Name:   name,
		Code:   normalizedCode,
		Pinyin: normalizedPinyin,
	}
	return m, s.r.CreateCategory(ctx, m)
}

func (s *Service) GetCategory(ctx context.Context, id string) (*domain.Category, error) {
	return s.r.GetCategory(ctx, id)
}

func (s *Service) ListCategories(ctx context.Context, keyword string, page, pageSize int) ([]domain.Category, int64, error) {
	return s.r.ListCategories(ctx, keyword, page, pageSize)
}

func (s *Service) UpdateCategory(ctx context.Context, id, name string, code *string, pinyin *string) error {
	normalizedCode, updateCode := normalizeCode(code)
	normalizedPinyin, _ := normalizePinyin(pinyin)
	return s.r.UpdateCategory(ctx, id, name, normalizedCode, normalizedPinyin, updateCode)
}

func (s *Service) DeleteCategory(ctx context.Context, id string) error {
	return s.r.DeleteCategory(ctx, id)
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

func normalizePinyin(pinyin *string) (*string, bool) {
	if pinyin == nil {
		return nil, false
	}
	trimmed := strings.TrimSpace(*pinyin)
	if trimmed == "" {
		return nil, true
	}
	normalized := trimmed
	return &normalized, true
}