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

	case "getSrcPkgMo":
		type moNode struct {
			PrimTypeTag session.MoValPrimType
			Nodes       []*moNode
			ClientInfo  struct {
				SrcFilePath string               `json:",omitempty"`
				SrcFileSpan *session.SrcFileSpan `json:",omitempty"`
				SrcFileText string               `json:",omitempty"`
			}
		}
		if len(params.Arguments) == 1 {
			src_file_path, ok := params.Arguments[0].(string)
			if ok && session.IsSrcFilePath(src_file_path) {
				session.LockedDo(func(sess session.StateAccess) {
					if src_pkg := sess.GetSrcPack(filepath.Dir(src_file_path), true); src_pkg != nil {
						var convert func(*session.MoExpr) *moNode
						convert = func(expr *session.MoExpr) *moNode {
							ret := moNode{PrimTypeTag: expr.Val.PrimType()}
							switch it := expr.Val.(type) {
							case session.MoValCall:
								for _, item := range it {
									ret.Nodes = append(ret.Nodes, convert(item))
								}
							case session.MoValDict:
								for _, pair := range it {
									node := &moNode{PrimTypeTag: session.MoPrimTypeDict, Nodes: []*moNode{convert(pair[0]), convert(pair[1])}}
									node.ClientInfo.SrcFilePath = node.Nodes[0].ClientInfo.SrcFilePath
									node.ClientInfo.SrcFileSpan = node.Nodes[0].ClientInfo.SrcFileSpan.ExtendBy(node.Nodes[1].ClientInfo.SrcFileSpan)
									node.ClientInfo.SrcFileText = node.Nodes[0].ClientInfo.SrcFileText + ": " + node.Nodes[1].ClientInfo.SrcFileText
									ret.Nodes = append(ret.Nodes, node)
								}
							case session.MoValErr:
								ret.Nodes = append(ret.Nodes, convert(it.Err))
							case *session.MoValFnLam:
								param_nodes := sl.As(it.Params, convert)
								ret.Nodes = append(ret.Nodes, &moNode{PrimTypeTag: session.MoPrimTypeFunc, Nodes: param_nodes}, convert(it.Body))
							case session.MoValList:
								for _, item := range it {
									ret.Nodes = append(ret.Nodes, convert(item))
								}
							}
							if expr.SrcSpan != nil {
								ret.ClientInfo.SrcFileSpan = expr.SrcSpan
							}
							if expr.SrcFile != nil {
								ret.ClientInfo.SrcFilePath = expr.SrcFile.FilePath
							}
							if (expr.SrcFile != nil) && (expr.SrcSpan != nil) {
								if node := expr.SrcFile.NodeAtSpan(expr.SrcSpan); node != nil {
									ret.ClientInfo.SrcFileText = node.Src
								}
							}
							return &ret
						}
						var top_level []*moNode
						for _, top_expr := range src_pkg.Sema.Pre {
							top_level = append(top_level, convert(top_expr))
						}
						ret = top_level
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
