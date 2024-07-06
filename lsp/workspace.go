package atmo_lsp

import (
	"atmo/util/str"

	lsp "github.com/metaleap/polyglot-lsp/lang_go/lsp_v3.17"
)

func init() {
	Server.On_initialized = func(params *lsp.InitializedParams) (any, error) {
		Server.Request_workspace_workspaceFolders(lsp.Void{}, func(workspaceFolders *[]lsp.WorkspaceFolder) {
			workspace_folders := dePtr(workspaceFolders)
			println("Request_workspace_workspaceFolders:", len(workspace_folders))
		})
		return nil, nil
	}

	Server.On_workspace_didChangeWatchedFiles = func(params *lsp.DidChangeWatchedFilesParams) (any, error) {
		return nil, nil
	}

	Server.On_workspace_didChangeWorkspaceFolders = func(params *lsp.DidChangeWorkspaceFoldersParams) (any, error) {
		return nil, nil
	}

	Server.On_textDocument_didChange = func(params *lsp.DidChangeTextDocumentParams) (any, error) {
		return nil, nil
	}

	Server.On_textDocument_didClose = func(params *lsp.DidCloseTextDocumentParams) (any, error) {
		src_file_path := srcFilePath(params.TextDocument.Uri)
		if IsSrcFilePath(src_file_path) {
			Server.Notify_textDocument_publishDiagnostics(lsp.PublishDiagnosticsParams{
				Uri:         lsp.String(srcFileUri(src_file_path)),
				Diagnostics: []lsp.Diagnostic{},
			})
		}
		return nil, nil
	}

	Server.On_textDocument_didOpen = func(params *lsp.DidOpenTextDocumentParams) (any, error) {
		src_file_path := srcFilePath(params.TextDocument.Uri)
		if IsSrcFilePath(src_file_path) {
			Server.Notify_textDocument_publishDiagnostics(lsp.PublishDiagnosticsParams{
				Uri: lsp.String(srcFileUri(src_file_path)),
				Diagnostics: []lsp.Diagnostic{{
					Range:           lsp.Range{Start: lsp.Position{Line: 0, Character: 0}, End: lsp.Position{Line: 0, Character: 4}},
					Severity:        lsp.DiagnosticSeverityHint,
					Code:            &lsp.IntegerOrString{String: ptr(lsp.String("0001"))},
					CodeDescription: &lsp.CodeDescription{Href: "http://github.com/metaleap/atmo?0001"},
					Source:          ptr(lsp.String("atmo")),
					Message:         str.Fmt("**HINT:** real diags for `%s`..", src_file_path),
				}, {
					Range:           lsp.Range{Start: lsp.Position{Line: 2, Character: 0}, End: lsp.Position{Line: 2, Character: 4}},
					Severity:        lsp.DiagnosticSeverityInformation,
					Code:            &lsp.IntegerOrString{String: ptr(lsp.String("0002"))},
					CodeDescription: &lsp.CodeDescription{Href: "http://github.com/metaleap/atmo?0002"},
					Source:          ptr(lsp.String("atmo")),
					Message:         str.Fmt("**INFO:** real diags for `%s`..", src_file_path),
				}},
			})
		}
		return nil, nil
	}
}
