package updater

import (
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"reflect"
	"strings"
)

type (
	UpdateBuilder[T any] struct {
		Val       *T
		OmitZero  bool
		GetIDFunc func(t *T) any
	}
)

func NewUpdateBuilder[T any](t *T) *UpdateBuilder[T] {
	res := &UpdateBuilder[T]{}
	res.Val = t
	return res
}

func (u *UpdateBuilder[T]) GetId() any {
	return u.GetIDFunc(u.Val)
}

func (u *UpdateBuilder[T]) GetBsonD() bson.D {
	res, _ := makeBsonDByReflect(u.Val, u.OmitZero)
	return res
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
