package collection

import (
	"context"

	tmorm "github.com/zzjha-cn/tm_orm"
	"github.com/zzjha-cn/tm_orm/expression"
	"github.com/zzjha-cn/tm_orm/impl"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Collection 统一的集合操作接口
type Collection[T any] struct {
	client         *tmorm.ORMClient
	DBName         string
	CollectionName string
	filter         impl.IBsonQuery
	msList         []tmorm.MiddlewareFunc
}

// NewCollection 创建新的集合操作器
func NewCollection[T any](cli *tmorm.ORMClient, db string, collectionName string, msList ...tmorm.MiddlewareFunc) *Collection[T] {
	return &Collection[T]{
		client:         cli,
		DBName:         db,
		CollectionName: collectionName,
		msList:         msList,
	}
}

// Where 添加查询条件 - 支持基础字段条件
func (c *Collection[T]) Where(field string) *CollectionFieldCondition[T] {
	return &CollectionFieldCondition[T]{
		collection: c,
		field:      field,
	}
}

// WhereExpr 使用表达式构建复杂查询条件
func (c *Collection[T]) WhereExpr(expr *expression.Expression) *Collection[T] {
	query := &ExpressionQuery{data: expr.Build()}
	return c.Filter(query)
}

// Expr 创建新的表达式构建器并返回字段表达式
func (c *Collection[T]) Expr(field string) *CollectionExpressionBuilder[T] {
	return &CollectionExpressionBuilder[T]{
		collection: c,
		expr:       expression.NewExpression(),
		field:      field,
	}
}

// ExpressionQuery 表达式查询包装器
type ExpressionQuery struct {
	data bson.M
}

// GetBsonD 实现IBsonQuery接口
func (q *ExpressionQuery) GetBsonD() bson.D {
	var result bson.D
	for k, v := range q.data {
		result = append(result, bson.E{Key: k, Value: v})
	}
	return result
}

// 便捷的表达式构建函数

// And 创建AND表达式组合
func And[T any](c *Collection[T], expressions ...*expression.Expression) *Collection[T] {
	combined := expression.And(expressions...)
	return c.WhereExpr(combined)
}

// Or 创建OR表达式组合
func Or[T any](c *Collection[T], expressions ...*expression.Expression) *Collection[T] {
	combined := expression.Or(expressions...)
	return c.WhereExpr(combined)
}

func Nor[T any](c *Collection[T], expressions ...*expression.Expression) *Collection[T] {
	combined := expression.Nor(expressions...)
	return c.WhereExpr(combined)
}

// Filter 使用查询构建器设置过滤条件
func (c *Collection[T]) Filter(query impl.IBsonQuery) *Collection[T] {
	newCollection := *c
	newCollection.filter = query
	return &newCollection
}
func (c *Collection[T]) combineChain(handle tmorm.MiddlewareFunc) (res []tmorm.MiddlewareFunc) {
	res = append(res, c.client.GetMiddleware()...)
	res = append(res, c.msList...)
	res = append(res, handle)
	return
}

// Find 查询多个文档
func (c *Collection[T]) Find(ctx context.Context, opts ...*options.FindOptions) ([]*T, error) {
	var r tmorm.MiddlewareFunc = func(mctx *tmorm.MiddleCtx, next func(m *tmorm.MiddleCtx)) {
		var (
			res    []*T
			filter = bson.D{}
		)
		if mctx.Operation != nil {
			filter = mctx.Operation.GetBsonD()
		}

		cursor, err := c.client.Database(mctx.DBName).MongoDatabase().Collection(mctx.CollectionName).Find(ctx, filter, opts...)
		if err == nil {
			err = cursor.All(ctx, &res)
			cursor.Close(ctx)
			if err != nil {
				mctx.Result = &tmorm.MResult{Val: nil, Err: err}
			} else {
				mctx.Result = &tmorm.MResult{Val: res, Err: nil}
			}
		} else {
			mctx.Result = &tmorm.MResult{Val: nil, Err: err}
		}

		next(mctx)
	}

	mctx := tmorm.NewMiddleContext(ctx, tmorm.FindMtd, c.DBName, c.CollectionName)
	mctx.Operation = c.filter

	tmorm.Executor(mctx, c.combineChain(r))
	res := mctx.Result
	if res.Val != nil {
		return res.Val.([]*T), res.Err
	}
	return nil, res.Err
}

// FindOne 查询单个文档
func (c *Collection[T]) FindOne(ctx context.Context, opts ...*options.FindOneOptions) (*T, error) {
	var r tmorm.MiddlewareFunc = func(mctx *tmorm.MiddleCtx, next func(m *tmorm.MiddleCtx)) {
		var (
			res    T
			filter = bson.D{}
		)
		if mctx.Operation != nil {
			filter = mctx.Operation.GetBsonD()
		}

		err := c.client.Database(mctx.DBName).MongoDatabase().Collection(mctx.CollectionName).FindOne(ctx, filter, opts...).Decode(&res)
		if err != nil {
			mctx.Result = &tmorm.MResult{Val: nil, Err: err}
		} else {
			mctx.Result = &tmorm.MResult{Val: &res, Err: nil}
		}

		next(mctx)
	}

	mctx := tmorm.NewMiddleContext(ctx, tmorm.FindOneMtd, c.DBName, c.CollectionName)
	mctx.Operation = c.filter

	tmorm.Executor(mctx, c.combineChain(r))
	res := mctx.Result
	if res.Val != nil {
		return res.Val.(*T), res.Err
	}
	return nil, res.Err
}

// Count 统计文档数量
func (c *Collection[T]) Count(ctx context.Context, opts ...*options.CountOptions) (int64, error) {
	var r tmorm.MiddlewareFunc = func(mctx *tmorm.MiddleCtx, next func(m *tmorm.MiddleCtx)) {
		var filter bson.D = bson.D{}
		if mctx.Operation != nil {
			filter = mctx.Operation.GetBsonD()
		}

		count, err := c.client.Database(mctx.DBName).MongoDatabase().Collection(mctx.CollectionName).CountDocuments(ctx, filter, opts...)
		mctx.Result = &tmorm.MResult{Val: count, Err: err}

		next(mctx)
	}

	mctx := tmorm.NewMiddleContext(ctx, tmorm.CountMtd, c.DBName, c.CollectionName)
	mctx.Operation = c.filter

	tmorm.Executor(mctx, c.combineChain(r))
	res := mctx.Result
	return res.Val.(int64), res.Err
}

// Set 设置字段值（用于更新）
func (c *Collection[T]) Set(field string, value any) *UpdateBuilder[T] {
	builder := NewUpdateBuilder[T](c)
	return builder.Set(field, value)
}

// Inc 增加字段值
func (c *Collection[T]) Inc(field string, value any) *UpdateBuilder[T] {
	builder := NewUpdateBuilder[T](c)
	return builder.Inc(field, value)
}

// Unset 删除字段
func (c *Collection[T]) Unset(fields ...string) *UpdateBuilder[T] {
	builder := NewUpdateBuilder[T](c)
	return builder.Unset(fields...)
}

// Insert 插入文档
func (c *Collection[T]) Insert(ctx context.Context, doc *T) (*mongo.InsertOneResult, error) {
	var r tmorm.MiddlewareFunc = func(mctx *tmorm.MiddleCtx, next func(m *tmorm.MiddleCtx)) {
		res, err := c.client.Database(mctx.DBName).MongoDatabase().Collection(mctx.CollectionName).InsertOne(ctx, doc)
		mctx.Result = &tmorm.MResult{Val: res, Err: err}
		next(mctx)
	}

	mctx := tmorm.NewMiddleContext(ctx, tmorm.InsertOneMtd, c.DBName, c.CollectionName)
	tmorm.Executor(mctx, c.combineChain(r))
	res := mctx.Result
	if res.Val != nil {
		return res.Val.(*mongo.InsertOneResult), res.Err
	}
	return nil, res.Err
}

// InsertMany 插入多个文档
func (c *Collection[T]) InsertMany(ctx context.Context, docs []*T) (*mongo.InsertManyResult, error) {
	var r tmorm.MiddlewareFunc = func(mctx *tmorm.MiddleCtx, next func(m *tmorm.MiddleCtx)) {
		interfaces := make([]any, len(docs))
		for i, doc := range docs {
			interfaces[i] = doc
		}
		res, err := c.client.Database(mctx.DBName).MongoDatabase().Collection(mctx.CollectionName).InsertMany(ctx, interfaces)
		mctx.Result = &tmorm.MResult{Val: res, Err: err}
		next(mctx)
	}

	mctx := tmorm.NewMiddleContext(ctx, tmorm.InsertManyMtd, c.DBName, c.CollectionName)
	tmorm.Executor(mctx, c.combineChain(r))
	res := mctx.Result
	if res.Val != nil {
		return res.Val.(*mongo.InsertManyResult), res.Err
	}
	return nil, res.Err
}

// Delete 删除文档
func (c *Collection[T]) Delete(ctx context.Context) (*mongo.DeleteResult, error) {
	var r tmorm.MiddlewareFunc = func(mctx *tmorm.MiddleCtx, next func(m *tmorm.MiddleCtx)) {
		var filter bson.D = bson.D{}
		if mctx.Operation != nil {
			filter = mctx.Operation.GetBsonD()
		}
		res, err := c.client.Database(mctx.DBName).MongoDatabase().Collection(mctx.CollectionName).DeleteMany(ctx, filter)
		mctx.Result = &tmorm.MResult{Val: res, Err: err}
		next(mctx)
	}

	mctx := tmorm.NewMiddleContext(ctx, tmorm.DeleteManyMtd, c.DBName, c.CollectionName)
	mctx.Operation = c.filter
	tmorm.Executor(mctx, c.combineChain(r))
	res := mctx.Result
	if res.Val != nil {
		return res.Val.(*mongo.DeleteResult), res.Err
	}
	return nil, res.Err
}

// DeleteOne 删除单个文档
func (c *Collection[T]) DeleteOne(ctx context.Context) (*mongo.DeleteResult, error) {
	var r tmorm.MiddlewareFunc = func(mctx *tmorm.MiddleCtx, next func(m *tmorm.MiddleCtx)) {
		var filter bson.D = bson.D{}
		if mctx.Operation != nil {
			filter = mctx.Operation.GetBsonD()
		}
		res, err := c.client.Database(mctx.DBName).MongoDatabase().Collection(mctx.CollectionName).DeleteOne(ctx, filter)
		mctx.Result = &tmorm.MResult{Val: res, Err: err}
		next(mctx)
	}

	mctx := tmorm.NewMiddleContext(ctx, tmorm.DeleteOneMtd, c.DBName, c.CollectionName)
	mctx.Operation = c.filter
	tmorm.Executor(mctx, c.combineChain(r))
	res := mctx.Result
	if res.Val != nil {
		return res.Val.(*mongo.DeleteResult), res.Err
	}
	return nil, res.Err
}
