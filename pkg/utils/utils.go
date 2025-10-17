package utils

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/mozillazg/go-pinyin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func NextColoumSort(tx *gorm.DB, tableName string) (int, error) {
	// 读取现有有效记录的 sort 并加锁，避免并发插入冲突
	var sorts []int
	if err := tx.
		Table(tableName).
		Select("sort").
		Where("is_deleted = 0").
		Order("sort ASC").
		Clauses(clause.Locking{Strength: "UPDATE"}). // 相当于 FOR UPDATE
		Scan(&sorts).Error; err != nil {
		return 0, err
	}

	// 计算最小缺口：1,2,3,...，谁缺谁就是答案；如果不缺，则=最大值+1
	next := 1
	for _, s := range sorts {
		if s < next {
			continue
		}
		if s == next {
			next++
			continue
		}
		// s > next，说明 next 缺失
		break
	}
	return next, nil
}

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

// ContainsChinese checks if the string contains Chinese characters
func ContainsChinese(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Han, r) {
			return true
		}
	}
	return false
}

func GetOrgCodeAndSortByID(ctx context.Context, db *gorm.DB, id string, forUpdate bool) (code string, sort int, err error) {
	if id == "" {
		return "", 0, errors.New("id 不能为空")
	}

	// 用一个轻量的投影结构体接收
	var row struct {
		Code string
		Sort int
	}

	q := db.WithContext(ctx).
		Table("base_org").
		Select("code, sort").
		Where("id = ? AND is_deleted = 0", id)

	// 需要在事务里加锁时，传 forUpdate=true
	if forUpdate {
		q = q.Clauses(clause.Locking{Strength: "UPDATE"})
	}

	if err = q.Take(&row).Error; err != nil {
		return "", 0, err // 可能是 gorm.ErrRecordNotFound
	}
	return row.Code, row.Sort, nil
}

// GeneratePinyin generates pinyin string from Chinese characters
// Returns empty string if input is empty or contains no Chinese characters
func GeneratePinyin(s string) string {
	if s == "" || !ContainsChinese(s) {
		return ""
	}

	args := pinyin.NewArgs()
	args.Style = pinyin.Normal // 不带声调
	args.Separator = ""        // 拼音之间不加分隔符
	args.Fallback = func(r rune, a pinyin.Args) []string {
		// 非汉字字符保留原样
		return []string{string(r)}
	}

	pinyinSlice := pinyin.LazyPinyin(s, args)
	return strings.Join(pinyinSlice, "")
}
