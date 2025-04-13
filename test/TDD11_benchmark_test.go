package test

import (
	"context"
	tmorm "github.com/zzjha-cn/tm_orm"
	"github.com/zzjha-cn/tm_orm/aggregator"
	"github.com/zzjha-cn/tm_orm/finder"
	"github.com/zzjha-cn/tm_orm/query"
	"github.com/zzjha-cn/tm_orm/updater"
	"math/rand"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type BenchUser struct {
	ID      primitive.ObjectID `bson:"_id"`
	Name    string             `bson:"name"`
	Age     int                `bson:"age"`
	Email   string             `bson:"email"`
	Address string             `bson:"address"`
}

var (
	benchDB   = "my_test"
	benchColl = "users"
	ctx       = context.Background()
)

// 初始化测试数据
func initBenchmarkData(b *testing.B, count int) (*tmorm.MDB, *mongo.Collection) {
	ConnectMongo()
	client := MongoClient

	// 清理已存在的数据
	coll := client.Database(benchDB).Collection(benchColl)
	coll.Drop(ctx)

	// 生成测试数据
	docs := make([]interface{}, count)
	for i := 0; i < count; i++ {
		docs[i] = BenchUser{
			ID:      primitive.NewObjectID(),
			Name:    randomString(10),
			Age:     rand.Intn(80),
			Email:   randomString(8) + "@example.com",
			Address: randomString(20),
		}
	}

	// 批量插入数据
	_, err := coll.InsertMany(ctx, docs)
	if err != nil {
		b.Fatal(err)
	}

	return tmorm.NewMDB(client), coll
}

func randomString(n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// 基本查询性能测试
func BenchmarkBasicQuery(b *testing.B) {
	db, coll := initBenchmarkData(b, 10000)
	fd := &finder.Finder[BenchUser]{}

	b.Run("ORM-SimpleQuery", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			filter := query.Query{}.Builder().
				K("age").Gte(18).
				ToQuery()
			_, err := fd.Find(db.Sess(ctx, benchDB, benchColl), filter)
			if err != nil {
				b.Fatal(err)
			}
		}
		b.StopTimer()

	})

	b.Run("Native-SimpleQuery", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			filter := bson.D{{"age", bson.D{{"$gte", 18}}}}
			cursor, err := coll.Find(ctx, filter)
			if err != nil {
				b.Fatal(err)
			}
			var results []BenchUser
			if err = cursor.All(ctx, &results); err != nil {
				b.Fatal(err)
			}
		}
		b.StopTimer()

	})
}

// 复杂查询性能测试
func BenchmarkComplexQuery(b *testing.B) {
	db, coll := initBenchmarkData(b, 10000)
	fd := &finder.Finder[BenchUser]{}

	b.Run("ORM-ComplexQuery", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			filter := query.Query{}.Builder().
				Or(func(a *query.QueryOr) query.Builder {
					return a.
						K("age").Gt(30).
						K("name").Regex("^A").
						And(func(and *query.QueryAnd) query.Builder {
							return and.
								K("age").Lt(50).
								K("email").Regex(".+@example.com$")
						})
				}).
				ToQuery()
			_, err := fd.Find(db.Sess(ctx, benchDB, benchColl), filter)
			if err != nil {
				b.Fatal(err)
			}
		}
		b.StopTimer()

	})

	b.Run("Native-ComplexQuery", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			filter := bson.D{{
				"$or", bson.A{
					bson.D{{"age", bson.D{{"$gt", 30}}}},
					bson.D{{"name", bson.D{{"$regex", "^A"}}}},
					bson.D{{
						"$and", bson.A{
							bson.D{{"age", bson.D{{"$lt", 50}}}},
							bson.D{{"email", bson.D{{"$regex", ".+@example.com$"}}}},
						},
					}},
				},
			}}
			cursor, err := coll.Find(ctx, filter)
			if err != nil {
				b.Fatal(err)
			}
			var results []BenchUser
			if err = cursor.All(ctx, &results); err != nil {
				b.Fatal(err)
			}
		}
		b.StopTimer()

	})
}

// 更新操作性能测试
func BenchmarkUpdate(b *testing.B) {
	db, coll := initBenchmarkData(b, 10000)
	up := &updater.MUpdater[BenchUser]{}

	b.Run("ORM-Update", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			filter := query.Query{}.Builder().
				K("age").Lt(30).
				ToQuery()
			bd := updater.NewReplaceBuilder[BenchUser]()
			bd.C().Set("address", "Updated Address")
			_, err := up.SetFilter(filter).UpdateMany(
				db.Sess(ctx, benchDB, benchColl),
				bd,
			)
			if err != nil {
				b.Fatal(err)
			}
		}
		b.StopTimer()

	})

	b.Run("Native-Update", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			filter := bson.D{{"age", bson.D{{"$lt", 30}}}}
			update := bson.D{{"$set", bson.D{{"address", "Updated Address"}}}}
			_, err := coll.UpdateMany(ctx, filter, update)
			if err != nil {
				b.Fatal(err)
			}
		}
		b.StopTimer()

	})
}

// 聚合操作性能测试
func BenchmarkAggregate(b *testing.B) {
	db, coll := initBenchmarkData(b, 10000)
	agg := aggregator.NewAggregator[BenchUser]()

	b.Run("ORM-Aggregate", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			p := aggregator.NewPipeline().
				Match(func(m *query.MatchCmd) query.Builder {
					return m.K("age").Gte(18)
				}).
				Group(func(g *query.GroupCmd) query.Builder {
					return g.IdWithField("age").
						Key("count1").Sum(g.AnyVal(1)).
						Key("avgAge").Avg(g.ToFd("age")).
						Build()
				})
			agg.SetPipe(p)

			type Result struct {
				ID     int     `bson:"_id"`
				Count  int     `bson:"count"`
				AvgAge float64 `bson:"avgAge"`
			}
			var results []*Result
			err := aggregator.WithParseAggregate(
				db.Sess(ctx, benchDB, benchColl),
				agg,
				&results,
			)
			if err != nil {
				b.Fatal(err)
			}
		}
		b.StopTimer()

	})

	b.Run("Native-Aggregate", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			pipeline := mongo.Pipeline{
				{{
					"$match", bson.D{
						{"age", bson.D{{"$gte", 18}}},
					},
				}},
				{{
					"$group", bson.D{
						{"_id", "$age"},
						{"count1", bson.D{{"$sum", 1}}},
						{"avgAge", bson.D{{"$avg", "$age"}}},
					},
				}},
			}

			cursor, err := coll.Aggregate(ctx, pipeline)
			if err != nil {
				b.Fatal(err)
			}
			var results []bson.M
			if err = cursor.All(ctx, &results); err != nil {
				b.Fatal(err)
			}
		}
		b.StopTimer()

	})
}
