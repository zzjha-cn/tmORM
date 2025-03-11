package aggregator

import (
	tmorm "tm_orm"
	"tm_orm/query"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"go.mongodb.org/mongo-driver/mongo"
)

type (
	Aggregator[T any] struct {
		pl *Pipeline
	}

	Pipeline struct {
		mongoPl mongo.Pipeline
		mt      query.MatchCmd
		group   *query.GroupCmd
	}
)

var (
	AggregateMtd tmorm.MethodTyp = "aggregate"
)

func NewAggregator[T any]() *Aggregator[T] {
	res := &Aggregator[T]{}
	res.pl = NewPipeline()
	return res
}

func (a *Aggregator[T]) GetPipeData() []bson.D {
	return a.pl.mongoPl
}

func (a *Aggregator[T]) Pipe() *Pipeline {
	return a.pl
}

func (a *Aggregator[T]) Aggregate(sess tmorm.MSession, opts ...*options.AggregateOptions) ([]*T, error) {
	var r tmorm.MHandlerFunc = func(mctx *tmorm.MiddleCtx) tmorm.MResult {
		p := a.GetPipeData()
		bs := make(bson.A, 0, len(p))
		for _, p1 := range p {
			bs = append(bs, p1)
		}
		cursor, err := mctx.Session.Conn().Aggregate(mctx.Session.Ctx, bs, opts...)
		if err != nil {
			return tmorm.MResult{
				Val: nil,
				Err: err,
			}
		}
		result := make([]*T, 0)
		err = cursor.All(mctx.Session.Ctx, &result)
		if err != nil {
			return tmorm.MResult{
				Val: nil,
				Err: err,
			}
		}

		return tmorm.MResult{
			Val: result,
			Err: err,
		}
	}

	ctx := tmorm.NewMiddleCtx(&sess, AggregateMtd)

	res := sess.BuildExecuteChain(r)(ctx)
	if res.Err != nil {
		return nil, res.Err
	}
	return res.Val.([]*T), nil
}

func WithParseAggregate[T any, R any](sess tmorm.MSession, a *Aggregator[T], res *([]*R), opts ...*options.AggregateOptions) error {
	var r tmorm.MHandlerFunc = func(mctx *tmorm.MiddleCtx) tmorm.MResult {
		cursor, err := mctx.Session.Conn().Aggregate(mctx.Session.Ctx, a.GetPipeData(), opts...)
		if err != nil {
			return tmorm.MResult{
				Val: nil,
				Err: err,
			}
		}
		qres := make([]*R, 0, 3)
		err = cursor.All(mctx.Session.Ctx, &qres)
		if err != nil {
			return tmorm.MResult{
				Val: nil,
				Err: err,
			}
		}

		*res = qres

		return tmorm.MResult{
			Val: nil,
			Err: err,
		}
	}

	ctx := tmorm.NewMiddleCtx(&sess, AggregateMtd)
	qr := sess.BuildExecuteChain(r)(ctx)
	if qr.Err != nil {
		return qr.Err
	}
	return nil
}
