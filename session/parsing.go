package session

import (
	"cmp"
	"errors"
	"strconv"
	"unicode/utf8"

	"atmo/util"
	"atmo/util/sl"
	"atmo/util/str"
)

type AstNodes sl.Of[*AstNode]

type AstNode struct {
	parent     *AstNode
	Kind       AstNodeKind
	Src        string
	Toks       Toks
	Nodes      AstNodes `json:",omitempty"`
	errParsing *Diag
	Lit        any `json:",omitempty"` // if AstNodeKindIdent or AstNodeKindLit, one of: float64 | int64 | uint64 | rune | string
}

type AstNodeKind int

const (
	AstNodeKindErr     AstNodeKind = iota
	AstNodeKindComment             // both /* multi-line */ and // single-line
	AstNodeKindIdent               // foo, #bar, @baz, $foo, %bar, ==, <==<
	AstNodeKindLit                 // 123, -321, 1.23, -3.21, "foo", `bar`, 'ö'
	AstNodeKindGroup               // foo bar, (), (foo), (foo bar), [], [foo], [foo bar], {}, {foo}, {foo bar}
	AstNodeKindBlockLine
)

// only called by EnsureSrcFile, just after tokenization, with `.diags.LexErrs` freshly set.
// mutates me.Content.TopLevelAstNodes and me.diags.ParseErrs.
func (me *SrcFile) parse() AstNodes {
	parsed := me.parseNodes(me.Src.Toks)

	// group huddled exprs: `foo x+z y` right now is `foo x + z y` BUT lets make it `foo (x + 1) y`:
	parsed.walk(nil, func(node *AstNode) {
		node.Nodes = node.Nodes.huddled(me, node)
	})

	// set all `AstNode.parent`s only after above re-arrangements; also
	parsed.walk(nil, func(node *AstNode) {
		for _, it := range node.Nodes {
			it.parent = node
		}
	})
	// for scripting use-case, a file beginning of `#!/usr/bin/env ` gets commented in AST
	for _, node := range parsed {
		if (node.Toks[0].Pos.Line == 1) && (node.Toks[0].Pos.Char == 1) && str.Begins(node.Src, "#!/usr/bin/env ") {
			node.Kind = AstNodeKindComment
			break
		}
	}
	return parsed
}

func (me *SrcFile) parseNode(toks Toks) *AstNode {
	nodes := me.parseNodes(toks)
	if len(nodes) == 1 {
		return nodes[0]
	}
	return &AstNode{Kind: AstNodeKindGroup, Nodes: nodes, Toks: toks, Src: toks.src(me.Src.Text)}
}

