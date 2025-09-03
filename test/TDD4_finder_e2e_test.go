package test

import (
	"context"
	"testing"
	tmorm "tm_orm"
	"tm_orm/collection"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

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
			name: "复合查询测试 - 使用原生bson进行复杂查询",
			check: func() error {
				// 使用原生bson进行复杂查询
				filter := bson.M{
					"$or": []bson.M{
						{"name": "sean"},
						{"age": bson.M{"$gt": 25}},
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
						"$gte": []interface{}{
							bson.M{"$multiply": []interface{}{"$age", 1}},
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

type CustomQuery struct {
	data bson.D
}

func (c *CustomQuery) GetBsonD() bson.D {
	return c.data
}
