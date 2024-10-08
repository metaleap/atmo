package session

import (
	"slices"
	"strings"

	"atmo/util"
	"atmo/util/sl"
	"atmo/util/str"
)

type SemType struct {
	DueTo     *SemExpr
	Prim      MoValPrimType
	TArgs     sl.Of[*SemType]
	Fields    sl.Of[MoValIdent]
	Singleton MoVal
}

func (me *SemType) checkIsPrimElseErrOn(expectedTyDueTo *SemExpr, errOwner *SemExpr, errSpan *SemExpr, wantPrim MoValPrimType, wantArity int) bool {
	if ty_expect := semTypeNew(expectedTyDueTo, wantPrim, sl.Repeat(wantArity, semTypeNew(expectedTyDueTo, MoPrimTypeAny))...); (me == nil) || !me.IsSubTypeOf(ty_expect) {
		if me != nil {
			errOwner.ErrAdd(semTypeErrOn(errSpan, ty_expect, me))
		}
		return false
	}
	return true
}

func (me *SemType) Eq(to *SemType) bool {
	sl_eq := util.If((me.Prim == MoPrimTypeOr), sl.EqAnyOrder, sl.Eq[sl.Of[*SemType]])
	return (me == to) || ((me != nil) && (to != nil) && (me.Prim == to.Prim) && (me.Singleton == to.Singleton) &&
		sl.Equal(me.Fields, to.Fields) && sl_eq(me.TArgs, to.TArgs, (*SemType).Eq))
}

func (me *SemType) fieldNamesIfObj(quoted bool) string {
	return str.Join(sl.To(me.Fields, func(it MoValIdent) string {
		return string(util.If(quoted, "`"+it+"`", it))
	}), ",")
}

func (me *SemType) hasSingletons() bool {
	return (me == nil) || (me.Singleton != nil) || sl.Any(me.TArgs, (*SemType).hasSingletons)
}

func (me *SemType) isTruthy() bool {
	switch me.Prim {
	case MoPrimTypeOr:
		return me.TArgs.All((*SemType).isTruthy)
	case MoPrimTypeAnd:
		return me.TArgs.Any((*SemType).isTruthy)
	}
	return me.Singleton == MoValBool(true)
}
func (me *SemType) isFalsy() bool {
	switch me.Prim {
	case MoPrimTypeOr:
		return me.TArgs.All((*SemType).isFalsy)
	case MoPrimTypeAnd:
		return me.TArgs.Any((*SemType).isFalsy)
	}
	return me.Singleton == MoValBool(false)
}

func (me *SemType) IsAny() bool { return me.Prim == MoPrimTypeAny }

