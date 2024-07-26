package lsp

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strconv"
	"sync"
	"time"

	"atmo/util"
	"atmo/util/str"
)

var StdErr = os.Stderr

type Void struct{}

type Server struct {
	stdout    io.Writer
	stdioMu   sync.Mutex // to sync writes to stdout
	waiters   map[any]func(any, any)
	waitersMu sync.Mutex

	LogPrefixSendRecvJsons string
	Initialized            struct {
		Fully  bool
		Client *InitializeParams
		Server *InitializeResult
	}

	Lang struct {
		TriggerChars struct {
			Completion []string
			Signature  []string
		}
		Commands                      []string
		DocumentSymbolsMultiTreeLabel string
	}

	On_workspace_didChangeWorkspaceFolders func(params *DidChangeWorkspaceFoldersParams) (any, error)
	On_initialized                         func(params *InitializedParams) (any, error)
	On_exit                                func(params *Void) (any, error)
	On_textDocument_didOpen                func(params *DidOpenTextDocumentParams) (any, error)
	On_textDocument_didChange              func(params *DidChangeTextDocumentParams) (any, error)
	On_textDocument_didClose               func(params *DidCloseTextDocumentParams) (any, error)
	On_textDocument_didSave                func(params *DidSaveTextDocumentParams) (any, error)
	On_workspace_didChangeWatchedFiles     func(params *DidChangeWatchedFilesParams) (any, error)
	On_textDocument_implementation         func(params *ImplementationParams) (any, error)
	On_textDocument_typeDefinition         func(params *TypeDefinitionParams) (any, error)
	On_textDocument_declaration            func(params *DeclarationParams) (any, error)
	On_textDocument_selectionRange         func(params *SelectionRangeParams) (any, error)
	On_shutdown                            func(params *Void) (any, error)
	On_textDocument_completion             func(params *CompletionParams) (any, error)
	On_textDocument_hover                  func(params *HoverParams) (*Hover, error)
	On_textDocument_signatureHelp          func(params *SignatureHelpParams) (any, error)
	On_textDocument_definition             func(params *DefinitionParams) (any, error)
	On_textDocument_references             func(params *ReferenceParams) (any, error)
	On_textDocument_documentHighlight      func(params *DocumentHighlightParams) (any, error)
	On_textDocument_documentSymbol         func(params *DocumentSymbolParams) ([]DocumentSymbol, error)
	On_textDocument_codeAction             func(params *CodeActionParams) (any, error)
	On_workspace_symbol                    func(params *WorkspaceSymbolParams) ([]WorkspaceSymbol, error)
	On_textDocument_formatting             func(params *DocumentFormattingParams) (any, error)
	On_textDocument_rangeFormatting        func(params *DocumentRangeFormattingParams) (any, error)
	On_textDocument_rename                 func(params *RenameParams) (any, error)
	On_textDocument_prepareRename          func(params *PrepareRenameParams) (any, error)
	On_workspace_executeCommand            func(params *ExecuteCommandParams) (any, error)
}

func (me *Server) Notify_window_showMessage(params ShowMessageParams) {
	go me.send("window/showMessage", params, false, nil)
}

func (me *Server) Notify_window_logMessage(params LogMessageParams) {
	go me.send("window/logMessage", params, false, nil)
}

func (me *Server) Notify_textDocument_publishDiagnostics(params PublishDiagnosticsParams) {
	go me.send("textDocument/publishDiagnostics", params, false, nil)
}

func (me *Server) Request_workspace_workspaceFolders(params Void, onResp func([]WorkspaceFolder)) {
	var on_resp func(any, any) = serverOnResp(me, onResp)
	go me.send("workspace/workspaceFolders", params, true, on_resp)
}

func (me *Server) Request_client_registerCapability(params RegistrationParams, onResp func(Void)) {
	var on_resp func(any, any) = serverOnResp(me, onResp)
	go me.send("client/registerCapability", params, true, on_resp)
}

