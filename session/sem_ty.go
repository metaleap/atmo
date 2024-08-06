package session

import "atmo/util/sl"

func (me *SrcPack) semTySynth() {
	for _, top_expr := range me.Trees.Sem.TopLevel {
		top_expr.Type = me.semTySynthFor(top_expr)
	}
}

func (me *SrcPack) semTySynthFor(self *SemExpr) *SemType {
	switch val := self.Val.(type) {
	case *SemValScalar:
		self.Type = semTypeNew(self, val.Value.PrimType())
	case *SemValIdent:
	case *SemValList:
		if val.IsTup {
			self.Type = semTypeNew(self, MoPrimTypeTup, sl.To(val.Items, me.semTySynthFor)...)
		}
	case *SemValDict:
		if val.IsObj {
			self.Type = semTypeNew(self, MoPrimTypeObj)
			self.Type.TArgs, self.Type.Fields = make(sl.Of[*SemType], len(val.Keys)), make(sl.Of[MoValIdent], len(val.Keys))
			for i, key := range val.Keys {
				self.Type.Fields[i] = key.MaybeIdent(true)
				self.Type.TArgs[i] = me.semTySynthFor(val.Vals[i])
			}
		}
	case *SemValFunc:
	case *SemValCall:
	}
	return self.Type
}
