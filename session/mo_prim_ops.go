package session

import (
	"strings"

	"atmo/util"
	"atmo/util/str"
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
	moPrimOpSet           MoValIdent = "@set"
	moPrimOpCaseOf        MoValIdent = "@caseOf"
	moPrimOpAnd           MoValIdent = "@and"
	moPrimOpOr            MoValIdent = "@or"
	moPrimOpMacro         MoValIdent = "@macro"
	moPrimOpExpand        MoValIdent = "@expand"
	moPrimOpFn            MoValIdent = "@fn"
	moPrimOpFnCall        MoValIdent = "@fnCall"

	moPrimFnReplEnv     MoValIdent = "@replEnv"
	moPrimFnReplPrint   MoValIdent = "@replPrint"
	moPrimFnReplReset   MoValIdent = "@replReset"
	moPrimFnNumIntAdd   MoValIdent = "@numIntAdd"
	moPrimFnNumIntSub   MoValIdent = "@numIntSub"
	moPrimFnNumIntMul   MoValIdent = "@numIntMul"
	moPrimFnNumIntDiv   MoValIdent = "@numIntDiv"
	moPrimFnNumIntMod   MoValIdent = "@numIntMod"
	moPrimFnNumUintAdd  MoValIdent = "@numUintAdd"
	moPrimFnNumUintSub  MoValIdent = "@numUintSub"
	moPrimFnNumUintMul  MoValIdent = "@numUintMul"
	moPrimFnNumUintDiv  MoValIdent = "@numUintDiv"
	moPrimFnNumUintMod  MoValIdent = "@numUintMod"
	moPrimFnNumFloatAdd MoValIdent = "@numFloatAdd"
	moPrimFnNumFloatSub MoValIdent = "@numFloatSub"
	moPrimFnNumFloatMul MoValIdent = "@numFloatMul"
	moPrimFnNumFloatDiv MoValIdent = "@numFloatDiv"
	moPrimFnCast        MoValIdent = "@cast"
	moPrimFnNot         MoValIdent = "@not"
	moPrimFnEq          MoValIdent = "@eq"
	moPrimFnNeq         MoValIdent = "@neq"
	moPrimFnGeq         MoValIdent = "@geq"
	moPrimFnLeq         MoValIdent = "@leq"
	moPrimFnLt          MoValIdent = "@lt"
	moPrimFnGt          MoValIdent = "@gt"
	moPrimFnPrimTypeTag MoValIdent = "@primTypeTag"
	moPrimFnListGet     MoValIdent = "@listGet"
	moPrimFnListSet     MoValIdent = "@listSet"
	moPrimFnListRange   MoValIdent = "@listRange"
	moPrimFnListLen     MoValIdent = "@listLen"
	moPrimFnListConcat  MoValIdent = "@listConcat"
	moPrimFnDictHas     MoValIdent = "@dictHas"
	moPrimFnDictGet     MoValIdent = "@dictGet"
	moPrimFnDictSet     MoValIdent = "@dictSet"
	moPrimFnDictDel     MoValIdent = "@dictDel"
	moPrimFnDictLen     MoValIdent = "@dictLen"
	moPrimFnErrNew      MoValIdent = "@errNew"
	moPrimFnErrVal      MoValIdent = "@errVal"
	moPrimFnStrConcat   MoValIdent = "@strConcat"
	moPrimFnStrLen      MoValIdent = "@strLen"
	moPrimFnStrCharAt   MoValIdent = "@strCharAt"
	moPrimFnStrRange    MoValIdent = "@strRange"
	moPrimFnStr         MoValIdent = "@str"
	moPrimFnExprStr     MoValIdent = "@exprStr"
	moPrimFnExprParse   MoValIdent = "@exprParse"
	moPrimFnExprEval    MoValIdent = "@exprEval"
)

