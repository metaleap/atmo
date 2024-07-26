package session

import (
	"fmt"
	"strings"

	"atmo/util"
)

var (
	moPrimOpsLazy = map[MoValIdent]moFnLazy{ // "lazy" prim-ops take unevaluated args to eval-or-not as needed. eg. `@match`, `@fn` etc
		// populated in `init()` below to avoid initialization-cycle error
	}
	moPrimOpsEager = map[MoValIdent]moFnEager{ // "eager" prim-ops receive already-evaluated args like any other func. eg. prim-type intrinsics like arithmetics, list concat etc
		"@replEnv":     (*Interp).primFnSessEnv,
		"@replPrintf":  (*Interp).primFnSessPrintf,
		"@replPrint":   (*Interp).primFnSessPrint,
		"@replPrintln": (*Interp).primFnSessPrintln,
		"@replReset":   (*Interp).primFnSessReset,
		"@numIntAdd":   makeArithPrimOp[MoValNumInt](MoPrimTypeNumInt, func(opl MoVal, opr MoVal) MoVal { return opl.(MoValNumInt) + opr.(MoValNumInt) }),
		"@numIntSub":   makeArithPrimOp[MoValNumInt](MoPrimTypeNumInt, func(opl MoVal, opr MoVal) MoVal { return opl.(MoValNumInt) - opr.(MoValNumInt) }),
		"@numIntMul":   makeArithPrimOp[MoValNumInt](MoPrimTypeNumInt, func(opl MoVal, opr MoVal) MoVal { return opl.(MoValNumInt) * opr.(MoValNumInt) }),
		"@numIntDiv":   makeArithPrimOp[MoValNumInt](MoPrimTypeNumInt, func(opl MoVal, opr MoVal) MoVal { return opl.(MoValNumInt) / opr.(MoValNumInt) }),
		"@numIntMod":   makeArithPrimOp[MoValNumInt](MoPrimTypeNumInt, func(opl MoVal, opr MoVal) MoVal { return opl.(MoValNumInt) % opr.(MoValNumInt) }),
		"@numUintAdd":  makeArithPrimOp[MoValNumUint](MoPrimTypeNumUint, func(opl MoVal, opr MoVal) MoVal { return opl.(MoValNumUint) + opr.(MoValNumUint) }),
		"@numUintSub":  makeArithPrimOp[MoValNumUint](MoPrimTypeNumUint, func(opl MoVal, opr MoVal) MoVal { return opl.(MoValNumUint) - opr.(MoValNumUint) }),
		"@numUintMul":  makeArithPrimOp[MoValNumUint](MoPrimTypeNumUint, func(opl MoVal, opr MoVal) MoVal { return opl.(MoValNumUint) * opr.(MoValNumUint) }),
		"@numUintDiv":  makeArithPrimOp[MoValNumUint](MoPrimTypeNumUint, func(opl MoVal, opr MoVal) MoVal { return opl.(MoValNumUint) / opr.(MoValNumUint) }),
		"@numUintMod":  makeArithPrimOp[MoValNumUint](MoPrimTypeNumUint, func(opl MoVal, opr MoVal) MoVal { return opl.(MoValNumUint) % opr.(MoValNumUint) }),
		"@numFloatAdd": makeArithPrimOp[MoValNumFloat](MoPrimTypeNumFloat, func(opl MoVal, opr MoVal) MoVal { return opl.(MoValNumFloat) + opr.(MoValNumFloat) }),
		"@numFloatSub": makeArithPrimOp[MoValNumFloat](MoPrimTypeNumFloat, func(opl MoVal, opr MoVal) MoVal { return opl.(MoValNumFloat) - opr.(MoValNumFloat) }),
		"@numFloatMul": makeArithPrimOp[MoValNumFloat](MoPrimTypeNumFloat, func(opl MoVal, opr MoVal) MoVal { return opl.(MoValNumFloat) * opr.(MoValNumFloat) }),
		"@numFloatDiv": makeArithPrimOp[MoValNumFloat](MoPrimTypeNumFloat, func(opl MoVal, opr MoVal) MoVal { return opl.(MoValNumFloat) / opr.(MoValNumFloat) }),
		"@eq":          (*Interp).primFnEq,
		"@neq":         (*Interp).primFnNeq,
		"@geq":         (*Interp).primFnGeq,
		"@leq":         (*Interp).primFnLeq,
		"@lt":          (*Interp).primFnLt,
		"@gt":          (*Interp).primFnGt,
		"@primTypeTag": (*Interp).primFnPrimTypeTag,
		"@listItemAt":  (*Interp).primFnListItemAt,
		"@listRange":   (*Interp).primFnListRange,
		"@listLen":     (*Interp).primFnListLen,
		"@listConcat":  (*Interp).primFnListConcat,
		"@dictHas":     (*Interp).primFnDictHas,
		"@dictGet":     (*Interp).primFnDictGet,
		"@dictWith":    (*Interp).primFnDictWith,
		"@dictWithout": (*Interp).primFnDictWithout,
		"@strConcat":   (*Interp).primFnStrConcat,
		"@strLen":      (*Interp).primFnStrLen,
		"@strCharAt":   (*Interp).primFnStrCharAt,
		"@strRange":    (*Interp).primFnStrRange,
		"@str":         (*Interp).primFnStr,
		"@exprStr":     (*Interp).primFnExprStr,
		"@exprParse":   (*Interp).primFnExprParse,
		"@exprEval":    (*Interp).primFnExprEval,
	}
)

