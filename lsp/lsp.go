package lsp

import (
	"errors"

	lsp "atmo/lsp/sdk"
	"atmo/session"
	"atmo/util"
	"atmo/util/sl"
	"atmo/util/str"
)

var Server lsp.Server
var ClientIsAtmoVscExt bool

func init() {
	session.OnNoticesChanged = func(pub map[string][]*session.SrcFileNotice) {
		util.Assert(Server.Initialized.Client != nil && Server.Initialized.Server != nil, nil)
		for file_path, diags := range pub {
			Server.Notify_textDocument_publishDiagnostics(lsp.PublishDiagnosticsParams{
				Uri: lsp.FsPathToUri(file_path),
				Diagnostics: sl.As(diags, func(it *session.SrcFileNotice) lsp.Diagnostic {
					code := str.Fmt("%04d", it.Code)
					return lsp.Diagnostic{
						Code:            code,
						CodeDescription: &lsp.CodeDescription{Href: "https://github.com/metaleap/atom/docs/errs.md#" + code},
						Range:           toLspRange(it.Span),
						Message:         it.Message,
						Severity:        toLspDiagSeverity(it.Kind),
					}
				}),
			})
		}
	}
}

func toLspPos(pos session.SrcFilePos) lsp.Position {
	return lsp.Position{Line: util.If(pos.Line <= 0, 0, pos.Line-1), Character: util.If(pos.Char <= 0, 0, pos.Char-1)}
}

func toLspRange(span session.SrcFileSpan) lsp.Range {
	return lsp.Range{Start: toLspPos(span.Start), End: toLspPos(span.End)}
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

func Main() {
	Server.LogPrefixSendRecvJsons = "atmo"
	Server.Lang.Commands = []string{"announceAtmoVscExt", "eval"}
	Server.Lang.TriggerChars.Completion = []string{"."}
	Server.Lang.TriggerChars.Signature = []string{","}
	Server.Lang.DocumentSymbolsMultiTreeLabel = "Atmo"

	Server.On_workspace_executeCommand = func(params *lsp.ExecuteCommandParams) (any, error) {
		switch params.Command {
		case "announceAtmoVscExt":
			ClientIsAtmoVscExt = true
			return nil, nil
		case "eval":
			code_action_params, err := util.JsonAs[lsp.CodeActionParams](params.Arguments[0])
			return str.Fmt("TODO: summon le Eval overlord for '%s' @ %d,%d", lsp.UriToFsPath(code_action_params.TextDocument.Uri)), err
		}
		return nil, errors.New("unknown command: '" + params.Command + "'")
	}

	panic(Server.Forever())
}

func ptr[T any](it T) *T { return &it }

func dePtr[T any](ptr *T) (ret T) {
	if ptr != nil {
		ret = *ptr
	}
	return
}
