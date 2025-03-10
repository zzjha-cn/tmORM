package query

import (
	"go.mongodb.org/mongo-driver/bson"
	tmorm "tm_orm"
)

// 查询表达式操作符
// 查询表达式之间是可以嵌套的，但是如果这样子需要将参数抽象为借口，带来使用的不便，所以当前还是以any为参数
// 如果存在嵌套的需求，再自己手写bson完成。

func (qh mongoCmd) In(vals ...any) Builder {
	return qh.sv(tmorm.InOp, any2BsonA(vals...))
}

func (qh mongoCmd) NIn(vals ...any) Builder {
	return qh.sv(tmorm.NinOp, any2BsonA(vals...))
}

func (qh mongoCmd) sv(key string, val any) Builder {
	e := EBonsE{}
	e.SetKey(key)
	e.SetV(val)
	qh.e.SetEAsBsonD(e.E)
	qh.b.data = append(qh.b.data, qh.e.E)
	return *qh.b
}

func (qh mongoCmd) Gte(val any) Builder {
	return qh.sv(tmorm.GteOp, val)
}

func (qh mongoCmd) Gt(val any) Builder {
	return qh.sv(tmorm.GtOp, val)
}

func (qh mongoCmd) Lte(val any) Builder {
	return qh.sv(tmorm.LteOp, val)
}

func (qh mongoCmd) Lt(val any) Builder {
	return qh.sv(tmorm.LtOp, val)
}

func (qh mongoCmd) Eq(val any) Builder {
	return qh.sv(tmorm.EqOp, val)
}

func (qh mongoCmd) Ne(val any) Builder {
	return qh.sv(tmorm.NeOp, val)
}

func (qh mongoCmd) Exists(val any) Builder {
	return qh.sv(tmorm.ExistsOp, val)
}

func (qh mongoCmd) Type(val any) Builder {
	return qh.sv(tmorm.TypeOp, val)
}

func (qh mongoCmd) Regex(val string) Builder {
	return qh.sv(tmorm.RegexOp, val)
}

func (qh mongoCmd) Mod(modNum, modRes any) Builder {
	return qh.sv(tmorm.ModOp, bson.A{modNum, modRes})
}

func (qh mongoCmd) All(vals ...any) Builder {
	return qh.sv(tmorm.ModOp, any2BsonA(vals...))
}

func (qh mongoCmd) ElemMatch(vals ...any) Builder {
	return qh.sv(tmorm.ElemMatchOp, any2BsonA(vals...))
}

func (qh mongoCmd) Size(val any) Builder {
	return qh.sv(tmorm.SizeOp, val)
}
