package market_inquiry

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	domain "hdzk.cn/foodapp/internal/domain/market_inquiry"
	repo "hdzk.cn/foodapp/internal/repository/market_inquiry"
)

type Service struct {
	r repo.MarketInquiryRepository
}

func NewService(r repo.MarketInquiryRepository) *Service { return &Service{r: r} }

type CreateParams struct {
	InquiryID      string
	ItemID         string
	MarketID       *string
	MarketNameSnap string
	Price          *float64
}

type UpdateParams struct {
	ID             string
	MarketID       *string
	MarketNameSnap *string
	Price          *float64
}

type ListParams struct {
	InquiryID string
	ItemID    *string
	MarketID  *string
	Page      int
	PageSize  int
}

func (s *Service) CreateMarketInquiry(ctx context.Context, params CreateParams) (*domain.MarketInquiry, error) {
	inquiryID, err := normalizeRequiredValue(params.InquiryID, "inquiry_id")
	if err != nil {
		return nil, err
	}
	itemID, err := normalizeRequiredValue(params.ItemID, "item_id")
	if err != nil {
		return nil, err
	}
	marketNameSnap, err := normalizeRequiredValue(params.MarketNameSnap, "market_name_snap")
	if err != nil {
		return nil, err
	}

	normalizedMarketID, _ := normalizeOptional(params.MarketID)

	m := &domain.MarketInquiry{
		ID:             uuid.NewString(),
		InquiryID:      inquiryID,
		ItemID:         itemID,
		MarketID:       normalizedMarketID,
		MarketNameSnap: marketNameSnap,
		Price:          params.Price,
	}
	return m, s.r.CreateMarketInquiry(ctx, m)
}

func (s *Service) GetMarketInquiry(ctx context.Context, id string) (*domain.MarketInquiry, error) {
	return s.r.GetMarketInquiry(ctx, strings.TrimSpace(id))
}

func (s *Service) ListMarketInquiries(ctx context.Context, params ListParams) ([]domain.MarketInquiry, int64, error) {
	trimmedInquiryID := strings.TrimSpace(params.InquiryID)
	if trimmedInquiryID == "" {
		return nil, 0, fmt.Errorf("inquiry_id 不能为空")
	}

	var itemID *string
	if params.ItemID != nil {
		normalized, err := normalizeOptionalWithOriginal(params.ItemID)
		if err != nil {
			return nil, 0, err
		}
		itemID = normalized
	}

	var marketID *string
	if params.MarketID != nil {
		normalized, err := normalizeOptionalWithOriginal(params.MarketID)
		if err != nil {
			return nil, 0, err
		}
		marketID = normalized
	}

	repoParams := repo.ListParams{
		InquiryID: trimmedInquiryID,
		ItemID:    itemID,
		MarketID:  marketID,
		Page:      params.Page,
		PageSize:  params.PageSize,
	}
	return s.r.ListMarketInquiries(ctx, repoParams)
}

func (s *Service) UpdateMarketInquiry(ctx context.Context, params UpdateParams) error {
	normalizedMarketID, _ := normalizeOptional(params.MarketID)
	normalizedMarketNameSnap, err := normalizeOptionalRequired(params.MarketNameSnap, "market_name_snap")
	if err != nil {
		return err
	}

	repoParams := repo.UpdateParams{
		ID:             strings.TrimSpace(params.ID),
		MarketID:       normalizedMarketID,
		MarketNameSnap: normalizedMarketNameSnap,
		Price:          params.Price,
	}
	return s.r.UpdateMarketInquiry(ctx, repoParams)
}

func (s *Service) SoftDeleteMarketInquiry(ctx context.Context, id string) error {
	return s.r.SoftDeleteMarketInquiry(ctx, strings.TrimSpace(id))
}

func (s *Service) HardDeleteMarketInquiry(ctx context.Context, id string) error {
	return s.r.HardDeleteMarketInquiry(ctx, strings.TrimSpace(id))
}

func (s *Service) BatchCreateMarketInquiries(ctx context.Context, items []domain.MarketInquiry) error {
	return s.r.BatchCreateMarketInquiries(ctx, items)
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