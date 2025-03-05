package test

import (
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"testing"
	"tm_orm/query"
)

func TestQueryExpr(t *testing.T) {
	testCase := []struct {
		name string

		bd *query.Query

		before func(bd *query.Query)
		check  func(bd *query.Query) error
	}{
		{
			name: "测试 normal",
			bd:   &query.Query{},
			before: func(bd *query.Query) {
				//a := bd.Builder().
				//	Expr().AggCmd().Gte(query.F("age"), 33)

				a := bd.Builder().
					Expr(func(m query.MExpr) query.Builder {
						return m.Gte(query.F("age"), 33)
					})

				*bd = a.ToQuery()
			},
			check: func(bd *query.Query) error {
				//want := bson.Data{{Key: "key1", Value: bson.E{Key: "$in", Value: []int{1, 2, 3, 4, 5}}}} // 没有泛型方法,所以没有[]int{}
				want := bson.D{
					{Key: "$expr", Value: bson.D{{
						Key:   "$gte",
						Value: bson.A{"$age", 33}}},
					},
				}
				d := bd.GetBsonD()
				assert.Equal(t, want, d)
				return nil
			},
		},
		{
			name: "测试嵌套expr",
			bd:   &query.Query{},
			before: func(bd *query.Query) {
				a := bd.Builder().
					Expr(func(m query.MExpr) query.Builder {
						return m.Gte(
							m.C().Multi(m.Fd("score"), 100),
							500,
						)
					})
				*bd = a.ToQuery()
			},
			check: func(bd *query.Query) error {
				//want := bson.Data{{Key: "key1", Value: bson.E{Key: "$in", Value: []int{1, 2, 3, 4, 5}}}} // 没有泛型方法,所以没有[]int{}
				want := bson.D{
					{
						Key: "$expr",
						Value: bson.D{{
							Key: "$gte",
							Value: bson.A{
								bson.D{{Key: "$multiply", Value: bson.A{"$score", 100}}},
								500,
							},
						}},
					},
				}
				d := bd.GetBsonD()
				assert.Equal(t, want, d)
				return nil
			},
		},
		{
			name: "测试 expr 嵌套 and 与 or",
			bd:   &query.Query{},
			before: func(bd *query.Query) {
				a := bd.Builder().
					Expr(func(m query.MExpr) query.Builder {
						return m.And(
							m.C().Gte(m.Fd("age"), 22),
							m.C().Gte(
								m.Fd("salary"),
								m.C().Multi(
									m.C().Avg(m.Fd("salary")),
									2,
								),
							),
							m.Er().Or(
								m.C().Eq(m.Fd("name"), "sean"),
								m.C().Gt(m.Fd("age"), 18),
							),
						)
					})
				*bd = a.ToQuery()
			},
			check: func(bd *query.Query) error {
				//want := bson.Data{{Key: "key1", Value: bson.E{Key: "$in", Value: []int{1, 2, 3, 4, 5}}}} // 没有泛型方法,所以没有[]int{}
				want := bson.D{
					{
						Key: "$expr",
						Value: bson.D{{
							Key: "$and",
							Value: bson.A{
								bson.D{{Key: "$gte", Value: bson.A{"$age", 22}}},
								bson.D{{Key: "$gte",
									Value: bson.A{
										"$salary",
										bson.D{{
											Key: "$multiply",
											Value: bson.A{
												bson.D{{
													Key:   "$avg",
													Value: bson.A{"$salary"},
												}},
												2,
											},
										}},
									}},
								},
								bson.D{{Key: "$or",
									Value: bson.A{
										bson.D{{"$eq", bson.A{"$name", "sean"}}},
										bson.D{{"$gt", bson.A{"$age", 18}}},
									}}},
							},
						}},
					},
				}
				d := bd.GetBsonD()
				assert.Equal(t, want, d)
				return nil
			},
		},
	}

	for _, t1 := range testCase {
		t.Run(t1.name, func(t *testing.T) {
			if t1.before != nil {
				t1.before(t1.bd)
			}

			if t1.check != nil {
				if err := t1.check(t1.bd); err != nil {
					t.Error(err)
				}
			}
		})
	}
}
