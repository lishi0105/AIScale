package configs

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

// ===== 原始指针结构（用于判断字段是否出现） =====

type AppConfig struct {
	Log    LogConfig    `json:"log"`
	Server ServerConfig `json:"server"`
	DB     DBConfig     `json:"db"`
	Auth   AuthConfig   `json:"auth"`
}

var DefaultConfig = AppConfig{
	Log:    DefaultLogConfig,
	Server: DefaultServerConfig,
	DB:     DefaultDBConfig,
	Auth:   DefaultAuthConfig,
}

type appConfigRaw struct {
	Log    *logConfigRaw    `json:"log"`
	Server *serverConfigRaw `json:"server"`
	DB     *dbConfigRaw     `json:"db"`
	Auth   *authConfigRaw   `json:"auth"`
}

func LoadConfig(path string) (*AppConfig, bool, error) {
	if path == "" {
		path = "configs/config.json"
	}

	// 文件不存在：写默认
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		if err := writeJSON(path, DefaultConfig); err != nil {
			return &DefaultConfig, true, fmt.Errorf("写默认配置失败: %w", err)
		}
		return &DefaultConfig, true, nil
	}

	// 读取&解析到 raw
	b, err := os.ReadFile(path)
	if err != nil {
		return &DefaultConfig, false, fmt.Errorf("读取配置失败: %w", err)
	}
	var raw appConfigRaw
	if err := json.Unmarshal(b, &raw); err != nil {
		// 解析失败：回落默认（但不强写回，保持宽容）
		return &DefaultConfig, false, nil
	}

	// 从默认出发，逐项 merge
	cfg := DefaultConfig
	mergeLog(&cfg.Log, raw.Log)
	mergeServer(&cfg.Server, raw.Server)
	mergeDB(&cfg.DB, raw.DB)
	mergeAuth(&cfg.Auth, raw.Auth)
	writeJSON(path, cfg)
	return &cfg, false, nil
}

// 可选：把最终合并后的配置“规范化写回”，便于观察当前生效值
func SaveCanonical(path string, cfg AppConfig) error {
	return writeJSON(path, cfg)
}
