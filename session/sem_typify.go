package session

import (
	"atmo/util"
	"atmo/util/sl"
	"atmo/util/str"
	"maps"
)

var (
	semTyPrimOps   map[MoValIdent]func(*SrcPack, *SemExpr)
	semTyPrimFns   map[MoValIdent]func(*SrcPack, *SemExpr)
	semPrimFnTypes map[MoValIdent]*SemType
)

func init() {
	semTyPrimOps = map[MoValIdent]func(*SrcPack, *SemExpr){
		moPrimOpSet:      (*SrcPack).semTyPrimOpSet,
		moPrimOpDo:       (*SrcPack).semTyPrimOpDo,
		moPrimOpFn:       (*SrcPack).semTyPrimOpFn,
		moPrimOpBoolAnd:  (*SrcPack).semTyPrimOpBoolAndOr,
		moPrimOpBoolOr:   (*SrcPack).semTyPrimOpBoolAndOr,
		moPrimOpQuote:    (*SrcPack).semTyPrimOpQuote,
		moPrimOpQQuote:   (*SrcPack).semTyPrimOpQuote,
		moPrimOpBoolCond: (*SrcPack).semTyPrimOpBoolCond,
		moPrimOpFnCall:   (*SrcPack).semTyPrimOpFnCall,
	}
	semTyPrimFns = map[MoValIdent]func(*SrcPack, *SemExpr){
		moPrimFnMacro:       (*SrcPack).semTyPrimFnMacro,
		moPrimFnMacroExpand: (*SrcPack).semTyPrimFnMacroExpand,
		moPrimFnBoolNot:     (*SrcPack).semTyPrimFnNot,
		moPrimFnReplEnv:     (*SrcPack).semTyPrimFnReplEnv,
		moPrimFnReplPrint:   (*SrcPack).semTyPrimFnReplPrint,
		moPrimFnReplReset:   (*SrcPack).semTyPrimFnReplReset,
		moPrimFnNumUintAdd:  semPrimFnArith[MoValNumUint](MoPrimTypeNumUint, false, func(opL MoValNumUint, opR MoValNumUint) MoValNumUint { return opL + opR }),
		moPrimFnNumUintSub:  semPrimFnArith[MoValNumUint](MoPrimTypeNumUint, false, func(opL MoValNumUint, opR MoValNumUint) MoValNumUint { return opL - opR }),
		moPrimFnNumUintMul:  semPrimFnArith[MoValNumUint](MoPrimTypeNumUint, false, func(opL MoValNumUint, opR MoValNumUint) MoValNumUint { return opL * opR }),
		moPrimFnNumUintDiv:  semPrimFnArith[MoValNumUint](MoPrimTypeNumUint, true, func(opL MoValNumUint, opR MoValNumUint) MoValNumUint { return opL / opR }),
		moPrimFnNumUintMod:  semPrimFnArith[MoValNumUint](MoPrimTypeNumUint, true, func(opL MoValNumUint, opR MoValNumUint) MoValNumUint { return opL % opR }),
		moPrimFnNumIntAdd:   semPrimFnArith[MoValNumInt](MoPrimTypeNumInt, false, func(opL MoValNumInt, opR MoValNumInt) MoValNumInt { return opL + opR }),
		moPrimFnNumIntSub:   semPrimFnArith[MoValNumInt](MoPrimTypeNumInt, false, func(opL MoValNumInt, opR MoValNumInt) MoValNumInt { return opL - opR }),
		moPrimFnNumIntMul:   semPrimFnArith[MoValNumInt](MoPrimTypeNumInt, false, func(opL MoValNumInt, opR MoValNumInt) MoValNumInt { return opL * opR }),
		moPrimFnNumIntDiv:   semPrimFnArith[MoValNumInt](MoPrimTypeNumInt, true, func(opL MoValNumInt, opR MoValNumInt) MoValNumInt { return opL / opR }),
		moPrimFnNumIntMod:   semPrimFnArith[MoValNumInt](MoPrimTypeNumInt, true, func(opL MoValNumInt, opR MoValNumInt) MoValNumInt { return opL % opR }),
		moPrimFnNumFloatAdd: semPrimFnArith[MoValNumFloat](MoPrimTypeNumFloat, false, func(opL MoValNumFloat, opR MoValNumFloat) MoValNumFloat { return opL + opR }),
		moPrimFnNumFloatSub: semPrimFnArith[MoValNumFloat](MoPrimTypeNumFloat, false, func(opL MoValNumFloat, opR MoValNumFloat) MoValNumFloat { return opL - opR }),
		moPrimFnNumFloatMul: semPrimFnArith[MoValNumFloat](MoPrimTypeNumFloat, false, func(opL MoValNumFloat, opR MoValNumFloat) MoValNumFloat { return opL * opR }),
		moPrimFnNumFloatDiv: semPrimFnArith[MoValNumFloat](MoPrimTypeNumFloat, true, func(opL MoValNumFloat, opR MoValNumFloat) MoValNumFloat { return opL / opR }),
		moPrimFnCast:        (*SrcPack).semTyPrimFnCast,
		moPrimFnCmpEq:       (*SrcPack).semTyPrimFnCmpEqNeq,
		moPrimFnCmpNeq:      (*SrcPack).semTyPrimFnCmpEqNeq,
		moPrimFnCmpGeq:      (*SrcPack).semTyPrimFnCmpOrd,
		moPrimFnCmpLeq:      (*SrcPack).semTyPrimFnCmpOrd,
		moPrimFnCmpLt:       (*SrcPack).semTyPrimFnCmpOrd,
		moPrimFnCmpGt:       (*SrcPack).semTyPrimFnCmpOrd,
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
		moPrimFnObjNew:      (*SrcPack).semTyPrimFnObjNew,
		moPrimFnObjGet:      (*SrcPack).semTyPrimFnObjGet,
		moPrimFnObjSet:      (*SrcPack).semTyPrimFnObjSet,
		moPrimFnTupGet:      (*SrcPack).semTyPrimFnTupGet,
		moPrimFnTupSet:      (*SrcPack).semTyPrimFnTupSet,
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
		t_ord, t_list, t_tup, t_dict, t_obj, t_err, t_func := t(nil, MoPrimTypeOr, t_int, t_uint, t_float, t_chr, t_str), t(nil, MoPrimTypeList, t_any), t(nil, MoPrimTypeTup, t_any), t(nil, MoPrimTypeList, t_any, t_any), semTypeNew(nil, MoPrimTypeObj), semTypeNew(nil, MoPrimTypeErr, t_any), semTypeNew(nil, MoPrimTypeFunc)
		semPrimFnTypes = map[MoValIdent]*SemType{
			moPrimFnMacro:       t(nil, fn, t_func, t_func),
			moPrimFnMacroExpand: t(nil, fn, t(nil, MoPrimTypeCall), t_any),
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
			moPrimFnBoolNot:     t(nil, fn, t_bool, t_bool),
			moPrimFnCmpEq:       t(nil, fn, t_any, t_any, t_bool),
			moPrimFnCmpNeq:      t(nil, fn, t_any, t_any, t_bool),
			moPrimFnCmpGeq:      t(nil, fn, t_ord, t_ord, t_bool),
			moPrimFnCmpLeq:      t(nil, fn, t_ord, t_ord, t_bool),
			moPrimFnCmpLt:       t(nil, fn, t_ord, t_ord, t_bool),
			moPrimFnCmpGt:       t(nil, fn, t_ord, t_ord, t_bool),
			moPrimFnPrimTypeTag: t(nil, fn, t_any, t_primtypetag),
			moPrimFnTupGet:      t(nil, fn, t_tup, t_uint, t_any),
			moPrimFnTupSet:      t(nil, fn, t_tup, t_uint, t_any, t_void),
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
			moPrimFnObjNew:      t(nil, fn, t_dict, t_obj),
			moPrimFnObjGet:      t(nil, fn, t_obj, t_ident, t_any),
			moPrimFnObjSet:      t(nil, fn, t_obj, t_ident, t_any, t_void),
			moPrimFnErrNew:      t(nil, fn, t_any, t_err),
			moPrimFnErrVal:      t(nil, fn, t_err, t_any),
			moPrimFnStrLen:      t(nil, fn, t_str, t_uint),
			moPrimFnStrConcat:   t(nil, fn, semTypeNew(nil, MoPrimTypeList, t_str), t_str),
			moPrimFnStrCharAt:   t(nil, fn, t_str, t_uint, t_chr),
			moPrimFnStrRange:    t(nil, fn, t_str, t_uint, t_uint, t_str),
			moPrimFnStr:         t(nil, fn, t_any, t_str),
			moPrimFnExprStr:     t(nil, fn, t_any, t_str),
			moPrimFnExprParse:   t(nil, fn, t_str, t_any),
			moPrimFnExprEval:    t(nil, fn, t_any, t_any),
		}
	}
}

