package lsp

import (
	"io/fs"

	"atmo/session"
	"atmo/util"
	"atmo/util/sl"
	"atmo/util/str"

	lsp "github.com/metaleap/polyglot-lsp/lang_go/lsp_v3.17"
)

func init() {
	Server.On_initialized = func(params *lsp.InitializedParams) (any, error) {
		Server.Request_workspace_workspaceFolders(lsp.Void{}, func(workspaceFolders *[]lsp.WorkspaceFolder) {
			onWorkspaceFoldersChanged(nil, dePtr(workspaceFolders))
		})
		return nil, nil
	}

	Server.On_workspace_didChangeWorkspaceFolders = func(params *lsp.DidChangeWorkspaceFoldersParams) (any, error) {
		onWorkspaceFoldersChanged(params.Event.Removed, params.Event.Added)
		return nil, nil
	}

	Server.On_workspace_didChangeWatchedFiles = func(params *lsp.DidChangeWatchedFilesParams) (any, error) {
		onWorkspaceDidChangeWatchedFiles(params.Changes)
		return nil, nil
	}

	Server.On_textDocument_didChange = func(params *lsp.DidChangeTextDocumentParams) (any, error) {
		if src_file_path := toFsPath(params.TextDocument.Uri); session.IsSrcFilePath(src_file_path) {
			session.OnSrcFileEdit(src_file_path, params.ContentChanges[0].TextString.Text)
		}
		return nil, nil
	}

	Server.On_textDocument_didClose = func(params *lsp.DidCloseTextDocumentParams) (any, error) {
		src_file_path := toFsPath(params.TextDocument.Uri)
		if !session.IsSrcFilePath(src_file_path) {
			return nil, nil
		}
		Server.Notify_textDocument_publishDiagnostics(lsp.PublishDiagnosticsParams{
			Uri:         lsp.String(toUri(src_file_path)),
			Diagnostics: []lsp.Diagnostic{},
		})
		return nil, nil
	}

	Server.On_textDocument_didOpen = func(params *lsp.DidOpenTextDocumentParams) (any, error) {
		src_file_path := toFsPath(params.TextDocument.Uri)
		if !session.IsSrcFilePath(src_file_path) {
			return nil, nil
		}
		Server.Notify_textDocument_publishDiagnostics(lsp.PublishDiagnosticsParams{
			Uri: lsp.String(toUri(src_file_path)),
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
		return nil, nil
	}
}

func onWorkspaceDidChangeWatchedFiles(fileEvents []lsp.FileEvent) {
	all_src_file_paths := func(fsPath string) (ret []string) {
		if util.FsIsDir(fsPath) {
			util.FsDirWalk(fsPath, func(fsPath string, fsEntry fs.DirEntry) {
				if (!fsEntry.IsDir()) && session.IsSrcFilePath(fsPath) {
					ret = append(ret, fsPath)
				}
			})
		} else if session.IsSrcFilePath(fsPath) {
			ret = append(ret, fsPath)
		}
		return
	}

	var removed, added, changed []string
	for _, it := range fileEvents {
		switch it.Type {
		case lsp.FileChangeTypeDeleted:
			removed = append(removed, all_src_file_paths(toFsPath(it.Uri))...)
		case lsp.FileChangeTypeCreated:
			added = append(added, all_src_file_paths(toFsPath(it.Uri))...)
		case lsp.FileChangeTypeChanged:
			changed = append(changed, all_src_file_paths(toFsPath(it.Uri))...)
		}
	}
	session.OnSrcFileEvents(removed, added, changed)
}

func onWorkspaceFoldersChanged(rootFoldersRemoved []lsp.WorkspaceFolder, rootFoldersAdded []lsp.WorkspaceFolder) {
	onWorkspaceDidChangeWatchedFiles(append(
		sl.As(rootFoldersRemoved, func(it lsp.WorkspaceFolder) lsp.FileEvent {
			return lsp.FileEvent{Type: lsp.FileChangeTypeDeleted, Uri: lsp.String(toFsPath(it.Uri))}
		}),
		sl.As(rootFoldersAdded, func(it lsp.WorkspaceFolder) lsp.FileEvent {
			return lsp.FileEvent{Type: lsp.FileChangeTypeCreated, Uri: lsp.String(toFsPath(it.Uri))}
		})...))
}
