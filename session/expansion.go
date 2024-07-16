package session

type EstNodes []*EstNode
type EstNode struct {
	Src struct {
		Node *AstNode
		File *SrcFile
	} `json:"-"`
	Kind       EstNodeKind
	ChildNodes EstNodes `json:"-"`
}

type EstNodeKind int

const (
	_ EstNodeKind = iota
	EstNodeKindIdent
	EstNodeKindLit
	EstNodeKindCall
)

func (me *SrcPkg) refreshEst() {
	me.Est = EstNodes{{Kind: EstNodeKindCall, ChildNodes: EstNodes{&EstNode{Kind: EstNodeKindIdent}, &EstNode{Kind: EstNodeKindLit}}}}
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
