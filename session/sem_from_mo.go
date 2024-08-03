package session

import (
	"atmo/util"
	"atmo/util/sl"
	"atmo/util/str"
)

func (me *SrcPack) semRefresh() {
	me.Trees.Sem.TopLevel = make(SemExprs, 0, len(me.Trees.MoOrig))
	me.Trees.Sem.Scope = SemScope{Own: map[MoValIdent]*SemScopeEntry{}}
	for _, top_expr := range me.Trees.MoOrig {
		it := me.semExprFromMoExpr(&me.Trees.Sem.Scope, top_expr, nil)
		me.Trees.Sem.TopLevel = append(me.Trees.Sem.TopLevel, it)
	}
	if me.Interp == nil {
		_ = newInterp(me, nil)
	}
	me.moPrePackEval()
	if !me.Trees.Sem.TopLevel.AnyErrs() {
		// me.semInferTypes()
		me.semPopulateRootScope()
		me.semInfer()
		// for _, top_expr := range me.Trees.Sem.TopLevel {
		// 	me.semTypify(top_expr, &me.Trees.Sem.Scope)
		// }
		if false && !me.Trees.Sem.TopLevel.AnyErrs() {
			me.Trees.Sem.TopLevel.Walk(nil, func(it *SemExpr) bool {
				ident, _ := it.Val.(*SemValIdent)
				if (it.Type == nil) && (!it.HasErrs()) && (!it.HasFact(SemFactPrimOp, nil, false, false)) && ((ident == nil) || !ident.IsParam) {
					it.ErrAdd(it.ErrNew(ErrCodeUntypifiable))
				}
				return true
			}, nil)
		}
	}
}

func (me *SrcPack) semExprFromMoExpr(scope *SemScope, moExpr *MoExpr, parent *SemExpr) *SemExpr {
	ret := &SemExpr{From: moExpr, Parent: parent, Scope: scope}
	switch it := moExpr.Val.(type) {
	default:
		panic(moExpr.String())
	case MoValVoid, MoValBool, MoValNumInt, MoValNumUint, MoValNumFloat, MoValChar, MoValStr, MoValPrimTypeTag:
		me.semPopulateScalar(ret, it)
	case MoValIdent:
		me.semPopulateIdent(ret, it)
	case *MoValList:
		me.semPopulateList(ret, it)
	case *MoValDict:
		me.semPopulateDict(ret, it)
	case MoValCall:
		me.semPopulateCall(ret, it)
	}
	return ret
}

func (me *SrcPack) semPopulateScalar(self *SemExpr, it MoVal) {
	scalar := &SemValScalar{Value: it}
	self.Val = scalar
	self.Fact(SemFact{Kind: SemFactPreComputed}, self)
	me.Trees.Sem.Index.Lits[it] = append(me.Trees.Sem.Index.Lits[it], self)
	if (it.PrimType() == MoPrimTypeBool) || (it.PrimType() == MoPrimTypePrimTypeTag) {
		self.Fact(SemFact{Kind: SemFactPrimIdent}, self)
	}
}

func (me *SrcPack) semPopulateIdent(self *SemExpr, it MoValIdent) {
	ident := &SemValIdent{Name: it}
	self.Val = ident
}

func (me *SrcPack) semPopulateList(self *SemExpr, it *MoValList) {
	list := &SemValList{Items: make(SemExprs, len(*it))}
	self.Val = list
	for i, item := range *it {
		list.Items[i] = me.semExprFromMoExpr(self.Scope, item, self)
	}
}

