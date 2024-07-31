package session

import (
	"atmo/util"
	"atmo/util/sl"
	"atmo/util/str"
)

var (
	semTypingPrimOpsEnv map[MoValIdent]SemType
	semTypingPrimOpsDo  map[MoValIdent]func(*SrcPack, *semTypeInfer, *SemExpr, SemType, map[MoValIdent]SemType) *SemExpr
)

func init() {
	ty_prim := func(t MoValPrimType) SemType { return semTypeNew(nil, t) }
	ty_fn_prims := func(t ...MoValPrimType) SemType {
		return semTypeNew(nil, MoPrimTypeFunc, sl.To(t, ty_prim)...)
	}
	ty_fn := func(t ...SemType) SemType { return semTypeNew(nil, MoPrimTypeFunc, t...) }
	semTypingPrimOpsEnv = map[MoValIdent]SemType{
		moPrimOpAnd:         ty_fn_prims(MoPrimTypeBool, MoPrimTypeBool, MoPrimTypeBool),
		moPrimOpOr:          ty_fn_prims(MoPrimTypeBool, MoPrimTypeBool, MoPrimTypeBool),
		moPrimFnNot:         ty_fn_prims(MoPrimTypeBool, MoPrimTypeBool),
		moPrimFnNumIntAdd:   ty_fn_prims(MoPrimTypeNumInt, MoPrimTypeNumInt, MoPrimTypeNumInt),
		moPrimFnNumIntSub:   ty_fn_prims(MoPrimTypeNumInt, MoPrimTypeNumInt, MoPrimTypeNumInt),
		moPrimFnNumIntMul:   ty_fn_prims(MoPrimTypeNumInt, MoPrimTypeNumInt, MoPrimTypeNumInt),
		moPrimFnNumIntDiv:   ty_fn_prims(MoPrimTypeNumInt, MoPrimTypeNumInt, MoPrimTypeNumInt),
		moPrimFnNumIntMod:   ty_fn_prims(MoPrimTypeNumInt, MoPrimTypeNumInt, MoPrimTypeNumInt),
		moPrimFnNumUintAdd:  ty_fn_prims(MoPrimTypeNumUint, MoPrimTypeNumUint, MoPrimTypeNumUint),
		moPrimFnNumUintSub:  ty_fn_prims(MoPrimTypeNumUint, MoPrimTypeNumUint, MoPrimTypeNumUint),
		moPrimFnNumUintMul:  ty_fn_prims(MoPrimTypeNumUint, MoPrimTypeNumUint, MoPrimTypeNumUint),
		moPrimFnNumUintDiv:  ty_fn_prims(MoPrimTypeNumUint, MoPrimTypeNumUint, MoPrimTypeNumUint),
		moPrimFnNumUintMod:  ty_fn_prims(MoPrimTypeNumUint, MoPrimTypeNumUint, MoPrimTypeNumUint),
		moPrimFnNumFloatAdd: ty_fn_prims(MoPrimTypeNumFloat, MoPrimTypeNumFloat, MoPrimTypeNumFloat),
		moPrimFnNumFloatSub: ty_fn_prims(MoPrimTypeNumFloat, MoPrimTypeNumFloat, MoPrimTypeNumFloat),
		moPrimFnNumFloatMul: ty_fn_prims(MoPrimTypeNumFloat, MoPrimTypeNumFloat, MoPrimTypeNumFloat),
		moPrimFnNumFloatDiv: ty_fn_prims(MoPrimTypeNumFloat, MoPrimTypeNumFloat, MoPrimTypeNumFloat),
		moPrimFnStrLen:      ty_fn_prims(MoPrimTypeStr, MoPrimTypeNumUint),
		moPrimFnStrCharAt:   ty_fn_prims(MoPrimTypeStr, MoPrimTypeNumUint, MoPrimTypeChar),
		moPrimFnStrRange:    ty_fn_prims(MoPrimTypeStr, MoPrimTypeNumUint, MoPrimTypeNumUint, MoPrimTypeStr),
		moPrimFnStrConcat:   ty_fn(semTypeNew(nil, MoPrimTypeList, ty_prim(MoPrimTypeStr)), ty_prim(MoPrimTypeStr)),
		moPrimFnReplEnv:     ty_fn(semTypeNew(nil, MoPrimTypeDict, ty_prim(MoPrimTypeIdent), ty_prim(MoPrimTypeUntyped))),
		moPrimFnReplReset:   ty_fn_prims(MoPrimTypeVoid),
	}
	semTypingPrimOpsDo = map[MoValIdent]func(*SrcPack, *semTypeInfer, *SemExpr, SemType, map[MoValIdent]SemType) *SemExpr{
		moPrimOpFn:            (*SrcPack).semTypingPrimOpFnOrMacro,
		moPrimOpMacro:         (*SrcPack).semTypingPrimOpFnOrMacro,
		moPrimOpFnCall:        (*SrcPack).semTypingPrimOpFnCall,
		moPrimOpSet:           (*SrcPack).semTypingPrimOpSet,
		moPrimOpCaseOf:        (*SrcPack).semTypingPrimOpCaseOf,
		moPrimOpDo:            (*SrcPack).semTypingPrimOpDo,
		moPrimOpExpand:        (*SrcPack).semTypingPrimOpExpand,
		moPrimOpQQuote:        (*SrcPack).semTypingPrimOpQQuote,
		moPrimOpQuote:         (*SrcPack).semTypingPrimOpQuote,
		moPrimOpSpliceUnquote: (*SrcPack).semTypingPrimOpSpliceUnquote,
		moPrimOpUnquote:       (*SrcPack).semTypingPrimOpUnquote,
		moPrimFnReplPrint:     (*SrcPack).semTypingPrimFnReplPrint,
		moPrimFnCast:          (*SrcPack).semTypingPrimFnCast,
		moPrimFnEq:            (*SrcPack).semTypingPrimFnEq,
		moPrimFnNeq:           (*SrcPack).semTypingPrimFnNeq,
		moPrimFnGeq:           (*SrcPack).semTypingPrimFnGeq,
		moPrimFnLeq:           (*SrcPack).semTypingPrimFnLeq,
		moPrimFnLt:            (*SrcPack).semTypingPrimFnLt,
		moPrimFnGt:            (*SrcPack).semTypingPrimFnGt,
		moPrimFnPrimTypeTag:   (*SrcPack).semTypingPrimFnPrimTypeTag,
		moPrimFnListGet:       (*SrcPack).semTypingPrimFnListGet,
		moPrimFnListSet:       (*SrcPack).semTypingPrimFnListSet,
		moPrimFnListRange:     (*SrcPack).semTypingPrimFnListRange,
		moPrimFnListLen:       (*SrcPack).semTypingPrimFnListLen,
		moPrimFnListConcat:    (*SrcPack).semTypingPrimFnListConcat,
		moPrimFnDictHas:       (*SrcPack).semTypingPrimFnDictHas,
		moPrimFnDictGet:       (*SrcPack).semTypingPrimFnDictGet,
		moPrimFnDictSet:       (*SrcPack).semTypingPrimFnDictSet,
		moPrimFnDictDel:       (*SrcPack).semTypingPrimFnDictDel,
		moPrimFnDictLen:       (*SrcPack).semTypingPrimFnDictLen,
		moPrimFnErrNew:        (*SrcPack).semTypingPrimFnErrNew,
		moPrimFnErrVal:        (*SrcPack).semTypingPrimFnErrVal,
		moPrimFnStr:           (*SrcPack).semTypingPrimFnStr,
		moPrimFnExprStr:       (*SrcPack).semTypingPrimFnExprStr,
		moPrimFnExprParse:     (*SrcPack).semTypingPrimFnExprParse,
		moPrimFnExprEval:      (*SrcPack).semTypingPrimFnExprEval,
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

func (me *SrcPack) semPrepScopeOnSet(self *SemExpr) {
	call := self.Val.(*SemValCall)
	if me.semCheckCount(2, 2, call.Args, self, true) {
		name, value := call.Args[0], call.Args[1]
		if ident := semCheckIs[SemValIdent](MoPrimTypeIdent, name); ident != nil {
			is_name_invalid := ident.Ident.IsReserved()
			if is_name_invalid {
				self.ErrsOwn.Add(name.From.SrcSpan.newDiagErr(ErrCodeReserved, ident.Ident, ident.Ident[0:1]))
			}

			if value_ident := value.MaybeIdent(); (value_ident != "") && (moPrimOpsLazy[value_ident] != nil) {
				self.ErrsOwn.Add(value.From.SrcSpan.newDiagErr(ErrCodeNotAValue, value_ident))
			}
			if !is_name_invalid {
				scope, resolved := self.Scope.Lookup(ident.Ident)
				if resolved == nil {
					self.Scope.Own[ident.Ident] = &SemScopeEntry{DeclParamOrSetCall: self}
				} else {
					resolved.SubsequentSetCalls = append(resolved.SubsequentSetCalls, self)
					if (scope == self.Scope) && (self.Parent == nil) {
						err := self.From.SrcSpan.newDiagErr(ErrCodeDuplTopDecl, ident.Ident)
						err.Rel = srcFileLocs([]string{str.Fmt("the other `%s` definition", ident.Ident)}, resolved.DeclParamOrSetCall)
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
					if ident.Ident.IsReserved() {
						self.ErrsOwn.Add(param.From.SrcSpan.newDiagErr(ErrCodeReserved, ident.Ident, ident.Ident[0:1]))
					} else {
						ok_params = append(ok_params, param)
					}
				}
			}
			fn := &SemValFunc{
				Scope:   &SemScope{Parent: self.Scope, Own: map[MoValIdent]*SemScopeEntry{}},
				Params:  ok_params,
				IsMacro: (call.Callee.Val.(*SemValIdent).Ident == moPrimOpMacro),
			}
			for _, param := range fn.Params {
				fn.Scope.Own[param.Val.(*SemValIdent).Ident] = &SemScopeEntry{DeclParamOrSetCall: param}
			}
			switch len(body_list.Items) {
			case 0:
				self.ErrsOwn.Add(call.Args[1].From.SrcSpan.newDiagErr(ErrCodeExpectedFoo, "one or more expressions"))
			case 1:
				fn.Body = body_list.Items[0]
			default:
				f, p, s := call.Args[1].From, call.Args[1], fn.Scope
				expr_do := &SemExpr{From: f, Parent: p, Scope: s, Val: &SemValCall{
					Callee: &SemExpr{Val: &SemValIdent{Ident: moPrimOpDo}, From: f, Parent: p, Scope: s},
					Args:   SemExprs{{Val: body_list, From: f, Parent: p, Scope: s}}}}
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

func (me *SrcPack) semTypingPrimOpFnOrMacro(ctx *semTypeInfer, expr *SemExpr, tyExpected SemType, env map[MoValIdent]SemType) *SemExpr {
	expr.ErrsOwn.Add(expr.From.SrcSpan.newDiagErr(ErrCodeAtmoTodo, "new bug intro'd: encountered `@fn` or `@macro` call in type-inference"))
	return expr.with(expr.newUntypable(), nil)
}

func (me *SrcPack) semTypingPrimOpFnCall(ctx *semTypeInfer, expr *SemExpr, tyExpected SemType, env map[MoValIdent]SemType) *SemExpr {
	expr.ErrsOwn.Add(expr.From.SrcSpan.newDiagErr(ErrCodeAtmoTodo, "semTypingPrimOpCaseOf"))
	return expr.with(expr.newUntypable(), nil)
	// call := expr.Val.(*SemValCall)
	// if me.semCheckCount(2, 2, call.Args, expr, true) {
	// 	if callee, call_args := call.Args[0], semCheckIs[SemValList](MoPrimTypeList, call.Args[1]); call_args != nil {
	// 		return ctx.inferForCallWith(me, env, expr, callee, call_args.Items...)
	// 	}
	// }
	// return expr.with(expr.newUntypable(),nil)
}

func (me *SrcPack) semTypingPrimOpSet(ctx *semTypeInfer, expr *SemExpr, tyExpected SemType, env map[MoValIdent]SemType) *SemExpr {
	expr.ErrsOwn.Add(expr.From.SrcSpan.newDiagErr(ErrCodeAtmoTodo, "semTypingPrimOpSet"))
	return expr.with(expr.newUntypable(), nil)
}

func (me *SrcPack) semTypingPrimOpCaseOf(ctx *semTypeInfer, expr *SemExpr, tyExpected SemType, env map[MoValIdent]SemType) *SemExpr {
	expr.ErrsOwn.Add(expr.From.SrcSpan.newDiagErr(ErrCodeAtmoTodo, "semTypingPrimOpCaseOf"))
	return expr.with(expr.newUntypable(), nil)
}

func (me *SrcPack) semTypingPrimOpDo(ctx *semTypeInfer, expr *SemExpr, tyExpected SemType, env map[MoValIdent]SemType) *SemExpr {
	expr.ErrsOwn.Add(expr.From.SrcSpan.newDiagErr(ErrCodeAtmoTodo, "semTypingPrimOpDo"))
	return expr.with(expr.newUntypable(), nil)
}

func (me *SrcPack) semTypingPrimOpExpand(ctx *semTypeInfer, expr *SemExpr, tyExpected SemType, env map[MoValIdent]SemType) *SemExpr {
	expr.ErrsOwn.Add(expr.From.SrcSpan.newDiagErr(ErrCodeAtmoTodo, "semTypingPrimOpExpand"))
	return expr.with(expr.newUntypable(), nil)
}

func (me *SrcPack) semTypingPrimOpQQuote(ctx *semTypeInfer, expr *SemExpr, tyExpected SemType, env map[MoValIdent]SemType) *SemExpr {
	expr.ErrsOwn.Add(expr.From.SrcSpan.newDiagErr(ErrCodeAtmoTodo, "semTypingPrimOpQQuote"))
	return expr.with(expr.newUntypable(), nil)
}

func (me *SrcPack) semTypingPrimOpQuote(ctx *semTypeInfer, expr *SemExpr, tyExpected SemType, env map[MoValIdent]SemType) *SemExpr {
	expr.ErrsOwn.Add(expr.From.SrcSpan.newDiagErr(ErrCodeAtmoTodo, "semTypingPrimOpQuote"))
	return expr.with(expr.newUntypable(), nil)
}

func (me *SrcPack) semTypingPrimOpSpliceUnquote(ctx *semTypeInfer, expr *SemExpr, tyExpected SemType, env map[MoValIdent]SemType) *SemExpr {
	expr.ErrsOwn.Add(expr.From.SrcSpan.newDiagErr(ErrCodeAtmoTodo, "semTypingPrimOpSpliceUnquote"))
	return expr.with(expr.newUntypable(), nil)
}

func (me *SrcPack) semTypingPrimOpUnquote(ctx *semTypeInfer, expr *SemExpr, tyExpected SemType, env map[MoValIdent]SemType) *SemExpr {
	expr.ErrsOwn.Add(expr.From.SrcSpan.newDiagErr(ErrCodeAtmoTodo, "semTypingPrimOpUnquote"))
	return expr.with(expr.newUntypable(), nil)
}

func (me *SrcPack) semTypingPrimFnReplPrint(ctx *semTypeInfer, expr *SemExpr, tyExpected SemType, env map[MoValIdent]SemType) *SemExpr {
	expr.ErrsOwn.Add(expr.From.SrcSpan.newDiagErr(ErrCodeAtmoTodo, "semTypingPrimFnReplPrint"))
	return expr.with(expr.newUntypable(), nil)
}

func (me *SrcPack) semTypingPrimFnCast(ctx *semTypeInfer, expr *SemExpr, tyExpected SemType, env map[MoValIdent]SemType) *SemExpr {
	expr.ErrsOwn.Add(expr.From.SrcSpan.newDiagErr(ErrCodeAtmoTodo, "semTypingPrimFnCast"))
	return expr.with(expr.newUntypable(), nil)
}

func (me *SrcPack) semTypingPrimFnEq(ctx *semTypeInfer, expr *SemExpr, tyExpected SemType, env map[MoValIdent]SemType) *SemExpr {
	expr.ErrsOwn.Add(expr.From.SrcSpan.newDiagErr(ErrCodeAtmoTodo, "semTypingPrimFnEq"))
	return expr.with(expr.newUntypable(), nil)
}

func (me *SrcPack) semTypingPrimFnNeq(ctx *semTypeInfer, expr *SemExpr, tyExpected SemType, env map[MoValIdent]SemType) *SemExpr {
	expr.ErrsOwn.Add(expr.From.SrcSpan.newDiagErr(ErrCodeAtmoTodo, "semTypingPrimFnNeq"))
	return expr.with(expr.newUntypable(), nil)
}

func (me *SrcPack) semTypingPrimFnGeq(ctx *semTypeInfer, expr *SemExpr, tyExpected SemType, env map[MoValIdent]SemType) *SemExpr {
	expr.ErrsOwn.Add(expr.From.SrcSpan.newDiagErr(ErrCodeAtmoTodo, "semTypingPrimFnGeq"))
	return expr.with(expr.newUntypable(), nil)
}

func (me *SrcPack) semTypingPrimFnLeq(ctx *semTypeInfer, expr *SemExpr, tyExpected SemType, env map[MoValIdent]SemType) *SemExpr {
	expr.ErrsOwn.Add(expr.From.SrcSpan.newDiagErr(ErrCodeAtmoTodo, "semTypingPrimFnLeq"))
	return expr.with(expr.newUntypable(), nil)
}

func (me *SrcPack) semTypingPrimFnLt(ctx *semTypeInfer, expr *SemExpr, tyExpected SemType, env map[MoValIdent]SemType) *SemExpr {
	expr.ErrsOwn.Add(expr.From.SrcSpan.newDiagErr(ErrCodeAtmoTodo, "semTypingPrimFnLt"))
	return expr.with(expr.newUntypable(), nil)
}

func (me *SrcPack) semTypingPrimFnGt(ctx *semTypeInfer, expr *SemExpr, tyExpected SemType, env map[MoValIdent]SemType) *SemExpr {
	expr.ErrsOwn.Add(expr.From.SrcSpan.newDiagErr(ErrCodeAtmoTodo, "semTypingPrimFnGt"))
	return expr.with(expr.newUntypable(), nil)
}

func (me *SrcPack) semTypingPrimFnPrimTypeTag(ctx *semTypeInfer, expr *SemExpr, tyExpected SemType, env map[MoValIdent]SemType) *SemExpr {
	expr.ErrsOwn.Add(expr.From.SrcSpan.newDiagErr(ErrCodeAtmoTodo, "semTypingPrimFnPrimTypeTag"))
	return expr.with(expr.newUntypable(), nil)
}

func (me *SrcPack) semTypingPrimFnListGet(ctx *semTypeInfer, expr *SemExpr, tyExpected SemType, env map[MoValIdent]SemType) *SemExpr {
	expr.ErrsOwn.Add(expr.From.SrcSpan.newDiagErr(ErrCodeAtmoTodo, "semTypingPrimFnListGet"))
	return expr.with(expr.newUntypable(), nil)
}

func (me *SrcPack) semTypingPrimFnListSet(ctx *semTypeInfer, expr *SemExpr, tyExpected SemType, env map[MoValIdent]SemType) *SemExpr {
	expr.ErrsOwn.Add(expr.From.SrcSpan.newDiagErr(ErrCodeAtmoTodo, "semTypingPrimFnListSet"))
	return expr.with(expr.newUntypable(), nil)
}

func (me *SrcPack) semTypingPrimFnListRange(ctx *semTypeInfer, expr *SemExpr, tyExpected SemType, env map[MoValIdent]SemType) *SemExpr {
	expr.ErrsOwn.Add(expr.From.SrcSpan.newDiagErr(ErrCodeAtmoTodo, "semTypingPrimFnListRange"))
	return expr.with(expr.newUntypable(), nil)
}

func (me *SrcPack) semTypingPrimFnListLen(ctx *semTypeInfer, expr *SemExpr, tyExpected SemType, env map[MoValIdent]SemType) *SemExpr {
	expr.ErrsOwn.Add(expr.From.SrcSpan.newDiagErr(ErrCodeAtmoTodo, "semTypingPrimFnListLen"))
	return expr.with(expr.newUntypable(), nil)
}

func (me *SrcPack) semTypingPrimFnListConcat(ctx *semTypeInfer, expr *SemExpr, tyExpected SemType, env map[MoValIdent]SemType) *SemExpr {
	expr.ErrsOwn.Add(expr.From.SrcSpan.newDiagErr(ErrCodeAtmoTodo, "semTypingPrimFnListConcat"))
	return expr.with(expr.newUntypable(), nil)
}

func (me *SrcPack) semTypingPrimFnDictHas(ctx *semTypeInfer, expr *SemExpr, tyExpected SemType, env map[MoValIdent]SemType) *SemExpr {
	expr.ErrsOwn.Add(expr.From.SrcSpan.newDiagErr(ErrCodeAtmoTodo, "semTypingPrimFnDictHas"))
	return expr.with(expr.newUntypable(), nil)
}

func (me *SrcPack) semTypingPrimFnDictGet(ctx *semTypeInfer, expr *SemExpr, tyExpected SemType, env map[MoValIdent]SemType) *SemExpr {
	expr.ErrsOwn.Add(expr.From.SrcSpan.newDiagErr(ErrCodeAtmoTodo, "semTypingPrimFnDictGet"))
	return expr.with(expr.newUntypable(), nil)
}

func (me *SrcPack) semTypingPrimFnDictSet(ctx *semTypeInfer, expr *SemExpr, tyExpected SemType, env map[MoValIdent]SemType) *SemExpr {
	expr.ErrsOwn.Add(expr.From.SrcSpan.newDiagErr(ErrCodeAtmoTodo, "semTypingPrimFnDictSet"))
	return expr.with(expr.newUntypable(), nil)
}

func (me *SrcPack) semTypingPrimFnDictDel(ctx *semTypeInfer, expr *SemExpr, tyExpected SemType, env map[MoValIdent]SemType) *SemExpr {
	expr.ErrsOwn.Add(expr.From.SrcSpan.newDiagErr(ErrCodeAtmoTodo, "semTypingPrimFnDictDel"))
	return expr.with(expr.newUntypable(), nil)
}

func (me *SrcPack) semTypingPrimFnDictLen(ctx *semTypeInfer, expr *SemExpr, tyExpected SemType, env map[MoValIdent]SemType) *SemExpr {
	expr.ErrsOwn.Add(expr.From.SrcSpan.newDiagErr(ErrCodeAtmoTodo, "semTypingPrimFnDictLen"))
	return expr.with(expr.newUntypable(), nil)
}

func (me *SrcPack) semTypingPrimFnErrNew(ctx *semTypeInfer, expr *SemExpr, tyExpected SemType, env map[MoValIdent]SemType) *SemExpr {
	expr.ErrsOwn.Add(expr.From.SrcSpan.newDiagErr(ErrCodeAtmoTodo, "semTypingPrimFnErrNew"))
	return expr.with(expr.newUntypable(), nil)
}

func (me *SrcPack) semTypingPrimFnErrVal(ctx *semTypeInfer, expr *SemExpr, tyExpected SemType, env map[MoValIdent]SemType) *SemExpr {
	expr.ErrsOwn.Add(expr.From.SrcSpan.newDiagErr(ErrCodeAtmoTodo, "semTypingPrimFnErrVal"))
	return expr.with(expr.newUntypable(), nil)
}

func (me *SrcPack) semTypingPrimFnStr(ctx *semTypeInfer, expr *SemExpr, tyExpected SemType, env map[MoValIdent]SemType) *SemExpr {
	expr.ErrsOwn.Add(expr.From.SrcSpan.newDiagErr(ErrCodeAtmoTodo, "semTypingPrimFnStr"))
	return expr.with(expr.newUntypable(), nil)
}

func (me *SrcPack) semTypingPrimFnExprStr(ctx *semTypeInfer, expr *SemExpr, tyExpected SemType, env map[MoValIdent]SemType) *SemExpr {
	expr.ErrsOwn.Add(expr.From.SrcSpan.newDiagErr(ErrCodeAtmoTodo, "semTypingPrimFnExprStr"))
	return expr.with(expr.newUntypable(), nil)
}

func (me *SrcPack) semTypingPrimFnExprParse(ctx *semTypeInfer, expr *SemExpr, tyExpected SemType, env map[MoValIdent]SemType) *SemExpr {
	expr.ErrsOwn.Add(expr.From.SrcSpan.newDiagErr(ErrCodeAtmoTodo, "semTypingPrimFnExprParse"))
	return expr.with(expr.newUntypable(), nil)
}

func (me *SrcPack) semTypingPrimFnExprEval(ctx *semTypeInfer, expr *SemExpr, tyExpected SemType, env map[MoValIdent]SemType) *SemExpr {
	expr.ErrsOwn.Add(expr.From.SrcSpan.newDiagErr(ErrCodeAtmoTodo, "semTypingPrimFnExprEval"))
	return expr.with(expr.newUntypable(), nil)
}
