package session

import (
	"atmo/util"
	"atmo/util/sl"
)

var (
	semTyPrimOps   map[MoValIdent]func(*SrcPack, *SemExpr, *SemScope)
	semTyPrimFns   map[MoValIdent]func(*SrcPack, *SemExpr, *SemScope)
	semPrimFnTypes map[MoValIdent]*SemType
)

func init() {
	semTyPrimOps = map[MoValIdent]func(*SrcPack, *SemExpr, *SemScope){
		moPrimOpSet:    (*SrcPack).semTyPrimOpSet,
		moPrimOpDo:     (*SrcPack).semTyPrimOpDo,
		moPrimOpFn:     (*SrcPack).semTyPrimOpFn,
		moPrimOpAnd:    (*SrcPack).semTyPrimOpAndOr,
		moPrimOpOr:     (*SrcPack).semTyPrimOpAndOr,
		moPrimOpQuote:  (*SrcPack).semTyPrimOpQuote,
		moPrimOpQQuote: (*SrcPack).semTyPrimOpQuote,
		moPrimOpCaseOf: (*SrcPack).semTyPrimOpCaseOf,
		moPrimOpMacro:  (*SrcPack).semTyPrimOpFn,
		moPrimOpExpand: (*SrcPack).semTyPrimOpExpand,
		moPrimOpFnCall: (*SrcPack).semTyPrimOpFnCall,
	}
	semTyPrimFns = map[MoValIdent]func(*SrcPack, *SemExpr, *SemScope){
		moPrimFnNot:         (*SrcPack).semTyPrimFnNot,
		moPrimFnReplEnv:     (*SrcPack).semTyPrimFnReplEnv,
		moPrimFnReplPrint:   (*SrcPack).semTyPrimFnReplPrint,
		moPrimFnReplReset:   (*SrcPack).semTyPrimFnReplReset,
		moPrimFnNumUintAdd:  semPrimFnArith[MoValNumUint](MoPrimTypeNumUint),
		moPrimFnNumUintSub:  semPrimFnArith[MoValNumUint](MoPrimTypeNumUint),
		moPrimFnNumUintMul:  semPrimFnArith[MoValNumUint](MoPrimTypeNumUint),
		moPrimFnNumUintDiv:  semPrimFnArith[MoValNumUint](MoPrimTypeNumUint),
		moPrimFnNumUintMod:  semPrimFnArith[MoValNumUint](MoPrimTypeNumUint),
		moPrimFnNumIntAdd:   semPrimFnArith[MoValNumInt](MoPrimTypeNumInt),
		moPrimFnNumIntSub:   semPrimFnArith[MoValNumInt](MoPrimTypeNumInt),
		moPrimFnNumIntMul:   semPrimFnArith[MoValNumInt](MoPrimTypeNumInt),
		moPrimFnNumIntDiv:   semPrimFnArith[MoValNumInt](MoPrimTypeNumInt),
		moPrimFnNumIntMod:   semPrimFnArith[MoValNumInt](MoPrimTypeNumInt),
		moPrimFnNumFloatAdd: semPrimFnArith[MoValNumFloat](MoPrimTypeNumFloat),
		moPrimFnNumFloatSub: semPrimFnArith[MoValNumFloat](MoPrimTypeNumFloat),
		moPrimFnNumFloatMul: semPrimFnArith[MoValNumFloat](MoPrimTypeNumFloat),
		moPrimFnNumFloatDiv: semPrimFnArith[MoValNumFloat](MoPrimTypeNumFloat),
		moPrimFnCast:        (*SrcPack).semTyPrimFnCast,
		moPrimFnEq:          (*SrcPack).semTyPrimFnEqNeq,
		moPrimFnNeq:         (*SrcPack).semTyPrimFnEqNeq,
		moPrimFnGeq:         (*SrcPack).semTyPrimFnCmpOrd,
		moPrimFnLeq:         (*SrcPack).semTyPrimFnCmpOrd,
		moPrimFnLt:          (*SrcPack).semTyPrimFnCmpOrd,
		moPrimFnGt:          (*SrcPack).semTyPrimFnCmpOrd,
		moPrimFnPrimTypeTag: (*SrcPack).semTyPrimFnPrimTypeTag,
		moPrimFnListLen:     (*SrcPack).semTyPrimFnListLen,
		moPrimFnListGet:     (*SrcPack).semTyPrimFnListGet,
		moPrimFnListSet:     (*SrcPack).semTyPrimFnListSet,
		moPrimFnListRange:   (*SrcPack).semTyPrimFnListRange,
		moPrimFnListConcat:  (*SrcPack).semTyPrimFnListConcat,
		moPrimFnDictHas:     (*SrcPack).semTyPrimFnDictHas,
		moPrimFnDictGet:     (*SrcPack).semTyPrimFnDictGet,
		moPrimFnDictSet:     (*SrcPack).semTyPrimFnDictSet,
		moPrimFnDictDel:     (*SrcPack).semTyPrimFnDictDel,
		moPrimFnDictLen:     (*SrcPack).semTyPrimFnDictLen,
		moPrimFnErrNew:      (*SrcPack).semTyPrimFnErrNew,
		moPrimFnErrVal:      (*SrcPack).semTyPrimFnErrVal,
		moPrimFnStrConcat:   (*SrcPack).semTyPrimFnStrConcat,
		moPrimFnStrLen:      (*SrcPack).semTyPrimFnStrLen,
		moPrimFnStrCharAt:   (*SrcPack).semTyPrimFnStrCharAt,
		moPrimFnStrRange:    (*SrcPack).semTyPrimFnStrRange,
		moPrimFnStr:         (*SrcPack).semTyPrimFnStr,
		moPrimFnExprStr:     (*SrcPack).semTyPrimFnExprStr,
		moPrimFnExprParse:   (*SrcPack).semTyPrimFnExprParse,
		moPrimFnExprEval:    (*SrcPack).semTyPrimFnExprEval,
	}
	{
		t, fn := semTypeNew, MoPrimTypeFunc
		t_any, t_void, t_bool, t_str, t_chr, t_int, t_uint, t_float, t_ident, t_primtypetag := t(nil, MoPrimTypeAny), t(nil, MoPrimTypeVoid), t(nil, MoPrimTypeBool), t(nil, MoPrimTypeStr), t(nil, MoPrimTypeChar), t(nil, MoPrimTypeNumInt), t(nil, MoPrimTypeNumUint), t(nil, MoPrimTypeNumFloat), t(nil, MoPrimTypeIdent), t(nil, MoPrimTypePrimTypeTag)
		t_ord, t_list, t_dict, t_err := t(nil, MoPrimTypeOr, t_int, t_uint, t_float, t_chr, t_str), t(nil, MoPrimTypeList, t_any), t(nil, MoPrimTypeList, t_any, t_any), semTypeNew(nil, MoPrimTypeErr, t_any)
		semPrimFnTypes = map[MoValIdent]*SemType{
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
			moPrimFnListConcat:  t(nil, fn, semTypeNew(nil, MoPrimTypeList, t_list), t_list),
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

func (me *SrcPack) semTypify(self *SemExpr, scope *SemScope) {
	if (self.Type != nil) || (len(self.ErrsOwn) > 0) {
		return
	}
	// if n := me.Trees.Sem.inFlight[self]; n > 123 {
	// 	self.ErrsOwn.Add(self.ErrNew(ErrCodeAtmoTodo, "approaching infinity, please share the code leading to this"))
	// }
	// me.Trees.Sem.inFlight[self] = me.Trees.Sem.inFlight[self] + 1
	// defer func() { me.Trees.Sem.inFlight[self] = me.Trees.Sem.inFlight[self] - 1 }()
	switch val := self.Val.(type) {
	case *SemValList:
		item_types := make(sl.Of[*SemType], len(val.Items))
		for i, item := range val.Items {
			me.semTypify(item, scope)
			item_types[i] = item.Type
		}
		item_type := semTypeFromMultiple(self, true, item_types...)
		if item_type != nil {
			self.Type = semTypeNew(self, MoPrimTypeList, item_type)
		}
	case *SemValDict:
		key_types, val_types := make(sl.Of[*SemType], len(val.Keys)), make(sl.Of[*SemType], len(val.Vals))
		for i, key := range val.Keys {
			val := val.Vals[i]
			me.semTypify(key, scope)
			me.semTypify(val, scope)
			key_types[i], val_types[i] = key.Type, val.Type
		}
		key_type, val_type := semTypeFromMultiple(self, true, key_types...), semTypeFromMultiple(self, true, val_types...)
		if (key_type != nil) && (val_type != nil) {
			self.Type = semTypeNew(self, MoPrimTypeDict, key_type, val_type)
		}
	case *SemValIdent:
		_, entry := scope.Lookup(val.Name)
		if entry != nil {
			entry.Refs[self] = util.Void{}
		}
		if (entry != nil) && (entry.Type != nil) {
			self.Type = semTypeEnsureDueTo(self, entry.Type)
			if decl, _ := entry.DeclParamOrCallOrFunc.Val.(*SemValFunc); decl != nil {
				self.Fact(SemFact{Kind: SemFactPrimFn}, self)
			}
		} else {
			is_prim_op := semTyPrimOps[val.Name] != nil
			if is_prim_op {
				self.Fact(SemFact{Kind: SemFactPrimOp}, self)
			}
			self.ErrsOwn.Add(self.From.SrcSpan.newDiagErr(util.If(is_prim_op, ErrCodeNotAValue, ErrCodeUndefined), val.Name))
		}
	case *SemValFunc:
		me.semTypify(val.Body, val.Scope)
		self.Type = semTypeNew(self, MoPrimTypeFunc, append(sl.To(val.Params, func(p *SemExpr) *SemType { return p.Type }), val.Body.Type)...)
	case *SemValCall:
		var prim_op func(*SrcPack, *SemExpr, *SemScope)
		if ident := val.Callee.MaybeIdent(false); ident != "" {
			prim_op = semTyPrimOps[ident]
		}
		if prim_op != nil {
			val.Callee.Fact(SemFact{Kind: SemFactPrimOp}, val.Callee)
			prim_op(me, self, scope)
		} else {
			me.semTypify(val.Callee, scope)
			fn, _ := val.Callee.Val.(*SemValFunc)
			if fn == nil {
				if _, entry := scope.Lookup(val.Callee.MaybeIdent(false)); (entry != nil) && (entry.Type.Prim == MoPrimTypeFunc) {
					switch decl := entry.DeclParamOrCallOrFunc.Val.(type) {
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
			} else {
				args := val.Args
				if len(args) > len(fn.Params) {
					args = args[:len(fn.Params)]
				}
				sl.Each(args, func(arg *SemExpr) { me.semTypify(arg, scope) })
				if fn.primImpl != nil {
					self.Type = semPrimFnTypes[val.Callee.MaybeIdent(false)]
					fn.primImpl(me, self, scope)
				} else {
					self.ErrsOwn.Add(self.ErrNew(ErrCodeAtmoTodo, "call lambda"))
				}
			}
		}
	}
}

func (me *SrcPack) semInterpMaybe(self *SemExpr, scope *SemScope) {
	if self.isPrecomputedPermissible() && (self.From != nil) {
		result := me.Interp.ExprEval(self.From)
		if err := result.Err(); err != nil {
			self.ErrsOwn.Add(result.Err())
		} else if to_sem := me.semExprFromMoExpr(scope, result, self.Parent); (to_sem != nil) && !to_sem.HasErrs() {
			if me.semTypify(to_sem, scope); !to_sem.HasErrs() {
				me.semReplaceExprValWithComputedValIfPermissible(self, to_sem.Val, to_sem.Type)
			}
		}
	}
}

func (me *SrcPack) semTyPrimOpSet(self *SemExpr, scope *SemScope) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeVoid)
	self.Fact(SemFact{Kind: SemFactNotPure}, self)
	if len(call.Args) > 1 {
		sl.Each(call.Args[1:], func(arg *SemExpr) { me.semTypify(arg, scope) })
		if name := call.Args[0].MaybeIdent(true); (name != "") && !self.HasErrs() {
			_, entry := scope.Lookup(name)
			if entry != nil { // no need to report, other errors in the same block are the cause
				entry.Refs[call.Args[0]] = util.Void{}
				ty_old := entry.Type
				if ty := call.Args[1].Type; entry.Type == nil {
					entry.Type = ty
				} else {
					entry.Type = semTypeFromMultiple(call.Args[1], false, entry.Type, ty)
				}
				is_same := (ty_old == entry.Type /*incl nilness*/) || ((entry.Type != nil) && entry.Type.Eq(ty_old))
				if !is_same {
					me.semScopePropagateTypeChangeToRefs(entry)
				}
			}
		}
	}
}

func (me *SrcPack) semTyPrimOpDo(self *SemExpr, scope *SemScope) {
	call := self.Val.(*SemValCall)
	if me.semCheckCount(1, 1, call.Args, self, true) {
		me.semTypify(call.Args[0], scope)
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

func (me *SrcPack) semTyPrimOpFn(self *SemExpr, _ *SemScope) {
	// no-op: whenever this call happens, it's always on a broken `@fn` or `@macro` call, because
	// otherwise `semPrepScopeOnFn` would have replaced the call with a `SemValFunc` expr already
	if !self.HasErrs() {
		self.ErrsOwn.Add(self.ErrNew(ErrCodeAtmoTodo, "encountered an `@fn` call despite no errors in it"))
	}
}

func (me *SrcPack) semTyPrimOpAndOr(self *SemExpr, scope *SemScope) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeBool)
	sl.Each(call.Args, func(arg *SemExpr) { me.semTypify(arg, scope); _ = me.semCheckType(arg, self.Type) })
	_ = me.semCheckCount(2, 2, call.Args, self, true)
}

func (me *SrcPack) semTyPrimOpQuote(self *SemExpr, scope *SemScope) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeAny)
	if me.semCheckCount(1, 1, call.Args, self, true) {
		call.Args[0].Walk(false, nil, func(it *SemExpr) {
			switch val := it.Val.(type) {
			case *SemValScalar:
				it.Type = semTypeNew(call.Callee, val.Value.PrimType())
			case *SemValList:
				it.Type = semTypeNew(call.Callee, MoPrimTypeList)
			case *SemValDict:
				it.Type = semTypeNew(call.Callee, MoPrimTypeDict)
			case *SemValIdent:
				it.Type = semTypeNew(call.Callee, MoPrimTypeIdent)
			case *SemValCall:
				it.Type = semTypeNew(call.Callee, MoPrimTypeCall)
			case *SemValFunc:
				it.Type = semTypeNew(call.Callee, MoPrimTypeFunc)
				it.ErrsOwn.Add(it.ErrNew(ErrCodeAtmoTodo, "encountered a Func expr inside a quote call"))
			}
		})
		self.Type = call.Args[0].Type
	}
}

func (me *SrcPack) semTyPrimOpCaseOf(self *SemExpr, scope *SemScope) {
	call := self.Val.(*SemValCall)
	sl.Each(call.Args, func(arg *SemExpr) { me.semTypify(arg, scope) })
	if me.semCheckCount(1, 1, call.Args, self, true) {
		if dict := semCheckIs[SemValDict](MoPrimTypeDict, call.Args[0]); dict != nil {
			if me.semCheckCount(1, -1, dict.Keys, call.Args[0], false) {
				new_ty := semTypeNew(self, MoPrimTypeOr)
				for i, dict_key := range dict.Keys {
					new_ty.TArgs = append(new_ty.TArgs, dict.Vals[i].Type)
					_ = me.semCheckType(dict_key, semTypeNew(call.Callee, MoPrimTypeBool))
				}
				if len(new_ty.TArgs) > 0 {
					if !new_ty.normalizeIfAdt() {
						new_ty = nil
					}
					self.Type = new_ty
				}
			}
		}
	}
}

func (me *SrcPack) semTyPrimOpExpand(self *SemExpr, scope *SemScope) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeAny)
	sl.Each(call.Args, func(arg *SemExpr) { me.semTypify(arg, scope) })
	if me.semCheckCount(1, 1, call.Args, self, true) {
		_ = me.semCheckTypePrim(call.Args[0], call.Callee, MoPrimTypeCall, 0)
	}
}

