package query

import (
	"go.mongodb.org/mongo-driver/bson"
)

type (
	EBonsE struct {
		E bson.E
	}

	mongoCmd struct {
		baseCmdBuilder
		//D bson.D
		// 其实bsonE就够了，但是防止后续有需要，还是定成bsonD

		// ！需要将数据中的bsonE转化为bsonD输出，因为mongo-drive中核心支持最小单位为bsonD，所以应该再套一层, 从E转D
	}

	baseCmdBuilder struct {
		e EBonsE   // 数据缓存，存储当前的命令片段，构造完成则将当前片段加入b *Builder中
		b *Builder // 构造链路的引用

		// ！需要将数据中的bsonE转化为bsonD输出，因为mongo-drive中核心支持最小单位为bsonD，所以应该再套一层, 从E转D
	}
)

func newMCommand() mongoCmd {
	return mongoCmd{}
}

func newBaseCommand() baseCmdBuilder {
	return baseCmdBuilder{}
}

func (b baseCmdBuilder) getE() bson.E {
	return b.e.E
}

func (b baseCmdBuilder) getD() bson.D {
	return bson.D{b.e.E}
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
