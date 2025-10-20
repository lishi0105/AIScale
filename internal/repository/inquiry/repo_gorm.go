package inquiry

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	domain "hdzk.cn/foodapp/internal/domain/inquiry"
)

type inquiryRepo struct{ db *gorm.DB }

func (r *inquiryRepo) CreateInquiry(ctx context.Context, m *domain.Inquiry) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *inquiryRepo) GetInquiry(ctx context.Context, id string) (*domain.Inquiry, error) {
	var out domain.Inquiry
	err := r.db.WithContext(ctx).
		Where("id = ? AND is_deleted = 0", id).
		First(&out).Error
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *inquiryRepo) ListInquiries(ctx context.Context, keyword string, orgID string, startDate, endDate *time.Time, page, pageSize int) ([]domain.Inquiry, int64, error) {
	var list []domain.Inquiry
	var total int64

	q := r.db.WithContext(ctx).Model(&domain.Inquiry{}).
		Where("is_deleted = 0 AND org_id = ?", orgID)

	if keyword != "" {
		pattern := "%" + keyword + "%"
		q = q.Where("inquiry_title LIKE ?", pattern)
	}

	// 日期范围过滤
	if startDate != nil {
		q = q.Where("inquiry_date >= ?", *startDate)
	}
	if endDate != nil {
		q = q.Where("inquiry_date <= ?", *endDate)
	}

	q.Count(&total)
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 1000 {
		pageSize = 20
	}

	err := q.
		Order("inquiry_date DESC").
		Order("inquiry_title ASC").
		Limit(pageSize).Offset((page - 1) * pageSize).
		Find(&list).Error
	return list, total, err
}

func (r *inquiryRepo) UpdateInquiry(ctx context.Context, params UpdateParams) error {
	updates := map[string]any{}
	
	if params.InquiryTitle != nil {
		updates["inquiry_title"] = *params.InquiryTitle
	}
	if params.InquiryDate != nil {
		updates["inquiry_date"] = *params.InquiryDate
	}
	if params.InquiryStartDate != nil {
		updates["inquiry_start_date"] = *params.InquiryStartDate
	}
	if params.InquiryEndDate != nil {
		updates["inquiry_end_date"] = *params.InquiryEndDate
	}
	
	if params.UpdateMarket1 {
		if params.Market1 != nil {
			updates["market_1"] = *params.Market1
		} else {
			updates["market_1"] = nil
		}
	}
	if params.UpdateMarket2 {
		if params.Market2 != nil {
			updates["market_2"] = *params.Market2
		} else {
			updates["market_2"] = nil
		}
	}
	if params.UpdateMarket3 {
		if params.Market3 != nil {
			updates["market_3"] = *params.Market3
		} else {
			updates["market_3"] = nil
		}
	}

	if len(updates) == 0 {
		return nil
	}

	// 如果同时更新了开始和结束时间，需要验证
	if params.InquiryStartDate != nil && params.InquiryEndDate != nil {
		if !params.InquiryEndDate.After(*params.InquiryStartDate) {
			return errors.New("结束时间必须晚于开始时间")
		}
	} else if params.InquiryStartDate != nil || params.InquiryEndDate != nil {
		// 只更新了一个时间，需要查询另一个时间进行验证
		var existing domain.Inquiry
		err := r.db.WithContext(ctx).
			Where("id = ? AND is_deleted = 0", params.ID).
			First(&existing).Error
		if err != nil {
			return err
		}
		
		startTime := existing.InquiryStartDate
		endTime := existing.InquiryEndDate
		
		if params.InquiryStartDate != nil {
			startTime = *params.InquiryStartDate
		}
		if params.InquiryEndDate != nil {
			endTime = *params.InquiryEndDate
		}
		
		if !endTime.After(startTime) {
			return errors.New("结束时间必须晚于开始时间")
		}
	}

	return r.db.WithContext(ctx).Model(&domain.Inquiry{}).
		Where("id = ? AND is_deleted = 0", params.ID).
		Updates(updates).Error
}

func (r *inquiryRepo) SoftDeleteInquiry(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&domain.Inquiry{}).
		Where("id = ?", id).
		Update("is_deleted", 1).Error
}

func (r *inquiryRepo) HardDeleteInquiry(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("id 不能为空")
	}
	return r.db.WithContext(ctx).
		Unscoped().
		Where("id = ?", id).
		Delete(&domain.Inquiry{}).Error
}
