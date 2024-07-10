package session

import (
	"cmp"
	"strconv"

	"atmo/util"
	"atmo/util/sl"
	"atmo/util/str"
)

type Nodes []*AstNode

type AstNode struct {
	parent      *AstNode
	Kind        NodeKind
	Children    Nodes
	Toks        Toks
	Src         string
	DocComments Toks
	errsParsing []*SrcFileNotice
	LitAtom     any // if NodeKindIdent or NodeKindLit, one of: float64 | int64 | uint64 | rune | string
}

type NodeKind int

const (
	NodeKindErr            NodeKind = iota
	NodeKindCallForm                // foo bar baz
	NodeKindCurlyBraces             // {}
	NodeKindSquareBrackets          // []
	NodeKindIdent                   // foo, #bar, @baz, $foo, %bar
	NodeKindLit                     // 123, -321, 1.23, -3.21, "foo", `bar`, 'ö'
)

// only called by EnsureSrcFile, just after tokenization, with `.Notices.LexErrs` freshly set.
// mutates me.Content.TopLevelAstNodes and me.Notices.ParseErrs.
func (me *SrcFile) parse(previously Nodes) {
	me.Notices.ParseErrs = nil

	top_level_nodes := previously

	// remove nodes whose src is no longer present in toks
	var gone_nodes Nodes
	for i := 0; i < len(top_level_nodes); i++ {
		if sl.None(me.Content.TopLevelToksChunks, func(topLevelChunk Toks) bool {
			return topLevelChunk.src(me.Content.Src) == top_level_nodes[i].Src
		}) {
			gone_nodes = append(gone_nodes, top_level_nodes[i])                     // keep around for potential reclaim below
			top_level_nodes = append(top_level_nodes[:i], top_level_nodes[i+1:]...) // remove
			i--
		}
	}

	// gather top-level chunks whose nodes do not yet exist
	var new_nodes Nodes
	for _, top_level_chunk := range me.Content.TopLevelToksChunks {
		toks_sans_comments := sl.Where(top_level_chunk, func(it *Tok) bool { return it.Kind != TokKindComment })
		if len(toks_sans_comments) == 0 {
			continue
		}
		node := sl.FirstWhere(top_level_nodes, func(it *AstNode) bool { return it.Src == top_level_chunk.src(me.Content.Src) })
		if node == nil {
			node, errs := me.parseNode(toks_sans_comments)
			me.Notices.ParseErrs = append(me.Notices.ParseErrs, errs...)
			if node != nil {
				node.Src = top_level_chunk.src(me.Content.Src) // parsed without comments, but need full Src for the very comparisons above
				node.setParents()
				// keep doc-comments in top-level node: those at the beginning of a top-level chunk
				for _, tok := range top_level_chunk {
					if tok.Kind != TokKindComment {
						break
					} else {
						node.DocComments = append(node.DocComments, tok)
					}
				}
				new_nodes = append(new_nodes, node)
			}
		}
	}

	// if a just-parsed node has existed before (ie had only inconsequential toks changes),
	// recover it (to keep annotations etc) but update wrt new toks, span & src
	for i := 0; i < len(new_nodes); i++ {
		new_node := new_nodes[i]
		if old_node := sl.FirstWhere(gone_nodes, func(it *AstNode) bool { return it.equals(new_node) }); old_node != nil {
			old_node.Toks, old_node.Src, old_node.errsParsing = new_node.Toks, new_node.Src, new_node.errsParsing
			gone_nodes = sl.Without(gone_nodes, true, old_node)
			new_nodes = append(new_nodes[:i], append(Nodes{old_node}, new_nodes[i+1:]...)...)
			i--
		}
	}

	top_level_nodes = append(top_level_nodes, new_nodes...)
	// sort all nodes to be in source-file order of appearance
	top_level_nodes = sl.SortedPer(top_level_nodes, func(node1 *AstNode, node2 *AstNode) int {
		return cmp.Compare(node1.Toks[0].byteOffset, node2.Toks[0].byteOffset)
	})

	me.Content.TopLevelAstNodes = top_level_nodes
}

func (me *SrcFile) parseNode(toks Toks) (*AstNode, []*SrcFileNotice) {
	var sub_nodes Nodes
	toks_full := toks
	{
		head, tail, err := toks.subChunks()
		if err != nil {
			return nil, []*SrcFileNotice{err}
		} else if len(tail) > 0 {
			toks = head
			for _, chunk_toks := range tail {
				sub_node, errs := me.parseNode(chunk_toks)
				if len(errs) > 0 {
					return nil, errs
				} else if sub_node != nil {
					sub_nodes = append(sub_nodes, sub_node)
				}
			}
		}
	}

	nodes, errs := me.parseNodes(toks)
	if len(errs) > 0 {
		return nil, errs
	}
	if len(nodes) == 1 && len(sub_nodes) == 0 {
		return nodes[0], nil
	}
	ret := &AstNode{Kind: NodeKindCallForm, Children: append(nodes, sub_nodes...), Toks: toks_full, Src: toks_full.src(me.Content.Src)}
	return ret, nil
}

