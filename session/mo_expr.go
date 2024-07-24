package session

import (
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
	MoPrimTypeRec
	MoPrimTypeArr
	MoPrimTypeCall
	MoPrimTypeFunc
)

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
type moValRec map[*MoExpr]*MoExpr
type moValArr []*MoExpr
type moValCall []*MoExpr
type moValFnPrim moFnEager
type moValFnLam struct {
	params  []*MoExpr // all are guaranteed to be ident before construction
	body    *MoExpr
	env     *MoEnv
	isMacro bool
}

func (moValType) primType() MoValPrimType   { return MoPrimTypeType }
func (moValIdent) primType() MoValPrimType  { return MoPrimTypeIdent }
func (moValInt) primType() MoValPrimType    { return MoPrimTypeInt }
func (moValUint) primType() MoValPrimType   { return MoPrimTypeUint }
func (moValFloat) primType() MoValPrimType  { return MoPrimTypeFloat }
func (moValChar) primType() MoValPrimType   { return MoPrimTypeChar }
func (moValStr) primType() MoValPrimType    { return MoPrimTypeStr }
func (moValErr) primType() MoValPrimType    { return MoPrimTypeErr }
func (moValRec) primType() MoValPrimType    { return MoPrimTypeRec }
func (moValArr) primType() MoValPrimType    { return MoPrimTypeArr }
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
func (me moValRec) String() string          { return moValToString(me) }
func (me moValArr) String() string          { return moValToString(me) }
func (me moValCall) String() string         { return moValToString(me) }
func (me moValFnPrim) String() string       { return moValToString(me) }
func (me *moValFnLam) String() string       { return moValToString(me) }

type MoExpr struct {
	Val     MoVal
	SrcSpan *SrcFileSpan `json:"-"`
}

func (me *MoExpr) Callee() *MoExpr {
	if call, is := me.Val.(moValCall); is {
		return call[0]
	}
	return nil
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
	case moValRec:
		w.WriteString("{")
		var n int
		for k, v := range it {
			if n > 0 {
				w.WriteString(", ")
			}
			if ident, _ := k.Val.(moValIdent); (ident != "") && (ident[0] == '@') {
				w.WriteString(str.Q(string(ident)))
			} else {
				k.WriteTo(w)
			}
			w.WriteString(": ")
			v.WriteTo(w)
			n++
		}
		w.WriteString("}")
	case moValArr:
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
		w.WriteString("<primFunc>")
	case *moValFnLam:
		w.WriteString("<lambdaFunc>")
	default:
		panic(it)
	}
}
func moValToString(it MoVal) string {
	var buf strings.Builder
	moValWriteTo(it, &buf)
	return buf.String()
}

func (me MoValPrimType) String() string { return me.Str(false) }
func (me MoValPrimType) Str(forDiag bool) string {
	switch me {
	case MoPrimTypeType:
		return util.If(forDiag, "type", "@Type")
	case MoPrimTypeIdent:
		return util.If(forDiag, "identifier", "@Ident")
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
	case MoPrimTypeRec:
		return util.If(forDiag, "record", "@Rec")
	case MoPrimTypeArr:
		return util.If(forDiag, "list", "@Arr")
	case MoPrimTypeCall:
		return util.If(forDiag, "call", "@Call")
	case MoPrimTypeFunc:
		return util.If(forDiag, "function", "@Func")
	}
	panic(me)
}

func (me *Interp) Parse(src string) (*MoExpr, *SrcFileNotice) {
	me.SrcFile.Src.Ast, me.SrcFile.Src.Toks, me.SrcFile.Src.Text = nil, nil, src
	toks, errs := tokenize(me.SrcFile.FilePath, src)
	if len(errs) > 0 {
		return nil, errs[0]
	}
	me.SrcFile.Src.Toks = toks
	me.SrcFile.Src.Ast = me.SrcFile.parse()
	for _, diag := range me.SrcFile.allNotices() {
		if diag.Kind == NoticeKindErr {
			return nil, diag
		}
	}
	filter := func(it *AstNode) bool { return (it.Kind != AstNodeKindErr) && (it.Kind != AstNodeKindComment) }
	me.SrcFile.Src.Ast = sl.Where(me.SrcFile.Src.Ast, filter)
	me.SrcFile.Src.Ast.walk(func(node *AstNode) bool {
		node.Nodes = sl.Where(node.Nodes, filter)
		return true
	}, nil)
	if len(me.SrcFile.Src.Ast) > 1 {
		return nil, me.SrcFile.Src.Ast.newDiagErr(me.SrcFile, NoticeCodeAtmoTodo, "odd case, please report exact input, namely: `"+src+"`")
	} else if (len(me.SrcFile.Src.Ast) == 0) || (len(me.SrcFile.Src.Ast[0].Nodes) == 0) {
		return nil, nil
	}

	return me.SrcFile.ExprFromAstNode(me.SrcFile.Src.Ast[0])
}

func (me *SrcFile) ExprFromAstNode(topNode *AstNode) (*MoExpr, *SrcFileNotice) {
	util.Assert((topNode.Kind == AstNodeKindGroup) && (len(topNode.Nodes) > 0), nil)
	if len(topNode.Nodes) > 1 {
		topNode.Nodes = []*AstNode{topNode.Nodes.toGroupNode(me, topNode, true, false)}
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
			arr := make(moValArr, 0, len(node.Nodes))
			for _, node := range node.Nodes {
				expr, err := me.exprFromAstNode(node)
				if err != nil {
					return nil, err
				}
				arr = append(arr, expr)
			}
			val = arr
		case node.IsCurlyBraces():
			rec := make(moValRec, len(node.Nodes))
			for _, node := range node.Nodes {
				expr_key, err := me.exprFromAstNode(node.Nodes[0])
				if err != nil {
					return nil, err
				}
				expr_val, err := me.exprFromAstNode(node.Nodes[1])
				if err != nil {
					return nil, err
				}
				rec[expr_key] = expr_val
			}
			val = rec
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
	ret := &MoExpr{SrcSpan: util.Ptr(node.Toks.Span()), Val: val}
	// if val != nil {
	// 	os.Stdout.WriteString(">>>")
	// 	ret.WriteTo(os.Stdout)
	// 	os.Stdout.WriteString("<<<\n")
	// }
	return ret, nil
}
