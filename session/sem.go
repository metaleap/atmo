package session

import (
	"atmo/util/sl"
)

type SemExprs sl.Of[*SemExpr]
type SemExpr struct {
	From             *MoExpr
	Parent           *SemExpr
	Scope            *SemScope
	ErrsOwn          SrcFileNotices
	Val              any
	DefinitelyUnused bool
}

type SemScope struct {
	Own    map[MoValIdent]*SemExpr
	Parent *SemScope
}

func (me *SemScope) Lookup(ident MoValIdent, ownOnly bool, recursivelyResolveUntilNonIdent bool) (ret SemExprs) {
	if resolved := me.Own[ident]; resolved != nil {
		if !recursivelyResolveUntilNonIdent {
			ret = append(ret, resolved)
		} else if alias, _ := resolved.Val.(*SemValIdent); alias == nil {
			ret = append(ret, resolved)
		} else {
			ret = append(ret, me.Lookup(resolved.From.Val.(MoValIdent), ownOnly, true)...)
		}
	}
	if (!ownOnly) && (me.Parent != nil) {
		ret = append(ret, me.Parent.Lookup(ident, false, recursivelyResolveUntilNonIdent)...)
	}
	return
}

type SemValIdent struct {
	Orig          SemExprs // these may be idents as well
	Full          SemExprs // these are fully-resolved, ie. never idents
	IsParamOfFunc *SemExpr
}

type SemValCall struct {
	Callee   *SemExpr
	Args     SemExprs
	Callable *SemExpr
	Inert    bool
	Decl     bool
}

type SemValErr struct {
	Val *SemExpr
}

type SemValList struct {
	Items SemExprs
}

type SemValDict struct {
	Keys SemExprs
	Vals SemExprs
}

type SemValFunc struct {
	Scope  *SemScope
	Params SemExprs
	Body   *SemExpr
}

func (me *SrcPack) semRefresh() {
	me.Trees.Sem.TopLevel = nil
	me.Trees.Sem.Scope = SemScope{Own: map[MoValIdent]*SemExpr{}}
	for _, top_expr := range me.Trees.MoOrig {
		it := me.semExprFromMoExpr(&me.Trees.Sem.Scope, top_expr, nil)
		if call, _ := it.Val.(*SemValCall); call == nil {
			it.DefinitelyUnused = true
		}
		me.Trees.Sem.TopLevel = append(me.Trees.Sem.TopLevel, it)
	}
}

func (me *SrcPack) semExprFromMoExpr(scope *SemScope, moExpr *MoExpr, parent *SemExpr) *SemExpr {
	ret := &SemExpr{From: moExpr, Parent: parent, Scope: scope}
	switch it := moExpr.Val.(type) {
	default:
		panic(it)
	case MoValPrimTypeTag:
		me.semIndexAddLit(ret)
	case MoValNumInt:
		me.semIndexAddLit(ret)
	case MoValNumUint:
		me.semIndexAddLit(ret)
	case MoValNumFloat:
		me.semIndexAddLit(ret)
	case MoValChar:
		me.semIndexAddLit(ret)
	case MoValStr:
		me.semIndexAddLit(ret)
	case MoValIdent:
		me.semPopulateIdent(ret, it)
	case MoValErr:
		me.semPopulateErr(ret, it)
	case MoValList:
		me.semPopulateList(ret, it)
	case MoValDict:
		me.semPopulateDict(ret, it)
	case MoValCall:
		me.semPopulateCall(ret, it)
	case *MoValFnLam:
		me.semPopulateFunc(ret, it)
	}
	return ret
}

func (me *SrcPack) semPopulateIdent(self *SemExpr, it MoValIdent) {
	ident := &SemValIdent{Orig: self.Scope.Lookup(it, false, false), Full: self.Scope.Lookup(it, false, true)}
	self.Val = ident
}

func (me *SrcPack) semPopulateErr(self *SemExpr, it MoValErr) {
	self.Val = &SemValErr{Val: me.semExprFromMoExpr(self.Scope, it.ErrVal, self)}
}

func (me *SrcPack) semPopulateList(self *SemExpr, it MoValList) {
	list := &SemValList{Items: make(SemExprs, len(it))}
	self.Val = list
	for i, item := range it {
		list.Items[i] = me.semExprFromMoExpr(self.Scope, item, self)
	}
}

