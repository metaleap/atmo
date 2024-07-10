package lsp

import (
	"atmo/session"
	"atmo/util/str"
)

func FsPathToLspUri(fsPath string) string { return "file://" + fsPath }
func PosToLspPos(pos *session.SrcFilePos) Position {
	return Position{Line: pos.Line - 1, Character: pos.Char - 1}
}
func SpanToLspRange(span session.SrcFileSpan) Range {
	return Range{Start: PosToLspPos(&span.Start), End: PosToLspPos(&span.End)}
}

type Registration struct {
	Id              string `json:"id,omitempty"`
	Method          string `json:"method,omitempty"`
	RegisterOptions any    `json:"registerOptions,omitempty"`
}

type RegistrationParams struct {
	Registrations []Registration `json:"registrations,omitempty"`
}

type FileSystemWatcher struct {
	Kind        WatchKind `json:"kind"`
	GlobPattern struct {
		Pattern string `json:"pattern"`
	} `json:"globPattern"`
}

type WatchKind int

const (
	WatchKindCreate WatchKind = 1
	WatchKindChange WatchKind = 2
	WatchKindDelete WatchKind = 4
	WatchKindAll              = WatchKindCreate | WatchKindChange | WatchKindDelete
)

type DidChangeWatchedFilesRegistrationOptions struct {
	Watchers []FileSystemWatcher `json:"watchers"`
}

type MarkupContent struct {
	Kind  MarkupKind `json:"kind"`
	Value string     `json:"value"`
}

type MarkupKind string

const (
	MarkupKindPlaintext MarkupKind = "plaintext"
	MarkupKindMarkdown  MarkupKind = "markdown"
)

type Hover struct {
	Contents MarkupContent `json:"contents"`
	Range    *Range        `json:"range,omitempty"`
}

type DocumentSymbol struct {
	Name           string           `json:"name"`
	Detail         string           `json:"detail,omitempty"`
	Kind           SymbolKind       `json:"kind,omitempty"`
	Tags           []SymbolTag      `json:"tags,omitempty"`
	Range          Range            `json:"range"`
	SelectionRange Range            `json:"selectionRange"`
	Children       []DocumentSymbol `json:"children,omitempty"`
}

// A symbol kind.
type SymbolKind uint

const (
	SymbolKindFile          SymbolKind = 1
	SymbolKindModule        SymbolKind = 2
	SymbolKindNamespace     SymbolKind = 3
	SymbolKindPackage       SymbolKind = 4
	SymbolKindClass         SymbolKind = 5
	SymbolKindMethod        SymbolKind = 6
	SymbolKindProperty      SymbolKind = 7
	SymbolKindField         SymbolKind = 8
	SymbolKindConstructor   SymbolKind = 9
	SymbolKindEnum          SymbolKind = 10
	SymbolKindInterface     SymbolKind = 11
	SymbolKindFunction      SymbolKind = 12
	SymbolKindVariable      SymbolKind = 13
	SymbolKindConstant      SymbolKind = 14
	SymbolKindString        SymbolKind = 15
	SymbolKindNumber        SymbolKind = 16
	SymbolKindBoolean       SymbolKind = 17
	SymbolKindArray         SymbolKind = 18
	SymbolKindObject        SymbolKind = 19
	SymbolKindKey           SymbolKind = 20
	SymbolKindNull          SymbolKind = 21
	SymbolKindEnumMember    SymbolKind = 22
	SymbolKindStruct        SymbolKind = 23
	SymbolKindEvent         SymbolKind = 24
	SymbolKindOperator      SymbolKind = 25
	SymbolKindTypeParameter SymbolKind = 26
)

func (me SymbolKind) String() string {
	switch me {
	case SymbolKindFile:
		return "File"
	case SymbolKindModule:
		return "Module"
	case SymbolKindNamespace:
		return "Namespace"
	case SymbolKindPackage:
		return "Package"
	case SymbolKindClass:
		return "Class"
	case SymbolKindMethod:
		return "Method"
	case SymbolKindProperty:
		return "Property"
	case SymbolKindField:
		return "Field"
	case SymbolKindConstructor:
		return "Constructor"
	case SymbolKindEnum:
		return "Enum"
	case SymbolKindInterface:
		return "Interface"
	case SymbolKindFunction:
		return "Function"
	case SymbolKindVariable:
		return "Variable"
	case SymbolKindConstant:
		return "Constant"
	case SymbolKindString:
		return "String"
	case SymbolKindNumber:
		return "Number"
	case SymbolKindBoolean:
		return "Boolean"
	case SymbolKindArray:
		return "Array"
	case SymbolKindObject:
		return "Object"
	case SymbolKindKey:
		return "Key"
	case SymbolKindNull:
		return "Null"
	case SymbolKindEnumMember:
		return "EnumMember"
	case SymbolKindStruct:
		return "Struct"
	case SymbolKindEvent:
		return "Event"
	case SymbolKindOperator:
		return "Operator"
	case SymbolKindTypeParameter:
		return "TypeParameter"
	}
	return str.FromInt(int(me))
}

