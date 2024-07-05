package atmo_lsp

import (
	lsp "github.com/metaleap/polyglot-lsp/lang_go/lsp_v3.17"
)

func init() {
	Server.On_textDocument_hover = func(params *lsp.HoverParams) (any, error) {
		str := lsp.String("**Test** _Hover_")
		return &lsp.Hover{
			Contents: &lsp.MarkupContentOrMarkedStringOrMarkedStrings{MarkedString: &lsp.StringOrLanguageStringWithValueString{String: &str}},
		}, nil
	}
}
