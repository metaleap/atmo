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
	parent     *AstNode
	Kind       AstNodeKind
	Src        string
	Toks       Toks
	ChildNodes AstNodes
	err        *SrcFileNotice
	Lit        any // if AstNodeKindIdent or AstNodeKindLit, one of: float64 | int64 | uint64 | rune | string
	Ann        any
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
func (me *SrcFile) parse() {
	parsed := me.parseNodes(me.Content.Toks)

	// group huddled exprs: `foo x+z y` right now is `foo x + z y` BUT lets make it `foo (x + 1) y`:
	parsed.walk(nil, func(node *AstNode) {
		node.ChildNodes = node.ChildNodes.huddled(me.Content.Src)
	})

	// only now, treat parens non-listish, unlike braces/brackets: hoist any 1-element parens-groups so that `(x)` becomes `x` and `(((foo bar)))` becomes `foo bar`
	parsed.walk(nil, func(node *AstNode) {
		for i, it := range node.ChildNodes {
			if it.isParens() && len(it.ChildNodes) == 1 {
				node.ChildNodes[i] = it.ChildNodes[0]
			}
		}
	})

	// rewrite nodes with an opish: everything to its left becomes its lhs expr, everything to its right becomes its rhs expr.
	parsed.walk(func(node *AstNode) bool {
		if (node.Kind == AstNodeKindGroup) && (len(node.ChildNodes) > 1) {
			if idx := sl.IdxWhere(node.ChildNodes, (*AstNode).isIdentOpish); idx >= 0 {
				op, lhs, rhs := node.ChildNodes[idx], node.ChildNodes[:idx], node.ChildNodes[idx+1:]
				node.ChildNodes = AstNodes{op, lhs.group(false, false, me.Content.Src), rhs.group(false, false, me.Content.Src)}
			}
		}
		return true
	}, nil)

	// sort all top-level nodes to be in source-file order of appearance; also set all `AstNode.parent`s
	parsed = sl.SortedPer(parsed, (*AstNode).cmp)
	parsed.walk(nil, func(node *AstNode) {
		for _, it := range node.ChildNodes {
			it.parent = node
		}
	})
	me.Content.Ast = parsed
}

func (me *SrcFile) parseNode(toks Toks) *AstNode {
	nodes := me.parseNodes(toks)
	if len(nodes) == 1 {
		return nodes[0]
	}
	return &AstNode{Kind: AstNodeKindGroup, ChildNodes: nodes, Toks: toks, Src: toks.src(me.Content.Src)}
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
				ret = append(ret, &AstNode{Kind: AstNodeKindErr, Toks: toks, Src: toks.src(me.Content.Src), err: err})
				toks = nil
			} else {
				node := &AstNode{
					Kind: AstNodeKindGroup, Toks: toks[0 : len(toks_inner)+2],
					ChildNodes: me.parseNodes(toks_inner),
				}
				node.Src = node.Toks.src(me.Content.Src) // .. and for Src to reflect that SrcFileSpan fully
				ret = append(ret, node)
				toks = toks_tail
			}
		case TokKindEnd:
			pop := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			ret = append(pop, &AstNode{Kind: AstNodeKindGroup, Toks: ret.toks(),
				Src: ret.toks().src(me.Content.Src), ChildNodes: ret})
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
		ret_toks := util.If(len(ret) == 0, ret, pop).toks()
		ret = append(pop, &AstNode{Kind: AstNodeKindErr, Toks: ret_toks,
			Src: ret_toks.src(me.Content.Src), ChildNodes: ret, err: ret_toks[0].newIndentErr()})
	}

	return
}

func parseLit[T cmp.Ordered](toks Toks, kind AstNodeKind, parseFunc func(string) (T, error)) *AstNode {
	tok := toks[0]
	lit, err := parseFunc(tok.Src)
	if err != nil {
		return &AstNode{Kind: AstNodeKindErr, Toks: toks[:1], Src: tok.Src, err: errToNotice(err, NoticeCodeLitSyntax, util.Ptr(tok.span()))}
	}
	return &AstNode{Kind: kind, Toks: toks[:1], Src: tok.Src, Lit: lit}
}

func (me *SrcFile) NodeAt(pos SrcFilePos, orAncestor bool) (ret *AstNode) {
	for _, node := range me.Content.Ast {
		if node.Toks.Span().contains(&pos) {
			ret = node.find(func(it *AstNode) bool {
				return (len(it.ChildNodes) == 0) && it.Toks.Span().contains(&pos)
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

	if me.Kind != it.Kind || !me.ChildNodes.equals(it.ChildNodes, withoutComments) {
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
		return (me.err == it.err) || ((me.err != nil) && (it.err != nil) &&
			(me.err.Code == it.err.Code) && (me.err.Kind == it.err.Kind) && (me.err.Message == it.err.Message))
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
	prev_tok := it.Toks[len(it.Toks)-1]
	return me.Toks[0].byteOffset == (prev_tok.byteOffset + len(prev_tok.Src))
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
	for _, it := range me.ChildNodes {
		it.sig(buf)
	}
	buf.WriteByte('>')
}

func (me *AstNode) Sig() string {
	var buf strings.Builder
	if me.Kind != AstNodeKindErr && !me.ChildNodes.hasKind(AstNodeKindErr) {
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
	for _, node := range me.ChildNodes {
		node.walk(onBefore, onAfter)
	}
	if onAfter != nil {
		onAfter(me)
	}
}

func (me AstNodes) equals(it AstNodes, withoutComments bool) bool {
	var idx int
	if withoutComments {
		me, it = me.withoutComments(), it.withoutComments()
	}
	return (len(me) == len(it)) && sl.All(me, func(node *AstNode) bool {
		it := it[idx]
		idx++
		return node.equals(it, withoutComments)
	})
}

func (me AstNodes) group(onlyIfMultiple bool, nilIfEmpty bool, curFullSrcFileContent string) *AstNode {
	if nilIfEmpty && (len(me) == 0) {
		return nil
	} else if onlyIfMultiple && (len(me) == 1) {
		return me[0]
	}
	return &AstNode{Kind: AstNodeKindGroup, Toks: me.toks(),
		ChildNodes: me, Src: me.toks().src(curFullSrcFileContent)}
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

func (me AstNodes) huddled(curFullSrcFileContent string) (ret AstNodes) {
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
			ret = append(ret, huddle.group(true, true, curFullSrcFileContent))
			huddle = AstNodes{cur}
		}
	}
	if all_huddled {
		ret = me
	} else {
		ret = append(ret, huddle.group(true, true, curFullSrcFileContent))
	}
	return
}

func (me AstNodes) toks() (ret Toks) {
	for _, node := range me {
		ret = append(ret, node.Toks...)
	}
	return
}

func (me AstNodes) walk(onBefore func(node *AstNode) bool, onAfter func(node *AstNode)) {
	for _, node := range me {
		node.walk(onBefore, onAfter)
	}
}

func (me AstNodes) withoutComments() AstNodes {
	return sl.Where(me, func(it *AstNode) bool { return it.Kind != AstNodeKindComment })
}