func (me *SrcPack) semTyPrimOpFnCall(self *SemExpr, scope *SemScope) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeAny)
	sl.Each(call.Args, func(arg *SemExpr) { me.semTypify(arg, scope) })
	if me.semCheckCount(2, 2, call.Args, self, true) && me.semCheckTypePrim(call.Args[0], call.Callee, MoPrimTypeFunc, -1) && me.semCheckTypePrim(call.Args[1], call.Callee, MoPrimTypeList, -1) {
		call.Callee.Type = semTypeNew(call.Args[0], MoPrimTypeFunc, call.Args[0].Type, call.Args[1].Type, self.Type)
	}
}

func semPrimFnArith[T MoValNumInt | MoValNumUint | MoValNumFloat](t MoValPrimType) func(*SrcPack, *SemExpr, *SemScope) {
	return func(me *SrcPack, self *SemExpr, scope *SemScope) {
		call := self.Val.(*SemValCall)
		self.Type = semTypeNew(call.Callee, t)
		if me.semCheckCount(2, 2, call.Args, self, true) {
			sl.Each(call.Args, func(arg *SemExpr) { me.semCheckType(arg, self.Type) })
		}
	}
}

func (me *SrcPack) semTyPrimFnNot(self *SemExpr, scope *SemScope) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeBool)
	sl.Each(call.Args, func(arg *SemExpr) { me.semCheckType(arg, self.Type) })
	_ = me.semCheckCount(1, 1, call.Args, self, true)
}

