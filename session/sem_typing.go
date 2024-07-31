package session

import (
	"strings"

	"atmo/util/sl"
	"atmo/util/str"
)

type semType interface {
	eq(semType) bool
	str(*strings.Builder)
}

type semTypeCtor struct {
	prim MoValPrimType
	args []semType
}

func (me *semTypeCtor) eq(to semType) bool {
	it, _ := to.(*semTypeCtor)
	return (it != nil) && ((me == it) || ((me.prim == it.prim) && sl.Eq(me.args, it.args, semType.eq)))
}

func (me *semTypeCtor) str(w *strings.Builder) {
	w.WriteString(me.prim.Str(false))
	w.WriteString("<")
	if w.Len() > 256 { // infinite-type guard
		w.WriteString("...>")
		return
	}
	for _, ty := range me.args {
		ty.str(w)
	}
	w.WriteString(">")
}

type semTypeVar struct {
	index int
}

func (me *semTypeVar) eq(to semType) bool {
	it, _ := to.(*semTypeVar)
	return (it != nil) && ((me == it) || (*me == *it))
}

func (me *semTypeVar) str(w *strings.Builder) {
	w.WriteString("<")
	w.WriteString(str.FromInt(me.index))
	w.WriteString(">")
}

func semTypeToString(ty semType) string {
	var buf strings.Builder
	ty.str(&buf)
	return buf.String()
}

var subst []semType

func semTypeUnify(t1 semType, t2 semType, errDst *SemExpr) *Diag {
	tc1, _ := t1.(*semTypeCtor)
	tc2, _ := t2.(*semTypeCtor)
	tv1, _ := t1.(*semTypeVar)
	tv2, _ := t2.(*semTypeVar)
	switch {

	case (tc1 != nil) && (tc2 != nil):
		if (tc1.prim != tc2.prim) || (len(tc1.args) != len(tc2.args)) {
			return errDst.From.SrcSpan.newDiagErr(ErrCodeTypeMismatch, semTypeToString(t1), semTypeToString(t2))
		}
		for i := range tc1.args {
			if err := semTypeUnify(tc1.args[i], tc2.args[i], errDst); err != nil {
				return err
			}
		}

	case (tv1 != nil) && (tv2 != nil) && (tv1.index == tv2.index):
		return nil

	case (tv1 != nil) && (subst[tv1.index] != tv1):
		return semTypeUnify(subst[tv1.index], t2, errDst)

	case (tv2 != nil) && (subst[tv2.index] != tv2):
		return semTypeUnify(t1, subst[tv2.index], errDst)

	case tv1 != nil:
		if semTypeOccursIn(tv1.index, t2) {
			return errDst.From.SrcSpan.newDiagErr(ErrCodeTypeInfinite)
		}
		subst[tv1.index] = t2

	case tv2 != nil:
		if semTypeOccursIn(tv2.index, t1) {
			return errDst.From.SrcSpan.newDiagErr(ErrCodeTypeInfinite)
		}
		subst[tv2.index] = t1

	}

	return nil
}

func semTypeOccursIn(index int, ty semType) bool {
	tv, _ := ty.(*semTypeVar)
	tc, _ := ty.(*semTypeCtor)
	switch {
	case (tv != nil) && !subst[tv.index].eq(ty):
		return semTypeOccursIn(index, subst[tv.index])
	case tv != nil:
		return tv.index == index
	case tc != nil:
		return sl.HasWhere(tc.args, func(it semType) bool { return semTypeOccursIn(index, it) })
	}
	return false
}