const (
	moPrimOpQuote         MoValIdent = "#"
	moPrimOpQQuote        MoValIdent = "#$"
	moPrimOpUnquote       MoValIdent = "$"
	moPrimOpSpliceUnquote MoValIdent = "$$"
	moPrimOpDo            MoValIdent = "@do"
)

func init() {
	for k, v := range map[MoValIdent]moFnLazy{
		"@set":         (*Interp).primOpSet,
		"@fn":          (*Interp).primOpFn,
		"@caseOf":      (*Interp).primOpCaseOf,
		"@and":         (*Interp).primOpBoolAnd,
		"@or":          (*Interp).primOpBoolOr,
		"@macro":       (*Interp).primOpMacro,
		"@expand":      (*Interp).primOpMacroExpand,
		"@fnCall":      (*Interp).primOpFnCall,
		moPrimOpDo:     (*Interp).primOpDo,
		moPrimOpQuote:  (*Interp).primOpQuote,
		moPrimOpQQuote: (*Interp).primOpQuasiQuote,
	} {
		moPrimOpsLazy[k] = v
	}
}

// lazy prim-ops first, eager prim-ops afterwards

func (me *Interp) primOpFnCall(env *MoEnv, args ...*MoExpr) (*MoEnv, *MoExpr, *SrcFileNotice) {
	if err := me.checkCount(2, 2, args); err != nil {
		return nil, me.exprNever(err), err
	}
	if err := me.checkIs(MoPrimTypeList, args[1]); err != nil {
		return nil, me.exprNever(err), err
	}
	callee, args_list := args[0], args[1].Val.(MoValList)
	return env, me.expr(append(MoValCall{callee}, args_list...), nil, nil, args...), nil
}

func (me *Interp) primOpSet(env *MoEnv, args ...*MoExpr) (*MoEnv, *MoExpr, *SrcFileNotice) {
	if err := me.checkCount(2, 2, args); err != nil {
		return nil, me.exprNever(err), err
	}
	if err := me.checkIs(MoPrimTypeIdent, args[0]); err != nil {
		return nil, me.exprNever(err), err
	}
	name := args[0].Val.(MoValIdent)
	if is_reserved := ((name[0] == '@') || (name[0] == moPrimOpQuote[0]) || (name[0] == moPrimOpUnquote[0]) || moPrimOpsLazy[name] != nil); is_reserved {
		err := me.diagSpan(false, true, args[0]).newDiagErr(NoticeCodeReserved, name, string(rune(name[0])))
		return nil, me.exprNever(err), err
	}
	owner_env, found := env.lookupOwner(name)
	if owner_env == nil {
		owner_env = env
	}

	const can_set_macros = false
	if (!can_set_macros) && (found != nil) {
		if fn, _ := found.Val.(*MoValFnLam); (fn != nil) && fn.IsMacro {
			err := me.diagSpan(true, false, args...).newDiagErr(NoticeCodeAtmoTodo, "mutating macros currently disabled, let us know whether you disagree with that or not")
			return nil, me.exprNever(err), err
		}
	}
	new_value, err := me.evalAndApply(env, args[1])
	if err != nil {
		return nil, me.exprNever(err), err
	}
	owner_env.set(name, new_value)
	return nil, me.exprFrom(moValNone, args...), nil
}

func (me *Interp) primOpDo(env *MoEnv, args ...*MoExpr) (*MoEnv, *MoExpr, *SrcFileNotice) {
	if err := me.checkCount(1, 1, args); err != nil {
		return nil, me.exprNever(err), err
	}
	if err := me.checkIs(MoPrimTypeList, args[0]); err != nil {
		return nil, me.exprNever(err), err
	}
	list := args[0].Val.(MoValList)
	for _, item := range list[:len(list)-1] {
		if _, err := me.evalAndApply(env, item); err != nil {
			return nil, me.exprNever(err), err
		}
	}
	return env, me.exprFrom(list[len(list)-1], args...), nil
}

