package test

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"testing"
	tmorm "tm_orm"
	"tm_orm/finder"
	"tm_orm/middleware"
	"tm_orm/query"
)

func TestMiddlewareE2E(t *testing.T) {
	ConnectMongo()
	fd := &finder.Finder[TestUser]{}
	md := tmorm.NewMDB(MongoClient)
	ndb := "mytest"
	ncoll := "db_test"
	md.GetMiddleware().Use(
		middleware.SLowQueryMiddleware{Threshold: 1}.Build(),
		func(next tmorm.MHandlerFunc) tmorm.MHandlerFunc {
			return func(mctx *tmorm.MiddleCtx) tmorm.MResult {
				println("db 前置")
				return next(mctx)
			}
		})

	type tcase struct {
		name   string
		finder *finder.Finder[TestUser]

		data any

		before func(*tcase)
		after  func(*tcase)

		check func(*tcase)
	}

	testCases := []tcase{
		{
			name:   "normal",
			finder: fd,
			data: &TestUser{
				ID:   primitive.ObjectID([12]byte{1, 2, 3, 4, 5}),
				Name: "sean",
				Age:  20,
			},
			before: func(tc *tcase) {
				data := tc.data.(*TestUser)
				MongoClient.Database("mytest").Collection("db_test").UpdateOne(context.Background(),
					bson.M{"_id": data.ID}, bson.M{"$set": data}, options.Update().SetUpsert(true))
			},
			after: func(tc *tcase) {
				data := tc.data.(*TestUser)
				MongoClient.Database("mytest").Collection("db_test").DeleteMany(context.Background(),
					bson.M{"_id": data.ID})
			},
			check: func(tc *tcase) {
				var (
					wantErr error
					wantRes = []*TestUser{
						tc.data.(*TestUser),
					}
					fil = query.Query{}.Builder().
						K("name").Eq("sean").
						K("age").Gte(18).ToQuery()
				)
				fmt.Println(fil.GetBsonD())

				ss := md.Sess(context.Background(), ndb, ncoll,
					func(next tmorm.MHandlerFunc) tmorm.MHandlerFunc {
						return func(mctx *tmorm.MiddleCtx) tmorm.MResult {
							r := next(mctx)
							println("session 后置")
							return r
						}
					})
				resList, err := tc.finder.Find(ss, fil)

				assert.Equal(t, wantErr, err)
				assert.Equal(t, wantRes, resList)
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
