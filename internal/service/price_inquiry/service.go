package price_inquiry

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	domain "hdzk.cn/foodapp/internal/domain/price_inquiry"
	repo "hdzk.cn/foodapp/internal/repository/price_inquiry"
)

type Service struct {
	r repo.PriceInquiryRepository
}

func NewService(r repo.PriceInquiryRepository) *Service { return &Service{r: r} }

type CreateParams struct {
	OrgID        string
	InquiryTitle string
	InquiryDate  time.Time
}

type UpdateParams struct {
	ID           string
	OrgID        *string
	InquiryTitle *string
	InquiryDate  *time.Time
}

type ListParams struct {
	OrgID    string
	Keyword  string
	Year     *int16
	Month    *int8
	TenDay   *int8
	Page     int
	PageSize int
}

func (s *Service) CreatePriceInquiry(ctx context.Context, params CreateParams) (*domain.PriceInquiry, error) {
	orgID, err := normalizeRequiredValue(params.OrgID, "org_id")
	if err != nil {
		return nil, err
	}
	title, err := normalizeRequiredValue(params.InquiryTitle, "inquiry_title")
	if err != nil {
		return nil, err
	}

	m := &domain.PriceInquiry{
		ID:           uuid.NewString(),
		OrgID:        orgID,
		InquiryTitle: title,
		InquiryDate:  params.InquiryDate,
	}
	return m, s.r.CreatePriceInquiry(ctx, m)
}

func (s *Service) GetPriceInquiry(ctx context.Context, id string) (*domain.PriceInquiry, error) {
	return s.r.GetPriceInquiry(ctx, strings.TrimSpace(id))
}

func (s *Service) ListPriceInquiries(ctx context.Context, params ListParams) ([]domain.PriceInquiry, int64, error) {
	trimmedOrg := strings.TrimSpace(params.OrgID)
	if trimmedOrg == "" {
		return nil, 0, fmt.Errorf("org_id 不能为空")
	}
	repoParams := repo.ListParams{
		OrgID:    trimmedOrg,
		Keyword:  strings.TrimSpace(params.Keyword),
		Year:     params.Year,
		Month:    params.Month,
		TenDay:   params.TenDay,
		Page:     params.Page,
		PageSize: params.PageSize,
	}
	return s.r.ListPriceInquiries(ctx, repoParams)
}

func (s *Service) UpdatePriceInquiry(ctx context.Context, params UpdateParams) error {
	normalizedOrgID, err := normalizeOptionalRequired(params.OrgID, "org_id")
	if err != nil {
		return err
	}
	normalizedTitle, err := normalizeOptionalRequired(params.InquiryTitle, "inquiry_title")
	if err != nil {
		return err
	}

	repoParams := repo.UpdateParams{
		ID:           strings.TrimSpace(params.ID),
		OrgID:        normalizedOrgID,
		InquiryTitle: normalizedTitle,
		InquiryDate:  params.InquiryDate,
	}
	return s.r.UpdatePriceInquiry(ctx, repoParams)
}

func (s *Service) SoftDeletePriceInquiry(ctx context.Context, id string) error {
	return s.r.SoftDeletePriceInquiry(ctx, strings.TrimSpace(id))
}

func (s *Service) HardDeletePriceInquiry(ctx context.Context, id string) error {
	return s.r.HardDeletePriceInquiry(ctx, strings.TrimSpace(id))
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
