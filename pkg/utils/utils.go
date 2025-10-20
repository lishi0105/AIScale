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

func NextSortSuffix(tx *gorm.DB, table_name, orgID string, base int, forUpdate bool) (int, error) {
	type rec struct{ Sort int }
	var rows []rec

	// 不在 SELECT 里用占位，直接选出 sort 再在 Go 里计算 suffix，避免参数计数错乱
	q := tx.Table(table_name).
		Select("sort").
		Where(`
			org_id = ?
			AND is_deleted = 0
			AND sort > ? AND sort <= ?`,
			orgID, base, base+999).
		Order("sort ASC")
	if forUpdate {
		q = q.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if err := q.Scan(&rows).Error; err != nil {
		return 0, fmt.Errorf("扫描 sort 失败: %w", err)
	}

	next := 1
	for _, r := range rows {
		suffix := r.Sort - base
		if suffix < next {
			continue
		}
		if suffix == next {
			next++
			continue
		}
		break
	}
	if next > 999 {
		return 0, fmt.Errorf("该 org 的 sort 段已满（1..999）")
	}
	return next, nil
}

// code 段的最小缺口：仅统计本 organ 形如 <orgCode><三位数字>
func NextCodeSuffixByPrefix(tx *gorm.DB, table_name, orgID, orgCode string, forUpdate bool) (int, error) {
	type rec struct{ Suffix int }
	var rows []rec

	// 用 SUBSTRING(code, ? + 1) + REGEXP/LIKE；以 len(orgCode) 作为位置参数，避免 CHAR_LENGTH(?) 触发计数差异
	prefixLen := len(orgCode)

	q := tx.Table(table_name).
		Select("CAST(SUBSTRING(code, ? + 1) AS UNSIGNED) AS suffix", prefixLen).
		Where(`
			org_id = ?
			AND is_deleted = 0
			AND code LIKE CONCAT(?, '___')
			AND code REGEXP CONCAT('^', ?, '[0-9]{3}$')`,
			orgID, orgCode, orgCode).
		Order("suffix ASC")
	if forUpdate {
		q = q.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if err := q.Scan(&rows).Error; err != nil {
		return 0, fmt.Errorf("扫描 code 后缀失败: %w", err)
	}

	next := 1
	for _, r := range rows {
		if r.Suffix < next {
			continue
		}
		if r.Suffix == next {
			next++
			continue
		}
		break
	}
	if next > 999 {
		return 0, fmt.Errorf("该 org 的 code 段已满（001..999）")
	}
	return next, nil
}

// GeneratePinyin generates pinyin string from Chinese characters
// Returns empty string if input is empty or contains no Chinese characters
func GeneratePinyin(s string) string {
	if s == "" || !ContainsChinese(s) {
		return ""
	}

	args := pinyin.NewArgs()
	args.Style = pinyin.FirstLetter // 不带声调
	args.Separator = ""             // 拼音之间不加分隔符
	args.Fallback = func(r rune, a pinyin.Args) []string {
		// 非汉字字符保留原样
		return []string{string(r)}
	}

	pinyinSlice := pinyin.LazyPinyin(s, args)
	return strings.Join(pinyinSlice, "")
}
