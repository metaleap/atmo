package lsp

import (
	"fmt"

	lsp "atmo/lsp/sdk"
	"atmo/session"
	"atmo/util"
	"atmo/util/sl"
	"atmo/util/str"
)

func init() {
	Server.On_textDocument_documentSymbol = func(params *lsp.DocumentSymbolParams) ([]lsp.DocumentSymbol, error) {
		src_file_path := lsp.LspUriToFsPath(params.TextDocument.Uri)
		return sl.As([]lsp.SymbolKind{lsp.SymbolKindArray, lsp.SymbolKindBoolean, lsp.SymbolKindClass, lsp.SymbolKindConstant, lsp.SymbolKindConstructor, lsp.SymbolKindEnum, lsp.SymbolKindEnumMember, lsp.SymbolKindEvent, lsp.SymbolKindField, lsp.SymbolKindFile, lsp.SymbolKindFunction, lsp.SymbolKindInterface, lsp.SymbolKindKey, lsp.SymbolKindMethod, lsp.SymbolKindModule, lsp.SymbolKindNamespace, lsp.SymbolKindNull, lsp.SymbolKindNumber, lsp.SymbolKindObject, lsp.SymbolKindOperator, lsp.SymbolKindPackage, lsp.SymbolKindProperty, lsp.SymbolKindString, lsp.SymbolKindStruct, lsp.SymbolKindTypeParameter, lsp.SymbolKindVariable},
			func(it lsp.SymbolKind) lsp.DocumentSymbol {
				return lsp.DocumentSymbol{
					Name:           it.String(),
					Detail:         fmt.Sprintf("**TODO:** documentSymbols for `%v`", src_file_path),
					Kind:           it,
					Range:          lsp.Range{Start: lsp.Position{Line: 2, Character: 1}, End: lsp.Position{Line: 2, Character: 8}},
					SelectionRange: lsp.Range{Start: lsp.Position{Line: 2, Character: 3}, End: lsp.Position{Line: 2, Character: 6}},
				}
			}), nil
	}

	Server.On_workspace_symbol = func(params *lsp.WorkspaceSymbolParams) ([]lsp.WorkspaceSymbol, error) {
		return []lsp.WorkspaceSymbol{{
			Name:          "Atmo",
			Kind:          lsp.SymbolKindInterface,
			ContainerName: "Container",
			Location: lsp.Location{
				Uri:   lsp.FsPathToLspUri("/home/_/c/at/foo.at"),
				Range: lsp.Range{Start: lsp.Position{Line: 2, Character: 1}, End: lsp.Position{Line: 2, Character: 8}},
			},
		}}, nil
	}

	Server.On_textDocument_definition = func(params *lsp.DefinitionParams) (any, error) {
		return dummyLocs(lsp.LspUriToFsPath(params.TextDocument.Uri)), nil
	}

	Server.On_textDocument_declaration = func(params *lsp.DeclarationParams) (any, error) {
		return dummyLocs(lsp.LspUriToFsPath(params.TextDocument.Uri)), nil
	}

	Server.On_textDocument_typeDefinition = func(params *lsp.TypeDefinitionParams) (any, error) {
		return dummyLocs(lsp.LspUriToFsPath(params.TextDocument.Uri)), nil
	}

	Server.On_textDocument_implementation = func(params *lsp.ImplementationParams) (any, error) {
		return dummyLocs(lsp.LspUriToFsPath(params.TextDocument.Uri)), nil
	}

	Server.On_textDocument_references = func(params *lsp.ReferenceParams) (any, error) {
		return dummyLocs(lsp.LspUriToFsPath(params.TextDocument.Uri)), nil
	}

	Server.On_textDocument_documentHighlight = func(params *lsp.DocumentHighlightParams) (any, error) {
		return sl.As(dummyLocs(lsp.LspUriToFsPath(params.TextDocument.Uri)), func(it lsp.Location) lsp.DocumentHighlight {
			return lsp.DocumentHighlight{Range: it.Range, Kind: lsp.DocumentHighlightKindText}
		}), nil
	}

	Server.On_textDocument_completion = func(params *lsp.CompletionParams) (any, error) {
		src_file_path := lsp.LspUriToFsPath(params.TextDocument.Uri)
		return sl.As([]lsp.CompletionItemKind{
			lsp.CompletionItemKindClass,
			lsp.CompletionItemKindColor,
			lsp.CompletionItemKindConstant,
			lsp.CompletionItemKindConstructor,
			lsp.CompletionItemKindEnum,
			lsp.CompletionItemKindEnumMember,
			lsp.CompletionItemKindEvent,
			lsp.CompletionItemKindField,
			lsp.CompletionItemKindFile,
			lsp.CompletionItemKindFolder,
			lsp.CompletionItemKindFunction,
			lsp.CompletionItemKindInterface,
			lsp.CompletionItemKindKeyword,
			lsp.CompletionItemKindMethod,
			lsp.CompletionItemKindModule,
			lsp.CompletionItemKindOperator,
			lsp.CompletionItemKindProperty,
			lsp.CompletionItemKindReference,
			lsp.CompletionItemKindSnippet,
			lsp.CompletionItemKindStruct,
			lsp.CompletionItemKindText,
			lsp.CompletionItemKindTypeParameter,
			lsp.CompletionItemKindUnit,
			lsp.CompletionItemKindValue,
			lsp.CompletionItemKindVariable,
		}, func(it lsp.CompletionItemKind) lsp.CompletionItem {
			return lsp.CompletionItem{
				Label: it.String(),
				Kind:  it,
				Documentation: &lsp.MarkupContent{Kind: lsp.MarkupKindMarkdown,
					Value: str.Fmt("**TODO** _%s_ for `%s` @ %d,%d", it.String(), src_file_path, params.Position.Line, params.Position.Character)},
				Detail: "Detail",
				LabelDetails: &lsp.CompletionItemLabelDetails{
					Detail:      " · LD_Detail " + it.String(),
					Description: "LD_Description " + it.String(),
				},
			}
		}), nil
	}

	Server.On_textDocument_hover = func(params *lsp.HoverParams) (*lsp.Hover, error) {
		src_file_path := lsp.LspUriToFsPath(params.TextDocument.Uri)
		str := str.Fmt("**TODO** _Hover_ for `%s` @ %d,%d", src_file_path, params.Position.Line, params.Position.Character)
		return &lsp.Hover{
			Contents: lsp.MarkupContent{Kind: lsp.MarkupKindMarkdown, Value: str},
		}, nil
	}

	Server.On_textDocument_prepareRename = func(params *lsp.PrepareRenameParams) (any, error) {
		// src_file_path := srcFilePath(params.TextDocument.Uri)
		return lsp.Range{Start: params.Position, End: lsp.Position{Line: params.Position.Line, Character: 4 + params.Position.Character}}, nil
	}

	Server.On_textDocument_rename = func(params *lsp.RenameParams) (any, error) {
		src_file_path := lsp.LspUriToFsPath(params.TextDocument.Uri)
		return lsp.WorkspaceEdit{
			Changes: map[string][]lsp.TextEdit{
				lsp.FsPathToLspUri(src_file_path): {{
					NewText: params.NewName,
					Range:   lsp.Range{Start: params.Position, End: lsp.Position{Line: params.Position.Line, Character: 4 + params.Position.Character}},
				}},
			},
		}, nil
	}

	Server.On_textDocument_signatureHelp = func(params *lsp.SignatureHelpParams) (any, error) {
		src_file_path := lsp.LspUriToFsPath(params.TextDocument.Uri)
		return lsp.SignatureHelp{
			Signatures: util.If(params.Position.Line > 0,
				nil,
				[]lsp.SignatureInformation{{
					Label: "(foo bar: #baz)",
					Documentation: &lsp.MarkupContent{
						Kind:  lsp.MarkupKindMarkdown,
						Value: str.Fmt("**TODO**: sig help for `%s` @ %d,%d", src_file_path, params.Position.Line, params.Position.Character)},
				}}),
		}, nil
	}

	Server.On_textDocument_codeAction = func(params *lsp.CodeActionParams) (any, error) {
		if ClientIsAtmoVscExt || (params.Range.Start == params.Range.End) {
			return nil, nil
		}
		return []lsp.Command{{Title: "Eval", Command: "eval", Arguments: []any{params}}}, nil
	}

	Server.On_textDocument_selectionRange = func(params *lsp.SelectionRangeParams) (any, error) {
		src_file_path := lsp.LspUriToFsPath(params.TextDocument.Uri)
		var ret []*lsp.SelectionRange
		if len(params.Positions) > 0 && session.IsSrcFilePath(src_file_path) {
			src_file := session.EnsureSrcFile(src_file_path, nil, true)
			for _, pos := range params.Positions {
				if node := src_file.NodeAt(lsp.LspPosToPos(&pos), true); node == nil {
					ret = nil
					break
				} else {
					all := sl.As(node.SelfAndAncestors(), func(it *session.AstNode) *lsp.SelectionRange {
						return &lsp.SelectionRange{Range: lsp.SpanToLspRange(it.Toks.Span())}
					})
					for i, it := range all[:len(all)-1] {
						it.Parent = all[i+1]
					}
					ret = append(ret, all[0])
				}
			}
		}
		return util.If[any](len(ret) > 0, ret, nil), nil
	}

}

func dummyLocs(srcFilePath string) []lsp.Location {
	return []lsp.Location{
		{Uri: lsp.FsPathToLspUri(srcFilePath), Range: lsp.Range{Start: lsp.Position{Line: 2, Character: 1}, End: lsp.Position{Line: 2, Character: 8}}},
		{Uri: lsp.FsPathToLspUri(srcFilePath), Range: lsp.Range{Start: lsp.Position{Line: 4, Character: 1}, End: lsp.Position{Line: 4, Character: 8}}},
	}
}

func toLspDiagSeverity(kind session.SrcFileNoticeKind) lsp.DiagnosticSeverity {
	switch kind {
	case session.NoticeKindErr:
		return lsp.DiagnosticSeverityError
	case session.NoticeKindWarn:
		return lsp.DiagnosticSeverityWarning
	case session.NoticeKindInfo:
		return lsp.DiagnosticSeverityInformation
	case session.NoticeKindHint:
		return lsp.DiagnosticSeverityHint
	default:
		panic(kind)
	}
}
