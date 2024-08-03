package session

import "maps"

func (me *SrcPack) semInfer() {
	ctx := &semInferCtx{}
	env := map[MoValIdent]*SemType{}
	for _, top_expr := range me.Trees.Sem.TopLevel {
		top_expr.Type, _ = me.semInferExpr(ctx, top_expr, env)
	}
}

type semInferCtx struct {
	next int
}

func (me *SrcPack) semInferExpr(ctx *semInferCtx, self *SemExpr, env map[MoValIdent]*SemType) (*SemType, map[int]*SemType) {
	switch val := self.Val.(type) {
	case *SemValScalar:
		return semTypeNew(self, val.Value.PrimType()), nil
	case *SemValIdent:
		ty := env[val.Name]
		if ty == nil {
			self.ErrAdd(self.ErrNew(ErrCodeAtmoTodo, "undefined:"+string(val.Name)))
		}
		return ty, map[int]*SemType{}
	case *SemValFunc:
		new_env := maps.Clone(env)
		ty_vars := make([]*SemType, len(val.Params))
		for i, param := range val.Params {
			ty_vars[i] = ctx.newTVar(param)
			new_env[param.MaybeIdent(true)] = ty_vars[i]
		}
		ty_body, subst := me.semInferExpr(ctx, val.Body, new_env)
		if ty_body == nil {
			return nil, nil
		}
		ty_fn_targs := make([]*SemType, len(val.Params)+1)
		for i := range val.Params {
			ty_fn_targs[i] = ctx.applySubst(subst, ty_vars[i])
		}
		ty_fn_targs[len(ty_fn_targs)-1] = ty_body
		ty_fn := semTypeNew(self, MoPrimTypeFunc, ty_fn_targs...)
		return ty_fn, subst
	}
	return nil, nil
}

func (me *semInferCtx) applySubst(subst map[int]*SemType, tyOrTVar *SemType) *SemType {
	if tyOrTVar.Prim < 0 { // type-var

	} else if tyOrTVar.Prim == MoPrimTypeFunc {
		ty_fn_targs := make([]*SemType, len(tyOrTVar.TArgs))
		for i := range ty_fn_targs {
			ty_fn_targs[i] = me.applySubst(subst, tyOrTVar.TArgs[i])
		}
		return semTypeNew(tyOrTVar.DueTo, MoPrimTypeFunc, ty_fn_targs...)
	}
	return tyOrTVar
}

func (me *semInferCtx) newTVar(dueTo *SemExpr) *SemType {
	me.next--
	return semTypeNew(dueTo, MoValPrimType(me.next))
}
