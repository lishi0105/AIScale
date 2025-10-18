package account

import (
	"context"
	"errors"
	"strings"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	domain "hdzk.cn/foodapp/internal/domain/account"
	"hdzk.cn/foodapp/pkg/logger"
)

type GormRepo struct{ db *gorm.DB }

// 可选：暴露一个构造器
func NewGorm(db *gorm.DB) *GormRepo { return &GormRepo{db: db} }

// -------- C --------
func (r *GormRepo) Create(ctx context.Context, a *domain.Account) error {
	if a == nil {
		return errors.New("account 不能为空")
	}
	if a.Username == "" {
		return errors.New("username 不能为空")
	}
	if a.PasswordHash == "" {
		return errors.New("password_hash 不能为空")
	}
	if a.OrgID == "" {
		return errors.New("org_id 不能为空")
	}

	logger.L().Info("creating account (upsert)",
		zap.String("id", a.ID),
		zap.String("username", a.Username),
		zap.String("org_id", a.OrgID),
		zap.Int("role", a.Role),
		zap.Int("is_deleted", a.IsDeleted),
	)

	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "username"}}, // 唯一键
			DoUpdates: clause.Assignments(map[string]any{
				"password_hash": a.PasswordHash,
				"role":          a.Role,
				"is_deleted":    0,
				// 仅当原表 org_id 为空/空串时才补齐
				"org_id": gorm.Expr("IFNULL(NULLIF(org_id, ''), ?)", a.OrgID),
				// 允许置空描述（a.Description 为 nil 时 -> NULL）
				"description": a.Description,
				"updated_at":  gorm.Expr("NOW()"),
			}),
		}).
		Create(a).Error
}

// -------- R --------

func (r *GormRepo) GetByID(ctx context.Context, id string) (*domain.Account, error) {
	var a domain.Account
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

func (r *GormRepo) List(ctx context.Context, keyword string, Deleted, Role *int, page, pageSize int) ([]domain.Account, int64, error) {
	var (
		list  []domain.Account
		total int64
	)

	q := r.db.WithContext(ctx).Model(&domain.Account{})

	// 可选条件
	if Deleted != nil {
		q = q.Where("is_deleted = ?", *Deleted)
	}
	if Role != nil {
		q = q.Where("role = ?", *Role)
	}

	// 关键字搜索：只用真实存在的列
	if kw := strings.TrimSpace(keyword); kw != "" {
		pattern := "%" + kw + "%"
		// 如果 accounts 表只有 username / description：
		q = q.Where("(username LIKE ? OR description LIKE ?)", pattern, pattern)
		// 如果没有 description，就用 username 一个字段：
		// q = q.Where("username LIKE ?", pattern)
	}

	// 分页参数
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 1000 {
		pageSize = 20
	}

	// 统计总数
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询结果（去掉不存在的 sort 排序）
	err := q.
		Order("sort ASC").
		Order("username ASC").
		Order("created_at DESC"). // 如无 created_at，则换成你表里的时间列
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&list).Error

	return list, total, err
}

// -------- U --------

func (r *GormRepo) UpdatePasswordHash(ctx context.Context, id string, hash string) error {
	if hash == "" {
		return errors.New("password_hash 不能为空")
	}
	return r.db.WithContext(ctx).Model(&domain.Account{}).
		Where("id = ? AND is_deleted = 0", id).
		Update("password_hash", hash).Error
}

func (r *GormRepo) UpdateFields(ctx context.Context, id string, fields map[string]any) error {
	if id == "" {
		return errors.New("id 不能为空")
	}
	if len(fields) == 0 {
		return errors.New("没有要更新项目")
	}
	return r.db.WithContext(ctx).
		Model(&domain.Account{}).
		Where("id = ? AND is_deleted = 0", id).
		Updates(fields).Error
}

// -------- D --------

func (r *GormRepo) SoftDelete(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("id 不能为空")
	}
	return r.db.WithContext(ctx).Model(&domain.Account{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"is_deleted": 1,
			"updated_at": gorm.Expr("NOW()"),
		}).Error
}

func (r *GormRepo) HardDelete(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("id 不能为空")
	}
	return r.db.WithContext(ctx).
		Unscoped().
		Where("id = ?", id).
		Delete(&domain.Account{}).Error
}
