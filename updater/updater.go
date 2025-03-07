package updater

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	tmorm "tm_orm"
	"tm_orm/impl"
	"tm_orm/query"
)

type (
	//go:generate mockgen -source=updater.go -destination=../mock/updater.mock.go -package=mocks
	IUpdater[T any] interface {
		UpdateOne(sess tmorm.MSession, upbd impl.IUpdateBuilder, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)
		UpdateMany(sess tmorm.MSession, upbd impl.IUpdateBuilder, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)
		ReplaceOne(sess tmorm.MSession, bd impl.IUpsertBuilder, opts ...*options.ReplaceOptions) (*mongo.UpdateResult, error)
		UpsertOne(sess tmorm.MSession, bd impl.IUpsertBuilder, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)
	}
)

var (
	_ IUpdater[any] = (*MUpdater[any])(nil)

	UpdateOneMtd  tmorm.MethodTyp = "UpdateOne"
	UpdateManyMtd tmorm.MethodTyp = "UpdateMany"
	ReplaceOneMtd tmorm.MethodTyp = "ReplaceOne"
	UpsertOneMtd  tmorm.MethodTyp = "UpsertOne"
)

type (
	MUpdater[T any] struct {
		filter impl.IBsonQuery
	}
)

func (p *MUpdater[T]) SetFilter(b impl.IBsonQuery) *MUpdater[T] {
	p.filter = b
	return p
}

func (p *MUpdater[T]) CommonFilter(f func(q query.Query) impl.IBsonQuery) *MUpdater[T] {
	if f != nil {
		p.filter = f(query.Query{})
	}
	return p
}

func (p *MUpdater[T]) UpdateOne(sess tmorm.MSession, upbd impl.IUpdateBuilder, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	var r tmorm.MHandlerFunc = func(mctx *tmorm.MiddleCtx) tmorm.MResult {
		var (
			fil bson.D
			upd bson.D
		)
		if mctx.Query != nil {
			fil = mctx.Query.GetBsonD()
		}
		if mctx.Update != nil {
			upd = mctx.Update.GetBsonD()
		}
		res, err := mctx.Session.Conn().UpdateOne(mctx.Session.Ctx, fil, upd, opts...)
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

	ctx := tmorm.NewMiddleCtx(&sess, UpdateOneMtd)
	ctx.Query = p.filter
	ctx.Update = upbd

	res := sess.BuildExecuteChain(r)(ctx)
	if res.Val != nil {
		return res.Val.(*mongo.UpdateResult), res.Err
	}
	return nil, res.Err
}

func (p *MUpdater[T]) UpdateMany(sess tmorm.MSession, upbd impl.IUpdateBuilder, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	var r tmorm.MHandlerFunc = func(mctx *tmorm.MiddleCtx) tmorm.MResult {
		var (
			fil bson.D
			upd bson.D
		)
		if mctx.Query != nil {
			fil = mctx.Query.GetBsonD()
		}
		if mctx.Update != nil {
			upd = mctx.Update.GetBsonD()
		}
		res, err := mctx.Session.Conn().UpdateMany(mctx.Session.Ctx, fil, upd, opts...)
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

	ctx := tmorm.NewMiddleCtx(&sess, UpdateManyMtd)
	ctx.Query = p.filter
	ctx.Update = upbd

	res := sess.BuildExecuteChain(r)(ctx)
	if res.Val != nil {
		return res.Val.(*mongo.UpdateResult), res.Err
	}
	return nil, res.Err

	//var (
	//	fil bson.D
	//)
	//if p.filter != nil {
	//	fil = p.filter.GetBsonD()
	//}
	//return sess.Conn().UpdateMany(sess.Ctx, fil, upbd.GetBsonD(), opts...)
}

/*
ReplaceOne
replaceOne：

	替换整个文档。
	传入的是一个完整的文档。
	不能使用更新操作符。

updateOne：

	部分更新文档。
	传入的是一个包含更新操作符的文档。
	可以使用$set、$inc、$push等操作符。
*/
func (p *MUpdater[T]) ReplaceOne(sess tmorm.MSession, bd impl.IUpsertBuilder, opts ...*options.ReplaceOptions) (*mongo.UpdateResult, error) {
	if len(opts) == 0 {
		opts = append(opts, options.Replace().SetUpsert(true))
	} else {
		opts[0].SetUpsert(true)
	}

	var r tmorm.MHandlerFunc = func(mctx *tmorm.MiddleCtx) tmorm.MResult {
		var (
			fil         bson.D
			updateBsonD bson.D
		)
		if mctx.Query != nil {
			fil = mctx.Query.GetBsonD()
		}
		if mctx.Upsert != nil {
			updateBsonD = mctx.Upsert.GetBsonD()
		}

		if id, ok := mctx.Upsert.GetId(); ok {
			fil = append(fil, bson.E{"_id", id})
		}

		// 使用replaceOne，不能包含更新操作符。
		var replaceVal any
		for _, d := range updateBsonD {
			if d.Key == tmorm.SetOp {
				replaceVal = d.Value
				break
			}
		}
		res, err := mctx.Session.Conn().ReplaceOne(mctx.Session.Ctx, fil, replaceVal, opts...)
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

	ctx := tmorm.NewMiddleCtx(&sess, ReplaceOneMtd)
	ctx.Query = p.filter
	ctx.Upsert = bd

	res := sess.BuildExecuteChain(r)(ctx)
	if res.Val != nil {
		return res.Val.(*mongo.UpdateResult), res.Err
	}
	return nil, res.Err
}

func (p *MUpdater[T]) UpsertOne(sess tmorm.MSession, bd impl.IUpsertBuilder, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	if len(opts) == 0 {
		opts = append(opts, options.Update().SetUpsert(true))
	} else {
		opts[0].SetUpsert(true)
	}

	var r tmorm.MHandlerFunc = func(mctx *tmorm.MiddleCtx) tmorm.MResult {
		var (
			fil         bson.D
			updateBsonD bson.D
		)
		if mctx.Query != nil {
			fil = mctx.Query.GetBsonD()
		}
		if mctx.Upsert != nil {
			updateBsonD = mctx.Upsert.GetBsonD()
		}

		if id, ok := mctx.Upsert.GetId(); ok {
			fil = append(fil, bson.E{"_id", id})
		}

		res, err := mctx.Session.Conn().UpdateOne(mctx.Session.Ctx, fil, updateBsonD, opts...)
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

	ctx := tmorm.NewMiddleCtx(&sess, UpsertOneMtd)
	ctx.Query = p.filter
	ctx.Upsert = bd

	res := sess.BuildExecuteChain(r)(ctx)
	if res.Val != nil {
		return res.Val.(*mongo.UpdateResult), res.Err
	}
	return nil, res.Err
}
