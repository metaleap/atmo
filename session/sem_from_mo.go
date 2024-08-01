package session

import (
	"atmo/util/sl"
	"atmo/util/str"
)

func (me *SrcPack) semRefresh() {
	me.Trees.Sem.TopLevel = make(SemExprs, 0, len(me.Trees.MoOrig))
	me.Trees.Sem.Scope = SemScope{Own: map[MoValIdent]*SemScopeEntry{}}
	for _, top_expr := range me.Trees.MoOrig {
		it := me.semExprFromMoExpr(&me.Trees.Sem.Scope, top_expr, nil)
		me.Trees.Sem.TopLevel = append(me.Trees.Sem.TopLevel, it)
	}
	if !me.Trees.Sem.TopLevel.AnyErrs() {
		// me.semInferTypes()
		me.semPopulateRootScope()
		for _, top_expr := range me.Trees.Sem.TopLevel {
			me.semEval(top_expr, &me.Trees.Sem.Scope)
		}
	}
}

func (me *SrcPack) semExprFromMoExpr(scope *SemScope, moExpr *MoExpr, parent *SemExpr) *SemExpr {
	ret := &SemExpr{From: moExpr, Parent: parent, Scope: scope}
	switch it := moExpr.Val.(type) {
	default:
		panic(it)
	case MoValVoid, MoValBool, MoValNumInt, MoValNumUint, MoValNumFloat, MoValChar, MoValStr:
		me.semPopulateScalar(ret, it)
	case MoValIdent:
		me.semPopulateIdent(ret, it)
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
	self.Type = semTypeNew(self, it.PrimType())
	me.Trees.Sem.Index.Lits[it] = append(me.Trees.Sem.Index.Lits[it], self)
}

func (me *SrcPack) semPopulateIdent(self *SemExpr, it MoValIdent) {
	ident := &SemValIdent{Ident: it}
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
}

func (me *SrcPack) semPopulateRootScope() {
	for name, prim_fn := range semEvalPrimFns {
		fn_val := &SemValFunc{primImpl: prim_fn}
		fn := &SemExpr{Scope: &me.Trees.Sem.Scope, Type: semEvalPrimFnTypes[name], Val: fn_val}
		fn.Type = semTypeEnsureDueTo(fn, fn.Type)
		var idx int
		fn_val.Params = SemExprs(sl.To(fn.Type.(*semTypeCtor).tyArgs, func(t SemType) *SemExpr {
			idx++
			return &SemExpr{Parent: fn, Scope: fn.Scope, Type: t, Val: &SemValIdent{Ident: MoValIdent("arg" + str.FromInt(idx))}}
		}))
		fn.Fact(SemFact{Kind: SemFactPrimFn}, fn)
		me.Trees.Sem.Scope.Own[name] = &SemScopeEntry{
			Type:                     fn.Type,
			DeclParamOrSetCallOrFunc: fn,
		}
	}

	me.Trees.Sem.TopLevel.Walk(nil, func(self *SemExpr) bool {
		if call, _ := self.Val.(*SemValCall); call != nil {
			if ident := call.Callee.MaybeIdent(false); (ident == moPrimOpQQuote) || (ident == moPrimOpQuote) {
				return false
			}
		}
		return true
	}, func(self *SemExpr) {
		if call, _ := self.Val.(*SemValCall); call != nil {
			switch ident := call.Callee.MaybeIdent(true); ident {
			case moPrimOpSet:
				me.semPrepScopeOnSet(self)
			case moPrimOpFn, moPrimOpMacro:
				me.semPrepScopeOnFn(self)
			}
		}
	})
}

type SemScope struct {
	Own    map[MoValIdent]*SemScopeEntry
	Parent *SemScope `json:"-"`
}

type SemScopeEntry struct {
	DeclParamOrSetCallOrFunc *SemExpr
	SubsequentSetCalls       SemExprs
	Type                     SemType
}

func (me *SemScope) Lookup(ident MoValIdent) (*SemScope, *SemScopeEntry) {
	if resolved := me.Own[ident]; resolved != nil {
		return me, resolved
	}
	if me.Parent != nil {
		return me.Parent.Lookup(ident)
	}
	return nil, nil
}
