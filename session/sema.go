package session

import (
	"cmp"
	"time"

	"atmo/util"
	"atmo/util/sl"
	"atmo/util/str"
)

func (me *SrcPack) refreshSema() (encounteredDiagsRelevantChanges bool) {
	return
	cur_paths := me.srcFilePaths()
	can_skip := (len(cur_paths) == len(me.Sema.last.files))
	defer func(timeStarted time.Time) {
		OnLogMsg(!can_skip, "SEMA %s for %s", str.DurationMs(time.Since(timeStarted).Nanoseconds()), me.DirPath)
	}(time.Now())

	if can_skip {
		cur_paths := me.srcFilePaths()
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
			if can_skip && (!src_file.isReplish()) && (me.Sema.last.files[src_file.FilePath] != util.ContentHash(src_file.Src.Text)) {
				can_skip = false
				break
			}
		}
	}
	if can_skip {
		return
	} else {
		me.Sema.last.files = map[string]string{}
		for _, src_file := range me.Files {
			if !src_file.isReplish() {
				me.Sema.last.files[src_file.FilePath] = util.ContentHash(src_file.Src.Text)
			}
		}
	}

	var top_level MoExprs
	var any_pre_errs bool
	for _, src_file := range me.Files {
		if !src_file.isReplish() {
			has_brace_errs, had_errs := src_file.Src.Ast.hasBraceErrors(), (len(src_file.notices.PreSema) > 0)
			src_file.notices.PreSema = nil
			if !has_brace_errs {
				for _, top_node := range src_file.Src.Ast {
					expr, err := src_file.ExprFromAstNode(top_node)
					if err != nil {
						src_file.notices.PreSema = append(src_file.notices.PreSema, err)
					} else if expr != nil {
						top_level = append(top_level, expr)
					}
				}
			}
			any_pre_errs = any_pre_errs || has_brace_errs || (len(src_file.notices.PreSema) > 0)
			encounteredDiagsRelevantChanges = encounteredDiagsRelevantChanges || (len(src_file.notices.PreSema) > 0) || had_errs || has_brace_errs
		}
	}
	me.Sema.Pre = top_level.sorted()

	old_had_errs := me.Sema.Post.AnyErrs()
	if any_pre_errs && !old_had_errs { // bug out & leave the old `.Sema.Post` intact in this case, for editor clients
		return
	}

	if me.Interp == nil {
		_ = newInterp(me, nil)
		util.Assert(me.Interp != nil, nil)
	}

	me.Sema.Post = nil
	me.Interp.resetEnv()
	eval := func(expr *MoExpr) {
		if evaled := me.Interp.Eval(expr); evaled != nil {
			me.Sema.Post = append(me.Sema.Post, evaled)
			encounteredDiagsRelevantChanges = encounteredDiagsRelevantChanges || evaled.HasErrs() || evaled.EqNever()
		}
	}
	// first, only top-level `@set` exprs, so that names are in env
	for _, top_expr := range me.Sema.Pre {
		if me.Interp.isSetCall(top_expr) {
			eval(top_expr)
		}
	}
	// now, all non-`@set` top-level exprs
	for _, top_expr := range me.Sema.Pre {
		if !me.Interp.isSetCall(top_expr) {
			eval(top_expr)
		}
	}
	me.Sema.Post = me.Sema.Post.sorted()
	encounteredDiagsRelevantChanges = encounteredDiagsRelevantChanges || old_had_errs || me.Sema.Post.AnyErrs()
	return
}

func (me MoExprs) sorted() MoExprs {
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
