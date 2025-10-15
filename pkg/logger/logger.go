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

var global *zap.Logger

func isTerminal() bool {
	fd := os.Stdout.Fd()
	return isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd)
}

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

	// 主文件名：每次启动独立 or 固定名
	filename := filepath.Join(cfg.Dir, cfg.BaseName+".log")
	if cfg.StartupTimestamp {
		ts := time.Now().Format("20060102-150405")
		filename = filepath.Join(cfg.Dir, fmt.Sprintf("%s-%s.log", cfg.BaseName, ts))
	}

	// -------- 编码器：微秒时间，TTY 才彩色 --------
	var levelEnc zapcore.LevelEncoder
	if isTerminal() {
		levelEnc = zapcore.CapitalColorLevelEncoder
	} else {
		levelEnc = zapcore.CapitalLevelEncoder
	}

	consoleCfg := zapcore.EncoderConfig{
		TimeKey:       "ts",
		LevelKey:      "lv",
		CallerKey:     "caller",
		MessageKey:    "msg",
		StacktraceKey: "stack",
		EncodeLevel:   levelEnc,
		EncodeCaller:  zapcore.ShortCallerEncoder,
	}
	consoleCfg.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000000")
	consoleEnc := zapcore.NewConsoleEncoder(consoleCfg)

	jsonCfg := zap.NewProductionEncoderConfig()
	jsonCfg.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000000")
	jsonEnc := zapcore.NewJSONEncoder(jsonCfg)

	// -------- 级别 --------
	level := zapcore.InfoLevel
	_ = level.Set(cfg.Level)

	// -------- 文件滚动（按大小） --------
	// 说明：MaxBackups 仅约束“当前主文件”的分卷数量。
	// 我们把它设成 MaxFiles 的一个保守值，目录层面再全局清理。
	lj := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    cfg.MaxSizeMB,          // 单文件大小(MB)
		MaxBackups: max(1, cfg.MaxFiles/4), // 分卷数(保守); 目录层面还有总量清理
		MaxAge:     cfg.MaxAgeDays,         // 按天的清理(对单主文件有效)
		Compress:   cfg.Compress,
	}
	fileWS := zapcore.AddSync(lj)

	// Windows/部分终端需要 colorable 才能显示 ANSI 颜色
	stdoutWS := zapcore.AddSync(colorable.NewColorableStdout())

	core := zapcore.NewTee(
		zapcore.NewCore(jsonEnc, fileWS, level),                   // 文件：JSON
		zapcore.NewCore(consoleEnc, stdoutWS, zapcore.DebugLevel), // 控制台：彩色
	)

	global = zap.New(core, zap.AddCaller())

	// -------- 目录级别的清理（数量 & 天数）--------
	// 1) 启动即清理一次
	cleanupLogDir(cfg, filename)
	// 2) 定时清理（可选：每小时/每天）
	go func() {
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			cleanupLogDir(cfg, filename)
		}
	}()

	return global
}

func L() *zap.Logger {
	if global == nil {
		panic("logger not initialized, call logger.Init() first")
	}
	return global
}

// 目录级别清理：
// - 删除超过 MaxAgeDays 的文件
// - 总数量超过 MaxFiles 时，删除最早的（不含当前正在写的文件）
func cleanupLogDir(cfg configs.LogConfig, currentFile string) {
	entries, _ := os.ReadDir(cfg.Dir)
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
		if !strings.HasPrefix(filepath.Ext(name), ".log") {
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
