package supplier_settlement

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	domain "hdzk.cn/foodapp/internal/domain/supplier_settlement"
	repo "hdzk.cn/foodapp/internal/repository/supplier_settlement"
)

type Service struct {
	r repo.SupplierSettlementRepository
}

func NewService(r repo.SupplierSettlementRepository) *Service { return &Service{r: r} }

type CreateParams struct {
	InquiryID        string
	ItemID           string
	SupplierID       *string
	SupplierNameSnap string
	FloatRatioSnap   float64
	SettlementPrice  float64
}

type UpdateParams struct {
	ID               string
	InquiryID        *string
	ItemID           *string
	SupplierID       *string
	SupplierNameSnap *string
	FloatRatioSnap   *float64
	SettlementPrice  *float64
}

func (s *Service) CreateSupplierSettlement(ctx context.Context, params CreateParams) (*domain.SupplierSettlement, error) {
	inquiryID, err := normalizeRequiredValue(params.InquiryID, "inquiry_id")
	if err != nil {
		return nil, err
	}
	itemID, err := normalizeRequiredValue(params.ItemID, "item_id")
	if err != nil {
		return nil, err
	}
	supplierNameSnap, err := normalizeRequiredValue(params.SupplierNameSnap, "supplier_name_snap")
	if err != nil {
		return nil, err
	}

	normalizedSupplierID, _ := normalizeOptional(params.SupplierID)

	m := &domain.SupplierSettlement{
		ID:               uuid.NewString(),
		InquiryID:        inquiryID,
		ItemID:           itemID,
		SupplierID:       normalizedSupplierID,
		SupplierNameSnap: supplierNameSnap,
		FloatRatioSnap:   params.FloatRatioSnap,
		SettlementPrice:  params.SettlementPrice,
	}
	return m, s.r.CreateSupplierSettlement(ctx, m)
}

func (s *Service) GetSupplierSettlement(ctx context.Context, id string) (*domain.SupplierSettlement, error) {
	return s.r.GetSupplierSettlement(ctx, strings.TrimSpace(id))
}

func (s *Service) ListSupplierSettlements(ctx context.Context, inquiryID *string, itemID *string, page, pageSize int) ([]domain.SupplierSettlement, int64, error) {
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
	return s.r.ListSupplierSettlements(ctx, inquiryPtr, itemPtr, page, pageSize)
}

func (s *Service) UpdateSupplierSettlement(ctx context.Context, params UpdateParams) error {
	normalizedInquiryID, err := normalizeOptionalRequired(params.InquiryID, "inquiry_id")
	if err != nil {
		return err
	}
	normalizedItemID, err := normalizeOptionalRequired(params.ItemID, "item_id")
	if err != nil {
		return err
	}
	normalizedSupplierName, err := normalizeOptionalRequired(params.SupplierNameSnap, "supplier_name_snap")
	if err != nil {
		return err
	}

	normalizedSupplierID, updateSupplierID := normalizeOptional(params.SupplierID)

	repoParams := repo.UpdateParams{
		ID:               strings.TrimSpace(params.ID),
		InquiryID:        normalizedInquiryID,
		ItemID:           normalizedItemID,
		SupplierID:       normalizedSupplierID,
		SupplierNameSnap: normalizedSupplierName,
		FloatRatioSnap:   params.FloatRatioSnap,
		SettlementPrice:  params.SettlementPrice,
		UpdateSupplierID: updateSupplierID,
	}
	return s.r.UpdateSupplierSettlement(ctx, repoParams)
}

func (s *Service) DeleteSupplierSettlement(ctx context.Context, id string) error {
	return s.r.DeleteSupplierSettlement(ctx, strings.TrimSpace(id))
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
