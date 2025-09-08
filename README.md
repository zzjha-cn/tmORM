# tmORM

tmORM是一个轻量级的MongoDB ORM框架，提供了简洁而强大的API来操作MongoDB数据库。支持链式调用、类型安全的查询构建器、更新操作和聚合表达式等特性。

## 特性

- 类型安全的查询构建器
- 支持复杂的查询条件（AND/OR/表达式）
- 支持文档更新和替换操作
- 支持数组操作（Push/Pull/AddToSet）
- 支持聚合管道操作
- 内置中间件机制，支持自定义中间件
- 完整的查询表达式支持


## 进度
- [x] 完成query builder
- [x] TDD测试驱动
- [x] 完成query expr, and , or
- [x] 完成query aggregate operator 聚合操作符
- [x] 完成Middleware
- [x] 简单完成Aggregater
- [x] 完成端到端测试
- [x] 查询的错误处理增强，支持错误分类与错误回调
- [ ] 完成Agg stage阶段的各个常用命令封装（https://www.mongodb.com/zh-cn/docs/manual/reference/operator/aggregation/sort/）
- [ ] 增加元数据管理中心
- [ ] 增加原子化操作
- [ ] 增加结果集处理封装

## 性能测试
```
goos: darwin
goarch: arm64
pkg: github.com/zzjha-cn/tm_orm/test
BenchmarkBasicQuery
BenchmarkBasicQuery/ORM-SimpleQuery
BenchmarkBasicQuery/ORM-SimpleQuery-8         	      12	  91372882 ns/op	14673922 B/op	  390403 allocs/op
BenchmarkBasicQuery/Native-SimpleQuery
BenchmarkBasicQuery/Native-SimpleQuery-8      	      13	  89366000 ns/op	21671820 B/op	  380370 allocs/op
BenchmarkComplexQuery
BenchmarkComplexQuery/ORM-ComplexQuery
BenchmarkComplexQuery/ORM-ComplexQuery-8      	      12	  94856938 ns/op	14675226 B/op	  390509 allocs/op
BenchmarkComplexQuery/Native-ComplexQuery
BenchmarkComplexQuery/Native-ComplexQuery-8   	      14	  76689161 ns/op	16657950 B/op	  296384 allocs/op
BenchmarkUpdate
BenchmarkUpdate/ORM-Update
BenchmarkUpdate/ORM-Update-8                  	     295	  11141763 ns/op	    7690 B/op	     128 allocs/op
BenchmarkUpdate/Native-Update
BenchmarkUpdate/Native-Update-8               	     306	   4011816 ns/op	    6079 B/op	      91 allocs/op
BenchmarkInsert
BenchmarkInsert/ORM-InsertOne
BenchmarkInsert/ORM-InsertOne-8               	   13258	     88546 ns/op	    7363 B/op	     123 allocs/op
BenchmarkInsert/Native-InsertOne
BenchmarkInsert/Native-InsertOne-8            	   13843	     84566 ns/op	    6046 B/op	      88 allocs/op
BenchmarkInsert/ORM-InsertMany
BenchmarkInsert/ORM-InsertMany-8              	     889	   3823420 ns/op	  255304 B/op	    4480 allocs/op
BenchmarkJoinQuery
BenchmarkJoinQuery/ORM-LookupJoin
BenchmarkJoinQuery/ORM-LookupJoin-8           	      19	  62841002 ns/op	 2022170 B/op	   39921 allocs/op
BenchmarkJoinQuery/Native-LookupJoin
BenchmarkJoinQuery/Native-LookupJoin-8        	      16	  63172367 ns/op	 2445467 B/op	   53703 allocs/op
BenchmarkRepository
BenchmarkRepository/Repository-FindByID
BenchmarkRepository/Repository-FindByID-8     	   12176	     97524 ns/op	    9287 B/op	     142 allocs/op
BenchmarkRepository/Repository-FindPage
BenchmarkRepository/Repository-FindPage-8     	     532	   5194989 ns/op	   40738 B/op	     943 allocs/op
BenchmarkRepository/Repository-UpdateByID
BenchmarkRepository/Repository-UpdateByID-8   	   10280	    110071 ns/op	    7657 B/op	     132 allocs/op
BenchmarkConcurrent
BenchmarkConcurrent/ORM-ConcurrentRead
BenchmarkConcurrent/ORM-ConcurrentRead-8      	      26	  62581518 ns/op	39395269 B/op	 1025371 allocs/op
BenchmarkConcurrent/ORM-ConcurrentWrite
BenchmarkConcurrent/ORM-ConcurrentWrite-8     	   35160	     33733 ns/op	    7372 B/op	     123 allocs/op
BenchmarkConcurrent/ORM-ConcurrentUpdate
BenchmarkConcurrent/ORM-ConcurrentUpdate-8    	     177	   7399120 ns/op	    7634 B/op	     129 allocs/op
BenchmarkMemoryUsage
BenchmarkMemoryUsage/ORM-LargeResultSet
BenchmarkMemoryUsage/ORM-LargeResultSet-8     	       2	 687059896 ns/op	136147088 B/op	 3514824 allocs/op
BenchmarkMemoryUsage/ORM-StreamingQuery
BenchmarkMemoryUsage/ORM-StreamingQuery-8     	       8	 137702141 ns/op	14462405 B/op	  392079 allocs/op
BenchmarkComplexConditions
BenchmarkComplexConditions/ORM-MultipleConditions
BenchmarkComplexConditions/ORM-MultipleConditions-8         	       3	 335019486 ns/op	56357880 B/op	 1458175 allocs/op
BenchmarkComplexConditions/Native-MultipleConditions
BenchmarkComplexConditions/Native-MultipleConditions-8      	       6	 205371618 ns/op	45832770 B/op	  779018 allocs/op
BenchmarkComplexConditions/ORM-RegexQuery
BenchmarkComplexConditions/ORM-RegexQuery-8                 	       3	 382153375 ns/op	75060738 B/op	 1956364 allocs/op
BenchmarkIndexPerformance
BenchmarkIndexPerformance/ORM-IndexedQuery
BenchmarkIndexPerformance/ORM-IndexedQuery-8                	      44	  32730887 ns/op	 4804644 B/op	  130477 allocs/op
BenchmarkIndexPerformance/ORM-NonIndexedQuery
BenchmarkIndexPerformance/ORM-NonIndexedQuery-8             	      24	  47871872 ns/op	    8132 B/op	     105 allocs/op
BenchmarkIndexPerformance/ORM-CompoundIndexQuery
BenchmarkIndexPerformance/ORM-CompoundIndexQuery-8          	       1	1599281500 ns/op	240992136 B/op	 6243937 allocs/op
BenchmarkDelete
BenchmarkDelete/ORM-DeleteMany
BenchmarkDelete/ORM-DeleteMany-8                            	   12594	     96674 ns/op	    5843 B/op	      95 allocs/op
BenchmarkDelete/Native-DeleteMany
BenchmarkDelete/Native-DeleteMany-8                         	   13174	     94205 ns/op	    4384 B/op	      60 allocs/op
BenchmarkDelete/ORM-DeleteOne
BenchmarkDelete/ORM-DeleteOne-8                             	    8694	    126434 ns/op	    5842 B/op	      95 allocs/op
BenchmarkPagination
BenchmarkPagination/ORM-Pagination
BenchmarkPagination/ORM-Pagination-8                        	     498	   2897570 ns/op	   38169 B/op	     931 allocs/op
BenchmarkPagination/Native-Pagination
BenchmarkPagination/Native-Pagination-8                     	     438	   3315105 ns/op	   45480 B/op	     872 allocs/op
BenchmarkCount
BenchmarkCount/ORM-Count
BenchmarkCount/ORM-Count-8                                  	      58	  18980491 ns/op	    7609 B/op	     118 allocs/op
BenchmarkCount/Native-Count
BenchmarkCount/Native-Count-8                               	      13	  82383907 ns/op	    6720 B/op	      82 allocs/op
BenchmarkAggregate
BenchmarkAggregate/ORM-SimpleAggregate
BenchmarkAggregate/Native-SimpleAggregate
BenchmarkAggregate/Native-SimpleAggregate-8                 	      24	  48347632 ns/op	   65619 B/op	    1276 allocs/op
BenchmarkAggregate/ORM-ComplexAggregate
BenchmarkAggregate/ORM-ComplexAggregate-8                   	      21	  51887315 ns/op	   25726 B/op	     417 allocs/op
BenchmarkAggregate/Native-ComplexAggregate
BenchmarkAggregate/Native-ComplexAggregate-8                	      21	  52452042 ns/op	   21222 B/op	     439 allocs/op
```

