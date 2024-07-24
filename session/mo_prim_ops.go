package session

import (
	"atmo/util"
)

var (
	moPrimOpsLazy  = map[moValIdent]moFnLazy{}
	moPrimOpsEager = map[moValIdent]moFnEager{}
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

func (me *DefaultEvaler) envWith(fn *moValFnLam, args []*MoExpr) (*MoEnv, *SrcFileNotice) {
	if err := me.checkCount(len(fn.params), len(fn.params), args); err != nil {
		return nil, err
	}
	return newMoEnv(fn.env, fn.params, args), nil
}

func primNumAddInt(...*MoExpr) (*MoExpr, *SrcFileNotice) {
	return nil, nil
}
