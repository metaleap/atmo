package lsp

import (
	lsp "atmo/lsp/sdk"
	"atmo/session"
	"atmo/util"
	"atmo/util/sl"
	"atmo/util/str"
)

const temporarilyListAllIcons = false

func init() {
	Server.Lang.DocumentSymbolsMultiTreeLabel = "Atmo"
	Server.Lang.TriggerChars.Completion = []string{".", "/"}
	Server.Lang.TriggerChars.Signature = []string{" "}

	Server.On_textDocument_documentSymbol = func(params *lsp.DocumentSymbolParams) (ret []lsp.DocumentSymbol, _ error) {
		src_file_path := lspUriToFsPath(params.TextDocument.Uri)
		session.Access(func(sess session.StateAccess, intel session.Intel) {
			if src_file := sess.SrcFile(src_file_path); src_file != nil {
				ret = sl.To(intel.Decls(nil, src_file, false, ""), toLspDocumentSymbol)
			}
		})
		if temporarilyListAllIcons {
			ret = sl.To([]lsp.SymbolKind{lsp.SymbolKindArray, lsp.SymbolKindBoolean, lsp.SymbolKindClass, lsp.SymbolKindConstant, lsp.SymbolKindConstructor, lsp.SymbolKindEnum, lsp.SymbolKindEnumMember, lsp.SymbolKindEvent, lsp.SymbolKindField, lsp.SymbolKindFile, lsp.SymbolKindFunction, lsp.SymbolKindInterface, lsp.SymbolKindKey, lsp.SymbolKindMethod, lsp.SymbolKindModule, lsp.SymbolKindNamespace, lsp.SymbolKindNull, lsp.SymbolKindNumber, lsp.SymbolKindObject, lsp.SymbolKindOperator, lsp.SymbolKindPackage, lsp.SymbolKindProperty, lsp.SymbolKindString, lsp.SymbolKindStruct, lsp.SymbolKindTypeParameter, lsp.SymbolKindVariable},
				func(it lsp.SymbolKind) lsp.DocumentSymbol {
					return lsp.DocumentSymbol{
						Name:           it.String(),
						Detail:         str.Fmt("**TODO:** documentSymbols for `%v`", src_file_path),
						Kind:           it,
						Range:          lsp.Range{Start: lsp.Position{Line: 2, Character: 1}, End: lsp.Position{Line: 2, Character: 8}},
						SelectionRange: lsp.Range{Start: lsp.Position{Line: 2, Character: 3}, End: lsp.Position{Line: 2, Character: 6}},
					}
				})
		}
		return
	}

	Server.On_workspace_symbol = func(params *lsp.WorkspaceSymbolParams) (ret []lsp.WorkspaceSymbol, _ error) {
		session.Access(func(sess session.StateAccess, intel session.Intel) {
			ret = sl.To(intel.Decls(nil, nil, true, params.Query), toLspWorkspaceSymbol)
		})
		return
	}

	Server.On_textDocument_definition = func(params *lsp.DefinitionParams) ([]lsp.Location, error) {
		return intelLookup(session.IntelLookupKindDefs, &params.TextDocumentPositionParams), nil
	}

	Server.On_textDocument_declaration = func(params *lsp.DeclarationParams) ([]lsp.Location, error) {
		return intelLookup(session.IntelLookupKindDecls, &params.TextDocumentPositionParams), nil
	}

	Server.On_textDocument_typeDefinition = func(params *lsp.TypeDefinitionParams) ([]lsp.Location, error) {
		return intelLookup(session.IntelLookupKindTypes, &params.TextDocumentPositionParams), nil
	}

	Server.On_textDocument_implementation = func(params *lsp.ImplementationParams) ([]lsp.Location, error) {
		return intelLookup(session.IntelLookupKindImpls, &params.TextDocumentPositionParams), nil
	}

	Server.On_textDocument_references = func(params *lsp.ReferenceParams) ([]lsp.Location, error) {
		return intelLookup(session.IntelLookupKindRefs, &params.TextDocumentPositionParams), nil
	}

	Server.On_textDocument_documentHighlight = func(params *lsp.DocumentHighlightParams) (ret []lsp.DocumentHighlight, _ error) {
		src_file_path := lspUriToFsPath(params.TextDocument.Uri)
		session.Access(func(sess session.StateAccess, intel session.Intel) {
			if src_file := sess.SrcFile(src_file_path); src_file != nil {
				for _, locs := range intel.Lookup(session.IntelLookupKindRefs, src_file, lspPosToPos(&params.Position), true) {
					for i, span := range locs.Spans {
						it := lsp.DocumentHighlight{Range: lspRangeFromSpan(span), Kind: lsp.DocumentHighlightKindText}
						if (len(locs.IsGet) == len(locs.Spans)) && (locs.IsGet[i]) {
							it.Kind = lsp.DocumentHighlightKindRead
						} else if (len(locs.IsSet) == len(locs.Spans)) && (locs.IsSet[i]) {
							it.Kind = lsp.DocumentHighlightKindWrite
						}
						ret = append(ret, it)
					}
				}
			}
		})
		return
	}

	Server.On_textDocument_completion = func(params *lsp.CompletionParams) ([]lsp.CompletionItem, error) {
		src_file_path := lspUriToFsPath(params.TextDocument.Uri)
		return sl.To([]lsp.CompletionItemKind{
			lsp.CompletionItemKindClass,
			lsp.CompletionItemKindColor,
			lsp.CompletionItemKindConstant,
			lsp.CompletionItemKindConstructor,
			lsp.CompletionItemKindEnum,
			lsp.CompletionItemKindEnumMember,
			lsp.CompletionItemKindEvent,
			lsp.CompletionItemKindField,
			lsp.CompletionItemKindFile,
			lsp.CompletionItemKindFolder,
			lsp.CompletionItemKindFunction,
			lsp.CompletionItemKindInterface,
			lsp.CompletionItemKindKeyword,
			lsp.CompletionItemKindMethod,
			lsp.CompletionItemKindModule,
			lsp.CompletionItemKindOperator,
			lsp.CompletionItemKindProperty,
			lsp.CompletionItemKindReference,
			lsp.CompletionItemKindSnippet,
			lsp.CompletionItemKindStruct,
			lsp.CompletionItemKindText,
			lsp.CompletionItemKindTypeParameter,
			lsp.CompletionItemKindUnit,
			lsp.CompletionItemKindValue,
			lsp.CompletionItemKindVariable,
		}, func(it lsp.CompletionItemKind) lsp.CompletionItem {
			return lsp.CompletionItem{
				Label: it.String(),
				Kind:  it,
				Documentation: &lsp.MarkupContent{Kind: lsp.MarkupKindMarkdown,
					Value: str.Fmt("**TODO** _%s_ for `%s` @ %d,%d", it.String(), src_file_path, params.Position.Line, params.Position.Character)},
				Detail: "Detail",
				LabelDetails: &lsp.CompletionItemLabelDetails{
					Detail:      " · LD_Detail " + it.String(),
					Description: "LD_Description " + it.String(),
				},
			}
		}), nil
	}

	Server.On_textDocument_hover = func(params *lsp.HoverParams) (ret *lsp.Hover, _ error) {
		src_file_path := lspUriToFsPath(params.TextDocument.Uri)
		session.Access(func(sess session.StateAccess, intel session.Intel) {
			if src_file := sess.SrcFile(src_file_path); src_file != nil {
				if info := intel.Info(src_file, lspPosToPos(&params.Position)); info != nil {
					items := info.Items.Where(session.IntelItemKindDescription)
					strs := sl.Where(sl.To(items, func(it session.IntelItem) string { return it.Value }), func(s string) bool { return s != "" })
					if text := str.Join(strs, "\n___\n"); text != "" {
						ret = &lsp.Hover{
							Contents: lsp.MarkupContent{Value: text, Kind: lsp.MarkupKindMarkdown},
						}
						if info.SpanFull != nil {
							ret.Range = util.Ptr(lspRangeFromSpan(info.SpanFull))
						}
					}
				}
			}
		})
		return
	}

	Server.On_textDocument_prepareRename = func(params *lsp.PrepareRenameParams) (ret *lsp.Range, _ error) {
		src_file_path := lspUriToFsPath(params.TextDocument.Uri)
		session.Access(func(sess session.StateAccess, intel session.Intel) {
			if src_file := sess.SrcFile(src_file_path); src_file != nil {
				if span := intel.CanRename(src_file, lspPosToPos(&params.Position)); span != nil {
					ret = util.Ptr(lspRangeFromSpan(span))
				}
			}
		})
		return
	}

	Server.On_textDocument_rename = func(params *lsp.RenameParams) (ret *lsp.WorkspaceEdit, _ error) {
		src_file_path := lspUriToFsPath(params.TextDocument.Uri)
		session.Access(func(sess session.StateAccess, intel session.Intel) {
			if src_file := sess.SrcFile(src_file_path); src_file != nil {
				if refs := intel.Lookup(session.IntelLookupKindRefs, src_file, lspPosToPos(&params.Position), false); len(refs) > 0 {
					ret = &lsp.WorkspaceEdit{Changes: map[string][]lsp.TextEdit{}}
					for _, locs := range refs {
						if len(locs.Spans) > 0 {
							ret.Changes[lspUriFromFsPath(locs.File.FilePath)] = sl.To(locs.Spans, func(span *session.SrcFileSpan) lsp.TextEdit {
								return lsp.TextEdit{Range: lspRangeFromSpan(span), NewText: params.NewName}
							})
						}
					}
				}
			}
		})
		return
	}

	Server.On_textDocument_signatureHelp = func(params *lsp.SignatureHelpParams) (ret *lsp.SignatureHelp, _ error) {
		src_file_path := lspUriToFsPath(params.TextDocument.Uri)
		return &lsp.SignatureHelp{
			Signatures: util.If(params.Position.Line > 0,
				nil,
				[]lsp.SignatureInformation{{
					Label: "(foo bar: #baz)",
					Documentation: &lsp.MarkupContent{
						Kind:  lsp.MarkupKindMarkdown,
						Value: str.Fmt("**TODO**: sig help for `%s` @ %d,%d", src_file_path, params.Position.Line, params.Position.Character)},
				}}),
		}, nil
	}

	Server.On_textDocument_selectionRange = func(params *lsp.SelectionRangeParams) (ret []*lsp.SelectionRange, _ error) {
		src_file_path := lspUriToFsPath(params.TextDocument.Uri)
		if len(params.Positions) > 0 && session.IsSrcFilePath(src_file_path) {
			session.Access(func(sess session.StateAccess, _ session.Intel) {
				if src_file := sess.SrcFile(src_file_path); src_file != nil {
					for _, pos := range params.Positions {
						if node := src_file.NodeAtPos(lspPosToPos(&pos), true); node == nil {
							ret = nil
							break
						} else {
							all := sl.To(node.SelfAndAncestors(), func(it *session.AstNode) *lsp.SelectionRange {
								return &lsp.SelectionRange{Range: lspRangeFromSpan(util.Ptr(it.Toks.Span()))}
							})
							for i, it := range all[:len(all)-1] {
								it.Parent = all[i+1]
							}
							ret = append(ret, all[0])
						}
					}
				}
			})
		}
		return
	}

}

