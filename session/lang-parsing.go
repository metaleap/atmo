package session

import (
	"cmp"
	"strconv"

	"atmo/util"
	"atmo/util/sl"
	"atmo/util/str"
)

type Nodes []*Node

type Node struct {
	parent      *Node
	Kind        NodeKind
	Children    Nodes
	Toks        Toks
	Src         string
	ErrsParsing []*SrcFileNotice
	LitAtom     any // if NodeKindIdent or some NodeKindLitFoo, one of: float64 | int64 | uint64 | rune | string
}

type NodeKind int

const (
	NodeKindErr            NodeKind = iota
	NodeKindCall                    // foo bar baz
	NodeKindCurlyBraces             // {}
	NodeKindSquareBrackets          // []
	NodeKindLitInt                  // 123, -321
	NodeKindLitFloat                // 1.23, -3.21
	NodeKindLitRune                 // '⅜'
	NodeKindLitStr                  // "foo", `bar`
	NodeKindIdent                   // foo, #bar, @baz, $foo, %bar, ()
)

// only called by EnsureSrcFile, just after tokenization, with `.Notices.LexErrs` freshly set.
func (me *SrcFile) parse() {
	me.Notices.ParseErrs = nil

	top_level_nodes := me.Content.TopLevelAstNodes

	// remove nodes whose src is no longer present in toks
	var gone_nodes Nodes
	for i := 0; i < len(top_level_nodes); i++ {
		if !sl.Any(me.Content.TopLevelToksChunks, func(topLevelChunk Toks) bool {
			return topLevelChunk.src(me.Content.Src) == top_level_nodes[i].Src
		}) {
			gone_nodes = append(gone_nodes, top_level_nodes[i])
			top_level_nodes = append(top_level_nodes[:i], top_level_nodes[i+1:]...)
			i--
		}
	}

	// gather top-level chunks whose nodes do not yet exist
	var new_nodes Nodes
	for _, top_level_chunk := range me.Content.TopLevelToksChunks {
		node := sl.FirstWhere(top_level_nodes, func(it *Node) bool { return it.Src == top_level_chunk.src(me.Content.Src) })
		if node == nil {
			node, errs := me.parseNode(top_level_chunk)
			if len(errs) > 0 {
				me.Notices.ParseErrs = append(me.Notices.ParseErrs, errs...)
			} else {
				new_nodes = append(new_nodes, node)
			}
		}
	}

	// if a just-parsed node has existed before (ie had only inconsequential toks changes),
	// recover it (to keep annotations etc) but update wrt new toks span
	for i := 0; i < len(new_nodes); i++ {
		new_node := new_nodes[i]
		if old_node := sl.FirstWhere(gone_nodes, func(it *Node) bool { return it.equals(new_node) }); old_node != nil {
			// TODO! would have to `node.walk` this whole logic, rethink & rework this once we have Anns
			old_node.Toks, old_node.Src, old_node.ErrsParsing = new_node.Toks, new_node.Src, new_node.ErrsParsing
			gone_nodes = sl.Without(gone_nodes, true, old_node)
			new_nodes = append(new_nodes[:i], append(Nodes{old_node}, new_nodes[i+1:]...)...)
			i--
		}
	}

	top_level_nodes = append(top_level_nodes, new_nodes...)
	// sort all nodes to be in source-file order of appearance
	top_level_nodes = sl.SortedPer(top_level_nodes, func(node1 *Node, node2 *Node) int {
		return cmp.Compare(node1.Toks[0].byteOffset, node2.Toks[0].byteOffset)
	})

	me.Content.TopLevelAstNodes = top_level_nodes
}

func (me *SrcFile) parseNode(toks Toks) (*Node, []*SrcFileNotice) {
	nodes, errs := me.parseNodes(toks)
	if len(errs) > 0 {
		return nil, errs
	}
	if len(nodes) == 1 {
		return nodes[0], nil
	}
	return &Node{Kind: NodeKindCall, Children: nodes, Toks: toks, Src: toks.src(me.Content.Src)}, nil
}

