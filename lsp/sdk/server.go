package lsp

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"sync"
	"time"

	"atmo/util"
)

type Void struct{}

type Server struct {
	stdioLock sync.Mutex // to sync writes to stdout
	stdout    io.Writer
	waiters   map[any]func(any)

	LogPrefixSendRecvJsons string
	Initialized            struct {
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

func (it *Server) Notify_window_showMessage(params ShowMessageParams) {
	var on_resp func(any)
	go it.send("window/showMessage", params, false, on_resp)
}

func (it *Server) Notify_window_logMessage(params LogMessageParams) {
	var on_resp func(any)
	go it.send("window/logMessage", params, false, on_resp)
}

func (it *Server) Notify_textDocument_publishDiagnostics(params PublishDiagnosticsParams) {
	var on_resp func(any)
	go it.send("textDocument/publishDiagnostics", params, false, on_resp)
}

func (it *Server) Request_workspace_workspaceFolders(params Void, onResp func([]WorkspaceFolder)) {
	var on_resp func(any) = serverOnResp(it, onResp)
	go it.send("workspace/workspaceFolders", params, true, on_resp)
}

func (it *Server) Request_client_registerCapability(params RegistrationParams, onResp func(Void)) {
	var on_resp func(any) = serverOnResp(it, onResp)
	go it.send("client/registerCapability", params, true, on_resp)
}

func (it *Server) Request_window_showMessageRequest(params ShowMessageRequestParams, onResp func(*MessageActionItem)) {
	var on_resp func(any) = serverOnResp(it, onResp)
	go it.send("window/showMessageRequest", params, true, on_resp)
}

func (it *Server) send(methodName string, params any, isReq bool, onResp func(any)) {
	req_id := strconv.FormatInt(time.Now().UnixNano(), 36)
	req := map[string]any{"method": methodName, "params": params}
	if onResp != nil {
		it.waiters[req_id] = onResp
	}
	if isReq {
		req["id"] = req_id
	}
	it.sendMsg(req)
}

func (it *Server) sendMsg(jsonable any) {
	json_bytes, _ := json.Marshal(jsonable)
	it.stdioLock.Lock()
	defer it.stdioLock.Unlock()
	if it.LogPrefixSendRecvJsons != "" {
		println(it.LogPrefixSendRecvJsons + ".SEND>>" + string(json_bytes) + ">>")
	}
	_, _ = it.stdout.Write([]byte("Content-Length: "))
	_, _ = it.stdout.Write([]byte(strconv.Itoa(len(json_bytes))))
	_, _ = it.stdout.Write([]byte("\r\n\r\n"))
	_, _ = it.stdout.Write(json_bytes)
}

type jsonRpcError struct {
	Code    ErrorCodes `json:"code"`
	Message string     `json:"message"`
}

func (it *Server) sendErrMsg(err any) {
	if err == nil {
		return
	}
	json_rpc_err_msg, is_json_rpc_err_msg := err.(*jsonRpcError)
	if json_rpc_err_msg == nil {
		if is_json_rpc_err_msg {
			return
		}
		json_rpc_err_msg = &jsonRpcError{Code: ErrorCodesInternalError, Message: fmt.Sprintf("%v", err)}
	}
	it.sendMsg(json_rpc_err_msg)
}

func (it *Server) handleIncoming(raw map[string]any) *jsonRpcError {
	msg_id, msg_method := raw["id"], raw["method"]

	switch msg_method, _ := msg_method.(string); msg_method {
	case "workspace/didChangeWorkspaceFolders":
		serverHandleIncoming(it, it.On_workspace_didChangeWorkspaceFolders, msg_method, msg_id, raw["params"])
	case "initialized":
		serverHandleIncoming(it, it.On_initialized, msg_method, msg_id, raw["params"])
	case "exit":
		serverHandleIncoming(it, it.On_exit, msg_method, msg_id, raw["params"])
	case "textDocument/didOpen":
		serverHandleIncoming(it, it.On_textDocument_didOpen, msg_method, msg_id, raw["params"])
	case "textDocument/didChange":
		serverHandleIncoming(it, it.On_textDocument_didChange, msg_method, msg_id, raw["params"])
	case "textDocument/didClose":
		serverHandleIncoming(it, it.On_textDocument_didClose, msg_method, msg_id, raw["params"])
	case "textDocument/didSave":
		serverHandleIncoming(it, it.On_textDocument_didSave, msg_method, msg_id, raw["params"])
	case "workspace/didChangeWatchedFiles":
		serverHandleIncoming(it, it.On_workspace_didChangeWatchedFiles, msg_method, msg_id, raw["params"])
	case "textDocument/implementation":
		serverHandleIncoming(it, it.On_textDocument_implementation, msg_method, msg_id, raw["params"])
	case "textDocument/typeDefinition":
		serverHandleIncoming(it, it.On_textDocument_typeDefinition, msg_method, msg_id, raw["params"])
	case "textDocument/declaration":
		serverHandleIncoming(it, it.On_textDocument_declaration, msg_method, msg_id, raw["params"])
	case "textDocument/selectionRange":
		serverHandleIncoming(it, it.On_textDocument_selectionRange, msg_method, msg_id, raw["params"])
	case "shutdown":
		serverHandleIncoming(it, it.On_shutdown, msg_method, msg_id, raw["params"])
	case "textDocument/completion":
		serverHandleIncoming(it, it.On_textDocument_completion, msg_method, msg_id, raw["params"])
	case "textDocument/hover":
		serverHandleIncoming(it, it.On_textDocument_hover, msg_method, msg_id, raw["params"])
	case "textDocument/signatureHelp":
		serverHandleIncoming(it, it.On_textDocument_signatureHelp, msg_method, msg_id, raw["params"])
	case "textDocument/definition":
		serverHandleIncoming(it, it.On_textDocument_definition, msg_method, msg_id, raw["params"])
	case "textDocument/references":
		serverHandleIncoming(it, it.On_textDocument_references, msg_method, msg_id, raw["params"])
	case "textDocument/documentHighlight":
		serverHandleIncoming(it, it.On_textDocument_documentHighlight, msg_method, msg_id, raw["params"])
	case "textDocument/documentSymbol":
		serverHandleIncoming(it, it.On_textDocument_documentSymbol, msg_method, msg_id, raw["params"])
	case "textDocument/codeAction":
		serverHandleIncoming(it, it.On_textDocument_codeAction, msg_method, msg_id, raw["params"])
	case "workspace/symbol":
		serverHandleIncoming(it, it.On_workspace_symbol, msg_method, msg_id, raw["params"])
	case "textDocument/formatting":
		serverHandleIncoming(it, it.On_textDocument_formatting, msg_method, msg_id, raw["params"])
	case "textDocument/rangeFormatting":
		serverHandleIncoming(it, it.On_textDocument_rangeFormatting, msg_method, msg_id, raw["params"])
	case "textDocument/rename":
		serverHandleIncoming(it, it.On_textDocument_rename, msg_method, msg_id, raw["params"])
	case "textDocument/prepareRename":
		serverHandleIncoming(it, it.On_textDocument_prepareRename, msg_method, msg_id, raw["params"])
	case "workspace/executeCommand":
		serverHandleIncoming(it, it.On_workspace_executeCommand, msg_method, msg_id, raw["params"])
	case "initialize":
		serverHandleIncoming(it, func(params *InitializeParams) (any, error) {
			init := &it.Initialized
			init.Client = params
			init.Server = &InitializeResult{
				ServerInfo: struct {
					Name    string "json:\"name\""
					Version string "json:\"version,omitempty\""
				}{Name: os.Args[0]},
			}
			caps := &init.Server.Capabilities
			if it.On_textDocument_didClose != nil || it.On_textDocument_didOpen != nil ||
				it.On_textDocument_didChange != nil || it.On_textDocument_didSave != nil {
				caps.TextDocumentSync = &TextDocumentSyncOptions{
					OpenClose: it.On_textDocument_didClose != nil || it.On_textDocument_didOpen != nil,
					Change:    util.If(it.On_textDocument_didChange != nil, TextDocumentSyncKindFull, TextDocumentSyncKindNone),
					Save:      util.If(it.On_textDocument_didSave != nil, &SaveOptions{IncludeText: true}, nil),
				}
			}
			if it.On_textDocument_completion != nil {
				caps.CompletionProvider = &CompletionOptions{TriggerCharacters: it.Lang.TriggerChars.Completion}
			}
			if it.On_textDocument_signatureHelp != nil {
				caps.SignatureHelpProvider = &SignatureHelpOptions{TriggerCharacters: it.Lang.TriggerChars.Signature}
			}
			if it.On_textDocument_rename != nil {
				caps.RenameProvider = &RenameOptions{
					PrepareProvider: (it.On_textDocument_prepareRename != nil),
				}
			}
			if it.On_workspace_executeCommand != nil {
				caps.ExecuteCommandProvider = &ExecuteCommandOptions{Commands: it.Lang.Commands}
			}
			caps.HoverProvider = (it.On_textDocument_hover != nil)
			caps.DeclarationProvider = (it.On_textDocument_declaration != nil)
			caps.DefinitionProvider = (it.On_textDocument_definition != nil)
			caps.TypeDefinitionProvider = (it.On_textDocument_typeDefinition != nil)
			caps.ImplementationProvider = (it.On_textDocument_implementation != nil)
			caps.ReferencesProvider = (it.On_textDocument_references != nil)
			caps.DocumentHighlightProvider = (it.On_textDocument_documentHighlight != nil)
			caps.CodeActionProvider = (it.On_textDocument_codeAction != nil)
			caps.DocumentFormattingProvider = (it.On_textDocument_formatting != nil)
			caps.DocumentRangeFormattingProvider = (it.On_textDocument_rangeFormatting != nil)
			caps.SelectionRangeProvider = (it.On_textDocument_selectionRange != nil)
			caps.WorkspaceSymbolProvider = (it.On_workspace_symbol != nil)
			if it.On_textDocument_documentSymbol != nil {
				caps.DocumentSymbolProvider = &DocumentSymbolOptions{
					Label: util.If(it.Lang.DocumentSymbolsMultiTreeLabel == "", "(lsp.Server.Lang.DocumentSymbolsMultiTreeLabel)", it.Lang.DocumentSymbolsMultiTreeLabel),
				}
			}
			if it.On_workspace_didChangeWorkspaceFolders != nil {
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
		return &jsonRpcError{Code: ErrorCodesMethodNotFound, Message: "unknown method: " + msg_method}
	}

	return nil
}

// Forever keeps reading and handling LSP JSON-RPC messages incoming over `os.Stdin`
// until reading from `os.Stdin` fails, then returns that IO read error.
func (it *Server) Forever() error {
	{ // users shouldn't have to set up no-op handlers for these routine teardown lifecycle messages:
		old_shutdown, old_exit, old_initialized := it.On_shutdown, it.On_exit, it.On_initialized
		it.On_shutdown = func(params *Void) (any, error) {
			if old_shutdown != nil {
				return old_shutdown(params)
			}
			return nil, nil
		}
		it.On_exit = func(params *Void) (any, error) {
			if old_exit != nil {
				return old_exit(params)
			}
			os.Exit(0)
			return nil, nil
		}
		it.On_initialized = func(params *InitializedParams) (any, error) {
			if it.On_workspace_didChangeWatchedFiles != nil {
				it.Request_client_registerCapability(RegistrationParams{
					Registrations: []Registration{
						{Method: "workspace/didChangeWatchedFiles", Id: "workspace/didChangeWatchedFiles",
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
			return nil, nil
		}
	}

	return it.forever(os.Stdin, os.Stdout, it.handleIncoming)
}

// forever keeps reading and handling LSP JSON-RPC messages incoming over
// `in` until reading from `in` fails, then returns that IO read error.
func (it *Server) forever(in io.Reader, out io.Writer, handleIncoming func(map[string]any) *jsonRpcError) error {
	const buf_cap = 1024 * 1024

	it.stdout = out
	it.waiters = map[any]func(any){}

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
		if it.LogPrefixSendRecvJsons != "" {
			it.stdioLock.Lock()
			println(it.LogPrefixSendRecvJsons + ".RECV<<" + string(json_bytes) + "<<")
			it.stdioLock.Unlock()
		}
		if err := json.Unmarshal(json_bytes, &raw); err != nil {
			it.sendErrMsg(&jsonRpcError{Code: ErrorCodesParseError, Message: err.Error()})
			continue
		}

		if raw["code"] != nil { // received an error message
			it.stdioLock.Lock()
			println(string(json_bytes)) // goes to stderr
			it.stdioLock.Unlock()
			continue
		}
		if msg_id := raw["id"]; raw["method"] == nil { // received a Response message
			handler := it.waiters[msg_id]
			delete(it.waiters, msg_id)
			go handler(raw["result"])
		} else {
			it.sendErrMsg(handleIncoming(raw))
		}
	}
	return stdin.Err()
}

func serverOnResp[T any](it *Server, onResp func(T)) func(any) {
	if onResp == nil {
		return nil
	}
	return func(resultAsMap any) {
		var result, none T
		if resultAsMap != nil {
			json_bytes, _ := json.Marshal(resultAsMap)
			if err := json.Unmarshal(json_bytes, &result); err != nil {
				it.sendErrMsg(err)
				return
			}
		}
		onResp(util.If(resultAsMap == nil, none, result))
	}
}

func serverHandleIncoming[TIn any, TOut any](it *Server, handler func(*TIn) (TOut, error), msgMethodName string, msgId any, msgParams any) {
	if handler == nil {
		if msgId != nil {
			it.sendErrMsg(errors.New("unimplemented: " + msgMethodName))
		}
		return
	}
	var params TIn
	if msgParams != nil {
		json_bytes, _ := json.Marshal(msgParams)
		if err := json.Unmarshal(json_bytes, &params); err != nil {
			it.sendErrMsg(&jsonRpcError{Code: ErrorCodesInvalidParams, Message: err.Error()})
			return
		}
	}
	go func(params *TIn) {
		if msgParams == nil {
			params = nil
		}
		result, err := handler(params)
		resp := map[string]any{
			"jsonrpc": "2.0",
			"result":  result,
			"id":      msgId,
		}
		if err != nil {
			if msgId != nil {
				resp["error"] = &jsonRpcError{Code: ErrorCodesInternalError, Message: fmt.Sprintf("%v", err)}
			} else {
				it.sendErrMsg(err)
				return
			}
		}
		if msgId != nil {
			it.sendMsg(resp)
		}
	}(&params)
}