## 安装

```bash
go get github.com/zzjha-cn/tmORM
```

## 快速开始
### 连接数据库
```go
import (
    "context"
    tmorm "github.com/zzjha-cn/tm_orm"
    "github.com/zzjha-cn/tm_orm/collection"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

// 创建 MongoDB 客户端
client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
if err != nil {
    panic(err)
}

// 创建 ORM 客户端
config := tmorm.DefaultClientConfig("mongodb://localhost:27017")
ormClient, err := tmorm.NewORMClient(config)
if err != nil {
    panic(err)
}
```
### 基本查询

```go
import (
    "github.com/zzjha-cn/tm_orm/collection"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

// 定义模型
type User struct {
    ID   primitive.ObjectID `bson:"_id,omitempty"`
    Name string            `bson:"name"`
    Age  int               `bson:"age"`
    Email string           `bson:"email"`
}

// 创建集合操作器
coll := collection.NewCollection[User](ormClient, "mydb", "users")

// 基本查询
users, err := coll.Where("name").Eq("sean").Where("age").Gte(18).Find(ctx)
if err != nil {
    panic(err)
}

// 查询单个文档
user, err := coll.Where("_id").Eq(userID).FindOne(ctx)
if err != nil {
    panic(err)
}

// 统计文档数量
count, err := coll.Where("age").Gte(18).Count(ctx)
```

