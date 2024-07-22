package main

import (
	"fmt"
)

const disableTcoFuncs = true // caution: if `true`, cannot def `macro`s; this bool is just for quick temporary via-REPL trouble-shootings to see if TCO got somehow broken (or to enable call tracing for trouble-shooting)
const disableTracing = true || !disableTcoFuncs
const fakeFuncNamesForDebugging = false // costly but can aid the occasional trouble-shooting

// to confirm TCO still works, uncomment the 2 commented lines in `evalAndApply` below that are referring to `id`.
// another way: run `(sum2 10000000 0)` with TCO disabled (stack overflow) and then re-enabled (no stack overflow), where `sum2` is in github.com/kanaka/mal/blob/master/impls/tests/step5_tco.mal

func evalAndApply(env *Env, expr Expr) (Expr, error) {
	// id := time.Now().UnixNano()
	var err error
	for env != nil {
		// println("ITER", id, printExpr(expr, true))
		if it, is_list := expr.(ExprList); (!is_list) || (len(it) == 0) {
			expr, err = evalExpr(env, expr)
			env = nil
		} else if expr, err = macroExpand(env, it); err != nil {
			return nil, err
		} else if it, is_list = expr.(ExprList); is_list && (len(it) > 0) { // macroExpand might have returned a non-list
			var special_form SpecialForm
			if ident, _ := it[0].(ExprIdent); ident != "" {
				special_form = specialForms[ident]
			}
			if special_form != nil {
				if env, expr, err = special_form(env, it[1:]); err != nil {
					return nil, err
				}
			} else {
				trace(true, func() string { return str(true, it) })
				maybe_ident, _ := it[0].(ExprIdent)
				if fakeFuncNamesForDebugging && maybe_ident == "" {
					maybe_ident = ExprIdent(str(true, it))
				}
				expr, err = evalExpr(env, it)
				if err != nil {
					return nil, err
				}
				call := expr.(ExprList)
				callee, args := call[0], call[1:]
				switch fn := callee.(type) {
				default:
					return nil, newErrNotCallable(callee)
				case ExprFunc:
					trace(false, func() string { return fmt.Sprintf("CALL>>>%s", str(true, call)) })
					expr, err = fn(args)
					trace(false, func() string { return fmt.Sprintf("RET<<<%s", str(true, expr)) })
					env = nil
				case *ExprFn:
					if (fn.nameMaybe == "") && (maybe_ident != "") {
						fn.nameMaybe = "`" + string(maybe_ident) + "`"
					}
					expr = fn.body
					env, err = fn.envWith(args)
					if err != nil {
						return nil, err
					}
				}
			}
		}
	}
	return expr, err
}

func evalExpr(env *Env, expr Expr) (Expr, error) {
	switch it := expr.(type) {
	case ExprIdent:
		return env.get(it)
	case ExprHashMap:
		var err error
		hash_map := make(ExprHashMap, len(it))
		for key, value := range it {
			hash_map[key], err = evalAndApply(env, value)
			if err != nil {
				return nil, err
			}
		}
		return hash_map, nil
	case ExprVec:
		var err error
		vec := make(ExprVec, len(it))
		for i, item := range it {
			vec[i], err = evalAndApply(env, item)
			if err != nil {
				return nil, err
			}
		}
		return vec, nil
	case ExprList:
		var err error
		list := make(ExprList, len(it))
		for i, item := range it {
			list[i], err = evalAndApply(env, item)
			if err != nil {
				return nil, err
			}
		}
		return list, nil
	}
	return expr, nil
}

func trace(isStackTraceAdd bool, msg func() string) {
	if disableTracing {
		return
	}
	str := msg()
	if isStackTraceAdd {
		stackTrace = append(stackTrace, str)
	}
	println(str)
}

var stackTrace []string
