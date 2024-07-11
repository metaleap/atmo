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
	AstNodeKindErr            AstNodeKind = iota
	AstNodeKindCallForm                   // foo bar baz
	AstNodeKindCurlyBraces                // {}
	AstNodeKindSquareBrackets             // []
	AstNodeKindIdent                      // foo, #bar, @baz, $foo, %bar
	AstNodeKindLit                        // 123, -321, 1.23, -3.21, "foo", `bar`, 'ö'
	AstNodeKindComment
)

// only called by EnsureSrcFile, just after tokenization, with `.Notices.LexErrs` freshly set.
// mutates me.Content.TopLevelAstNodes and me.Notices.ParseErrs.
func (me *SrcFile) parse() {
	parsed := me.parseNodes(me.Content.Toks, true)

	// // flip infix-operator call forms to prefix form
	// parsed.walk(nil, func(node *AstNode) {
	// 	if node.Kind == AstNodeKindCallForm && (len(node.ChildNodes) > 2) && node.ChildNodes.has(false, (*AstNode).isIdentOp) {
	// 		for i, it := range node.ChildNodes {
	// 			if must, is := ((i % 2) != 0), it.isIdentOp(); must != is {
	// 				node.Kind = AstNodeKindErr
	// 				node.err = &SrcFileNotice{Kind: NoticeKindErr, Code: NoticeCodeBadInfixExpr, Span: it.Toks.Span(),
	// 					Message: util.If(must, "operator", "non-operator") + " expected due to infix expression"}
	// 				break
	// 			}
	// 		}
	// 	}
	// })

	// sort all nodes to be in source-file order of appearance
	parsed.walk(nil, func(node *AstNode) {
		node.ChildNodes = sl.SortedPer(node.ChildNodes, func(node1 *AstNode, node2 *AstNode) int {
			return cmp.Compare(node1.Toks[0].byteOffset, node2.Toks[0].byteOffset)
		})
	})
	me.Content.Ast = parsed
}

func (me *SrcFile) parseNode(toks Toks, checkForThrong bool) *AstNode {
	nodes := me.parseNodes(toks, checkForThrong)
	if len(nodes) == 1 {
		return nodes[0]
	}
	return &AstNode{Kind: AstNodeKindCallForm, ChildNodes: nodes, Toks: toks, Src: toks.src(me.Content.Src)}
}

func (me *SrcFile) parseNodes(toks Toks, checkForThrong bool) (ret AstNodes) {
	// pos_line, pos_char := toks[0].Pos.Line, toks[0].Pos.Char
	for len(toks) > 0 {
		if checkForThrong {
			if thronged, rest := toks.throng(); len(thronged) > 1 && ((len(rest) > 0) || (len(ret) > 0)) {
				ret = append(ret, me.parseNode(thronged, false))
				toks = rest
				continue
			}
		}

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
		case TokKindIdent, TokKindOp:
			ret = append(ret, parseLit[string](toks, AstNodeKindIdent, func(src string) (string, error) { return src, nil }))
			toks = toks[1:]
		case TokKindBrace:
			toks_inner, toks_tail, err := toks.braceMatch()
			if err != nil {
				ret = append(ret, &AstNode{Kind: AstNodeKindErr, Toks: toks, Src: toks.src(me.Content.Src), err: err})
				toks = nil
			} else {
				switch tok.Src[0] {
				case '(':
					if len(toks_inner) == 0 {
						ret = append(ret, &AstNode{Kind: AstNodeKindErr, Toks: toks[:2], Src: toks[:2].src(me.Content.Src), err: &SrcFileNotice{
							Kind: NoticeKindErr, Message: "expression expected", Span: toks[1].span(), Code: NoticeCodeExprExpected,
						}})
					} else {
						node := me.parseNode(toks_inner, true)
						node.Toks = toks[0 : len(toks_inner)+2]  // want to include the parens in the node's SrcFileSpan..
						node.Src = node.Toks.src(me.Content.Src) // .. and for Src to reflect that SrcFileSpan fully
						ret = append(ret, node)
					}
				case '[', '{':
					split := toks_inner.split(TokKindSep)
					node := &AstNode{Kind: util.If(tok.Src[0] == '[', AstNodeKindSquareBrackets, AstNodeKindCurlyBraces)}
					node.Toks = toks[0 : len(toks_inner)+2]  // want to include the braces/brackets in the node's SrcFileSpan..
					node.Src = node.Toks.src(me.Content.Src) // .. and for Src to reflect that SrcFileSpan fully
					for _, toks := range split {
						node.ChildNodes = append(node.ChildNodes, me.parseNode(toks, true))
					}
					ret = append(ret, node)
				}
				toks = toks_tail
			}
		default:
			ret = append(ret, &AstNode{Kind: AstNodeKindErr, Toks: toks[:1], Src: tok.Src, err: &SrcFileNotice{
				Kind: NoticeKindErr, Message: "unexpected: '" + tok.Src + "'", Span: tok.span(), Code: NoticeCodeMisplaced,
			}})
			toks = toks[1:]
		}
	}

	return
}

func parseLit[T cmp.Ordered](toks Toks, kind AstNodeKind, parseFunc func(string) (T, error)) *AstNode {
	tok := toks[0]
	lit, err := parseFunc(tok.Src)
	if err != nil {
		return &AstNode{Kind: AstNodeKindErr, Toks: toks[:1], Src: tok.Src, err: errToNotice(err, NoticeCodeBadLitSyntax, util.Ptr(tok.span()))}
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

func (me *AstNode) equals(it *AstNode) bool {
	if me.Kind != it.Kind || len(me.ChildNodes) != len(it.ChildNodes) {
		return false
	}
	util.Assert(me != it, nil)
	switch me.Kind {
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
	case AstNodeKindCallForm, AstNodeKindCurlyBraces, AstNodeKindSquareBrackets:
		var idx int
		return sl.All(me.ChildNodes, func(node *AstNode) (isEq bool) {
			isEq = node.equals(it.ChildNodes[idx])
			idx++
			return
		})
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

func (me *AstNode) isIdentOp() bool {
	return me.Kind == AstNodeKindIdent && me.Toks[0].Kind == TokKindOp
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

func (me *AstNode) setParents() {
	me.walk(nil, func(it *AstNode) {
		for _, node := range it.ChildNodes {
			node.parent = it
		}
	})
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
