package session

import (
	"maps"
	"strings"

	"atmo/util/sl"
	"atmo/util/str"
)

func (me *SrcPack) semInferTypes() {
	env := maps.Clone(semTypingPrimOpsEnv)
	for i, top_expr := range me.Trees.Sem.TopLevel {
		var it semTypeInfer
		top_expr.Type = it.newTypeVar(top_expr)
		it.infer(me, top_expr, env)
		errs := it.solveConstraints(top_expr)
		top_expr.ErrsOwn.Add(errs...)
		it.substExpr(top_expr)
		me.Trees.Sem.TopLevel[i] = top_expr
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
	args  sl.Of[SemType]
}
type semTypeVar struct {
	dueTo *SemExpr
	idx   int
}

func (me *semTypeVar) Eq(to SemType) bool {
	it, _ := to.(*semTypeVar)
	return (it != nil) && ((me == it) || (me.idx == it.idx))
}
func (me *semTypeCtor) Eq(to SemType) bool {
	it, _ := to.(*semTypeCtor)
	return (it != nil) && ((me == it) || ((me.prim == it.prim) && sl.Eq(me.args, it.args, SemType.Eq)))
}
func (me *semTypeVar) From() *SemExpr  { return me.dueTo }
func (me *semTypeCtor) From() *SemExpr { return me.dueTo }
func (me *semTypeVar) Str(w *strings.Builder) {
	w.WriteString("¿")
	w.WriteString(str.FromInt(me.idx))
}
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

func SemTypeToString(ty SemType) string {
	if ty == nil {
		return MoPrimTypeUntyped.Str(false)
	}
	var buf strings.Builder
	ty.Str(&buf)
	return buf.String()
}

type semTypeInfer struct {
	substs      []SemType
	constraints sl.Of[SemTypeConstraint]
}

func (me *semTypeInfer) solveConstraints(errDst *SemExpr) (ret []*Diag) {
	for _, constraint := range me.constraints {
		switch it := constraint.(type) {
		default:
			panic(it)
		case *semTypeConstraintEq:
			if err := me.unify(it.T1, it.T2, errDst); err != nil {
				ret = append(ret, err)
			}
		}
	}
	return
}

func (me *semTypeInfer) substExpr(expr *SemExpr) {
	switch val := expr.Val.(type) {
	case *SemValFunc:
		_ = val
	}
}

func (me *semTypeInfer) substType(ty SemType) SemType {
	tv, _ := ty.(*semTypeVar)
	tc, _ := ty.(*semTypeCtor)
	switch {
	case (tv != nil) && !tv.Eq(me.substs[tv.idx]):
		return me.substType(me.substs[tv.idx])
	case tc != nil:
		return &semTypeCtor{dueTo: tc.dueTo, prim: tc.prim, args: sl.To(tc.args, me.substType)}
	}
	return ty
}

func (me *semTypeInfer) infer(ctx *SrcPack, expr *SemExpr, env map[MoValIdent]SemType) {
	switch val := expr.Val.(type) {
	case *SemValFunc:

	case *SemValCall:
		ty_args := sl.To(val.Args, func(arg *SemExpr) SemType { return me.newTypeVar(arg) })
		ty_fn := semTypeNew(expr, MoPrimTypeFunc, append(ty_args, expr.Type)...)
		val.Callee.Type = ty_fn
		me.infer(ctx, val.Callee, env)
		var idx int
		sl.Each(val.Args, func(arg *SemExpr) { arg.Type = ty_args[idx]; me.infer(ctx, arg, env); idx++ })
	case *SemValIdent:
		ty_ident := env[val.Ident]
		if ty_ident == nil {
			expr.ErrsOwn.Add(expr.From.SrcSpan.newDiagErr(ErrCodeUndefined, val.Ident))
		} else {
			me.constraints.Add(semTypeEq(expr, expr.Type, ty_ident))
		}
	case *SemValScalar:
		new_ty_expr := semTypeNew(expr, val.MoVal.PrimType())
		me.constraints.Add(semTypeEq(expr, expr.Type, new_ty_expr))
	case *SemValList:
		var new_ty_items SemType
		if ty_list := expr.Type.(*semTypeCtor); (ty_list != nil) && (ty_list.prim == MoPrimTypeList) && (len(ty_list.args) == 1) {
			new_ty_items = ty_list.args[0]
		} else {
			new_ty_items = me.newTypeVar(expr)
		}
		new_ty_expr := semTypeNew(expr, MoPrimTypeList, new_ty_items)
		me.constraints.Add(semTypeEq(expr, expr.Type, new_ty_expr))
		sl.Each(val.Items, func(item *SemExpr) { item.Type = new_ty_items; me.infer(ctx, item, env) })
	case *SemValDict:
		expr.ErrsOwn.Add(expr.From.SrcSpan.newDiagErr(ErrCodeAtmoTodo, "semTypeInfer.infer(someDict)"))
	}
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

	case (tv1 != nil) && (tv2 != nil) && (tv1.idx == tv2.idx):
		return

	case (tv1 != nil) && !tv1.Eq(me.substs[tv1.idx]):
		return me.unify(me.substs[tv1.idx], t2, errDst)

	case (tv2 != nil) && !tv2.Eq(me.substs[tv2.idx]):
		return me.unify(t1, me.substs[tv2.idx], errDst)

	case tv1 != nil:
		if me.occursIn(tv1.idx, t2) {
			err = errDst.From.SrcSpan.newDiagErr(ErrCodeTypeInfinite, SemTypeToString(t2))
			break
		}
		me.substs[tv1.idx] = t2

	case tv2 != nil:
		if me.occursIn(tv2.idx, t1) {
			err = errDst.From.SrcSpan.newDiagErr(ErrCodeTypeInfinite, SemTypeToString(t1))
			break
		}
		me.substs[tv2.idx] = t1

	}

	if err != nil {
		from := errDst
		from1, from2 := t1.From(), t2.From()
		if (from1 == nil) || (from2 == nil) {
			if call, _ := from.Val.(*SemValCall); call != nil {
				from = call.Callee
			}
		}
		err.Rel = srcFileLocs([]string{
			str.Fmt("type `%s` decided here", SemTypeToString(t1)),
			str.Fmt("type `%s` decided here", SemTypeToString(t2)),
		}, sl.FirstNonNil(from1, from), sl.FirstNonNil(from2, from))
	}
	return
}

func (me *semTypeInfer) occursIn(index int, ty SemType) bool {
	tv, _ := ty.(*semTypeVar)
	tc, _ := ty.(*semTypeCtor)
	switch {
	case (tv != nil) && !tv.Eq(me.substs[tv.idx]):
		return me.occursIn(index, me.substs[tv.idx])
	case tv != nil:
		return tv.idx == index
	case tc != nil:
		return sl.Any(tc.args, func(tArg SemType) bool { return me.occursIn(index, tArg) })
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
	ret = &semTypeVar{dueTo: dueTo, idx: len(me.substs)}
	me.substs = append(me.substs, ret)
	return
}

type SemTypeConstraint interface {
	isConstraint()
	String() string
}

type semTypeConstraintEq struct {
	dueTo *SemExpr
	T1    SemType
	T2    SemType
}

func (*semTypeConstraintEq) isConstraint() {}
func (me *semTypeConstraintEq) String() string {
	var buf strings.Builder
	me.T1.Str(&buf)
	buf.WriteString("==")
	me.T2.Str(&buf)
	return buf.String()
}