func (me *Server) Request_window_showMessageRequest(params ShowMessageRequestParams, onResp func(*MessageActionItem)) {
	var on_resp func(any, any) = serverOnResp(me, onResp)
	go me.send("window/showMessageRequest", params, true, on_resp)
}

func (*Server) newId() string { return strconv.FormatInt(time.Now().UnixNano(), 36) }

func (me *Server) send(methodName string, params any, isReq bool, onResp func(any, any)) {
	req_id := me.newId()
	req := map[string]any{"method": methodName, "params": params}
	if onResp != nil {
		me.waitersMu.Lock()
		me.waiters[req_id] = onResp
		me.waitersMu.Unlock()
	}
	if isReq {
		req["id"] = req_id
	}
	me.sendMsg(req)
}

func (me *Server) sendMsg(jsonable any) {
	json_bytes, _ := json.Marshal(jsonable)
	me.stdioMu.Lock()
	defer me.stdioMu.Unlock()
	if me.LogPrefixSendRecvJsons != "" {
		StdErr.WriteString(me.LogPrefixSendRecvJsons + ".SEND>>" + string(json_bytes) + ">>\n")
		_ = StdErr.Sync()
	}
	_, _ = me.stdout.Write([]byte("Content-Length: "))
	_, _ = me.stdout.Write([]byte(strconv.Itoa(len(json_bytes))))
	_, _ = me.stdout.Write([]byte("\r\n\r\n"))
	_, _ = me.stdout.Write(json_bytes)
	_ = os.Stdout.Sync()
}

type jsonRpcError struct {
	Code    ErrorCodes `json:"code"`
	Message string     `json:"message"`
}

func (me *Server) sendErrMsg(err any, msgId any) {
	if err == nil {
		return
	}
	json_rpc_err_msg, is_json_rpc_err_msg := err.(*jsonRpcError)
	if json_rpc_err_msg == nil {
		if is_json_rpc_err_msg {
			return
		}
		json_rpc_err_msg = &jsonRpcError{Code: ErrorCodesInternalError, Message: str.Fmt("%v", err)}
	}
	me.sendMsg(map[string]any{
		"jsonrpc": "2.0",
		"error":   json_rpc_err_msg,
		"id":      msgId,
	})
}

