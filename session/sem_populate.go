package session

import "atmo/util/sl"

func (me *SrcPack) semRefresh() {
	me.Trees.Sem.TopLevel = nil
	me.Trees.Sem.Scope = SemScope{Own: map[MoValIdent]*SemScopeEntry{}}
	for _, top_expr := range me.Trees.MoOrig {
		it := me.semExprFromMoExpr(&me.Trees.Sem.Scope, top_expr, nil)
		if call, _ := it.Val.(*SemValCall); call == nil {
			it.Fact(SemFact{Kind: SemFactUnused}, it)
		}
		me.Trees.Sem.TopLevel = append(me.Trees.Sem.TopLevel, it)
	}
}

func (me *SrcPack) semExprFromMoExpr(scope *SemScope, moExpr *MoExpr, parent *SemExpr) *SemExpr {
	ret := &SemExpr{From: moExpr, Parent: parent, Scope: scope}
	switch it := moExpr.Val.(type) {
	default:
		panic(it)
	case MoValNumInt:
		me.semPopulateScalar(ret, it)
	case MoValNumUint:
		me.semPopulateScalar(ret, it)
	case MoValNumFloat:
		me.semPopulateScalar(ret, it)
	case MoValChar:
		me.semPopulateScalar(ret, it)
	case MoValStr:
		me.semPopulateScalar(ret, it)
	case MoValIdent:
		switch it {
		case moValTrue.Val.(MoValIdent), moValFalse.Val.(MoValIdent), moValNone.Val.(MoValIdent):
			me.semPopulateScalar(ret, it)
		default:
			me.semPopulateIdent(ret, it)
		}
	case *MoValList:
		me.semPopulateList(ret, it)
	case *MoValDict:
		me.semPopulateDict(ret, it)
	case MoValCall:
		me.semPopulateCall(ret, it)
	}
	return ret
}

func (me *SrcPack) semPopulateScalar(self *SemExpr, it MoVal) {
	scalar := &SemValScalar{MoVal: it}
	self.Val = scalar
	me.Trees.Sem.Index.Lits[it] = append(me.Trees.Sem.Index.Lits[it], self)
	self.Fact(SemFact{Kind: SemFactScalar, Of: it}, self)
	self.Fact(SemFact{Kind: SemFactPrimType, Of: it.PrimType()}, self)
}

func (me *SrcPack) semPopulateIdent(self *SemExpr, it MoValIdent) {
	ident := &SemValIdent{MoVal: it}
	self.Val = ident
}

func (me *SrcPack) semPopulateList(self *SemExpr, it *MoValList) {
	list := &SemValList{Items: make(SemExprs, len(*it))}
	self.Val = list
	for i, item := range *it {
		list.Items[i] = me.semExprFromMoExpr(self.Scope, item, self)
	}
}

func (me *SrcPack) semPopulateDict(self *SemExpr, it *MoValDict) {
	dict := &SemValDict{Keys: make(SemExprs, len(*it)), Vals: make(SemExprs, len(*it))}
	self.Val = dict
	for i, item := range *it {
		dict.Keys[i] = me.semExprFromMoExpr(self.Scope, item.Key, self)
		dict.Vals[i] = me.semExprFromMoExpr(self.Scope, item.Val, self)
	}
}

func (me *SrcPack) semPopulateCall(self *SemExpr, it MoValCall) {
	call := &SemValCall{Callee: me.semExprFromMoExpr(self.Scope, it[0], self)}
	self.Val = call

	for _, arg := range it[1:] {
		call.Args = append(call.Args, me.semExprFromMoExpr(self.Scope, arg, self))
	}

	if ident := call.Callee.MaybeIdent(); ident != "" {
		if prim_op := semPrimOps[ident]; prim_op != nil {
			prim_op(me, self)
			return
		}
	}

	call.Callee.Fact(SemFact{Kind: SemFactCallable}, self)
}

type SemScope struct {
	Own    map[MoValIdent]*SemScopeEntry
	Parent *SemScope `json:"-"`
}

type SemScopeEntry struct {
	DeclParamOrSetCall *SemExpr
	SubsequentSetCalls SemExprs
}

func (me *SemScope) Lookup(ident MoValIdent, ownOnly bool, identExpr *SemExpr) (*SemScope, *SemScopeEntry) {
	if resolved := me.Own[ident]; resolved != nil {
		return me, resolved
	}
	if (!ownOnly) && (me.Parent != nil) {
		return me.Parent.Lookup(ident, false, identExpr)
	}
	if (identExpr != nil) && (semPrimOps[ident] == nil) && !sl.HasWhere(identExpr.ErrsOwn, func(it *Diag) bool { return it.Code == ErrCodeUndefined }) {
		identExpr.ErrsOwn.Add(identExpr.From.SrcSpan.newDiagErr(ErrCodeUndefined, ident))
	}
	return nil, nil
}
