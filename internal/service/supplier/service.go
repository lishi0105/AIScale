package supplier

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

type CreateParams struct {
	Name           string
	OrgID          string
	Description    string
	FloatRatio     float64
	Code           *string
	Pinyin         *string
	ContactName    *string
	ContactPhone   *string
	ContactEmail   *string
	ContactAddress *string
	Status         *int
	StartTime      *time.Time
	EndTime        *time.Time
}

type UpdateParams struct {
	ID              string
	Name            *string
	Code            *string
	Pinyin          *string
	Sort            *int
	Status          *int
	Description     *string
	FloatRatio      *float64
	ContactName     *string
	ContactPhone    *string
	ContactEmail    *string
	ContactAddress  *string
	StartTime       *time.Time
	EndTime         *time.Time
	UpdateSort      bool
	UpdateStartTime bool
	UpdateEndTime   bool
}

func (s *Service) CreateSupplier(ctx context.Context, params CreateParams) (*domain.Supplier, error) {
	normalizedCode, _ := normalizeString(params.Code)
	normalizedPinyin, _ := normalizeString(params.Pinyin)
	contactName, _ := normalizeString(params.ContactName)
	contactPhone, _ := normalizeString(params.ContactPhone)
	contactEmail, _ := normalizeString(params.ContactEmail)
	contactAddress, _ := normalizeString(params.ContactAddress)

	status := 1
	if params.Status != nil {
		status = *params.Status
	}

	m := &domain.Supplier{
		ID:             uuid.NewString(),
		Name:           params.Name,
		OrgID:          params.OrgID,
		Description:    params.Description,
		FloatRatio:     params.FloatRatio,
		Code:           normalizedCode,
		Pinyin:         normalizedPinyin,
		ContactName:    contactName,
		ContactPhone:   contactPhone,
		ContactEmail:   contactEmail,
		ContactAddress: contactAddress,
		Status:         status,
		StartTime:      params.StartTime,
		EndTime:        params.EndTime,
	}
	return m, s.r.CreateSupplier(ctx, m)
}

func (s *Service) GetSupplier(ctx context.Context, id string) (*domain.Supplier, error) {
	return s.r.GetSupplier(ctx, id)
}

func (s *Service) ListSuppliers(ctx context.Context, keyword string, orgID *string, status *int, contactName, contactPhone, contactEmail, contactAddress *string, page, pageSize int) ([]domain.Supplier, int64, error) {
	return s.r.ListSuppliers(ctx, keyword, orgID, status, contactName, contactPhone, contactEmail, contactAddress, page, pageSize)
}

func (s *Service) UpdateSupplier(ctx context.Context, params UpdateParams) error {
	normalizedCode, updateCode := normalizeString(params.Code)
	normalizedPinyin, updatePinyin := normalizeString(params.Pinyin)
	normalizedContactName, updateContactName := normalizeString(params.ContactName)
	normalizedContactPhone, updateContactPhone := normalizeString(params.ContactPhone)
	normalizedContactEmail, updateContactEmail := normalizeString(params.ContactEmail)
	normalizedContactAddress, updateContactAddress := normalizeString(params.ContactAddress)

	repoParams := repo.UpdateParams{
		ID:                   params.ID,
		Name:                 params.Name,
		Code:                 normalizedCode,
		Pinyin:               normalizedPinyin,
		Sort:                 params.Sort,
		Status:               params.Status,
		Description:          params.Description,
		FloatRatio:           params.FloatRatio,
		ContactName:          normalizedContactName,
		ContactPhone:         normalizedContactPhone,
		ContactEmail:         normalizedContactEmail,
		ContactAddress:       normalizedContactAddress,
		StartTime:            params.StartTime,
		EndTime:              params.EndTime,
		UpdateCode:           updateCode,
		UpdatePinyin:         updatePinyin,
		UpdateSort:           params.UpdateSort,
		UpdateContactName:    updateContactName,
		UpdateContactPhone:   updateContactPhone,
		UpdateContactEmail:   updateContactEmail,
		UpdateContactAddress: updateContactAddress,
		UpdateStartTime:      params.UpdateStartTime,
		UpdateEndTime:        params.UpdateEndTime,
	}
	return s.r.UpdateSupplier(ctx, repoParams)
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
