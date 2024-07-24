package session

import (
	"atmo/util"
	"fmt"
	"os"
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
		"@call":        (*Interp).primFnCall,
		"@env":         (*Interp).primFnEnv,
		"@printf":      (*Interp).primFnPrintf,
		"@print":       (*Interp).primFnPrint,
		"@println":     (*Interp).primFnPrintln,
	}
)

func init() {
	for k, v := range map[moValIdent]moFnLazy{
		"@set":   (*Interp).primOpSet,
		"@quote": (*Interp).primOpQuote,
		"#":      (*Interp).primOpQuote,
	} {
		moPrimOpsLazy[k] = v
	}
}

type MoEnv struct {
	Outer *MoEnv
	Own   map[moValIdent]*MoExpr
}

func newMoEnv(outer *MoEnv, names []*MoExpr, values []*MoExpr) *MoEnv {
	util.Assert(len(names) == len(values), "newMoEnv: len(names) != len(values)")
	ret := MoEnv{Outer: outer, Own: map[moValIdent]*MoExpr{}}
	for i, name := range names {
		ret.Own[name.Val.(moValIdent)] = values[i]
	}
	return &ret
}

func (me *MoEnv) hasOwn(name moValIdent) (ret bool) {
	_, ret = me.Own[name]
	return
}

func (me *MoEnv) set(name moValIdent, value *MoExpr) {
	util.Assert(value != nil, "MoEnv.set(name, nil)")
	me.Own[name] = value
}

func (me *MoEnv) lookup(name moValIdent) *MoExpr {
	_, found := me.lookupOwner(name)
	return found
}

func (me *MoEnv) lookupOwner(name moValIdent) (*MoEnv, *MoExpr) {
	found := me.Own[name]
	if found == nil {
		if me.Outer != nil {
			return me.Outer.lookupOwner(name)
		} else {
			return nil, nil
		}
	}
	return me, found
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
	if is_reserved := ((name[0] == '@') || moPrimOpsLazy[name] != nil); is_reserved {
		return nil, nil, me.diagNode(false, true, args[0]).newDiagErr(false, NoticeCodeReserved, name, string(rune(name[0])))
	}
	owner_env, _ := env.lookupOwner(name)
	if owner_env == nil {
		owner_env = env
	}
	new_value, err := me.evalAndApply(env, args[1])
	if err != nil {
		return nil, nil, err
	}
	owner_env.set(name, new_value)
	return nil, moValNone, nil
}

func (me *Interp) primOpQuote(_ *MoEnv, args ...*MoExpr) (*MoEnv, *MoExpr, *SrcFileNotice) {
	if err := me.checkCount(1, 1, args); err != nil {
		return nil, nil, err
	}
	return nil, args[0], nil
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

func (me *Interp) primFnEnv(env *MoEnv, args ...*MoExpr) (*MoExpr, *SrcFileNotice) {
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

func (me *Interp) primFnCall(env *MoEnv, args ...*MoExpr) (*MoExpr, *SrcFileNotice) {
	if err := me.checkCount(1, -1, args); err != nil {
		return nil, err
	}
	if err := me.checkIs(MoPrimTypeFunc, args[0]); err != nil {
		return nil, err
	}
	return me.callWithDiagCtxSet(env, args[0], args[1:]...)
}

func (me *Interp) primFnPrintf(_ *MoEnv, args ...*MoExpr) (*MoExpr, *SrcFileNotice) {
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

func (me *Interp) primFnPrint(_ *MoEnv, args ...*MoExpr) (*MoExpr, *SrcFileNotice) {
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

func (me *Interp) primFnPrintln(env *MoEnv, args ...*MoExpr) (*MoExpr, *SrcFileNotice) {
	expr, err := me.primFnPrint(env, args...)
	if err == nil {
		os.Stdout.WriteString("\n")
	}
	return expr, err
}
