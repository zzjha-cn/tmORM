package test

import (
	"context"
	"testing"
	"time"

	tmorm "github.com/zzjha-cn/tm_orm"
	"github.com/zzjha-cn/tm_orm/collection"
	"github.com/zzjha-cn/tm_orm/expression"

	"github.com/stretchr/testify/assert"
)

// ExpressionTestUser 测试用户结构
type ExpressionTestUser struct {
	ID        string    `bson:"_id,omitempty"`
	Name      string    `bson:"name"`
	Age       int       `bson:"age"`
	Email     string    `bson:"email"`
	Status    string    `bson:"status"`
	Tags      []string  `bson:"tags"`
	CreatedAt time.Time `bson:"created_at"`
	Salary    int       `bson:"salary"`
}

func TestExpressionIntegration(t *testing.T) {
	client, err := tmorm.NewORMClient(nil)
	assert.NoError(t, err)
	defer client.Close()

	coll := collection.NewCollection[ExpressionTestUser](client, "test_db", "expr_users")
	ctx := context.Background()

	// 清理测试数据
	_, _ = coll.Delete(ctx)

	// 插入测试数据
	testUsers := []*ExpressionTestUser{
		{
			Name:      "John Doe",
			Age:       25,
			Email:     "john@gmail.com",
			Status:    "active",
			Tags:      []string{"developer", "golang"},
			CreatedAt: time.Now(),
			Salary:    75000,
		},
		{
			Name:      "Jane Smith",
			Age:       30,
			Email:     "jane@yahoo.com",
			Status:    "inactive",
			Tags:      []string{"manager", "python"},
			CreatedAt: time.Now().AddDate(0, 0, -1),
			Salary:    85000,
		},
		{
			Name:      "Bob Johnson",
			Age:       35,
			Email:     "bob@gmail.com",
			Status:    "active",
			Tags:      []string{"developer", "java", "golang"},
			CreatedAt: time.Now().AddDate(0, 0, -7),
			Salary:    90000,
		},
	}

	for _, user := range testUsers {
		_, err := coll.Insert(ctx, user)
		assert.NoError(t, err)
	}

	// 定义测试用例结构
	testCases := []struct {
		name     string
		setup    func() ([]*ExpressionTestUser, error)
		expected int
		validate func([]*ExpressionTestUser) bool
	}{
		{
			name: "基础表达式查询 - age > 25",
			setup: func() ([]*ExpressionTestUser, error) {
				return coll.Expr("age").Gt(25).Find(ctx)
			},
			expected: 2,
			validate: func(users []*ExpressionTestUser) bool {
				return len(users) == 2 // Jane 和 Bob
			},
		},
		{
			name: "字符串匹配查询 - name StartsWith John",
			setup: func() ([]*ExpressionTestUser, error) {
				return coll.Expr("name").StartsWith("John").Find(ctx)
			},
			expected: 1,
			validate: func(users []*ExpressionTestUser) bool {
				return len(users) == 1 && users[0].Name == "John Doe"
			},
		},
		{
			name: "字符串匹配查询 - email Contains gmail",
			setup: func() ([]*ExpressionTestUser, error) {
				return coll.Expr("email").Contains("gmail").Find(ctx)
			},
			expected: 2,
			validate: func(users []*ExpressionTestUser) bool {
				return len(users) == 2 // John 和 Bob
			},
		},
		{
			name: "范围查询 - salary Between 70000-80000",
			setup: func() ([]*ExpressionTestUser, error) {
				return coll.Expr("salary").Between(70000, 80000).Find(ctx)
			},
			expected: 1,
			validate: func(users []*ExpressionTestUser) bool {
				return len(users) == 1 && users[0].Name == "John Doe"
			},
		},
		{
			name: "数组查询 - tags Size 3",
			setup: func() ([]*ExpressionTestUser, error) {
				return coll.Expr("tags").Size(3).Find(ctx)
			},
			expected: 1,
			validate: func(users []*ExpressionTestUser) bool {
				return len(users) == 1 && users[0].Name == "Bob Johnson"
			},
		},
		{
			name: "数组查询 - tags All developer,golang",
			setup: func() ([]*ExpressionTestUser, error) {
				return coll.Expr("tags").All("developer", "golang").Find(ctx)
			},
			expected: 2,
			validate: func(users []*ExpressionTestUser) bool {
				return len(users) == 2 // John 和 Bob
			},
		},
		{
			name: "WhereExpr方法 - status=active AND age>=30",
			setup: func() ([]*ExpressionTestUser, error) {
				expr := expression.NewExpression().Field("status").Eq("active").And().Field("age").Gte(30)
				return coll.WhereExpr(expr).Find(ctx)
			},
			expected: 1,
			validate: func(users []*ExpressionTestUser) bool {
				return len(users) == 1 && users[0].Name == "Bob Johnson"
			},
		},
		{
			name: "组合表达式 - And函数",
			setup: func() ([]*ExpressionTestUser, error) {
				expr1 := expression.NewExpression().Field("status").Eq("active")
				expr2 := expression.NewExpression().Field("age").Lt(30)
				return collection.And(coll, expr1, expr2).Find(ctx)
			},
			expected: 1,
			validate: func(users []*ExpressionTestUser) bool {
				return len(users) == 1 && users[0].Name == "John Doe"
			},
		},
		{
			name: "组合表达式 - Or函数",
			setup: func() ([]*ExpressionTestUser, error) {
				expr3 := expression.NewExpression().Field("age").Lt(26)
				expr4 := expression.NewExpression().Field("age").Gt(34)
				return collection.Or(coll, expr3, expr4).Find(ctx)
			},
			expected: 2,
			validate: func(users []*ExpressionTestUser) bool {
				return len(users) == 2 // John (25) 和 Bob (35)
			},
		},
		{
			name: "组合表达式 - Nor函数",
			setup: func() ([]*ExpressionTestUser, error) {
				expr5 := expression.NewExpression().Field("age").Eq(25)
				expr6 := expression.NewExpression().Field("age").Eq(35)
				return collection.Nor(coll, expr5, expr6).Find(ctx)
			},
			expected: 1,
			validate: func(users []*ExpressionTestUser) bool {
				return len(users) == 1
			},
		},
		{
			name: "NOT表达式 - status Not inactive",
			setup: func() ([]*ExpressionTestUser, error) {
				return coll.Where("status").Not().Eq("inactive").Find(ctx)
			},
			expected: 2,
			validate: func(users []*ExpressionTestUser) bool {
				return len(users) == 2 // John 和 Bob (active)
			},
		},
		{
			name: "字段存在性查询 - name Exists",
			setup: func() ([]*ExpressionTestUser, error) {
				return coll.Expr("name").Exists(true).Find(ctx)
			},
			expected: 3,
			validate: func(users []*ExpressionTestUser) bool {
				return len(users) == 3 // 所有用户都有name字段
			},
		},
		{
			name: "字段存在性查询 - email IsNotNull",
			setup: func() ([]*ExpressionTestUser, error) {
				return coll.Expr("email").IsNotNull().Find(ctx)
			},
			expected: 3,
			validate: func(users []*ExpressionTestUser) bool {
				return len(users) == 3 // 所有用户都有email字段
			},
		},
		{
			name: "正则表达式查询 - email Regex gmail",
			setup: func() ([]*ExpressionTestUser, error) {
				return coll.Expr("email").Regex(".*@gmail\\.com$").Find(ctx)
			},
			expected: 2,
			validate: func(users []*ExpressionTestUser) bool {
				return len(users) == 2 // John 和 Bob
			},
		},
	}

	// 循环执行测试用例
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			users, err := tc.setup()
			assert.NoError(t, err)
			assert.Len(t, users, tc.expected)
			assert.True(t, tc.validate(users))
		})
	}

	// 清理测试数据
	_, _ = coll.Delete(ctx)
}

