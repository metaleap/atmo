package atmo_lsp

import (
	"fmt"

	"github.com/metaleap/atmo/util/sl"

	lsp "github.com/metaleap/polyglot-lsp/lang_go/lsp_v3.17"
)

func srcFilePath(it lsp.TextDocumentIdentifier) string {
	return string(it.Uri)
}

func init() {
	Server.On_textDocument_documentSymbol = func(params *lsp.DocumentSymbolParams) (any, error) {
		source_file_path := srcFilePath(params.TextDocument)
		ret := sl.As([]lsp.SymbolKind{lsp.SymbolKindArray, lsp.SymbolKindBoolean, lsp.SymbolKindClass, lsp.SymbolKindConstant, lsp.SymbolKindConstructor, lsp.SymbolKindEnum, lsp.SymbolKindEnumMember, lsp.SymbolKindEvent, lsp.SymbolKindField, lsp.SymbolKindFile, lsp.SymbolKindFunction, lsp.SymbolKindInterface, lsp.SymbolKindKey, lsp.SymbolKindMethod, lsp.SymbolKindModule, lsp.SymbolKindNamespace, lsp.SymbolKindNull, lsp.SymbolKindNumber, lsp.SymbolKindObject, lsp.SymbolKindOperator, lsp.SymbolKindPackage, lsp.SymbolKindProperty, lsp.SymbolKindString, lsp.SymbolKindStruct, lsp.SymbolKindTypeParameter, lsp.SymbolKindVariable},
			func(it lsp.SymbolKind) lsp.DocumentSymbol {
				return lsp.DocumentSymbol{
					Name:           it.String(),
					Detail:         ptr(lsp.String(fmt.Sprintf("**TODO:** documentSymbols for `%v`", source_file_path))),
					Kind:           it,
					Range:          lsp.Range{Start: lsp.Position{Line: 2, Character: 1}, End: lsp.Position{Line: 2, Character: 8}},
					SelectionRange: lsp.Range{Start: lsp.Position{Line: 2, Character: 3}, End: lsp.Position{Line: 2, Character: 6}},
				}
			})
		return ret, nil
	}

	Server.On_textDocument_hover = func(params *lsp.HoverParams) (any, error) {
		str := lsp.String("**Test** _Hover_")
		return &lsp.Hover{
			Contents: &lsp.MarkupContentOrMarkedStringOrMarkedStrings{MarkedString: &lsp.StringOrLanguageStringWithValueString{String: &str}},
		}, nil
	}

}
