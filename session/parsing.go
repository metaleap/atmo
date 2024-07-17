package session

import (
	"cmp"
	"errors"
	"strconv"
	"strings"
	"unicode/utf8"

	"atmo/util"
	"atmo/util/sl"
	"atmo/util/str"
)

type AstNodes []*AstNode

type AstNode struct {
	parent        *AstNode
	Kind          AstNodeKind
	Src           string
	Toks          Toks
	Nodes         AstNodes `json:",omitempty"`
	errParsing    *SrcFileNotice
	errsExpansion SrcFileNotices
	Lit           any `json:",omitempty"` // if AstNodeKindIdent or AstNodeKindLit, one of: float64 | int64 | uint64 | rune | string
}

type AstNodeKind int

const (
	AstNodeKindErr     AstNodeKind = iota
	AstNodeKindComment             // both /* multi-line */ and // single-line
	AstNodeKindIdent               // foo, #bar, @baz, $foo, %bar, ==, <==<
	AstNodeKindLit                 // 123, -321, 1.23, -3.21, "foo", `bar`, 'ö'
	AstNodeKindGroup               // foo bar, (), (foo), (foo bar), [], [foo], [foo bar], {}, {foo}, {foo bar}
)

// only called by EnsureSrcFile, just after tokenization, with `.Notices.LexErrs` freshly set.
// mutates me.Content.TopLevelAstNodes and me.Notices.ParseErrs.
func (me *SrcFile) parse() AstNodes {
	parsed := me.parseNodes(me.Content.Toks)

	// group huddled exprs: `foo x+z y` right now is `foo x + z y` BUT lets make it `foo (x + 1) y`:
	parsed.walk(nil, func(node *AstNode) {
		node.Nodes = node.Nodes.huddled(me)
	})

	// only now, treat parens non-listish, unlike braces/brackets: hoist any 1-element parens-groups so that `(x)` becomes `x` and `(((foo bar)))` becomes `(foo bar)`
	parsed.walk(nil, func(node *AstNode) {
		for i, it := range node.Nodes {
			if it.isParens() && len(it.Nodes) == 1 {
				node.Nodes[i] = it.Nodes[0]
			}
		}
	})

	// sort all top-level nodes to be in source-file order of appearance; also set all `AstNode.parent`s
	parsed = sl.SortedPer(parsed, (*AstNode).cmp)
	parsed.walk(nil, func(node *AstNode) {
		for _, it := range node.Nodes {
			it.parent = node
		}
	})
	return parsed
}

func (me *SrcFile) parseNode(toks Toks) *AstNode {
	nodes := me.parseNodes(toks)
	if len(nodes) == 1 {
		return nodes[0]
	}
	return &AstNode{Kind: AstNodeKindGroup, Nodes: nodes, Toks: toks, Src: toks.src(me.Content.Src)}
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
		case TokKindBrace:
			toks_inner, toks_tail, err := toks.braceMatch()
			if err != nil {
				had_brace_err = true
				ret = append(ret, &AstNode{Kind: AstNodeKindErr, Toks: toks, Src: toks.src(me.Content.Src), errParsing: err})
				toks = nil
			} else {
				node := &AstNode{
					Kind: AstNodeKindGroup, Toks: toks[0 : len(toks_inner)+2],
					Nodes: me.parseNodes(toks_inner),
				}
				node.Src = node.Toks.src(me.Content.Src) // .. and for Src to reflect that SrcFileSpan fully
				ret = append(ret, node)
				toks = toks_tail
			}
		case TokKindEnd:
			pop := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			ret = append(pop, &AstNode{Kind: AstNodeKindGroup, Toks: ret.toks(me),
				Src: ret.src(me), Nodes: ret})
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
			Src: ret_toks.src(me.Content.Src), Nodes: ret, errParsing: ret_toks[0].newIndentErr()})
	}

	return
}

func parseLit[T cmp.Ordered](toks Toks, kind AstNodeKind, parseFunc func(string) (T, error)) *AstNode {
	tok := toks[0]
	lit, err := parseFunc(tok.Src)
	if err != nil {
		return &AstNode{Kind: AstNodeKindErr, Toks: toks[:1], Src: tok.Src, errParsing: errToNotice(err, NoticeCodeLitSyntax, tok.span())}
	}
	return &AstNode{Kind: kind, Toks: toks[:1], Src: tok.Src, Lit: lit}
}

