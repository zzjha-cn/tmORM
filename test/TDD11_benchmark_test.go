package test

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"testing"
	"time"

	tmorm "github.com/zzjha-cn/tm_orm"
	"github.com/zzjha-cn/tm_orm/aggregator"
	"github.com/zzjha-cn/tm_orm/collection"
	"github.com/zzjha-cn/tm_orm/repo"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type BenchUser struct {
	ID        primitive.ObjectID `bson:"_id"`
	Name      string             `bson:"name"`
	Age       int                `bson:"age"`
	Email     string             `bson:"email"`
	Address   string             `bson:"address"`
	City      string             `bson:"city"`
	Country   string             `bson:"country"`
	Salary    float64            `bson:"salary"`
	Tags      []string           `bson:"tags"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
	IsActive  bool               `bson:"is_active"`
}

type BenchOrder struct {
	ID       primitive.ObjectID `bson:"_id"`
	UserID   primitive.ObjectID `bson:"user_id"`
	Amount   float64            `bson:"amount"`
	Status   string             `bson:"status"`
	Products []BenchProduct     `bson:"products"`
	OrderAt  time.Time          `bson:"order_at"`
}

type BenchProduct struct {
	ID       primitive.ObjectID `bson:"_id"`
	Name     string             `bson:"name"`
	Price    float64            `bson:"price"`
	Quantity int                `bson:"quantity"`
}

var (
	benchDB     = "my_test"
	benchColl   = "users"
	orderColl   = "orders"
	productColl = "products"
	ctx         = context.Background()
	cities      = []string{"Beijing", "Shanghai", "Guangzhou", "Shenzhen", "Hangzhou", "Nanjing", "Chengdu", "Wuhan"}
	countries   = []string{"China", "USA", "Japan", "Germany", "France", "UK", "Canada", "Australia"}
	tags        = []string{"developer", "manager", "designer", "analyst", "tester", "architect", "devops", "product"}
	statuses    = []string{"pending", "processing", "completed", "cancelled", "refunded"}
	
	// 全局数据库连接，避免重复初始化
	benchMongoClient *mongo.Client
	benchORMClient   *tmorm.ORMClient
	benchInitOnce    sync.Once
)

// 确保数据库连接只初始化一次
func ensureBenchmarkConnection() {
	benchInitOnce.Do(func() {
		ConnectMongo()
		benchMongoClient = MongoClient
		var err error
		benchORMClient, err = tmorm.NewORMClient(nil)
		if err != nil {
			panic(fmt.Sprintf("Failed to create ORM client: %v", err))
		}
	})
}

// 初始化测试数据
func initBenchmarkData(b *testing.B, count int) (*collection.Collection[BenchUser], *mongo.Collection) {
	ensureBenchmarkConnection()
	client := benchMongoClient

	// 清理已存在的数据
	coll := client.Database(benchDB).Collection(benchColl)
	coll.Drop(ctx)

	// 生成测试数据
	docs := make([]any, count)
	for i := 0; i < count; i++ {
		userTags := make([]string, rand.Intn(3)+1)
		for j := range userTags {
			userTags[j] = tags[rand.Intn(len(tags))]
		}

		docs[i] = BenchUser{
			ID:        primitive.NewObjectID(),
			Name:      randomString(10),
			Age:       rand.Intn(60) + 18, // 18-78岁
			Email:     randomString(8) + "@example.com",
			Address:   randomString(20),
			City:      cities[rand.Intn(len(cities))],
			Country:   countries[rand.Intn(len(countries))],
			Salary:    float64(rand.Intn(200000) + 30000), // 30k-230k
			Tags:      userTags,
			CreatedAt: time.Now().Add(-time.Duration(rand.Intn(365*24)) * time.Hour),
			UpdatedAt: time.Now().Add(-time.Duration(rand.Intn(30*24)) * time.Hour),
			IsActive:  rand.Float32() > 0.1, // 90%活跃用户
		}
	}

	// 批量插入数据
	_, err := coll.InsertMany(ctx, docs)
	if err != nil {
		b.Fatal(err)
	}

	// 创建索引以提高查询性能
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{"age", 1}}},
		{Keys: bson.D{{"email", 1}}},
		{Keys: bson.D{{"city", 1}}},
		{Keys: bson.D{{"salary", 1}}},
		{Keys: bson.D{{"created_at", 1}}},
		{Keys: bson.D{{"is_active", 1}}},
	}
	coll.Indexes().CreateMany(ctx, indexes)

	// 创建ORM collection
	ormCollection := collection.NewCollection[BenchUser](benchORMClient, benchDB, benchColl)

	return ormCollection, coll
}

// 初始化订单测试数据
func initOrderBenchmarkData(b *testing.B, userCount, orderCount int) (*collection.Collection[BenchOrder], *mongo.Collection) {
	ensureBenchmarkConnection()
	client := benchMongoClient

	// 清理已存在的数据
	coll := client.Database(benchDB).Collection(orderColl)
	coll.Drop(ctx)

	// 先创建用户数据
	userColl := client.Database(benchDB).Collection(benchColl)
	userColl.Drop(ctx)

	// 生成用户ID列表
	userIDs := make([]primitive.ObjectID, userCount)
	userDocs := make([]any, userCount)
	for i := 0; i < userCount; i++ {
		userIDs[i] = primitive.NewObjectID()
		userDocs[i] = BenchUser{
			ID:        userIDs[i],
			Name:      randomString(10),
			Age:       rand.Intn(60) + 18,
			Email:     randomString(8) + "@example.com",
			Address:   randomString(20),
			City:      cities[rand.Intn(len(cities))],
			Country:   countries[rand.Intn(len(countries))],
			Salary:    float64(rand.Intn(200000) + 30000),
			Tags:      []string{tags[rand.Intn(len(tags))]},
			CreatedAt: time.Now().Add(-time.Duration(rand.Intn(365*24)) * time.Hour),
			UpdatedAt: time.Now(),
			IsActive:  true,
		}
	}
	userColl.InsertMany(ctx, userDocs)

	// 生成订单数据
	orderDocs := make([]any, orderCount)
	for i := 0; i < orderCount; i++ {
		productCount := rand.Intn(5) + 1
		products := make([]BenchProduct, productCount)
		totalAmount := 0.0

		for j := 0; j < productCount; j++ {
			price := float64(rand.Intn(1000) + 10)
			quantity := rand.Intn(5) + 1
			products[j] = BenchProduct{
				ID:       primitive.NewObjectID(),
				Name:     fmt.Sprintf("Product_%s", randomString(5)),
				Price:    price,
				Quantity: quantity,
			}
			totalAmount += price * float64(quantity)
		}

		orderDocs[i] = BenchOrder{
			ID:       primitive.NewObjectID(),
			UserID:   userIDs[rand.Intn(len(userIDs))],
			Amount:   totalAmount,
			Status:   statuses[rand.Intn(len(statuses))],
			Products: products,
			OrderAt:  time.Now().Add(-time.Duration(rand.Intn(90*24)) * time.Hour),
		}
	}

	// 批量插入订单数据
	_, err := coll.InsertMany(ctx, orderDocs)
	if err != nil {
		b.Fatal(err)
	}

	// 创建索引
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{"user_id", 1}}},
		{Keys: bson.D{{"status", 1}}},
		{Keys: bson.D{{"amount", 1}}},
		{Keys: bson.D{{"order_at", 1}}},
	}
	coll.Indexes().CreateMany(ctx, indexes)

	// 创建ORM collection
	ormCollection := collection.NewCollection[BenchOrder](benchORMClient, benchDB, orderColl)

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

// 插入操作性能测试
func BenchmarkInsert(b *testing.B) {
	ensureBenchmarkConnection()
	client := benchMongoClient

	// 清理数据
	coll := client.Database(benchDB).Collection(benchColl)
	coll.Drop(ctx)

	ormCollection := collection.NewCollection[BenchUser](benchORMClient, benchDB, benchColl)

	b.Run("ORM-InsertOne", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			user := &BenchUser{
				ID:        primitive.NewObjectID(),
				Name:      randomString(10),
				Age:       rand.Intn(60) + 18,
				Email:     randomString(8) + "@example.com",
				Address:   randomString(20),
				City:      cities[rand.Intn(len(cities))],
				Country:   countries[rand.Intn(len(countries))],
				Salary:    float64(rand.Intn(200000) + 30000),
				Tags:      []string{tags[rand.Intn(len(tags))]},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				IsActive:  true,
			}
			_, err := ormCollection.Insert(ctx, user)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Native-InsertOne", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			user := BenchUser{
				ID:        primitive.NewObjectID(),
				Name:      randomString(10),
				Age:       rand.Intn(60) + 18,
				Email:     randomString(8) + "@example.com",
				Address:   randomString(20),
				City:      cities[rand.Intn(len(cities))],
				Country:   countries[rand.Intn(len(countries))],
				Salary:    float64(rand.Intn(200000) + 30000),
				Tags:      []string{tags[rand.Intn(len(tags))]},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				IsActive:  true,
			}
			_, err := coll.InsertOne(ctx, user)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("ORM-InsertMany", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			users := make([]*BenchUser, 100)
			for j := 0; j < 100; j++ {
				users[j] = &BenchUser{
					ID:        primitive.NewObjectID(),
					Name:      randomString(10),
					Age:       rand.Intn(60) + 18,
					Email:     randomString(8) + "@example.com",
					Address:   randomString(20),
					City:      cities[rand.Intn(len(cities))],
					Country:   countries[rand.Intn(len(countries))],
					Salary:    float64(rand.Intn(200000) + 30000),
					Tags:      []string{tags[rand.Intn(len(tags))]},
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
					IsActive:  true,
				}
			}
			_, err := ormCollection.InsertMany(ctx, users)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// 关联查询性能测试
func BenchmarkJoinQuery(b *testing.B) {
	_, nativeOrderColl := initOrderBenchmarkData(b, 1000, 10000)
	ensureBenchmarkConnection()

	b.Run("ORM-LookupJoin", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			agg := aggregator.NewAggregator[any](benchORMClient, benchDB, orderColl)
			_, err := agg.
				Match(bson.M{"status": "completed"}).
				Lookup(benchColl, "user_id", "_id", "user_info").
				Unwind("$user_info", true).
				Project(bson.M{
					"_id":        1,
					"amount":     1,
					"user_name":  "$user_info.name",
					"user_email": "$user_info.email",
				}).
				Execute(ctx)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Native-LookupJoin", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			pipeline := mongo.Pipeline{
				{{"$match", bson.D{{"status", "completed"}}}},
				{{"$lookup", bson.D{
					{"from", benchColl},
					{"localField", "user_id"},
					{"foreignField", "_id"},
					{"as", "user_info"},
				}}},
				{{"$unwind", bson.D{
					{"path", "$user_info"},
					{"preserveNullAndEmptyArrays", true},
				}}},
				{{"$project", bson.D{
					{"_id", 1},
					{"amount", 1},
					{"user_name", "$user_info.name"},
					{"user_email", "$user_info.email"},
				}}},
			}
			cursor, err := nativeOrderColl.Aggregate(ctx, pipeline)
			if err != nil {
				b.Fatal(err)
			}
			var results []bson.M
			if err = cursor.All(ctx, &results); err != nil {
				b.Fatal(err)
			}
		}
	})
}

// Repository模式性能测试
func BenchmarkRepository(b *testing.B) {
	ensureBenchmarkConnection()
	repo := repo.NewRepository[BenchUser](benchORMClient, benchDB, benchColl)

	// 初始化数据
	client := benchMongoClient
	nativeColl := client.Database(benchDB).Collection(benchColl)
	nativeColl.Drop(ctx)

	// 插入测试数据
	users := make([]*BenchUser, 10000)
	for i := 0; i < 10000; i++ {
		users[i] = &BenchUser{
			ID:        primitive.NewObjectID(),
			Name:      randomString(10),
			Age:       rand.Intn(60) + 18,
			Email:     randomString(8) + "@example.com",
			Address:   randomString(20),
			City:      cities[rand.Intn(len(cities))],
			Country:   countries[rand.Intn(len(countries))],
			Salary:    float64(rand.Intn(200000) + 30000),
			Tags:      []string{tags[rand.Intn(len(tags))]},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			IsActive:  true,
		}
	}
	repo.CreateMany(ctx, users)

	b.Run("Repository-FindByID", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			userID := users[rand.Intn(len(users))].ID
			_, err := repo.FindByID(ctx, userID)
			if err != nil && err != mongo.ErrNoDocuments {
				b.Fatal(err)
			}
		}
	})

	b.Run("Repository-FindPage", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			page := int64(rand.Intn(50) + 1)
			pageSize := int64(20)
			_, _, err := repo.FindPage(ctx, page, pageSize)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Repository-UpdateByID", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			userID := users[rand.Intn(len(users))].ID
			updates := map[string]any{
				"salary":     float64(rand.Intn(200000) + 30000),
				"updated_at": time.Now(),
			}
			_, err := repo.UpdateByID(ctx, userID, updates)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// 并发性能测试
func BenchmarkConcurrent(b *testing.B) {
	ormColl, _ := initBenchmarkData(b, 50000)

	b.Run("ORM-ConcurrentRead", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				age := rand.Intn(60) + 18
				_, err := ormColl.Where("age").Gte(age).Find(ctx)
				if err != nil {
					b.Error(err)
				}
			}
		})
	})

	b.Run("ORM-ConcurrentWrite", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				user := &BenchUser{
					ID:        primitive.NewObjectID(),
					Name:      randomString(10),
					Age:       rand.Intn(60) + 18,
					Email:     randomString(8) + "@example.com",
					Address:   randomString(20),
					City:      cities[rand.Intn(len(cities))],
					Country:   countries[rand.Intn(len(countries))],
					Salary:    float64(rand.Intn(200000) + 30000),
					Tags:      []string{tags[rand.Intn(len(tags))]},
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
					IsActive:  true,
				}
				_, err := ormColl.Insert(ctx, user)
				if err != nil {
					b.Error(err)
				}
			}
		})
	})

	b.Run("ORM-ConcurrentUpdate", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				age := rand.Intn(60) + 18
				newSalary := float64(rand.Intn(200000) + 30000)
				_, err := ormColl.Where("age").Eq(age).Set("salary", newSalary).Update(ctx)
				if err != nil {
					b.Error(err)
				}
			}
		})
	})
}

// 内存使用和GC压力测试
func BenchmarkMemoryUsage(b *testing.B) {
	ormColl, _ := initBenchmarkData(b, 100000)

	b.Run("ORM-LargeResultSet", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// 强制GC以获得更准确的内存分配统计
			runtime.GC()
			_, err := ormColl.Where("is_active").Eq(true).Find(ctx)
			if err != nil {
				b.Fatal(err)
			}
			// 再次强制GC
			runtime.GC()
		}
	})

	b.Run("ORM-StreamingQuery", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// 分批查询以减少内存压力
			pageSize := int64(1000)
			for page := int64(1); page <= 10; page++ {
				skip := (page - 1) * pageSize
				opts := options.Find().SetSkip(skip).SetLimit(pageSize)
				_, err := ormColl.Where("is_active").Eq(true).Find(ctx, opts)
				if err != nil {
					b.Fatal(err)
				}
			}
		}
	})
}

// 复杂查询条件性能测试
func BenchmarkComplexConditions(b *testing.B) {
	ormColl, nativeColl := initBenchmarkData(b, 100000)

	b.Run("ORM-MultipleConditions", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := ormColl.
				Where("age").Gte(25).Where("age").Lte(65).
				Where("salary").Gte(50000).
				Where("is_active").Eq(true).
				Where("city").In("Beijing", "Shanghai", "Guangzhou").
				Find(ctx)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Native-MultipleConditions", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			filter := bson.D{{
				"$and", bson.A{
					bson.D{{"age", bson.D{{"$gte", 25}}}},
					bson.D{{"age", bson.D{{"$lte", 65}}}},
					bson.D{{"salary", bson.D{{"$gte", 50000}}}},
					bson.D{{"is_active", true}},
					bson.D{{"city", bson.D{{"$in", bson.A{"Beijing", "Shanghai", "Guangzhou"}}}}},
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
	})

	b.Run("ORM-RegexQuery", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := ormColl.
				Where("email").Regex("@example.com$").
				Where("name").Regex("^[A-Z]").
				Find(ctx)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// 数据类型和索引性能测试
func BenchmarkIndexPerformance(b *testing.B) {
	ormColl, nativeColl := initBenchmarkData(b, 200000)

	b.Run("ORM-IndexedQuery", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// 使用已建索引的字段查询
			age := rand.Intn(60) + 18
			_, err := ormColl.Where("age").Eq(age).Find(ctx)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("ORM-NonIndexedQuery", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// 使用未建索引的字段查询
			address := randomString(5)
			_, err := ormColl.Where("address").Regex(address).Find(ctx)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("ORM-CompoundIndexQuery", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		// 创建复合索引
		indexModel := mongo.IndexModel{
			Keys: bson.D{{"city", 1}, {"age", 1}, {"salary", 1}},
		}
		nativeColl.Indexes().CreateOne(ctx, indexModel)

		for i := 0; i < b.N; i++ {
			city := cities[rand.Intn(len(cities))]
			age := rand.Intn(60) + 18
			salary := float64(rand.Intn(200000) + 30000)
			_, err := ormColl.
				Where("city").Eq(city).
				Where("age").Gte(age).
				Where("salary").Gte(salary).
				Find(ctx)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// 删除操作性能测试
func BenchmarkDelete(b *testing.B) {
	ormColl, nativeColl := initBenchmarkData(b, 50000)

	b.Run("ORM-DeleteMany", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := ormColl.Where("age").Lt(25).Delete(ctx)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Native-DeleteMany", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			filter := bson.D{{"age", bson.D{{"$lt", 25}}}}
			_, err := nativeColl.DeleteMany(ctx, filter)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("ORM-DeleteOne", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := ormColl.Where("age").Eq(30).DeleteOne(ctx)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// 分页查询性能测试
func BenchmarkPagination(b *testing.B) {
	ormColl, nativeColl := initBenchmarkData(b, 100000)

	b.Run("ORM-Pagination", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			page := int64(rand.Intn(100) + 1)
			pageSize := int64(20)
			skip := (page - 1) * pageSize

			opts := options.Find().SetSkip(skip).SetLimit(pageSize).SetSort(bson.D{{"created_at", -1}})
			_, err := ormColl.Where("is_active").Eq(true).Find(ctx, opts)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Native-Pagination", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			page := int64(rand.Intn(100) + 1)
			pageSize := int64(20)
			skip := (page - 1) * pageSize

			filter := bson.D{{"is_active", true}}
			opts := options.Find().SetSkip(skip).SetLimit(pageSize).SetSort(bson.D{{"created_at", -1}})
			cursor, err := nativeColl.Find(ctx, filter, opts)
			if err != nil {
				b.Fatal(err)
			}
			var results []BenchUser
			if err = cursor.All(ctx, &results); err != nil {
				b.Fatal(err)
			}
		}
	})
}

// 统计查询性能测试
func BenchmarkCount(b *testing.B) {
	ormColl, nativeColl := initBenchmarkData(b, 100000)

	b.Run("ORM-Count", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := ormColl.Where("age").Gte(25).Where("salary").Gte(50000).Count(ctx)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Native-Count", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			filter := bson.D{{
				"$and", bson.A{
					bson.D{{"age", bson.D{{"$gte", 25}}}},
					bson.D{{"salary", bson.D{{"$gte", 50000}}}},
				},
			}}
			_, err := nativeColl.CountDocuments(ctx, filter)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// 聚合操作性能测试
func BenchmarkAggregate(b *testing.B) {
	_, nativeColl := initBenchmarkData(b, 50000)
	ormClient, _ := tmorm.NewORMClient(nil)

	b.Run("ORM-SimpleAggregate", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			agg := aggregator.NewAggregator[BenchUser](ormClient, benchDB, benchColl)
			_, err := agg.
				Match(bson.M{"age": bson.M{"$gte": 18}}).
				Group("$age", bson.M{
					"count":      bson.M{"$sum": 1},
					"avg_salary": bson.M{"$avg": "$salary"},
				}).
				Execute(ctx)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Native-SimpleAggregate", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			pipeline := mongo.Pipeline{
				{{"$match", bson.D{{"age", bson.D{{"$gte", 18}}}}}},
				{{"$group", bson.D{
					{"_id", "$age"},
					{"count", bson.D{{"$sum", 1}}},
					{"avg_salary", bson.D{{"$avg", "$salary"}}},
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
	})

	b.Run("ORM-ComplexAggregate", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			agg := aggregator.NewAggregator[any](ormClient, benchDB, benchColl)
			_, err := agg.
				Match(bson.M{"is_active": true}).
				Group("$city", bson.M{
					"total_users": bson.M{"$sum": 1},
					"avg_age":     bson.M{"$avg": "$age"},
					"avg_salary":  bson.M{"$avg": "$salary"},
					"max_salary":  bson.M{"$max": "$salary"},
					"min_salary":  bson.M{"$min": "$salary"},
				}).
				Sort(bson.D{{"avg_salary", -1}}).
				Limit(10).
				Execute(ctx)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Native-ComplexAggregate", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			pipeline := mongo.Pipeline{
				{{"$match", bson.D{{"is_active", true}}}},
				{{"$group", bson.D{
					{"_id", "$city"},
					{"total_users", bson.D{{"$sum", 1}}},
					{"avg_age", bson.D{{"$avg", "$age"}}},
					{"avg_salary", bson.D{{"$avg", "$salary"}}},
					{"max_salary", bson.D{{"$max", "$salary"}}},
					{"min_salary", bson.D{{"$min", "$salary"}}},
				}}},
				{{"$sort", bson.D{{"avg_salary", -1}}}},
				{{"$limit", 10}},
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
	})
}
