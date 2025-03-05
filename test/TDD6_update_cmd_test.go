package test

import (
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"testing"
	"tm_orm/updater"
)

func TestUpdateCmd(t *testing.T) {
	testCase := []struct {
		name string

		bd *updater.ReplaceBuilder[TestUser]

		before func(bd *updater.ReplaceBuilder[TestUser])
		check  func(bd *updater.ReplaceBuilder[TestUser]) error
	}{
		{
			name: "测试SetObj",
			bd:   updater.NewReplaceBuilder[TestUser](),
			before: func(bd *updater.ReplaceBuilder[TestUser]) {
				bd.C().SetObj(&TestUser{Name: "sean", Age: 20}, true)
			},
			check: func(bd *updater.ReplaceBuilder[TestUser]) error {
				want := bson.D{
					{Key: "$set", Value: bson.D{
						{Key: "name", Value: "sean"},
						{Key: "age", Value: int64(20)},
					}},
				}
				d := bd.GetBsonD()
				assert.Equal(t, want, d)
				return nil
			},
		},
		{
			name: "测试Set和Unset",
			bd:   updater.NewReplaceBuilder[TestUser](),
			before: func(bd *updater.ReplaceBuilder[TestUser]) {
				bd.C().Set("name", "sean").Unset("age")
			},
			check: func(bd *updater.ReplaceBuilder[TestUser]) error {
				want := bson.D{
					{Key: "$set", Value: bson.D{{Key: "name", Value: "sean"}}},
					{Key: "$unset", Value: bson.D{{Key: "age", Value: ""}}},
				}
				d := bd.GetBsonD()
				assert.Equal(t, want, d)
				return nil
			},
		},
		{
			name: "测试数组操作",
			bd:   updater.NewReplaceBuilder[TestUser](),
			before: func(bd *updater.ReplaceBuilder[TestUser]) {
				bd.C().AddToSet("tags", []string{"tag1", "tag2"}).
					Push("scores", []int{90, 95}).
					Pull("oldTags", "oldTag")
			},
			check: func(bd *updater.ReplaceBuilder[TestUser]) error {
				want := bson.D{
					{Key: "$addToSet", Value: bson.D{{Key: "tags", Value: []string{"tag1", "tag2"}}}},
					{Key: "$push", Value: bson.D{{Key: "scores", Value: []int{90, 95}}}},
					{Key: "$pull", Value: bson.D{{Key: "oldTags", Value: "oldTag"}}},
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