func (me *SrcPack) semPopulateDict(self *SemExpr, it MoValDict) {
	dict := &SemValDict{Keys: make(SemExprs, len(it)), Vals: make(SemExprs, len(it))}
	self.Val = dict
	for i, item := range it {
		dict.Keys[i] = me.semExprFromMoExpr(self.Scope, item.Key, self)
		dict.Vals[i] = me.semExprFromMoExpr(self.Scope, item.Val, self)
	}
}

func (me *SrcPack) semPopulateCall(self *SemExpr, it MoValCall) {
	call := &SemValCall{Callee: me.semExprFromMoExpr(self.Scope, it[0], self)}
	self.Val = call

	var err_non_callable bool
	switch callee := call.Callee.Val.(type) {
	case *SemValFunc:
		call.Callable = call.Callee
	case *SemValIdent:
		if len(callee.Orig) == 0 {
			call.Callee.ErrsOwn.Add(call.Callee.From.SrcNode.newDiagErr(false, NoticeCodeUndefined, it))
		} else if len(callee.Full) > 0 {
			call.Callable = callee.Full[0]
			if _, is := call.Callable.Val.(*SemValFunc); !is {
				err_non_callable = true
			}
		}
	default:
		err_non_callable = true
	}
	if err_non_callable {
		self.ErrsOwn.Add(self.From.SrcNode.newDiagErr(false, NoticeCodeUncallable, self.From.SrcNode.Src))
	}

	for _, arg := range it[1:] {
		call.Args = append(call.Args, me.semExprFromMoExpr(self.Scope, arg, self))
	}
	// TODO!

	call.Inert = call.Callee.HasErrs() || call.Args.AnyErrs() || (call.Callable == nil) || call.Callable.HasErrs()
}

func (me *SrcPack) semPopulateFunc(self *SemExpr, it *MoValFnLam) {
	fn := &SemValFunc{
		Scope:  &SemScope{Parent: self.Scope, Own: map[MoValIdent]*SemExpr{}},
		Params: make(SemExprs, len(it.Params)),
	}
	self.Val = fn
	for i, param := range it.Params {
		fn.Params[i] = me.semExprFromMoExpr(self.Scope, param, self)
		fn.Scope.Own[param.Val.(MoValIdent)] = fn.Params[i]
		fn.Params[i].Val.(*SemValIdent).IsParamOfFunc = self
	}
	fn.Body = me.semExprFromMoExpr(fn.Scope, it.Body, self)
}

func (me *SrcPack) semIndexAddLit(it *SemExpr) {
	lit_val := it.From.Val
	me.Trees.Sem.Index.Lits[lit_val] = append(me.Trees.Sem.Index.Lits[lit_val], it)
}

func (me *SemExpr) HasErrs() (ret bool) {
	me.Walk(func(it *SemExpr) bool {
		ret = ret || (len(it.ErrsOwn) > 0)
		return !ret
	}, nil)
	return
}

func (me *SemExpr) Errs() (ret SrcFileNotices) {
	me.Walk(nil, func(it *SemExpr) {
		ret.Add(it.ErrsOwn...)
	})
	return
}

func (me SemExprs) AnyErrs() bool {
	return sl.Any(me, (*SemExpr).HasErrs)
}

func (me SemExprs) Errs() (ret SrcFileNotices) {
	for _, top_expr := range me {
		ret.Add(top_expr.Errs()...)
	}
	return
}

func (me SemExprs) Walk(filterBy *SrcFile, onBefore func(it *SemExpr) bool, onAfter func(it *SemExpr)) {
	for _, expr := range me {
		if (filterBy == nil) || (expr.From.SrcFile == filterBy) {
			expr.Walk(onBefore, onAfter)
		}
	}
}

func (me *SemExpr) Walk(onBefore func(it *SemExpr) bool, onAfter func(it *SemExpr)) {
	if onBefore != nil && !onBefore(me) {
		return
	}
	switch it := me.Val.(type) {
	case *SemValCall:
		it.Callee.Walk(onBefore, onAfter)
		for _, arg := range it.Args {
			arg.Walk(onBefore, onAfter)
		}
	case *SemValErr:
		it.Val.Walk(onBefore, onAfter)
	case *SemValList:
		for _, item := range it.Items {
			item.Walk(onBefore, onAfter)
		}
	case *SemValDict:
		for i, key := range it.Keys {
			key.Walk(onBefore, onAfter)
			it.Vals[i].Walk(onBefore, onAfter)
		}
	case *SemValFunc:
		for _, param := range it.Params {
			param.Walk(onBefore, onAfter)
		}
		it.Body.Walk(onBefore, onAfter)
	}
	if onAfter != nil {
		onAfter(me)
	}
}