type semTyEnv map[MoValIdent]*SemType

func (me semTyEnv) set(name MoValIdent, ty *SemType) semTyEnv {
	ret := maps.Clone(me)
	ret[name] = ty
	return ret
}

func (me *SrcPack) semTySynth() {
	env := semTyEnv{}
	for _, top_expr := range me.Trees.Sem.TopLevel {
		top_expr.Type = me.semTypify(top_expr, env)
	}
}

func (me *SrcPack) semTypify(self *SemExpr, env semTyEnv) *SemType {
	switch val := self.Val.(type) {
	case *SemValScalar:
		self.Type = semTypeNew(self, val.Value.PrimType())
		self.Type.Singleton = val.Value
	case *SemValIdent:
		if self.Type = env[val.Name]; self.Type == nil {
			self.Type = semPrimFnTypes[val.Name]
		}
	case *SemValList:
		item_types := make(sl.Of[*SemType], len(val.Items))
		for i, item := range val.Items {
			me.semTypify(item, env)
			item_types[i] = item.Type
		}
		if val.IsTup {
			self.Type = semTypeNew(self, MoPrimTypeTup, item_types...)
		} else if item_type := semTypeOr(self, true, item_types...); item_type != nil {
			self.Type = semTypeNew(self, MoPrimTypeList, item_type)
			if len(val.Items) == 0 { // need sentinel so that `[]` (which types as `[@Any]`) will satisfy type `[@Foo]`
				self.Type.Singleton = MoValPrimTypeTag(MoPrimTypeAny)
			}
		}
	case *SemValDict:
		key_types, val_types := make(sl.Of[*SemType], len(val.Keys)), make(sl.Of[*SemType], len(val.Vals))
		for i, key := range val.Keys {
			val := val.Vals[i]
			me.semTypify(key, env)
			me.semTypify(val, env)
			key_types[i], val_types[i] = key.Type, val.Type
		}
		if !val.IsObj {
			key_type, val_type := semTypeOr(self, true, key_types...), semTypeOr(self, true, val_types...)
			self.Type = semTypeNew(self, MoPrimTypeDict, key_type, val_type)
		} else if self.Type = semTypeNew(self, MoPrimTypeObj, val_types...); self.Type != nil {
			self.Type.Fields = sl.To(val.Keys, func(it *SemExpr) MoValIdent { return it.Val.(*SemValIdent).Name })
			if len(val.Keys) == 0 { // need sentinel so that `{}` (which types as `{@Any:@Any}`) will satisfy type `{@Foo:@Bar}`
				self.Type.Singleton = MoValPrimTypeTag(MoPrimTypeAny)
			}
		}
	case *SemValFunc:
		sub_env := maps.Clone(env)
		for _, param := range val.Params {
			param.Type = semTypeNew(param, MoPrimTypeAny)
			sub_env[param.MaybeIdent(true)] = param.Type
		}
		ty_ret := me.semTypify(val.Body, sub_env)
		if ty_ret != nil {
			self.Type = semTypeNew(self, MoPrimTypeFunc, append(sl.To(val.Params, func(it *SemExpr) *SemType { return it.Type }), ty_ret)...)
		}
	case *SemValCall:
		me.semTypify(val.Callee, env)
		sl.Each(val.Args, func(it *SemExpr) { me.semTypify(it, env) })
		callee_name := val.Callee.MaybeIdent(false)
		prim_op, prim_fn := semTyPrimOps[callee_name], semTyPrimFns[callee_name]
		if (prim_op == nil) && sl.Any(val.Args, func(it *SemExpr) bool { return it.Type == nil }) {
			break
		}
		if prim_op != nil {
			val.Callee.Fact(SemFact{Kind: SemFactPrimOp}, val.Callee)
			prim_op(me, self)
		} else if prim_fn != nil {
			val.Callee.Fact(SemFact{Kind: SemFactPrimFn}, val.Callee)
			prim_fn(me, self)
		} else if val.Callee.Type != nil {
			n_params := -1
			_ = val.Callee.Type.mapIfAndOr(val.Callee, func(tyFn *SemType) *SemType {
				tyFn = semTypeEnsureDueTo(val.Callee, tyFn)
				if num_params := len(tyFn.TArgs) - 1; tyFn.Prim != MoPrimTypeFunc {
					self.ErrAdd(val.Callee.ErrNew(ErrCodeNotCallable, str.Shorten(val.Callee.String(true), 22)))
				} else if (n_params >= 0) && (num_params != n_params) {
					if n_params < 0 {
						n_params = num_params
					} else {
						self.ErrAdd(val.Callee.ErrNew(ErrCodeOrFuncsParamsMismatch, n_params, num_params))
					}
				} else if fn, _ := val.Callee.Val.(*SemValFunc); me.semCheckCount(num_params, num_params, val.Args, self, true) &&
					((fn == nil) || me.semCheckCount(len(fn.Params), len(fn.Params), val.Args, self, true)) {
					self.Type = tyFn.TArgs[num_params] // the func's return type
					var idx int
					sl.Each(val.Args, func(arg *SemExpr) {
						_ = me.semCheckType(arg, tyFn.TArgs[idx])
						idx++
					})
				}
				n_params = len(tyFn.TArgs) - 1
				return tyFn
			})
		}
	}
	return self.Type
}