func (me *SrcFile) NodeAt(pos SrcFilePos, orAncestor bool) (ret *AstNode) {
	for _, node := range me.Content.Ast {
		if node.Toks.Span().contains(&pos) {
			ret = node.find(func(it *AstNode) bool {
				return (len(it.Nodes) == 0) && it.Toks.Span().contains(&pos)
			})
			if ret == nil && orAncestor {
				ret = node
			}
			break
		}
	}
	return
}

func (me *AstNode) canHuddle() bool {
	return (me.Kind == AstNodeKindLit) || (me.Kind == AstNodeKindIdent)
}

func (me *AstNode) cmp(it *AstNode) int {
	return cmp.Compare(me.Toks[0].byteOffset, it.Toks[0].byteOffset)
}

func (me *AstNode) equals(it *AstNode, withoutComments bool) bool {
	util.Assert(me != it, nil)

	if me.Kind != it.Kind || (!me.Nodes.equals(it.Nodes, withoutComments)) ||
		!sl.Eq(me.errsExpansion, it.errsExpansion, (*SrcFileNotice).equals) {
		return false
	}

	switch me.Kind {
	case AstNodeKindGroup:
		return (me.Src[0] == it.Src[0]) // covers parens,brackets,braces
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
		return me.errParsing.equals(it.errParsing)
	default:
		panic(me.Kind)
	}
}

