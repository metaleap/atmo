package lsp

import (
	lsp "atmo/lsp/sdk"
	"atmo/session"
	"atmo/util"
	"atmo/util/sl"
)

func init() {
	session.OnNoticesChanged = func() {
		util.Assert(Server.Initialized.Client != nil && Server.Initialized.Server != nil, nil)
		session.WithAllCurrentSrcFileNoticesDo(func(all_notices map[string][]*session.SrcFileNotice) {
			for file_path, diags := range all_notices {
				Server.Notify_textDocument_publishDiagnostics(lsp.PublishDiagnosticsParams{
					Uri:         lsp.FsPathToLspUri(file_path),
					Diagnostics: sl.As(diags, srcFileNoticeToLspDiag),
				})
			}
		})
	}

	Server.On_textDocument_codeAction = func(params *lsp.CodeActionParams) (any, error) {
		var ret []lsp.CodeAction
		src_file_path := lsp.LspUriToFsPath(params.TextDocument.Uri)

		if session.IsSrcFilePath(src_file_path) {
			// if a text selection and not our VSCode extension, offer an Eval command
			if (!ClientIsAtmoVscExt) && (params.Range.Start != params.Range.End) {
				ret = append(ret, lsp.CodeAction{
					Title:   "Eval",
					Command: &lsp.Command{Title: "Eval", Command: "eval", Arguments: []any{params}},
				})
			}

			// gather any actions deriving from current `SrcFileNotice`s on the file, if any
			session.WithAllCurrentSrcFileNoticesDo(func(all_notices map[string][]*session.SrcFileNotice) {
				notices := all_notices[src_file_path]
				if len(notices) > 0 {
					session.WithSrcFileDo(src_file_path, true, func(srcFile *session.SrcFile) {
						for _, it := range notices {
							switch it.Code {
							case session.NoticeCodeWhitespace:
								// keep comment here. we dont offer quick-fix, having tried it out:
								// CR->LF replacings are ignored by vscode; sensible tab-replacings would require knowing the current client-side tab size
							}
						}
					})
				}
			})
		}

		return ret, nil
	}
}

func srcFileNoticeToLspDiag(it *session.SrcFileNotice) lsp.Diagnostic {
	return lsp.Diagnostic{
		Code:            string(it.Code),
		CodeDescription: &lsp.CodeDescription{Href: "https://github.com/atmo-lang/atmo/docs/err-codes.md#" + string(it.Code)},
		Range:           toLspRange(it.Span),
		Message:         it.Message,
		Severity:        toLspDiagSeverity(it.Kind),
	}
}
