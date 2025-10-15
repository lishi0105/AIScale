package account

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	domain "hdzk.cn/foodapp/internal/domain/account"
)

type GormRepo struct{ db *gorm.DB }

func NewGorm(db *gorm.DB) *GormRepo { return &GormRepo{db: db} }

// -------- C --------

func (r *GormRepo) Create(ctx context.Context, a *domain.Account) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "username"}}, // uk_account_username
			DoUpdates: clause.Assignments(map[string]any{
				"password_hash": a.PasswordHash,
				"status":        a.Status,
				"role":          a.Role,
				"is_deleted":    0,
				"updated_at":    gorm.Expr("NOW()"),
			}),
		}).
		Create(a).Error
}

// -------- R --------
func (r *GormRepo) GetByID(ctx context.Context, id string) (*domain.Account, error) {
	var a domain.Account
	// 修复：UUID 主键必须显式 where 条件
	if err := r.db.WithContext(ctx).
		Where("id = ? AND is_deleted = 0", id).
		First(&a).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *GormRepo) GetByUsername(ctx context.Context, username string) (*domain.Account, error) {
	var a domain.Account
	if err := r.db.WithContext(ctx).
		Where("username = ? AND is_deleted = 0", username).
		First(&a).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *GormRepo) List(ctx context.Context, q domain.ListQuery) ([]domain.Account, int64, error) {
	tx := r.db.WithContext(ctx).Model(&domain.Account{}).
		Where("is_deleted = 0")
	if q.UsernameLike != "" {
		tx = tx.Where("username LIKE ?", "%"+q.UsernameLike+"%")
	}
	if q.Status != nil {
		tx = tx.Where("status = ?", *q.Status)
	}
	if q.Role != nil {
		tx = tx.Where("role = ?", *q.Role)
	}

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	limit := q.Limit
	if limit <= 0 || limit > 200 {
		limit = 20
	}
	offset := q.Offset
	if offset < 0 {
		offset = 0
	}

	var items []domain.Account
	if err := tx.
		// UUID 不适合作为时间序排序，建议按创建时间
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// -------- U --------

func (r *GormRepo) UpdatePasswordHash(ctx context.Context, id string, hash string) error {
	return r.db.WithContext(ctx).Model(&domain.Account{}).
		Where("id = ? AND is_deleted = 0", id).
		Update("password_hash", hash).Error
}

func (r *GormRepo) UpdateStatus(ctx context.Context, id string, status int) error {
	return r.db.WithContext(ctx).Model(&domain.Account{}).
		Where("id = ? AND is_deleted = 0", id).
		Update("status", status).Error
}

// -------- D --------
func (r *GormRepo) SoftDeleteByID(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&domain.Account{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"is_deleted": 1,
			"updated_at": gorm.Expr("NOW()"),
		}).Error
}

func (r *GormRepo) HardDeleteByID(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Unscoped().
		Where("id = ?", id).
		Delete(&domain.Account{}).Error
}

// 可选：行级锁
func (r *GormRepo) lockByID(ctx context.Context, id string) (*domain.Account, error) {
	var a domain.Account
	if err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND is_deleted = 0", id). // 修复这里
		First(&a).Error; err != nil {
		return nil, err
	}
	return &a, nil
}
