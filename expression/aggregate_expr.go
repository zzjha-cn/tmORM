package expression

import "go.mongodb.org/mongo-driver/bson"

// $expr中的聚合表达式支持

// AggregationExpression 聚合表达式构建器
type AggregationExpression struct {
	expr any
}

// Field 字段引用
func AggField(name string) *AggregationExpression {
	return &AggregationExpression{expr: "$" + name}
}

// Literal 字面值
func AggLiteral(value any) *AggregationExpression {
	return &AggregationExpression{expr: value}
}

// Add 加法
func (a *AggregationExpression) Add(other *AggregationExpression) *AggregationExpression {
	return &AggregationExpression{expr: bson.M{"$add": []any{a.expr, other.expr}}}
}

// Subtract 减法
func (a *AggregationExpression) Subtract(other *AggregationExpression) *AggregationExpression {
	return &AggregationExpression{expr: bson.M{"$subtract": []any{a.expr, other.expr}}}
}

// Multiply 乘法
func (a *AggregationExpression) Multiply(other *AggregationExpression) *AggregationExpression {
	return &AggregationExpression{expr: bson.M{"$multiply": []any{a.expr, other.expr}}}
}

// Divide 除法
func (a *AggregationExpression) Divide(other *AggregationExpression) *AggregationExpression {
	return &AggregationExpression{expr: bson.M{"$divide": []any{a.expr, other.expr}}}
}

// Mod 取模
func (a *AggregationExpression) Mod(other *AggregationExpression) *AggregationExpression {
	return &AggregationExpression{expr: bson.M{"$mod": []any{a.expr, other.expr}}}
}

// Gt 大于
func (a *AggregationExpression) Gt(other *AggregationExpression) *AggregationExpression {
	return &AggregationExpression{expr: bson.M{"$gt": []any{a.expr, other.expr}}}
}

// Gte 大于等于
func (a *AggregationExpression) Gte(other *AggregationExpression) *AggregationExpression {
	return &AggregationExpression{expr: bson.M{"$gte": []any{a.expr, other.expr}}}
}

// Lt 小于
func (a *AggregationExpression) Lt(other *AggregationExpression) *AggregationExpression {
	return &AggregationExpression{expr: bson.M{"$lt": []any{a.expr, other.expr}}}
}

// Lte 小于等于
func (a *AggregationExpression) Lte(other *AggregationExpression) *AggregationExpression {
	return &AggregationExpression{expr: bson.M{"$lte": []any{a.expr, other.expr}}}
}

// Eq 等于
func (a *AggregationExpression) Eq(other *AggregationExpression) *AggregationExpression {
	return &AggregationExpression{expr: bson.M{"$eq": []any{a.expr, other.expr}}}
}

// Ne 不等于
func (a *AggregationExpression) Ne(other *AggregationExpression) *AggregationExpression {
	return &AggregationExpression{expr: bson.M{"$ne": []any{a.expr, other.expr}}}
}

// And 逻辑与
func AggAnd(expressions ...*AggregationExpression) *AggregationExpression {
	args := make([]any, len(expressions))
	for i, expr := range expressions {
		args[i] = expr.expr
	}
	return &AggregationExpression{expr: bson.M{"$and": args}}
}

// Or 逻辑或
func AggOr(expressions ...*AggregationExpression) *AggregationExpression {
	args := make([]any, len(expressions))
	for i, expr := range expressions {
		args[i] = expr.expr
	}
	return &AggregationExpression{expr: bson.M{"$or": args}}
}

// Cond 条件表达式
func AggCond(condition, ifTrue, ifFalse *AggregationExpression) *AggregationExpression {
	return &AggregationExpression{expr: bson.M{"$cond": []any{condition.expr, ifTrue.expr, ifFalse.expr}}}
}

// GetExpr 获取表达式
func (a *AggregationExpression) GetExpr() any {
	return a.expr
}

// Expr 创建 $expr 查询表达式
func Expr(aggExpr *AggregationExpression) *Expression {
	return NewExpression().AddCondition(bson.M{"$expr": aggExpr.GetExpr()})
}

// 为 FieldExpression 添加 Expr 方法
func (f *FieldExpression) Expr(aggExpr *AggregationExpression) *Expression {
	return f.parent.AddCondition(bson.M{"$expr": aggExpr.GetExpr()})
}
