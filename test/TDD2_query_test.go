package test

import (
	"testing"
	"tm_orm/collection"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/stretchr/testify/assert"
)

func TestQueryBase(t *testing.T) {
	// ctx := context.Background()
	// ormClient, _ := tmorm.NewORMClient(nil)
	// coll := collection.NewCollection[QueryTestUser](ormClient, "test_db", "query_test")

	testCase := []struct {
		name string

		bd bson.D

		before func(bd *bson.D)
		check  func(bd *bson.D) error
	}{
		{
			name: "测试简单查询 in nin",
			before: func(bd *bson.D) {
				b := collection.NewQuery().Where("key1").In(1, 2, 3, 4, 5).
					Where("age").NotIn(19, 20).
					Build()
				*bd = b.GetBsonD()
			},
			check: func(bd *bson.D) error {
				//want := bson.Data{{Key: "key1", Value: bson.E{Key: "$in", Value: []int{1, 2, 3, 4, 5}}}} // 没有泛型方法,所以没有[]int{}
				want := bson.D{
					{Key: "key1", Value: bson.D{{Key: "$in", Value: bson.A{1, 2, 3, 4, 5}}}},
					{Key: "age", Value: bson.D{{Key: "$nin", Value: bson.A{19, 20}}}},
				}
				assert.Equal(t, want, *bd)
				return nil
			},
		},
		{
			name: "测试简单查询 gt gte lt lte",
			before: func(bd *bson.D) {
				*bd = collection.NewQuery().
					Where("key1").Gt(1).
					Where("key2").Gte(2).
					Where("key3").Lte(3).
					Where("key4").Lt(4).
					Build().GetBsonD()
			},
			check: func(bd *bson.D) error {
				want := bson.D{
					{Key: "key1", Value: bson.D{{Key: "$gt", Value: 1}}},
					{Key: "key2", Value: bson.D{{Key: "$gte", Value: 2}}},
					{Key: "key3", Value: bson.D{{Key: "$lte", Value: 3}}},
					{Key: "key4", Value: bson.D{{Key: "$lt", Value: 4}}},
				}
				assert.Equal(t, want, *bd)
				return nil
			},
		},
		{
			name: "测试简单查询 eq ne exist regex",
			before: func(bd *bson.D) {
				*bd = collection.NewQuery().
					Where("key1").Eq(1).
					Where("key2").Ne(2).
					Where("key3").Exists(true).
					Where("key4").Regex(`(http://[.]+[^/])`).
					Build().GetBsonD()
			},
			check: func(bd *bson.D) error {
				want := bson.D{
					{Key: "key1", Value: bson.D{{Key: "$eq", Value: 1}}},
					{Key: "key2", Value: bson.D{{Key: "$ne", Value: 2}}},
					{Key: "key3", Value: bson.D{{Key: "$exists", Value: true}}},
					{Key: "key4", Value: bson.D{{Key: "$regex", Value: `(http://[.]+[^/])`}}},
				}
				assert.Equal(t, want, *bd)
				return nil
			},
		},
		{
			name: "原生bson查询",
			before: func(bd *bson.D) {
				*bd = collection.NewQuery().
					Where("key1").Eq(2).
					Where("key2").In(1, 2).Build().GetBsonD()

			},
			check: func(bd *bson.D) error {
				want := bson.D{
					{Key: "key1", Value: bson.D{{Key: "$eq", Value: 2}}},
					{Key: "key2", Value: bson.D{{"$in", bson.A{1, 2}}}},
				}
				assert.Equal(t, want, *bd)
				return nil
			},
		},
	}

	for _, t1 := range testCase {
		t.Run(t1.name, func(t *testing.T) {
			if t1.before != nil {
				t1.before(&t1.bd)
			}

			if t1.check != nil {
				if err := t1.check(&t1.bd); err != nil {
					t.Error(err)
				}
			}
		})
	}
}
