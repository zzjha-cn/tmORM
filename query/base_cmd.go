package query

import (
	"go.mongodb.org/mongo-driver/bson"
	tmorm "tm_orm"
)

type (
	EBonsE struct {
		E bson.E
	}

	mCommand struct {
		baseCommand
		//D bson.D
		// 其实bsonE就够了，但是防止后续有需要，还是定成bsonD

		// ！需要将数据中的bsonE转化为bsonD输出，因为mongo-drive中核心支持最小单位为bsonD，所以应该再套一层, 从E转D
	}

	baseCommand struct {
		e EBonsE
		b *Builder

		// ！需要将数据中的bsonE转化为bsonD输出，因为mongo-drive中核心支持最小单位为bsonD，所以应该再套一层, 从E转D
	}
)

func newMCommand() mCommand {
	return mCommand{}
}

func newBaseCommand() baseCommand {
	return baseCommand{}
}

func (b baseCommand) getE() bson.E {
	return b.e.E
}

func (b baseCommand) getD() bson.D {
	return bson.D{b.e.E}
}

func (qh mCommand) In(vals ...any) Builder {
	e := EBonsE{}
	e.SetKey(tmorm.InOp)
	e.SetV(vals)
	qh.e.SetEAsBsonD(e.E)
	qh.b.data = append(qh.b.data, qh.e.E)
	return *qh.b
}

func (qh mCommand) NIn(vals ...any) Builder {
	e := EBonsE{}
	e.SetKey(tmorm.NinOp)
	e.SetV(vals)
	qh.e.SetEAsBsonD(e.E)
	qh.b.data = append(qh.b.data, qh.e.E)
	return *qh.b
}

func (qh mCommand) sv(key string, val any) Builder {
	e := EBonsE{}
	e.SetKey(key)
	e.SetV(val)
	qh.e.SetEAsBsonD(e.E)
	qh.b.data = append(qh.b.data, qh.e.E)
	return *qh.b
}

func (qh mCommand) Gte(val any) Builder {
	return qh.sv(tmorm.GteOp, val)
}

func (qh mCommand) Gt(val any) Builder {
	return qh.sv(tmorm.GtOp, val)
}

func (qh mCommand) Lte(val any) Builder {
	return qh.sv(tmorm.LteOp, val)
}

func (qh mCommand) Lt(val any) Builder {
	return qh.sv(tmorm.LtOp, val)
}

func (qh mCommand) Eq(val any) Builder {
	return qh.sv(tmorm.EqOp, val)
}

func (qh mCommand) Ne(val any) Builder {
	return qh.sv(tmorm.NeOp, val)
}

func (qh mCommand) Exists(val any) Builder {
	return qh.sv(tmorm.ExistsOp, val)
}

func (qh mCommand) Regex(val string) Builder {
	return qh.sv(tmorm.RegexOp, val)
}

// =================================================

func (e *EBonsE) SetKey(k string) {
	e.E.Key = k
}
func (e *EBonsE) SetV(v any) {
	e.E.Value = v
}

// 将bsonE转为bsonD再作为EBsonE的value，因为mongo-driver是以bsonD为最小单位调度
func (e *EBonsE) SetEAsBsonD(v bson.E) {
	e.E.Value = bson.D{v}
}
