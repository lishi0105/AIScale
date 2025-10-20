package inquiry

import (
	"context"
	"fmt"
	"strings"
	"time"

	domain "hdzk.cn/foodapp/internal/domain/inquiry"
	repo "hdzk.cn/foodapp/internal/repository/inquiry"
)

type Service struct{ r repo.Repository }

func NewService(r repo.Repository) *Service { return &Service{r: r} }

type CreateParams struct {
	OrgID        string
	InquiryTitle string
	InquiryDate  time.Time
	Market1      *string
	Market2      *string
	Market3      *string
}

type UpdateParams struct {
	ID           string
	InquiryTitle *string
	InquiryDate  *time.Time
	Market1      *string
	Market2      *string
	Market3      *string
}

func (s *Service) Create(ctx context.Context, p CreateParams) (*domain.PriceInquiry, error) {
	if strings.TrimSpace(p.OrgID) == "" {
		return nil, fmt.Errorf("org_id 不能为空")
	}
	if strings.TrimSpace(p.InquiryTitle) == "" {
		return nil, fmt.Errorf("inquiry_title 不能为空")
	}

	m := &domain.PriceInquiry{
		OrgID:        strings.TrimSpace(p.OrgID),
		InquiryTitle: strings.TrimSpace(p.InquiryTitle),
		InquiryDate:  p.InquiryDate,
		Market1:      normalizePtr(p.Market1),
		Market2:      normalizePtr(p.Market2),
		Market3:      normalizePtr(p.Market3),
	}
	return m, s.r.Create(ctx, m)
}

func (s *Service) Get(ctx context.Context, id string) (*domain.PriceInquiry, error) {
	return s.r.Get(ctx, strings.TrimSpace(id))
}

func (s *Service) List(ctx context.Context, orgID string, keyword string, dateFrom, dateTo *time.Time, page, pageSize int) ([]domain.PriceInquiry, int64, error) {
	trimmedOrg := strings.TrimSpace(orgID)
	if trimmedOrg == "" {
		return nil, 0, fmt.Errorf("org_id 不能为空")
	}
	kw := strings.TrimSpace(keyword)
	return s.r.List(ctx, trimmedOrg, kw, dateFrom, dateTo, page, pageSize)
}

func (s *Service) Update(ctx context.Context, p UpdateParams) error {
	rp := repo.UpdateParams{
		ID:           strings.TrimSpace(p.ID),
		InquiryTitle: normalizePtr(p.InquiryTitle),
		InquiryDate:  p.InquiryDate,
		Market1:      normalizePtr(p.Market1),
		Market2:      normalizePtr(p.Market2),
		Market3:      normalizePtr(p.Market3),
	}
	return s.r.Update(ctx, rp)
}

func (s *Service) SoftDelete(ctx context.Context, id string) error {
	return s.r.SoftDelete(ctx, strings.TrimSpace(id))
}

func (s *Service) HardDelete(ctx context.Context, id string) error {
	return s.r.HardDelete(ctx, strings.TrimSpace(id))
}

func normalizePtr(p *string) *string {
	if p == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*p)
	if trimmed == "" {
		return nil
	}
	v := trimmed
	return &v
}
