package lsp

import (
	lsp "atmo/lsp/sdk"
	"atmo/session"
	"atmo/util"
)

var Server = lsp.Server{LogPrefixSendRecvJsons: "atmo"}
var ClientIsAtmoVscExt bool

func Main() {
	panic(Server.Forever())
}

func init() {
	Server.Lang.DocumentSymbolsMultiTreeLabel = "Atmo"
	Server.Lang.TriggerChars.Completion = []string{"."}
	Server.Lang.TriggerChars.Signature = []string{","}
}

func toLspPos(pos session.SrcFilePos) lsp.Position {
	return lsp.Position{Line: util.If(pos.Line <= 0, 0, pos.Line-1), Character: util.If(pos.Char <= 0, 0, pos.Char-1)}
}

func toLspRange(span session.SrcFileSpan) lsp.Range {
	return lsp.Range{Start: toLspPos(span.Start), End: toLspPos(span.End)}
}
