package test

import (
	"context"
	"testing"
	tmorm "tm_orm"
	"tm_orm/aggregator"
	"tm_orm/query"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestAggregateCmd(t *testing.T) {
	ConnectMongo()
	md := tmorm.NewMDB(MongoClient)
	ndb := "mytest"
	ncoll := "db_test"

	type tcc struct {
		name   string
		agg    *aggregator.Aggregator[TestUser]
		data   any
		before func(cc *tcc)
		after  func(cc *tcc)
		check  func(cc *tcc)
	}

	testCase := []tcc{
		{
			name: "测试基本聚合操作",
			agg:  aggregator.NewAggregator[TestUser](),
			data: []*TestUser{
				{
					ID:   primitive.ObjectID([12]byte{1, 2, 3, 4, 5}),
					Name: "sean",
					Age:  20,
				},
				{
					ID:   primitive.ObjectID([12]byte{1, 2, 3, 4, 6}),
					Name: "sean",
					Age:  25,
				},
			},
			before: func(tc *tcc) {
				data := tc.data.([]*TestUser)
				for _, user := range data {
					MongoClient.Database("mytest").Collection("db_test").UpdateOne(context.Background(),
						bson.M{"_id": user.ID}, bson.M{"$set": user}, options.Update().SetUpsert(true))
				}
			},
			after: func(tc *tcc) {
				data := tc.data.([]*TestUser)
				for _, user := range data {
					MongoClient.Database("mytest").Collection("db_test").DeleteMany(context.Background(),
						bson.M{"_id": user.ID})
				}
			},
			check: func(tc *tcc) {
				var (
					wantErr error
				)

				tc.agg.Pipe().
					Match(func(m *query.MatchCmd) query.Builder {
						return m.K("name").Eq("sean")
					}).
					Group(func(gc *query.GroupCmd) query.Builder {
						return gc.IdWithField("name").
							Key("avgAge").Avg(gc.ToFd("age")).
							Key("count").Sum(gc.AnyVal(1)).Build()
					})

				m := []*struct {
					Id     string  `bson:"_id"`
					AvgAge float64 `bson:"avgAge"`
					Count  int     `bson:"count"`
				}{}
				err := aggregator.WithParseAggregate(md.Sess(context.Background(), ndb, ncoll), tc.agg, &m)
				if err != nil {
					t.Error(err)
					return
				}
				assert.Equal(t, wantErr, err)
				assert.Equal(t, 1, len(m))

				// 验证聚合结果
				expectedResult := &struct {
					Id     string  `bson:"_id"`
					AvgAge float64 `bson:"avgAge"`
					Count  int     `bson:"count"`
				}{Id: "sean", AvgAge: 22.5, Count: 2}

				assert.Equal(t, expectedResult, m[0])
			},
		},
		{
			name: "测试复杂分组和多聚合函数",
			agg:  aggregator.NewAggregator[TestUser](),
			data: []*TestUser{
				{
					ID:   primitive.ObjectID([12]byte{1, 2, 3, 4, 5}),
					Name: "sean",
					Age:  20,
				},
				{
					ID:   primitive.ObjectID([12]byte{1, 2, 3, 4, 6}),
					Name: "sean",
					Age:  25,
				},
				{
					ID:   primitive.ObjectID([12]byte{1, 2, 3, 4, 7}),
					Name: "tom",
					Age:  30,
				},
			},
			before: func(tc *tcc) {
				data := tc.data.([]*TestUser)
				for _, user := range data {
					MongoClient.Database("mytest").Collection("db_test").UpdateOne(context.Background(),
						bson.M{"_id": user.ID}, bson.M{"$set": user}, options.Update().SetUpsert(true))
				}
			},
			after: func(tc *tcc) {
				data := tc.data.([]*TestUser)
				for _, user := range data {
					MongoClient.Database("mytest").Collection("db_test").DeleteMany(context.Background(),
						bson.M{"_id": user.ID})
				}
			},
			check: func(tc *tcc) {
				var (
					wantErr error
				)

				tc.agg.Pipe().
					Group(func(gc *query.GroupCmd) query.Builder {
						return gc.IdWithField("name").
							Key("minAge").Min(gc.ToFd("age")).
							Key("maxAge").Max(gc.ToFd("age")).
							Key("avgAge").Avg(gc.ToFd("age")).
							Key("count").Sum(gc.AnyVal(1)).Build()
					}).
					Sort("avgAge")

				m := []*struct {
					Id     string  `bson:"_id"`
					MinAge int64   `bson:"minAge"`
					MaxAge int64   `bson:"maxAge"`
					AvgAge float64 `bson:"avgAge"`
					Count  int     `bson:"count"`
				}{}
				err := aggregator.WithParseAggregate(md.Sess(context.Background(), ndb, ncoll), tc.agg, &m)
				if err != nil {
					t.Error(err)
					return
				}
				assert.Equal(t, wantErr, err)
				assert.Equal(t, 2, len(m))

				// 验证聚合结果
				expectedResults := []*struct {
					Id     string  `bson:"_id"`
					MinAge int64   `bson:"minAge"`
					MaxAge int64   `bson:"maxAge"`
					AvgAge float64 `bson:"avgAge"`
					Count  int     `bson:"count"`
				}{
					{Id: "sean", MinAge: 20, MaxAge: 25, AvgAge: 22.5, Count: 2},
					{Id: "tom", MinAge: 30, MaxAge: 30, AvgAge: 30, Count: 1},
				}

				// 验证结果
				for i, expected := range expectedResults {
					assert.Equal(t, expected.Id, m[i].Id)
					assert.Equal(t, expected.MinAge, m[i].MinAge)
					assert.Equal(t, expected.MaxAge, m[i].MaxAge)
					assert.Equal(t, expected.AvgAge, m[i].AvgAge)
					assert.Equal(t, expected.Count, m[i].Count)
				}
			},
		},
		{
			name: "测试多级分组和复杂管道操作",
			agg:  aggregator.NewAggregator[TestUser](),
			data: []*TestUser{
				{
					ID:         primitive.ObjectID([12]byte{1, 2, 3, 4, 5}),
					Name:       "sean",
					Age:        20,
					Department: "IT",
				},
				{
					ID:         primitive.ObjectID([12]byte{1, 2, 3, 4, 6}),
					Name:       "tom",
					Age:        25,
					Department: "IT",
				},
				{
					ID:         primitive.ObjectID([12]byte{1, 2, 3, 4, 7}),
					Name:       "jerry",
					Age:        30,
					Department: "HR",
				},
			},
			before: func(tc *tcc) {
				data := tc.data.([]*TestUser)
				for _, user := range data {
					MongoClient.Database("mytest").Collection("db_test").UpdateOne(context.Background(),
						bson.M{"_id": user.ID}, bson.M{"$set": user}, options.Update().SetUpsert(true))
				}
			},
			after: func(tc *tcc) {
				data := tc.data.([]*TestUser)
				for _, user := range data {
					MongoClient.Database("mytest").Collection("db_test").DeleteMany(context.Background(),
						bson.M{"_id": user.ID})
				}
			},
			check: func(tc *tcc) {
				var (
					wantErr error
				)

				tc.agg.Pipe().
					Match(func(m *query.MatchCmd) query.Builder {
						return m.K("age").Gte(20)
					}).
					Group(func(gc *query.GroupCmd) query.Builder {
						return gc.Id(
							gc.IdBuilder().
								SetKeyField("department", "department").
								Key("ageRange").Cond(
								gc.AggC().Lte(gc.ToFd("age"), gc.AnyVal(25)),
								query.V("young"),
								query.V("old"),
							),
						).
							Key("avgAge").Avg(gc.ToFd("age")).
							Key("names").Push(gc.ToFd("name")).Build()
					}).
					Sort("avgAge")

				m := []*struct {
					Id struct {
						Department string `bson:"department"`
						AgeRange   string `bson:"ageRange"`
					} `bson:"_id"`
					AvgAge float64  `bson:"avgAge"`
					Names  []string `bson:"names"`
				}{}
				err := aggregator.WithParseAggregate(md.Sess(context.Background(), ndb, ncoll), tc.agg, &m)
				if err != nil {
					t.Error(err)
					return
				}
				assert.Equal(t, wantErr, err)
				assert.Equal(t, 2, len(m))

				// 验证聚合结果
				expectedResults := []*struct {
					Id struct {
						Department string `bson:"department"`
						AgeRange   string `bson:"ageRange"`
					} `bson:"_id"`
					AvgAge float64  `bson:"avgAge"`
					Names  []string `bson:"names"`
				}{
					{Id: struct {
						Department string `bson:"department"`
						AgeRange   string `bson:"ageRange"`
					}{Department: "IT", AgeRange: "young"}, AvgAge: 22.5, Names: []string{"sean", "tom"}},
					{Id: struct {
						Department string `bson:"department"`
						AgeRange   string `bson:"ageRange"`
					}{Department: "HR", AgeRange: "old"}, AvgAge: 30, Names: []string{"jerry"}},
				}

				// 验证结果
				for i, expected := range expectedResults {
					assert.Equal(t, expected.Id.Department, m[i].Id.Department)
					assert.Equal(t, expected.Id.AgeRange, m[i].Id.AgeRange)
					assert.Equal(t, expected.AvgAge, m[i].AvgAge)
					assert.Equal(t, expected.Names, m[i].Names)
				}
			},
		},
	}

	for _, tc := range testCase {
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