func intelLookup(kind session.IntelLookupKind, params *lsp.TextDocumentPositionParams) (ret []lsp.Location) {
	src_file_path := lspUriToFsPath(params.TextDocument.Uri)
	session.Access(func(sess session.StateAccess, intel session.Intel) {
		if src_file := sess.SrcFile(src_file_path); src_file != nil {
			for _, locs := range intel.Lookup(kind, src_file, lspPosToPos(&params.Position), false) {
				ret = append(ret, toLspLocations(locs)...)
			}
		}
	})
	return
}

func toLspLocations(from ...*session.SrcFileLocs) (ret []lsp.Location) {
	for _, loc := range from {
		for _, span := range loc.Spans {
			ret = append(ret, lsp.Location{Range: lspRangeFromSpan(span), Uri: lspUriFromFsPath(loc.File.FilePath)})
		}
	}
	return
}

func toLspDocumentSymbol(info *session.IntelInfo) (sym lsp.DocumentSymbol) {
	sym.Kind, sym.Name = lsp.SymbolKindVariable, info.Items.Name().Value
	if descr := info.Items.First(session.IntelItemKindDescription); descr != nil {
		sym.Detail = descr.Value
	}
	if (info.SpanIdent != nil) && (info.SpanFull != nil) {
		sym.SelectionRange = lspRangeFromSpan(info.SpanIdent)
		sym.Range = lspRangeFromSpan(info.SpanFull)
	}
	for _, item := range info.Items.Where(session.IntelItemKindKind) {
		switch item.Value {
		case string(session.IntelDeclKindFunc):
			sym.Kind = lsp.SymbolKindFunction
		}
	}
	sym.Children = sl.To(info.Sub, toLspDocumentSymbol)
	return
}

func toLspWorkspaceSymbol(info *session.IntelInfo) (sym lsp.WorkspaceSymbol) {
	sym.Kind, sym.Name = lsp.SymbolKindVariable, info.Items.Name().Value
	if pack_dir_path := info.Items.First(session.IntelItemKindSrcPackDirPath); pack_dir_path != nil {
		sym.ContainerName = pack_dir_path.Value
	}
	if src_file_path := info.Items.First(session.IntelItemKindSrcFilePath); (src_file_path != nil) && (info.SpanIdent != nil) {
		sym.Location = lsp.Location{
			Uri:   lspUriFromFsPath(src_file_path.Value),
			Range: lspRangeFromSpan(info.SpanIdent),
		}
	}
	for _, item := range info.Items.Where(session.IntelItemKindKind) {
		switch item.Value {
		case string(session.IntelDeclKindFunc):
			sym.Kind = lsp.SymbolKindFunction
		}
	}
	return
}
