package main

import (
	"fmt"
)

var (
	exprIdentDo            = ExprIdent("do")
	exprIdentQuote         = ExprIdent("quote")
	exprIdentQuasiQuote    = ExprIdent("quasiQuote")
	exprIdentUnquote       = ExprIdent("unquote")
	exprIdentSpliceUnquote = ExprIdent("spliceUnquote")
	exprIdentMacro         = ExprIdent("macro")

	specialForms map[ExprIdent]SpecialForm
)

func init() { // in here, instead of above, to avoid "initialization cycle" error:
	specialForms = map[ExprIdent]SpecialForm{
		"def":               stdDef,
		"set":               stdSet,
		"if":                stdIf,
		"let":               stdLet,
		"fn":                stdFn,
		exprIdentDo:         stdDo,
		exprIdentQuote:      stdQuote,
		exprIdentQuasiQuote: stdQuasiQuote,
		"macroExpand":       stdMacroExpand,
		"try":               stdTryCatch,
	}
}

// like the funcs in `std.go`, the special-forms return an `Expr, error`. but:
// in addition, they also return an `*Env`. if that is `nil`, the returned Expr
// is a complete result, just like with the std funcs. but if not, the TCO loop
// in `evalAndApply` goes through another iteration with the returned Expr.

func stdDef(env *Env, args []Expr) (*Env, Expr, error) {
	return defOrSet(true, env, args)
}
func stdSet(env *Env, args []Expr) (*Env, Expr, error) {
	return defOrSet(false, env, args)
}
func defOrSet(isDef bool, env *Env, args []Expr) (*Env, Expr, error) {
	if err := checkArgsCount(2, 2, "`def` and `set`", args); err != nil {
		return nil, nil, err
	}
	name, err := checkIs[ExprIdent](args[0])
	if err != nil {
		return nil, nil, err
	}
	if _, is_reserved := specialForms[name]; is_reserved {
		return nil, nil, fmt.Errorf("cannot redefine `%s`", name)
	}
	if !isDef { // check if this `set` refers to something that's been `def`d
		if _, err := env.get(name); err != nil {
			return nil, nil, err
		}
	} else if env.hasOwn(name) && !malCompat { // cannot re`def` locals
		return nil, nil, fmt.Errorf("cannot redefine `%s` (use `set` here instead of `def`)", name)
	}

	var ret Expr
	if maybe_macro, is_macro, err := isListStartingWithIdent(args[1], exprIdentMacro, -1); err != nil {
		return nil, nil, err
	} else if is_macro {
		_, maybe_macro, err := stdFn(env, maybe_macro[1:])
		if err != nil {
			return nil, nil, err
		}
		macro, err := checkIs[*ExprFn](maybe_macro)
		if err != nil {
			return nil, nil, err
		}
		macro.isMacro = true
		ret = macro
	} else if ret, err = evalAndApply(env, args[1]); err != nil {
		return nil, nil, err
	}
	env.set(name, ret)
	return nil, ret, nil
}

func stdDo(env *Env, args []Expr) (tailEnv *Env, expr Expr, err error) {
	if err = checkArgsCount(1, -1, "`do`", args); err != nil {
		return
	}
	for _, arg := range args[:len(args)-1] {
		if expr, err = evalAndApply(env, arg); err != nil {
			return
		}
	}
	tailEnv, expr = env, args[len(args)-1]
	return
}

func stdLet(env *Env, args []Expr) (*Env, Expr, error) {
	if err := checkArgsCount(2, -1, "`let`", args); err != nil {
		return nil, nil, err
	}
	bindings, err := checkIsSeq(args[0])
	if err != nil {
		return nil, nil, err
	}
	let_env := newEnv(env, nil, nil)
	for _, binding := range bindings {
		pair, err := checkIsSeq(binding)
		if err != nil {
			return nil, nil, err
		}
		if err = checkArgsCount(2, 2, "a `let` pairing", pair); err != nil {
			return nil, nil, err
		}
		name, err := checkIs[ExprIdent](pair[0])
		if err != nil {
			return nil, nil, err
		}
		if _, is_special := specialForms[name]; is_special {
			return nil, nil, fmt.Errorf("cannot redefine `%s`", name)
		}
		expr, err := evalAndApply(let_env, pair[1])
		if err != nil {
			return nil, nil, err
		}
		let_env.set(name, expr)
	}
	return stdDo(let_env, args[1:])
}

func stdIf(env *Env, args []Expr) (*Env, Expr, error) {
	if err := checkArgsCount(2, 3, "`if`", args); err != nil {
		return nil, nil, err
	}
	if len(args) < 3 {
		args = append(args, exprNil)
	}
	expr, err := evalAndApply(env, args[0])
	if err != nil {
		return nil, nil, err
	}
	idx := 1
	if isNilOrFalse(expr) {
		idx = 2
	}
	return env, args[idx], nil
}

