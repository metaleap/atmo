package session

import (
	"atmo/util"
	"atmo/util/sl"
)

var (
	semEvalPrimOps map[MoValIdent]func(*SrcPack, *SemExpr, *SemScope)
	semEvalPrimFns map[MoValIdent]func(*SrcPack, *SemExpr, *SemScope)
)

func init() {
	semEvalPrimOps = map[MoValIdent]func(*SrcPack, *SemExpr, *SemScope){
		moPrimOpSet:    (*SrcPack).semPrimOpSet,
		moPrimOpDo:     (*SrcPack).semPrimOpDo,
		moPrimOpFn:     (*SrcPack).semPrimOpFn,
		moPrimOpAnd:    (*SrcPack).semPrimOpAndOr,
		moPrimOpOr:     (*SrcPack).semPrimOpAndOr,
		moPrimOpQuote:  (*SrcPack).semPrimOpQuote,
		moPrimOpQQuote: (*SrcPack).semPrimOpQuote,
		moPrimOpCaseOf: (*SrcPack).semPrimOpCaseOf,
	}
	semEvalPrimFns = map[MoValIdent]func(*SrcPack, *SemExpr, *SemScope){}
}

func (me *SrcPack) semEval(self *SemExpr, scope *SemScope) {
	if (self.Type != nil) || (len(self.ErrsOwn) > 0) {
		return
	}
	switch val := self.Val.(type) {
	case *SemValScalar:
		self.Type = semTypeNew(self, val.MoVal.PrimType())
	case *SemValList:
		item_types := make(sl.Of[SemType], len(val.Items))
		for i, item := range val.Items {
			me.semEval(item, scope)
			item_types[i] = item.Type
		}
		self.Type = semTypeNew(self, MoPrimTypeList, semTypeFromMultiple(self, item_types...))
	case *SemValDict:
		key_types, val_types := make(sl.Of[SemType], len(val.Keys)), make(sl.Of[SemType], len(val.Vals))
		for i, key := range val.Keys {
			val := val.Vals[i]
			me.semEval(key, scope)
			me.semEval(val, scope)
			key_types[i], val_types[i] = key.Type, val.Type
		}
		self.Type = semTypeNew(self, MoPrimTypeDict, semTypeFromMultiple(self, key_types...), semTypeFromMultiple(self, val_types...))
	case *SemValIdent:
		_, entry := scope.Lookup(val.Ident)
		if entry == nil {
			self.Type = self.newUntyped()
			self.ErrsOwn.Add(self.From.SrcSpan.newDiagErr(util.If(semEvalPrimOps[val.Ident] != nil, ErrCodeNotAValue, ErrCodeUndefined), val.Ident))
		} else {
			self.Type = entry.Type
		}
	case *SemValFunc:
		me.semEval(val.Body, val.Scope)
		self.Type = semTypeNew(self, MoPrimTypeFunc, append(sl.To(val.Params, func(p *SemExpr) SemType { return p.Type }), val.Body.Type)...)
	case *SemValCall:
		var prim_op func(*SrcPack, *SemExpr, *SemScope)
		if ident := val.Callee.MaybeIdent(); ident != "" {
			prim_op = semEvalPrimOps[ident]
		}
		if prim_op != nil {
			prim_op(me, self, scope)
		} else {
			me.semEval(val.Callee, scope)
			sl.Each(val.Args, func(arg *SemExpr) { me.semEval(arg, scope) })
			switch callee := val.Callee.Val.(type) {
			case *SemValFunc:
				// dupl := *callee
				// dupl.Scope = &SemScope{Own: maps.Clone(callee.Scope.Own), Parent: callee.Scope.Parent}
				self.ErrsOwn.Add(self.ErrNew(ErrCodeAtmoTodo, "CALL OF FUNC"))
				_ = callee
				self.Type = self.newUntyped()
			default:
				if !val.Callee.HasErrs() { // dont wanna be too noisy
					val.Callee.ErrsOwn.Add(val.Callee.ErrNew(ErrCodeUncallable, val.Callee.From.String()))
				}
				self.Type = self.newUntyped()
			}
		}
	}
}

func (me *SrcPack) semPrimOpSet(self *SemExpr, scope *SemScope) {
	// need no checks on args count or the ident being @set since those were performed by semPrepScopeOnSet
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeVoid)
	sl.Each(call.Args[1:], func(arg *SemExpr) { me.semEval(arg, scope) })
	ty := call.Args[1].Type
	_, entry := scope.Lookup(call.Args[0].Val.(*SemValIdent).Ident)
	if entry.Type == nil {
		entry.Type = ty
	} else {
		entry.Type.(*semTypeCtor).ensure(ty)
	}
	call.Args[0].Type = entry.Type
	self.Fact(SemFact{Kind: SemFactEffectful}, self)
}

