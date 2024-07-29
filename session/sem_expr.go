package session

import (
	"atmo/util/sl"
)

type SemExprs sl.Of[*SemExpr]
type SemExpr struct {
	From   *MoExpr        `json:"-"`
	Parent *SemExpr       `json:"-"`
	Scope  *SemScope      `json:"-"`
	ErrOwn *SrcFileNotice `json:",omitempty"`
	Val    any
	Facts  map[SemValFact]SemExprs
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
	IsMacro bool `json:",omitempty"`
}

func (me *SemExpr) MaybeCalleeOfCall() bool {
	if call, _ := me.Parent.Val.(*SemValCall); call != nil {
		return (call.Callee == me)
	}
	return false
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

func (me *SemExpr) MaybeIdent() MoValIdent {
	ident, _ := me.Val.(*SemValIdent)
	if ident != nil {
		return ident.MoVal
	}
	return ""
}

func (me *SemExpr) EnsureResolvesIfIdent() *SemScopeEntry {
	if me.ErrOwn != nil {
		return nil
	}
	if ident := me.MaybeIdent(); ident != "" {
		if _, entry := me.Scope.Lookup(ident, false, me); entry != nil {
			return entry
		}
	}
	return nil
}

func (me *SemExpr) Errs() (ret SrcFileNotices) {
	me.Walk(nil, func(it *SemExpr) {
		if it.ErrOwn != nil {
			ret.Add(it.ErrOwn)
		}
	})
	return
}

func (me *SemExpr) Fact(fact SemValFact, from *SemExpr) {
	if me.Facts == nil {
		me.Facts = map[SemValFact]SemExprs{}
	}
	if it := me.Facts[fact]; !sl.Has(it, from) {
		me.Facts[fact] = append(it, from)
		if entry := me.EnsureResolvesIfIdent(); entry != nil {
			entry.DeclVal.Fact(fact, from)
			for _, it := range entry.SubsequentSetVals {
				it.Fact(fact, from)
			}
		}
	}
}

func (me *SemExpr) HasErrs() (ret bool) {
	me.Walk(func(it *SemExpr) bool {
		ret = ret || (it.ErrOwn != nil)
		return !ret
	}, nil)
	return
}

func (me *SemExpr) HasFact(kind SemValFactKind, of any) SemExprs {
	if me.Facts == nil {
		return nil
	}
	return me.Facts[SemValFact{Kind: kind, Of: of}]
}

func (me SemExprs) AnyErrs() bool {
	return sl.Any(me, (*SemExpr).HasErrs)
}

func (me SemExprs) Errs() (ret SrcFileNotices) {
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

type SemValFactKind int

const (
	SemValFactCallable SemValFactKind = iota
	SemValFactUnused
)

type SemValFact struct {
	Kind SemValFactKind
	Of   any
}
