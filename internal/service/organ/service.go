package organ

import (
	"context"
	"strings"

	"github.com/google/uuid"
	domain "hdzk.cn/foodapp/internal/domain/organ"
	repo "hdzk.cn/foodapp/internal/repository/organ"
)

type Service struct {
	repo repo.Repository
}

func New(repo repo.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, input *domain.Organ) (*domain.Organ, error) {
	input.ID = uuid.NewString()
	normalize(input)
	if err := s.repo.Create(ctx, input); err != nil {
		return nil, err
	}
	return input, nil
}

func (s *Service) Update(ctx context.Context, input *domain.Organ) error {
	normalize(input)
	return s.repo.Update(ctx, input)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.SoftDelete(ctx, id)
}

func (s *Service) Get(ctx context.Context, id string) (*domain.Organ, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context, query domain.ListQuery) ([]domain.Organ, int64, error) {
	if query.Limit <= 0 {
		query.Limit = 20
	}
	if query.Offset < 0 {
		query.Offset = 0
	}
	query.Keyword = strings.TrimSpace(query.Keyword)
	return s.repo.List(ctx, query)
}

func normalize(o *domain.Organ) {
	o.Name = strings.TrimSpace(o.Name)
	if o.Code != nil {
		trimmed := strings.TrimSpace(*o.Code)
		if trimmed == "" {
			o.Code = nil
		} else {
			upper := strings.ToUpper(trimmed)
			o.Code = &upper
		}
	}
	o.Leader = strings.TrimSpace(o.Leader)
	o.Phone = strings.TrimSpace(o.Phone)
	o.Remark = strings.TrimSpace(o.Remark)
	if o.Sort < 0 {
		o.Sort = 0
	}
	if o.Status != 0 {
		o.Status = 1
	}
}
