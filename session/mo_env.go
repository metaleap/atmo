package session

import (
	"atmo/util"
	"atmo/util/kv"
)

var rootEnv = MoEnv{Own: map[moValIdent]*MoExpr{}}

type MoEnv struct {
	Parent *MoEnv
	Own    map[moValIdent]*MoExpr
}

func (me *Interp) ensureRootEnvPopulated() {
	if len(rootEnv.Own) > 0 {
		return
	}
	// prim idents (@true, @false, @nil) into rootEnv
	for name, expr := range moPrimIdents {
		rootEnv.set(name, expr)
	}
	// builtin eager prim-op funcs into rootEnv
	for name, fn := range moPrimOpsEager {
		rootEnv.set(name, me.expr(moValFnPrim(fn), nil, nil))
	}
}

func newMoEnv(parent *MoEnv, names MoExprs, values MoExprs) *MoEnv {
	util.Assert(len(names) == len(values), "newMoEnv: len(names) != len(values)")
	ret := MoEnv{Parent: parent, Own: map[moValIdent]*MoExpr{}}
	for i, name := range names {
		ret.Own[name.Val.(moValIdent)] = values[i]
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
	return me.Parent.eq(to.Parent) && kv.Eq(me.Own, to.Own, (*MoExpr).eq)
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
		if me.Parent != nil {
			return me.Parent.lookupOwner(name)
		} else {
			return nil, nil
		}
	}
	return me, found
}
