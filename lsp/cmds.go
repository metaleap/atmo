package lsp

import (
	"errors"

	lsp "atmo/lsp/sdk"
	"atmo/session"
	"atmo/util"
	"atmo/util/str"
)

func init() {
	Server.Lang.Commands = []string{"announceAtmoVscExt", "eval", "getSrcFileToks"}
	Server.On_workspace_executeCommand = executeCommand
}

func executeCommand(params *lsp.ExecuteCommandParams) (any, error) {
	switch params.Command {

	case "announceAtmoVscExt":
		ClientIsAtmoVscExt = true
		return nil, nil

	case "eval":
		if len(params.Arguments) == 1 {
			code_action_params, err := util.JsonAs[lsp.CodeActionParams](params.Arguments[0])
			return str.Fmt("TODO: summon le Eval overlord for '%s' @ %d,%d", lsp.UriToFsPath(code_action_params.TextDocument.Uri)), err
		}

	case "getSrcFileToks":
		if len(params.Arguments) == 1 {
			src_file_path, ok := params.Arguments[0].(string)
			if ok && session.IsSrcFilePath(src_file_path) {
				src_file := session.EnsureSrcFile(src_file_path, nil, true)
				return src_file.Content.TopLevelToksChunks, nil
			}
		}

	}
	return nil, errors.New("unknown command or invalid `arguments`: '" + params.Command + "'")
}