func (me *Server) handleIncoming(raw map[string]any) *jsonRpcError {
	msg_id, msg_method := raw["id"], raw["method"]

	switch msg_method, _ := msg_method.(string); msg_method {
	case "workspace/didChangeWorkspaceFolders":
		serverHandleIncoming(me, me.On_workspace_didChangeWorkspaceFolders, msg_method, msg_id, raw["params"])
	case "initialized":
		serverHandleIncoming(me, me.On_initialized, msg_method, msg_id, raw["params"])
	case "exit":
		serverHandleIncoming(me, me.On_exit, msg_method, msg_id, raw["params"])
	case "textDocument/didOpen":
		serverHandleIncoming(me, me.On_textDocument_didOpen, msg_method, msg_id, raw["params"])
	case "textDocument/didChange":
		serverHandleIncoming(me, me.On_textDocument_didChange, msg_method, msg_id, raw["params"])
	case "textDocument/didClose":
		serverHandleIncoming(me, me.On_textDocument_didClose, msg_method, msg_id, raw["params"])
	case "textDocument/didSave":
		serverHandleIncoming(me, me.On_textDocument_didSave, msg_method, msg_id, raw["params"])
	case "workspace/didChangeWatchedFiles":
		serverHandleIncoming(me, me.On_workspace_didChangeWatchedFiles, msg_method, msg_id, raw["params"])
	case "textDocument/implementation":
		serverHandleIncoming(me, me.On_textDocument_implementation, msg_method, msg_id, raw["params"])
	case "textDocument/typeDefinition":
		serverHandleIncoming(me, me.On_textDocument_typeDefinition, msg_method, msg_id, raw["params"])
	case "textDocument/declaration":
		serverHandleIncoming(me, me.On_textDocument_declaration, msg_method, msg_id, raw["params"])
	case "textDocument/selectionRange":
		serverHandleIncoming(me, me.On_textDocument_selectionRange, msg_method, msg_id, raw["params"])
	case "shutdown":
		serverHandleIncoming(me, me.On_shutdown, msg_method, msg_id, raw["params"])
	case "textDocument/completion":
		serverHandleIncoming(me, me.On_textDocument_completion, msg_method, msg_id, raw["params"])
	case "textDocument/hover":
		serverHandleIncoming(me, me.On_textDocument_hover, msg_method, msg_id, raw["params"])
	case "textDocument/signatureHelp":
		serverHandleIncoming(me, me.On_textDocument_signatureHelp, msg_method, msg_id, raw["params"])
	case "textDocument/definition":
		serverHandleIncoming(me, me.On_textDocument_definition, msg_method, msg_id, raw["params"])
	case "textDocument/references":
		serverHandleIncoming(me, me.On_textDocument_references, msg_method, msg_id, raw["params"])
	case "textDocument/documentHighlight":
		serverHandleIncoming(me, me.On_textDocument_documentHighlight, msg_method, msg_id, raw["params"])
	case "textDocument/documentSymbol":
		serverHandleIncoming(me, me.On_textDocument_documentSymbol, msg_method, msg_id, raw["params"])
	case "textDocument/codeAction":
		serverHandleIncoming(me, me.On_textDocument_codeAction, msg_method, msg_id, raw["params"])
	case "workspace/symbol":
		serverHandleIncoming(me, me.On_workspace_symbol, msg_method, msg_id, raw["params"])
	case "textDocument/formatting":
		serverHandleIncoming(me, me.On_textDocument_formatting, msg_method, msg_id, raw["params"])
	case "textDocument/rangeFormatting":
		serverHandleIncoming(me, me.On_textDocument_rangeFormatting, msg_method, msg_id, raw["params"])
	case "textDocument/rename":
		serverHandleIncoming(me, me.On_textDocument_rename, msg_method, msg_id, raw["params"])
	case "textDocument/prepareRename":
		serverHandleIncoming(me, me.On_textDocument_prepareRename, msg_method, msg_id, raw["params"])
	case "workspace/executeCommand":
		serverHandleIncoming(me, me.On_workspace_executeCommand, msg_method, msg_id, raw["params"])
	case "initialize":
		serverHandleIncoming(me, func(params *InitializeParams) (any, error) {
			init := &me.Initialized
			init.Client = params
			init.Server = &InitializeResult{
				ServerInfo: struct {
					Name    string "json:\"name\""
					Version string "json:\"version,omitempty\""
				}{Name: os.Args[0]},
			}
			caps := &init.Server.Capabilities
			if me.On_textDocument_didClose != nil || me.On_textDocument_didOpen != nil ||
				me.On_textDocument_didChange != nil || me.On_textDocument_didSave != nil {
				caps.TextDocumentSync = &TextDocumentSyncOptions{
					OpenClose: me.On_textDocument_didClose != nil || me.On_textDocument_didOpen != nil,
					Change:    util.If(me.On_textDocument_didChange != nil, TextDocumentSyncKindFull, TextDocumentSyncKindNone),
					Save:      util.If(me.On_textDocument_didSave != nil, &SaveOptions{IncludeText: true}, nil),
				}
			}
			if me.On_textDocument_completion != nil {
				caps.CompletionProvider = &CompletionOptions{TriggerCharacters: me.Lang.TriggerChars.Completion}
			}
			if me.On_textDocument_signatureHelp != nil {
				caps.SignatureHelpProvider = &SignatureHelpOptions{TriggerCharacters: me.Lang.TriggerChars.Signature}
			}
			if me.On_textDocument_rename != nil {
				caps.RenameProvider = &RenameOptions{
					PrepareProvider: (me.On_textDocument_prepareRename != nil),
				}
			}
			if me.On_workspace_executeCommand != nil {
				caps.ExecuteCommandProvider = &ExecuteCommandOptions{Commands: me.Lang.Commands}
			}
			caps.HoverProvider = (me.On_textDocument_hover != nil)
			caps.DeclarationProvider = (me.On_textDocument_declaration != nil)
			caps.DefinitionProvider = (me.On_textDocument_definition != nil)
			caps.TypeDefinitionProvider = (me.On_textDocument_typeDefinition != nil)
			caps.ImplementationProvider = (me.On_textDocument_implementation != nil)
			caps.ReferencesProvider = (me.On_textDocument_references != nil)
			caps.DocumentHighlightProvider = (me.On_textDocument_documentHighlight != nil)
			caps.CodeActionProvider = (me.On_textDocument_codeAction != nil)
			caps.DocumentFormattingProvider = (me.On_textDocument_formatting != nil)
			caps.DocumentRangeFormattingProvider = (me.On_textDocument_rangeFormatting != nil)
			caps.SelectionRangeProvider = (me.On_textDocument_selectionRange != nil)
			caps.WorkspaceSymbolProvider = (me.On_workspace_symbol != nil)
			if me.On_textDocument_documentSymbol != nil {
				caps.DocumentSymbolProvider = &DocumentSymbolOptions{
					Label: util.If(me.Lang.DocumentSymbolsMultiTreeLabel == "", "(lsp.Server.Lang.DocumentSymbolsMultiTreeLabel)", me.Lang.DocumentSymbolsMultiTreeLabel),
				}
			}
			if me.On_workspace_didChangeWorkspaceFolders != nil {
				caps.Workspace = struct {
					WorkspaceFolders WorkspaceFoldersServerCapabilities "json:\"workspaceFolders,omitempty\""
				}{
					WorkspaceFolders: WorkspaceFoldersServerCapabilities{
						Supported:           true,
						ChangeNotifications: true,
					},
				}
			}
			return init.Server, nil
		}, msg_method, msg_id, raw["params"])
	default: // msg is an incoming Request or Notification
		if msg_id != nil { // a Request (not a Notification) that was sent despite lacking server support
			return &jsonRpcError{Code: ErrorCodesMethodNotFound, Message: "unknown method: " + msg_method}
		}
	}

	return nil
}

