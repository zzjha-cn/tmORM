package collection

import (
	"context"
	tmorm "tm_orm"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// UpdateBuilder 更新构建器
type UpdateBuilder[T any] struct {
	collection *Collection[T]
	updates    bson.D
}

// Set 继续设置字段
func (u *UpdateBuilder[T]) Set(field string, value any) *UpdateBuilder[T] {
	// 查找是否已有$set操作
	for i, update := range u.updates {
		if update.Key == tmorm.SetOp {
			setDoc := update.Value.(bson.D)
			setDoc = append(setDoc, bson.E{Key: field, Value: value})
			u.updates[i].Value = setDoc
			return u
		}
	}
	// 如果没有$set操作，创建新的
	u.updates = append(u.updates, bson.E{Key: tmorm.SetOp, Value: bson.D{{field, value}}})
	return u
}

// Inc 继续增加字段
func (u *UpdateBuilder[T]) Inc(field string, value any) *UpdateBuilder[T] {
	// 查找是否已有$inc操作
	for i, update := range u.updates {
		if update.Key == tmorm.IncOp {
			incDoc := update.Value.(bson.D)
			incDoc = append(incDoc, bson.E{Key: field, Value: value})
			u.updates[i].Value = incDoc
			return u
		}
	}
	// 如果没有$inc操作，创建新的
	u.updates = append(u.updates, bson.E{Key: tmorm.IncOp, Value: bson.D{{field, value}}})
	return u
}

// Update 执行更新操作
func (u *UpdateBuilder[T]) Update(ctx context.Context, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	var r tmorm.MiddlewareFunc = func(mctx *tmorm.MiddleCtx, next func(m *tmorm.MiddleCtx)) {
		var filter = bson.D{}
		if mctx.Operation != nil {
			filter = mctx.Operation.GetBsonD()
		}

		res, err := u.collection.client.Database(u.collection.DBName).MongoDatabase().Collection(u.collection.CollectionName).UpdateMany(ctx, filter, u.updates, opts...)
		mctx.Result = &tmorm.MResult{Val: res, Err: err}
		next(mctx)
	}

	mctx := tmorm.NewMiddleContext(ctx, "UpdateMany")
	mctx.Operation = u.collection.filter

	tmorm.Executor(mctx, u.collection.combineChain(r))
	res := mctx.Result
	if res.Val != nil {
		return res.Val.(*mongo.UpdateResult), res.Err
	}
	return nil, res.Err
}

// UpdateOne 执行单个文档更新
func (u *UpdateBuilder[T]) UpdateOne(ctx context.Context, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	var r tmorm.MiddlewareFunc = func(mctx *tmorm.MiddleCtx, next func(m *tmorm.MiddleCtx)) {
		var filter = bson.D{}
		if mctx.Operation != nil {
			filter = mctx.Operation.GetBsonD()
		}

		res, err := u.collection.client.Database(u.collection.DBName).MongoDatabase().Collection(u.collection.CollectionName).UpdateOne(ctx, filter, u.updates, opts...)
		mctx.Result = &tmorm.MResult{Val: res, Err: err}
		next(mctx)
	}

	mctx := tmorm.NewMiddleContext(ctx, "UpdateOne")
	mctx.Operation = u.collection.filter

	tmorm.Executor(mctx, u.collection.combineChain(r))
	res := mctx.Result
	if res.Val != nil {
		return res.Val.(*mongo.UpdateResult), res.Err
	}
	return nil, res.Err
}
