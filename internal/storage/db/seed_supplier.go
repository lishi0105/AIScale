package foodDB

import (
	"context"
	"time"

	"hdzk.cn/foodapp/internal/domain/supplier"
	"hdzk.cn/foodapp/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ① 商品单位
var defaultSupplier = []struct {
	Name           string
	ContactName    string
	ContactPhone   string
	ContactAddress string
	ContactEmail   string
	FloatRatio     float64
}{
	{"贵阳恒阳食品贸易公司", "张三", "13812345678", "北京市xxxxxx路", "test1@example.com", 0.12},
	{"贵阳李四副食", "李四", "17687654321", "贵州市******路32号", "test2@example.com", 0.11},
	{"贵阳网二食品", "王二", "15114785236", "贵州市******路132号", "test3@example.com", 0.13},
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// 共用的 upsert（按 name 唯一冲突更新 sort）
func upsertSupplier(ctx context.Context, db *gorm.DB, row *supplier.Supplier) error {
	assignments := map[string]any{
		"float_ratio": row.FloatRatio,
		"is_deleted":  false, // 0
		"updated_at":  time.Now(),
	}
	if row.ContactName != nil {
		assignments["contact_name"] = *row.ContactName
	} else {
		assignments["contact_name"] = nil
	}
	if row.ContactPhone != nil {
		assignments["contact_phone"] = *row.ContactPhone
	} else {
		assignments["contact_phone"] = nil
	}
	if row.ContactEmail != nil {
		assignments["contact_email"] = *row.ContactEmail
	} else {
		assignments["contact_email"] = nil
	}
	if row.ContactAddress != nil {
		assignments["contact_address"] = *row.ContactAddress
	} else {
		assignments["contact_address"] = nil
	}

	return db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "org_id"},
				{Name: "name"},
			},
			DoUpdates: clause.Assignments(assignments),
		}).
		Create(row).Error
}

// 修正函数名和逻辑：确保默认供应商
func EnsureDefaultSupplier(ctx context.Context, gdb *gorm.DB) error {
	for _, it := range defaultSupplier {
		s := &supplier.Supplier{
			ID:             "", // GORM 会自动生成 UUID（如果你有钩子）或留空由数据库处理
			Name:           it.Name,
			OrgID:          DefaultOrgID,
			ContactName:    strPtr(it.ContactName),
			ContactPhone:   strPtr(it.ContactPhone),
			ContactEmail:   strPtr(it.ContactEmail),
			ContactAddress: strPtr(it.ContactAddress),
			FloatRatio:     it.FloatRatio,
			Sort:           0,
			Status:         1,
			Description:    "默认测试供应商",
		}
		if err := upsertSupplier(ctx, gdb, s); err != nil {
			logger.L().Error("Failed to seed supplier", zap.Error(err), zap.String("name", it.Name))
			return err
		}
	}

	logger.L().Info("Supplier seeded/ensured",
		zap.Int("count", len(defaultSupplier)),
		zap.String("org_id", DefaultOrgID),
	)
	return nil
}
