package query

import (
	"go.mongodb.org/mongo-driver/bson"
)

type (
	Query struct {
		bd Builder
	}
)

func (q Query) GetBsonD() bson.D {
	return q.bd.data
}

func (q Query) Builder() Builder {
	return q.bd
}
