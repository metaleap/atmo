package session

import (
	"maps"

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
		ctx.next = 0
		var subst semInferSubst
		top_expr.Type, subst = ctx.inferExpr(top_expr, env)

		// // at this point, fully-typified top-level exprs have their `.Type` set alright. but many/most of their
		// // inner Exprs have type-vars still that we now resolve by walk, thus also normalizing any remaining type-var
		// // expr `.Type`s to `nil` which then means "expr untypifyable" (due to errors already reported along the way)
		// var fully_typified func(*SemType) bool
		// fully_typified = func(t *SemType) bool {
		// 	return (t.Prim >= 0) && sl.All(t.TArgs, fully_typified)
		// }
		top_expr.Walk(nil, func(it *SemExpr) {
			if it.Type != nil {
				it.Type = ctx.applySubstToType(subst, it.Type)
				// if !fully_typified(it.Type) {
				// 	it.Type = nil
				// }
			}
		})
	}
}

func (me *semInferCtx) inferExpr(self *SemExpr, env semInferEnv) (*SemType, semInferSubst) {
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
			if ty == nil && val.Unresolved { // can still typify the unresolved symbol
				ty = me.newTVar(self)
				env[val.Name] = ty
			}
		}
		self.Type = ty
		return ty, semInferSubst{}

	case *SemValFunc:
		new_env := maps.Clone(env)
		ty_vars := make([]*SemType, len(val.Params))
		for i, param := range val.Params {
			if !param.Val.(*SemValIdent).IsDeclUsed {
				ty_vars[i] = semTypeNew(val.Body, MoPrimTypeAny)
				param.Fact(SemFact{Kind: SemFactUnused}, val.Body)
			} else {
				ty_vars[i] = me.newTVar(param)
			}
			new_env[param.Val.(*SemValIdent).Name] = ty_vars[i]
		}
		ty_body, subst := me.inferExpr(val.Body, new_env)
		if ty_body == nil {
			return nil, nil
		}
		targs_fn := make([]*SemType, len(val.Params)+1)
		for i := range val.Params {
			targs_fn[i] = me.applySubstToType(subst, ty_vars[i])
			val.Params[i].Type = targs_fn[i]
		}
		targs_fn[len(targs_fn)-1] = ty_body
		self.Type = semTypeNew(self, MoPrimTypeFunc, targs_fn...)
		return semTypeNew(self, MoPrimTypeFunc, targs_fn...), subst

	case *SemValCall:
		ty_fn_0, s1 := me.inferExpr(val.Callee, env)
		ty_arg, s2 := make([]*SemType, len(val.Args)), make([]semInferSubst, len(val.Args))
		for i, arg := range val.Args {
			ty_arg[i], s2[i] = me.inferExpr(arg, me.applySubstToEnv(s1, env))
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

func (me *semInferCtx) unify(errDst *SemExpr, t1 *SemType, t2 *SemType) (semInferSubst, *Diag) {
	if t1.Prim < 0 {
		return me.tVarBind(int(t1.Prim), t2)
	} else if t2.Prim < 0 {
		return me.tVarBind(int(t2.Prim), t1)
	} else if (t1.Prim == MoPrimTypeFunc) && (t2.Prim == MoPrimTypeFunc) {
		var err *Diag
		s1s := make([]semInferSubst, len(t1.TArgs)-1)
		for i := range t1.TArgs[:len(t1.TArgs)-1] {
			if s1s[i], err = me.unify(errDst, t1.TArgs[i], t2.TArgs[i]); err != nil {
				return nil, err
			}
		}
		var s1 semInferSubst
		if len(s1s) == 0 {
			s1 = semInferSubst{}
		} else {
			s1 = me.composeSubst(s1s[0], s1s[1:]...)
		}
		s2, err := me.unify(errDst, me.applySubstToType(s1, t1.TArgs[len(t1.TArgs)-1]), me.applySubstToType(s1, t2.TArgs[len(t2.TArgs)-1]))
		if err != nil {
			return nil, err
		}
		return me.composeSubst(s1, s2), nil
	} else if t1.Eq(t2) {
		return semInferSubst{}, nil
	}

	if t1.DueTo.isCallArg() {
		errDst = t1.DueTo
	} else if t2.DueTo.isCallArg() {
		errDst = t2.DueTo
	}
	return nil, semTypeErrOn(errDst, t1, t2)
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
