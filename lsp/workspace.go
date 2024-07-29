package lsp

import (
	"errors"
	"io/fs"

	lsp "atmo/lsp/sdk"
	"atmo/session"
	"atmo/util"
	"atmo/util/sl"
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
		src_file_path := lspUriToFsPath(params.TextDocument.Uri)
		if session.IsSrcFilePath(src_file_path) && len(params.ContentChanges) == 1 {
			session.Access(func(sess session.StateAccess, _ session.Intel) {
				sess.OnSrcFileEdit(src_file_path, params.ContentChanges[0].Text)
			})
		} else if len(params.ContentChanges) > 1 {
			return nil, errors.New("'textDocument/didChange' notifications based on `TextDocumentSyncKind.Incremental` not supported")
		}
		return nil, nil
	}

	Server.On_textDocument_didSave = func(params *lsp.DidSaveTextDocumentParams) (any, error) {
		if src_file_path := lspUriToFsPath(params.TextDocument.Uri); session.IsSrcFilePath(src_file_path) {
			session.Access(func(sess session.StateAccess, _ session.Intel) {
				sess.OnSrcFileEvents(nil, false, src_file_path)
			})
		}
		return nil, nil
	}

	Server.On_textDocument_didClose = func(params *lsp.DidCloseTextDocumentParams) (any, error) {
		if src_file_path := lspUriToFsPath(params.TextDocument.Uri); session.IsSrcFilePath(src_file_path) {
			session.Access(func(sess session.StateAccess, _ session.Intel) {
				sess.OnSrcFileEvents(nil, true, src_file_path)
			})
		}
		return nil, nil
	}

	Server.On_textDocument_didOpen = func(params *lsp.DidOpenTextDocumentParams) (any, error) {
		if src_file_path := lspUriToFsPath(params.TextDocument.Uri); session.IsSrcFilePath(src_file_path) {
			session.Access(func(sess session.StateAccess, _ session.Intel) {
				sess.OnSrcFileEvents(nil, true, src_file_path)
			})
		}
		return nil, nil
	}
}

func onWorkspaceDidChangeWatchedFiles(fileEvents []lsp.FileEvent) {
	session.Access(func(sess session.StateAccess, _ session.Intel) {
		all_src_file_paths := func(fsPath string) (ret []string) {
			if session.IsSrcFilePath(fsPath) {
				ret = append(ret, fsPath)
			} else if util.FsIsDir(fsPath) {
				util.FsDirWalk(fsPath, func(fsPath string, fsEntry fs.DirEntry) {
					if session.IsSrcFilePath(fsPath) {
						ret = append(ret, fsPath)
					}
				})
			} else if pkg := sess.GetSrcPack(fsPath, false); pkg != nil {
				for _, src_file := range pkg.Files {
					ret = append(ret, src_file.FilePath)
				}
			}
			return
		}

		var removed, added, changed []string
		for _, it := range fileEvents {
			switch path := lspUriToFsPath(it.Uri); it.Type {
			case lsp.FileChangeTypeDeleted:
				removed = append(removed, all_src_file_paths(path)...)
			case lsp.FileChangeTypeCreated:
				added = append(added, all_src_file_paths(path)...)
			case lsp.FileChangeTypeChanged:
				changed = append(changed, all_src_file_paths(path)...)
			}
		}
		sess.OnSrcFileEvents(removed, false, append(added, changed...)...)
	})
}

func onWorkspaceFoldersChanged(rootFoldersRemoved []lsp.WorkspaceFolder, rootFoldersAdded []lsp.WorkspaceFolder) {
	onWorkspaceDidChangeWatchedFiles(append(
		sl.To(rootFoldersRemoved, func(it lsp.WorkspaceFolder) lsp.FileEvent {
			return lsp.FileEvent{Type: lsp.FileChangeTypeDeleted, Uri: lspUriToFsPath(it.Uri)}
		}),
		sl.To(rootFoldersAdded, func(it lsp.WorkspaceFolder) lsp.FileEvent {
			return lsp.FileEvent{Type: lsp.FileChangeTypeCreated, Uri: lspUriToFsPath(it.Uri)}
		})...))
}
