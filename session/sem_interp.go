package session

import "atmo/util/sl"

var (
	semEvalPrimOps map[MoValIdent]func(*SrcPack, *SemExpr, *SemScope)
	semEvalPrimFns map[MoValIdent]func(*SrcPack, *SemExpr, *SemScope)
)

func init() {
	semEvalPrimOps = map[MoValIdent]func(*SrcPack, *SemExpr, *SemScope){
		moPrimOpSet: (*SrcPack).semPrimOpSet,
	}
	semEvalPrimFns = map[MoValIdent]func(*SrcPack, *SemExpr, *SemScope){}
}

func (me *SrcPack) semEval(self *SemExpr, scope *SemScope) {
	if (self.Type != nil) || (len(self.ErrsOwn) > 0) {
		return
	}
	switch val := self.Val.(type) {
	case *SemValScalar:
		self.Type = semTypeNew(self, val.MoVal.PrimType())
	case *SemValList:
		item_types := make(sl.Of[SemType], len(val.Items))
		for i, item := range val.Items {
			me.semEval(item, scope)
			item_types[i] = item.Type
		}

		var item_type SemType
		switch item_types.EnsureAllUnique(SemType.Eq); len(item_types) {
		case 0:
			item_type = semTypeNew(self, MoPrimTypeAny)
		case 1:
			item_type = item_types[0]
		default:
			item_type = semTypeNew(self, MoPrimTypeOr, item_types...)
		}
		self.Type = semTypeNew(self, MoPrimTypeList, item_type)
	case *SemValIdent:
		_, entry := scope.Lookup(val.Ident)
		if entry == nil {
			self.Type = semTypeNew(self, MoPrimTypeAny)
			self.ErrsOwn.Add(self.From.SrcSpan.newDiagErr(ErrCodeUndefined, val.Ident))
		}
	case *SemValFunc:
		me.semEval(val.Body, val.Scope)
		self.Type = semTypeNew(self, MoPrimTypeFunc, append(sl.To(val.Params, func(p *SemExpr) SemType { return p.Type }), val.Body.Type)...)
	case *SemValCall:
		sl.Each(val.Args, func(arg *SemExpr) { me.semEval(arg, scope) })

		me.semEval(val.Callee, scope)
	}
}

func (me *SrcPack) semPrimOpSet(self *SemExpr, scope *SemScope) {
	call := self.Val.(*SemValCall)
	_ = call
}
