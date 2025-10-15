package configs

import (
	"os"
	"strings"
)

// DBConfig 非机密走配置；机密通过 env/file 注入
type DBConfig struct {
	Host   string `json:"host"`   // 例如 127.0.0.1
	Port   int    `json:"port"`   // 3306
	Name   string `json:"name"`   // 数据库名
	User   string `json:"user"`   // 用户名（非机密）
	Params string `json:"params"` // 连接参数，如 charset=utf8mb4&parseTime=True&loc=Local

	// 机密与直连优先级（高→低）：DSNEnv > PassEnv > PassFile > 空密码
	DSNEnv   string `json:"dsn_env"`   // 如 "DB_DSN"，若设置且存在则直接使用
	PassEnv  string `json:"pass_env"`  // 如 "DB_PASS"
	PassFile string `json:"pass_file"` // 如 "/run/secrets/db_pass"

	// 连接池
	MaxOpenConns int `json:"max_open_conns"`
	MaxIdleConns int `json:"max_idle_conns"`
	ConnMaxLifeS int `json:"conn_max_life_seconds"`
}

type dbConfigRaw struct {
	Host         *string `json:"host"`
	Port         *int    `json:"port"`
	Name         *string `json:"name"`
	User         *string `json:"user"`
	Params       *string `json:"params"`
	DSNEnv       *string `json:"dsn_env"`
	PassEnv      *string `json:"pass_env"`
	PassFile     *string `json:"pass_file"`
	MaxOpenConns *int    `json:"max_open_conns"`
	MaxIdleConns *int    `json:"max_idle_conns"`
	ConnMaxLifeS *int    `json:"conn_max_life_seconds"`
}

var DefaultDBConfig = DBConfig{
	Host:         "172.16.66.35",
	Port:         3306,
	Name:         "main",
	User:         "food_user",
	Params:       "charset=utf8mb4&parseTime=True&loc=Local",
	DSNEnv:       "DB_DSN",
	PassEnv:      "DB_PASS",
	PassFile:     "",
	MaxOpenConns: 100,
	MaxIdleConns: 10,
	ConnMaxLifeS: 3600,
}

// —— 合并（出现且合法则覆盖默认）——

func mergeDB(dst *DBConfig, raw *dbConfigRaw) {
	if raw == nil {
		return
	}
	if s := strPtrNonEmpty(raw.Host); s != "" {
		dst.Host = s
	}
	if v := intPtrInRange(raw.Port, 1, 65535); v > 0 {
		dst.Port = v
	}
	if s := strPtrNonEmpty(raw.Name); s != "" {
		dst.Name = s
	}
	if s := strPtrNonEmpty(raw.User); s != "" {
		dst.User = s
	}
	if s := strPtrNonEmpty(raw.Params); s != "" {
		dst.Params = s
	}
	if s := strPtrNonEmpty(raw.DSNEnv); s != "" {
		dst.DSNEnv = s
	}
	if s := strPtrNonEmpty(raw.PassEnv); s != "" {
		dst.PassEnv = s
	}
	if s := strPtrNonEmpty(raw.PassFile); s != "" && fileReadable(s) {
		dst.PassFile = s
	}
	if v := intPtrPositive(raw.MaxOpenConns); v > 0 {
		dst.MaxOpenConns = v
	}
	if v := intPtrNonNegative(raw.MaxIdleConns); v >= 0 {
		dst.MaxIdleConns = v
	}
	if v := intPtrNonNegative(raw.ConnMaxLifeS); v >= 0 {
		dst.ConnMaxLifeS = v
	}
}

// —— 本文件内小工具（避免与其他文件重名）——

func strPtrNonEmpty(p *string) string {
	if p == nil {
		return ""
	}
	s := strings.TrimSpace(*p)
	if s == "" {
		return ""
	}
	return s
}

func intPtrPositive(p *int) int {
	if p == nil || *p <= 0 {
		return 0
	}
	return *p
}

func intPtrNonNegative(p *int) int {
	if p == nil || *p < 0 {
		return -1
	}
	return *p
}

func fileReadable(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
