package session

import (
	"atmo/util"
	"atmo/util/sl"
)

var (
	semEvalPrimOps     map[MoValIdent]func(*SrcPack, *SemExpr, *SemScope)
	semEvalPrimFns     map[MoValIdent]func(*SrcPack, *SemExpr, *SemScope)
	semEvalPrimFnTypes map[MoValIdent]SemType
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
	semEvalPrimFns = map[MoValIdent]func(*SrcPack, *SemExpr, *SemScope){
		moPrimFnNot:         (*SrcPack).semPrimFnNot,
		moPrimFnReplEnv:     (*SrcPack).semPrimFnReplEnv,
		moPrimFnReplPrint:   (*SrcPack).semPrimFnReplPrint,
		moPrimFnReplReset:   (*SrcPack).semPrimFnReplReset,
		moPrimFnNumUintAdd:  semPrimFnArith[MoValNumUint](MoPrimTypeNumUint, func(opl MoVal, opr MoVal) MoVal { return opl.(MoValNumUint) + opr.(MoValNumUint) }),
		moPrimFnNumUintSub:  semPrimFnArith[MoValNumUint](MoPrimTypeNumUint, func(opl MoVal, opr MoVal) MoVal { return opl.(MoValNumUint) - opr.(MoValNumUint) }),
		moPrimFnNumUintMul:  semPrimFnArith[MoValNumUint](MoPrimTypeNumUint, func(opl MoVal, opr MoVal) MoVal { return opl.(MoValNumUint) * opr.(MoValNumUint) }),
		moPrimFnNumUintDiv:  semPrimFnArith[MoValNumUint](MoPrimTypeNumUint, func(opl MoVal, opr MoVal) MoVal { return opl.(MoValNumUint) / opr.(MoValNumUint) }),
		moPrimFnNumUintMod:  semPrimFnArith[MoValNumUint](MoPrimTypeNumUint, func(opl MoVal, opr MoVal) MoVal { return opl.(MoValNumUint) % opr.(MoValNumUint) }),
		moPrimFnNumIntAdd:   semPrimFnArith[MoValNumInt](MoPrimTypeNumInt, func(opl MoVal, opr MoVal) MoVal { return opl.(MoValNumInt) + opr.(MoValNumInt) }),
		moPrimFnNumIntSub:   semPrimFnArith[MoValNumInt](MoPrimTypeNumInt, func(opl MoVal, opr MoVal) MoVal { return opl.(MoValNumInt) - opr.(MoValNumInt) }),
		moPrimFnNumIntMul:   semPrimFnArith[MoValNumInt](MoPrimTypeNumInt, func(opl MoVal, opr MoVal) MoVal { return opl.(MoValNumInt) * opr.(MoValNumInt) }),
		moPrimFnNumIntDiv:   semPrimFnArith[MoValNumInt](MoPrimTypeNumInt, func(opl MoVal, opr MoVal) MoVal { return opl.(MoValNumInt) / opr.(MoValNumInt) }),
		moPrimFnNumIntMod:   semPrimFnArith[MoValNumInt](MoPrimTypeNumInt, func(opl MoVal, opr MoVal) MoVal { return opl.(MoValNumInt) % opr.(MoValNumInt) }),
		moPrimFnNumFloatAdd: semPrimFnArith[MoValNumFloat](MoPrimTypeNumFloat, func(opl MoVal, opr MoVal) MoVal { return opl.(MoValNumFloat) + opr.(MoValNumFloat) }),
		moPrimFnNumFloatSub: semPrimFnArith[MoValNumFloat](MoPrimTypeNumFloat, func(opl MoVal, opr MoVal) MoVal { return opl.(MoValNumFloat) - opr.(MoValNumFloat) }),
		moPrimFnNumFloatMul: semPrimFnArith[MoValNumFloat](MoPrimTypeNumFloat, func(opl MoVal, opr MoVal) MoVal { return opl.(MoValNumFloat) * opr.(MoValNumFloat) }),
		moPrimFnNumFloatDiv: semPrimFnArith[MoValNumFloat](MoPrimTypeNumFloat, func(opl MoVal, opr MoVal) MoVal { return opl.(MoValNumFloat) / opr.(MoValNumFloat) }),
		moPrimFnCast:        (*SrcPack).semPrimFnCast,
	}
	{
		t, fn := semTypeNew, MoPrimTypeFunc
		t_any, t_void, t_bool, t_str, t_chr, t_int, t_uint, t_float, t_ident, t_primtypetag := t(nil, MoPrimTypeUntyped), t(nil, MoPrimTypeVoid), t(nil, MoPrimTypeBool), t(nil, MoPrimTypeStr), t(nil, MoPrimTypeChar), t(nil, MoPrimTypeNumInt), t(nil, MoPrimTypeNumUint), t(nil, MoPrimTypeNumFloat), t(nil, MoPrimTypeIdent), t(nil, MoPrimTypePrimTypeTag)
		t_ord, t_list, t_dict, t_err := t(nil, MoPrimTypeOr, t_int, t_uint, t_float, t_chr, t_str), t(nil, MoPrimTypeList, t_any), t(nil, MoPrimTypeList, t_any, t_any), semTypeNew(nil, MoPrimTypeErr, t_any)
		semEvalPrimFnTypes = map[MoValIdent]SemType{
			moPrimFnReplEnv:     t(nil, fn, t(nil, MoPrimTypeDict, t_ident, t_any)),
			moPrimFnReplPrint:   t(nil, fn, t_any, t_void),
			moPrimFnReplReset:   t(nil, fn, t_void),
			moPrimFnNumUintAdd:  t(nil, fn, t_uint, t_uint, t_uint),
			moPrimFnNumUintSub:  t(nil, fn, t_uint, t_uint, t_uint),
			moPrimFnNumUintMul:  t(nil, fn, t_uint, t_uint, t_uint),
			moPrimFnNumUintDiv:  t(nil, fn, t_uint, t_uint, t_uint),
			moPrimFnNumUintMod:  t(nil, fn, t_uint, t_uint, t_uint),
			moPrimFnNumFloatAdd: t(nil, fn, t_float, t_float, t_float),
			moPrimFnNumFloatSub: t(nil, fn, t_float, t_float, t_float),
			moPrimFnNumFloatMul: t(nil, fn, t_float, t_float, t_float),
			moPrimFnNumFloatDiv: t(nil, fn, t_float, t_float, t_float),
			moPrimFnNumIntAdd:   t(nil, fn, t_int, t_int, t_int),
			moPrimFnNumIntSub:   t(nil, fn, t_int, t_int, t_int),
			moPrimFnNumIntMul:   t(nil, fn, t_int, t_int, t_int),
			moPrimFnNumIntDiv:   t(nil, fn, t_int, t_int, t_int),
			moPrimFnNumIntMod:   t(nil, fn, t_int, t_int, t_int),
			moPrimFnCast:        t(nil, fn, t_primtypetag, t_any, t_any),
			moPrimFnNot:         t(nil, fn, t_bool, t_bool),
			moPrimFnEq:          t(nil, fn, t_any, t_any, t_bool),
			moPrimFnNeq:         t(nil, fn, t_any, t_any, t_bool),
			moPrimFnGeq:         t(nil, fn, t_ord, t_ord, t_bool),
			moPrimFnLeq:         t(nil, fn, t_ord, t_ord, t_bool),
			moPrimFnLt:          t(nil, fn, t_ord, t_ord, t_bool),
			moPrimFnGt:          t(nil, fn, t_ord, t_ord, t_bool),
			moPrimFnPrimTypeTag: t(nil, fn, t_any, t_primtypetag),
			moPrimFnListGet:     t(nil, fn, t_list, t_uint, t_any),
			moPrimFnListSet:     t(nil, fn, t_list, t_uint, t_any, t_void),
			moPrimFnListRange:   t(nil, fn, t_list, t_uint, t_uint, t_list),
			moPrimFnListLen:     t(nil, fn, t_list, t_uint),
			moPrimFnListConcat:  t(nil, fn, semTypeNew(nil, MoPrimTypeList, t_list)),
			moPrimFnDictHas:     t(nil, fn, t_dict, t_any, t_bool),
			moPrimFnDictGet:     t(nil, fn, t_dict, t_any, t_any),
			moPrimFnDictSet:     t(nil, fn, t_dict, t_any, t_any, t_void),
			moPrimFnDictDel:     t(nil, fn, t_dict, t_list, t_void),
			moPrimFnDictLen:     t(nil, fn, t_dict, t_uint),
			moPrimFnErrNew:      t(nil, fn, t_any, t_err),
			moPrimFnErrVal:      t(nil, fn, t_err, t_any),
			moPrimFnStrLen:      t(nil, fn, t_str, t_uint),
			moPrimFnStrConcat:   t(nil, fn, t_str, semTypeNew(nil, MoPrimTypeList, t_str), t_str),
			moPrimFnStrCharAt:   t(nil, fn, t_str, t_uint, t_chr),
			moPrimFnStrRange:    t(nil, fn, t_str, t_uint, t_uint, t_str),
			moPrimFnStr:         t(nil, fn, t_any, t_str),
			moPrimFnExprStr:     t(nil, fn, t_any, t_str),
			moPrimFnExprParse:   t(nil, fn, t_str, t_any),
			moPrimFnExprEval:    t(nil, fn, t_any, t_any),
		}
	}
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
			is_prim_op := semEvalPrimOps[val.Ident] != nil
			if is_prim_op {
				self.Fact(SemFact{Kind: SemFactPrimOp}, self)
			}
			self.ErrsOwn.Add(self.From.SrcSpan.newDiagErr(util.If(is_prim_op, ErrCodeNotAValue, ErrCodeUndefined), val.Ident))
		} else {
			self.Type = semTypeEnsureDueTo(self, entry.Type)
			if decl, _ := entry.DeclParamOrSetCallOrFunc.Val.(*SemValFunc); decl != nil {
				self.Fact(SemFact{Kind: SemFactPrimFn}, self)
			}
		}
	case *SemValFunc:
		me.semEval(val.Body, val.Scope)
		self.Type = semTypeNew(self, MoPrimTypeFunc, append(sl.To(val.Params, func(p *SemExpr) SemType { return p.Type }), val.Body.Type)...)
	case *SemValCall:
		var prim_op func(*SrcPack, *SemExpr, *SemScope)
		if ident := val.Callee.MaybeIdent(false); ident != "" {
			prim_op = semEvalPrimOps[ident]
		}
		if prim_op != nil {
			val.Callee.Fact(SemFact{Kind: SemFactPrimOp}, val.Callee)
			prim_op(me, self, scope)
		} else {
			self.Type = self.newUntyped()
			me.semEval(val.Callee, scope)
			sl.Each(val.Args, func(arg *SemExpr) { me.semEval(arg, scope) })
			fn, _ := val.Callee.Val.(*SemValFunc)
			if fn == nil {
				if _, entry := scope.Lookup(val.Callee.MaybeIdent(false)); (entry != nil) && (entry.Type.(*semTypeCtor).prim == MoPrimTypeFunc) {
					switch decl := entry.DeclParamOrSetCallOrFunc.Val.(type) {
					case *SemValFunc:
						fn = decl
					case *SemValIdent:
						self.ErrsOwn.Add(self.ErrNew(ErrCodeAtmoTodo, "FOO"))
					}
				}
			}
			if fn == nil {
				if !val.Callee.HasErrs() { // dont wanna be too noisy
					val.Callee.ErrsOwn.Add(val.Callee.ErrNew(ErrCodeUncallable, val.Callee.From.String()))
				}
			} else if fn.primImpl != nil {
				self.Type = semEvalPrimFnTypes[val.Callee.MaybeIdent(false)]
				fn.primImpl(me, self, scope)
			} else {
				self.ErrsOwn.Add(self.ErrNew(ErrCodeAtmoTodo, "call lambda"))
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
	name := call.Args[0].MaybeIdent(true)
	if name == "" {

	} else {
		_, entry := scope.Lookup(name)
		if entry.Type == nil {
			entry.Type = ty
		} else {
			entry.Type.(*semTypeCtor).ensureOrOr(ty)
		}
		call.Args[0].Type = entry.Type
	}
	self.Fact(SemFact{Kind: SemFactNotPure}, self)
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
					if (i < len(list.Items)-1) && !expr.HasFact(SemFactNotPure, nil, false, true) {
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
	if me.semCheckCount(2, 2, call.Args, self, true) && self.isPrecomputedPermissible() && sl.All(call.Args, func(arg *SemExpr) bool {
		val, _ := arg.Val.(*SemValScalar)
		return (val != nil) && (val.MoVal.PrimType() == MoPrimTypeBool)
	}) {
		is_and := (call.Callee.MaybeIdent(false) == moPrimOpAnd)
		all_true, any_true := true, false
		sl.Each(call.Args, func(arg *SemExpr) {
			b := bool(arg.Val.(*SemValScalar).MoVal.(MoValBool))
			any_true, all_true = any_true || b, all_true && b
		})
		if is_and {
			me.semReplaceExprValWithComputedValIfPermissible(self, MoValBool(all_true), semTypeNew(call.Args[1], MoPrimTypeBool))
		} else {
			me.semReplaceExprValWithComputedValIfPermissible(self, MoValBool(any_true), semTypeNew(call.Args[1], MoPrimTypeBool))
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
				new_ty := semTypeNew(self, MoPrimTypeOr).(*semTypeCtor)
				all_case_preds_statically_known := (!self.HasErrs())
				for i, dict_key := range dict.Keys {
					dict_val := dict.Vals[i]
					new_ty.tyArgs = append(new_ty.tyArgs, dict_val.Type)
					if dict_key.HasFact(SemFactNotPure, nil, false, true) {
						all_case_preds_statically_known = false
					}
					if me.semCheckType(dict_key, semTypeNew(call.Callee, MoPrimTypeBool)) {
						if scalar, _ := dict_key.Val.(*SemValScalar); (scalar == nil) || (scalar.MoVal.PrimType() != MoPrimTypeBool) {
							all_case_preds_statically_known = false
						} else if b := scalar.MoVal.(MoValBool); b && (new_val == nil) { // any prior @true branch _stays_ the winner
							new_val = dict_val
						} else {
							dict_val.Fact(SemFact{Kind: SemFactUnused}, dict_key)
						}
					}
				}
				if len(new_ty.tyArgs) > 0 {
					new_ty.normalizeIfOr()
					self.Type = new_ty
					if all_case_preds_statically_known && (new_val != nil) {
						me.semReplaceExprValWithComputedValIfPermissible(self, new_val.Val, new_val.Type)
					}
				}
			}
		}
	}
}

func (me *SrcPack) semPrimFnNot(self *SemExpr, scope *SemScope) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeBool)
	sl.Each(call.Args, func(arg *SemExpr) { me.semEval(arg, scope) })
	if me.semCheckCount(1, 1, call.Args, self, true) {
		if me.semCheckType(call.Args[0], self.Type) {
			if scalar, _ := call.Args[0].Val.(*SemValScalar); (scalar != nil) && (scalar.MoVal.PrimType() == MoPrimTypeBool) {
				me.semReplaceExprValWithComputedValIfPermissible(self, !scalar.MoVal.(MoValBool), nil)
			}
		}
	}
}

