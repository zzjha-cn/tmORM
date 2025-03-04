package query

import (
	tmorm "tm_orm"

	"go.mongodb.org/mongo-driver/bson"
)

type (
	// 聚合表达式中
	// 与key组合的复杂命令，如 "{"$gte":["$age",10]}"。 (区别于 {"age":{"$gte":10}} )
	aggCommand struct {
		m baseCommand
	}

	// 命令的 key 的抽象
	IAggCommandKey interface {
		CmdKey()
	}
	// 字段名称 or 字符串类型的"$$xxx"
	Field string

	// 命令 val 的抽象
	IAggCommandValue interface {
		CmdVal()
	}
)

func (ch aggCommand) CmdVal() {}
func (ch aggCommand) CmdKey() {}

func newAggCommand() aggCommand {
	b := newBaseCommand()
	b.b = &Builder{}
	return aggCommand{b}
}

// ==============================================

func (f Field) CmdKey() {}
func F(s string) Field {
	return Field("$" + s)
}

// ==============================================

func (ch aggCommand) combineBsonKv(op string, k IAggCommandKey, v any) Builder {
	valArr := make(bson.A, 2)
	switch k1 := k.(type) {
	case Field:
		valArr[0] = string(k1)
	case aggCommand:
		valArr[0] = k1.m.getD()
	case Builder:
		valArr[0] = k1.data
	default:
		//valArr[0] = k1 // 去掉，强调key的类型绑定，不允许别的类型进入。
	}

	switch v1 := v.(type) {
	case aggCommand:
		valArr[1] = v1.m.getD()
	case Builder:
		valArr[1] = v1.data
	default:
		valArr[1] = v
	}

	ch.m.b.data = appendBsonD(ch.m.b.data, op, valArr)
	//ch.m.e.SetEAsBsonD(bson.E{
	//	Key:   op,
	//	Value: valArr,
	//})
	return *ch.m.b
}

func (ch aggCommand) combineBsonArray(op string, k IAggCommandKey, v ...any) Builder {
	e1 := bson.E{}
	e1.Key = op
	valArr := make(bson.A, 1)
	switch k1 := k.(type) {
	case Field:
		valArr[0] = string(k1)
	case aggCommand:
		valArr[0] = k1.m.getD()
	case Builder:
		valArr[0] = k1.data
	default:
		//valArr[0] = k1
	}

	//valArr[1] = any2BsonA(v) // 注意三个点
	if len(v) > 0 {
		valArr = append(valArr, any2BsonA(v...))
	}
	ch.m.b.data = appendBsonD(ch.m.b.data, op, valArr)

	return *ch.m.b
}

func any2BsonA(list ...any) bson.A {
	res := make(bson.A, 0, len(list))
	for _, a := range list {
		res = append(res, a)
	}
	return res
}

// -------------------比较命令------------------

func (ch aggCommand) Gte(k IAggCommandKey, v any) Builder {
	return ch.combineBsonKv(tmorm.GteOp, k, v)
}
func (ch aggCommand) Gt(k IAggCommandKey, v any) Builder {
	return ch.combineBsonKv(tmorm.GtOp, k, v)
}
func (ch aggCommand) Lte(k IAggCommandKey, v any) Builder {
	return ch.combineBsonKv(tmorm.LteOp, k, v)
}
func (ch aggCommand) Lt(k IAggCommandKey, v any) Builder {
	return ch.combineBsonKv(tmorm.LtOp, k, v)
}
func (ch aggCommand) Eq(k IAggCommandKey, v any) Builder {
	return ch.combineBsonKv(tmorm.EqOp, k, v)
}
func (ch aggCommand) Ne(k IAggCommandKey, v any) Builder {
	return ch.combineBsonKv(tmorm.NeOp, k, v)
}

// --------------------算术逻辑----------------

func (ch aggCommand) Divide(k IAggCommandKey, v any) Builder {
	return ch.combineBsonKv(tmorm.DivideOp, k, v)
}
func (ch aggCommand) Add(k IAggCommandKey, v any) Builder {
	return ch.combineBsonKv(tmorm.AddOp, k, v)
}
func (ch aggCommand) Subtract(k IAggCommandKey, v any) Builder {
	return ch.combineBsonKv(tmorm.SubtractOp, k, v)
}
func (ch aggCommand) Multi(k IAggCommandKey, v any) Builder {
	return ch.combineBsonKv(tmorm.MultiplyOp, k, v)
}
func (ch aggCommand) Mod(k IAggCommandKey, v any) Builder {
	return ch.combineBsonKv(tmorm.ModOp, k, v)
}

//------------------------------------------------------

//func (ch aggCommand) Cond(ifExpr IAggCommandKey, thenVal, elseVal any) Builder {
//	e1, e2 := bson.E{}, bson.E{}
//	e1.Key = key
//
//	valArr := make(bson.A, 1)
//	switch k1 := ifExpr.(type) {
//	case Field:
//		valArr[0] = string(k1)
//	case aggCommand:
//		valArr[0] = k1.m.getD()
//	case Builder:
//		valArr[0] = k1.data
//	default:
//		//valArr[0] = k1
//	}
//
//	e2.Key = tmorm.CondOp
//	e2.Value = bson.A{
//		ifExpr.m.getD(),
//		thenVal,
//		elseVal,
//	}
//
//	e1.Value = e2
//
//	ch.m.e.SetEAsBsonD(e1)
//	ch.m.b.data = appendBsonD(ch.m.b.data, op, valArr)
//	return *ch.m.b
//}

//------------------------------------------------------

func (ch aggCommand) In(k IAggCommandKey, v ...any) Builder {
	return ch.combineBsonArray(tmorm.InOp, k, v...) // 注意三个点
}
func (ch aggCommand) NIn(k IAggCommandKey, v ...any) Builder {
	return ch.combineBsonArray(tmorm.NinOp, k, v...) // 注意三个点
}
func (ch aggCommand) Avg(k IAggCommandKey, v ...any) Builder {
	return ch.combineBsonArray(tmorm.AvgOp, k, v...) // 注意三个点
}
