package session

import (
	"io"
	"strconv"
	"strings"

	"atmo/util"
	"atmo/util/sl"
	"atmo/util/str"
)

type moFnEager = func(ctx *Interp, args ...*MoExpr) (*MoExpr, *SrcFileNotice)
type moFnLazy = func(ctx *Interp, env *MoEnv, args ...*MoExpr) (*MoEnv, *MoExpr, *SrcFileNotice)

type MoValType int

const (
	MoValTypeType MoValType = iota
	MoValTypeIdent
	MoValTypeInt
	MoValTypeUint
	MoValTypeFloat
	MoValTypeChar
	MoValTypeStr
	MoValTypeErr
	MoValTypeRec
	MoValTypeArr
	MoValTypeCall
	MoValTypeFunc
)

type MoVal interface {
	valType() MoValType
}

type moValType MoValType
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

func (MoValType) valType() MoValType   { return MoValTypeType }
func (moValIdent) valType() MoValType  { return MoValTypeIdent }
func (moValInt) valType() MoValType    { return MoValTypeInt }
func (moValUint) valType() MoValType   { return MoValTypeUint }
func (moValFloat) valType() MoValType  { return MoValTypeFloat }
func (moValChar) valType() MoValType   { return MoValTypeChar }
func (moValStr) valType() MoValType    { return MoValTypeStr }
func (moValErr) valType() MoValType    { return MoValTypeErr }
func (moValRec) valType() MoValType    { return MoValTypeRec }
func (moValArr) valType() MoValType    { return MoValTypeArr }
func (moValCall) valType() MoValType   { return MoValTypeCall }
func (moValFnPrim) valType() MoValType { return MoValTypeFunc }
func (*moValFnLam) valType() MoValType { return MoValTypeFunc }

type MoExpr struct {
	SrcNode *AstNode `json:"-"`
	Val     MoVal
}

func (me *MoExpr) Callee() *MoExpr {
	if call, is := me.Val.(moValCall); is {
		return call[0]
	}
	return nil
}

func (me *MoExpr) String() string {
	var buf strings.Builder
	me.WriteTo(&buf)
	return buf.String()
}

func (me *MoExpr) WriteTo(w io.StringWriter) {
	switch it := me.Val.(type) {
	case MoValType:
		w.WriteString(MoValType(it).String())
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
			k.WriteTo(w)
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
	case moValFnPrim, *moValFnLam:
		w.WriteString(me.SrcNode.Src)
	default:
		panic(it)
	}
}

func (me MoValType) String() string {
	switch me {
	case MoValTypeType:
		return "@Type"
	case MoValTypeIdent:
		return "@Ident"
	case MoValTypeInt:
		return "@Int"
	case MoValTypeUint:
		return "@Uint"
	case MoValTypeFloat:
		return "@Float"
	case MoValTypeChar:
		return "@Char"
	case MoValTypeStr:
		return "@Str"
	case MoValTypeErr:
		return "@Err"
	case MoValTypeRec:
		return "@Rec"
	case MoValTypeArr:
		return "@Arr"
	case MoValTypeCall:
		return "@Call"
	case MoValTypeFunc:
		return "@Func"
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
	ret := &MoExpr{SrcNode: node, Val: val}
	// if val != nil {
	// 	os.Stdout.WriteString(">>>")
	// 	ret.WriteTo(os.Stdout)
	// 	os.Stdout.WriteString("<<<\n")
	// }
	return ret, nil
}
