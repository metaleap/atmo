package session

import (
	"io"
	"os"

	"atmo/util"
	"atmo/util/sl"
	"atmo/util/str"
)

type Writer interface {
	io.Writer
	io.StringWriter
}

type Interp struct {
	Pack           *SrcPack
	replFauxFile   *SrcFile
	Env            *MoEnv
	StackTraces    bool
	LastStackTrace MoExprs
	StdIo          struct {
		In  io.Reader
		Out Writer
		Err Writer
	}
	diagCtxCall *MoExpr // set to a full call-expr just before it is entered into, for use in producing that call's error (if any) unwinding the whole eval
}

func newInterp(inPack *SrcPack, replFauxFile *SrcFile) *Interp {
	me := Interp{Env: newMoEnv(&rootEnv, nil, nil), Pack: inPack, replFauxFile: replFauxFile}
	me.StdIo.In, me.StdIo.Out, me.StdIo.Err = os.Stdin, os.Stdout, os.Stderr
	me.ensureRootEnvPopulated()
	me.Pack.Sema.Eval = &me
	return &me
}

func (me *Interp) reset() {
	LockedDo(func(sess StateAccess) {
		allNotices = map[string]sl.Of[*SrcFileNotice]{}
		me.ClearStackTrace()
		me.Env = newMoEnv(&rootEnv, nil, nil)
		me.Pack.Sema.Eval, me.Pack.Sema.Pre = me, nil
		for _, src_file := range me.Pack.Files {
			src_file.notices.LexErrs, src_file.notices.LastReadErr, src_file.notices.Sema, src_file.Src.Ast, src_file.Src.Toks, src_file.Src.Text =
				nil, nil, nil, nil, nil, ""
		}
		_ = ensureSrcFiles(nil, false, me.Pack.srcFilePaths()...)
		me.Pack.refreshSema()
		refreshAndPublishNotices(true, me.Pack.srcFilePaths()...)
	})
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
	diag_ctx_orig, expr_orig := me.diagCtxCall, expr
	// id := strconv.FormatInt(time.Now().UnixNano(), 36) // uncomment and print `id` to check for TCO loop
	for (err == nil) && (env != nil) {
		if _, is_call := expr.Val.(MoValCall); !is_call {
			expr, err = me.evalExpr(env, expr)
			env = nil
		} else if expr, err = me.macroExpand(env, expr); err != nil {
			return nil, err
		} else if call, is_call := expr.Val.(MoValCall); is_call {
			callee, call_args := call[0], (MoExprs)(call[1:])

			var prim_op_lazy moFnLazy
			if ident, _ := callee.Val.(MoValIdent); ident != "" {
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
				call = expr.Val.(MoValCall)
				callee, call_args = call[0], (MoExprs)(call[1:])
				switch fn := callee.Val.(type) {
				default:
					return nil, me.diagSpan(true, false).newDiagErr(NoticeCodeUncallable, callee.String())
				case MoValFnPrim:
					if expr, err = fn(me, env, call_args...); err != nil {
						return nil, err
					}
					expr.setSrcSpanIfNone(diag_ctx_cur)
					env, me.diagCtxCall = nil, diag_ctx_prev
				case *MoValFnLam:
					expr = fn.Body
					env, err = me.envWith(fn, call_args)
					if err != nil {
						return nil, err
					}
				}
			}
		}
	}
	me.diagCtxCall = diag_ctx_orig
	if expr != nil && ((expr.SrcFile == nil) || (expr.SrcSpan == nil)) {
		expr = &MoExpr{Val: expr.Val, SrcSpan: sl.FirstNonNil(expr.SrcSpan, expr_orig.SrcSpan), SrcFile: sl.FirstNonNil(expr.SrcFile, expr_orig.SrcFile)}
	}
	return expr, err
}

func (me *Interp) evalExpr(env *MoEnv, expr *MoExpr) (*MoExpr, *SrcFileNotice) {
	switch val := expr.Val.(type) {
	case MoValIdent:
		if (val[0] != '@') || (moPrimIdents[val] == nil) { // using prim idents as values (outside of stdlib) would be obscurely-rare or more likely mistaken, so the map-lookup on `@` prefix is OK
			found := env.lookup(val)
			if found == nil {
				return nil, me.diagSpan(false, true, expr).newDiagErr(NoticeCodeUndefined, val)
			}
			return me.expr(found.Val, expr.SrcFile, expr.SrcSpan), nil
		} // else: prefer to return expr itself so that there's a better-fitting SrcNode for diags
	case MoValList:
		list := make(MoValList, len(val))
		for i, item := range val {
			it, err := me.evalAndApply(env, item)
			if err != nil {
				return nil, err
			}
			list[i] = it
		}
		return me.expr(list, expr.SrcFile, expr.SrcSpan), nil
	case MoValDict:
		dict := make(MoValDict, 0, len(val))
		for _, pair := range val {
			k, v := pair[0], pair[1]
			key, err := me.evalAndApply(env, k)
			if err != nil {
				return nil, err
			}
			val, err := me.evalAndApply(env, v)
			if err != nil {
				return nil, err
			}
			if dict.Has(key) {
				return nil, k.SrcSpan.newDiagErr(NoticeCodeDictDuplKey, key)
			}
			dict.Set(key, val)
		}
		return me.expr(dict, expr.SrcFile, expr.SrcSpan), nil
	case MoValCall:
		call := make(MoValCall, len(val))
		for i, item := range val {
			it, err := me.evalAndApply(env, item)
			if err != nil {
				return nil, err
			}
			call[i] = it
		}
		return me.expr(call, expr.SrcFile, expr.SrcSpan), nil
	}
	return expr, nil
}

