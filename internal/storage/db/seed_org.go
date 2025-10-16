package foodDB

import (
	"context"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"hdzk.cn/foodapp/pkg/logger"
)

const (
	DefaultOrgID          = "6f1a9b16-8e3a-4c7f-b2d3-9c0f5b8a12e4" // 你指定的固定 UUID
	DefaultOrgName        = "默认机构"
	DefaultOrgCode        = "00"
	DefaultOrgDescription = "系统初始组织（种子数据）"
)

// 映射 base_org（仅用于种子，不耦合你现有 domain）
type seedOrg struct {
	ID          string `gorm:"column:id;primaryKey;type:char(36)"`
	Name        string `gorm:"column:name"`
	Code        string `gorm:"column:code"`
	Sort        int    `gorm:"column:sort"`
	ParentID    string `gorm:"column:parent_id"`
	Description string `gorm:"column:description"`
	IsDeleted   int    `gorm:"column:is_deleted"`
}

func (seedOrg) TableName() string { return "base_org" }

// EnsureDefaultOrganization 若不存在则用固定 UUID 创建根组织；存在则直接返回
func EnsureDefaultOrganization(ctx context.Context, db *gorm.DB) error {
	// 1) 先按固定 ID 查
	var got seedOrg
	if err := db.WithContext(ctx).First(&got, "id = ?", DefaultOrgID).Error; err == nil {
		logger.L().Info("default organization exists",
			zap.String("id", got.ID), zap.String("name", got.Name), zap.String("code", got.Code))
		return nil
	}

	// 2) 若按 ID 不存在，再按 code 查（避免重复插入）
	if err := db.WithContext(ctx).First(&got, "code = ?", DefaultOrgCode).Error; err == nil {
		logger.L().Info("default organization exists(by code)",
			zap.String("id", got.ID), zap.String("name", got.Name), zap.String("code", got.Code))
		return nil
	}

	// 3) 插入（根组织 parent_id 指向自身）
	root := seedOrg{
		ID:          DefaultOrgID,
		Name:        DefaultOrgName,
		Code:        DefaultOrgCode,
		Sort:        -1,
		ParentID:    DefaultOrgID,
		Description: DefaultOrgDescription,
		IsDeleted:   0,
	}
	if err := db.WithContext(ctx).Create(&root).Error; err != nil {
		return err
	}
	logger.L().Info("created default organization",
		zap.String("id", root.ID), zap.String("name", root.Name), zap.String("code", root.Code))
	return nil
}