func (me *Interp) primOpMacro(env *MoEnv, args ...*MoExpr) (*MoEnv, *MoExpr, *SrcFileNotice) {
	_, expr, err := me.primOpFn(env, args...)
	if err != nil {
		return nil, me.exprNever(err), err
	}
	expr.Val.(*MoValFnLam).IsMacro = true
	return nil, expr, nil
}

func (me *Interp) primOpFn(env *MoEnv, args ...*MoExpr) (*MoEnv, *MoExpr, *SrcFileNotice) {
	if err := me.checkCount(2, 2, args); err != nil {
		return nil, me.exprNever(err), err
	}
	if err := me.checkIsListOf(MoPrimTypeIdent, args[0]); err != nil {
		return nil, me.exprNever(err), err
	}
	if err := me.checkIs(MoPrimTypeList, args[1]); err != nil {
		return nil, me.exprNever(err), err
	}
	list := args[1].Val.(MoValList)
	body := list[0]
	if len(list) > 1 {
		do := me.expr(moPrimOpDo, body.SrcFile, srcSpanFrom(MoExprs(list)))
		body = me.expr(MoValCall{do, me.expr(list, do.SrcFile, do.SrcSpan)}, do.SrcFile, do.SrcSpan)
	}
	expr := me.expr(
		&MoValFnLam{Params: MoExprs(args[0].Val.(MoValList)), Body: body, Env: env},
		body.SrcFile, srcSpanFrom(args))
	return nil, expr, nil
}

func (me *Interp) primOpQuote(_ *MoEnv, args ...*MoExpr) (*MoEnv, *MoExpr, *SrcFileNotice) {
	if err := me.checkCount(1, 1, args); err != nil {
		return nil, me.exprNever(err), err
	}
	return nil, me.exprFrom(args[0], args...), nil
}

func (me *Interp) primOpQuasiQuote(env *MoEnv, args ...*MoExpr) (*MoEnv, *MoExpr, *SrcFileNotice) {
	if err := me.checkCount(1, 1, args); err != nil {
		return nil, me.exprNever(err), err
	}

	// is atomic arg? then just like primOpQuote
	if args[0].Val.PrimType().isAtomic() {
		return nil, me.exprFrom(args[0], args...), nil
	}

	// is this call directly quoting an unquote call?
	if is_unquote, err := me.checkIsCallOnIdent(args[0], moPrimOpUnquote, 1); err != nil {
		return nil, me.exprNever(err), err
	} else if is_unquote {
		ret, err := me.evalAndApply(env, args[0].Val.(MoValCall)[1])
		if err != nil {
			return nil, me.exprNever(err), err
		}
		return nil, me.exprFrom(ret, args...), nil
	}

	// dict? call ourselves on each key and value
	if dict, is := args[0].Val.(MoValDict); is {
		ret := make(MoValDict, 0, len(dict))
		for _, pair := range dict {
			k, v := pair[0], pair[1]
			_, key, err := me.primOpQuasiQuote(env, k)
			if err != nil {
				return nil, me.exprNever(err), err
			}
			_, val, err := me.primOpQuasiQuote(env, v)
			if err != nil {
				return nil, me.exprNever(err), err
			}
			if dict.Has(key) {
				err := k.SrcSpan.newDiagErr(NoticeCodeDictDuplKey, key)
				return nil, me.exprNever(err), err
			}
			ret.Set(key, val)
		}
		return nil, me.expr(ret, me.srcFile(false, true, args...), me.diagSpan(false, true, args...)), nil
	}

	// must be list or call then: we handle them the same, per item iteration

	is_list := (args[0].Val.PrimType() == MoPrimTypeList)
	var call_or_arr MoExprs
	if call, is := args[0].Val.(MoValCall); is {
		call_or_arr = MoExprs(call)
	} else if is_list {
		call_or_arr = MoExprs(args[0].Val.(MoValList))
	} else {
		err := me.diagSpan(false, true, args[0]).newDiagErr(NoticeCodeAtmoTodo, "NEW BUG intro'd in primOpQuasiQuote")
		return nil, me.exprNever(err), err
	}

	ret := make(MoExprs, 0, len(call_or_arr))
	for _, item := range call_or_arr {
		if is_unquote, err := me.checkIsCallOnIdent(item, moPrimOpUnquote, 1); err != nil {
			return nil, me.exprNever(err), err
		} else if is_unquote {
			unquotee := item.Val.(MoValCall)[1]
			if evaled, err := me.evalAndApply(env, unquotee); err != nil {
				return nil, me.exprNever(err), err
			} else {
				ret = append(ret, evaled)
			}
		} else if is_splice_unquote, err := me.checkIsCallOnIdent(item, moPrimOpSpliceUnquote, 1); err != nil {
			return nil, me.exprNever(err), err
		} else if is_splice_unquote {
			unquotee := item.Val.(MoValCall)[1]
			evaled, err := me.evalAndApply(env, unquotee)
			if err != nil {
				return nil, me.exprNever(err), err
			}
			if err = me.checkIsListOf(-1, evaled); err != nil {
				return nil, me.exprNever(err), err
			}
			for _, splicee := range evaled.Val.(MoValList) {
				if evaled, err := me.evalAndApply(env, splicee); err != nil {
					return nil, me.exprNever(err), err
				} else {
					ret = append(ret, evaled)
				}
			}
		} else {
			_, evaled, err := me.primOpQuasiQuote(env, item)
			if err != nil {
				return nil, me.exprNever(err), err
			}
			ret = append(ret, evaled)
		}
	}
	return nil, me.expr(util.If[MoVal](is_list, MoValList(ret), MoValCall(ret)),
		me.srcFile(false, true, args...), me.diagSpan(false, true, args...)), nil
}