func init() {
	moPrimOpsLazy = map[MoValIdent]moFnLazy{
		moPrimOpFn:     (*Interp).primOpFn,
		moPrimOpFnCall: (*Interp).primOpFnCall,
		moPrimOpCaseOf: (*Interp).primOpCaseOf,
		moPrimOpAnd:    (*Interp).primOpBoolAnd,
		moPrimOpOr:     (*Interp).primOpBoolOr,
		moPrimOpMacro:  (*Interp).primOpMacro,
		moPrimOpExpand: (*Interp).primOpMacroExpand,
		moPrimOpQuote:  (*Interp).primOpQuote,
		moPrimOpQQuote: (*Interp).primOpQuasiQuote,
		moPrimOpSet:    (*Interp).primOpSet,
		moPrimOpDo:     (*Interp).primOpDo,
	}
	moPrimOpsEager = map[MoValIdent]moFnEager{
		moPrimFnReplEnv:     (*Interp).primFnReplEnv,
		moPrimFnReplPrint:   (*Interp).primFnReplPrint,
		moPrimFnReplReset:   (*Interp).primFnReplReset,
		moPrimFnNumIntAdd:   moPrimFnArith[MoValNumInt](MoPrimTypeNumInt, func(opl MoVal, opr MoVal) MoVal { return opl.(MoValNumInt) + opr.(MoValNumInt) }),
		moPrimFnNumIntSub:   moPrimFnArith[MoValNumInt](MoPrimTypeNumInt, func(opl MoVal, opr MoVal) MoVal { return opl.(MoValNumInt) - opr.(MoValNumInt) }),
		moPrimFnNumIntMul:   moPrimFnArith[MoValNumInt](MoPrimTypeNumInt, func(opl MoVal, opr MoVal) MoVal { return opl.(MoValNumInt) * opr.(MoValNumInt) }),
		moPrimFnNumIntDiv:   moPrimFnArith[MoValNumInt](MoPrimTypeNumInt, func(opl MoVal, opr MoVal) MoVal { return opl.(MoValNumInt) / opr.(MoValNumInt) }),
		moPrimFnNumIntMod:   moPrimFnArith[MoValNumInt](MoPrimTypeNumInt, func(opl MoVal, opr MoVal) MoVal { return opl.(MoValNumInt) % opr.(MoValNumInt) }),
		moPrimFnNumUintAdd:  moPrimFnArith[MoValNumUint](MoPrimTypeNumUint, func(opl MoVal, opr MoVal) MoVal { return opl.(MoValNumUint) + opr.(MoValNumUint) }),
		moPrimFnNumUintSub:  moPrimFnArith[MoValNumUint](MoPrimTypeNumUint, func(opl MoVal, opr MoVal) MoVal { return opl.(MoValNumUint) - opr.(MoValNumUint) }),
		moPrimFnNumUintMul:  moPrimFnArith[MoValNumUint](MoPrimTypeNumUint, func(opl MoVal, opr MoVal) MoVal { return opl.(MoValNumUint) * opr.(MoValNumUint) }),
		moPrimFnNumUintDiv:  moPrimFnArith[MoValNumUint](MoPrimTypeNumUint, func(opl MoVal, opr MoVal) MoVal { return opl.(MoValNumUint) / opr.(MoValNumUint) }),
		moPrimFnNumUintMod:  moPrimFnArith[MoValNumUint](MoPrimTypeNumUint, func(opl MoVal, opr MoVal) MoVal { return opl.(MoValNumUint) % opr.(MoValNumUint) }),
		moPrimFnNumFloatAdd: moPrimFnArith[MoValNumFloat](MoPrimTypeNumFloat, func(opl MoVal, opr MoVal) MoVal { return opl.(MoValNumFloat) + opr.(MoValNumFloat) }),
		moPrimFnNumFloatSub: moPrimFnArith[MoValNumFloat](MoPrimTypeNumFloat, func(opl MoVal, opr MoVal) MoVal { return opl.(MoValNumFloat) - opr.(MoValNumFloat) }),
		moPrimFnNumFloatMul: moPrimFnArith[MoValNumFloat](MoPrimTypeNumFloat, func(opl MoVal, opr MoVal) MoVal { return opl.(MoValNumFloat) * opr.(MoValNumFloat) }),
		moPrimFnNumFloatDiv: moPrimFnArith[MoValNumFloat](MoPrimTypeNumFloat, func(opl MoVal, opr MoVal) MoVal { return opl.(MoValNumFloat) / opr.(MoValNumFloat) }),
		moPrimFnNot:         (*Interp).primFnBoolNot,
		moPrimFnCast:        (*Interp).primFnCast,
		moPrimFnEq:          (*Interp).primFnEq,
		moPrimFnNeq:         (*Interp).primFnNeq,
		moPrimFnGeq:         (*Interp).primFnGeq,
		moPrimFnLeq:         (*Interp).primFnLeq,
		moPrimFnLt:          (*Interp).primFnLt,
		moPrimFnGt:          (*Interp).primFnGt,
		moPrimFnPrimTypeTag: (*Interp).primFnPrimTypeTag,
		moPrimFnListGet:     (*Interp).primFnListGet,
		moPrimFnListSet:     (*Interp).primFnListSet,
		moPrimFnListRange:   (*Interp).primFnListRange,
		moPrimFnListLen:     (*Interp).primFnListLen,
		moPrimFnListConcat:  (*Interp).primFnListConcat,
		moPrimFnDictHas:     (*Interp).primFnDictHas,
		moPrimFnDictGet:     (*Interp).primFnDictGet,
		moPrimFnDictSet:     (*Interp).primFnDictSet,
		moPrimFnDictDel:     (*Interp).primFnDictDel,
		moPrimFnDictLen:     (*Interp).primFnDictLen,
		moPrimFnErrNew:      (*Interp).primFnErrNew,
		moPrimFnErrVal:      (*Interp).primFnErrVal,
		moPrimFnStrConcat:   (*Interp).primFnStrConcat,
		moPrimFnStrLen:      (*Interp).primFnStrLen,
		moPrimFnStrCharAt:   (*Interp).primFnStrCharAt,
		moPrimFnStrRange:    (*Interp).primFnStrRange,
		moPrimFnStr:         (*Interp).primFnStr,
		moPrimFnExprStr:     (*Interp).primFnExprStr,
		moPrimFnExprParse:   (*Interp).primFnExprParse,
		moPrimFnExprEval:    (*Interp).primFnExprEval,
	}
}

// lazy prim-ops first, eager prim-ops afterwards

func (me *Interp) primOpFnCall(env *MoEnv, args ...*MoExpr) (*MoEnv, *MoExpr) {
	if err := me.checkCount(2, 2, args); err != nil {
		return nil, me.exprErr(err)
	}
	if err := me.checkIs(MoPrimTypeList, args[1]); err != nil {
		return nil, me.exprErr(err)
	}
	callee, args_list := args[0], args[1].Val.(*MoValList)
	return env, me.expr(append(MoValCall{callee}, (*args_list)...), nil, nil, args...)
}

