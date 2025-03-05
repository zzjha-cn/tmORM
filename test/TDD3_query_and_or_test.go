package test

import (
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"testing"
	"tm_orm/query"
)

func TestQueryAndOr(t *testing.T) {
	testCase := []struct {
		name string

		bd *query.Query

		before func(bd *query.Query)
		check  func(bd *query.Query) error
	}{
		{
			name: "测试and",
			bd:   &query.Query{},
			before: func(bd *query.Query) {
				a := bd.Builder().
					And(func(a *query.QueryAnd) query.Builder {
						b := a.K("key1").Gt(13).
							K("key2").In(11, 22, 33)
						return b
					})

				*bd = a.ToQuery()
			},
			check: func(bd *query.Query) error {
				//want := bson.Data{{Key: "key1", Value: bson.E{Key: "$in", Value: []int{1, 2, 3, 4, 5}}}} // 没有泛型方法,所以没有[]int{}
				want := bson.D{
					{Key: "$and", Value: bson.A{
						bson.D{{Key: "key1", Value: bson.D{{"$gt", 13}}}},
						bson.D{{Key: "key2", Value: bson.D{{"$in", []any{11, 22, 33}}}}},
					}},
				}
				d := bd.GetBsonD()
				assert.Equal(t, want, d)
				return nil
			},
		},
		{
			name: "测试or",
			bd:   &query.Query{},
			before: func(bd *query.Query) {
				a := bd.Builder().
					Or(func(a *query.QueryOr) query.Builder {
						b := a.K("key1").Gt(13).
							K("key2").In(11, 22, 33)
						return b
					})

				*bd = a.ToQuery()
			},
			check: func(bd *query.Query) error {
				//want := bson.Data{{Key: "key1", Value: bson.E{Key: "$in", Value: []int{1, 2, 3, 4, 5}}}} // 没有泛型方法,所以没有[]int{}
				want := bson.D{
					{Key: "$or", Value: bson.A{
						bson.D{{Key: "key1", Value: bson.D{{"$gt", 13}}}},
						bson.D{{Key: "key2", Value: bson.D{{"$in", []any{11, 22, 33}}}}},
					}},
				}
				d := bd.GetBsonD()
				assert.Equal(t, want, d)
				return nil
			},
		},
		{
			name: "测试or 与 and 嵌套",
			bd:   &query.Query{},
			before: func(bd *query.Query) {
				a := bd.Builder().
					Or(func(a *query.QueryOr) query.Builder {
						b := a.
							K("key1").Gt(13).
							K("key2").In(11, 22, 33).
							And(func(and *query.QueryAnd) query.Builder {
								return and.K("key3").Exists("sean").K("key4").Lte(100)
							})
						return b
					})

				*bd = a.ToQuery()
			},
			check: func(bd *query.Query) error {
				//want := bson.Data{{Key: "key1", Value: bson.E{Key: "$in", Value: []int{1, 2, 3, 4, 5}}}} // 没有泛型方法,所以没有[]int{}
				want := bson.D{
					{Key: "$or", Value: bson.A{
						bson.D{{Key: "key1", Value: bson.D{{"$gt", 13}}}},
						bson.D{{Key: "key2", Value: bson.D{{"$in", []any{11, 22, 33}}}}},
						bson.D{{Key: "$and", Value: bson.A{
							bson.D{{Key: "key3", Value: bson.D{{"$exists", "sean"}}}},
							bson.D{{Key: "key4", Value: bson.D{{"$lte", 100}}}},
						}}},
					}},
				}
				d := bd.GetBsonD()
				assert.Equal(t, want, d)
				return nil
			},
		},
		{
			name: "测试and与expr嵌套",
			bd:   &query.Query{},
			before: func(bd *query.Query) {
				a := bd.Builder().
					And(func(a *query.QueryAnd) query.Builder {
						b := a.Expr(func(m query.MExpr) query.Builder {
							return m.Eq(m.Fd("name"), "sean")
						})
						return b
					})

				*bd = a.ToQuery()
			},
			check: func(bd *query.Query) error {
				want := bson.D{
					{Key: "$and", Value: bson.A{
						bson.D{{Key: "$expr", Value: bson.D{{"$eq", bson.A{"$name", "sean"}}}}},
					}},
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
