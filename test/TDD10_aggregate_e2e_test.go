package test

import (
	"context"
	"testing"
	tmorm "tm_orm"
	"tm_orm/aggregator"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AggTestUser 聚合测试用户结构
type AggTestUser struct {
	ID     primitive.ObjectID `bson:"_id,omitempty"`
	Name   string             `bson:"name"`
	Age    int                `bson:"age"`
	City   string             `bson:"city,omitempty"`
	Salary int                `bson:"salary,omitempty"`
	Tags   []string           `bson:"tags,omitempty"`
}

// TestOrder 测试订单结构
type TestOrder struct {
	ID     primitive.ObjectID `bson:"_id,omitempty"`
	UserID primitive.ObjectID `bson:"user_id"`
	Amount float64            `bson:"amount"`
	Status string             `bson:"status"`
}

func TestAggregateE2E(t *testing.T) {
	ConnectMongo()
	ctx := context.Background()
	ormClient, _ := tmorm.NewORMClient(nil)

	type tcc struct {
		name   string
		data   any
		before func(cc *tcc)
		after  func(cc *tcc)
		check  func(cc *tcc)
	}

	testCase := []tcc{
		{
			name: "basic_aggregation_match_group",
			data: []*AggTestUser{
				{
					ID:   primitive.NewObjectID(),
					Name: "alice",
					Age:  25,
					City: "beijing",
				},
				{
					ID:   primitive.NewObjectID(),
					Name: "bob",
					Age:  30,
					City: "beijing",
				},
				{
					ID:   primitive.NewObjectID(),
					Name: "charlie",
					Age:  35,
					City: "shanghai",
				},
			},
			before: func(tc *tcc) {
				data := tc.data.([]*AggTestUser)
				for _, user := range data {
					_, err := MongoClient.Database("mytest").Collection("users").InsertOne(ctx, user)
					assert.NoError(t, err)
				}
			},
			after: func(tc *tcc) {
				data := tc.data.([]*AggTestUser)
				for _, user := range data {
					_, err := MongoClient.Database("mytest").Collection("users").DeleteMany(ctx,
						bson.M{"_id": user.ID})
					assert.NoError(t, err)
				}
			},
			check: func(tc *tcc) {
				agg := aggregator.NewAggregator[any](ormClient, "mytest", "users")

				// 匹配年龄大于等于25的用户，按城市分组统计
				agg.Match(bson.M{"age": bson.M{"$gte": 25}}).
					Group("$city", bson.M{
						"user_count": bson.M{"$sum": 1},
						"avg_age":    bson.M{"$avg": "$age"},
						"max_age":    bson.M{"$max": "$age"},
						"min_age":    bson.M{"$min": "$age"},
					}).
					SortDesc("user_count")

				results, err := agg.Execute(ctx)
				assert.NoError(t, err)
				assert.Equal(t, 2, len(results))

				// 验证北京的统计结果（用户数更多）
				firstResult := (*results[0]).(map[string]any)
				assert.Equal(t, "beijing", firstResult["_id"])
				assert.Equal(t, int32(2), firstResult["user_count"])
				assert.Equal(t, float64(27.5), firstResult["avg_age"])
			},
		},
		{
			name: "complex_grouping_with_salary",
			data: []*AggTestUser{
				{
					ID:     primitive.NewObjectID(),
					Name:   "alice",
					Age:    25,
					City:   "beijing",
					Salary: 50000,
				},
				{
					ID:     primitive.NewObjectID(),
					Name:   "bob",
					Age:    30,
					City:   "beijing",
					Salary: 60000,
				},
				{
					ID:     primitive.NewObjectID(),
					Name:   "charlie",
					Age:    35,
					City:   "shanghai",
					Salary: 70000,
				},
			},
			before: func(tc *tcc) {
				data := tc.data.([]*AggTestUser)
				for _, user := range data {
					_, err := MongoClient.Database("mytest").Collection("users").InsertOne(ctx, user)
					assert.NoError(t, err)
				}
			},
			after: func(tc *tcc) {
				data := tc.data.([]*AggTestUser)
				for _, user := range data {
					_, err := MongoClient.Database("mytest").Collection("users").DeleteMany(ctx,
						bson.M{"_id": user.ID})
					assert.NoError(t, err)
				}
			},
			check: func(tc *tcc) {
				agg := aggregator.NewAggregator[any](ormClient, "mytest", "users")

				// 按城市分组，计算薪资统计
				agg.Group("$city", bson.M{
					"user_count": bson.M{"$sum": 1},
					"avg_salary": bson.M{"$avg": "$salary"},
					"max_salary": bson.M{"$max": "$salary"},
					"min_salary": bson.M{"$min": "$salary"},
					"users":      bson.M{"$push": "$name"},
				}).
					SortDesc("avg_salary")

				results, err := agg.Execute(ctx)
				assert.NoError(t, err)
				assert.Equal(t, 2, len(results))

				// 验证上海的统计结果（平均薪资更高）
				firstResult := (*results[0]).(map[string]any)
				assert.Equal(t, "shanghai", firstResult["_id"])
				assert.Equal(t, int32(1), firstResult["user_count"])
				assert.Equal(t, float64(70000), firstResult["avg_salary"])
			},
		},
		{
			name: "projection_and_transformation",
			data: []*AggTestUser{
				{
					ID:   primitive.NewObjectID(),
					Name: "alice",
					Age:  28,
					City: "beijing",
				},
				{
					ID:   primitive.NewObjectID(),
					Name: "bob",
					Age:  32,
					City: "shanghai",
				},
			},
			before: func(tc *tcc) {
				data := tc.data.([]*AggTestUser)
				for _, user := range data {
					_, err := MongoClient.Database("mytest").Collection("users").InsertOne(ctx, user)
					assert.NoError(t, err)
				}
			},
			after: func(tc *tcc) {
				data := tc.data.([]*AggTestUser)
				for _, user := range data {
					_, err := MongoClient.Database("mytest").Collection("users").DeleteMany(ctx,
						bson.M{"_id": user.ID})
					assert.NoError(t, err)
				}
			},
			check: func(tc *tcc) {
				agg := aggregator.NewAggregator[any](ormClient, "mytest", "users")

				// 投影选择特定字段并添加计算字段
				agg.Project(bson.M{
					"name":     1,
					"age":      1,
					"city":     1,
					"is_adult": bson.M{"$gte": []any{"$age", 18}},
					"age_group": bson.M{
						"$cond": []any{
							bson.M{"$lt": []any{"$age", 30}},
							"young",
							"mature",
						},
					},
					"_id": 0,
				})

				results, err := agg.Execute(ctx)
				assert.NoError(t, err)
				assert.Equal(t, 2, len(results))

				// 验证投影结果
				for _, result := range results {
					resultMap := (*result).(map[string]any)
					assert.Contains(t, resultMap, "name")
					assert.Contains(t, resultMap, "age")
					assert.Contains(t, resultMap, "city")
					assert.Contains(t, resultMap, "is_adult")
					assert.Contains(t, resultMap, "age_group")
					assert.NotContains(t, resultMap, "_id")
					assert.Equal(t, true, resultMap["is_adult"])
				}
			},
		},
		{
			name: "lookup_join_users_orders",
			data: map[string]any{
				"users": []*AggTestUser{
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
				},
				"orders": []*TestOrder{},
			},
			before: func(tc *tcc) {
				data := tc.data.(map[string]any)
				users := data["users"].([]*AggTestUser)
				orders := []*TestOrder{
					{
						ID:     primitive.NewObjectID(),
						UserID: users[0].ID,
						Amount: 100.0,
						Status: "completed",
					},
					{
						ID:     primitive.NewObjectID(),
						UserID: users[0].ID,
						Amount: 200.0,
						Status: "completed",
					},
					{
						ID:     primitive.NewObjectID(),
						UserID: users[1].ID,
						Amount: 150.0,
						Status: "pending",
					},
				}
				data["orders"] = orders

				// 插入用户数据
				for _, user := range users {
					_, err := MongoClient.Database("mytest").Collection("users").InsertOne(ctx, user)
					assert.NoError(t, err)
				}

				// 插入订单数据
				for _, order := range orders {
					_, err := MongoClient.Database("mytest").Collection("orders").InsertOne(ctx, order)
					assert.NoError(t, err)
				}
			},
			after: func(tc *tcc) {
				data := tc.data.(map[string]any)
				users := data["users"].([]*TestUser)
				orders := data["orders"].([]*TestOrder)

				// 清理用户数据
				for _, user := range users {
					_, err := MongoClient.Database("mytest").Collection("users").DeleteMany(ctx,
						bson.M{"_id": user.ID})
					assert.NoError(t, err)
				}

				// 清理订单数据
				for _, order := range orders {
					_, err := MongoClient.Database("mytest").Collection("orders").DeleteMany(ctx,
						bson.M{"_id": order.ID})
					assert.NoError(t, err)
				}
			},
			check: func(tc *tcc) {
				agg := aggregator.NewAggregator[any](ormClient, "mytest", "users")

				// 关联查询用户和订单，计算订单统计
				agg.Lookup("orders", "_id", "user_id", "user_orders").
					AddStage(bson.M{
						"$addFields": bson.M{
							"order_count": bson.M{"$size": "$user_orders"},
							"total_amount": bson.M{
								"$sum": "$user_orders.amount",
							},
						},
					}).
					Match(bson.M{"order_count": bson.M{"$gt": 0}}).
					SortDesc("total_amount")

				results, err := agg.Execute(ctx)
				assert.NoError(t, err)
				assert.Equal(t, 2, len(results))

				// 验证第一个结果（alice，总金额更高）
				firstResult := (*results[0]).(map[string]any)
				assert.Equal(t, "alice", firstResult["name"])
				assert.Equal(t, int32(2), firstResult["order_count"])
				assert.Equal(t, float64(300), firstResult["total_amount"])
			},
		},
		{
			name: "unwind_array_tags",
			data: []*AggTestUser{
				{
					ID:   primitive.NewObjectID(),
					Name: "alice",
					Age:  25,
					Tags: []string{"developer", "golang", "mongodb"},
				},
				{
					ID:   primitive.NewObjectID(),
					Name: "bob",
					Age:  30,
					Tags: []string{"designer", "ui", "ux"},
				},
				{
					ID:   primitive.NewObjectID(),
					Name: "charlie",
					Age:  35,
					Tags: []string{"developer", "python"},
				},
			},
			before: func(tc *tcc) {
				data := tc.data.([]*AggTestUser)
				for _, user := range data {
					_, err := MongoClient.Database("mytest").Collection("users").InsertOne(ctx, user)
					assert.NoError(t, err)
				}
			},
			after: func(tc *tcc) {
				data := tc.data.([]*AggTestUser)
				for _, user := range data {
					_, err := MongoClient.Database("mytest").Collection("users").DeleteMany(ctx,
						bson.M{"_id": user.ID})
					assert.NoError(t, err)
				}
			},
			check: func(tc *tcc) {
				agg := aggregator.NewAggregator[any](ormClient, "mytest", "users")

				// 展开标签数组，按标签分组统计
				agg.Match(bson.M{"tags": bson.M{"$exists": true, "$ne": nil}}).
					Unwind("$tags").
					Group("$tags", bson.M{
						"user_count": bson.M{"$sum": 1},
						"users":      bson.M{"$push": "$name"},
					}).
					SortDesc("user_count")

				results, err := agg.Execute(ctx)
				assert.NoError(t, err)
				assert.True(t, len(results) > 0)

				// 验证developer标签的统计结果
				found := false
				for _, result := range results {
					resultMap := (*result).(map[string]any)
					if resultMap["_id"] == "developer" {
						found = true
						assert.Equal(t, int32(2), resultMap["user_count"])
						users := resultMap["users"].(primitive.A)
						assert.Equal(t, 2, len(users))
						break
					}
				}
				assert.True(t, found, "developer tag should be found")
			},
		},
		{
			name: "conditional_aggregation_with_switch",
			data: []*AggTestUser{
				{
					ID:     primitive.NewObjectID(),
					Name:   "alice",
					Age:    22,
					Salary: 45000,
				},
				{
					ID:     primitive.NewObjectID(),
					Name:   "bob",
					Age:    28,
					Salary: 55000,
				},
				{
					ID:     primitive.NewObjectID(),
					Name:   "charlie",
					Age:    35,
					Salary: 65000,
				},
				{
					ID:     primitive.NewObjectID(),
					Name:   "david",
					Age:    45,
					Salary: 75000,
				},
			},
			before: func(tc *tcc) {
				data := tc.data.([]*TestUser)
				for _, user := range data {
					_, err := MongoClient.Database("mytest").Collection("users").InsertOne(ctx, user)
					assert.NoError(t, err)
				}
			},
			after: func(tc *tcc) {
				data := tc.data.([]*TestUser)
				for _, user := range data {
					_, err := MongoClient.Database("mytest").Collection("users").DeleteMany(ctx,
						bson.M{"_id": user.ID})
					assert.NoError(t, err)
				}
			},
			check: func(tc *tcc) {
				agg := aggregator.NewAggregator[any](ormClient, "mytest", "users")

				// 使用switch条件添加年龄段字段，然后按年龄段分组
				agg.AddStage(bson.M{
					"$addFields": bson.M{
						"age_group": bson.M{
							"$switch": bson.M{
								"branches": []bson.M{
									{"case": bson.M{"$lt": []any{"$age", 25}}, "then": "18-24"},
									{"case": bson.M{"$lt": []any{"$age", 35}}, "then": "25-34"},
									{"case": bson.M{"$lt": []any{"$age", 45}}, "then": "35-44"},
								},
								"default": "45+",
							},
						},
					},
				}).
					Group("$age_group", bson.M{
						"total_users": bson.M{"$sum": 1},
						"high_salary_users": bson.M{
							"$sum": bson.M{
								"$cond": []any{
									bson.M{"$gte": []any{"$salary", 50000}},
									1,
									0,
								},
							},
						},
						"avg_salary": bson.M{"$avg": "$salary"},
					}).
					Sort(bson.D{{"_id", 1}})

				results, err := agg.Execute(ctx)
				assert.NoError(t, err)
				assert.Equal(t, 4, len(results))

				// 验证各年龄段的统计结果
				for _, result := range results {
					resultMap := (*result).(map[string]any)
					assert.Contains(t, resultMap, "_id")
					assert.Contains(t, resultMap, "total_users")
					assert.Contains(t, resultMap, "high_salary_users")
					assert.Contains(t, resultMap, "avg_salary")
				}
			},
		},
		{
			name: "time_based_aggregation_simulation",
			data: []*TestOrder{
				{
					ID:     primitive.NewObjectID(),
					UserID: primitive.NewObjectID(),
					Amount: 100.0,
					Status: "completed",
				},
				{
					ID:     primitive.NewObjectID(),
					UserID: primitive.NewObjectID(),
					Amount: 200.0,
					Status: "completed",
				},
				{
					ID:     primitive.NewObjectID(),
					UserID: primitive.NewObjectID(),
					Amount: 150.0,
					Status: "pending",
				},
			},
			before: func(tc *tcc) {
				data := tc.data.([]*TestOrder)
				for _, order := range data {
					_, err := MongoClient.Database("mytest").Collection("orders").InsertOne(ctx, order)
					assert.NoError(t, err)
				}
			},
			after: func(tc *tcc) {
				data := tc.data.([]*TestOrder)
				for _, order := range data {
					_, err := MongoClient.Database("mytest").Collection("orders").DeleteMany(ctx,
						bson.M{"_id": order.ID})
					assert.NoError(t, err)
				}
			},
			check: func(tc *tcc) {
				agg := aggregator.NewAggregator[any](ormClient, "mytest", "orders")

				// 按状态分组统计订单
				agg.Match(bson.M{"status": bson.M{"$in": []string{"completed", "pending"}}}).
					Group("$status", bson.M{
						"order_count":  bson.M{"$sum": 1},
						"total_amount": bson.M{"$sum": "$amount"},
						"avg_amount":   bson.M{"$avg": "$amount"},
						"max_amount":   bson.M{"$max": "$amount"},
						"min_amount":   bson.M{"$min": "$amount"},
					}).
					SortDesc("total_amount")

				results, err := agg.Execute(ctx)
				assert.NoError(t, err)
				assert.Equal(t, 2, len(results))

				// 验证completed状态的统计结果（总金额更高）
				firstResult := (*results[0]).(map[string]any)
				assert.Equal(t, "completed", firstResult["_id"])
				assert.Equal(t, int32(2), firstResult["order_count"])
				assert.Equal(t, float64(300), firstResult["total_amount"])
				assert.Equal(t, float64(150), firstResult["avg_amount"])
			},
		},
		{
			name: "complex_pipeline_with_multiple_stages",
			data: []*AggTestUser{
				{
					ID:     primitive.NewObjectID(),
					Name:   "alice",
					Age:    25,
					City:   "beijing",
					Salary: 50000,
					Tags:   []string{"developer", "golang"},
				},
				{
					ID:     primitive.NewObjectID(),
					Name:   "bob",
					Age:    30,
					City:   "beijing",
					Salary: 60000,
					Tags:   []string{"designer", "ui"},
				},
				{
					ID:     primitive.NewObjectID(),
					Name:   "charlie",
					Age:    35,
					City:   "shanghai",
					Salary: 70000,
					Tags:   []string{"developer", "python"},
				},
			},
			before: func(tc *tcc) {
				data := tc.data.([]*TestUser)
				for _, user := range data {
					_, err := MongoClient.Database("mytest").Collection("users").InsertOne(ctx, user)
					assert.NoError(t, err)
				}
			},
			after: func(tc *tcc) {
				data := tc.data.([]*TestUser)
				for _, user := range data {
					_, err := MongoClient.Database("mytest").Collection("users").DeleteMany(ctx,
						bson.M{"_id": user.ID})
					assert.NoError(t, err)
				}
			},
			check: func(tc *tcc) {
				agg := aggregator.NewAggregator[any](ormClient, "mytest", "users")

				// 复杂管道：匹配、展开标签、按标签分组、添加计算字段、排序、限制
				agg.Match(bson.M{"salary": bson.M{"$gte": 50000}}).
					Unwind("$tags").
					Group("$tags", bson.M{
						"user_count":   bson.M{"$sum": 1},
						"avg_salary":   bson.M{"$avg": "$salary"},
						"total_salary": bson.M{"$sum": "$salary"},
						"users":        bson.M{"$push": "$name"},
					}).
					AddStage(bson.M{
						"$addFields": bson.M{
							"salary_level": bson.M{
								"$cond": []any{
									bson.M{"$gte": []any{"$avg_salary", 60000}},
									"high",
									"medium",
								},
							},
						},
					}).
					SortDesc("user_count").
					Limit(10)

				results, err := agg.Execute(ctx)
				assert.NoError(t, err)
				assert.True(t, len(results) > 0)

				// 验证结果包含所需字段
				for _, result := range results {
					resultMap := (*result).(map[string]any)
					assert.Contains(t, resultMap, "_id")
					assert.Contains(t, resultMap, "user_count")
					assert.Contains(t, resultMap, "avg_salary")
					assert.Contains(t, resultMap, "total_salary")
					assert.Contains(t, resultMap, "users")
					assert.Contains(t, resultMap, "salary_level")
				}
			},
		},
	}

	for _, tc := range testCase {
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
