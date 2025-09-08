package test

import (
	"context"
	tmorm "github.com/zzjha-cn/tm_orm"
	"github.com/zzjha-cn/tm_orm/collection"
	"github.com/zzjha-cn/tm_orm/expression"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TestUser2 struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	Name       string             `bson:"name"`
	Email      string             `bson:"email"`
	Age        int64              `bson:"age"`
	Status     string             `bson:"status"`
	Department string             `bson:"department"`
	Tags       []string           `bson:"tags"`
	Salary     float64            `bson:"salary"`
	CreatedAt  time.Time          `bson:"created_at"`
	IsActive   bool               `bson:"is_active"`
}

// 产品测试结构
type TestProduct struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	Name     string             `bson:"name"`
	Price    float64            `bson:"price"`
	Category string             `bson:"category"`
	InStock  bool               `bson:"in_stock"`
}

// 端到端测试collection finder功能

func TestFinderE2E(t *testing.T) {
	ctx := context.Background()
	ormClient, _ := tmorm.NewORMClient(nil)
	coll := collection.NewCollection[TestUser](ormClient, "mytest", "db_test")

	// 插入测试数据
	testUsers := []*TestUser{
		{
			ID:   primitive.NewObjectID(),
			Name: "sean",
			Age:  20,
		},
		{
			ID:   primitive.NewObjectID(),
			Name: "alice",
			Age:  25,
		},
		{
			ID:   primitive.NewObjectID(),
			Name: "bob",
			Age:  30,
		},
	}

	for _, user := range testUsers {
		_, err := coll.Insert(ctx, user)
		assert.NoError(t, err)
	}

	testCases := []struct {
		name   string
		before func() error
		check  func() error
	}{
		{
			name: "基本查询测试 - 按名字查询",
			check: func() error {
				results, err := coll.Where("name").Eq("sean").Find(ctx)
				assert.NoError(t, err)
				assert.Len(t, results, 1)
				assert.Equal(t, "sean", results[0].Name)
				return nil
			},
		},
		{
			name: "基本查询测试 - 按年龄范围查询",
			check: func() error {
				results, err := coll.Where("age").Gte(25).Find(ctx)
				assert.NoError(t, err)
				assert.Len(t, results, 2) // alice and bob
				return nil
			},
		},
		{
			name: "复合查询测试 - 使用原生bson",
			check: func() error {
				// 使用原生bson进行复杂查询
				bsonD := bson.D{{
					"$or", bson.A{
						bson.D{{"name", "sean"}},
						bson.D{{"age", bson.M{"$gt": 25}}},
					},
				}}

				// 创建自定义查询实例
				query := &CustomQuery{data: bsonD}
				results, err := coll.Filter(query).Find(ctx)
				assert.NoError(t, err)
				assert.Len(t, results, 2) // sean and bob
				return nil
			},
		},
		{
			name: "表达式查询测试 - 使用$expr进行数学运算",
			check: func() error {
				// 测试表达式：age * 1 >= 25的用户
				filter := bson.M{
					"$expr": bson.M{
						"$gte": []any{
							bson.M{"$multiply": []any{"$age", 1}},
							25,
						},
					},
				}

				// 将bson.M转换为bson.D
				var bsonD bson.D
				for k, v := range filter {
					bsonD = append(bsonD, bson.E{Key: k, Value: v})
				}

				// 创建自定义查询实例
				query := &CustomQuery{data: bsonD}
				results, err := coll.Filter(query).Find(ctx)
				assert.NoError(t, err)
				assert.Len(t, results, 2) // alice and bob (age >= 25)
				return nil
			},
		},
		{
			name: "计数测试 - 测试总数",
			check: func() error {
				//count, err := coll.Where("age").Gt(1).Count(ctx)
				count, err := coll.Count(ctx)
				assert.NoError(t, err)
				assert.Equal(t, int64(3), count)
				return nil
			},
		},
		{
			name: "计数测试 - 测试条件计数",
			check: func() error {
				count, err := coll.Where("age").Gt(20).Count(ctx)
				assert.NoError(t, err)
				assert.Equal(t, int64(2), count) // alice and bob
				return nil
			},
		},
		{
			name: "单个文档查询测试 - 测试FindOne",
			check: func() error {
				user, err := coll.Where("name").Eq("alice").FindOne(ctx)
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, "alice", user.Name)
				assert.Equal(t, int64(25), user.Age)
				return nil
			},
		},
		{
			name: "测试expr表达式",
			check: func() error {
				expr1 := expression.NewExpression().Field("name").Eq("alice")
				expr2 := expression.NewExpression().Field("age").Lt(30)
				user, err := collection.And(coll, expr1, expr2).FindOne(ctx)

				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, "alice", user.Name)
				assert.Equal(t, int64(25), user.Age)
				return nil
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.before != nil {
				if err := tc.before(); err != nil {
					t.Error(err)
				}
			}

			if tc.check != nil {
				if err := tc.check(); err != nil {
					t.Error(err)
				}
			}
		})
	}

	// 清理测试数据
	t.Cleanup(func() {
		_, err := coll.Where("_id").Ne("").Delete(ctx)
		assert.NoError(t, err)
	})
}

