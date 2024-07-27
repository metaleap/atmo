package session

import (
	"atmo/util"
	"atmo/util/str"
	"fmt"
)

// SrcFilePos Line and Char both start at 1
type SrcFilePos struct {
	// Line starts at 1
	Line int
	// Char starts at 1
	Char int
}

func (me *SrcFilePos) after(it *SrcFilePos) bool {
	return util.If(me.Line == it.Line, me.Char > it.Char, me.Line > it.Line)
}
func (me *SrcFilePos) afterOrAt(it *SrcFilePos) bool {
	return util.If(me.Line == it.Line, me.Char >= it.Char, me.Line > it.Line)
}
func (me *SrcFilePos) before(it *SrcFilePos) bool {
	return util.If(me.Line == it.Line, me.Char < it.Char, me.Line < it.Line)
}
func (me *SrcFilePos) beforeOrAt(it *SrcFilePos) bool {
	return util.If(me.Line == it.Line, me.Char <= it.Char, me.Line < it.Line)
}
func (me *SrcFilePos) String() string { return str.Fmt("%d,%d", me.Line, me.Char) }
func (me SrcFilePos) ToSpan() (ret SrcFileSpan) {
	ret.Start, ret.End = me, me
	return
}

type SrcFileSpan struct {
	Start SrcFilePos
	End   SrcFilePos
}

func (me SrcFileSpan) Contains(it *SrcFilePos) bool {
	return it.afterOrAt(&me.Start) && it.beforeOrAt(&me.End)
}

func (me *SrcFileSpan) IsSinglePos() bool { return me.Start == me.End }

func (me SrcFileSpan) Eq(to SrcFileSpan) bool {
	return (me.Start == to.Start) && (me.End == to.End)
}

func (me *SrcFileSpan) Expanded(to *SrcFileSpan) *SrcFileSpan {
	if me == to {
		return me
	}
	return &SrcFileSpan{Start: util.If(to.Start.before(&me.Start), to.Start, me.Start),
		End: util.If(to.End.after(&me.End), to.End, me.End)}
}

func (me SrcFileSpan) String() string {
	if me.IsSinglePos() {
		return me.Start.String()
	}
	return str.Fmt("%s-%s", me.Start.String(), me.End.String())
}

func srcSpanFrom(exprs MoExprs) (ret *SrcFileSpan) {
	for _, expr := range exprs {
		if expr.SrcSpan != nil {
			if ret == nil {
				ret = util.Ptr(*expr.SrcSpan)
			} else {
				ret.End = expr.SrcSpan.End
			}
		}
	}
	return
}

func (me *SrcFile) Span() (ret SrcFileSpan) {
	ret.Start, ret.End = SrcFilePos{Line: 1, Char: 1}, SrcFilePos{Line: 1, Char: 1}
	for i := 0; i < len(me.Src.Text); i++ {
		if me.Src.Text[i] == '\n' {
			ret.End.Line++
		}
	}
	if (me.Src.Text != "") && (me.Src.Text[len(me.Src.Text)-1] != '\n') {
		ret.End.Line++
	}
	return
}

func (me SrcFileSpan) LocStr(srcFilePath string) string {
	if srcFilePath == "" {
		return me.String()
	}
	return fmt.Sprintf("%s:%s", srcFilePath, me.String())
}
func (me Toks) LocStr(srcFilePath string) string { return me.Span().LocStr(srcFilePath) }

func (me *SrcFileSpan) newDiag(kind SrcFileNoticeKind, code SrcFileNoticeCode, args ...any) *SrcFileNotice {
	return &SrcFileNotice{Kind: kind, Code: code, Span: *me, Message: errMsg(code, args...)}
}
func (me *SrcFileSpan) newDiagInfo(code SrcFileNoticeCode, args ...any) *SrcFileNotice {
	return me.newDiag(NoticeKindInfo, code, args...)
}
func (me *SrcFileSpan) newDiagHint(code SrcFileNoticeCode, args ...any) *SrcFileNotice {
	return me.newDiag(NoticeKindHint, code, args...)
}
func (me *SrcFileSpan) newDiagWarn(code SrcFileNoticeCode, args ...any) *SrcFileNotice {
	return me.newDiag(NoticeKindWarn, code, args...)
}
func (me *SrcFileSpan) newDiagErr(code SrcFileNoticeCode, args ...any) *SrcFileNotice {
	return me.newDiag(NoticeKindErr, code, args...)
}
