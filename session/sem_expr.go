package session

import (
	"atmo/util/sl"
	"atmo/util/str"
)

type SemExprs sl.Of[*SemExpr]
type SemExpr struct {
	From    *MoExpr   `json:"-"`
	Parent  *SemExpr  `json:"-"`
	Scope   *SemScope `json:"-"`
	ErrsOwn Diags     `json:",omitempty"`
	Val     any
	Facts   map[SemFact]SemExprs `json:"-"`
}

type SemValScalar struct {
	MoVal MoVal
}

type SemValIdent struct {
	MoVal MoValIdent
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
	Scope   *SemScope
	Params  SemExprs
	Body    *SemExpr
	IsMacro bool
}

type SemValHole struct {
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

func (me *SemExpr) Errs() (ret Diags) {
	me.Walk(nil, func(it *SemExpr) {
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

func (me *SemExpr) FactsFrom(from *SemExpr) {
	for fact := range from.Facts {
		me.Fact(fact, from)
	}
}

func (me *SemExpr) HasErrs() (ret bool) {
	me.Walk(func(it *SemExpr) bool {
		ret = ret || (len(it.ErrsOwn) > 0)
		return !ret
	}, nil)
	return
}

func (me *SemExpr) HasFact(kind SemFactKind, of any, orAncestor bool, orDescendant bool) (dueTo SemExprs) {
	if me.Facts == nil {
		return
	}
	dueTo = me.Facts[SemFact{Kind: kind, Data: of}]
	if (len(dueTo) == 0) && orAncestor && (me.Parent != nil) {
		dueTo = me.Parent.HasFact(kind, of, true, false)
	}
	if (len(dueTo) == 0) && orDescendant {
		me.Each(func(it *SemExpr) {
			dueTo = append(dueTo, it.HasFact(kind, of, false, true)...)
		})
	}
	return
}

func (me *SemExpr) MaybeArgOfCall() int {
	if call, _ := me.Parent.Val.(*SemValCall); call != nil {
		for i, arg := range call.Args {
			if arg == me {
				return i
			}
		}
	}
	return -1
}

func (me *SemExpr) MaybeBodyOfFunc() bool {
	if fn, _ := me.Parent.Val.(*SemValFunc); fn != nil {
		return (fn.Body == me)
	}
	return false
}

func (me *SemExpr) MaybeCalleeOfCall() bool {
	if call, _ := me.Parent.Val.(*SemValCall); call != nil {
		return (call.Callee == me)
	}
	return false
}

func (me *SemExpr) MaybeIdent() MoValIdent {
	ident, _ := me.Val.(*SemValIdent)
	if ident != nil {
		return ident.MoVal
	}
	return ""
}

func (me *SemExpr) MaybeParamOfFunc() int {
	if fn, _ := me.Parent.Val.(*SemValFunc); fn != nil {
		for i, param := range fn.Params {
			if param == me {
				return i
			}
		}
	}
	return -1
}

func (me *SemExpr) Walk(onBefore func(it *SemExpr) bool, onAfter func(it *SemExpr)) {
	if onBefore != nil && !onBefore(me) {
		return
	}
	switch it := me.Val.(type) {
	case *SemValCall:
		it.Callee.Walk(onBefore, onAfter)
		for _, arg := range it.Args {
			arg.Walk(onBefore, onAfter)
		}
	case *SemValList:
		for _, item := range it.Items {
			item.Walk(onBefore, onAfter)
		}
	case *SemValDict:
		for i, key := range it.Keys {
			key.Walk(onBefore, onAfter)
			it.Vals[i].Walk(onBefore, onAfter)
		}
	case *SemValFunc:
		for _, param := range it.Params {
			param.Walk(onBefore, onAfter)
		}
		it.Body.Walk(onBefore, onAfter)
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
			top_expr.Walk(func(it *SemExpr) bool {
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

func (me SemExprs) Walk(filterBy *SrcFile, onBefore func(it *SemExpr) bool, onAfter func(it *SemExpr)) {
	for _, expr := range me {
		if (filterBy == nil) || (expr.From.SrcFile == filterBy) {
			expr.Walk(onBefore, onAfter)
		}
	}
}

type SemFactKind int

const (
	_ SemFactKind = iota
	SemFactCallable
	SemFactUnused
	SemFactEffectful
	SemFactScalar
	SemFactPrimType
	SemFactFuncIsMacro
	SemFactQQuote
)

type SemFact struct {
	Kind SemFactKind
	Data any `json:",omitempty"`
}

func (me *SemFact) String() (ret string) {
	switch me.Kind {
	default:
		ret = "?!"
	case SemFactCallable:
		ret = "callable"
	case SemFactEffectful:
		ret = "effectful"
	case SemFactUnused:
		ret = "unused"
	case SemFactScalar:
		ret = "scalar"
	case SemFactPrimType:
		ret = "primType"
	case SemFactFuncIsMacro:
		ret = "fnMacro"
	case SemFactQQuote:
		ret = "qQuote"
	}
	if me.Data != nil {
		of := me.Data
		if val, _ := of.(MoVal); val != nil {
			of = MoValToString(val)
		}
		ret += str.Fmt("(%v)", of)
	}
	return
}
