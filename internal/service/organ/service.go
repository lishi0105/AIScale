package organ

import (
	"context"
	"errors"

	domain "hdzk.cn/foodapp/internal/domain/organ"
	repo "hdzk.cn/foodapp/internal/repository/organ"
)

/************ 对外接口（handler 只依赖这个接口） ************/
// type Service interface {
// 	Create(ctx context.Context, in CreateInput) (*domain.Organ, error)
// 	GetByID(ctx context.Context, id string) (*domain.Organ, error)
// 	List(ctx context.Context, q domain.ListQuery) ([]*domain.Organ, int64, error)
// 	Update(ctx context.Context, in UpdateInput) (*domain.Organ, error)
// 	SoftDelete(ctx context.Context, id string) error
// 	HardDelete(ctx context.Context, id string) error
// }

/************ 具体实现（持有 repo.Repository 接口字段） ************/
type Service struct {
	r repo.Repository
}

func NewService(r repo.Repository) *Service { return &Service{r: r} }

/************ DTO ************/
type UpdateInput struct {
	ID          string
	Name        *string
	Parent      *string
	Code        *string
	Description *string
}

/************ 方法实现 ************/
func (s *Service) Create(ctx context.Context, in *domain.Organ) error {
	if in.Name == "" {
		return errors.New("名称不能为空")
	}
	if err := s.r.Create(ctx, in); err != nil {
		return err
	}
	return nil
}

func (s *Service) GetByID(ctx context.Context, id string) (*domain.Organ, error) {
	return s.r.GetByID(ctx, id)
}

type ListQuery struct {
	NameLike string // 模糊匹配
	Deleted  *int
	Role     *int
	Limit    int
	Offset   int
}

func (s *Service) List(ctx context.Context, NameLike string, Deleted, Role *int, page, page_size int) ([]domain.Organ, int64, error) {
	return s.r.List(ctx, NameLike, Deleted, Role, page, page_size)
}

func (s *Service) Update(ctx context.Context, in UpdateInput) error {
	if in.ID == "" {
		return errors.New("ID 不能为空")
	}
	updates := map[string]any{}
	if in.Name != nil {
		updates["name"] = *in.Name
	}
	if in.Parent != nil {
		updates["parent"] = *in.Parent
	}
	if in.Code != nil {
		updates["code"] = *in.Code // 允许置为 NULL
	}
	if in.Description != nil {
		updates["description"] = *in.Description
	}
	return s.r.UpdateFields(ctx, in.ID, updates)
}

func (s *Service) SoftDelete(ctx context.Context, id string) error {
	return s.r.SoftDelete(ctx, id)
}

func (s *Service) HardDelete(ctx context.Context, id string) error {
	return s.r.HardDelete(ctx, id)
}
