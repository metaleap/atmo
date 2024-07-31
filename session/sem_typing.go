package session

import (
	"maps"
	"strings"

	"atmo/util/sl"
	"atmo/util/str"
)

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
	var buf strings.Builder
	ty.Str(&buf)
	return buf.String()
}

type semTypeInfer struct {
	subst       []SemType
	constraints []SemTypeConstraint
}

func (me *semTypeInfer) infer(expr *SemExpr, env map[MoValIdent]SemType) SemType {
	switch val := expr.Val.(type) {
	case *SemValScalar:
		return semTypeNew(expr, val.MoVal.PrimType())
	case *SemValList:
		return semTypeNew(expr, MoPrimTypeList, me.newTypeVar(expr))
	case *SemValDict:
		return semTypeNew(expr, MoPrimTypeDict, me.newTypeVar(expr), me.newTypeVar(expr))
	case *SemValIdent:
		return env[val.MoVal]
	case *SemValFunc:
		own_env := maps.Clone(env)
		param_type_vars := make([]SemType, len(val.Params))
		for i, param := range val.Params {
			param_type_vars[i] = me.newTypeVar(param)
			own_env[param.Val.(*SemValIdent).MoVal] = param_type_vars[i]
		}
		ty_ret := me.infer(val.Body, own_env)
		return semTypeNew(expr, MoPrimTypeFunc, append(param_type_vars, ty_ret)...)
	case *SemValCall:
		switch callee:=val.Callee.MaybeIdent() {
		case moPrimOpFn,moPrimOpMacro:
			panic("new bug intro'd: encountered @fn or @macro call in type-inference")
		case moPrimOpCaseOf:
		case moPrimOpDo:
		case moPrimOpExpand,moPrimOpQQuote,moPrimOpQuote,moPrimOpSpliceUnquote,moPrimOpUnquote:
		case moPrimOpFnCall:
		case moPrimOpSet:
		default:

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
	T1 SemType
	T2 SemType
}