func stdFn(env *Env, args []Expr) (*Env, Expr, error) {
	if err := checkArgsCount(2, -1, "`fn`", args); err != nil {
		return nil, nil, err
	}

	params, err := checkIsSeq(args[0])
	if err != nil {
		return nil, nil, err
	}
	if err := checkAre[ExprIdent](params...); err != nil {
		return nil, nil, err
	}
	var is_variadic bool
	if len(params) >= 2 {
		if amper := params[len(params)-2].(ExprIdent); amper == "&" {
			is_variadic = true // and we remove the ampersand param:
			params = append(params[:len(params)-2], params[len(params)-1])
		}
	}

	body := args[1]
	if len(args) > 2 {
		body = append(ExprList{exprIdentDo}, args[1:]...)
	}
	var expr Expr = &ExprFn{params: params, body: body, env: env, isVariadic: is_variadic}

	if disableTcoFuncs {
		expr = (ExprFunc)(expr.(*ExprFn).Call)
	}
	return nil, expr, nil
}

func stdMacroExpand(env *Env, args []Expr) (*Env, Expr, error) {
	if err := checkArgsCount(1, 1, "`macroExpand`", args); err != nil {
		return nil, nil, err
	}
	list, err := checkIs[ExprList](args[0])
	if err != nil {
		return nil, nil, err
	}

	expr, err := macroExpand(env, list)
	if err != nil {
		return nil, nil, err
	}
	return nil, expr, nil
}

func stdQuote(_ *Env, args []Expr) (*Env, Expr, error) {
	if err := checkArgsCount(1, 1, "`quote`", args); err != nil {
		return nil, nil, err
	}
	return nil, args[0], nil
}

func stdQuasiQuote(env *Env, args []Expr) (*Env, Expr, error) {
	if err := checkArgsCount(1, 1, "`quasiQuote`", args); err != nil {
		return nil, nil, err
	}
	_, is_vec := args[0].(ExprVec)
	list, not_a_seq := checkIsSeq(args[0])
	if not_a_seq != nil {
		return stdQuote(env, args)
	}

	if unquote, ok, err := isListStartingWithIdent(args[0], exprIdentUnquote, 2); err != nil {
		return nil, nil, err
	} else if ok {
		return env, unquote[1], nil
	}

	expr := make(ExprList, 0, len(list))
	for _, item := range list {
		if unquote, ok, err := isListStartingWithIdent(item, exprIdentUnquote, 2); err != nil {
			return nil, nil, err
		} else if ok {
			if unquoted, err := evalAndApply(env, unquote[1]); err != nil {
				return nil, nil, err
			} else {
				expr = append(expr, unquoted)
			}
		} else if splice_unquote, ok, err := isListStartingWithIdent(item, exprIdentSpliceUnquote, 2); err != nil {
			return nil, nil, err
		} else if ok {
			evaled, err := evalAndApply(env, splice_unquote[1])
			if err != nil {
				return nil, nil, err
			}
			splicees, err := checkIs[ExprList](evaled)
			if err != nil {
				return nil, nil, err
			}
			for _, splicee := range splicees {
				if unquoted, err := evalAndApply(env, splicee); err != nil {
					return nil, nil, err
				} else {
					expr = append(expr, unquoted)
				}
			}
		} else {
			_, evaled, err := stdQuasiQuote(env, []Expr{item})
			if err != nil {
				return nil, nil, err
			}
			expr = append(expr, evaled)
		}
	}
	if is_vec {
		return nil, (ExprVec)(expr), nil
	}
	return nil, expr, nil
}

func stdTryCatch(env *Env, args []Expr) (*Env, Expr, error) {
	if err := checkArgsCount(1, 2, "`try`", args); err != nil {
		return nil, nil, err
	}
	if len(args) == 1 {
		return env, args[0], nil
	}

	catch, ok, err := isListStartingWithIdent(args[1], "catch", 3)
	if err != nil {
		return nil, nil, err
	} else if !ok {
		if malCompat {
			catch, ok, _ = isListStartingWithIdent(args[1], "catch*", 3) // MAL compat for testing `./self-hosted-mal/*.mal`s
		}
		if !ok {
			return nil, nil, fmt.Errorf("expected `(catch theErr exprHandlingIt)` as the last form in `try` , instead of `%s`", str(true, args[1]))
		}
	}

	expr, err := evalAndApply(env, args[0])
	if err != nil {
		err_expr, is := err.(ExprErr)
		if !is {
			err_expr = ExprErr{It: err}
		}
		if expr, is = err_expr.It.(Expr); !is {
			expr = err_expr
		}
		env.set(catch[1].(ExprIdent), expr)
		return env, catch[2], nil
	}

	return nil, expr, nil
}
