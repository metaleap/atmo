package atmo_lsp

import (
	"fmt"
	"yo/util/str"

	"github.com/metaleap/atmo/util/sl"

	lsp "github.com/metaleap/polyglot-lsp/lang_go/lsp_v3.17"
)

func srcFilePath(it lsp.TextDocumentIdentifier) string {
	return str.TrimPref(string(it.Uri), "file://")
}

func srcFileUri(srcFilePath string) string {
	return "file://" + srcFilePath
}

func init() {
	Server.On_textDocument_documentSymbol = func(params *lsp.DocumentSymbolParams) (any, error) {
		src_file_path := srcFilePath(params.TextDocument)
		ret := sl.As([]lsp.SymbolKind{lsp.SymbolKindArray, lsp.SymbolKindBoolean, lsp.SymbolKindClass, lsp.SymbolKindConstant, lsp.SymbolKindConstructor, lsp.SymbolKindEnum, lsp.SymbolKindEnumMember, lsp.SymbolKindEvent, lsp.SymbolKindField, lsp.SymbolKindFile, lsp.SymbolKindFunction, lsp.SymbolKindInterface, lsp.SymbolKindKey, lsp.SymbolKindMethod, lsp.SymbolKindModule, lsp.SymbolKindNamespace, lsp.SymbolKindNull, lsp.SymbolKindNumber, lsp.SymbolKindObject, lsp.SymbolKindOperator, lsp.SymbolKindPackage, lsp.SymbolKindProperty, lsp.SymbolKindString, lsp.SymbolKindStruct, lsp.SymbolKindTypeParameter, lsp.SymbolKindVariable},
			func(it lsp.SymbolKind) lsp.DocumentSymbol {
				return lsp.DocumentSymbol{
					Name:           it.String(),
					Detail:         ptr(lsp.String(fmt.Sprintf("**TODO:** documentSymbols for `%v`", src_file_path))),
					Kind:           it,
					Range:          lsp.Range{Start: lsp.Position{Line: 2, Character: 1}, End: lsp.Position{Line: 2, Character: 8}},
					SelectionRange: lsp.Range{Start: lsp.Position{Line: 2, Character: 3}, End: lsp.Position{Line: 2, Character: 6}},
				}
			})
		return ret, nil
	}

	Server.On_workspace_symbol = func(params *lsp.WorkspaceSymbolParams) (any, error) {
		return []lsp.WorkspaceSymbol{{
			BaseSymbolInformation: lsp.BaseSymbolInformation{Name: "Atmo", Kind: lsp.SymbolKindInterface, ContainerName: ptr(lsp.String("Container"))},
			Location: &lsp.LocationOrUriDocumentUri{Location: &lsp.Location{
				Uri:   lsp.String("file://" + "/home/_/c/at/foo.at"),
				Range: lsp.Range{Start: lsp.Position{Line: 2, Character: 1}, End: lsp.Position{Line: 2, Character: 8}}}},
		}}, nil
	}

	Server.On_textDocument_hover = func(params *lsp.HoverParams) (any, error) {
		str := lsp.String("**Test** _Hover_")
		return &lsp.Hover{
			Contents: &lsp.MarkupContentOrMarkedStringOrMarkedStrings{MarkedString: &lsp.StringOrLanguageStringWithValueString{String: &str}},
		}, nil
	}

	Server.On_textDocument_definition = func(params *lsp.DefinitionParams) (any, error) {
		return dummyLocs(srcFilePath(params.TextDocument)), nil
	}

	Server.On_textDocument_declaration = func(params *lsp.DeclarationParams) (any, error) {
		return dummyLocs(srcFilePath(params.TextDocument)), nil
	}

	Server.On_textDocument_typeDefinition = func(params *lsp.TypeDefinitionParams) (any, error) {
		return dummyLocs(srcFilePath(params.TextDocument)), nil
	}

	Server.On_textDocument_implementation = func(params *lsp.ImplementationParams) (any, error) {
		return dummyLocs(srcFilePath(params.TextDocument)), nil
	}

	Server.On_textDocument_references = func(params *lsp.ReferenceParams) (any, error) {
		return dummyLocs(srcFilePath(params.TextDocument)), nil
	}

	Server.On_textDocument_documentHighlight = func(params *lsp.DocumentHighlightParams) (any, error) {
		return sl.As(dummyLocs(srcFilePath(params.TextDocument)), func(it lsp.Location) lsp.DocumentHighlight {
			return lsp.DocumentHighlight{Range: it.Range, Kind: lsp.DocumentHighlightKindText}
		}), nil
	}

}

func dummyLocs(srcFilePath string) []lsp.Location {
	return []lsp.Location{
		{Uri: lsp.String(srcFileUri(srcFilePath)), Range: lsp.Range{Start: lsp.Position{Line: 2, Character: 1}, End: lsp.Position{Line: 2, Character: 8}}},
		{Uri: lsp.String(srcFileUri(srcFilePath)), Range: lsp.Range{Start: lsp.Position{Line: 4, Character: 1}, End: lsp.Position{Line: 4, Character: 8}}},
	}
}
