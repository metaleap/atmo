package session

type Nodes []*Node

type Node struct {
	parent   *Node
	Children Nodes
	Toks     Toks
	ToksSrc  string
}

func (*SrcFile) parseNode(toks Toks) (*Node, *SrcFileNotice) {
	return nil, nil
}

func (*SrcFile) parseNodes(toks Toks) (ret Nodes, err *SrcFileNotice) {
	return
}

func (me *Node) equals(it *Node) bool {
	return false
}