func (me *Interp) primOpSet(env *MoEnv, args ...*MoExpr) (*MoEnv, *MoExpr) {
	if err := me.checkCount(2, 2, args); err != nil {
		return nil, me.exprErr(err)
	}
	if err := me.checkIs(MoPrimTypeIdent, args[0]); err != nil {
		return nil, me.exprErr(err)
	}
	name := args[0].Val.(MoValIdent)
	if name.IsReserved() {
		return nil, me.exprErr(me.diagSpan(false, true, args[0]).newDiagErr(ErrCodeReserved, name, string(rune(name[0]))), args[0])
	}
	owner_env, found := env.lookupOwner(name)
	if owner_env == nil {
		owner_env = env
	}

	const can_set_macros = false
	if (!can_set_macros) && (found != nil) {
		if fn, _ := found.Val.(*MoValFnLam); (fn != nil) && fn.IsMacro {
			return nil, me.exprErr(me.diagSpan(true, false, args...).newDiagErr(ErrCodeAtmoTodo, "mutating macros currently disabled, let us know whether you disagree with that or not"))
		}
	}
	new_value := me.evalAndApply(env, args[1])
	if err := new_value.Err(); err != nil {
		return nil, me.exprErr(err, args[1])
	}
	owner_env.set(name, new_value)
	return nil, me.exprVoid(args...)
}

func (me *Interp) primOpDo(env *MoEnv, args ...*MoExpr) (*MoEnv, *MoExpr) {
	if err := me.checkCount(1, 1, args); err != nil {
		return nil, me.exprErr(err)
	}
	if err := me.checkIs(MoPrimTypeList, args[0]); err != nil {
		return nil, me.exprErr(err)
	}
	list := *(args[0].Val.(*MoValList))
	for _, item := range list[:len(list)-1] {
		evaled := me.evalAndApply(env, item)
		if err := evaled.Err(); err != nil {
			return nil, me.exprErr(err, item)
		}
	}
	return env, me.exprFrom(list[len(list)-1], args...)
}

func (me *Interp) primOpMacro(env *MoEnv, args ...*MoExpr) (*MoEnv, *MoExpr) {
	_, expr := me.primOpFn(env, args...)
	if lam, _ := expr.Val.(*MoValFnLam); lam != nil {
		lam.IsMacro = true
	}
	return nil, expr
}

func (me *Interp) primOpFn(env *MoEnv, args ...*MoExpr) (*MoEnv, *MoExpr) {
	if err := me.checkCount(2, 2, args); err != nil {
		return nil, me.exprErr(err)
	}
	if err := me.checkIsListOf(MoPrimTypeIdent, args[0]); err != nil {
		return nil, me.exprErr(err)
	}
	for _, param := range *(args[0].Val.(*MoValList)) {
		if ident := param.Val.(MoValIdent); ident.IsReserved() {
			return nil, me.exprErr(param.SrcNode.newDiagErr(false, ErrCodeReserved, ident), param)
		}
	}
	if err := me.checkIs(MoPrimTypeList, args[1]); err != nil {
		return nil, me.exprErr(err)
	}
	list := *(args[1].Val.(*MoValList))
	body := list[0]
	if len(list) > 1 {
		do := me.expr(moPrimOpDo, body.SrcFile, srcSpanFrom(MoExprs(list)))
		body = me.expr(MoValCall{do, me.expr(&list, do.SrcFile, do.SrcSpan)}, do.SrcFile, do.SrcSpan)
	}
	expr := me.expr(
		&MoValFnLam{Params: MoExprs(*(args[0].Val.(*MoValList))), Body: body, Env: env},
		body.SrcFile, srcSpanFrom(args))
	return nil, expr
}

func (me *Interp) primOpQuote(_ *MoEnv, args ...*MoExpr) (*MoEnv, *MoExpr) {
	if err := me.checkCount(1, 1, args); err != nil {
		return nil, me.exprErr(err)
	}
	return nil, me.exprFrom(args[0], args...)
}

func (me *Interp) primOpQuasiQuote(env *MoEnv, args ...*MoExpr) (*MoEnv, *MoExpr) {
	if err := me.checkCount(1, 1, args); err != nil {
		return nil, me.exprErr(err)
	}

	// is atomic arg? then just like primOpQuote
	if args[0].Val.PrimType().isAtomic() {
		return nil, me.exprFrom(args[0], args...)
	}

	// is this call directly quoting an unquote call?
	if is_unquote, err := me.checkIsCallOnIdent(args[0], moPrimOpUnquote, 1); err != nil {
		return nil, me.exprErr(err, args...)
	} else if is_unquote {
		evaled := me.evalAndApply(env, args[0].Val.(MoValCall)[1])
		if err := evaled.Err(); err != nil {
			return nil, me.exprErr(err, args[0])
		}
		return nil, me.exprFrom(evaled, args[0])
	}

	// dict? call ourselves on each key and value
	if dict, is := args[0].Val.(*MoValDict); is {
		ret := make(MoValDict, 0, len(*dict))
		for _, entry := range *dict {
			_, key := me.primOpQuasiQuote(env, entry.Key)
			if err := key.Err(); err != nil {
				return nil, me.exprErr(err, entry.Key)
			}
			_, val := me.primOpQuasiQuote(env, entry.Val)
			if err := val.Err(); err != nil {
				return nil, me.exprErr(err, entry.Val)
			}
			if dict.Has(key) {
				return nil, me.exprErr(entry.Key.SrcSpan.newDiagErr(ErrCodeDictDuplKey, key))
			}
			ret.Set(key, val)
		}
		return nil, me.expr(&ret, me.srcFile(false, true, args...), me.diagSpan(false, true, args...))
	}

	// must be list or call then: we handle them the same, per item iteration

	is_list := (args[0].Val.PrimType() == MoPrimTypeList)
	var call_or_arr MoExprs
	if call, is := args[0].Val.(MoValCall); is {
		call_or_arr = MoExprs(call)
	} else if is_list {
		call_or_arr = MoExprs(*(args[0].Val.(*MoValList)))
	} else {
		return nil, me.exprErr(me.diagSpan(false, true, args[0]).newDiagErr(ErrCodeAtmoTodo, "NEW BUG intro'd in primOpQuasiQuote"))
	}

	ret := make(MoExprs, 0, len(call_or_arr))
	for _, item := range call_or_arr {
		if is_unquote, err := me.checkIsCallOnIdent(item, moPrimOpUnquote, 1); err != nil {
			return nil, me.exprErr(err)
		} else if is_unquote {
			unquotee := item.Val.(MoValCall)[1]
			evaled := me.evalAndApply(env, unquotee)
			if err = evaled.Err(); err != nil {
				return nil, me.exprErr(err, unquotee)
			}
			ret = append(ret, evaled)
		} else if is_splice_unquote, err := me.checkIsCallOnIdent(item, moPrimOpSpliceUnquote, 1); err != nil {
			return nil, me.exprErr(err)
		} else if is_splice_unquote {
			unquotee := item.Val.(MoValCall)[1]
			evaled := me.evalAndApply(env, unquotee)
			if err = evaled.Err(); err != nil {
				return nil, me.exprErr(err, unquotee)
			} else if err = me.checkIsListOf(-1, evaled); err != nil {
				return nil, me.exprErr(err)
			}
			for _, splicee := range *(evaled.Val.(*MoValList)) {
				evaled = me.evalAndApply(env, splicee)
				if err = evaled.Err(); err != nil {
					return nil, me.exprErr(err, splicee)
				}
				ret = append(ret, evaled)
			}
		} else {
			_, evaled := me.primOpQuasiQuote(env, item)
			if err = evaled.Err(); err != nil {
				return nil, me.exprErr(err, item)
			}
			ret = append(ret, evaled)
		}
	}
	return nil, me.expr(util.If[MoVal](is_list, util.Ptr(MoValList(ret)), MoValCall(ret)),
		me.srcFile(false, true, args...), me.diagSpan(false, true, args...))
}