### 复杂查询

#### 多条件查询
```go
// 链式条件查询
users, err := coll.
    Where("age").Gte(18).Where("age").Lte(65).
    Where("name").Regex("^[A-Z]").
    Where("email").Exists(true).
    Find(ctx)

// 使用 In 操作符
users, err := coll.Where("city").In("Beijing", "Shanghai", "Guangzhou").Find(ctx)

// 使用 Filter 进行复杂条件
import "go.mongodb.org/mongo-driver/bson"

filter := bson.M{
    "$and": []bson.M{
        {"age": bson.M{"$gte": 18}},
        {"$or": []bson.M{
            {"name": "sean"},
            {"email": bson.M{"$regex": "@example.com$"}},
        }},
    },
}
users, err := coll.Filter(filter).Find(ctx)
```

#### 表达式查询
```go
import "github.com/zzjha-cn/tm_orm/expression"

// 使用表达式构建复杂查询
expr := expression.NewExpression().Field("name").Eq("sean").
    Or().Field("age").Gt(25).And().Field("email").Regex("@company.com$")

users, err := coll.WhereExpr(expr).Find(ctx)

// 或者使用组合表达式
expr1 := expression.NewExpression().Field("name").Eq("sean")
expr2 := expression.NewExpression().Field("age").Gt(25)
combinedExpr := expression.Or(expr1, expr2)

users, err := coll.WhereExpr(combinedExpr).Find(ctx)

// 使用链式表达式构建器
users, err := coll.Expr("age").Gt(25).And("status").Eq("active").Find(ctx)
```

### 更新操作

