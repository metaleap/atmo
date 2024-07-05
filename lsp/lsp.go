package atmo_lsp

import (
	lsp "github.com/metaleap/polyglot-lsp/lang_go/lsp_v3.17"
)

var Server lsp.Server

func Main() {
	Server.LogPrefixSendRecvJsons = "atmo"
	Server.Lang.Commands = []string{}
	Server.Lang.CompletionTriggerChars = []string{"."}
	Server.Lang.SignatureTriggerChars = []string{","}

	Server.On_workspace_executeCommand = func(params *lsp.ExecuteCommandParams) (any, error) {
		if params.Command == "announce-atmo-vscode-ext" {
		}
		return nil, nil
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
