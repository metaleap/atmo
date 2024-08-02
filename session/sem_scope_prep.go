package session

import (
	"atmo/util"
	"atmo/util/str"
)

func (me *SrcPack) semCheckCount(wantAtLeast int, wantAtMost int, have SemExprs, errDst *SemExpr, forArgs bool) bool {
	if wantAtLeast >= 0 {
		plural := util.If((wantAtLeast <= wantAtMost) && (wantAtLeast != 1), "s", "")
		moniker := util.If(!forArgs, "expression"+plural+" in here", "arg"+plural+" for callee")
		if forArgs && (errDst != nil) {
			if call, _ := errDst.Val.(*SemValCall); (call != nil) && (call.Callee.From != nil) && (call.Callee.From.SrcNode != nil) && (call.Callee.From.SrcNode.Src != "") {
				moniker += " `" + call.Callee.From.SrcNode.Src + "`"
			}
		}
		if (wantAtLeast == wantAtMost) && (wantAtLeast != len(have)) {
			errDst.ErrsOwn.Add(errDst.From.SrcSpan.newDiagErr(ErrCodeExpectedFoo, str.Fmt("%d %s instead of %d", wantAtLeast, moniker, len(have))))
			return false
		} else if len(have) < wantAtLeast {
			errDst.ErrsOwn.Add(errDst.From.SrcSpan.newDiagErr(ErrCodeExpectedFoo, str.Fmt("at least %d %s instead of %d", wantAtLeast, moniker, len(have))))
			return false
		} else if (wantAtMost > wantAtLeast) && (len(have) > wantAtMost) {
			errDst.ErrsOwn.Add(errDst.From.SrcSpan.newDiagErr(ErrCodeExpectedFoo, str.Fmt("%d to %d %s instead of %d", wantAtLeast, wantAtMost, moniker, len(have))))
			return false
		}
	}
	return true
}

func semCheckIs[T any](equivPrimType MoValPrimType, expr *SemExpr) *T {
	if ret, is := expr.Val.(*T); is {
		return ret
	}
	expr.ErrsOwn.Add(expr.From.SrcSpan.newDiagErr(ErrCodeExpectedFoo, str.Fmt("%s here instead of `%s`",
		util.If(equivPrimType < 0, "a comparable value", equivPrimType.Str(true)),
		expr.From.SrcNode.Src)))
	return nil
}

func (me *SrcPack) semCheckType(expr *SemExpr, expect *SemType) bool {
	if !expect.Eq(expr.Type) {
		if !expr.HasErrs() { // dont wanna be too noisy
			err := expr.ErrNew(ErrCodeTypeMismatch, expect.String(), expr.Type.String())
			err.Rel = srcFileLocs([]string{
				str.Fmt("type `%s` decided here", expect.String()),
				str.Fmt("type `%s` decided here", expr.Type.String()),
			}, expect.DueTo, expr.Type.DueTo)
			expr.ErrsOwn.Add(err)
		}
		return false
	}
	return true
}

func (me *SrcPack) semPrepScopeOnSet(self *SemExpr) {
	call := self.Val.(*SemValCall)
	if me.semCheckCount(2, 2, call.Args, self, true) {
		expr_name, expr_value := call.Args[0], call.Args[1]
		if ident := semCheckIs[SemValIdent](MoPrimTypeIdent, expr_name); ident != nil {
			ident.IsSet = true
			is_name_invalid := ident.Name.IsReserved()
			if is_name_invalid {
				self.ErrsOwn.Add(expr_name.From.SrcSpan.newDiagErr(ErrCodeReserved, ident.Name, ident.Name[0:1]))
			}

			if value_ident := expr_value.MaybeIdent(false); (value_ident != "") && (moPrimOpsLazy[value_ident] != nil) {
				self.ErrsOwn.Add(expr_value.From.SrcSpan.newDiagErr(ErrCodeNotAValue, value_ident))
			}
			if !is_name_invalid {
				scope, resolved := self.Scope.Lookup(ident.Name)
				if resolved == nil {
					ident.IsDecl = true
					self.Scope.Own[ident.Name] = &SemScopeEntry{DeclParamOrCallOrFunc: self, Refs: map[*SemExpr]util.Void{}}
				} else {
					resolved.SubsequentSetCalls = append(resolved.SubsequentSetCalls, self)
					if (scope == self.Scope) && (scope == &me.Trees.Sem.Scope) {
						err := self.From.SrcSpan.newDiagErr(ErrCodeDuplTopDecl, ident.Name)
						err.Rel = srcFileLocs([]string{str.Fmt("the other `%s` definition", ident.Name)}, resolved.DeclParamOrCallOrFunc)
						self.ErrsOwn.Add(err)
					}
				}
			}
		}
	}
}

func (me *SrcPack) semPrepScopeOnFn(self *SemExpr) {
	call := self.Val.(*SemValCall)
	if me.semCheckCount(2, 2, call.Args, self, true) {
		if params_list, body_list := semCheckIs[SemValList](MoPrimTypeList, call.Args[0]), semCheckIs[SemValList](MoPrimTypeList, call.Args[1]); (params_list != nil) && (body_list != nil) {
			var ok_params SemExprs
			for _, param := range params_list.Items {
				if ident := semCheckIs[SemValIdent](MoPrimTypeIdent, param); ident != nil {
					if ident.Name.IsReserved() {
						self.ErrsOwn.Add(param.From.SrcSpan.newDiagErr(ErrCodeReserved, ident.Name, ident.Name[0:1]))
					} else {
						ident.IsParam, ident.IsDecl = true, true
						ok_params = append(ok_params, param)
					}
				}
			}
			fn := &SemValFunc{
				Scope:   &SemScope{Parent: self.Scope, Own: map[MoValIdent]*SemScopeEntry{}},
				Params:  ok_params,
				IsMacro: (call.Callee.Val.(*SemValIdent).Name == moPrimOpMacro),
			}
			for _, param := range fn.Params {
				fn.Scope.Own[param.Val.(*SemValIdent).Name] = &SemScopeEntry{DeclParamOrCallOrFunc: param, Refs: map[*SemExpr]util.Void{}}
			}
			switch len(body_list.Items) {
			case 0:
				self.ErrsOwn.Add(call.Args[1].From.SrcSpan.newDiagErr(ErrCodeExpectedFoo, "one or more expressions"))
			case 1:
				fn.Body = body_list.Items[0]
			default:
				f, p, s := call.Args[1].From, call.Args[1], fn.Scope
				expr_do := &SemExpr{From: f, Parent: p, Scope: s, Val: &SemValCall{
					Callee: &SemExpr{Val: &SemValIdent{Name: moPrimOpDo}, From: f, Parent: p, Scope: s},
					Args:   SemExprs{{Val: body_list, From: f, Parent: p, Scope: s}}}}
				fn.Body = expr_do
			}
			if (fn.Body != nil) && (len(ok_params) == len(params_list.Items)) {
				fn.Body.Walk(true, nil, func(it *SemExpr) {
					it.Scope = fn.Scope
				})
				self.Val = fn
			}
		}
	}
}
