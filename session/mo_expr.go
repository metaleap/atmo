package session

import (
	"cmp"
	"fmt"
	"io"
	"strconv"
	"strings"

	"atmo/util"
	"atmo/util/sl"
	"atmo/util/str"
)

var (
	moPrimIdents = map[moValIdent]*MoExpr{
		moValNone.Val.(moValIdent):  moValNone,
		moValTrue.Val.(moValIdent):  moValTrue,
		moValFalse.Val.(moValIdent): moValFalse,
	}
	moValNone  = &MoExpr{Val: moValIdent("@none")}
	moValTrue  = &MoExpr{Val: moValIdent("@true")}
	moValFalse = &MoExpr{Val: moValIdent("@false")}
)

type moFnEager = func(ctx *Interp, env *MoEnv, args ...*MoExpr) (*MoExpr, *SrcFileNotice)
type moFnLazy = func(ctx *Interp, env *MoEnv, args ...*MoExpr) (*MoEnv, *MoExpr, *SrcFileNotice)

type MoValPrimType int

const (
	MoPrimTypeType MoValPrimType = iota
	MoPrimTypeIdent
	MoPrimTypeInt
	MoPrimTypeUint
	MoPrimTypeFloat
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
	case MoPrimTypeType:
		return util.If(forDiag, "primitive-type tag", "@PrimTypeTag")
	case MoPrimTypeIdent:
		return util.If(forDiag, "quoted-identifier", "@Ident")
	case MoPrimTypeInt:
		return util.If(forDiag, "signed integer number", "@Int")
	case MoPrimTypeUint:
		return util.If(forDiag, "unsigned integer number", "@Uint")
	case MoPrimTypeFloat:
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
	fmt.Stringer
	primType() MoValPrimType
}

type moValType MoValPrimType
type moValIdent string
type moValInt int64
type moValUint uint64
type moValFloat float64
type moValChar rune
type moValStr string
type moValErr struct{ Err *MoExpr }
type moValDict [][2]*MoExpr
type moValList MoExprs
type moValCall MoExprs
type moValFnPrim moFnEager
type moValFnLam struct {
	params              MoExprs // all are guaranteed to be ident before construction
	body                *MoExpr
	env                 *MoEnv
	isMacro             bool
	isLazyPrimOpWrapper bool
}

func (moValType) primType() MoValPrimType   { return MoPrimTypeType }
func (moValIdent) primType() MoValPrimType  { return MoPrimTypeIdent }
func (moValInt) primType() MoValPrimType    { return MoPrimTypeInt }
func (moValUint) primType() MoValPrimType   { return MoPrimTypeUint }
func (moValFloat) primType() MoValPrimType  { return MoPrimTypeFloat }
func (moValChar) primType() MoValPrimType   { return MoPrimTypeChar }
func (moValStr) primType() MoValPrimType    { return MoPrimTypeStr }
func (moValErr) primType() MoValPrimType    { return MoPrimTypeErr }
func (moValDict) primType() MoValPrimType   { return MoPrimTypeDict }
func (moValList) primType() MoValPrimType   { return MoPrimTypeList }
func (moValCall) primType() MoValPrimType   { return MoPrimTypeCall }
func (moValFnPrim) primType() MoValPrimType { return MoPrimTypeFunc }
func (*moValFnLam) primType() MoValPrimType { return MoPrimTypeFunc }
func (me moValType) String() string         { return moValToString(me) }
func (me moValIdent) String() string        { return moValToString(me) }
func (me moValInt) String() string          { return moValToString(me) }
func (me moValUint) String() string         { return moValToString(me) }
func (me moValFloat) String() string        { return moValToString(me) }
func (me moValChar) String() string         { return moValToString(me) }
func (me moValStr) String() string          { return moValToString(me) }
func (me moValErr) String() string          { return moValToString(me) }
func (me moValDict) String() string         { return moValToString(me) }
func (me moValList) String() string         { return moValToString(me) }
func (me moValCall) String() string         { return moValToString(me) }
func (me moValFnPrim) String() string       { return moValToString(me) }
func (me *moValFnLam) String() string       { return moValToString(me) }