func (me *Interp) primOpMacroExpand(env *MoEnv, args ...*MoExpr) (*MoEnv, *MoExpr, *SrcFileNotice) {
	if err := me.checkCount(1, 1, args); err != nil {
		return nil, me.exprNever(err), err
	}
	if err := me.checkIs(MoPrimTypeCall, args[0]); err != nil {
		return nil, me.exprNever(err), err
	}
	ret, err := me.macroExpand(env, args[0])
	if err != nil {
		return nil, me.exprNever(err), err
	}
	return nil, me.exprFrom(ret, args...), nil
}

func (me *Interp) primOpCaseOf(env *MoEnv, args ...*MoExpr) (*MoEnv, *MoExpr, *SrcFileNotice) {
	if err := me.check(MoPrimTypeDict, 1, 1, args...); err != nil {
		return nil, me.exprNever(err), err
	}
	pairs := args[0].Val.(MoValDict)
	for _, pair := range pairs {
		key, val := pair[0], pair[1]
		pred, err := me.evalAndApply(env, key)
		if err != nil {
			return nil, me.exprNever(err, key), err
		}
		if pred.EqTrue() {
			return env, me.exprFrom(val, val), nil
		} else if !pred.EqFalse() {
			err := me.diagSpan(false, true, key).newDiagErr(NoticeCodeExpectedFoo, "a boolean expression")
			return nil, me.exprNever(err), err
		}
	}
	err := me.diagSpan(true, false, args...).newDiagErr(NoticeCodeNoElseCase)
	return nil, me.exprNever(err), err
}

func (me *Interp) primOpBoolAnd(env *MoEnv, args ...*MoExpr) (*MoEnv, *MoExpr, *SrcFileNotice) {
	if err := me.checkCount(2, 2, args); err != nil {
		return nil, me.exprNever(err), err
	}
	for _, arg := range args {
		evaled, err := me.evalAndApply(env, arg)
		if err != nil {
			return nil, me.exprNever(err), err
		}
		if evaled.EqFalse() {
			return nil, me.exprBool(false, args...), nil
		} else if !evaled.EqTrue() {
			err := me.diagSpan(false, true, arg).newDiagErr(NoticeCodeExpectedFoo, "a boolean expression")
			return nil, me.exprNever(err), err
		}
	}
	return nil, me.exprBool(true, args...), nil
}

func (me *Interp) primOpBoolOr(env *MoEnv, args ...*MoExpr) (*MoEnv, *MoExpr, *SrcFileNotice) {
	if err := me.checkCount(2, 2, args); err != nil {
		return nil, me.exprNever(err), err
	}
	for _, arg := range args {
		evaled, err := me.evalAndApply(env, arg)
		if err != nil {
			return nil, me.exprNever(err), err
		}
		if evaled.EqTrue() {
			return nil, me.exprBool(true, args...), nil
		} else if !evaled.EqFalse() {
			err := me.diagSpan(false, true, arg).newDiagErr(NoticeCodeExpectedFoo, "a boolean expression")
			return nil, me.exprNever(err), err
		}
	}
	return nil, me.exprBool(false, args...), nil
}

// eager prim-ops below, lazy ones above

