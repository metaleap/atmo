package lsp

import (
	lsp "atmo/lsp/sdk"
	"atmo/session"
	"atmo/util"
	"atmo/util/str"
)

const (
	logJsonMsgs = false
)

var (
	Server             = lsp.Server{LogPrefixSendRecvJsons: util.If(logJsonMsgs, "atmo", "")}
	ClientIsAtmoVscExt bool
)

func Main() {
	panic(Server.Forever())
}

func init() {
	Server.Lang.DocumentSymbolsMultiTreeLabel = "Atmo"
	Server.Lang.TriggerChars.Completion = []string{"."}
	Server.Lang.TriggerChars.Signature = []string{","}
	session.OnDbgMsg = func(should bool, msg string, args ...any) {
		if should {
			if len(args) > 0 {
				msg = str.Fmt(msg, args...)
			}
			Server.Notify_window_showMessage(lsp.ShowMessageParams{Type: lsp.MessageTypeInfo, Message: "DBG:" + msg})
		}
	}
}

func toLspPos(pos session.SrcFilePos) lsp.Position {
	return lsp.Position{Line: util.If(pos.Line <= 0, 0, pos.Line-1), Character: util.If(pos.Char <= 0, 0, pos.Char-1)}
}

func toLspRange(span session.SrcFileSpan) lsp.Range {
	return lsp.Range{Start: toLspPos(span.Start), End: toLspPos(span.End)}
}