// TestAdvancedQueries 高级查询测试
func TestAdvancedQueries(t *testing.T) {
	ctx := context.Background()
	ormClient, _ := tmorm.NewORMClient(nil)
	userColl := collection.NewCollection[TestUser2](ormClient, "mytest", "enhanced_users")
	productColl := collection.NewCollection[TestProduct](ormClient, "mytest", "test_products")

	// 插入增强的测试用户数据
	enhancedUsers := []*TestUser2{
		{
			ID:         primitive.NewObjectID(),
			Name:       "John Doe",
			Email:      "john@company.com",
			Age:        28,
			Status:     "active",
			Department: "Engineering",
			Tags:       []string{"developer", "golang", "senior"},
			Salary:     85000.0,
			CreatedAt:  time.Now().AddDate(0, -6, 0), // 6个月前
			IsActive:   true,
		},
		{
			ID:         primitive.NewObjectID(),
			Name:       "Jane Smith",
			Email:      "jane@company.com",
			Age:        32,
			Status:     "active",
			Department: "IT",
			Tags:       []string{"manager", "python"},
			Salary:     95000.0,
			CreatedAt:  time.Now().AddDate(0, -3, 0), // 3个月前
			IsActive:   true,
		},
		{
			ID:         primitive.NewObjectID(),
			Name:       "Bob Johnson",
			Email:      "bob@external.com",
			Age:        45,
			Status:     "inactive",
			Department: "HR",
			Tags:       []string{"hr", "management"},
			Salary:     75000.0,
			CreatedAt:  time.Now().AddDate(-1, 0, 0), // 1年前
			IsActive:   false,
		},
		{
			ID:         primitive.NewObjectID(),
			Name:       "Alice Brown",
			Email:      "alice@company.com",
			Age:        26,
			Status:     "active",
			Department: "Engineering",
			Tags:       []string{"developer", "javascript", "junior"},
			Salary:     65000.0,
			CreatedAt:  time.Now().AddDate(0, -1, 0), // 1个月前
			IsActive:   true,
		},
	}

	// 插入产品测试数据
	testProducts := []*TestProduct{
		{
			ID:       primitive.NewObjectID(),
			Name:     "Laptop Pro",
			Price:    1299.99,
			Category: "Electronics",
			InStock:  true,
		},
		{
			ID:       primitive.NewObjectID(),
			Name:     "Wireless Mouse",
			Price:    29.99,
			Category: "Accessories",
			InStock:  true,
		},
		{
			ID:       primitive.NewObjectID(),
			Name:     "Gaming Chair",
			Price:    399.99,
			Category: "Furniture",
			InStock:  false,
		},
		{
			ID:       primitive.NewObjectID(),
			Name:     "Mechanical Keyboard",
			Price:    149.99,
			Category: "Electronics",
			InStock:  true,
		},
	}

	// 插入测试数据
	for _, user := range enhancedUsers {
		_, err := userColl.Insert(ctx, user)
		assert.NoError(t, err)
	}

	for _, product := range testProducts {
		_, err := productColl.Insert(ctx, product)
		assert.NoError(t, err)
	}

	// 复杂查询测试用例
	advancedTestCases := []struct {
		name     string
		setup    func() (interface{}, error)
		validate func(interface{}) bool
	}{
		{
			name: "复杂条件查询 - 技术部门活跃用户",
			setup: func() (interface{}, error) {
				// 复杂查询：(Engineering OR IT) AND age 25-45 AND active AND 最近一年注册
				expr1 := expression.Or(
					expression.NewExpression().Field("department").Eq("Engineering"),
					expression.NewExpression().Field("department").Eq("IT"),
				)
				expr2 := expression.NewExpression().Field("age").Between(25, 45)
				expr3 := expression.NewExpression().Field("status").Eq("active")
				combinedExpr := expression.And(expr1, expr2, expr3)
				return userColl.WhereExpr(combinedExpr).Find(ctx)
			},
			validate: func(result interface{}) bool {
				users := result.([]*TestUser2)
				return len(users) == 3 // John, Jane, Alice
			},
		},
		{
			name: "字符串匹配查询 - 公司邮箱用户",
			setup: func() (interface{}, error) {
				return userColl.Expr("email").Contains("@company.com").Find(ctx)
			},
			validate: func(result interface{}) bool {
				users := result.([]*TestUser2)
				return len(users) == 3 // John, Jane, Alice
			},
		},
		{
			name: "数组查询 - 包含特定标签的用户",
			setup: func() (interface{}, error) {
				return userColl.Expr("tags").All("developer").Find(ctx)
			},
			validate: func(result interface{}) bool {
				users := result.([]*TestUser2)
				return len(users) == 2 // John, Alice
			},
		},
		{
			name: "薪资范围查询",
			setup: func() (interface{}, error) {
				return userColl.Expr("salary").Between(70000.0, 90000.0).Find(ctx)
			},
			validate: func(result interface{}) bool {
				users := result.([]*TestUser2)
				return len(users) == 2 // John, Bob
			},
		},
		{
			name: "产品查询 - 价格范围和库存状态",
			setup: func() (interface{}, error) {
				expr1 := expression.NewExpression().Field("price").Between(100.0, 500.0)
				expr2 := expression.NewExpression().Field("in_stock").Eq(true)
				combinedExpr := expression.And(expr1, expr2)
				return productColl.WhereExpr(combinedExpr).Find(ctx)
			},
			validate: func(result interface{}) bool {
				products := result.([]*TestProduct)
				return len(products) == 1 // Mechanical Keyboard
			},
		},
		{
			name: "排序查询 - 按价格降序",
			setup: func() (interface{}, error) {
				opts := options.Find().SetSort(bson.D{{"price", -1}}).SetLimit(2)
				return productColl.Expr("in_stock").Eq(true).Find(ctx, opts)
			},
			validate: func(result interface{}) bool {
				products := result.([]*TestProduct)
				return len(products) == 2 && products[0].Price > products[1].Price
			},
		},
		{
			name: "组合表达式 - Or查询",
			setup: func() (interface{}, error) {
				expr1 := expression.NewExpression().Field("department").Eq("HR")
				expr2 := expression.NewExpression().Field("age").Lt(30)
				return collection.Or(userColl, expr1, expr2).Find(ctx)
			},
			validate: func(result interface{}) bool {
				users := result.([]*TestUser2)
				return len(users) == 2 // Bob (HR), Alice (age < 30)
			},
		},
		{
			name: "NOT查询 - 非活跃用户",
			setup: func() (interface{}, error) {
				return userColl.Where("status").Not().Eq("active").Find(ctx)
			},
			validate: func(result interface{}) bool {
				users := result.([]*TestUser2)
				return len(users) == 1 // Bob
			},
		},
		{
			name: "字段存在性查询",
			setup: func() (interface{}, error) {
				return userColl.Expr("email").Exists(true).Find(ctx)
			},
			validate: func(result interface{}) bool {
				users := result.([]*TestUser2)
				return len(users) == 4 // 所有用户都有email
			},
		},
		{
			name: "正则表达式查询 - 邮箱域名",
			setup: func() (interface{}, error) {
				return userColl.Expr("email").Regex(".*@company\\.com$").Find(ctx)
			},
			validate: func(result interface{}) bool {
				users := result.([]*TestUser2)
				return len(users) == 3 // John, Jane, Alice
			},
		},
	}

	// 执行高级测试用例
	for _, tc := range advancedTestCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := tc.setup()
			assert.NoError(t, err)
			assert.True(t, tc.validate(result))
		})
	}

	// 清理测试数据
	t.Cleanup(func() {
		_, err := userColl.Where("_id").Ne("").Delete(ctx)
		assert.NoError(t, err)
		_, err = productColl.Where("_id").Ne("").Delete(ctx)
		assert.NoError(t, err)
	})
}

type CustomQuery struct {
	data bson.D
}

func (c *CustomQuery) GetBsonD() bson.D {
	return c.data
}