func (me moValDict) Has(key *MoExpr) bool {
	for _, pair := range me {
		if found := pair[0].eq(key); found {
			return true
		}
	}
	return false
}

func (me moValDict) Get(key *MoExpr) *MoExpr {
	for _, pair := range me {
		if found := pair[0].eq(key); found {
			return pair[1]
		}
	}
	return nil
}

func (me moValDict) Without(keys ...*MoExpr) moValDict {
	if len(keys) == 0 {
		return me
	}
	return sl.Where(me, func(pair [2]*MoExpr) bool {
		return !sl.HasWhere(keys, func(k *MoExpr) bool { return k.eq(pair[0]) })
	})
}

func (me moValDict) With(key *MoExpr, val *MoExpr) moValDict {
	ret := make(moValDict, len(me))
	for i, pair := range me {
		k, v := *pair[0], *pair[1]
		ret[i][0], ret[i][1] = &k, &v
	}
	ret.Set(key, val)
	return ret
}

func (me *moValDict) Set(key *MoExpr, val *MoExpr) {
	this := *me
	var found bool
	for i, pair := range this {
		if found = pair[0].eq(key); found {
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
	Val     MoVal
	SrcSpan *SrcFileSpan `json:"-"` // caution: `nil` for prims / builtins
	SrcFile *SrcFile     `json:"-"` // dito
}

func (me *MoExpr) srcNode() *AstNode {
	if (me.SrcFile != nil) && (me.SrcSpan != nil) {
		me.SrcFile.NodeAtPos(me.SrcSpan.Start, false)
	}
	return nil
}

func (me *MoExpr) Callee() *MoExpr {
	if call, is := me.Val.(moValCall); is {
		return call[0]
	}
	return nil
}

func (me *MoExpr) eqTrue() bool  { return (me == moValTrue) || (me.Val == moValTrue.Val) }
func (me *MoExpr) eqFalse() bool { return (me == moValFalse) || (me.Val == moValFalse.Val) }
func (me *MoExpr) eqNone() bool  { return (me == moValNone) || (me.Val == moValNone.Val) }
func (me *MoExpr) eq(to *MoExpr) bool {
	if me == to {
		return true
	}
	if (me == nil) || (to == nil) || me.Val.primType() != to.Val.primType() {
		return false
	}
	switch it := me.Val.(type) {
	case moValErr:
		other := to.Val.(moValErr)
		return it.Err.eq(other.Err)
	case moValList:
		other := to.Val.(moValList)
		return sl.Eq(it, other, (*MoExpr).eq)
	case moValCall:
		other := to.Val.(moValCall)
		return sl.Eq(it, other, (*MoExpr).eq)
	case *moValFnLam:
		other := to.Val.(*moValFnLam)
		return it.body.eq(other.body) && (it.isMacro == other.isMacro) && sl.Eq(it.params, other.params, (*MoExpr).eq) && it.env.eq(other.env)
	case moValFnPrim:
		other := to.Val.(moValFnPrim)
		return (it == nil) && (other == nil)
	case moValDict:
		other := to.Val.(moValDict)
		if len(it) != len(other) {
			return false
		}
		for i := range it {
			if !sl.HasWhere(other, func(other_pair [2]*MoExpr) bool { return other_pair[0].eq(it[i][0]) && other_pair[1].eq(it[i][1]) }) {
				return false
			}
		}
		return true
	}
	return me.Val == to.Val
}

func (me *Interp) cmp(it *MoExpr, to *MoExpr, diagMsgOpMoniker string) (int, *SrcFileNotice) {
	switch it := it.Val.(type) {
	case moValChar:
		if other, is := to.Val.(moValChar); is {
			return cmp.Compare(it, other), nil
		}
	case moValStr:
		if other, is := to.Val.(moValStr); is {
			return cmp.Compare(it, other), nil
		}
	case moValFloat:
		if other, is := to.Val.(moValFloat); is {
			return cmp.Compare(it, other), nil
		}
	case moValInt:
		if other, is := to.Val.(moValInt); is {
			return cmp.Compare(it, other), nil
		}
	case moValUint:
		if other, is := to.Val.(moValUint); is {
			return cmp.Compare(it, other), nil
		}
	}
	return 0, me.diagSpan(true, false, it, to).newDiagErr(NoticeCodeNotComparable, it, to, diagMsgOpMoniker)
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
	return me.expr(expr.Val, expr.SrcFile, expr.SrcSpan, srcSpanCtx...)
}
func (me *Interp) exprBool(b bool, srcSpanCtx ...*MoExpr) *MoExpr {
	return me.exprFrom(util.If(b, moValTrue, moValFalse), srcSpanCtx...)
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
	case moValType:
		w.WriteString(MoValPrimType(it).Str(false))
	case moValIdent:
		w.WriteString(string(it))
	case moValInt:
		w.WriteString(str.FromI64(int64(it), 10))
	case moValUint:
		w.WriteString(str.FromU64(uint64(it), 10))
	case moValFloat:
		w.WriteString(str.FromFloat(float64(it), -1))
	case moValChar:
		w.WriteString(strconv.QuoteRune(rune(it)))
	case moValStr:
		w.WriteString(str.Q(string(it)))
	case moValErr:
		w.WriteString("(@Err ")
		it.Err.WriteTo(w)
		w.WriteString(")")
	case moValDict:
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
	case moValList:
		w.WriteString("[")
		for i, item := range it {
			if i > 0 {
				w.WriteString(", ")
			}
			item.WriteTo(w)
		}
		w.WriteString("]")
	case moValCall:
		w.WriteString("(")
		for i, item := range it {
			if i > 0 {
				w.WriteString(" ")
			}
			item.WriteTo(w)
		}
		w.WriteString(")")
	case moValFnPrim:
		w.WriteString("<builtin>")
	case *moValFnLam:
		w.WriteString("<lambda>")
	default:
		panic(it)
	}
}
func moValToString(it MoVal) string {
	var buf strings.Builder
	moValWriteTo(it, &buf)
	return buf.String()
}

func (me *Interp) Parse(src string) (*MoExpr, *SrcFileNotice) {
	me.replFauxFile.Src.Ast, me.replFauxFile.Src.Toks, me.replFauxFile.Src.Text = nil, nil, src
	toks, errs := tokenize(me.replFauxFile.FilePath, src)
	if len(errs) > 0 {
		return nil, errs[0]
	}
	me.replFauxFile.Src.Toks = toks
	me.replFauxFile.Src.Ast = me.replFauxFile.parse()
	for _, diag := range me.replFauxFile.allNotices() {
		if diag.Kind == NoticeKindErr {
			return nil, diag
		}
	}
	filter := func(it *AstNode) bool { return (it.Kind != AstNodeKindErr) && (it.Kind != AstNodeKindComment) }
	me.replFauxFile.Src.Ast = sl.Where(me.replFauxFile.Src.Ast, filter)
	me.replFauxFile.Src.Ast.walk(func(node *AstNode) bool {
		node.Nodes = sl.Where(node.Nodes, filter)
		return true
	}, nil)
	if len(me.replFauxFile.Src.Ast) > 1 {
		return nil, me.replFauxFile.Src.Ast.newDiagErr(me.replFauxFile, NoticeCodeAtmoTodo, "odd case: please report, quoting exact input, namely: `"+src+"`")
	} else if (len(me.replFauxFile.Src.Ast) == 0) || (len(me.replFauxFile.Src.Ast[0].Nodes) == 0) {
		return nil, nil
	}

	expr, err := me.replFauxFile.ExprFromAstNode(me.replFauxFile.Src.Ast[0])
	if err != nil {
		return nil, err
	}
	return expr, nil
}

func (me *SrcFile) ExprFromAstNode(topNode *AstNode) (*MoExpr, *SrcFileNotice) {
	util.Assert((topNode.Kind == AstNodeKindGroup) && (len(topNode.Nodes) > 0), nil)
	if len(topNode.Nodes) > 1 {
		topNode.Nodes = AstNodes{topNode.Nodes.toGroupNode(me, topNode, true, false)}
	}
	return me.exprFromAstNode(topNode.Nodes[0])
}

func (me *SrcFile) exprFromAstNode(node *AstNode) (*MoExpr, *SrcFileNotice) {
	var val MoVal
	switch node.Kind {
	case AstNodeKindIdent:
		if node.Toks[0].isSep() {
			return nil, node.newDiagErr(false, NoticeCodeExpectedFoo, "expression instead of `"+node.Src+"` here")
		}
		val = moValIdent(node.Src)
	case AstNodeKindLit:
		switch it := node.Lit.(type) {
		case rune:
			val = moValChar(it)
		case string:
			val = moValStr(it)
		case float64:
			val = moValFloat(it)
		case int64:
			val = moValInt(it)
		case uint64:
			val = moValUint(it)
		default:
			panic(str.Fmt("TODO: lit type %T", it))
		}
	case AstNodeKindGroup:
		switch {
		case node.IsSquareBrackets():
			list := make(moValList, 0, len(node.Nodes))
			for _, node := range node.Nodes {
				expr, err := me.exprFromAstNode(node)
				if err != nil {
					return nil, err
				}
				list = append(list, expr)
			}
			val = list
		case node.IsCurlyBraces():
			dict := make(moValDict, 0, len(node.Nodes))
			for _, node := range node.Nodes {
				util.Assert(len(node.Nodes) == 2, len(node.Nodes))
				expr_key, err := me.exprFromAstNode(node.Nodes[0])
				if err != nil {
					return nil, err
				}
				expr_val, err := me.exprFromAstNode(node.Nodes[1])
				if err != nil {
					return nil, err
				}
				if dict.Has(expr_key) {
					return nil, expr_key.SrcSpan.newDiagErr(NoticeCodeDictDuplKey, expr_key)
				}
				dict.Set(expr_key, expr_val)
			}
			val = dict
		default:
			if len(node.Nodes) == 1 {
				return me.exprFromAstNode(node.Nodes[0])
			} else if len(node.Nodes) == 0 {
				return nil, node.newDiagErr(false, NoticeCodeExpectedFoo, "expression inside these empty parens")
			}

			call_form := make(moValCall, 0, len(node.Nodes))
			for _, node := range node.Nodes {
				expr, err := me.exprFromAstNode(node)
				if err != nil {
					return nil, err
				}
				call_form = append(call_form, expr)
			}
			val = call_form
		}
	}
	return &MoExpr{SrcFile: me, SrcSpan: util.Ptr(node.Toks.Span()), Val: val}, nil
}

func (me *MoExpr) walk(onBefore func(it *MoExpr) bool, onAfter func(it *MoExpr)) {
	if onBefore != nil && !onBefore(me) {
		return
	}
	switch it := me.Val.(type) {
	case moValCall:
		for _, item := range it {
			item.walk(onBefore, onAfter)
		}
	case moValDict:
		for _, pair := range it {
			pair[0].walk(onBefore, onAfter)
			pair[1].walk(onBefore, onAfter)
		}
	case moValErr:
		it.Err.walk(onBefore, onAfter)
	case *moValFnLam:
		for _, item := range it.params {
			item.walk(onBefore, onAfter)
		}
		it.body.walk(onBefore, onAfter)
	case moValList:
		for _, item := range it {
			item.walk(onBefore, onAfter)
		}
	}
	if onAfter != nil {
		onAfter(me)
	}
}

type MoExprs []*MoExpr

func (me MoExprs) walk(onBefore func(it *MoExpr) bool, onAfter func(it *MoExpr)) {
	for _, expr := range me {
		expr.walk(onBefore, onAfter)
	}
}