func makeArithPrimOp[T MoValNumInt | MoValNumUint | MoValNumFloat](t MoValPrimType, f func(opl MoVal, opr MoVal) MoVal) moFnEager {
	return func(me *Interp, _ *MoEnv, args ...*MoExpr) (*MoExpr, *SrcFileNotice) {
		if err := me.check(t, 2, 2, args...); err != nil {
			return me.exprNever(err), err
		}
		return me.expr(f(args[0].Val, args[1].Val), nil, nil, args...), nil
	}
}

func (me *Interp) primFnSessEnv(env *MoEnv, args ...*MoExpr) (*MoExpr, *SrcFileNotice) {
	if err := me.checkCount(0, 0, args); err != nil {
		return me.exprNever(err), err
	}
	src_file := me.ReplFauxFile
	if (me.diagCtxCall != nil) && (me.diagCtxCall.SrcFile != nil) {
		src_file = me.diagCtxCall.SrcFile
	}
	src_span := me.diagSpan(false, false, args...)

	ret := me.expr(make(MoValDict, 0, len(me.Env.Own)+len(env.Own)), src_file, src_span)
	var populate func(it *MoEnv, into *MoExpr) *MoExpr
	populate = func(it *MoEnv, into *MoExpr) *MoExpr {
		dict := into.Val.(MoValDict)
		for k, v := range it.Own {
			dict.Set(me.expr(k, src_file, src_span), v)
		}
		if it.Parent != nil {
			dict_parent := me.expr(make(MoValDict, 0, len(env.Parent.Own)), src_file, src_span)
			dict.Set(me.expr(MoValIdent(""), src_file, src_span),
				populate(it.Parent, dict_parent))
		}
		into.Val = dict
		return into
	}
	populate(env, ret)
	return ret, nil
}

func (me *Interp) primFnSessPrintf(_ *MoEnv, args ...*MoExpr) (*MoExpr, *SrcFileNotice) {
	if err := me.checkCount(2, 2, args); err != nil {
		return me.exprNever(err), err
	}
	if err := me.checkIs(MoPrimTypeStr, args[0]); err != nil {
		return me.exprNever(err), err
	}
	if err := me.checkIs(MoPrimTypeList, args[1]); err != nil {
		return me.exprNever(err), err
	}
	list := args[1].Val.(MoValList)
	fmt_args := make([]any, 0, len(list))
	for _, arg := range list {
		fmt_args = append(fmt_args, arg.Val)
	}
	fmt.Fprintf(me.StdIo.Out, string(args[0].Val.(MoValStr)), fmt_args...)
	return me.exprFrom(moValNone, args...), nil
}

func (me *Interp) primFnSessPrint(_ *MoEnv, args ...*MoExpr) (*MoExpr, *SrcFileNotice) {
	if err := me.checkCount(1, 1, args); err != nil {
		return me.exprNever(err), err
	}
	switch arg0 := args[0].Val.(type) {
	case MoValStr:
		me.StdIo.Out.WriteString(string(arg0))
	default:
		args[0].WriteTo(me.StdIo.Out)
	}
	return me.exprFrom(moValNone, args...), nil
}

func (me *Interp) primFnSessPrintln(env *MoEnv, args ...*MoExpr) (*MoExpr, *SrcFileNotice) {
	expr, err := me.primFnSessPrint(env, args...)
	if (expr.Diag.Err == nil) && (err == nil) {
		me.StdIo.Out.WriteString("\n")
	}
	return expr, err
}

func (me *Interp) primFnListLen(_ *MoEnv, args ...*MoExpr) (*MoExpr, *SrcFileNotice) {
	if err := me.checkCount(1, 1, args); err != nil {
		return me.exprNever(err), err
	}
	if err := me.checkIs(MoPrimTypeList, args[0]); err != nil {
		return me.exprNever(err), err
	}
	return me.expr(MoValNumUint(len(args[0].Val.(MoValList))), nil, nil, args...), nil
}

func (me *Interp) primFnListItemAt(_ *MoEnv, args ...*MoExpr) (*MoExpr, *SrcFileNotice) {
	if err := me.checkCount(2, 2, args); err != nil {
		return me.exprNever(err), err
	}
	if err := me.checkIs(MoPrimTypeList, args[0]); err != nil {
		return me.exprNever(err), err
	}
	if err := me.checkIs(MoPrimTypeNumUint, args[1]); err != nil {
		return me.exprNever(err), err
	}
	list, idx := args[0].Val.(MoValList), args[1].Val.(MoValNumUint)
	if idx_downcast := int(idx); (idx_downcast < 0) || (idx_downcast >= len(list)) {
		err := me.diagSpan(false, true, args[1]).newDiagErr(NoticeCodeIndexOutOfBounds, idx_downcast, len(list))
		return me.exprNever(err), err
	}
	return me.exprFrom(list[idx], args...), nil
}

