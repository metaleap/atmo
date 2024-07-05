package atmo_lsp

import (
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
		return nil, nil
	}

	Server.On_textDocument_didOpen = func(params *lsp.DidOpenTextDocumentParams) (any, error) {
		return nil, nil
	}
}
