package session

import (
	"atmo/util/sl"
)

type SemExprs sl.Of[*SemExpr]
type SemExpr struct {
	Src     *MoExpr
	Parent  *SemExpr
	Scope   *SemScope
	ErrsOwn []*SrcFileNotice
	Refs    SemExprs
}

type SemScope struct {
	Own    map[MoValIdent]*SemExpr
	Parent *SemScope
}

func (me *SrcPack) semRefresh() {
	me.Trees.Sem.Scope = SemScope{Own: map[MoValIdent]*SemExpr{}}
	for _, top_expr := range me.Trees.MoOrig {
		me.Trees.Sem.TopLevel = append(me.Trees.Sem.TopLevel, me.semExprFromMoExpr(&me.Trees.Sem.Scope, top_expr, nil))
	}
}

func (me *SrcPack) semExprFromMoExpr(scope *SemScope, moExpr *MoExpr, parent *SemExpr) *SemExpr {
	ret := &SemExpr{Src: moExpr, Parent: parent, Scope: scope}
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
		if resolved := scope.Own[it]; resolved != nil {
			ret.Refs = append(ret.Refs, resolved)
		} else {
			ret.ErrsOwn = append(ret.ErrsOwn, moExpr.SrcNode.newDiagErr(false, NoticeCodeUndefined, it))
		}
	case MoValErr:
	case MoValDict:
	case MoValList:
	case MoValCall:
		ret.ErrsOwn = append(ret.ErrsOwn, moExpr.SrcNode.newDiagErr(false, NoticeCodeUndefined, it.PrimType()))
	case MoValFnPrim:
	case *MoValFnLam:
	}
	return ret
}

func (me *SrcPack) semIndexAddLit(it *SemExpr) {
	lit_val := it.Src.Val
	me.Trees.Sem.Index.Lits[lit_val] = append(me.Trees.Sem.Index.Lits[lit_val], it)
}

func (me *SemExpr) Errs() (ret SrcFileNotices) {
	me.Walk(nil, func(it *SemExpr) {
		ret = append(ret, it.ErrsOwn...)
	})
	return
}

func (me SemExprs) Errs() (ret SrcFileNotices) {
	for _, top_expr := range me {
		ret = append(ret, top_expr.Errs()...)
	}
	return
}

func (me SemExprs) Walk(filterBy *SrcFile, onBefore func(it *SemExpr) bool, onAfter func(it *SemExpr)) {
	for _, expr := range me {
		if (filterBy == nil) || (expr.Src.SrcFile == filterBy) {
			expr.Walk(onBefore, onAfter)
		}
	}
}

func (me *SemExpr) Walk(onBefore func(it *SemExpr) bool, onAfter func(it *SemExpr)) {
	if onBefore != nil && !onBefore(me) {
		return
	}
	// TODO: traverse
	if onAfter != nil {
		onAfter(me)
	}
}
