package test

import (
	"context"
	"fmt"
	"testing"
	tmorm "tm_orm"
	"tm_orm/collection"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestMiddlewareE2E(t *testing.T) {
	ConnectMongo()
	ctx := context.Background()
	ndb := "mytest"
	ncoll := "db_test"

	// 创建带有middleware的collection
	coll := collection.NewCollection[TestUser](client, ndb, ncoll,
		// 慢查询中间件
		func(mctx *tmorm.MiddleCtx, next func(m *tmorm.MiddleCtx)) {
			fmt.Println("慢查询检测中间件")
			next(mctx)
		},
		// 日志中间件
		func(mctx *tmorm.MiddleCtx, next func(m *tmorm.MiddleCtx)) {
			fmt.Printf("执行操作: %s\n", string(mctx.Typ))
			next(mctx)
			fmt.Printf("操作完成: %s\n", string(mctx.Typ))
		},
	)

	type tcase struct {
		name   string
		data   *TestUser
		before func(*tcase)
		after  func(*tcase)
		check  func(*tcase)
	}

	testCases := []tcase{
		{
			name: "normal",
			data: &TestUser{
				ID:   primitive.ObjectID([12]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}),
				Name: "sean",
				Age:  25,
			},
			before: func(tc *tcase) {
				// 清理测试数据
				coll.Where("name").Eq("sean").Delete(ctx)
			},
			after: func(tc *tcase) {
				// 插入测试数据
				_, err := coll.Insert(ctx, tc.data)
				assert.NoError(t, err)
			},
			check: func(tc *tcase) {
				// 验证数据是否插入成功
				user, err := coll.Where("name").Eq("sean").FindOne(ctx)
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, "sean", user.Name)
				assert.Equal(t, int64(25), user.Age)
			},
		},
		{
			name: "update_test",
			data: &TestUser{
				ID:   primitive.ObjectID([12]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 13}),
				Name: "alice",
				Age:  30,
			},
			before: func(tc *tcase) {
				// 清理并插入测试数据
				coll.Where("name").Eq("alice").Delete(ctx)
				_, err := coll.Insert(ctx, tc.data)
				assert.NoError(t, err)
			},
			after: func(tc *tcase) {
				// 更新数据
				_, err := coll.Where("name").Eq("alice").Set("age", 31).Update(ctx)
				assert.NoError(t, err)
			},
			check: func(tc *tcase) {
				// 验证更新是否成功
				user, err := coll.Where("name").Eq("alice").FindOne(ctx)
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, "alice", user.Name)
				assert.Equal(t, int64(31), user.Age)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.before != nil {
				tc.before(&tc)
			}
			if tc.after != nil {
				tc.after(&tc)
			}
			if tc.check != nil {
				tc.check(&tc)
			}
		})
	}
}