func (me *SrcFile) parseNodes(toks Toks) (ret Nodes, errs []*SrcFileNotice) {
	for len(toks) > 0 {
		tok := toks[0]
		switch tok.Kind {
		case TokKindComment:
			toks = toks[1:]
			continue
		case TokKindErr:
			ret = append(ret, &Node{Kind: NodeKindErr, Toks: toks[:1], Src: tok.Src})
			toks = toks[1:]
		case TokKindLitFloat:
			ret = append(ret, parseLit[float64](toks, NodeKindLitFloat, func(src string) (float64, error) { return str.ToF(src, 64) }))
			toks = toks[1:]
		case TokKindLitRune:
			ret = append(ret, parseLit[rune](toks, NodeKindLitRune, func(src string) (rune, error) {
				ret, _, _, err := strconv.UnquoteChar(src, '\'')
				return ret, err
			}))
			toks = toks[1:]
		case TokKindLitStr:
			ret = append(ret, parseLit[string](toks, NodeKindLitStr, strconv.Unquote))
			toks = toks[1:]
		case TokKindLitInt:
			if tok.Src[0] == '-' {
				ret = append(ret, parseLit[int64](toks, NodeKindLitInt, func(src string) (int64, error) {
					return str.ToI64(src, 0, 64)
				}))
			} else {
				ret = append(ret, parseLit[uint64](toks, NodeKindLitInt, func(src string) (uint64, error) {
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
				ret = append(ret, &Node{Kind: NodeKindErr, Toks: toks, Src: toks.src(me.Content.Src), ErrsParsing: []*SrcFileNotice{err}})
				toks = nil
			} else {
				toks = toks_tail
				switch tok.Src {
				case "(":
					node, errs_inner := me.parseNode(toks_inner)
					if len(errs_inner) == 0 {
						ret = append(ret, node)
					} else {
						errs = append(errs, errs_inner...)
					}
				case "[", "{":
					split := toks_inner.split(TokKindSep)
					node := &Node{
						Kind: util.If(tok.Src == "[", NodeKindSquareBrackets, NodeKindCurlyBraces),
						Toks: toks_inner,
						Src:  toks_inner.src(me.Content.Src),
					}
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
			}
		default:
			ret = append(ret, &Node{Kind: NodeKindErr, Toks: toks[:1], Src: tok.Src, ErrsParsing: []*SrcFileNotice{{
				Kind: NoticeKindErr, Message: "unexpected: '" + tok.Src + "'", Span: tok.span(), Code: NoticeCodeMisplaced,
			}}})
			toks = toks[1:]
		}
	}
	return
}

func parseLit[T cmp.Ordered](toks Toks, kind NodeKind, parseFunc func(string) (T, error)) *Node {
	tok := toks[0]
	lit, err := parseFunc(tok.Src)
	if err != nil {
		return &Node{Kind: NodeKindErr, Toks: toks[:1], Src: tok.Src,
			ErrsParsing: []*SrcFileNotice{errToNotice(err, NoticeCodeBadLitSyntax, util.Ptr(tok.span()))}}
	}
	return &Node{Kind: kind, Toks: toks[:1], Src: tok.Src, LitAtom: lit}
}

func (me *Node) equals(it *Node) bool {
	if me.Kind != it.Kind || len(me.Children) == len(it.Children) {
		return false
	}
	switch me.Kind {
	case NodeKindLitFloat:
		return (me.LitAtom.(float64) == it.LitAtom.(float64))
	case NodeKindLitInt:
		switch mine := me.LitAtom.(type) {
		case int64:
			other, ok := it.LitAtom.(int64)
			return ok && (mine == other)
		case uint64:
			other, ok := it.LitAtom.(uint64)
			return ok && (mine == other)
		default:
			panic(me.LitAtom)
		}
	case NodeKindLitRune:
		return (me.LitAtom.(rune) == it.LitAtom.(rune))
	case NodeKindLitStr:
		return (me.LitAtom.(string) == it.LitAtom.(string))
	case NodeKindIdent:
		return me.LitAtom.(string) == it.LitAtom.(string)
	case NodeKindErr:
		var idx int
		return (len(me.ErrsParsing) == len(it.ErrsParsing)) && sl.All(me.ErrsParsing, func(err *SrcFileNotice) (isEq bool) {
			isEq = (*err == *it.ErrsParsing[idx])
			idx++
			return
		})
	case NodeKindCall, NodeKindCurlyBraces, NodeKindSquareBrackets:
		var idx int
		return sl.All(me.Children, func(node *Node) (isEq bool) {
			isEq = node.equals(it.Children[idx])
			idx++
			return
		})
	default:
		panic(me.Kind)
	}
}

func (n *Node) walk(onBefore func(n *Node) bool, onAfter func(n *Node)) {
	if onBefore != nil && !onBefore(n) {
		return
	}
	for _, child_node := range n.Children {
		child_node.walk(onBefore, onAfter)
	}
	if onAfter != nil {
		onAfter(n)
	}
}
