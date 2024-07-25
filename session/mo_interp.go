package session

import (
	"atmo/util"
	"atmo/util/str"
)

type Interp struct {
	SrcFile        *SrcFile
	Env            *MoEnv
	StackTraces    bool
	LastStackTrace []*MoExpr
	diagCtxCall    *MoExpr // set to a full call-expr just before it is entered into, for use in producing that call's error (if any) unwinding the whole eval
}

func newInterp(srcFile *SrcFile) *Interp {
	return &Interp{Env: newMoEnv(&rootEnv, nil, nil), SrcFile: srcFile}
}

func (me *Interp) ClearStackTrace() {
	me.LastStackTrace = me.LastStackTrace[:0] // keeps currently-already-alloc'd capacity, for reduced GC churn and reduced alloc times
}

func (me *Interp) Eval(expr *MoExpr) (*MoExpr, *SrcFileNotice) {
	me.ClearStackTrace()
	me.diagCtxCall = nil
	return me.evalAndApply(me.Env, expr)
}

func (me *Interp) evalAndApply(env *MoEnv, expr *MoExpr) (*MoExpr, *SrcFileNotice) {
	var err *SrcFileNotice
	diag_ctx_orig := me.diagCtxCall
	// id := strconv.FormatInt(time.Now().UnixNano(), 36) // uncomment and print `id` to check for TCO loop
	for (err == nil) && (env != nil) {
		if _, is_call := expr.Val.(moValCall); !is_call {
			expr, err = me.evalExpr(env, expr)
			env = nil
		} else if expr, err = me.macroExpand(env, expr); err != nil {
			return nil, err
		} else if call, is_call := expr.Val.(moValCall); is_call {
			callee, call_args := call[0], ([]*MoExpr)(call[1:])

			var prim_op_lazy moFnLazy
			if ident, _ := callee.Val.(moValIdent); ident != "" {
				prim_op_lazy = moPrimOpsLazy[ident]
			}

			if prim_op_lazy != nil {
				diag_ctx_cur, diag_ctx_prev := expr, me.diagCtxCall
				me.diagCtxCall = diag_ctx_cur
				if env, expr, err = prim_op_lazy(me, env, call_args...); err != nil {
					return nil, err
				}
				expr.setSrcSpanIfNone(diag_ctx_cur)
				me.diagCtxCall = diag_ctx_prev
			} else {
				if me.StackTraces {
					me.LastStackTrace = append(me.LastStackTrace, expr)
				}
				diag_ctx_cur, diag_ctx_prev := expr, me.diagCtxCall
				me.diagCtxCall = diag_ctx_cur
				if expr, err = me.evalExpr(env, expr); err != nil {
					return nil, err
				}
				me.diagCtxCall = diag_ctx_cur
				call = expr.Val.(moValCall)
				callee, call_args = call[0], ([]*MoExpr)(call[1:])
				switch fn := callee.Val.(type) {
				default:
					return nil, me.diagSpan(true, false).newDiagErr(NoticeCodeUncallable, callee.String())
				case moValFnPrim:
					if expr, err = fn(me, env, call_args...); err != nil {
						return nil, err
					}
					expr.setSrcSpanIfNone(diag_ctx_cur)
					env, me.diagCtxCall = nil, diag_ctx_prev
				case *moValFnLam:
					expr = fn.body
					env, err = me.envWith(fn, call_args)
					if err != nil {
						return nil, err
					}
				}
			}
		}
	}
	me.diagCtxCall = diag_ctx_orig
	return expr, err
}

func (me *Interp) evalExpr(env *MoEnv, expr *MoExpr) (*MoExpr, *SrcFileNotice) {
	switch val := expr.Val.(type) {
	case moValIdent:
		if (val[0] != '@') || (moPrimIdents[val] == nil) { // using prim idents as values (outside of stdlib) would be obscurely-rare or more likely mistaken, so the map-lookup on `@` prefix is OK
			found := env.lookup(val)
			if found == nil {
				return nil, me.diagSpan(false, true, expr).newDiagErr(NoticeCodeUndefined, val)
			} else if found.SrcSpan == nil {
				return &MoExpr{Val: found.Val, SrcSpan: expr.SrcSpan}, nil
			}
			return found, nil
		} // else: prefer to return expr itself so that there's a better-fitting SrcNode for diags
	case moValList:
		list := make(moValList, len(val))
		for i, item := range val {
			it, err := me.evalAndApply(env, item)
			if err != nil {
				return nil, err
			}
			list[i] = it
		}
		return &MoExpr{Val: list, SrcSpan: expr.SrcSpan}, nil
	case moValRec:
		rec := make(moValRec, len(val))
		for k, v := range val {
			key, err := me.evalAndApply(env, k)
			if err != nil {
				return nil, err
			}
			val, err := me.evalAndApply(env, v)
			if err != nil {
				return nil, err
			}
			rec[key] = val
		}
		return &MoExpr{Val: rec, SrcSpan: expr.SrcSpan}, nil
	case moValCall:
		call := make(moValCall, len(val))
		for i, item := range val {
			it, err := me.evalAndApply(env, item)
			if err != nil {
				return nil, err
			}
			call[i] = it
		}
		return &MoExpr{Val: call, SrcSpan: expr.SrcSpan}, nil
	}
	return expr, nil
}

