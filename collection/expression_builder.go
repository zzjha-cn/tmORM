package collection

import (
	"go.mongodb.org/mongo-driver/bson"
	"tm_orm/expression"
)

// CollectionExpressionBuilder 集合表达式构建器
type CollectionExpressionBuilder[T any] struct {
	collection *Collection[T]
	expr       *expression.Expression
	field      string
}

// Eq 等于条件
func (e *CollectionExpressionBuilder[T]) Eq(value any) *Collection[T] {
	expr := e.expr.Field(e.field).Eq(value)
	return e.collection.WhereExpr(expr)
}

// Ne 不等于条件
func (e *CollectionExpressionBuilder[T]) Ne(value any) *Collection[T] {
	expr := e.expr.Field(e.field).Ne(value)
	return e.collection.WhereExpr(expr)
}

// Gt 大于条件
func (e *CollectionExpressionBuilder[T]) Gt(value any) *Collection[T] {
	expr := e.expr.Field(e.field).Gt(value)
	return e.collection.WhereExpr(expr)
}

// Gte 大于等于条件
func (e *CollectionExpressionBuilder[T]) Gte(value any) *Collection[T] {
	expr := e.expr.Field(e.field).Gte(value)
	return e.collection.WhereExpr(expr)
}

// Lt 小于条件
func (e *CollectionExpressionBuilder[T]) Lt(value any) *Collection[T] {
	expr := e.expr.Field(e.field).Lt(value)
	return e.collection.WhereExpr(expr)
}

// Lte 小于等于条件
func (e *CollectionExpressionBuilder[T]) Lte(value any) *Collection[T] {
	expr := e.expr.Field(e.field).Lte(value)
	return e.collection.WhereExpr(expr)
}

// In 包含条件
func (e *CollectionExpressionBuilder[T]) In(values ...any) *Collection[T] {
	expr := e.expr.Field(e.field).In(values...)
	return e.collection.WhereExpr(expr)
}

// NotIn 不包含条件
func (e *CollectionExpressionBuilder[T]) NotIn(values ...any) *Collection[T] {
	expr := e.expr.Field(e.field).NotIn(values...)
	return e.collection.WhereExpr(expr)
}

// Exists 字段存在条件
func (e *CollectionExpressionBuilder[T]) Exists(exists bool) *Collection[T] {
	expr := e.expr.Field(e.field).Exists(exists)
	return e.collection.WhereExpr(expr)
}

// Regex 正则表达式条件
func (e *CollectionExpressionBuilder[T]) Regex(pattern string, options ...string) *Collection[T] {
	expr := e.expr.Field(e.field).Regex(pattern, options...)
	return e.collection.WhereExpr(expr)
}

// Contains 包含子字符串
func (e *CollectionExpressionBuilder[T]) Contains(substr string) *Collection[T] {
	expr := e.expr.Field(e.field).Contains(substr)
	return e.collection.WhereExpr(expr)
}

// StartsWith 以指定前缀开始
func (e *CollectionExpressionBuilder[T]) StartsWith(prefix string) *Collection[T] {
	expr := e.expr.Field(e.field).StartsWith(prefix)
	return e.collection.WhereExpr(expr)
}

// EndsWith 以指定后缀结束
func (e *CollectionExpressionBuilder[T]) EndsWith(suffix string) *Collection[T] {
	expr := e.expr.Field(e.field).EndsWith(suffix)
	return e.collection.WhereExpr(expr)
}

// Between 范围条件
func (e *CollectionExpressionBuilder[T]) Between(min, max any) *Collection[T] {
	expr := e.expr.Field(e.field).Between(min, max)
	return e.collection.WhereExpr(expr)
}

// IsNull 字段为null
func (e *CollectionExpressionBuilder[T]) IsNull() *Collection[T] {
	expr := e.expr.Field(e.field).IsNull()
	return e.collection.WhereExpr(expr)
}

// IsNotNull 字段不为null
func (e *CollectionExpressionBuilder[T]) IsNotNull() *Collection[T] {
	expr := e.expr.Field(e.field).IsNotNull()
	return e.collection.WhereExpr(expr)
}

// Size 数组大小条件
func (e *CollectionExpressionBuilder[T]) Size(size int) *Collection[T] {
	expr := e.expr.Field(e.field).Size(size)
	return e.collection.WhereExpr(expr)
}

// All 数组包含所有指定元素
func (e *CollectionExpressionBuilder[T]) All(values ...any) *Collection[T] {
	expr := e.expr.Field(e.field).All(values...)
	return e.collection.WhereExpr(expr)
}

// ElemMatch 数组元素匹配条件
func (e *CollectionExpressionBuilder[T]) ElemMatch(condition bson.M) *Collection[T] {
	expr := e.expr.Field(e.field).ElemMatch(condition)
	return e.collection.WhereExpr(expr)
}

// Today 今天的日期条件
func (e *CollectionExpressionBuilder[T]) Today() *Collection[T] {
	expr := e.expr.Field(e.field).Today()
	return e.collection.WhereExpr(expr)
}

// Yesterday 昨天的日期条件
func (e *CollectionExpressionBuilder[T]) Yesterday() *Collection[T] {
	expr := e.expr.Field(e.field).Yesterday()
	return e.collection.WhereExpr(expr)
}

// ThisWeek 本周的日期条件
func (e *CollectionExpressionBuilder[T]) ThisWeek() *Collection[T] {
	expr := e.expr.Field(e.field).ThisWeek()
	return e.collection.WhereExpr(expr)
}

// ThisMonth 本月的日期条件
func (e *CollectionExpressionBuilder[T]) ThisMonth() *Collection[T] {
	expr := e.expr.Field(e.field).ThisMonth()
	return e.collection.WhereExpr(expr)
}

// LastNDays 最近N天的日期条件
func (e *CollectionExpressionBuilder[T]) LastNDays(days int) *Collection[T] {
	expr := e.expr.Field(e.field).LastNDays(days)
	return e.collection.WhereExpr(expr)
}

// And 添加AND条件并返回新的表达式构建器
func (e *CollectionExpressionBuilder[T]) And(field string) *CollectionExpressionBuilder[T] {
	return &CollectionExpressionBuilder[T]{
		collection: e.collection,
		expr:       e.expr.And(),
		field:      field,
	}
}

// Or 添加OR条件并返回新的表达式构建器
func (e *CollectionExpressionBuilder[T]) Or(field string) *CollectionExpressionBuilder[T] {
	return &CollectionExpressionBuilder[T]{
		collection: e.collection,
		expr:       e.expr.Or(),
		field:      field,
	}
}

// Expr 使用聚合表达式查询
func (e *CollectionExpressionBuilder[T]) Expr(aggExpr *expression.AggregationExpression) *Collection[T] {
	expr := e.expr.AddCondition(bson.M{"$expr": aggExpr.GetExpr()})
	return e.collection.WhereExpr(expr)
}