func (me *SrcPack) semTyPrimOpSet(self *SemExpr) {
	call := self.Val.(*SemValCall)
	self.Fact(SemFact{Kind: SemFactNotPure}, self)
	self.Type = semTypeNew(call.Callee, MoPrimTypeVoid)
	if me.semCheckCount(2, 2, call.Args, self, true) {
		call.Args[0].Type = call.Args[1].Type
	}
}

func (me *SrcPack) semTyPrimOpDo(self *SemExpr) {
	call := self.Val.(*SemValCall)
	if me.semCheckCount(1, 1, call.Args, self, true) {
		if list := semCheckIs[SemValList](MoPrimTypeList, call.Args[0]); list != nil {
			if me.semCheckCount(1, -1, list.Items, call.Args[0], false) {
				self.Type = list.Items[len(list.Items)-1].Type
				for i, expr := range list.Items {
					if (i < len(list.Items)-1) && !expr.HasFact(SemFactNotPure, nil, false, false) {
						expr.Fact(SemFact{Kind: SemFactUnused}, expr)
					}
				}
			}
		}
	}
}

func (me *SrcPack) semTyPrimOpFn(self *SemExpr) {
	// no-op: whenever this call happens, it's always on a broken `@fn` call, because
	// otherwise `semPrepScopeOnFn` would have replaced the call with a `SemValFunc` expr already
	if !self.HasErrs() {
		self.ErrAdd(self.ErrNew(ErrCodeAtmoTodo, "encountered an `@fn` call despite no errors in it"))
	}
}

func (me *SrcPack) semTyPrimOpBoolAndOr(self *SemExpr) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeBool)
	sl.Each(call.Args, func(arg *SemExpr) { _ = me.semCheckType(arg, self.Type) })
	_ = me.semCheckCount(2, 2, call.Args, self, true)
}