func (me *SrcPack) semTyPrimFnReplEnv(self *SemExpr, scope *SemScope) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeDict, semTypeNew(call.Callee, MoPrimTypeIdent), semTypeNew(call.Callee, MoPrimTypeAny))
	self.Fact(SemFact{Kind: SemFactNotPure}, call.Callee)
	_ = me.semCheckCount(0, 0, call.Args, self, true)
}

func (me *SrcPack) semTyPrimFnReplPrint(self *SemExpr, scope *SemScope) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeVoid)
	self.Fact(SemFact{Kind: SemFactNotPure}, call.Callee)
	if me.semCheckCount(1, 1, call.Args, self, true) {
		call.Callee.Type = semTypeNew(call.Args[0], MoPrimTypeFunc, call.Args[0].Type, self.Type)
	}
}

func (me *SrcPack) semTyPrimFnReplReset(self *SemExpr, scope *SemScope) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeVoid)
	self.Fact(SemFact{Kind: SemFactNotPure}, call.Callee)
	_ = me.semCheckCount(0, 0, call.Args, self, true)
}

func (me *SrcPack) semTyPrimFnCast(self *SemExpr, scope *SemScope) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeAny)
	if me.semCheckCount(2, 2, call.Args, self, true) {
		ty_prim := semTypeNew(call.Callee, MoPrimTypePrimTypeTag)
		if me.semCheckType(call.Args[0], ty_prim) {
			if cast_to, _ := call.Args[0].Val.(*SemValScalar); (cast_to != nil) && (cast_to.Value.PrimType() == MoPrimTypePrimTypeTag) {
				self.Type = semTypeNew(call.Args[0], MoValPrimType(cast_to.Value.(MoValPrimTypeTag)))
				call.Callee.Type = semTypeNew(call.Args[0], MoPrimTypeFunc, ty_prim, call.Args[1].Type, self.Type)
			}
		}
	}
}

