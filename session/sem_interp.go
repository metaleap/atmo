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

func (me *SrcPack) semEval(expr *SemExpr, scope *SemScope) {
	if expr.Type != nil {
		return
	}
	switch val := expr.Val.(type) {
	case *SemValScalar:
		expr.Type = semTypeNew(expr, val.MoVal.PrimType())
	case *SemValList:
		item_types := make(sl.Of[SemType], len(val.Items))
		for i, item := range val.Items {
			me.semEval(item, scope)
			item_types[i] = item.Type
		}

		var item_type SemType
		switch item_types.EnsureAllUnique(SemType.Eq); len(item_types) {
		case 0:
			item_type = semTypeNew(expr, MoPrimTypeAny)
		case 1:
			item_type = item_types[0]
		default:
			item_type = semTypeNew(expr, MoPrimTypeOr, item_types...)
		}
		expr.Type = semTypeNew(expr, MoPrimTypeList, item_type)
	case *SemValIdent:
		_, entry := scope.Lookup(val.Ident)
		if entry == nil {
			expr.Type = semTypeNew(expr, MoPrimTypeAny)
			expr.ErrsOwn.Add(expr.From.SrcSpan.newDiagErr(ErrCodeUndefined, val.Ident))
		}
	case *SemValFunc:
		me.semEval(val.Body, val.Scope)
		expr.Type = semTypeNew(expr, MoPrimTypeFunc, append(sl.To(val.Params, func(p *SemExpr) SemType { return p.Type }), val.Body.Type)...)
	case *SemValCall:
		sl.Each(val.Args, func(arg *SemExpr) { me.semEval(arg, scope) })

		me.semEval(val.Callee, scope)
	}
}

func (me *SrcPack) semPrimOpSet(expr *SemExpr, scope *SemScope) {

}
