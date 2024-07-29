package session

import (
	"atmo/util/sl"
)

type SemExprs sl.Of[*SemExpr]
type SemExpr struct {
	From             *MoExpr        `json:"-"`
	Parent           *SemExpr       `json:"-"`
	Scope            *SemScope      `json:"-"`
	ErrOwn           *SrcFileNotice `json:",omitempty"`
	Val              any
	DefinitelyUnused bool `json:",omitempty"`
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

func (me *SrcPack) semRefresh() {
	me.Trees.Sem.TopLevel = nil
	me.Trees.Sem.Scope = SemScope{Own: map[MoValIdent]*SemExpr{}}
	for _, top_expr := range me.Trees.MoOrig {
		it := me.semExprFromMoExpr(&me.Trees.Sem.Scope, top_expr, nil)
		if call, _ := it.Val.(*SemValCall); call == nil {
			it.DefinitelyUnused = true
		}
		me.Trees.Sem.TopLevel = append(me.Trees.Sem.TopLevel, it)
	}
}

func (me *SrcPack) semExprFromMoExpr(scope *SemScope, moExpr *MoExpr, parent *SemExpr) *SemExpr {
	ret := &SemExpr{From: moExpr, Parent: parent, Scope: scope}
	switch it := moExpr.Val.(type) {
	default:
		panic(it)
	case MoValNumInt:
		me.semPopulateScalar(ret, it)
	case MoValNumUint:
		me.semPopulateScalar(ret, it)
	case MoValNumFloat:
		me.semPopulateScalar(ret, it)
	case MoValChar:
		me.semPopulateScalar(ret, it)
	case MoValStr:
		me.semPopulateScalar(ret, it)
	case MoValIdent:
		me.semPopulateIdent(ret, it)
	case MoValList:
		me.semPopulateList(ret, it)
	case MoValDict:
		me.semPopulateDict(ret, it)
	case MoValCall:
		me.semPopulateCall(ret, it)
	}
	return ret
}

func (me *SrcPack) semPopulateScalar(self *SemExpr, it MoVal) {
	scalar := &SemValScalar{MoVal: it}
	self.Val = scalar
	me.Trees.Sem.Index.Lits[it] = append(me.Trees.Sem.Index.Lits[it], self)
}

func (me *SrcPack) semPopulateIdent(self *SemExpr, it MoValIdent) {
	ident := &SemValIdent{MoVal: it}
	self.Val = ident
}

func (me *SrcPack) semPopulateList(self *SemExpr, it MoValList) {
	list := &SemValList{Items: make(SemExprs, len(it))}
	self.Val = list
	for i, item := range it {
		list.Items[i] = me.semExprFromMoExpr(self.Scope, item, self)
	}
}

func (me *SrcPack) semPopulateDict(self *SemExpr, it MoValDict) {
	dict := &SemValDict{Keys: make(SemExprs, len(it)), Vals: make(SemExprs, len(it))}
	self.Val = dict
	for i, item := range it {
		dict.Keys[i] = me.semExprFromMoExpr(self.Scope, item.Key, self)
		dict.Vals[i] = me.semExprFromMoExpr(self.Scope, item.Val, self)
	}
}

func (me *SrcPack) semPopulateCall(self *SemExpr, it MoValCall) {
	call := &SemValCall{Callee: me.semExprFromMoExpr(self.Scope, it[0], self)}
	self.Val = call

	for _, arg := range it[1:] {
		call.Args = append(call.Args, me.semExprFromMoExpr(self.Scope, arg, self))
	}

	if self.HasErrs() {
		return
	}

	if ident := call.Callee.MaybeIdent(); ident != "" {
		if prim_op := semPrimOps[ident]; prim_op != nil {
			prim_op(me, self)
			return
		}
	}

	var callee_fn *SemValFunc
	if callee := call.Callee.ResolvedIfIdent(); callee != nil {
		if callee_fn = callee.Val.(*SemValFunc); callee_fn == nil {
			self.ErrOwn = self.From.SrcNode.newDiagErr(false, NoticeCodeUncallable, self.From.SrcNode.Src)
		}
	}
}

func (me *SrcPack) semPopulateFunc(self *SemExpr, it *MoValFnLam) {
	fn := &SemValFunc{
		Scope:  &SemScope{Parent: self.Scope, Own: map[MoValIdent]*SemExpr{}},
		Params: make(SemExprs, len(it.Params)),
	}
	self.Val = fn
	for i, param := range it.Params {
		ident := me.semExprFromMoExpr(self.Scope, param, self)
		fn.Scope.Own[ident.MaybeIdent()] = ident
		fn.Params[i] = ident
	}
	fn.Body = me.semExprFromMoExpr(fn.Scope, it.Body, self)
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

func (me *SemExpr) ResolvedIfIdent() *SemExpr {
	if me.ErrOwn != nil {
		return nil
	}
	ident := me.MaybeIdent()
	if ident == "" {
		return me
	}
	_, resolved := me.Scope.Lookup(ident, false, true)
	if resolved == nil {
		me.ErrOwn = me.From.SrcNode.newDiagErr(false, NoticeCodeUndefined, ident)
	}
	return resolved
}

func (me *SemExpr) HasErrs() (ret bool) {
	me.Walk(func(it *SemExpr) bool {
		ret = ret || (it.ErrOwn != nil)
		return !ret
	}, nil)
	return
}

func (me *SemExpr) Errs() (ret SrcFileNotices) {
	me.Walk(nil, func(it *SemExpr) {
		if it.ErrOwn != nil {
			ret.Add(it.ErrOwn)
		}
	})
	return
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

type SemScope struct {
	Own    map[MoValIdent]*SemExpr
	Parent *SemScope `json:"-"`
}

func (me *SemScope) Lookup(ident MoValIdent, ownOnly bool, deepResolveUntilNonIdent bool) (*SemScope, *SemExpr) {
	if resolved := me.Own[ident]; resolved != nil {
		if !deepResolveUntilNonIdent {
			return me, resolved
		} else if alias, _ := resolved.Val.(*SemValIdent); alias == nil {
			return me, resolved
		} else if _, resolved = me.Lookup(alias.MoVal, ownOnly, true); resolved != nil {
			return me, resolved
		}
	}
	if (!ownOnly) && (me.Parent != nil) {
		return me.Parent.Lookup(ident, false, deepResolveUntilNonIdent)
	}
	return nil, nil
}
