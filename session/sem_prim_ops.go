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
	}
}

func (me *SrcPack) semCheckCount(wantAtLeast int, wantAtMost int, have SemExprs, errDst *SemExpr, forArgs bool) bool {
	if wantAtLeast >= 0 {
		plural := util.If((wantAtLeast <= wantAtMost) && (wantAtLeast != 1), "s", "")
		moniker := util.If(!forArgs, "item"+plural, "arg"+plural+" for this call")
		if (wantAtLeast == wantAtMost) && (wantAtLeast != len(have)) {
			errDst.ErrOwn = errDst.From.SrcSpan.newDiagErr(NoticeCodeExpectedFoo, str.Fmt("%d %s, not %d", wantAtLeast, moniker, len(have)))
			return false
		} else if len(have) < wantAtLeast {
			errDst.ErrOwn = errDst.From.SrcSpan.newDiagErr(NoticeCodeExpectedFoo, str.Fmt("at least %d %s, not %d", wantAtLeast, moniker, len(have)))
			return false
		} else if (wantAtMost > wantAtLeast) && (len(have) > wantAtMost) {
			errDst.ErrOwn = errDst.From.SrcSpan.newDiagErr(NoticeCodeExpectedFoo, str.Fmt("%d to %d %s, not %d", wantAtLeast, wantAtMost, moniker, len(have)))
			return false
		}
	}
	return true
}

func semCheckIs[T any](equivPrimType MoValPrimType, expr *SemExpr) *T {
	if ret, is := expr.Val.(*T); is {
		return ret
	}
	expr.ErrOwn = expr.From.SrcSpan.newDiagErr(NoticeCodeExpectedFoo, str.Fmt("%s here instead of `%s`", equivPrimType.Str(true), expr.From.SrcNode.Src))
	return nil
}

func (me *SrcPack) semPrimOpSet(self *SemExpr) {
	call := self.Val.(*SemValCall)
	if me.semCheckCount(2, 2, call.Args, self, true) {
		name, value := call.Args[0], call.Args[1]
		if ident := semCheckIs[SemValIdent](MoPrimTypeIdent, name); ident != nil {
			if ident.MoVal.IsReserved() {
				name.ErrOwn = name.From.SrcSpan.newDiagErr(NoticeCodeReserved, ident.MoVal, ident.MoVal[0:1])
				return
			}

			if _ = value.EnsureResolvesIfIdent(); value.HasErrs() {
				return
			} else if value_ident := value.MaybeIdent(); (value_ident != "") && (semPrimOps[value_ident] != nil) {
				value.ErrOwn = value.From.SrcSpan.newDiagErr(NoticeCodeNotAValue, value_ident)
				return
			}
			self.Fact(SemFact{Kind: SemFactSideEffects}, call.Callee)

			scope, resolved := self.Scope.Lookup(ident.MoVal, false, nil)
			if resolved == nil {
				self.Scope.Own[ident.MoVal] = &SemScopeEntry{DeclVal: value}
			} else if (scope == self.Scope) && (self.Parent == nil) {
				self.ErrOwn = self.From.SrcSpan.newDiagErr(NoticeCodeDuplTopDecl, ident.MoVal)
				self.ErrOwn.Rel = &SrcFileLocs{File: resolved.DeclVal.From.SrcFile, Spans: []*SrcFileSpan{resolved.DeclVal.From.SrcSpan}, IsSet: []bool{true}, IsGet: []bool{false}}
				return
			} else {
				resolved.SubsequentSetVals = append(resolved.SubsequentSetVals, value)
			}
		}
	}
}
