package test

import (
	"testing"
	tmorm "tm_orm"
	"tm_orm/aggregator"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

// TestAggregatorMatch 测试 Match 阶段的 BSON 构建
func TestAggregatorMatch(t *testing.T) {
	// 创建一个模拟的 ORMClient
	client := &tmorm.ORMClient{}

	tests := []struct {
		name     string
		buildFn  func() *aggregator.Aggregator[any]
		expected []bson.M
	}{
		{
			name: "simple_match",
			buildFn: func() *aggregator.Aggregator[any] {
				return aggregator.NewAggregator[any](client, "testdb", "users").Match(bson.M{"age": bson.M{tmorm.GtOp: 18}})
			},
			expected: []bson.M{
				{tmorm.MatchOp: bson.M{"age": bson.M{tmorm.GtOp: 18}}},
			},
		},
		{
			name: "complex_match_with_and",
			buildFn: func() *aggregator.Aggregator[any] {
				return aggregator.NewAggregator[any](client, "testdb", "users").Match(bson.M{
					tmorm.AndOp: []bson.M{
						{"age": bson.M{tmorm.GteOp: 18}},
						{"status": "active"},
					},
				})
			},
			expected: []bson.M{
				{tmorm.MatchOp: bson.M{
					tmorm.AndOp: []bson.M{
						{"age": bson.M{tmorm.GteOp: 18}},
						{"status": "active"},
					},
				}},
			},
		},
		{
			name: "match_with_exists_and_ne",
			buildFn: func() *aggregator.Aggregator[any] {
				return aggregator.NewAggregator[any](client, "testdb", "users").Match(bson.M{
					"tags": bson.M{"$exists": true, "$ne": nil},
				})
			},
			expected: []bson.M{
				{tmorm.MatchOp: bson.M{
					"tags": bson.M{"$exists": true, "$ne": nil},
				}},
			},
		},
		{
			name: "match_with_gte_condition",
			buildFn: func() *aggregator.Aggregator[any] {
				return aggregator.NewAggregator[any](client, "testdb", "users").Match(bson.M{"age": bson.M{"$gte": 25}})
			},
			expected: []bson.M{
				{tmorm.MatchOp: bson.M{"age": bson.M{"$gte": 25}}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agg := tt.buildFn()
			pipeline := agg.Pipeline()
			assert.Equal(t, tt.expected, pipeline)
		})
	}
}

// TestAggregatorProject 测试 Project 阶段的 BSON 构建
func TestAggregatorProject(t *testing.T) {
	client := &tmorm.ORMClient{}

	tests := []struct {
		name     string
		buildFn  func() *aggregator.Aggregator[any]
		expected []bson.M
	}{
		{
			name: "simple_project",
			buildFn: func() *aggregator.Aggregator[any] {
				return aggregator.NewAggregator[any](client, "testdb", "users").Project(bson.M{
					"name": 1,
					"age":  1,
					"_id":  0,
				})
			},
			expected: []bson.M{
				{tmorm.ProjectOp: bson.M{
					"name": 1,
					"age":  1,
					"_id":  0,
				}},
			},
		},
		{
			name: "computed_fields_project",
			buildFn: func() *aggregator.Aggregator[any] {
				return aggregator.NewAggregator[any](client, "testdb", "users").Project(bson.M{
					"name":     1,
					"age":      1,
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
			},
			expected: []bson.M{
				{tmorm.ProjectOp: bson.M{
					"name":     1,
					"age":      1,
					"is_adult": bson.M{"$gte": []any{"$age", 18}},
					"age_group": bson.M{
						"$cond": []any{
							bson.M{"$lt": []any{"$age", 30}},
							"young",
							"mature",
						},
					},
					"_id": 0,
				}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agg := tt.buildFn()
			pipeline := agg.Pipeline()
			assert.Equal(t, tt.expected, pipeline)
		})
	}
}

// TestAggregatorSort 测试 Sort 阶段的 BSON 构建
func TestAggregatorSort(t *testing.T) {
	client := &tmorm.ORMClient{}

	tests := []struct {
		name     string
		buildFn  func() *aggregator.Aggregator[any]
		expected []bson.M
	}{
		{
			name: "single_field_asc",
			buildFn: func() *aggregator.Aggregator[any] {
				return aggregator.NewAggregator[any](client, "testdb", "users").SortAsc("name")
			},
			expected: []bson.M{
				{tmorm.SortOp: bson.D{{"name", 1}}},
			},
		},
		{
			name: "single_field_desc",
			buildFn: func() *aggregator.Aggregator[any] {
				return aggregator.NewAggregator[any](client, "testdb", "users").SortDesc("user_count")
			},
			expected: []bson.M{
				{tmorm.SortOp: bson.D{{"user_count", -1}}},
			},
		},
		{
			name: "multiple_fields_mixed",
			buildFn: func() *aggregator.Aggregator[any] {
				return aggregator.NewAggregator[any](client, "testdb", "users").SortAsc("name").SortDesc("age")
			},
			expected: []bson.M{
				{tmorm.SortOp: bson.D{{"name", 1}}},
				{tmorm.SortOp: bson.D{{"age", -1}}},
			},
		},
		{
			name: "sort_with_bson_d",
			buildFn: func() *aggregator.Aggregator[any] {
				return aggregator.NewAggregator[any](client, "testdb", "users").Sort(bson.D{{"age", 1}})
			},
			expected: []bson.M{
				{tmorm.SortOp: bson.D{{"age", 1}}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agg := tt.buildFn()
			pipeline := agg.Pipeline()
			assert.Equal(t, tt.expected, pipeline)
		})
	}
}

// TestAggregatorGroup 测试 Group 阶段的 BSON 构建
func TestAggregatorGroup(t *testing.T) {
	client := &tmorm.ORMClient{}

	tests := []struct {
		name     string
		buildFn  func() *aggregator.Aggregator[any]
		expected []bson.M
	}{
		{
			name: "simple_group",
			buildFn: func() *aggregator.Aggregator[any] {
				return aggregator.NewAggregator[any](client, "testdb", "orders").Group("$customerId", bson.M{
					"totalAmount": bson.M{tmorm.SumOp: "$amount"},
					"count":       bson.M{tmorm.SumOp: 1},
				})
			},
			expected: []bson.M{
				{tmorm.GroupOp: bson.M{
					tmorm.IdOp:    "$customerId",
					"totalAmount": bson.M{tmorm.SumOp: "$amount"},
					"count":       bson.M{tmorm.SumOp: 1},
				}},
			},
		},
		{
			name: "group_by_city_with_aggregations",
			buildFn: func() *aggregator.Aggregator[any] {
				return aggregator.NewAggregator[any](client, "testdb", "users").Group("$city", bson.M{
					"user_count": bson.M{tmorm.SumOp: 1},
					"avg_salary": bson.M{tmorm.AvgOp: "$salary"},
					"max_salary": bson.M{tmorm.MaxOp: "$salary"},
					"min_salary": bson.M{tmorm.MinOp: "$salary"},
				})
			},
			expected: []bson.M{
				{tmorm.GroupOp: bson.M{
					tmorm.IdOp:   "$city",
					"user_count": bson.M{tmorm.SumOp: 1},
					"avg_salary": bson.M{tmorm.AvgOp: "$salary"},
					"max_salary": bson.M{tmorm.MaxOp: "$salary"},
					"min_salary": bson.M{tmorm.MinOp: "$salary"},
				}},
			},
		},
		{
			name: "group_by_name_with_push",
			buildFn: func() *aggregator.Aggregator[any] {
				return aggregator.NewAggregator[any](client, "testdb", "users").Group("$name", bson.M{
					"avgAge": bson.M{"$avg": "$age"},
					"count":  bson.M{"$sum": 1},
				})
			},
			expected: []bson.M{
				{tmorm.GroupOp: bson.M{
					tmorm.IdOp: "$name",
					"avgAge":   bson.M{"$avg": "$age"},
					"count":    bson.M{"$sum": 1},
				}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agg := tt.buildFn()
			pipeline := agg.Pipeline()
			assert.Equal(t, tt.expected, pipeline)
		})
	}
}

// TestAggregatorLookup 测试 Lookup 阶段的 BSON 构建
func TestAggregatorLookup(t *testing.T) {
	client := &tmorm.ORMClient{}

	tests := []struct {
		name     string
		buildFn  func() *aggregator.Aggregator[any]
		expected []bson.M
	}{
		{
			name: "simple_lookup",
			buildFn: func() *aggregator.Aggregator[any] {
				return aggregator.NewAggregator[any](client, "testdb", "orders").Lookup("users", "userId", "_id", "user")
			},
			expected: []bson.M{
				{tmorm.StageLookUpOp: bson.M{
					"from":         "users",
					"localField":   "userId",
					"foreignField": "_id",
					"as":           "user",
				}},
			},
		},
		{
			name: "lookup_orders_to_users",
			buildFn: func() *aggregator.Aggregator[any] {
				return aggregator.NewAggregator[any](client, "testdb", "users").Lookup("orders", "_id", "user_id", "user_orders")
			},
			expected: []bson.M{
				{tmorm.StageLookUpOp: bson.M{
					"from":         "orders",
					"localField":   "_id",
					"foreignField": "user_id",
					"as":           "user_orders",
				}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agg := tt.buildFn()
			pipeline := agg.Pipeline()
			assert.Equal(t, tt.expected, pipeline)
		})
	}
}

// TestAggregatorUnwind 测试 Unwind 阶段的 BSON 构建
func TestAggregatorUnwind(t *testing.T) {
	client := &tmorm.ORMClient{}

	tests := []struct {
		name     string
		buildFn  func() *aggregator.Aggregator[any]
		expected []bson.M
	}{
		{
			name: "simple_unwind",
			buildFn: func() *aggregator.Aggregator[any] {
				return aggregator.NewAggregator[any](client, "testdb", "users").Unwind("$tags")
			},
			expected: []bson.M{
				{tmorm.UnwindOp: "$tags"},
			},
		},
		{
			name: "unwind_with_options",
			buildFn: func() *aggregator.Aggregator[any] {
				return aggregator.NewAggregator[any](client, "testdb", "users").AddStage(bson.M{
					tmorm.UnwindOp: bson.M{
						"path":                    "$tags",
						"includeArrayIndex":       "tag_index",
						"preserveNullAndEmptyArrays": true,
					},
				})
			},
			expected: []bson.M{
				{tmorm.UnwindOp: bson.M{
					"path":                    "$tags",
					"includeArrayIndex":       "tag_index",
					"preserveNullAndEmptyArrays": true,
				}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agg := tt.buildFn()
			pipeline := agg.Pipeline()
			assert.Equal(t, tt.expected, pipeline)
		})
	}
}

// TestAggregatorAddFields 测试 AddFields 阶段的 BSON 构建
func TestAggregatorAddFields(t *testing.T) {
	client := &tmorm.ORMClient{}

	tests := []struct {
		name     string
		buildFn  func() *aggregator.Aggregator[any]
		expected []bson.M
	}{
		{
			name: "add_fields_with_switch",
			buildFn: func() *aggregator.Aggregator[any] {
				return aggregator.NewAggregator[any](client, "testdb", "users").AddStage(bson.M{
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
				})
			},
			expected: []bson.M{
				{"$addFields": bson.M{
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
				}},
			},
		},
		{
			name: "add_fields_with_size_and_sum",
			buildFn: func() *aggregator.Aggregator[any] {
				return aggregator.NewAggregator[any](client, "testdb", "users").AddStage(bson.M{
					"$addFields": bson.M{
						"order_count": bson.M{"$size": "$user_orders"},
						"total_amount": bson.M{
							"$sum": "$user_orders.amount",
						},
					},
				})
			},
			expected: []bson.M{
				{"$addFields": bson.M{
					"order_count": bson.M{"$size": "$user_orders"},
					"total_amount": bson.M{
						"$sum": "$user_orders.amount",
					},
				}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agg := tt.buildFn()
			pipeline := agg.Pipeline()
			assert.Equal(t, tt.expected, pipeline)
		})
	}
}

// TestAggregatorLimit 测试 Limit 阶段的 BSON 构建
func TestAggregatorLimit(t *testing.T) {
	client := &tmorm.ORMClient{}

	tests := []struct {
		name     string
		buildFn  func() *aggregator.Aggregator[any]
		expected []bson.M
	}{
		{
			name: "limit_10",
			buildFn: func() *aggregator.Aggregator[any] {
				return aggregator.NewAggregator[any](client, "testdb", "users").Limit(10)
			},
			expected: []bson.M{
				{tmorm.LimitOp: int64(10)},
			},
		},
		{
			name: "limit_5",
			buildFn: func() *aggregator.Aggregator[any] {
				return aggregator.NewAggregator[any](client, "testdb", "users").Limit(5)
			},
			expected: []bson.M{
				{tmorm.LimitOp: int64(5)},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agg := tt.buildFn()
			pipeline := agg.Pipeline()
			assert.Equal(t, tt.expected, pipeline)
		})
	}
}

// TestAggregatorComplexPipeline 测试复杂管道的 BSON 构建
func TestAggregatorComplexPipeline(t *testing.T) {
	client := &tmorm.ORMClient{}

	tests := []struct {
		name     string
		buildFn  func() *aggregator.Aggregator[any]
		expected []bson.M
	}{
		{
			name: "match_group_sort_limit",
			buildFn: func() *aggregator.Aggregator[any] {
				return aggregator.NewAggregator[any](client, "testdb", "orders").
					Match(bson.M{"status": "completed"}).
					Group("$customerId", bson.M{
						"totalAmount": bson.M{tmorm.SumOp: "$amount"},
						"orderCount":  bson.M{tmorm.SumOp: 1},
					}).
					SortDesc("totalAmount").
					Limit(10)
			},
			expected: []bson.M{
				{tmorm.MatchOp: bson.M{"status": "completed"}},
				{tmorm.GroupOp: bson.M{
					tmorm.IdOp:    "$customerId",
					"totalAmount": bson.M{tmorm.SumOp: "$amount"},
					"orderCount":  bson.M{tmorm.SumOp: 1},
				}},
				{tmorm.SortOp: bson.D{{"totalAmount", -1}}},
				{tmorm.LimitOp: int64(10)},
			},
		},
		{
			name: "match_lookup_addfields_match_sort_limit",
			buildFn: func() *aggregator.Aggregator[any] {
				return aggregator.NewAggregator[any](client, "testdb", "users").
					Match(bson.M{"age": bson.M{"$gte": 25}}).
					Lookup("orders", "_id", "user_id", "user_orders").
					AddStage(bson.M{
						"$addFields": bson.M{
							"order_count": bson.M{"$size": "$user_orders"},
							"total_amount": bson.M{
								"$sum": "$user_orders.amount",
							},
						},
					}).
					Match(bson.M{"order_count": bson.M{"$gt": 0}}).
					SortDesc("total_amount").
					Limit(10)
			},
			expected: []bson.M{
				{tmorm.MatchOp: bson.M{"age": bson.M{"$gte": 25}}},
				{tmorm.StageLookUpOp: bson.M{
					"from":         "orders",
					"localField":   "_id",
					"foreignField": "user_id",
					"as":           "user_orders",
				}},
				{"$addFields": bson.M{
					"order_count": bson.M{"$size": "$user_orders"},
					"total_amount": bson.M{
						"$sum": "$user_orders.amount",
					},
				}},
				{tmorm.MatchOp: bson.M{"order_count": bson.M{"$gt": 0}}},
				{tmorm.SortOp: bson.D{{"total_amount", -1}}},
				{tmorm.LimitOp: int64(10)},
			},
		},
		{
			name: "match_unwind_group_sort",
			buildFn: func() *aggregator.Aggregator[any] {
				return aggregator.NewAggregator[any](client, "testdb", "users").
					Match(bson.M{"tags": bson.M{"$exists": true, "$ne": nil}}).
					Unwind("$tags").
					Group("$tags", bson.M{
						"user_count": bson.M{"$sum": 1},
						"users":      bson.M{"$push": "$name"},
					}).
					SortDesc("user_count")
			},
			expected: []bson.M{
				{tmorm.MatchOp: bson.M{"tags": bson.M{"$exists": true, "$ne": nil}}},
				{tmorm.UnwindOp: "$tags"},
				{tmorm.GroupOp: bson.M{
					tmorm.IdOp:   "$tags",
					"user_count": bson.M{"$sum": 1},
					"users":      bson.M{"$push": "$name"},
				}},
				{tmorm.SortOp: bson.D{{"user_count", -1}}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agg := tt.buildFn()
			pipeline := agg.Pipeline()
			assert.Equal(t, tt.expected, pipeline)
		})
	}
}

// TestAggregatorConditionalAggregation 测试条件聚合的 BSON 构建
func TestAggregatorConditionalAggregation(t *testing.T) {
	client := &tmorm.ORMClient{}

	tests := []struct {
		name     string
		buildFn  func() *aggregator.Aggregator[any]
		expected []bson.M
	}{
		{
			name: "conditional_sum_with_cond",
			buildFn: func() *aggregator.Aggregator[any] {
				return aggregator.NewAggregator[any](client, "testdb", "users").
					Group("$city", bson.M{
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
					})
			},
			expected: []bson.M{
				{tmorm.GroupOp: bson.M{
					tmorm.IdOp:    "$city",
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
				}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agg := tt.buildFn()
			pipeline := agg.Pipeline()
			assert.Equal(t, tt.expected, pipeline)
		})
	}
}
