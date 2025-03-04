package query

import (
	tmorm "tm_orm"

	"go.mongodb.org/mongo-driver/bson"
)

type (
	// and 中可以嵌套 expr ， expr中又可以嵌套and

	MExpr struct {
		aggCommand
	}
)

//func (m MExpr) AggCmd() aggCommand {
//	am := newAggCommand()
//	//am.m = m.baseCommand
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

func (m MExpr) Er() MExpr {
	mc := newAggCommand()
	b := &Builder{}
	ex := MExpr{mc}
	ex.m.b = b
	return ex
}

func (m MExpr) And(ebd ...IAggCommandKey) Builder {
	valArr := make(bson.A, 0, len(ebd))
	for _, i := range ebd {
		switch v := i.(type) {
		case Builder:
			valArr = append(valArr, v.data)
		}
	}
	return m.m.b.KV(tmorm.AndOp, valArr)
}

func (m MExpr) Or(ebd ...IAggCommandKey) Builder {
	valArr := make(bson.A, 0, len(ebd))
	for _, i := range ebd {
		switch v := i.(type) {
		case Builder:
			valArr = append(valArr, v.data)
		}
	}
	return m.m.b.KV(tmorm.OrOp, valArr)
}