func (me *SrcFile) parseNodes(toks Toks) (ret AstNodes) {
	var stack []AstNodes // in case of indents/dedents in toks
	var had_brace_err bool
	for len(toks) > 0 {
		tok := toks[0]
		switch tok.Kind {
		case TokKindComment:
			ret = append(ret, &AstNode{Kind: AstNodeKindComment, Toks: toks[:1], Src: tok.Src, Lit: tok.Src})
			toks = toks[1:]
		case TokKindLitStr:
			ret = append(ret, parseLit[string](toks, AstNodeKindLit, strconv.Unquote))
			toks = toks[1:]
		case TokKindLitFloat:
			ret = append(ret, parseLit[float64](toks, AstNodeKindLit, func(src string) (float64, error) {
				return str.ToF(src, 64)
			}))
			toks = toks[1:]
		case TokKindLitRune:
			ret = append(ret, parseLit[rune](toks, AstNodeKindLit, func(src string) (ret rune, err error) {
				util.Assert(len(src) > 2 && src[0] == '\'' && src[len(src)-1] == '\'', nil)
				ret, _ = utf8.DecodeRuneInString(src[1 : len(src)-1])
				if ret == utf8.RuneError {
					err = errors.New("invalid UTF-8 encoding")
				}
				return
			}))
			toks = toks[1:]
		case TokKindLitInt:
			if tok.Src[0] == '-' {
				ret = append(ret, parseLit[int64](toks, AstNodeKindLit, func(src string) (int64, error) {
					return str.ToI64(src, 0, 64)
				}))
			} else {
				ret = append(ret, parseLit[uint64](toks, AstNodeKindLit, func(src string) (uint64, error) {
					return str.ToU64(src, 0, 64)
				}))
			}
			toks = toks[1:]
		case TokKindIdentWord, TokKindIdentOpish:
			ret = append(ret, parseLit[string](toks, AstNodeKindIdent, func(src string) (string, error) { return src, nil }))
			toks = toks[1:]
		case TokKindBracketing:
			toks_inner, toks_tail, err := toks.bracketingMatch()
			if err != nil {
				had_brace_err = true
				ret = append(ret, &AstNode{Kind: AstNodeKindErr, Toks: toks, Src: toks.src(me.Src.Text), errParsing: err})
				toks = nil
			} else {
				node := &AstNode{Kind: AstNodeKindGroup, Toks: toks[0 : len(toks_inner)+2], Lit: toks[0].Src[0]}
				node.Src = node.Toks.src(me.Src.Text)
				if len(toks_inner) > 0 {
					split_by_comma, is_curly, is_square := toks_inner.split(','), node.IsCurlyBraces(), node.IsSquareBrackets()
					if (!is_curly) && (!is_square) && ((len(toks_inner) == 0) || (len(split_by_comma) == 1)) {
						node.Nodes = me.parseNodes(toks_inner)
					} else {
						if (!is_curly) && (!is_square) {
							node.Lit = byte(',')
						}
						err_toks := node.Toks[1:2]
						for _, item_toks := range split_by_comma {
							if len(item_toks) == 0 {
								node.Nodes = append(node.Nodes, &AstNode{Kind: AstNodeKindErr, Toks: err_toks, Src: err_toks.src(me.Src.Text),
									errParsing: err_toks[util.If(is_curly, 0, len(err_toks)-1)].newErr(ErrCodeExpectedFoo, "expression before the superfluous comma")})
							} else {
								err_toks = item_toks[len(item_toks)-1:]
								if !is_curly {
									node.Nodes = append(node.Nodes, me.parseNode(item_toks))
								} else if pair := item_toks.split(':'); (len(pair) != 2) || (len(pair[0]) == 0) || (len(pair[1]) == 0) {
									node.Nodes = append(node.Nodes, &AstNode{Kind: AstNodeKindErr, Toks: err_toks, Src: err_toks.src(me.Src.Text),
										errParsing: err_toks[0].newErr(ErrCodeExpectedFoo, "expression pair separated by `:`")})
								} else {
									node_key, node_val := me.parseNode(pair[0]), me.parseNode(pair[1])
									node.Nodes = append(node.Nodes, AstNodes{node_key, node_val}.toGroupNode(me, node, true, true))
								}
							}
						}
					}
				}
				ret = append(ret, node)
				toks = toks_tail
			}
		case TokKindEnd:
			pop := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			ret = append(pop, &AstNode{Kind: AstNodeKindBlockLine, Toks: ret.toks(me), Src: ret.src(me), Nodes: ret})
			toks = toks[1:]
		case TokKindBegin:
			stack = append(stack, ret)
			ret = nil
			toks = toks[1:]
		default:
			panic(tok)
		}
	}

	if (len(stack) > 0) && !had_brace_err {
		pop := stack[len(stack)-1]
		ret_toks := util.If(len(ret) == 0, ret, pop).toks(me)
		ret = append(pop, &AstNode{Kind: AstNodeKindErr, Toks: ret_toks,
			Src: ret_toks.src(me.Src.Text), Nodes: ret, errParsing: ret_toks[0].newIndentErr()})
	}

	return
}

func parseLit[T cmp.Ordered](toks Toks, kind AstNodeKind, parseFunc func(string) (T, error)) *AstNode {
	tok := toks[0]
	lit, err := parseFunc(tok.Src)
	if err != nil {
		return &AstNode{Kind: AstNodeKindErr, Toks: toks[:1], Src: tok.Src, errParsing: errToDiag(err, ErrCodeLitWontParse, tok.span())}
	}
	return &AstNode{Kind: kind, Toks: toks[:1], Src: tok.Src, Lit: lit}
}

func (me *SrcFile) NodeAtPos(pos SrcFilePos, orAncestor bool) (ret *AstNode) {
	for _, node := range me.Src.Ast {
		if node.Toks.Span().Contains(&pos) {
			ret = node.find(func(it *AstNode) bool {
				return (len(it.Nodes) == 0) && it.Toks.Span().Contains(&pos)
			})
			if ret == nil && orAncestor {
				ret = node
			}
			break
		}
	}
	return
}

func (me *SrcFile) NodeAtSpan(span *SrcFileSpan) (ret *AstNode) {
	for _, node := range me.Src.Ast {
		if node_span := node.Toks.Span(); node_span.Contains(&span.Start) || node_span.Contains(&span.End) {
			if ret = node.find(func(it *AstNode) bool {
				return it.Toks.Span() == *span
			}); ret != nil {
				break
			}
		}
	}
	return
}

func (me *AstNode) canHuddle() bool {
	if me.Kind == AstNodeKindIdent {
		return !me.Toks[0].isSep()
	}
	return (me.Kind == AstNodeKindLit) || (me.Kind == AstNodeKindGroup)
}

func (me *AstNode) Cmp(it *AstNode) int {
	return cmp.Compare(me.Toks[0].byteOffset, it.Toks[0].byteOffset)
}

