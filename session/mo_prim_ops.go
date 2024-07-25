package session

import (
	"fmt"
	"os"

	"atmo/util"
)

var (
	moPrimOpsLazy = map[moValIdent]moFnLazy{ // "lazy" prim-ops take unevaluated args to eval-or-not as needed. eg. `@match`, `@fn` etc
		// populated in `init()` below to avoid initialization-cycle error
	}
	moPrimOpsEager = map[moValIdent]moFnEager{ // "eager" prim-ops receive already-evaluated args like any other func. eg. prim-type intrinsics like arithmetics, list concat etc
		"@numIntAdd":   makeArithPrimOp[moValInt](MoPrimTypeInt, func(opl MoVal, opr MoVal) MoVal { return opl.(moValInt) + opr.(moValInt) }),
		"@numIntSub":   makeArithPrimOp[moValInt](MoPrimTypeInt, func(opl MoVal, opr MoVal) MoVal { return opl.(moValInt) - opr.(moValInt) }),
		"@numIntMul":   makeArithPrimOp[moValInt](MoPrimTypeInt, func(opl MoVal, opr MoVal) MoVal { return opl.(moValInt) * opr.(moValInt) }),
		"@numIntDiv":   makeArithPrimOp[moValInt](MoPrimTypeInt, func(opl MoVal, opr MoVal) MoVal { return opl.(moValInt) / opr.(moValInt) }),
		"@numIntMod":   makeArithPrimOp[moValInt](MoPrimTypeInt, func(opl MoVal, opr MoVal) MoVal { return opl.(moValInt) % opr.(moValInt) }),
		"@numUintAdd":  makeArithPrimOp[moValUint](MoPrimTypeUint, func(opl MoVal, opr MoVal) MoVal { return opl.(moValUint) + opr.(moValUint) }),
		"@numUintSub":  makeArithPrimOp[moValUint](MoPrimTypeUint, func(opl MoVal, opr MoVal) MoVal { return opl.(moValUint) - opr.(moValUint) }),
		"@numUintMul":  makeArithPrimOp[moValUint](MoPrimTypeUint, func(opl MoVal, opr MoVal) MoVal { return opl.(moValUint) * opr.(moValUint) }),
		"@numUintDiv":  makeArithPrimOp[moValUint](MoPrimTypeUint, func(opl MoVal, opr MoVal) MoVal { return opl.(moValUint) / opr.(moValUint) }),
		"@numUintMod":  makeArithPrimOp[moValUint](MoPrimTypeUint, func(opl MoVal, opr MoVal) MoVal { return opl.(moValUint) % opr.(moValUint) }),
		"@numFloatAdd": makeArithPrimOp[moValFloat](MoPrimTypeFloat, func(opl MoVal, opr MoVal) MoVal { return opl.(moValFloat) + opr.(moValFloat) }),
		"@numFloatSub": makeArithPrimOp[moValFloat](MoPrimTypeFloat, func(opl MoVal, opr MoVal) MoVal { return opl.(moValFloat) - opr.(moValFloat) }),
		"@numFloatMul": makeArithPrimOp[moValFloat](MoPrimTypeFloat, func(opl MoVal, opr MoVal) MoVal { return opl.(moValFloat) * opr.(moValFloat) }),
		"@numFloatDiv": makeArithPrimOp[moValFloat](MoPrimTypeFloat, func(opl MoVal, opr MoVal) MoVal { return opl.(moValFloat) / opr.(moValFloat) }),
		"@fnCall":      (*Interp).primFnFuncCall,
		"@sessEnv":     (*Interp).primFnSessEnv,
		"@sessPrintf":  (*Interp).primFnSessPrintf,
		"@sessPrint":   (*Interp).primFnSessPrint,
		"@sessPrintln": (*Interp).primFnSessPrintln,
		"@listItemAt":  (*Interp).primFnListItemAt,
		"@listRange":   (*Interp).primFnListRange,
		"@exprStr":     (*Interp).primFnExprStr,
		"@exprParse":   (*Interp).primFnExprParse,
		"@exprEval":    (*Interp).primFnExprEval,
		"@str":         (*Interp).primFnStr,
	}
)