type SymbolTag uint

const (
	SymbolTagDeprecated SymbolTag = 1
)

func (me SymbolTag) String() string {
	switch me {
	case SymbolTagDeprecated:
		return "Deprecated"
	}
	return ""
}

type Location struct {
	Uri   string `json:"uri"`
	Range Range  `json:"range"`
}

type WorkspaceSymbol struct {
	Name          string      `json:"name"`
	Kind          SymbolKind  `json:"kind,omitempty"`
	Tags          []SymbolTag `json:"tags,omitempty"`
	ContainerName string      `json:"containerName,omitempty"`
	Location      Location    `json:"location"`
}

type DocumentHighlight struct {
	Range Range                 `json:"range"`
	Kind  DocumentHighlightKind `json:"kind"`
}

type DocumentHighlightKind uint

const (
	DocumentHighlightKindText  DocumentHighlightKind = 1
	DocumentHighlightKindRead  DocumentHighlightKind = 2
	DocumentHighlightKindWrite DocumentHighlightKind = 3
)

type CompletionItemLabelDetails struct {
	Detail      string `json:"detail,omitempty"`
	Description string `json:"description,omitempty"`
}

type CompletionItem struct {
	Label         string                      `json:"label"`
	LabelDetails  *CompletionItemLabelDetails `json:"labelDetails,omitempty"`
	Kind          CompletionItemKind          `json:"kind,omitempty"`
	Tags          []CompletionItemTag         `json:"tags,omitempty"`
	Detail        string                      `json:"detail,omitempty"`
	Documentation *MarkupContent              `json:"documentation,omitempty"`
}

type CompletionItemTag uint

const (
	CompletionItemTagDeprecated CompletionItemTag = 1
)

func (me CompletionItemTag) String() string {
	switch me {
	case CompletionItemTagDeprecated:
		return "Deprecated"
	}
	return str.FromInt(int(me))
}

type CompletionItemKind uint

const (
	CompletionItemKindText          CompletionItemKind = 1
	CompletionItemKindMethod        CompletionItemKind = 2
	CompletionItemKindFunction      CompletionItemKind = 3
	CompletionItemKindConstructor   CompletionItemKind = 4
	CompletionItemKindField         CompletionItemKind = 5
	CompletionItemKindVariable      CompletionItemKind = 6
	CompletionItemKindClass         CompletionItemKind = 7
	CompletionItemKindInterface     CompletionItemKind = 8
	CompletionItemKindModule        CompletionItemKind = 9
	CompletionItemKindProperty      CompletionItemKind = 10
	CompletionItemKindUnit          CompletionItemKind = 11
	CompletionItemKindValue         CompletionItemKind = 12
	CompletionItemKindEnum          CompletionItemKind = 13
	CompletionItemKindKeyword       CompletionItemKind = 14
	CompletionItemKindSnippet       CompletionItemKind = 15
	CompletionItemKindColor         CompletionItemKind = 16
	CompletionItemKindFile          CompletionItemKind = 17
	CompletionItemKindReference     CompletionItemKind = 18
	CompletionItemKindFolder        CompletionItemKind = 19
	CompletionItemKindEnumMember    CompletionItemKind = 20
	CompletionItemKindConstant      CompletionItemKind = 21
	CompletionItemKindStruct        CompletionItemKind = 22
	CompletionItemKindEvent         CompletionItemKind = 23
	CompletionItemKindOperator      CompletionItemKind = 24
	CompletionItemKindTypeParameter CompletionItemKind = 25
)

