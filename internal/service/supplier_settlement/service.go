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
	SettlementPrice  *float64
}

type UpdateParams struct {
	ID               string
	SupplierID       *string
	SupplierNameSnap *string
	FloatRatioSnap   *float64
	SettlementPrice  *float64
}

type ListParams struct {
	InquiryID  string
	ItemID     *string
	SupplierID *string
	Page       int
	PageSize   int
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
	if params.FloatRatioSnap <= 0 {
		return nil, fmt.Errorf("float_ratio_snap 必须大于0")
	}

	normalizedSupplierID, _ := normalizeOptional(params.SupplierID)

	m := &domain.SupplierSettlement{
		ID:                uuid.NewString(),
		InquiryID:         inquiryID,
		ItemID:            itemID,
		SupplierID:        normalizedSupplierID,
		SupplierNameSnap:  supplierNameSnap,
		FloatRatioSnap:    params.FloatRatioSnap,
		SettlementPrice:   params.SettlementPrice,
	}
	return m, s.r.CreateSupplierSettlement(ctx, m)
}

func (s *Service) GetSupplierSettlement(ctx context.Context, id string) (*domain.SupplierSettlement, error) {
	return s.r.GetSupplierSettlement(ctx, strings.TrimSpace(id))
}

func (s *Service) ListSupplierSettlements(ctx context.Context, params ListParams) ([]domain.SupplierSettlement, int64, error) {
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

	var supplierID *string
	if params.SupplierID != nil {
		normalized, err := normalizeOptionalWithOriginal(params.SupplierID)
		if err != nil {
			return nil, 0, err
		}
		supplierID = normalized
	}

	repoParams := repo.ListParams{
		InquiryID:  trimmedInquiryID,
		ItemID:     itemID,
		SupplierID: supplierID,
		Page:       params.Page,
		PageSize:   params.PageSize,
	}
	return s.r.ListSupplierSettlements(ctx, repoParams)
}

func (s *Service) UpdateSupplierSettlement(ctx context.Context, params UpdateParams) error {
	normalizedSupplierID, _ := normalizeOptional(params.SupplierID)
	normalizedSupplierNameSnap, err := normalizeOptionalRequired(params.SupplierNameSnap, "supplier_name_snap")
	if err != nil {
		return err
	}

	repoParams := repo.UpdateParams{
		ID:               strings.TrimSpace(params.ID),
		SupplierID:       normalizedSupplierID,
		SupplierNameSnap: normalizedSupplierNameSnap,
		FloatRatioSnap:   params.FloatRatioSnap,
		SettlementPrice:  params.SettlementPrice,
	}
	return s.r.UpdateSupplierSettlement(ctx, repoParams)
}

func (s *Service) SoftDeleteSupplierSettlement(ctx context.Context, id string) error {
	return s.r.SoftDeleteSupplierSettlement(ctx, strings.TrimSpace(id))
}

func (s *Service) HardDeleteSupplierSettlement(ctx context.Context, id string) error {
	return s.r.HardDeleteSupplierSettlement(ctx, strings.TrimSpace(id))
}

func (s *Service) BatchCreateSupplierSettlements(ctx context.Context, items []domain.SupplierSettlement) error {
	return s.r.BatchCreateSupplierSettlements(ctx, items)
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