func (me *SrcPack) semTyPrimFnEqNeq(self *SemExpr, scope *SemScope) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeBool)
	if (me.semCheckCount(2, 2, call.Args, self, true)) && (call.Args[0].Type != nil) && (call.Args[1].Type != nil) {
		if me.semCheckType(call.Args[1], call.Args[0].Type) {
			call.Callee.Type = semTypeNew(call.Args[0], MoPrimTypeFunc, call.Args[0].Type, call.Args[1].Type, self.Type)
		}
	}
}

func (me *SrcPack) semTyPrimFnCmpOrd(self *SemExpr, scope *SemScope) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeBool)
	if me.semCheckCount(2, 2, call.Args, self, true) {
		if t0, t1 := call.Args[0].Type, call.Args[1].Type; (t0 != nil) && (t1 != nil) && me.semCheckType(call.Args[1], t0) {
			if lhs, rhs := semCheckIs[SemValScalar](-1, call.Args[0]), semCheckIs[SemValScalar](-1, call.Args[1]); (lhs != nil) && (rhs != nil) {
				ok_types := []MoValPrimType{MoPrimTypeChar, MoPrimTypeStr, MoPrimTypeNumFloat, MoPrimTypeNumInt, MoPrimTypeNumUint}
				if !sl.Has(ok_types, lhs.Value.PrimType()) {
					call.Args[0].ErrsOwn.Add(call.Args[0].ErrNew(ErrCodeExpectedFoo, "a comparable value here"))
				} else if !sl.Has(ok_types, rhs.Value.PrimType()) {
					call.Args[1].ErrsOwn.Add(call.Args[1].ErrNew(ErrCodeExpectedFoo, "a comparable value here"))
				} else {
					call.Callee.Type = semTypeNew(call.Args[0], MoPrimTypeFunc, call.Args[0].Type, call.Args[1].Type, self.Type)
				}
			}
		}
	}
}

