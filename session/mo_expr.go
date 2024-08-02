package session

import (
	"cmp"
	"io"
	"strconv"
	"strings"

	"atmo/util"
	"atmo/util/sl"
	"atmo/util/str"
)

type moFnEager = func(ctx *Interp, env *MoEnv, args ...*MoExpr) *MoExpr
type moFnLazy = func(ctx *Interp, env *MoEnv, args ...*MoExpr) (*MoEnv, *MoExpr)

type MoValPrimType int

const (
	MoPrimTypeAny MoValPrimType = iota
	MoPrimTypeVoid
	MoPrimTypePrimTypeTag
	MoPrimTypeIdent
	MoPrimTypeBool
	MoPrimTypeNumInt
	MoPrimTypeNumUint
	MoPrimTypeNumFloat
	MoPrimTypeChar
	MoPrimTypeStr
	MoPrimTypeErr
	MoPrimTypeDict
	MoPrimTypeList
	MoPrimTypeCall
	MoPrimTypeFunc
	MoPrimTypeOr
)

func (me MoValPrimType) isAtomic() bool {
	return (me != MoPrimTypeAny) && (me != MoPrimTypeErr) && (me != MoPrimTypeList) && (me != MoPrimTypeCall) && (me != MoPrimTypeDict) && (me != MoPrimTypeFunc) && (me != MoPrimTypeOr)
}

func (me MoValPrimType) String() string { return me.Str(false) }
func (me MoValPrimType) Str(forDiag bool) string {
	switch me {
	case MoPrimTypeAny:
		return util.If(forDiag, "untyped-value", "@Any")
	case MoPrimTypePrimTypeTag:
		return util.If(forDiag, "primitive-type tag", "@PrimTypeTag")
	case MoPrimTypeIdent:
		return util.If(forDiag, "identifier", "@Ident")
	case MoPrimTypeVoid:
		return util.If(forDiag, "void", "@Void")
	case MoPrimTypeBool:
		return util.If(forDiag, "boolean", "@Bool")
	case MoPrimTypeNumInt:
		return util.If(forDiag, "signed-integer number", "@Int")
	case MoPrimTypeNumUint:
		return util.If(forDiag, "unsigned-integer number", "@Uint")
	case MoPrimTypeNumFloat:
		return util.If(forDiag, "floating-point number", "@Float")
	case MoPrimTypeChar:
		return util.If(forDiag, "character", "@Char")
	case MoPrimTypeStr:
		return util.If(forDiag, "text string", "@Str")
	case MoPrimTypeErr:
		return util.If(forDiag, "error", "@Err")
	case MoPrimTypeDict:
		return util.If(forDiag, "dictionary", "@Dict")
	case MoPrimTypeList:
		return util.If(forDiag, "list", "@List")
	case MoPrimTypeCall:
		return util.If(forDiag, "call expression", "@Call")
	case MoPrimTypeFunc:
		return util.If(forDiag, "function", "@Func")
	case MoPrimTypeOr:
		return util.If(forDiag, "union", "@Or")
	}
	return "NewAtmoBugReportPlz(" + str.FromInt(int(me)) + ")"
}

type MoVal interface {
	PrimType() MoValPrimType
}

type MoValPrimTypeTag MoValPrimType
type MoValIdent string
type MoValVoid struct{}
type MoValBool bool
type MoValNumInt int64
type MoValNumUint uint64
type MoValNumFloat float64
type MoValChar rune
type MoValStr string
type MoValErr struct{ ErrVal *MoExpr }
type MoValDict []moDictEntry
type MoValList MoExprs
type MoValCall MoExprs
type MoValFnPrim moFnEager
type MoValFnLam struct {
	Params  MoExprs // all are guaranteed to be ident before construction
	Body    *MoExpr
	Env     *MoEnv
	IsMacro bool
}

type moDictEntry struct {
	Key *MoExpr
	Val *MoExpr
}

