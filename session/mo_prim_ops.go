package session

import (
	"atmo/util"
)

var (
	moPrimOpsLazy = map[moValIdent]moFnLazy{ // "lazy" prim-ops take unevaluated args to eval-or-not as needed. eg. `@match`, `@fn` etc
		// populated in `init()` below to avoid initialization-cycle error
	}
	moPrimOpsEager = map[moValIdent]moFnEager{ // "eager" prim-ops receive already-evaluated args like any other func. eg. prim-type intrinsics like arithmetics, list concat etc
		"@numIntAdd":   makeArithPrimOp[moValInt](MoValTypeInt, func(opl MoVal, opr MoVal) MoVal { return opl.(moValInt) + opr.(moValInt) }),
		"@numIntSub":   makeArithPrimOp[moValInt](MoValTypeInt, func(opl MoVal, opr MoVal) MoVal { return opl.(moValInt) - opr.(moValInt) }),
		"@numIntMul":   makeArithPrimOp[moValInt](MoValTypeInt, func(opl MoVal, opr MoVal) MoVal { return opl.(moValInt) * opr.(moValInt) }),
		"@numIntDiv":   makeArithPrimOp[moValInt](MoValTypeInt, func(opl MoVal, opr MoVal) MoVal { return opl.(moValInt) / opr.(moValInt) }),
		"@numIntMod":   makeArithPrimOp[moValInt](MoValTypeInt, func(opl MoVal, opr MoVal) MoVal { return opl.(moValInt) % opr.(moValInt) }),
		"@numUintAdd":  makeArithPrimOp[moValUint](MoValTypeUint, func(opl MoVal, opr MoVal) MoVal { return opl.(moValUint) + opr.(moValUint) }),
		"@numUintSub":  makeArithPrimOp[moValUint](MoValTypeUint, func(opl MoVal, opr MoVal) MoVal { return opl.(moValUint) - opr.(moValUint) }),
		"@numUintMul":  makeArithPrimOp[moValUint](MoValTypeUint, func(opl MoVal, opr MoVal) MoVal { return opl.(moValUint) * opr.(moValUint) }),
		"@numUintDiv":  makeArithPrimOp[moValUint](MoValTypeUint, func(opl MoVal, opr MoVal) MoVal { return opl.(moValUint) / opr.(moValUint) }),
		"@numUintMod":  makeArithPrimOp[moValUint](MoValTypeUint, func(opl MoVal, opr MoVal) MoVal { return opl.(moValUint) % opr.(moValUint) }),
		"@numFloatAdd": makeArithPrimOp[moValFloat](MoValTypeFloat, func(opl MoVal, opr MoVal) MoVal { return opl.(moValFloat) + opr.(moValFloat) }),
		"@numFloatSub": makeArithPrimOp[moValFloat](MoValTypeFloat, func(opl MoVal, opr MoVal) MoVal { return opl.(moValFloat) - opr.(moValFloat) }),
		"@numFloatMul": makeArithPrimOp[moValFloat](MoValTypeFloat, func(opl MoVal, opr MoVal) MoVal { return opl.(moValFloat) * opr.(moValFloat) }),
		"@numFloatDiv": makeArithPrimOp[moValFloat](MoValTypeFloat, func(opl MoVal, opr MoVal) MoVal { return opl.(moValFloat) / opr.(moValFloat) }),
		"@call":        (*Interp).primFnCall,
		"@env":         (*Interp).primFnEnv,
	}
)

func init() {
	for k, v := range map[moValIdent]moFnLazy{
		"@set": (*Interp).primOpSet,
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
	if err := me.checkIs(MoValTypeIdent, args[0]); err != nil {
		return nil, nil, err
	}
	name := args[0].Val.(moValIdent)
	if is_reserved := (name[0] == '@'); is_reserved {
		return nil, nil, me.diagNode(false, true, args[0]).newDiagErr(false, NoticeCodeReserved, name, "@")
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

// eager prim-ops below, lazy ones above

func makeArithPrimOp[T moValInt | moValUint | moValFloat](t MoValType, f func(opl MoVal, opr MoVal) MoVal) moFnEager {
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
	// if ident,_ := args[0].Val.(moValIdent);ident!="" {

	// }
	if err := me.checkIs(MoValTypeFunc, args[0]); err != nil {
		return nil, err
	}
	return me.callWithDiagCtxSet(env, args[0], args[1:]...)
}