func (me *SrcPack) semTyPrimOpQuote(self *SemExpr) {
	// TODO: handle unquotes if quasi-quote call
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(self, MoPrimTypeAny)
	if me.semCheckCount(1, 1, call.Args, self, true) {
		call.Args[0].Walk(nil, func(it *SemExpr) {
			switch val := it.Val.(type) {
			case *SemValScalar:
				it.Type = semTypeNew(self, val.Value.PrimType())
				it.Type.Singleton = val.Value
			case *SemValList:
				it.Type = semTypeNew(self, MoPrimTypeList)
			case *SemValDict:
				it.Type = semTypeNew(self, MoPrimTypeDict)
			case *SemValIdent:
				it.Type = semTypeNew(self, MoPrimTypeIdent)
				it.Type.Singleton = val.Name
			case *SemValCall:
				it.Type = semTypeNew(self, MoPrimTypeCall)
			case *SemValFunc:
				it.Type = semTypeNew(self, MoPrimTypeFunc)
				it.ErrAdd(it.ErrNew(ErrCodeAtmoTodo, "encountered a Func expr inside a quote call"))
			}
		})
		self.Type = call.Args[0].Type
	}
}

func (me *SrcPack) semTyPrimOpBoolCond(self *SemExpr) {
	call := self.Val.(*SemValCall)
	if me.semCheckCount(1, 1, call.Args, self, true) {
		if dict := semCheckIs[SemValDict](MoPrimTypeDict, call.Args[0]); dict != nil {
			if me.semCheckCount(1, -1, dict.Keys, call.Args[0], false) {
				new_ty := semTypeNew(self, MoPrimTypeOr)
				for i, dict_key := range dict.Keys {
					ty_val := dict.Vals[i].Type // me.semTypify(dict.Vals[i], me.semTypeNarrow(env, dict_key, true))
					new_ty.TArgs = append(new_ty.TArgs, ty_val)
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

func (me *SrcPack) semTyPrimOpFnCall(self *SemExpr) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeAny)
	if me.semCheckCount(2, 2, call.Args, self, true) {
		_ = call.Args[1].Type.mapIfAndOr(call.Callee, func(ty *SemType) *SemType {
			_ = ty.checkIsPrimElseErrOn(call.Callee, self, call.Args[1], MoPrimTypeList, 1)
			return ty
		})
		self.Type = call.Args[0].Type.mapIfAndOr(call.Callee, func(ty *SemType) *SemType {
			if ty.checkIsPrimElseErrOn(call.Callee, self, call.Args[0], MoPrimTypeFunc, -1) {
				return ty.TArgs[len(ty.TArgs)-1]
			}
			return nil
		})
	}
}

func semPrimFnArith[T MoValNumInt | MoValNumUint | MoValNumFloat](t MoValPrimType, isDivOrMod bool, do func(T, T) T) func(*SrcPack, *SemExpr) {
	return func(me *SrcPack, self *SemExpr) {
		call := self.Val.(*SemValCall)
		self.Type = semTypeNew(call.Callee, t)
		if me.semCheckCount(2, 2, call.Args, self, true) {
			ok, tl, tr := true, call.Args[0].Type, call.Args[1].Type
			if sl.Each(call.Args, func(arg *SemExpr) { ok = me.semCheckType(arg, self.Type) && ok }); ok {
				self.Type = semTypeMapIfAndOr(call.Callee, tl, tr, func(t1, t2 *SemType) *SemType {
					if (t1.Singleton != nil) && (t2.Singleton != nil) {
						var zero T
						if isDivOrMod && (t2.Singleton.(T) == zero) {
							self.ErrAdd(call.Args[1].ErrNew(ErrCodeDivModZero))
						} else {
							ret := semTypeNew(call.Args[0], self.Type.Prim)
							ret.Singleton = MoVal(do(t1.Singleton.(T), t2.Singleton.(T)))
							return ret
						}
					}
					return self.Type
				})
			}
			call.Callee.Type = semTypeNew(call.Callee, MoPrimTypeFunc, tl, tr, self.Type)
		}
	}
}

func (me *SrcPack) semTyPrimFnMacro(self *SemExpr) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeFunc)
	if me.semCheckCount(1, 1, call.Args, self, true) {
		self.Type = call.Args[0].Type.mapIfAndOr(call.Callee, func(ty *SemType) *SemType {
			if ty.checkIsPrimElseErrOn(call.Callee, self, call.Args[0], MoPrimTypeFunc, -1) {
				return ty
			}
			return nil
		})
		call.Callee.Type = semTypeNew(call.Args[0], MoPrimTypeFunc, call.Args[0].Type, self.Type)
	}
}

func (me *SrcPack) semTyPrimFnMacroExpand(self *SemExpr) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeAny)
	if me.semCheckCount(1, 1, call.Args, self, true) {
		_ = call.Args[0].Type.mapIfAndOr(call.Callee, func(ty *SemType) *SemType {
			_ = ty.checkIsPrimElseErrOn(call.Callee, self, call.Args[0], MoPrimTypeCall, 0)
			return ty
		})
		call.Callee.Type = semTypeNew(call.Args[0], MoPrimTypeFunc, call.Args[0].Type, self.Type)
	}
}

func (me *SrcPack) semTyPrimFnNot(self *SemExpr) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeBool)
	if me.semCheckCount(1, 1, call.Args, self, true) {
		me.semCheckType(call.Args[0], self.Type)
	}
}

func (me *SrcPack) semTyPrimFnReplEnv(self *SemExpr) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeDict, semTypeNew(call.Callee, MoPrimTypeIdent), semTypeNew(call.Callee, MoPrimTypeAny))
	self.Fact(SemFact{Kind: SemFactNotPure}, call.Callee)
	_ = me.semCheckCount(0, 0, call.Args, self, true)
}