func (me *Interp) primOpMacroExpand(env *MoEnv, args ...*MoExpr) (*MoEnv, *MoExpr) {
	if err := me.checkCount(1, 1, args); err != nil {
		return nil, me.exprErr(err)
	}
	if err := me.checkIs(MoPrimTypeCall, args[0]); err != nil {
		return nil, me.exprErr(err)
	}
	return nil, me.exprFrom(me.macroExpand(env, args[0]), args...)
}

func (me *Interp) primOpCaseOf(env *MoEnv, args ...*MoExpr) (*MoEnv, *MoExpr) {
	if err := me.check(MoPrimTypeDict, 1, 1, args...); err != nil {
		return nil, me.exprErr(err)
	}
	entries := args[0].Val.(*MoValDict)
	for _, entry := range *entries {
		pred := me.evalAndApply(env, entry.Key)
		if err := pred.Err(); err != nil {
			return nil, me.exprErr(err, entry.Key)
		} else if pred.EqTrue() {
			return env, me.exprFrom(entry.Val, entry.Val)
		} else if !pred.EqFalse() {
			return nil, me.exprErr(me.newErrExpectedBool(entry.Key))
		}
	}
	return nil, me.exprErr(me.diagSpan(true, false, args...).newDiagErr(ErrCodeNoElseCase))
}

func (me *Interp) primOpBoolAnd(env *MoEnv, args ...*MoExpr) (*MoEnv, *MoExpr) {
	if err := me.checkCount(2, 2, args); err != nil {
		return nil, me.exprErr(err)
	}
	for _, arg := range args {
		evaled := me.evalAndApply(env, arg)
		if err := evaled.Err(); err != nil {
			return nil, me.exprErr(err, arg)
		} else if evaled.EqFalse() {
			return nil, me.exprBool(false, args...)
		} else if !evaled.EqTrue() {
			return nil, me.exprErr(me.newErrExpectedBool(arg))
		}
	}
	return nil, me.exprBool(true, args...)
}

func (me *Interp) primOpBoolOr(env *MoEnv, args ...*MoExpr) (*MoEnv, *MoExpr) {
	if err := me.checkCount(2, 2, args); err != nil {
		return nil, me.exprErr(err)
	}
	for _, arg := range args {
		evaled := me.evalAndApply(env, arg)
		if err := evaled.Err(); err != nil {
			return nil, me.exprErr(err, arg)
		} else if evaled.EqTrue() {
			return nil, me.exprBool(true, args...)
		} else if !evaled.EqFalse() {
			return nil, me.exprErr(me.newErrExpectedBool(arg))
		}
	}
	return nil, me.exprBool(false, args...)
}

// eager prim-ops below, lazy ones above

func (me *Interp) primFnBoolNot(env *MoEnv, args ...*MoExpr) *MoExpr {
	if err := me.checkCount(1, 1, args); err != nil {
		return me.exprErr(err)
	}
	switch {
	case args[0].EqFalse():
		return me.exprBool(true, args...)
	case args[0].EqTrue():
		return me.exprBool(false, args...)
	default:
		return me.exprErr(me.newErrExpectedBool(args[0]))
	}
}

func moPrimFnArith[T MoValNumInt | MoValNumUint | MoValNumFloat](t MoValPrimType, f func(opl MoVal, opr MoVal) MoVal) moFnEager {
	return func(me *Interp, _ *MoEnv, args ...*MoExpr) *MoExpr {
		if err := me.check(t, 2, 2, args...); err != nil {
			return me.exprErr(err)
		}
		if err := me.checkArgErrs(args...); err != nil {
			return err
		}
		return me.expr(f(args[0].Val, args[1].Val), nil, nil, args...)
	}
}

