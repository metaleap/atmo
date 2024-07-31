package session

import (
	"atmo/util/str"
	"strings"
)

type semType interface {
	str(Writer)
}

type semTypeCtor struct {
	prim MoValPrimType
	args []semType
}

func (me *semTypeCtor) str(w Writer) {
	w.WriteString(me.prim.Str(false))
	w.WriteString("<")
	for _, ty := range me.args {
		ty.str(w)
	}
	w.WriteString(">")
}

type semTypeVar struct {
	index int
}

func (me *semTypeVar) str(w Writer) {
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
			return errDst.From.SrcSpan.newDiagErr(ErrCodeAtmoTodo, "occurs-in")
		}
		subst[tv1.index] = t2

	case tv2 != nil:
		if semTypeOccursIn(tv2.index, t1) {
			return errDst.From.SrcSpan.newDiagErr(ErrCodeAtmoTodo, "occurs-in")
		}
		subst[tv2.index] = t1

	}

	return nil
}

func semTypeOccursIn(int, semType) bool {
	return false
}
