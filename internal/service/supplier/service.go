package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	domain "hdzk.cn/foodapp/internal/domain/supplier"
	repo "hdzk.cn/foodapp/internal/repository/supplier"
)

type Service struct {
	r repo.SupplierRepository
}

func NewService(r repo.SupplierRepository) *Service { return &Service{r: r} }

func (s *Service) CreateSupplier(ctx context.Context, name string, code *string, pinyin *string, status *int, description string, floatRatio float64, orgID *string, startTime *time.Time, endTime *time.Time) (*domain.Supplier, error) {
	normalizedCode, _ := normalizeString(code)
	normalizedPinyin, _ := normalizeString(pinyin)
	normalizedOrgID, _ := normalizeString(orgID)
	
	// 默认状态为 1（正常）
	supplierStatus := 1
	if status != nil {
		supplierStatus = *status
	}
	
	m := &domain.Supplier{
		ID:          uuid.NewString(),
		Name:        name,
		Code:        normalizedCode,
		Pinyin:      normalizedPinyin,
		Status:      supplierStatus,
		Description: description,
		FloatRatio:  floatRatio,
		OrgID:       normalizedOrgID,
		StartTime:   startTime,
		EndTime:     endTime,
	}
	return m, s.r.CreateSupplier(ctx, m)
}

func (s *Service) GetSupplier(ctx context.Context, id string) (*domain.Supplier, error) {
	return s.r.GetSupplier(ctx, id)
}

func (s *Service) ListSuppliers(ctx context.Context, keyword string, orgID *string, status *int, page, pageSize int) ([]domain.Supplier, int64, error) {
	return s.r.ListSuppliers(ctx, keyword, orgID, status, page, pageSize)
}

func (s *Service) UpdateSupplier(ctx context.Context, id, name string, code *string, pinyin *string, sort *int, status *int, description *string, floatRatio *float64, orgID *string, startTime *time.Time, endTime *time.Time) error {
	normalizedCode, updateCode := normalizeString(code)
	normalizedPinyin, updatePinyin := normalizeString(pinyin)
	normalizedOrgID, updateOrgID := normalizeString(orgID)
	updateSort := sort != nil
	updateStatus := status != nil
	updateDescription := description != nil
	updateFloatRatio := floatRatio != nil
	updateStartTime := startTime != nil
	updateEndTime := endTime != nil
	
	return s.r.UpdateSupplier(ctx, id, name, normalizedCode, normalizedPinyin, sort, status, description, floatRatio, normalizedOrgID, startTime, endTime, updateCode, updatePinyin, updateSort, updateStatus, updateDescription, updateFloatRatio, updateOrgID, updateStartTime, updateEndTime)
}

func (s *Service) SoftDeleteSupplier(ctx context.Context, id string) error {
	return s.r.SoftDeleteSupplier(ctx, id)
}

func (s *Service) HardDeleteSupplier(ctx context.Context, id string) error {
	return s.r.HardDeleteSupplier(ctx, id)
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
