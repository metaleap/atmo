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
	AstNodeKindErr      AstNodeKind = iota
	AstNodeKindComment              // both /* multi-line */ and // single-line
	AstNodeKindIdent                // foo, #bar, @baz, $foo, %bar, ==, <==<
	AstNodeKindLit                  // 123, -321, 1.23, -3.21, "foo", `bar`, 'ö'
	AstNodeKindMultiple             // foo bar, (), (foo), (foo bar), [], [foo], [foo bar], {}, {foo}, {foo bar}
)

// only called by EnsureSrcFile, just after tokenization, with `.Notices.LexErrs` freshly set.
// mutates me.Content.TopLevelAstNodes and me.Notices.ParseErrs.
func (me *SrcFile) parse() {
	parsed := me.parseNodes(me.Content.Toks, true)

	// multi-line call-forms in parens: make each subsequent multi-tok line its own call-form
	parsed.walk(nil, func(node *AstNode) {
		// if node.Kind == AstNodeKindMultiple && node.isParens() && node.Toks.isMultiLine() {
		// 	lines := node.ChildNodes.splitByLines()
		// 	node.ChildNodes = lines[0]

		// 	for _, line := range lines[1:] {
		// 		if len(line) == 1 {
		// 			node.ChildNodes = append(node.ChildNodes, line[0])
		// 		} else {
		// 			call_node := &AstNode{Kind: AstNodeKindMultiple, Src: line.toks().src(me.Content.Src),
		// 				Toks: line.toks(), ChildNodes: line}
		// 			node.ChildNodes = append(node.ChildNodes, call_node)
		// 		}
		// 	}
		// }
	})

	// rewrite all call-forms with an infix operator: `foo bar · baz mojo + times 10` => `(· (foo bar) (+ (baz mojo) (times 10)))`
	// that is: everything to its left is its lhs expr, everything to its right is its rhs expr.
	parsed.walk(nil, func(node *AstNode) {
		// if node.Kind == AstNodeKindMultiple {
		// 	idx := 1 + sl.IdxWhere(node.ChildNodes[1:], (*AstNode).isIdentOpish)
		// 	if idx > 0 {
		// 		op, lhs, rhs := node.ChildNodes[idx], node.ChildNodes[:idx], node.ChildNodes[idx+1:]
		// 		// println(">>>>>>>>>INFIXX>>>>>>>>>>>>>>" + op.Src + "<<<<<<<<<<<<<<<<<<INSIDE>>>>>>>" + node.Src + "<<<<<<<<<<<<<<<<<")
		// 		lhs = AstNodes{{Kind: AstNodeKindMultiple, Src: lhs.toks().src(me.Content.Src),
		// 			Toks: lhs.toks(), ChildNodes: lhs}}
		// 		rhs = AstNodes{{Kind: AstNodeKindMultiple, Src: rhs.toks().src(me.Content.Src),
		// 			Toks: rhs.toks(), ChildNodes: rhs}}
		// 		node.ChildNodes = AstNodes{op, lhs[0], rhs[0]}
		// 	}
		// }
	})

	// sort all top-level nodes to be in source-file order of appearance; also set all `AstNode.parent`s
	parsed = sl.SortedPer(parsed, (*AstNode).cmp)
	parsed.walk(nil, func(node *AstNode) {
		for _, it := range node.ChildNodes {
			it.parent = node
		}
	})
	me.Content.Ast = parsed
}

func (me *SrcFile) parseNode(toks Toks, checkForHuddle bool) *AstNode {
	nodes := me.parseNodes(toks, checkForHuddle)
	if len(nodes) == 1 {
		return nodes[0]
	}
	return &AstNode{Kind: AstNodeKindMultiple, ChildNodes: nodes, Toks: toks, Src: toks.src(me.Content.Src)}
}

func (me *SrcFile) parseNodes(toks Toks, checkForHuddle bool) (ret AstNodes) {
	for len(toks) > 0 {
		if checkForHuddle {
			if huddled, rest := toks.huddle(); len(huddled) > 1 && ((len(rest) > 0) || (len(ret) > 0)) {
				ret = append(ret, me.parseNode(huddled, false))
				toks = rest
				continue
			}
		}

		tok := toks[0]
		switch tok.Kind {
		case TokKindNewLine, TokKindIndent, TokKindDedent:
			toks = toks[1:]
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
				ret = append(ret, &AstNode{Kind: AstNodeKindErr, Toks: toks, Src: toks.src(me.Content.Src), err: err})
				toks = nil
			} else {
				node := &AstNode{
					Kind: AstNodeKindMultiple, Toks: toks[0 : len(toks_inner)+2],
					ChildNodes: me.parseNodes(toks_inner, checkForHuddle),
				}
				node.Src = node.Toks.src(me.Content.Src) // .. and for Src to reflect that SrcFileSpan fully
				ret = append(ret, node)
				toks = toks_tail
			}
		default:
			panic(tok)
		}
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

func (me *AstNode) cmp(it *AstNode) int {
	return cmp.Compare(me.Toks[0].byteOffset, it.Toks[0].byteOffset)
}

func (me *AstNode) equals(it *AstNode) bool {
	util.Assert(me != it, nil)

	if me.Kind != it.Kind || len(me.ChildNodes) != len(it.ChildNodes) {
		return false
	}

	var idx int
	if !sl.All(me.ChildNodes, func(node *AstNode) (isEq bool) {
		isEq, idx = node.equals(it.ChildNodes[idx]), idx+1
		return
	}) {
		return false
	}

	switch me.Kind {
	case AstNodeKindComment:
		return true
	case AstNodeKindMultiple:
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
	return me.Kind == AstNodeKindIdent && me.Toks[0].Kind == TokKindIdentOpish
}

func (me *AstNode) isParens() bool {
	return me.Src[0] == '('
}

func (me *AstNode) sig(buf *strings.Builder) {
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

func (me AstNodes) has(recurse bool, where func(*AstNode) bool) (ret bool) {
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

func (me AstNodes) hasKind(kind AstNodeKind) (ret bool) {
	me.walk(func(it *AstNode) bool {
		ret = ret || (it.Kind == kind)
		return !ret
	}, nil)
	return
}

func (me AstNodes) splitByLines() (ret []AstNodes) {
	cur_line := me[0].Toks[0].Pos.Line
	var cur AstNodes
	for _, node := range me {
		span := node.Toks.Span()
		if span.Start.Line == cur_line {
			cur = append(cur, node)
		} else if len(cur) > 0 {
			ret = append(ret, cur)
			cur = AstNodes{node}
		}
		cur_line = span.End.Line
	}
	if len(cur) > 0 {
		ret = append(ret, cur)
	}
	return
}

func (me AstNodes) toks() (ret Toks) {
	for _, node := range me {
		ret = append(ret, node.Toks...)
	}
	return
}

func (me AstNodes) walk(onBefore func(*AstNode) bool, onAfter func(*AstNode)) {
	for _, node := range me {
		node.walk(onBefore, onAfter)
	}
}