```go
// 单字段更新
result, err := coll.Where("_id").Eq(userID).Set("age", 25).UpdateOne(ctx)
if err != nil {
    panic(err)
}

// 多字段更新
result, err := coll.Where("name").Eq("sean").
    Set("age", 26).
    Set("email", "sean@example.com").
    Update(ctx)

// 数值操作
result, err := coll.Where("_id").Eq(userID).
    Inc("age", 1).        // 年龄加1
    Mul("salary", 1.1).   // 薪资乘以1.1
    UpdateOne(ctx)

// 数组操作
result, err := coll.Where("_id").Eq(userID).
    Push("tags", "new-tag").              // 添加到数组
    Pull("old_tags", "deprecated-tag").   // 从数组移除
    AddToSet("skills", "golang").         // 添加到集合（去重）
    UpdateOne(ctx)

// 字段重命名和删除
result, err := coll.Where("_id").Eq(userID).
    Rename("old_field", "new_field").     // 重命名字段
    Unset("deprecated_field").            // 删除字段
    UpdateOne(ctx)

// 使用 UpdateBuilder 进行复杂更新
import "github.com/zzjha-cn/tm_orm/collection"

builder := collection.NewUpdateBuilder().
    Set("name", "new name").
    Inc("age", 1).
    Push("tags", "tag1")

result, err := coll.Where("_id").Eq(userID).UpdateWithBuilder(ctx, builder)
```

### 插入操作

```go
// 插入单个文档
user := &User{
    Name:  "sean",
    Age:   25,
    Email: "sean@example.com",
}
result, err := coll.Insert(ctx, user)
if err != nil {
    panic(err)
}

// 批量插入
users := []*User{
    {Name: "alice", Age: 23, Email: "alice@example.com"},
    {Name: "bob", Age: 27, Email: "bob@example.com"},
}
result, err := coll.InsertMany(ctx, users)
```

### 删除操作

```go
// 删除单个文档
result, err := coll.Where("_id").Eq(userID).DeleteOne(ctx)

// 批量删除
result, err := coll.Where("age").Lt(18).Delete(ctx)
```

### Repository 模式

```go
import "github.com/zzjha-cn/tm_orm/repo"

// 创建 Repository
repo := repo.NewRepository[User](ormClient, "mydb", "users")

// 基本 CRUD 操作
user := &User{Name: "sean", Age: 25, Email: "sean@example.com"}

// 创建
createdUser, err := repo.Create(ctx, user)

// 根据 ID 查找
foundUser, err := repo.FindByID(ctx, createdUser.ID)

// 根据 ID 更新
updates := map[string]any{"age": 26, "email": "newemail@example.com"}
updatedUser, err := repo.UpdateByID(ctx, createdUser.ID, updates)

// 根据 ID 删除
err = repo.DeleteByID(ctx, createdUser.ID)

// 分页查询
users, total, err := repo.FindPage(ctx, 1, 10) // 第1页，每页10条

// 检查是否存在
exists, err := repo.Exists(ctx, createdUser.ID)

// 统计数量
count, err := repo.Count(ctx)
```

### 聚合操作