const (
	moPrimOpQuote         moValIdent = "#"
	moPrimOpQQuote        moValIdent = "%"
	moPrimOpUnquote       moValIdent = "$"
	moPrimOpSpliceUnquote moValIdent = "$$"
	moPrimOpDo            moValIdent = "@do"
)

func init() {
	for k, v := range map[moValIdent]moFnLazy{
		"@set":         (*Interp).primOpSet,
		"@fn":          (*Interp).primOpFn,
		"@caseOf":      (*Interp).primOpCaseOf,
		"@macro":       (*Interp).primOpMacro,
		"@expand":      (*Interp).primOpMacroExpand,
		moPrimOpDo:     (*Interp).primOpDo,
		moPrimOpQuote:  (*Interp).primOpQuote,
		moPrimOpQQuote: (*Interp).primOpQuasiQuote,
	} {
		moPrimOpsLazy[k] = v
	}

	rootEnv.populateWithPrims()
}

// lazy prim-ops first, eager prim-ops afterwards

func (me *Interp) primOpSet(env *MoEnv, args ...*MoExpr) (*MoEnv, *MoExpr, *SrcFileNotice) {
	if err := me.checkCount(2, 2, args); err != nil {
		return nil, nil, err
	}
	if err := me.checkIs(MoPrimTypeIdent, args[0]); err != nil {
		return nil, nil, err
	}
	name := args[0].Val.(moValIdent)
	if is_reserved := ((name[0] == '@') || (name[0] == moPrimOpUnquote[0]) || moPrimOpsLazy[name] != nil); is_reserved {
		return nil, nil, me.diagSpan(false, true, args[0]).newDiagErr(NoticeCodeReserved, name, string(rune(name[0])))
	}
	owner_env, found := env.lookupOwner(name)
	if owner_env == nil {
		owner_env = env
	}

	const can_set_macros = false
	if (!can_set_macros) && (found != nil) {
		if fn, _ := found.Val.(*moValFnLam); (fn != nil) && fn.isMacro {
			return nil, nil, me.diagSpan(true, false, args...).newDiagErr(NoticeCodeAtmoTodo, "mutating macros currently disabled, let us know whether you disagree with that or not")
		}
	}
	new_value, err := me.evalAndApply(env, args[1])
	if err != nil {
		return nil, nil, err
	}
	owner_env.set(name, new_value)
	return nil, moValNone, nil
}

func (me *Interp) primOpDo(env *MoEnv, args ...*MoExpr) (tailEnv *MoEnv, expr *MoExpr, err *SrcFileNotice) {
	if err = me.checkCount(1, -1, args); err != nil {
		return
	}
	for _, arg := range args[:len(args)-1] {
		if expr, err = me.evalAndApply(env, arg); err != nil {
			return
		}
	}
	tailEnv, expr = env, args[len(args)-1]
	return
}

func (me *Interp) primOpMacro(env *MoEnv, args ...*MoExpr) (*MoEnv, *MoExpr, *SrcFileNotice) {
	_, expr, err := me.primOpFn(env, args...)
	if err != nil {
		return nil, nil, err
	}
	expr.Val.(*moValFnLam).isMacro = true
	return nil, expr, nil
}

func (me *Interp) primOpFn(env *MoEnv, args ...*MoExpr) (*MoEnv, *MoExpr, *SrcFileNotice) {
	if err := me.checkCount(2, -1, args); err != nil {
		return nil, nil, err
	}
	if err := me.checkIsListOf(MoPrimTypeIdent, args[0]); err != nil {
		return nil, nil, err
	}
	body := args[1]
	if len(args) > 2 {
		do := &MoExpr{SrcSpan: srcSpanFrom(args[1:]), Val: moPrimOpDo}
		body = &MoExpr{
			SrcSpan: do.SrcSpan,
			Val:     append(moValCall{do}, args[1:]...),
		}
	}
	expr := &MoExpr{
		SrcSpan: srcSpanFrom(args),
		Val:     &moValFnLam{params: args[0].Val.(moValList), body: body, env: env},
	}
	return nil, expr, nil
}

func (me *Interp) primOpQuote(_ *MoEnv, args ...*MoExpr) (*MoEnv, *MoExpr, *SrcFileNotice) {
	if err := me.checkCount(1, 1, args); err != nil {
		return nil, nil, err
	}
	return nil, args[0], nil
}

