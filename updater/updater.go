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
		UpdateOne(sess tmorm.MSession, upbd IUpdateBuilder, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)
		UpdateMany(sess tmorm.MSession, upbd IUpdateBuilder, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)
		ReplaceOne(sess tmorm.MSession, bd IUpsertBuilder, opts ...*options.ReplaceOptions) (*mongo.UpdateResult, error)
		UpsertOne(sess tmorm.MSession, bd IUpsertBuilder, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)
	}

	IUpdateBuilder interface {
		GetBsonD() bson.D
	}

	IUpsertBuilder interface {
		GetId() (any, bool)
		IUpdateBuilder
	}
)

type (
	MUpdater[T any] struct {
		filter query.IBsonQuery
	}
)

func (p *MUpdater[T]) SetFilter(b query.IBsonQuery) *MUpdater[T] {
	p.filter = b
	return p
}

func (p *MUpdater[T]) CommonFilter(f func(q query.Query) query.IBsonQuery) *MUpdater[T] {
	if f != nil {
		p.filter = f(query.Query{})
	}
	return p
}

// 操作update命令
func (p *MUpdater[T]) C() *MUpdater[T] {
	return p
}

func (p *MUpdater[T]) UpdateOne(sess tmorm.MSession, upbd IUpdateBuilder, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	var (
		fil bson.D
	)
	if p.filter != nil {
		fil = p.filter.GetBsonD()
	}
	return sess.Conn().UpdateOne(sess.Ctx, fil, upbd.GetBsonD(), opts...)
}

func (p *MUpdater[T]) UpdateMany(sess tmorm.MSession, upbd IUpdateBuilder, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	var (
		fil bson.D
	)
	if p.filter != nil {
		fil = p.filter.GetBsonD()
	}
	return sess.Conn().UpdateMany(sess.Ctx, fil, upbd.GetBsonD(), opts...)
}

func (p *MUpdater[T]) ReplaceOne(sess tmorm.MSession, bd IUpsertBuilder, opts ...*options.ReplaceOptions) (*mongo.UpdateResult, error) {
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

	if id, ok := bd.GetId(); ok {
		fil = append(fil, bson.E{"_id", id})
	}

	// 使用replaceOne，不能包含更新操作符。
	updateBsonD := bd.GetBsonD()
	var replaceVal any
	for _, d := range updateBsonD {
		if d.Key == tmorm.SetOp {
			replaceVal = d.Value
			break
		}
	}

	return sess.Conn().ReplaceOne(sess.Ctx, fil, replaceVal, opts...)
}

func (p *MUpdater[T]) UpsertOne(sess tmorm.MSession, bd IUpsertBuilder, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	if len(opts) == 0 {
		opts = append(opts, options.Update().SetUpsert(true))
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

	if id, ok := bd.GetId(); ok {
		fil = append(fil, bson.E{"_id", id})
	}

	updateBsonD := bd.GetBsonD()

	return sess.Conn().UpdateOne(sess.Ctx, fil, updateBsonD, opts...)
}
