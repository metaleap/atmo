package session

import (
	"atmo/util"
	"atmo/util/sl"
	"atmo/util/str"
	"cmp"
	"fmt"
)

type SrcFileLocs struct {
	File  *SrcFile
	Spans []*SrcFileSpan
	IsSet []bool
	IsGet []bool
}

// SrcFilePos Line and Char both start at 1
type SrcFilePos struct {
	// Line starts at 1
	Line int
	// Char starts at 1
	Char int
}

func (me *SrcFilePos) After(it *SrcFilePos) bool {
	return util.If(me.Line == it.Line, me.Char > it.Char, me.Line > it.Line)
}
func (me *SrcFilePos) AfterOrAt(it *SrcFilePos) bool {
	return util.If(me.Line == it.Line, me.Char >= it.Char, me.Line > it.Line)
}
func (me *SrcFilePos) Before(it *SrcFilePos) bool {
	return util.If(me.Line == it.Line, me.Char < it.Char, me.Line < it.Line)
}
func (me *SrcFilePos) BeforeOrAt(it *SrcFilePos) bool {
	return util.If(me.Line == it.Line, me.Char <= it.Char, me.Line < it.Line)
}
func (me *SrcFilePos) Cmp(to *SrcFilePos) int {
	if me.Line == to.Line {
		return cmp.Compare(me.Char, to.Char)
	}
	return cmp.Compare(me.Line, to.Line)
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
	return it.AfterOrAt(&me.Start) && it.BeforeOrAt(&me.End)
}

func (me *SrcFileSpan) IsSinglePos() bool { return me.Start == me.End }

func (me SrcFileSpan) Eq(to SrcFileSpan) bool {
	return (me.Start == to.Start) && (me.End == to.End)
}

func (me *SrcFileSpan) Expanded(to *SrcFileSpan) *SrcFileSpan {
	if me == to {
		return me
	}
	return &SrcFileSpan{Start: util.If(to.Start.Before(&me.Start), to.Start, me.Start),
		End: util.If(to.End.After(&me.End), to.End, me.End)}
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

func (me *SrcFileSpan) newDiag(kind DiagKind, code DiagCode, args ...any) *Diag {
	return &Diag{Kind: kind, Code: code, Span: *me, Message: errMsg(code, args...)}
}
func (me *SrcFileSpan) newDiagInfo(code DiagCode, args ...any) *Diag {
	return me.newDiag(DiagKindInfo, code, args...)
}
func (me *SrcFileSpan) newDiagHint(code DiagCode, args ...any) *Diag {
	return me.newDiag(DiagKindHint, code, args...)
}
func (me *SrcFileSpan) newDiagWarn(code DiagCode, args ...any) *Diag {
	return me.newDiag(DiagKindWarn, code, args...)
}
func (me *SrcFileSpan) newDiagErr(code DiagCode, args ...any) *Diag {
	return me.newDiag(DiagKindErr, code, args...)
}
func (me *SrcFileSpan) Cmp(to *SrcFileSpan) int {
	return me.Start.Cmp(&to.Start)
}

func (me *SrcFileLocs) SortSpans(reversed bool) {
	me.Spans = sl.SortedPer(me.Spans, func(lhs *SrcFileSpan, rhs *SrcFileSpan) int {
		if reversed {
			lhs, rhs = rhs, lhs
		}
		return lhs.Cmp(rhs)
	})
}
