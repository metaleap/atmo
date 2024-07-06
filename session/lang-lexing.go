package session

import (
	"atmo/util"
	"atmo/util/str"

	"strings"
)

const tokenizerBraceChars = "(){}[]"

type Tokenizer struct {
	strLitDelimChars  string
	sepChars          string
	opChars           string
	lineCommentPrefix string
}

type Tokens []Token

type Token struct {
	kind       TokenKind
	idx        int
	idxLnStart int
	ln0        int
	src        string
}
type TokenKind int

const (
	_ TokenKind = iota
	tokKindIdentName
	tokKindIdentOp
	tokKindSep
	tokKindBrace
	tokKindStrLit
	tokKindNumLit
	tokKindComment
)

func tokenize(src string) (toks Tokens) {
	me := Tokenizer{
		strLitDelimChars:  "'\"`",
		sepChars:          ",;",
		opChars:           ".:$\\?@=<>!&|~^+-*/%",
		lineCommentPrefix: "//",
	}

	var cur Token
	var idxln, lnnr int
	tokdone := func(idxLastChar int) {
		if cur.kind != 0 && idxLastChar >= cur.idx {
			cur.src = src[cur.idx : idxLastChar+'1']
			toks = append(toks, cur)
		}
		cur = Token{}
	}
	for idx := 0; idx < len(src); idx++ {
		c := src[idx]
		switch {
		case cur.kind == tokKindComment: // already in comment?
			if c == '\n' {
				tokdone(idx - 1)
			}
		case cur.kind == tokKindStrLit: // already in string?
			if c == src[cur.idx] && src[idx-1] != '\\' {
				tokdone(idx)
			}
		case cur.kind == tokKindNumLit: // already in number?
			if !(c >= '0' && c <= '9') {
				tokdone(idx - 1)
				idx-- // revisit cur char
				continue
			}
		case cur.kind == tokKindIdentOp: // already in opish ident?
			if strings.IndexByte(me.opChars, c) < 0 {
				tokdone(idx - 1)
				idx-- // revisit cur char
				continue
			}
		case strings.HasPrefix(src[idx:], me.lineCommentPrefix): // start of comment?
			tokdone(idx - 1)
			cur.idx, cur.idxLnStart, cur.ln0, cur.kind = idx, idxln, lnnr, tokKindComment
		case c >= '0' && c <= '9': // start of number?
			if cur.kind != tokKindIdentName {
				tokdone(idx - 1)
				cur.idx, cur.idxLnStart, cur.ln0, cur.kind = idx, idxln, lnnr, tokKindNumLit
			}
		case strings.IndexByte(me.strLitDelimChars, c) >= 0: // start of string?
			tokdone(idx - 1)
			cur.idx, cur.idxLnStart, cur.ln0, cur.kind = idx, idxln, lnnr, tokKindStrLit
		case strings.IndexByte(me.sepChars, c) >= 0: // a sep?
			tokdone(idx - 1)
			cur.idx, cur.idxLnStart, cur.ln0, cur.kind = idx, idxln, lnnr, tokKindSep
			tokdone(idx)
		case strings.IndexByte(tokenizerBraceChars, c) >= 0: // a brace?
			tokdone(idx - 1)
			cur.idx, cur.idxLnStart, cur.ln0, cur.kind = idx, idxln, lnnr, tokKindBrace
			tokdone(idx)
		case strings.IndexByte(me.opChars, c) >= 0: // start of opish ident?
			tokdone(idx - 1)
			cur.idx, cur.idxLnStart, cur.ln0, cur.kind = idx, idxln, lnnr, tokKindIdentOp
		case c == ' ' || c == '\t' || c == '\r' || c == '\n' || c == '\v' || c == '\b':
			tokdone(idx - 1)
		default:
			if cur.kind == 0 {
				cur.idx, cur.idxLnStart, cur.ln0, cur.kind = idx, idxln, lnnr, tokKindIdentName
			}
		}
		if c == '\n' {
			idxln, lnnr = idx+1, lnnr+1
		}
	}
	tokdone(len(src) - 1)
	return
}

func (me *Token) braceMatch() byte {
	idx := strings.IndexByte(tokenizerBraceChars, me.src[0])
	if idx < 0 {
		return 0
	}
	return tokenizerBraceChars[util.If((idx%2) == 0, idx+1, idx-1)]
}

func (me *Token) col0() int {
	return me.idx - me.idxLnStart
}

func (me *Token) Pos() SrcFilePos {
	return SrcFilePos{Line: me.ln0, Char: me.col0()}
}

