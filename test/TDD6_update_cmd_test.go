package test

import (
	"github.com/stretchr/testify/assert"
	"testing"
	tmorm "tm_orm"
	"tm_orm/collection"

	"go.mongodb.org/mongo-driver/bson"
)

func TestUpdateCmd(t *testing.T) {
	ConnectMongo()
	ormClient, _ := tmorm.NewORMClient(nil)
	coll := collection.NewCollection[TestUser](ormClient, "mytest", "db_test")

	testCase := []struct {
		name   string
		before func() *collection.UpdateBuilder[TestUser]
		check  func(bd *collection.UpdateBuilder[TestUser]) error
	}{
		{
			name: "测试Set操作",
			before: func() *collection.UpdateBuilder[TestUser] {
				return coll.Where("_id").Eq("test").Set("name", "sean").Set("age", 20)
			},
			check: func(bd *collection.UpdateBuilder[TestUser]) error {
			updates := bd.GetUpdates().GetBsonD()
			// 验证包含 $set 操作
			assert.True(t, bd.GetUpdates().HasOperation(tmorm.SetOp))
			// 验证 $set 操作的内容
			setOp, exists := bd.GetUpdates().GetOperation(tmorm.SetOp)
			assert.True(t, exists)
			assert.Contains(t, setOp, bson.E{Key: "name", Value: "sean"})
			assert.Contains(t, setOp, bson.E{Key: "age", Value: 20})
			// 验证整体 BSON 结构
			assert.NotEmpty(t, updates)
			return nil
		},
		},
		{
			name: "测试Set和Unset",
			before: func() *collection.UpdateBuilder[TestUser] {
				return coll.Where("_id").Eq("test").Set("name", "sean").Unset("age")
			},
			check: func(bd *collection.UpdateBuilder[TestUser]) error {
			// 验证包含 $set 和 $unset 操作
			assert.True(t, bd.GetUpdates().HasOperation(tmorm.SetOp))
			assert.True(t, bd.GetUpdates().HasOperation(tmorm.UnsetOp))
			// 验证 $set 操作内容
			setOp, _ := bd.GetUpdates().GetOperation(tmorm.SetOp)
			assert.Contains(t, setOp, bson.E{Key: "name", Value: "sean"})
			// 验证 $unset 操作内容
			unsetOp, _ := bd.GetUpdates().GetOperation(tmorm.UnsetOp)
			assert.Contains(t, unsetOp, bson.E{Key: "age", Value: ""})
			return nil
		},
		},
		{
			name: "测试数组操作",
			before: func() *collection.UpdateBuilder[TestUser] {
				return coll.Where("_id").Eq("test").Set("name", "test").
					AddToSet("tags", []string{"tag1", "tag2"}).
					Push("scores", []int{90, 95}).
					Pull("oldTags", "oldTag")
			},
			check: func(bd *collection.UpdateBuilder[TestUser]) error {
				// 验证包含所有数组操作
				assert.True(t, bd.GetUpdates().HasOperation(tmorm.SetOp))
				assert.True(t, bd.GetUpdates().HasOperation(tmorm.AddToSetOp))
				assert.True(t, bd.GetUpdates().HasOperation(tmorm.PushOp))
				assert.True(t, bd.GetUpdates().HasOperation(tmorm.PullOp))
				// 验证具体操作内容
				addToSetOp, _ := bd.GetUpdates().GetOperation(tmorm.AddToSetOp)
				assert.Contains(t, addToSetOp, bson.E{Key: "tags", Value: []string{"tag1", "tag2"}})
				return nil
			},
		},
		{
			name: "测试Inc和Mul操作",
			before: func() *collection.UpdateBuilder[TestUser] {
				return coll.Where("_id").Eq("test").Inc("age", 1).Mul("score", 1.5)
			},
			check: func(bd *collection.UpdateBuilder[TestUser]) error {
				// 验证包含 $inc 和 $mul 操作
				assert.True(t, bd.GetUpdates().HasOperation(tmorm.IncOp))
				assert.True(t, bd.GetUpdates().HasOperation(tmorm.MulOp))
				// 验证操作内容
				incOp, _ := bd.GetUpdates().GetOperation(tmorm.IncOp)
				assert.Contains(t, incOp, bson.E{Key: "age", Value: 1})
				mulOp, _ := bd.GetUpdates().GetOperation(tmorm.MulOp)
				assert.Contains(t, mulOp, bson.E{Key: "score", Value: 1.5})
				return nil
			},
		},
		{
			name: "测试Rename操作",
			before: func() *collection.UpdateBuilder[TestUser] {
				return coll.Where("_id").Eq("test").Set("name", "test").Rename("oldName", "newName")
			},
			check: func(bd *collection.UpdateBuilder[TestUser]) error {
				// 验证包含 $set 和 $rename 操作
				assert.True(t, bd.GetUpdates().HasOperation(tmorm.SetOp))
				assert.True(t, bd.GetUpdates().HasOperation(tmorm.RenameOp))
				// 验证 $rename 操作内容
				renameOp, _ := bd.GetUpdates().GetOperation(tmorm.RenameOp)
				assert.Contains(t, renameOp, bson.E{Key: "oldName", Value: "newName"})
				return nil
			},
		},
		{
			name: "测试SetOnInsert和CurrentDate操作",
			before: func() *collection.UpdateBuilder[TestUser] {
				return coll.Where("_id").Eq("test").Set("name", "test").SetOnInsert("createdBy", "system").CurrentDate("updatedAt")
			},
			check: func(bd *collection.UpdateBuilder[TestUser]) error {
				// 验证包含所有操作
				assert.True(t, bd.GetUpdates().HasOperation(tmorm.SetOp))
				assert.True(t, bd.GetUpdates().HasOperation(tmorm.SetOnInsertOp))
				assert.True(t, bd.GetUpdates().HasOperation(tmorm.CurrentDateOp))
				// 验证操作内容
				setOnInsertOp, _ := bd.GetUpdates().GetOperation(tmorm.SetOnInsertOp)
				assert.Contains(t, setOnInsertOp, bson.E{Key: "createdBy", Value: "system"})
				currentDateOp, _ := bd.GetUpdates().GetOperation(tmorm.CurrentDateOp)
				assert.Contains(t, currentDateOp, bson.E{Key: "updatedAt", Value: true})
				return nil
			},
		},
		{
			name: "测试Pop和PullAll操作",
			before: func() *collection.UpdateBuilder[TestUser] {
				return coll.Where("_id").Eq("test").Set("name", "test").Pop("items", 1).PullAll("tags", []any{"tag1", "tag2"})
			},
			check: func(bd *collection.UpdateBuilder[TestUser]) error {
				// 验证包含所有操作
				assert.True(t, bd.GetUpdates().HasOperation(tmorm.SetOp))
				assert.True(t, bd.GetUpdates().HasOperation(tmorm.PopOp))
				assert.True(t, bd.GetUpdates().HasOperation(tmorm.PullAllOp))
				// 验证操作内容
				popOp, _ := bd.GetUpdates().GetOperation(tmorm.PopOp)
				assert.Contains(t, popOp, bson.E{Key: "items", Value: 1})
				pullAllOp, _ := bd.GetUpdates().GetOperation(tmorm.PullAllOp)
				assert.Contains(t, pullAllOp, bson.E{Key: "tags", Value: []any{"tag1", "tag2"}})
				return nil
			},
		},
		{
			name: "测试复合操作",
			before: func() *collection.UpdateBuilder[TestUser] {
				return coll.Where("_id").Eq("test").
					Set("name", "john").
					Inc("age", 1).
					Unset("oldField").
					CurrentDate("lastModified")
			},
			check: func(bd *collection.UpdateBuilder[TestUser]) error {
				// 验证包含所有复合操作
				assert.True(t, bd.GetUpdates().HasOperation(tmorm.SetOp))
				assert.True(t, bd.GetUpdates().HasOperation(tmorm.IncOp))
				assert.True(t, bd.GetUpdates().HasOperation(tmorm.UnsetOp))
				assert.True(t, bd.GetUpdates().HasOperation(tmorm.CurrentDateOp))
				// 验证整体 BSON 结构完整性
				updates := bd.GetUpdates().GetBsonD()
				assert.Len(t, updates, 4) // 应该有4个不同的操作
				// 验证具体操作内容
				setOp, _ := bd.GetUpdates().GetOperation(tmorm.SetOp)
				assert.Contains(t, setOp, bson.E{Key: "name", Value: "john"})
				incOp, _ := bd.GetUpdates().GetOperation(tmorm.IncOp)
				assert.Contains(t, incOp, bson.E{Key: "age", Value: 1})
				return nil
			},
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			bd := tc.before()
			err := tc.check(bd)
			assert.NoError(t, err)
		})
	}
}
