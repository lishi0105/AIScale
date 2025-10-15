package security

import (
	"fmt"
	"strings"
	"unicode"
)

// Validate 返回所有违反项；若切片为空则表示通过
// 口径：只允许 A-Z a-z 0-9 以及 asciiSpecials 中的字符（不含空格）。
func Validate(pw string) []string {
	var errs []string

	// 0) 白名单检查：发现不支持字符则直接报错
	if bad := unsupportedRunesASCII(pw); len(bad) > 0 {
		// 展示前几个非法字符，便于前端提示
		var samples []string
		maxShow := 3
		for i := 0; i < len(bad) && i < maxShow; i++ {
			// 打印出字符本身与位置（从0计）
			samples = append(samples, fmt.Sprintf("%q@%d", bad[i].R, bad[i].Pos))
		}
		msg := "包含不支持的字符（只允许英文字母、数字及以下符号：" + asciiSpecials + "）"
		if len(samples) > 0 {
			msg += "，示例：" + strings.Join(samples, ", ")
		}
		errs = append(errs, msg)
		return errs
	}

	// 1) 长度
	if len([]rune(pw)) < 8 {
		errs = append(errs, "长度至少 8 位")
	}
	if len([]rune(pw)) > 64 {
		errs = append(errs, "长度不得超过 64 位")
	}

	// 2) 至少三类（大写/小写/数字/特殊）
	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, r := range pw {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		default:
			if isSpecialASCII(r) {
				hasSpecial = true
			}
		}
	}
	categories := 0
	for _, ok := range []bool{hasUpper, hasLower, hasDigit, hasSpecial} {
		if ok {
			categories++
		}
	}
	if categories < 3 {
		errs = append(errs, "必须至少包含三类字符：大写、小写、数字、特殊字符")
	}

	// 3a) 连续相同字符（长度≥3）
	if hasTripleRepeat(pw) {
		errs = append(errs, "不能包含连续重复字符（如 aaa、111）")
	}

	// 3b) 单调序列（长度≥3）：abc/ABC/123 及其逆序 cba/321
	if hasMonotonicSequence(pw) {
		errs = append(errs, "不能包含明显顺序或逆序序列（如 abc、cba、123、321）")
	}

	return errs
}

// 只允许的 ASCII 特殊字符集合（不包含空格）
const asciiSpecials = `!"#$%&'()*+,-./:;<=>?@[\]^_` + "`" + `{|}~`

// 仅允许：A-Z / a-z / 0-9 / asciiSpecials；其他一律视为不支持
func unsupportedRunesASCII(s string) []struct {
	Pos int
	R   rune
} {
	var bad []struct {
		Pos int
		R   rune
	}
	pos := 0
	for _, r := range s {
		allowed := unicode.IsLetter(r) || unicode.IsDigit(r) || isSpecialASCII(r)
		if !allowed {
			bad = append(bad, struct {
				Pos int
				R   rune
			}{Pos: pos, R: r})
		}
		pos++
	}
	return bad
}

func isSpecialASCII(r rune) bool {
	// 仅可见 ASCII 段（33..126），排除空格等
	if r < 33 || r > 126 {
		return false
	}
	return strings.ContainsRune(asciiSpecials, r)
}

func hasTripleRepeat(s string) bool {
	var last rune
	run := 0
	first := true
	for _, r := range s {
		if first {
			first = false
			last = r
			run = 1
			continue
		}
		if r == last {
			run++
			if run >= 3 {
				return true
			}
		} else {
			last = r
			run = 1
		}
	}
	return false
}

// 仅对“连续的字母序列”或“连续的数字序列”检测单调 +/-1；字母大小写不敏感
func hasMonotonicSequence(s string) bool {
	// 把字母统一成小写处理，便于 A→B 检测
	rs := []rune(strings.ToLower(s))

	// 针对连续的字母序列与数字序列分别滚动窗口
	checkSeq := func(isSameClass func(rune) bool) bool {
		n := len(rs)
		i := 0
		for i < n {
			// 找一段同类（全字母或全数字）的连续区间 [i, j)
			if !isSameClass(rs[i]) {
				i++
				continue
			}
			j := i + 1
			for j < n && isSameClass(rs[j]) {
				j++
			}
			// 区间长度 >= 3 才有意义
			if j-i >= 3 {
				if isMonotonicStepOne(rs[i:j]) {
					return true
				}
			}
			i = j
		}
		return false
	}

	isAlpha := func(r rune) bool { return unicode.IsLetter(r) }
	isDigit := func(r rune) bool { return unicode.IsDigit(r) }

	return checkSeq(isAlpha) || checkSeq(isDigit)
}

// 判断切片是否存在长度≥3的连续子串满足相邻差值恒为 +1 或 -1
func isMonotonicStepOne(arr []rune) bool {
	if len(arr) < 3 {
		return false
	}
	// 以滑动法寻找最长单调段
	start := 0
	for i := 1; i < len(arr); i++ {
		diff := int(arr[i]) - int(arr[i-1])
		if diff != 1 && diff != -1 {
			// 断开
			if i-start >= 3 {
				return true
			}
			start = i
		}
	}
	return len(arr)-start >= 3
}
