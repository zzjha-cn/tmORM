package finder

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	tmorm "tm_orm"
	"tm_orm/impl"
)

type (
	//go:generate mockgen -source=finder.go -destination=../mock/finder.mock.go -package=mocks IFinder[TestUser]
	IFinder[T any] interface {
		FindOne(sess tmorm.MSession, q impl.IBsonQuery, opts ...*options.FindOneOptions) (*T, error)
		Find(sess tmorm.MSession, q impl.IBsonQuery, opts ...*options.FindOptions) ([]*T, error)
		Count(sess tmorm.MSession, q impl.IBsonQuery, opts ...*options.CountOptions) (int64, error)
		Distinct(sess tmorm.MSession, q impl.IBsonQuery, fieldName string, opts ...*options.DistinctOptions) ([]any, error)
	}

	Finder[T any] struct {
	}
)

var (
	_ IFinder[any] = (*Finder[any])(nil)

	FindMtd     tmorm.MethodTyp = "Find"
	FindOneMtd  tmorm.MethodTyp = "FindOne"
	CountMtd    tmorm.MethodTyp = "Count"
	DistinctMtd tmorm.MethodTyp = "Distinct"
)

func (f *Finder[T]) Find(sess tmorm.MSession, q impl.IBsonQuery, opts ...*options.FindOptions) ([]*T, error) {

	var r tmorm.MHandlerFunc = func(mctx *tmorm.MiddleCtx) tmorm.MResult {
		var (
			res    []*T
			filter bson.D
		)
		if mctx.Query != nil {
			filter = mctx.Query.GetBsonD()
		}

		cursor, err := mctx.Session.Conn().Find(mctx.Session.Ctx, filter, opts...)
		if err != nil {
			return tmorm.MResult{
				Val: nil,
				Err: err,
			}
		}

		err = cursor.All(mctx.Session.Ctx, &res)
		cursor.Close(mctx.Session.Ctx)
		if err != nil {
			return tmorm.MResult{
				Val: nil,
				Err: err,
			}
		}

		return tmorm.MResult{
			Val: res,
			Err: err,
		}
	}

	ctx := tmorm.NewMiddleCtx(&sess, FindMtd)
	ctx.Query = q

	res := sess.BuildExecuteChain(r)(ctx)

	if res.Val != nil {
		return res.Val.([]*T), res.Err
	}
	return nil, res.Err
}

func (f *Finder[T]) FindOne(sess tmorm.MSession, q impl.IBsonQuery, opts ...*options.FindOneOptions) (*T, error) {

	var r tmorm.MHandlerFunc = func(mctx *tmorm.MiddleCtx) tmorm.MResult {
		var (
			res    []*T
			filter bson.D
		)
		if mctx.Query != nil {
			filter = mctx.Query.GetBsonD()
		}

		err := mctx.Session.Conn().FindOne(mctx.Session.Ctx, filter, opts...).Decode(&res)
		if err != nil {
			return tmorm.MResult{
				Val: nil,
				Err: err,
			}
		}

		return tmorm.MResult{
			Val: res,
			Err: err,
		}
	}

	ctx := tmorm.NewMiddleCtx(&sess, FindOneMtd)
	ctx.Query = q

	res := sess.BuildExecuteChain(r)(ctx)

	if res.Val != nil {
		return res.Val.(*T), res.Err
	}
	return nil, res.Err
}

func (f *Finder[T]) Count(sess tmorm.MSession, q impl.IBsonQuery, opts ...*options.CountOptions) (int64, error) {
	var r tmorm.MHandlerFunc = func(mctx *tmorm.MiddleCtx) tmorm.MResult {
		var (
			filter bson.D
		)
		if mctx.Query != nil {
			filter = mctx.Query.GetBsonD()
		}

		count, err := mctx.Session.Conn().CountDocuments(mctx.Session.Ctx, filter, opts...)
		if err != nil {
			return tmorm.MResult{
				Val: nil,
				Err: err,
			}
		}

		return tmorm.MResult{
			Val: count,
			Err: err,
		}
	}

	ctx := tmorm.NewMiddleCtx(&sess, CountMtd)
	ctx.Query = q

	res := sess.BuildExecuteChain(r)(ctx)

	if res.Val != nil {
		return res.Val.(int64), res.Err
	}
	return 0, res.Err
}

func (f *Finder[T]) Distinct(sess tmorm.MSession, q impl.IBsonQuery, fieldName string, opts ...*options.DistinctOptions) ([]any, error) {
	var r tmorm.MHandlerFunc = func(mctx *tmorm.MiddleCtx) tmorm.MResult {
		var (
			filter bson.D
		)
		if mctx.Query != nil {
			filter = mctx.Query.GetBsonD()
		}

		dval, err := mctx.Session.Conn().Distinct(mctx.Session.Ctx, fieldName, filter, opts...)
		if err != nil {
			return tmorm.MResult{
				Val: nil,
				Err: err,
			}
		}

		return tmorm.MResult{
			Val: dval,
			Err: err,
		}
	}

	ctx := tmorm.NewMiddleCtx(&sess, DistinctMtd)
	ctx.Query = q
	res := sess.BuildExecuteChain(r)(ctx)

	if res.Val != nil {
		return res.Val.([]any), res.Err
	}
	return nil, res.Err
}
