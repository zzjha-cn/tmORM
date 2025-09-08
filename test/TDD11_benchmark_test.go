package test

import (
	"context"
	tmorm "github.com/zzjha-cn/tm_orm"
	"github.com/zzjha-cn/tm_orm/aggregator"
	"github.com/zzjha-cn/tm_orm/collection"
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
func initBenchmarkData(b *testing.B, count int) (*collection.Collection[BenchUser], *mongo.Collection) {
	ConnectMongo()
	client := MongoClient

	// 清理已存在的数据
	coll := client.Database(benchDB).Collection(benchColl)
	coll.Drop(ctx)

	// 生成测试数据
	docs := make([]any, count)
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

	// 创建ORM client和collection
	ormClient, _ := tmorm.NewORMClient(nil)
	ormCollection := collection.NewCollection[BenchUser](ormClient, benchDB, benchColl)

	return ormCollection, coll
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
	ormColl, nativeColl := initBenchmarkData(b, 10000)

	b.Run("ORM-SimpleQuery", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, err := ormColl.Where("age").Gte(18).Find(ctx)
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
			cursor, err := nativeColl.Find(ctx, filter)
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
	ormColl, nativeColl := initBenchmarkData(b, 10000)

	b.Run("ORM-ComplexQuery", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, err := ormColl.
				Where("age").Gt(30).
				Where("email").Regex("@example.com").
				Find(ctx)
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
				"$and", bson.A{
					bson.D{{"age", bson.D{{"$gt", 30}}}},
					bson.D{{"email", bson.D{{"$regex", "@example.com"}}}},
				},
			}}
			cursor, err := nativeColl.Find(ctx, filter)
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
	ormColl, nativeColl := initBenchmarkData(b, 10000)

	b.Run("ORM-Update", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, err := ormColl.
				Where("age").Lt(30).
				Set("address", "Updated Address").
				Update(ctx)
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
			_, err := nativeColl.UpdateMany(ctx, filter, update)
			if err != nil {
				b.Fatal(err)
			}
		}
		b.StopTimer()
	})
}

// 聚合操作性能测试
func BenchmarkAggregate(b *testing.B) {
	_, nativeColl := initBenchmarkData(b, 10000)
	ormClient, _ := tmorm.NewORMClient(nil)

	b.Run("ORM-Aggregate", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			agg := aggregator.NewAggregator[BenchUser](ormClient, benchDB, benchColl)
			_, err := agg.
				Match(bson.M{"age": bson.M{"$gte": 18}}).
				Group("$age", bson.M{
					"count": bson.M{"$sum": 1},
				}).
				Execute(ctx)
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
				{{"$match", bson.D{{"age", bson.D{{"$gte", 18}}}}}},
				{{"$group", bson.D{
					{"_id", "$age"},
					{"count", bson.D{{"$sum", 1}}},
				}}},
			}
			cursor, err := nativeColl.Aggregate(ctx, pipeline)
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
