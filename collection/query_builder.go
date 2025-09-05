package collection

import (
	tmorm "tm_orm"
	"tm_orm/impl"

	"go.mongodb.org/mongo-driver/bson"
)

// SimpleQueryBuilder 简化的查询构建器
type SimpleQueryBuilder struct {
	filters   []bson.E
	orGroups  [][]bson.E
	currentOr []bson.E
	inOr      bool // 是否在构建or
	inNot     bool // 是否在构建not
}

// NewQuery 创建新的查询构建器
func NewQuery() *SimpleQueryBuilder {
	return &SimpleQueryBuilder{
		filters:  make([]bson.E, 0),
		orGroups: make([][]bson.E, 0),
	}
}

// Where 添加字段条件
func (q *SimpleQueryBuilder) Where(field string) *FieldCondition {
	return &FieldCondition{
		builder: q,
		field:   field,
	}
}
func (q *SimpleQueryBuilder) WhereNot(field string) *FieldCondition {
	f1 := &FieldCondition{
		builder: q,
		field:   field,
	}
	q.inNot = true
	return f1
}

// Or 开始OR条件组
func (q *SimpleQueryBuilder) Or() *SimpleQueryBuilder {
	if q.inOr {
		// 结束当前OR组
		q.orGroups = append(q.orGroups, q.currentOr)
		q.currentOr = make([]bson.E, 0)
	} else {
		// 开始新的OR组
		q.inOr = true
		q.currentOr = make([]bson.E, 0)
	}
	return q
}

// And 结束OR条件组，回到AND模式
func (q *SimpleQueryBuilder) And() *SimpleQueryBuilder {
	if q.inOr {
		q.orGroups = append(q.orGroups, q.currentOr)
		q.currentOr = make([]bson.E, 0)
		q.inOr = false
	}
	return q
}

// Build 构建最终的查询条件
func (q *SimpleQueryBuilder) Build() impl.IBsonQuery {
	// 处理最后的OR组
	if q.inOr && len(q.currentOr) > 0 {
		q.orGroups = append(q.orGroups, q.currentOr)
	}

	var result bson.D

	// 添加普通AND条件
	for _, filter := range q.filters {
		result = append(result, filter)
	}

	// 添加OR条件组
	if len(q.orGroups) > 0 {
		orConditions := make(bson.A, 0)
		for _, orGroup := range q.orGroups {
			if len(orGroup) == 1 {
				orConditions = append(orConditions, bson.D{orGroup[0]})
			} else if len(orGroup) > 1 {
				// 多个条件需要用$and包装
				andGroup := make(bson.A, 0)
				for _, cond := range orGroup {
					andGroup = append(andGroup, bson.D{cond})
				}
				orConditions = append(orConditions, bson.D{{tmorm.AndOp, andGroup}})
			}
		}
		result = append(result, bson.E{Key: tmorm.OrOp, Value: orConditions})
	}

	return &SimpleQuery{data: result}
}

// addCondition 添加条件到当前上下文
func (q *SimpleQueryBuilder) addCondition(field string, operator string, value any) {
	condition := bson.E{
		Key:   field,
		Value: bson.D{{operator, value}},
	}

	if q.inNot {
		// $not 比较特殊。{"age":{"$not":{"$lt":11}}} vs {"age":{"$gte":11}}
		condition.Value = bson.D{{tmorm.NotOp, condition.Value}}
	}

	if q.inOr {
		q.currentOr = append(q.currentOr, condition)
	} else {
		q.filters = append(q.filters, condition)
	}
}

// FieldCondition 字段条件构建器
type FieldCondition struct {
	builder *SimpleQueryBuilder
	field   string
}

// Eq 等于条件
func (f *FieldCondition) Eq(value any) *SimpleQueryBuilder {
	f.builder.addCondition(f.field, tmorm.EqOp, value)
	return f.builder
}

// Ne 不等于条件
func (f *FieldCondition) Ne(value any) *SimpleQueryBuilder {
	f.builder.addCondition(f.field, tmorm.NeOp, value)
	return f.builder
}

// Gt 大于条件
func (f *FieldCondition) Gt(value any) *SimpleQueryBuilder {
	f.builder.addCondition(f.field, tmorm.GtOp, value)
	return f.builder
}

// Gte 大于等于条件
func (f *FieldCondition) Gte(value any) *SimpleQueryBuilder {
	f.builder.addCondition(f.field, tmorm.GteOp, value)
	return f.builder
}

// Lt 小于条件
func (f *FieldCondition) Lt(value any) *SimpleQueryBuilder {
	f.builder.addCondition(f.field, tmorm.LtOp, value)
	return f.builder
}

// Lte 小于等于条件
func (f *FieldCondition) Lte(value any) *SimpleQueryBuilder {
	f.builder.addCondition(f.field, tmorm.LteOp, value)
	return f.builder
}

// In 包含条件
func (f *FieldCondition) In(values ...any) *SimpleQueryBuilder {
	f.builder.addCondition(f.field, tmorm.InOp, any2BsonA(values...))
	return f.builder
}

// NotIn 不包含条件
func (f *FieldCondition) NotIn(values ...any) *SimpleQueryBuilder {
	f.builder.addCondition(f.field, tmorm.NinOp, any2BsonA(values...))
	return f.builder
}

// Regex 正则表达式条件
func (f *FieldCondition) Regex(pattern string) *SimpleQueryBuilder {
	f.builder.addCondition(f.field, tmorm.RegexOp, pattern)
	return f.builder
}

// Exists 字段存在条件
func (f *FieldCondition) Exists(exists bool) *SimpleQueryBuilder {
	f.builder.addCondition(f.field, tmorm.ExistsOp, exists)
	return f.builder
}

type SimpleQuery struct {
	data bson.D
}

// GetBsonD 实现 IBsonQuery 接口
func (q *SimpleQuery) GetBsonD() bson.D {
	return q.data
}
