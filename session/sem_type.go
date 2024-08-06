package session

import (
	"strings"

	"atmo/util"
	"atmo/util/sl"
	"atmo/util/str"
)

type SemType struct {
	DueTo  *SemExpr
	Prim   MoValPrimType
	TArgs  sl.Of[*SemType]
	Fields sl.Of[MoValIdent]
}

func (me *SemType) Eq(to *SemType) bool {
	sl_eq := util.If((me.Prim == MoPrimTypeOr), sl.EqAnyOrder, sl.Eq[sl.Of[*SemType]])
	return (me == to) || ((me != nil) && (to != nil) && (me.Prim == to.Prim) && sl_eq(me.TArgs, to.TArgs, (*SemType).Eq))
}

func (me *SemType) IsSubOf(of *SemType) bool {
	switch {
	case (me == nil) || (of == nil):
		return (me == of)
	case (me == of) || (of.Prim == MoPrimTypeAny) || (me.Prim == MoPrimTypeNever):
		return true
	case of.Prim == MoPrimTypeNever:
		return false
	case (len(me.TArgs) == 0) && (len(of.TArgs) == 0):
		return me.Prim == of.Prim
	case (me.Prim == MoPrimTypeObj) && (of.Prim == MoPrimTypeObj):
		for i, targ := range of.TArgs {
			field_name := of.Fields[i]
			if idx := sl.IdxOf(me.Fields, field_name); idx < 0 {
				return false
			} else if !me.TArgs[idx].IsSubOf(targ) {
				return false
			}
		}
		return true
	case (me.Prim == MoPrimTypeTup) && (of.Prim == MoPrimTypeTup):
		for i, targ := range of.TArgs {
			if (i >= len(me.TArgs)) || !me.TArgs[i].IsSubOf(targ) {
				return false
			}
		}
		return true
	case (me.Prim == MoPrimTypeFunc) && (of.Prim == MoPrimTypeFunc) && (len(me.TArgs) == len(of.TArgs)):
		for i, targ := range me.TArgs { // last (return) tArg covariant, the others contravariant:
			if is := util.If((i == (len(me.TArgs) - 1)), targ.IsSubOf(of.TArgs[i]), of.TArgs[i].IsSubOf(targ)); !is {
				return false
			}
		}
	}
	return false
}

func (me *SemType) Sats(expect *SemType) bool {
	switch {
	case me == expect:
		return true
	case (me == nil) || (expect == nil):
		return false
	case expect.Prim == MoPrimTypeAny:
		return true
	case me.Prim == MoPrimTypeOr:
		return sl.All(me.TArgs, func(targ *SemType) bool { return targ.Sats(expect) })
	case expect.Prim == MoPrimTypeOr:
		return sl.Any(expect.TArgs, func(targ *SemType) bool { return me.Sats(targ) })
	case (expect.Prim == MoPrimTypeFunc) && (len(expect.TArgs) == 0): // means "any function" — note a func type with zero args would still have the return t-arg, so 0-arity is an OK sentinel for any-function
		return me.Prim == MoPrimTypeFunc
	}
	sl_eq := util.If((me.Prim == MoPrimTypeOr), sl.EqAnyOrder, sl.Eq[sl.Of[*SemType]])
	return (me.Prim == expect.Prim) && sl_eq(me.TArgs, expect.TArgs, (*SemType).Sats)
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
	} else if w.Len() > 123 { // infinite-type guard
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
	if (me.Prim == MoPrimTypeOr) || (me.Prim == MoPrimTypeAnd) {
		for i := 0; i < me.TArgs.Len(); i++ {
			if t := me.TArgs[i]; t.Prim == MoPrimTypeNever {
				return false
			} else if t.Prim == MoPrimTypeAny {
				*me = *t
				return true
			} else if t.Prim == me.Prim {
				me.TArgs = append(append(me.TArgs[:i], me.TArgs[i+1:]...), t.TArgs...)
				i--
			}
		}
		me.TArgs.EnsureAllUnique((*SemType).Eq)
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
		should := func(t *SemType) bool {
			return (t.DueTo == nil) || (t.DueTo.From == nil) || (t.DueTo.From.SrcFile == nil) || (t.DueTo.From.SrcSpan == nil)
		}
		if use := should(ty); use || sl.Any(ty.TArgs, should) {
			return semTypeNew(util.If(use, dueTo, ty.DueTo), ty.Prim, sl.To(ty.TArgs, func(targ *SemType) *SemType { return semTypeEnsureDueTo(dueTo, targ) })...)
		}
	}
	return ty
}

