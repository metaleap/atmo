package session

import (
	"atmo/util/sl"
)

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

func (me *SrcPkg) refreshEst() (encounteredDiagRelevantChanges bool) {
	new_ast_nodes := map[*AstNode]bool{}
	same_est_nodes := map[*EstNode]bool{}
	for _, src_file := range me.Files {
		for _, ast_node := range src_file.Content.Ast {
			var found bool
			for _, est_node := range me.Est {
				if est_node.Src.Node.equals(ast_node, true) {
					est_node.Src.File, est_node.Src.Node = src_file, ast_node
					found, same_est_nodes[est_node] = true, true
				}
			}
			if !found {
				new_ast_nodes[ast_node] = true
			}
		}
	}
	new_est := sl.Where(me.Est, func(it *EstNode) bool {
		return same_est_nodes[it]
	})

	encounteredDiagRelevantChanges = (len(new_est) != len(me.Est)) || (len(new_ast_nodes) > 0)
	for ast_node := range new_ast_nodes {
		_ = ast_node
	}
	me.Est = new_est
	return
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
