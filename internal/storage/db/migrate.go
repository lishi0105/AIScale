// internal/storage/db/migrate.go
package foodDB

import (
	"gorm.io/gorm"
	acc "hdzk.cn/foodapp/internal/domain/account"
	category "hdzk.cn/foodapp/internal/domain/category"
	dict "hdzk.cn/foodapp/internal/domain/dict"
	organ "hdzk.cn/foodapp/internal/domain/organ"
)

func AutoMigrate(gdb *gorm.DB) error {
	return gdb.AutoMigrate(
		&organ.Organ{},
		&acc.Account{},
		&dict.Unit{},
		&dict.Spec{},
		&dict.MealTime{},
		&category.Category{},
		// 其他模型
		// 以后新增模型都放这里
	)
}
