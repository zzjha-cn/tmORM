package test

import (
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"testing"
	"tm_orm/query"
)

func TestQueryBase(t *testing.T) {
	testCase := []struct {
		name string

		bd *query.Query

		before func(bd *query.Query)
		check  func(bd *query.Query) error
	}{
		{
			name: "测试简单查询 in nin",
			bd:   &query.Query{},
			before: func(bd *query.Query) {
				a := bd.Builder().
					K("key1").In(1, 2, 3, 4, 5).
					K("age").NIn(19, 20).ToQuery()
				*bd = a
			},
			check: func(bd *query.Query) error {
				//want := bson.Data{{Key: "key1", Value: bson.E{Key: "$in", Value: []int{1, 2, 3, 4, 5}}}} // 没有泛型方法,所以没有[]int{}
				want := bson.D{
					{Key: "key1", Value: bson.D{{Key: "$in", Value: bson.A{1, 2, 3, 4, 5}}}},
					{Key: "age", Value: bson.D{{Key: "$nin", Value: bson.A{19, 20}}}},
				}
				d := bd.GetBsonD()
				assert.Equal(t, want, d)
				return nil
			},
		},
		{
			name: "测试简单查询 gt gte lt lte",
			bd:   &query.Query{},
			before: func(bd *query.Query) {
				*bd = bd.Builder().
					K("key1").Gt(1).
					K("key2").Gte(2).
					K("key3").Lte(3).
					K("key4").Lt(4).ToQuery()

			},
			check: func(bd *query.Query) error {
				want := bson.D{
					{Key: "key1", Value: bson.D{{Key: "$gt", Value: 1}}},
					{Key: "key2", Value: bson.D{{Key: "$gte", Value: 2}}},
					{Key: "key3", Value: bson.D{{Key: "$lte", Value: 3}}},
					{Key: "key4", Value: bson.D{{Key: "$lt", Value: 4}}},
				}
				d := bd.GetBsonD()
				assert.Equal(t, want, d)
				return nil
			},
		},
		{
			name: "测试简单查询 eq ne exist regex",
			bd:   &query.Query{},
			before: func(bd *query.Query) {
				*bd = bd.Builder().
					K("key1").Eq(1).
					K("key2").Ne(2).
					K("key3").Exists("val3").
					K("key4").Regex(`(http://[.]+[^/])`).ToQuery()

			},
			check: func(bd *query.Query) error {
				want := bson.D{
					{Key: "key1", Value: bson.D{{Key: "$eq", Value: 1}}},
					{Key: "key2", Value: bson.D{{Key: "$ne", Value: 2}}},
					{Key: "key3", Value: bson.D{{Key: "$exists", Value: "val3"}}},
					{Key: "key4", Value: bson.D{{Key: "$regex", Value: `(http://[.]+[^/])`}}},
				}
				d := bd.GetBsonD()
				assert.Equal(t, want, d)
				return nil
			},
		},
		{
			name: "原生bson查询",
			bd:   &query.Query{},
			before: func(bd *query.Query) {
				*bd = bd.Builder().
					KV("key1", 2).
					KV("key2", bson.M{"$in": []int{1, 2}}).ToQuery()

			},
			check: func(bd *query.Query) error {
				want := bson.D{
					{Key: "key1", Value: 2},
					{Key: "key2", Value: bson.M{"$in": []int{1, 2}}},
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
