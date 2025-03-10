package query

type (
	// and 中可以嵌套 expr ， expr中又可以嵌套and

	MExpr struct {
		aggCommand
	}
)

//func (m MExpr) AggCmd() aggCommand {
//	am := newAggCommand()
//	//am.m = m.baseCmdBuilder
//	am.m.e.SetKey(tmorm.ExprOp)
//	return am
//}

func (m MExpr) Fd(s string) Field {
	return F(s)
}
func (m MExpr) C() aggCommand {
	res := newAggCommand()
	return res
}

func (m MExpr) Val(s any) AggVal[any] {
	return V(s)
}

func (m MExpr) Er() MExpr {
	mc := newAggCommand()
	b := &Builder{}
	ex := MExpr{mc}
	ex.m.b = b
	return ex
}
