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
	interp := &Interp{Env: newMoEnv(nil, nil, nil), SrcFile: srcFile}
	for prim_op_name, prim_op_func := range moPrimOpsEager {
		interp.Env.set(prim_op_name, &MoExpr{Val: moValFnPrim(prim_op_func)})
	}
	for prim_ident_name, prim_ident_expr := range moPrimIdents {
		interp.Env.set(prim_ident_name, prim_ident_expr)
	}
	return interp
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
				diag_ctx_prev := me.diagCtxCall
				me.diagCtxCall = expr
				if env, expr, err = prim_op_lazy(me, env, call_args...); err != nil {
					return nil, err
				}
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
					return nil, me.diagNode(true, false).newDiagErr(false, NoticeCodeUncallable, callee.String())
				case moValFnPrim:
					if expr, err = fn(me, call_args...); err != nil {
						return nil, err
					}
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
		found := env.lookup(val)
		if found == nil {
			return nil, me.diagNode(false, true, expr).newDiagErr(false, NoticeCodeUndefined, val)
		}
		return found, nil
	case moValArr:
		arr := make(moValArr, len(val))
		for i, item := range val {
			it, err := me.evalAndApply(env, item)
			if err != nil {
				return nil, err
			}
			arr[i] = it
		}
		return &MoExpr{Val: arr, SrcNode: expr.SrcNode}, nil
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
		return &MoExpr{Val: rec, SrcNode: expr.SrcNode}, nil
	case moValCall:
		call := make(moValCall, len(val))
		for i, item := range val {
			it, err := me.evalAndApply(env, item)
			if err != nil {
				return nil, err
			}
			call[i] = it
		}
		return &MoExpr{Val: call, SrcNode: expr.SrcNode}, nil
	}
	return expr, nil
}

func (me *Interp) diagNode(preferCalleeOverCall bool, preferTheseEvenMore bool, have ...*MoExpr) (ret *AstNode) {
	if me.diagCtxCall != nil {
		ret = me.diagCtxCall.SrcNode
		if callee := me.diagCtxCall.Val.(moValCall)[0]; preferCalleeOverCall && (callee.SrcNode != nil) {
			ret = callee.SrcNode
		}
		if ret == nil {
			for _, expr := range me.diagCtxCall.Val.(moValCall) {
				if ret = expr.SrcNode; ret != nil {
					break
				}
			}
		}
	}
	if preferTheseEvenMore || (ret == nil) {
		for _, expr := range have {
			if expr.SrcNode != nil {
				return expr.SrcNode
			}
		}
	}
	return
}

func (me *Interp) callWithDiagCtxSet(fnOrFuncExpr *MoExpr, args ...*MoExpr) (*MoExpr, *SrcFileNotice) {
	util.Assert(me.diagCtxCall != nil, nil)
	switch fn := fnOrFuncExpr.Val.(type) {
	case moValFnPrim:
		return fn(me, args...)
	case *moValFnLam:
		env, err := me.envWith(fn, args)
		if err != nil {
			return nil, err
		}
		return me.evalAndApply(env, fn.body)
	}
	callee := me.diagCtxCall.Callee()
	return nil, me.diagNode(true, false).newDiagErr(false, NoticeCodeUncallable, callee.String())
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
		it, err := me.callWithDiagCtxSet(fn, expr.Val.(moValCall)[1:]...)
		if err != nil {
			return nil, err
		}
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
	diag_src_node := me.diagNode(false, false, have...)
	if wantAtLeast < 0 {
		return nil
	} else if (wantAtLeast == wantAtMost) && (wantAtLeast != len(have)) {
		return diag_src_node.newDiagErr(false, NoticeCodeExpectedFoo, str.Fmt("%d arg(s), not %d", wantAtLeast, len(have)))
	} else if len(have) < wantAtLeast {
		return diag_src_node.newDiagErr(false, NoticeCodeExpectedFoo, str.Fmt("at least %d arg(s), not %d", wantAtLeast, len(have)))
	} else if (wantAtMost > wantAtLeast) && (len(have) > wantAtMost) {
		return diag_src_node.newDiagErr(false, NoticeCodeExpectedFoo, str.Fmt("%d to %d arg(s), not %d", wantAtLeast, wantAtMost, len(have)))
	}
	return nil
}

func (me *Interp) checkIs(want MoValType, have *MoExpr) *SrcFileNotice {
	if have_type := have.Val.valType(); have_type != want {
		return me.diagNode(false, true, have).newDiagErr(false, NoticeCodeExpectedFoo, str.Fmt("`%s`, not `%s`", want, have_type))
	}
	return nil
}

func (me *Interp) checkAre(want MoValType, wantAtLeast int, wantAtMost int, have ...*MoExpr) *SrcFileNotice {
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
