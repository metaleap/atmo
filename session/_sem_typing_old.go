package session

/*
import (
	"maps"
	"strings"

	"atmo/util"
	"atmo/util/sl"
	"atmo/util/str"
)

var (
	semTypingPrimOpsEnv map[MoValIdent]OldSemType
	semTypingPrimOpsDo  map[MoValIdent]func(*SrcPack, *oldSemTypeInfer, *SemExpr, map[MoValIdent]OldSemType)
	semTypingPrimFnsDo  map[MoValIdent]func(*SrcPack, *oldSemTypeInfer, *SemExpr, map[MoValIdent]OldSemType)
)

func (me *SrcPack) oldSemInferTypes() {
	env := maps.Clone(semTypingPrimOpsEnv)
	for i, top_expr := range me.Trees.Sem.TopLevel {
		var it oldSemTypeInfer
		top_expr.Type = it.newTypeVar(top_expr)
		it.infer(me, top_expr, env)
		errs := it.solveConstraints()
		top_expr.ErrAdd(errs...)
		it.substExpr(top_expr)
		me.Trees.Sem.TopLevel[i] = top_expr
	}
}

type OldSemType interface {
	Eq(OldSemType) bool
	From() *SemExpr
	Str(*strings.Builder)
}

type oldSemTypeCtor struct {
	dueTo  *SemExpr
	prim   MoValPrimType
	tyArgs sl.Of[OldSemType]
}
type oldSemTypeVar struct {
	dueTo *SemExpr
	idx   int
}

func (me *oldSemTypeVar) Eq(to OldSemType) bool {
	it, _ := to.(*oldSemTypeVar)
	return (me == it) || ((me != nil) && (it != nil) && (me.idx == it.idx))
}
func (me *oldSemTypeCtor) Eq(to OldSemType) bool {
	it, _ := to.(*oldSemTypeCtor)
	return (me == it) || ((me != nil) && (it != nil) && (me.prim == it.prim) && sl.Eq(me.tyArgs, it.tyArgs, OldSemType.Eq))
}
func (me *oldSemTypeVar) From() *SemExpr  { return me.dueTo }
func (me *oldSemTypeCtor) From() *SemExpr { return me.dueTo }
func (me *oldSemTypeVar) Str(w *strings.Builder) {
	w.WriteByte('T')
	w.WriteString(str.FromInt(me.idx))
}
func (me *oldSemTypeCtor) Str(w *strings.Builder) {
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
		w.WriteString(": ")
		me.tyArgs[1].Str(w)
		w.WriteByte('}')
	case (me.prim == MoPrimTypeFunc) && (len(me.tyArgs) > 0):
		w.WriteByte('(')
		for i, targ := range me.tyArgs {
			if i > 0 {
				w.WriteString(" ")
			}
			if i == (len(me.tyArgs) - 1) {
				w.WriteString("=> ")
			}
			if targ == nil {
				w.WriteString("<NIL?!?!?!>")
			} else {
				targ.Str(w)
			}
		}
		w.WriteByte(')')
	case (me.prim == MoPrimTypeOr) && (len(me.tyArgs) > 0):
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

func OldSemTypeToString(ty OldSemType) string {
	if ty == nil {
		return "<untypifyable>"
	}
	var buf strings.Builder
	ty.Str(&buf)
	return buf.String()
}

type oldSemTypeInfer struct {
	substs      []OldSemType
	constraints sl.Of[OldSemTypeConstraint]
}

func (me *oldSemTypeInfer) solveConstraints() (ret []*Diag) {
	for _, constraint := range me.constraints {
		switch it := constraint.(type) {
		default:
			panic(it)
		case *oldSemTypeConstraintEq:
			if err := me.unify(it.T1, it.T2, it.dueTo); err != nil {
				ret = append(ret, err)
			}
		}
	}
	return
}

func (me *oldSemTypeInfer) substExpr(expr *SemExpr) {
	switch val := expr.Val.(type) {
	case *SemValFunc:
		var ty_ret OldSemType
		if ty_fn, _ := expr.Type.(*oldSemTypeCtor); (ty_fn != nil) && (ty_fn.prim == MoPrimTypeFunc) && (len(ty_fn.tyArgs) == (1 + len(val.Params))) {
			ty_ret = ty_fn.tyArgs[len(ty_fn.tyArgs)-1]
		}
		sl.Each(val.Params, func(p *SemExpr) {
			if p.Type != nil {
				p.Type = me.substType(p.Type)
			}
		})
		val.Body.Type = ty_ret
		me.substExpr(val.Body)
		expr.Type = oldSemTypeNew(expr, MoPrimTypeFunc, append(sl.To(val.Params, func(p *SemExpr) OldSemType { return p.Type }), ty_ret)...)
	case *SemValCall:
		me.substExpr(val.Callee)
		sl.Each(val.Args, me.substExpr)
		if ty_fn, _ := val.Callee.Type.(*oldSemTypeCtor); (ty_fn != nil) && (ty_fn.prim == MoPrimTypeFunc) && ((len(ty_fn.tyArgs)) == (1 + len(val.Args))) {
			expr.Type = ty_fn.tyArgs[len(ty_fn.tyArgs)-1]
		}
	case *SemValList:
		sl.Each(val.Items, me.substExpr)
		var ty_item OldSemType
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
			ty_item = oldSemTypeNew(expr, MoPrimTypeAny)
		}
		expr.Type = oldSemTypeNew(expr, MoPrimTypeList, ty_item)
	case *SemValDict:
		expr.ErrAdd(expr.From.SrcSpan.newDiagErr(ErrCodeAtmoTodo, "semTypeInfer.substExpr(someDict)"))
	}
}

func (me *oldSemTypeInfer) substType(ty OldSemType) OldSemType {
	tv, _ := ty.(*oldSemTypeVar)
	tc, _ := ty.(*oldSemTypeCtor)
	switch {
	case (tv != nil) && !tv.Eq(me.substs[tv.idx]):
		return me.substType(me.substs[tv.idx])
	case tc != nil:
		return &oldSemTypeCtor{dueTo: tc.dueTo, prim: tc.prim, tyArgs: sl.To(tc.tyArgs, me.substType)}
	}
	return ty
}

func (me *oldSemTypeInfer) infer(ctx *SrcPack, expr *SemExpr, env map[MoValIdent]OldSemType) {
	switch val := expr.Val.(type) {
	case *SemValFunc:
		var new_ty_ret OldSemType
		if ty_fn, _ := expr.Type.(*oldSemTypeCtor); (ty_fn != nil) && (ty_fn.prim == MoPrimTypeFunc) && (len(ty_fn.tyArgs) == (1 + len(val.Params))) {
			new_ty_ret = ty_fn.tyArgs[len(ty_fn.tyArgs)-1]
		} else {
			new_ty_ret = me.newTypeVar(expr)
		}
		new_env := maps.Clone(env)
		sl.Each(val.Params, func(p *SemExpr) {
			if p.Type == nil {
				p.Type = me.newTypeVar(p)
			}
			new_env[p.Val.(*SemValIdent).Name] = p.Type
		})
		val.Body.Type = new_ty_ret
		me.infer(ctx, val.Body, new_env)
		me.constraints.Add(oldSemTypeEq(expr, expr.Type, oldSemTypeNew(expr, MoPrimTypeFunc, append(sl.To(val.Params, func(p *SemExpr) OldSemType { return p.Type }), new_ty_ret)...)))
	case *SemValCall:
		ty_args := sl.To(val.Args, func(arg *SemExpr) OldSemType { return me.newTypeVar(arg) })
		ty_fn := oldSemTypeNew(val.Callee, MoPrimTypeFunc, append(ty_args, expr.Type)...)
		val.Callee.Type = ty_fn

		var prim_op func(*SrcPack, *oldSemTypeInfer, *SemExpr, map[MoValIdent]OldSemType)
		if ident := val.Callee.MaybeIdent(false); ident != "" {
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
		ty_ident := env[val.Name]
		if ty_ident == nil {
			expr.ErrAdd(expr.From.SrcSpan.newDiagErr(ErrCodeUndefined, val.Name))
		} else {
			if ty_ident.From() == nil { // for idents referencing the built-in prim-ops
				ty_ident = oldSemTypeEnsureDueTo(expr, ty_ident)
			}
			me.constraints.Add(oldSemTypeEq(expr, expr.Type, ty_ident))
			expr.Type = ty_ident
		}
	case *SemValScalar:
		new_ty_expr := oldSemTypeNew(expr, val.Value.PrimType())
		me.constraints.Add(oldSemTypeEq(expr, expr.Type, new_ty_expr))
		expr.Type = new_ty_expr
	case *SemValList:
		var new_ty_items OldSemType
		if ty_list, _ := expr.Type.(*oldSemTypeCtor); (ty_list != nil) && (ty_list.prim == MoPrimTypeList) && (len(ty_list.tyArgs) == 1) {
			new_ty_items = ty_list.tyArgs[0]
		} else {
			new_ty_items = me.newTypeVar(expr)
		}
		new_ty_expr := oldSemTypeNew(expr, MoPrimTypeList, new_ty_items)
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
		me.constraints.Add(oldSemTypeEq(expr, expr.Type, new_ty_expr))
		expr.Type = new_ty_expr
	case *SemValDict:
		expr.ErrAdd(expr.From.SrcSpan.newDiagErr(ErrCodeAtmoTodo, "semTypeInfer.infer(someDict)"))
	}
}

func (me *oldSemTypeInfer) unify(t1 OldSemType, t2 OldSemType, errDst *SemExpr) (err *Diag) {
	tc1, _ := t1.(*oldSemTypeCtor)
	tc2, _ := t2.(*oldSemTypeCtor)
	tv1, _ := t1.(*oldSemTypeVar)
	tv2, _ := t2.(*oldSemTypeVar)
	switch {

	case (tc1 != nil) && (tc2 != nil):
		if (tc1.prim != tc2.prim) || (len(tc1.tyArgs) != len(tc2.tyArgs)) {
			err = errDst.From.SrcSpan.newDiagErr(ErrCodeTypeMismatch, OldSemTypeToString(t1), OldSemTypeToString(t2))
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
			err = errDst.From.SrcSpan.newDiagErr(ErrCodeTypeInfinite, OldSemTypeToString(t2))
			break
		}
		me.substs[tv1.idx] = t2

	case tv2 != nil:
		if me.occursIn(tv2.idx, t1) {
			err = errDst.From.SrcSpan.newDiagErr(ErrCodeTypeInfinite, OldSemTypeToString(t1))
			break
		}
		me.substs[tv2.idx] = t1

	}

	if err != nil {
		err.Rel = srcFileLocs([]string{
			str.Fmt("type `%s` decided here", OldSemTypeToString(t1)),
			str.Fmt("type `%s` decided here", OldSemTypeToString(t2)),
		}, t1.From(), t2.From())
	}
	return
}

func (me *oldSemTypeInfer) occursIn(index int, ty OldSemType) bool {
	tv, _ := ty.(*oldSemTypeVar)
	tc, _ := ty.(*oldSemTypeCtor)
	switch {
	case (tv != nil) && !tv.Eq(me.substs[tv.idx]):
		return me.occursIn(index, me.substs[tv.idx])
	case tv != nil:
		return tv.idx == index
	case tc != nil:
		return sl.Any(tc.tyArgs, func(tArg OldSemType) bool { return me.occursIn(index, tArg) })
	}
	return false
}

func oldSemTypeEq(dueTo *SemExpr, t1 OldSemType, t2 OldSemType) OldSemTypeConstraint {
	return &oldSemTypeConstraintEq{dueTo: dueTo, T1: t1, T2: t2}
}
func oldSemTypeNew(dueTo *SemExpr, prim MoValPrimType, tyArgs ...OldSemType) OldSemType {
	ret := &oldSemTypeCtor{dueTo: dueTo, prim: prim, tyArgs: sl.To(tyArgs, func(targ OldSemType) OldSemType { return oldSemTypeEnsureDueTo(dueTo, targ) })}
	if len(tyArgs) > 0 {
		if !ret.normalizeIfAdt() {
			ret = nil
		}
	}
	return ret
}
func (me *oldSemTypeInfer) newTypeVar(dueTo *SemExpr) (ret OldSemType) {
	ret = &oldSemTypeVar{dueTo: dueTo, idx: len(me.substs)}
	me.substs = append(me.substs, ret)
	return
}

func oldSemTypeEnsureDueTo(dueTo *SemExpr, ty OldSemType) OldSemType {
	if dueTo != nil {
		nay := func(expr *SemExpr) bool {
			return (expr == nil) || (expr.From == nil) || (expr.From.SrcFile == nil) || (expr.From.SrcSpan == nil)
		}
		switch ty := ty.(type) {
		case *oldSemTypeCtor:
			if nah := nay(ty.dueTo); nah || sl.Any(ty.tyArgs, func(targ OldSemType) bool { return nay(targ.From()) }) {
				return oldSemTypeNew(util.If(nah, dueTo, ty.dueTo), ty.prim, sl.To(ty.tyArgs, func(targ OldSemType) OldSemType { return oldSemTypeEnsureDueTo(dueTo, targ) })...)
			}
		case *oldSemTypeVar:
			if nay(ty.dueTo) {
				ty.dueTo = dueTo
			}
		}
	}
	return ty
}

type OldSemTypeConstraint interface {
	isConstraint()
	String() string
}

type oldSemTypeConstraintEq struct {
	dueTo *SemExpr
	T1    OldSemType
	T2    OldSemType
}

func (*oldSemTypeConstraintEq) isConstraint() {}
func (me *oldSemTypeConstraintEq) String() string {
	var buf strings.Builder
	me.T1.Str(&buf)
	buf.WriteString("==")
	me.T2.Str(&buf)
	return buf.String()
}

func (me *oldSemTypeCtor) normalizeIfAdt() bool {
	if me.prim == MoPrimTypeOr {
		for i := 0; i < me.tyArgs.Len(); i++ {
			if t := me.tyArgs[i].(*oldSemTypeCtor); t.prim == MoPrimTypeOr {
				me.tyArgs = append(append(me.tyArgs[:i], me.tyArgs[i+1:]...), t.tyArgs...)
				i--
			}
		}
		me.tyArgs.EnsureAllUnique(OldSemType.Eq)
		me.tyArgs = me.tyArgs.Without(func(it OldSemType) bool { return it.(*oldSemTypeCtor).prim == MoPrimTypeAny })
		switch len(me.tyArgs) {
		case 0:
			return false
		case 1:
			*me = *(me.tyArgs[0].(*oldSemTypeCtor))
		}
	}
	return true
}

func oldSemTypeFromMultiple(dueTo *SemExpr, anyIfEmpty bool, ty ...OldSemType) OldSemType {
	types := (sl.Of[OldSemType])(ty)
	use_any := anyIfEmpty && (len(types) == 0)
	types = types.Without(func(t OldSemType) bool { return t == nil })
	switch types.EnsureAllUnique(OldSemType.Eq); len(types) {
	case 0:
		return util.If(use_any, oldSemTypeNew(dueTo, MoPrimTypeAny), nil)
	case 1:
		return types[0]
	default:
		return oldSemTypeNew(dueTo, MoPrimTypeOr, types...)
	}
}
*/
