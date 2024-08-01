package session

import (
	"atmo/util/sl"
)

var (
	semEvalPrimOps map[MoValIdent]func(*SrcPack, *SemExpr, *SemScope)
	semEvalPrimFns map[MoValIdent]func(*SrcPack, *SemExpr, *SemScope)
)

func init() {
	semEvalPrimOps = map[MoValIdent]func(*SrcPack, *SemExpr, *SemScope){
		moPrimOpSet: (*SrcPack).semPrimOpSet,
		moPrimOpDo:  (*SrcPack).semPrimOpDo,
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
			item_type = self.newUntyped()
		case 1:
			item_type = item_types[0]
		default:
			item_type = semTypeNew(self, MoPrimTypeOr, item_types...)
		}
		self.Type = semTypeNew(self, MoPrimTypeList, item_type)
	case *SemValIdent:
		_, entry := scope.Lookup(val.Ident)
		if entry == nil {
			self.Type = self.newUntyped()
			self.ErrsOwn.Add(self.From.SrcSpan.newDiagErr(ErrCodeUndefined, val.Ident))
		} else {
			self.Type = entry.Type
		}
	case *SemValFunc:
		me.semEval(val.Body, val.Scope)
		self.Type = semTypeNew(self, MoPrimTypeFunc, append(sl.To(val.Params, func(p *SemExpr) SemType { return p.Type }), val.Body.Type)...)
	case *SemValCall:
		var prim_op func(*SrcPack, *SemExpr, *SemScope)
		if ident := val.Callee.MaybeIdent(); ident != "" {
			prim_op = semEvalPrimOps[ident]
		}
		if prim_op != nil {
			prim_op(me, self, scope)
		} else {
			me.semEval(val.Callee, scope)
			sl.Each(val.Args, func(arg *SemExpr) { me.semEval(arg, scope) })
			switch callee := val.Callee.Val.(type) {
			case *SemValFunc:
				// dupl := *callee
				// dupl.Scope = &SemScope{Own: maps.Clone(callee.Scope.Own), Parent: callee.Scope.Parent}
				self.ErrsOwn.Add(self.ErrNew(ErrCodeAtmoTodo, "CALL OF FUNC"))
				_ = callee
				self.Type = self.newUntyped()
			default:
				val.Callee.ErrsOwn.Add(val.Callee.ErrNew(ErrCodeUncallable, val.Callee.From.String()))
				self.Type = self.newUntyped()
			}
		}
	}
}

func (me *SrcPack) semPrimOpSet(self *SemExpr, scope *SemScope) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeVoid)
	sl.Each(call.Args[1:], func(arg *SemExpr) { me.semEval(arg, scope) })
	ty := call.Args[1].Type
	_, entry := scope.Lookup(call.Args[0].Val.(*SemValIdent).Ident)
	if entry.Type == nil {
		entry.Type = ty
	} else {
		entry.Type.(*semTypeCtor).ensure(ty)
	}
	call.Args[0].Type = entry.Type
}

func (me *SrcPack) semPrimOpDo(self *SemExpr, scope *SemScope) {
	call := self.Val.(*SemValCall)
	if me.semCheckCount(1, 1, call.Args, self, true) {
		me.semEval(call.Args[0], scope)
		if list := semCheckIs[SemValList](MoPrimTypeList, call.Args[0]); list != nil {
			if me.semCheckCount(1, -1, list.Items, call.Args[0], false) {
				self.Type = list.Items[len(list.Items)-1].Type
			}
		}
	}
	if self.Type == nil {
		self.Type = self.newUntyped()
	}
}
