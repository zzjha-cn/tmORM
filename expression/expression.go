package expression

import (
	"regexp"
	"time"
	tmorm "tm_orm"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Expression 表达式构建器
type Expression struct {
	conditions []bson.M
	operator   string // "$and" 或 "$or"
}

// NewExpression 创建新的表达式构建器
func NewExpression() *Expression {
	return &Expression{
		conditions: make([]bson.M, 0),
		operator:   tmorm.AndOp,
	}
}

// And 设置为AND操作
func (e *Expression) And() *Expression {
	e.operator = tmorm.AndOp
	return e
}

// Or 设置为OR操作
func (e *Expression) Or() *Expression {
	e.operator = tmorm.OrOp
	return e
}

// Field 开始构建字段条件
func (e *Expression) Field(name string) *FieldExpression {
	return &FieldExpression{
		parent:    e,
		fieldName: name,
	}
}

// Where 字段条件的别名
func (e *Expression) Where(name string) *FieldExpression {
	return e.Field(name)
}

// AddCondition 添加条件
func (e *Expression) AddCondition(condition bson.M) *Expression {
	e.conditions = append(e.conditions, condition)
	return e
}

// Build 构建最终的BSON查询
func (e *Expression) Build() bson.M {
	if len(e.conditions) == 0 {
		return bson.M{}
	}

	if len(e.conditions) == 1 {
		return e.conditions[0]
	}

	return bson.M{e.operator: e.conditions}
}

// FieldExpression 字段表达式构建器
type FieldExpression struct {
	parent    *Expression
	fieldName string
}

// Eq 等于
func (f *FieldExpression) Eq(value any) *Expression {
	return f.parent.AddCondition(bson.M{f.fieldName: value})
}

// Ne 不等于
func (f *FieldExpression) Ne(value any) *Expression {
	return f.parent.AddCondition(bson.M{f.fieldName: bson.M{tmorm.NeOp: value}})
}

func (f *FieldExpression) Not(expr *Expression) *Expression {
	return f.parent.AddCondition(bson.M{f.fieldName: bson.M{tmorm.NotOp: expr.Build()}})
}

// Gt 大于
func (f *FieldExpression) Gt(value any) *Expression {
	return f.parent.AddCondition(bson.M{f.fieldName: bson.M{tmorm.GtOp: value}})
}

// Gte 大于等于
func (f *FieldExpression) Gte(value any) *Expression {
	return f.parent.AddCondition(bson.M{f.fieldName: bson.M{tmorm.GteOp: value}})
}

// Lt 小于
func (f *FieldExpression) Lt(value any) *Expression {
	return f.parent.AddCondition(bson.M{f.fieldName: bson.M{tmorm.LtOp: value}})
}

// Lte 小于等于
func (f *FieldExpression) Lte(value any) *Expression {
	return f.parent.AddCondition(bson.M{f.fieldName: bson.M{tmorm.LteOp: value}})
}

// In 在列表中
func (f *FieldExpression) In(values ...any) *Expression {
	return f.parent.AddCondition(bson.M{f.fieldName: bson.M{tmorm.InOp: values}})
}

// NotIn 不在列表中
func (f *FieldExpression) NotIn(values ...any) *Expression {
	return f.parent.AddCondition(bson.M{f.fieldName: bson.M{tmorm.NinOp: values}})
}

// Exists 字段存在
func (f *FieldExpression) Exists(exists bool) *Expression {
	return f.parent.AddCondition(bson.M{f.fieldName: bson.M{tmorm.ExistsOp: exists}})
}

// Regex 正则表达式匹配
func (f *FieldExpression) Regex(pattern string, options ...string) *Expression {
	regexValue := bson.M{tmorm.RegexOp: pattern}
	if len(options) > 0 {
		regexValue[tmorm.OptionsOp] = options[0]
	}
	return f.parent.AddCondition(bson.M{f.fieldName: regexValue})
}

// Contains 包含字符串（不区分大小写）
func (f *FieldExpression) Contains(substr string) *Expression {
	return f.Regex(regexp.QuoteMeta(substr), "i")
}

// StartsWith 以字符串开头
func (f *FieldExpression) StartsWith(prefix string) *Expression {
	return f.Regex("^" + regexp.QuoteMeta(prefix))
}

// EndsWith 以字符串结尾
func (f *FieldExpression) EndsWith(suffix string) *Expression {
	return f.Regex(regexp.QuoteMeta(suffix) + "$")
}

// Between 在范围内（包含边界）
func (f *FieldExpression) Between(min, max any) *Expression {
	return f.parent.AddCondition(bson.M{
		f.fieldName: bson.M{
			tmorm.GteOp: min,
			tmorm.LteOp: max,
		},
	})
}

// IsNull 字段为null
func (f *FieldExpression) IsNull() *Expression {
	return f.parent.AddCondition(bson.M{f.fieldName: nil})
}

// IsNotNull 字段不为null
func (f *FieldExpression) IsNotNull() *Expression {
	return f.parent.AddCondition(bson.M{f.fieldName: bson.M{tmorm.NeOp: nil}})
}

// Size 数组大小
func (f *FieldExpression) Size(size int) *Expression {
	return f.parent.AddCondition(bson.M{f.fieldName: bson.M{tmorm.SizeOp: size}})
}

// All 数组包含所有元素
func (f *FieldExpression) All(values ...any) *Expression {
	return f.parent.AddCondition(bson.M{f.fieldName: bson.M{tmorm.AllOp: values}})
}

// ElemMatch 数组元素匹配
func (f *FieldExpression) ElemMatch(condition bson.M) *Expression {
	return f.parent.AddCondition(bson.M{f.fieldName: bson.M{tmorm.ElemMatchOp: condition}})
}

// 时间相关的便利方法

// Today 今天
func (f *FieldExpression) Today() *Expression {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	end := start.Add(24 * time.Hour)
	return f.Between(primitive.NewDateTimeFromTime(start), primitive.NewDateTimeFromTime(end))
}

// Yesterday 昨天
func (f *FieldExpression) Yesterday() *Expression {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, now.Location())
	end := start.Add(24 * time.Hour)
	return f.Between(primitive.NewDateTimeFromTime(start), primitive.NewDateTimeFromTime(end))
}