func (me *Interp) primFnReplEnv(env *MoEnv, args ...*MoExpr) *MoExpr {
	if err := me.checkCount(0, 0, args); err != nil {
		return me.exprErr(err)
	}
	src_file := me.FauxFile
	if (me.diagCtxCall != nil) && (me.diagCtxCall.SrcFile != nil) {
		src_file = me.diagCtxCall.SrcFile
	}
	src_span := me.diagSpan(false, false, args...)

	ret := me.expr(util.Ptr(make(MoValDict, 0, len(me.Env.Own)+len(env.Own))), src_file, src_span)
	var populate func(it *MoEnv, into *MoExpr) *MoExpr
	populate = func(it *MoEnv, into *MoExpr) *MoExpr {
		dict := into.Val.(*MoValDict)
		for k, v := range it.Own {
			dict.Set(me.expr(k, src_file, src_span), v)
		}
		if it.Parent != nil {
			dict_parent := me.expr(util.Ptr(make(MoValDict, 0, len(env.Parent.Own))), src_file, src_span)
			dict.Set(me.expr(MoValIdent(""), src_file, src_span),
				populate(it.Parent, dict_parent))
		}
		into.Val = dict
		return into
	}
	populate(env, ret)
	return ret
}

func (me *Interp) primFnReplPrint(_ *MoEnv, args ...*MoExpr) *MoExpr {
	if err := me.checkCount(1, 1, args); err != nil {
		return me.exprErr(err)
	}
	if err := me.checkArgErrs(args...); err != nil {
		return err
	}
	switch arg0 := args[0].Val.(type) {
	case MoValStr:
		InterpStdout.WriteString(string(arg0))
	default:
		args[0].WriteTo(InterpStdout)
	}
	return me.exprVoid(args...)
}

func (me *Interp) primFnListLen(_ *MoEnv, args ...*MoExpr) *MoExpr {
	if err := me.checkCount(1, 1, args); err != nil {
		return me.exprErr(err)
	}
	if err := me.checkArgErrs(args...); err != nil {
		return err
	}
	if err := me.checkIs(MoPrimTypeList, args[0]); err != nil {
		return me.exprErr(err)
	}
	return me.expr(MoValNumUint(len(*(args[0].Val.(*MoValList)))), nil, nil, args...)
}

func (me *Interp) primFnListGet(_ *MoEnv, args ...*MoExpr) *MoExpr {
	if err := me.checkCount(2, 2, args); err != nil {
		return me.exprErr(err)
	}
	if err := me.checkArgErrs(args...); err != nil {
		return err
	}
	if err := me.checkIs(MoPrimTypeList, args[0]); err != nil {
		return me.exprErr(err)
	}
	if err := me.checkIs(MoPrimTypeNumUint, args[1]); err != nil {
		return me.exprErr(err)
	}
	list, idx := *(args[0].Val.(*MoValList)), args[1].Val.(MoValNumUint)
	if idx_downcast := int(idx); (idx_downcast < 0) || (idx_downcast >= len(list)) {
		return me.exprErr(me.diagSpan(false, true, args[1]).newDiagErr(ErrCodeIndexOutOfBounds, idx_downcast, len(list)))
	}
	return me.exprFrom(list[idx], args...)
}

func (me *Interp) primFnListSet(_ *MoEnv, args ...*MoExpr) *MoExpr {
	if err := me.checkCount(3, 3, args); err != nil {
		return me.exprErr(err)
	}
	if err := me.checkArgErrs(args...); err != nil {
		return err
	}
	if err := me.checkIs(MoPrimTypeList, args[0]); err != nil {
		return me.exprErr(err)
	}
	if err := me.checkIs(MoPrimTypeNumUint, args[1]); err != nil {
		return me.exprErr(err)
	}
	list, idx := args[0].Val.(*MoValList), args[1].Val.(MoValNumUint)
	if idx_downcast := int(idx); (idx_downcast < 0) || (idx_downcast >= len(*list)) {
		return me.exprErr(me.diagSpan(false, true, args[1]).newDiagErr(ErrCodeIndexOutOfBounds, idx_downcast, len(*list)))
	}
	it := *list
	it[idx] = args[2]
	*list = it
	return me.exprVoid(args...)
}

func (me *Interp) primFnListRange(_ *MoEnv, args ...*MoExpr) *MoExpr {
	if err := me.checkCount(2, 3, args); err != nil {
		return me.exprErr(err)
	}
	if err := me.checkArgErrs(args...); err != nil {
		return err
	}
	if err := me.checkIs(MoPrimTypeList, args[0]); err != nil {
		return me.exprErr(err)
	}
	if err := me.checkIs(MoPrimTypeNumUint, args[1]); err != nil {
		return me.exprErr(err)
	}
	list, idx_start := *(args[0].Val.(*MoValList)), args[1].Val.(MoValNumUint)
	if len(args) == 2 {
		args = append(args, me.expr(MoValNumUint(len(list)), me.srcFile(false, true, args[1]), me.diagSpan(false, true, args[1])))
	} else if err := me.checkIs(MoPrimTypeNumUint, args[2]); err != nil {
		return me.exprErr(err)
	}
	idx_end := args[2].Val.(MoValNumUint)
	if idx_end < idx_start {
		return me.exprErr(me.diagSpan(false, true, args[2]).newDiagErr(ErrCodeRangeNegative, idx_end, idx_start))
	} else if idx_downcast := int(idx_start); (idx_downcast < 0) || (idx_downcast > len(list)) {
		return me.exprErr(me.diagSpan(false, true, args[1]).newDiagErr(ErrCodeIndexOutOfBounds, idx_downcast, len(list)))
	}
	if idx_downcast := int(idx_end); (idx_downcast < 0) || (idx_downcast > len(list)) {
		return me.exprErr(me.diagSpan(false, true, args[2]).newDiagErr(ErrCodeIndexOutOfBounds, idx_downcast, len(list)))
	}
	return me.expr(util.Ptr(list[idx_start:idx_end]), me.srcFile(false, idx_start != idx_end, list[idx_start:idx_end]...), me.diagSpan(false, idx_start != idx_end, list[idx_start:idx_end]...))
}

