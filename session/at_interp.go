package session

import (
	"atmo/util"
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
	} else if (len(me.SrcFile.Src.Ast) == 0) || (len(me.SrcFile.Src.Ast[0].Nodes) == 0) {
		return nil, nil
	}

	util.Assert(me.SrcFile.Src.Ast[0].Kind == AstNodeKindGroup, nil)
	if len(me.SrcFile.Src.Ast[0].Nodes) > 1 {
		me.SrcFile.Src.Ast[0].Nodes = []*AstNode{me.SrcFile.Src.Ast[0].Nodes.group(me.SrcFile, me.SrcFile.Src.Ast[0], true, false)}
	}
	return me.SrcFile.toExpr(me.SrcFile.Src.Ast[0].Nodes[0])
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
		case node.isBrackets():
			items := node.Nodes.splitByIdentWithGrouping(me, node, ",")
			arr := make(atValArr, 0, len(items))
			err_node := node
			for _, expr_node := range items {
				if expr_node == nil {
					return nil, err_node.newDiagErr(err_node != node, NoticeCodeExpectedFooHere, "expression", "before the superfluous comma")
				}
				err_node = expr_node
				expr, err := me.toExpr(expr_node)
				if err != nil {
					return nil, err
				}
				arr = append(arr, expr)
			}
			val = arr
		case node.isBraces():
			items := node.Nodes.splitByIdentWithGrouping(me, node, ",")
			rec := make(atValRec, len(items))
			err_node := node
			for _, expr_node := range items {
				if expr_node == nil {
					return nil, err_node.newDiagErr(false, NoticeCodeExpectedFooHere, "expression", "in between the 2 consecutive commas")
				}
				err_node = expr_node
				pair := expr_node.Nodes.splitByIdentWithGrouping(me, expr_node, ":")
				if len(pair) != 2 || pair[0] == nil || pair[1] == nil {
					return nil, err_node.newDiagErr(true, NoticeCodeExpectedFooHere, "expression pair separated by `:`", "")
				}
				expr_key, err := me.toExpr(pair[0])
				if err != nil {
					return nil, err
				}
				expr_val, err := me.toExpr(pair[1])
				if err != nil {
					return nil, err
				}
				rec[expr_key] = expr_val
			}
			val = rec
		default:
			if len(node.Nodes) == 1 {
				return me.toExpr(node.Nodes[0])
			} else if len(node.Nodes) == 0 {
				return nil, node.newDiagErr(false, NoticeCodeExpectedFooHere, "expression", "inside the parens")
			}

			call_form := make(atValCall, 0, len(node.Nodes))
			for _, node := range node.Nodes {
				expr, err := me.toExpr(node)
				if err != nil {
					return nil, err
				}
				call_form = append(call_form, expr)
			}
			val = call_form
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