func (me *SemType) IsSuperTypeOf(of *SemType) bool { return of.IsSubTypeOf(me) }
func (me *SemType) IsSubTypeOf(of *SemType) bool {
	switch {
	case (me == nil) || (of == nil):
		return (me == of)
	case (me == of) || (of.Prim == MoPrimTypeAny) || (me.Prim == MoPrimTypeNever):
		return true
	case of.Prim == MoPrimTypeNever:
		return false
	case (len(me.TArgs) == 0) && (len(of.TArgs) == 0):
		return (me.Prim == of.Prim) && util.If((me.Singleton == nil), (of.Singleton == nil), (of.Singleton == nil) || (of.Singleton == me.Singleton))
	case (me.Prim == MoPrimTypeObj) && (of.Prim == MoPrimTypeObj):
		for i, targ := range of.TArgs {
			field_name := of.Fields[i]
			if idx := sl.IdxOf(me.Fields, field_name); idx < 0 {
				return false
			} else if !me.TArgs[idx].IsSubTypeOf(targ) {
				return false
			}
		}
		return true
	case (me.Prim == MoPrimTypeTup) && (of.Prim == MoPrimTypeTup):
		for i, targ := range of.TArgs {
			if (i >= len(me.TArgs)) || !me.TArgs[i].IsSubTypeOf(targ) {
				return false
			}
		}
		return true
	case ((me.Prim == MoPrimTypeList) || (me.Prim == MoPrimTypeDict) || (me.Prim == MoPrimTypeErr)) && (of.Prim == me.Prim):
		return (me.Singleton == MoValPrimTypeTag(MoPrimTypeAny) /*used as sentinel on [] and {} literals */) ||
			sl.Eq(me.TArgs, of.TArgs, (*SemType).IsSubTypeOf)
	case (me.Prim == MoPrimTypeFunc) && (of.Prim == MoPrimTypeFunc) && (len(me.TArgs) == len(of.TArgs)):
		for i, targ := range me.TArgs { // last (return) tArg covariant, the others contravariant:
			if is := util.If((i == (len(me.TArgs) - 1)), targ.IsSubTypeOf(of.TArgs[i]), of.TArgs[i].IsSubTypeOf(targ)); !is {
				return false
			}
		}
	case (me.Prim == MoPrimTypeFunc) && (of.Prim == MoPrimTypeFunc) && (len(of.TArgs) == 0): // match "is function at all (whatever arity)"
		return true
	case me.Prim == MoPrimTypeOr:
		return sl.All(me.TArgs, func(ty *SemType) bool { return ty.IsSubTypeOf(of) })
	case of.Prim == MoPrimTypeOr: // TODO: {foo:1|2} is in fact sub of {foo:1}|{foo:2} — yet the below code doesn't agree yet, and should.
		return sl.Any(of.TArgs, func(ty *SemType) bool { return me.IsSubTypeOf(ty) }) // per above: this is correct but not yet complete.
	case me.Prim == MoPrimTypeAnd:
		return me.TArgs.Any(func(ty *SemType) bool { return ty.IsSubTypeOf(of) })
	case of.Prim == MoPrimTypeAnd: // TODO: similar to above comment on Or, see also jaked.org/blog/2021-10-28-Reconstructing-TypeScript-part-5#subtyping-intersection-types
		return of.TArgs.All(func(ty *SemType) bool { return me.IsSubTypeOf(ty) })
	}
	return me.Eq(of)
}

func (me *SemType) mapIfAndOr(dueTo *SemExpr, f func(ty *SemType) *SemType) *SemType {
	switch {
	case (me == nil):
		return f(me)
	case (me.Prim == MoPrimTypeOr):
		ts := sl.To(me.TArgs, func(t *SemType) *SemType { return t.mapIfAndOr(dueTo, f) })
		if sl.Has(ts, nil) {
			return nil
		} else {
			return semTypeOr(dueTo, false, ts...)
		}
	case (me.Prim == MoPrimTypeAnd):
		ts := sl.To(me.TArgs, f)
		if sl.Has(ts, nil) {
			return nil
		} else {
			return semTypeAnd(dueTo, ts...)
		}
	default:
		return f(me)
	}
}

