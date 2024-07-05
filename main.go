package main

import (
	lsp "github.com/metaleap/polyglot-lsp/lang_go/lsp_v3.17"
)

func main() {
	var lsp_server lsp.Server

	lsp_server.LogPrefixSendRecvJsons = "atom"

	lsp_server.On_textDocument_hover = func(params *lsp.HoverParams) (any, error) {
		str := lsp.String("Test Hover")
		return &lsp.Hover{
			Contents: &lsp.MarkupContentOrMarkedStringOrMarkedStrings{MarkedString: &lsp.StringOrLanguageStringWithValueString{String: &str}},
		}, nil
	}

	panic(lsp_server.Forever())
}
