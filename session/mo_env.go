package session

import (
	"atmo/util"
)

var (
	moStdLazy  = map[moValIdent]moFnLazy{}
	moStdEager = map[moValIdent]moFnEager{}
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

func (me *moValFunc) envWith(args []*MoExpr, ctxExpr *MoExpr) (*MoEnv, *SrcFileNotice) {
	num_args_min, num_args_max := len(me.params), len(me.params)
	if err := checkCount(num_args_min, num_args_max, args, ctxExpr); err != nil {
		return nil, err
	}
	return newMoEnv(me.env, me.params, args), nil
}