func (me *SemType) normalizeIfAdt() bool {
	if (me.Prim == MoPrimTypeOr) || (me.Prim == MoPrimTypeAnd) {
		// flatten:
		for i := 0; i < me.TArgs.Len(); i++ {
			if t := me.TArgs[i]; (t == nil) || (t.Prim == MoPrimTypeNever) {
				return false
			} else if t.Prim == MoPrimTypeAny {
				*me = *t
				return true
			} else if t.Prim == me.Prim { // lift inner or/and parts to outer
				me.TArgs = append(append(me.TArgs[:i], me.TArgs[i+1:]...), t.TArgs...)
				i--
			}
		}
		collapse := func(ts []*SemType) []*SemType {
			i1 := -1
			return sl.Where(ts, func(t1 *SemType) bool {
				i1++
				i2 := -1
				return sl.All(ts, func(t2 *SemType) bool {
					i2++
					fn_is := util.If(me.Prim == MoPrimTypeAnd, (*SemType).IsSuperTypeOf, (*SemType).IsSubTypeOf)
					return (t1 == t2) || (!fn_is(t1, t2) || ((i1 < i2) && fn_is(t2, t1)))
				})
			})
		}
		if me.Prim == MoPrimTypeOr {
			me.TArgs = collapse(me.TArgs)
		} else { // still following along jaked.org/blog/2021-10-28-Reconstructing-TypeScript-part-5#normalizing-intersections
			me.Prim, me.TArgs = MoPrimTypeOr, sl.To(semTypeDistributeUnions(me.TArgs...), func(ts []*SemType) *SemType {
				for i := 0; i < len(ts); i++ {
					for j := 0; j < len(ts); j++ {
						if (i < j) && !ts[i].overlaps(ts[j]) {
							return semTypeNew(me.DueTo, MoPrimTypeNever)
						}
					}
				}
				ts = collapse(ts)
				if len(ts) == 0 {
					return semTypeNew(me.DueTo, MoPrimTypeNever)
				} else if len(ts) == 1 {
					return ts[0]
				}
				return semTypeNew(me.DueTo, MoPrimTypeAnd, ts...)
			})
		}
		switch len(me.TArgs) {
		case 0:
			return false
		case 1:
			*me = *(me.TArgs[0])
		}
	}
	return true
}

func (me *SemType) overlaps(ty *SemType) bool {
	switch {
	case (me.Prim == MoPrimTypeNever) || (ty.Prim == MoPrimTypeNever):
		return false
	case (me.Prim == MoPrimTypeAny) || (ty.Prim == MoPrimTypeAny):
		return true
	case me.Prim == MoPrimTypeOr:
		return sl.Any(me.TArgs, func(tyArg *SemType) bool { return tyArg.overlaps(ty) })
	case ty.Prim == MoPrimTypeOr:
		return sl.Any(ty.TArgs, func(tyArg *SemType) bool { return me.overlaps(tyArg) })
	case me.Prim == MoPrimTypeAnd:
		return sl.All(me.TArgs, func(tyArg *SemType) bool { return tyArg.overlaps(ty) })
	case ty.Prim == MoPrimTypeAnd:
		return sl.All(ty.TArgs, func(tyArg *SemType) bool { return me.overlaps(tyArg) })
	case (me.Singleton != nil) && (ty.Singleton != nil):
		return (me.Singleton == ty.Singleton)
	case (me.Singleton != nil) || (ty.Singleton != nil):
		return (me.Prim == ty.Prim)
	case (me.Prim == MoPrimTypeObj) && (ty.Prim == MoPrimTypeObj):
		idx := -1
		return sl.All(me.TArgs, func(tyArg *SemType) bool {
			idx++
			field_name := me.Fields[idx]
			if field_idx_other := sl.IdxOf(ty.Fields, field_name); field_idx_other < 0 {
				return true
			} else {
				return tyArg.overlaps(ty.TArgs[field_idx_other])
			}
		})
	case (me.Prim == MoPrimTypeTup) && (ty.Prim == MoPrimTypeTup):
		idx := -1
		return sl.All(me.TArgs, func(tyArg *SemType) bool {
			idx++
			if idx >= len(ty.TArgs) {
				return true
			} else {
				return tyArg.overlaps(ty.TArgs[idx])
			}
		})
	case (me.Prim == MoPrimTypeList) && (ty.Prim == MoPrimTypeList):
		return me.TArgs[0].overlaps(ty.TArgs[0])
	case (me.Prim == MoPrimTypeDict) && (ty.Prim == MoPrimTypeDict):
		return me.TArgs[0].overlaps(ty.TArgs[0]) && me.TArgs[1].overlaps(ty.TArgs[1])
	}
	return (me.Prim == ty.Prim)
}

