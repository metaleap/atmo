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
		src_file_path := lsp.UriToFsPath(params.TextDocument.Uri)
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
				Uri:   lsp.FsPathToUri("/home/_/c/at/foo.at"),
				Range: lsp.Range{Start: lsp.Position{Line: 2, Character: 1}, End: lsp.Position{Line: 2, Character: 8}},
			},
		}}, nil
	}

	Server.On_textDocument_definition = func(params *lsp.DefinitionParams) (any, error) {
		return dummyLocs(lsp.UriToFsPath(params.TextDocument.Uri)), nil
	}

	Server.On_textDocument_declaration = func(params *lsp.DeclarationParams) (any, error) {
		return dummyLocs(lsp.UriToFsPath(params.TextDocument.Uri)), nil
	}

	Server.On_textDocument_typeDefinition = func(params *lsp.TypeDefinitionParams) (any, error) {
		return dummyLocs(lsp.UriToFsPath(params.TextDocument.Uri)), nil
	}

	Server.On_textDocument_implementation = func(params *lsp.ImplementationParams) (any, error) {
		return dummyLocs(lsp.UriToFsPath(params.TextDocument.Uri)), nil
	}

	Server.On_textDocument_references = func(params *lsp.ReferenceParams) (any, error) {
		return dummyLocs(lsp.UriToFsPath(params.TextDocument.Uri)), nil
	}

	Server.On_textDocument_documentHighlight = func(params *lsp.DocumentHighlightParams) (any, error) {
		return sl.As(dummyLocs(lsp.UriToFsPath(params.TextDocument.Uri)), func(it lsp.Location) lsp.DocumentHighlight {
			return lsp.DocumentHighlight{Range: it.Range, Kind: lsp.DocumentHighlightKindText}
		}), nil
	}

	Server.On_textDocument_completion = func(params *lsp.CompletionParams) (any, error) {
		src_file_path := lsp.UriToFsPath(params.TextDocument.Uri)
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
		src_file_path := lsp.UriToFsPath(params.TextDocument.Uri)
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
		src_file_path := lsp.UriToFsPath(params.TextDocument.Uri)
		return lsp.WorkspaceEdit{
			Changes: map[string][]lsp.TextEdit{
				lsp.FsPathToUri(src_file_path): {{
					NewText: params.NewName,
					Range:   lsp.Range{Start: params.Position, End: lsp.Position{Line: params.Position.Line, Character: 4 + params.Position.Character}},
				}},
			},
		}, nil
	}

	Server.On_textDocument_signatureHelp = func(params *lsp.SignatureHelpParams) (any, error) {
		src_file_path := lsp.UriToFsPath(params.TextDocument.Uri)
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

	session.OnNoticesChanged = func(pub map[string][]*session.SrcFileNotice) {
		util.Assert(Server.Initialized.Client != nil && Server.Initialized.Server != nil, nil)
		for file_path, diags := range pub {
			Server.Notify_textDocument_publishDiagnostics(lsp.PublishDiagnosticsParams{
				Uri: lsp.FsPathToUri(file_path),
				Diagnostics: sl.As(diags, func(it *session.SrcFileNotice) lsp.Diagnostic {
					return lsp.Diagnostic{
						Code:            string(it.Code),
						CodeDescription: &lsp.CodeDescription{Href: "https://github.com/metaleap/atom/docs/errs.md#" + string(it.Code)},
						Range:           toLspRange(it.Span),
						Message:         it.Message,
						Severity:        toLspDiagSeverity(it.Kind),
					}
				}),
			})
		}
	}

}

func dummyLocs(srcFilePath string) []lsp.Location {
	return []lsp.Location{
		{Uri: lsp.FsPathToUri(srcFilePath), Range: lsp.Range{Start: lsp.Position{Line: 2, Character: 1}, End: lsp.Position{Line: 2, Character: 8}}},
		{Uri: lsp.FsPathToUri(srcFilePath), Range: lsp.Range{Start: lsp.Position{Line: 4, Character: 1}, End: lsp.Position{Line: 4, Character: 8}}},
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
