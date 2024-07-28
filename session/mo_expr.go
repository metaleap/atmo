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

var (
	moPrimIdents = map[MoValIdent]*MoExpr{
		moValNone.Val.(MoValIdent):  moValNone,
		moValTrue.Val.(MoValIdent):  moValTrue,
		moValFalse.Val.(MoValIdent): moValFalse,
	}
	moValNone  = &MoExpr{Val: MoValIdent("@none")}
	moValTrue  = &MoExpr{Val: MoValIdent("@true")}
	moValFalse = &MoExpr{Val: MoValIdent("@false")}
)

type moFnEager = func(ctx *Interp, env *MoEnv, args ...*MoExpr) *MoExpr
type moFnLazy = func(ctx *Interp, env *MoEnv, args ...*MoExpr) (*MoEnv, *MoExpr)

type MoValPrimType int

const (
	MoPrimTypePrimTypeTag MoValPrimType = iota
	MoPrimTypeIdent
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
)

func (me MoValPrimType) isAtomic() bool {
	return (me != MoPrimTypeErr) && (me != MoPrimTypeList) && (me != MoPrimTypeCall) && (me != MoPrimTypeDict) && (me != MoPrimTypeFunc)
}

func (me MoValPrimType) String() string { return me.Str(false) }
func (me MoValPrimType) Str(forDiag bool) string {
	switch me {
	case MoPrimTypePrimTypeTag:
		return util.If(forDiag, "primitive-type tag", "@PrimTypeTag")
	case MoPrimTypeIdent:
		return util.If(forDiag, "identifier", "@Ident")
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
	}
	panic(me)
}

type MoVal interface {
	PrimType() MoValPrimType
}

type MoValType MoValPrimType
type MoValIdent string
type MoValNumInt int64
type MoValNumUint uint64
type MoValNumFloat float64
type MoValChar rune
type MoValStr string
type MoValErr struct{ ErrVal *MoExpr }
type MoValDict [][2]*MoExpr
type MoValList MoExprs
type MoValCall MoExprs
type MoValFnPrim moFnEager
type MoValFnLam struct {
	Params  MoExprs // all are guaranteed to be ident before construction
	Body    *MoExpr
	Env     *MoEnv
	IsMacro bool
}

func (MoValType) PrimType() MoValPrimType     { return MoPrimTypePrimTypeTag }
func (MoValIdent) PrimType() MoValPrimType    { return MoPrimTypeIdent }
func (MoValNumInt) PrimType() MoValPrimType   { return MoPrimTypeNumInt }
func (MoValNumUint) PrimType() MoValPrimType  { return MoPrimTypeNumUint }
func (MoValNumFloat) PrimType() MoValPrimType { return MoPrimTypeNumFloat }
func (MoValChar) PrimType() MoValPrimType     { return MoPrimTypeChar }
func (MoValStr) PrimType() MoValPrimType      { return MoPrimTypeStr }
func (MoValErr) PrimType() MoValPrimType      { return MoPrimTypeErr }
func (MoValDict) PrimType() MoValPrimType     { return MoPrimTypeDict }
func (MoValList) PrimType() MoValPrimType     { return MoPrimTypeList }
func (MoValCall) PrimType() MoValPrimType     { return MoPrimTypeCall }
func (MoValFnPrim) PrimType() MoValPrimType   { return MoPrimTypeFunc }
func (*MoValFnLam) PrimType() MoValPrimType   { return MoPrimTypeFunc }

func (me MoValDict) Has(key *MoExpr) bool {
	for _, pair := range me {
		if found := pair[0].Eq(key); found {
			return true
		}
	}
	return false
}

func (me MoValDict) Get(key *MoExpr) *MoExpr {
	for _, pair := range me {
		if found := pair[0].Eq(key); found {
			return pair[1]
		}
	}
	return nil
}

func (me MoValDict) Without(keys ...*MoExpr) MoValDict {
	if len(keys) == 0 {
		return me
	}
	return sl.Where(me, func(pair [2]*MoExpr) bool {
		return !sl.HasWhere(keys, func(k *MoExpr) bool { return k.Eq(pair[0]) })
	})
}

