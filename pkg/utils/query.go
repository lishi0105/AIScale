// query/opt.go
package utils

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// QueryOptional 统一封装：从 query 中取 key，解析为 T。
// 返回：(*T 值; provided 是否传入该 key; err 解析/校验错误)
// 约定：provided=true 且 val==nil 表示“传了空字符串”。
func QueryOptional[T any](c *gin.Context, key string, parse func(string) (T, error), validate func(T) error) (*T, bool, error) {
	raw, exists := c.GetQuery(key)
	if !exists {
		return nil, false, nil // 未传
	}
	if strings.TrimSpace(raw) == "" {
		return nil, true, nil // 传了空
	}
	v, err := parse(raw)
	if err != nil {
		return nil, true, fmt.Errorf("invalid %s: %w", key, err)
	}
	if validate != nil {
		if err := validate(v); err != nil {
			return nil, true, fmt.Errorf("invalid %s: %w", key, err)
		}
	}
	return &v, true, nil
}

// -------- 具体类型便捷封装 --------

// QueryOptionalInt 解析为 *int；validate 可为 nil，或用于范围/枚举校验。
func GetQueryIntPointer(c *gin.Context, key string) (*int, error) {
	raw := c.Query(key)
	if strings.TrimSpace(raw) == "" {
		return nil, nil // 未传或空字符串 -> nil
	}

	v, err := strconv.Atoi(raw)
	if err != nil {
		return nil, err
	}
	return &v, nil
}
