package session

import (
	"cmp"
	"time"

	"atmo/util/sl"
)

func (me *SrcPack) refreshSema() (encounteredDiagsRelevantChanges bool) {
	defer func(timeStarted time.Time) { OnDbgMsg(true, "SEMA %s for %s", time.Since(timeStarted), me.DirPath) }(time.Now())

	var top_level MoExprs
	for _, src_file := range me.Files {
		if !src_file.isReplish() {
			src_file.notices.Sema = nil
			for _, top_node := range src_file.Src.Ast {
				if (top_node.Kind == AstNodeKindComment) || (top_node.Kind == AstNodeKindErr) {
					continue
				}
				expr, err := src_file.ExprFromAstNode(top_node)
				if err != nil {
					src_file.notices.Sema = append(src_file.notices.Sema, err)
				} else {
					top_level = append(top_level, expr)
				}
			}
			encounteredDiagsRelevantChanges = encounteredDiagsRelevantChanges || (len(src_file.notices.Sema) > 0)
		}
	}

	me.Sema.Top = top_level.sorted()
	return
}

func (me MoExprs) sorted() MoExprs {
	return sl.SortedPer(me, func(expr1 *MoExpr, expr2 *MoExpr) int {
		node1, node2 := expr1.srcNode(), expr2.srcNode()
		if (node1 == nil) || (node2 == nil) || (expr1.SrcFile.FilePath != expr2.SrcFile.FilePath) {
			return cmp.Compare(expr1.SrcFile.FilePath, expr2.SrcFile.FilePath)
		}
		if node1.Toks[0].Pos.Line != node2.Toks[0].Pos.Line {
			return cmp.Compare(node1.Toks[0].Pos.Line, node2.Toks[0].Pos.Line)
		}
		return cmp.Compare(node1.Toks[0].Pos.Char, node2.Toks[0].Pos.Char)
	})
}
