package session

import (
	"atmo/util"
	"atmo/util/sl"
	"atmo/util/str"
)

type Evaler interface {
	eval(ctx *Interp, expr *MoExpr) (*MoExpr, *SrcFileNotice)
}

type Interp struct {
	SrcFile        *SrcFile
	Env            *MoEnv
	evaler         Evaler
	StackTraces    bool
	LastStackTrace []*MoExpr
}

type DefaultEvaler struct {
	ctx *Interp
}

func newInterp(srcFile *SrcFile, evaler Evaler) *Interp {
	if evaler == nil {
		evaler = &DefaultEvaler{}
	}
	interp := &Interp{Env: newMoEnv(nil, nil, nil), SrcFile: srcFile, evaler: evaler}
	return interp
}

func (me *Interp) Eval(expr *MoExpr) (*MoExpr, *SrcFileNotice) {
	me.LastStackTrace = me.LastStackTrace[:0] // keeps old capacity allocated
	return me.evaler.eval(me, expr)
}

func (me *DefaultEvaler) eval(ctx *Interp, expr *MoExpr) (*MoExpr, *SrcFileNotice) {
	me.ctx = ctx
	return me.evalAndApply(ctx.Env, expr)
}

func (me *DefaultEvaler) evalAndApply(env *MoEnv, expr *MoExpr) (*MoExpr, *SrcFileNotice) {
	var err *SrcFileNotice
	for (err == nil) && (env != nil) {
		if _, is_call := expr.Val.(moValCall); !is_call {
			expr, err = me.evalExpr(env, expr)
			env = nil
		} else if expr, err = me.macroExpand(env, expr); err != nil {
			return nil, err
		} else if call, _ := expr.Val.(moValCall); len(call) > 0 {
			callee, call_args := call[0], ([]*MoExpr)(call[1:])
			var special_form moFnLazy
			if ident, _ := callee.Val.(moValIdent); ident != "" {
				special_form = moStdLazy[ident]
			}
			if special_form != nil {
				if env, expr, err = special_form(me.ctx, env, call_args...); err != nil {
					return nil, err
				}
			} else {
				if expr, err = me.evalExpr(env, expr); err != nil {
					return nil, err
				}
				if me.ctx.StackTraces {
					me.ctx.LastStackTrace = append(me.ctx.LastStackTrace, expr)
				}
				call = expr.Val.(moValCall)
				callee, call_args = call[0], ([]*MoExpr)(call[1:])
				switch fn := callee.Val.(type) {
				default:
					return nil, callee.SrcNode.newDiagErr(false, NoticeCodeUncallable, callee.String())
				case moValFn:
					if expr, err = fn(call_args...); err != nil {
						return nil, err
					}
					env = nil
				case *moValFunc:
					expr = fn.body
					env, err = fn.envWith(call_args, expr)
					if err != nil {
						return nil, err
					}
				}
			}
		}
	}
	return expr, err
}

func (me *DefaultEvaler) evalExpr(env *MoEnv, expr *MoExpr) (*MoExpr, *SrcFileNotice) {
	switch val := expr.Val.(type) {
	case moValIdent:
		found := env.lookup(val)
		if found == nil {
			return nil, expr.SrcNode.newDiagErr(false, NoticeCodeUndefined, val)
		}
	case moValArr:
		arr := make(moValArr, len(val))
		for i, item := range val {
			it, err := me.evalAndApply(env, item)
			if err != nil {
				return nil, err
			}
			arr[i] = it
		}
		return &MoExpr{Val: arr, SrcNode: expr.SrcNode}, nil
	case moValRec:
		rec := make(moValRec, len(val))
		for k, v := range val {
			key, err := me.evalAndApply(env, k)
			if err != nil {
				return nil, err
			}
			val, err := me.evalAndApply(env, v)
			if err != nil {
				return nil, err
			}
			rec[key] = val
		}
		return &MoExpr{Val: rec, SrcNode: expr.SrcNode}, nil
	case moValCall:
		call := make(moValCall, len(val))
		for i, item := range val {
			it, err := me.evalAndApply(env, item)
			if err != nil {
				return nil, err
			}
			call[i] = it
		}
		return &MoExpr{Val: call, SrcNode: expr.SrcNode}, nil
	}
	return expr, nil
}

func (me *DefaultEvaler) macroExpand(_ *MoEnv, expr *MoExpr) (*MoExpr, *SrcFileNotice) {
	return expr, nil
}

func checkCount(wantAtLeast int, wantAtMost int, have []*MoExpr, diagCtx *MoExpr) *SrcFileNotice {
	if wantAtLeast < 0 {
		return nil
	} else if (wantAtLeast == wantAtMost) && (wantAtLeast != len(have)) {
		return diagCtx.SrcNode.newDiagErr(false, NoticeCodeExpectedFoo, str.Fmt("%d arg(s), not %d", wantAtLeast, len(have)))
	} else if len(have) < wantAtLeast {
		return diagCtx.SrcNode.newDiagErr(false, NoticeCodeExpectedFoo, str.Fmt("at least %d arg(s), not %d", wantAtLeast, len(have)))
	} else if (wantAtMost > wantAtLeast) && (len(have) > wantAtMost) {
		return diagCtx.SrcNode.newDiagErr(false, NoticeCodeExpectedFoo, str.Fmt("%d to %d arg(s), not %d", wantAtLeast, wantAtMost, len(have)))
	}
	return nil
}

func checkIs(want MoValType, have *MoExpr) *SrcFileNotice {
	if have_type := have.Val.valType(); have_type != want {
		return have.SrcNode.newDiagErr(false, NoticeCodeExpectedFoo, str.Fmt("`%s`, not `%s`", want, have_type))
	}
	return nil
}

func checkAre(want MoValType, have ...*MoExpr) *SrcFileNotice {
	for _, expr := range have {
		if err := checkIs(want, expr); err != nil {
			return err
		}
	}
	return nil
}

func checkAreBoth(want MoValType, have []*MoExpr, exactArgsCount bool, diagCtx *MoExpr) (err *SrcFileNotice) {
	max_args_count := -1
	if exactArgsCount {
		max_args_count = 2
	}
	if err = checkCount(2, max_args_count, have, diagCtx); err == nil {
		if err = checkIs(want, have[0]); err == nil {
			err = checkIs(want, have[1])
		}
	}
	return
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
