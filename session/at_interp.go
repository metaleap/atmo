package session

type Interp interface {
	Eval(*AtExpr) (*AtExpr, error)
	Parse(src string) (*AtExpr, error)
}

type interp struct {
	SrcFile    *SrcFile
	env        *AtEnv
	stackTrace []string
}

func (me *interp) Eval(*AtExpr) (*AtExpr, error) {
	me.stackTrace = nil
	return nil, nil
}

func (*interp) evalAndApply() {
}

func (*interp) evalExpr() {
}

func (me *interp) Parse(src string) (*AtExpr, error) {
	me.SrcFile.Src.Ast, me.SrcFile.Src.Toks, me.SrcFile.Src.Text = nil, nil, src
	toks, errs := tokenize(me.SrcFile.FilePath, src)
	if len(errs) > 0 {
		return nil, errs[0]
	}
	me.SrcFile.Src.Toks = toks
	me.SrcFile.Src.Ast = me.SrcFile.parse().withoutComments()
	for _, diag := range me.SrcFile.allNotices() {
		if diag.Kind == NoticeKindErr {
			return nil, diag
		}
	}
	return nil, nil
}

func (me *SrcFile) toExpr(node *AstNode) (*AtExpr, error) {
	switch node.Kind {

	}
	return nil, nil
}
