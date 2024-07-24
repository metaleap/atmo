package session

import (
	"atmo/util"
)

var (
	moPrimOpsLazy = map[moValIdent]moFnLazy{ // "lazy" prim-ops take unevaluated args to eval-or-not as needed. eg. `@match`, `@fn` etc
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
	}
)

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
	found := me.Own[name]
	if (found == nil) && (me.Outer != nil) {
		return me.Outer.lookup(name)
	}
	return found
}

// lazy prim-ops first, eager prim-ops afterwards

// eager prim-ops below, lazy ones above

func makeArithPrimOp[T moValInt | moValUint | moValFloat](t MoValType, f func(opl MoVal, opr MoVal) MoVal) moFnEager {
	return func(me *Interp, args ...*MoExpr) (*MoExpr, *SrcFileNotice) {
		if err := me.checkAre(t, 2, 2, args...); err != nil {
			return nil, err
		}
		return &MoExpr{Val: f(args[0].Val, args[1].Val)}, nil
	}
}