func (me MoValDict) With(key *MoExpr, val *MoExpr) MoValDict {
	ret := make(MoValDict, len(me))
	for i, pair := range me {
		k, v := *pair[0], *pair[1]
		ret[i][0], ret[i][1] = &k, &v
	}
	ret.Set(key, val)
	return ret
}

func (me *MoValDict) Set(key *MoExpr, val *MoExpr) {
	this := *me
	var found bool
	for i, pair := range this {
		if found = pair[0].Eq(key); found {
			this[i][1] = val
			break
		}
	}
	if !found {
		this = append(this, [2]*MoExpr{key, val})
	}
	*me = this
}

type MoExpr struct {
	Val  MoVal
	Diag struct {
		Err *SrcFileNotice
	}
	SrcSpan                          *SrcFileSpan // caution: `nil` for prims / builtins
	SrcFile                          *SrcFile     // dito
	PreEvalTopLevelPreEnvUnevaledYet bool
}

func (me *MoExpr) srcNode() *AstNode {
	if (me.SrcFile != nil) && (me.SrcSpan != nil) {
		me.SrcFile.NodeAtPos(me.SrcSpan.Start, false)
	}
	return nil
}

func (me *MoExpr) EqTrue() bool  { return (me == moValTrue) || (me.Val == moValTrue.Val) }
func (me *MoExpr) EqFalse() bool { return (me == moValFalse) || (me.Val == moValFalse.Val) }
func (me *MoExpr) EqNone() bool  { return (me == moValNone) || (me.Val == moValNone.Val) }
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
	case MoValList:
		other := to.Val.(MoValList)
		return sl.Eq(it, other, (*MoExpr).Eq)
	case MoValCall:
		other := to.Val.(MoValCall)
		return sl.Eq(it, other, (*MoExpr).Eq)
	case *MoValFnLam:
		other := to.Val.(*MoValFnLam)
		return it.Body.Eq(other.Body) && (it.IsMacro == other.IsMacro) && sl.Eq(it.Params, other.Params, (*MoExpr).Eq) && it.Env.eq(other.Env)
	case MoValFnPrim:
		other := to.Val.(MoValFnPrim)
		return (it == nil) && (other == nil)
	case MoValDict:
		other := to.Val.(MoValDict)
		if len(it) != len(other) {
			return false
		}
		for i := range it {
			if !sl.HasWhere(other, func(other_pair [2]*MoExpr) bool { return other_pair[0].Eq(it[i][0]) && other_pair[1].Eq(it[i][1]) }) {
				return false
			}
		}
		return true
	}
	return me.Val == to.Val
}

func (me *Interp) ExprCmp(it *MoExpr, to *MoExpr, diagMsgOpMoniker string) (int, *SrcFileNotice) {
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
	return 0, me.diagSpan(true, false, it, to).newDiagErr(NoticeCodeNotComparable, it, to, diagMsgOpMoniker)
}

