package updater

import (
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"reflect"
	"strings"
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
		res, _ = makeBsonDByReflect(b.val, b.omiZero)
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

func makeBsonDByReflect(t any, omitZero bool) (bson.D, error) {
	var (
		val = reflect.ValueOf(t)
		typ = reflect.TypeOf(t)
		res = bson.D{}
	)

	for typ.Kind() == reflect.Pointer {
		val = val.Elem()
		typ = typ.Elem()
	}

	if typ.Kind() != reflect.Struct {
		return nil, errors.New("must struct")
	}

	for i := 0; i < typ.NumField(); i++ {
		ftyp := typ.Field(i).Type
		fval := val.Field(i)
		fd := typ.Field(i)
		bsonKey, ok := GetTagString(fd, "bson")
		if ok {
			if ftyp.Kind() == reflect.Pointer {
				if fval.IsNil() {
					continue
				}
				ftyp = ftyp.Elem()
				fval = fval.Elem()
			}

			if ftyp.Kind() == reflect.Struct {
				if subRes, err := makeBsonDByReflect(fval.Interface(), omitZero); err == nil && len(subRes) > 0 {
					res = append(res, bson.E{Key: bsonKey, Value: subRes})
				}
			} else {
				if fval.IsZero() && omitZero {
					// 如果是类型零值，则不处理
					continue
				}
				res = append(res, bson.E{Key: bsonKey, Value: fval.Interface()})

			}
		}
	}
	return res, nil
}

func GetTagString(structTag reflect.StructField, tag string) (string, bool) {
	strcutTag := structTag.Tag
	tagValStr := strcutTag.Get(tag)
	tagValArr := strings.Split(tagValStr, ",")
	if len(tagValArr) == 0 {
		return "", false
	}
	tagVal := tagValArr[0]
	if len(tagVal) == 0 || tagVal == "-" {
		return "", false
	}
	return tagValArr[0], true
}
