package test

import (
	"context"
	"fmt"
	"testing"
	tmorm "tm_orm"
	"tm_orm/finder"
	"tm_orm/query"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// 端到端测试finder

func TestFinderE2E(t *testing.T) {
	ConnectMongo()
	fd := &finder.Finder[TestUser]{}
	md := tmorm.NewMDB(MongoClient)
	ndb := "mytest"
	ncoll := "db_test"

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
				resList, err := tc.finder.Find(
					md.Sess(context.Background(), ndb, ncoll),
					fil,
				)

				assert.Equal(t, wantErr, err)
				assert.Equal(t, wantRes, resList)
			},
		},
		{
			name:   "and or",
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
						Or(func(a *query.QueryOr) query.Builder {
							return a.
								K("age").Gt(10).
								K("name").Eq("xx").
								And(func(and *query.QueryAnd) query.Builder {
									return and.
										K("age").Gt(100).
										K("name").Eq("sean")
								})
						}).ToQuery()
				)
				//fmt.Println(fil.GetBsonD())
				resList, err := tc.finder.Find(
					md.Sess(context.Background(), ndb, ncoll),
					fil,
				)

				assert.Equal(t, wantErr, err)
				assert.Equal(t, wantRes, resList)
			},
		},
		{
			name:   "expr",
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
				MongoClient.Database("mytest").Collection("db_test").UpdateOne(context.Background(),
					bson.M{"_id": primitive.ObjectID([12]byte{1, 2, 3, 4, 5, 9, 8, 6})}, bson.M{"$set": &TestUser{
						ID:   primitive.ObjectID([12]byte{1, 2, 3, 4, 5, 9, 8, 6}),
						Name: "sean1",
						Age:  50,
					}}, options.Update().SetUpsert(true))
			},
			after: func(tc *tcase) {
				data := tc.data.(*TestUser)
				MongoClient.Database("mytest").Collection("db_test").DeleteMany(context.Background(),
					bson.M{"_id": data.ID})
				MongoClient.Database("mytest").Collection("db_test").DeleteMany(context.Background(),
					bson.M{"_id": primitive.ObjectID([12]byte{1, 2, 3, 4, 5, 9, 8, 6})})
			},
			check: func(tc *tcase) {
				var (
					wantErr error
					wantRes = []*TestUser{
						tc.data.(*TestUser),
					}
					fil = query.Query{}.Builder().
						Expr(func(m query.MExpr) query.Builder {
							return m.Or(
								m.C().Eq(m.Fd("name"), m.Val("sean2")),
								m.Er().And(
									m.C().Eq(
										m.C().Multi(m.Fd("age"), m.Val(2)),
										m.Val(40),
									),
									m.C().In(m.Fd("name"), m.Val("sean")),
								),
							)
						}).
						ToQuery()
				)
				//fmt.Println(fil.GetBsonD())
				resList, err := tc.finder.Find(
					md.Sess(context.Background(), ndb, ncoll),
					fil,
				)

				assert.Equal(t, wantErr, err)
				assert.Equal(t, wantRes, resList)
			},
		},
		{
			name:   "count",
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
				_, err := MongoClient.Database("mytest").Collection("db_test").UpdateOne(context.Background(),
					bson.M{"_id": primitive.ObjectID([12]byte{1, 2, 3, 4, 5, 8, 6})}, bson.M{"$set": &TestUser{
						ID:   primitive.ObjectID([12]byte{1, 2, 3, 4, 5, 8, 6}),
						Name: "sean1",
						Age:  50,
					}}, options.Update().SetUpsert(true))
				if err != nil {
					panic(err)
				}
			},
			after: func(tc *tcase) {
				data := tc.data.(*TestUser)
				MongoClient.Database("mytest").Collection("db_test").DeleteMany(context.Background(),
					bson.M{"_id": data.ID})
				MongoClient.Database("mytest").Collection("db_test").DeleteMany(context.Background(),
					bson.M{"_id": primitive.ObjectID([12]byte{1, 2, 3, 4, 5, 8, 6})})
			},
			check: func(tc *tcase) {
				var (
					wantErr error
					wantRes int64 = 2
					fil           = query.Query{}.Builder().
						Or(func(a *query.QueryOr) query.Builder {
							return a.K("age").Gt(10)
						}).
						ToQuery()
				)
				//fmt.Println(fil.GetBsonD())
				co, err := tc.finder.Count(
					md.Sess(context.Background(), ndb, ncoll),
					fil,
				)

				assert.Equal(t, wantErr, err)
				assert.Equal(t, wantRes, co)
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
