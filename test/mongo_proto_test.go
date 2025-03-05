package test

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"testing"
	"tm_orm/finder"
)

// 端到端测试finder

var MongoClient *mongo.Client

func TestMongoPro(t *testing.T) {
	ConnectMongo()
	fd := &finder.Finder[TestUser]{}
	tq := &Tquery{}

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
			name:   "测试mongo协议_and",
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
				tq.E = bson.E{
					Key: "$and", Value: bson.A{
						//bson.E{"name", bson.E{"$eq", "sean"}},
						bson.E{"age", bson.E{"$gt", 15}},
					},
				}

				d := tq.GetBsonD()
				fmt.Println(d)
				l, err := MongoClient.Database("mytest").Collection("db_test").Find(
					context.Background(),
					//bson.M{"$and": bson.A{bson.D{{"age", bson.D{{"$gt", 10}}}}}},
					//bson.M{"$and": bson.A{bson.M{"age": bson.E{"$gt", 10}}}}, // 查无数据
					//bson.E{"$and", bson.A{bson.M{"age": bson.D{{"$gt", 10}}}}}, // 查无数据
					//bson.D{{"$and", bson.A{bson.E{"age", bson.D{{"$gt", 10}}}}}}, // 查无数据
					//bson.D{{"$and", bson.A{bson.D{{"age", bson.E{"$gt", 10}}}}}},  // 查无数据
					bson.D{{"$and", bson.A{
						bson.D{{"name", bson.D{{"$eq", "sean"}}}},
						bson.D{{"age", bson.D{{"$gt", 10}}}},
					}}}, // 有数据
					//bson.D{{"$and", bson.A{bson.D{{"age", bson.D{{"$gt", 10}}}}}}}, // 有数据
					//bson.D{{"$and", bson.A{bson.M{"age": bson.D{{"$gt", 10}}}}}}, // 有数据
					//bson.M{"$and": bson.A{bson.M{"age": bson.D{{"$gt", 10}}}}}, // 有数据
					//bson.M{"$and": bson.A{bson.M{"age": bson.M{"$gt": 10}}}}, // 有数据
					// 因为driver以bsonD为最小单位，而不是bsonE
				)
				if err != nil {
					panic(err)
				}

				var u []*TestUser
				err = l.All(context.Background(), &u)
				if err != nil {
					panic(err)
				}
			},
		},
		{
			name:   "测试mongo协议_and_2",
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
				l, err := MongoClient.Database("mytest").Collection("db_test").Find(
					context.Background(),
					bson.D{{"$and", bson.A{
						bson.D{{"$gte", bson.A{"$age", 10}}},
					}}}, // 有数据
				)
				if err != nil {
					panic(err)
				}

				var u []*TestUser
				err = l.All(context.Background(), &u)
				if err != nil {
					panic(err)
				}

				println(u)
			},
		},
		{
			name:   "测试mongo协议_and与or嵌套",
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
				l, err := MongoClient.Database("mytest").Collection("db_test").Find(
					context.Background(),
					bson.D{
						{Key: "$or", Value: bson.A{
							bson.D{{Key: "name", Value: bson.D{{"$eq", "x"}}}},
							bson.D{{Key: "age", Value: bson.D{{"$gte", 10}}}},
							bson.D{{Key: "$and", Value: bson.A{
								bson.D{{Key: "name", Value: bson.D{{"$eq", "sean"}}}},
								bson.D{{Key: "age", Value: bson.D{{"$eq", 21}}}},
							}}},
						}},
					},
				)
				if err != nil {
					panic(err)
				}

				var u []*TestUser
				err = l.All(context.Background(), &u)
				if err != nil {
					panic(err)
				}

				println(u)
			},
		},
		{
			name:   "测试mongo协议_expr",
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
				l, err := MongoClient.Database("mytest").Collection("db_test").Find(
					context.Background(),
					bson.D{{Key: "$expr", Value: bson.D{{
						Key:   "$in",
						Value: bson.A{"$name", []any{"sean", "jean", "mike"}}}},
					}},
				)
				if err != nil {
					panic(err)
				}

				var u []*TestUser
				err = l.All(context.Background(), &u)
				if err != nil {
					panic(err)
				}

				println(u)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.before != nil {
				tc.before(&tc)
			}
			tc.check(&tc)
			//if tc.after != nil {
			//	tc.after(&tc)
			//}
		})
	}

}

func ConnectMongo() {
	conf := &mongoCfg{
		Username:   "test",
		Password:   "1234",
		AuthSource: "mytest",

		Timeout:        10,
		HostsWithPorts: []string{"127.0.0.1:27017"},
		MaxPool:        20,
		MinPool:        10,
		ReplicaSet:     "",
	}

	credential := options.Credential{
		Username:      conf.Username,
		Password:      conf.Password,
		AuthMechanism: conf.AuthMechanism,
		AuthSource:    conf.AuthSource,
	}

	clientOps := options.Client().SetAuth(credential).
		// SetConnectTimeout(time.Duration(conf.Timeout * 1000)).
		SetHosts(conf.HostsWithPorts).
		SetMaxPoolSize(conf.MaxPool).
		SetMinPoolSize(conf.MinPool)
	// SetReadPreference()

	var ctx = context.TODO()
	client, err := mongo.Connect(ctx, clientOps)
	if nil != err {
		panic(err)
	}

	err = client.Ping(ctx, nil)
	if nil != err {
		fmt.Println(err)
		return
	}
	MongoClient = client
}