func TestAggregationExpressions(t *testing.T) {
	client, err := tmorm.NewORMClient(nil)
	assert.NoError(t, err)
	defer client.Close()

	coll := collection.NewCollection[ExpressionTestUser](client, "test_db", "expr_users")
	ctx := context.Background()

	// 清理测试数据
	_, _ = coll.Delete(ctx)

	// 插入测试数据
	testUsers := []*ExpressionTestUser{
		{
			Name:   "Alice",
			Age:    25,
			Salary: 50000,
			Status: "active",
		},
		{
			Name:   "Bob",
			Age:    30,
			Salary: 60000,
			Status: "active",
		},
		{
			Name:   "Charlie",
			Age:    35,
			Salary: 70000,
			Status: "inactive",
		},
		{
			Name:   "s1",
			Age:    30,
			Salary: 700,
			Status: "inactive",
		},
	}

	for _, user := range testUsers {
		_, err := coll.Insert(ctx, user)
		assert.NoError(t, err)
	}

	testCases := []struct {
		name   string
		before func()
		check  func(t *testing.T)
	}{
		{
			name:   "$expr 基础数学运算 - 薪资大于年龄*1000",
			before: func() {},
			check: func(t *testing.T) {
				// 查找薪资大于年龄*1000的用户
				aggExpr := expression.AggField("salary").Gt(
					expression.AggField("age").Multiply(expression.AggLiteral(1000)),
				)
				users, err := coll.Expr("_id").Expr(aggExpr).Find(ctx)
				assert.NoError(t, err)
				assert.Len(t, users, 3) // Alice (50000 > 25*1000) 和 Bob (60000 > 30*1000)
			},
		},
		{
			name:   "$expr 复杂逻辑运算 - AND 条件",
			before: func() {},
			check: func(t *testing.T) {
				// 查找薪资大于55000且年龄小于32的用户
				aggExpr := expression.AggAnd(
					expression.AggField("salary").Gt(expression.AggLiteral(55000)),
					expression.AggField("age").Lt(expression.AggLiteral(32)),
				)
				users, err := coll.Expr("_id").Expr(aggExpr).Find(ctx)
				assert.NoError(t, err)
				assert.Len(t, users, 1) // 只有 Bob
				assert.Equal(t, "Bob", users[0].Name)
			},
		},
		{
			name:   "$expr 复杂逻辑运算 - OR 条件",
			before: func() {},
			check: func(t *testing.T) {
				// 查找年龄小于27或薪资大于65000的用户
				aggExpr := expression.AggOr(
					expression.AggField("age").Lt(expression.AggLiteral(27)),
					expression.AggField("salary").Gt(expression.AggLiteral(65000)),
				)
				users, err := coll.Expr("_id").Expr(aggExpr).Find(ctx)
				assert.NoError(t, err)
				assert.Len(t, users, 2) // Alice (age < 27) 和 Charlie (salary > 65000)
			},
		},
		{
			name:   "$expr 数学运算 - 除法和取模",
			before: func() {},
			check: func(t *testing.T) {
				// 查找薪资除以年龄大于1800的用户
				aggExpr := expression.AggField("salary").Divide(
					expression.AggField("age"),
				).Eq(expression.AggLiteral(2000))
				users, err := coll.Expr("_id").Expr(aggExpr).Find(ctx)
				assert.NoError(t, err)
				assert.Len(t, users, 3)
			},
		},
		{
			name:   "$expr 条件表达式 - Cond",
			before: func() {},
			check: func(t *testing.T) {
				// 使用条件表达式：如果年龄>30则薪资应该>65000，否则薪资应该>45000
				condition := expression.AggField("age").Gt(expression.AggLiteral(30))
				ifTrue := expression.AggLiteral(65000)
				ifFalse := expression.AggLiteral(45000)
				aggExpr := expression.AggField("salary").Gt(
					expression.AggCond(condition, ifTrue, ifFalse),
				)
				users, err := coll.Expr("_id").Expr(aggExpr).Find(ctx)
				assert.NoError(t, err)
				assert.Len(t, users, 3) // Alice (25<=30, 50000>45000) 和 Charlie (35>30, 70000>65000)
			},
		},
		{
			name:   "$expr 复杂乘法表达式",
			before: func() {},
			check: func(t *testing.T) {
				aggExpr := expression.AggField("salary").Multiply(
					expression.AggLiteral(10),
				).Gt(expression.AggLiteral(7000))

				// 验证生成的表达式结构
				expr := expression.Expr(aggExpr)
				built := expr.Build()
				t.Logf("Generated expression: %+v", built)

				// 执行查询
				users, err := coll.Expr("_id").Expr(aggExpr).Find(ctx)
				assert.NoError(t, err)
				// 所有用户的 salary * 100 都会 >= 500，因为最小薪资是 50000
				assert.Len(t, users, 3)
			},
		},
		{
			name:   "$expr 使用全局 Expr 函数",
			before: func() {},
			check: func(t *testing.T) {
				// 使用全局 Expr 函数创建表达式
				aggExpr := expression.AggField("salary").Gte(
					expression.AggField("age").Multiply(expression.AggLiteral(2000)),
				)
				expr := expression.Expr(aggExpr)
				users, err := coll.WhereExpr(expr).Find(ctx)
				assert.NoError(t, err)
				assert.Len(t, users, 3) // 只有 Bob (60000 >= 30*2000)
				//assert.Equal(t, "Bob", users[0].Name)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before()
			tc.check(t)
		})
	}

	// 清理测试数据
	_, _ = coll.Delete(ctx)
}