// Forever keeps reading and handling LSP JSON-RPC messages incoming over `os.Stdin`
// until reading from `os.Stdin` fails, then returns that IO read error.
func (me *Server) Forever() error {
	{ // users shouldn't have to set up no-op handlers for these routine teardown lifecycle messages:
		old_shutdown, old_exit, old_initialized := me.On_shutdown, me.On_exit, me.On_initialized
		me.On_shutdown = func(params *Void) (any, error) {
			if old_shutdown != nil {
				return old_shutdown(params)
			}
			return nil, nil
		}
		me.On_exit = func(params *Void) (any, error) {
			if old_exit != nil {
				return old_exit(params)
			}
			os.Exit(0)
			return nil, nil
		}
		me.On_initialized = func(params *InitializedParams) (any, error) {
			if me.On_workspace_didChangeWatchedFiles != nil {
				me.Request_client_registerCapability(RegistrationParams{
					Registrations: []Registration{
						{Method: "workspace/didChangeWatchedFiles", Id: me.newId(),
							RegisterOptions: DidChangeWatchedFilesRegistrationOptions{Watchers: []FileSystemWatcher{
								{Kind: WatchKindAll,
									GlobPattern: struct {
										Pattern string "json:\"pattern\""
									}{Pattern: "**/*"}}}}},
					},
				}, func(Void) {})
			}
			if old_initialized != nil {
				return old_initialized(params)
			}
			me.Initialized.Fully = true
			return nil, nil
		}
	}

	return me.forever(os.Stdin, os.Stdout, me.handleIncoming)
}