func (me *SrcFile) parseNodes(toks Toks) (ret Nodes, errs []*SrcFileNotice) {
	for len(toks) > 0 {
		tok := toks[0]
		switch tok.Kind {
		case TokKindErr:
			ret = append(ret, &AstNode{Kind: NodeKindErr, Toks: toks[:1], Src: tok.Src})
			toks = toks[1:]
		case TokKindLitFloat:
			ret = append(ret, parseLit[float64](toks, NodeKindLit, func(src string) (float64, error) { return str.ToF(src, 64) }))
			toks = toks[1:]
		case TokKindLitRune:
			ret = append(ret, parseLit[rune](toks, NodeKindLit, func(src string) (rune, error) {
				ret, _, _, err := strconv.UnquoteChar(src, '\'')
				return ret, err
			}))
			toks = toks[1:]
		case TokKindLitStr:
			ret = append(ret, parseLit[string](toks, NodeKindLit, strconv.Unquote))
			toks = toks[1:]
		case TokKindLitInt:
			if tok.Src[0] == '-' {
				ret = append(ret, parseLit[int64](toks, NodeKindLit, func(src string) (int64, error) {
					return str.ToI64(src, 0, 64)
				}))
			} else {
				ret = append(ret, parseLit[uint64](toks, NodeKindLit, func(src string) (uint64, error) {
					return str.ToU64(src, 0, 64)
				}))
			}
			toks = toks[1:]
		case TokKindIdent, TokKindOp:
			ret = append(ret, parseLit[string](toks, NodeKindIdent, func(src string) (string, error) { return src, nil }))
			toks = toks[1:]
		case TokKindBrace:
			toks_inner, toks_tail, err := toks.braceMatch()
			if err != nil {
				ret = append(ret, &AstNode{Kind: NodeKindErr, Toks: toks, Src: toks.src(me.Content.Src), errsParsing: []*SrcFileNotice{err}})
				toks = nil
			} else {
				switch tok.Src {
				case "(":
					if len(toks_inner) == 0 {
						ret = append(ret, &AstNode{Kind: NodeKindErr, Toks: toks[:2], Src: toks[:2].src(me.Content.Src), errsParsing: []*SrcFileNotice{{
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
					node := &AstNode{Kind: util.If(tok.Src == "[", NodeKindSquareBrackets, NodeKindCurlyBraces)}
					node.Toks = toks[0 : len(toks_inner)+2]  // want to include the braces/brackets in the node's SrcFileSpan..
					node.Src = node.Toks.src(me.Content.Src) // .. and for Src to reflect that SrcFileSpan fully
					for _, toks := range split {
						sub_node, errs_sub_node := me.parseNode(toks)
						if len(errs_sub_node) == 0 {
							node.Children = append(node.Children, sub_node)
						} else {
							errs = append(errs, errs_sub_node...)
						}
					}
					ret = append(ret, node)
				}
				toks = toks_tail
			}
		default:
			ret = append(ret, &AstNode{Kind: NodeKindErr, Toks: toks[:1], Src: tok.Src, errsParsing: []*SrcFileNotice{{
				Kind: NoticeKindErr, Message: "unexpected: '" + tok.Src + "'", Span: tok.span(), Code: NoticeCodeMisplaced,
			}}})
			toks = toks[1:]
		}
	}
	return
}

func parseLit[T cmp.Ordered](toks Toks, kind NodeKind, parseFunc func(string) (T, error)) *AstNode {
	tok := toks[0]
	lit, err := parseFunc(tok.Src)
	if err != nil {
		return &AstNode{Kind: NodeKindErr, Toks: toks[:1], Src: tok.Src,
			errsParsing: []*SrcFileNotice{errToNotice(err, NoticeCodeBadLitSyntax, util.Ptr(tok.span()))}}
	}
	return &AstNode{Kind: kind, Toks: toks[:1], Src: tok.Src, LitAtom: lit}
}

func (me *SrcFile) NodeAt(pos SrcFilePos, orAncestor bool) (ret *AstNode) {
	for _, node := range me.Content.TopLevelAstNodes {
		if node.Toks.Span().contains(&pos) {
			ret = node.find(func(it *AstNode) bool {
				return (len(it.Children) == 0) && it.Toks.Span().contains(&pos)
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
	if me.Kind != it.Kind || len(me.Children) != len(it.Children) {
		return false
	}
	util.Assert(me != it, nil)
	switch me.Kind {
	case NodeKindLit:
		switch mine := me.LitAtom.(type) {
		case float64:
			other, ok := it.LitAtom.(float64)
			return ok && (mine == other)
		case int64:
			other, ok := it.LitAtom.(int64)
			return ok && (mine == other)
		case uint64:
			other, ok := it.LitAtom.(uint64)
			return ok && (mine == other)
		case rune:
			other, ok := it.LitAtom.(rune)
			return ok && (mine == other)
		case string:
			other, ok := it.LitAtom.(string)
			return ok && (mine == other)
		default:
			panic(me.LitAtom)
		}
	case NodeKindIdent:
		return (me.LitAtom.(string) == it.LitAtom.(string))
	case NodeKindErr:
		var idx int
		return (len(me.errsParsing) == len(it.errsParsing)) && sl.All(me.errsParsing, func(err *SrcFileNotice) (isEq bool) {
			isEq = (err == it.errsParsing[idx]) || (*err == *it.errsParsing[idx])
			idx++
			return
		})
	case NodeKindCallForm, NodeKindCurlyBraces, NodeKindSquareBrackets:
		var idx int
		return sl.All(me.Children, func(node *AstNode) (isEq bool) {
			isEq = node.equals(it.Children[idx])
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

func (me *AstNode) SelfAndAncestors() (ret Nodes) {
	for it := me; it != nil; it = it.parent {
		ret = append(ret, it)
	}
	return
}

func (me *AstNode) setParents() {
	me.walk(nil, func(it *AstNode) {
		for _, node := range it.Children {
			node.parent = it
		}
	})
}

func (me *AstNode) walk(onBefore func(n *AstNode) bool, onAfter func(n *AstNode)) {
	if onBefore != nil && !onBefore(me) {
		return
	}
	for _, node := range me.Children {
		node.walk(onBefore, onAfter)
	}
	if onAfter != nil {
		onAfter(me)
	}
}