func (me *SrcPack) semTyPrimFnPrimTypeTag(self *SemExpr, scope *SemScope) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypePrimTypeTag)
	if me.semCheckCount(1, 1, call.Args, self, true) {
		call.Callee.Type = semTypeNew(call.Args[0], MoPrimTypeFunc, call.Args[0].Type, semTypeNew(call.Callee, MoPrimTypePrimTypeTag))
	}
}

func (me *SrcPack) semTyPrimFnListGet(self *SemExpr, scope *SemScope) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeAny)
	if me.semCheckCount(2, 2, call.Args, self, true) {
		if me.semCheckType(call.Args[1], semTypeNew(call.Callee, MoPrimTypeNumUint)) && me.semCheckTypePrim(call.Args[0], call.Callee, MoPrimTypeList, 1) {
			ty_list := call.Args[0].Type
			self.Type = ty_list.TArgs[0]
			call.Callee.Type = semTypeNew(call.Args[0], MoPrimTypeFunc, ty_list, call.Args[1].Type, self.Type)
		}
	}
}

func (me *SrcPack) semTyPrimFnListLen(self *SemExpr, scope *SemScope) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeNumUint)
	if me.semCheckCount(1, 1, call.Args, self, true) && me.semCheckTypePrim(call.Args[0], call.Callee, MoPrimTypeList, 1) {
		call.Callee.Type = semTypeNew(call.Args[0], MoPrimTypeFunc, call.Args[0].Type, self.Type)
	}
}

