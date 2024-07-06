package session

import (
	"atmo/util/str"

	"strings"
)

var (
	parsingOpsNumeric         = []string{"+", "-", "*", "/", "%"}
	parsingOpsShortcircuiting = []string{"&&", "||"}
	parsingOpsBitwise         = []string{"!", "&", "|", "~", "^", "<<", ">>"}
	parsingOpsCmp             = []string{"<", ">", "==", "!=", ">=", "<="}
)

type AstFile struct {
	srcFilePath string
	origSrc     string
	toks        Tokens
	topLevel    []AstNode
}

type AstNode interface {
	base() AstNodeBase
	String(int) string
}

type AstNodeBase struct {
	toks Tokens
}

func (me AstNodeBase) base() AstNodeBase { return me }

type AstNodeBraced struct {
	AstNodeBase
	square bool
	curly  bool
	list   AstNodeList
}

type AstNodeList struct {
	AstNodeBase
	sep   string
	nodes []AstNode
}

type AstNodePair struct {
	AstNodeBase
	sep string
	lhs AstNode
	rhs AstNode
}

type AstNodeAtom struct {
	AstNodeBase
}

func (me AstNodeAtom) String(indent int) (s string) {
	s = strings.Repeat("\t", indent) + "<ATOM tk='" + me.toks[0].kind.String() + "'>"
	s += me.toks.String("", "")
	s += "</ATOM>\n"
	return
}

func (me AstNodePair) String(indent int) (s string) {
	s = strings.Repeat("\t", indent) + "<PAIR>\n"
	s += me.lhs.String(indent + 1)
	s += strings.Repeat("\t", indent+1) + me.sep + "\n"
	s += me.rhs.String(indent + 1)
	s += strings.Repeat("\t", indent) + "</PAIR>\n"
	return
}

func (me AstNodeList) String(indent int) (s string) {
	s = strings.Repeat("\t", indent) + "<LIST sep='" + me.sep + "'>\n"
	for i, node := range me.nodes {
		s += strings.Repeat("\t", indent+1) + "<" + str.FromInt(i) + ">\n"
		s += node.String(indent + 2)
		s += strings.Repeat("\t", indent+1) + "</" + str.FromInt(i) + ">\n"
	}
	s += strings.Repeat("\t", indent) + "</LIST>\n"
	return
}

func (me AstNodeBraced) String(indent int) (s string) {
	s = strings.Repeat("\t", indent) + "<BR t='" + me.toks[0].src + "'>\n"
	s += me.list.String(indent + 1)
	s += strings.Repeat("\t", indent) + "</BR>\n"
	return
}

func parse(toks Tokens, origSrc string, srcFilePath string) (ret *AstFile, errs []*SrcFileNotice) {
	ret = &AstFile{toks: toks, srcFilePath: srcFilePath, origSrc: origSrc}
	toplevelchunks := toks.indentLevelChunks(0)
	for _, tlc := range toplevelchunks {
		if node, err := ret.parseNode(tlc); err != nil {
			errs = append(errs, err)
		} else if node != nil {
			ret.topLevel = append(ret.topLevel, node)
		}
	}
	return
}

func (me *AstFile) parseNode(toks Tokens) (AstNode, *SrcFileNotice) {
	nodes, err := me.parseNodes(toks)
	if err != nil || len(nodes) == 0 {
		return nil, err
	} else if len(nodes) > 1 {
		return nil, &SrcFileNotice{Kind: NoticeKindErr, Code: NoticeCodeMultipleAstNodes, Span: toks.Span(),
			Message: "expected single expression, not multiple"}
	}
	return nodes[0], err
}

