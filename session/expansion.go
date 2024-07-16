package session

import (
	"atmo/util/sl"
	"atmo/util/str"
)

type EstNodes []*EstNode
type EstNode struct {
	SrcNode    *AstNode `json:"-"`
	SrcFile    *SrcFile `json:"-"`
	Kind       EstNodeKind
	ChildNodes EstNodes `json:"-"`
}

type EstNodeKind int

const (
	EstNodeKindInvalid EstNodeKind = iota
	EstNodeKindIdent
	EstNodeKindLit
	EstNodeKindCall
	EstNodeKindMacro
)

func (me *SrcPkg) refreshEst() (encounteredDiagsRelevantChanges bool) {
	new_ast_nodes, same_est_nodes := map[*AstNode]*SrcFile{}, map[*EstNode]bool{}
	for _, src_file := range me.Files {
		for _, ast_node := range src_file.Content.Ast {
			var found bool
			for _, est_node := range me.Est {
				if est_node.SrcNode.equals(ast_node, true) {
					est_node.SrcFile, est_node.SrcNode = src_file, ast_node
					found, same_est_nodes[est_node] = true, true
				}
			}
			if !found {
				new_ast_nodes[ast_node] = src_file
			}
		}
	}
	ctx := ctxExpand{pkg: me, est: sl.Where(me.Est, func(it *EstNode) bool {
		return same_est_nodes[it]
	})}

	encounteredDiagsRelevantChanges = (len(ctx.est) != len(me.Est)) || (len(new_ast_nodes) > 0)
	for top_level_ast_node, src_file := range new_ast_nodes {
		ctx.curAstNodesSrcFile = src_file
		ctx.absorb(top_level_ast_node)
	}
	me.Est = ctx.est
	return
}

type ctxExpand struct {
	pkg                *SrcPkg
	curAstNodesSrcFile *SrcFile
	est                EstNodes
}

func (me *ctxExpand) absorb(astNode *AstNode) {
	switch astNode.Kind {
	default:
		astNode.errsExpansion = append(astNode.errsExpansion, astNode.newDiagWarn(NoticeCodeAtmoTodo, str.Fmt("Absorb node kind %d", astNode.Kind)))
	}
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
