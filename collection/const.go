package collection

import (
	tmorm "github.com/zzjha-cn/tm_orm"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	FindMtd     tmorm.MethodTyp = "Find"
	FindOneMtd  tmorm.MethodTyp = "FindOne"
	CountMtd    tmorm.MethodTyp = "Count"
	DistinctMtd tmorm.MethodTyp = "Distinct"

	InsertOneMtd  tmorm.MethodTyp = "InsertOne"
	InsertManyMtd tmorm.MethodTyp = "InsertMany"

	UpdateOneMtd  tmorm.MethodTyp = "UpdateOne"
	UpdateManyMtd tmorm.MethodTyp = "UpdateMany"
	ReplaceOneMtd tmorm.MethodTyp = "ReplaceOne"
	UpsertOneMtd  tmorm.MethodTyp = "UpsertOne"

	DeleteOneMtd  tmorm.MethodTyp = "DeleteOne"
	DeleteManyMtd tmorm.MethodTyp = "DeleteMany"
)

func any2BsonA(list ...any) bson.A {
	res := make(bson.A, 0, len(list))
	for _, a := range list {
		res = append(res, a)
	}
	return res
}
