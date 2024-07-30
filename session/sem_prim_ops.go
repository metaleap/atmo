package session

import (
	"atmo/util"
	"atmo/util/str"
)

type SemPrimOpOrFn func(*SrcPack, *SemExpr)

var (
	semPrimOps map[MoValIdent]SemPrimOpOrFn
	semPrimFns map[MoValIdent]SemPrimOpOrFn
)

func init() {
	semPrimOps = map[MoValIdent]SemPrimOpOrFn{
		moPrimOpSet: (*SrcPack).semPrimOpSet,
		moPrimOpDo:  (*SrcPack).semPrimOpDo,
		moPrimOpFn:  (*SrcPack).semPrimOpFn,
	}
}

func (me *SrcPack) semCheckCount(wantAtLeast int, wantAtMost int, have SemExprs, errDst *SemExpr, forArgs bool) bool {
	if wantAtLeast >= 0 {
		plural := util.If((wantAtLeast <= wantAtMost) && (wantAtLeast != 1), "s", "")
		moniker := util.If(!forArgs, "expression"+plural, "arg"+plural+" for this call")
		if (wantAtLeast == wantAtMost) && (wantAtLeast != len(have)) {
			errDst.ErrsOwn.Add(errDst.From.SrcSpan.newDiagErr(ErrCodeExpectedFoo, str.Fmt("%d %s, not %d", wantAtLeast, moniker, len(have))))
			return false
		} else if len(have) < wantAtLeast {
			errDst.ErrsOwn.Add(errDst.From.SrcSpan.newDiagErr(ErrCodeExpectedFoo, str.Fmt("at least %d %s, not %d", wantAtLeast, moniker, len(have))))
			return false
		} else if (wantAtMost > wantAtLeast) && (len(have) > wantAtMost) {
			errDst.ErrsOwn.Add(errDst.From.SrcSpan.newDiagErr(ErrCodeExpectedFoo, str.Fmt("%d to %d %s, not %d", wantAtLeast, wantAtMost, moniker, len(have))))
			return false
		}
	}
	return true
}

func semCheckIs[T any](equivPrimType MoValPrimType, expr *SemExpr) *T {
	if ret, is := expr.Val.(*T); is {
		return ret
	}
	expr.ErrsOwn.Add(expr.From.SrcSpan.newDiagErr(ErrCodeExpectedFoo, str.Fmt("%s here instead of `%s`", equivPrimType.Str(true), expr.From.SrcNode.Src)))
	return nil
}

func (me *SrcPack) semPrimOpSet(self *SemExpr) {
	call := self.Val.(*SemValCall)
	if me.semCheckCount(2, 2, call.Args, self, true) {
		name, value := call.Args[0], call.Args[1]
		if ident := semCheckIs[SemValIdent](MoPrimTypeIdent, name); ident != nil {
			is_name_invalid := ident.MoVal.IsReserved()
			if is_name_invalid {
				self.ErrsOwn.Add(name.From.SrcSpan.newDiagErr(ErrCodeReserved, ident.MoVal, ident.MoVal[0:1]))
			}

			_ = value.EnsureResolvesIfIdent()
			if value_ident := value.MaybeIdent(); (value_ident != "") && (semPrimOps[value_ident] != nil) {
				self.ErrsOwn.Add(value.From.SrcSpan.newDiagErr(ErrCodeNotAValue, value_ident))
			}
			self.Fact(SemFact{Kind: SemFactEffectful}, call.Callee)
			if !is_name_invalid {
				scope, resolved := self.Scope.Lookup(ident.MoVal, false, nil)
				if resolved == nil {
					self.Scope.Own[ident.MoVal] = &SemScopeEntry{DeclVal: value}
				} else {
					resolved.SubsequentSetVals = append(resolved.SubsequentSetVals, value)
					if (scope == self.Scope) && (self.Parent == nil) {
						err := self.From.SrcSpan.newDiagErr(ErrCodeDuplTopDecl, ident.MoVal)
						err.Rel = &SrcFileLocs{File: resolved.DeclVal.From.SrcFile, Spans: []*SrcFileSpan{resolved.DeclVal.From.SrcSpan}, IsSet: []bool{true}, IsGet: []bool{false}}
						self.ErrsOwn.Add(err)
					}
				}
			}
		}
	}
}

func (me *SrcPack) semPrimOpDo(self *SemExpr) {
	call := self.Val.(*SemValCall)
	if me.semCheckCount(1, 1, call.Args, self, true) {
		if list := semCheckIs[SemValList](MoPrimTypeList, call.Args[0]); list != nil {
			for i, expr := range list.Items {
				is_last := (i == len(list.Items)-1)
				if is_last {
					self.FactsFrom(expr)
				} else if len(expr.HasFact(SemFactEffectful, nil, false, false, false)) == 0 {
					expr.Fact(SemFact{Kind: SemFactUnused}, expr)
				}
			}
		}
	}
}

func (me *SrcPack) semPrimOpFn(self *SemExpr) {
	call := self.Val.(*SemValCall)
	self.Fact(SemFact{Kind: SemFactCallable}, self)
	self.Fact(SemFact{Kind: SemFactPrimType, Of: MoPrimTypeFunc}, self)
	if me.semCheckCount(2, 2, call.Args, self, true) {
		if params_list, body_list := semCheckIs[SemValList](MoPrimTypeList, call.Args[0]), semCheckIs[SemValList](MoPrimTypeList, call.Args[1]); (params_list != nil) && (body_list != nil) {
			for _, param := range params_list.Items {
				if ident := semCheckIs[SemValIdent](MoPrimTypeIdent, param); ident != nil {
					if ident.MoVal.IsReserved() {
						self.ErrsOwn.Add(param.From.SrcSpan.newDiagErr(ErrCodeReserved, ident.MoVal, ident.MoVal[0:1]))
					}
				}
			}
			fn := &SemValFunc{
				Scope:  &SemScope{Parent: self.Scope, Own: map[MoValIdent]*SemScopeEntry{}},
				Params: params_list.Items,
			}
			switch len(body_list.Items) {
			case 0:
				self.ErrsOwn.Add(call.Args[1].From.SrcSpan.newDiagErr(ErrCodeExpectedFoo, "one or more expressions"))
			case 1:
				fn.Body = body_list.Items[0]
			default:
				f, p, s := call.Args[1].From, call.Args[1], fn.Scope
				expr_do := &SemExpr{From: f, Parent: p, Scope: s, Val: &SemValCall{
					Callee: &SemExpr{Val: &SemValIdent{MoVal: moPrimOpDo}, From: f, Parent: p, Scope: s},
					Args:   SemExprs{{Val: body_list, From: f, Parent: p, Scope: s}}}}
				me.semPrimOpDo(expr_do)
				fn.Body = expr_do
			}
			if fn.Body != nil {
				fn.Body.Walk(nil, func(it *SemExpr) {
					it.Scope = fn.Scope
				})
				self.Val = fn
			}
		}
	}
}
