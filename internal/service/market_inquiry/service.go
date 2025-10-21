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
	Price          float64
}

type UpdateParams struct {
	ID             string
	InquiryID      *string
	ItemID         *string
	MarketID       *string
	MarketNameSnap *string
	Price          *float64
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

func (s *Service) ListMarketInquiries(ctx context.Context, inquiryID *string, itemID *string, page, pageSize int) ([]domain.MarketInquiry, int64, error) {
	var inquiryPtr *string
	if inquiryID != nil {
		normalized, err := normalizeOptionalWithOriginal(inquiryID)
		if err != nil {
			return nil, 0, err
		}
		inquiryPtr = normalized
	}
	var itemPtr *string
	if itemID != nil {
		normalized, err := normalizeOptionalWithOriginal(itemID)
		if err != nil {
			return nil, 0, err
		}
		itemPtr = normalized
	}
	return s.r.ListMarketInquiries(ctx, inquiryPtr, itemPtr, page, pageSize)
}

func (s *Service) UpdateMarketInquiry(ctx context.Context, params UpdateParams) error {
	normalizedInquiryID, err := normalizeOptionalRequired(params.InquiryID, "inquiry_id")
	if err != nil {
		return err
	}
	normalizedItemID, err := normalizeOptionalRequired(params.ItemID, "item_id")
	if err != nil {
		return err
	}
	normalizedMarketName, err := normalizeOptionalRequired(params.MarketNameSnap, "market_name_snap")
	if err != nil {
		return err
	}

	normalizedMarketID, updateMarketID := normalizeOptional(params.MarketID)

	repoParams := repo.UpdateParams{
		ID:             strings.TrimSpace(params.ID),
		InquiryID:      normalizedInquiryID,
		ItemID:         normalizedItemID,
		MarketID:       normalizedMarketID,
		MarketNameSnap: normalizedMarketName,
		Price:          params.Price,
		UpdateMarketID: updateMarketID,
	}
	return s.r.UpdateMarketInquiry(ctx, repoParams)
}

func (s *Service) DeleteMarketInquiry(ctx context.Context, id string) error {
	return s.r.DeleteMarketInquiry(ctx, strings.TrimSpace(id))
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