func (me *Interp) primFnListConcat(_ *MoEnv, args ...*MoExpr) *MoExpr {
	if err := me.checkCount(1, 1, args); err != nil {
		return me.exprErr(err)
	}
	if err := me.checkArgErrs(args...); err != nil {
		return err
	}
	if err := me.checkIs(MoPrimTypeList, args[0]); err != nil {
		return me.exprErr(err)
	}
	list := *(args[0].Val.(*MoValList))
	ret := make(MoValList, 0, len(list)*3)
	for _, arg := range list {
		if err := me.checkIs(MoPrimTypeList, arg); err != nil {
			return me.exprErr(err)
		}
		ret = append(ret, *(arg.Val.(*MoValList))...)
	}
	return me.expr(&ret, nil, nil, args...)
}

func (me *Interp) primFnStrCharAt(_ *MoEnv, args ...*MoExpr) *MoExpr {
	if err := me.checkCount(2, 2, args); err != nil {
		return me.exprErr(err)
	}
	if err := me.checkArgErrs(args...); err != nil {
		return err
	}
	if err := me.checkIs(MoPrimTypeStr, args[0]); err != nil {
		return me.exprErr(err)
	}
	str := args[0].Val.(MoValStr)
	if err := me.checkIs(MoPrimTypeNumUint, args[1]); err != nil {
		return me.exprErr(err)
	}
	idx := int(args[1].Val.(MoValNumUint))
	if (idx < 0) || (idx >= len(str)) {
		return me.exprErr(me.diagSpan(false, true, args[1]).newDiagErr(ErrCodeIndexOutOfBounds, idx, len(str)))
	}
	return me.expr(MoValChar(str[idx]), nil, nil, args...)
}

func (me *Interp) primFnStrRange(_ *MoEnv, args ...*MoExpr) *MoExpr {
	if err := me.checkCount(2, 3, args); err != nil {
		return me.exprErr(err)
	}
	if err := me.checkArgErrs(args...); err != nil {
		return err
	}
	if err := me.checkIs(MoPrimTypeStr, args[0]); err != nil {
		return me.exprErr(err)
	}
	str := args[0].Val.(MoValStr)
	if err := me.checkIs(MoPrimTypeNumUint, args[1]); err != nil {
		return me.exprErr(err)
	}
	idx_start, idx_end := int(args[1].Val.(MoValNumUint)), len(str)
	if (idx_start < 0) || (idx_start > len(str)) {
		return me.exprErr(me.diagSpan(false, true, args[1]).newDiagErr(ErrCodeIndexOutOfBounds, idx_start, len(str)))
	}
	if len(args) > 2 {
		if err := me.checkIs(MoPrimTypeNumUint, args[2]); err != nil {
			return me.exprErr(err)
		}
		idx_end = int(args[2].Val.(MoValNumUint))
	}
	if (idx_end < 0) || (idx_end > len(str)) {
		return me.exprErr(me.diagSpan(false, true, args[1]).newDiagErr(ErrCodeIndexOutOfBounds, idx_end, len(str)))
	} else if idx_end < idx_start {
		return me.exprErr(me.diagSpan(false, true, args[1]).newDiagErr(ErrCodeRangeNegative, idx_end, idx_start))
	}
	return me.expr(MoValStr(str[idx_start:idx_end]), nil, nil, args...)
}

func (me *Interp) primFnStrLen(_ *MoEnv, args ...*MoExpr) *MoExpr {
	if err := me.checkCount(1, 1, args); err != nil {
		return me.exprErr(err)
	}
	if err := me.checkArgErrs(args...); err != nil {
		return err
	}
	if err := me.checkIs(MoPrimTypeStr, args[0]); err != nil {
		return me.exprErr(err)
	}
	return me.expr(MoValNumUint(len(args[0].Val.(MoValStr))), nil, nil, args...)
}

func (me *Interp) primFnStrConcat(_ *MoEnv, args ...*MoExpr) *MoExpr {
	if err := me.checkCount(1, 1, args); err != nil {
		return me.exprErr(err)
	}
	if err := me.checkArgErrs(args...); err != nil {
		return err
	}
	if err := me.checkIs(MoPrimTypeList, args[0]); err != nil {
		return me.exprErr(err)
	}
	var buf strings.Builder
	for _, arg := range *(args[0].Val.(*MoValList)) {
		if err := me.checkIs(MoPrimTypeStr, arg); err != nil {
			return me.exprErr(err)
		}
		buf.WriteString(string(arg.Val.(MoValStr)))
	}
	return me.expr(MoValStr(buf.String()), nil, nil, args...)
}

func (me *Interp) primFnStr(_ *MoEnv, args ...*MoExpr) *MoExpr {
	if err := me.checkCount(1, 1, args); err != nil {
		return me.exprErr(err)
	}
	if err := me.checkArgErrs(args...); err != nil {
		return err
	}
	var str MoValStr
	switch it := args[0].Val.(type) {
	case MoValChar:
		str = MoValStr(string(it))
	case MoValStr:
		str = MoValStr(it)
	case MoValPrimTypeTag:
		str = MoValStr(((MoValPrimType)(it)).Str(true))
	default:
		str = MoValStr(MoValToString(it))
	}
	return me.expr(str, nil, nil, args...)
}

