package utils

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

func NextDictionaryCode(tx *gorm.DB, tableName, base string) (string, error) {
	var codes []string
	if err := tx.Table(tableName).
		Select("code").
		Where("code IS NOT NULL AND code <> ''").
		Pluck("code", &codes).Error; err != nil {
		return "", err
	}

	numbers := make([]int, 0, len(codes))
	for _, c := range codes {
		if !strings.HasPrefix(c, base) {
			continue
		}
		suffix := strings.TrimPrefix(c, base)
		if suffix == "" {
			continue
		}
		n, err := strconv.Atoi(suffix)
		if err != nil {
			continue
		}
		numbers = append(numbers, n)
	}

	sort.Ints(numbers)
	expected := 1
	for _, n := range numbers {
		if n < expected {
			continue
		}
		if n == expected {
			expected++
			continue
		}
		if n > expected {
			break
		}
	}

	return fmt.Sprintf("%s%03d", base, expected), nil
}
