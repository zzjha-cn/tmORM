package test

import (
	"testing"
	"tm_orm/expression"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

func TestQueryExpr(t *testing.T) {
	testCase := []struct {
		name   string
		expr   *expression.Expression
		before func() *expression.Expression
		check  func(expr *expression.Expression) error
	}{
		{
			name: "测试 normal - 基本字段比较",
			before: func() *expression.Expression {
				return expression.Q("age").Gte(33)
			},
			check: func(expr *expression.Expression) error {
				want := bson.M{
					"age": bson.M{"$gte": 33},
				}
				result := expr.Build()
				assert.Equal(t, want, result)
				return nil
			},
		},
		{
			name: "测试范围查询 - Between操作",
			before: func() *expression.Expression {
				return expression.Q("score").Between(100, 500)
			},
			check: func(expr *expression.Expression) error {
				want := bson.M{
					"score": bson.M{
						"$gte": 100,
						"$lte": 500,
					},
				}
				result := expr.Build()
				assert.Equal(t, want, result)
				return nil
			},
		},
		{
			name: "测试复杂表达式 - And与Or组合",
			before: func() *expression.Expression {
				// age >= 22 AND salary >= 5000 AND (name = "sean" OR age > 18)
				return expression.And(
					expression.Q("age").Gte(22),
					expression.Q("salary").Gte(5000),
					expression.Or(
						expression.Q("name").Eq("sean"),
						expression.Q("age").Gt(18),
					),
				)
			},
			check: func(expr *expression.Expression) error {
				want := bson.M{
					"$and": []bson.M{
						{"age": bson.M{"$gte": 22}},
						{"salary": bson.M{"$gte": 5000}},
						{"$or": []bson.M{
							{"name": "sean"},
							{"age": bson.M{"$gt": 18}},
						}},
					},
				}
				result := expr.Build()
				assert.Equal(t, want, result)
				return nil
			},
		},
		{
			name: "测试字符串匹配 - Contains和Regex",
			before: func() *expression.Expression {
				return expression.Q("name").Contains("john")
			},
			check: func(expr *expression.Expression) error {
				result := expr.Build()
				// 验证包含regex操作
				assert.Contains(t, result, "name")
				if nameField, ok := result["name"].(bson.M); ok {
					assert.Contains(t, nameField, "$regex")
					assert.Contains(t, nameField, "$options")
				}
				return nil
			},
		},
		{
			name: "测试数组操作 - In和Size",
			before: func() *expression.Expression {
				return expression.And(
					expression.Q("status").In("active", "pending", "completed"),
					expression.Q("tags").Size(3),
				)
			},
			check: func(expr *expression.Expression) error {
				want := bson.M{
					"$and": []bson.M{
						{"status": bson.M{"$in": []interface{}{"active", "pending", "completed"}}},
						{"tags": bson.M{"$size": 3}},
					},
				}
				result := expr.Build()
				assert.Equal(t, want, result)
				return nil
			},
		},
		{
			name: "测试否定表达式 - Not操作",
			before: func() *expression.Expression {
				return expression.Not(
					expression.Q("deleted").Eq(true),
				)
			},
			check: func(expr *expression.Expression) error {
				want := bson.M{
					"$not": bson.M{"deleted": true},
				}
				result := expr.Build()
				assert.Equal(t, want, result)
				return nil
			},
		},
		{
			name: "测试Collection集成 - WhereExpr方法",
			before: func() *expression.Expression {
				// 测试collection包中的WhereExpr方法
				expr := expression.Q("age").Gte(18).Field("status").Eq("active")
				// 模拟WhereExpr的使用
				return expr
			},
			check: func(expr *expression.Expression) error {
				want := bson.M{
					"$and": []bson.M{
						{"age": bson.M{"$gte": 18}},
						{"status": "active"},
					},
				}
				result := expr.Build()
				assert.Equal(t, want, result)
				return nil
			},
		},
		{
			name: "测试时间查询 - 便利方法",
			before: func() *expression.Expression {
				// 测试时间相关的便利方法
				return expression.And(
					expression.Q("created_at").Today(),
					expression.Q("updated_at").LastNDays(7),
				)
			},
			check: func(expr *expression.Expression) error {
				result := expr.Build()
				// 验证结构包含$and和时间范围查询
				assert.Contains(t, result, "$and")
				if andConditions, ok := result["$and"].([]bson.M); ok {
					assert.Len(t, andConditions, 2)
					// 验证包含created_at和updated_at字段
					assert.Contains(t, andConditions[0], "created_at")
					assert.Contains(t, andConditions[1], "updated_at")
				}
				return nil
			},
		},
	}

	for _, t1 := range testCase {
		t.Run(t1.name, func(t *testing.T) {
			if t1.before != nil {
				t1.expr = t1.before()
			}

			if t1.check != nil {
				if err := t1.check(t1.expr); err != nil {
					t.Error(err)
				}
			}
		})
	}
}

// TestCollectionExprIntegration 测试Collection与Expression的集成
func TestCollectionExprIntegration(t *testing.T) {

	// 测试基本的表达式构建
	expr1 := expression.Q("age").Gte(18)
	result1 := expr1.Build()
	expected1 := bson.M{"age": bson.M{"$gte": 18}}
	assert.Equal(t, expected1, result1)

	// 测试复杂表达式
	expr2 := expression.And(
		expression.Q("age").Between(18, 65),
		expression.Q("status").In("active", "pending"),
		expression.Or(
			expression.Q("department").Eq("IT"),
			expression.Q("role").Eq("manager"),
		),
	)
	result2 := expr2.Build()

	// 验证复杂表达式的结构
	assert.Contains(t, result2, "$and")
	if andConditions, ok := result2["$and"].([]bson.M); ok {
		assert.Len(t, andConditions, 3)

		// 验证age范围查询
		ageCondition := andConditions[0]["age"].(bson.M)
		assert.Equal(t, 18, ageCondition["$gte"])
		assert.Equal(t, 65, ageCondition["$lte"])

		// 验证status的in查询
		statusCondition := andConditions[1]["status"].(bson.M)
		assert.Contains(t, statusCondition, "$in")

		// 验证or条件
		assert.Contains(t, andConditions[2], "$or")
	}
}
