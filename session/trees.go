package session

import (
	"cmp"
	"time"

	"atmo/util"
	"atmo/util/sl"
	"atmo/util/str"
)

var (
	DoSrcPackEvals bool
	DoSrcPackSems  bool
)

func (me *SrcPack) treesRefresh() (encounteredDiagsRelevantChanges bool) {
	if me.treesRefreshCanSkip() {
		return
	}
	defer func(timeStarted time.Time) {
		OnLogMsg(true, "TR %s for %s", str.DurationMs(time.Since(timeStarted).Nanoseconds()), me.DirPath)
	}(time.Now())

	var top_level MoExprs
	var any_pre_errs bool
	for _, src_file := range me.Files {
		if !src_file.IsInterpFauxFile() {
			had_errs := (len(src_file.notices.AstToMo) > 0)
			src_file.notices.AstToMo = nil
			for _, top_node := range src_file.Src.Ast {
				expr, err := src_file.MoExprFromAstNode(top_node)
				if err != nil {
					src_file.notices.AstToMo = append(src_file.notices.AstToMo, err)
				} else if expr != nil {
					top_level = append(top_level, expr)
				}
			}
			any_pre_errs = any_pre_errs || (len(src_file.notices.AstToMo) > 0)
			encounteredDiagsRelevantChanges = encounteredDiagsRelevantChanges || (len(src_file.notices.AstToMo) > 0) || had_errs
		}
	}
	me.Trees.MoOrig = top_level

	old_had_errs := me.Trees.MoEvaled.AnyErrs()
	if any_pre_errs && !old_had_errs { // bug out & leave the old `.Trees.MoEvaled` intact in this case, for editor clients
		return
	}

	if DoSrcPackEvals {
		if me.Interp == nil {
			_ = newInterp(me, nil)
			util.Assert(me.Interp != nil, nil)
		}

		me.Trees.MoEvaled = nil
		me.Interp.envReset()
		// first, handle top-level `@set` exprs: we add them to env but un-evaled, just so ident evals (env lookups) will find them (and eval them)
		for _, top_expr := range me.Trees.MoOrig {
			if ident := me.Interp.isSetCall(top_expr); ident != "" {
				dup := *top_expr
				dup.PreEvalTopLevelPreEnvUnevaledYet = true
				me.Interp.Env.set(ident, &dup)
			}
		}
		// now, we eval
		for _, top_expr := range me.Trees.MoOrig {
			dup := *top_expr
			if evaled := me.Interp.ExprEval(&dup); evaled != nil {
				me.Trees.MoEvaled = append(me.Trees.MoEvaled, evaled)
				encounteredDiagsRelevantChanges = encounteredDiagsRelevantChanges || evaled.HasErrs()
			}
		}
		encounteredDiagsRelevantChanges = encounteredDiagsRelevantChanges || old_had_errs || me.Trees.MoEvaled.AnyErrs()
	}

	if DoSrcPackSems {
		me.semRefresh()
	}

	return
}

func (me *SrcPack) treesRefreshCanSkip() bool {
	cur_paths := me.srcFilePaths()
	can_skip := (len(cur_paths) == len(me.Trees.last.files))
	if can_skip {
		cur_paths := cur_paths
		for _, path := range cur_paths {
			if _, ok := me.Trees.last.files[path]; !ok {
				can_skip = false
				break
			}
		}
		for path := range me.Trees.last.files {
			if can_skip && !sl.Has(cur_paths, path) {
				can_skip = false
				break
			}
		}
		for _, src_file := range me.Files {
			if can_skip && (!src_file.IsInterpFauxFile()) && (me.Trees.last.files[src_file.FilePath] != util.ContentHash(src_file.Src.Text)) {
				can_skip = false
				break
			}
		}
	}
	if !can_skip {
		for _, src_file := range me.Files {
			if (!src_file.IsInterpFauxFile()) && src_file.HasMoPrecludingErrs() {
				can_skip = true
				break
			}
		}
	}
	if !can_skip {
		me.Trees.last.files = map[string]string{}
		for _, src_file := range me.Files {
			if !src_file.IsInterpFauxFile() {
				me.Trees.last.files[src_file.FilePath] = util.ContentHash(src_file.Src.Text)
			}
		}
	}
	return can_skip
}

func (me MoExprs) Sorted() MoExprs {
	return sl.SortedPer(me, func(expr1 *MoExpr, expr2 *MoExpr) int {
		var node1, node2 *AstNode
		if expr1.SrcFile.FilePath == expr2.SrcFile.FilePath {
			node1, node2 = expr1.srcNode(), expr2.srcNode()
		}
		if (node1 == nil) || (node2 == nil) {
			return cmp.Compare(expr1.SrcFile.FilePath, expr2.SrcFile.FilePath)
		}
		if node1.Toks[0].Pos.Line != node2.Toks[0].Pos.Line {
			return cmp.Compare(node1.Toks[0].Pos.Line, node2.Toks[0].Pos.Line)
		}
		// we shouldn't really ever get to here, since we only sort top-level exprs and there are no two of them on the same line in the same file
		return cmp.Compare(node1.Toks[0].Pos.Char, node2.Toks[0].Pos.Char)
	})
}