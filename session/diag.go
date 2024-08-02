package session

import (
	"path/filepath"
	"slices"

	"atmo/util"
	"atmo/util/sl"
	"atmo/util/str"
)

type DiagKind int

const (
	_ DiagKind = iota
	DiagKindErr
	DiagKindWarn
	DiagKindInfo
	DiagKindHint
)

type DiagCode string

const (
	ErrCodeAtmoTodo      DiagCode = "AtmoTodo"
	ErrCodeFileReadError DiagCode = "FileReadError"

	// lexing
	ErrCodeWhitespace  DiagCode = "Whitespace"
	ErrCodeLexingError DiagCode = "LexingError"
	ErrCodeIndentation DiagCode = "Indentation"

	// parsing
	ErrCodeBracketingMismatch DiagCode = "BracketingMismatch"
	ErrCodeLitSyntax          DiagCode = "LitSyntax"

	// semantic (errors)
	ErrCodeExpectedFoo       DiagCode = "Unexpected"
	ErrCodeUndefined         DiagCode = "Unresolved"
	ErrCodeNotAValue         DiagCode = "NotAValue"
	ErrCodeUncallable        DiagCode = "NotCallable"
	ErrCodeReserved          DiagCode = "Reserved"
	ErrCodeNoElseCase        DiagCode = "ElseCaseMissing"
	ErrCodeIndexOutOfBounds  DiagCode = "IndexOutOfBounds"
	ErrCodeRangeNegative     DiagCode = "RangeNegative"
	ErrCodeDictDuplKey       DiagCode = "DictDuplKey"
	ErrCodeNotComparable     DiagCode = "NotComparable"
	ErrCodeNotConvertible    DiagCode = "NotConvertible"
	ErrCodeDuplTopDecl       DiagCode = "DuplTopDecl"
	ErrCodeTypeMismatch      DiagCode = "TypeMismatch"
	ErrCodeTypeInfinite      DiagCode = "TypeInfinite"
	ErrCodeComputationFailed DiagCode = "ComputationFailed"
	ErrCodeCyclic            DiagCode = "Cyclic"

	// semantic (warnings / infos / hints)
	HintCodeUnused DiagCode = "Unused"
)

var (
	allDiags       = map[string]Diags{}
	OnDiagsChanged = func() {}
	OnDbgMsg       = func(showIf bool, fmt string, args ...any) {}
	OnLogMsg       = func(showIf bool, fmt string, args ...any) {}
	errMsgs        = map[DiagCode]string{
		ErrCodeAtmoTodo:      "TODO by Atmo team, please report this detail message: \"%s\"",
		ErrCodeFileReadError: "%s", // actual error msg in %s

		ErrCodeWhitespace:  "unsupported white-space; ensure both: no line-leading tabs, and LF-only line endings (no CR or CRLF)",
		ErrCodeLexingError: "invalid token: %s", // actual error msg in %s
		ErrCodeIndentation: "incorrect indentation",

		ErrCodeLitSyntax:          "invalid literal: %s", // actual error msg in %s
		ErrCodeBracketingMismatch: "opening and closing %s don't match up",

		ErrCodeExpectedFoo:       "expected %s",
		ErrCodeUndefined:         "`%s` is not defined or not in scope",
		ErrCodeNotAValue:         "`%s` cannot be used as a value",
		ErrCodeUncallable:        "`%s` is not callable",
		ErrCodeReserved:          "cannot assign to or define `%s` or any other `%s`-prefixed identifier",
		ErrCodeNoElseCase:        "missing a fallback case",
		ErrCodeIndexOutOfBounds:  "index %d out of bounds, given length %d",
		ErrCodeRangeNegative:     "range end %d is smaller than range start %d",
		ErrCodeDictDuplKey:       "duplicate key `%s` in dict constructor",
		ErrCodeNotComparable:     "operands `%s` and `%s` cannot be compared in %s terms",
		ErrCodeNotConvertible:    "cannot convert `%s` to %s",
		ErrCodeDuplTopDecl:       "top-level declaration `%s` already defined",
		ErrCodeTypeMismatch:      "type mismatch: `%s` vs. `%s`",
		ErrCodeTypeInfinite:      "infinite type detected: `%s`",
		ErrCodeComputationFailed: "%v",
		ErrCodeCyclic:            "cyclic reference: `%s` refers directly or indirectly to itself",

		HintCodeUnused: "code unreachable or without effects (and will be discarded by code generation)",
	}
)

type Diags = sl.Of[*Diag]
type Diag struct {
	Kind    DiagKind
	Message string
	Span    SrcFileSpan `json:"-"`
	Code    DiagCode
	Rel     []*SrcFileLocs `json:",omitempty"`
}