func (me *Interp) primOpQuasiQuote(env *MoEnv, args ...*MoExpr) (*MoEnv, *MoExpr, *SrcFileNotice) {
	if err := me.checkCount(1, 1, args); err != nil {
		return nil, nil, err
	}

	// is atomic arg? then just like primOpQuote
	if args[0].Val.primType().isAtomic() {
		return nil, args[0], nil
	}

	// is this call directly quoting an unquote call?
	if is_unquote, err := me.checkIsCallOnIdent(args[0], moPrimOpUnquote, 1); err != nil {
		return nil, nil, err
	} else if is_unquote {
		ret, err := me.evalAndApply(env, args[0].Val.(moValCall)[1])
		if err != nil {
			return nil, nil, err
		}
		return nil, ret, err
	}

	// hash-table? call ourselves on each key and value
	if rec, is := args[0].Val.(moValRec); is {
		ret := make(moValRec, len(rec))
		for k, v := range rec {
			_, key, err := me.primOpQuasiQuote(env, k)
			if err != nil {
				return nil, nil, err
			}
			_, val, err := me.primOpQuasiQuote(env, v)
			if err != nil {
				return nil, nil, err
			}
			ret[key] = val
		}
		return nil, &MoExpr{Val: ret, SrcSpan: args[0].SrcSpan}, nil
	}

	// must be list or call then: we handle them the same, per item iteration

	is_list := (args[0].Val.primType() == MoPrimTypeList)
	var call_or_arr []*MoExpr
	if call, is := args[0].Val.(moValCall); is {
		call_or_arr = call
	} else if is_list {
		call_or_arr = args[0].Val.(moValList)
	} else {
		return nil, nil, args[0].SrcSpan.newDiagErr(NoticeCodeAtmoTodo, "NEW BUG intro'd in primOpQuasiQuote")
	}

	ret := make([]*MoExpr, 0, len(call_or_arr))
	for _, item := range call_or_arr {
		if is_unquote, err := me.checkIsCallOnIdent(item, moPrimOpUnquote, 1); err != nil {
			return nil, nil, err
		} else if is_unquote {
			if evaled, err := me.evalAndApply(env, item.Val.(moValCall)[1]); err != nil {
				return nil, nil, err
			} else {
				ret = append(ret, evaled)
			}
		} else if is_splice_unquote, err := me.checkIsCallOnIdent(item, moPrimOpSpliceUnquote, 1); err != nil {
			return nil, nil, err
		} else if is_splice_unquote {
			evaled, err := me.evalAndApply(env, item.Val.(moValCall)[1])
			if err != nil {
				return nil, nil, err
			}
			if err = me.checkIsListOf(-1, evaled); err != nil {
				return nil, nil, err
			}
			for _, splicee := range evaled.Val.(moValList) {
				if evaled, err := me.evalAndApply(env, splicee); err != nil {
					return nil, nil, err
				} else {
					ret = append(ret, evaled)
				}
			}
		} else {
			_, evaled, err := me.primOpQuasiQuote(env, item)
			if err != nil {
				return nil, nil, err
			}
			ret = append(ret, evaled)
		}
	}
	return nil, &MoExpr{Val: util.If[MoVal](is_list, moValList(ret), moValCall(ret)), SrcSpan: args[0].SrcSpan}, nil
}

func (me *Interp) primOpMacroExpand(env *MoEnv, args ...*MoExpr) (*MoEnv, *MoExpr, *SrcFileNotice) {
	if err := me.checkCount(1, 1, args); err != nil {
		return nil, nil, err
	}
	if err := me.checkIs(MoPrimTypeCall, args[0]); err != nil {
		return nil, nil, err
	}
	ret, err := me.macroExpand(env, args[0])
	return nil, ret, err
}

