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
		it.substituteExpr(top_expr, ty)
	}
}

func (me *SemExpr) newUntypable() SemType {
	return semTypeNew(me, MoPrimTypeUntyped)
}

type SemType interface {
	Eq(SemType) bool
	From() *SemExpr
	Str(*strings.Builder)
}

type semTypeCtor struct {
	dueTo *SemExpr
	prim  MoValPrimType
	args  []SemType
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
	constraints []SemTypeConstraint
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

func (me *semTypeInfer) substituteExpr(expr *SemExpr, ty SemType) {
	switch val := expr.Val.(type) {
	case *SemValScalar:
		_ = val
		expr.Type = ty
	case *SemValIdent:
		expr.Type = ty
	case *SemValList:
	case *SemValDict:
	case **SemValFunc:
	case *SemValCall:
	}
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

func (me *semTypeInfer) infer(ctx *SrcPack, expr *SemExpr, env map[MoValIdent]SemType) SemType {
	ty_on_err := expr.newUntypable()
	if expr.Type != nil {
		return expr.Type
	}
	switch val := expr.Val.(type) {
	case *SemValScalar:
		return semTypeNew(expr, val.MoVal.PrimType())
	case *SemValList:
		return semTypeNew(expr, MoPrimTypeList, me.newTypeVar(expr))
	case *SemValDict:
		return semTypeNew(expr, MoPrimTypeDict, me.newTypeVar(expr), me.newTypeVar(expr))
	case *SemValIdent:
		ty := env[val.MoVal]
		if ty == nil {
			ty = ty_on_err
		}
		return ty
	case *SemValFunc:
		own_env := maps.Clone(env)
		param_type_vars := make([]SemType, len(val.Params))
		for i, param := range val.Params {
			param_type_vars[i] = me.newTypeVar(param)
			own_env[param.Val.(*SemValIdent).MoVal] = param_type_vars[i]
		}
		ty_ret := me.infer(ctx, val.Body, own_env)
		return semTypeNew(expr, MoPrimTypeFunc, append(param_type_vars, ty_ret)...)
	case *SemValCall:
		var prim_op func(*SrcPack, *semTypeInfer, *SemExpr, map[MoValIdent]SemType) SemType
		if callee := val.Callee.MaybeIdent(); callee != "" {
			prim_op = semTypingPrimOpsDo[callee]
		}
		if prim_op != nil {
			return prim_op(ctx, me, expr, env)
		} else {
			return me.inferForCallWith(ctx, env, expr, val.Callee, val.Args...)
		}
	}
	return ty_on_err
}

func (me *semTypeInfer) inferForCallWith(ctx *SrcPack, env map[MoValIdent]SemType, call *SemExpr, callee *SemExpr, callArgs ...*SemExpr) SemType {
	ty_callee := me.infer(ctx, callee, env)
	ty_args := make([]SemType, len(callArgs))
	for i, arg := range callArgs {
		ty_args[i] = me.infer(ctx, arg, env)
	}
	ty_ret := me.newTypeVar(call)
	me.constraints = append(me.constraints, &semTypeConstraintEq{dueTo: call,
		T1: ty_callee,
		T2: semTypeNew(call, MoPrimTypeFunc, append(ty_args, ty_ret)...)})
	return ty_ret
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
