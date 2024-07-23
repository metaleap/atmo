package lsp

import (
	"errors"
	"path/filepath"

	lsp "atmo/lsp/sdk"
	"atmo/session"
	"atmo/util"
	"atmo/util/sl"
	"atmo/util/str"
)

func init() {
	Server.Lang.Commands = []string{"announceAtmoVscExt", "eval", "pkgsFsRefresh", "getSrcPkgs",
		"getSrcFileToks", "getSrcFileAst"}
	Server.On_workspace_executeCommand = executeCommand
}

func executeCommand(params *lsp.ExecuteCommandParams) (ret any, err error) {
	switch params.Command {

	default:
		err = errors.New("unknown command or invalid `arguments`: '" + params.Command + "'")

	case "announceAtmoVscExt":
		ClientIsAtmoVscExt = true

	case "eval":
		if len(params.Arguments) == 1 {
			code_action_params, err_json := util.JsonAs[lsp.CodeActionParams](params.Arguments[0])
			ret, err = str.Fmt("TODO: summon le Eval overlord for '%s' @ %d,%d", lsp.LspUriToFsPath(code_action_params.TextDocument.Uri)), err_json
		}

	case "pkgsFsRefresh":
		session.LockedDo(session.StateAccess.PacksFsRefresh)

	case "getSrcPkgs":
		session.LockedDo(func(sess session.StateAccess) {
			ret = sess.AllCurrentSrcPacks()
		})

	case "getSrcPkgEst":
		if len(params.Arguments) == 1 {
			src_file_path, ok := params.Arguments[0].(string)
			if ok && session.IsSrcFilePath(src_file_path) {
				session.LockedDo(func(sess session.StateAccess) {
					if src_pkg := sess.GetSrcPack(filepath.Dir(src_file_path), true); src_pkg != nil {
						type EstNode struct {
							*session.EstNode
							ClientInfo struct {
								SrcFilePath string               `json:",omitempty"`
								SrcFileSpan *session.SrcFileSpan `json:",omitempty"`
								SrcFileText string               `json:",omitempty"`
							} `json:",omitempty"`
						}
						convert := func(it *session.EstNode) *EstNode {
							ret := &EstNode{EstNode: it}
							if it.SrcFile != nil {
								ret.ClientInfo.SrcFilePath = it.SrcFile.FilePath
								if it.SrcNode != nil {
									ret.ClientInfo.SrcFileSpan = util.Ptr(it.SrcNode.Toks.Span())
									ret.ClientInfo.SrcFileText = it.SrcNode.Src
								}
							}
							return ret
						}
						ret = sl.As(src_pkg.Est, convert)
					}
				})
			}
		}

	case "getSrcFileToks":
		if len(params.Arguments) == 1 {
			src_file_path, ok := params.Arguments[0].(string)
			if ok && session.IsSrcFilePath(src_file_path) {
				session.LockedDo(func(sess session.StateAccess) {
					if src_file := sess.SrcFile(src_file_path, true); src_file != nil {
						ret = src_file.Src.Toks
					}
				})
			}
		}

	case "getSrcFileAst":
		if len(params.Arguments) == 1 {
			src_file_path, ok := params.Arguments[0].(string)
			if ok && session.IsSrcFilePath(src_file_path) {
				session.LockedDo(func(sess session.StateAccess) {
					if src_file := sess.SrcFile(src_file_path, true); src_file != nil {
						ret = src_file.Src.Ast
					}
				})
			}
		}

	}

	return
}