func semTypeNew(dueTo *SemExpr, prim MoValPrimType, tyArgs ...*SemType) *SemType {
	if sl.Has(tyArgs, nil) {
		return nil
	}
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

func (me *SrcPack) semCheckCount(wantAtLeast int, wantAtMost int, have SemExprs, errDst *SemExpr, forArgs bool) bool {
	if wantAtLeast >= 0 {
		plural := util.If((wantAtLeast <= wantAtMost) && (wantAtLeast != 1), "s", "")
		moniker := util.If(!forArgs, "expression"+plural+" in here", "arg"+plural+" for callee")
		if forArgs && (errDst != nil) {
			if call, _ := errDst.Val.(*SemValCall); (call != nil) && (call.Callee.From != nil) && (call.Callee.From.SrcNode != nil) && (call.Callee.From.SrcNode.Src != "") {
				moniker += " `" + call.Callee.From.SrcNode.Src + "`"
			}
		}
		err_loc := errDst
		if (wantAtMost >= wantAtLeast) && (len(have) > wantAtMost) {
			err_loc = have[wantAtMost]
		}
		if (wantAtLeast == wantAtMost) && (wantAtLeast != len(have)) {
			errDst.ErrAdd(err_loc.From.SrcSpan.newDiagErr(ErrCodeExpectedFoo, str.Fmt("%d %s instead of %d", wantAtLeast, moniker, len(have))))
			return false
		} else if len(have) < wantAtLeast {
			errDst.ErrAdd(err_loc.From.SrcSpan.newDiagErr(ErrCodeExpectedFoo, str.Fmt("at least %d %s instead of %d", wantAtLeast, moniker, len(have))))
			return false
		} else if (wantAtMost > wantAtLeast) && (len(have) > wantAtMost) {
			errDst.ErrAdd(err_loc.From.SrcSpan.newDiagErr(ErrCodeExpectedFoo, str.Fmt("%d to %d %s instead of %d", wantAtLeast, wantAtMost, moniker, len(have))))
			return false
		}
	}
	return true
}

func semCheckIs[T any](equivPrimType MoValPrimType, expr *SemExpr) *T {
	if ret, is := expr.Val.(*T); is {
		return ret
	}
	expr.ErrAdd(expr.From.SrcSpan.newDiagErr(ErrCodeExpectedFoo, str.Fmt("%s here instead of `%s`",
		util.If((equivPrimType < 0), "a comparable value", equivPrimType.Str(true)),
		expr.From.SrcNode.Src)))
	return nil
}

func (me *SrcPack) semCheckTypePrim(expr *SemExpr, dueTo *SemExpr, expect MoValPrimType, arity int) bool {
	if (expr.Type != nil) && (expr.Type.Prim == expect) && ((arity < 0) || (len(expr.Type.TArgs) == arity)) {
		return true
	}
	arity = util.If((arity < 0), 0, arity)
	targs, targ := make([]*SemType, arity), semTypeNew(dueTo, MoPrimTypeAny)
	for i := range targs {
		targs[i] = targ
	}
	return me.semCheckType(expr, semTypeNew(dueTo, expect, targs...))
}

func (me *SrcPack) semCheckType(expr *SemExpr, expect *SemType) bool {
	if expr.Type == nil {
		expr.Type = expect
		if ident, _ := expr.Val.(*SemValIdent); ident != nil {
			if _, entry := expr.Scope.Lookup(ident.Name); entry == nil {
				expr.Type = nil
				return true
			} else {
				me.semScopeEntrySetType(entry, expr)
			}
		}
	} else if !expr.Type.Sats(expect) {
		if !expr.HasErrs() { // dont wanna be too noisy
			expr.ErrAdd(semTypeErr(expr, expect))
		}
		return false
	}
	return true
}

func semTypeErr(expr *SemExpr, expect *SemType) *Diag {
	return semTypeErrOn(expr, expect, expr.Type)
}

func semTypeErrOn(self *SemExpr, expect *SemType, have *SemType) *Diag {
	t1, t2 := expect, have
	dt1, dt2 := expect.DueTo, have.DueTo
	s1, s2 := "`"+t1.String()+"` value", "`"+t2.String()+"` value"
	if t1.Prim != t2.Prim {
		s1, s2 = t1.Prim.Str(true), t2.Prim.Str(true)
	}
	err := self.ErrNew(ErrCodeTypeMismatch, s1, s2)
	err.Rel = srcFileLocs([]string{
		str.Fmt("%s decided by `%s`", s1, dt1.String()),
		str.Fmt("%s decided by `%s`", s2, dt2.String()),
	}, t1.DueTo, t2.DueTo)
	return err
}
