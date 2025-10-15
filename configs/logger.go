package configs

import "strings"

// LogConfig 日志相关配置
type LogConfig struct {
	Dir              string `json:"dir"`               // 日志目录(必填/可给默认 ./logs)
	BaseName         string `json:"base_name"`         // 基础名(默认 app)
	Level            string `json:"level"`             // debug/info/warn/error
	MaxSizeMB        int    `json:"max_size_mb"`       // 单个日志文件大小上限(MB)
	MaxFiles         int    `json:"max_files"`         // 目录内最多保留多少个文件(含分卷)
	MaxAgeDays       int    `json:"max_age_days"`      // 保留天数(超过删除)
	Compress         bool   `json:"compress"`          // 旧文件是否压缩
	StartupTimestamp bool   `json:"startup_timestamp"` // 是否每次启动新建带时间戳文件
}

type logConfigRaw struct {
	Dir              *string `json:"dir"`
	BaseName         *string `json:"base_name"`
	Level            *string `json:"level"`
	MaxSizeMB        *int    `json:"max_size_mb"`
	MaxFiles         *int    `json:"max_files"`
	MaxAgeDays       *int    `json:"max_age_days"`
	Compress         *bool   `json:"compress"`
	StartupTimestamp *bool   `json:"startup_timestamp"`
}

// 默认日志配置
var DefaultLogConfig = LogConfig{
	Dir:              "logs",
	BaseName:         "app",
	Level:            "info",
	MaxSizeMB:        50,
	MaxFiles:         10,
	MaxAgeDays:       30,
	Compress:         true,
	StartupTimestamp: true,
}

func normalLevel(p *string) string {
	if p == nil {
		return ""
	}
	l := strings.ToLower(strings.TrimSpace(*p))
	switch l {
	case "debug", "info", "warn", "error", "d", "i", "w", "e":
		switch l {
		case "d":
			return "debug"
		case "i":
			return "info"
		case "w":
			return "warn"
		case "e":
			return "error"
		default:
			return l
		}
	default:
		return ""
	}
}

func normalMaxSizeMB(v int) bool {
	if v <= 0 || v > 1024 {
		return false
	}
	return true
}

func normalMaxFiles(v int) bool {
	if v <= 0 || v > 100 {
		return false
	}
	return true
}

func normalMaxAgeDays(v int) bool {
	if v <= 0 || v > 365 {
		return false
	}
	return true
}

func mergeLog(dst *LogConfig, raw *logConfigRaw) {
	if raw == nil {
		return
	}
	if s := strPtrValid(raw.Dir); s != "" {
		dst.Dir = s
	}
	if s := strPtrValid(raw.BaseName); s != "" {
		dst.BaseName = s
	}
	if lvl := normalLevel(raw.Level); lvl != "" {
		dst.Level = lvl
	}
	if v := intPtrPos(raw.MaxSizeMB); normalMaxSizeMB(v) {
		dst.MaxSizeMB = v
	}
	if v := intPtrPos(raw.MaxFiles); normalMaxFiles(v) {
		dst.MaxFiles = v
	}
	if v := intPtrPos(raw.MaxAgeDays); normalMaxAgeDays(v) {
		dst.MaxAgeDays = v
	}
	if raw.Compress != nil {
		dst.Compress = *raw.Compress
	}
	if raw.StartupTimestamp != nil {
		dst.StartupTimestamp = *raw.StartupTimestamp
	}
}
