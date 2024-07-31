package session

import (
	"maps"
	"strings"

	"atmo/util/sl"
	"atmo/util/str"
)

func (me *SrcPack) semInferTypes() {
	env := maps.Clone(semTypingPrimOpsEnv)
	for _, top_expr := range me.Trees.Sem.TopLevel {
		var it semTypeInfer
		ty := it.infer(me, top_expr, env)
		if err := it.solveConstraints(top_expr); err != nil {
			top_expr.ErrsOwn.Add(err)
		}
		top_expr.Type = it.substitute(ty)
	}
}

func (me *SemExpr) newUntypable() SemType {
	return semTypeNew(me, MoPrimTypeUntyped)
}

func (me *SemExpr) with(ty SemType, val any) *SemExpr {
	dup := *me
	if val != nil {
		dup.Val = val
	}

	dup.Each(func(it *SemExpr) { it.Parent = &dup })

	var ty_fixup func(SemType) SemType
	ty_fixup = func(ty SemType) SemType {
		switch ty := ty.(type) {
		default:
			panic(ty)
		case *semTypeVar:
			if ty.dueTo == me {
				ty.dueTo = &dup
			}
			return ty
		case *semTypeCtor:
			if ty.dueTo == me {
				ty.dueTo = &dup
			}
			ty.args = sl.To(ty.args, ty_fixup)
			return ty
		}
	}
	dup.Type = ty_fixup(ty)

	return &dup
}

type SemType interface {
	Eq(SemType) bool
	From() *SemExpr
	Str(*strings.Builder)
}

type semTypeCtor struct {
	dueTo *SemExpr
	prim  MoValPrimType
	args  sl.Of[SemType]
}
type semTypeVar struct {
	dueTo *SemExpr
	index int
}

func (me *semTypeCtor) Eq(to SemType) bool {
	it, _ := to.(*semTypeCtor)
	return (it != nil) && ((me == it) || ((me.prim == it.prim) && sl.Eq(me.args, it.args, SemType.Eq)))
}
func (me *semTypeVar) Eq(to SemType) bool {
	it, _ := to.(*semTypeVar)
	return (it != nil) && ((me == it) || (*me == *it))
}
func (me *semTypeCtor) From() *SemExpr { return me.dueTo }
func (me *semTypeVar) From() *SemExpr  { return me.dueTo }
func (me *semTypeCtor) Str(w *strings.Builder) {
	if w.Len() > 123 { // infinite-type guard
		w.WriteString("..")
		return
	}
	w.WriteString(me.prim.Str(false))
	if len(me.args) > 0 {
		w.WriteByte('<')
		for i, ty := range me.args {
			if i > 0 {
				w.WriteByte(',')
			}
			ty.Str(w)
		}
		w.WriteByte('>')
	}
}
func (me *semTypeVar) Str(w *strings.Builder) {
	w.WriteString(str.FromInt(me.index))
}

func SemTypeToString(ty SemType) string {
	if ty == nil {
		return MoPrimTypeUntyped.Str(false)
	}
	var buf strings.Builder
	ty.Str(&buf)
	return buf.String()
}

type semTypeInfer struct {
	subst       []SemType
	constraints sl.Of[SemTypeConstraint]
}

func (me *semTypeInfer) solveConstraints(errDst *SemExpr) *Diag {
	for _, constraint := range me.constraints {
		switch it := constraint.(type) {
		default:
			panic(it)
		case *semTypeConstraintEq:
			return me.unify(it.T1, it.T2, errDst)
		}
	}
	me.constraints = nil
	return nil
}

func (me *semTypeInfer) substitute(ty SemType) SemType {
	switch it := ty.(type) {
	case *semTypeCtor:
		return semTypeNew(it.dueTo, it.prim, sl.To(it.args, me.substitute)...)
	case *semTypeVar:
		if !me.subst[it.index].Eq(ty) {
			return me.substitute(me.subst[it.index])
		}
	}
	return ty
}

