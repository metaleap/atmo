package lsp

import (
	"errors"

	"atmo/session"
	"atmo/util"
	"atmo/util/sl"
	"atmo/util/str"

	lsp "github.com/metaleap/polyglot-lsp/lang_go/lsp_v3.17"
)

var Server lsp.Server
var ClientIsAtmoVscExt bool

func init() {
	session.OnNoticesChanged = func(pub map[string][]*session.SrcFileNotice) {
		return
		util.Assert(Server.Initialized.Client != nil && Server.Initialized.Server != nil, nil)
		for file_path, diags := range pub {
			Server.Notify_textDocument_publishDiagnostics(lsp.PublishDiagnosticsParams{
				Uri: lsp.String(toUri(file_path)),
				Diagnostics: sl.As(diags, func(it *session.SrcFileNotice) lsp.Diagnostic {
					code := str.Fmt("%04d", it.Code)
					return lsp.Diagnostic{
						Code:            &lsp.IntegerOrString{String: ptr(lsp.String(code))},
						CodeDescription: &lsp.CodeDescription{Href: lsp.String("https://github.com/metaleap/atom/docs/errs.md#" + code)},
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
	return lsp.Position{Line: uint(pos.Line), Character: uint(pos.Char)}
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
			return str.Fmt("TODO: summon le Eval overlord for '%s' @ %d,%d", toFsPath(code_action_params.TextDocument.Uri)), err
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
