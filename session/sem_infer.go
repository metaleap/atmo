package session

import (
	"atmo/util/sl"
	"maps"
)

func (me *SrcPack) semInfer() {
	ctx := &semInferCtx{}
	env := semInferEnv{}
	for _, top_expr := range me.Trees.Sem.TopLevel {
		top_expr.Type, _ = me.inferExpr(ctx, top_expr, env)
	}
}

type semInferCtx struct {
	next int
}

type semInferEnv map[MoValIdent]*SemType
type semInferSubst map[int]*SemType

func (me *SrcPack) inferExpr(ctx *semInferCtx, self *SemExpr, env semInferEnv) (*SemType, semInferSubst) {
	switch val := self.Val.(type) {
	case *SemValScalar:
		return semTypeNew(self, val.Value.PrimType()), semInferSubst{}
	case *SemValIdent:
		ty := env[val.Name]
		if ty == nil {
			self.ErrAdd(self.ErrNew(ErrCodeAtmoTodo, "undefined:"+string(val.Name)))
		}
		return ty, semInferSubst{}
	case *SemValFunc:
		new_env := maps.Clone(env)
		ty_vars := make([]*SemType, len(val.Params))
		for i, param := range val.Params {
			ty_vars[i] = ctx.newTVar(param)
			new_env[param.MaybeIdent(true)] = ty_vars[i]
		}
		ty_body, subst := me.inferExpr(ctx, val.Body, new_env)
		if ty_body == nil {
			return nil, nil
		}
		ty_fn_targs := make([]*SemType, len(val.Params)+1)
		for i := range val.Params {
			ty_fn_targs[i] = ctx.applySubstToType(subst, ty_vars[i])
		}
		ty_fn_targs[len(ty_fn_targs)-1] = ty_body
		ty_fn := semTypeNew(self, MoPrimTypeFunc, ty_fn_targs...)
		return ty_fn, subst
	case *SemValCall:
		ty_fn_0, s1 := me.inferExpr(ctx, val.Callee, env)
		ty_arg, s2 := make([]*SemType, len(val.Args)), make([]semInferSubst, len(val.Args))
		for i, arg := range val.Args {
			ty_arg[i], s2[i] = me.inferExpr(ctx, arg, ctx.applySubstToEnv(s1, env))
		}
		new_var := ctx.newTVar(self)
		s3 := ctx.composeSubst(s1, s2...)
		if (ty_fn_0 == nil) || sl.Has(ty_arg, nil) {
			return nil, nil
		}
		s4, err := ctx.unify(self, semTypeNew(self, MoPrimTypeFunc, append(ty_arg, new_var)...), ty_fn_0)
		if err != nil {
			self.ErrAdd(err)
			return nil, nil
		}
		ty_fn_1 := ctx.applySubstToType(s4, ty_fn_0)
		s5 := ctx.composeSubst(s3, s4)
		idx, erred := -1, false
		s6 := sl.To(ty_arg, func(tyArg *SemType) semInferSubst {
			idx++
			ret, err := ctx.unify(val.Args[idx], ctx.applySubstToType(s5, ty_fn_1.TArgs[idx]), tyArg)
			if err != nil {
				self.ErrAdd(err)
				erred = true
			}
			return ret
		})
		if erred {
			return nil, nil
		}
		result_subst := ctx.composeSubst(s5, s6...)
		return ctx.applySubstToType(result_subst, ty_fn_1.TArgs[len(ty_fn_1.TArgs)-1]), result_subst
	}
	return nil, nil
}

func (me *semInferCtx) applySubstToEnv(subst semInferSubst, env semInferEnv) semInferEnv {
	ret := maps.Clone(env)
	for k, t := range ret {
		ret[k] = me.applySubstToType(subst, t)
	}
	return ret
}

func (me *semInferCtx) applySubstToType(subst semInferSubst, tyOrTVar *SemType) *SemType {
	if tyOrTVar.Prim < 0 { // type-var
		if ty := subst[int(tyOrTVar.Prim)]; ty != nil {
			return ty
		}
	} else if tyOrTVar.Prim == MoPrimTypeFunc {
		ty_fn_targs := make([]*SemType, len(tyOrTVar.TArgs))
		for i := range ty_fn_targs {
			ty_fn_targs[i] = me.applySubstToType(subst, tyOrTVar.TArgs[i])
		}
		return semTypeNew(tyOrTVar.DueTo, MoPrimTypeFunc, ty_fn_targs...)
	}
	return tyOrTVar
}

func (me *semInferCtx) composeSubst(s1 semInferSubst, s2 ...semInferSubst) semInferSubst {
	result := maps.Clone(s1)
	for _, s2 := range s2 {
		for k, t2 := range s2 {
			result[k] = me.applySubstToType(result, t2)
		}
	}
	return result
}

func (me *semInferCtx) newTVar(dueTo *SemExpr) *SemType {
	me.next--
	return semTypeNew(dueTo, MoValPrimType(me.next))
}

func (me *semInferCtx) unify(self *SemExpr, t1 *SemType, t2 *SemType) (semInferSubst, *Diag) {
	if t1.Prim < 0 {
		return me.tVarBind(int(t1.Prim), t2)
	} else if t2.Prim < 0 {
		return me.tVarBind(int(t2.Prim), t1)
	} else if (t1.Prim == MoPrimTypeFunc) && (t2.Prim == MoPrimTypeFunc) {
		var err *Diag
		s1s := make([]semInferSubst, len(t1.TArgs)-1)
		for i := range t1.TArgs[:len(t1.TArgs)-1] {
			if s1s[i], err = me.unify(self, t1.TArgs[i], t2.TArgs[i]); err != nil {
				return nil, err
			}
		}
		var s1 semInferSubst
		if len(s1s) == 0 {
			s1 = semInferSubst{}
		} else {
			s1 = me.composeSubst(s1s[0], s1s[1:]...)
		}
		s2, err := me.unify(self, me.applySubstToType(s1, t1.TArgs[len(t1.TArgs)-1]), me.applySubstToType(s1, t2.TArgs[len(t2.TArgs)-1]))
		if err != nil {
			return nil, err
		}
		return me.composeSubst(s1, s2), nil
	} else if t1.Eq(t2) {
		return semInferSubst{}, nil
	}
	return nil, self.ErrNew(ErrCodeAtmoTodo, "typeMismatch")
}

func (me *semInferCtx) tVarBind(n int, ty *SemType) (semInferSubst, *Diag) {
	if n == int(ty.Prim) {
		return semInferSubst{}, nil
	} else if me.tVarOccurs(ty, n) {
		return nil, ty.DueTo.ErrNew(ErrCodeAtmoTodo, "cyclic")
	}
	return semInferSubst{n: ty}, nil
}

func (me *semInferCtx) tVarOccurs(ty *SemType, n int) bool {
	if ty.Prim < 0 {
		return n == int(ty.Prim)
	} else if ty.Prim == MoPrimTypeFunc {
		return sl.Any(ty.TArgs, func(tyArg *SemType) bool { return me.tVarOccurs(tyArg, n) })
	}
	return false
}
