package session

import (
	"io"
	"strconv"

	"atmo/util/str"
)

type atFnEager = func(...*AtExpr) (*AtExpr, *SrcFileNotice)
type atFnLazy = func(*AtEnv, []*AtExpr) (*AtEnv, *AtExpr, error)

type AtValType int

const (
	AtValTypeType AtValType = iota
	AtValTypeIdent
	AtValTypeInt
	AtValTypeUint
	AtValTypeFloat
	AtValTypeChar
	AtValTypeStr
	AtValTypeErr
	AtValTypeRec
	AtValTypeArr
	AtValTypeCall
	AtValTypeFunc
)

type AtVal interface {
	valType() AtValType
}

type atValType AtValType
type atValIdent string
type atValInt int64
type atValUint uint64
type atValFloat float64
type atValChar rune
type atValStr string
type atValErr struct{ Err *AtExpr }
type atValRec map[*AtExpr]*AtExpr
type atValArr []*AtExpr
type atValCall []*AtExpr
type atValFn atFnEager
type atValFunc struct {
	params  []*AtExpr // all are guaranteed to be ident before construction
	body    *AtExpr
	env     *AtEnv
	isMacro bool
}

func (atValType) valType() AtValType  { return AtValTypeType }
func (atValIdent) valType() AtValType { return AtValTypeIdent }
func (atValInt) valType() AtValType   { return AtValTypeInt }
func (atValUint) valType() AtValType  { return AtValTypeUint }
func (atValFloat) valType() AtValType { return AtValTypeFloat }
func (atValChar) valType() AtValType  { return AtValTypeChar }
func (atValStr) valType() AtValType   { return AtValTypeStr }
func (atValErr) valType() AtValType   { return AtValTypeErr }
func (atValRec) valType() AtValType   { return AtValTypeRec }
func (atValArr) valType() AtValType   { return AtValTypeArr }
func (atValCall) valType() AtValType  { return AtValTypeCall }
func (atValFn) valType() AtValType    { return AtValTypeFunc }
func (*atValFunc) valType() AtValType { return AtValTypeFunc }

type AtExpr struct {
	SrcNode *AstNode `json:"-"`
	SrcFile *SrcFile `json:"-"`
	Val     AtVal
}

func (me *AtExpr) WriteTo(w io.StringWriter) {
	switch it := me.Val.(type) {
	case atValType:
		w.WriteString(AtValType(it).String())
	case atValIdent:
		w.WriteString(string(it))
	case atValInt:
		w.WriteString(str.FromI64(int64(it), 10))
	case atValUint:
		w.WriteString(str.FromU64(uint64(it), 10))
	case atValFloat:
		w.WriteString(str.FromFloat(float64(it), -1))
	case atValChar:
		w.WriteString(strconv.QuoteRune(rune(it)))
	case atValStr:
		w.WriteString(str.Q(string(it)))
	case atValErr:
		w.WriteString("(@Err ")
		it.Err.WriteTo(w)
		w.WriteString(")")
	case atValRec:
		w.WriteString("{")
		var n int
		for k, v := range it {
			if n > 0 {
				w.WriteString(", ")
			}
			k.WriteTo(w)
			w.WriteString(": ")
			v.WriteTo(w)
			n++
		}
		w.WriteString("}")
	case atValArr:
		w.WriteString("[")
		for i, item := range it {
			if i > 0 {
				w.WriteString(", ")
			}
			item.WriteTo(w)
		}
		w.WriteString("]")
	case atValCall:
		w.WriteString("(")
		for i, item := range it {
			if i > 0 {
				w.WriteString(" ")
			}
			item.WriteTo(w)
		}
		w.WriteString(")")
	case atValFn, *atValFunc:
		w.WriteString(me.SrcNode.Src)
	default:
		panic(it)
	}
}

func (me AtValType) String() string {
	switch me {
	case AtValTypeType:
		return "@Type"
	case AtValTypeIdent:
		return "@Ident"
	case AtValTypeInt:
		return "@Int"
	case AtValTypeUint:
		return "@Uint"
	case AtValTypeFloat:
		return "@Float"
	case AtValTypeChar:
		return "@Char"
	case AtValTypeStr:
		return "@Str"
	case AtValTypeErr:
		return "@Err"
	case AtValTypeRec:
		return "@Rec"
	case AtValTypeArr:
		return "@Arr"
	case AtValTypeCall:
		return "@Call"
	case AtValTypeFunc:
		return "@Func"
	}
	panic(me)
}
