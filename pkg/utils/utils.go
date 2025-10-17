package utils

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode"

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

// GeneratePinyin generates pinyin for Chinese characters
// This is a simple implementation that converts Chinese characters to pinyin
// For production use, consider using a more robust pinyin library
func GeneratePinyin(text string) string {
	if text == "" {
		return ""
	}
	
	// Check if the text contains Chinese characters
	hasChinese := false
	for _, r := range text {
		if unicode.Is(unicode.Scripts["Han"], r) {
			hasChinese = true
			break
		}
	}
	
	if !hasChinese {
		return ""
	}
	
	// Simple pinyin mapping for common characters
	// In production, you should use a proper pinyin library like github.com/mozillazg/go-pinyin
	pinyinMap := map[rune]string{
		'食': "shi",
		'品': "pin", 
		'类': "lei",
		'别': "bie",
		'蔬': "shu",
		'菜': "cai",
		'水': "shui",
		'果': "guo",
		'肉': "rou",
		'海': "hai",
		'鲜': "xian",
		'调': "tiao",
		'料': "liao",
		'饮': "yin",
		'主': "zhu",
		'副': "fu",
		'零': "ling",
		'甜': "tian",
		'点': "dian",
		'面': "mian",
		'包': "bao",
		'米': "mi",
		'饭': "fan",
		'条': "tiao",
		'汤': "tang",
		'粥': "zhou",
		'豆': "dou",
		'制': "zhi",
		'奶': "nai",
		'蛋': "dan",
		'油': "you",
		'脂': "zhi",
		'味': "wei",
		'糖': "tang",
		'坚': "jian",
		'种': "zhong",
		'子': "zi",
		'茶': "cha",
		'叶': "ye",
		'咖': "ka",
		'啡': "fei",
		'酒': "jiu",
		'保': "bao",
		'健': "jian",
		'营': "ying",
		'养': "yang",
		'补': "bu",
		'充': "chong",
		'剂': "ji",
		'其': "qi",
		'他': "ta",
	}
	
	var result strings.Builder
	for _, r := range text {
		if pinyin, exists := pinyinMap[r]; exists {
			result.WriteString(pinyin)
		} else if unicode.IsLetter(r) || unicode.IsDigit(r) {
			// Keep non-Chinese characters as is
			result.WriteRune(r)
		}
	}
	
	return result.String()
}

// IsChineseText checks if the text contains Chinese characters
func IsChineseText(text string) bool {
	for _, r := range text {
		if unicode.Is(unicode.Scripts["Han"], r) {
			return true
		}
	}
	return false
}
