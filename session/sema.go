package session

import (
	"cmp"
	"time"

	"atmo/util"
	"atmo/util/sl"
	"atmo/util/str"
)

func (me *SrcPack) semaRefresh() (encounteredDiagsRelevantChanges bool) {
	if me.semaRefreshCanSkip() {
		return
	}
	defer func(timeStarted time.Time) {
		OnLogMsg(true, "SEMA %s for %s", str.DurationMs(time.Since(timeStarted).Nanoseconds()), me.DirPath)
	}(time.Now())

	var top_level MoExprs
	var any_pre_errs bool
	for _, src_file := range me.Files {
		if !src_file.IsInterpFauxFile() {
			had_errs := (len(src_file.notices.PreSema) > 0)
			src_file.notices.PreSema = nil
			for _, top_node := range src_file.Src.Ast {
				expr, err := src_file.MoExprFromAstNode(top_node)
				if err != nil {
					src_file.notices.PreSema = append(src_file.notices.PreSema, err)
				} else if expr != nil {
					top_level = append(top_level, expr)
				}
			}
			any_pre_errs = any_pre_errs || (len(src_file.notices.PreSema) > 0)
			encounteredDiagsRelevantChanges = encounteredDiagsRelevantChanges || (len(src_file.notices.PreSema) > 0) || had_errs
		}
	}
	me.Sema.Pre = top_level

	old_had_errs := me.Sema.Post.AnyErrs()
	if any_pre_errs && !old_had_errs { // bug out & leave the old `.Sema.Post` intact in this case, for editor clients
		return
	}

	if me.Interp == nil {
		_ = newInterp(me, nil)
		util.Assert(me.Interp != nil, nil)
	}

	me.Sema.Post = nil
	me.Interp.envReset()
	// first, handle top-level `@set` exprs: we add them to env but un-evaled, just so sema evals will find them
	for _, top_expr := range me.Sema.Pre {
		if ident := me.Interp.isSetCall(top_expr); ident != "" {
			dup := *top_expr
			dup.Sema = &SemaExpr{topLevelPreEnvUnevaled: true}
			me.Interp.Env.set(ident, &dup)
		}
	}
	// now, we sema
	for _, top_expr := range me.Sema.Pre {
		dup := *top_expr
		if evaled := me.Interp.ExprEval(&dup, true); evaled != nil {
			me.Sema.Post = append(me.Sema.Post, evaled)
			encounteredDiagsRelevantChanges = encounteredDiagsRelevantChanges || evaled.HasErrs() || evaled.EqNever()
		}
	}
	encounteredDiagsRelevantChanges = encounteredDiagsRelevantChanges || old_had_errs || me.Sema.Post.AnyErrs()
	return
}

func (me *SrcPack) semaRefreshCanSkip() bool {
	cur_paths := me.srcFilePaths()
	can_skip := (len(cur_paths) == len(me.Sema.last.files))
	if can_skip {
		cur_paths := cur_paths
		for _, path := range cur_paths {
			if _, ok := me.Sema.last.files[path]; !ok {
				can_skip = false
				break
			}
		}
		for path := range me.Sema.last.files {
			if can_skip && !sl.Has(cur_paths, path) {
				can_skip = false
				break
			}
		}
		for _, src_file := range me.Files {
			if can_skip && (!src_file.IsInterpFauxFile()) && (me.Sema.last.files[src_file.FilePath] != util.ContentHash(src_file.Src.Text)) {
				can_skip = false
				break
			}
		}
	}
	if !can_skip {
		for _, src_file := range me.Files {
			if (!src_file.IsInterpFauxFile()) && src_file.HasSemaPrecludingErrs() {
				can_skip = true
				break
			}
		}
	}
	if !can_skip {
		me.Sema.last.files = map[string]string{}
		for _, src_file := range me.Files {
			if !src_file.IsInterpFauxFile() {
				me.Sema.last.files[src_file.FilePath] = util.ContentHash(src_file.Src.Text)
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

type SemaExpr struct {
	topLevelPreEnvUnevaled bool
}
