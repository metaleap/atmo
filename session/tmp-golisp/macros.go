package main

func macroExpand(env *Env, expr Expr) (Expr, error) {
	for callee := macroCallCallee(env, expr); callee != nil; callee = macroCallCallee(env, expr) {
		it, err := callee.Call(expr.(ExprList)[1:])
		if err != nil {
			return nil, err
		}
		expr = it
	}
	return expr, nil
}

func macroCallCallee(env *Env, expr Expr) *ExprFn {
	if list, _ := expr.(ExprList); len(list) > 0 {
		if ident, _ := list[0].(ExprIdent); ident != "" {
			if maybe_fn := env.find(ident); maybe_fn != nil {
				if fn, _ := maybe_fn.(*ExprFn); (fn != nil) && fn.isMacro {
					return fn
				}
			}
		}
	}
	return nil
}
