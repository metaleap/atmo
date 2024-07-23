package session

import (
	"atmo/util/sl"
	"atmo/util/str"
	"errors"
	"os"
)

type Interp interface {
	Eval(*AtExpr) (*AtExpr, error)
	Parse(src string) (*AtExpr, error)
}

type interp struct {
	SrcFile    *SrcFile
	env        *AtEnv
	stackTrace []string
}

func (me *interp) Eval(*AtExpr) (*AtExpr, error) {
	me.stackTrace = nil
	return nil, nil
}

func (*interp) evalAndApply() {
}

func (*interp) evalExpr() {
}

func (me *interp) Parse(src string) (*AtExpr, error) {
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
		return nil, errors.New("one at a time, please")
	} else if len(me.SrcFile.Src.Ast) == 0 {
		return nil, nil
	}
	return me.SrcFile.toExpr(me.SrcFile.Src.Ast[0])
}

func (me *SrcFile) toExpr(node *AstNode) (*AtExpr, error) {
	var val AtVal
	switch node.Kind {
	case AstNodeKindIdent:
		val = atValIdent(node.Src)
	case AstNodeKindLit:
		switch it := node.Lit.(type) {
		case rune:
			val = atValChar(it)
		case string:
			val = atValStr(it)
		case float64:
			val = atValFloat(it)
		case int64:
			val = atValInt(it)
		case uint64:
			val = atValUint(it)
		default:
			panic(str.Fmt("TODO: lit type %T", it))
		}
	case AstNodeKindGroup:
		switch {
		case node.isBraces():
			rec := make(atValRec, len(node.Nodes)/4)
			val = rec
		case node.isBrackets():
			rec := make(atValArr, 0, len(node.Nodes)/2)
			val = rec
		default:
			if len(node.Nodes) == 1 {
				return me.toExpr(node.Nodes[0])
			} else if len(node.Nodes) == 0 {
				return nil, node.newDiagErr(false, NoticeCodeExpectedFooHere, "expression", "inside the parens")
			}

			rec := make(atValCall, 0, len(node.Nodes))
			for _, node := range node.Nodes {
				expr, err := me.toExpr(node)
				if err != nil {
					return nil, err
				}
				rec = append(rec, expr)
			}
			val = rec
		}
	}
	if val != nil {
		ret := &AtExpr{SrcNode: node, Val: val}
		os.Stdout.WriteString(">>>")
		ret.WriteTo(os.Stdout)
		os.Stdout.WriteString("<<<\n")
		return ret, nil
	}
	return nil, nil
}