func (MoValPrimTypeTag) PrimType() MoValPrimType { return MoPrimTypePrimTypeTag }
func (MoValIdent) PrimType() MoValPrimType       { return MoPrimTypeIdent }
func (MoValVoid) PrimType() MoValPrimType        { return MoPrimTypeVoid }
func (MoValBool) PrimType() MoValPrimType        { return MoPrimTypeBool }
func (MoValNumInt) PrimType() MoValPrimType      { return MoPrimTypeNumInt }
func (MoValNumUint) PrimType() MoValPrimType     { return MoPrimTypeNumUint }
func (MoValNumFloat) PrimType() MoValPrimType    { return MoPrimTypeNumFloat }
func (MoValChar) PrimType() MoValPrimType        { return MoPrimTypeChar }
func (MoValStr) PrimType() MoValPrimType         { return MoPrimTypeStr }
func (MoValErr) PrimType() MoValPrimType         { return MoPrimTypeErr }
func (*MoValDict) PrimType() MoValPrimType       { return MoPrimTypeDict }
func (*MoValList) PrimType() MoValPrimType       { return MoPrimTypeList }
func (MoValCall) PrimType() MoValPrimType        { return MoPrimTypeCall }
func (MoValFnPrim) PrimType() MoValPrimType      { return MoPrimTypeFunc }
func (*MoValFnLam) PrimType() MoValPrimType      { return MoPrimTypeFunc }

func (me *MoValDict) Has(key *MoExpr) bool {
	for _, entry := range *me {
		if found := entry.Key.Eq(key); found {
			return true
		}
	}
	return false
}

func (me *MoValDict) Get(key *MoExpr) *MoExpr {
	for _, entry := range *me {
		if found := entry.Key.Eq(key); found {
			return entry.Val
		}
	}
	return nil
}

func (me *MoValDict) Without(keys ...*MoExpr) *MoValDict {
	if len(keys) == 0 {
		return me
	}
	return util.Ptr(sl.Where(*me, func(entry moDictEntry) bool {
		return !sl.Any(keys, func(k *MoExpr) bool { return k.Eq(entry.Key) })
	}))
}

func (me *MoValDict) With(key *MoExpr, val *MoExpr) *MoValDict {
	ret := make(MoValDict, len(*me))
	for i, entry := range *me {
		ret[i].Key, ret[i].Val = entry.Key, entry.Val
	}
	ret.Set(key, val)
	return &ret
}

func (me *MoValDict) Del(key *MoExpr) {
	this := *me
	for i, entry := range this {
		if entry.Key.Eq(key) {
			this = append(this[:i], this[i+1:]...)
			break
		}
	}
	*me = this
}

func (me *MoValDict) Set(key *MoExpr, val *MoExpr) {
	this := *me
	var found bool
	for i, entry := range this {
		if found = entry.Key.Eq(key); found {
			this[i].Val = val
			break
		}
	}
	if !found {
		this = append(this, moDictEntry{key, val})
	}
	*me = this
}

func (me MoValIdent) IsReserved() bool {
	return (me[0] == '@') || (me[0] == ':') || (me[0] == '#') || (me[0] == '$')
}

type MoExpr struct {
	Val  MoVal
	Diag struct {
		Err *Diag
	}
	SrcSpan                          *SrcFileSpan // caution: `nil` for prims / builtins
	SrcFile                          *SrcFile     // dito
	SrcNode                          *AstNode     // dito
	PreEvalTopLevelPreEnvUnevaledYet bool
}