func (me *Interp) primFnListRange(_ *MoEnv, args ...*MoExpr) (*MoExpr, *SrcFileNotice) {
	if err := me.checkCount(2, 3, args); err != nil {
		return me.exprNever(err), err
	}
	if err := me.checkIs(MoPrimTypeList, args[0]); err != nil {
		return me.exprNever(err), err
	}
	if err := me.checkIs(MoPrimTypeNumUint, args[1]); err != nil {
		return me.exprNever(err), err
	}
	list, idx_start := args[0].Val.(MoValList), args[1].Val.(MoValNumUint)
	if len(args) == 2 {
		args = append(args, me.expr(MoValNumUint(len(list)), me.srcFile(false, true, args[1]), me.diagSpan(false, true, args[1])))
	} else if err := me.checkIs(MoPrimTypeNumUint, args[2]); err != nil {
		return me.exprNever(err), err
	}
	idx_end := args[2].Val.(MoValNumUint)
	if idx_end < idx_start {
		err := me.diagSpan(false, true, args[2]).newDiagErr(NoticeCodeRangeNegative, idx_end, idx_start)
		return me.exprNever(err), err
	} else if idx_downcast := int(idx_start); (idx_downcast < 0) || (idx_downcast > len(list)) {
		err := me.diagSpan(false, true, args[1]).newDiagErr(NoticeCodeIndexOutOfBounds, idx_downcast, len(list))
		return me.exprNever(err), err
	}
	if idx_downcast := int(idx_end); (idx_downcast < 0) || (idx_downcast > len(list)) {
		err := me.diagSpan(false, true, args[2]).newDiagErr(NoticeCodeIndexOutOfBounds, idx_downcast, len(list))
		return me.exprNever(err), err
	}
	return me.expr(list[idx_start:idx_end], me.srcFile(false, idx_start != idx_end, list[idx_start:idx_end]...), me.diagSpan(false, idx_start != idx_end, list[idx_start:idx_end]...)), nil
}

func (me *Interp) primFnListConcat(_ *MoEnv, args ...*MoExpr) (*MoExpr, *SrcFileNotice) {
	ret := make(MoValList, 0, len(args)*3)
	for _, arg := range args {
		if err := me.checkIs(MoPrimTypeList, arg); err != nil {
			return me.exprNever(err), err
		}
		ret = append(ret, arg.Val.(MoValList)...)
	}
	return me.expr(ret, nil, nil, args...), nil
}

func (me *Interp) primFnStrCharAt(_ *MoEnv, args ...*MoExpr) (*MoExpr, *SrcFileNotice) {
	if err := me.checkCount(2, 2, args); err != nil {
		return me.exprNever(err), err
	}
	if err := me.checkIs(MoPrimTypeStr, args[0]); err != nil {
		return me.exprNever(err), err
	}
	str := args[0].Val.(MoValStr)
	if err := me.checkIs(MoPrimTypeNumUint, args[1]); err != nil {
		return me.exprNever(err), err
	}
	idx := int(args[1].Val.(MoValNumUint))
	if (idx < 0) || (idx >= len(str)) {
		err := me.diagSpan(false, true, args[1]).newDiagErr(NoticeCodeIndexOutOfBounds, idx, len(str))
		return me.exprNever(err), err
	}
	return me.expr(MoValChar(str[idx]), nil, nil, args...), nil
}

func (me *Interp) primFnStrRange(_ *MoEnv, args ...*MoExpr) (*MoExpr, *SrcFileNotice) {
	if err := me.checkCount(2, 3, args); err != nil {
		return me.exprNever(err), err
	}
	if err := me.checkIs(MoPrimTypeStr, args[0]); err != nil {
		return me.exprNever(err), err
	}
	str := args[0].Val.(MoValStr)
	if err := me.checkIs(MoPrimTypeNumUint, args[1]); err != nil {
		return me.exprNever(err), err
	}
	idx_start, idx_end := int(args[1].Val.(MoValNumUint)), len(str)
	if (idx_start < 0) || (idx_start > len(str)) {
		err := me.diagSpan(false, true, args[1]).newDiagErr(NoticeCodeIndexOutOfBounds, idx_start, len(str))
		return me.exprNever(err), err
	}
	if len(args) > 2 {
		if err := me.checkIs(MoPrimTypeNumUint, args[2]); err != nil {
			return me.exprNever(err), err
		}
		idx_end = int(args[2].Val.(MoValNumUint))
	}
	if (idx_end < 0) || (idx_end > len(str)) {
		err := me.diagSpan(false, true, args[1]).newDiagErr(NoticeCodeIndexOutOfBounds, idx_end, len(str))
		return me.exprNever(err), err
	} else if idx_end < idx_start {
		err := me.diagSpan(false, true, args[1]).newDiagErr(NoticeCodeRangeNegative, idx_end, idx_start)
		return me.exprNever(err), err
	}
	return me.expr(MoValStr(str[idx_start:idx_end]), nil, nil, args...), nil
}

