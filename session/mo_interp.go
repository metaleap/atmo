package session

import (
	"io"
	"os"

	"atmo/util"
	"atmo/util/kv"
	"atmo/util/sl"
	"atmo/util/str"
)

var (
	InterpStderr    Writer    = os.Stderr
	InterpStdout    Writer    = os.Stdout
	InterpStdin     io.Reader = os.Stdin
	interpRootEnv             = MoEnv{Own: map[MoValIdent]*MoExpr{}}
	InterpCallLimit int
	InterpLoopLimit int
)

type Writer interface {
	io.Writer
	io.StringWriter
}

type Interp struct {
	Pack           *SrcPack
	FauxFile       *SrcFile
	Env            *MoEnv
	SubCallListing struct {
		Use  bool
		Last MoExprs
	}
	callCounter int
	loopCounter int
	diagCtxCall *MoExpr // set to a full call-expr just before it is entered into, for use in producing that call's error (if any) unwinding the whole eval
}

func newInterp(inPack *SrcPack, replFauxFile *SrcFile) *Interp {
	if replFauxFile == nil {
		src_file_path := newInterpFauxFilePath(inPack.DirPath)
		_ = ensureSrcFiles(nil, true, src_file_path)
		replFauxFile = state.srcFiles[src_file_path]
	}

	me := Interp{Pack: inPack, FauxFile: replFauxFile}
	me.resetEnv()
	me.Pack.Interp = &me
	return &me
}

func (me *Interp) resetEnv() {
	me.Env = newMoEnv(&interpRootEnv, nil, nil)
	me.ensureRootEnvPopulated()
}

func (me *Interp) resetForSem() {
	me.ClearStackTrace()
	me.resetEnv()
}

// for REPL use-cases only!
func (me *Interp) resetForRepl() {
	Access(func(sess StateAccess, _ Intel) {
		allDiags = map[string]sl.Of[*Diag]{}
		me.ClearStackTrace()
		me.resetEnv()
		me.Pack.Interp, me.Pack.Trees.MoOrig, me.Pack.Trees.MoEvaled = me, nil, nil
		for _, src_file := range me.Pack.Files {
			src_file.diags.LexErrs, src_file.diags.LastReadErr, src_file.diags.Ast2Mo, src_file.Src.Ast, src_file.Src.Toks, src_file.Src.Text =
				nil, nil, nil, nil, nil, ""
		}
		_ = ensureSrcFiles(nil, false, me.Pack.srcFilePaths()...) // does `SrcPack.treesRefresh()` too
		refreshAndPublishDiags(true, me.Pack.srcFilePaths()...)
	})
}

func (me *Interp) ClearStackTrace() {
	me.SubCallListing.Last = me.SubCallListing.Last[:0] // keeps currently-already-alloc'd capacity, for reduced GC churn and reduced alloc times
}

func (me *Interp) ExprEval(expr *MoExpr) *MoExpr {
	me.ClearStackTrace()
	me.diagCtxCall = nil
	if expr == nil {
		return nil
	}
	me.callCounter = 0
	return me.evalAndApply(me.Env, expr)
}

