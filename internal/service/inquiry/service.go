package inquiry

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	domain "hdzk.cn/foodapp/internal/domain/inquiry"
	repo "hdzk.cn/foodapp/internal/repository/inquiry"
)

type Service struct {
	r repo.InquiryRepository
}

func NewService(r repo.InquiryRepository) *Service { return &Service{r: r} }

type CreateParams struct {
	InquiryTitle     string
	InquiryDate      time.Time
	Market1          *string
	Market2          *string
	Market3          *string
	OrgID            string
	InquiryStartDate time.Time
	InquiryEndDate   time.Time
}

type UpdateParams struct {
	ID               string
	InquiryTitle     *string
	InquiryDate      *time.Time
	Market1          *string
	Market2          *string
	Market3          *string
	InquiryStartDate *time.Time
	InquiryEndDate   *time.Time
}

func (s *Service) CreateInquiry(ctx context.Context, params CreateParams) (*domain.Inquiry, error) {
	title, err := normalizeRequiredValue(params.InquiryTitle, "inquiry_title")
	if err != nil {
		return nil, err
	}
	orgID, err := normalizeRequiredValue(params.OrgID, "org_id")
	if err != nil {
		return nil, err
	}

	// 验证时间范围
	if !params.InquiryEndDate.After(params.InquiryStartDate) {
		return nil, fmt.Errorf("结束时间必须晚于开始时间")
	}

	normalizedMarket1, _ := normalizeOptional(params.Market1)
	normalizedMarket2, _ := normalizeOptional(params.Market2)
	normalizedMarket3, _ := normalizeOptional(params.Market3)

	m := &domain.Inquiry{
		ID:               uuid.NewString(),
		InquiryTitle:     title,
		InquiryDate:      params.InquiryDate,
		Market1:          normalizedMarket1,
		Market2:          normalizedMarket2,
		Market3:          normalizedMarket3,
		OrgID:            orgID,
		InquiryStartDate: params.InquiryStartDate,
		InquiryEndDate:   params.InquiryEndDate,
	}
	return m, s.r.CreateInquiry(ctx, m)
}

func (s *Service) GetInquiry(ctx context.Context, id string) (*domain.Inquiry, error) {
	return s.r.GetInquiry(ctx, strings.TrimSpace(id))
}

func (s *Service) ListInquiries(ctx context.Context, keyword string, orgID string, startDate, endDate *time.Time, page, pageSize int) ([]domain.Inquiry, int64, error) {
	trimmedOrg := strings.TrimSpace(orgID)
	if trimmedOrg == "" {
		return nil, 0, fmt.Errorf("org_id 不能为空")
	}
	kw := strings.TrimSpace(keyword)
	return s.r.ListInquiries(ctx, kw, trimmedOrg, startDate, endDate, page, pageSize)
}

func (s *Service) UpdateInquiry(ctx context.Context, params UpdateParams) error {
	normalizedTitle, err := normalizeOptionalRequired(params.InquiryTitle, "inquiry_title")
	if err != nil {
		return err
	}

	normalizedMarket1, updateMarket1 := normalizeOptional(params.Market1)
	normalizedMarket2, updateMarket2 := normalizeOptional(params.Market2)
	normalizedMarket3, updateMarket3 := normalizeOptional(params.Market3)

	repoParams := repo.UpdateParams{
		ID:               strings.TrimSpace(params.ID),
		InquiryTitle:     normalizedTitle,
		InquiryDate:      params.InquiryDate,
		Market1:          normalizedMarket1,
		Market2:          normalizedMarket2,
		Market3:          normalizedMarket3,
		InquiryStartDate: params.InquiryStartDate,
		InquiryEndDate:   params.InquiryEndDate,
		UpdateMarket1:    updateMarket1,
		UpdateMarket2:    updateMarket2,
		UpdateMarket3:    updateMarket3,
	}
	return s.r.UpdateInquiry(ctx, repoParams)
}

func (s *Service) SoftDeleteInquiry(ctx context.Context, id string) error {
	return s.r.SoftDeleteInquiry(ctx, strings.TrimSpace(id))
}

func (s *Service) HardDeleteInquiry(ctx context.Context, id string) error {
	return s.r.HardDeleteInquiry(ctx, strings.TrimSpace(id))
}

// 辅助函数
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
