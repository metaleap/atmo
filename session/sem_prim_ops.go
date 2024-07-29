package session

type SemPrimOp func(*SrcPack, *SemExpr, MoValCall)

var (
	semPrimOps map[MoValIdent]SemPrimOp
)

func init() {
	semPrimOps = map[MoValIdent]SemPrimOp{
		moPrimOpSet: (*SrcPack).semPrimOpSet,
	}
}

func (me *SrcPack) semPrimOpSet(self *SemExpr, it MoValCall) {
}
