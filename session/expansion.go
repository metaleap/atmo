package session

import (
	"atmo/util/sl"
)

type EstNodes []*EstNode
type EstNode struct {
	parent  *EstNode
	SrcNode *AstNode `json:"-"`
	SrcFile *SrcFile `json:"-"`
	Kind    EstNodeKind
	Nodes   EstNodes `json:"-"`
	Self    estNode  `json:",omitempty"`
}

type EstNodeKind int

const (
	EstNodeKindInvalid EstNodeKind = iota
	EstNodeKindIdent
	EstNodeKindLit
	EstNodeKindCall
	EstNodeKindMacro
)

type Scope struct {
	parent *Scope
	lookup map[string]Ref
}

type Ref struct {
	pkg  *SrcPkg
	node *EstNode
}

type estNode interface{}

type EstNodeMacro struct {
	Pattern AstNodes
	Body    AstNodes
}

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
		ctx.addMacroFrom(top_level_ast_node)
	}
	for top_level_ast_node, src_file := range new_ast_nodes {
		ctx.curAstNodesSrcFile = src_file
		ctx.addCallFrom(top_level_ast_node, true)
	}
	me.Est = ctx.est
	return
}

type ctxExpand struct {
	pkg                *SrcPkg
	curAstNodesSrcFile *SrcFile
	est                EstNodes
}

func (me *ctxExpand) addMacroFrom(astNode *AstNode) {
	if !astNode.isMacro() {
		return
	}
	switch {
	case (len(astNode.Nodes) < 2):
		astNode.errsExpansion.Add(astNode.Nodes.last().newDiagErr(true, NoticeCodeExpectedFooHere, "macro pattern and body", "after `@macro`"))
	case (len(astNode.Nodes[1].Nodes) <= 1):
		astNode.errsExpansion.Add(astNode.Nodes[1].newDiagErr(true, NoticeCodeExpectedFooHere, "macro pattern", ""))
	case (len(astNode.Nodes) < 3):
		astNode.errsExpansion.Add(astNode.Nodes.last().newDiagErr(true, NoticeCodeExpectedFooHere, "macro body", "after `@macro` and pattern"))
	default:
		self := EstNodeMacro{
			Pattern: astNode.Nodes[1].Nodes,
			Body:    astNode.Nodes[2:],
		}
		macro_node := &EstNode{SrcNode: astNode, SrcFile: me.curAstNodesSrcFile, Kind: EstNodeKindMacro,
			Self: self}
		me.est = append(me.est, macro_node)
	}
}

func (me *ctxExpand) addCallFrom(astNode *AstNode, must bool) {
	switch astNode.Kind {
	case AstNodeKindComment, AstNodeKindErr:
	case AstNodeKindGroup:
	default:
		if must {
			astNode.errsExpansion.Add(astNode.newDiagErr(false, NoticeCodeExpectedFooHere, "call", ""))
		}
	}
}

func (me *EstNode) walk(onBefore func(*EstNode) bool, onAfter func(*EstNode)) {
	if onBefore != nil && !onBefore(me) {
		return
	}
	for _, node := range me.Nodes {
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

func (me *AstNode) isMacro() bool {
	return (me.Kind == AstNodeKindGroup) && (len(me.Nodes) > 0) && (me.Nodes[0].ident() == "@macro")
}