func (me *SrcPack) semTyPrimFnReplPrint(self *SemExpr) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeVoid)
	self.Fact(SemFact{Kind: SemFactNotPure}, call.Callee)
	if me.semCheckCount(1, 1, call.Args, self, true) {
		call.Callee.Type = semTypeNew(call.Args[0], MoPrimTypeFunc, call.Args[0].Type, self.Type)
	}
}

func (me *SrcPack) semTyPrimFnReplReset(self *SemExpr) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeVoid)
	self.Fact(SemFact{Kind: SemFactNotPure}, call.Callee)
	_ = me.semCheckCount(0, 0, call.Args, self, true)
}

func (me *SrcPack) semTyPrimFnCast(self *SemExpr) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeAny)
	if me.semCheckCount(2, 2, call.Args, self, true) {
		ty_prim := semTypeNew(call.Callee, MoPrimTypePrimTypeTag)
		if me.semCheckType(call.Args[0], ty_prim) {
			self.Type = call.Args[0].Type.mapIfAndOr(call.Callee, func(ty *SemType) *SemType {
				if ty.Singleton != nil {
					return semTypeNew(call.Args[0], MoValPrimType(ty.Singleton.(MoValPrimTypeTag)))
				}
				return nil
			})
		}
		call.Callee.Type = semTypeNew(call.Args[0], MoPrimTypeFunc, call.Args[0].Type, call.Args[1].Type, self.Type)
	}
}

func (me *SrcPack) semTyPrimFnCmpEqNeq(self *SemExpr) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeBool)
	if me.semCheckCount(2, 2, call.Args, self, true) {
		call.Callee.Type = semTypeNew(call.Args[0], MoPrimTypeFunc, call.Args[0].Type, call.Args[1].Type, self.Type)
		if (call.Args[0].Type != nil) && (call.Args[1].Type != nil) && me.semCheckTypeLax(call.Args[1], call.Args[0].Type, true) {
			_ = me.semCheckTypeLax(call.Args[0], call.Args[1].Type, true)
		}
	}
}

func (me *SrcPack) semTyPrimFnCmpOrd(self *SemExpr) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeBool)
	if me.semCheckCount(2, 2, call.Args, self, true) {
		if t0, t1 := call.Args[0].Type, call.Args[1].Type; (t0 != nil) && (t1 != nil) && me.semCheckTypeLax(call.Args[1], t0, true) && me.semCheckTypeLax(call.Args[0], t1, true) {
			ty_cmp := semTypeNew(call.Callee, MoPrimTypeOr,
				semTypeNew(call.Callee, MoPrimTypeChar),
				semTypeNew(call.Callee, MoPrimTypeStr),
				semTypeNew(call.Callee, MoPrimTypeNumFloat),
				semTypeNew(call.Callee, MoPrimTypeNumInt),
				semTypeNew(call.Callee, MoPrimTypeNumUint),
			)
			if me.semCheckTypeLax(call.Args[0], ty_cmp, true) {
				_ = me.semCheckTypeLax(call.Args[1], ty_cmp, true)
			}
		}
		call.Callee.Type = semTypeNew(call.Args[0], MoPrimTypeFunc, call.Args[0].Type, call.Args[1].Type, self.Type)
	}
}

func (me *SrcPack) semTyPrimFnPrimTypeTag(self *SemExpr) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypePrimTypeTag)
	if me.semCheckCount(1, 1, call.Args, self, true) {
		call.Callee.Type = semTypeNew(call.Args[0], MoPrimTypeFunc, call.Args[0].Type, self.Type)
	}
}

func (me *SrcPack) semTyPrimFnTupGet(self *SemExpr) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeAny)
	if me.semCheckCount(2, 2, call.Args, self, true) {
		if me.semCheckType(call.Args[1], semTypeNew(call.Callee, MoPrimTypeNumUint)) {
			self.Type = semTypeMapIfAndOr(call.Callee, call.Args[0].Type, call.Args[1].Type, func(tyTup, tyIdx *SemType) *SemType {
				if tyTup.checkIsPrimElseErrOn(call.Callee, self, call.Args[0], MoPrimTypeTup, -1) && tyIdx.checkIsPrimElseErrOn(call.Callee, self, call.Args[1], MoPrimTypeNumUint, 0) {
					idx, has_idx := tyIdx.Singleton.(MoValNumUint)
					if (!has_idx) || (int(idx) >= len(tyTup.TArgs)) {
						self.ErrAdd(call.Args[1].ErrNew(ErrCodeNoSuchField, util.If(has_idx, MoValToString(idx), str.Shorten(call.Args[1].String(true), 11))))
						return nil
					}
					return tyTup.TArgs[idx]
				}
				return nil
			})
		}
		call.Callee.Type = semTypeNew(call.Args[0], MoPrimTypeFunc, call.Args[0].Type, call.Args[1].Type, self.Type)
	}
}