func (me *Interp) primOpCaseOf(env *MoEnv, args ...*MoExpr) (*MoEnv, *MoExpr, *SrcFileNotice) {
	if err := me.check(MoPrimTypeList, 1, 1, args...); err != nil {
		return nil, nil, err
	}
	pairs := args[0].Val.(moValList)
	for _, pair := range pairs {
		if err := me.checkIs(MoPrimTypeList, pair); err != nil {
			return nil, nil, err
		}
		if err := me.checkCountWithSrcSpan(2, 2, pair.Val.(moValList), true); err != nil {
			return nil, nil, err
		}
		pred := pair.Val.(moValList)[0]
		evaled, err := me.evalAndApply(env, pred)
		if err != nil {
			return nil, nil, err
		}
		if evaled.eqTrue() {
			return env, pair.Val.(moValList)[1], nil
		} else if !evaled.isFalsy() {
			return nil, nil, me.diagSpan(false, true, pred).newDiagErr(NoticeCodeExpectedFoo, "a boolean expression")
		}
	}
	return nil, nil, me.diagSpan(true, false, args...).newDiagErr(NoticeCodeNoElseCase)
}

// eager prim-ops below, lazy ones above

func makeArithPrimOp[T moValInt | moValUint | moValFloat](t MoValPrimType, f func(opl MoVal, opr MoVal) MoVal) moFnEager {
	return func(me *Interp, _ *MoEnv, args ...*MoExpr) (*MoExpr, *SrcFileNotice) {
		if err := me.check(t, 2, 2, args...); err != nil {
			return nil, err
		}
		return &MoExpr{Val: f(args[0].Val, args[1].Val)}, nil
	}
}

func (me *Interp) primFnSessEnv(env *MoEnv, args ...*MoExpr) (*MoExpr, *SrcFileNotice) {
	if err := me.checkCount(0, 0, args); err != nil {
		return nil, err
	}
	ret := &MoExpr{Val: make(moValRec, len(me.Env.Own)+len(env.Own))}
	var populate func(it *MoEnv, into *MoExpr) *MoExpr
	populate = func(it *MoEnv, into *MoExpr) *MoExpr {
		for k, v := range it.Own {
			into.Val.(moValRec)[&MoExpr{Val: k}] = v
		}
		if it.Outer != nil {
			rec := &MoExpr{Val: make(moValRec, len(env.Outer.Own))}
			into.Val.(moValRec)[&MoExpr{Val: moValIdent("")}] = populate(it.Outer, rec)
		}
		return into
	}
	populate(env, ret)
	return ret, nil
}

func (me *Interp) primFnFuncCall(env *MoEnv, args ...*MoExpr) (*MoExpr, *SrcFileNotice) {
	if err := me.checkCount(2, 2, args); err != nil {
		return nil, err
	}
	if err := me.checkIs(MoPrimTypeFunc, args[0]); err != nil {
		return nil, err
	}
	if err := me.checkIs(MoPrimTypeList, args[1]); err != nil {
		return nil, err
	}
	return me.callWithDiagCtxSet(env, args[0], args[1].Val.(moValList)...)
}

func (me *Interp) primFnSessPrintf(_ *MoEnv, args ...*MoExpr) (*MoExpr, *SrcFileNotice) {
	if err := me.checkCount(1, -1, args); err != nil {
		return nil, err
	}
	if err := me.checkIs(MoPrimTypeStr, args[0]); err != nil {
		return nil, err
	}
	fmt_args := make([]any, 0, len(args)-1)
	for _, arg := range args[1:] {
		fmt_args = append(fmt_args, arg.Val)
	}
	fmt.Fprintf(os.Stdout, string(args[0].Val.(moValStr)), fmt_args...)
	return moValNone, nil
}

func (me *Interp) primFnSessPrint(_ *MoEnv, args ...*MoExpr) (*MoExpr, *SrcFileNotice) {
	if len(args) > 0 {
		switch arg0 := args[0].Val; true {
		case (arg0.primType() == MoPrimTypeStr) && (len(args) == 1):
			os.Stdout.WriteString(string(arg0.(moValStr)))
		default:
			for i, arg := range args {
				if i > 0 {
					os.Stdout.WriteString(" ")
				}
				arg.WriteTo(os.Stdout)
			}
		}
	}
	return moValNone, nil
}

func (me *Interp) primFnSessPrintln(env *MoEnv, args ...*MoExpr) (*MoExpr, *SrcFileNotice) {
	expr, err := me.primFnSessPrint(env, args...)
	if err == nil {
		os.Stdout.WriteString("\n")
	}
	return expr, err
}

