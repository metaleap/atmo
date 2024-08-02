package session

import (
	"strings"

	"atmo/util"
	"atmo/util/sl"
	"atmo/util/str"
)

type SemType struct {
	DueTo *SemExpr
	Prim  MoValPrimType
	TArgs sl.Of[*SemType]
}

func (me *SemType) Eq(to *SemType) bool {
	sl_eq := util.If((me.Prim == MoPrimTypeOr), sl.EqOrderless, sl.Eq[sl.Of[*SemType]])
	return (me == to) || ((me != nil) && (to != nil) && (me.Prim == to.Prim) && sl_eq(me.TArgs, to.TArgs, (*SemType).Eq))
}

func (me *SemType) String() (ret string) {
	var buf str.Buf
	me.stringifyTo(&buf)
	if ret = buf.String(); str.Begins(ret, "(") && str.Ends(ret, ")") {
		ret = ret[1 : len(ret)-1]
	}
	return
}
func (me *SemType) stringifyTo(w *strings.Builder) {
	if me == nil {
		w.WriteString("<untypifyable>")
		return
	}
	if w.Len() > 123 { // infinite-type guard
		w.WriteString("..")
		return
	}

	switch {
	case len(me.TArgs) == 0:
		w.WriteString(me.Prim.Str(false))
	case (me.Prim == MoPrimTypeList) && (len(me.TArgs) == 1):
		w.WriteByte('[')
		me.TArgs[0].stringifyTo(w)
		w.WriteByte(']')
	case (me.Prim == MoPrimTypeDict) && (len(me.TArgs) == 2):
		w.WriteByte('{')
		me.TArgs[0].stringifyTo(w)
		w.WriteString(": ")
		me.TArgs[1].stringifyTo(w)
		w.WriteByte('}')
	case (me.Prim == MoPrimTypeFunc) && (len(me.TArgs) > 0):
		w.WriteByte('(')
		for i, targ := range me.TArgs {
			if i > 0 {
				w.WriteString(" ")
			}
			if i == (len(me.TArgs) - 1) {
				w.WriteString("=> ")
			}
			if targ == nil {
				w.WriteString("<NIL?!?!?!TODO>")
			} else {
				targ.stringifyTo(w)
			}
		}
		w.WriteByte(')')
	case (me.Prim == MoPrimTypeOr) && (len(me.TArgs) > 0):
		w.WriteByte('(')
		for i, targ := range me.TArgs {
			if (i > 0) || (len(me.TArgs) == 1) {
				w.WriteString(" | ")
			}
			targ.stringifyTo(w)
		}
		w.WriteByte(')')
	default:
		w.WriteString(me.Prim.Str(false))
		w.WriteByte('<')
		for i, ty := range me.TArgs {
			if i > 0 {
				w.WriteByte(',')
			}
			ty.stringifyTo(w)
		}
		w.WriteByte('>')
	}
}

func (me *SemType) IsAny() bool { return me.Prim == MoPrimTypeAny }

func (me *SemType) normalizeIfAdt() bool {
	if me.Prim == MoPrimTypeOr {
		for i := 0; i < me.TArgs.Len(); i++ {
			if t := me.TArgs[i]; t.Prim == MoPrimTypeAny {
				*me = *t
				return true
			} else if t.Prim == MoPrimTypeOr {
				me.TArgs = append(append(me.TArgs[:i], me.TArgs[i+1:]...), t.TArgs...)
				i--
			}
		}
		me.TArgs.EnsureAllUnique((*SemType).Eq)
		me.TArgs = me.TArgs.Without(func(it *SemType) bool { return it.Prim == MoPrimTypeAny })
		switch len(me.TArgs) {
		case 0:
			return false
		case 1:
			*me = *(me.TArgs[0])
		}
	}
	return true
}

func semTypeEnsureDueTo(dueTo *SemExpr, ty *SemType) *SemType {
	if dueTo != nil {
		nay := func(t *SemType) bool {
			return (t.DueTo == nil) || (t.DueTo.From == nil) || (t.DueTo.From.SrcFile == nil) || (t.DueTo.From.SrcSpan == nil)
		}
		if nah := nay(ty); nah || sl.Any(ty.TArgs, nay) {
			return semTypeNew(util.If(nah, dueTo, ty.DueTo), ty.Prim, sl.To(ty.TArgs, func(targ *SemType) *SemType { return semTypeEnsureDueTo(dueTo, targ) })...)
		}
	}
	return ty
}

func semTypeNew(dueTo *SemExpr, prim MoValPrimType, tyArgs ...*SemType) *SemType {
	ret := &SemType{DueTo: dueTo, Prim: prim, TArgs: sl.To(tyArgs, func(targ *SemType) *SemType { return semTypeEnsureDueTo(dueTo, targ) })}
	if len(tyArgs) > 0 {
		if !ret.normalizeIfAdt() {
			ret = nil
		}
	}
	return ret
}

func semTypeFromMultiple(dueTo *SemExpr, anyIfEmpty bool, ty ...*SemType) *SemType {
	types := (sl.Of[*SemType])(ty)
	use_any := anyIfEmpty && (len(types) == 0)
	types = types.Without(func(t *SemType) bool { return t == nil })
	switch types.EnsureAllUnique((*SemType).Eq); len(types) {
	case 0:
		return util.If(use_any, semTypeNew(dueTo, MoPrimTypeAny), nil)
	case 1:
		return types[0]
	default:
		return semTypeNew(dueTo, MoPrimTypeOr, types...)
	}
}
