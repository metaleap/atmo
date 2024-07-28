package session

import "atmo/util/sl"

type SemExprs sl.Of[*SemExpr]
type SemExpr struct {
	Src     *MoExpr
	Parent  *SemExpr
	ErrsOwn []*SrcFileNotice
}

func (me *SrcPack) semRefresh() {
	for _, top_expr := range me.Trees.MoOrig {
		me.Trees.Sem.TopLevel = append(me.Trees.Sem.TopLevel, me.semExprFromMoExpr(top_expr, nil))
	}
}

func (me *SrcPack) semExprFromMoExpr(moExpr *MoExpr, parent *SemExpr) *SemExpr {
	ret := &SemExpr{Src: moExpr, Parent: parent}
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
	case MoValErr:
	case MoValDict:
	case MoValList:
	case MoValCall:
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
