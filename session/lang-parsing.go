package session

import (
	"cmp"

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
	LitAtom     any // if NodeKindIdent or some NodeKindLitFoo, one of: float64 | uint64 | rune | string
}

type NodeKind int

const (
	NodeKindErr            NodeKind = iota
	NodeKindCall                    // foo bar baz
	NodeKindCurlyBraces             // {}
	NodeKindSquareBrackets          // []
	NodeKindLitUInt                 // 123, -321
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
	return nil, []*SrcFileNotice{{Kind: NoticeKindErr, Span: toks.span(), Code: NoticeCodeMultipleNodes,
		Message: str.Fmt("expected a single expression, not %d", len(nodes))}}
}

func (*SrcFile) parseNodes(toks Toks) (ret Nodes, errs []*SrcFileNotice) {
	for len(toks) > 0 {
		tok := toks[0]
		switch tok.Kind {
		case TokKindErr:
			ret = append(ret, &Node{Kind: NodeKindErr, Toks: toks[:1], Src: tok.Src})

		}
	}
	return
}

func (me *Node) equals(it *Node) bool {
	if me.Kind != it.Kind || len(me.Children) == len(it.Children) {
		return false
	}
	switch me.Kind {
	case NodeKindLitFloat:
		return (me.LitAtom.(float64) == it.LitAtom.(float64))
	case NodeKindLitUInt:
		return (me.LitAtom.(uint64) == it.LitAtom.(uint64))
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