func (me *SrcPack) semTyPrimFnTupSet(self *SemExpr) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeVoid)
	self.Fact(SemFact{Kind: SemFactNotPure}, call.Callee)
	if me.semCheckCount(3, 3, call.Args, self, true) {
		if me.semCheckType(call.Args[1], semTypeNew(call.Callee, MoPrimTypeNumUint)) {
			ty_val := semTypeMapIfAndOr(call.Callee, call.Args[0].Type, call.Args[1].Type, func(tyTup, tyIdx *SemType) *SemType {
				if tyTup.checkIsPrimElseErrOn(call.Callee, self, call.Args[0], MoPrimTypeTup, -1) && tyIdx.checkIsPrimElseErrOn(call.Callee, self, call.Args[1], MoPrimTypeNumUint, 0) {
					idx, has_idx := tyIdx.Singleton.(MoValNumUint)
					if (!has_idx) || (int(idx) >= len(tyTup.TArgs)) {
						self.ErrAdd(call.Args[1].ErrNew(ErrCodeNoSuchField, util.If(has_idx, MoValToString(idx), str.Shorten(call.Args[1].String(true), 11))))
						return nil
					}
					return tyTup.TArgs[idx]
				}
				return nil
			})
			if ty_val == nil {
				self.Type = nil
			} else if me.semCheckTypeLax(call.Args[2], ty_val, true) {
				call.Callee.Type = semTypeNew(call.Args[0], MoPrimTypeFunc, call.Args[0].Type, call.Args[1].Type, call.Args[2].Type, self.Type)
			}
		}
	}
}

func (me *SrcPack) semTyPrimFnObjGet(self *SemExpr) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeAny)
	if me.semCheckCount(2, 2, call.Args, self, true) {
		self.Type = semTypeMapIfAndOr(call.Callee, call.Args[0].Type, call.Args[1].Type, func(tyObj, tyKey *SemType) *SemType {
			ident, _ := tyKey.Singleton.(MoValIdent)
			if idx := sl.IdxOf(tyObj.Fields, ident); idx >= 0 {
				return tyObj.TArgs[idx]
			}
			self.ErrAdd(call.Args[1].ErrNew(ErrCodeNoSuchField, util.If(ident != "", string(ident), str.Shorten(call.Args[1].String(true), 11))))
			return nil
		})
		if self.Type != nil {
			call.Callee.Type = semTypeNew(call.Args[0], MoPrimTypeFunc, call.Args[0].Type, call.Args[1].Type, self.Type)
		}
	}
}

func (me *SrcPack) semTyPrimFnObjSet(self *SemExpr) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeVoid)
	self.Fact(SemFact{Kind: SemFactNotPure}, call.Callee)
	if me.semCheckCount(3, 3, call.Args, self, true) {
		ty_val := semTypeMapIfAndOr(call.Callee, call.Args[0].Type, call.Args[1].Type, func(tyObj, tyKey *SemType) *SemType {
			ident, _ := tyKey.Singleton.(MoValIdent)
			if idx := sl.IdxOf(tyObj.Fields, ident); idx >= 0 {
				return tyObj.TArgs[idx]
			}
			self.ErrAdd(call.Args[1].ErrNew(ErrCodeNoSuchField, util.If(ident != "", string(ident), str.Shorten(call.Args[1].String(true), 11))))
			return nil
		})
		if ty_val == nil {
			self.Type = nil
		} else if me.semCheckTypeLax(call.Args[2], ty_val, true) {
			call.Callee.Type = semTypeNew(call.Args[0], MoPrimTypeFunc, call.Args[0].Type, call.Args[1].Type, call.Args[2].Type, self.Type)
		}
	}
}

func (me *SrcPack) semTyPrimFnObjNew(self *SemExpr) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeObj)
	if me.semCheckCount(1, 1, call.Args, self, true) {
		dict := semCheckIs[SemValDict](MoPrimTypeDict, call.Args[0])
		if dict != nil {
			for i, key := range dict.Keys {
				if ty_field := dict.Vals[i].Type; (!me.semCheckTypePrim(key, call.Callee, MoPrimTypeIdent, 0)) || (ty_field == nil) {
					self.Type.TArgs, self.Type.Fields = nil, nil
					break
				} else {
					self.Type.TArgs.Add(ty_field)
					self.Type.Fields.Add(key.UnquotedIfQuoteCall().MaybeIdent(true))
				}
			}
		}
		call.Callee.Type = semTypeNew(call.Args[0], MoPrimTypeFunc, call.Args[0].Type, self.Type)
	}
}

func (me *SrcPack) semTyPrimFnListLen(self *SemExpr) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeNumUint)
	if me.semCheckCount(1, 1, call.Args, self, true) && me.semCheckTypePrim(call.Args[0], call.Callee, MoPrimTypeList, 1) {
		call.Callee.Type = semTypeNew(call.Args[0], MoPrimTypeFunc, call.Args[0].Type, self.Type)
	}
}

func (me *SrcPack) semTyPrimFnListGet(self *SemExpr) {
	call := self.Val.(*SemValCall)
	if me.semCheckCount(2, 2, call.Args, self, true) && me.semCheckType(call.Args[1], semTypeNew(call.Callee, MoPrimTypeNumUint)) {
		self.Type = call.Args[0].Type.mapIfAndOr(call.Callee, func(tyList *SemType) *SemType {
			if !tyList.checkIsPrimElseErrOn(call.Callee, self, call.Args[0], MoPrimTypeList, 1) {
				return nil
			}
			return tyList.TArgs[0]
		})
		if self.Type != nil {
			call.Callee.Type = semTypeNew(call.Args[0], MoPrimTypeFunc, call.Args[0].Type, call.Args[1].Type, self.Type)
		}
	}
}