func (me *AstNode) find(where func(*AstNode) bool) (ret *AstNode) {
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

func (me *AstNode) isBraces() bool {
	return me.Src[0] == '{'
}

func (me *AstNode) isBrackets() bool {
	return me.Src[0] == '['
}

func (me *AstNode) isIdentOpish() bool {
	return (me.Kind == AstNodeKindIdent) && (me.Toks[0].Kind == TokKindIdentOpish)
}

func (me *AstNode) isParens() bool {
	return me.Src[0] == '('
}

func (me *AstNode) isWhitespacelesslyRightAfter(it *AstNode) bool {
	return me.Toks[0].isWhitespacelesslyRightAfter(it.Toks[len(it.Toks)-1])
}

func (me *AstNode) newDiag(kind SrcFileNoticeKind, atEnd bool, code SrcFileNoticeCode, args ...any) *SrcFileNotice {
	return &SrcFileNotice{Kind: kind, Code: code, Span: util.If(atEnd, Toks.SpanEnd, Toks.Span)(me.Toks), Message: errMsg(code, args...)}
}
func (me *AstNode) newDiagInfo(atEnd bool, code SrcFileNoticeCode, args ...any) *SrcFileNotice {
	return me.newDiag(NoticeKindInfo, atEnd, code, args...)
}
func (me *AstNode) newDiagHint(atEnd bool, code SrcFileNoticeCode, args ...any) *SrcFileNotice {
	return me.newDiag(NoticeKindHint, atEnd, code, args...)
}
func (me *AstNode) newDiagWarn(atEnd bool, code SrcFileNoticeCode, args ...any) *SrcFileNotice {
	return me.newDiag(NoticeKindWarn, atEnd, code, args...)
}
func (me *AstNode) newDiagErr(atEnd bool, code SrcFileNoticeCode, args ...any) *SrcFileNotice {
	return me.newDiag(NoticeKindErr, atEnd, code, args...)
}

func (me *AstNode) sig(buf *strings.Builder) {
	if me.Kind == AstNodeKindComment {
		return
	}
	buf.WriteByte('<')
	buf.WriteString(str.FromInt(int(me.Kind)))
	buf.WriteByte(',')
	switch me.Kind {
	case AstNodeKindIdent:
		buf.WriteString(me.Lit.(string))
	case AstNodeKindLit:
		switch lit := me.Lit.(type) {
		case float64:
			buf.WriteString(str.FromFloat(lit, -1))
		case int64:
			buf.WriteString(str.FromI64(lit, 36))
		case uint64:
			buf.WriteString(str.FromU64(lit, 36))
		case rune:
			buf.WriteString(strconv.QuoteRune(lit))
		case string:
			buf.WriteString(str.Q(lit))
		default:
			panic(me.Lit)
		}
	}
	buf.WriteByte(',')
	for _, it := range me.Nodes {
		it.sig(buf)
	}
	buf.WriteByte('>')
}

func (me *AstNode) Sig() string {
	var buf strings.Builder
	if me.Kind != AstNodeKindErr && !me.Nodes.hasKind(AstNodeKindErr) {
		me.sig(&buf)
	}
	return buf.String()
}

func (me *AstNode) SelfAndAncestors() (ret AstNodes) {
	for it := me; it != nil; it = it.parent {
		ret = append(ret, it)
	}
	return
}

func (me *AstNode) walk(onBefore func(*AstNode) bool, onAfter func(*AstNode)) {
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

func (me AstNodes) equals(it AstNodes, withoutComments bool) bool {
	if withoutComments {
		me, it = me.withoutComments(), it.withoutComments()
	}
	return sl.Eq(me, it, func(node1 *AstNode, node2 *AstNode) bool {
		return node1.equals(node2, withoutComments)
	})
}

func (me AstNodes) first() *AstNode { return me[0] }

func (me AstNodes) group(srcFile *SrcFile, onlyIfMultiple bool, nilIfEmpty bool) *AstNode {
	if nilIfEmpty && (len(me) == 0) {
		return nil
	} else if onlyIfMultiple && (len(me) == 1) {
		return me[0]
	}
	return &AstNode{Kind: AstNodeKindGroup, Toks: me.toks(srcFile),
		Nodes: me, Src: me.src(srcFile)}
}

func (me AstNodes) has(recurse bool, where func(node *AstNode) bool) (ret bool) {
	if !recurse {
		ret = sl.HasWhere(me, where)
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

func (me AstNodes) huddled(srcFile *SrcFile) (ret AstNodes) {
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
			ret = append(ret, huddle.group(srcFile, true, true))
			huddle = AstNodes{cur}
		}
	}
	if all_huddled {
		ret = me
	} else {
		ret = append(ret, huddle.group(srcFile, true, true))
	}
	return
}

func (me AstNodes) last() *AstNode { return me[len(me)-1] }

func (me AstNodes) newDiag(srcFile *SrcFile, kind SrcFileNoticeKind, code SrcFileNoticeCode, args ...any) *SrcFileNotice {
	return &SrcFileNotice{Kind: kind, Code: code, Span: me.toks(srcFile).Span(), Message: errMsg(code, args...)}
}
func (me AstNodes) newDiagInfo(srcFile *SrcFile, code SrcFileNoticeCode, args ...any) *SrcFileNotice {
	return me.newDiag(srcFile, NoticeKindInfo, code, args...)
}
func (me AstNodes) newDiagHint(srcFile *SrcFile, code SrcFileNoticeCode, args ...any) *SrcFileNotice {
	return me.newDiag(srcFile, NoticeKindHint, code, args...)
}
func (me AstNodes) newDiagWarn(srcFile *SrcFile, code SrcFileNoticeCode, args ...any) *SrcFileNotice {
	return me.newDiag(srcFile, NoticeKindWarn, code, args...)
}
func (me AstNodes) newDiagErr(srcFile *SrcFile, code SrcFileNoticeCode, args ...any) *SrcFileNotice {
	return me.newDiag(srcFile, NoticeKindErr, code, args...)
}

func (me AstNodes) src(srcFile *SrcFile) string {
	return me.toks(srcFile).src(srcFile.Content.Src)
}

func (me AstNodes) toks(srcFile *SrcFile) Toks {
	if len(me) == 0 {
		return nil
	}
	node_first, node_last := me[0], me[len(me)-1]
	tok_first, tok_last := node_first.Toks[0], node_last.Toks[len(node_last.Toks)-1]
	idx_first := -1
	for i, tok := range srcFile.Content.Toks {
		if tok == tok_first {
			idx_first = i
		}
		if (tok == tok_last) && (idx_first >= 0) {
			return srcFile.Content.Toks[idx_first : i+1]
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