func (me *MoExpr) EqTrue() bool {
	val, is := me.Val.(MoValBool)
	return is && bool(val)
}
func (me *MoExpr) EqFalse() bool {
	val, is := me.Val.(MoValBool)
	return is && !bool(val)
}
func (me *MoExpr) EqVoid() bool { return (me.Val.PrimType() == MoPrimTypeVoid) }
func (me *MoExpr) Eq(to *MoExpr) bool {
	if me == to {
		return true
	}
	if (me == nil) || (to == nil) || me.Val.PrimType() != to.Val.PrimType() {
		return false
	}
	switch it := me.Val.(type) {
	case MoValErr:
		other := to.Val.(MoValErr)
		return it.ErrVal.Eq(other.ErrVal)
	case *MoValList:
		other := to.Val.(*MoValList)
		return sl.Eq(*it, *other, (*MoExpr).Eq)
	case MoValCall:
		other := to.Val.(MoValCall)
		return sl.Eq(it, other, (*MoExpr).Eq)
	case *MoValFnLam:
		other := to.Val.(*MoValFnLam)
		return it.Body.Eq(other.Body) && (it.IsMacro == other.IsMacro) && sl.Eq(it.Params, other.Params, (*MoExpr).Eq) && it.Env.eq(other.Env)
	case MoValFnPrim:
		other := to.Val.(MoValFnPrim)
		return (it == nil) && (other == nil)
	case *MoValDict:
		other := to.Val.(*MoValDict)
		if len(*it) != len(*other) {
			return false
		}
		dict := *it
		for i := range dict {
			if !sl.Any(*other, func(other moDictEntry) bool { return other.Key.Eq(dict[i].Key) && other.Val.Eq(dict[i].Val) }) {
				return false
			}
		}
		return true
	}
	return me.Val == to.Val
}

func (me *Interp) ExprCmp(it *MoExpr, to *MoExpr, diagMsgOpMoniker string) (int, *Diag) {
	switch it := it.Val.(type) {
	case MoValChar:
		if other, is := to.Val.(MoValChar); is {
			return cmp.Compare(it, other), nil
		}
	case MoValStr:
		if other, is := to.Val.(MoValStr); is {
			return cmp.Compare(it, other), nil
		}
	case MoValNumFloat:
		if other, is := to.Val.(MoValNumFloat); is {
			return cmp.Compare(it, other), nil
		}
	case MoValNumInt:
		if other, is := to.Val.(MoValNumInt); is {
			return cmp.Compare(it, other), nil
		}
	case MoValNumUint:
		if other, is := to.Val.(MoValNumUint); is {
			return cmp.Compare(it, other), nil
		}
	}
	return 0, me.diagSpan(true, false, it, to).newDiagErr(ErrCodeNotComparable, it, to, diagMsgOpMoniker)
}

func (me *Interp) exprErr(err *Diag, srcSpanCtx ...*MoExpr) *MoExpr {
	util.Assert(err != nil, nil)
	ret := me.expr(MoValErr{}, nil, nil, srcSpanCtx...)
	ret.Diag.Err = err
	return ret
}
func (me *Interp) expr(val MoVal, srcFile *SrcFile, srcSpan *SrcFileSpan, srcSpanCtx ...*MoExpr) *MoExpr {
	if srcSpan == nil {
		srcSpan = me.diagSpan(false, false, srcSpanCtx...)
	}
	if srcFile == nil {
		srcFile = me.srcFile(false, false, srcSpanCtx...)
	}
	return &MoExpr{Val: val, SrcSpan: srcSpan, SrcFile: srcFile}
}
func (me *Interp) exprFrom(expr *MoExpr, srcSpanCtx ...*MoExpr) *MoExpr {
	ret := me.expr(expr.Val, expr.SrcFile, expr.SrcSpan, srcSpanCtx...)
	ret.Diag.Err = expr.Diag.Err
	return ret
}
func (me *Interp) exprBool(b bool, srcSpanCtx ...*MoExpr) *MoExpr {
	return me.expr(MoValBool(b), nil, nil, srcSpanCtx...)
}
func (me *Interp) exprVoid(srcSpanCtx ...*MoExpr) *MoExpr {
	return me.expr(MoValVoid{}, nil, nil, srcSpanCtx...)
}

func (me *Interp) isSetCall(expr *MoExpr) (ret MoValIdent) {
	if is, _ := me.checkIsCallOnIdent(expr, moPrimOpSet, -1); is {
		if call := expr.Val.(MoValCall); len(call) > 1 {
			ret, _ = call[1].Val.(MoValIdent)
		}
	}
	return
}