func (me *semTypeInfer) infer(ctx *SrcPack, expr *SemExpr, tyExpected SemType, env map[MoValIdent]SemType) *SemExpr {
	ty_on_err := expr.newUntypable()
	switch val := expr.Val.(type) {
	case *SemValScalar:
		me.constraints.Add(semTypeEq(expr, tyExpected, semTypeNew(expr, val.MoVal.PrimType())))
		return expr.with(tyExpected, nil)
	case *SemValList:
		ty_item := val.TItem
		if ty_item == nil {
			ty_item = me.newTypeVar(expr)
		}
		items := sl.To(val.Items, func(item *SemExpr) *SemExpr { return me.infer(ctx, item, ty_item, env) })
		me.constraints.Add(semTypeEq(expr, tyExpected, semTypeNew(expr, MoPrimTypeList, ty_item)))
		return expr.with(tyExpected, &SemValList{TItem: ty_item, Items: SemExprs(items)})
	case *SemValDict:
		ty_key, ty_val := val.TKey, val.TVal
		if ty_key == nil {
			ty_key = me.newTypeVar(expr)
		}
		if ty_val == nil {
			ty_val = me.newTypeVar(expr)
		}
		keys := sl.To(val.Keys, func(k *SemExpr) *SemExpr { return me.infer(ctx, k, ty_key, env) })
		vals := sl.To(val.Keys, func(v *SemExpr) *SemExpr { return me.infer(ctx, v, ty_key, env) })
		me.constraints.Add(semTypeEq(expr, tyExpected, semTypeNew(expr, MoPrimTypeDict, ty_key, ty_val)))
		return expr.with(tyExpected, &SemValDict{TKey: ty_key, TVal: ty_val, Keys: SemExprs(keys), Vals: SemExprs(vals)})
	case *SemValIdent:
		ty_var := env[val.Ident]
		if ty_var == nil {
			expr.ErrsOwn.Add(expr.From.SrcSpan.newDiagErr(ErrCodeUndefined, val.Ident))
			ty_var = ty_on_err
		}
		me.constraints.Add(semTypeEq(expr, tyExpected, ty_var))
		return expr.with(tyExpected, nil)
	case *SemValFunc:
		new_ty_ret := val.TRet
		if new_ty_ret == nil {
			new_ty_ret = me.newTypeVar(expr)
		}
		new_ty_params := sl.To(val.Params, func(p *SemExpr) SemType {
			if p.Type == nil {
				return me.newTypeVar(p)
			}
			return p.Type
		})
		var idx int
		new_params := sl.To(new_ty_params, func(t SemType) *SemExpr {
			p := val.Params[idx]
			idx++
			return p.with(t, nil)
		})
		new_env := maps.Clone(env)
		for _, param := range new_params {
			new_env[param.MaybeIdent()] = param.Type
		}
		new_body := me.infer(ctx, val.Body, new_ty_ret, new_env)
		me.constraints.Add(semTypeEq(expr, tyExpected, semTypeNew(expr, MoPrimTypeFunc, append(new_ty_params, new_ty_ret)...)))
		return expr.with(tyExpected, &SemValFunc{Scope: val.Scope, Params: SemExprs(new_params), Body: new_body, TRet: new_ty_ret, IsMacro: val.IsMacro})
	case *SemValCall:
		var prim_op func(*SrcPack, *semTypeInfer, *SemExpr, SemType, map[MoValIdent]SemType) *SemExpr
		if callee := val.Callee.MaybeIdent(); callee != "" {
			prim_op = semTypingPrimOpsDo[callee]
		}
		if prim_op != nil {
			return prim_op(ctx, me, expr, tyExpected, env)
		} else {
			ty_args := sl.To(val.Args, func(arg *SemExpr) SemType { return me.newTypeVar(arg) })
			ty_fn := semTypeNew(expr, MoPrimTypeFunc, append(ty_args, tyExpected)...)
			var idx int
			new_fn := me.infer(ctx, val.Callee, ty_fn, env)
			new_args := sl.To(ty_args, func(ty SemType) *SemExpr {
				arg := val.Args[idx]
				idx++
				return me.infer(ctx, arg, ty, env)
			})
			return expr.with(tyExpected, &SemValCall{Callee: new_fn, Args: SemExprs(new_args)})
		}
	}
	panic(expr)
}

func (me *semTypeInfer) unify(t1 SemType, t2 SemType, errDst *SemExpr) (err *Diag) {
	tc1, _ := t1.(*semTypeCtor)
	tc2, _ := t2.(*semTypeCtor)
	tv1, _ := t1.(*semTypeVar)
	tv2, _ := t2.(*semTypeVar)
	switch {

	case (tc1 != nil) && (tc2 != nil):
		if (tc1.prim != tc2.prim) || (len(tc1.args) != len(tc2.args)) {
			err = errDst.From.SrcSpan.newDiagErr(ErrCodeTypeMismatch, SemTypeToString(t1), SemTypeToString(t2))
			break
		}
		for i := range tc1.args {
			if err := me.unify(tc1.args[i], tc2.args[i], errDst); err != nil {
				return err
			}
		}

	case (tv1 != nil) && (tv2 != nil) && (tv1.index == tv2.index):
		return

	case (tv1 != nil) && !me.subst[tv1.index].Eq(t1):
		return me.unify(me.subst[tv1.index], t2, errDst)

	case (tv2 != nil) && !me.subst[tv2.index].Eq(t2):
		return me.unify(t1, me.subst[tv2.index], errDst)

	case tv1 != nil:
		if me.occursIn(tv1.index, t2) {
			err = errDst.From.SrcSpan.newDiagErr(ErrCodeTypeInfinite, SemTypeToString(t2))
			break
		}
		me.subst[tv1.index] = t2

	case tv2 != nil:
		if me.occursIn(tv2.index, t1) {
			err = errDst.From.SrcSpan.newDiagErr(ErrCodeTypeInfinite, SemTypeToString(t1))
			break
		}
		me.subst[tv2.index] = t1

	}

	if err != nil {
		err.Rel = srcFileLocs([]string{
			str.Fmt("type `%s` decided here", SemTypeToString(t1)),
			str.Fmt("type `%s` decided here", SemTypeToString(t2)),
		}, t1.From(), t2.From())
	}
	return
}

func (me *semTypeInfer) occursIn(index int, ty SemType) bool {
	tv, _ := ty.(*semTypeVar)
	tc, _ := ty.(*semTypeCtor)
	switch {
	case (tv != nil) && !me.subst[tv.index].Eq(ty):
		return me.occursIn(index, me.subst[tv.index])
	case tv != nil:
		return tv.index == index
	case tc != nil:
		return sl.HasWhere(tc.args, func(it SemType) bool { return me.occursIn(index, it) })
	}
	return false
}

func semTypeEq(dueTo *SemExpr, t1 SemType, t2 SemType) SemTypeConstraint {
	return &semTypeConstraintEq{dueTo: dueTo, T1: t1, T2: t2}
}
func semTypeNew(dueTo *SemExpr, prim MoValPrimType, args ...SemType) SemType {
	return &semTypeCtor{dueTo: dueTo, prim: prim, args: args}
}
func (me *semTypeInfer) newTypeVar(dueTo *SemExpr) (ret SemType) {
	ret = &semTypeVar{dueTo: dueTo, index: len(me.subst)}
	me.subst = append(me.subst, ret)
	return
}

type SemTypeConstraint interface{}

type semTypeConstraintEq struct {
	dueTo *SemExpr
	T1    SemType
	T2    SemType
}
