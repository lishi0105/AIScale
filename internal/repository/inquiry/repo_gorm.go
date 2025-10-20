package inquiry

import (
    "context"
    "time"

    "gorm.io/gorm"
    domain "hdzk.cn/foodapp/internal/domain/inquiry"
)

type repo struct{ db *gorm.DB }

func (r *repo) Create(ctx context.Context, m *domain.PriceInquiry) error {
    return r.db.WithContext(ctx).Create(m).Error
}

func (r *repo) Get(ctx context.Context, id string) (*domain.PriceInquiry, error) {
    var out domain.PriceInquiry
    err := r.db.WithContext(ctx).Where("id = ? AND is_deleted = 0", id).First(&out).Error
    if err != nil {
        return nil, err
    }
    return &out, nil
}

func (r *repo) List(ctx context.Context, orgID string, keyword string, dateFrom, dateTo *time.Time, page, pageSize int) ([]domain.PriceInquiry, int64, error) {
    var list []domain.PriceInquiry
    var total int64

    q := r.db.WithContext(ctx).Model(&domain.PriceInquiry{}).
        Where("is_deleted = 0 AND org_id = ?", orgID)

    if keyword != "" {
        like := "%" + keyword + "%"
        q = q.Where("inquiry_title LIKE ?", like)
    }
    if dateFrom != nil {
        q = q.Where("inquiry_date >= ?", *dateFrom)
    }
    if dateTo != nil {
        q = q.Where("inquiry_date <= ?", *dateTo)
    }

    q.Count(&total)
    if page < 1 {
        page = 1
    }
    if pageSize <= 0 || pageSize > 1000 {
        pageSize = 20
    }

    err := q.Order("inquiry_date DESC").
        Limit(pageSize).Offset((page-1)*pageSize).
        Find(&list).Error
    return list, total, err
}

func (r *repo) Update(ctx context.Context, params UpdateParams) error {
    updates := map[string]any{}
    if params.InquiryTitle != nil {
        updates["inquiry_title"] = *params.InquiryTitle
    }
    if params.InquiryDate != nil {
        updates["inquiry_date"] = *params.InquiryDate
    }
    if params.Market1 != nil {
        updates["market_1"] = *params.Market1
    }
    if params.Market2 != nil {
        updates["market_2"] = *params.Market2
    }
    if params.Market3 != nil {
        updates["market_3"] = *params.Market3
    }
    if params.InquiryStartDate != nil {
        updates["inquiry_start_date"] = *params.InquiryStartDate
    }
    if params.InquiryEndDate != nil {
        updates["inquiry_end_date"] = *params.InquiryEndDate
    }
    if len(updates) == 0 {
        return nil
    }
    return r.db.WithContext(ctx).Model(&domain.PriceInquiry{}).
        Where("id = ? AND is_deleted = 0", params.ID).
        Updates(updates).Error
}

func (r *repo) SoftDelete(ctx context.Context, id string) error {
    return r.db.WithContext(ctx).Model(&domain.PriceInquiry{}).
        Where("id = ?", id).Update("is_deleted", 1).Error
}

func (r *repo) HardDelete(ctx context.Context, id string) error {
    return r.db.WithContext(ctx).Unscoped().
        Where("id = ?", id).Delete(&domain.PriceInquiry{}).Error
}