func (me *SemType) sansSingletons() *SemType {
	if (me == nil) || !me.hasSingletons() {
		return me
	}
	dupl := *me
	dupl.Singleton, dupl.TArgs = nil, sl.To(dupl.TArgs, (*SemType).sansSingletons)
	return util.If(dupl.normalizeIfAdt(), &dupl, nil)
}

func (me *SemType) String() (ret string) {
	var buf str.Buf
	me.stringifyTo(&buf)
	if ret = buf.String(); str.Begins(ret, "(") && str.Ends(ret, ")") && (me.Prim != MoPrimTypeTup) {
		ret = ret[1 : len(ret)-1]
	}
	return
}
func (me *SemType) stringifyTo(w *strings.Builder) {
	if me == nil {
		w.WriteString("<untypifyable>")
		return
	} else if w.Len() > 321 { // infinite-type guard
		w.WriteString("..")
		return
	}

	switch {
	case len(me.TArgs) == 0:
		w.WriteString(me.Prim.Str(false))
		if me.Singleton != nil {
			w.WriteString("=")
			moValStringifyTo(me.Singleton, w)
		}
	case (me.Prim == MoPrimTypeList) && (len(me.TArgs) == 1):
		w.WriteByte('[')
		me.TArgs[0].stringifyTo(w)
		w.WriteByte(']')
	case me.Prim == MoPrimTypeTup:
		w.WriteByte('(')
		for i, targ := range me.TArgs {
			if i > 0 {
				w.WriteString(", ")
			}
			targ.stringifyTo(w)
		}
		w.WriteByte(')')
	case (me.Prim == MoPrimTypeDict) && (len(me.TArgs) == 2):
		w.WriteByte('{')
		me.TArgs[0].stringifyTo(w)
		w.WriteString(": ")
		me.TArgs[1].stringifyTo(w)
		w.WriteByte('}')
	case me.Prim == MoPrimTypeObj:
		w.WriteByte('{')
		for i, targ := range me.TArgs {
			if i > 0 {
				w.WriteString(", ")
			}
			moValStringifyTo(me.Fields[i], w)
			w.WriteString(": ")
			targ.stringifyTo(w)
		}
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

func semTypeDistributeUnions(tys ...*SemType) (ret sl.Of[[]*SemType]) {
	var dist func([]*SemType, int)
	dist = func(tys []*SemType, i int) {
		if i == len(tys) {
			ret.Add(tys)
		} else if ti := tys[i]; ti.Prim != MoPrimTypeOr {
			dist(tys, i+1)
		} else {
			for _, t := range ti.TArgs {
				ts2 := append(append(slices.Clone(tys[:i]), t), tys[i+1:]...)
				dist(ts2, i+1)
			}
		}
	}
	if len(tys) > 0 {
		dist(tys, 0)
	}
	return
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

func semTypeMapIfAndOr(dueTo *SemExpr, t1 *SemType, t2 *SemType, f func(t1 *SemType, t2 *SemType) *SemType) *SemType {
	or1, or2, and1, and2 := (t1.Prim == MoPrimTypeOr), (t2.Prim == MoPrimTypeOr), (t1.Prim == MoPrimTypeAnd), (t2.Prim == MoPrimTypeAnd)
	if is_or := or1 || or2; is_or || and1 || and2 {
		var t1s, t2s []*SemType
		if is_or {
			t1s, t2s = util.If(or1, t1.TArgs, []*SemType{t1}), util.If(or2, t2.TArgs, []*SemType{t2})
		} else {
			t1s, t2s = util.If(and1, t1.TArgs, []*SemType{t1}), util.If(and2, t2.TArgs, []*SemType{t2})
		}
		var ret sl.Of[*SemType]
		for _, t1 := range t1s {
			for _, t2 := range t2s {
				if is_or {
					ret.Add(semTypeMapIfAndOr(dueTo, t1, t2, f))
				} else {
					ret.Add(f(t1, t2))
				}
			}
		}
		if is_or {
			return semTypeOr(dueTo, false, ret...)
		}
		return semTypeAnd(dueTo, ret...)
	}
	return f(t1, t2)
}

func semTypeNew(dueTo *SemExpr, prim MoValPrimType, tyArgs ...*SemType) *SemType {
	if sl.Has(tyArgs, nil) {
		return nil
	}
	ret := &SemType{DueTo: dueTo, Prim: prim, TArgs: sl.To(tyArgs, func(targ *SemType) *SemType { return semTypeEnsureDueTo(dueTo, targ) })}
	if (len(tyArgs) > 0) && !ret.normalizeIfAdt() {
		return nil
	}
	return ret
}

func semTypeOr(dueTo *SemExpr, anyIfEmpty bool, ty ...*SemType) *SemType {
	return semTypeMulti(dueTo, anyIfEmpty, false, ty...)
}
func semTypeAnd(dueTo *SemExpr, ty ...*SemType) *SemType {
	return semTypeMulti(dueTo, false, true, ty...)
}
func semTypeMulti(dueTo *SemExpr, anyIfEmpty bool, isAnd bool, ty ...*SemType) *SemType {
	types := (sl.Of[*SemType])(ty)
	use_any := anyIfEmpty && (len(types) == 0)
	types = types.Without(func(t *SemType) bool { return t == nil })
	switch types.EnsureAllUnique((*SemType).Eq); len(types) {
	case 0:
		return util.If(use_any, semTypeNew(dueTo, MoPrimTypeAny), nil)
	case 1:
		return types[0]
	default:
		return semTypeNew(dueTo, util.If(isAnd, MoPrimTypeAnd, MoPrimTypeOr), types...)
	}
}

func (me *SrcPack) semCheckCount(wantAtLeast int, wantAtMost int, have SemExprs, errDst *SemExpr, forArgs bool) bool {
	if wantAtLeast >= 0 {
		plural := util.If((wantAtLeast <= wantAtMost) && (wantAtLeast != 1), "s", "")
		moniker := util.If(!forArgs, "expression"+plural+" in here", "arg"+plural+" for callee")
		if forArgs && (errDst != nil) {
			if call, _ := errDst.Val.(*SemValCall); (call != nil) && (call.Callee.From != nil) && (call.Callee.From.SrcNode != nil) && (call.Callee.From.SrcNode.Src != "") {
				moniker += " `" + str.Shorten(call.Callee.From.SrcNode.Src, 11) + "`"
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
	t_anys, t_any := make([]*SemType, arity), semTypeNew(dueTo, MoPrimTypeAny)
	for i := range t_anys {
		t_anys[i] = t_any
	}
	return me.semCheckType(expr, semTypeNew(dueTo, expect, t_anys...))
}

func (me *SrcPack) semCheckType(expr *SemExpr, expect *SemType) bool {
	return me.semCheckTypeLax(expr, expect, false)
}

func (me *SrcPack) semCheckTypeLax(expr *SemExpr, expect *SemType, sansSingletons bool) bool {
	ty := expr.Type
	if sansSingletons {
		ty, expect = ty.sansSingletons(), expect.sansSingletons()
	}
	if !ty.IsSubTypeOf(expect) {
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
	s1, s2 := "`"+t1.String()+"`", "`"+t2.String()+"`"
	err := self.ErrNew(ErrCodeTypeMismatch, s1+" value", s2+" value")
	err.Rel = srcFileLocs([]string{
		str.Fmt("wanting %s due to `%s`", s1, str.Shorten(dt1.String(false), 22)),
		str.Fmt("but given %s due to `%s`", s2, str.Shorten(dt2.String(false), 22)),
	}, t1.DueTo, t2.DueTo)
	return err
}

func (me *SrcPack) semTypeNarrow(env semTyEnv, self *SemExpr, isTrue bool) semTyEnv {
	switch val := self.Val.(type) {
	case *SemValCall:
		switch callee := val.Callee.MaybeIdent(false); callee {
		case moPrimFnBoolNot:
			return me.semTypeNarrow(env, val.Args[0], !isTrue)
		case moPrimOpBoolAnd:
			if isTrue {
				return me.semTypeNarrow(me.semTypeNarrow(env, val.Args[0], true), val.Args[1], true)
			} else if me.semTypify(val.Args[0], env); (val.Args[0].Type != nil) && val.Args[0].Type.isTruthy() {
				return me.semTypeNarrow(env, val.Args[1], false)
			} else if me.semTypify(val.Args[1], env); (val.Args[1].Type != nil) && val.Args[1].Type.isTruthy() {
				return me.semTypeNarrow(env, val.Args[0], false)
			}
		case moPrimOpBoolOr:
			if !isTrue {
				return me.semTypeNarrow(me.semTypeNarrow(env, val.Args[0], false), val.Args[1], false)
			} else if me.semTypify(val.Args[0], env); (val.Args[0].Type != nil) && val.Args[0].Type.isFalsy() {
				return me.semTypeNarrow(env, val.Args[1], true)
			} else if me.semTypify(val.Args[1], env); (val.Args[1].Type != nil) && val.Args[1].Type.isFalsy() {
				return me.semTypeNarrow(env, val.Args[0], true)
			}
		case moPrimFnCmpEq, moPrimFnCmpNeq:
			me.semTypify(val.Args[0], env)
			me.semTypify(val.Args[1], env)
			if ((callee == moPrimFnCmpEq) && isTrue) || ((callee == moPrimFnCmpNeq) && !isTrue) {
				return me.semTypeNarrowPath(me.semTypeNarrowPath(env, val.Args[0], val.Args[1].Type), val.Args[1], val.Args[0].Type)
			} else if ((callee == moPrimFnCmpNeq) && isTrue) || ((callee == moPrimFnCmpEq) && !isTrue) {
				if val.Args[1].Type.Singleton != nil {
					env = me.semTypeNarrowPath(env, val.Args[0], semTypeNew(val.Callee, MoPrimTypeNot, val.Args[1].Type))
				}
				if val.Args[0].Type.Singleton != nil {
					env = me.semTypeNarrowPath(env, val.Args[1], semTypeNew(val.Callee, MoPrimTypeNot, val.Args[0].Type))
				}
			}
		}
	}
	return env
}

func (me *SrcPack) semTypeNarrowPath(env semTyEnv, self *SemExpr, ty *SemType) semTyEnv {
	switch val := self.Val.(type) {
	case *SemValIdent:
		if t := env[val.Name]; t != nil {
			return env.set(val.Name, me.semTypeNarrowType(self, t, ty))
		}
	case *SemValCall:
		if self.MaybeIdent(false) == moPrimFnObjGet {
			ty_obj := semTypeNew(val.Callee, MoPrimTypeObj, ty)
			ty_obj.Fields = sl.Of[MoValIdent]{val.Args[1].Val.(*SemValIdent).Name}
			return me.semTypeNarrowPath(env, val.Args[0], ty_obj)
		}
	}
	return env
}

func (me *SrcPack) semTypeWidenNots(ty *SemType) *SemType {
	switch ty.Prim {
	case MoPrimTypeNot:
		return semTypeNew(ty.DueTo, MoPrimTypeAny)
	case MoPrimTypeOr:
		return semTypeOr(ty.DueTo, false, sl.To(ty.TArgs, me.semTypeWidenNots)...)
	case MoPrimTypeAnd:
		return semTypeAnd(ty.DueTo, sl.To(ty.TArgs, me.semTypeWidenNots)...)
	case MoPrimTypeList, MoPrimTypeDict, MoPrimTypeTup, MoPrimTypeObj:
		dupl := *ty
		dupl.TArgs = sl.To(dupl.TArgs, me.semTypeWidenNots)
		return &dupl
	}
	return ty
}

func (me *SrcPack) semTypeNarrowType(dueTo *SemExpr, t1 *SemType, t2 *SemType) *SemType {
	ty_never := semTypeNew(dueTo, MoPrimTypeNever)
	switch {
	case (t1 == nil) || (t2 == nil) || (t1.Prim == MoPrimTypeNever) || (t2.Prim == MoPrimTypeNever):
		return ty_never
	case t1.Prim == MoPrimTypeAny:
		return me.semTypeWidenNots(t2)
	case t2.Prim == MoPrimTypeAny:
		return t1
	case t1.Prim == MoPrimTypeOr:
		return semTypeOr(t1.DueTo, false, sl.To(t1.TArgs, func(t *SemType) *SemType { return me.semTypeNarrowType(t.DueTo, t, t2) })...)
	case t2.Prim == MoPrimTypeOr:
		return semTypeOr(t2.DueTo, false, sl.To(t2.TArgs, func(t *SemType) *SemType { return me.semTypeNarrowType(t.DueTo, t1, t) })...)
	case t1.Prim == MoPrimTypeAnd:
		return semTypeAnd(t1.DueTo, sl.To(t1.TArgs, func(t *SemType) *SemType { return me.semTypeNarrowType(t.DueTo, t, t2) })...)
	case t2.Prim == MoPrimTypeAnd:
		return semTypeAnd(t2.DueTo, sl.To(t2.TArgs, func(t *SemType) *SemType { return me.semTypeNarrowType(t.DueTo, t1, t) })...)
	case t2.Prim == MoPrimTypeNot:
		if t1.IsSubTypeOf(t2.TArgs[0]) {
			return ty_never
		} else if (t1.Prim == MoPrimTypeBool) && (t2.TArgs[0].Singleton != nil) && (t2.TArgs[0].Prim == MoPrimTypeBool) {
			ret := semTypeNew(dueTo, MoPrimTypeBool)
			ret.Singleton = !t2.TArgs[0].Singleton.(MoValBool)
			return ret
		} else {
			return t1
		}
	case (t1.Singleton != nil) && (t2.Singleton != nil):
		return util.If(t1.Singleton == t2.Singleton, t1, ty_never)
	case t1.Singleton != nil:
		return util.If(t1.Prim == t2.Prim, t1, ty_never)
	case t2.Singleton != nil:
		return util.If(t2.Prim == t1.Prim, t2, ty_never)
	case (t2.Prim == t1.Prim) && (MoPrimTypeObj == t1.Prim):
		var fields []MoValIdent
		var types []*SemType
		for i, targ := range t1.TArgs {
			name := t1.Fields[i]
			idx := sl.IdxOf(t2.Fields, name)
			tf := targ
			if idx >= 0 {
				tf = me.semTypeNarrowType(dueTo, targ, t2.TArgs[idx])
			}
			fields, types = append(fields, name), append(types, tf)
		}
		if sl.Any(types, func(t *SemType) bool { return (t == nil) || (t.Prim == MoPrimTypeNever) }) {
			return ty_never
		}
		ret := semTypeNew(dueTo, MoPrimTypeObj, types...)
		ret.Fields = fields
		return ret
	case (t2.Prim == t1.Prim) && ((MoPrimTypeTup == t1.Prim) || (MoPrimTypeDict == t1.Prim) || (MoPrimTypeList == t1.Prim)):
		var types []*SemType
		for i, targ := range t1.TArgs {
			tf := targ
			if i < len(t2.TArgs) {
				tf = me.semTypeNarrowType(dueTo, targ, t2.TArgs[i])
			}
			types = append(types, tf)
		}
		if sl.Any(types, func(t *SemType) bool { return (t == nil) || (t.Prim == MoPrimTypeNever) }) {
			return ty_never
		}
		return semTypeNew(dueTo, t1.Prim /*same as t2 remember*/, types...)
	}
	return semTypeAnd(dueTo, t1, t2)
}
