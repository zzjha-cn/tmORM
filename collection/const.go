package collection

import (
	"go.mongodb.org/mongo-driver/bson"
	tmorm "tm_orm"
)

var (
	FindMtd     tmorm.MethodTyp = "Find"
	FindOneMtd  tmorm.MethodTyp = "FindOne"
	CountMtd    tmorm.MethodTyp = "Count"
	DistinctMtd tmorm.MethodTyp = "Distinct"

	UpdateOneMtd  tmorm.MethodTyp = "UpdateOne"
	UpdateManyMtd tmorm.MethodTyp = "UpdateMany"
	ReplaceOneMtd tmorm.MethodTyp = "ReplaceOne"
	UpsertOneMtd  tmorm.MethodTyp = "UpsertOne"
)

func any2BsonA(list ...any) bson.A {
	res := make(bson.A, 0, len(list))
	for _, a := range list {
		res = append(res, a)
	}
	return res
}
