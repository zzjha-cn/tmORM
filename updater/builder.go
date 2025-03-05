package updater

import (
	"go.mongodb.org/mongo-driver/bson"
	"tm_orm/utils"
)

type (
	// update语句构造器
	UpdateBuilder[T any] struct {
		cmd *UpdateCmd[T]
	}

	// upsert语句构造器（仅支持单文档replace）
	ReplaceBuilder[T any] struct {
		GetIDFunc func() (any, bool)
		*UpdateBuilder[T]
	}

	BaseSetBuilder[T any] struct {
		val       *T
		omiZero   bool // 默认为false
		GetIDFunc func(val *T, isValid bool) (any, bool)
	}
)

func (b *BaseSetBuilder[T]) GetId() (any, bool) {
	if b.GetIDFunc != nil {
		return b.GetIDFunc(b.val, b.val != nil)
	}
	return nil, false
}

func (b *BaseSetBuilder[T]) SetOmiZero(o bool) {
	b.omiZero = o
}

func NewBaseSetBuilder[T any](t ...*T) *BaseSetBuilder[T] {
	res := &BaseSetBuilder[T]{}
	if len(t) > 0 {
		res.val = t[0]
	}
	return res
}

func (b *BaseSetBuilder[T]) GetBsonD() bson.D {
	var res bson.D
	if b.val != nil {
		res, _ = utils.MakeBsonDByReflect(b.val, b.omiZero)
	}
	return res
}

// ===========================================

func NewUpdateBuilder[T any]() *UpdateBuilder[T] {
	res := &UpdateBuilder[T]{}
	res.cmd = newUpdateCmd[T]()
	return res
}

func (u *UpdateBuilder[T]) GetBsonD() bson.D {
	return u.cmd.bd.GetData()
}

func (u *UpdateBuilder[T]) C() *UpdateCmd[T] {
	return u.cmd
}

// ===========================================

func NewReplaceBuilder[T any]() *ReplaceBuilder[T] {
	res := &ReplaceBuilder[T]{}
	res.UpdateBuilder = NewUpdateBuilder[T]()
	return res
}

func (u *ReplaceBuilder[T]) GetId() (any, bool) {
	if u.GetIDFunc != nil {
		return u.GetIDFunc()
	}
	return nil, false
}

func (u *ReplaceBuilder[T]) SetGetIdFunc(f func() (any, bool)) {
	u.GetIDFunc = f
}