func (me *Diag) equals(to *Diag, includingSpans bool) bool {
	return (me == to) || ((me != nil) && (to != nil) &&
		(me.Code == to.Code) && (me.Kind == to.Kind) && (me.Message == to.Message) &&
		((!includingSpans) || (me.Span.Eq(&to.Span) && sl.Eq(me.Rel, to.Rel, func(l1 *SrcFileLocs, l2 *SrcFileLocs) bool {
			return (l1.File == l2.File) && (sl.Eq(l1.Spans, l2.Spans, (*SrcFileSpan).Eq))
		}))))
}

func (me *Diag) Error() string  { return me.String() }
func (me *Diag) String() string { return str.Fmt("[%s] %s", me.Code, me.Message) }

func (me *Diag) LocStr(srcFilePath string) string {
	if tmp, err := filepath.Rel(".", srcFilePath); (srcFilePath != "") && (err != nil) && (tmp != "") {
		srcFilePath = tmp
	}
	return me.Span.LocStr(srcFilePath)
}

func errMsg(code DiagCode, args ...any) string {
	return str.Trim(str.Fmt(errMsgs[code], args...))
}

func errToDiag(err error, code DiagCode, span SrcFileSpan) *Diag {
	if err == nil {
		return nil
	}
	err_msg, err_msg_fmt := err.Error(), errMsgs[code]
	err_msg = str.Trim(util.If(err_msg_fmt == "", err_msg, str.Fmt(err_msg_fmt, err_msg)))
	return &Diag{Kind: DiagKindErr, Message: err_msg, Code: code, Span: span}
}

func (me *SrcFile) allDiags() (ret Diags) {
	if me.diags.LastReadErr != nil {
		ret.Add(me.diags.LastReadErr)
	}
	ret.Add(me.diags.LexErrs...)
	me.Src.Ast.walk(nil, func(node *AstNode) {
		if node.errParsing != nil {
			ret.Add(node.errParsing)
		}
	})
	if len(ret) == 0 {
		ret.Add(me.diags.Ast2Mo...)
		add := func(it *MoExpr) {
			if it.Diag.Err != nil {
				ret.Add(it.Diag.Err)
			}
		}
		me.pack.Trees.MoOrig.Walk(me, nil, add)
		me.pack.Trees.MoEvaled.Walk(me, nil, add)
		ret.Add(me.pack.Trees.Sem.TopLevel.Errs()...)
		ret.Add(me.pack.semNonErrDiags()...)
	}
	return
}

func (me *SrcPack) semNonErrDiags() (ret Diags) {
	me.Trees.Sem.TopLevel.Walk(nil, nil, func(it *SemExpr) {
		if it.HasFact(SemFactUnused, nil, false, false) {
			ret.Add(it.From.SrcSpan.newDiag(DiagKindHint, HintCodeUnused))
		}
	})
	return
}

// callers have already `sharedState.Lock`ed.
// `force` is ONLY for repl-reset use-case (fully reload pack), NOT to work around any possible/future diags-refresh/diags-pub bugs for LSP clients!
func refreshAndPublishDiags(force bool, provokingFilePaths ...string) {
	if (len(provokingFilePaths) == 0) && !force {
		return
	}
	new_diags := map[string]Diags{}

	for _, src_file_path := range provokingFilePaths {
		var file_diags Diags
		if src_file := state.srcFiles[src_file_path]; src_file != nil {
			file_diags.Add(src_file.allDiags()...)
		}
		new_diags[src_file_path] = file_diags
	}

	// sorting is mainly for the later equality-comparison further down below
	for src_file_path := range new_diags {
		new_diags[src_file_path] = sl.SortedPer(new_diags[src_file_path], func(diag1 *Diag, diag2 *Diag) int {
			return diag1.Span.Cmp(&diag2.Span)
		})
	}

	var have_changes bool
	for src_file_path := range allDiags {
		if _, still_exists := state.srcFiles[src_file_path]; !still_exists {
			have_changes = true
			delete(allDiags, src_file_path)
		}
	}
	for src_file_path, new_diags := range new_diags {
		old_diags := allDiags[src_file_path]
		if !slices.EqualFunc(old_diags, new_diags, func(diag1 *Diag, diag2 *Diag) bool {
			return (diag1.equals(diag2, true))
		}) {
			have_changes = true
			break
		}
	}
	if have_changes {
		for src_file_path, new_diags := range new_diags {
			allDiags[src_file_path] = new_diags
		}
	}

	if have_changes || force {
		go OnDiagsChanged()
	}
}
