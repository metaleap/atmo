package session

import (
	"atmo/util"
)

var (
	semPrimOpsLazy map[MoValIdent]func(*SrcPack, *SemExpr)
)

func init() {
	semPrimOpsLazy = map[MoValIdent]func(*SrcPack, *SemExpr){
		moPrimOpSet: (*SrcPack).semPrimOpSet,
	}
}

func (me *SrcPack) semRefresh() {
	me.Trees.Sem.TopLevel = nil
	me.Trees.Sem.Scope = SemScope{Own: map[MoValIdent]*SemScopeEntry{}}
	for _, top_expr := range me.Trees.MoOrig {
		it := me.semExprFromMoExpr(&me.Trees.Sem.Scope, top_expr, nil)
		me.Trees.Sem.TopLevel = append(me.Trees.Sem.TopLevel, it)
	}

	for _, top_expr := range me.Trees.Sem.TopLevel {
		me.semInterpApply(top_expr)
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
	me.Trees.Sem.Index.Lits[it] = append(me.Trees.Sem.Index.Lits[it], self)
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

	switch ident := call.Callee.MaybeIdent(); ident {
	case moPrimOpSet:
		me.semPrepScopeOnSet(self)
	case moPrimOpFn, moPrimOpMacro:
		me.semPrepScopeOnFn(self)
	}
}

type SemScope struct {
	Own    map[MoValIdent]*SemScopeEntry
	Parent *SemScope `json:"-"`
}

type SemScopeEntry struct {
	DeclParamOrSetCall *SemExpr
	SubsequentSetCalls SemExprs
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

func (me *SrcPack) semInterpExpr(self *SemExpr) {
	switch val := self.Val.(type) {
	case *SemValIdent:
		_, entry := self.Scope.Lookup(val.MoVal)
		if entry != nil {
			switch decl := entry.DeclParamOrSetCall.Val.(type) {
			default:
				panic("new bug introduced")
			case *SemValIdent: // ident refers to func param
			case *SemValCall:
				self.Val = decl.Args[1]
				me.semInterpApply(self)
			}
		} else if prim_fn := moPrimOpsEager[val.MoVal]; prim_fn == nil {
			_, is_lazy_prim_op := moPrimOpsLazy[val.MoVal]
			self.ErrsOwn.Add(self.From.SrcSpan.newDiagErr(util.If(!is_lazy_prim_op, ErrCodeUndefined, ErrCodeNotAValue), val.MoVal))
		} else {
			self.Val = prim_fn
		}
	case *SemValList:
		for _, item := range val.Items {
			me.semInterpApply(item)
		}
	case *SemValDict:
		for i, key := range val.Keys {
			me.semInterpApply(key)
			me.semInterpApply(val.Vals[i])
		}
	case *SemValCall:
		me.semInterpApply(val.Callee)
		for _, arg := range val.Args {
			me.semInterpApply(arg)
		}
	}
}

func (me *SrcPack) semInterpApply(self *SemExpr) {
	call, _ := self.Val.(*SemValCall)
	if call == nil {
		me.semInterpExpr(self)
		return
	}

	var prim_op func(*SrcPack, *SemExpr)
	if ident := call.Callee.MaybeIdent(); ident != "" {
		prim_op = semPrimOpsLazy[ident]
	}
	if prim_op != nil {
		prim_op(me, self)
		return
	}

	me.semInterpExpr(self)
	switch fn := call.Callee.Val.(type) {
	case func(*SrcPack, *SemExpr):
		fn(me, self)
	case *SemValFunc:
		// the big TODO...
	default:
		self.ErrsOwn.Add(call.Callee.From.SrcSpan.newDiagErr(ErrCodeUncallable, call.Callee.From.SrcNode.Src))
	}
}

func (me *SrcPack) semPrimOpSet(self *SemExpr) {
	call := self.Val.(*SemValCall)
	self.Fact(SemFact{Kind: SemFactPrimType, Data: MoPrimTypeVoid}, call.Callee)
	if len(call.Args) > 1 {
		me.semInterpApply(call.Args[1])
	}
}