```go
import "github.com/zzjha-cn/tm_orm/aggregator"

// 创建聚合器
agg := aggregator.NewAggregator[User](ormClient, "mydb", "users")

// 基本聚合操作
results, err := agg.
    Match(bson.M{"age": bson.M{"$gte": 18}}).
    Group("$department", bson.M{
        "avgAge": bson.M{"$avg": "$age"},
        "count": bson.M{"$sum": 1},
    }).
    SortDesc("avgAge").
    Execute(ctx)

if err != nil {
    panic(err)
}

// 使用聚合操作符
results, err := agg.
    Match(bson.M{"status": "active"}).
    Project(bson.M{
        "name": 1,
        "age": 1,
        "ageGroup": bson.M{
            "$cond": bson.M{
                "if":   bson.M{"$lt": []any{"$age", 30}},
                "then": "young",
                "else": "senior",
            },
        },
    }).
    Execute(ctx)

// 关联查询 (Lookup)
type OrderResult struct {
    ID       primitive.ObjectID `bson:"_id"`
    Amount   float64           `bson:"amount"`
    UserInfo User              `bson:"user_info"`
}

orderAgg := aggregator.NewAggregator[OrderResult](ormClient, "mydb", "orders")
orders, err := orderAgg.
    Match(bson.M{"status": "completed"}).
    Lookup("users", "user_id", "_id", "user_info").
    Unwind("$user_info").
    Project(bson.M{
        "_id": 1,
        "amount": 1,
        "user_name": "$user_info.name",
        "user_email": "$user_info.email",
    }).
    Execute(ctx)
```
### 复杂聚合
```go
// 多级分组和条件聚合
results, err := agg.
    Match(bson.M{"age": bson.M{"$gte": 20}}).
    Group(bson.M{
        "department": "$department",
        "ageRange": bson.M{
            "$cond": bson.M{
                "if":   bson.M{"$lte": []any{"$age", 25}},
                "then": "young",
                "else": "senior",
            },
        },
    }, bson.M{
        "avgAge": bson.M{"$avg": "$age"},
        "names":  bson.M{"$push": "$name"},
        "count":  bson.M{"$sum": 1},
    }).
    SortDesc("avgAge").
    Execute(ctx)

// 分面聚合 (Facet)
facetResults, err := agg.
    Match(bson.M{"status": "active"}).
    Facet(bson.M{
        "ageStats": []bson.M{
            {"$group": bson.M{
                "_id": nil,
                "avgAge": bson.M{"$avg": "$age"},
                "minAge": bson.M{"$min": "$age"},
                "maxAge": bson.M{"$max": "$age"},
            }},
        },
        "departmentCount": []bson.M{
            {"$group": bson.M{
                "_id": "$department",
                "count": bson.M{"$sum": 1},
            }},
            {"$sort": bson.M{"count": -1}},
        },
    }).
    Execute(ctx)

// 分页聚合
page := int64(1)
pageSize := int64(10)
results, total, err := agg.
    Match(bson.M{"age": bson.M{"$gte": 18}}).
    SortAsc("name").
    Paginate(ctx, page, pageSize)
```

### 中间件使用

```go
import (
    "context"
    "fmt"
    "time"
    tmorm "github.com/zzjha-cn/tm_orm"
    "github.com/zzjha-cn/tm_orm/collection"
)

// 创建自定义中间件
loggingMiddleware := func(mctx *tmorm.MiddleCtx, next func(m *tmorm.MiddleCtx)) {
    start := time.Now()
    fmt.Printf("[%s] Starting operation: %s\n", start.Format(time.RFC3339), mctx.Typ)

    next(mctx)

    duration := time.Since(start)
    if mctx.Result.Err != nil {
        fmt.Printf("[%s] Operation %s failed after %v: %v\n",
            time.Now().Format(time.RFC3339), mctx.Typ, duration, mctx.Result.Err)
    } else {
        fmt.Printf("[%s] Operation %s completed in %v\n",
            time.Now().Format(time.RFC3339), mctx.Typ, duration)
    }
}

// 创建带中间件的 ORM 客户端
config := tmorm.DefaultClientConfig()
config.URI = "mongodb://localhost:27017"
ormClient, err := tmorm.NewORMClient(config)
if err != nil {
    panic(err)
}

// 为collection创建中间件
coll := collection.NewCollection[User](ormClient, "mydb", "users", loggingMiddleware)

// 所有通过此客户端的操作都会经过中间件处理
users, err := coll.Where("age").Gte(18).Find(ctx) // 会触发中间件
```

## 支持的查询操作符

框架支持丰富的MongoDB查询操作符，包括：

- 比较操作符：`$eq`, `$gt`, `$gte`, `$lt`, `$lte`, `$ne`, `$in`, `$nin`
- 逻辑操作符：`$and`, `$or`
- 元素操作符：`$exists`, `$type`
- 评估操作符：`$expr`, `$regex`, `$mod`
- 数组操作符：`$all`, `$elemMatch`, `$size`

更多操作符支持详见[文档](docs/support_exprssion.md)。

## 许可证

本项目采用MIT许可证。详见[LICENSE](LICENSE)文件。