func (me *SrcPack) semPopulateDict(self *SemExpr, it *MoValDict) {
	dict := &SemValDict{Keys: make(SemExprs, len(*it)), Vals: make(SemExprs, len(*it))}
	self.Val = dict
	for i, item := range *it {
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
}

func (me *SrcPack) semPopulateRootScope() {
	for name, prim_fn := range semTyPrimFns {
		fn_val := &SemValFunc{primImpl: prim_fn}
		fn := &SemExpr{Scope: &me.Trees.Sem.Scope, Type: semPrimFnTypes[name], Val: fn_val}
		fn.Type = semTypeEnsureDueTo(fn, fn.Type)
		var idx int
		fn_val.Params = SemExprs(sl.To(fn.Type.TArgs, func(t *SemType) *SemExpr {
			idx++
			return &SemExpr{Parent: fn, Scope: fn.Scope, Type: t, Val: &SemValIdent{Name: MoValIdent("arg" + str.FromInt(idx))}}
		}))
		fn_val.Params = fn_val.Params[:len(fn_val.Params)-1] // remove the return tyArg from params
		fn.Fact(SemFact{Kind: SemFactPrimFn}, fn)
		me.Trees.Sem.Scope.Own[name] = &SemScopeEntry{
			Type:                  fn.Type,
			DeclParamOrCallOrFunc: fn,
			Refs:                  map[*SemExpr]util.Void{},
		}
	}

	me.Trees.Sem.TopLevel.Walk(nil, func(self *SemExpr) bool {
		if call, _ := self.Val.(*SemValCall); call != nil {
			switch ident := call.Callee.MaybeIdent(true); ident {
			case moPrimOpQQuote, moPrimOpQuote:
				return false
			case moPrimOpFn, moPrimOpMacro:
				me.semScopePrepOnFn(self)
			case moPrimOpSet:
				me.semScopePrepOnSet(self)
			}
		}
		return true
	}, nil)
}

func (me *SrcPack) semReplaceExprValWithComputedValIfPermissible(self *SemExpr, val any, ty *SemType) {
	if self.isPrecomputedPermissible() && (ty != nil) {
		if self.ValOrig == nil {
			self.ValOrig = self.Val
		}
		if moval, is := val.(MoVal); is {
			me.semPopulateScalar(self, moval)
		} else if scalar, _ := val.(*SemValScalar); scalar != nil {
			me.semPopulateScalar(self, scalar.Value)
		} else {
			self.Val = val
		}
		self.Type = ty
		self.Fact(SemFact{Kind: SemFactPreComputed}, self.Type.DueTo)
	}
}

type SemScope struct {
	Own    map[MoValIdent]*SemScopeEntry
	Parent *SemScope `json:"-"`
}

type SemScopeEntry struct {
	DeclParamOrCallOrFunc *SemExpr
	SubsequentSetCalls    SemExprs
	Type                  *SemType
	Refs                  map[*SemExpr]util.Void
}

func (me *SemScope) Lookup(ident MoValIdent) (*SemScope, *SemScopeEntry) {
	if resolved := me.Own[ident]; resolved != nil {
		return me, resolved
	}
	if me.Parent != nil {
		return me.Parent.Lookup(ident)
	}
	return nil, nil
}

func (me *SrcPack) semScopeEntrySetType(entry *SemScopeEntry, from *SemExpr) {
	ty_old := entry.Type
	if ty := from.Type; entry.Type == nil {
		entry.Type = ty
	} else {
		entry.Type = semTypeFromMultiple(from, false, entry.Type, ty)
	}
	is_same := (ty_old == entry.Type /*incl nilness*/) || ((entry.Type != nil) && entry.Type.Eq(ty_old))
	if !is_same {
		for ref := range entry.Refs {
			ref.Type = entry.Type
			var top *SemExpr
			for p := ref.Parent; p != nil; p = p.Parent {
				top, p.Type = p, nil
			}
			if top != nil {
				me.semTypify(top, top.Scope)
			}
		}
	}
}

func (me *SrcPack) semScopePrepOnSet(self *SemExpr) {
	call := self.Val.(*SemValCall)
	if me.semCheckCount(2, 2, call.Args, self, true) {
		expr_name, expr_value := call.Args[0], call.Args[1]
		if ident := semCheckIs[SemValIdent](MoPrimTypeIdent, expr_name); ident != nil {
			ident.IsSet = true
			is_name_invalid := ident.Name.IsReserved()
			if is_name_invalid {
				self.ErrAdd(expr_name.From.SrcSpan.newDiagErr(ErrCodeReserved, ident.Name, ident.Name[0:1]))
			}

			if value_ident := expr_value.MaybeIdent(false); (value_ident != "") && (moPrimOpsLazy[value_ident] != nil) {
				self.ErrAdd(expr_value.From.SrcSpan.newDiagErr(ErrCodeNotAValue, value_ident))
			}
			if !is_name_invalid {
				scope, resolved := self.Scope.Lookup(ident.Name)
				if resolved == nil {
					ident.IsDecl = true
					self.Scope.Own[ident.Name] = &SemScopeEntry{DeclParamOrCallOrFunc: self, Refs: map[*SemExpr]util.Void{}}
				} else {
					resolved.SubsequentSetCalls = append(resolved.SubsequentSetCalls, self)
					if (scope == self.Scope) && (scope == &me.Trees.Sem.Scope) {
						err := self.From.SrcSpan.newDiagErr(ErrCodeDuplTopDecl, ident.Name)
						err.Rel = srcFileLocs([]string{str.Fmt("the other `%s` definition", ident.Name)}, resolved.DeclParamOrCallOrFunc)
						self.ErrAdd(err)
					}
				}
			}
		}
	}
}

func (me *SrcPack) semScopePrepOnFn(self *SemExpr) {
	call := self.Val.(*SemValCall)
	if me.semCheckCount(2, 2, call.Args, self, true) {
		if params_list, body_list := semCheckIs[SemValList](MoPrimTypeList, call.Args[0]), semCheckIs[SemValList](MoPrimTypeList, call.Args[1]); (params_list != nil) && (body_list != nil) {
			var ok_params SemExprs
			for _, param := range params_list.Items {
				if ident := semCheckIs[SemValIdent](MoPrimTypeIdent, param); ident != nil {
					if ident.Name.IsReserved() {
						self.ErrAdd(param.From.SrcSpan.newDiagErr(ErrCodeReserved, ident.Name, ident.Name[0:1]))
					} else {
						ident.IsParam, ident.IsDecl = true, true
						ok_params = append(ok_params, param)
					}
				}
			}
			fn := &SemValFunc{
				Scope:   &SemScope{Parent: self.Scope, Own: map[MoValIdent]*SemScopeEntry{}},
				Params:  ok_params,
				IsMacro: (call.Callee.Val.(*SemValIdent).Name == moPrimOpMacro),
			}
			for _, param := range fn.Params {
				fn.Scope.Own[param.Val.(*SemValIdent).Name] = &SemScopeEntry{DeclParamOrCallOrFunc: param, Refs: map[*SemExpr]util.Void{param: {}}}
			}
			switch len(body_list.Items) {
			case 0:
				self.ErrAdd(call.Args[1].From.SrcSpan.newDiagErr(ErrCodeExpectedFoo, "one or more expressions"))
			case 1:
				fn.Body = body_list.Items[0]
			default:
				f, p, s := call.Args[1].From, call.Args[1], fn.Scope
				expr_do := &SemExpr{From: f, Parent: p, Scope: s, Val: &SemValCall{
					Callee: &SemExpr{Val: &SemValIdent{Name: moPrimOpDo}, From: f, Parent: p, Scope: s},
					Args:   SemExprs{{Val: body_list, From: f, Parent: p, Scope: s}}}}
				fn.Body = expr_do
			}
			if (fn.Body != nil) && (len(ok_params) == len(params_list.Items)) {
				fn.Body.Walk(true, nil, func(it *SemExpr) {
					it.Scope = fn.Scope
				})
				self.Val = fn
			}
		}
	}
}