func (me *MoExpr) setSrcSpanIfNone(from *MoExpr) {
	if me.SrcSpan == nil {
		me.SrcSpan = from.SrcSpan
	}
}

func (me *MoExpr) String() string {
	var buf strings.Builder
	me.WriteTo(&buf)
	return buf.String()
}

func (me *MoExpr) WriteTo(w io.StringWriter) { moValWriteTo(me.Val, w) }
func moValWriteTo(it MoVal, w io.StringWriter) {
	switch it := it.(type) {
	case MoValPrimTypeTag:
		w.WriteString(MoValPrimType(it).Str(false))
	case MoValIdent:
		w.WriteString(string(it))
	case MoValVoid:
		w.WriteString("@void")
	case MoValBool:
		w.WriteString(util.If(bool(it), "@true", "@false"))
	case MoValNumInt:
		w.WriteString(str.FromI64(int64(it), 10))
	case MoValNumUint:
		w.WriteString(str.FromU64(uint64(it), 10))
	case MoValNumFloat:
		w.WriteString(str.FromFloat(float64(it), -1))
	case MoValChar:
		w.WriteString(strconv.QuoteRune(rune(it)))
	case MoValStr:
		w.WriteString(str.Q(string(it)))
	case MoValErr:
		w.WriteString("(@Err ")
		if it.ErrVal != nil {
			it.ErrVal.WriteTo(w)
		}
		w.WriteString(")")
	case *MoValDict:
		w.WriteString("{")
		for i, item := range *it {
			if i > 0 {
				w.WriteString(", ")
			}
			item.Key.WriteTo(w)
			w.WriteString(": ")
			item.Val.WriteTo(w)
		}
		w.WriteString("}")
	case *MoValList:
		w.WriteString("[")
		for i, item := range *it {
			if i > 0 {
				w.WriteString(", ")
			}
			item.WriteTo(w)
		}
		w.WriteString("]")
	case MoValCall:
		w.WriteString("(")
		for i, item := range it {
			if i > 0 {
				w.WriteString(" ")
			}
			item.WriteTo(w)
		}
		w.WriteString(")")
	case MoValFnPrim:
		w.WriteString("<builtin>")
	case *MoValFnLam:
		w.WriteString("<lambda>")
	default:
		panic(it)
	}
}
func MoValToString(it MoVal) string {
	var buf strings.Builder
	moValWriteTo(it, &buf)
	return buf.String()
}

func (me *Interp) ExprParse(src string) (*MoExpr, *Diag) {
	me.FauxFile.Src.Ast, me.FauxFile.Src.Toks, me.FauxFile.Src.Text = nil, nil, src
	toks, errs := tokenize(me.FauxFile.FilePath, src)
	if len(errs) > 0 {
		return nil, errs[0]
	}
	me.FauxFile.Src.Toks = toks
	me.FauxFile.Src.Ast = me.FauxFile.parse()
	for _, diag := range me.FauxFile.allDiags() {
		if diag.Kind == DiagKindErr {
			return nil, diag
		}
	}
	if me.FauxFile.Src.Ast = me.FauxFile.Src.Ast.withoutComments(); len(me.FauxFile.Src.Ast) > 1 {
		return nil, me.FauxFile.Src.Ast.newDiagErr(me.FauxFile, ErrCodeExpectedFoo, str.Fmt("a single expression only, rather than %d", len(me.FauxFile.Src.Ast)))
	} else if (len(me.FauxFile.Src.Ast) == 0) || (len(me.FauxFile.Src.Ast[0].Nodes) == 0) {
		return nil, nil
	}

	expr, err := me.FauxFile.MoExprFromAstNode(me.FauxFile.Src.Ast[0])
	if err != nil {
		return nil, err
	}
	return expr, nil
}