func (me *AstFile) parseNodes(toks Tokens) (ret []AstNode, err *SrcFileNotice) {
	for len(toks) > 0 {
		var node AstNode
		if t := &toks[0]; t.kind == tokKindComment {
			toks = toks[1:]
		} else if t.src == "[" || t.src == "(" || t.src == "{" {
			node, toks, err = me.parseNodeBraced(toks)
		} else if toks.idxAtLevel0(",") >= 0 {
			node, err = me.parseNodeList(toks, ",")
			toks = nil
		} else if idx := toks.idxAtLevel0("="); idx > 0 {
			node, err = me.parseNodePair(toks, idx)
			toks = nil
		} else if idx = toks.idxAtLevel0(":"); idx > 0 {
			node, err = me.parseNodePair(toks, idx)
			toks = nil
		} else if toks.anyAtLevel0(parsingOpsShortcircuiting...) {
			node, err = me.parseNodeList(toks, parsingOpsShortcircuiting...)
			toks = nil
		} else if toks.anyAtLevel0(parsingOpsCmp...) {
			node, err = me.parseNodeList(toks, parsingOpsCmp...)
			toks = nil
		} else if toks.anyAtLevel0(parsingOpsNumeric...) {
			node, err = me.parseNodeList(toks, parsingOpsNumeric...)
			toks = nil
		} else if toks.anyAtLevel0(parsingOpsBitwise...) {
			node, err = me.parseNodeList(toks, parsingOpsBitwise...)
			toks = nil
		} else {
			node, toks = AstNodeAtom{AstNodeBase: AstNodeBase{toks: toks[:1]}}, toks[1:]
		}
		if node != nil {
			ret = append(ret, node)
		}
	}
	return
}

func (me *AstFile) parseNodeBraced(toks Tokens) (ret AstNodeBraced, tail Tokens, err *SrcFileNotice) {
	ret.toks, ret.square, ret.curly = toks, (toks[0].src == "["), (toks[0].src == "{")
	idx := toks.idxOfClosingBrace()
	if idx <= 0 {
		err = &SrcFileNotice{Kind: NoticeKindErr, Code: NoticeCodeUnmatchedBrace, Span: toks.Span(),
			Message: "unmatched brace '" + toks[0].src + "'"}
		return
	}
	ret.list, err = me.parseNodeList(toks[1:idx], ",")
	tail = toks[idx+1:]
	return
}

func (me *AstFile) parseNodeList(toks Tokens, sepOrOps ...string) (ret AstNodeList, err *SrcFileNotice) {
	tokss, sep, err := toks.split(sepOrOps...)
	ret.sep, ret.toks = sep, toks
	for _, nodetoks := range tokss {
		node, err_parse_node := me.parseNode(nodetoks)
		if err = err_parse_node; err != nil {
			return
		}
		ret.nodes = append(ret.nodes, node)
	}
	return
}

func (me *AstFile) parseNodePair(toks Tokens, idx int) (ret AstNodePair, err *SrcFileNotice) {
	ret.toks, ret.sep = toks, toks[idx].src
	if ret.lhs, err = me.parseNode(toks[:idx]); err == nil {
		ret.rhs, err = me.parseNode(toks[idx+1:])
	}
	return
}

/*
L1 C1 'IdentName'>>>>str<<<<
L1 C4 'Sep'>>>>:<<<<
L1 C6 'IdentOp'>>>>@<<<<
L1 C8 'IdentOp'>>>>=<<<<
L1 C10 'StrLit'>>>>"hello\nworld"<<<<
L3 C1 'IdentName'>>>>c_puts<<<<
L3 C7 'Sep'>>>>:<<<<
L3 C9 'Sep'>>>>(<<<<
L3 C10 'IdentOp'>>>>@<<<<
L3 C11 'Sep'>>>>)<<<<
L3 C12 'IdentName'>>>>·I32<<<<
L3 C18 'IdentOp'>>>>=<<<<
L3 C20 'IdentName'>>>>·extern<<<<
L3 C29 'StrLit'>>>>"puts"<<<<
L5 C1 'IdentName'>>>>main<<<<
L5 C5 'Sep'>>>>:<<<<
L5 C7 'Sep'>>>>(<<<<
L5 C8 'Sep'>>>>)<<<<
L5 C9 'IdentName'>>>>·I32<<<<
L5 C15 'IdentOp'>>>>=<<<<
L5 C17 'Sep'>>>>(<<<<
L5 C18 'Sep'>>>>)<<<<
L6 C5 'IdentName'>>>>c_puts<<<<
L6 C11 'Sep'>>>>(<<<<
L6 C12 'IdentName'>>>>str<<<<
L6 C15 'Sep'>>>>)<<<<
L7 C5 'IdentName'>>>>·ret<<<<
L7 C11 'NumLit'>>>>0<<<<
*/
