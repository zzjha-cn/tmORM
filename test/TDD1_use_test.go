package test

import (
	"context"
	"fmt"
	tmorm "github.com/zzjha-cn/tm_orm"
	"github.com/zzjha-cn/tm_orm/collection"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	dbName = "mytest"
	coll   = "db_test"
)

var client, _ = tmorm.NewORMClient(nil)

func TestUserType(t *testing.T) {
	return
	ctx := context.Background()
	// 使用新的collection API
	collection := collection.NewCollection[TestUser](client, dbName, coll)

	// 基本查询
	users, err := collection.Find(ctx)
	if err != nil {
		t.Errorf("Find error: %v", err)
	}
	fmt.Printf("Found %d users\n", len(users))

	// 条件查询
	adults, err := collection.Where("age").Gte(18).Find(ctx)
	if err != nil {
		t.Errorf("Where Find error: %v", err)
	}
	fmt.Printf("Found %d adults\n", len(adults))

	// 查找单个文档
	user, err := collection.Where("name").Eq("John").FindOne(ctx)
	if err != nil {
		t.Logf("FindOne error: %v", err)
	} else {
		fmt.Printf("Found user: %+v\n", user)
	}

	// 插入文档
	newUser := &TestUser{
		ID:         primitive.NewObjectID(),
		Name:       "Alice",
		Age:        25,
		Department: "Engineering",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	result, err := collection.Insert(ctx, newUser)
	if err != nil {
		t.Errorf("Insert error: %v", err)
	} else {
		fmt.Printf("Inserted user with ID: %v\n", result.InsertedID)
	}

	// 更新文档
	updateResult, err := collection.Where("name").Eq("Alice").Set("age", 26).Update(ctx)
	if err != nil {
		t.Errorf("Update error: %v", err)
	} else {
		fmt.Printf("Updated %d documents\n", updateResult.ModifiedCount)
	}

	// 删除文档
	deleteResult, err := collection.Where("name").Eq("Alice").Delete(ctx)
	if err != nil {
		t.Errorf("Delete error: %v", err)
	} else {
		fmt.Printf("Deleted %d documents\n", deleteResult.DeletedCount)
	}

	// 统计文档数量
	count, err := collection.Count(ctx)
	if err != nil {
		t.Errorf("Count error: %v", err)
	} else {
		fmt.Printf("Total documents: %d\n", count)
	}

	// 复杂查询示例
	seniorEngineers, err := collection.
		Where("age").Gte(30).
		Where("department").Eq("Engineering").
		Find(ctx)
	if err != nil {
		t.Errorf("Complex query error: %v", err)
	} else {
		fmt.Printf("Found %d senior engineers\n", len(seniorEngineers))
	}
}

type (
	TestUser struct {
		ID           primitive.ObjectID `bson:"_id,omitempty"`
		Name         string             `bson:"name"`
		Age          int64              `bson:"age"`
		Department   string             `bson:"department"`
		UnknownField string             `bson:"-"`
		CreatedAt    time.Time          `bson:"created_at"`
		UpdatedAt    time.Time          `bson:"updated_at"`
	}

	Tquery struct {
		E bson.E
	}

	mongoCfg struct {
		Username      string
		Password      string
		AuthMechanism string
		AuthSource    string

		Timeout        int
		HostsWithPorts []string
		MaxPool        uint64
		MinPool        uint64
		ReplicaSet     string
	}
)

func (t *Tquery) GetBsonD() bson.D {
	println(t.E.Value)
	a := bson.D{}
	a = append(a, t.E)
	return a
}