func (me *SrcPack) semTyPrimFnListSet(self *SemExpr, scope *SemScope) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeVoid)
	self.Fact(SemFact{Kind: SemFactNotPure}, call.Callee)
	if me.semCheckCount(3, 3, call.Args, self, true) {
		if me.semCheckType(call.Args[1], semTypeNew(call.Callee, MoPrimTypeNumUint)) && me.semCheckTypePrim(call.Args[0], call.Callee, MoPrimTypeList, 1) {
			ty_list := call.Args[0].Type
			if !ty_list.TArgs[0].IsAny() {
				_ = me.semCheckType(call.Args[2], ty_list.TArgs[0])
			}
			call.Callee.Type = semTypeNew(call.Args[0], MoPrimTypeFunc, call.Args[0].Type, call.Args[1].Type, ty_list.TArgs[0], self.Type)
		}
	}
}

func (me *SrcPack) semTyPrimFnListRange(self *SemExpr, scope *SemScope) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeList)
	if me.semCheckCount(3, 3, call.Args, self, true) {
		if me.semCheckType(call.Args[1], semTypeNew(call.Callee, MoPrimTypeNumUint)) && me.semCheckType(call.Args[2], semTypeNew(call.Callee, MoPrimTypeNumUint)) && me.semCheckTypePrim(call.Args[0], call.Callee, MoPrimTypeList, 1) {
			ty_list := call.Args[0].Type
			self.Type = ty_list
			call.Callee.Type = semTypeNew(call.Args[0], MoPrimTypeFunc, call.Args[0].Type, call.Args[1].Type, call.Args[2].Type, self.Type)
		}
	}
}

func (me *SrcPack) semTyPrimFnListConcat(self *SemExpr, scope *SemScope) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeList, semTypeNew(call.Callee, MoPrimTypeAny))
	if me.semCheckCount(1, 1, call.Args, self, true) && me.semCheckTypePrim(call.Args[0], call.Callee, MoPrimTypeList, 1) {
		if ty_list := call.Args[0].Type; (ty_list.TArgs[0] == nil) || (ty_list.TArgs[0].Prim != MoPrimTypeList) {
			_ = me.semCheckType(call.Args[0], semTypeNew(call.Callee, MoPrimTypeList, semTypeNew(call.Callee, MoPrimTypeList, semTypeNew(call.Callee, MoPrimTypeAny)))) // to provoke type error diag
		} else {
			self.Type = ty_list.TArgs[0]
			call.Callee.Type = semTypeNew(call.Args[0], MoPrimTypeFunc, call.Args[0].Type, self.Type)
		}
	}
}

