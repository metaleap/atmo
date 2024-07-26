package session

import (
	"fmt"
	"strings"

	"atmo/util"
)

var (
	moPrimOpsLazy  map[MoValIdent]moFnLazy  // "lazy" prim-ops take unevaluated args to eval-or-not as needed. eg. `@match`, `@fn` etc
	moPrimOpsEager map[MoValIdent]moFnEager // "eager" prim-ops receive already-evaluated args like any other func. eg. prim-type intrinsics like arithmetics, list concat etc
)

const (
	moPrimOpQuote         MoValIdent = "#"
	moPrimOpQQuote        MoValIdent = "#$"
	moPrimOpUnquote       MoValIdent = "$"
	moPrimOpSpliceUnquote MoValIdent = "$$"
	moPrimOpDo            MoValIdent = "@do"
)

func init() {
	moPrimOpsLazy = map[MoValIdent]moFnLazy{
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
	}
	moPrimOpsEager = map[MoValIdent]moFnEager{
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
}

// lazy prim-ops first, eager prim-ops afterwards

func (me *Interp) primOpFnCall(env *MoEnv, args ...*MoExpr) (*MoEnv, *MoExpr) {
	if err := me.checkCount(2, 2, args); err != nil {
		return nil, me.exprNever(err)
	}
	if err := me.checkIs(MoPrimTypeList, args[1]); err != nil {
		return nil, me.exprNever(err)
	}
	callee, args_list := args[0], args[1].Val.(MoValList)
	return env, me.expr(append(MoValCall{callee}, args_list...), nil, nil, args...)
}

func (me *Interp) primOpSet(env *MoEnv, args ...*MoExpr) (*MoEnv, *MoExpr) {
	if err := me.checkCount(2, 2, args); err != nil {
		return nil, me.exprNever(err)
	}
	if err := me.checkIs(MoPrimTypeIdent, args[0]); err != nil {
		return nil, me.exprNever(err)
	}
	name := args[0].Val.(MoValIdent)
	if is_reserved := ((name[0] == '@') || (name[0] == moPrimOpQuote[0]) || (name[0] == moPrimOpUnquote[0]) || moPrimOpsLazy[name] != nil); is_reserved {
		return nil, me.exprNever(me.diagSpan(false, true, args[0]).newDiagErr(NoticeCodeReserved, name, string(rune(name[0]))))
	}
	owner_env, found := env.lookupOwner(name)
	if owner_env == nil {
		owner_env = env
	}

	const can_set_macros = false
	if (!can_set_macros) && (found != nil) {
		if fn, _ := found.Val.(*MoValFnLam); (fn != nil) && fn.IsMacro {
			return nil, me.exprNever(me.diagSpan(true, false, args...).newDiagErr(NoticeCodeAtmoTodo, "mutating macros currently disabled, let us know whether you disagree with that or not"))
		}
	}
	new_value := me.evalAndApply(env, args[1])
	owner_env.set(name, new_value)
	return nil, me.exprFrom(moValNone, args...)
}

func (me *Interp) primOpDo(env *MoEnv, args ...*MoExpr) (*MoEnv, *MoExpr) {
	if err := me.checkCount(1, 1, args); err != nil {
		return nil, me.exprNever(err)
	}
	if err := me.checkIs(MoPrimTypeList, args[0]); err != nil {
		return nil, me.exprNever(err)
	}
	list := args[0].Val.(MoValList)
	for _, item := range list[:len(list)-1] {
		if expr := me.evalAndApply(env, item); expr.IsErr() {
			return nil, expr
		}
	}
	return env, me.exprFrom(list[len(list)-1], args...)
}

func (me *Interp) primOpMacro(env *MoEnv, args ...*MoExpr) (*MoEnv, *MoExpr) {
	_, expr := me.primOpFn(env, args...)
	if expr.IsErr() {
		return nil, expr
	}
	expr.Val.(*MoValFnLam).IsMacro = true
	return nil, expr
}

func (me *Interp) primOpFn(env *MoEnv, args ...*MoExpr) (*MoEnv, *MoExpr) {
	if err := me.checkCount(2, 2, args); err != nil {
		return nil, me.exprNever(err)
	}
	if err := me.checkIsListOf(MoPrimTypeIdent, args[0]); err != nil {
		return nil, me.exprNever(err)
	}
	if err := me.checkIs(MoPrimTypeList, args[1]); err != nil {
		return nil, me.exprNever(err)
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
	return nil, expr
}

func (me *Interp) primOpQuote(_ *MoEnv, args ...*MoExpr) (*MoEnv, *MoExpr) {
	if err := me.checkCount(1, 1, args); err != nil {
		return nil, me.exprNever(err)
	}
	return nil, me.exprFrom(args[0], args...)
}

func (me *Interp) primOpQuasiQuote(env *MoEnv, args ...*MoExpr) (*MoEnv, *MoExpr) {
	if err := me.checkCount(1, 1, args); err != nil {
		return nil, me.exprNever(err)
	}

	// is atomic arg? then just like primOpQuote
	if args[0].Val.PrimType().isAtomic() {
		return nil, me.exprFrom(args[0], args...)
	}

	// is this call directly quoting an unquote call?
	if is_unquote, err := me.checkIsCallOnIdent(args[0], moPrimOpUnquote, 1); err != nil {
		return nil, me.exprNever(err)
	} else if is_unquote {
		return nil, me.exprFrom(me.evalAndApply(env, args[0].Val.(MoValCall)[1]), args...)
	}

	// dict? call ourselves on each key and value
	if dict, is := args[0].Val.(MoValDict); is {
		ret := make(MoValDict, 0, len(dict))
		for _, pair := range dict {
			k, v := pair[0], pair[1]
			_, key := me.primOpQuasiQuote(env, k)
			_, val := me.primOpQuasiQuote(env, v)
			if dict.Has(key) {
				return nil, me.exprNever(k.SrcSpan.newDiagErr(NoticeCodeDictDuplKey, key))
			}
			ret.Set(key, val)
		}
		return nil, me.expr(ret, me.srcFile(false, true, args...), me.diagSpan(false, true, args...))
	}

	// must be list or call then: we handle them the same, per item iteration

	is_list := (args[0].Val.PrimType() == MoPrimTypeList)
	var call_or_arr MoExprs
	if call, is := args[0].Val.(MoValCall); is {
		call_or_arr = MoExprs(call)
	} else if is_list {
		call_or_arr = MoExprs(args[0].Val.(MoValList))
	} else {
		return nil, me.exprNever(me.diagSpan(false, true, args[0]).newDiagErr(NoticeCodeAtmoTodo, "NEW BUG intro'd in primOpQuasiQuote"))
	}

	ret := make(MoExprs, 0, len(call_or_arr))
	for _, item := range call_or_arr {
		if is_unquote, err := me.checkIsCallOnIdent(item, moPrimOpUnquote, 1); err != nil {
			return nil, me.exprNever(err)
		} else if is_unquote {
			unquotee := item.Val.(MoValCall)[1]
			ret = append(ret, me.evalAndApply(env, unquotee))
		} else if is_splice_unquote, err := me.checkIsCallOnIdent(item, moPrimOpSpliceUnquote, 1); err != nil {
			return nil, me.exprNever(err)
		} else if is_splice_unquote {
			unquotee := item.Val.(MoValCall)[1]
			evaled := me.evalAndApply(env, unquotee)
			if evaled.IsErr() {
				return nil, me.exprFrom(evaled)
			}
			if err = me.checkIsListOf(-1, evaled); err != nil {
				return nil, me.exprNever(err)
			}
			for _, splicee := range evaled.Val.(MoValList) {
				ret = append(ret, me.evalAndApply(env, splicee))
			}
		} else {
			_, evaled := me.primOpQuasiQuote(env, item)
			ret = append(ret, evaled)
		}
	}
	return nil, me.expr(util.If[MoVal](is_list, MoValList(ret), MoValCall(ret)),
		me.srcFile(false, true, args...), me.diagSpan(false, true, args...))
}

func (me *Interp) primOpMacroExpand(env *MoEnv, args ...*MoExpr) (*MoEnv, *MoExpr) {
	if err := me.checkCount(1, 1, args); err != nil {
		return nil, me.exprNever(err)
	}
	if err := me.checkIs(MoPrimTypeCall, args[0]); err != nil {
		return nil, me.exprNever(err)
	}
	return nil, me.exprFrom(me.macroExpand(env, args[0]), args...)
}

func (me *Interp) primOpCaseOf(env *MoEnv, args ...*MoExpr) (*MoEnv, *MoExpr) {
	if err := me.check(MoPrimTypeDict, 1, 1, args...); err != nil {
		return nil, me.exprNever(err)
	}
	pairs := args[0].Val.(MoValDict)
	for _, pair := range pairs {
		key, val := pair[0], pair[1]
		pred := me.evalAndApply(env, key)
		if pred.IsErr() {
			return nil, me.exprFrom(pred)
		} else if pred.EqTrue() {
			return env, me.exprFrom(val, val)
		} else if !pred.EqFalse() {
			return nil, me.exprNever(me.newErrExpectedBool(key))
		}
	}
	return nil, me.exprNever(me.diagSpan(true, false, args...).newDiagErr(NoticeCodeNoElseCase))
}

func (me *Interp) primOpBoolAnd(env *MoEnv, args ...*MoExpr) (*MoEnv, *MoExpr) {
	if err := me.checkCount(2, 2, args); err != nil {
		return nil, me.exprNever(err)
	}
	for _, arg := range args {
		evaled := me.evalAndApply(env, arg)
		if evaled.IsErr() {
			return nil, me.exprFrom(evaled)
		} else if evaled.EqFalse() {
			return nil, me.exprBool(false, args...)
		} else if !evaled.EqTrue() {
			return nil, me.exprNever(me.newErrExpectedBool(arg))
		}
	}
	return nil, me.exprBool(true, args...)
}

func (me *Interp) primOpBoolOr(env *MoEnv, args ...*MoExpr) (*MoEnv, *MoExpr) {
	if err := me.checkCount(2, 2, args); err != nil {
		return nil, me.exprNever(err)
	}
	for _, arg := range args {
		evaled := me.evalAndApply(env, arg)
		if evaled.IsErr() {
			return nil, me.exprFrom(evaled)
		} else if evaled.EqTrue() {
			return nil, me.exprBool(true, args...)
		} else if !evaled.EqFalse() {
			return nil, me.exprNever(me.newErrExpectedBool(arg))
		}
	}
	return nil, me.exprBool(false, args...)
}

func (me *Interp) newErrExpectedBool(noBool *MoExpr) *SrcFileNotice {
	return me.diagSpan(false, true, noBool).newDiagErr(NoticeCodeExpectedFoo, "a boolean expression instead of `"+noBool.String()+"`")
}

// eager prim-ops below, lazy ones above

func makeArithPrimOp[T MoValNumInt | MoValNumUint | MoValNumFloat](t MoValPrimType, f func(opl MoVal, opr MoVal) MoVal) moFnEager {
	return func(me *Interp, _ *MoEnv, args ...*MoExpr) *MoExpr {
		if err := me.check(t, 2, 2, args...); err != nil {
			return me.exprNever(err)
		}
		return me.expr(f(args[0].Val, args[1].Val), nil, nil, args...)
	}
}

func (me *Interp) primFnSessEnv(env *MoEnv, args ...*MoExpr) *MoExpr {
	if err := me.checkCount(0, 0, args); err != nil {
		return me.exprNever(err)
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
	return ret
}

func (me *Interp) primFnSessPrintf(_ *MoEnv, args ...*MoExpr) *MoExpr {
	if err := me.checkCount(2, 2, args); err != nil {
		return me.exprNever(err)
	}
	if err := me.checkIs(MoPrimTypeStr, args[0]); err != nil {
		return me.exprNever(err)
	}
	if err := me.checkIs(MoPrimTypeList, args[1]); err != nil {
		return me.exprNever(err)
	}
	list := args[1].Val.(MoValList)
	fmt_args := make([]any, 0, len(list))
	for _, arg := range list {
		fmt_args = append(fmt_args, arg.Val)
	}
	fmt.Fprintf(me.StdIo.Out, string(args[0].Val.(MoValStr)), fmt_args...)
	return me.exprFrom(moValNone, args...)
}

func (me *Interp) primFnSessPrint(_ *MoEnv, args ...*MoExpr) *MoExpr {
	if err := me.checkCount(1, 1, args); err != nil {
		return me.exprNever(err)
	}
	switch arg0 := args[0].Val.(type) {
	case MoValStr:
		me.StdIo.Out.WriteString(string(arg0))
	default:
		args[0].WriteTo(me.StdIo.Out)
	}
	return me.exprFrom(moValNone, args...)
}

func (me *Interp) primFnSessPrintln(env *MoEnv, args ...*MoExpr) *MoExpr {
	expr := me.primFnSessPrint(env, args...)
	if !expr.IsErr() {
		me.StdIo.Out.WriteString("\n")
	}
	return expr
}

func (me *Interp) primFnListLen(_ *MoEnv, args ...*MoExpr) *MoExpr {
	if err := me.checkCount(1, 1, args); err != nil {
		return me.exprNever(err)
	}
	if err := me.checkIs(MoPrimTypeList, args[0]); err != nil {
		return me.exprNever(err)
	}
	return me.expr(MoValNumUint(len(args[0].Val.(MoValList))), nil, nil, args...)
}

func (me *Interp) primFnListItemAt(_ *MoEnv, args ...*MoExpr) *MoExpr {
	if err := me.checkCount(2, 2, args); err != nil {
		return me.exprNever(err)
	}
	if err := me.checkIs(MoPrimTypeList, args[0]); err != nil {
		return me.exprNever(err)
	}
	if err := me.checkIs(MoPrimTypeNumUint, args[1]); err != nil {
		return me.exprNever(err)
	}
	list, idx := args[0].Val.(MoValList), args[1].Val.(MoValNumUint)
	if idx_downcast := int(idx); (idx_downcast < 0) || (idx_downcast >= len(list)) {
		return me.exprNever(me.diagSpan(false, true, args[1]).newDiagErr(NoticeCodeIndexOutOfBounds, idx_downcast, len(list)))
	}
	return me.exprFrom(list[idx], args...)
}

func (me *Interp) primFnListRange(_ *MoEnv, args ...*MoExpr) *MoExpr {
	if err := me.checkCount(2, 3, args); err != nil {
		return me.exprNever(err)
	}
	if err := me.checkIs(MoPrimTypeList, args[0]); err != nil {
		return me.exprNever(err)
	}
	if err := me.checkIs(MoPrimTypeNumUint, args[1]); err != nil {
		return me.exprNever(err)
	}
	list, idx_start := args[0].Val.(MoValList), args[1].Val.(MoValNumUint)
	if len(args) == 2 {
		args = append(args, me.expr(MoValNumUint(len(list)), me.srcFile(false, true, args[1]), me.diagSpan(false, true, args[1])))
	} else if err := me.checkIs(MoPrimTypeNumUint, args[2]); err != nil {
		return me.exprNever(err)
	}
	idx_end := args[2].Val.(MoValNumUint)
	if idx_end < idx_start {
		return me.exprNever(me.diagSpan(false, true, args[2]).newDiagErr(NoticeCodeRangeNegative, idx_end, idx_start))
	} else if idx_downcast := int(idx_start); (idx_downcast < 0) || (idx_downcast > len(list)) {
		return me.exprNever(me.diagSpan(false, true, args[1]).newDiagErr(NoticeCodeIndexOutOfBounds, idx_downcast, len(list)))
	}
	if idx_downcast := int(idx_end); (idx_downcast < 0) || (idx_downcast > len(list)) {
		return me.exprNever(me.diagSpan(false, true, args[2]).newDiagErr(NoticeCodeIndexOutOfBounds, idx_downcast, len(list)))
	}
	return me.expr(list[idx_start:idx_end], me.srcFile(false, idx_start != idx_end, list[idx_start:idx_end]...), me.diagSpan(false, idx_start != idx_end, list[idx_start:idx_end]...))
}

func (me *Interp) primFnListConcat(_ *MoEnv, args ...*MoExpr) *MoExpr {
	ret := make(MoValList, 0, len(args)*3)
	for _, arg := range args {
		if err := me.checkIs(MoPrimTypeList, arg); err != nil {
			return me.exprNever(err)
		}
		ret = append(ret, arg.Val.(MoValList)...)
	}
	return me.expr(ret, nil, nil, args...)
}

func (me *Interp) primFnStrCharAt(_ *MoEnv, args ...*MoExpr) *MoExpr {
	if err := me.checkCount(2, 2, args); err != nil {
		return me.exprNever(err)
	}
	if err := me.checkIs(MoPrimTypeStr, args[0]); err != nil {
		return me.exprNever(err)
	}
	str := args[0].Val.(MoValStr)
	if err := me.checkIs(MoPrimTypeNumUint, args[1]); err != nil {
		return me.exprNever(err)
	}
	idx := int(args[1].Val.(MoValNumUint))
	if (idx < 0) || (idx >= len(str)) {
		return me.exprNever(me.diagSpan(false, true, args[1]).newDiagErr(NoticeCodeIndexOutOfBounds, idx, len(str)))
	}
	return me.expr(MoValChar(str[idx]), nil, nil, args...)
}

func (me *Interp) primFnStrRange(_ *MoEnv, args ...*MoExpr) *MoExpr {
	if err := me.checkCount(2, 3, args); err != nil {
		return me.exprNever(err)
	}
	if err := me.checkIs(MoPrimTypeStr, args[0]); err != nil {
		return me.exprNever(err)
	}
	str := args[0].Val.(MoValStr)
	if err := me.checkIs(MoPrimTypeNumUint, args[1]); err != nil {
		return me.exprNever(err)
	}
	idx_start, idx_end := int(args[1].Val.(MoValNumUint)), len(str)
	if (idx_start < 0) || (idx_start > len(str)) {
		return me.exprNever(me.diagSpan(false, true, args[1]).newDiagErr(NoticeCodeIndexOutOfBounds, idx_start, len(str)))
	}
	if len(args) > 2 {
		if err := me.checkIs(MoPrimTypeNumUint, args[2]); err != nil {
			return me.exprNever(err)
		}
		idx_end = int(args[2].Val.(MoValNumUint))
	}
	if (idx_end < 0) || (idx_end > len(str)) {
		return me.exprNever(me.diagSpan(false, true, args[1]).newDiagErr(NoticeCodeIndexOutOfBounds, idx_end, len(str)))
	} else if idx_end < idx_start {
		return me.exprNever(me.diagSpan(false, true, args[1]).newDiagErr(NoticeCodeRangeNegative, idx_end, idx_start))
	}
	return me.expr(MoValStr(str[idx_start:idx_end]), nil, nil, args...)
}

func (me *Interp) primFnStrLen(_ *MoEnv, args ...*MoExpr) *MoExpr {
	if err := me.checkCount(1, 1, args); err != nil {
		return me.exprNever(err)
	}
	if err := me.checkIs(MoPrimTypeStr, args[0]); err != nil {
		return me.exprNever(err)
	}
	return me.expr(MoValNumUint(len(args[0].Val.(MoValStr))), nil, nil, args...)
}

func (me *Interp) primFnStrConcat(_ *MoEnv, args ...*MoExpr) *MoExpr {
	if err := me.checkCount(1, 1, args); err != nil {
		return me.exprNever(err)
	}
	if err := me.checkIs(MoPrimTypeList, args[0]); err != nil {
		return me.exprNever(err)
	}
	var buf strings.Builder
	for _, arg := range args[0].Val.(MoValList) {
		if err := me.checkIs(MoPrimTypeStr, arg); err != nil {
			return me.exprNever(err)
		}
		buf.WriteString(string(arg.Val.(MoValStr)))
	}
	return me.expr(MoValStr(buf.String()), nil, nil, args...)
}

func (me *Interp) primFnStr(_ *MoEnv, args ...*MoExpr) *MoExpr {
	if err := me.checkCount(1, 1, args); err != nil {
		return me.exprNever(err)
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
	return me.expr(str, nil, nil, args...)
}

func (me *Interp) primFnExprStr(_ *MoEnv, args ...*MoExpr) *MoExpr {
	if err := me.checkCount(1, 1, args); err != nil {
		return me.exprNever(err)
	}
	return me.expr(MoValStr(args[0].String()), nil, nil, args...)
}

func (me *Interp) primFnExprEval(_ *MoEnv, args ...*MoExpr) *MoExpr {
	if err := me.checkCount(1, 1, args); err != nil {
		return me.exprNever(err)
	}
	return me.exprFrom(me.evalAndApply(me.Env, args[0]), args...)
}

func (me *Interp) primFnExprParse(_ *MoEnv, args ...*MoExpr) *MoExpr {
	if err := me.checkCount(1, 1, args); err != nil {
		return me.exprNever(err)
	}
	if err := me.checkIs(MoPrimTypeStr, args[0]); err != nil {
		return me.exprNever(err)
	}
	ret, err := me.ParseExpr(string(args[0].Val.(MoValStr)))
	if err != nil {
		return me.exprNever(err)
	}
	return ret
}

func (me *Interp) primFnDictHas(_ *MoEnv, args ...*MoExpr) *MoExpr {
	if err := me.checkCount(2, 2, args); err != nil {
		return me.exprNever(err)
	}
	if err := me.checkIs(MoPrimTypeDict, args[0]); err != nil {
		return me.exprNever(err)
	}
	return me.exprBool(args[0].Val.(MoValDict).Has(args[1]), args...)
}

func (me *Interp) primFnDictGet(_ *MoEnv, args ...*MoExpr) *MoExpr {
	if err := me.checkCount(2, 2, args); err != nil {
		return me.exprNever(err)
	}
	if err := me.checkIs(MoPrimTypeDict, args[0]); err != nil {
		return me.exprNever(err)
	}
	found := args[0].Val.(MoValDict).Get(args[1])
	if found == nil {
		return me.exprFrom(moValNone, args...)
	}
	return me.exprFrom(found, args...)
}

func (me *Interp) primFnDictWith(_ *MoEnv, args ...*MoExpr) *MoExpr {
	if err := me.checkCount(3, 3, args); err != nil {
		return me.exprNever(err)
	}
	if err := me.checkIs(MoPrimTypeDict, args[0]); err != nil {
		return me.exprNever(err)
	}
	ret := me.exprFrom(args[0], args...)
	ret.Val = ret.Val.(MoValDict).With(args[1], args[2])
	return ret
}

func (me *Interp) primFnDictWithout(_ *MoEnv, args ...*MoExpr) *MoExpr {
	if err := me.checkCount(2, 2, args); err != nil {
		return me.exprNever(err)
	}
	if err := me.checkIs(MoPrimTypeDict, args[0]); err != nil {
		return me.exprNever(err)
	}
	if err := me.checkIs(MoPrimTypeList, args[1]); err != nil {
		return me.exprNever(err)
	}
	ret := me.exprFrom(args[0], args...)
	ret.Val = ret.Val.(MoValDict).Without(args[1].Val.(MoValList)...)
	return ret
}

func (me *Interp) primFnPrimTypeTag(_ *MoEnv, args ...*MoExpr) *MoExpr {
	if err := me.checkCount(1, 1, args); err != nil {
		return me.exprNever(err)
	}
	return me.expr(MoValType(args[0].Val.PrimType()), nil, nil, args...)
}

func (me *Interp) primFnSessReset(_ *MoEnv, args ...*MoExpr) *MoExpr {
	me.reset()
	return moValNone
}

func (me *Interp) primFnEq(_ *MoEnv, args ...*MoExpr) *MoExpr {
	if err := me.checkCount(2, 2, args); err != nil {
		return me.exprNever(err)
	}
	if !me.checkNoneArePrimFuncs(args...) {
		return me.exprNever(me.diagSpan(false, true, args...).newDiagErr(NoticeCodeNotComparable, args[0], args[1], "equality"))
	}
	return me.exprBool(args[0].Eq(args[1]), args...)
}

func (me *Interp) primFnNeq(env *MoEnv, args ...*MoExpr) *MoExpr {
	if err := me.checkCount(2, 2, args); err != nil {
		return me.exprNever(err)
	}
	if !me.checkNoneArePrimFuncs(args...) {
		return me.exprNever(me.diagSpan(false, true, args...).newDiagErr(NoticeCodeNotComparable, args[0], args[1], "not-equal"))
	}
	return me.exprBool(!args[0].Eq(args[1]), args...)
}

func (me *Interp) primFnLeq(_ *MoEnv, args ...*MoExpr) *MoExpr {
	cmp, err := me.primCmpHelper("less-or-equal", args...)
	if err != nil {
		return me.exprNever(err)
	}
	return me.exprBool(cmp <= 0, args...)
}

func (me *Interp) primFnGeq(_ *MoEnv, args ...*MoExpr) *MoExpr {
	cmp, err := me.primCmpHelper("greater-or-equal", args...)
	if err != nil {
		return me.exprNever(err)
	}
	return me.exprBool(cmp >= 0, args...)
}

func (me *Interp) primFnLt(_ *MoEnv, args ...*MoExpr) *MoExpr {
	cmp, err := me.primCmpHelper("less-than", args...)
	if err != nil {
		return me.exprNever(err)
	}
	return me.exprBool(cmp < 0, args...)
}

func (me *Interp) primFnGt(_ *MoEnv, args ...*MoExpr) *MoExpr {
	cmp, err := me.primCmpHelper("greater-than", args...)
	if err != nil {
		return me.exprNever(err)
	}
	return me.exprBool(cmp > 0, args...)
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
