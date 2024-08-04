package session

import (
	"maps"

	"atmo/util"
	"atmo/util/sl"
)

type semInferCtx struct {
	next int
}
type semInferEnv map[MoValIdent]*SemType
type semInferSubst map[int]*SemType

func (me *SrcPack) semInferTypes() {
	ctx, env := semInferCtx{}, semInferEnv{}
	for _, top_expr := range me.Trees.Sem.TopLevel {
		var subst semInferSubst
		top_expr.Type, subst = ctx.semTypeInferExpr(top_expr, env)

		// at this point, fully-typified top-level exprs have their `.Type` set alright. but many/most of their
		// inner Exprs have type-vars still that we now resolve by walk, thus also normalizing any remaining type-var
		// expr `.Type`s to `nil` which then means "expr untypifyable" (due to errors already reported along the way)
		var fully_typified func(*SemType) bool
		fully_typified = func(t *SemType) bool {
			return (t != nil) && (t.Prim >= 0) && sl.All(t.TArgs, fully_typified)
		}
		top_expr.Walk(false, nil, func(it *SemExpr) {
			if it.Type != nil {
				it.Type = ctx.applySubstToType(subst, it.Type)
				if !fully_typified(it.Type) {
					it.Type = nil
				} else if fn, _ := it.Val.(*SemValFunc); fn != nil { // check all funcs for untypified (ie. unreferenced) Params
					for _, param := range fn.Params {
						if param.Type == nil {
							param.Type = semTypeNew(fn.Body, MoPrimTypeAny)
							param.Fact(SemFact{Kind: SemFactUnused}, fn.Body)
						}
					}
				}
			}
		})
	}
}

func (me *semInferCtx) semTypeInferExpr(self *SemExpr, env semInferEnv) (*SemType, semInferSubst) {
	switch val := self.Val.(type) {
	case *SemValScalar:
		self.Type = semTypeNew(self, val.Value.PrimType())
		return self.Type, semInferSubst{}
	case *SemValIdent:
		ty := env[val.Name]
		if ty == nil {
			if ty = semPrimFnTypes[val.Name]; ty != nil {
				ty = semTypeEnsureDueTo(self, ty)
			}
		}
		if ty == nil {
			is_prim_op := semTyPrimOps[val.Name] != nil
			self.ErrAdd(self.ErrNew(util.If(is_prim_op, ErrCodeNotAValue, ErrCodeUndefined), val.Name))
		}
		self.Type = ty
		return ty, semInferSubst{}
	case *SemValFunc:
		new_env := maps.Clone(env)
		ty_vars := make([]*SemType, len(val.Params))
		for i, param := range val.Params {
			ty_vars[i] = me.newTVar(param)
			new_env[param.MaybeIdent(true)] = ty_vars[i]
		}
		ty_body, subst := me.semTypeInferExpr(val.Body, new_env)
		if ty_body == nil {
			return nil, nil
		}
		ty_fn_targs := make([]*SemType, len(val.Params)+1)
		for i := range val.Params {
			ty_fn_targs[i] = me.applySubstToType(subst, ty_vars[i])
			val.Params[i].Type = ty_fn_targs[i]
		}
		ty_fn_targs[len(ty_fn_targs)-1] = ty_body
		ty_fn := semTypeNew(self, MoPrimTypeFunc, ty_fn_targs...)
		self.Type = ty_fn
		return ty_fn, subst
	case *SemValCall:
		ty_fn_0, s1 := me.semTypeInferExpr(val.Callee, env)
		ty_arg, s2 := make([]*SemType, len(val.Args)), make([]semInferSubst, len(val.Args))
		for i, arg := range val.Args {
			ty_arg[i], s2[i] = me.semTypeInferExpr(arg, me.applySubstToEnv(s1, env))
		}
		if (ty_fn_0 == nil) || sl.Has(ty_arg, nil) {
			return nil, nil
		}
		new_var := me.newTVar(self)
		s3 := me.composeSubst(s1, s2...)
		s4, err := me.unify(self, semTypeNew(self, MoPrimTypeFunc, append(ty_arg, new_var)...), ty_fn_0)
		if err != nil {
			self.ErrAdd(err)
			return nil, nil
		}
		ty_fn_1 := me.applySubstToType(s4, ty_fn_0)
		s5 := me.composeSubst(s3, s4)
		idx, erred := -1, false
		s6 := sl.To(ty_arg, func(tyArg *SemType) semInferSubst {
			idx++
			ret, err := me.unify(val.Args[idx], me.applySubstToType(s5, ty_fn_1.TArgs[idx]), tyArg)
			if err != nil {
				self.ErrAdd(err)
				erred = true
			}
			return ret
		})
		if erred {
			return nil, nil
		}
		result_subst := me.composeSubst(s5, s6...)
		self.Type = me.applySubstToType(result_subst, ty_fn_1.TArgs[len(ty_fn_1.TArgs)-1])
		return self.Type, result_subst
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
	return nil, semTypeErrOn(self, t1, t2)
}

func (me *semInferCtx) tVarBind(n int, ty *SemType) (semInferSubst, *Diag) {
	if n == int(ty.Prim) {
		return semInferSubst{}, nil
	} else if me.tVarOccurs(ty, n) {
		return nil, ty.DueTo.ErrNew(ErrCodeTypeInfinite, ty.String())
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
