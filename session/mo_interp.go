package session

import (
	"atmo/util"
	"atmo/util/str"
)

type Evaler interface {
	eval(ctx *Interp, expr *MoExpr) (*MoExpr, *SrcFileNotice)
}

type Interp struct {
	SrcFile        *SrcFile
	Env            *MoEnv
	evaler         Evaler
	StackTraces    bool
	LastStackTrace []*MoExpr
}

type DefaultEvaler struct {
	ctx         *Interp
	diagCtxCall *MoExpr // set to a full call-expr just before it is entered into, for use in producing that call's error (if any) unwinding the whole eval
}

func newInterp(srcFile *SrcFile, evaler Evaler) *Interp {
	if evaler == nil {
		evaler = &DefaultEvaler{}
	}
	interp := &Interp{Env: newMoEnv(nil, nil, nil), SrcFile: srcFile, evaler: evaler}

	return interp
}

func (me *Interp) Eval(expr *MoExpr) (*MoExpr, *SrcFileNotice) {
	me.LastStackTrace = me.LastStackTrace[:0] // keeps currently-already-alloc'd capacity, for reduced GC churn and reduced alloc times
	return me.evaler.eval(me, expr)
}

func (me *DefaultEvaler) eval(ctx *Interp, expr *MoExpr) (*MoExpr, *SrcFileNotice) {
	me.ctx, me.diagCtxCall = ctx, nil
	return me.evalAndApply(ctx.Env, expr)
}

func (me *DefaultEvaler) evalAndApply(env *MoEnv, expr *MoExpr) (*MoExpr, *SrcFileNotice) {
	var err *SrcFileNotice
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
				me.diagCtxCall = expr
				if env, expr, err = prim_op_lazy(me.ctx, env, call_args...); err != nil {
					return nil, err
				}
			} else {
				if me.ctx.StackTraces {
					me.ctx.LastStackTrace = append(me.ctx.LastStackTrace, expr)
				}
				diag_ctx := expr
				me.diagCtxCall = diag_ctx
				if expr, err = me.evalExpr(env, expr); err != nil {
					return nil, err
				}
				me.diagCtxCall = diag_ctx
				call = expr.Val.(moValCall)
				callee, call_args = call[0], ([]*MoExpr)(call[1:])
				switch fn := callee.Val.(type) {
				default:
					return nil, callee.SrcNode.newDiagErr(false, NoticeCodeUncallable, callee.String())
				case moValFnPrim:
					if expr, err = fn(call_args...); err != nil {
						return nil, err
					}
					env = nil
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
	return expr, err
}

func (me *DefaultEvaler) evalExpr(env *MoEnv, expr *MoExpr) (*MoExpr, *SrcFileNotice) {
	switch val := expr.Val.(type) {
	case moValIdent:
		found := env.lookup(val)
		if found == nil {
			return nil, expr.SrcNode.newDiagErr(false, NoticeCodeUndefined, val)
		}
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

func (me *DefaultEvaler) callWithDiagCtxSet(fnOrFuncExpr *MoExpr, args ...*MoExpr) (*MoExpr, *SrcFileNotice) {
	util.Assert(me.diagCtxCall != nil, nil)
	switch fn := fnOrFuncExpr.Val.(type) {
	case moValFnPrim:
		return fn(args...)
	case *moValFnLam:
		env, err := me.envWith(fn, args)
		if err != nil {
			return nil, err
		}
		return me.evalAndApply(env, fn.body)
	}
	callee := me.diagCtxCall.Callee()
	return nil, callee.SrcNode.newDiagErr(false, NoticeCodeUncallable, callee.String())
}

func (me *DefaultEvaler) macroExpand(env *MoEnv, expr *MoExpr) (*MoExpr, *SrcFileNotice) {
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

func (me *DefaultEvaler) checkCount(wantAtLeast int, wantAtMost int, have []*MoExpr) *SrcFileNotice {
	if wantAtLeast < 0 {
		return nil
	} else if (wantAtLeast == wantAtMost) && (wantAtLeast != len(have)) {
		return me.diagCtxCall.SrcNode.newDiagErr(false, NoticeCodeExpectedFoo, str.Fmt("%d arg(s), not %d", wantAtLeast, len(have)))
	} else if len(have) < wantAtLeast {
		return me.diagCtxCall.SrcNode.newDiagErr(false, NoticeCodeExpectedFoo, str.Fmt("at least %d arg(s), not %d", wantAtLeast, len(have)))
	} else if (wantAtMost > wantAtLeast) && (len(have) > wantAtMost) {
		return me.diagCtxCall.SrcNode.newDiagErr(false, NoticeCodeExpectedFoo, str.Fmt("%d to %d arg(s), not %d", wantAtLeast, wantAtMost, len(have)))
	}
	return nil
}

func (*DefaultEvaler) checkIs(want MoValType, have *MoExpr) *SrcFileNotice {
	if have_type := have.Val.valType(); have_type != want {
		return have.SrcNode.newDiagErr(false, NoticeCodeExpectedFoo, str.Fmt("`%s`, not `%s`", want, have_type))
	}
	return nil
}

func (me *DefaultEvaler) checkAre(want MoValType, have ...*MoExpr) *SrcFileNotice {
	for _, expr := range have {
		if err := me.checkIs(want, expr); err != nil {
			return err
		}
	}
	return nil
}

func (me *DefaultEvaler) checkAreBoth(want MoValType, have []*MoExpr, exactArgsCount bool) (err *SrcFileNotice) {
	max_args_count := -1
	if exactArgsCount {
		max_args_count = 2
	}
	if err = me.checkCount(2, max_args_count, have); err == nil {
		if err = me.checkIs(want, have[0]); err == nil {
			err = me.checkIs(want, have[1])
		}
	}
	return
}
