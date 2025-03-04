package finder

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	tmorm "tm_orm"
	"tm_orm/query"
)

type (
	//go:generate mockgen -source=finder.go -destination=../mock/finder.mock.go -package=mocks IFinder[TestUser]
	IFinder[T any] interface {
		FindOne(sess tmorm.MSession, q query.IBsonQuery, opts ...*options.FindOneOptions) (*T, error)
		Find(sess tmorm.MSession, q query.IBsonQuery, opts ...*options.FindOptions) ([]*T, error)
		Count(sess tmorm.MSession, q query.IBsonQuery, opts ...*options.CountOptions) (int64, error)
		Distinct(sess tmorm.MSession, q query.IBsonQuery, fieldName string, opts ...*options.DistinctOptions) ([]any, error)
	}

	Finder[T any] struct {
	}
)

var (
	_ IFinder[any] = (*Finder[any])(nil)
)

func (f *Finder[T]) Find(sess tmorm.MSession, q query.IBsonQuery, opts ...*options.FindOptions) ([]*T, error) {
	var (
		res    []*T
		filter bson.D = q.GetBsonD()
	)
	cursor, err := sess.Conn().Find(sess.Ctx, filter, opts...)
	if err != nil {
		return nil, err
	}

	defer cursor.Close(sess.Ctx)
	err = cursor.All(sess.Ctx, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (f *Finder[T]) FindOne(sess tmorm.MSession, q query.IBsonQuery, opts ...*options.FindOneOptions) (*T, error) {
	var (
		res    *T
		filter bson.D = q.GetBsonD()
	)

	err := sess.Conn().FindOne(sess.Ctx, filter, opts...).Decode(&res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (f *Finder[T]) Count(sess tmorm.MSession, q query.IBsonQuery, opts ...*options.CountOptions) (int64, error) {
	var (
		filter bson.D = q.GetBsonD()
	)
	return sess.Conn().CountDocuments(sess.Ctx, filter, opts...)
}

func (f *Finder[T]) Distinct(sess tmorm.MSession, q query.IBsonQuery, fieldName string, opts ...*options.DistinctOptions) ([]any, error) {
	var (
		filter bson.D = q.GetBsonD()
	)
	return sess.Conn().Distinct(sess.Ctx, fieldName, filter, opts...)
}
