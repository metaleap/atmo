package session

import (
	"atmo/util"
	"atmo/util/sl"
	"atmo/util/str"
)

type SemPrimOpOrFn func(*SrcPack, *SemExpr)

var (
	semPrimOps map[MoValIdent]SemPrimOpOrFn
	semPrimFns map[MoValIdent]SemPrimOpOrFn
)

func init() {
	semPrimOps = map[MoValIdent]SemPrimOpOrFn{
		moPrimOpSet:    (*SrcPack).semPrimOpSet,
		moPrimOpDo:     (*SrcPack).semPrimOpDo,
		moPrimOpFn:     (*SrcPack).semPrimOpFn,
		moPrimOpMacro:  (*SrcPack).semPrimOpFn,
		moPrimOpFnCall: (*SrcPack).semPrimOpFnCall,
		moPrimOpQuote:  (*SrcPack).semPrimOpQuote,
		moPrimOpQQuote: (*SrcPack).semPrimOpQQuote,
		moPrimOpExpand: (*SrcPack).semPrimOpExpand,
		moPrimOpCaseOf: (*SrcPack).semPrimOpCaseOf,
		moPrimOpAnd:    (*SrcPack).semPrimOpAndOr,
		moPrimOpOr:     (*SrcPack).semPrimOpAndOr,
	}
}

func (me *SrcPack) semCheckCount(wantAtLeast int, wantAtMost int, have SemExprs, errDst *SemExpr, forArgs bool) bool {
	if wantAtLeast >= 0 {
		plural := util.If((wantAtLeast <= wantAtMost) && (wantAtLeast != 1), "s", "")
		moniker := util.If(!forArgs, "expression"+plural, "arg"+plural+" for callee")
		if forArgs && (errDst != nil) {
			if call, _ := errDst.Val.(*SemValCall); (call != nil) && (call.Callee.From != nil) && (call.Callee.From.SrcNode != nil) && (call.Callee.From.SrcNode.Src != "") {
				moniker += " `" + call.Callee.From.SrcNode.Src + "`"
			}
		}
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

			if value_ident := value.MaybeIdent(); (value_ident != "") && (semPrimOps[value_ident] != nil) {
				self.ErrsOwn.Add(value.From.SrcSpan.newDiagErr(ErrCodeNotAValue, value_ident))
			}
			self.Fact(SemFact{Kind: SemFactEffectful}, call.Callee)
			if !is_name_invalid {
				scope, resolved := self.Scope.Lookup(ident.MoVal, false, nil)
				if resolved == nil {
					self.Scope.Own[ident.MoVal] = &SemScopeEntry{DeclParamOrSetCall: self}
				} else {
					resolved.SubsequentSetCalls = append(resolved.SubsequentSetCalls, self)
					if (scope == self.Scope) && (self.Parent == nil) {
						err := self.From.SrcSpan.newDiagErr(ErrCodeDuplTopDecl, ident.MoVal)
						err.Rel = &SrcFileLocs{File: resolved.DeclParamOrSetCall.From.SrcFile, Spans: []*SrcFileSpan{resolved.DeclParamOrSetCall.From.SrcSpan}, IsSet: []bool{true}, IsGet: []bool{false}}
						self.ErrsOwn.Add(err)
					}
				}
			}
		}
	}
	self.Fact(SemFact{Kind: SemFactPrimType, Data: MoPrimTypeVoid}, self)
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
	self.Fact(SemFact{Kind: SemFactPrimType, Data: MoPrimTypeFunc}, self)
	is_macro := (call.Callee.Val.(*SemValIdent).MoVal == moPrimOpMacro)
	if is_macro {
		self.Fact(SemFact{Kind: SemFactFuncIsMacro}, call.Callee)
	}
	if me.semCheckCount(2, 2, call.Args, self, true) {
		if params_list, body_list := semCheckIs[SemValList](MoPrimTypeList, call.Args[0]), semCheckIs[SemValList](MoPrimTypeList, call.Args[1]); (params_list != nil) && (body_list != nil) {
			var ok_params SemExprs
			for _, param := range params_list.Items {
				if ident := semCheckIs[SemValIdent](MoPrimTypeIdent, param); ident != nil {
					if ident.MoVal.IsReserved() {
						self.ErrsOwn.Add(param.From.SrcSpan.newDiagErr(ErrCodeReserved, ident.MoVal, ident.MoVal[0:1]))
					} else {
						ok_params = append(ok_params, param)
					}
				}
			}
			fn := &SemValFunc{
				Scope:  &SemScope{Parent: self.Scope, Own: map[MoValIdent]*SemScopeEntry{}},
				Params: ok_params,
			}
			for _, param := range fn.Params {
				fn.Scope.Own[param.Val.(*SemValIdent).MoVal] = &SemScopeEntry{DeclParamOrSetCall: param}
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
					if entry := it.ResolvedIfIdent(false); (entry != nil) && sl.Has(fn.Params, entry.DeclParamOrSetCall) {
						entry.DeclParamOrSetCall.FactsFrom(it)
					}
				})
				self.Val = fn
			}
		}
	}
}

func (me *SrcPack) semPrimOpFnCall(self *SemExpr) {
	call := self.Val.(*SemValCall)
	if me.semCheckCount(2, 2, call.Args, self, true) {
		if args_list := semCheckIs[SemValList](MoPrimTypeList, call.Args[1]); args_list != nil {
			call.Args[0].Fact(SemFact{Kind: SemFactCallable}, self)
		}
	}
}

func (me *SrcPack) semPrimOpQuote(self *SemExpr) {
	call := self.Val.(*SemValCall)
	if me.semCheckCount(1, 1, call.Args, self, true) {
		self.Fact(SemFact{Kind: SemFactQQuote, Data: false}, call.Callee)
	}
}

func (me *SrcPack) semPrimOpQQuote(self *SemExpr) {
	call := self.Val.(*SemValCall)
	if me.semCheckCount(1, 1, call.Args, self, true) {
		self.Fact(SemFact{Kind: SemFactQQuote, Data: true}, call.Callee)
	}
}

func (me *SrcPack) semPrimOpExpand(self *SemExpr) {
	call := self.Val.(*SemValCall)
	if me.semCheckCount(1, 1, call.Args, self, true) {
		_ = semCheckIs[SemValCall](MoPrimTypeCall, call.Args[0])
	}
}

func (me *SrcPack) semPrimOpCaseOf(self *SemExpr) {
	call := self.Val.(*SemValCall)
	if me.semCheckCount(1, 1, call.Args, self, true) {
		if dict := semCheckIs[SemValDict](MoPrimTypeDict, call.Args[0]); dict != nil {
			for _, key := range dict.Keys {
				key.Fact(SemFact{Kind: SemFactPrimType, Data: MoPrimTypeBool}, call.Callee)
			}
		}
	}
}

func (me *SrcPack) semPrimOpAndOr(self *SemExpr) {
	call := self.Val.(*SemValCall)
	if me.semCheckCount(2, 2, call.Args, self, true) {
		call.Args[0].Fact(SemFact{Kind: SemFactPrimType, Data: MoPrimTypeBool}, call.Callee)
		call.Args[1].Fact(SemFact{Kind: SemFactPrimType, Data: MoPrimTypeBool}, call.Callee)
	}
}