func (me *Interp) primFnExprStr(_ *MoEnv, args ...*MoExpr) *MoExpr {
	if err := me.checkCount(1, 1, args); err != nil {
		return me.exprErr(err)
	}
	if err := me.checkArgErrs(args...); err != nil {
		return err
	}
	return me.expr(MoValStr(args[0].String()), nil, nil, args...)
}

func (me *Interp) primFnExprEval(_ *MoEnv, args ...*MoExpr) *MoExpr {
	if err := me.checkCount(1, 1, args); err != nil {
		return me.exprErr(err)
	}
	if err := me.checkArgErrs(args...); err != nil {
		return err
	}
	return me.exprFrom(me.evalAndApply(me.Env, args[0]), args...)
}

func (me *Interp) primFnExprParse(_ *MoEnv, args ...*MoExpr) *MoExpr {
	if err := me.checkCount(1, 1, args); err != nil {
		return me.exprErr(err)
	}
	if err := me.checkArgErrs(args...); err != nil {
		return err
	}
	if err := me.checkIs(MoPrimTypeStr, args[0]); err != nil {
		return me.exprErr(err)
	}
	ret, err := me.ExprParse(string(args[0].Val.(MoValStr)))
	if err != nil {
		return me.exprErr(err)
	}
	return ret
}

func (me *Interp) primFnDictHas(_ *MoEnv, args ...*MoExpr) *MoExpr {
	if err := me.checkCount(2, 2, args); err != nil {
		return me.exprErr(err)
	}
	if err := me.checkArgErrs(args...); err != nil {
		return err
	}
	if err := me.checkIs(MoPrimTypeDict, args[0]); err != nil {
		return me.exprErr(err)
	}
	return me.exprBool(args[0].Val.(*MoValDict).Has(args[1]), args...)
}

func (me *Interp) primFnDictGet(_ *MoEnv, args ...*MoExpr) *MoExpr {
	if err := me.checkCount(2, 2, args); err != nil {
		return me.exprErr(err)
	}
	if err := me.checkArgErrs(args...); err != nil {
		return err
	}
	if err := me.checkIs(MoPrimTypeDict, args[0]); err != nil {
		return me.exprErr(err)
	}
	found := args[0].Val.(*MoValDict).Get(args[1])
	if found == nil {
		return me.exprVoid(args...)
	}
	return me.exprFrom(found, args...)
}

func (me *Interp) primFnDictSet(_ *MoEnv, args ...*MoExpr) *MoExpr {
	if err := me.checkCount(3, 3, args); err != nil {
		return me.exprErr(err)
	}
	if err := me.checkArgErrs(args...); err != nil {
		return err
	}
	if err := me.checkIs(MoPrimTypeDict, args[0]); err != nil {
		return me.exprErr(err)
	}
	dict := args[0].Val.(*MoValDict)
	dict.Set(args[1], args[2])
	return me.exprVoid(args...)
}

func (me *Interp) primFnDictDel(_ *MoEnv, args ...*MoExpr) *MoExpr {
	if err := me.checkCount(2, 2, args); err != nil {
		return me.exprErr(err)
	}
	if err := me.checkArgErrs(args...); err != nil {
		return err
	}
	if err := me.checkIs(MoPrimTypeDict, args[0]); err != nil {
		return me.exprErr(err)
	}
	dict := args[0].Val.(*MoValDict)
	if err := me.checkIs(MoPrimTypeList, args[1]); err != nil {
		return me.exprErr(err)
	}
	for _, key_to_del := range *(args[1].Val.(*MoValList)) {
		dict.Del(key_to_del)
	}
	return me.exprVoid(args...)
}

func (me *Interp) primFnDictLen(_ *MoEnv, args ...*MoExpr) *MoExpr {
	if err := me.checkCount(1, 1, args); err != nil {
		return me.exprErr(err)
	}
	if err := me.checkArgErrs(args...); err != nil {
		return err
	}
	if err := me.checkIs(MoPrimTypeDict, args[0]); err != nil {
		return me.exprErr(err)
	}
	return me.expr(MoValNumUint(len(*(args[0].Val.(*MoValDict)))), nil, nil, args...)
}

func (me *Interp) primFnPrimTypeTag(_ *MoEnv, args ...*MoExpr) *MoExpr {
	if err := me.checkCount(1, 1, args); err != nil {
		return me.exprErr(err)
	}
	if err := me.checkArgErrs(args...); err != nil {
		return err
	}
	return me.expr(MoValPrimTypeTag(args[0].Val.PrimType()), nil, nil, args...)
}

