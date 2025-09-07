package collection

import (
	"context"
	tmorm "tm_orm"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IUpdateOperation interface {
	// GetBsonD 获取 BSON 文档
	GetBsonD() bson.D
	// AddOperation 添加操作
	AddOperation(op string, field string, value any)
	// HasOperation 检查是否存在操作
	HasOperation(op string) bool
	// GetOperation 获取特定操作的值
	GetOperation(op string) (bson.D, bool)
}

// DefaultUpdateOperation 默认更新操作实现
type DefaultUpdateOperation struct {
	updates bson.D
}

// NewUpdateOperation 创建新的更新操作
func NewUpdateOperation() IUpdateOperation {
	return &DefaultUpdateOperation{
		updates: bson.D{},
	}
}

// GetBsonD 获取 BSON 文档
func (d *DefaultUpdateOperation) GetBsonD() bson.D {
	return d.updates
}

// AddOperation 添加操作
func (d *DefaultUpdateOperation) AddOperation(op string, field string, value any) {
	// 查找是否已有该操作
	for i, update := range d.updates {
		if update.Key == op {
			opDoc := update.Value.(bson.D)
			opDoc = append(opDoc, bson.E{Key: field, Value: value})
			d.updates[i].Value = opDoc
			return
		}
	}
	// 如果没有该操作，创建新的
	d.updates = append(d.updates, bson.E{Key: op, Value: bson.D{{field, value}}})
}

// HasOperation 检查是否存在操作
func (d *DefaultUpdateOperation) HasOperation(op string) bool {
	for _, update := range d.updates {
		if update.Key == op {
			return true
		}
	}
	return false
}

// GetOperation 获取特定操作的值
func (d *DefaultUpdateOperation) GetOperation(op string) (bson.D, bool) {
	for _, update := range d.updates {
		if update.Key == op {
			return update.Value.(bson.D), true
		}
	}
	return nil, false
}

// UpdateBuilder 更新构建器
type UpdateBuilder[T any] struct {
	collection *Collection[T]
	updates    IUpdateOperation
}

// NewUpdateBuilder 创建新的更新构建器
func NewUpdateBuilder[T any](collection *Collection[T]) *UpdateBuilder[T] {
	return &UpdateBuilder[T]{
		collection: collection,
		updates:    NewUpdateOperation(),
	}
}

// GetUpdates 获取更新操作接口（用于测试）
func (u *UpdateBuilder[T]) GetUpdates() IUpdateOperation {
	return u.updates
}

// Set 继续设置字段
func (u *UpdateBuilder[T]) Set(field string, value any) *UpdateBuilder[T] {
	u.updates.AddOperation(tmorm.SetOp, field, value)
	return u
}

// Inc 继续增加字段
func (u *UpdateBuilder[T]) Inc(field string, value any) *UpdateBuilder[T] {
	u.updates.AddOperation(tmorm.IncOp, field, value)
	return u
}

// Unset 删除字段
func (u *UpdateBuilder[T]) Unset(fields ...string) *UpdateBuilder[T] {
	for _, field := range fields {
		u.updates.AddOperation(tmorm.UnsetOp, field, "")
	}
	return u
}

// Mul 乘法操作
func (u *UpdateBuilder[T]) Mul(field string, value any) *UpdateBuilder[T] {
	u.updates.AddOperation(tmorm.MulOp, field, value)
	return u
}

// Rename 重命名字段
func (u *UpdateBuilder[T]) Rename(oldField, newField string) *UpdateBuilder[T] {
	u.updates.AddOperation(tmorm.RenameOp, oldField, newField)
	return u
}

// SetOnInsert 仅在插入时设置字段
func (u *UpdateBuilder[T]) SetOnInsert(field string, value any) *UpdateBuilder[T] {
	u.updates.AddOperation(tmorm.SetOnInsertOp, field, value)
	return u
}

// CurrentDate 设置当前日期
func (u *UpdateBuilder[T]) CurrentDate(field string, typeSpec ...any) *UpdateBuilder[T] {
	var value any = true
	if len(typeSpec) > 0 {
		value = typeSpec[0]
	}
	u.updates.AddOperation(tmorm.CurrentDateOp, field, value)
	return u
}

// AddToSet 向数组添加唯一元素
func (u *UpdateBuilder[T]) AddToSet(field string, value any) *UpdateBuilder[T] {
	u.updates.AddOperation(tmorm.AddToSetOp, field, value)
	return u
}

// Push 向数组添加元素
func (u *UpdateBuilder[T]) Push(field string, value any) *UpdateBuilder[T] {
	u.updates.AddOperation(tmorm.PushOp, field, value)
	return u
}

// Pop 从数组移除元素
func (u *UpdateBuilder[T]) Pop(field string, position int) *UpdateBuilder[T] {
	u.updates.AddOperation(tmorm.PopOp, field, position)
	return u
}

// Pull 从数组移除匹配的元素
func (u *UpdateBuilder[T]) Pull(field string, condition any) *UpdateBuilder[T] {
	u.updates.AddOperation(tmorm.PullOp, field, condition)
	return u
}

// PullAll 从数组移除所有匹配的值
func (u *UpdateBuilder[T]) PullAll(field string, values []any) *UpdateBuilder[T] {
	u.updates.AddOperation(tmorm.PullAllOp, field, values)
	return u
}

// Update 执行更新操作
func (u *UpdateBuilder[T]) Update(ctx context.Context, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	var r tmorm.MiddlewareFunc = func(mctx *tmorm.MiddleCtx, next func(m *tmorm.MiddleCtx)) {
		var filter = bson.D{}
		if mctx.Operation != nil {
			filter = mctx.Operation.GetBsonD()
		}

		res, err := u.collection.client.Database(u.collection.DBName).MongoDatabase().Collection(u.collection.CollectionName).UpdateMany(ctx, filter, u.updates.GetBsonD(), opts...)
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

		res, err := u.collection.client.Database(u.collection.DBName).MongoDatabase().Collection(u.collection.CollectionName).UpdateOne(ctx, filter, u.updates.GetBsonD(), opts...)
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