func (me *Interp) diagSpan(preferCalleeOverCall bool, preferTheseEvenMore bool, have ...*MoExpr) (ret *SrcFileSpan) {
	if me.diagCtxCall != nil {
		ret = me.diagCtxCall.SrcSpan
		if callee := me.diagCtxCall.Val.(moValCall)[0]; preferCalleeOverCall && (callee.SrcSpan != nil) {
			ret = callee.SrcSpan
		}
		if ret == nil {
			for _, expr := range me.diagCtxCall.Val.(moValCall) {
				if ret = expr.SrcSpan; ret != nil {
					break
				}
			}
		}
	}
	if preferTheseEvenMore || (ret == nil) {
		for _, expr := range have {
			if expr.SrcSpan != nil {
				return expr.SrcSpan
			}
		}
	}
	return
}

func (me *Interp) callWithDiagCtxSet(env *MoEnv, fnOrFuncExpr *MoExpr, args ...*MoExpr) (*MoExpr, *SrcFileNotice) {
	util.Assert(me.diagCtxCall != nil, nil)
	switch fn := fnOrFuncExpr.Val.(type) {
	case moValFnPrim:
		return fn(me, env, args...)
	case *moValFnLam:
		env, err := me.envWith(fn, args)
		if err != nil {
			return nil, err
		}
		return me.evalAndApply(env, fn.body)
	}
	callee := me.diagCtxCall.Callee()
	return nil, me.diagSpan(true, false).newDiagErr(NoticeCodeUncallable, callee.String())
}

func (me *Interp) envWith(fn *moValFnLam, args []*MoExpr) (*MoEnv, *SrcFileNotice) {
	if err := me.checkCount(len(fn.params), len(fn.params), args); err != nil {
		return nil, err
	}
	return newMoEnv(fn.env, fn.params, args), nil
}

func (me *Interp) macroExpand(env *MoEnv, expr *MoExpr) (*MoExpr, *SrcFileNotice) {
	diag_ctx := me.diagCtxCall
	for fn := expr.macroCallCallee(env); fn != nil; fn = expr.macroCallCallee(env) {
		me.diagCtxCall = expr
		it, err := me.callWithDiagCtxSet(env, fn, expr.Val.(moValCall)[1:]...)
		if err != nil {
			return nil, err
		}
		it.setSrcSpanIfNone(expr)
		expr = it
	}
	me.diagCtxCall = diag_ctx
	return expr, nil
}

func (me *MoExpr) macroCallCallee(env *MoEnv) *MoExpr {
	if call, is := me.Val.(moValCall); is {
		if ident, _ := call[0].Val.(moValIdent); ident != "" {
			if expr := env.lookup(ident); expr != nil {
				if fn, _ := expr.Val.(*moValFnLam); fn != nil && fn.isMacro {
					return expr
				}
			}
		}
	}
	return nil
}

func (me *Interp) checkCount(wantAtLeast int, wantAtMost int, have []*MoExpr) *SrcFileNotice {
	return me.checkCountWithSrcSpan(wantAtLeast, wantAtMost, have, false)
}

func (me *Interp) checkCountWithSrcSpan(wantAtLeast int, wantAtMost int, have []*MoExpr, preferSrcSpan bool) *SrcFileNotice {
	diag_src_span := me.diagSpan(false, preferSrcSpan, have...)
	moniker := util.If(preferSrcSpan, "item", "arg")
	if wantAtLeast < 0 {
		return nil
	} else if (wantAtLeast == wantAtMost) && (wantAtLeast != len(have)) {
		return diag_src_span.newDiagErr(NoticeCodeExpectedFoo, str.Fmt("%d %s(s), not %d", wantAtLeast, moniker, len(have)))
	} else if len(have) < wantAtLeast {
		return diag_src_span.newDiagErr(NoticeCodeExpectedFoo, str.Fmt("at least %d %s(s), not %d", wantAtLeast, moniker, len(have)))
	} else if (wantAtMost > wantAtLeast) && (len(have) > wantAtMost) {
		return diag_src_span.newDiagErr(NoticeCodeExpectedFoo, str.Fmt("%d to %d %s(s), not %d", wantAtLeast, wantAtMost, moniker, len(have)))
	}
	return nil
}

func (me *Interp) checkIs(want MoValPrimType, have *MoExpr) *SrcFileNotice {
	if have_type := have.Val.primType(); have_type != want {
		have_str := util.If(have_type == MoPrimTypeFunc, me.SrcFile.srcAt(have.SrcSpan, '`'), "`"+have.String()+"`")
		return me.diagSpan(false, true, have).newDiagErr(NoticeCodeExpectedFoo, str.Fmt("%s instead of %s %s", want.Str(true), have_type.Str(true), have_str))
	}
	return nil
}

func (me *Interp) checkIsListOf(of MoValPrimType, expr *MoExpr) *SrcFileNotice {
	if err := me.checkIs(MoPrimTypeList, expr); err != nil {
		return err
	}
	if of >= 0 {
		return me.check(of, -1, -1, expr.Val.(moValList)...)
	}
	return nil
}

func (me *Interp) checkIsCallOnIdent(call *MoExpr, ident moValIdent, errIfNumArgsNot int) (bool, *SrcFileNotice) {
	if call, is := call.Val.(moValCall); is {
		if callee, _ := call[0].Val.(moValIdent); callee == ident {
			return true, me.checkCount(errIfNumArgsNot, errIfNumArgsNot, call[1:])
		}
	}
	return false, nil
}

func (me *Interp) check(want MoValPrimType, wantAtLeast int, wantAtMost int, have ...*MoExpr) *SrcFileNotice {
	if err := me.checkCount(wantAtLeast, wantAtMost, have); err != nil {
		return err
	}
	for _, expr := range have {
		if err := me.checkIs(want, expr); err != nil {
			return err
		}
	}
	return nil
}