func (me *SrcPack) semTyPrimFnListSet(self *SemExpr) {
	call := self.Val.(*SemValCall)
	self.Fact(SemFact{Kind: SemFactNotPure}, call.Callee)
	if me.semCheckCount(3, 3, call.Args, self, true) && me.semCheckType(call.Args[1], semTypeNew(call.Callee, MoPrimTypeNumUint)) {
		ty_val := call.Args[0].Type.mapIfAndOr(call.Callee, func(tyList *SemType) *SemType {
			if !tyList.checkIsPrimElseErrOn(call.Callee, self, call.Args[0], MoPrimTypeList, 1) {
				return nil
			}
			return tyList.TArgs[0]
		})
		if (ty_val != nil) && me.semCheckTypeLax(call.Args[2], ty_val, true) {
			self.Type = semTypeNew(call.Callee, MoPrimTypeVoid)
			call.Callee.Type = semTypeNew(call.Args[0], MoPrimTypeFunc, call.Args[0].Type, call.Args[1].Type, call.Args[2].Type, self.Type)
		}
	}
}

func (me *SrcPack) semTyPrimFnListRange(self *SemExpr) {
	call := self.Val.(*SemValCall)
	if me.semCheckCount(3, 3, call.Args, self, true) {
		if me.semCheckType(call.Args[1], semTypeNew(call.Callee, MoPrimTypeNumUint)) && me.semCheckType(call.Args[2], semTypeNew(call.Callee, MoPrimTypeNumUint)) && me.semCheckTypePrim(call.Args[0], call.Callee, MoPrimTypeList, 1) {
			ty_list := call.Args[0].Type
			self.Type = ty_list
		}
		call.Callee.Type = semTypeNew(call.Args[0], MoPrimTypeFunc, call.Args[0].Type, call.Args[1].Type, call.Args[2].Type, self.Type)
	}
}

func (me *SrcPack) semTyPrimFnListConcat(self *SemExpr) {
	call := self.Val.(*SemValCall)
	if me.semCheckCount(1, 1, call.Args, self, true) && call.Args[0].Type != nil {
		ty_item := call.Args[0].Type.mapIfAndOr(call.Callee, func(tyList *SemType) *SemType {
			if !tyList.checkIsPrimElseErrOn(call.Callee, self, call.Args[0], MoPrimTypeList, 1) {
				return nil
			}
			return tyList.TArgs[0]
		})
		if self.Type = semTypeNew(call.Callee, MoPrimTypeList, ty_item); self.Type != nil {
			call.Callee.Type = semTypeNew(call.Args[0], MoPrimTypeFunc, call.Args[0].Type, self.Type)
		}
	}
}

func (me *SrcPack) semTyPrimFnDictHas(self *SemExpr) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeBool)
	if me.semCheckCount(2, 2, call.Args, self, true) && (call.Args[0].Type != nil) {
		ty_key := call.Args[0].Type.mapIfAndOr(call.Args[0], func(tyDict *SemType) *SemType {
			if tyDict.checkIsPrimElseErrOn(call.Callee, self, call.Args[0], MoPrimTypeDict, 2) {
				return tyDict.TArgs[0]
			}
			return nil
		})
		if (ty_key != nil) && me.semCheckTypeLax(call.Args[1], ty_key, true) {
			call.Callee.Type = semTypeNew(self, MoPrimTypeFunc, call.Args[0].Type, call.Args[1].Type, self.Type)
		}
	}
}

func (me *SrcPack) semTyPrimFnDictGet(self *SemExpr) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeAny)
	if me.semCheckCount(2, 2, call.Args, self, true) && (call.Args[0].Type != nil) {
		var ty_rets sl.Of[*SemType]
		ty_key := call.Args[0].Type.mapIfAndOr(call.Args[0], func(tyDict *SemType) *SemType {
			if tyDict.checkIsPrimElseErrOn(call.Callee, self, call.Args[0], MoPrimTypeDict, 2) {
				ty_rets.Add(tyDict.TArgs[1])
				return tyDict.TArgs[0]
			}
			return nil
		})
		if (ty_key != nil) && me.semCheckTypeLax(call.Args[1], ty_key, true) {
			self.Type = semTypeOr(call.Args[0], false, append(ty_rets, semTypeNew(call.Callee, MoPrimTypeVoid))...)
			call.Callee.Type = semTypeNew(self, MoPrimTypeFunc, call.Args[0].Type, call.Args[1].Type, self.Type)
		}
	}
}

func (me *SrcPack) semTyPrimFnDictSet(self *SemExpr) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeVoid)
	self.Fact(SemFact{Kind: SemFactNotPure}, call.Callee)
	if me.semCheckCount(3, 3, call.Args, self, true) && me.semCheckTypePrim(call.Args[0], call.Callee, MoPrimTypeDict, 2) {
		var ty_rets sl.Of[*SemType]
		ty_key := call.Args[0].Type.mapIfAndOr(call.Args[0], func(tyDict *SemType) *SemType {
			if tyDict.checkIsPrimElseErrOn(call.Callee, self, call.Args[0], MoPrimTypeDict, 2) {
				ty_rets.Add(tyDict.TArgs[1])
				return tyDict.TArgs[0]
			}
			return nil
		})
		if (ty_key != nil) && me.semCheckTypeLax(call.Args[1], ty_key, true) && me.semCheckTypeLax(call.Args[2], semTypeOr(call.Args[0], false, ty_rets...), true) {
			call.Callee.Type = semTypeNew(self, MoPrimTypeFunc, call.Args[0].Type, call.Args[1].Type, call.Args[2].Type, self.Type)
		}
	}
}