func (me *AstNode) equals(it *AstNode, includingSpans bool, withoutComments bool) bool {
	util.Assert(me != it, nil)

	if (me.Kind != it.Kind) ||
		(includingSpans && !me.Toks.Span().eq(it.Toks.Span())) ||
		(!me.Nodes.equals(it.Nodes, includingSpans, withoutComments)) {
		return false
	}

	switch me.Kind {
	case AstNodeKindGroup, AstNodeKindBlockLine:
		return (me.Lit == it.Lit) // covers parens,brackets,braces
	case AstNodeKindLit:
		switch mine := me.Lit.(type) {
		case float64:
			other, ok := it.Lit.(float64)
			return ok && (mine == other)
		case int64:
			other, ok := it.Lit.(int64)
			return ok && (mine == other)
		case uint64:
			other, ok := it.Lit.(uint64)
			return ok && (mine == other)
		case rune:
			other, ok := it.Lit.(rune)
			return ok && (mine == other)
		case string:
			other, ok := it.Lit.(string)
			return ok && (mine == other)
		default:
			panic(me.Lit)
		}
	case AstNodeKindIdent:
		return (me.Lit.(string) == it.Lit.(string))
	case AstNodeKindErr:
		return me.errParsing.equals(it.errParsing, includingSpans)
	default:
		panic(me.Kind)
	}
}

func (me *AstNode) find(where func(node *AstNode) bool) (ret *AstNode) {
	me.walk(func(node *AstNode) bool {
		if ret == nil && where(node) {
			ret = node
		}
		return (ret == nil)
	}, nil)
	return
}

func (me *AstNode) ident() string {
	return util.If(me.Kind == AstNodeKindIdent, me.Src, "")
}

func (me *AstNode) IsBracketingWith(opener byte) bool {
	if me.Kind == AstNodeKindGroup {
		return (me.Lit == opener)
	}
	return false
}
func (me *AstNode) IsCurlyBraces() bool    { return me.IsBracketingWith('{') }
func (me *AstNode) IsSquareBrackets() bool { return me.IsBracketingWith('[') }
func (me *AstNode) IsParensCallish() bool  { return me.IsBracketingWith('(') }
func (me *AstNode) IsParensTuplish() bool  { return me.IsBracketingWith(',') }

func (me *AstNode) IsIdentOpish() bool {
	return (me.Kind == AstNodeKindIdent) && (me.Toks[0].Kind == TokKindIdentOpish)
}
func (me *AstNode) IsIdentSepish() bool { return (me.Kind == AstNodeKindIdent) && me.Toks[0].isSep() }
func (me *AstNode) IsIdentPrim() bool {
	return (me.Kind == AstNodeKindIdent) && (len(me.Src) > 1) && (me.Src[0] == '@')
}
func (me *AstNode) IsIdentKeyword() bool {
	return (me.Kind == AstNodeKindIdent) && (len(me.Src) > 1) && (me.Src[0] == ':')
}

func (me *AstNode) isWhitespacelesslyRightAfter(it *AstNode) bool {
	return me.Toks[0].isWhitespacelesslyRightAfter(it.Toks[len(it.Toks)-1])
}

func (me *AstNode) newDiag(kind DiagKind, atEnd bool, code DiagCode, args ...any) *Diag {
	return me.Toks.newDiag(kind, atEnd, code, args...)
}
func (me *AstNode) newDiagInfo(atEnd bool, code DiagCode, args ...any) *Diag {
	return me.newDiag(DiagKindInfo, atEnd, code, args...)
}
func (me *AstNode) newDiagHint(atEnd bool, code DiagCode, args ...any) *Diag {
	return me.newDiag(DiagKindHint, atEnd, code, args...)
}
func (me *AstNode) newDiagWarn(atEnd bool, code DiagCode, args ...any) *Diag {
	return me.newDiag(DiagKindWarn, atEnd, code, args...)
}
func (me *AstNode) newDiagErr(atEnd bool, code DiagCode, args ...any) *Diag {
	return me.newDiag(DiagKindErr, atEnd, code, args...)
}

func (me *AstNode) SelfAndAncestors() (ret AstNodes) {
	for it := me; it != nil; it = it.parent {
		ret = append(ret, it)
	}
	return
}

func (me *AstNode) walk(onBefore func(node *AstNode) bool, onAfter func(node *AstNode)) {
	if onBefore != nil && !onBefore(me) {
		return
	}
	for _, node := range me.Nodes {
		node.walk(onBefore, onAfter)
	}
	if onAfter != nil {
		onAfter(me)
	}
}

func (me AstNodes) AnyErrs() (ret bool) {
	me.walk(func(node *AstNode) bool {
		ret = ret || (node.Kind == AstNodeKindErr)
		return !ret
	}, nil)
	return
}

