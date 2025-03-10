package query

import (
	tmorm "tm_orm"

	"go.mongodb.org/mongo-driver/bson"
)

type (
	// 聚合表达式中
	// 与key组合的复杂命令，如 "{"$gte":["$age",10]}"。 (区别于查询表达式 {"age":{"$gte":10}} )
	aggCommand struct {
		m baseCmdBuilder
	}

	// 命令的 key 的抽象
	IAggCommandField interface {
		cmdFd()
		GetValue() any
	}
	// 字段名称
	Field string
	// 为任意值实现IAggCommandField接口
	AggVal[T any] struct {
		val T
	}
)

func (ch aggCommand) cmdFd() {}

func newAggCommand() aggCommand {
	b := newBaseCommand()
	b.b = &Builder{}
	return aggCommand{b}
}

// ==============================================

func (f Field) cmdFd()        {}
func (f Field) GetValue() any { return string(f) }
func F(s string) Field {
	return Field("$" + s)
}
func (v AggVal[T]) cmdFd()        {}
func (v AggVal[T]) GetValue() any { return v.val }

func V[T any](t T) AggVal[T] {
	return AggVal[T]{
		val: t,
	}
}

// ==============================================
func (ch aggCommand) combineSingleVal(op string, k IAggCommandField) Builder {
	ch.m.b.data = appendBsonD(ch.m.b.data, op, k.GetValue())
	return *ch.m.b
}

func (ch aggCommand) combineBsonKv(op string, k IAggCommandField, v IAggCommandField) Builder {
	valArr := make(bson.A, 2)
	valArr[0] = k.GetValue()
	valArr[1] = v.GetValue()

	ch.m.b.data = appendBsonD(ch.m.b.data, op, valArr)
	return *ch.m.b
}