// forever keeps reading and handling LSP JSON-RPC messages incoming over
// `in` until reading from `in` fails, then returns that IO read error.
func (me *Server) forever(in io.Reader, out io.Writer, handleIncoming func(map[string]any) *jsonRpcError) error {
	const buf_cap = 1024 * 1024

	me.stdout = out
	me.waiters = map[any]func(any, any){}

	stdin := bufio.NewScanner(in)
	stdin.Split(func(data []byte, ateof bool) (advance int, token []byte, err error) {
		if i_cl1 := bytes.Index(data, []byte("Content-Length: ")); i_cl1 >= 0 {
			datafromclen := data[i_cl1+16:]
			if i_cl2 := bytes.IndexAny(datafromclen, "\r\n"); i_cl2 > 0 {
				if clen, e := strconv.Atoi(string(datafromclen[:i_cl2])); e != nil {
					err = e
				} else if i_js1 := bytes.Index(datafromclen, []byte("{\"")); i_js1 > i_cl2 {
					if i_js2 := i_js1 + clen; len(datafromclen) >= i_js2 {
						advance = i_cl1 + 16 + i_js2
						token = datafromclen[i_js1:i_js2]
					}
				}
			}
		}
		return
	})

	for stdin.Scan() {
		raw := map[string]any{}
		json_bytes := stdin.Bytes()
		if me.LogPrefixSendRecvJsons != "" {
			me.stdioMu.Lock()
			StdErr.WriteString(me.LogPrefixSendRecvJsons + ".RECV<<" + string(json_bytes) + "<<\n")
			_ = StdErr.Sync()
			me.stdioMu.Unlock()
		}
		if err := json.Unmarshal(json_bytes, &raw); err != nil {
			StdErr.WriteString("failed to parse incoming JSON message '" + string(json_bytes) + "': " + err.Error() + "\n")
			continue
		}
		msg_id := raw["id"]
		me.waitersMu.Lock()
		handler := me.waiters[msg_id]
		delete(me.waiters, msg_id)
		me.waitersMu.Unlock()

		if raw["code"] != nil { // received an error message
			me.stdioMu.Lock()
			StdErr.WriteString(string(json_bytes) + "\n")
			me.stdioMu.Unlock()
			continue
		}

		if raw["method"] == nil { // received a Response message
			go handler(raw["result"], msg_id)
		} else {
			me.sendErrMsg(handleIncoming(raw), msg_id)
		}
	}
	return stdin.Err()
}

func serverOnResp[T any](me *Server, onResp func(T)) func(any, any) {
	if onResp == nil {
		return nil
	}
	return func(resultAsMap any, msgId any) {
		var result, none T
		if resultAsMap != nil {
			json_bytes, _ := json.Marshal(resultAsMap)
			if err := json.Unmarshal(json_bytes, &result); err != nil {
				me.sendErrMsg(err, msgId)
				return
			}
		}
		onResp(util.If(resultAsMap == nil, none, result))
	}
}

func serverHandleIncoming[TIn any, TOut any](me *Server, handler func(*TIn) (TOut, error), msgMethodName string, msgId any, msgParams any) {
	if handler == nil {
		if msgId != nil {
			me.sendErrMsg(errors.New("unimplemented: "+msgMethodName), msgId)
		}
		return
	}
	var params TIn
	if msgParams != nil {
		json_bytes, _ := json.Marshal(msgParams)
		if err := json.Unmarshal(json_bytes, &params); err != nil {
			me.sendErrMsg(&jsonRpcError{Code: ErrorCodesInvalidParams, Message: err.Error()}, msgId)
			return
		}
	}
	go func(params *TIn) {
		if msgParams == nil {
			params = nil
		}
		result, err := handler(params)
		if msgId != nil {
			resp := map[string]any{
				"jsonrpc": "2.0",
				"result":  result,
				"id":      msgId,
			}
			if err != nil {
				delete(resp, "result")
				resp["error"] = &jsonRpcError{Code: ErrorCodesInternalError, Message: str.Fmt("%v", err)}
			}
			me.sendMsg(resp)
		} else if err != nil {
			StdErr.WriteString("handler for Notification '" + msgMethodName + "' failed: " + err.Error() + "\n")
		}
	}(&params)
}
