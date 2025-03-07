package impl

import "go.mongodb.org/mongo-driver/bson"

type (
	IBsonQuery interface {
		GetBsonD() bson.D
	}
)

type (
	IUpdateBuilder interface {
		GetBsonD() bson.D
	}

	IUpsertBuilder interface {
		GetId() (any, bool)
		IUpdateBuilder
	}
)
