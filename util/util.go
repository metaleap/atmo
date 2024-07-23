package util

import (
	"cmp"
	"encoding/json"
	"io"
	"strings"

	"atmo/util/str"
)

func Ptr[T any](it T) *T { return &it }
func UnPtr[T any](it *T) (ret T) {
	if it != nil {
		ret = *it
	}
	return
}

func Assert(alwaysTrue bool, show any) {
	if !alwaysTrue {
		var err any = "unreachable"
		if show != nil {
			if show_fn, _ := show.(func() any); show_fn != nil {
				err = show_fn()
			} else {
				err = show
			}
		}
		panic(str.FmtV(err))
	}
}

func If[T any](b bool, t T, f T) T {
	if b {
		return t
	}
	return f
}

func IfF[T any](b bool, t func() T, f func() T) T {
	if b {
		return t()
	}
	return f()
}

func Clamp[T cmp.Ordered](min T, max T, v T) T {
	return If(v < min, min, If(v > max, max, v))
}

func Min[T cmp.Ordered](values ...T) (ret T) {
	ret = values[0]
	for _, value := range values[1:] {
		if value < ret {
			ret = value
		}
	}
	return
}

func Max[T cmp.Ordered](values ...T) (ret T) {
	ret = values[0]
	for _, value := range values[1:] {
		if value > ret {
			ret = value
		}
	}
	return
}

func ReadUntil(r io.Reader, until byte, initialBufCapacity int) ([]byte, error) {
	buf := make([]byte, 0, initialBufCapacity)
	var b [1]byte
	for {
		_, err := r.Read(b[0:1])
		if err != nil {
			return nil, err
		} else if b[0] == until {
			break
		} else {
			buf = append(buf, b[0])
		}
	}
	return buf, nil
}

func ToIdent(s string) string { return ToIdentWith(s, '_') }

func ToIdentWith(s string, replaceChar byte) string {
	s = strings.TrimSpace(s)
	var buf strings.Builder
	buf.Grow(len(s))
	next_up := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if is_ident := (c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z'); is_ident {
			if next_up && (c >= 'a' && c <= 'z') {
				c, next_up = c-32, false
			}
			buf.WriteByte(c)
		} else if replaceChar != 0 {
			buf.WriteByte(replaceChar)
		} else {
			next_up = true
		}
	}
	return buf.String()
}

func JsonAs[T any](v any) (ret T, err error) {
	json_bytes, _ := json.Marshal(v)
	err = json.Unmarshal(json_bytes, &ret)
	return
}

func JsonFrom(it any) string {
	json_bytes, _ := json.Marshal(it)
	return string(json_bytes)
}
