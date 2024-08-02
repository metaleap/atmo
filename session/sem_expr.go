package session

import (
	"atmo/util/sl"
	"atmo/util/str"
	"strings"
)

type SemExprs sl.Of[*SemExpr]
type SemExpr struct {
	From    *MoExpr   `json:"-"`
	Parent  *SemExpr  `json:"-"`
	Scope   *SemScope `json:"-"`
	ErrsOwn Diags     `json:",omitempty"`
	Val     any
	ValOrig any                  `json:"-"` // nil unless Val was statically pre-computed from its orig expr, which then is preserved in here
	Facts   map[SemFact]SemExprs `json:"-"`
	Type    SemType              `json:"-"`
}

type SemValScalar struct {
	Value MoVal
}

type SemValIdent struct {
	Name    MoValIdent
	IsDecl  bool // if true, either IsParam or IsSet also is
	IsParam bool // if true, so is IsDecl
	IsSet   bool
}

type SemValCall struct {
	Callee *SemExpr
	Args   SemExprs
}

type SemValList struct {
	Items SemExprs
}

type SemValDict struct {
	Keys SemExprs
	Vals SemExprs
}

type SemValFunc struct {
	Scope    *SemScope
	Params   SemExprs
	Body     *SemExpr
	primImpl func(*SrcPack, *SemExpr, *SemScope)
	IsMacro  bool
}

func (me *SemExpr) Each(do func(it *SemExpr)) {
	switch val := me.Val.(type) {
	case SemValCall:
		do(val.Callee)
		sl.Each(val.Args, do)
	case SemValDict:
		sl.Each(val.Keys, do)
		sl.Each(val.Vals, do)
	case SemValFunc:
		sl.Each(val.Params, do)
		do(val.Body)
	case SemValList:
		sl.Each(val.Items, do)
	}
}

func (me *SemExpr) ErrNew(code DiagCode, args ...any) *Diag {
	return me.From.SrcSpan.newDiagErr(code, args...)
}

func (me *SemExpr) Errs() (ret Diags) {
	me.Walk(true, nil, func(it *SemExpr) {
		ret.Add(it.ErrsOwn...)
	})
	return
}

func (me *SemExpr) Fact(fact SemFact, from *SemExpr) {
	if me.Facts == nil {
		me.Facts = map[SemFact]SemExprs{}
	}
	if it := me.Facts[fact]; !sl.Has(it, from) {
		me.Facts[fact] = append(it, from)
	}
}

func (me *SemExpr) FactCauses(kind SemFactKind, of any, orAncestor bool, orDescendant bool) (dueTo SemExprs) {
	if me.Facts == nil {
		return
	}
	dueTo = me.Facts[SemFact{Kind: kind, Data: of}]
	if (len(dueTo) == 0) && orAncestor && (me.Parent != nil) {
		dueTo = me.Parent.FactCauses(kind, of, true, false)
	}
	if (len(dueTo) == 0) && orDescendant {
		me.Each(func(it *SemExpr) {
			dueTo = append(dueTo, it.FactCauses(kind, of, false, true)...)
		})
	}
	return
}

func (me *SemExpr) FactsFrom(from *SemExpr) {
	for fact := range from.Facts {
		me.Fact(fact, from)
	}
}

func (me *SemExpr) HasErrs() (ret bool) {
	me.Walk(true, func(it *SemExpr) bool {
		ret = ret || (len(it.ErrsOwn) > 0)
		return !ret
	}, nil)
	return
}

func (me *SemExpr) HasFact(kind SemFactKind, of any, orAncestor bool, orDescendant bool) (ret bool) {
	if me.Facts == nil {
		return
	}
	ret = (len(me.Facts[SemFact{Kind: kind, Data: of}]) > 0)
	if (!ret) && orAncestor && (me.Parent != nil) {
		ret = me.Parent.HasFact(kind, of, true, false)
	}
	if (!ret) && orDescendant {
		me.Each(func(it *SemExpr) {
			ret = it.HasFact(kind, of, false, true)
		})
	}
	return
}

func (me *SemExpr) isPrecomputedPermissible() bool {
	return (!me.HasErrs()) && !me.HasFact(SemFactNotPure, nil, false, true)
}

func (me *SemExpr) MaybeIdent(canBeDecl bool) MoValIdent {
	ident, _ := me.Val.(*SemValIdent)
	if (ident != nil) && (canBeDecl || !ident.IsDecl) {
		return ident.Name
	}
	return ""
}

func (me *SrcPack) Refs(self *SemExpr, onlyInFile *SrcFile) (sets SemExprs, gets SemExprs) {
	name := self.MaybeIdent(true)
	if _, entry := self.Scope.Lookup(name); entry != nil {
		sets = append(SemExprs{entry.DeclParamOrCallOrFunc}, entry.SubsequentSetCalls...)
		walk_from := SemExprs{entry.DeclParamOrCallOrFunc.Parent}
		if fn, _ := entry.DeclParamOrCallOrFunc.Val.(*SemValFunc); (fn != nil) || (walk_from[0] == nil) {
			walk_from = me.Trees.Sem.TopLevel
		}
		walk_from.Walk(onlyInFile, nil, func(it *SemExpr) {
			if ident, _ := it.Val.(*SemValIdent); (ident != nil) && !(ident.IsDecl || ident.IsSet || ident.IsParam) {
				if _, e := it.Scope.Lookup(ident.Name); e == entry {
					gets = append(gets, it)
				}
			}
		})
	}
	return
}

