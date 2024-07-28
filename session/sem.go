package session

import (
	"atmo/util/sl"
)

type SemExprs sl.Of[*SemExpr]
type SemExpr struct {
	From    *MoExpr
	Parent  *SemExpr
	Scope   *SemScope
	ErrsOwn SrcFileNotices
	Val     any
}

type SemScope struct {
	Own    map[MoValIdent]*SemExpr
	Parent *SemScope
}

type SemValIdent struct {
	Refs SemExprs
}

type SemValCall struct {
	Callee   *SemExpr
	Args     SemExprs
	Callable *SemValCallable
}

type SemValCallable struct {
}

type SemValErr struct {
	Val *SemExpr
}

func (me *SrcPack) semRefresh() {
	me.Trees.Sem.Scope = SemScope{Own: map[MoValIdent]*SemExpr{}}
	for _, top_expr := range me.Trees.MoOrig {
		me.Trees.Sem.TopLevel = append(me.Trees.Sem.TopLevel, me.semExprFromMoExpr(&me.Trees.Sem.Scope, top_expr, nil))
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
	case MoValCall:
		me.semPopulateCall(ret, it)
	case MoValErr:
		me.semPopulateErr(ret, it)
	case MoValDict:
	case MoValList:
	case MoValFnPrim:
	case *MoValFnLam:
	}
	return ret
}

func (me *SrcPack) semPopulateIdent(self *SemExpr, it MoValIdent) {
	ident := &SemValIdent{}
	if resolved := self.Scope.Own[it]; resolved != nil {
		ident.Refs = append(ident.Refs, resolved)
	} else {
		self.ErrsOwn = append(self.ErrsOwn, self.From.SrcNode.newDiagErr(false, NoticeCodeUndefined, it))
	}
	self.Val = ident
}

func (me *SrcPack) semPopulateErr(self *SemExpr, it MoValErr) {
	self.Val = &SemValErr{Val: me.semExprFromMoExpr(self.Scope, it.ErrVal, self)}
}

func (me *SrcPack) semPopulateCall(self *SemExpr, it MoValCall) {
	call := &SemValCall{Callee: me.semExprFromMoExpr(self.Scope, it[0], self)}
	for _, arg := range it[1:] {
		call.Args = append(call.Args, me.semExprFromMoExpr(self.Scope, arg, self))
	}
	call.Callable, _ = call.Callee.Val.(*SemValCallable)
	var err_non_callable bool
	if call.Callable == nil {
		if ident, _ := call.Callee.Val.(*SemValIdent); ident == nil {
			err_non_callable = true
		} else if len(ident.Refs) > 0 {
			call.Callable, _ = ident.Refs[0].Val.(*SemValCallable)
			err_non_callable = (call.Callable == nil)
		}
	}
	if err_non_callable {
		self.ErrsOwn.Add(self.From.SrcNode.newDiagErr(false, NoticeCodeUncallable, self.From.SrcNode.Src))
	}
	self.Val = call
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
	}
	if onAfter != nil {
		onAfter(me)
	}
}
