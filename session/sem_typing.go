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
		errs := it.solveConstraints()
		top_expr.ErrsOwn.Add(errs...)
		it.substExpr(top_expr)
		me.Trees.Sem.TopLevel[i] = top_expr
	}
}

func (me *SemExpr) newUntyped() SemType {
	return semTypeNew(me, MoPrimTypeAny)
}

type SemType interface {
	Eq(SemType) bool
	From() *SemExpr
	Str(*strings.Builder)
}

type semTypeCtor struct {
	dueTo  *SemExpr
	prim   MoValPrimType
	tyArgs sl.Of[SemType]
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
	return (it != nil) && ((me == it) || ((me.prim == it.prim) && sl.Eq(me.tyArgs, it.tyArgs, SemType.Eq)))
}
func (me *semTypeVar) From() *SemExpr  { return me.dueTo }
func (me *semTypeCtor) From() *SemExpr { return me.dueTo }
func (me *semTypeVar) Str(w *strings.Builder) {
	w.WriteByte('T')
	w.WriteString(str.FromInt(me.idx))
}
func (me *semTypeCtor) Str(w *strings.Builder) {
	if w.Len() > 123 { // infinite-type guard
		w.WriteString("..")
		return
	}

	switch {
	case len(me.tyArgs) == 0:
		w.WriteString(me.prim.Str(false))
	case (me.prim == MoPrimTypeList) && (len(me.tyArgs) == 1):
		w.WriteByte('[')
		me.tyArgs[0].Str(w)
		w.WriteByte(']')
	case (me.prim == MoPrimTypeDict) && (len(me.tyArgs) == 2):
		w.WriteByte('{')
		me.tyArgs[0].Str(w)
		w.WriteByte(':')
		me.tyArgs[1].Str(w)
		w.WriteByte('}')
	case (me.prim == MoPrimTypeFunc) && (len(me.tyArgs) > 0):
		w.WriteByte('(')
		for i, targ := range me.tyArgs {
			if (i > 0) || (len(me.tyArgs) == 1) {
				w.WriteString("→")
			}
			targ.Str(w)
		}
		w.WriteByte(')')
	case (me.prim == MoPrimTypeFunc) && (len(me.tyArgs) > 0):
		w.WriteByte('(')
		for i, targ := range me.tyArgs {
			if (i > 0) || (len(me.tyArgs) == 1) {
				w.WriteString(" | ")
			}
			targ.Str(w)
		}
		w.WriteByte(')')
	default:
		w.WriteString(me.prim.Str(false))
		w.WriteByte('<')
		for i, ty := range me.tyArgs {
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
		return MoPrimTypeAny.Str(false)
	}
	var buf strings.Builder
	ty.Str(&buf)
	return buf.String()
}

type semTypeInfer struct {
	substs      []SemType
	constraints sl.Of[SemTypeConstraint]
}

func (me *semTypeInfer) solveConstraints() (ret []*Diag) {
	for _, constraint := range me.constraints {
		switch it := constraint.(type) {
		default:
			panic(it)
		case *semTypeConstraintEq:
			if err := me.unify(it.T1, it.T2, it.dueTo); err != nil {
				ret = append(ret, err)
			}
		}
	}
	return
}

func (me *semTypeInfer) substExpr(expr *SemExpr) {
	switch val := expr.Val.(type) {
	case *SemValFunc:
		var ty_ret SemType
		if ty_fn, _ := expr.Type.(*semTypeCtor); (ty_fn != nil) && (ty_fn.prim == MoPrimTypeFunc) && (len(ty_fn.tyArgs) == (1 + len(val.Params))) {
			ty_ret = ty_fn.tyArgs[len(ty_fn.tyArgs)-1]
		}
		sl.Each(val.Params, func(p *SemExpr) {
			if p.Type != nil {
				p.Type = me.substType(p.Type)
			}
		})
		val.Body.Type = ty_ret
		me.substExpr(val.Body)
		expr.Type = semTypeNew(expr, MoPrimTypeFunc, append(sl.To(val.Params, func(p *SemExpr) SemType { return p.Type }), ty_ret)...)
	case *SemValCall:
		me.substExpr(val.Callee)
		sl.Each(val.Args, me.substExpr)
		if ty_fn, _ := val.Callee.Type.(*semTypeCtor); (ty_fn != nil) && (ty_fn.prim == MoPrimTypeFunc) && ((len(ty_fn.tyArgs)) == (1 + len(val.Args))) {
			expr.Type = ty_fn.tyArgs[len(ty_fn.tyArgs)-1]
		}
	case *SemValList:
		sl.Each(val.Items, me.substExpr)
		var ty_item SemType
		for _, item := range val.Items {
			if ty_item == nil {
				ty_item = item.Type
			} else if !ty_item.Eq(item.Type) {
				// no error reporting needed, unify will have done it
				ty_item = nil
				break
			}
		}
		if ty_item == nil {
			ty_item = expr.newUntyped()
		}
		expr.Type = semTypeNew(expr, MoPrimTypeList, ty_item)
	case *SemValDict:
		expr.ErrsOwn.Add(expr.From.SrcSpan.newDiagErr(ErrCodeAtmoTodo, "semTypeInfer.substExpr(someDict)"))
	}
}

func (me *semTypeInfer) substType(ty SemType) SemType {
	tv, _ := ty.(*semTypeVar)
	tc, _ := ty.(*semTypeCtor)
	switch {
	case (tv != nil) && !tv.Eq(me.substs[tv.idx]):
		return me.substType(me.substs[tv.idx])
	case tc != nil:
		return &semTypeCtor{dueTo: tc.dueTo, prim: tc.prim, tyArgs: sl.To(tc.tyArgs, me.substType)}
	}
	return ty
}

func (me *semTypeInfer) infer(ctx *SrcPack, expr *SemExpr, env map[MoValIdent]SemType) {
	switch val := expr.Val.(type) {
	case *SemValFunc:
		var new_ty_ret SemType
		if ty_fn, _ := expr.Type.(*semTypeCtor); (ty_fn != nil) && (ty_fn.prim == MoPrimTypeFunc) && (len(ty_fn.tyArgs) == (1 + len(val.Params))) {
			new_ty_ret = ty_fn.tyArgs[len(ty_fn.tyArgs)-1]
		} else {
			new_ty_ret = me.newTypeVar(expr)
		}
		new_env := maps.Clone(env)
		sl.Each(val.Params, func(p *SemExpr) {
			if p.Type == nil {
				p.Type = me.newTypeVar(p)
			}
			new_env[p.Val.(*SemValIdent).Ident] = p.Type
		})
		val.Body.Type = new_ty_ret
		me.infer(ctx, val.Body, new_env)
		me.constraints.Add(semTypeEq(expr, expr.Type, semTypeNew(expr, MoPrimTypeFunc, append(sl.To(val.Params, func(p *SemExpr) SemType { return p.Type }), new_ty_ret)...)))
	case *SemValCall:
		ty_args := sl.To(val.Args, func(arg *SemExpr) SemType { return me.newTypeVar(arg) })
		ty_fn := semTypeNew(val.Callee, MoPrimTypeFunc, append(ty_args, expr.Type)...)
		val.Callee.Type = ty_fn

		var prim_op func(*SrcPack, *semTypeInfer, *SemExpr, map[MoValIdent]SemType)
		if ident := val.Callee.MaybeIdent(); ident != "" {
			if prim_op = semTypingPrimOpsDo[ident]; prim_op == nil {
				prim_op = semTypingPrimFnsDo[ident]
			}
		}
		if prim_op != nil {
			prim_op(ctx, me, expr, env)
		} else {
			me.infer(ctx, val.Callee, env)
			var idx int
			sl.Each(val.Args, func(arg *SemExpr) { arg.Type = ty_args[idx]; me.infer(ctx, arg, env); idx++ })
		}
	case *SemValIdent:
		ty_ident := env[val.Ident]
		if ty_ident == nil {
			expr.ErrsOwn.Add(expr.From.SrcSpan.newDiagErr(ErrCodeUndefined, val.Ident))
		} else {
			if ty_ident.From() == nil { // for idents referencing the built-in prim-ops
				ty_ident = semTypeEnsureDueTo(expr, ty_ident)
			}
			me.constraints.Add(semTypeEq(expr, expr.Type, ty_ident))
			expr.Type = ty_ident
		}
	case *SemValScalar:
		new_ty_expr := semTypeNew(expr, val.MoVal.PrimType())
		me.constraints.Add(semTypeEq(expr, expr.Type, new_ty_expr))
		expr.Type = new_ty_expr
	case *SemValList:
		var new_ty_items SemType
		if ty_list, _ := expr.Type.(*semTypeCtor); (ty_list != nil) && (ty_list.prim == MoPrimTypeList) && (len(ty_list.tyArgs) == 1) {
			new_ty_items = ty_list.tyArgs[0]
		} else {
			new_ty_items = me.newTypeVar(expr)
		}
		new_ty_expr := semTypeNew(expr, MoPrimTypeList, new_ty_items)
		sl.Each(val.Items, func(item *SemExpr) { item.Type = new_ty_items; me.infer(ctx, item, env) })
		// { // technically superfluous block. just for err-msg UX purposes so we can see "@Foo vs. [@Bar]" rather than "@Foo vs [T4]" or some such
		// 	var ty_item SemType
		// 	for _, item := range val.Items {
		// 		if ty_item == nil {
		// 			ty_item = item.Type
		// 		} else if !ty_item.Eq(item.Type) {
		// 			ty_item = nil
		// 			break
		// 		}
		// 	}
		// 	if tc, _ := ty_item.(*semTypeCtor); tc != nil {
		// 		new_ty_expr = semTypeNew(expr, MoPrimTypeList, ty_item)
		// 	}
		// }
		me.constraints.Add(semTypeEq(expr, expr.Type, new_ty_expr))
		expr.Type = new_ty_expr
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
		if (tc1.prim != tc2.prim) || (len(tc1.tyArgs) != len(tc2.tyArgs)) {
			err = errDst.From.SrcSpan.newDiagErr(ErrCodeTypeMismatch, SemTypeToString(t1), SemTypeToString(t2))
			break
		}
		for i := range tc1.tyArgs {
			if err = me.unify(tc1.tyArgs[i], tc2.tyArgs[i], errDst); err != nil {
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
	case (tv != nil) && !tv.Eq(me.substs[tv.idx]):
		return me.occursIn(index, me.substs[tv.idx])
	case tv != nil:
		return tv.idx == index
	case tc != nil:
		return sl.Any(tc.tyArgs, func(tArg SemType) bool { return me.occursIn(index, tArg) })
	}
	return false
}

func semTypeEq(dueTo *SemExpr, t1 SemType, t2 SemType) SemTypeConstraint {
	return &semTypeConstraintEq{dueTo: dueTo, T1: t1, T2: t2}
}
func semTypeNew(dueTo *SemExpr, prim MoValPrimType, tyArgs ...SemType) SemType {
	ret := &semTypeCtor{dueTo: dueTo, prim: prim, tyArgs: sl.To(tyArgs, func(targ SemType) SemType { return semTypeEnsureDueTo(dueTo, targ) })}
	return ret
}
func (me *semTypeInfer) newTypeVar(dueTo *SemExpr) (ret SemType) {
	ret = &semTypeVar{dueTo: dueTo, idx: len(me.substs)}
	me.substs = append(me.substs, ret)
	return
}

func semTypeEnsureDueTo(dueTo *SemExpr, ty SemType) SemType {
	switch ty := ty.(type) {
	case *semTypeCtor:
		if (ty.dueTo == nil) || sl.Any(ty.tyArgs, func(targ SemType) bool { return targ.From() == nil }) {
			return semTypeNew(dueTo, ty.prim, sl.To(ty.tyArgs, func(targ SemType) SemType { return semTypeEnsureDueTo(dueTo, targ) })...)
		}
	case *semTypeVar:
		if ty.dueTo == nil {
			ty.dueTo = dueTo
		}
	}
	return ty
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

func (me *semTypeCtor) ensure(ty SemType) {
	if me.Eq(ty) {
		return
	} else if me.prim != MoPrimTypeOr {
		*me = *(semTypeNew(me.dueTo, MoPrimTypeOr, me, ty).(*semTypeCtor))
	} else if !sl.Any(me.tyArgs, func(targ SemType) bool { return targ.Eq(ty) }) {
		me.tyArgs = append(me.tyArgs, ty)
	}
}
