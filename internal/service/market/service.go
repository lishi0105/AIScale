package market

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	domain "hdzk.cn/foodapp/internal/domain/market"
	repo "hdzk.cn/foodapp/internal/repository/market"
)

type Service struct {
	r repo.MarketRepository
}

func NewService(r repo.MarketRepository) *Service { return &Service{r: r} }

type CreateParams struct {
	Name   string
	OrgID  string
	Code   *string
	Sort   *int
}

type UpdateParams struct {
	ID   string
	Name *string
	Code *string
	Sort *int
}

func (s *Service) CreateMarket(ctx context.Context, params CreateParams) (*domain.Market, error) {
	name, err := normalizeRequiredValue(params.Name, "name")
	if err != nil {
		return nil, err
	}
	orgID, err := normalizeRequiredValue(params.OrgID, "org_id")
	if err != nil {
		return nil, err
	}

	normalizedCode, _ := normalizeOptional(params.Code)

	m := &domain.Market{
		ID:    uuid.NewString(),
		Name:  name,
		OrgID: orgID,
		Code:  normalizedCode,
	}
	if params.Sort != nil {
		m.Sort = *params.Sort
	}
	return m, s.r.CreateMarket(ctx, m)
}

func (s *Service) GetMarket(ctx context.Context, id string) (*domain.Market, error) {
	return s.r.GetMarket(ctx, strings.TrimSpace(id))
}

func (s *Service) ListMarkets(ctx context.Context, keyword string, orgID string, page, pageSize int) ([]domain.Market, int64, error) {
	trimmedOrg := strings.TrimSpace(orgID)
	if trimmedOrg == "" {
		return nil, 0, fmt.Errorf("org_id 不能为空")
	}
	kw := strings.TrimSpace(keyword)
	return s.r.ListMarkets(ctx, kw, trimmedOrg, page, pageSize)
}

func (s *Service) UpdateMarket(ctx context.Context, params UpdateParams) error {
	normalizedName, err := normalizeOptionalRequired(params.Name, "name")
	if err != nil {
		return err
	}
	normalizedCode, err := normalizeOptionalRequired(params.Code, "code")
	if err != nil {
		return err
	}

	repoParams := repo.UpdateParams{
		ID:   strings.TrimSpace(params.ID),
		Name: normalizedName,
		Code: normalizedCode,
		Sort: params.Sort,
	}
	return s.r.UpdateMarket(ctx, repoParams)
}

func (s *Service) SoftDeleteMarket(ctx context.Context, id string) error {
	return s.r.SoftDeleteMarket(ctx, strings.TrimSpace(id))
}

func (s *Service) HardDeleteMarket(ctx context.Context, id string) error {
	return s.r.HardDeleteMarket(ctx, strings.TrimSpace(id))
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