package main

import (
	lsp "github.com/metaleap/polyglot-lsp/lang_go/lsp_v3.17"
)

func main() {
	var lsp_server lsp.Server

	lsp_server.LogPrefixSendRecvJsons = "atom"

	lsp_server.On_textDocument_hover = func(params *lsp.HoverParams) (any, error) {
		str := lsp.String("**Test** _Hover_")
		return &lsp.Hover{
			Contents: &lsp.MarkupContentOrMarkedStringOrMarkedStrings{MarkedString: &lsp.StringOrLanguageStringWithValueString{String: &str}},
		}, nil
	}

	lsp_server.On_workspace_executeCommand = func(params *lsp.ExecuteCommandParams) (any, error) {
		if params.Command == "announce-atmo-vscode-ext" {
		}
		return nil, nil
	}

	panic(lsp_server.Forever())
}
