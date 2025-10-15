package foodDB

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"hdzk.cn/foodapp/configs"
)

func OpenFromConfig(c configs.DBConfig) (*gorm.DB, error) {
	// 1) 最高优先级：DSN 环境变量（如存在）
	if s := strings.TrimSpace(os.Getenv(c.DSNEnv)); c.DSNEnv != "" && s != "" {
		return openGorm(s, c)
	}

	// 2) 拼接 DSN（密码来自 env 或文件）
	pass := ""
	if c.PassEnv != "" {
		pass = os.Getenv(c.PassEnv)
	}
	if pass == "" && c.PassFile != "" {
		if b, err := os.ReadFile(c.PassFile); err == nil {
			pass = strings.TrimSpace(string(b))
		} else {
			return nil, fmt.Errorf("读取 DB 密码文件失败: %w", err)
		}
	}

	if c.Host == "" || c.Port <= 0 || c.User == "" || c.Name == "" {
		return nil, errors.New("数据库配置缺少 host/port/user/name")
	}
	params := c.Params
	if strings.TrimSpace(params) == "" {
		params = "charset=utf8mb4&parseTime=True&loc=Local"
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s",
		c.User, pass, c.Host, c.Port, c.Name, params)

	return openGorm(dsn, c)
}

func openGorm(dsn string, c configs.DBConfig) (*gorm.DB, error) {
	gdb, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
	})
	if err != nil {
		return nil, err
	}
	sqlDB, _ := gdb.DB()
	tunePool(sqlDB, c)
	return gdb, nil
}

func tunePool(sqlDB *sql.DB, c configs.DBConfig) {
	if c.MaxIdleConns >= 0 { // 允许 0
		sqlDB.SetMaxIdleConns(c.MaxIdleConns)
	}
	if c.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(c.MaxOpenConns)
	}
	if c.ConnMaxLifeS > 0 {
		sqlDB.SetConnMaxLifetime(time.Duration(c.ConnMaxLifeS) * time.Second)
	}
}

func Close(gdb *gorm.DB) error {
	if gdb == nil {
		return nil
	}
	sqlDB, err := gdb.DB()
	if err != nil {
		// 理论上不太会出错，但还是返回给上层
		return err
	}
	return sqlDB.Close()
}
