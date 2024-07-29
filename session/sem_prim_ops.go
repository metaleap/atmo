package session

type SemPrimOp func(*SrcPack, *SemExpr, MoValCall) *SemExpr

var (
	semPrimValTrue  = &SemExpr{}
	semPrimValFalse = &SemExpr{}
	semPrimValNone  = &SemExpr{}
	semPrimVals     = map[MoValIdent]*SemExpr{
		moValNone.Val.(MoValIdent):  semPrimValNone,
		moValTrue.Val.(MoValIdent):  semPrimValTrue,
		moValFalse.Val.(MoValIdent): semPrimValFalse,
	}
	semPrimOps map[MoValIdent]SemPrimOp
)

func init() {
	semPrimOps = map[MoValIdent]SemPrimOp{
		moPrimOpSet: (*SrcPack).semPrimOpSet,
	}
}

func (me *SrcPack) semPrimOpSet(self *SemExpr, it MoValCall) *SemExpr {
	return nil
}
