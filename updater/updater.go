package updater

import (
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	tmorm "tm_orm"
	"tm_orm/query"
)

type (
	//go:generate mockgen -source=updater.go -destination=../mock/updater.mock.go -package=mocks
	IUpdater[T any] interface {
		UpdateOne(sess tmorm.MSession, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)
		UpdateMany(sess tmorm.MSession, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)
		Upsert(sess tmorm.MSession, opts ...*options.ReplaceOptions) (*mongo.UpdateResult, error)
	}

	IUpdateBuilder interface {
		GetBsonD() bson.D
	}

	IUpsertBuilder interface {
		GetId() any
		IUpdateBuilder
	}
)

type (
	MUpdater[T any] struct {
		filter query.IBsonQuery
	}
)

func (p *MUpdater[T]) Filter(b query.IBsonQuery) *MUpdater[T] {
	p.filter = b
	return p
}

// 操作update命令
func (p *MUpdater[T]) C() *MUpdater[T] {
	return p
}

func (p *MUpdater[T]) UpdateOne(sess tmorm.MSession, upbd IUpdateBuilder, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return sess.Conn().UpdateOne(sess.Ctx, bson.D{}, bson.D{}, opts...)
}

func (p *MUpdater[T]) UpdateMany(sess tmorm.MSession, upbd IUpdateBuilder, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	var (
		fil bson.D
	)
	if p.filter != nil {
		fil = p.filter.GetBsonD()
	}
	return sess.Conn().UpdateMany(sess.Ctx, fil, bson.D{}, opts...)
}

func (p *MUpdater[T]) Upsert(sess tmorm.MSession, bd IUpsertBuilder, opts ...*options.ReplaceOptions) (*mongo.UpdateResult, error) {
	if len(opts) == 0 {
		opts = append(opts, options.Replace().SetUpsert(true))
	} else {
		opts[0].SetUpsert(true)
	}

	var (
		fil bson.D
	)
	if p.filter != nil {
		fil = p.filter.GetBsonD()
	}

	if bd == nil {
		return nil, errors.New("need upsert")
	}

	fil = append(fil, bson.E{"_id", bd.GetId()})

	return sess.Conn().ReplaceOne(sess.Ctx, fil, bd.GetBsonD(), opts...)
}
