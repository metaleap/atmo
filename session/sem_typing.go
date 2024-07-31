package session

import (
	"strings"

	"atmo/util/sl"
	"atmo/util/str"
)

type semType interface {
	eq(semType) bool
	from() *SemExpr
	str(*strings.Builder)
}

type semTypeCtor struct {
	dueTo *SemExpr
	prim  MoValPrimType
	args  []semType
}
type semTypeVar struct {
	dueTo *SemExpr
	index int
}

func (me *semTypeCtor) eq(to semType) bool {
	it, _ := to.(*semTypeCtor)
	return (it != nil) && ((me == it) || ((me.prim == it.prim) && sl.Eq(me.args, it.args, semType.eq)))
}
func (me *semTypeVar) eq(to semType) bool {
	it, _ := to.(*semTypeVar)
	return (it != nil) && ((me == it) || (*me == *it))
}
func (me *semTypeCtor) from() *SemExpr { return me.dueTo }
func (me *semTypeVar) from() *SemExpr  { return me.dueTo }
func (me *semTypeCtor) str(w *strings.Builder) {
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
			ty.str(w)
		}
		w.WriteByte('>')
	}
}
func (me *semTypeVar) str(w *strings.Builder) {
	w.WriteString(str.FromInt(me.index))
}

func semTypeToString(ty semType) string {
	var buf strings.Builder
	ty.str(&buf)
	return buf.String()
}

type semTypeInfer struct {
	subst []semType
}

func (me *semTypeInfer) unify(t1 semType, t2 semType, errDst *SemExpr) (err *Diag) {
	tc1, _ := t1.(*semTypeCtor)
	tc2, _ := t2.(*semTypeCtor)
	tv1, _ := t1.(*semTypeVar)
	tv2, _ := t2.(*semTypeVar)
	switch {

	case (tc1 != nil) && (tc2 != nil):
		if (tc1.prim != tc2.prim) || (len(tc1.args) != len(tc2.args)) {
			err = errDst.From.SrcSpan.newDiagErr(ErrCodeTypeMismatch, semTypeToString(t1), semTypeToString(t2))
			break
		}
		for i := range tc1.args {
			if err := me.unify(tc1.args[i], tc2.args[i], errDst); err != nil {
				return err
			}
		}

	case (tv1 != nil) && (tv2 != nil) && (tv1.index == tv2.index):
		return

	case (tv1 != nil) && !me.subst[tv1.index].eq(t1):
		return me.unify(me.subst[tv1.index], t2, errDst)

	case (tv2 != nil) && !me.subst[tv2.index].eq(t2):
		return me.unify(t1, me.subst[tv2.index], errDst)

	case tv1 != nil:
		if me.occursIn(tv1.index, t2) {
			err = errDst.From.SrcSpan.newDiagErr(ErrCodeTypeInfinite, semTypeToString(t2))
			break
		}
		me.subst[tv1.index] = t2

	case tv2 != nil:
		if me.occursIn(tv2.index, t1) {
			err = errDst.From.SrcSpan.newDiagErr(ErrCodeTypeInfinite, semTypeToString(t1))
			break
		}
		me.subst[tv2.index] = t1

	}

	if err != nil {
		err.Rel = srcFileLocs([]string{
			str.Fmt("type `%s` decided here", semTypeToString(t1)),
			str.Fmt("type `%s` decided here", semTypeToString(t2)),
		}, t1.from(), t2.from())
	}
	return
}

func (me *semTypeInfer) occursIn(index int, ty semType) bool {
	tv, _ := ty.(*semTypeVar)
	tc, _ := ty.(*semTypeCtor)
	switch {
	case (tv != nil) && !me.subst[tv.index].eq(ty):
		return me.occursIn(index, me.subst[tv.index])
	case tv != nil:
		return tv.index == index
	case tc != nil:
		return sl.HasWhere(tc.args, func(it semType) bool { return me.occursIn(index, it) })
	}
	return false
}

func semTypeNew(dueTo *SemExpr, prim MoValPrimType, args ...semType) semType {
	return &semTypeCtor{dueTo: dueTo, prim: prim, args: args}
}
