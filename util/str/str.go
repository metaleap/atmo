package str

import (
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

type Buf = strings.Builder
type Dict = map[string]string

var (
	ReflType = reflect.TypeOf("")
	Has      = strings.Contains
	Begins   = strings.HasPrefix
	Ends     = strings.HasSuffix
	Trim     = strings.TrimSpace
	TrimPref = strings.TrimPrefix
	TrimSuff = strings.TrimSuffix
	Idx      = strings.IndexByte
	IdxSub   = strings.Index
	IdxLast  = strings.LastIndexByte
	IdxRune  = strings.IndexRune
	Join     = strings.Join
	Split    = strings.Split
	Cut      = strings.Cut
	FromInt  = strconv.Itoa
	FromBool = strconv.FormatBool
	FromI64  = strconv.FormatInt
	FromU64  = strconv.FormatUint
	ToInt    = strconv.Atoi
	ToI64    = strconv.ParseInt
	ToU64    = strconv.ParseUint
	ToF      = strconv.ParseFloat
	Fmt      = fmt.Sprintf
	Q        = strconv.Quote
	FromQ    = strconv.Unquote
	Lo       = strings.ToLower
	Up       = strings.ToUpper
	Repeat   = strings.Repeat
)

func FmtV(v any) string   { return Fmt("%v", v) }
func GoLike(v any) string { return Fmt("%#v", v) }
func Base36(i int) string { return FromI64(int64(i), 36) }
func FromFloat(f float64, prec int) string {
	ret := strconv.FormatFloat(f, 'f', prec, 64)
	if (prec < 0) && Idx(ret, '.') < 0 {
		ret += ".0"
	}
	return ret
}

func Shorter(s1 string, s2 string) (string, string) {
	if len(s2) < len(s1) {
		return s2, s1
	}
	return s1, s2
}

func Shorten(s string, lenMax int) string {
	if len(s) > lenMax {
		s = s[:lenMax] + "..."
	}
	return s
}

func Replace(s string, repl Dict) string {
	replacer := Replacer(s, repl)
	if replacer == nil {
		return s
	}
	return replacer.Replace(s)
}

func Replacer(s string, repl Dict) *strings.Replacer {
	if len(repl) == 0 {
		return nil
	}
	repl_old_new := make([]string, 0, len(repl)*2)
	for k, v := range repl {
		repl_old_new = append(repl_old_new, k, v)
	}
	return strings.NewReplacer(repl_old_new...)
}

func RePrefix(s string, oldPrefix string, newPrefix string) string {
	return newPrefix + TrimPref(s, oldPrefix)
}

func ReSuffix(s string, oldSuffix string, newSuffix string) string {
	return TrimSuff(s, oldSuffix) + newSuffix
}

func DurationMs(nanos int64) string {
	ms := float64(nanos) * 0.000001
	return FromFloat(ms, 2) + "ms"
}

func IsLo(s string) bool {
	for _, r := range s {
		if !unicode.IsLower(r) {
			return false
		}
	}
	return true
}

func IsUp(s string) bool {
	for _, r := range s {
		if !unicode.IsUpper(r) {
			return false
		}
	}
	return true
}

func IsPrtAscii(s string) bool {
	for i := 0; i < len(s); i++ {
		if (s[i] < 0x20) || (s[i] > 0x7e) {
			return false
		}
	}
	return true
}

// ascii only
func Lo0(s string) string {
	if (s == "") || (s[0] < 'A' || s[0] > 'Z') {
		return s
	}
	return Lo(s[:1]) + s[1:]
}

// ascii only
func Up0(s string) string {
	if (s == "") || ((s[0] >= 'A') && (s[0] <= 'Z')) {
		return s
	}
	return Up(s[:1]) + s[1:]
}

func Sub(s string, runeIdx int, runesLen int) string {
	if (s == "") || (runesLen == 0) || (runeIdx < 0) {
		return ""
	}
	rune_idx, idxStart, idxEnd := 0, -1, -1
	for i := range s { // iter by runes
		if rune_idx == runeIdx {
			if idxStart = i; runesLen < 0 {
				break
			}
		} else if (idxStart >= 0) && ((rune_idx - idxStart) == runesLen) {
			idxEnd = i
			break
		}
		rune_idx++
	}
	if idxStart < 0 {
		return ""
	} else if (runesLen < 0) || (idxEnd < idxStart) {
		return s[idxStart:]
	}
	return s[idxStart:idxEnd]
}

// whether `str` matches at least _@_._
func IsEmailishEnough(str string) bool {
	l, idx_at, idx_last_dot := len(str), Idx(str, '@'), IdxLast(str, '.')
	return (l >= 5) && (l <= 255) && (idx_at > 0) && (idx_at < l-1) && (idx_at <= 64) && (idx_at == IdxLast(str, '@') &&
		(idx_last_dot > idx_at) && (idx_last_dot < l-1))
}

func In[T ~string](str T, set ...T) bool {
	for i := range set {
		if set[i] == str {
			return true
		}
	}
	return false
}

func Repl(str string, namedReplacements Dict) string {
	if (len(namedReplacements) == 0) || (len(str) == 0) {
		return str
	}
	new_len := len(str)
	for k, v := range namedReplacements {
		new_len -= (len(k) + 2)
		new_len += len(v)
	}
	var buf Buf
	if new_len > 0 {
		buf.Grow(new_len)
	}

	var skip_until, accum_from int
	var accum string
	for i := 0; i < len(str); i++ {
		if skip_until > i {
			continue
		} else if str[i] == '{' {
			if idx := i + Idx(str[i:], '}'); idx > i {
				name := str[i+1 : idx]
				if repl, exists := namedReplacements[name]; exists {
					_, _ = buf.WriteString(accum)
					_, _ = buf.WriteString(repl)
					skip_until = idx + 1
					accum_from, accum = skip_until, ""
					continue
				}
			}
		}
		accum = str[accum_from : i+1]
	}
	_, _ = buf.WriteString(accum)
	return buf.String()
}

func AsciiRand(minLen int, maxLen int) (ret string) {
	max := big.NewInt(math.MaxInt64)
	for len(ret) < minLen {
		big, err := rand.Int(rand.Reader, max)
		if err != nil {
			panic(err)
		}
		ret += FromI64(big.Int64(), 36)
	}
	if (maxLen > 0) && (len(ret) > maxLen) {
		ret = ret[:maxLen]
	}
	return
}
