package session

import (
	"atmo/util"
)

type AtEnv struct {
	Outer *AtEnv
	Own   map[atValIdent]*AtExpr
}

func newAtEnv(outer *AtEnv, names []*AtExpr, values []*AtExpr) *AtEnv {
	util.Assert(len(names) == len(values), "newAtEnv: len(names) != len(values)")
	ret := AtEnv{Outer: outer, Own: map[atValIdent]*AtExpr{}}
	for i, name := range names {
		ret.Own[name.Val.(atValIdent)] = values[i]
	}
	return &ret
}

func (me *AtEnv) hasOwn(name atValIdent) (ret bool) {
	_, ret = me.Own[name]
	return
}

func (me *AtEnv) set(name atValIdent, value *AtExpr) {
	util.Assert(value != nil, "AtEnv.set(name, nil)")
	me.Own[name] = value
}

func (me *AtEnv) lookup(name atValIdent) *AtExpr {
	found := me.Own[name]
	if (found == nil) && (me.Outer != nil) {
		return me.Outer.lookup(name)
	}
	return found
}
