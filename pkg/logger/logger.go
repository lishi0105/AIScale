// logger/logger.go
package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/mattn/go-colorable"
	"github.com/mattn/go-isatty"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"hdzk.cn/foodapp/configs"
)

var (
	global      *zap.Logger
	globalLevel zap.AtomicLevel // 统一控制文件 + 控制台
)

// isTerminal 判断是否是交互式终端（用于是否彩色）
func isTerminal() bool {
	fd := os.Stdout.Fd()
	return isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd)
}

// Init 初始化全局 Logger。
// - 文件输出：JSON + lumberjack 分卷
// - 控制台输出：Console（TTY 彩色）
// - 日志级别：默认 info，cfg.Level（大小写不敏感）
// - 目录清理：按 MaxAgeDays & MaxFiles
func Init(cfg configs.LogConfig) *zap.Logger {
	// -------- 默认值 --------
	if cfg.Dir == "" {
		cfg.Dir = "./logs"
	}
	if cfg.BaseName == "" {
		cfg.BaseName = "app"
	}
	if cfg.Level == "" {
		cfg.Level = "info"
	}
	if cfg.MaxSizeMB <= 0 {
		cfg.MaxSizeMB = 100
	}
	if cfg.MaxFiles <= 0 {
		cfg.MaxFiles = 50
	}
	if cfg.MaxAgeDays <= 0 {
		cfg.MaxAgeDays = 30
	}
	_ = os.MkdirAll(cfg.Dir, 0o755)

	// -------- 主文件名（固定/带时间戳） --------
	filename := filepath.Join(cfg.Dir, cfg.BaseName+".log")
	if cfg.StartupTimestamp {
		ts := time.Now().Format("20060102-150405")
		filename = filepath.Join(cfg.Dir, fmt.Sprintf("%s-%s.log", cfg.BaseName, ts))
	}

	// -------- 级别（默认 info；大小写不敏感；失败兜底） --------
	globalLevel = zap.NewAtomicLevel()
	lvlTxt := strings.ToLower(strings.TrimSpace(cfg.Level))
	if lvlTxt == "" {
		lvlTxt = "info"
	}
	if err := globalLevel.UnmarshalText([]byte(lvlTxt)); err != nil {
		globalLevel.SetLevel(zap.InfoLevel)
	}

	// -------- 编码器（控制台 + JSON 文件） --------
	var levelEnc zapcore.LevelEncoder
	if isTerminal() {
		levelEnc = zapcore.CapitalColorLevelEncoder
	} else {
		levelEnc = zapcore.CapitalLevelEncoder
	}

	consoleEncCfg := zapcore.EncoderConfig{
		TimeKey:       "ts",
		LevelKey:      "lv",
		CallerKey:     "caller",
		MessageKey:    "msg",
		StacktraceKey: "stack",
		EncodeLevel:   levelEnc,
		EncodeCaller:  zapcore.ShortCallerEncoder,
	}
	consoleEncCfg.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000000")
	consoleEnc := zapcore.NewConsoleEncoder(consoleEncCfg)

	jsonEncCfg := zap.NewProductionEncoderConfig()
	jsonEncCfg.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000000")
	jsonEnc := zapcore.NewJSONEncoder(jsonEncCfg)

	// -------- 文件输出（lumberjack 分卷） --------
	// 注意：MaxBackups 仅管理“当前主文件”的分卷数量；目录层面的总量由 cleanup 统一管控
	lj := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    cfg.MaxSizeMB,          // 单文件大小上限(MB)
		MaxBackups: max(1, cfg.MaxFiles/4), // 分卷数（保守取值；目录层面还有总量清理）
		MaxAge:     cfg.MaxAgeDays,         // 单文件保留天数
		Compress:   cfg.Compress,
	}
	fileWS := zapcore.AddSync(lj)

	// Windows/部分终端需要 colorable 才能显示 ANSI 颜色
	stdoutWS := zapcore.AddSync(colorable.NewColorableStdout())

	// -------- Core（统一使用 globalLevel） --------
	core := zapcore.NewTee(
		zapcore.NewCore(jsonEnc, fileWS, globalLevel),      // 文件：JSON
		zapcore.NewCore(consoleEnc, stdoutWS, globalLevel), // 控制台：Console
	)

	global = zap.New(
		core,
		zap.AddCaller(),
		zap.AddStacktrace(zap.ErrorLevel), // 仅 error+ 打栈（降低 info 噪音）
	)

	// -------- 目录级别的清理（数量 & 天数）--------
	cleanupLogDir(cfg, filename) // 启动即清理一次

	// 可选：定时清理（每小时）
	go func() {
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			cleanupLogDir(cfg, filename)
		}
	}()

	return global
}

