package lsp

import (
	lsp "atmo/lsp/sdk"
	"atmo/session"
	"atmo/util"
	"atmo/util/sl"
)

func init() {
	session.OnNoticesChanged = func(pub map[string][]*session.SrcFileNotice) {
		util.Assert(Server.Initialized.Client != nil && Server.Initialized.Server != nil, nil)
		for file_path, diags := range pub {
			Server.Notify_textDocument_publishDiagnostics(lsp.PublishDiagnosticsParams{
				Uri: lsp.FsPathToLspUri(file_path),
				Diagnostics: sl.As(diags, func(it *session.SrcFileNotice) lsp.Diagnostic {
					return lsp.Diagnostic{
						Code:            string(it.Code),
						CodeDescription: &lsp.CodeDescription{Href: "https://github.com/atmo-lang/atmo/docs/err-codes.md#" + string(it.Code)},
						Range:           toLspRange(it.Span),
						Message:         it.Message,
						Severity:        toLspDiagSeverity(it.Kind),
					}
				}),
			})
		}
	}
}
