package collection

import (
	"go.mongodb.org/mongo-driver/bson"
)

func any2BsonA(list ...any) bson.A {
	res := make(bson.A, 0, len(list))
	for _, a := range list {
		res = append(res, a)
	}
	return res
}