// L 获取全局 Logger。未初始化时 panic。
func L() *zap.Logger {
	if global == nil {
		panic("logger not initialized, call logger.Init() first")
	}
	return global
}

// SetLevel 动态设置日志级别（运行时）
// 例如：logger.SetLevel(zap.DebugLevel)
func SetLevel(l zapcore.Level) {
	globalLevel.SetLevel(l)
}

// SetLevelByString 动态设置日志级别（字符串）
// 例如：logger.SetLevelByString("debug")
func SetLevelByString(level string) {
	_ = globalLevel.UnmarshalText([]byte(strings.ToLower(strings.TrimSpace(level))))
}

// GetLevel 获取当前日志级别
func GetLevel() zapcore.Level {
	return globalLevel.Level()
}

// Sync 尽量刷新缓冲输出（在程序退出时调用）
func Sync() {
	if global != nil {
		_ = global.Sync()
	}
}

// 目录级别清理：
// - 删除超过 MaxAgeDays 的文件
// - 总数量超过 MaxFiles 时，删除最早的（不含 currentFile）
func cleanupLogDir(cfg configs.LogConfig, currentFile string) {
	entries, err := os.ReadDir(cfg.Dir)
	if err != nil {
		return
	}

	type fmeta struct {
		path string
		info os.FileInfo
	}
	var files []fmeta

	// 匹配我们生成的日志文件：basename-*.log 或 basename.log 及其分卷(.1/.2/.gz)
	prefix1 := cfg.BaseName + "-"
	prefix2 := cfg.BaseName + ".log"

	now := time.Now()
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()

		// 仅处理我们的日志
		if !(strings.HasPrefix(name, prefix1) || strings.HasPrefix(name, prefix2)) {
			continue
		}
		// 只认 .log 开头的文件及其分卷/压缩
		ext := filepath.Ext(name) // 例如 ".log" / ".gz"
		if ext == "" {
			continue
		}
		// name 可能是 "app.log.1" / "app.log.2.gz" 等，这里只要含 ".log" 即可
		if !strings.Contains(name, ".log") {
			continue
		}

		fi, err := e.Info()
		if err != nil {
			continue
		}
		full := filepath.Join(cfg.Dir, name)

		// 超过天数直接删
		if cfg.MaxAgeDays > 0 {
			age := now.Sub(fi.ModTime())
			if age > (time.Duration(cfg.MaxAgeDays) * 24 * time.Hour) {
				_ = os.Remove(full)
				continue
			}
		}
		// 收集候选
		files = append(files, fmeta{path: full, info: fi})
	}

	// 按修改时间从新到旧排序
	sort.Slice(files, func(i, j int) bool {
		return files[i].info.ModTime().After(files[j].info.ModTime())
	})

	// 若数量超限，删除最旧的（保留最新的若干个；永远保留 currentFile）
	if cfg.MaxFiles > 0 && len(files) > cfg.MaxFiles {
		for _, f := range files[cfg.MaxFiles:] {
			if f.path == currentFile {
				continue
			}
			_ = os.Remove(f.path)
		}
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
