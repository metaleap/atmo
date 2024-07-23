package session

import (
	"errors"

	"atmo/util"
	"atmo/util/sl"
	"atmo/util/str"
)

type Evaler interface {
	eval(ctx *Interp, expr *MoExpr) (*MoExpr, error)
}

type Interp struct {
	SrcFile *SrcFile
	Env     *MoEnv
	evaler  Evaler
}

type DefaultEvaler struct {
	LastStackTrace []string
}

func newInterp(srcFile *SrcFile, evaler Evaler) *Interp {
	if evaler == nil {
		evaler = &DefaultEvaler{}
	}
	interp := &Interp{Env: newMoEnv(nil, nil, nil), SrcFile: srcFile, evaler: evaler}
	return interp
}

func (me *Interp) Eval(expr *MoExpr) (*MoExpr, error) {
	return me.evaler.eval(me, expr)
}

func (me *DefaultEvaler) eval(ctx *Interp, expr *MoExpr) (*MoExpr, error) {
	me.LastStackTrace = nil
	return me.evalAndApply(ctx.Env, expr)
}

func (me *DefaultEvaler) evalAndApply(env *MoEnv, expr *MoExpr) (*MoExpr, error) {
	return me.evalExpr(env, expr)
}

func (me *DefaultEvaler) evalExpr(env *MoEnv, expr *MoExpr) (*MoExpr, error) {
	switch val := expr.Val.(type) {
	case moValIdent:
		found := env.lookup(val)
		if found == nil {
			return nil, expr.SrcNode.newDiagErr(false, NoticeCodeUndefined, val)
		}
	}
	return nil, nil
}

func (me *Interp) Parse(src string) (*MoExpr, error) {
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

	return me.SrcFile.ExprFromAstNode(me.SrcFile.Src.Ast[0])
}

func (me *SrcFile) ExprFromAstNode(topNode *AstNode) (*MoExpr, error) {
	util.Assert((topNode.Kind == AstNodeKindGroup) && (len(topNode.Nodes) > 0), nil)
	if len(topNode.Nodes) > 1 {
		topNode.Nodes = []*AstNode{topNode.Nodes.toGroupNode(me, topNode, true, false)}
	}
	return me.exprFromAstNode(topNode.Nodes[0])
}

func (me *SrcFile) exprFromAstNode(node *AstNode) (*MoExpr, error) {
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
