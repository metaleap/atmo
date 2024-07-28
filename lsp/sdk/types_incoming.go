package lsp

type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

type TextDocumentIdentifier struct {
	Uri string `json:"uri,omitempty"`
}

type TextDocumentPositionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

type HoverParams struct {
	TextDocumentPositionParams
}

type FileEvent struct {
	Uri  string         `json:"uri,omitempty"`
	Type FileChangeType `json:"type"`
}

type FileChangeType int

const (
	FileChangeTypeCreated FileChangeType = 1
	FileChangeTypeChanged FileChangeType = 2
	FileChangeTypeDeleted FileChangeType = 3
)

type DidChangeWatchedFilesParams struct {
	Changes []FileEvent `json:"changes"`
}

type WorkspaceFolder struct {
	Uri  string `json:"uri,omitempty"`
	Name string `json:"name,omitempty"`
}

type WorkspaceFoldersChangeEvent struct {
	Added   []WorkspaceFolder `json:"added"`
	Removed []WorkspaceFolder `json:"removed"`
}

type DidChangeWorkspaceFoldersParams struct {
	Event WorkspaceFoldersChangeEvent `json:"event"`
}

type TextDocumentItem struct {
	Uri        string `json:"uri,omitempty"`
	LanguageId string `json:"languageId"`
	Version    int    `json:"version,omitempty"`
	Text       string `json:"text"`
}

type DidOpenTextDocumentParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}

type DidCloseTextDocumentParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

type VersionedTextDocumentIdentifier struct {
	TextDocumentIdentifier
	Version int `json:"version,omitempty"`
}

type TextDocumentContentChangeEvent struct {
	Range *Range `json:"range"`
	Text  string `json:"text"`
}

type DidChangeTextDocumentParams struct {
	TextDocument   VersionedTextDocumentIdentifier  `json:"textDocument"`
	ContentChanges []TextDocumentContentChangeEvent `json:"contentChanges"`
}

type DocumentSymbolParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

type WorkspaceSymbolParams struct {
	Query string `json:"query"`
}

type DefinitionParams struct {
	TextDocumentPositionParams
}

type DeclarationParams struct {
	TextDocumentPositionParams
}

type TypeDefinitionParams struct {
	TextDocumentPositionParams
}

type ImplementationParams struct {
	TextDocumentPositionParams
}

type ReferenceParams struct {
	TextDocumentPositionParams
}

type DocumentHighlightParams struct {
	TextDocumentPositionParams
}

type CompletionParams struct {
	TextDocumentPositionParams
}

type SignatureHelpParams struct {
	TextDocumentPositionParams
}

type PrepareRenameParams struct {
	TextDocumentPositionParams
}

type RenameParams struct {
	TextDocumentPositionParams
	NewName string `json:"newName"`
}

type CodeActionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Range        Range                  `json:"range"`
}

type ExecuteCommandParams struct {
	Command   string `json:"command"`
	Arguments []any  `json:"arguments"`
}

type InitializeParams struct {
	WorkspaceFolders []WorkspaceFolder `json:"workspaceFolders"`
	ClientInfo       struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	}
}

type InitializedParams struct {
}

type DidSaveTextDocumentParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Text         *string                `json:"text,omitempty"`
}

type SelectionRangeParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Positions    []Position             `json:"positions"`
}

type DocumentFormattingParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

type DocumentRangeFormattingParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Range        Range                  `json:"range"`
}

type MessageActionItem struct {
	Title string `json:"title"`
}