func (me *SemExpr) Walk(origValTooIfValPrecomputedFromExpr bool, onBefore func(it *SemExpr) bool, onAfter func(it *SemExpr)) {
	if onBefore != nil && !onBefore(me) {
		return
	}
	vals := []any{me.Val}
	if origValTooIfValPrecomputedFromExpr && (me.ValOrig != nil) {
		vals = append(vals, me.ValOrig)
	}
	for _, val := range vals {
		switch it := val.(type) {
		case *SemValCall:
			it.Callee.Walk(origValTooIfValPrecomputedFromExpr, onBefore, onAfter)
			for _, arg := range it.Args {
				arg.Walk(origValTooIfValPrecomputedFromExpr, onBefore, onAfter)
			}
		case *SemValList:
			for _, item := range it.Items {
				item.Walk(origValTooIfValPrecomputedFromExpr, onBefore, onAfter)
			}
		case *SemValDict:
			for i, key := range it.Keys {
				key.Walk(origValTooIfValPrecomputedFromExpr, onBefore, onAfter)
				it.Vals[i].Walk(origValTooIfValPrecomputedFromExpr, onBefore, onAfter)
			}
		case *SemValFunc:
			for _, param := range it.Params {
				param.Walk(origValTooIfValPrecomputedFromExpr, onBefore, onAfter)
			}
			it.Body.Walk(origValTooIfValPrecomputedFromExpr, onBefore, onAfter)
		}
	}
	if onAfter != nil {
		onAfter(me)
	}
}

func (me SemExprs) AnyErrs() bool {
	return sl.Any(me, (*SemExpr).HasErrs)
}

func (me SemExprs) At(srcFile *SrcFile, pos *SrcFilePos) (ret *SemExpr) {
	for _, top_expr := range me {
		if (top_expr.From != nil) && (top_expr.From.SrcFile == srcFile) && (top_expr.From.SrcSpan != nil) && (top_expr.From.SrcSpan.Contains(pos)) {
			ret = top_expr
			top_expr.Walk(true, func(it *SemExpr) bool {
				if (it.From != nil) && (it.From.SrcSpan != nil) && it.From.SrcSpan.Contains(pos) {
					ret = it
				}
				return true
			}, nil)
			break
		}
	}
	return
}

func (me SemExprs) Errs() (ret Diags) {
	for _, top_expr := range me {
		ret.Add(top_expr.Errs()...)
	}
	return
}

func (me SemExprs) Walk(onlyInFile *SrcFile, onBefore func(it *SemExpr) bool, onAfter func(it *SemExpr)) {
	for _, expr := range me {
		if (onlyInFile == nil) || (expr.From.SrcFile == onlyInFile) {
			expr.Walk(true, onBefore, onAfter)
		}
	}
}

func (me *SemExpr) str(w *str.Buf) {
	switch val := me.Val.(type) {
	case *SemValIdent:
		w.WriteString(string(val.Name))
	case *SemValScalar:
		moValWriteTo(val.Value, w)
	case *SemValList:
		w.WriteByte('[')
		for i, item := range val.Items {
			if i > 0 {
				w.WriteString(", ")
			}
			item.str(w)
		}
		w.WriteByte(']')
	case *SemValDict:
		w.WriteByte('{')
		for i, key := range val.Keys {
			if i > 0 {
				w.WriteString(", ")
			}
			key.str(w)
			w.WriteString(": ")
			val.Vals[i].str(w)
		}
		w.WriteByte('}')
	case *SemValCall:
		w.WriteByte('(')
		val.Callee.str(w)
		for _, arg := range val.Args {
			w.WriteByte(' ')
			arg.str(w)
		}
		w.WriteByte(')')
	case *SemValFunc:
		w.WriteString("@fn [")
		for i, param := range val.Params {
			if i > 0 {
				w.WriteString(", ")
			}
			param.str(w)
		}
		w.WriteString("] [")
		if val.Body == nil {
			w.WriteString("…")
		} else {
			val.Body.str(w)
		}
		w.WriteByte(']')
	}
}

func (me *SemExpr) String() string {
	if me == nil {
		return "<nil>"
	}
	var buf strings.Builder
	me.str(&buf)
	return buf.String()
}

type SemFactKind int

const (
	_ SemFactKind = iota
	SemFactUnused
	SemFactNotPure
	SemFactPreComputed
	SemFactPrimOp
	SemFactPrimFn
	SemFactPrimIdent
)

type SemFact struct {
	Kind SemFactKind
	Data any `json:",omitempty"`
}

func (me *SemFact) String() (ret string) {
	switch me.Kind {
	default:
		ret = "?!"
	case SemFactNotPure:
		ret = "notPure"
	case SemFactUnused:
		ret = "unused"
	case SemFactPreComputed:
		ret = "staticallyComputed"
	case SemFactPrimFn:
		ret = "primFn"
	case SemFactPrimOp:
		ret = "primOp"
	case SemFactPrimIdent:
		ret = "primIdent"
	}
	if me.Data != nil {
		of := me.Data
		if val, _ := of.(MoVal); val != nil {
			of = MoValToString(val)
		} else if ty, _ := of.(SemType); ty != nil {
			of = SemTypeToString(ty)
		}
		ret += str.Fmt("(%v)", of)
	}
	return
}