func (me AstNodes) equals(it AstNodes, includingSpans bool, withoutComments bool) bool {
	if withoutComments {
		me, it = me.withoutComments(), it.withoutComments()
	}
	return sl.Eq(me, it, func(node1 *AstNode, node2 *AstNode) bool {
		return node1.equals(node2, includingSpans, withoutComments)
	})
}

func (me AstNodes) first() *AstNode { return me[0] }

func (me AstNodes) has(recurse bool, where func(node *AstNode) bool) (ret bool) {
	if !recurse {
		ret = sl.Any(me, where)
	} else {
		me.walk(func(it *AstNode) bool {
			ret = ret || where(it)
			return !ret
		}, nil)
	}
	return
}

func (me AstNodes) hasKind(kind AstNodeKind) bool {
	return me.has(true, func(it *AstNode) bool { return it.Kind == kind })
}

func (me AstNodes) huddled(srcFile *SrcFile, parent *AstNode) (ret AstNodes) {
	if len(me) <= 1 {
		return me
	}
	all_huddled, huddle := true, AstNodes{me[0]}
	for i := 1; i < len(me); i++ {
		prev, cur := me[i-1], me[i]
		if prev.canHuddle() && cur.canHuddle() && cur.isWhitespacelesslyRightAfter(prev) {
			huddle = append(huddle, cur)
		} else {
			all_huddled = false
			ret = append(ret, huddle.toGroupNode(srcFile, parent, true, true))
			huddle = AstNodes{cur}
		}
	}
	if all_huddled {
		ret = me
	} else {
		ret = append(ret, huddle.toGroupNode(srcFile, parent, true, true))
	}
	return
}

func (me AstNodes) last() *AstNode { return me[len(me)-1] }

func (me AstNodes) newDiag(srcFile *SrcFile, kind DiagKind, code DiagCode, args ...any) *Diag {
	return &Diag{Kind: kind, Code: code, Span: me.toks(srcFile).Span(), Message: errMsg(code, args...)}
}
func (me AstNodes) newDiagInfo(srcFile *SrcFile, code DiagCode, args ...any) *Diag {
	return me.newDiag(srcFile, DiagKindInfo, code, args...)
}
func (me AstNodes) newDiagHint(srcFile *SrcFile, code DiagCode, args ...any) *Diag {
	return me.newDiag(srcFile, DiagKindHint, code, args...)
}
func (me AstNodes) newDiagWarn(srcFile *SrcFile, code DiagCode, args ...any) *Diag {
	return me.newDiag(srcFile, DiagKindWarn, code, args...)
}
func (me AstNodes) newDiagErr(srcFile *SrcFile, code DiagCode, args ...any) *Diag {
	return me.newDiag(srcFile, DiagKindErr, code, args...)
}

func (me AstNodes) src(srcFile *SrcFile) string {
	return me.toks(srcFile).src(srcFile.Src.Text)
}

func (me AstNodes) toGroupNode(srcFile *SrcFile, parent *AstNode, onlyIfMultiple bool, nilIfEmpty bool) *AstNode {
	if nilIfEmpty && (len(me) == 0) {
		return nil
	} else if onlyIfMultiple && (len(me) == 1) {
		return me[0]
	}
	return &AstNode{Kind: AstNodeKindGroup, Toks: me.toks(srcFile), parent: parent, Nodes: me, Src: me.src(srcFile)}
}

func (me AstNodes) toks(srcFile *SrcFile) Toks {
	if len(me) == 0 {
		return nil
	}
	node_first, node_last := me[0], me[len(me)-1]
	tok_first, tok_last := node_first.Toks[0], node_last.Toks[len(node_last.Toks)-1]
	idx_first := -1
	for i, tok := range srcFile.Src.Toks {
		if tok == tok_first {
			idx_first = i
		}
		if (tok == tok_last) && (idx_first >= 0) {
			return srcFile.Src.Toks[idx_first : i+1]
		}
	}
	panic(str.Fmt("%d %d >>%s<< >>%s<<", idx_first, len(me), node_first.Src, node_last.Src))
}

func (me AstNodes) walk(onBefore func(node *AstNode) bool, onAfter func(node *AstNode)) {
	for _, node := range me {
		node.walk(onBefore, onAfter)
	}
}

func (me AstNodes) withoutComments() AstNodes {
	return sl.Where(me, func(it *AstNode) bool { return it.Kind != AstNodeKindComment })
}

// any basic syntax errs from the lexing or parsing stages preclude a `SrcPack.treesRefresh` (the prior trees are kept)
func (me *SrcFile) HasMoPrecludingErrs() bool {
	return (me.diags.LastReadErr != nil) || (len(me.diags.LexErrs) > 0) || me.Src.Ast.AnyErrs()
}
