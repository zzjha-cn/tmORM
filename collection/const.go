package collection

import tmorm "tm_orm"

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
