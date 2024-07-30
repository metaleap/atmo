package lsp

import (
	lsp "atmo/lsp/sdk"
	"atmo/session"
	"atmo/util"
	"atmo/util/sl"
	"atmo/util/str"
)

func init() {
	session.OnDiagsChanged = func() {
		util.Assert(Server.Initialized.Fully, nil)
		session.Access(func(sess session.StateAccess, _ session.Intel) {
			all_diags := sess.AllCurrentSrcFileDiags()
			for file_path, diags := range all_diags {
				Server.Notify_textDocument_publishDiagnostics(lsp.PublishDiagnosticsParams{
					Uri:         lspUriFromFsPath(file_path),
					Diagnostics: sl.To(diags, diagToLspDiag),
				})
			}
		})
	}

	Server.On_textDocument_codeAction = func(params *lsp.CodeActionParams) (ret []lsp.CodeAction, _ error) {
		src_file_path := lspUriToFsPath(params.TextDocument.Uri)

		if session.IsSrcFilePath(src_file_path) {
			// if a text selection and not our VSCode extension, offer an Eval command
			if (!ClientIsAtmoVscExt) && (params.Range.Start != params.Range.End) {
				ret = append(ret, lsp.CodeAction{
					Title:   "Eval",
					Command: &lsp.Command{Title: "Eval", Command: "eval", Arguments: []any{params}},
				})
			}

			// gather any actions deriving from current `Diag`s on the file, if any
			session.Access(func(sess session.StateAccess, _ session.Intel) {
				diags := sess.AllCurrentSrcFileDiags()[src_file_path]
				if len(diags) == 0 {
					return
				}
				src_file := sess.SrcFile(src_file_path)
				if src_file == nil {
					return
				}
				for _, it := range diags {
					switch it.Code {
					case session.ErrCodeIndentation:
						if src_file.Src.Toks[0].Pos.Char > 1 {
							diags := []lsp.Diagnostic{diagToLspDiag(it)}
							cmd_title := "Fix first-line mis-indentation"
							ret = append(ret, lsp.CodeAction{
								Title:       cmd_title,
								Kind:        lsp.CodeActionKindQuickFix,
								Diagnostics: diags,
								Edit: &lsp.WorkspaceEdit{Changes: map[string][]lsp.TextEdit{
									src_file_path: {{NewText: str.Trim(src_file.Src.Text), Range: lspRangeFromSpan(util.Ptr(src_file.Span()))}},
								}},
							})
						}
					case session.ErrCodeWhitespace:
						if ClientIsAtmoVscExt {
							diags := []lsp.Diagnostic{diagToLspDiag(it)}
							if cmd_title := "Convert all line-leading tabs to spaces"; str.Idx(src_file.Src.Text, '\t') >= 0 {
								ret = append(ret, lsp.CodeAction{
									Title:       cmd_title,
									Kind:        lsp.CodeActionKindQuickFix,
									Diagnostics: diags,
									Command:     &lsp.Command{Title: cmd_title, Command: "editor.action.indentationToSpaces"},
								})
							}
							if cmd_title := "Fix end-of-line sequences"; str.Idx(src_file.Src.Text, '\r') >= 0 {
								ret = append(ret, lsp.CodeAction{
									Title:       cmd_title,
									Kind:        lsp.CodeActionKindQuickFix,
									Diagnostics: diags,
									Command:     &lsp.Command{Title: cmd_title, Command: "workbench.action.editor.changeEOL"},
								})
							}
						}
					}
				}
			})
		}
		return
	}
}

func diagToLspDiag(it *session.Diag) lsp.Diagnostic {
	ret := lsp.Diagnostic{
		Code:            string(it.Code),
		CodeDescription: &lsp.CodeDescription{Href: "https://github.com/atmo-lang/atmo/docs/err-codes.md#" + string(it.Code)},
		Range:           lspRangeFromSpan(&it.Span),
		Message:         it.Message,
		Severity:        toLspDiagSeverity(it.Kind),
		Source:          "atmo",
	}
	if it.Code == session.HintCodeUnused {
		ret.Tags = append(ret.Tags, lsp.DiagnosticTagUnnecessary)
	}
	for _, locs := range it.Rel {
		for i, span := range locs.Spans {
			hint := "namely, here"
			if len(locs.Hints) == len(locs.Spans) {
				hint = locs.Hints[i]
			}
			ret.RelatedInformation = append(ret.RelatedInformation, lsp.DiagnosticRelatedInformation{
				Location: lsp.Location{Uri: lspUriFromFsPath(locs.File.FilePath), Range: lspRangeFromSpan(span)},
				Message:  hint,
			})
		}
	}
	return ret
}

func toLspDiagSeverity(kind session.DiagKind) lsp.DiagnosticSeverity {
	switch kind {
	case session.DiagKindErr:
		return lsp.DiagnosticSeverityError
	case session.DiagKindWarn:
		return lsp.DiagnosticSeverityWarning
	case session.DiagKindInfo:
		return lsp.DiagnosticSeverityInformation
	case session.DiagKindHint:
		return lsp.DiagnosticSeverityHint
	default:
		panic(kind)
	}
}
