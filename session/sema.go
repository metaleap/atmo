package session

import (
	"cmp"
	"time"

	"atmo/util/sl"
	"atmo/util/str"
)

func (me *SrcPack) refreshSema() (encounteredDiagsRelevantChanges bool) {
	return
	defer func(timeStarted time.Time) {
		OnLogMsg(true, "SEMA %s for %s", str.DurationMs(time.Since(timeStarted).Nanoseconds()), me.DirPath)
	}(time.Now())

	var top_level MoExprs
	for _, src_file := range me.Files {
		if !src_file.isReplish() {
			had_errs := (len(src_file.notices.Sema) > 0)
			src_file.notices.Sema = nil
			for _, top_node := range src_file.Src.Ast {
				expr, err := src_file.ExprFromAstNode(top_node)
				if err != nil {
					src_file.notices.Sema = append(src_file.notices.Sema, err)
				} else if expr != nil {
					top_level = append(top_level, expr)
				}
			}
			encounteredDiagsRelevantChanges = encounteredDiagsRelevantChanges || (len(src_file.notices.Sema) > 0) || had_errs
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
