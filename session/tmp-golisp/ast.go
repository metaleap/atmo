package main

import (
	"cmp"
	"fmt"
	"reflect"
	"strings"
)

var (
	exprTrue  = ExprKeyword(":true")
	exprFalse = ExprKeyword(":false")
	exprNil   = ExprKeyword(":nil")
)

type Expr interface {
	isExpr()
}

func (ExprIdent) isExpr()   {}
func (ExprKeyword) isExpr() {}
func (ExprStr) isExpr()     {}
func (ExprNum) isExpr()     {}
func (ExprList) isExpr()    {}
func (ExprVec) isExpr()     {}
func (ExprHashMap) isExpr() {}
func (*ExprAtom) isExpr()   {}
func (ExprErr) isExpr()     {}
func (ExprFunc) isExpr()    {}
func (*ExprFn) isExpr()     {}

type ExprIdent string
type ExprKeyword string
type ExprStr string
type ExprNum int
type ExprList []Expr
type ExprVec []Expr
type ExprHashMap map[string]Expr
type ExprAtom struct{ Ref Expr }
type ExprErr struct{ It any }
type ExprFunc func([]Expr) (Expr, error)
type ExprFn struct { // if it weren't for TCO, just the above `ExprFunc` would suffice.
	params     []Expr // all are guaranteed to be `ExprIdent` before constructing an `ExprFn`
	body       Expr
	env        *Env
	isMacro    bool
	isVariadic bool
	nameMaybe  string
}

func (me *ExprFn) envWith(args []Expr) (*Env, error) {
	num_args_min, num_args_max := len(me.params), len(me.params)
	if me.isVariadic {
		num_args_min, num_args_max = len(me.params)-1, -1
	}
	if err := checkArgsCount(num_args_min, num_args_max, strings.TrimSpace("function "+me.nameMaybe), args); err != nil {
		return nil, err
	}
	if me.isVariadic {
		the_var_args := args[len(me.params)-1:]
		args = append(append(make([]Expr, 0, len(me.params)-1+len(the_var_args)), args[:len(me.params)-1]...), (ExprList)(the_var_args))
	}
	return newEnv(me.env, me.params, args), nil
}

// note, `(*ExprFn).Call` is itself an `ExprFunc`
func (me *ExprFn) Call(args []Expr) (Expr, error) {
	env, err := me.envWith(args)
	if err != nil {
		return nil, err
	}
	return evalAndApply(env, me.body)
}

func (me ExprErr) Error() string {
	if err, _ := me.It.(error); err != nil {
		return err.Error()
	}
	if expr, _ := me.It.(Expr); expr != nil {
		return str(false, expr)
	}
	return fmt.Sprintf("%#v", me.It)
}

func exprBool(b bool) ExprKeyword {
	if b {
		return exprTrue
	}
	return exprFalse
}

func exprStrOrKeyword(s string) Expr {
	if (s != "") && (s[0] == ':') {
		return ExprKeyword(s)
	}
	return ExprStr(s)
}

func compare(args []Expr) (int, error) {
	if err := checkArgsCount(2, 2, "comparer", args); err != nil {
		return 0, err
	}
	switch it := args[0].(type) {
	case ExprNum:
		if other, ok := args[1].(ExprNum); ok {
			return cmp.Compare(it, other), nil
		}
	case ExprStr:
		if other, ok := args[1].(ExprStr); ok {
			return cmp.Compare(it, other), nil
		}
	}
	return 0, fmt.Errorf("specified operands `%#v` and `%#v` are not comparable", args[0], args[1])
}

func isListOrVec(seq Expr) bool {
	ty := reflect.TypeOf(seq)
	return (ty == reflect.TypeFor[ExprList]()) || (ty == reflect.TypeFor[ExprVec]())
}

func isListStartingWithIdent(maybeList Expr, ident ExprIdent, mustHaveLen int) (list []Expr, doesListStartWithIdent bool, err error) {
	if list, _ = maybeList.(ExprList); len(list) > 0 {
		if maybe_ident, _ := list[0].(ExprIdent); maybe_ident == ident {
			doesListStartWithIdent, err = true, checkArgsCount(mustHaveLen, mustHaveLen, "form `"+string(ident)+"`", list)
		}
	}
	return
}

func isEq(arg1 Expr, arg2 Expr) bool {
	ty1, ty2 := reflect.TypeOf(arg1), reflect.TypeOf(arg2)
	if (ty1 != ty2) && ((!isListOrVec(arg1)) || !isListOrVec(arg2)) {
		return false
	}
	switch arg1.(type) {
	case ExprErr: // TODO: not fully correct this way:
		return ((reflect.TypeOf(arg1.(ExprErr).It) == reflect.TypeOf(arg2.(ExprErr).It)) && (arg1.(ExprErr).Error() == arg2.(ExprErr).Error()))
	case *ExprAtom:
		return arg1.(*ExprAtom).Ref == arg2.(*ExprAtom).Ref
	case ExprVec, ExprList:
		sl1, _ := checkIsSeq(arg1)
		sl2, _ := checkIsSeq(arg2)
		if len(sl1) != len(sl2) {
			return false
		}
		for i := 0; i < len(sl1); i += 1 {
			if !isEq(sl1[i], sl2[i]) {
				return false
			}
		}
		return true
	case ExprHashMap:
		hm1, hm2 := arg1.(ExprHashMap), arg2.(ExprHashMap)
		if len(hm1) != len(hm2) {
			return false
		}
		for k, v := range hm1 {
			if !isEq(v, hm2[k]) {
				return false
			}
		}
		return true
	default:
		return arg1 == arg2
	}
}

func str(srcLike bool, args ...Expr) string {
	var buf strings.Builder
	for i, arg := range args {
		if i > 0 && srcLike {
			buf.WriteByte(' ')
		}
		exprWriteTo(&buf, arg, srcLike)
	}
	return buf.String()
}