func (me *Interp) evalAndApply(env *MoEnv, expr *MoExpr) *MoExpr {
	if me.callCounter++; (InterpCallLimit > 0) && (me.callCounter > InterpCallLimit) {
		return me.exprErr(expr.SrcNode.newDiagErr(false, ErrCodeInterpLimitExceeded, InterpLoopLimit, InterpCallLimit), expr)
	}

	diag_ctx_orig, expr_orig, did_call := me.diagCtxCall, expr, false
	// id := strconv.FormatInt(time.Now().UnixNano(), 36) // uncomment and print `id` to check for TCO loop
	me.loopCounter = 0
tco_loop:
	for env != nil {
		if me.loopCounter++; (InterpLoopLimit > 0) && (me.loopCounter > InterpLoopLimit) {
			return me.exprErr(expr.SrcNode.newDiagErr(false, ErrCodeInterpLimitExceeded, InterpLoopLimit, InterpCallLimit), expr)
		}
		if _, is_call := expr.Val.(MoValCall); !is_call {
			env, expr = nil, me.evalExpr(env, expr)
			break tco_loop
		} else if expr = me.macroExpand(env, expr); !expr.IsErr() {
			if call, is_call := expr.Val.(MoValCall); is_call { // checking once more now after macro-expansion
				callee, call_args := call[0], (MoExprs)(call[1:])

				var prim_op_lazy moFnLazy
				if ident, _ := callee.Val.(MoValIdent); ident != "" {
					prim_op_lazy = moPrimOpsLazy[ident]
				}

				if prim_op_lazy != nil {
					diag_ctx_cur, diag_ctx_prev := expr, me.diagCtxCall
					me.diagCtxCall = diag_ctx_cur
					env, expr = prim_op_lazy(me, env, call_args...)
					expr.setSrcSpanIfNone(diag_ctx_cur)
					me.diagCtxCall = diag_ctx_prev
				} else {
					if me.SubCallListing.Use {
						me.SubCallListing.Last = append(me.SubCallListing.Last, expr)
					}
					diag_ctx_cur, diag_ctx_prev := expr, me.diagCtxCall
					me.diagCtxCall = diag_ctx_cur
					expr = me.evalExpr(env, expr)
					if expr.IsErr() {
						break tco_loop
					}
					me.diagCtxCall = diag_ctx_cur
					evaled_call := expr.Val.(MoValCall)
					evaled_callee, evaled_call_args := evaled_call[0], (MoExprs)(evaled_call[1:])
					if evaled_callee.IsErr() {
						return me.exprErr(evaled_callee.Diag.Err, callee)
					}
					for i, arg := range evaled_call_args {
						if arg.IsErr() {
							return me.exprErr(arg.Diag.Err, call_args[i])
						}
					}
					did_call = true
					switch fn := evaled_callee.Val.(type) {
					default:
						env, expr = nil, me.exprErr(me.diagSpan(true, false).newDiagErr(ErrCodeNotCallable, evaled_callee.String()))
					case MoValFnPrim:
						expr = fn(me, env, evaled_call_args...)
						expr.setSrcSpanIfNone(diag_ctx_cur)
						env, me.diagCtxCall = nil, diag_ctx_prev
					case *MoValFnLam:
						var err *Diag
						env, err = me.envWith(fn, evaled_call_args)
						if err != nil {
							env, expr = nil, me.exprErr(err, diag_ctx_cur)
						} else {
							expr = fn.Body
						}
					}
				}
			}
		}
	}
	me.diagCtxCall = diag_ctx_orig
	if false && did_call {
		if err := expr.Err(); (err != nil) && (expr_orig.SrcSpan != nil) {
			err.Span = *expr_orig.SrcSpan
		}
	}
	if (expr != nil) && ((expr.SrcFile == nil) || (expr.SrcSpan == nil)) {
		diag := expr.Diag
		expr = &MoExpr{Val: expr.Val, SrcSpan: sl.FirstNonNil(expr.SrcSpan, expr_orig.SrcSpan), SrcFile: sl.FirstNonNil(expr.SrcFile, expr_orig.SrcFile)}
		expr.Diag = diag
	}
	me.callCounter--
	return expr
}

