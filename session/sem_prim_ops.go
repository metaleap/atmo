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
	}
}

func (me *SrcPack) semCheckCount(wantAtLeast int, wantAtMost int, have SemExprs, errDst *SemExpr, forArgs bool) bool {
	if wantAtLeast >= 0 {
		plural := util.If((wantAtLeast <= wantAtMost) && (wantAtLeast != 1), "s", "")
		moniker := util.If(!forArgs, "item"+plural, "arg"+plural+" for this call")
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
			if ident.MoVal.IsReserved() {
				self.ErrsOwn.Add(name.From.SrcSpan.newDiagErr(ErrCodeReserved, ident.MoVal, ident.MoVal[0:1]))
			}

			_ = value.EnsureResolvesIfIdent()
			if value_ident := value.MaybeIdent(); (value_ident != "") && (semPrimOps[value_ident] != nil) {
				self.ErrsOwn.Add(value.From.SrcSpan.newDiagErr(ErrCodeNotAValue, value_ident))
			}
			self.Fact(SemFact{Kind: SemFactEffectful}, call.Callee)
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
