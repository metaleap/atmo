package session

type AtValType int

const (
	AtValTypeType AtValType = iota
	AtValTypeIdent
	AtValTypeInt
	AtValTypeUint
	AtValTypeFloat
	AtValTypeChar
	AtValTypeStr
	AtValTypeRef
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
type atValRef struct{ To *AtExpr }
type atValErr struct{ Err any }
type atValRec map[*AtExpr]*AtExpr
type atValArr []*AtExpr
type atValCall []*AtExpr
type atValFn func(...*AtVal) (*AtVal, *SrcFileNotice)
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
func (atValRef) valType() AtValType   { return AtValTypeRef }
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
