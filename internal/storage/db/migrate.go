// internal/storage/db/migrate.go
package foodDB

import (
	"gorm.io/gorm"
	acc "hdzk.cn/foodapp/internal/domain/account"
	category "hdzk.cn/foodapp/internal/domain/category"
	dict "hdzk.cn/foodapp/internal/domain/dict"
	market "hdzk.cn/foodapp/internal/domain/market"
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
		&market.BaseMarket{},
		&market.BasePriceInquiry{},
		&market.PriceInquiryItem{},
		&market.PriceMarketInquiry{},
		&market.PriceSupplierSettlement{},
	)
}