func (me *SrcPack) semPrimFnReplEnv(self *SemExpr, scope *SemScope) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeDict, semTypeNew(call.Callee, MoPrimTypeIdent), call.Callee.newUntyped())
	self.Fact(SemFact{Kind: SemFactNotPure}, call.Callee)
	sl.Each(call.Args, func(arg *SemExpr) { me.semEval(arg, scope) })
	_ = me.semCheckCount(0, 0, call.Args, self, true)
}

func (me *SrcPack) semPrimFnReplPrint(self *SemExpr, scope *SemScope) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeVoid)
	self.Fact(SemFact{Kind: SemFactNotPure}, call.Callee)
	sl.Each(call.Args, func(arg *SemExpr) { me.semEval(arg, scope) })
	_ = me.semCheckCount(1, 1, call.Args, self, true)
}

func (me *SrcPack) semPrimFnReplReset(self *SemExpr, scope *SemScope) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeVoid)
	self.Fact(SemFact{Kind: SemFactNotPure}, call.Callee)
	sl.Each(call.Args, func(arg *SemExpr) { me.semEval(arg, scope) })
	_ = me.semCheckCount(0, 0, call.Args, self, true)
}

func (me *SrcPack) semPrimFnCast(self *SemExpr, scope *SemScope) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeUntyped)
	sl.Each(call.Args, func(arg *SemExpr) { me.semEval(arg, scope) })
	if me.semCheckCount(2, 2, call.Args, self, true) {
		if me.semCheckType(call.Args[0], semTypeNew(call.Callee, MoPrimTypePrimTypeTag)) {

		}
	}
}

