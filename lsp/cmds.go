package lsp

import (
	"errors"
	"path/filepath"

	lsp "atmo/lsp/sdk"
	"atmo/session"
	"atmo/util"
	"atmo/util/kv"
	"atmo/util/sl"
	"atmo/util/str"
)

func init() {
	Server.Lang.Commands = []string{"announceAtmoVscExt", "eval",
		"packsFsRefresh", "getSrcPacks", "getSrcFileToks", "getSrcFileAst", "getSrcPackMoOrig", "getSrcPackMoSem"}
	Server.On_workspace_executeCommand = executeCommand
}

type treeNodeClientInfo struct {
	SrcFilePath string               `json:",omitempty"`
	SrcFileSpan *session.SrcFileSpan `json:",omitempty"`
	SrcFileText string               `json:",omitempty"`
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
			ret, err = str.Fmt("TODO: summon le Eval overlord for '%s' @ %d,%d", lspUriToFsPath(code_action_params.TextDocument.Uri)), err_json
		}

	case "packsFsRefresh":
		session.Access(func(sess session.StateAccess, _ session.Intel) {
			sess.PacksFsRefresh()
		})

	case "getSrcPacks":
		session.Access(func(sess session.StateAccess, _ session.Intel) {
			ret = sess.AllCurrentSrcPacks()
		})

	case "getSrcFileToks":
		if len(params.Arguments) == 1 {
			src_file_path, ok := params.Arguments[0].(string)
			if ok && session.IsSrcFilePath(src_file_path) {
				session.Access(func(sess session.StateAccess, _ session.Intel) {
					if src_file := sess.SrcFile(src_file_path); src_file != nil {
						ret = src_file.Src.Toks
					}
				})
			}
		}

	case "getSrcFileAst":
		if len(params.Arguments) == 1 {
			src_file_path, ok := params.Arguments[0].(string)
			if ok && session.IsSrcFilePath(src_file_path) {
				session.Access(func(sess session.StateAccess, _ session.Intel) {
					if src_file := sess.SrcFile(src_file_path); src_file != nil {
						ret = sl.SortedPer(src_file.Src.Ast, (*session.AstNode).Cmp)
					}
				})
			}
		}

	case "getSrcPackMoOrig":
		type moNode struct {
			PrimTypeTag session.MoValPrimType
			Nodes       []*moNode
			ClientInfo  treeNodeClientInfo
		}
		if len(params.Arguments) == 1 {
			src_file_path, ok := params.Arguments[0].(string)
			if ok && session.IsSrcFilePath(src_file_path) {
				session.Access(func(sess session.StateAccess, _ session.Intel) {
					if src_pack := sess.GetSrcPack(filepath.Dir(src_file_path), true); src_pack != nil {
						var convert func(*session.MoExpr) *moNode
						convert = func(expr *session.MoExpr) *moNode {
							ret := moNode{PrimTypeTag: expr.Val.PrimType()}
							if expr.IsErr() {
								ret.PrimTypeTag = session.MoPrimTypeErr
							}
							switch it := expr.Val.(type) {
							case session.MoValCall:
								for _, item := range it {
									ret.Nodes = append(ret.Nodes, convert(item))
								}
							case session.MoValDict:
								const fake_prim_tag_dictentry = -1
								for _, pair := range it {
									node := &moNode{PrimTypeTag: fake_prim_tag_dictentry, Nodes: []*moNode{convert(pair.Key), convert(pair.Val)}}
									node.ClientInfo.SrcFilePath = node.Nodes[0].ClientInfo.SrcFilePath
									node.ClientInfo.SrcFileSpan = node.Nodes[0].ClientInfo.SrcFileSpan.Expanded(node.Nodes[1].ClientInfo.SrcFileSpan)
									node.ClientInfo.SrcFileText = node.Nodes[0].ClientInfo.SrcFileText + ": " + node.Nodes[1].ClientInfo.SrcFileText
									ret.Nodes = append(ret.Nodes, node)
								}
							case session.MoValErr:
								node := moNode{PrimTypeTag: session.MoPrimTypeErr}
								if it.ErrVal != nil {
									node.Nodes = append(node.Nodes, convert(it.ErrVal))
									node.ClientInfo.SrcFilePath = node.Nodes[0].ClientInfo.SrcFilePath
									node.ClientInfo.SrcFileSpan = node.Nodes[0].ClientInfo.SrcFileSpan.Expanded(node.Nodes[1].ClientInfo.SrcFileSpan)
									node.ClientInfo.SrcFileText = node.Nodes[0].ClientInfo.SrcFileText + ": " + node.Nodes[1].ClientInfo.SrcFileText
								}
								ret.Nodes = append(ret.Nodes, &node)
							case *session.MoValFnLam:
								param_nodes := sl.To(it.Params, convert)
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
						for _, top_expr := range src_pack.Trees.MoOrig.Sorted() {
							top_level = append(top_level, convert(top_expr))
						}
						ret = top_level
					}
				})
			}
		}

	case "getSrcPackMoSem":
		src_file_path, ok := params.Arguments[0].(string)
		if ok && session.IsSrcFilePath(src_file_path) {
			session.Access(func(sess session.StateAccess, _ session.Intel) {
				if src_pack := sess.GetSrcPack(filepath.Dir(src_file_path), true); src_pack != nil {
					type SemNode struct {
						session.SemExpr
						ClientInfo treeNodeClientInfo `json:",omitempty"`
					}
					var convert func(from *session.SemExpr) (ret SemNode)
					convert = func(from *session.SemExpr) (ret SemNode) {
						ret.SemExpr = *from
						if from.From != nil {
							ret.ClientInfo.SrcFileSpan = from.From.SrcSpan
							if from.From.SrcFile != nil {
								ret.ClientInfo.SrcFilePath = from.From.SrcFile.FilePath
							}
							if from.From.SrcNode != nil {
								ret.ClientInfo.SrcFileText = from.From.SrcNode.Src
							}
						}
						switch val := from.Val.(type) {
						default:
							panic(val)
						case *session.SemValIdent:
							ret.Val = map[string]any{"Kind": "scalar", "MoVal": val.MoVal}
						case *session.SemValScalar:
							ret.Val = map[string]any{"Kind": "scalar", "MoVal": val.MoVal}
						case *session.SemValList:
							ret.Val = map[string]any{"Kind": "list", "Items": sl.To(val.Items, convert)}
						case *session.SemValDict:
							ret.Val = map[string]any{"Kind": "dict", "Keys": sl.To(val.Keys, convert), "Vals": sl.To(val.Vals, convert)}
						case *session.SemValCall:
							ret.Val = map[string]any{"Kind": "call", "Callee": convert(val.Callee), "Args": sl.To(val.Args, convert)}
						case *session.SemValFunc:
							ret.Val = map[string]any{"Kind": "func", "Params": sl.To(val.Params, convert), "Body": convert(val.Body), "IsMacro": val.IsMacro,
								"Scope": kv.Keys(val.Scope.Own)}
						}
						return
					}

					ret = sl.To(src_pack.Trees.Sem.TopLevel, convert)
				}
			})
		}

	}

	return
}