func (me *SrcFile) MoExprFromAstNode(node *AstNode) (*MoExpr, *Diag) {
	var val MoVal
	switch node.Kind {
	case AstNodeKindComment:
		return nil, nil
	case AstNodeKindErr:
		return nil, node.newDiagErr(false, ErrCodeAtmoTodo, "newly introduced ast2mo bug (encountered error AST node, when the idea is not to run this at all with any such present in AST)")
	case AstNodeKindIdent:
		if node.IsIdentSepish() {
			return nil, node.newDiagErr(false, ErrCodeExpectedFoo, "expression instead of `"+node.Src+"` here")
		}
		switch ident := MoValIdent(node.Src); ident {
		case "@true":
			val = MoValBool(true)
		case "@false":
			val = MoValBool(false)
		case "@void":
			val = MoValVoid{}
		default:
			if (ident[0] == '@') && (len(ident) > 1) && (ident[1] >= 'A') && (ident[1] <= 'Z') {
				for _, prim_type_tag := range []MoValPrimType{
					MoPrimTypeAny,
					MoPrimTypeVoid,
					MoPrimTypePrimTypeTag,
					MoPrimTypeIdent,
					MoPrimTypeBool,
					MoPrimTypeNumInt,
					MoPrimTypeNumUint,
					MoPrimTypeNumFloat,
					MoPrimTypeChar,
					MoPrimTypeStr,
					MoPrimTypeErr,
					MoPrimTypeDict,
					MoPrimTypeList,
					MoPrimTypeCall,
					MoPrimTypeFunc,
					MoPrimTypeOr,
				} {
					if prim_type_tag.Str(false) == string(ident) {
						val = MoValPrimTypeTag(prim_type_tag)
						break
					}
				}
			}
			if val == nil {
				val = MoValIdent(node.Src)
			}
		}
	case AstNodeKindLit:
		switch it := node.Lit.(type) {
		case rune:
			val = MoValChar(it)
		case string:
			val = MoValStr(it)
		case float64:
			val = MoValNumFloat(it)
		case int64:
			val = MoValNumInt(it)
		case uint64:
			val = MoValNumUint(it)
		default:
			panic(str.Fmt("TODO: lit type %T", it))
		}
	case AstNodeKindBlockLine:
		nodes := node.Nodes.withoutComments()
		if len(nodes) == 0 {
			return nil, nil
		}
		var block_own_line MoExprs
		var block_sub_lines MoExprs
		for _, node := range nodes {
			expr, err := me.MoExprFromAstNode(node)
			if err != nil {
				return nil, err
			} else if expr != nil {
				if node.Kind == AstNodeKindBlockLine {
					if len(block_own_line) == 0 {
						return nil, node.newDiagErr(false, ErrCodeAtmoTodo, "newly introduced block-parsing bug (encountered block sub-line with no prior non-block own-line exprs)")
					}
					block_sub_lines = append(block_sub_lines, expr)
				} else if len(block_sub_lines) > 0 {
					return nil, node.newDiagErr(false, ErrCodeAtmoTodo, "newly introduced block-parsing bug (encountered non-block sub-line)")
				} else {
					block_own_line = append(block_own_line, expr)
				}
			}
		}
		call_form := MoValCall(block_own_line)
		if (len(call_form) == 1) && (len(block_sub_lines) == 0) {
			val = call_form[0].Val
		} else {
			if len(block_sub_lines) > 0 {
				call_form = append(call_form, &MoExpr{Val: util.Ptr(MoValList(block_sub_lines)), SrcFile: me,
					SrcSpan: block_sub_lines[0].SrcSpan.Expanded(block_sub_lines[len(block_sub_lines)-1].SrcSpan)})
			}
			val = call_form
		}
	case AstNodeKindGroup:
		switch {
		case node.IsSquareBrackets():
			nodes := node.Nodes.withoutComments()
			list := make(MoValList, 0, len(nodes))
			for _, node := range nodes {
				expr, err := me.MoExprFromAstNode(node)
				if err != nil {
					return nil, err
				} else if expr != nil {
					list = append(list, expr)
				}
			}
			val = &list
		case node.IsCurlyBraces():
			nodes := node.Nodes.withoutComments()
			dict := make(MoValDict, 0, len(nodes))
			for _, kv_node := range nodes {
				nodes_of_dict_entry := kv_node.Nodes.withoutComments()
				if kv_node.Kind == AstNodeKindErr {
					continue
				} else if len(nodes_of_dict_entry) != 2 {
					return nil, kv_node.newDiagErr(false, ErrCodeAtmoTodo, str.Fmt("new dict parsing bug: KV node has len %d with kind %d", len(nodes_of_dict_entry), kv_node.Kind))
				}
				expr_key, err := me.MoExprFromAstNode(nodes_of_dict_entry[0])
				if err != nil {
					return nil, err
				}
				expr_val, err := me.MoExprFromAstNode(nodes_of_dict_entry[1])
				if err != nil {
					return nil, err
				}
				if (expr_key == nil) || (expr_val == nil) {
					return nil, kv_node.newDiagErr(false, ErrCodeAtmoTodo, str.Fmt("new dict parsing bug: had only Comment-kind node for key or value"))
				} else if dict.Has(expr_key) {
					return nil, expr_key.SrcSpan.newDiagErr(ErrCodeDictDuplKey, expr_key)
				}
				dict.Set(expr_key, expr_val)
			}
			val = &dict
		default: // parensed or huddled
			nodes := node.Nodes.withoutComments()
			if len(nodes) == 1 {
				return me.MoExprFromAstNode(nodes[0])
			} else if len(nodes) == 0 {
				return nil, node.newDiagErr(false, ErrCodeExpectedFoo, "expression inside these empty parens")
			}

			call_form := make(MoValCall, 0, len(nodes))
			for _, node := range nodes {
				expr, err := me.MoExprFromAstNode(node)
				if err != nil {
					return nil, err
				} else if expr != nil {
					call_form = append(call_form, expr)
				}
			}
			val = call_form
		}
	}
	return &MoExpr{SrcNode: node, SrcFile: me, SrcSpan: util.Ptr(node.Toks.Span()), Val: val}, nil
}