func (me *Interp) exprErr(err *SrcFileNotice, srcSpanCtx ...*MoExpr) *MoExpr {
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
	return me.exprFrom(util.If(b, moValTrue, moValFalse), srcSpanCtx...)
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
	case MoValType:
		w.WriteString(MoValPrimType(it).Str(false))
	case MoValIdent:
		w.WriteString(string(it))
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
	case MoValDict:
		w.WriteString("{")
		for i, pair := range it {
			if i > 0 {
				w.WriteString(", ")
			}
			k, v := pair[0], pair[1]
			k.WriteTo(w)
			w.WriteString(": ")
			v.WriteTo(w)
		}
		w.WriteString("}")
	case MoValList:
		w.WriteString("[")
		for i, item := range it {
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

func (me *Interp) ExprParse(src string) (*MoExpr, *SrcFileNotice) {
	me.FauxFile.Src.Ast, me.FauxFile.Src.Toks, me.FauxFile.Src.Text = nil, nil, src
	toks, errs := tokenize(me.FauxFile.FilePath, src)
	if len(errs) > 0 {
		return nil, errs[0]
	}
	me.FauxFile.Src.Toks = toks
	me.FauxFile.Src.Ast = me.FauxFile.parse()
	for _, diag := range me.FauxFile.allNotices() {
		if diag.Kind == NoticeKindErr {
			return nil, diag
		}
	}
	if me.FauxFile.Src.Ast = me.FauxFile.Src.Ast.withoutComments(); len(me.FauxFile.Src.Ast) > 1 {
		return nil, me.FauxFile.Src.Ast.newDiagErr(me.FauxFile, NoticeCodeExpectedFoo, str.Fmt("a single expression only, rather than %d", len(me.FauxFile.Src.Ast)))
	} else if (len(me.FauxFile.Src.Ast) == 0) || (len(me.FauxFile.Src.Ast[0].Nodes) == 0) {
		return nil, nil
	}

	expr, err := me.FauxFile.MoExprFromAstNode(me.FauxFile.Src.Ast[0])
	if err != nil {
		return nil, err
	}
	return expr, nil
}

func (me *SrcFile) MoExprFromAstNode(node *AstNode) (*MoExpr, *SrcFileNotice) {
	var val MoVal
	switch node.Kind {
	case AstNodeKindComment:
		return nil, nil
	case AstNodeKindErr:
		return nil, node.newDiagErr(false, NoticeCodeAtmoTodo, "newly introduced ast2mo bug (encountered error AST node, when the idea is not to run this at all with any such present in AST)")
	case AstNodeKindIdent:
		if node.IsIdentSepish() {
			return nil, node.newDiagErr(false, NoticeCodeExpectedFoo, "expression instead of `"+node.Src+"` here")
		}
		val = MoValIdent(node.Src)
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
						return nil, node.newDiagErr(false, NoticeCodeAtmoTodo, "newly introduced block-parsing bug (encountered block sub-line with no prior non-block own-line exprs)")
					}
					block_sub_lines = append(block_sub_lines, expr)
				} else if len(block_sub_lines) > 0 {
					return nil, node.newDiagErr(false, NoticeCodeAtmoTodo, "newly introduced block-parsing bug (encountered non-block sub-line)")
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
				call_form = append(call_form, &MoExpr{Val: MoValList(block_sub_lines), SrcFile: me,
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
			val = list
		case node.IsCurlyBraces():
			nodes := node.Nodes.withoutComments()
			dict := make(MoValDict, 0, len(nodes))
			for _, kv_node := range nodes {
				nodes_of_pair := kv_node.Nodes.withoutComments()
				if kv_node.Kind == AstNodeKindErr {
					continue
				} else if len(nodes_of_pair) != 2 {
					return nil, kv_node.newDiagErr(false, NoticeCodeAtmoTodo, str.Fmt("new dict parsing bug: KV node has len %d with kind %d", len(nodes_of_pair), kv_node.Kind))
				}
				expr_key, err := me.MoExprFromAstNode(nodes_of_pair[0])
				if err != nil {
					return nil, err
				}
				expr_val, err := me.MoExprFromAstNode(nodes_of_pair[1])
				if err != nil {
					return nil, err
				}
				if (expr_key == nil) || (expr_val == nil) {
					return nil, kv_node.newDiagErr(false, NoticeCodeAtmoTodo, str.Fmt("new dict parsing bug: had only Comment-kind node for key or value"))
				} else if dict.Has(expr_key) {
					return nil, expr_key.SrcSpan.newDiagErr(NoticeCodeDictDuplKey, expr_key)
				}
				dict.Set(expr_key, expr_val)
			}
			val = dict
		default: // parensed or huddled
			nodes := node.Nodes.withoutComments()
			if len(nodes) == 1 {
				return me.MoExprFromAstNode(nodes[0])
			} else if len(nodes) == 0 {
				return nil, node.newDiagErr(false, NoticeCodeExpectedFoo, "expression inside these empty parens")
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
	return &MoExpr{SrcFile: me, SrcSpan: util.Ptr(node.Toks.Span()), Val: val}, nil
}

func (me *MoExpr) IsErr() (ret bool) {
	return me.Diag.Err != nil
}

func (me *MoExpr) Err() (ret *SrcFileNotice) {
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
	case MoValDict:
		for _, pair := range it {
			pair[0].Walk(onBefore, onAfter)
			pair[1].Walk(onBefore, onAfter)
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
	case MoValList:
		for _, item := range it {
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
