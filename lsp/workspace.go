package lsp

import (
	"io/fs"

	lsp "atmo/lsp/sdk"
	"atmo/session"
	"atmo/util"
	"atmo/util/sl"
	"atmo/util/str"
)

func init() {
	Server.On_initialized = func(params *lsp.InitializedParams) (any, error) {
		Server.Request_workspace_workspaceFolders(lsp.Void{}, func(workspaceFolders []lsp.WorkspaceFolder) {
			onWorkspaceFoldersChanged(nil, workspaceFolders)
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
		if src_file_path := lsp.UriToFsPath(params.TextDocument.Uri); session.IsSrcFilePath(src_file_path) {
			println(str.Fmt(">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>%#v<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<", params.ContentChanges[0]))

			/*&lsp_v3_17.RangeWithRangeLengthUintegerWithTextStringOrTextString{
			RangeWithRangeLengthUintegerWithTextString:(*struct { Range lsp_v3_17.Range "json:\"range,omitempty\""; RangeLength *lsp_v3_17.Uinteger "json:\"rangeLength,omitempty\""; Text string "json:\"text,omitempty\"" })(nil),
			TextString:(*struct { Text string "json:\"text,omitempty\"" })(nil)}
			*/
			//{"params":{"textDocument":{"uri":"file:///home/_/c/at/foo.at","version":2},"contentChanges":[{"text":"foo-bar-{baz} :=\n  (\"Hello World\"[0])\n"}]}}

			session.OnSrcFileEdit(src_file_path, params.ContentChanges[0].Text)
		}
		return nil, nil
	}

	Server.On_textDocument_didSave = func(params *lsp.DidSaveTextDocumentParams) (any, error) {
		if src_file_path := lsp.UriToFsPath(params.TextDocument.Uri); session.IsSrcFilePath(src_file_path) {
			session.OnSrcFileEvents(nil, true, src_file_path)
		}
		return nil, nil
	}

	Server.On_textDocument_didClose = func(params *lsp.DidCloseTextDocumentParams) (any, error) {
		if src_file_path := lsp.UriToFsPath(params.TextDocument.Uri); session.IsSrcFilePath(src_file_path) {
			session.OnSrcFileEvents(nil, true, src_file_path)
		}
		return nil, nil
	}

	Server.On_textDocument_didOpen = func(params *lsp.DidOpenTextDocumentParams) (any, error) {
		if src_file_path := lsp.UriToFsPath(params.TextDocument.Uri); session.IsSrcFilePath(src_file_path) {
			session.OnSrcFileEvents(nil, true, src_file_path)
		}
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
			removed = append(removed, all_src_file_paths(lsp.UriToFsPath(it.Uri))...)
		case lsp.FileChangeTypeCreated:
			added = append(added, all_src_file_paths(lsp.UriToFsPath(it.Uri))...)
		case lsp.FileChangeTypeChanged:
			changed = append(changed, all_src_file_paths(lsp.UriToFsPath(it.Uri))...)
		}
	}
	session.OnSrcFileEvents(removed, false, append(added, changed...)...)
}

func onWorkspaceFoldersChanged(rootFoldersRemoved []lsp.WorkspaceFolder, rootFoldersAdded []lsp.WorkspaceFolder) {
	onWorkspaceDidChangeWatchedFiles(append(
		sl.As(rootFoldersRemoved, func(it lsp.WorkspaceFolder) lsp.FileEvent {
			return lsp.FileEvent{Type: lsp.FileChangeTypeDeleted, Uri: lsp.UriToFsPath(it.Uri)}
		}),
		sl.As(rootFoldersAdded, func(it lsp.WorkspaceFolder) lsp.FileEvent {
			return lsp.FileEvent{Type: lsp.FileChangeTypeCreated, Uri: lsp.UriToFsPath(it.Uri)}
		})...))
}
