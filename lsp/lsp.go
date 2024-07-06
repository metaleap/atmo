package atmo_lsp

import (
	"errors"

	"atmo/util"
	"atmo/util/str"

	lsp "github.com/metaleap/polyglot-lsp/lang_go/lsp_v3.17"
)

var Server lsp.Server
var ClientIsAtmoVscExt bool

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
			cap, err := util.JsonAs[lsp.CodeActionParams](params.Arguments[0])
			return str.Fmt("TODO: summon le Eval overlord for '%s' @ %d,%d", srcFilePath(cap.TextDocument)), err
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
