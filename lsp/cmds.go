package lsp

import (
	"errors"

	lsp "atmo/lsp/sdk"
	"atmo/session"
	"atmo/util"
	"atmo/util/str"
)

func init() {
	Server.Lang.Commands = []string{"announceAtmoVscExt", "eval", "getSrcFileToks", "getSrcFileAstOrig"}
	Server.On_workspace_executeCommand = executeCommand
}

func executeCommand(params *lsp.ExecuteCommandParams) (ret any, err error) {
	switch params.Command {

	default:
		err = errors.New("unknown command or invalid `arguments`: '" + params.Command + "'")

	case "announceAtmoVscExt":
		ClientIsAtmoVscExt = true
		return

	case "eval":
		if len(params.Arguments) == 1 {
			code_action_params, err_json := util.JsonAs[lsp.CodeActionParams](params.Arguments[0])
			ret, err = str.Fmt("TODO: summon le Eval overlord for '%s' @ %d,%d", lsp.LspUriToFsPath(code_action_params.TextDocument.Uri)), err_json
		}

	case "getSrcFileToks":
		if len(params.Arguments) == 1 {
			src_file_path, ok := params.Arguments[0].(string)
			if ok && session.IsSrcFilePath(src_file_path) {
				session.WithSrcFileDo(src_file_path, true, func(srcFile *session.SrcFile) {
					ret = srcFile.Content.Toks
				})
				return
			}
		}

	case "getSrcFileAstOrig":
		if len(params.Arguments) == 1 {
			src_file_path, ok := params.Arguments[0].(string)
			if ok && session.IsSrcFilePath(src_file_path) {
				session.WithSrcFileDo(src_file_path, true, func(srcFile *session.SrcFile) {
					ret = srcFile.Content.Ast
				})
				return
			}
		}

	}

	return
}
