package session

import (
	"io"
	"strconv"
	"strings"

	"atmo/util/str"
)

type moFnEager = func(args ...*MoExpr) (*MoExpr, *SrcFileNotice)
type moFnLazy = func(ctx *Interp, env *MoEnv, args ...*MoExpr) (*MoEnv, *MoExpr, *SrcFileNotice)

type MoValType int

const (
	MoValTypeType MoValType = iota
	MoValTypeIdent
	MoValTypeInt
	MoValTypeUint
	MoValTypeFloat
	MoValTypeChar
	MoValTypeStr
	MoValTypeErr
	MoValTypeRec
	MoValTypeArr
	MoValTypeCall
	MoValTypeFunc
)

type MoVal interface {
	valType() MoValType
}

type moValType MoValType
type moValIdent string
type moValInt int64
type moValUint uint64
type moValFloat float64
type moValChar rune
type moValStr string
type moValErr struct{ Err *MoExpr }
type moValRec map[*MoExpr]*MoExpr
type moValArr []*MoExpr
type moValCall []*MoExpr
type moValFn moFnEager
type moValFunc struct {
	params  []*MoExpr // all are guaranteed to be ident before construction
	body    *MoExpr
	env     *MoEnv
	isMacro bool
}

func (MoValType) valType() MoValType  { return MoValTypeType }
func (moValIdent) valType() MoValType { return MoValTypeIdent }
func (moValInt) valType() MoValType   { return MoValTypeInt }
func (moValUint) valType() MoValType  { return MoValTypeUint }
func (moValFloat) valType() MoValType { return MoValTypeFloat }
func (moValChar) valType() MoValType  { return MoValTypeChar }
func (moValStr) valType() MoValType   { return MoValTypeStr }
func (moValErr) valType() MoValType   { return MoValTypeErr }
func (moValRec) valType() MoValType   { return MoValTypeRec }
func (moValArr) valType() MoValType   { return MoValTypeArr }
func (moValCall) valType() MoValType  { return MoValTypeCall }
func (moValFn) valType() MoValType    { return MoValTypeFunc }
func (*moValFunc) valType() MoValType { return MoValTypeFunc }

type MoExpr struct {
	SrcNode *AstNode `json:"-"`
	Val     MoVal
}

func (me *MoExpr) String() string {
	var buf strings.Builder
	me.WriteTo(&buf)
	return buf.String()
}

func (me *MoExpr) WriteTo(w io.StringWriter) {
	switch it := me.Val.(type) {
	case MoValType:
		w.WriteString(MoValType(it).String())
	case moValIdent:
		w.WriteString(string(it))
	case moValInt:
		w.WriteString(str.FromI64(int64(it), 10))
	case moValUint:
		w.WriteString(str.FromU64(uint64(it), 10))
	case moValFloat:
		w.WriteString(str.FromFloat(float64(it), -1))
	case moValChar:
		w.WriteString(strconv.QuoteRune(rune(it)))
	case moValStr:
		w.WriteString(str.Q(string(it)))
	case moValErr:
		w.WriteString("(@Err ")
		it.Err.WriteTo(w)
		w.WriteString(")")
	case moValRec:
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
	case moValArr:
		w.WriteString("[")
		for i, item := range it {
			if i > 0 {
				w.WriteString(", ")
			}
			item.WriteTo(w)
		}
		w.WriteString("]")
	case moValCall:
		w.WriteString("(")
		for i, item := range it {
			if i > 0 {
				w.WriteString(" ")
			}
			item.WriteTo(w)
		}
		w.WriteString(")")
	case moValFn, *moValFunc:
		w.WriteString(me.SrcNode.Src)
	default:
		panic(it)
	}
}

func (me MoValType) String() string {
	switch me {
	case MoValTypeType:
		return "@Type"
	case MoValTypeIdent:
		return "@Ident"
	case MoValTypeInt:
		return "@Int"
	case MoValTypeUint:
		return "@Uint"
	case MoValTypeFloat:
		return "@Float"
	case MoValTypeChar:
		return "@Char"
	case MoValTypeStr:
		return "@Str"
	case MoValTypeErr:
		return "@Err"
	case MoValTypeRec:
		return "@Rec"
	case MoValTypeArr:
		return "@Arr"
	case MoValTypeCall:
		return "@Call"
	case MoValTypeFunc:
		return "@Func"
	}
	panic(me)
}