func (me *SrcPack) semTyPrimFnDictHas(self *SemExpr, scope *SemScope) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeBool)
	if me.semCheckCount(2, 2, call.Args, self, true) && me.semCheckTypePrim(call.Args[0], call.Callee, MoPrimTypeDict, 2) {
		ty_dict := call.Args[0].Type
		ty_key := ty_dict.TArgs[0]
		if ty_key.Prim == MoPrimTypeAny {
			ty_key = call.Args[1].Type
			ty_dict.TArgs[0] = ty_key
		} else {
			_ = me.semCheckType(call.Args[1], ty_dict.TArgs[0])
		}
		call.Callee.Type = semTypeNew(self, MoPrimTypeFunc, ty_dict, ty_key, self.Type)
	}
}

func (me *SrcPack) semTyPrimFnDictGet(self *SemExpr, scope *SemScope) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeAny)
	if me.semCheckCount(2, 2, call.Args, self, true) && me.semCheckTypePrim(call.Args[0], call.Callee, MoPrimTypeDict, 2) {
		ty_dict := call.Args[0].Type
		ty_key, ty_ret := ty_dict.TArgs[0], ty_dict.TArgs[1]
		if ty_key.Prim == MoPrimTypeAny {
			ty_key = call.Args[1].Type
			ty_dict.TArgs[0] = ty_key
		} else {
			_ = me.semCheckType(call.Args[1], ty_key)
		}
		ty_ret = semTypeFromMultiple(call.Callee, true, ty_ret, semTypeNew(call.Callee, MoPrimTypeVoid))
		self.Type = ty_ret
		call.Callee.Type = semTypeNew(self, MoPrimTypeFunc, ty_dict, ty_key, self.Type)
	}
}

func (me *SrcPack) semTyPrimFnDictSet(self *SemExpr, scope *SemScope) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeVoid)
	self.Fact(SemFact{Kind: SemFactNotPure}, call.Callee)
	if me.semCheckCount(3, 3, call.Args, self, true) && me.semCheckTypePrim(call.Args[0], call.Callee, MoPrimTypeDict, 2) {
		ty_dict := call.Args[0].Type
		ty_key, ty_val := ty_dict.TArgs[0], ty_dict.TArgs[1]
		if ty_key.Prim == MoPrimTypeAny {
			ty_key = call.Args[1].Type
			ty_dict.TArgs[0] = ty_key
		} else {
			_ = me.semCheckType(call.Args[1], ty_key)
		}
		if ty_val.Prim == MoPrimTypeAny {
			ty_val = call.Args[2].Type
			ty_dict.TArgs[1] = ty_val
		} else {
			_ = me.semCheckType(call.Args[2], ty_val)
		}
		call.Callee.Type = semTypeNew(self, MoPrimTypeFunc, ty_dict, ty_key, ty_val, self.Type)
	}
}

func (me *SrcPack) semTyPrimFnDictDel(self *SemExpr, scope *SemScope) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeVoid)
	self.Fact(SemFact{Kind: SemFactNotPure}, call.Callee)
	if me.semCheckCount(2, 2, call.Args, self, true) && me.semCheckTypePrim(call.Args[0], call.Callee, MoPrimTypeDict, 2) && me.semCheckTypePrim(call.Args[1], call.Callee, MoPrimTypeList, 1) {
		ty_dict := call.Args[0].Type
		ty_key := ty_dict.TArgs[0]
		if ty_key.Prim == MoPrimTypeAny {
			ty_key = call.Args[1].Type.TArgs[0]
			ty_dict.TArgs[0] = ty_key
		} else {
			_ = me.semCheckType(call.Args[1], semTypeNew(call.Args[0], MoPrimTypeList, ty_key))
		}
		call.Callee.Type = semTypeNew(self, MoPrimTypeFunc, ty_dict, ty_key, self.Type)
	}
}

func (me *SrcPack) semTyPrimFnDictLen(self *SemExpr, scope *SemScope) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeNumUint)
	if me.semCheckCount(1, 1, call.Args, self, true) && me.semCheckTypePrim(call.Args[0], call.Callee, MoPrimTypeDict, 2) {
		ty_dict := call.Args[0].Type
		call.Callee.Type = semTypeNew(call.Args[0], MoPrimTypeFunc, ty_dict, self.Type)
	}
}

func (me *SrcPack) semTyPrimFnErrNew(self *SemExpr, scope *SemScope) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeErr)
	if me.semCheckCount(1, 1, call.Args, self, true) {
		self.Type.TArgs = append(self.Type.TArgs, call.Args[0].Type)
		call.Callee.Type = semTypeNew(self, MoPrimTypeFunc, call.Args[0].Type, self.Type)
	}
}

