package session

import (
	"atmo/util/sl"
)

type semTyEnv map[MoValIdent]*SemType

func (me *SrcPack) semTySynth() {
	env := semTyEnv{}
	for _, top_expr := range me.Trees.Sem.TopLevel {
		top_expr.Type = me.semTySynthFor(top_expr, env)
	}
}

func (me *SrcPack) semTySynthFor(self *SemExpr, env semTyEnv) *SemType {
	switch val := self.Val.(type) {
	case *SemValScalar:
		self.Type = semTypeNew(self, val.Value.PrimType())
	case *SemValIdent:
		self.Type = env[val.Name]
	case *SemValList:
		if val.IsTup {
			self.Type = semTypeNew(self, MoPrimTypeTup, sl.To(val.Items, func(it *SemExpr) *SemType { return me.semTySynthFor(it, env) })...)
		}
	case *SemValDict:
		if val.IsObj {
			self.Type = semTypeNew(self, MoPrimTypeObj)
			self.Type.TArgs, self.Type.Fields = make(sl.Of[*SemType], len(val.Keys)), make(sl.Of[MoValIdent], len(val.Keys))
			for i, key := range val.Keys {
				self.Type.Fields[i] = key.MaybeIdent(true)
				self.Type.TArgs[i] = me.semTySynthFor(val.Vals[i])
			}
			if sl.Has(self.Type.TArgs, nil) {
				self.Type = nil
			}
		}
	case *SemValFunc:
	case *SemValCall:
	}
	return self.Type
}
