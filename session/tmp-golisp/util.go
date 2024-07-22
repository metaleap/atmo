package main

import (
	"fmt"
	"io"
	"os"
)

func isNilOrFalse(expr Expr) bool {
	return isEq(exprNil, expr) || isEq(exprFalse, expr)
}

func readUntil(r io.Reader, until byte, initialBufCapacity int) (string, error) {
	buf := make([]byte, 0, initialBufCapacity)
	var b [1]byte
	for {
		_, err := r.Read(b[0:1])
		if err != nil {
			return "", err
		} else if b[0] == until {
			break
		} else {
			buf = append(buf, b[0])
		}
	}
	line := string(buf)
	return line, nil
}

func checkArgsCount(wantAtLeast int, wantAtMost int, name string, have []Expr) error {
	if wantAtLeast < 0 {
		return nil
	} else if want_exactly := wantAtLeast; (want_exactly == wantAtMost) && (want_exactly != len(have)) {
		return fmt.Errorf("%s expects %d arg(s), not %d", name, want_exactly, len(have))
	} else if len(have) < wantAtLeast {
		return fmt.Errorf("%s expects at least %d arg(s), not %d", name, wantAtLeast, len(have))
	} else if (wantAtMost > wantAtLeast) && (len(have) > wantAtMost) {
		return fmt.Errorf("%s expects %d to %d arg(s), not %d", name, wantAtLeast, wantAtMost, len(have))
	}
	return nil
}

func checkIs[T Expr](have Expr) (T, error) {
	ret, ok := have.(T)
	if !ok {
		return ret, fmt.Errorf("expected %T, not %T", ret, have)
	}
	return ret, nil
}

func checkAre[T Expr](have ...Expr) error {
	for _, expr := range have {
		if _, err := checkIs[T](expr); err != nil {
			return err
		}
	}
	return nil
}

func checkAreBoth[T1 Expr, T2 Expr](have []Expr, exactArgsCount bool) (ret1 T1, ret2 T2, err error) {
	max_args_count := -1
	if exactArgsCount {
		max_args_count = 2
	}
	if err = checkArgsCount(2, max_args_count, "operator", have); err != nil {
		return
	}
	if ret1, err = checkIs[T1](have[0]); err != nil {
		return
	}
	if ret2, err = checkIs[T2](have[1]); err != nil {
		return
	}
	return
}

func checkIsStrOrKeyword(expr Expr) (string, Expr, error) {
	switch it := expr.(type) {
	case ExprStr:
		return string(it), it, nil
	case ExprKeyword:
		return string(it), it, nil
	}
	return "", nil, fmt.Errorf("expected string or keyword, not `%s`", str(true, expr))
}

func checkIsSeq(expr Expr) ([]Expr, error) {
	switch expr := expr.(type) {
	case ExprList:
		return ([]Expr)(expr), nil
	case ExprVec:
		return ([]Expr)(expr), nil
	default:
		return nil, fmt.Errorf("expected list or vector, not %T", expr)
	}
}

func newErrNotCallable(expr Expr) error {
	return fmt.Errorf("not callable: `%s`", str(true, expr))
}

func addOsArgsToEnv() {
	if len(os.Args) > 1 {
		args := make(ExprList, 0, len(os.Args)-2)
		for _, arg := range os.Args[2:] {
			args = append(args, ExprStr(arg))
		}
		envMain.set("osArgs", args)
	}
}
