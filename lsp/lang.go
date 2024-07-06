package atmo_lsp

import (
	"fmt"

	"atmo/util"
	"atmo/util/sl"
	"atmo/util/str"

	lsp "github.com/metaleap/polyglot-lsp/lang_go/lsp_v3.17"
)

func srcFilePath(it lsp.URI) string {
	return str.TrimPref(string(it), "file://")
}

func srcFileUri(srcFilePath string) string {
	return "file://" + srcFilePath
}

func init() {
	Server.On_textDocument_documentSymbol = func(params *lsp.DocumentSymbolParams) (any, error) {
		src_file_path := srcFilePath(params.TextDocument.Uri)
		return sl.As([]lsp.SymbolKind{lsp.SymbolKindArray, lsp.SymbolKindBoolean, lsp.SymbolKindClass, lsp.SymbolKindConstant, lsp.SymbolKindConstructor, lsp.SymbolKindEnum, lsp.SymbolKindEnumMember, lsp.SymbolKindEvent, lsp.SymbolKindField, lsp.SymbolKindFile, lsp.SymbolKindFunction, lsp.SymbolKindInterface, lsp.SymbolKindKey, lsp.SymbolKindMethod, lsp.SymbolKindModule, lsp.SymbolKindNamespace, lsp.SymbolKindNull, lsp.SymbolKindNumber, lsp.SymbolKindObject, lsp.SymbolKindOperator, lsp.SymbolKindPackage, lsp.SymbolKindProperty, lsp.SymbolKindString, lsp.SymbolKindStruct, lsp.SymbolKindTypeParameter, lsp.SymbolKindVariable},
			func(it lsp.SymbolKind) lsp.DocumentSymbol {
				return lsp.DocumentSymbol{
					Name:           it.String(),
					Detail:         ptr(lsp.String(fmt.Sprintf("**TODO:** documentSymbols for `%v`", src_file_path))),
					Kind:           it,
					Range:          lsp.Range{Start: lsp.Position{Line: 2, Character: 1}, End: lsp.Position{Line: 2, Character: 8}},
					SelectionRange: lsp.Range{Start: lsp.Position{Line: 2, Character: 3}, End: lsp.Position{Line: 2, Character: 6}},
				}
			}), nil
	}

	Server.On_workspace_symbol = func(params *lsp.WorkspaceSymbolParams) (any, error) {
		return []lsp.WorkspaceSymbol{{
			BaseSymbolInformation: lsp.BaseSymbolInformation{Name: "Atmo", Kind: lsp.SymbolKindInterface, ContainerName: ptr(lsp.String("Container"))},
			Location: &lsp.LocationOrUriDocumentUri{Location: &lsp.Location{
				Uri:   lsp.String("file://" + "/home/_/c/at/foo.at"),
				Range: lsp.Range{Start: lsp.Position{Line: 2, Character: 1}, End: lsp.Position{Line: 2, Character: 8}}}},
		}}, nil
	}

	Server.On_textDocument_definition = func(params *lsp.DefinitionParams) (any, error) {
		return dummyLocs(srcFilePath(params.TextDocument.Uri)), nil
	}

	Server.On_textDocument_declaration = func(params *lsp.DeclarationParams) (any, error) {
		return dummyLocs(srcFilePath(params.TextDocument.Uri)), nil
	}

	Server.On_textDocument_typeDefinition = func(params *lsp.TypeDefinitionParams) (any, error) {
		return dummyLocs(srcFilePath(params.TextDocument.Uri)), nil
	}

	Server.On_textDocument_implementation = func(params *lsp.ImplementationParams) (any, error) {
		return dummyLocs(srcFilePath(params.TextDocument.Uri)), nil
	}

	Server.On_textDocument_references = func(params *lsp.ReferenceParams) (any, error) {
		return dummyLocs(srcFilePath(params.TextDocument.Uri)), nil
	}

	Server.On_textDocument_documentHighlight = func(params *lsp.DocumentHighlightParams) (any, error) {
		return sl.As(dummyLocs(srcFilePath(params.TextDocument.Uri)), func(it lsp.Location) lsp.DocumentHighlight {
			return lsp.DocumentHighlight{Range: it.Range, Kind: lsp.DocumentHighlightKindText}
		}), nil
	}

	Server.On_textDocument_completion = func(params *lsp.CompletionParams) (any, error) {
		src_file_path := srcFilePath(params.TextDocument.Uri)
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
				Documentation: &lsp.StringOrMarkupContent{MarkupContent: &lsp.MarkupContent{Kind: lsp.MarkupKindMarkdown,
					Value: str.Fmt("**TODO** _%s_ for `%s` @ %d,%d", it.String(), src_file_path, params.Position.Line, params.Position.Character)}},
				Detail: ptr(lsp.String("Detail")),
				LabelDetails: &lsp.CompletionItemLabelDetails{
					Detail:      ptr(lsp.String(" · LD_Detail " + it.String())),
					Description: ptr(lsp.String("LD_Description " + it.String())),
				},
			}
		}), nil
	}

	Server.On_textDocument_hover = func(params *lsp.HoverParams) (any, error) {
		src_file_path := srcFilePath(params.TextDocument.Uri)
		str := lsp.String(str.Fmt("**TODO** _Hover_ for `%s` @ %d,%d", src_file_path, params.Position.Line, params.Position.Character))
		return lsp.Hover{
			Contents: &lsp.MarkupContentOrMarkedStringOrMarkedStrings{MarkedString: &lsp.StringOrLanguageStringWithValueString{String: &str}},
		}, nil
	}

	Server.On_textDocument_prepareRename = func(params *lsp.PrepareRenameParams) (any, error) {
		// src_file_path := srcFilePath(params.TextDocument.Uri)
		return lsp.Range{Start: params.Position, End: lsp.Position{Line: params.Position.Line, Character: 4 + params.Position.Character}}, nil
	}

	Server.On_textDocument_rename = func(params *lsp.RenameParams) (any, error) {
		src_file_path := srcFilePath(params.TextDocument.Uri)
		return lsp.WorkspaceEdit{
			Changes: map[lsp.String][]lsp.TextEdit{
				lsp.String(srcFileUri(src_file_path)): {{
					NewText: params.NewName,
					Range:   lsp.Range{Start: params.Position, End: lsp.Position{Line: params.Position.Line, Character: 4 + params.Position.Character}},
				}},
			},
		}, nil
	}

	Server.On_textDocument_signatureHelp = func(params *lsp.SignatureHelpParams) (any, error) {
		src_file_path := srcFilePath(params.TextDocument.Uri)
		return lsp.SignatureHelp{
			Signatures: util.If(params.Position.Line > 0,
				nil,
				[]lsp.SignatureInformation{{
					Label: "(foo bar: #baz)",
					Documentation: &lsp.StringOrMarkupContent{MarkupContent: &lsp.MarkupContent{
						Kind:  lsp.MarkupKindMarkdown,
						Value: str.Fmt("**TODO**: sig help for `%s` @ %d,%d", src_file_path, params.Position.Line, params.Position.Character)}},
				}}),
		}, nil
	}

	Server.On_textDocument_codeAction = func(params *lsp.CodeActionParams) (any, error) {
		if ClientIsAtmoVscExt || (params.Range.Start == params.Range.End) {
			return nil, nil
		}
		return []lsp.Command{{Title: "Eval", Command: "eval", Arguments: []any{params}}}, nil
	}

}

func dummyLocs(srcFilePath string) []lsp.Location {
	return []lsp.Location{
		{Uri: lsp.String(srcFileUri(srcFilePath)), Range: lsp.Range{Start: lsp.Position{Line: 2, Character: 1}, End: lsp.Position{Line: 2, Character: 8}}},
		{Uri: lsp.String(srcFileUri(srcFilePath)), Range: lsp.Range{Start: lsp.Position{Line: 4, Character: 1}, End: lsp.Position{Line: 4, Character: 8}}},
	}
}