func (me *SrcPack) semPrimOpDo(self *SemExpr, scope *SemScope) {
	call := self.Val.(*SemValCall)
	self.Type = self.newUntyped()
	if me.semCheckCount(1, 1, call.Args, self, true) {
		me.semEval(call.Args[0], scope)
		if list := semCheckIs[SemValList](MoPrimTypeList, call.Args[0]); list != nil {
			if me.semCheckCount(1, -1, list.Items, call.Args[0], false) {
				self.Type = list.Items[len(list.Items)-1].Type
				for i, expr := range list.Items {
					if (i < len(list.Items)-1) && (len(expr.HasFact(SemFactEffectful, nil, false, true)) == 0) {
						expr.Fact(SemFact{Kind: SemFactUnused}, expr)
					}
				}
			}
		}
	}
}

func (me *SrcPack) semPrimOpFn(self *SemExpr, _ *SemScope) {
	// whenever this func is invoked, it's always on a broken `@fn` or `@macro` call, because
	// otherwise semPrepScopeOnFn would have turned the call into a SemValFunc already
	self.Type = self.newUntyped()
}

func (me *SrcPack) semPrimOpAndOr(self *SemExpr, scope *SemScope) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeBool)
	sl.Each(call.Args, func(arg *SemExpr) {
		me.semEval(arg, scope)
		_ = me.semCheckType(arg, self.Type)
	})
	if me.semCheckCount(2, 2, call.Args, self, true) && sl.All(call.Args, func(arg *SemExpr) bool {
		val, _ := arg.Val.(*SemValScalar)
		return (val != nil) && (val.MoVal.PrimType() == MoPrimTypeBool)
	}) {
		is_and := (call.Callee.MaybeIdent() == moPrimOpAnd)
		all_true, any_true := true, false
		sl.Each(call.Args, func(arg *SemExpr) {
			b := bool(arg.Val.(*SemValScalar).MoVal.(MoValBool))
			any_true, all_true = any_true || b, all_true && b
		})
		if self.ValOrig == nil {
			self.ValOrig = self.Val
		}
		if is_and {
			me.semPopulateScalar(self, MoValBool(all_true))
		} else {
			me.semPopulateScalar(self, MoValBool(any_true))
		}
	}
}

func (me *SrcPack) semPrimOpQuote(self *SemExpr, scope *SemScope) {
	call := self.Val.(*SemValCall)
	self.Type = self.newUntyped()
	if me.semCheckCount(1, 1, call.Args, self, true) {
		switch val := call.Args[0].Val.(type) {
		case *SemValScalar:
			self.Type = semTypeNew(call.Callee, val.MoVal.PrimType())
		case *SemValList:
			self.Type = semTypeNew(call.Callee, MoPrimTypeList)
		case *SemValDict:
			self.Type = semTypeNew(call.Callee, MoPrimTypeDict)
		case *SemValIdent:
			self.Type = semTypeNew(call.Callee, MoPrimTypeIdent)
		case *SemValCall:
			self.Type = semTypeNew(call.Callee, MoPrimTypeCall)
		case *SemValFunc:
			self.ErrsOwn.Add(self.ErrNew(ErrCodeAtmoTodo, "shouldn't happen"))
			self.Type = semTypeNew(call.Callee, MoPrimTypeFunc)
		}
	}
}

func (me *SrcPack) semPrimOpCaseOf(self *SemExpr, scope *SemScope) {
	call := self.Val.(*SemValCall)
	self.Type = self.newUntyped()
	sl.Each(call.Args, func(arg *SemExpr) { me.semEval(arg, scope) })
	if me.semCheckCount(1, 1, call.Args, self, true) {
		if dict := semCheckIs[SemValDict](MoPrimTypeDict, call.Args[0]); dict != nil {
			if me.semCheckCount(1, -1, dict.Keys, call.Args[0], false) {
				var new_val *SemExpr
				var new_ty SemType
				all_case_preds_statically_known := !self.HasErrs()
				for i, key := range dict.Keys {
					val := dict.Vals[i]
					new_ty = semTypeFromMultiple(val, new_ty, val.Type)
					if me.semCheckType(key, semTypeNew(call.Callee, MoPrimTypeBool)) {
						if scalar, _ := key.Val.(*SemValScalar); (scalar == nil) || (scalar.MoVal.PrimType() != MoPrimTypeBool) {
							all_case_preds_statically_known = false
						} else if b := scalar.MoVal.(MoValBool); b {
							new_val = val
						}
					}
				}
				self.Type = new_ty
				if all_case_preds_statically_known && (new_val != nil) && (new_ty != nil) && (len(self.ErrsOwn) == 0) {
					if self.ValOrig == nil {
						self.ValOrig = self.Val
					}
					self.Val, self.Type = new_val.Val, new_val.Type
				}
			}
		}
	}
}