func (me *Interp) primFnCast(_ *MoEnv, args ...*MoExpr) *MoExpr {
	if err := me.checkCount(2, 2, args); err != nil {
		return me.exprErr(err)
	}
	if err := me.checkArgErrs(args...); err != nil {
		return err
	}
	if err := me.checkIs(MoPrimTypePrimTypeTag, args[0]); err != nil {
		return me.exprErr(err)
	}
	convert_to, convertee := MoValPrimType(args[0].Val.(MoValPrimTypeTag)), args[1]
	if check := me.checkIs(convert_to, convertee); check == nil {
		return me.exprFrom(convertee)
	}
	if err_val, is := convertee.Val.(MoValErr); is && (err_val.ErrVal != nil) {
		convertee = err_val.ErrVal
	}
	var ret MoVal
	switch convert_to {
	case MoPrimTypeCall, MoPrimTypeDict, MoPrimTypeFunc, MoPrimTypeList:
		break
	case MoPrimTypeChar:
		switch it := convertee.Val.(type) {
		case MoValNumUint:
			ret = MoValChar(rune(int32(it)))
		}
	case MoPrimTypeErr:
		ret = MoValErr{ErrVal: convertee}
	case MoPrimTypeIdent:
		switch it := convertee.Val.(type) {
		case MoValStr:
			ret = MoValIdent(it)
		}
	case MoPrimTypeNumFloat:
		switch it := convertee.Val.(type) {
		case MoValNumUint:
			ret = MoValNumFloat(it)
		case MoValNumInt:
			ret = MoValNumFloat(it)
		case MoValStr:
			if f, err := str.ToF(string(it), 64); err == nil {
				ret = MoValNumFloat(f)
			}
		}
	case MoPrimTypeNumInt:
		switch it := convertee.Val.(type) {
		case MoValNumUint:
			ret = MoValNumInt(it)
		case MoValNumFloat:
			ret = MoValNumInt(it)
		case MoValStr:
			if i, err := str.ToI64(string(it), 10, 64); err == nil {
				ret = MoValNumInt(i)
			}
		}
	case MoPrimTypeNumUint:
		switch it := convertee.Val.(type) {
		case MoValNumFloat:
			ret = MoValNumUint(it)
		case MoValNumInt:
			ret = MoValNumUint(it)
		case MoValStr:
			if ui, err := str.ToU64(string(it), 10, 64); err == nil {
				ret = MoValNumUint(ui)
			}
		}
	case MoPrimTypeStr:
		ret = MoValStr(convertee.String())
	default:
		return me.exprErr(me.diagSpan(false, true, args...).newDiagErr(ErrCodeAtmoTodo, "primFnCast: unhandled prim-type-tag "+convert_to.Str(false)))
	}
	if ret != nil {
		return me.expr(ret, nil, nil, args[1])
	}
	return me.exprErr(me.diagSpan(false, false, args...).newDiagErr(ErrCodeNotConvertible, convertee.String(), convert_to.Str(true)))
}

func (me *Interp) primFnErrNew(_ *MoEnv, args ...*MoExpr) *MoExpr {
	if err := me.checkCount(1, 1, args); err != nil {
		return me.exprErr(err)
	}
	if err := me.checkArgErrs(args...); err != nil {
		return err
	}
	return me.expr(MoValErr{ErrVal: args[0]}, nil, nil, args...)
}

func (me *Interp) primFnErrVal(_ *MoEnv, args ...*MoExpr) *MoExpr {
	if err := me.checkCount(1, 1, args); err != nil {
		return me.exprErr(err)
	}
	if err := me.checkIs(MoPrimTypeErr, args[0]); err != nil {
		return me.exprErr(err, args[0])
	}
	err_val := args[0].Val.(MoValErr).ErrVal
	return me.exprFrom(err_val, args...)
}

func (me *Interp) primFnReplReset(_ *MoEnv, args ...*MoExpr) *MoExpr {
	me.replReset()
	return me.exprVoid(args...)
}

func (me *Interp) primFnEq(_ *MoEnv, args ...*MoExpr) *MoExpr {
	if err := me.checkCount(2, 2, args); err != nil {
		return me.exprErr(err)
	}
	if err := me.checkArgErrs(args...); err != nil {
		return err
	}
	if !me.checkNoneArePrimFuncs(args...) {
		return me.exprErr(me.diagSpan(false, true, args...).newDiagErr(ErrCodeNotComparable, args[0], args[1], "equality"))
	}
	return me.exprBool(args[0].Eq(args[1]), args...)
}

func (me *Interp) primFnNeq(env *MoEnv, args ...*MoExpr) *MoExpr {
	if err := me.checkCount(2, 2, args); err != nil {
		return me.exprErr(err)
	}
	if err := me.checkArgErrs(args...); err != nil {
		return err
	}
	if !me.checkNoneArePrimFuncs(args...) {
		return me.exprErr(me.diagSpan(false, true, args...).newDiagErr(ErrCodeNotComparable, args[0], args[1], "not-equal"))
	}
	return me.exprBool(!args[0].Eq(args[1]), args...)
}

func (me *Interp) primFnLeq(_ *MoEnv, args ...*MoExpr) *MoExpr {
	cmp, err := me.primCmpHelper("less-or-equal", args...)
	if err != nil {
		return me.exprErr(err)
	}
	return me.exprBool(cmp <= 0, args...)
}

func (me *Interp) primFnGeq(_ *MoEnv, args ...*MoExpr) *MoExpr {
	cmp, err := me.primCmpHelper("greater-or-equal", args...)
	if err != nil {
		return me.exprErr(err)
	}
	return me.exprBool(cmp >= 0, args...)
}

func (me *Interp) primFnLt(_ *MoEnv, args ...*MoExpr) *MoExpr {
	cmp, err := me.primCmpHelper("less-than", args...)
	if err != nil {
		return me.exprErr(err)
	}
	return me.exprBool(cmp < 0, args...)
}

func (me *Interp) primFnGt(_ *MoEnv, args ...*MoExpr) *MoExpr {
	cmp, err := me.primCmpHelper("greater-than", args...)
	if err != nil {
		return me.exprErr(err)
	}
	return me.exprBool(cmp > 0, args...)
}
func (me *Interp) primCmpHelper(diagMoniker string, args ...*MoExpr) (int, *Diag) {
	if err := me.checkCount(2, 2, args); err != nil {
		return 0, err
	}
	for _, arg := range args {
		if err := arg.Err(); err != nil {
			return 0, err
		}
	}
	if !me.checkNoneArePrimFuncs(args...) {
		return 0, me.diagSpan(false, true, args...).newDiagErr(ErrCodeNotComparable, args[0], args[1], diagMoniker)
	}
	cmp, err := me.ExprCmp(args[0], args[1], diagMoniker)
	if err != nil {
		return 0, err
	}
	return cmp, nil
}