func (me CompletionItemKind) String() string {
	switch me {
	case CompletionItemKindText:
		return "Text"
	case CompletionItemKindMethod:
		return "Method"
	case CompletionItemKindFunction:
		return "Function"
	case CompletionItemKindConstructor:
		return "Constructor"
	case CompletionItemKindField:
		return "Field"
	case CompletionItemKindVariable:
		return "Variable"
	case CompletionItemKindClass:
		return "Class"
	case CompletionItemKindInterface:
		return "Interface"
	case CompletionItemKindModule:
		return "Module"
	case CompletionItemKindProperty:
		return "Property"
	case CompletionItemKindUnit:
		return "Unit"
	case CompletionItemKindValue:
		return "Value"
	case CompletionItemKindEnum:
		return "Enum"
	case CompletionItemKindKeyword:
		return "Keyword"
	case CompletionItemKindSnippet:
		return "Snippet"
	case CompletionItemKindColor:
		return "Color"
	case CompletionItemKindFile:
		return "File"
	case CompletionItemKindReference:
		return "Reference"
	case CompletionItemKindFolder:
		return "Folder"
	case CompletionItemKindEnumMember:
		return "EnumMember"
	case CompletionItemKindConstant:
		return "Constant"
	case CompletionItemKindStruct:
		return "Struct"
	case CompletionItemKindEvent:
		return "Event"
	case CompletionItemKindOperator:
		return "Operator"
	case CompletionItemKindTypeParameter:
		return "TypeParameter"
	}
	return str.FromInt(int(me))
}

type TextEdit struct {
	Range   Range  `json:"range"`
	NewText string `json:"newText"`
}

type WorkspaceEdit struct {
	Changes map[string][]TextEdit `json:"changes"`
}

type SignatureHelp struct {
	Signatures []SignatureInformation `json:"signatures"`
}

type SignatureInformation struct {
	Label         string         `json:"label"`
	Documentation *MarkupContent `json:"documentation,omitempty"`
}

type PublishDiagnosticsParams struct {
	Uri         string       `json:"uri,omitempty"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

type Diagnostic struct {
	Range           Range              `json:"range"`
	Severity        DiagnosticSeverity `json:"severity,omitempty"`
	Code            string             `json:"code,omitempty"`
	CodeDescription *CodeDescription   `json:"codeDescription,omitempty"`
	Source          string             `json:"source,omitempty"`
	Message         string             `json:"message"`
	Tags            []DiagnosticTag    `json:"tags,omitempty"`
}

type CodeDescription struct {
	Href string `json:"href"`
}

type DiagnosticSeverity uint

const (
	DiagnosticSeverityError       DiagnosticSeverity = 1
	DiagnosticSeverityWarning     DiagnosticSeverity = 2
	DiagnosticSeverityInformation DiagnosticSeverity = 3
	DiagnosticSeverityHint        DiagnosticSeverity = 4
)

func (me DiagnosticSeverity) String() string {
	switch me {
	case DiagnosticSeverityError:
		return "Error"
	case DiagnosticSeverityWarning:
		return "Warning"
	case DiagnosticSeverityInformation:
		return "Information"
	case DiagnosticSeverityHint:
		return "Hint"
	}
	return str.FromInt(int(me))
}

type DiagnosticTag uint

const (
	DiagnosticTagUnnecessary DiagnosticTag = 1
	DiagnosticTagDeprecated  DiagnosticTag = 2
)

func (me DiagnosticTag) String() string {
	switch me {
	case DiagnosticTagUnnecessary:
		return "Unnecessary"
	case DiagnosticTagDeprecated:
		return "Deprecated"
	}
	return str.FromInt(int(me))
}

type Command struct {
	Title     string `json:"title,omitempty"`
	Command   string `json:"command"`
	Arguments []any  `json:"arguments,omitempty"`
}

type LogMessageParams struct {
	Type    MessageType `json:"type"`
	Message string      `json:"message"`
}

type ShowMessageParams struct {
	Type    MessageType `json:"type"`
	Message string      `json:"message"`
}

type MessageType uint

const (
	MessageTypeError   MessageType = 1
	MessageTypeWarning MessageType = 2
	MessageTypeInfo    MessageType = 3
	MessageTypeLog     MessageType = 4
	MessageTypeDebug   MessageType = 5
)

func (me MessageType) String() string {
	switch me {
	case MessageTypeError:
		return "Error"
	case MessageTypeWarning:
		return "Warning"
	case MessageTypeInfo:
		return "Info"
	case MessageTypeLog:
		return "Log"
	case MessageTypeDebug:
		return "Debug"
	}
	return str.FromInt(int(me))
}

type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities,omitempty"`
	ServerInfo   struct {
		Name    string `json:"name"`
		Version string `json:"version,omitempty"`
	} `json:"serverInfo"`
}