func (me *MoExpr) IsErr() (ret bool) {
	return me.Diag.Err != nil
}

func (me *MoExpr) Err() (ret *Diag) {
	me.Walk(func(it *MoExpr) bool {
		if ret == nil {
			ret = it.Diag.Err
		}
		return (ret == nil)
	}, nil)
	return
}

func (me *MoExpr) HasErrs() (ret bool) {
	me.Walk(func(it *MoExpr) bool {
		ret = ret || (it.Diag.Err != nil)
		return !ret
	}, nil)
	return
}

func (me *MoExpr) Walk(onBefore func(it *MoExpr) bool, onAfter func(it *MoExpr)) {
	if onBefore != nil && !onBefore(me) {
		return
	}
	switch it := me.Val.(type) {
	case MoValCall:
		for _, item := range it {
			item.Walk(onBefore, onAfter)
		}
	case *MoValDict:
		for _, item := range *it {
			item.Key.Walk(onBefore, onAfter)
			item.Val.Walk(onBefore, onAfter)
		}
	case MoValErr:
		if it.ErrVal != nil {
			it.ErrVal.Walk(onBefore, onAfter)
		}
	case *MoValFnLam:
		for _, item := range it.Params {
			item.Walk(onBefore, onAfter)
		}
		it.Body.Walk(onBefore, onAfter)
	case *MoValList:
		for _, item := range *it {
			item.Walk(onBefore, onAfter)
		}
	}
	if onAfter != nil {
		onAfter(me)
	}
}

type MoExprs sl.Of[*MoExpr]

func (me MoExprs) AnyErrs() bool {
	return sl.Any(me, func(it *MoExpr) bool { return it.HasErrs() })
}

func (me MoExprs) Walk(filterBy *SrcFile, onBefore func(it *MoExpr) bool, onAfter func(it *MoExpr)) {
	for _, expr := range me {
		if (filterBy == nil) || (expr.SrcFile == filterBy) {
			expr.Walk(onBefore, onAfter)
		}
	}
}