// ThisWeek 本周
func (f *FieldExpression) ThisWeek() *Expression {
	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7 // 将周日从0改为7
	}
	start := now.AddDate(0, 0, -(weekday - 1))
	start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
	end := start.AddDate(0, 0, 7)
	return f.Between(primitive.NewDateTimeFromTime(start), primitive.NewDateTimeFromTime(end))
}

// ThisMonth 本月
func (f *FieldExpression) ThisMonth() *Expression {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	end := start.AddDate(0, 1, 0)
	return f.Between(primitive.NewDateTimeFromTime(start), primitive.NewDateTimeFromTime(end))
}

// LastNDays 最近N天
func (f *FieldExpression) LastNDays(days int) *Expression {
	now := time.Now()
	start := now.AddDate(0, 0, -days)
	return f.Between(primitive.NewDateTimeFromTime(start), primitive.NewDateTimeFromTime(now))
}

// 复合表达式构建器

// Q 快速创建表达式（类似Django ORM的Q对象）
func Q(field string) *FieldExpression {
	return NewExpression().Field(field)
}

// And 组合多个表达式（AND）
func And(expressions ...*Expression) *Expression {
	result := NewExpression().And()
	for _, expr := range expressions {
		built := expr.Build()
		if len(built) > 0 {
			result.AddCondition(built)
		}
	}
	return result
}

// Or 组合多个表达式（OR）
func Or(expressions ...*Expression) *Expression {
	result := NewExpression().Or()
	for _, expr := range expressions {
		built := expr.Build()
		if len(built) > 0 {
			result.AddCondition(built)
		}
	}
	return result
}

func Nor(expressions ...*Expression) *Expression {
	result := NewExpression()
	result.operator = tmorm.NorOp
	for _, expr := range expressions {
		built := expr.Build()
		if len(built) > 0 {
			result.AddCondition(built)
		}
	}
	return result
}

// Not 否定表达式
func Not(expression *Expression) *Expression {
	built := expression.Build()
	if len(built) == 0 {
		return NewExpression()
	}
	return NewExpression().AddCondition(bson.M{tmorm.NotOp: built})
}
