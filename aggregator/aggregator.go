package aggregator

import (
	"tm_orm/query"

	"go.mongodb.org/mongo-driver/mongo"
)

type (
	Aggregator[T any] struct {
		pl *Pipeline
	}

	Pipeline struct {
		pl    mongo.Pipeline
		mt    query.MatchCmd
		group query.GroupCmd
	}
)

func (a *Aggregator[T]) Pipe() *Pipeline {
	return a.pl
}

// ---------------------------------------------------------
