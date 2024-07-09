package session

import (
	"cmp"

	"atmo/util/sl"
	"atmo/util/str"
)

type Nodes []*Node

type Node struct {
	parent   *Node
	Children Nodes
	Toks     Toks
	Src      string
	Errs     struct {
		Parsing []*SrcFileNotice
	}
}

type NodeKind int

const (
	NodeKindErr NodeKind = iota
	NodeKindCallForm
	NodeKindCurlyBraces
	NodeKindSquareBrackets
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
			old_node.Toks, old_node.Src, old_node.Errs.Parsing = new_node.Toks, new_node.Src, new_node.Errs.Parsing
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
	return
}

func (me *Node) equals(it *Node) bool {
	return false
}

func (n *Node) walk(in func(n *Node) bool, out func(n *Node)) {
	if in != nil && !in(n) {
		return
	}
	for _, child_node := range n.Children {
		child_node.walk(in, out)
	}
	if out != nil {
		out(n)
	}
}