func (me *Interp) primFnStrLen(_ *MoEnv, args ...*MoExpr) (*MoExpr, *SrcFileNotice) {
	if err := me.checkCount(1, 1, args); err != nil {
		return me.exprNever(err), err
	}
	if err := me.checkIs(MoPrimTypeStr, args[0]); err != nil {
		return me.exprNever(err), err
	}
	return me.expr(MoValNumUint(len(args[0].Val.(MoValStr))), nil, nil, args...), nil
}

func (me *Interp) primFnStrConcat(_ *MoEnv, args ...*MoExpr) (*MoExpr, *SrcFileNotice) {
	if err := me.checkCount(1, 1, args); err != nil {
		return me.exprNever(err), err
	}
	if err := me.checkIs(MoPrimTypeList, args[0]); err != nil {
		return me.exprNever(err), err
	}
	var buf strings.Builder
	for _, arg := range args[0].Val.(MoValList) {
		if err := me.checkIs(MoPrimTypeStr, arg); err != nil {
			return me.exprNever(err), err
		}
		buf.WriteString(string(arg.Val.(MoValStr)))
	}
	return me.expr(MoValStr(buf.String()), nil, nil, args...), nil
}

func (me *Interp) primFnStr(_ *MoEnv, args ...*MoExpr) (*MoExpr, *SrcFileNotice) {
	if err := me.checkCount(1, 1, args); err != nil {
		return me.exprNever(err), err
	}
	var str MoValStr
	switch it := args[0].Val.(type) {
	case MoValChar:
		str = MoValStr(string(it))
	case MoValStr:
		str = MoValStr(it)
	case MoValType:
		str = MoValStr(((MoValPrimType)(it)).Str(true))
	default:
		str = MoValStr(MoValToString(it))
	}
	return me.expr(str, nil, nil, args...), nil
}

func (me *Interp) primFnExprStr(_ *MoEnv, args ...*MoExpr) (*MoExpr, *SrcFileNotice) {
	if err := me.checkCount(1, 1, args); err != nil {
		return me.exprNever(err), err
	}
	return me.expr(MoValStr(args[0].String()), nil, nil, args...), nil
}

func (me *Interp) primFnExprEval(_ *MoEnv, args ...*MoExpr) (*MoExpr, *SrcFileNotice) {
	if err := me.checkCount(1, 1, args); err != nil {
		return me.exprNever(err), err
	}
	ret, err := me.evalAndApply(me.Env, args[0])
	if err != nil {
		return me.exprNever(err), err
	}
	return me.exprFrom(ret, args...), nil
}

func (me *Interp) primFnExprParse(_ *MoEnv, args ...*MoExpr) (*MoExpr, *SrcFileNotice) {
	if err := me.checkCount(1, 1, args); err != nil {
		return me.exprNever(err), err
	}
	if err := me.checkIs(MoPrimTypeStr, args[0]); err != nil {
		return me.exprNever(err), err
	}
	return me.Parse(string(args[0].Val.(MoValStr)))
}

func (me *Interp) primFnDictHas(_ *MoEnv, args ...*MoExpr) (*MoExpr, *SrcFileNotice) {
	if err := me.checkCount(2, 2, args); err != nil {
		return me.exprNever(err), err
	}
	if err := me.checkIs(MoPrimTypeDict, args[0]); err != nil {
		return me.exprNever(err), err
	}
	return me.exprBool(args[0].Val.(MoValDict).Has(args[1]), args...), nil
}

func (me *Interp) primFnDictGet(_ *MoEnv, args ...*MoExpr) (*MoExpr, *SrcFileNotice) {
	if err := me.checkCount(2, 2, args); err != nil {
		return me.exprNever(err), err
	}
	if err := me.checkIs(MoPrimTypeDict, args[0]); err != nil {
		return me.exprNever(err), err
	}
	found := args[0].Val.(MoValDict).Get(args[1])
	if found == nil {
		return me.exprFrom(moValNone, args...), nil
	}
	return me.exprFrom(found, args...), nil
}