func semPrimFnArith[T MoValNumInt | MoValNumUint | MoValNumFloat](t MoValPrimType, f func(opl MoVal, opr MoVal) MoVal) func(*SrcPack, *SemExpr, *SemScope) {
	return func(me *SrcPack, self *SemExpr, scope *SemScope) {
		defer func() { // statically-computed div/mod by 0
			if err := recover(); err != nil {
				self.ErrsOwn.Add(self.ErrNew(ErrCodeComputationFailed, err))
			}
		}()
		call := self.Val.(*SemValCall)
		self.Type = semTypeNew(call.Callee, t)
		sl.Each(call.Args, func(arg *SemExpr) { me.semEval(arg, scope); me.semCheckType(arg, self.Type) })
		if me.semCheckCount(2, 2, call.Args, self, true) && self.isPrecomputedPermissible() {
			if scalar1, _ := call.Args[0].Val.(*SemValScalar); (scalar1 != nil) && (scalar1.MoVal.PrimType() == t) {
				if scalar2, _ := call.Args[1].Val.(*SemValScalar); (scalar2 != nil) && (scalar2.MoVal.PrimType() == t) {
					result := f(scalar1.MoVal, scalar2.MoVal)
					me.semReplaceExprValWithComputedValIfPermissible(self, result, nil)
				}
			}
		}
	}
}