func (ch aggCommand) combineBsonArray(op string, k ...IAggCommandField) Builder {
	val := make(bson.A, 0, len(k))
	for _, v := range k {
		val = append(val, v.GetValue())
	}

	if len(val) > 0 {
		ch.m.b.data = appendBsonD(ch.m.b.data, op, val)
	}

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

func (ch aggCommand) Gte(k IAggCommandField, v IAggCommandField) Builder {
	return ch.combineBsonKv(tmorm.GteOp, k, v)
}
func (ch aggCommand) Gt(k IAggCommandField, v IAggCommandField) Builder {
	return ch.combineBsonKv(tmorm.GtOp, k, v)
}
func (ch aggCommand) Lte(k IAggCommandField, v IAggCommandField) Builder {
	return ch.combineBsonKv(tmorm.LteOp, k, v)
}
func (ch aggCommand) Lt(k IAggCommandField, v IAggCommandField) Builder {
	return ch.combineBsonKv(tmorm.LtOp, k, v)
}
func (ch aggCommand) Eq(k IAggCommandField, v IAggCommandField) Builder {
	return ch.combineBsonKv(tmorm.EqOp, k, v)
}
func (ch aggCommand) Ne(k IAggCommandField, v IAggCommandField) Builder {
	return ch.combineBsonKv(tmorm.NeOp, k, v)
}

// --------------------算术逻辑----------------

func (ch aggCommand) Abs(k IAggCommandField) Builder {
	return ch.combineSingleVal(tmorm.DivideOp, k)
}

func (ch aggCommand) Floor(k IAggCommandField) Builder {
	return ch.combineSingleVal(tmorm.FloorOp, k)
}

func (ch aggCommand) Divide(k IAggCommandField, v IAggCommandField) Builder {
	return ch.combineBsonKv(tmorm.DivideOp, k, v)
}
func (ch aggCommand) Add(k ...IAggCommandField) Builder {
	return ch.combineBsonArray(tmorm.AddOp, k...)
}
func (ch aggCommand) Subtract(k IAggCommandField, v IAggCommandField) Builder {
	return ch.combineBsonKv(tmorm.SubtractOp, k, v)
}

func (ch aggCommand) Mod(k IAggCommandField, v IAggCommandField) Builder {
	return ch.combineBsonKv(tmorm.ModOp, k, v)
}

func (ch aggCommand) Avg(k ...IAggCommandField) Builder {
	return ch.combineBsonArray(tmorm.AvgOp, k...) // 注意三个点
}

func (ch aggCommand) Sum(k ...IAggCommandField) Builder {
	return ch.combineBsonArray(tmorm.SumOp, k...) // 注意三个点
}

func (ch aggCommand) Multi(k ...IAggCommandField) Builder {
	return ch.combineBsonArray(tmorm.MultiplyOp, k...)
}

func (ch aggCommand) Min(k ...IAggCommandField) Builder {
	return ch.combineBsonArray(tmorm.MinOp, k...) // 注意三个点
}

func (ch aggCommand) Max(k ...IAggCommandField) Builder {
	return ch.combineBsonArray(tmorm.MaxOp, k...) // 注意三个点
}

func (ch aggCommand) First(k IAggCommandField) Builder {
	return ch.combineSingleVal(tmorm.FirstOp, k)
}

func (ch aggCommand) Last(k IAggCommandField) Builder {
	return ch.combineSingleVal(tmorm.LastOp, k)
}

func (ch aggCommand) Push(k IAggCommandField) Builder {
	return ch.combineSingleVal(tmorm.PushOp, k)
}

func (ch aggCommand) AddToSet(k IAggCommandField) Builder {
	return ch.combineSingleVal(tmorm.AddToSetOp, k)
}

//------------------------------------------------------

func (ch aggCommand) ArrayElemAt(k IAggCommandField, v IAggCommandField) Builder {
	return ch.combineBsonKv(tmorm.ArrayElemAtOp, k, v)
}

func (ch aggCommand) ArrayToObj(k IAggCommandField) Builder {
	return ch.combineSingleVal(tmorm.ArrayToObjectOp, k)
}

func (ch aggCommand) ReverseArray(k IAggCommandField) Builder {
	return ch.combineSingleVal(tmorm.ReverseArrayOp, k)
}

func (ch aggCommand) Size(k IAggCommandField) Builder {
	return ch.combineSingleVal(tmorm.SizeOp, k)
}

func (ch aggCommand) ConcatArray(k ...IAggCommandField) Builder {
	return ch.combineBsonArray(tmorm.ConcatArraysOp, k...)
}

func (ch aggCommand) SliceArray(k1, k2, k3 IAggCommandField) Builder {
	return ch.combineBsonArray(tmorm.SliceOp, k1, k2, k3)
}

func (ch aggCommand) Concat(k ...IAggCommandField) Builder {
	return ch.combineBsonArray(tmorm.ConcatOp, k...)
}

func (ch aggCommand) And(k ...IAggCommandField) Builder {
	return ch.combineBsonArray(tmorm.AndOp, k...)
}
func (ch aggCommand) Or(k ...IAggCommandField) Builder {
	return ch.combineBsonArray(tmorm.OrOp, k...)
}

func (ch aggCommand) Type(k IAggCommandField) Builder {
	return ch.combineSingleVal(tmorm.TypeOp, k)
}

func (ch aggCommand) Cond(ifExpr, thenVal, elseVal IAggCommandField) Builder {
	return ch.combineBsonArray(tmorm.CondOp, ifExpr, thenVal, elseVal)
}

//------------------------------------------------------

func (ch aggCommand) In(key IAggCommandField, k ...IAggCommandField) Builder {
	inVal := make(bson.A, 2)
	inVal[0] = key.GetValue()
	val := make(bson.A, 0, len(k))
	for _, v := range k {
		val = append(val, v.GetValue())
	}
	inVal[1] = val

	ch.m.b.data = appendBsonD(ch.m.b.data, tmorm.InOp, inVal)

	return *ch.m.b
}

func (ch aggCommand) NIn(key IAggCommandField, k ...IAggCommandField) Builder {
	inVal := make(bson.A, 2)
	inVal[0] = key.GetValue()
	val := make(bson.A, 0, len(k))
	for _, v := range k {
		val = append(val, v.GetValue())
	}
	inVal[1] = val

	ch.m.b.data = appendBsonD(ch.m.b.data, tmorm.NinOp, inVal)

	return *ch.m.b
}
