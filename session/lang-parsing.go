package session

import (
	"cmp"
	"strconv"
	"strings"

	"atmo/util"
	"atmo/util/sl"
	"atmo/util/str"
)

type AstNodes []*AstNode

type AstNode struct {
	parent      *AstNode
	Kind        AstNodeKind
	ChildNodes  AstNodes
	Toks        Toks
	Src         string
	errsParsing []*SrcFileNotice
	Lit         any // if AstNodeKindIdent or AstNodeKindLit, one of: float64 | int64 | uint64 | rune | string
	Ann         any
}

type AstNodeKind int

const (
	AstNodeKindErr            AstNodeKind = iota
	AstNodeKindCallForm                   // foo bar baz
	AstNodeKindCurlyBraces                // {}
	AstNodeKindSquareBrackets             // []
	AstNodeKindIdent                      // foo, #bar, @baz, $foo, %bar
	AstNodeKindLit                        // 123, -321, 1.23, -3.21, "foo", `bar`, 'ö'
)

// only called by EnsureSrcFile, just after tokenization, with `.Notices.LexErrs` freshly set.
// mutates me.Content.TopLevelAstNodes and me.Notices.ParseErrs.
func (me *SrcFile) parse() {
	me.Notices.ParseErrs = nil

	parsed, errs := me.parseNodes(me.Content.Toks)
	me.Notices.ParseErrs = errs

	// sort all nodes to be in source-file order of appearance
	parsed = sl.SortedPer(parsed, func(node1 *AstNode, node2 *AstNode) int {
		return cmp.Compare(node1.Toks[0].byteOffset, node2.Toks[0].byteOffset)
	})
	me.Content.Ast = parsed
}

func (me *SrcFile) parseNode(toks Toks) (*AstNode, []*SrcFileNotice) {
	nodes, errs := me.parseNodes(toks)
	if len(errs) > 0 {
		return nil, errs
	}
	if len(nodes) == 1 {
		return nodes[0], nil
	}
	ret := &AstNode{Kind: AstNodeKindCallForm, ChildNodes: nodes, Toks: toks, Src: toks.src(me.Content.Src)}
	return ret, nil
}

func (me *SrcFile) parseNodes(toks Toks) (ret AstNodes, errs []*SrcFileNotice) {
	for len(toks) > 0 {
		if thronged, tail := toks.throng(); len(thronged) > 0 {
			node, errs_node := me.parseNode(thronged)
			errs = append(errs, errs_node...)
			if node != nil {
				ret = append(ret, node)
			}
			toks = tail
			continue
		}

		tok := toks[0]
		switch tok.Kind {
		case TokKindErr:
			ret = append(ret, &AstNode{Kind: AstNodeKindErr, Toks: toks[:1], Src: tok.Src})
			toks = toks[1:]
		case TokKindLitFloat:
			ret = append(ret, parseLit[float64](toks, AstNodeKindLit, func(src string) (float64, error) { return str.ToF(src, 64) }))
			toks = toks[1:]
		case TokKindLitRune:
			ret = append(ret, parseLit[rune](toks, AstNodeKindLit, func(src string) (rune, error) {
				ret, _, _, err := strconv.UnquoteChar(src, '\'')
				return ret, err
			}))
			toks = toks[1:]
		case TokKindLitStr:
			ret = append(ret, parseLit[string](toks, AstNodeKindLit, strconv.Unquote))
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
				ret = append(ret, &AstNode{Kind: AstNodeKindErr, Toks: toks, Src: toks.src(me.Content.Src), errsParsing: []*SrcFileNotice{err}})
				toks = nil
			} else {
				switch tok.Src {
				case "(":
					if len(toks_inner) == 0 {
						ret = append(ret, &AstNode{Kind: AstNodeKindErr, Toks: toks[:2], Src: toks[:2].src(me.Content.Src), errsParsing: []*SrcFileNotice{{
							Kind: NoticeKindErr, Message: "expression expected", Span: toks[1].span(), Code: NoticeCodeExprExpected,
						}}})
					} else {
						node, errs_inner := me.parseNode(toks_inner)
						errs = append(errs, errs_inner...)
						if node != nil {
							node.Toks = toks[0 : len(toks_inner)+2]  // want to include the parens in the node's SrcFileSpan..
							node.Src = node.Toks.src(me.Content.Src) // .. and for Src to reflect that SrcFileSpan fully
							ret = append(ret, node)
						}
					}
				case "[", "{":
					split := toks_inner.split(TokKindSep)
					node := &AstNode{Kind: util.If(tok.Src == "[", AstNodeKindSquareBrackets, AstNodeKindCurlyBraces)}
					node.Toks = toks[0 : len(toks_inner)+2]  // want to include the braces/brackets in the node's SrcFileSpan..
					node.Src = node.Toks.src(me.Content.Src) // .. and for Src to reflect that SrcFileSpan fully
					for _, toks := range split {
						sub_node, errs_sub_node := me.parseNode(toks)
						if len(errs_sub_node) == 0 {
							node.ChildNodes = append(node.ChildNodes, sub_node)
						} else {
							errs = append(errs, errs_sub_node...)
						}
					}
					ret = append(ret, node)
				}
				toks = toks_tail
			}
		default:
			ret = append(ret, &AstNode{Kind: AstNodeKindErr, Toks: toks[:1], Src: tok.Src, errsParsing: []*SrcFileNotice{{
				Kind: NoticeKindErr, Message: "unexpected: '" + tok.Src + "'", Span: tok.span(), Code: NoticeCodeMisplaced,
			}}})
			toks = toks[1:]
		}
	}
	return
}

func parseLit[T cmp.Ordered](toks Toks, kind AstNodeKind, parseFunc func(string) (T, error)) *AstNode {
	tok := toks[0]
	lit, err := parseFunc(tok.Src)
	if err != nil {
		return &AstNode{Kind: AstNodeKindErr, Toks: toks[:1], Src: tok.Src,
			errsParsing: []*SrcFileNotice{errToNotice(err, NoticeCodeBadLitSyntax, util.Ptr(tok.span()))}}
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
		var idx int
		return (len(me.errsParsing) == len(it.errsParsing)) && sl.All(me.errsParsing, func(err *SrcFileNotice) (isEq bool) {
			isEq = (err == it.errsParsing[idx]) || (*err == *it.errsParsing[idx])
			idx++
			return
		})
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

func (me AstNodes) hasKind(kind AstNodeKind) (ret bool) {
	me.walk(func(it *AstNode) bool {
		ret = ret || (it.Kind == kind)
		return !ret
	}, nil)
	return
}

func (me AstNodes) walk(onBefore func(*AstNode) bool, onAfter func(*AstNode)) {
	for _, node := range me {
		node.walk(onBefore, onAfter)
	}
}