func (me *Interp) diagSpan(preferCalleeOverCall bool, preferTheseEvenMore bool, have ...*MoExpr) (ret *SrcFileSpan) {
	if me.diagCtxCall != nil {
		ret = me.diagCtxCall.SrcSpan
		if callee := me.diagCtxCall.Val.(MoValCall)[0]; preferCalleeOverCall && (callee.SrcSpan != nil) {
			ret = callee.SrcSpan
		}
		if ret == nil {
			for _, expr := range me.diagCtxCall.Val.(MoValCall) {
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
func (me *Interp) srcFile(preferCalleeOverCall bool, preferTheseEvenMore bool, have ...*MoExpr) (ret *SrcFile) {
	if me.diagCtxCall != nil {
		ret = me.diagCtxCall.SrcFile
		if callee := me.diagCtxCall.Val.(MoValCall)[0]; preferCalleeOverCall && (callee.SrcFile != nil) {
			ret = callee.SrcFile
		}
		if ret == nil {
			for _, expr := range me.diagCtxCall.Val.(MoValCall) {
				if ret = expr.SrcFile; ret != nil {
					break
				}
			}
		}
	}
	if preferTheseEvenMore || (ret == nil) {
		for _, expr := range have {
			if expr.SrcFile != nil {
				return expr.SrcFile
			}
		}
	}
	return
}

func (me *Interp) envWith(fn *MoValFnLam, args MoExprs) (*MoEnv, *SrcFileNotice) {
	if err := me.checkCount(len(fn.Params), len(fn.Params), args); err != nil {
		return nil, err
	}
	return newMoEnv(fn.Env, fn.Params, args), nil
}

func (me *Interp) macroExpand(env *MoEnv, expr *MoExpr) (*MoExpr, *SrcFileNotice) {
	diag_ctx := me.diagCtxCall
	for fn := expr.macroCallCallee(env); fn != nil; fn = expr.macroCallCallee(env) {
		me.diagCtxCall = expr

		fn := fn.Val.(*MoValFnLam)
		call_env, err := me.envWith(fn, MoExprs(expr.Val.(MoValCall)[1:]))
		if err != nil {
			return nil, err
		}
		it, err := me.evalAndApply(call_env, fn.Body)
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
	if call, is := me.Val.(MoValCall); is {
		if ident, _ := call[0].Val.(MoValIdent); ident != "" {
			if expr := env.lookup(ident); expr != nil {
				if fn, _ := expr.Val.(*MoValFnLam); fn != nil && fn.IsMacro {
					return expr
				}
			}
		}
	}
	return nil
}

func (me *Interp) checkCount(wantAtLeast int, wantAtMost int, have MoExprs) *SrcFileNotice {
	return me.checkCountWithSrcSpan(wantAtLeast, wantAtMost, have, false)
}

func (me *Interp) checkCountWithSrcSpan(wantAtLeast int, wantAtMost int, have MoExprs, preferSrcSpan bool) *SrcFileNotice {
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
	if have_type := have.Val.PrimType(); have_type != want {
		have_str := util.If(have_type == MoPrimTypeFunc, have.SrcFile.srcAt(have.SrcSpan, '`'), "`"+have.String()+"`")
		return me.diagSpan(false, true, have).newDiagErr(NoticeCodeExpectedFoo, str.Fmt("%s instead of %s %s", want.Str(true), have_type.Str(true), have_str))
	}
	return nil
}

func (me *Interp) checkIsListOf(of MoValPrimType, expr *MoExpr) *SrcFileNotice {
	if err := me.checkIs(MoPrimTypeList, expr); err != nil {
		return err
	}
	if of >= 0 {
		return me.check(of, -1, -1, expr.Val.(MoValList)...)
	}
	return nil
}

func (me *Interp) checkIsCallOnIdent(call *MoExpr, ident MoValIdent, errIfNumArgsNot int) (bool, *SrcFileNotice) {
	if call, is := call.Val.(MoValCall); is {
		if callee, _ := call[0].Val.(MoValIdent); callee == ident {
			return true, me.checkCount(errIfNumArgsNot, errIfNumArgsNot, MoExprs(call[1:]))
		}
	}
	return false, nil
}

func (me *Interp) checkNoneArePrimFuncs(have ...*MoExpr) bool {
	for _, arg := range have {
		if _, is := arg.Val.(MoValFnPrim); is {
			return false
		}
	}
	return true
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