func (me *SrcPack) semTyPrimFnDictDel(self *SemExpr) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeVoid)
	self.Fact(SemFact{Kind: SemFactNotPure}, call.Callee)
	if me.semCheckCount(2, 2, call.Args, self, true) && me.semCheckTypePrim(call.Args[0], call.Callee, MoPrimTypeDict, 2) && me.semCheckTypePrim(call.Args[1], call.Callee, MoPrimTypeList, 1) {
		ty_keys := call.Args[0].Type.mapIfAndOr(call.Callee, func(tyDict *SemType) *SemType {
			if !tyDict.checkIsPrimElseErrOn(call.Callee, self, call.Args[0], MoPrimTypeDict, 2) {
				return nil
			}
			return tyDict.TArgs[0]
		})
		if ty_key := semTypeOr(call.Callee, false, ty_keys); ty_key != nil {
			_ = me.semCheckTypeLax(call.Args[1], semTypeNew(call.Args[0], MoPrimTypeList, ty_key), true)
		}
		call.Callee.Type = semTypeNew(self, MoPrimTypeFunc, call.Args[0].Type, call.Args[1].Type, self.Type)
	}
}

func (me *SrcPack) semTyPrimFnDictLen(self *SemExpr) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeNumUint)
	if me.semCheckCount(1, 1, call.Args, self, true) && me.semCheckTypePrim(call.Args[0], call.Callee, MoPrimTypeDict, 2) {
		ty_dict := call.Args[0].Type
		call.Callee.Type = semTypeNew(call.Args[0], MoPrimTypeFunc, ty_dict, self.Type)
	}
}

func (me *SrcPack) semTyPrimFnErrNew(self *SemExpr) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeErr)
	if me.semCheckCount(1, 1, call.Args, self, true) {
		self.Type.TArgs.Add(call.Args[0].Type)
		call.Callee.Type = semTypeNew(self, MoPrimTypeFunc, call.Args[0].Type, self.Type)
	}
}

func (me *SrcPack) semTyPrimFnErrVal(self *SemExpr) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeAny)
	if me.semCheckCount(1, 1, call.Args, self, true) {
		ty_vals := call.Args[0].Type.mapIfAndOr(call.Callee, func(ty *SemType) *SemType {
			if (ty == nil) || !ty.checkIsPrimElseErrOn(call.Callee, self, call.Args[0], MoPrimTypeErr, 1) {
				return nil
			}
			return ty.TArgs[0]
		})
		self.Type = semTypeOr(call.Callee, false, ty_vals)
		call.Callee.Type = semTypeNew(self, MoPrimTypeFunc, call.Args[0].Type, self.Type)
	}
}

func (me *SrcPack) semTyPrimFnStrConcat(self *SemExpr) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeStr)
	if me.semCheckCount(1, 1, call.Args, self, true) {
		ty_list_str := semTypeNew(call.Callee, MoPrimTypeList, semTypeNew(call.Callee, MoPrimTypeStr))
		_ = me.semCheckType(call.Args[0], ty_list_str)
		call.Callee.Type = semTypeNew(call.Callee, MoPrimTypeFunc, ty_list_str, self.Type)
	}
}

func (me *SrcPack) semTyPrimFnStrLen(self *SemExpr) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeNumUint)
	if me.semCheckCount(1, 1, call.Args, self, true) {
		ty_str := semTypeNew(call.Callee, MoPrimTypeStr)
		_ = me.semCheckType(call.Args[0], ty_str)
		call.Callee.Type = semTypeNew(call.Callee, MoPrimTypeFunc, ty_str, self.Type)
	}
}

func (me *SrcPack) semTyPrimFnStrCharAt(self *SemExpr) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeChar)
	if me.semCheckCount(2, 2, call.Args, self, true) {
		ty_str, ty_uint := semTypeNew(call.Callee, MoPrimTypeStr), semTypeNew(call.Callee, MoPrimTypeNumUint)
		_ = me.semCheckType(call.Args[0], ty_str)
		_ = me.semCheckType(call.Args[1], ty_uint)
		call.Callee.Type = semTypeNew(call.Callee, MoPrimTypeFunc, ty_str, ty_uint, self.Type)
	}
}

func (me *SrcPack) semTyPrimFnStrRange(self *SemExpr) {
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

func (me *SrcPack) semTyPrimFnStr(self *SemExpr) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeStr)
	if me.semCheckCount(1, 1, call.Args, self, true) {
		call.Callee.Type = semTypeNew(call.Callee, MoPrimTypeFunc, call.Args[0].Type, self.Type)
	}
}

func (me *SrcPack) semTyPrimFnExprStr(self *SemExpr) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeStr)
	if me.semCheckCount(1, 1, call.Args, self, true) {
		call.Callee.Type = semTypeNew(call.Callee, MoPrimTypeFunc, call.Args[0].Type, self.Type)
	}
}

func (me *SrcPack) semTyPrimFnExprParse(self *SemExpr) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeAny)
	if me.semCheckCount(1, 1, call.Args, self, true) {
		ty_str := semTypeNew(call.Callee, MoPrimTypeStr)
		_ = me.semCheckType(call.Args[0], ty_str)
		call.Callee.Type = semTypeNew(call.Callee, MoPrimTypeFunc, ty_str, self.Type)
	}
}

func (me *SrcPack) semTyPrimFnExprEval(self *SemExpr) {
	call := self.Val.(*SemValCall)
	self.Type = semTypeNew(call.Callee, MoPrimTypeAny)
	if me.semCheckCount(1, 1, call.Args, self, true) {
		call.Callee.Type = semTypeNew(call.Callee, MoPrimTypeFunc, call.Args[0].Type, self.Type)
	}
}
