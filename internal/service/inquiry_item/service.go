package inquiry_item

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	domain "hdzk.cn/foodapp/internal/domain/inquiry_item"
	repo "hdzk.cn/foodapp/internal/repository/inquiry_item"
)

type Service struct {
	r repo.InquiryItemRepository
}

func NewService(r repo.InquiryItemRepository) *Service { return &Service{r: r} }

type CreateParams struct {
	InquiryID         string
	GoodsID           string
	CategoryID        string
	SpecID            *string
	UnitID            *string
	GoodsNameSnap     string
	CategoryNameSnap  string
	SpecNameSnap      *string
	UnitNameSnap      *string
	GuidePrice        *float64
	LastMonthAvgPrice *float64
	CurrentAvgPrice   *float64
	Sort              *int
}

type UpdateParams struct {
	ID                string
	InquiryID         *string
	GoodsID           *string
	CategoryID        *string
	SpecID            *string
	UnitID            *string
	GoodsNameSnap     *string
	CategoryNameSnap  *string
	SpecNameSnap      *string
	UnitNameSnap      *string
	GuidePrice        *float64
	LastMonthAvgPrice *float64
	CurrentAvgPrice   *float64
	Sort              *int
}

func (s *Service) CreateInquiryItem(ctx context.Context, params CreateParams) (*domain.InquiryItem, error) {
	inquiryID, err := normalizeRequiredValue(params.InquiryID, "inquiry_id")
	if err != nil {
		return nil, err
	}
	goodsID, err := normalizeRequiredValue(params.GoodsID, "goods_id")
	if err != nil {
		return nil, err
	}
	categoryID, err := normalizeRequiredValue(params.CategoryID, "category_id")
	if err != nil {
		return nil, err
	}
	goodsNameSnap, err := normalizeRequiredValue(params.GoodsNameSnap, "goods_name_snap")
	if err != nil {
		return nil, err
	}
	categoryNameSnap, err := normalizeRequiredValue(params.CategoryNameSnap, "category_name_snap")
	if err != nil {
		return nil, err
	}

	normalizedSpecID, _ := normalizeOptional(params.SpecID)
	normalizedUnitID, _ := normalizeOptional(params.UnitID)
	normalizedSpecName, _ := normalizeOptional(params.SpecNameSnap)
	normalizedUnitName, _ := normalizeOptional(params.UnitNameSnap)

	m := &domain.InquiryItem{
		ID:                uuid.NewString(),
		InquiryID:         inquiryID,
		GoodsID:           goodsID,
		CategoryID:        categoryID,
		SpecID:            normalizedSpecID,
		UnitID:            normalizedUnitID,
		GoodsNameSnap:     goodsNameSnap,
		CategoryNameSnap:  categoryNameSnap,
		SpecNameSnap:      normalizedSpecName,
		UnitNameSnap:      normalizedUnitName,
		GuidePrice:        params.GuidePrice,
		LastMonthAvgPrice: params.LastMonthAvgPrice,
		CurrentAvgPrice:   params.CurrentAvgPrice,
	}
	if params.Sort != nil {
		m.Sort = *params.Sort
	}
	return m, s.r.CreateInquiryItem(ctx, m)
}

func (s *Service) GetInquiryItem(ctx context.Context, id string) (*domain.InquiryItem, error) {
	return s.r.GetInquiryItem(ctx, strings.TrimSpace(id))
}

func (s *Service) ListInquiryItems(ctx context.Context, inquiryID string, categoryID *string, page, pageSize int) ([]domain.InquiryItem, int64, error) {
	trimmedInquiryID := strings.TrimSpace(inquiryID)
	if trimmedInquiryID == "" {
		return nil, 0, fmt.Errorf("inquiry_id 不能为空")
	}
	var categoryPtr *string
	if categoryID != nil {
		normalized, err := normalizeOptionalWithOriginal(categoryID)
		if err != nil {
			return nil, 0, err
		}
		categoryPtr = normalized
	}
	return s.r.ListInquiryItems(ctx, trimmedInquiryID, categoryPtr, page, pageSize)
}

