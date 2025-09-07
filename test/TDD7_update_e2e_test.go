package test

import (
	"context"
	"testing"
	tmorm "tm_orm"
	"tm_orm/collection"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// 端到端测试updater
func TestUpdaterE2E(t *testing.T) {
	ConnectMongo()
	ctx := context.Background()
	ormClient, _ := tmorm.NewORMClient(nil)
	coll := collection.NewCollection[TestUser](ormClient, "mytest", "db_test")

	type tcase struct {
		name   string
		data   any
		before func(*tcase)
		after  func(*tcase)
		check  func(*tcase)
	}

	testCases := []tcase{
		{
			name: "update_one",
			data: &TestUser{
				ID:   primitive.ObjectID([12]byte{1, 2, 3, 4, 5}),
				Name: "sean",
				Age:  20,
			},
			before: func(tc *tcase) {
				data := tc.data.(*TestUser)
				_, err := coll.Insert(ctx, data)
				assert.NoError(t, err)
			},
			after: func(tc *tcase) {
				data := tc.data.(*TestUser)
				_, err := coll.Where("_id").Eq(data.ID).Delete(ctx)
				assert.NoError(t, err)
			},
			check: func(tc *tcase) {
				data := tc.data.(*TestUser)

				_, err := coll.Where("_id").Eq(data.ID).Set("age", 25).UpdateOne(ctx)
				assert.NoError(t, err)

				// 验证更新结果
				result, err := coll.Where("_id").Eq(data.ID).FindOne(ctx)
				assert.NoError(t, err)
				assert.Equal(t, int64(25), result.Age)
			},
		},
		{
			name: "update_many",
			data: []*TestUser{
				{
					ID:   primitive.ObjectID([12]byte{1, 2, 3, 4, 5}),
					Name: "sean",
					Age:  20,
				},
				{
					ID:   primitive.ObjectID([12]byte{1, 2, 3, 4, 6}),
					Name: "sean",
					Age:  22,
				},
			},
			before: func(tc *tcase) {
				data := tc.data.([]*TestUser)
				for _, user := range data {
					_, err := coll.Insert(ctx, user)
					assert.NoError(t, err)
				}
			},
			after: func(tc *tcase) {
				data := tc.data.([]*TestUser)
				for _, user := range data {
					_, err := coll.Where("_id").Eq(user.ID).Delete(ctx)
					assert.NoError(t, err)
				}
			},
			check: func(tc *tcase) {
				_, err := coll.Where("name").Eq("sean").Set("age", 30).Update(ctx)
				assert.NoError(t, err)

				// 验证更新结果
				results, err := coll.Where("name").Eq("sean").Find(ctx)
				assert.NoError(t, err)
				assert.Equal(t, 2, len(results))
				for _, result := range results {
					assert.Equal(t, int64(30), result.Age)
				}
			},
		},
		{
			name: "inc_operation",
			data: &TestUser{
				ID:   primitive.ObjectID([12]byte{1, 2, 3, 4, 7}),
				Name: "alice",
				Age:  25,
			},
			before: func(tc *tcase) {
				data := tc.data.(*TestUser)
				_, err := coll.Insert(ctx, data)
				assert.NoError(t, err)
			},
			after: func(tc *tcase) {
				data := tc.data.(*TestUser)
				_, err := coll.Where("_id").Eq(data.ID).Delete(ctx)
				assert.NoError(t, err)
			},
			check: func(tc *tcase) {
				data := tc.data.(*TestUser)

				_, err := coll.Where("_id").Eq(data.ID).Set("name", data.Name).Inc("age", 5).UpdateOne(ctx)
				assert.NoError(t, err)

				// 验证Inc操作结果
				result, err := coll.Where("_id").Eq(data.ID).FindOne(ctx)
				assert.NoError(t, err)
				assert.Equal(t, int64(30), result.Age)
			},
		},
		{
			name: "mul_operation",
			data: &TestUser{
				ID:   primitive.ObjectID([12]byte{1, 2, 3, 4, 8}),
				Name: "bob",
				Age:  10,
			},
			before: func(tc *tcase) {
				data := tc.data.(*TestUser)
				_, err := coll.Insert(ctx, data)
				assert.NoError(t, err)
			},
			after: func(tc *tcase) {
				data := tc.data.(*TestUser)
				_, err := coll.Where("_id").Eq(data.ID).Delete(ctx)
				assert.NoError(t, err)
			},
			check: func(tc *tcase) {
				data := tc.data.(*TestUser)

				_, err := coll.Where("_id").Eq(data.ID).Set("name", data.Name).Mul("age", 3).UpdateOne(ctx)
				assert.NoError(t, err)

				// 验证Mul操作结果
				result, err := coll.Where("_id").Eq(data.ID).FindOne(ctx)
				assert.NoError(t, err)
				assert.Equal(t, int64(30), result.Age)
			},
		},
		{
			name: "unset_operation",
			data: &TestUser{
				ID:   primitive.ObjectID([12]byte{1, 2, 3, 4, 9}),
				Name: "charlie",
				Age:  35,
			},
			before: func(tc *tcase) {
				data := tc.data.(*TestUser)
				_, err := coll.Insert(ctx, data)
				assert.NoError(t, err)
			},
			after: func(tc *tcase) {
				data := tc.data.(*TestUser)
				_, err := coll.Where("_id").Eq(data.ID).Delete(ctx)
				assert.NoError(t, err)
			},
			check: func(tc *tcase) {
				data := tc.data.(*TestUser)

				_, err := coll.Where("_id").Eq(data.ID).Set("name", data.Name).Unset("age").UpdateOne(ctx)
				assert.NoError(t, err)

				// 验证Unset操作结果
				result, err := coll.Where("_id").Eq(data.ID).FindOne(ctx)
				assert.NoError(t, err)
				assert.Equal(t, int64(0), result.Age) // 字段被删除后应该是零值
			},
		},
		{
			name: "complex_update_operations",
			data: &TestUser{
				ID:   primitive.ObjectID([12]byte{1, 2, 3, 5, 2}),
				Name: "frank",
				Age:  45,
			},
			before: func(tc *tcase) {
				data := tc.data.(*TestUser)
				_, err := coll.Insert(ctx, data)
				assert.NoError(t, err)
			},
			after: func(tc *tcase) {
				data := tc.data.(*TestUser)
				_, err := coll.Where("_id").Eq(data.ID).Delete(ctx)
				assert.NoError(t, err)
			},
			check: func(tc *tcase) {
				data := tc.data.(*TestUser)

				// 复合更新操作：Set + Inc + CurrentDate
				_, err := coll.Where("_id").Eq(data.ID).
					Set("name", "frank_updated").
					Inc("age", 5).
					CurrentDate("lastModified").
					UpdateOne(ctx)
				assert.NoError(t, err)

				// 验证复合操作结果
				result, err := coll.Where("_id").Eq(data.ID).FindOne(ctx)
				assert.NoError(t, err)
				assert.Equal(t, "frank_updated", result.Name)
				assert.Equal(t, int64(50), result.Age)
			},
		},
		{
			name: "array_operations",
			data: &TestUser{
				ID:   primitive.ObjectID([12]byte{1, 2, 3, 5, 3}),
				Name: "david",
				Age:  28,
			},
			before: func(tc *tcase) {
				data := tc.data.(*TestUser)
				_, err := coll.Insert(ctx, data)
				assert.NoError(t, err)
			},
			after: func(tc *tcase) {
				data := tc.data.(*TestUser)
				_, err := coll.Where("_id").Eq(data.ID).Delete(ctx)
				assert.NoError(t, err)
			},
			check: func(tc *tcase) {
					data := tc.data.(*TestUser)

					// 测试数组操作：AddToSet, Push
					_, err := coll.Where("_id").Eq(data.ID).
						Set("name", data.Name).
						AddToSet("tags", "golang").
						Push("scores", 95).
						UpdateOne(ctx)
					assert.NoError(t, err)

					// 验证数组操作结果
					result, err := coll.Where("_id").Eq(data.ID).FindOne(ctx)
					assert.NoError(t, err)
					assert.Equal(t, data.Name, result.Name)
				},
		},
		{
			name: "rename_operation",
			data: &TestUser{
				ID:   primitive.ObjectID([12]byte{1, 2, 3, 5, 4}),
				Name: "eve",
				Age:  32,
			},
			before: func(tc *tcase) {
				data := tc.data.(*TestUser)
				_, err := coll.Insert(ctx, data)
				assert.NoError(t, err)
			},
			after: func(tc *tcase) {
				data := tc.data.(*TestUser)
				_, err := coll.Where("_id").Eq(data.ID).Delete(ctx)
				assert.NoError(t, err)
			},
			check: func(tc *tcase) {
					data := tc.data.(*TestUser)

					// 测试字段重命名操作
					_, err := coll.Where("_id").Eq(data.ID).
						Set("fullName", data.Name).
						Rename("name", "oldName").
						UpdateOne(ctx)
					assert.NoError(t, err)

					// 验证重命名操作结果
					result, err := coll.Where("_id").Eq(data.ID).FindOne(ctx)
					assert.NoError(t, err)
					// 原字段应该被重命名，新字段应该存在
					assert.Equal(t, int64(32), result.Age)
				},
		},
		{
			name: "conditional_update_with_upsert",
			data: &TestUser{
				ID:   primitive.ObjectID([12]byte{1, 2, 3, 5, 5}),
				Name: "grace",
				Age:  27,
			},
			before: func(tc *tcase) {
				// 不插入数据，测试 upsert 功能
			},
			after: func(tc *tcase) {
				data := tc.data.(*TestUser)
				_, err := coll.Where("_id").Eq(data.ID).Delete(ctx)
				assert.NoError(t, err)
			},
			check: func(tc *tcase) {
					data := tc.data.(*TestUser)

					// 测试条件更新：如果文档不存在则插入
					_, err := coll.Where("_id").Eq(data.ID).
						Set("name", data.Name).
						Set("age", data.Age).
						SetOnInsert("createdBy", "system").
						CurrentDate("createdAt").
						UpdateOne(ctx)
					assert.NoError(t, err)

					// 验证文档被创建
					result, err := coll.Where("_id").Eq(data.ID).FindOne(ctx)
					assert.NoError(t, err)
					assert.Equal(t, data.Name, result.Name)
					assert.Equal(t, data.Age, result.Age)
				},
		},
		{
			name: "batch_update_with_different_conditions",
			data: []*TestUser{
				{
					ID:   primitive.ObjectID([12]byte{1, 2, 3, 5, 6}),
					Name: "henry",
					Age:  40,
				},
				{
					ID:   primitive.ObjectID([12]byte{1, 2, 3, 5, 7}),
					Name: "ivy",
					Age:  35,
				},
				{
					ID:   primitive.ObjectID([12]byte{1, 2, 3, 5, 8}),
					Name: "jack",
					Age:  25,
				},
			},
			before: func(tc *tcase) {
				data := tc.data.([]*TestUser)
				for _, user := range data {
					_, err := coll.Insert(ctx, user)
					assert.NoError(t, err)
				}
			},
			after: func(tc *tcase) {
				data := tc.data.([]*TestUser)
				for _, user := range data {
					_, err := coll.Where("_id").Eq(user.ID).Delete(ctx)
					assert.NoError(t, err)
				}
			},
			check: func(tc *tcase) {
					// 测试批量更新：年龄大于30的用户增加5岁
					_, err := coll.Where("age").Gt(30).Inc("age", 5).Update(ctx)
					assert.NoError(t, err)

					// 验证批量更新结果
					results, err := coll.Where("age").Gt(35).Find(ctx)
					assert.NoError(t, err)
					assert.Equal(t, 2, len(results)) // henry(45) 和 ivy(40)

					// 验证年龄小于等于30的用户未被更新
					jackResult, err := coll.Where("name").Eq("jack").FindOne(ctx)
					assert.NoError(t, err)
					assert.Equal(t, int64(25), jackResult.Age) // 未被更新
				},
		},
		{
			name: "complex_nested_update_operations",
			data: &TestUser{
				ID:   primitive.ObjectID([12]byte{1, 2, 3, 5, 9}),
				Name: "karen",
				Age:  29,
			},
			before: func(tc *tcase) {
				data := tc.data.(*TestUser)
				_, err := coll.Insert(ctx, data)
				assert.NoError(t, err)
			},
			after: func(tc *tcase) {
				data := tc.data.(*TestUser)
				_, err := coll.Where("_id").Eq(data.ID).Delete(ctx)
				assert.NoError(t, err)
			},
			check: func(tc *tcase) {
					data := tc.data.(*TestUser)

					// 测试复杂嵌套更新操作：多种操作组合
					_, err := coll.Where("_id").Eq(data.ID).
						Set("name", "karen_updated").
						Inc("age", 1).
						Mul("score", 2).
						AddToSet("skills", "golang").
						Push("achievements", "expert").
						CurrentDate("lastUpdated").
						SetOnInsert("createdBy", "admin").
						UpdateOne(ctx)
					assert.NoError(t, err)

					// 验证复杂更新操作结果
					result, err := coll.Where("_id").Eq(data.ID).FindOne(ctx)
					assert.NoError(t, err)
					assert.Equal(t, "karen_updated", result.Name)
					assert.Equal(t, int64(30), result.Age)
				},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.before != nil {
				tc.before(&tc)
			}

			if tc.check != nil {
				tc.check(&tc)
			}

			if tc.after != nil {
				tc.after(&tc)
			}
		})
	}
}
