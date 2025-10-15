// ===== 小工具：校验/归一化 =====
package configs

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

func strPtrValid(p *string) string {
	if p == nil {
		return ""
	}
	s := strings.TrimSpace(*p)
	if s == "" {
		return ""
	}
	return s
}

func intPtrPos(p *int) int {
	if p == nil || *p <= 0 {
		return 0
	}
	return *p
}

func intPtrInRange(p *int, lo, hi int) int {
	if p == nil {
		return 0
	}
	if *p < lo || *p > hi {
		return 0
	}
	return *p
}

// 原子写（简单版）
func writeJSON(path string, v any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
