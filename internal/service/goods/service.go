package goods

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	domain "hdzk.cn/foodapp/internal/domain/goods"
	repo "hdzk.cn/foodapp/internal/repository/goods"
)

type Service struct {
	r repo.GoodsRepository
}

func NewService(r repo.GoodsRepository) *Service { return &Service{r: r} }

type CreateParams struct {
	Name        string
	OrgID       string
	SpecID      string
	UnitID      string
	CategoryID  string
	Sort        *int
	Code        *string
	Pinyin      *string
	ImageURL    *string
	Description *string
}

type UpdateParams struct {
	ID          string
	Name        *string
	Code        *string
	Pinyin      *string
	Sort        *int
	SpecID      *string
	UnitID      *string
	CategoryID  *string
	ImageURL    *string
	Description *string
}

func (s *Service) CreateGoods(ctx context.Context, params CreateParams) (*domain.Goods, error) {
	name, err := normalizeRequiredValue(params.Name, "name")
	if err != nil {
		return nil, err
	}
	orgID, err := normalizeRequiredValue(params.OrgID, "org_id")
	if err != nil {
		return nil, err
	}
	specID, err := normalizeRequiredValue(params.SpecID, "spec_id")
	if err != nil {
		return nil, err
	}
	unitID, err := normalizeRequiredValue(params.UnitID, "unit_id")
	if err != nil {
		return nil, err
	}
	categoryID, err := normalizeRequiredValue(params.CategoryID, "category_id")
	if err != nil {
		return nil, err
	}

	normalizedCode, _ := normalizeOptional(params.Code)
	normalizedPinyin, _ := normalizeOptional(params.Pinyin)
	normalizedImageURL, _ := normalizeOptional(params.ImageURL)
	normalizedDescription, _ := normalizeOptional(params.Description)

	m := &domain.Goods{
		ID:          uuid.NewString(),
		Name:        name,
		OrgID:       orgID,
		SpecID:      specID,
		UnitID:      unitID,
		CategoryID:  categoryID,
		Code:        normalizedCode,
		Pinyin:      normalizedPinyin,
		ImageURL:    normalizedImageURL,
		Description: normalizedDescription,
	}
	if params.Sort != nil {
		m.Sort = *params.Sort
	}
	return m, s.r.CreateGoods(ctx, m)
}

func (s *Service) GetGoods(ctx context.Context, id string) (*domain.Goods, error) {
	return s.r.GetGoods(ctx, strings.TrimSpace(id))
}

func (s *Service) ListGoods(ctx context.Context, keyword string, orgID string, categoryID, specID, unitID *string, page, pageSize int) ([]domain.Goods, int64, error) {
	trimmedOrg := strings.TrimSpace(orgID)
	if trimmedOrg == "" {
		return nil, 0, fmt.Errorf("org_id 不能为空")
	}
	var categoryPtr *string
	if categoryID != nil {
		normalized, err := normalizeOptionalWithOriginal(categoryID)
		if err != nil {
			return nil, 0, err
		}
		categoryPtr = normalized
	}
	var specPtr *string
	if specID != nil {
		normalized, err := normalizeOptionalWithOriginal(specID)
		if err != nil {
			return nil, 0, err
		}
		specPtr = normalized
	}

	var unitPtr *string
	if unitID != nil {
		normalized, err := normalizeOptionalWithOriginal(unitID)
		if err != nil {
			return nil, 0, err
		}
		unitPtr = normalized
	}
	kw := strings.TrimSpace(keyword)
	return s.r.ListGoods(ctx, kw, trimmedOrg, categoryPtr, specPtr, unitPtr, page, pageSize)
}

func (s *Service) UpdateGoods(ctx context.Context, params UpdateParams) error {
	normalizedName, err := normalizeOptionalRequired(params.Name, "name")
	if err != nil {
		return err
	}
	normalizedCode, err := normalizeOptionalRequired(params.Code, "code")
	if err != nil {
		return err
	}
	normalizedSpec, err := normalizeOptionalRequired(params.SpecID, "spec_id")
	if err != nil {
		return err
	}
	normalizedCategory, err := normalizeOptionalRequired(params.CategoryID, "category_id")
	if err != nil {
		return err
	}
	normalizedPinyin, updatePinyin := normalizeOptional(params.Pinyin)
	normalizedImage, updateImage := normalizeOptional(params.ImageURL)
	normalizedDescription, updateDescription := normalizeOptional(params.Description)

	repoParams := repo.UpdateParams{
		ID:                strings.TrimSpace(params.ID),
		Name:              normalizedName,
		Code:              normalizedCode,
		Pinyin:            normalizedPinyin,
		Sort:              params.Sort,
		SpecID:            normalizedSpec,
		CategoryID:        normalizedCategory,
		ImageURL:          normalizedImage,
		UpdatePinyin:      updatePinyin,
		UpdateImageURL:    updateImage,
		Description:       normalizedDescription,
		UpdateDescription: updateDescription,
	}
	return s.r.UpdateGoods(ctx, repoParams)
}

func (s *Service) SoftDeleteGoods(ctx context.Context, id string) error {
	return s.r.SoftDeleteGoods(ctx, strings.TrimSpace(id))
}

func (s *Service) HardDeleteGoods(ctx context.Context, id string) error {
	return s.r.HardDeleteGoods(ctx, strings.TrimSpace(id))
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

func normalizeRequiredOptional(str *string) (*string, bool) {
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
