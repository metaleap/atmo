package session

type EstNodes []*EstNode
type EstNode struct {
	Ast struct {
		Node *AstNode
		File *SrcFile
	}
	ChildNodes EstNodes
	diags      []*SrcFileNotice
}

func (me *SrcFile) expand() {

}

func (me *EstNode) walk(onBefore func(*EstNode) bool, onAfter func(*EstNode)) {
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

func (me EstNodes) walk(onBefore func(node *EstNode) bool, onAfter func(node *EstNode)) {
	for _, node := range me {
		node.walk(onBefore, onAfter)
	}
}