func (me *Interp) primFnDictWith(_ *MoEnv, args ...*MoExpr) (*MoExpr, *SrcFileNotice) {
	if err := me.checkCount(3, 3, args); err != nil {
		return me.exprNever(err), err
	}
	if err := me.checkIs(MoPrimTypeDict, args[0]); err != nil {
		return me.exprNever(err), err
	}
	ret := me.exprFrom(args[0], args...)
	ret.Val = ret.Val.(MoValDict).With(args[1], args[2])
	return ret, nil
}

func (me *Interp) primFnDictWithout(_ *MoEnv, args ...*MoExpr) (*MoExpr, *SrcFileNotice) {
	if err := me.checkCount(2, 2, args); err != nil {
		return me.exprNever(err), err
	}
	if err := me.checkIs(MoPrimTypeDict, args[0]); err != nil {
		return me.exprNever(err), err
	}
	if err := me.checkIs(MoPrimTypeList, args[1]); err != nil {
		return me.exprNever(err), err
	}
	ret := me.exprFrom(args[0], args...)
	ret.Val = ret.Val.(MoValDict).Without(args[1].Val.(MoValList)...)
	return ret, nil
}

func (me *Interp) primFnPrimTypeTag(_ *MoEnv, args ...*MoExpr) (*MoExpr, *SrcFileNotice) {
	if err := me.checkCount(1, 1, args); err != nil {
		return me.exprNever(err), err
	}
	return me.expr(MoValType(args[0].Val.PrimType()), nil, nil, args...), nil
}

func (me *Interp) primFnSessReset(_ *MoEnv, args ...*MoExpr) (*MoExpr, *SrcFileNotice) {
	me.reset()
	return moValNone, nil
}

func (me *Interp) primFnEq(_ *MoEnv, args ...*MoExpr) (*MoExpr, *SrcFileNotice) {
	if err := me.checkCount(2, 2, args); err != nil {
		return me.exprNever(err), err
	}
	if !me.checkNoneArePrimFuncs(args...) {
		err := me.diagSpan(false, true, args...).newDiagErr(NoticeCodeNotComparable, args[0], args[1], "equality")
		return me.exprNever(err), err
	}
	return me.exprBool(args[0].Eq(args[1]), args...), nil
}

func (me *Interp) primFnNeq(env *MoEnv, args ...*MoExpr) (*MoExpr, *SrcFileNotice) {
	if err := me.checkCount(2, 2, args); err != nil {
		return me.exprNever(err), err
	}
	if !me.checkNoneArePrimFuncs(args...) {
		err := me.diagSpan(false, true, args...).newDiagErr(NoticeCodeNotComparable, args[0], args[1], "equality")
		return me.exprNever(err), err
	}
	return me.exprBool(!args[0].Eq(args[1]), args...), nil
}

func (me *Interp) primFnLeq(_ *MoEnv, args ...*MoExpr) (*MoExpr, *SrcFileNotice) {
	cmp, err := me.primCmpHelper("less-or-equal", args...)
	if err != nil {
		return me.exprNever(err), err
	}
	return me.exprBool(cmp <= 0, args...), nil
}

func (me *Interp) primFnGeq(_ *MoEnv, args ...*MoExpr) (*MoExpr, *SrcFileNotice) {
	cmp, err := me.primCmpHelper("greater-or-equal", args...)
	if err != nil {
		return me.exprNever(err), err
	}
	return me.exprBool(cmp >= 0, args...), nil
}

func (me *Interp) primFnLt(_ *MoEnv, args ...*MoExpr) (*MoExpr, *SrcFileNotice) {
	cmp, err := me.primCmpHelper("less-than", args...)
	if err != nil {
		return me.exprNever(err), err
	}
	return me.exprBool(cmp < 0, args...), nil
}

func (me *Interp) primFnGt(_ *MoEnv, args ...*MoExpr) (*MoExpr, *SrcFileNotice) {
	cmp, err := me.primCmpHelper("greater-than", args...)
	if err != nil {
		return me.exprNever(err), err
	}
	return me.exprBool(cmp > 0, args...), nil
}

func (me *Interp) primCmpHelper(diagMoniker string, args ...*MoExpr) (int, *SrcFileNotice) {
	if err := me.checkCount(2, 2, args); err != nil {
		return 0, err
	}
	if !me.checkNoneArePrimFuncs(args...) {
		return 0, me.diagSpan(false, true, args...).newDiagErr(NoticeCodeNotComparable, args[0], args[1], diagMoniker)
	}
	cmp, err := me.Cmp(args[0], args[1], diagMoniker)
	if err != nil {
		return 0, err
	}
	return cmp, nil
}