type ServerCapabilities struct {
	TextDocumentSync                *TextDocumentSyncOptions `json:"textDocumentSync,omitempty"`
	CompletionProvider              *CompletionOptions       `json:"completionProvider,omitempty"`
	HoverProvider                   bool                     `json:"hoverProvider,omitempty"`
	SignatureHelpProvider           *SignatureHelpOptions    `json:"signatureHelpProvider,omitempty"`
	DeclarationProvider             bool                     `json:"declarationProvider,omitempty"`
	DefinitionProvider              bool                     `json:"definitionProvider,omitempty"`
	TypeDefinitionProvider          bool                     `json:"typeDefinitionProvider,omitempty"`
	ImplementationProvider          bool                     `json:"implementationProvider,omitempty"`
	ReferencesProvider              bool                     `json:"referencesProvider,omitempty"`
	DocumentHighlightProvider       bool                     `json:"documentHighlightProvider,omitempty"`
	DocumentSymbolProvider          *DocumentSymbolOptions   `json:"documentSymbolProvider,omitempty"`
	CodeActionProvider              bool                     `json:"codeActionProvider,omitempty"`
	WorkspaceSymbolProvider         bool                     `json:"workspaceSymbolProvider,omitempty"`
	DocumentFormattingProvider      bool                     `json:"documentFormattingProvider,omitempty"`
	DocumentRangeFormattingProvider bool                     `json:"documentRangeFormattingProvider,omitempty"`
	RenameProvider                  *RenameOptions           `json:"renameProvider,omitempty"`
	SelectionRangeProvider          bool                     `json:"selectionRangeProvider,omitempty"`
	ExecuteCommandProvider          *ExecuteCommandOptions   `json:"executeCommandProvider,omitempty"`
	Workspace                       struct {
		WorkspaceFolders WorkspaceFoldersServerCapabilities `json:"workspaceFolders,omitempty"`
	} `json:"workspace"`
}

type CompletionOptions struct {
	TriggerCharacters []string `json:"triggerCharacters,omitempty"`
}

type SignatureHelpOptions struct {
	TriggerCharacters []string `json:"triggerCharacters,omitempty"`
}

type DocumentSymbolOptions struct {
	Label string `json:"label,omitempty"`
}

type RenameOptions struct {
	PrepareProvider bool `json:"prepareProvider,omitempty"`
}

type ExecuteCommandOptions struct {
	Commands []string `json:"commands,omitempty"`
}

type WorkspaceFoldersServerCapabilities struct {
	Supported           bool `json:"supported,omitempty"`
	ChangeNotifications bool `json:"changeNotifications,omitempty"`
}

type TextDocumentSyncOptions struct {
	OpenClose bool                 `json:"openClose,omitempty"`
	Change    TextDocumentSyncKind `json:"change,omitempty"`
	Save      *SaveOptions         `json:"save,omitempty"`
}

type SaveOptions struct {
	IncludeText bool `json:"includeText,omitempty"`
}

type TextDocumentSyncKind uint

const (
	TextDocumentSyncKindNone        TextDocumentSyncKind = 0
	TextDocumentSyncKindFull        TextDocumentSyncKind = 1
	TextDocumentSyncKindIncremental TextDocumentSyncKind = 2
)

func (me TextDocumentSyncKind) String() string {
	switch me {
	case TextDocumentSyncKindNone:
		return "None"
	case TextDocumentSyncKindFull:
		return "Full"
	case TextDocumentSyncKindIncremental:
		return "Incremental"
	}
	return str.FromInt(int(me))
}

type ErrorCodes int

const (
	ErrorCodesParseError           ErrorCodes = -32700
	ErrorCodesInvalidRequest       ErrorCodes = -32600
	ErrorCodesMethodNotFound       ErrorCodes = -32601
	ErrorCodesInvalidParams        ErrorCodes = -32602
	ErrorCodesInternalError        ErrorCodes = -32603
	ErrorCodesServerNotInitialized ErrorCodes = -32002
	ErrorCodesUnknownErrorCode     ErrorCodes = -32001
)

func (me ErrorCodes) String() string {
	switch me {
	case ErrorCodesParseError:
		return "ParseError"
	case ErrorCodesInvalidRequest:
		return "InvalidRequest"
	case ErrorCodesMethodNotFound:
		return "MethodNotFound"
	case ErrorCodesInvalidParams:
		return "InvalidParams"
	case ErrorCodesInternalError:
		return "InternalError"
	case ErrorCodesServerNotInitialized:
		return "ServerNotInitialized"
	case ErrorCodesUnknownErrorCode:
		return "UnknownErrorCode"
	}
	return str.FromInt(int(me))
}

type ShowMessageRequestParams struct {
	Type    MessageType         `json:"type,omitempty"`
	Message string              `json:"message,omitempty"`
	Actions []MessageActionItem `json:"actions,omitempty"`
}

type SelectionRange struct {
	Range  Range           `json:"range"`
	Parent *SelectionRange `json:"parent,omitempty"`
	Dbg    string
}