func (me *Interp) evalExpr(env *MoEnv, expr *MoExpr) *MoExpr {
	switch val := expr.Val.(type) {
	case MoValIdent:
		owner_env, found := env.lookupOwner(val)
		if (found != nil) && found.PreEvalTopLevelPreEnvUnevaledYet { // top-level @set was put unevaled into env under the name it has yet to set
			_ = me.evalAndApply(owner_env, found) // eval that @set call, to put the actual value for the name in env instead, return is ignored because it's just @void or an @Err
			found = env.lookup(val)
		}
		if found == nil {
			_, is_lazy_prim_op := moPrimOpsLazy[val]
			return me.exprErr(me.diagSpan(false, true, expr).newDiagErr(util.If(!is_lazy_prim_op, ErrCodeNotDefined, ErrCodeNotAValue), val))
		}
		return me.expr(found.Val, expr.SrcFile, expr.SrcSpan)
	case *MoValList:
		list := make(MoValList, len(*val))
		for i, item := range *val {
			list[i] = me.evalAndApply(env, item)
		}
		return me.expr(&list, expr.SrcFile, expr.SrcSpan)
	case *MoValTup:
		tup := make(MoValTup, len(*val))
		for i, item := range *val {
			tup[i] = me.evalAndApply(env, item)
		}
		return me.expr(&tup, expr.SrcFile, expr.SrcSpan)
	case *MoValObj:
		obj := make(MoValDict, 0, len(*val))
		for _, entry := range *val {
			key := me.evalAndApply(env, entry.Key)
			val := me.evalAndApply(env, entry.Val)
			if (!key.IsErr()) && obj.Has(key) {
				return me.exprErr(entry.Key.SrcSpan.newDiagErr(ErrCodeDictDuplKey, key))
			}
			obj.Set(key, val)
		}
		return me.expr((*MoValObj)(&obj), expr.SrcFile, expr.SrcSpan)
	case *MoValDict:
		dict := make(MoValDict, 0, len(*val))
		for _, entry := range *val {
			key := me.evalAndApply(env, entry.Key)
			val := me.evalAndApply(env, entry.Val)
			if (!key.IsErr()) && dict.Has(key) {
				return me.exprErr(entry.Key.SrcSpan.newDiagErr(ErrCodeDictDuplKey, key))
			}
			dict.Set(key, val)
		}
		return me.expr(&dict, expr.SrcFile, expr.SrcSpan)
	case MoValCall:
		call := make(MoValCall, len(val))
		for i, item := range val {
			call[i] = me.evalAndApply(env, item)
		}
		return me.expr(call, expr.SrcFile, expr.SrcSpan)
	}
	return expr
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

func (me *Interp) envWith(fn *MoValFnLam, args MoExprs) (*MoEnv, *Diag) {
	if err := me.checkCount(len(fn.Params), len(fn.Params), args); err != nil {
		return nil, err
	}
	if err := me.checkArgErrs(args...); err != nil {
		return nil, err.Diag.Err
	}
	return newMoEnv(fn.Env, fn.Params, args), nil
}

func (me *Interp) macroExpand(env *MoEnv, expr *MoExpr) *MoExpr {
	diag_ctx := me.diagCtxCall
	for fn := expr.macroCallCallee(env); fn != nil; fn = expr.macroCallCallee(env) {
		me.diagCtxCall = expr

		fn := fn.Val.(*MoValFnLam)
		call_env, err := me.envWith(fn, MoExprs(expr.Val.(MoValCall)[1:]))
		if err != nil {
			return me.exprErr(err)
		}
		expr_now := me.evalAndApply(call_env, fn.Body)
		expr_now.setSrcSpanIfNone(expr)
		expr = expr_now
	}
	me.diagCtxCall = diag_ctx
	return expr
}

func (me *MoExpr) callee() *MoExpr {
	if me == nil {
		return nil
	}
	if call, is := me.Val.(MoValCall); is {
		return call[0]
	}
	return nil
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

func (me *Interp) checkArgErrs(args ...*MoExpr) *MoExpr {
	for _, arg := range args {
		if err := arg.Err(); err != nil {
			return me.exprErr(err, arg)
		}
	}
	return nil
}

func (me *Interp) checkCount(wantAtLeast int, wantAtMost int, have MoExprs) *Diag {
	return me.checkCountWithSrcSpan(wantAtLeast, wantAtMost, have, false)
}

func (me *Interp) checkCountWithSrcSpan(wantAtLeast int, wantAtMost int, have MoExprs, preferSrcSpan bool) *Diag {
	if wantAtLeast < 0 {
		return nil
	}
	diag_src_span := me.diagSpan(false, preferSrcSpan, have...)
	plural := util.If((wantAtLeast <= wantAtMost) && (wantAtLeast != 1), "s", "")
	moniker := util.If(preferSrcSpan, "item"+plural, "arg"+plural+" for this call")
	if (wantAtLeast == wantAtMost) && (wantAtLeast != len(have)) {
		return diag_src_span.newDiagErr(ErrCodeExpectedFoo, str.Fmt("%d %s, not %d", wantAtLeast, moniker, len(have)))
	} else if len(have) < wantAtLeast {
		return diag_src_span.newDiagErr(ErrCodeExpectedFoo, str.Fmt("at least %d %s, not %d", wantAtLeast, moniker, len(have)))
	} else if (wantAtMost > wantAtLeast) && (len(have) > wantAtMost) {
		return diag_src_span.newDiagErr(ErrCodeExpectedFoo, str.Fmt("%d to %d %s, not %d", wantAtLeast, wantAtMost, moniker, len(have)))
	}
	return nil
}

func (me *Interp) checkIs(want MoValPrimType, have *MoExpr) *Diag {
	if have_type := have.Val.PrimType(); have_type != want {
		return me.diagSpan(false, true, have).newDiagErr(ErrCodeTypeMismatch, want.Str(true), have_type.Str(true))
	}
	return nil
}

func (me *Interp) checkIsListOf(of MoValPrimType, expr *MoExpr) *Diag {
	if err := me.checkIs(MoPrimTypeList, expr); err != nil {
		return err
	}
	if of >= 0 {
		return me.check(of, -1, -1, *(expr.Val.(*MoValList))...)
	}
	return nil
}

func (me *Interp) checkIsCallOnIdent(call *MoExpr, ident MoValIdent, errIfNumArgsNot int) (bool, *Diag) {
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

func (me *Interp) check(want MoValPrimType, wantAtLeast int, wantAtMost int, have ...*MoExpr) *Diag {
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

func (me *Interp) newErrExpectedBool(noBool *MoExpr) *Diag {
	return me.diagSpan(false, true, noBool).newDiagErr(ErrCodeTypeMismatch, MoPrimTypeBool.Str(true), noBool.Val.PrimType().Str(true))
}

type MoEnv struct {
	Parent *MoEnv
	Own    map[MoValIdent]*MoExpr
}

func (me *Interp) ensureRootEnvPopulated() {
	if len(interpRootEnv.Own) > 0 {
		return
	}
	// builtin eager prim-op funcs into rootEnv
	for name, fn := range moPrimOpsEager {
		interpRootEnv.set(name, me.expr(MoValFnPrim(fn), nil, nil))
	}
}

func newMoEnv(parent *MoEnv, names MoExprs, values MoExprs) *MoEnv {
	util.Assert(len(names) == len(values), "newMoEnv: len(names) != len(values)")
	ret := MoEnv{Parent: parent, Own: map[MoValIdent]*MoExpr{}}
	for i, name := range names {
		ret.Own[name.Val.(MoValIdent)] = values[i]
	}
	return &ret
}

func (me *MoEnv) eq(to *MoEnv) bool {
	if me == to {
		return true
	}
	if (me == nil) || (to == nil) {
		return false
	}
	return me.Parent.eq(to.Parent) && kv.Eq(me.Own, to.Own, (*MoExpr).Eq)
}

func (me *MoEnv) hasOwn(name MoValIdent) (ret bool) {
	_, ret = me.Own[name]
	return
}

func (me *MoEnv) set(name MoValIdent, value *MoExpr) {
	util.Assert(value != nil, "MoEnv.set(name, nil)")
	me.Own[name] = value
}

func (me *MoEnv) lookup(name MoValIdent) *MoExpr {
	_, found := me.lookupOwner(name)
	return found
}

func (me *MoEnv) lookupOwner(name MoValIdent) (*MoEnv, *MoExpr) {
	found := me.Own[name]
	if found == nil {
		if me.Parent != nil {
			return me.Parent.lookupOwner(name)
		} else {
			return nil, nil
		}
	}
	return me, found
}