func (s *Service) UpdateInquiryItem(ctx context.Context, params UpdateParams) error {
	normalizedInquiryID, err := normalizeOptionalRequired(params.InquiryID, "inquiry_id")
	if err != nil {
		return err
	}
	normalizedGoodsID, err := normalizeOptionalRequired(params.GoodsID, "goods_id")
	if err != nil {
		return err
	}
	normalizedCategoryID, err := normalizeOptionalRequired(params.CategoryID, "category_id")
	if err != nil {
		return err
	}
	normalizedGoodsName, err := normalizeOptionalRequired(params.GoodsNameSnap, "goods_name_snap")
	if err != nil {
		return err
	}
	normalizedCategoryName, err := normalizeOptionalRequired(params.CategoryNameSnap, "category_name_snap")
	if err != nil {
		return err
	}

	normalizedSpecID, updateSpecID := normalizeOptional(params.SpecID)
	normalizedUnitID, updateUnitID := normalizeOptional(params.UnitID)
	normalizedSpecName, updateSpecName := normalizeOptional(params.SpecNameSnap)
	normalizedUnitName, updateUnitName := normalizeOptional(params.UnitNameSnap)
	_, updateGuidePrice := normalizeOptionalFloat(params.GuidePrice)
	_, updateLastMonth := normalizeOptionalFloat(params.LastMonthAvgPrice)
	_, updateCurrentAvg := normalizeOptionalFloat(params.CurrentAvgPrice)

	repoParams := repo.UpdateParams{
		ID:                strings.TrimSpace(params.ID),
		InquiryID:         normalizedInquiryID,
		GoodsID:           normalizedGoodsID,
		CategoryID:        normalizedCategoryID,
		SpecID:            normalizedSpecID,
		UnitID:            normalizedUnitID,
		GoodsNameSnap:     normalizedGoodsName,
		CategoryNameSnap:  normalizedCategoryName,
		SpecNameSnap:      normalizedSpecName,
		UnitNameSnap:      normalizedUnitName,
		GuidePrice:        params.GuidePrice,
		LastMonthAvgPrice: params.LastMonthAvgPrice,
		CurrentAvgPrice:   params.CurrentAvgPrice,
		Sort:              params.Sort,
		UpdateSpecID:      updateSpecID,
		UpdateUnitID:      updateUnitID,
		UpdateSpecName:    updateSpecName,
		UpdateUnitName:    updateUnitName,
		UpdateGuidePrice:  updateGuidePrice,
		UpdateLastMonth:   updateLastMonth,
		UpdateCurrentAvg:  updateCurrentAvg,
	}
	return s.r.UpdateInquiryItem(ctx, repoParams)
}

func (s *Service) SoftDeleteInquiryItem(ctx context.Context, id string) error {
	return s.r.SoftDeleteInquiryItem(ctx, strings.TrimSpace(id))
}

func (s *Service) HardDeleteInquiryItem(ctx context.Context, id string) error {
	return s.r.HardDeleteInquiryItem(ctx, strings.TrimSpace(id))
}

func normalizeOptional(str *string) (*string, bool) {
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

func normalizeOptionalFloat(f *float64) (*float64, bool) {
	if f == nil {
		return nil, false
	}
	return f, true
}

func normalizeRequiredValue(val string, field string) (string, error) {
	trimmed := strings.TrimSpace(val)
	if trimmed == "" {
		return "", fmt.Errorf("%s 不能为空", field)
	}
	return trimmed, nil
}

func normalizeOptionalRequired(str *string, field string) (*string, error) {
	if str == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*str)
	if trimmed == "" {
		return nil, fmt.Errorf("%s 不能为空", field)
	}
	normalized := trimmed
	return &normalized, nil
}

func normalizeOptionalWithOriginal(str *string) (*string, error) {
	if str == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*str)
	if trimmed == "" {
		return nil, nil
	}
	normalized := trimmed
	return &normalized, nil
}
