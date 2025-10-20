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

func NewService(r repo.InquiryRepository) *Service {
	return &Service{r: r}
}

// 创建询价记录
func (s *Service) CreateInquiry(ctx context.Context, params domain.CreateInquiryParams) (*domain.BasePriceInquiry, error) {
	// 验证必填字段
	if err := s.validateCreateParams(params); err != nil {
		return nil, err
	}

	// 检查标题和日期的唯一性
	exists, err := s.r.CheckTitleDateUnique(ctx, params.OrgID, params.InquiryTitle, params.InquiryDate, nil)
	if err != nil {
		return nil, fmt.Errorf("检查唯一性失败: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("该组织下已存在相同标题和日期的询价记录")
	}

	// 创建询价记录
	inquiry := &domain.BasePriceInquiry{
		ID:               uuid.NewString(),
		InquiryTitle:     params.InquiryTitle,
		InquiryDate:      params.InquiryDate,
		Market1:          params.Market1,
		Market2:          params.Market2,
		Market3:          params.Market3,
		OrgID:            params.OrgID,
		IsDeleted:        0,
		InquiryStartDate: params.InquiryStartDate,
		InquiryEndDate:   params.InquiryEndDate,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// 设置 ActiveTitle
	inquiry.ActiveTitle = &inquiry.InquiryTitle

	if err := s.r.CreateInquiry(ctx, inquiry); err != nil {
		return nil, fmt.Errorf("创建询价记录失败: %w", err)
	}

	return inquiry, nil
}

// 获取询价记录
func (s *Service) GetInquiry(ctx context.Context, id string) (*domain.BasePriceInquiry, error) {
	trimmedID := strings.TrimSpace(id)
	if trimmedID == "" {
		return nil, fmt.Errorf("id 不能为空")
	}

	return s.r.GetInquiry(ctx, trimmedID)
}

// 查询询价记录列表
func (s *Service) ListInquiries(ctx context.Context, params domain.InquiryQueryParams) ([]domain.BasePriceInquiry, int64, error) {
	// 验证组织ID
	if strings.TrimSpace(params.OrgID) == "" {
		return nil, 0, fmt.Errorf("org_id 不能为空")
	}

	// 设置默认分页参数
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 20
	}
	if params.PageSize > 100 {
		params.PageSize = 100 // 限制最大页面大小
	}

	// 清理关键词
	params.Keyword = strings.TrimSpace(params.Keyword)

	return s.r.ListInquiries(ctx, params)
}

// 更新询价记录
func (s *Service) UpdateInquiry(ctx context.Context, params domain.UpdateInquiryParams) error {
	// 验证ID
	if strings.TrimSpace(params.ID) == "" {
		return fmt.Errorf("id 不能为空")
	}

	// 检查记录是否存在
	existing, err := s.r.GetInquiry(ctx, params.ID)
	if err != nil {
		return fmt.Errorf("获取询价记录失败: %w", err)
	}

	// 如果更新标题或日期，需要检查唯一性
	if params.InquiryTitle != nil || params.InquiryDate != nil {
		title := existing.InquiryTitle
		date := existing.InquiryDate
		
		if params.InquiryTitle != nil {
			title = *params.InquiryTitle
		}
		if params.InquiryDate != nil {
			date = *params.InquiryDate
		}

		exists, err := s.r.CheckTitleDateUnique(ctx, existing.OrgID, title, date, &params.ID)
		if err != nil {
			return fmt.Errorf("检查唯一性失败: %w", err)
		}
		if exists {
			return fmt.Errorf("该组织下已存在相同标题和日期的询价记录")
		}
	}

	// 验证时间逻辑
	if params.InquiryStartDate != nil && params.InquiryEndDate != nil {
		if params.InquiryEndDate.Before(*params.InquiryStartDate) {
			return fmt.Errorf("结束时间必须晚于开始时间")
		}
	} else if params.InquiryStartDate != nil && existing.InquiryEndDate.Before(*params.InquiryStartDate) {
		return fmt.Errorf("结束时间必须晚于开始时间")
	} else if params.InquiryEndDate != nil && params.InquiryEndDate.Before(existing.InquiryStartDate) {
		return fmt.Errorf("结束时间必须晚于开始时间")
	}

	return s.r.UpdateInquiry(ctx, params)
}

// 软删除询价记录
func (s *Service) SoftDeleteInquiry(ctx context.Context, id string) error {
	trimmedID := strings.TrimSpace(id)
	if trimmedID == "" {
		return fmt.Errorf("id 不能为空")
	}

	// 检查记录是否存在
	_, err := s.r.GetInquiry(ctx, trimmedID)
	if err != nil {
		return fmt.Errorf("获取询价记录失败: %w", err)
	}

	return s.r.SoftDeleteInquiry(ctx, trimmedID)
}

// 硬删除询价记录
func (s *Service) HardDeleteInquiry(ctx context.Context, id string) error {
	trimmedID := strings.TrimSpace(id)
	if trimmedID == "" {
		return fmt.Errorf("id 不能为空")
	}

	return s.r.HardDeleteInquiry(ctx, trimmedID)
}

// 验证创建参数
func (s *Service) validateCreateParams(params domain.CreateInquiryParams) error {
	if strings.TrimSpace(params.InquiryTitle) == "" {
		return fmt.Errorf("inquiry_title 不能为空")
	}
	if strings.TrimSpace(params.OrgID) == "" {
		return fmt.Errorf("org_id 不能为空")
	}
	if params.InquiryDate.IsZero() {
		return fmt.Errorf("inquiry_date 不能为空")
	}
	if params.InquiryStartDate.IsZero() {
		return fmt.Errorf("inquiry_start_date 不能为空")
	}
	if params.InquiryEndDate.IsZero() {
		return fmt.Errorf("inquiry_end_date 不能为空")
	}
	if params.InquiryEndDate.Before(params.InquiryStartDate) {
		return fmt.Errorf("结束时间必须晚于开始时间")
	}

	return nil
}