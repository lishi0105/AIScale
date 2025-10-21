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
	GoodsNameSnap     *string
	CategoryNameSnap  *string
	SpecNameSnap      *string
	UnitNameSnap      *string
	GuidePrice        *float64
	LastMonthAvgPrice *float64
	CurrentAvgPrice   *float64
	Sort              *int
}

type ListParams struct {
	InquiryID  string
	CategoryID *string
	GoodsID    *string
	Page       int
	PageSize   int
}

type BatchCreateParams struct {
	Items []CreateParams
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
	normalizedSpecNameSnap, _ := normalizeOptional(params.SpecNameSnap)
	normalizedUnitNameSnap, _ := normalizeOptional(params.UnitNameSnap)

	m := &domain.InquiryItem{
		ID:                uuid.NewString(),
		InquiryID:         inquiryID,
		GoodsID:           goodsID,
		CategoryID:        categoryID,
		SpecID:            normalizedSpecID,
		UnitID:            normalizedUnitID,
		GoodsNameSnap:     goodsNameSnap,
		CategoryNameSnap:  categoryNameSnap,
		SpecNameSnap:      normalizedSpecNameSnap,
		UnitNameSnap:      normalizedUnitNameSnap,
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

func (s *Service) ListInquiryItems(ctx context.Context, params ListParams) ([]domain.InquiryItem, int64, error) {
	inquiryID := strings.TrimSpace(params.InquiryID)
	if inquiryID == "" {
		return nil, 0, fmt.Errorf("inquiry_id 不能为空")
	}

	var categoryPtr *string
	if params.CategoryID != nil {
		normalized, err := normalizeOptionalWithOriginal(params.CategoryID)
		if err != nil {
			return nil, 0, err
		}
		categoryPtr = normalized
	}

	var goodsPtr *string
	if params.GoodsID != nil {
		normalized, err := normalizeOptionalWithOriginal(params.GoodsID)
		if err != nil {
			return nil, 0, err
		}
		goodsPtr = normalized
	}

	repoParams := repo.ListParams{
		InquiryID:  inquiryID,
		CategoryID: categoryPtr,
		GoodsID:    goodsPtr,
		Page:       params.Page,
		PageSize:   params.PageSize,
	}
	return s.r.ListInquiryItems(ctx, repoParams)
}

func (s *Service) UpdateInquiryItem(ctx context.Context, params UpdateParams) error {
	normalizedGoodsNameSnap, err := normalizeOptionalRequired(params.GoodsNameSnap, "goods_name_snap")
	if err != nil {
		return err
	}
	normalizedCategoryNameSnap, err := normalizeOptionalRequired(params.CategoryNameSnap, "category_name_snap")
	if err != nil {
		return err
	}
	normalizedSpecNameSnap, _ := normalizeOptional(params.SpecNameSnap)
	normalizedUnitNameSnap, _ := normalizeOptional(params.UnitNameSnap)

	repoParams := repo.UpdateParams{
		ID:                strings.TrimSpace(params.ID),
		GoodsNameSnap:     normalizedGoodsNameSnap,
		CategoryNameSnap:  normalizedCategoryNameSnap,
		SpecNameSnap:      normalizedSpecNameSnap,
		UnitNameSnap:      normalizedUnitNameSnap,
		GuidePrice:        params.GuidePrice,
		LastMonthAvgPrice: params.LastMonthAvgPrice,
		CurrentAvgPrice:   params.CurrentAvgPrice,
		Sort:              params.Sort,
	}
	return s.r.UpdateInquiryItem(ctx, repoParams)
}

func (s *Service) SoftDeleteInquiryItem(ctx context.Context, id string) error {
	return s.r.SoftDeleteInquiryItem(ctx, strings.TrimSpace(id))
}

func (s *Service) HardDeleteInquiryItem(ctx context.Context, id string) error {
	return s.r.HardDeleteInquiryItem(ctx, strings.TrimSpace(id))
}

func (s *Service) BatchCreateInquiryItems(ctx context.Context, params BatchCreateParams) error {
	if len(params.Items) == 0 {
		return nil
	}

	items := make([]domain.InquiryItem, 0, len(params.Items))
	for _, itemParams := range params.Items {
		item, err := s.CreateInquiryItem(ctx, itemParams)
		if err != nil {
			return fmt.Errorf("创建询价商品失败: %w", err)
		}
		items = append(items, *item)
	}

	return s.r.BatchCreateInquiryItems(ctx, items)
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