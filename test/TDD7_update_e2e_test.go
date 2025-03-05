package test

import (
	"context"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"testing"
	"time"
	tmorm "tm_orm"
	"tm_orm/query"
	"tm_orm/updater"
)

// 端到端测试updater
func TestUpdaterE2E(t *testing.T) {
	ConnectMongo()
	up := &updater.MUpdater[TestUser]{}
	md := tmorm.NewMDB(MongoClient)
	ndb := "mytest"
	ncoll := "db_test"

	type tcase struct {
		name    string
		updater *updater.MUpdater[TestUser]

		data any

		before func(*tcase)
		after  func(*tcase)

		check func(*tcase)
	}

	testCases := []tcase{
		{
			name:    "update_one",
			updater: up,
			data: &TestUser{
				ID:   primitive.ObjectID([12]byte{1, 2, 3, 4, 5}),
				Name: "sean",
				Age:  20,
			},
			before: func(tc *tcase) {
				data := tc.data.(*TestUser)
				MongoClient.Database("mytest").Collection("db_test").InsertOne(context.Background(), data)
			},
			after: func(tc *tcase) {
				data := tc.data.(*TestUser)
				MongoClient.Database("mytest").Collection("db_test").DeleteMany(context.Background(),
					bson.M{"_id": data.ID})
			},
			check: func(tc *tcase) {
				data := tc.data.(*TestUser)
				var (
					wantErr error
					bd      = updater.NewReplaceBuilder[TestUser]()
				)
				bd.C().Set("age", 25)

				_, err := tc.updater.SetFilter(
					query.Query{}.Builder().K("_id").Eq(data.ID).ToQuery(),
				).UpdateOne(md.Sess(context.Background(), ndb, ncoll), bd)

				assert.Equal(t, wantErr, err)

				// 验证更新结果
				var result TestUser
				err = MongoClient.Database("mytest").Collection("db_test").FindOne(context.Background(),
					bson.M{"_id": data.ID}).Decode(&result)
				assert.NoError(t, err)
				assert.Equal(t, int64(25), result.Age)
			},
		},
		{
			name:    "update_many",
			updater: up,
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
					MongoClient.Database("mytest").Collection("db_test").InsertOne(context.Background(), user)
				}
			},
			after: func(tc *tcase) {
				data := tc.data.([]*TestUser)
				for _, user := range data {
					MongoClient.Database("mytest").Collection("db_test").DeleteMany(context.Background(),
						bson.M{"_id": user.ID})
				}
			},
			check: func(tc *tcase) {
				var (
					wantErr error
					bd      = updater.NewReplaceBuilder[TestUser]()
				)
				bd.C().Set("age", 30)

				_, err := tc.updater.SetFilter(
					query.Query{}.Builder().K("name").Eq("sean").ToQuery(),
				).UpdateMany(md.Sess(context.Background(), ndb, ncoll), bd)

				assert.Equal(t, wantErr, err)

				// 验证更新结果
				cursor, err := MongoClient.Database("mytest").Collection("db_test").Find(context.Background(),
					bson.M{"name": "sean"})
				assert.NoError(t, err)

				var results []TestUser
				err = cursor.All(context.Background(), &results)
				assert.NoError(t, err)
				assert.Equal(t, 2, len(results))
				for _, result := range results {
					assert.Equal(t, int64(30), result.Age)
				}
			},
		},
		{
			name:    "replaceOne",
			updater: up,
			data: &TestUser{
				ID:   primitive.ObjectID([12]byte{1, 2, 3, 4, 5}),
				Name: "sean",
				Age:  20,
			},
			before: func(tc *tcase) {},
			after: func(tc *tcase) {
				data := tc.data.(*TestUser)
				MongoClient.Database("mytest").Collection("db_test").DeleteMany(context.Background(),
					bson.M{"_id": data.ID})
			},
			check: func(tc *tcase) {
				data := tc.data.(*TestUser)
				var (
					wantErr error
					bd      = updater.NewReplaceBuilder[TestUser]()
				)
				bd.SetGetIdFunc(func() (any, bool) {
					return data.ID, true
				})
				bd.C().SetObj(data, false)

				_, err := tc.updater.ReplaceOne(md.Sess(context.Background(), ndb, ncoll), bd)
				assert.Equal(t, wantErr, err)

				// 验证更新结果
				var result TestUser
				err = MongoClient.Database("mytest").Collection("db_test").FindOne(context.Background(),
					bson.M{"_id": data.ID}).Decode(&result)
				assert.NoError(t, err)
				assert.Equal(t, data.Name, result.Name)
				assert.Equal(t, data.Age, result.Age)
			},
		},
		{
			name:    "upsertOne",
			updater: up,
			data: &TestUser{
				ID:        primitive.ObjectID([12]byte{1, 2, 3, 4, 6}),
				Name:      "sean_upsert",
				Age:       20,
				CreatedAt: time.Now(),
			},
			before: func(tc *tcase) {},
			after: func(tc *tcase) {
				data := tc.data.(*TestUser)
				MongoClient.Database("mytest").Collection("db_test").DeleteMany(context.Background(),
					bson.M{"_id": data.ID})
			},
			check: func(tc *tcase) {
				data := tc.data.(*TestUser)
				var (
					wantErr error
					bd      = updater.NewReplaceBuilder[TestUser]()
				)
				bd.SetGetIdFunc(func() (any, bool) {
					return data.ID, true
				})
				bd.C().SetObj(data, false)

				_, err := tc.updater.UpsertOne(md.Sess(context.Background(), ndb, ncoll), bd)
				assert.Equal(t, wantErr, err)

				// 验证更新结果
				var result TestUser
				err = MongoClient.Database("mytest").Collection("db_test").FindOne(context.Background(),
					bson.M{"_id": data.ID}).Decode(&result)
				assert.NoError(t, err)
				assert.Equal(t, data.Name, result.Name)
				assert.Equal(t, data.Age, result.Age)
				assert.Equal(t, data.CreatedAt.Unix(), result.CreatedAt.Unix())
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.before != nil {
				tc.before(&tc)
			}
			tc.check(&tc)
			if tc.after != nil {
				tc.after(&tc)
			}
		})
	}
}