func (me *Interp) primFnListItemAt(env *MoEnv, args ...*MoExpr) (*MoExpr, *SrcFileNotice) {
	if err := me.checkCount(2, 2, args); err != nil {
		return nil, err
	}
	if err := me.checkIs(MoPrimTypeList, args[0]); err != nil {
		return nil, err
	}
	if err := me.checkIs(MoPrimTypeUint, args[1]); err != nil {
		return nil, err
	}
	list, idx := args[0].Val.(moValList), args[1].Val.(moValUint)
	if idx_downcast := int(idx); (idx_downcast < 0) || (idx_downcast >= len(list)) {
		return nil, me.diagSpan(false, true, args[1]).newDiagErr(NoticeCodeListIndexOutOfBounds, idx_downcast, len(list))
	}
	return list[idx], nil
}

func (me *Interp) primFnListRange(env *MoEnv, args ...*MoExpr) (*MoExpr, *SrcFileNotice) {
	if err := me.checkCount(2, 3, args); err != nil {
		return nil, err
	}
	if err := me.checkIs(MoPrimTypeList, args[0]); err != nil {
		return nil, err
	}
	if err := me.checkIs(MoPrimTypeUint, args[1]); err != nil {
		return nil, err
	}
	list, idx_start := args[0].Val.(moValList), args[1].Val.(moValUint)
	if len(args) == 2 {
		args = append(args, &MoExpr{Val: moValUint(len(list)), SrcSpan: srcSpanFrom(args)})
	} else if err := me.checkIs(MoPrimTypeUint, args[2]); err != nil {
		return nil, err
	}
	idx_end := args[2].Val.(moValUint)
	if idx_end < idx_start {
		return nil, me.diagSpan(false, true, args[2]).newDiagErr(NoticeCodeListRangeNegative, idx_end, idx_start)
	} else if idx_downcast := int(idx_start); (idx_downcast < 0) || (idx_downcast > len(list)) {
		return nil, me.diagSpan(false, true, args[1]).newDiagErr(NoticeCodeListIndexOutOfBounds, idx_downcast, len(list))
	}
	if idx_downcast := int(idx_end); (idx_downcast < 0) || (idx_downcast > len(list)) {
		return nil, me.diagSpan(false, true, args[2]).newDiagErr(NoticeCodeListIndexOutOfBounds, idx_downcast, len(list))
	}

	return &MoExpr{Val: list[idx_start:idx_end], SrcSpan: srcSpanFrom(util.If(idx_start == idx_end, args, list[idx_start:idx_end]))}, nil
}

func (me *Interp) primFnStr(env *MoEnv, args ...*MoExpr) (*MoExpr, *SrcFileNotice) {
	if err := me.checkCount(1, 1, args); err != nil {
		return nil, err
	}
	var str moValStr
	switch it := args[0].Val.(type) {
	case moValChar:
		str = moValStr(string(it))
	case moValStr:
		str = moValStr(it)
	case moValType:
		str = moValStr(((MoValPrimType)(it)).Str(true))
	default:
		str = moValStr(it.String())
	}
	return &MoExpr{SrcSpan: srcSpanFrom(args), Val: str}, nil
}

func (me *Interp) primFnExprStr(env *MoEnv, args ...*MoExpr) (*MoExpr, *SrcFileNotice) {
	if err := me.checkCount(1, 1, args); err != nil {
		return nil, err
	}
	return &MoExpr{SrcSpan: srcSpanFrom(args), Val: moValStr(args[0].String())}, nil
}

func (me *Interp) primFnExprEval(env *MoEnv, args ...*MoExpr) (*MoExpr, *SrcFileNotice) {
	if err := me.checkCount(1, 1, args); err != nil {
		return nil, err
	}
	return me.evalAndApply(env, args[0])
}

func (me *Interp) primFnExprParse(env *MoEnv, args ...*MoExpr) (*MoExpr, *SrcFileNotice) {
	if err := me.checkCount(1, 1, args); err != nil {
		return nil, err
	}
	if err := me.checkIs(MoPrimTypeStr, args[0]); err != nil {
		return nil, err
	}
	return me.Parse(string(args[0].Val.(moValStr)))
}