func (me Tokens) anyAtLevel0(strs ...string) bool {
	var level int
	for i := range me {
		if idx := strings.IndexByte(tokenizerBraceChars, me[i].src[0]); idx >= 0 && (idx%2) == 0 {
			level++
		} else if idx >= 0 && (idx%2) == 1 {
			level--
		} else if level == 0 {
			for _, s := range strs {
				if me[i].src == s {
					return true
				}
			}
		}
	}
	return false
}

func (me Tokens) idxAtLevel0(src string) int {
	var level int
	for i := range me {
		if idx := strings.IndexByte(tokenizerBraceChars, me[i].src[0]); idx >= 0 && (idx%2) == 0 {
			level++
		} else if idx >= 0 && (idx%2) == 1 {
			level--
		} else if level == 0 && me[i].src == src {
			return i
		}
	}
	return -1
}

func (me Tokens) idxOfClosingBrace() int {
	if bracematch := me[0].braceMatch(); bracematch != 0 {
		var level int
		for i := range me {
			if idx := strings.IndexByte(tokenizerBraceChars, me[i].src[0]); idx < 0 || len(me[i].src) > 1 {
				continue
			} else if (idx % 2) == 0 {
				level++
			} else if (idx % 2) == 1 {
				if level--; level == 0 && me[i].src[0] == bracematch {
					return i
				}
			}
		}
	}
	return -1
}

func (me Tokens) indentLevelChunks(col0 int) (ret []Tokens) {
	var idxchunk int
	for i := range me {
		if i > 0 && me[i].col0() == col0 {
			ret = append(ret, me[idxchunk:i])
			idxchunk = i
		}
	}
	if tail := me[idxchunk:]; len(tail) > 0 {
		ret = append(ret, tail)
	}
	// stitch full-line comment lines together, and prefix them to where they should be
	for i := 0; i < len(ret)-1; i++ {
		if ret[i][len(ret[i])-1].kind == tokKindComment &&
			ret[i+1][0].ln0 == ret[i][len(ret[i])-1].ln0+1 {
			ret[i+1] = append(ret[i], ret[i+1]...)
			ret = append(ret[:i], ret[i+1:]...)
			i--
		}
	}
	return
}

func (me Tokens) split(sepsOrOps ...string) (ret []Tokens, sep string, err *SrcFileNotice) {
	var idxlast int
	var level int
	for i := range me {
		if idx := strings.IndexByte(tokenizerBraceChars, me[i].src[0]); idx >= 0 && (idx%2) == 0 {
			level++
		} else if idx >= 0 && (idx%2) == 1 {
			level--
		} else if level == 0 {
			var is bool
			for _, s := range sepsOrOps {
				if is := (me[i].src == s); is {
					if sep == "" {
						sep = s
					} else if sep != s {
						err = &SrcFileNotice{Kind: NoticeKindErr, Code: NoticeCodeNoPrecedence, Span: me.Span(),
							Message: "parenthesize to disambiguate precedence of '" + sep + "' vs. '" + s + "'"}
						return
					}
					break
				}
			}
			if is {
				ret = append(ret, me[idxlast:i])
				idxlast = i + 1
			}
		}
	}
	if tail := me[idxlast:]; len(tail) > 0 {
		ret = append(ret, tail)
	}
	return
}

func (me Tokens) Span() SrcFileSpan {
	return SrcFileSpan{Start: me[0].Pos(), End: me[len(me)-1].Pos()}
}

func (me Tokens) String(origSrc string, maybeErrMsg string) (ret string) {
	last := &me[len(me)-1]
	if len(origSrc) > 0 {
		ret = origSrc[me[0].idx : last.idx+len(last.src)]
	} else {
		for i := range me {
			if i > 0 && me[i].ln0 != me[i-1].ln0 {
				ret += "\n"
			}
			ret += me[i].src + " "
		}
		ret = strings.TrimSpace(ret)
	}
	if maybeErrMsg != "" {
		if max := 44; len(ret) > max {
			ret = ret[:max]
		}
		ret = "L" + str.FromInt(me[0].ln0+1) + "C" + str.FromInt(me[0].col0()+1) + " " + maybeErrMsg + ": " + ret
	}
	return
}

func (me TokenKind) String() string {
	switch me {
	case tokKindIdentName:
		return "IdentName"
	case tokKindIdentOp:
		return "IdentOp"
	case tokKindSep:
		return "Sep"
	case tokKindBrace:
		return "Brace"
	case tokKindStrLit:
		return "StrLit"
	case tokKindNumLit:
		return "NumLit"
	case tokKindComment:
		return "Comment"
	}
	return ""
}