func (me *SrcPack) semTyPrimFnErrVal(self *SemExpr, scope *SemScope) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeAny)
	if me.semCheckCount(1, 1, call.Args, self, true) {
		if me.semCheckTypePrim(call.Args[0], call.Callee, MoPrimTypeErr, 1) {
			self.Type = call.Args[0].Type.TArgs[0]
			call.Callee.Type = semTypeNew(self, MoPrimTypeFunc, call.Args[0].Type, self.Type)
		}
	}
}

func (me *SrcPack) semTyPrimFnStrConcat(self *SemExpr, scope *SemScope) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeStr)
	if me.semCheckCount(1, 1, call.Args, self, true) {
		ty_list_str := semTypeNew(call.Callee, MoPrimTypeList, semTypeNew(call.Callee, MoPrimTypeStr))
		_ = me.semCheckType(call.Args[0], ty_list_str)
		call.Callee.Type = semTypeNew(call.Callee, MoPrimTypeFunc, ty_list_str, self.Type)
	}
}

func (me *SrcPack) semTyPrimFnStrLen(self *SemExpr, scope *SemScope) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeNumUint)
	if me.semCheckCount(1, 1, call.Args, self, true) {
		ty_str := semTypeNew(call.Callee, MoPrimTypeStr)
		_ = me.semCheckType(call.Args[0], ty_str)
		call.Callee.Type = semTypeNew(call.Callee, MoPrimTypeFunc, ty_str, self.Type)
	}
}

func (me *SrcPack) semTyPrimFnStrCharAt(self *SemExpr, scope *SemScope) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeChar)
	if me.semCheckCount(2, 2, call.Args, self, true) {
		ty_str, ty_uint := semTypeNew(call.Callee, MoPrimTypeStr), semTypeNew(call.Callee, MoPrimTypeNumUint)
		_ = me.semCheckType(call.Args[0], ty_str)
		_ = me.semCheckType(call.Args[1], ty_uint)
		call.Callee.Type = semTypeNew(call.Callee, MoPrimTypeFunc, ty_str, ty_uint, self.Type)
	}
}

func (me *SrcPack) semTyPrimFnStrRange(self *SemExpr, scope *SemScope) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeStr)
	if me.semCheckCount(3, 3, call.Args, self, true) {
		ty_str, ty_uint := semTypeNew(call.Callee, MoPrimTypeStr), semTypeNew(call.Callee, MoPrimTypeNumUint)
		_ = me.semCheckType(call.Args[0], ty_str)
		_ = me.semCheckType(call.Args[1], ty_uint)
		_ = me.semCheckType(call.Args[2], ty_uint)
		call.Callee.Type = semTypeNew(call.Callee, MoPrimTypeFunc, ty_str, ty_uint, ty_uint, self.Type)
	}
}

func (me *SrcPack) semTyPrimFnStr(self *SemExpr, scope *SemScope) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeStr)
	if me.semCheckCount(1, 1, call.Args, self, true) {
		call.Callee.Type = semTypeNew(call.Callee, MoPrimTypeFunc, call.Args[0].Type, self.Type)
	}
}

func (me *SrcPack) semTyPrimFnExprStr(self *SemExpr, scope *SemScope) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeStr)
	if me.semCheckCount(1, 1, call.Args, self, true) {
		call.Callee.Type = semTypeNew(call.Callee, MoPrimTypeFunc, call.Args[0].Type, self.Type)
	}
}

func (me *SrcPack) semTyPrimFnExprParse(self *SemExpr, scope *SemScope) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeAny)
	if me.semCheckCount(1, 1, call.Args, self, true) {
		ty_str := semTypeNew(call.Callee, MoPrimTypeStr)
		_ = me.semCheckType(call.Args[0], ty_str)
		call.Callee.Type = semTypeNew(call.Callee, MoPrimTypeFunc, ty_str, self.Type)
	}
}

func (me *SrcPack) semTyPrimFnExprEval(self *SemExpr, scope *SemScope) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeAny)
	if me.semCheckCount(1, 1, call.Args, self, true) {
		call.Callee.Type = semTypeNew(call.Callee, MoPrimTypeFunc, call.Args[0].Type, self.Type)
	}
}
