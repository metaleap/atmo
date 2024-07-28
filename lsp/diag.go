package lsp

import (
	lsp "atmo/lsp/sdk"
	"atmo/session"
	"atmo/util"
	"atmo/util/sl"
	"atmo/util/str"
)

func init() {
	session.OnNoticesChanged = func() {
		util.Assert(Server.Initialized.Fully, nil)
		session.Access(func(sess session.StateAccess, _ session.Intel) {
			all_notices := sess.AllCurrentSrcFileNotices()
			for file_path, diags := range all_notices {
				Server.Notify_textDocument_publishDiagnostics(lsp.PublishDiagnosticsParams{
					Uri:         lspUriFromFsPath(file_path),
					Diagnostics: sl.As(diags, srcFileNoticeToLspDiag),
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

			// gather any actions deriving from current `SrcFileNotice`s on the file, if any
			session.Access(func(sess session.StateAccess, _ session.Intel) {
				notices := sess.AllCurrentSrcFileNotices()[src_file_path]
				if len(notices) == 0 {
					return
				}
				src_file := sess.SrcFile(src_file_path, true)
				if src_file == nil {
					return
				}
				for _, it := range notices {
					switch it.Code {
					case session.NoticeCodeIndentation:
						if src_file.Src.Toks[0].Pos.Char > 1 {
							diags := []lsp.Diagnostic{srcFileNoticeToLspDiag(it)}
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
					case session.NoticeCodeWhitespace:
						if ClientIsAtmoVscExt {
							diags := []lsp.Diagnostic{srcFileNoticeToLspDiag(it)}
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

func srcFileNoticeToLspDiag(it *session.SrcFileNotice) lsp.Diagnostic {
	return lsp.Diagnostic{
		Code:            string(it.Code),
		CodeDescription: &lsp.CodeDescription{Href: "https://github.com/atmo-lang/atmo/docs/err-codes.md#" + string(it.Code)},
		Range:           lspRangeFromSpan(&it.Span),
		Message:         it.Message,
		Severity:        toLspDiagSeverity(it.Kind),
	}
}

func toLspDiagSeverity(kind session.SrcFileNoticeKind) lsp.DiagnosticSeverity {
	switch kind {
	case session.NoticeKindErr:
		return lsp.DiagnosticSeverityError
	case session.NoticeKindWarn:
		return lsp.DiagnosticSeverityWarning
	case session.NoticeKindInfo:
		return lsp.DiagnosticSeverityInformation
	case session.NoticeKindHint:
		return lsp.DiagnosticSeverityHint
	default:
		panic(kind)
	}
}
